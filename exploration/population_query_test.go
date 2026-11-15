package exploration

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

type conservation454Case struct {
	Name             string
	Query            json.RawMessage
	Population       Query
	Limit, Offset    int
	SpecializedCount string
}

func conservation454Bank(t *testing.T) ([]conservation454Case, []conservation454Case) {
	t.Helper()
	raw, err := os.ReadFile("../tests/fixtures/selector-conservation454.json")
	if err != nil {
		t.Fatal(err)
	}
	var bank struct {
		Version string
		Cases   []conservation454Case
		Invalid []conservation454Case
	}
	if err := json.Unmarshal(raw, &bank); err != nil || bank.Version != "goo/selector-conservation/v1" || len(bank.Cases) != 9 || len(bank.Invalid) != 7 {
		t.Fatalf("invalid shared case bank: %v", err)
	}
	return bank.Cases, bank.Invalid
}

func TestPopulationQuery454ConservesCompleteSelectors(t *testing.T) {
	cases, invalid := conservation454Bank(t)
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			var parent Query
			if err := json.Unmarshal(tc.Query, &parent); err != nil {
				t.Fatal(err)
			}
			before, _ := json.Marshal(parent)
			out, err := PopulationQuery(parent, tc.Limit, tc.Offset)
			if err != nil || !reflect.DeepEqual(out, tc.Population) {
				t.Fatalf("population changed: %v\ngot %#v\nwant %#v", err, out, tc.Population)
			}
			// Both compiler source clauses must be identical: grouping and measures
			// cannot erase a population restriction or alter its parameter binding.
			original, err := Compile(parent)
			if err != nil {
				t.Fatal(err)
			}
			projected, err := Compile(out)
			if err != nil {
				t.Fatal(err)
			}
			source := func(sql string) string {
				i := strings.Index(sql, " FROM otel_")
				if i < 0 {
					t.Fatal("missing source clause", sql)
				}
				return sql[i:]
			}
			if source(original.Count) != source(projected.Count) {
				t.Fatal("projection changed the compiled source population")
			}
			for key, value := range projected.Parameters {
				if strings.Contains(projected.Count, "{"+key+":") && original.Parameters[key] != value {
					t.Fatalf("population parameter %s changed", key)
				}
			}
			grouped := cloneQuery(out)
			grouped.GroupBy = []Field{{Kind: "column", Key: "service"}}
			grouped.Offset = 0
			if refused := ValidateCapturedCountQuery(grouped) != nil; refused != (tc.SpecializedCount == "refuse") {
				t.Fatal("specialized profile did not preserve its refusal policy")
			}
			// Mutating any copied restriction must never mutate the parent bank.
			if out.Where != nil {
				out.Where.Value = "mutated"
				if len(out.Where.All) > 0 {
					out.Where.All[0].Field.Key = "scope"
					if last := &out.Where.All[len(out.Where.All)-1]; last.Numeric != nil {
						last.Numeric.Field.Key = "mutated"
					}
				}
			}
			if out.ReportedEdge != nil {
				out.ReportedEdge.Parent.Namespace.Value = "mutated"
			}
			if out.LinksTo != nil {
				out.LinksTo.TraceID = strings.Repeat("f", 32)
			}
			if out.EnvironmentSelector != nil {
				out.EnvironmentSelector.Scope.Value = "mutated"
			}
			after, _ := json.Marshal(parent)
			if string(before) != string(after) {
				t.Fatal("projection aliases its parent")
			}
		})
	}
	for _, tc := range invalid {
		t.Run(tc.Name, func(t *testing.T) {
			var q Query
			if json.Unmarshal(tc.Query, &q) != nil {
				return // Literal wire admission refuses before a projection exists.
			}
			before, _ := json.Marshal(q)
			out, err := PopulationQuery(q, tc.Limit, tc.Offset)
			if err == nil || !reflect.DeepEqual(out, Query{}) {
				t.Fatal("invalid projection returned a partial question")
			}
			after, _ := json.Marshal(q)
			if string(before) != string(after) {
				t.Fatal("refusal mutated the parent")
			}
		})
	}
}

func TestPopulationQuery454ExplanationAdoption(t *testing.T) {
	cases, _ := conservation454Bank(t)
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			var q Query
			if err := json.Unmarshal(tc.Query, &q); err != nil {
				t.Fatal(err)
			}
			r := ExplanationRequest{Version: ExplanationVersion, Baseline: q, Selection: q, Fields: []Field{{Kind: "column", Key: "service"}}}
			plan, err := PlanExplanation(r)
			if err != nil {
				t.Fatal(err)
			}
			want, err := PopulationQuery(q, 100, 0)
			if err != nil || !reflect.DeepEqual(plan.Baseline, want) || !reflect.DeepEqual(plan.Selection, want) {
				t.Fatal("explanation did not use complete checked population", err)
			}
			if plan.Baseline.LinksTo != nil {
				plan.Baseline.LinksTo.SpanID = strings.Repeat("f", 16)
				if plan.Request.Baseline.LinksTo.SpanID != q.LinksTo.SpanID || plan.Selection.LinksTo.SpanID != q.LinksTo.SpanID {
					t.Fatal("explanation populations share mutable selectors")
				}
			}
		})
	}
}
