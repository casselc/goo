"""Compare the Go reference evaluator and generated SQL with independent fixtures.

Requires Go on PATH plus requirements-dev.txt. Uses embedded ClickHouse (chDB),
not a Databricks emulator. Generated Databricks YAML is structurally checked only.
"""
import json
import math
from pathlib import Path
import shutil
import subprocess
import tempfile

from chdb import session
import yaml

ROOT = Path(__file__).resolve().parents[1]
COUNTS = ("eligible", "good", "bad", "unknown", "impaired")


def expand(blocks):
    facts = []
    for block in blocks:
        for i in range(block["count"]):
            fact = dict(evaluation_id="basis-1", environment="test", tenant_id="tenant-a",
                        reservation_id=f'{block["id"]}-{i}', service_name="reservations",
                        quality_impaired=False, is_mature=True)
            fact.update({k: v for k, v in block.items() if k not in ("count", "id")})
            facts.append(fact)
    return facts


def assert_expected(got, expected):
    for name in COUNTS:
        assert got[name] == expected[name], (name, got, expected)
    assert got["state"] == expected["state"], (got, expected)
    parts = expected["ratio_parts"]
    if parts is None:
        assert got["ratio"] is None, got
    else:
        assert got["ratio"] is not None and math.isclose(
            got["ratio"], parts[0] / parts[1], rel_tol=1e-12, abs_tol=1e-12), (got, expected)


def sql_literal(value):
    # Only the fixture runner substitutes this known typed binding because the
    # embedded session API does not offer the server HTTP parameter interface.
    return "'" + value.replace("\\", "\\\\").replace("'", "''") + "'"


def main():
    go = shutil.which("go")
    if not go:
        raise SystemExit("Go is required on PATH; see docs/PORTABLE_COMPILER.md")
    with tempfile.TemporaryDirectory(prefix="goo-parity-") as directory:
        temp = Path(directory)
        binary = temp / "goo"
        subprocess.run([go, "build", "-o", str(binary), "./cmd/goo"], cwd=ROOT, check=True)
        definition = ROOT / "examples/portable/reservations.yaml"

        def cli(command, *args, definition_path=definition):
            result = subprocess.run([str(binary), command, "-definition", str(definition_path), *args],
                                    cwd=ROOT, check=True, text=True, capture_output=True)
            return json.loads(result.stdout)

        db = session.Session(str(temp / "clickhouse"))
        try:
            db.query("""CREATE TABLE reservation_subject_facts (
                evaluation_id String, environment String, tenant_id String,
                reservation_id String, service_name String, release String,
                classification String, quality_impaired Bool, is_mature Bool
            ) ENGINE=Memory""")
            cases = json.loads((ROOT / "measurement/testdata/cohorts.json").read_text())["cases"]
            comparisons = 0
            for case in cases:
                facts = expand(case["blocks"])
                path = temp / "facts.json"
                path.write_text(json.dumps(facts))
                db.query("TRUNCATE TABLE reservation_subject_facts")
                if facts:
                    db.query("INSERT INTO reservation_subject_facts FORMAT JSONEachRow\n" +
                             "\n".join(json.dumps(f) for f in facts))
                for grouping in ("", "release"):
                    args = ("-group-by", grouping)
                    reference = cli("evaluate", "-facts", str(path), "-evaluation", "basis-1", *args)
                    plan = cli("compile", "-target", "clickhouse", *args)
                    assert "LIMIT" not in plan["query"]
                    sql = plan["query"].replace("{evaluation_id:String}", sql_literal("basis-1"))
                    sql = sql.replace("{evidence_impaired:Bool}", "false")
                    actual = [json.loads(line) for line in str(db.query(sql)).splitlines() if line]
                    expected = case["by_release"] if grouping else {"total": case["total"]}
                    ref_rows = {}
                    for row in reference:
                        measures = row["measures"]
                        key = row["dimensions"]["release"] if grouping else "total"
                        ref_rows[key] = {**{n: measures[n]["count"] for n in COUNTS},
                                         "state": measures["success"]["state"],
                                         "ratio": measures["success"].get("ratio")}
                    sql_rows = {}
                    for row in actual:
                        key = row["release"] if grouping else "total"
                        sql_rows[key] = {**{n: row[n] for n in COUNTS},
                                         "state": row["success_state"], "ratio": row["success"]}
                    assert ref_rows.keys() == sql_rows.keys() == expected.keys(), (case["name"], grouping)
                    for key, want in expected.items():
                        assert_expected(ref_rows[key], want)
                        assert_expected(sql_rows[key], want)
                    comparisons += 1
                print("PASS", case["name"], "(total and release groups)")

            # A new measure and a hostile literal exercise the generic compiler,
            # not a mode-specific branch or merely an output snapshot.
            extended = yaml.safe_load(definition.read_text())
            literal = "v' OR 1=1 -- \\ path"
            extended["measures"]["chosen_release"] = {
                "op": "count", "where": {"field": "release", "eq": literal}}
            # This measure deliberately shares a physical field name. SQL
            # source qualification must prevent alias substitution in filters.
            extended["measures"]["service_name"] = {
                "op": "count", "where": {"field": "classification", "eq": "good"}}
            extended_path = temp / "extended.yaml"
            extended_path.write_text(yaml.safe_dump(extended))
            facts = expand([{"count": 1, "id": "chosen", "release": literal, "classification": "good"},
                            {"count": 2, "id": "other", "release": "v1", "classification": "bad"}])
            path.write_text(json.dumps(facts))
            db.query("TRUNCATE TABLE reservation_subject_facts")
            db.query("INSERT INTO reservation_subject_facts FORMAT JSONEachRow\n" +
                     "\n".join(json.dumps(f) for f in facts))
            reference = cli("evaluate", "-facts", str(path), "-evaluation", "basis-1",
                            definition_path=extended_path)
            plan = cli("compile", "-target", "clickhouse", definition_path=extended_path)
            sql = plan["query"].replace("{evaluation_id:String}", sql_literal("basis-1"))
            sql = sql.replace("{evidence_impaired:Bool}", "false")
            actual = json.loads(str(db.query(sql)))
            assert actual["chosen_release"] == reference[0]["measures"]["chosen_release"]["count"] == 1
            assert actual["eligible"] == 3
            assert actual["service_name"] == 1
            grouped = cli("compile", "-target", "clickhouse", "-group-by", "service",
                          definition_path=extended_path)
            grouped_sql = grouped["query"].replace("{evaluation_id:String}", sql_literal("basis-1"))
            grouped_sql = grouped_sql.replace("{evidence_impaired:Bool}", "false")
            assert json.loads(str(db.query(grouped_sql)))["service"] == "reservations"
            print("PASS declarative extension and escaped literal")

            for grouping in ("", "release"):
                plan = cli("compile", "-target", "databricks", "-group-by", grouping)
                view = yaml.safe_load(plan["metricYaml"])
                assert view["version"] == "1.1"
                assert len(view["measures"]) == 6
                assert ":evaluation_id" in plan["query"]
                assert plan["status"] == "requires-prepared-source"
                assert "CREATE VIEW" in plan["ddl"] and "OR REPLACE" not in plan["ddl"]
            print(f"PASS {comparisons} fixture/group comparisons plus extension; "
                  "Databricks artifacts structurally checked, not executed")
        finally:
            db.close()


if __name__ == "__main__":
    main()
