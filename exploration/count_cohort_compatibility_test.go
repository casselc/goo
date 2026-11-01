package exploration

import (
	"encoding/json"
	"github.com/casselc/goo/identity"
	"reflect"
	"testing"
)

func TestCountCohortCompatibilityAllowsOnlyIndependentPredicates(t *testing.T) {
	a := capturedCountQuestion()
	b := detached(a)
	b.Where = &Predicate{Any: []Predicate{{Field: &Field{Kind: "attribute", Key: "team"}, Op: "eq", Value: "other"}, {Field: &Field{Kind: "resource", Key: "team"}, Op: "missing"}}}
	b.FromMS = 61000
	b.ToMS = 181000
	b.Limit = 100
	b.Offset = 100
	b.Environment = ""
	b.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "current-only", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}}
	beforeA, beforeB := detached(a), detached(b)
	if err := CompatibleCountCohortQueries(a, b); err != nil {
		t.Fatal(err)
	}
	if CompatibleCapturedCountQueries(a, b) == nil {
		t.Fatal("cohort display broadened captured/template reuse")
	}
	if !reflect.DeepEqual(a, beforeA) || !reflect.DeepEqual(b, beforeB) {
		t.Fatal("compatibility mutated original predicates")
	}
	b.Where = nil
	if err := CompatibleCountCohortQueries(a, b); err != nil {
		t.Fatal("whole population versus subset refused", err)
	}
	changes := map[string]func(*Query){
		"invalid predicate": func(q *Query) {
			q.Where = &Predicate{Field: &Field{Kind: "column", Key: "no-such-field"}, Op: "eq", Value: "x"}
		},
		"group":   func(q *Query) { q.GroupBy[0] = Field{Kind: "resource", Key: "team"} },
		"summary": func(q *Query) { q.Summary = "span-duration" }, "bucket": func(q *Query) { q.BucketMS = 10000 },
		"policy": func(q *Query) { q.EnvironmentSelector.Policy = "effective" }, "scope": func(q *Query) { q.EnvironmentSelector.Scope.Value = "other" },
		"signal": func(q *Query) { q.Signal = "logs" }, "status": func(q *Query) { q.SpanStatus = true }, "page": func(q *Query) { q.Offset = 1001 },
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			wrong := detached(b)
			change(&wrong)
			if CompatibleCountCohortQueries(a, wrong) == nil {
				t.Fatal("cohort compatibility ignored non-predicate difference")
			}
		})
	}
}
func TestCountCohortGroupAlignmentRetainsIndependentScopesAndUncertainty(t *testing.T) {
	label := "api"
	a := groupCountCohort(ValueCount{Value: &label, Count: "2"})
	b := groupCountCohort(ValueCount{Value: &label, Count: "5"})
	b.Question.Where = nil
	result, err := CompareCohorts(ComparisonRequest{Baseline: a, Selection: b})
	if err != nil || result.GroupComparison == nil {
		t.Fatal(err)
	}
	group := result.GroupComparison
	var baselineQuery, selectionQuery Query
	json.Unmarshal(group.Baseline.Scope.Selector, &baselineQuery)
	json.Unmarshal(group.Selection.Scope.Selector, &selectionQuery)
	if *group.Rows[0].Change.Value != "3" || !reflect.DeepEqual(baselineQuery, a.Question) || !reflect.DeepEqual(selectionQuery, b.Question) {
		t.Fatal("independent checked predicates lost")
	}
	if group.Baseline.Completeness != nil || group.Selection.Completeness != nil || group.Baseline.Quality.Extent != "unknown" || group.Selection.Origin != "browser-held" {
		t.Fatal("cohort gained source authority")
	}
	b.State = "changed_between_reads"
	b.Count = "1"
	result, err = CompareCohorts(ComparisonRequest{Baseline: a, Selection: b})
	if err != nil || result.GroupComparison == nil {
		t.Fatal(err)
	}
	if result.GroupComparison.Rows[0].Change.Value != nil || result.GroupComparison.SelectionAccounting != "diagnostic" {
		t.Fatal("different predicates bypassed diagnostic accounting")
	}
}
