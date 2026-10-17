package measurement

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func example(t *testing.T) *Definition {
	t.Helper()
	f, err := os.Open("../examples/portable/reservations.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d, err := Read(f)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func fact(id, class, release string) Fact {
	return Fact{"evaluation_id": "basis-1", "environment": "test", "tenant_id": "tenant-a",
		"reservation_id": id, "service_name": "reservations", "release": release,
		"classification": class, "quality_impaired": false, "is_mature": true}
}

func TestFullPopulationAndWeightedRollup(t *testing.T) {
	d := example(t)
	facts := []Fact{}
	for i := 0; i < 1000; i++ {
		facts = append(facts, fact(string(rune(i+1000)), "good", "v1"))
	}
	facts = append(facts, fact("bad", "bad", "v2"))
	other := fact("bad", "good", "v2")
	other["evaluation_id"] = "basis-2"
	facts = append(facts, other)
	pending := fact("pending", "pending", "v2")
	pending["is_mature"] = false
	facts = append(facts, pending)
	rows, err := d.Evaluate(facts, "basis-1", nil)
	if err != nil {
		t.Fatal(err)
	}
	m := rows[0].Measures
	if *m["eligible"].Count != 1001 || *m["good"].Count != 1000 || *m["success"].Ratio != 1000.0/1001 {
		t.Fatalf("incorrect full-population totals: %+v", m)
	}
	groups, err := d.Evaluate(facts, "basis-1", []string{"release"})
	if err != nil || len(groups) != 2 {
		t.Fatalf("groups: %v, %v", groups, err)
	}
	if *groups[0].Measures["success"].Ratio != 1 || *groups[1].Measures["success"].Ratio != 0 {
		t.Fatal("unexpected group ratios")
	}
	if *m["success"].Ratio == 0.5 {
		t.Fatal("must not average group ratios")
	}
}

func TestQualityAndNoData(t *testing.T) {
	d := example(t)
	for _, kind := range []string{"unknown", "impaired", "no_data"} {
		t.Run(kind, func(t *testing.T) {
			facts := []Fact{}
			if kind != "no_data" {
				f := fact("x", "good", "v1")
				if kind == "unknown" {
					f["classification"] = "unknown"
				} else {
					f["quality_impaired"] = true
				}
				facts = append(facts, f)
			}
			rows, err := d.Evaluate(facts, "basis-1", nil)
			if err != nil {
				t.Fatal(err)
			}
			want := "unknown"
			if kind == "no_data" {
				want = kind
			}
			v := rows[0].Measures["success"]
			if v.State != want || v.Ratio != nil {
				t.Fatalf("quality result %+v, want %s and no numeric ratio", v, want)
			}
		})
	}
	rows, err := d.Evaluate(nil, "basis-1", []string{"release"})
	if err != nil || len(rows) != 0 {
		t.Fatal("empty grouped input must not invent a dimension member")
	}
}

func TestScopedIdentityAndInputRejection(t *testing.T) {
	d := example(t)
	a, b := fact("same", "good", "v1"), fact("same", "bad", "v1")
	b["tenant_id"] = "tenant-b"
	rows, err := d.Evaluate([]Fact{a, b}, "basis-1", nil)
	if err != nil || *rows[0].Measures["eligible"].Count != 2 {
		t.Fatal("cross-tenant IDs must remain distinct")
	}
	for _, kind := range []string{"duplicate", "empty_key", "missing", "wrong_type", "bad_enum", "hidden_duplicate"} {
		t.Run(kind, func(t *testing.T) {
			x := fact("same", "good", "v1")
			facts := []Fact{x}
			switch kind {
			case "duplicate":
				facts = append(facts, fact("same", "bad", "v2"))
			case "empty_key":
				x["tenant_id"] = ""
			case "missing":
				delete(x, "classification")
			case "wrong_type":
				x["quality_impaired"] = "false"
			case "bad_enum":
				x["classification"] = "healthy-ish"
			case "hidden_duplicate":
				x["is_mature"] = false
				facts = append(facts, fact("same", "good", "v1"))
			}
			if _, err := d.Evaluate(facts, "basis-1", nil); err == nil {
				t.Fatal("invalid prepared source accepted")
			}
		})
	}
	if _, err := d.Evaluate(nil, "", nil); err == nil {
		t.Fatal("evaluation binding required")
	}
}

func TestDefinitionRejections(t *testing.T) {
	for _, kind := range []string{"op", "field", "literal", "missing_ref", "ratio_ref", "denominator_filter", "name", "table", "dimension", "enum"} {
		t.Run(kind, func(t *testing.T) {
			d := example(t)
			switch kind {
			case "op":
				d.Measures["new"] = Measure{Op: "sql"}
			case "field":
				d.Population.Field = "missing"
			case "literal":
				d.Population.Eq = "true"
			case "missing_ref":
				d.Measures["success"] = Measure{Op: "ratio", Numerator: "missing", Denominator: "eligible"}
			case "ratio_ref":
				d.Measures["another"] = Measure{Op: "ratio", Numerator: "success", Denominator: "eligible"}
			case "denominator_filter":
				d.Measures["eligible"] = Measure{Op: "count", Where: &Predicate{"is_mature", true}}
			case "name":
				d.Name = "x; DROP TABLE x"
			case "table":
				d.Source.Bindings["clickhouse"] = "x WHERE true"
			case "dimension":
				d.Dimensions["success_state"] = "release"
			case "enum":
				d.Source.Enums["classification"] = []string{"good", "good"}
			}
			if err := d.Validate(); err == nil {
				t.Fatal("invalid definition accepted")
			}
		})
	}
	b, _ := os.ReadFile("../examples/portable/reservations.yaml")
	for _, suffix := range []string{"\nunknownField: true\n", "\nname: duplicate\n", "\n---\n{}\n"} {
		if _, err := Read(strings.NewReader(string(b) + suffix)); err == nil {
			t.Fatal("unknown field, duplicate key, or multiple documents accepted")
		}
	}
}

func TestDeclarativeExtensionAndTargetContracts(t *testing.T) {
	d := example(t)
	// Adding a measurement changes data, not compiler branches.
	d.Measures["v2_subjects"] = Measure{Op: "count", Where: &Predicate{"release", "v2"}}
	for _, target := range []string{"clickhouse", "databricks"} {
		p, err := d.Compile(target, []string{"release"})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(p.Query, "v2_subjects") || strings.Contains(p.Query, "LIMIT") || len(p.Requirements) == 0 {
			t.Fatal("missing declarative count, full-population query or requirements")
		}
		if target == "databricks" {
			var view metricView
			if err := yaml.Unmarshal([]byte(p.MetricYAML), &view); err != nil || view.Version != "1.1" {
				t.Fatal("generated YAML is invalid")
			}
			if !strings.Contains(p.Query, ":evaluation_id") || !strings.Contains(p.DDL, "WITH METRICS") || strings.Contains(p.DDL, "OR REPLACE") {
				t.Fatal("unsafe DDL or missing evaluation binding")
			}
			if view.Measures[len(view.Measures)-1].Name != "success" {
				t.Fatal("atomic measures must precede composed ratios")
			}
		} else if !strings.Contains(p.Query, "{evaluation_id:String}") {
			t.Fatal("missing typed ClickHouse parameter")
		}
		first, _ := json.Marshal(p)
		secondPlan, _ := d.Compile(target, []string{"release"})
		second, _ := json.Marshal(secondPlan)
		if string(first) != string(second) {
			t.Fatal("nondeterministic compiler output")
		}
	}
	for _, groups := range [][]string{{"missing"}, {"release", "release"}} {
		if _, err := d.Compile("clickhouse", groups); err == nil {
			t.Fatal("invalid grouping accepted")
		}
	}
	delete(d.Source.Bindings, "databricks")
	if _, err := d.Compile("databricks", nil); err == nil {
		t.Fatal("missing backend binding accepted")
	}
}

func TestDollarDelimiterRejected(t *testing.T) {
	d := example(t)
	d.Measures["special"] = Measure{Op: "count", Where: &Predicate{"release", "$$"}}
	if _, err := d.Compile("databricks", nil); err == nil {
		t.Fatal("DDL delimiter must not be injected through a literal")
	}
}
