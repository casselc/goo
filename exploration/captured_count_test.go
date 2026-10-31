package exploration

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/casselc/goo/identity"
)

func capturedCountQuestion() Query {
	return Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 2, GroupBy: []Field{{Kind: "column", Key: "service"}}, Where: &Predicate{Field: &Field{Kind: "attribute", Key: "team"}, Op: "eq", Value: "core"}}
}
func TestCapturedCountCompatibilityRetainsDefinitionAndTypedEnvironment(t *testing.T) {
	a := capturedCountQuestion()
	b := detached(a)
	b.FromMS = 61000
	b.ToMS = 181000
	b.Offset = 100
	b.Limit = 1
	b.Environment = ""
	b.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "current-only", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}}
	beforeA, beforeB := detached(a), detached(b)
	if err := CompatibleCapturedCountQueries(a, b); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, beforeA) || !reflect.DeepEqual(b, beforeB) {
		t.Fatal("compatibility mutated an operand")
	}
	changes := map[string]func(*Query){
		"predicate": func(q *Query) { q.Where.Value = "other" },
		"group":     func(q *Query) { q.GroupBy[0] = Field{Kind: "attribute", Key: "team"} },
		"summary":   func(q *Query) { q.Summary = "span-duration" },
		"bucket":    func(q *Query) { q.BucketMS = 10000 },
		"policy":    func(q *Query) { q.EnvironmentSelector.Policy = "effective" },
		"scope":     func(q *Query) { q.EnvironmentSelector.Scope.Value = "other" },
		"signal":    func(q *Query) { q.Signal = "logs" },
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			wrong := detached(b)
			change(&wrong)
			if CompatibleCapturedCountQueries(a, wrong) == nil {
				t.Fatal("different root definition accepted")
			}
		})
	}
}
func TestCapturedCountProfileRefusesUnsupportedAndInvalidOriginal(t *testing.T) {
	q := capturedCountQuestion()
	changes := map[string]func(*Query){
		"no grouping":       func(q *Query) { q.GroupBy = nil },
		"multiple grouping": func(q *Query) { q.GroupBy = append(q.GroupBy, Field{Kind: "attribute", Key: "team"}) },
		"pattern":           func(q *Query) { q.Signal = "logs"; q.GroupBy = []Field{{Kind: "column", Key: "body_pattern_v1"}} },
		"duration bucket":   func(q *Query) { q.GroupBy = []Field{{Kind: "column", Key: "duration_bucket"}} },
		"status":            func(q *Query) { q.SpanStatus = true },
		"split":             func(q *Query) { q.BucketMS = 10000; q.GroupTimeSeries = true },
		"numeric":           func(q *Query) { v := decimalBinding(2); q.Numeric = &v },
		"invalid summary":   func(q *Query) { q.Summary = "bad" },
		"invalid page":      func(q *Query) { q.Offset = 1001 },
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			wrong := detached(q)
			change(&wrong)
			if ValidateCapturedCountQuery(wrong) == nil || CompatibleCapturedCountQueries(q, wrong) == nil {
				t.Fatal("unsupported original accepted")
			}
		})
	}
	q.Summary = "span-duration"
	q.BucketMS = 10000
	if err := ValidateCapturedCountQuery(q); err != nil {
		t.Fatal("eligible duration parent refused", err)
	}
}
func TestCapturedCountUsesCanonicalKnownZeroAndDisagreement(t *testing.T) {
	q := capturedCountQuestion()
	value := "api"
	cohort := func(count, state string) Cohort {
		c := Cohort{Question: detached(q), Basis: "held-content", CapturedAt: "2026-10-31T00:00:00Z", Count: count, State: state, Groups: []ValueCount{}}
		if count != "0" {
			c.Groups = []ValueCount{{Value: &value, Count: count}}
		}
		return c
	}
	for _, tc := range []struct{ a, b, want string }{{"0", "0", "0"}, {"0", "3", "3"}, {"3", "0", "-3"}, {"2", "5", "3"}} {
		aState, bState := "observed", "observed"
		if tc.a == "0" {
			aState = "no_data"
		}
		if tc.b == "0" {
			bState = "no_data"
		}
		result, err := CompareCohorts(ComparisonRequest{Baseline: cohort(tc.a, aState), Selection: cohort(tc.b, bState)})
		if err != nil || result.ViewComparison == nil {
			t.Fatal("canonical count comparison missing", err)
		}
		change := result.ViewComparison.Changes["count"]
		if change.State != "derived" || change.Value == nil || *change.Value != tc.want || change.Unit != "records" {
			t.Fatalf("wrong canonical count change: %+v", change)
		}
		if result.ViewComparison.Baseline.Completeness != nil || result.ViewComparison.Selection.Completeness != nil {
			t.Fatal("held inputs gained source completeness")
		}
		if _, err := json.Marshal(result); err != nil {
			t.Fatal(err)
		}
	}
	result, err := CompareCohorts(ComparisonRequest{Baseline: cohort("3", "observed"), Selection: cohort("4", "changed_between_reads")})
	if err != nil || result.ViewComparison == nil || result.ViewComparison.State != "diagnostic" || result.ViewComparison.Changes["count"].Value != nil || result.ViewComparison.Changes["count"].State != "unavailable" {
		t.Fatal("source disagreement gained a count delta", err)
	}
}
