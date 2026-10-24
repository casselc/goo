package exploration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func numericPredicateFixture(typ, op, value string) *Predicate {
	return &Predicate{Numeric: &NumericBinding{Field: Field{Kind: "attribute", Key: "measure"}, Type: typ, Unit: "ms", Parse: "canonical-decimal"}, Op: op, Value: value}
}
func TestNumericPredicateValidation(t *testing.T) {
	for _, typ := range []string{"int64", "uint64"} {
		for _, op := range []string{"eq", "ne", "gt", "gte", "lt", "lte", "valid", "missing", "empty", "invalid"} {
			value := "0"
			if op == "valid" || op == "missing" || op == "empty" || op == "invalid" {
				value = ""
			}
			q := basic()
			q.Where = numericPredicateFixture(typ, op, value)
			if e := q.Validate(); e != nil {
				t.Fatal(typ, op, e)
			}
		}
	}
	for _, mutate := range []func(*Predicate){
		func(p *Predicate) { p.Field = &Field{Kind: "attribute", Key: "other"} },
		func(p *Predicate) { p.All = []Predicate{} }, func(p *Predicate) { p.Any = []Predicate{} },
		func(p *Predicate) { p.Numeric.Field.Kind = "column" }, func(p *Predicate) { p.Numeric.Type = "float64" },
		func(p *Predicate) { p.Numeric.Unit = "" }, func(p *Predicate) { p.Numeric.Parse = "permissive" },
		func(p *Predicate) { p.Op = "contains" }, func(p *Predicate) { p.Op = "exists" },
		func(p *Predicate) { p.Op = "valid"; p.Value = "0" },
	} {
		q := basic()
		q.Where = numericPredicateFixture("int64", "eq", "0")
		mutate(q.Where)
		if _, e := Compile(q); e == nil {
			t.Fatal("malformed numeric predicate compiled", q.Where)
		}
	}
	for _, typ := range []string{"int64", "uint64"} {
		invalid := []string{"", "01", "-0", "+1", " 1", "1 ", "1.0", "1e3", "NaN", "∞", "１", "92233720368547758080", "1\x00"}
		if typ == "int64" {
			invalid = append(invalid, "9223372036854775808", "-9223372036854775809")
		} else {
			invalid = append(invalid, "-1", "18446744073709551616")
		}
		for _, value := range invalid {
			q := basic()
			q.Where = numericPredicateFixture(typ, "eq", value)
			if q.Validate() == nil {
				t.Fatal("invalid threshold accepted", typ, value)
			}
		}
	}
	// Preserve the historical untyped uint64 threshold contract, including leading
	// zero equality: the new canonical contract is opt-in through Numeric only.
	q := basic()
	q.Where = &Predicate{Field: &Field{Kind: "column", Key: "duration_ns"}, Op: "eq", Value: "01"}
	if q.Validate() != nil || !Match(q.Where, map[Field]string{*q.Where.Field: "1"}) {
		t.Fatal("legacy column behavior changed")
	}
	wire, e := json.Marshal(q.Where)
	if e != nil || string(wire) != `{"field":{"kind":"column","key":"duration_ns"},"op":"eq","value":"01"}` {
		t.Fatal("absent numeric discriminator changed legacy predicate bytes", string(wire), e)
	}
}

