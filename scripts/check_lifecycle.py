"""Exercise real lifecycle preparation, receipt guards and aggregate SQL locally.

Uses independent normalized-event fixtures and embedded ClickHouse. It does not
execute a SQL lifecycle-preparation backend or connect to Databricks.
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


def expand(defaults, recipes):
    events = []
    for recipe in recipes:
        for i in range(recipe.get("repeat", 1)):
            event = {**defaults, **{k: v for k, v in recipe.items() if k != "repeat"}}
            events.append({k: v.replace("{i}", str(i)) if isinstance(v, str) else v
                           for k, v in event.items()})
    return events


def main():
    go = shutil.which("go")
    if not go:
        raise SystemExit("Go is required on PATH")
    fixtures = json.loads((ROOT / "lifecycle/testdata/cases.json").read_text())
    definition = ROOT / "examples/lifecycle/measurement.yaml"
    with tempfile.TemporaryDirectory(prefix="goo-lifecycle-") as directory:
        temp = Path(directory)
        binary = temp / "goo"
        subprocess.run([go, "build", "-o", str(binary), "./cmd/goo"], cwd=ROOT, check=True)

        def cli(command, *args):
            r = subprocess.run([str(binary), command, *map(str, args)], cwd=ROOT,
                               check=False, capture_output=True, text=True)
            if r.returncode:
                raise RuntimeError(f'{command} exited {r.returncode}: {r.stderr.strip()}')
            return json.loads(r.stdout)

        db = session.Session(str(temp / "clickhouse"))
        try:
            db.query("""CREATE TABLE reservation_subject_facts (
                evaluation_id String, environment String, tenant_id String,
                reservation_id String, service_name String, release String,
                classification String, quality_impaired Bool, is_mature Bool,
                in_cohort Bool
            ) ENGINE=Memory""")
            snapshots = {}
            comparisons = 0
            for case in fixtures["cases"]:
                policy = {**fixtures["policy"], **case.get("policy", {})}
                policy_path = temp / "policy.yaml"
                policy_path.write_text(yaml.safe_dump(policy))
                events_path = temp / "events.json"
                events_path.write_text(json.dumps(expand(fixtures["eventDefaults"], case["events"])))
                prepare_args = ["-policy", policy_path, "-events", events_path]
                if policy.get("supersedesEvaluationId"):
                    prior = snapshots[case["previousCase"]]
                    assert prior["receipt"]["policy"]["evaluationId"] == policy["supersedesEvaluationId"]
                    prior_path = temp / "previous.json"
                    prior_path.write_text(json.dumps(prior))
                    prepare_args += ["-previous", prior_path]
                prepared = cli("prepare", *prepare_args)
                snapshots[case["name"]] = prepared
                receipt = prepared["receipt"]
                assert receipt["counts"] == case["counts"], (case["name"], receipt["counts"])
                assert receipt["state"] == case["state"], (case["name"], receipt["state"])
                assert receipt["cohort_start_ns"] == "57" and receipt["cohort_end_ns"] == "87"
                assert len(prepared["subjects"]) == case["subjectCount"]
                if "subjects" in case:
                    subjects = {s["tenant_id"] + "/" + s["subject_id"]:
                                [s["classification"], s["reason"], s["in_cohort"], s["is_mature"]]
                                for s in prepared["subjects"]}
                    assert subjects == case["subjects"], (case["name"], subjects)
                artifact = temp / "prepared.json"
                artifact.write_text(json.dumps(prepared))
                db.query("TRUNCATE TABLE reservation_subject_facts")
                if prepared["facts"]:
                    db.query("INSERT INTO reservation_subject_facts FORMAT JSONEachRow\n" +
                             "\n".join(json.dumps(f) for f in prepared["facts"]))
                for grouping in ("", "service,release"):
                    result = cli("evaluate-prepared", "-definition", definition,
                                 "-prepared", artifact, "-group-by", grouping)
                    plan = cli("compile", "-definition", definition,
                               "-target", "clickhouse", "-group-by", grouping)
                    evaluation = policy["evaluationId"].replace("\\", "\\\\").replace("'", "''")
                    sql = plan["query"].replace("{evaluation_id:String}", "'" + evaluation + "'")
                    sql = sql.replace("{evidence_impaired:Bool}",
                                      "true" if receipt["state"] == "unknown" else "false")
                    actual = [json.loads(line) for line in str(db.query(sql)).splitlines() if line]
                    groups = grouping.split(",") if grouping else []
                    expected_rows = {tuple(row["dimensions"][g] for g in groups): row["measures"]
                                     for row in result["results"]}
                    sql_rows = {tuple(row[g] for g in groups): row for row in actual}
                    assert sql_rows.keys() == expected_rows.keys(), (case["name"], grouping)
                    for key, measures in expected_rows.items():
                        row = sql_rows[key]
                        for name, value in measures.items():
                            if "count" in value:
                                assert row[name] == value["count"], (case["name"], name, row)
                            else:
                                assert row[name + "_state"] == value["state"]
                                if "ratio" not in value:
                                    assert row[name] is None
                                else:
                                    assert math.isclose(row[name], value["ratio"], rel_tol=1e-12, abs_tol=1e-12)
                        if not groups:
                            for name in ("eligible", "good", "bad", "unknown"):
                                assert row[name] == case["counts"][name]
                            assert row["success_state"] == case["state"]
                            if case["ratio"] is None:
                                assert row["success"] is None
                            else:
                                assert math.isclose(row["success"], case["ratio"], abs_tol=1e-12)
                    comparisons += 1
                print("PASS", case["name"])

            before = snapshots["original_snapshot_before_late_delivery"]
            after = snapshots["corrected_snapshot_after_late_delivery"]
            assert before["receipt"]["snapshot_digest"] != after["receipt"]["snapshot_digest"]
            db.query("TRUNCATE TABLE reservation_subject_facts")
            db.query("INSERT INTO reservation_subject_facts FORMAT JSONEachRow\n" +
                     "\n".join(json.dumps(f) for f in before["facts"] + after["facts"]))
            plan = cli("compile", "-definition", definition, "-target", "clickhouse")
            for snapshot in (before, after):
                receipt = snapshot["receipt"]
                sql = plan["query"].replace("{evaluation_id:String}",
                                             "'" + receipt["policy"]["evaluationId"] + "'")
                sql = sql.replace("{evidence_impaired:Bool}", "false")
                row = json.loads(str(db.query(sql)))
                assert row["eligible"] == 1 and row["good"] == receipt["counts"]["good"]
            print("PASS original/corrected snapshots coexist without double counting")

            plan = cli("compile", "-definition", definition, "-target", "databricks")
            assert ":evidence_impaired" in plan["query"]
            assert yaml.safe_load(plan["metricYaml"])["filter"] == "`in_cohort` = true"
            print(f"PASS {comparisons} lifecycle aggregate comparisons; Databricks output not executed")
        finally:
            db.close()


if __name__ == "__main__":
    main()
