package exploration

import (
	"encoding/json"
	"reflect"
	"testing"
)

func statusCohort419(status SpanStatusSummary) Cohort {
	c := cohort(status.Count)
	c.Question.SpanStatus = true
	c.SpanStatus = &status
	if status.Count != "0" {
		groupStatus := status
		c.Groups = []ValueCount{{Value: str("operation"), Count: status.Count, SpanStatus: &groupStatus}}
	} else {
		c.State = "no_data"
	}
	return c
}

func TestRecordedStatusComparison419PreservesCountsSupportAndExactChanges(t *testing.T) {
	a := statusCohort419(SpanStatusSummary{"7", "2", "1", "3", "1"})
	b := statusCohort419(SpanStatusSummary{"11", "5", "2", "4", "0"})
	before, _ := json.Marshal(ComparisonRequest{Baseline: a, Selection: b})
	result, err := CompareCohorts(ComparisonRequest{Baseline: a, Selection: b})
	if err != nil || result.ViewComparison == nil {
		t.Fatal(result, err)
	}
	v := result.ViewComparison
	for key, want := range map[string]string{"count": "4", "status.error": "3", "status.ok": "1", "status.unset": "1", "status.unknown": "-1", "status.error_fraction": "13/77"} {
		got := v.Changes[key]
		if got.State != "derived" || got.Value == nil || *got.Value != want {
			t.Fatalf("%s: %#v want %s", key, got, want)
		}
	}
	delta := v.Changes["status.error_fraction"]
	if delta.Unit != "1" || !reflect.DeepEqual(delta.Aggregation, &ViewAggregation{Family: "ratio", Kind: "difference", Method: "exact-rational", Numerator: "status.error", Denominator: "count"}) {
		t.Fatal("delta mislabeled bounded fraction", delta)
	}
	for _, side := range []*ViewDataset{v.Baseline, v.Selection} {
		if side.Completeness != nil || side.Quality.Extent != "unknown" || side.Origin != "browser-held" {
			t.Fatal("held status acquired authority", side)
		}
	}
	if v.Baseline.Rows[0].Values["status.unset"] == nil || *v.Baseline.Rows[0].Values["status.unset"] != "3" || result.GroupComparison != nil {
		t.Fatal("status support lost or status group display invented")
	}
	after, _ := json.Marshal(ComparisonRequest{Baseline: a, Selection: b})
	if string(before) != string(after) {
		t.Fatal("comparison mutated cohorts")
	}
}

func TestRecordedStatusComparison419ZeroDiagnosticAndInvalidConservation(t *testing.T) {
	zero := statusCohort419(SpanStatusSummary{"0", "0", "0", "0", "0"})
	b := statusCohort419(SpanStatusSummary{"3", "1", "0", "2", "0"})
	result, err := CompareCohorts(ComparisonRequest{Baseline: zero, Selection: b})
	if err != nil || result.ViewComparison == nil {
		t.Fatal(result, err)
	}
	v := result.ViewComparison
	if v.State != "no_data" || *v.Baseline.Rows[0].Values["status.error"] != "0" || v.Baseline.Rows[0].Values["status.error_fraction"] != nil || v.Baseline.Rows[0].Cells["status.error_fraction"].State != "no_data" || v.Changes["status.error_fraction"].State != "unavailable" || v.Changes["status.error_fraction"].Value != nil || *v.Changes["status.error"].Value != "1" {
		t.Fatal("empty fraction became zero or valid count lost", v)
	}
	b.State = "changed_between_reads"
	result, err = CompareCohorts(ComparisonRequest{Baseline: zero, Selection: b})
	if err != nil || result.ViewComparison.State != "diagnostic" {
		t.Fatal(result, err)
	}
	for key, delta := range result.ViewComparison.Changes {
		if delta.State != "unavailable" || delta.Value != nil {
			t.Fatal("diagnostic delta invented", key, delta)
		}
	}
	for name, mutate := range map[string]func(*Cohort){
		"missing status":  func(c *Cohort) { c.SpanStatus = nil },
		"invalid total":   func(c *Cohort) { c.SpanStatus.Count = "4" },
		"partition":       func(c *Cohort) { c.SpanStatus.Unset = "1" },
		"group partition": func(c *Cohort) { c.Groups[0].SpanStatus.Unset = "1" },
	} {
		t.Run(name, func(t *testing.T) {
			c := statusCohort419(SpanStatusSummary{"3", "1", "0", "2", "0"})
			mutate(&c)
			if _, err := CompareCohorts(ComparisonRequest{Baseline: zero, Selection: c}); err == nil {
				t.Fatal("invalid source conservation accepted")
			}
		})
	}
}

func TestRecordedStatusFreshProfile419DoesNotWidenCapturedReuse(t *testing.T) {
	a := statusCohort419(SpanStatusSummary{"3", "1", "0", "2", "0"}).Question
	b := detached(a)
	b.Where = &Predicate{Field: &Field{Kind: "resource", Key: "service.version"}, Op: "eq", Value: "v2"}
	if err := CompatibleCountCohortQueries(a, b); err != nil {
		t.Fatal(err)
	}
	if ValidateCapturedCountQuery(a) == nil || CompatibleCapturedCountQueries(a, a) == nil {
		t.Fatal("fresh status widened captured profiles")
	}
	b.SpanStatus = false
	if CompatibleCountCohortQueries(a, b) == nil {
		t.Fatal("status directive drift accepted")
	}
}