func TestNumericPredicateCoverageAndExactComparisons(t *testing.T) {
	values := []*string{nil}
	for _, v := range []string{"", "0", "-0", "01", "+1", "1 ", "1.0", "-1", "1", "9007199254740993", "9223372036854775807", "-9223372036854775808", "18446744073709551615", "18446744073709551616", "-9223372036854775809"} {
		x := v
		values = append(values, &x)
	}
	for _, typ := range []string{"int64", "uint64"} {
		p := numericPredicateFixture(typ, "valid", "")
		summary, e := EvaluateNumericValues(*p.Numeric, values)
		if e != nil {
			t.Fatal(e)
		}
		counts := map[string]int{"valid": 0, "missing": 0, "empty": 0, "invalid": 0}
		for _, raw := range values {
			row := map[Field]string{}
			if raw != nil {
				row[p.Numeric.Field] = *raw
			}
			matched := 0
			for _, op := range []string{"valid", "missing", "empty", "invalid"} {
				p.Op = op
				if Match(p, row) {
					counts[op]++
					matched++
				}
			}
			if matched != 1 {
				t.Fatal("numeric states do not partition record", typ, raw, matched)
			}
			p.Op = "ne"
			p.Value = "0"
			if (raw == nil || *raw == "" || *raw == "01" || *raw == "-0") && Match(p, row) {
				t.Fatal("invalid value matched ne", raw)
			}
			p.Value = ""
		}
		if fmt.Sprint(counts["valid"]) != summary.Valid || fmt.Sprint(counts["missing"]) != summary.Missing || fmt.Sprint(counts["empty"]) != summary.Empty || fmt.Sprint(counts["invalid"]) != summary.Invalid {
			t.Fatal("predicate/aggregate coverage disagreement", counts, summary)
		}
	}
	for _, tc := range []struct {
		typ, raw, op, threshold string
		want                    bool
	}{
		{"int64", "9007199254740993", "gt", "9007199254740992", true},
		{"uint64", "18446744073709551615", "eq", "18446744073709551615", true},
		{"int64", "-9223372036854775808", "lt", "-9223372036854775807", true},
		{"int64", "9223372036854775807", "gte", "9223372036854775807", true},
		{"uint64", "18446744073709551615", "lt", "18446744073709551615", false},
		{"uint64", "-1", "ne", "0", false},
	} {
		p := numericPredicateFixture(tc.typ, tc.op, tc.threshold)
		if got := Match(p, map[Field]string{p.Numeric.Field: tc.raw}); got != tc.want {
			t.Fatal(tc, got)
		}
	}
}

func TestNumericPredicateCompileParameterizationAndPaging(t *testing.T) {
	for _, signal := range []string{"spans", "logs"} {
		for _, kind := range []string{"attribute", "resource"} {
			for _, typ := range []string{"int64", "uint64"} {
				q := basic()
				q.Signal = signal
				q.Where = numericPredicateFixture(typ, "ne", "9007199254740993")
				q.Where.Numeric.Field = Field{Kind: kind, Key: "weird'\\key"}
				q.Numeric = q.Where.Numeric
				plan, e := Compile(q)
				if e != nil {
					t.Fatal(e)
				}
				sqlType := "Int64"
				cast := "toInt64OrNull"
				if typ == "uint64" {
					sqlType = "UInt64"
					cast = "toUInt64OrNull"
				}
				if !strings.Contains(plan.Count, "ifNull((") || !strings.Contains(plan.Count, cast) || !strings.Contains(plan.Count, ":"+sqlType+"}") || strings.Contains(plan.Count, q.Where.Value) || strings.Contains(plan.Count, q.Where.Numeric.Field.Key) {
					t.Fatal("unsafe or untyped numeric lowering", plan)
				}
				q.Limit = 100
				q.Offset = 1000
				other, e := Compile(q)
				if e != nil || plan.Count != other.Count || plan.Fields != other.Fields || !reflect.DeepEqual(plan.Parameters, other.Parameters) {
					t.Fatal("presentation paging changed population", e)
				}
			}
		}
	}
}

func TestNumericPredicatePackageAndPopulationComposition(t *testing.T) {
	q := basic()
	q.Where = numericPredicateFixture("int64", "gte", "-2")
	pkg, e := DraftTelemetryPackage(q, "typed-population", "Typed population", 1)
	if e != nil {
		t.Fatal(e)
	}
	plan, e := PlanTelemetryPackage(pkg, q.FromMS, q.ToMS, nil)
	if e != nil {
		t.Fatal(e)
	}
	parsed, e := ReadTelemetryPackage(plan.YAML)
	if e != nil || !reflect.DeepEqual(parsed.Where, q.Where) {
		t.Fatal("typed predicate lost in package", e)
	}
	found := false
	for _, f := range plan.Handoff.Fields {
		found = found || f == q.Where.Numeric.Field
	}
	if !found {
		t.Fatal("numeric field absent from instrumentation handoff")
	}
	population, e := PlanPopulation(PopulationRequest{Question: q})
	if e != nil || len(population.Queries) != 2 {
		t.Fatal(population, e)
	}
	if !reflect.DeepEqual(population.Queries[0].Where, q.Where) || !reflect.DeepEqual(population.Queries[1].Where, q.Where) {
		t.Fatal("population dropped typed parent restriction")
	}
	group := Field{Kind: "attribute", Key: "http.route"}
	choice := Predicate{Field: &group, Op: "eq", Value: "/selected"}
	// Generated categorical selected/rest intent composes with the unchanged typed
	// parent; no generic NOT is introduced over nullable numeric comparisons.
	selected := Predicate{All: []Predicate{*q.Where, choice}}
	rest := Predicate{All: []Predicate{*q.Where, {Any: []Predicate{{Field: &group, Op: "missing"}, {Field: &group, Op: "ne", Value: choice.Value}}}}}
	for _, numeric := range []string{"-3", "-2", "0", "bad", "", "01"} {
		for _, route := range []*string{nil, func() *string { s := "/selected"; return &s }(), func() *string { s := "/other"; return &s }()} {
			row := map[Field]string{q.Where.Numeric.Field: numeric}
			if route != nil {
				row[group] = *route
			}
			a, b, parent := Match(&selected, row), Match(&rest, row), Match(q.Where, row)
			if a && b || (a || b) != parent {
				t.Fatal("selected/rest lost parent partition", row, a, b, parent)
			}
		}
	}
	for _, p := range []*Predicate{&selected, &rest} {
		v := q
		v.Where = p
		if _, e := Compile(v); e != nil {
			t.Fatal(e)
		}
	}
}

// The inline fixture uses real ClickHouse parsing/lowering with the normal
// read-only HTTP adapter. It creates no tables and mutates no telemetry.
func TestClickHouseNumericPredicateParity(t *testing.T) {
	endpoint := os.Getenv("GOO_CLICKHOUSE_TEST_URL")
	if endpoint == "" {
		t.Skip("GOO_CLICKHOUSE_TEST_URL is not configured")
	}
	source := Source{Endpoint: endpoint, User: os.Getenv("GOO_CLICKHOUSE_TEST_USER"), Password: os.Getenv("GOO_CLICKHOUSE_TEST_PASSWORD")}
	raw := []string{"", "0", "-0", "01", "+1", " 1", "1 ", "1.0", "1e3", "-1", "1", "9007199254740993", "9223372036854775807", "-9223372036854775808", "18446744073709551615", "18446744073709551616", "-9223372036854775809", "\x00", "1\n"}
	for _, signal := range []string{"spans", "logs"} {
		for _, kind := range []string{"attribute", "resource"} {
			for _, typ := range []string{"int64", "uint64"} {
				for _, op := range []string{"valid", "missing", "empty", "invalid", "eq", "ne", "gt", "gte", "lt", "lte"} {
					thresholds := []string{"0", "9007199254740993", "-9223372036854775808", "9223372036854775807"}
					if typ == "uint64" {
						thresholds = []string{"0", "9007199254740993", "18446744073709551615"}
					}
					if op == "valid" || op == "missing" || op == "empty" || op == "invalid" {
						thresholds = []string{""}
					}
					for _, threshold := range thresholds {
						p := numericPredicateFixture(typ, op, threshold)
						p.Numeric.Field.Kind = kind
						c := compiler{signal: signal, params: map[string]string{}}
						expr := c.predicate(p)
						data, _ := json.Marshal(raw)
						array := c.param(string(data), "String")
						column := "SpanAttributes"
						if signal == "logs" {
							column = "LogAttributes"
						}
						if kind == "resource" {
							column = "ResourceAttributes"
						}
						sql := "SELECT toString(number) AS n, " + expr + " AS matched FROM (SELECT number, if(number = 0, CAST(map(), 'Map(String, String)'), map('measure', JSONExtract(" + array + ", 'Array(String)')[number])) AS " + column + " FROM numbers(" + fmt.Sprint(len(raw)+1) + ")) ORDER BY number FORMAT JSONEachRow"
						rows, e := source.read(context.Background(), sql, c.params)
						if e != nil || len(rows) != len(raw)+1 {
							t.Fatal(signal, kind, typ, op, len(rows), e)
						}
						for i, row := range rows {
							var observed struct {
								N       string
								Matched *uint8
							}
							if e := json.Unmarshal(row, &observed); e != nil {
								t.Fatal(string(row), e)
							}
							values := map[Field]string{}
							if i > 0 {
								values[p.Numeric.Field] = raw[i-1]
							}
							want := Match(p, values)
							if observed.N != fmt.Sprint(i) || observed.Matched == nil || (*observed.Matched == 1) != want || *observed.Matched > 1 {
								t.Fatal("ClickHouse/reference mismatch", signal, kind, typ, op, i, string(row), want)
							}
						}
					}
				}
			}
		}
	}
}
