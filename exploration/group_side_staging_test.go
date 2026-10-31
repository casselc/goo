package exploration

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/identity"
)

func groupSideIntentFor(t *testing.T, request ComparisonRequest, side, key string) GroupSideIntent {
	t.Helper()
	result, err := CompareCohorts(request)
	if err != nil || result.GroupComparison == nil {
		t.Fatal("group fixture is not comparable", err)
	}
	snapshot := result.GroupComparison.Baseline
	if side == "selection" {
		snapshot = result.GroupComparison.Selection
	}
	return GroupSideIntent{Type: "scope-group-side", Side: side, Key: key, DatasetID: snapshot.ID, Revision: snapshot.Revision, ScopeID: snapshot.Scope.ID, Generation: snapshot.Scope.Generation}
}

func TestStageGroupSidePreservesChosenRootAndTypedMissingness(t *testing.T) {
	empty := ""
	for _, env := range []string{"legacy", "current", "effective", "empty", "unreported"} {
		t.Run(env, func(t *testing.T) {
			a := groupCountCohort(ValueCount{Count: "1"}, ValueCount{Value: &empty, Count: "1"})
			a.Question.Where = &Predicate{Any: []Predicate{{Field: &Field{Kind: "attribute", Key: "region"}, Op: "missing"}, {Field: &Field{Kind: "attribute", Key: "region"}, Op: "eq", Value: "west"}}}
			a.Question.BucketMS = 10000
			if env != "legacy" {
				mode, value, policy := "named", "prod", "effective"
				if env == "current" {
					policy = "current-only"
				}
				if env == "empty" || env == "unreported" {
					mode, value = env, ""
				}
				a.Question.Environment = ""
				a.Question.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: policy, Scope: identity.EnvironmentScope{Mode: mode, Value: value}}
			}
			b := detached(a)
			a.Question.Limit, a.Question.Offset = 7, 14
			b.Question.FromMS, b.Question.ToMS, b.Question.Limit, b.Question.Offset = 61000, 181000, 13, 19
			request := ComparisonRequest{Baseline: a, Selection: b}
			before := detached(request)
			for _, side := range []string{"baseline", "selection"} {
				for _, key := range []string{"0", "1"} {
					intent := groupSideIntentFor(t, request, side, key)
					got, err := StageGroupSide(request, intent)
					if err != nil {
						t.Fatal(err)
					}
					want := detached(a.Question)
					if side == "selection" {
						want = detached(b.Question)
					}
					want.Offset = 0
					leaf := Predicate{Field: &Field{Kind: "attribute", Key: "cohort"}, Op: "missing"}
					if key == "1" {
						leaf.Op = "eq"
					}
					want.Where = &Predicate{All: []Predicate{*want.Where, leaf}}
					if !reflect.DeepEqual(got, want) {
						t.Fatalf("chosen question changed: got%+v want%+v", got, want)
					}
					if _, err := Compile(got); err != nil {
						t.Fatal(err)
					}
					got.Where.All[0].Any[0].Field.Key = "changed"
					got.GroupBy[0].Key = "changed"
					if got.EnvironmentSelector != nil {
						got.EnvironmentSelector.Scope.Mode = "changed"
					}
					if !reflect.DeepEqual(request, before) {
						t.Fatal("staged query aliases original operands")
					}
				}
			}
		})
	}
}

func TestStageGroupSideRetainsMeasureAndLimitConfiguration(t *testing.T) {
	value := "api"
	a := groupCountCohort(ValueCount{Value: &value, Count: "2"})
	a.Question.Where = nil
	a.Question.Summary = "span-duration"
	a.Question.BucketMS = 10000
	a.Duration = &DurationSummary{State: "observed", Count: "2", SumNS: "0", MinNS: "0", MaxNS: "0", P50NS: "0", P95NS: "0", P99NS: "0"}
	a.Groups[0].Duration = detached(a.Duration)
	request := ComparisonRequest{Baseline: a, Selection: detached(a)}
	got, err := StageGroupSide(request, groupSideIntentFor(t, request, "baseline", "1api"))
	if err != nil {
		t.Fatal(err)
	}
	if got.Summary != a.Question.Summary || got.BucketMS != a.Question.BucketMS || got.Limit != a.Question.Limit || !reflect.DeepEqual(got.GroupBy, a.Question.GroupBy) || got.Where.Op != "eq" || got.Where.Value != "api" || got.Where.All != nil {
		t.Fatal("measure/limit scope changed", got)
	}
}

func TestStageGroupSideRejectsWrongOrUnavailableIntent(t *testing.T) {
	value := "api"
	a := groupCountCohort(ValueCount{Value: &value, Count: "2"})
	b := groupCountCohort()
	request := ComparisonRequest{Baseline: a, Selection: b}
	intent := groupSideIntentFor(t, request, "baseline", "1api")
	if _, err := StageGroupSide(request, intent); err != nil {
		t.Fatal("known-empty opposite side blocked positive side", err)
	}
	for name, edit := range map[string]func(*GroupSideIntent){
		"type":         func(i *GroupSideIntent) { i.Type = "scope-group" },
		"side":         func(i *GroupSideIntent) { i.Side = "current" },
		"swapped side": func(i *GroupSideIntent) { i.Side = "selection" },
		"key":          func(i *GroupSideIntent) { i.Key = "1unknown" },
		"dataset":      func(i *GroupSideIntent) { i.DatasetID += "x" },
		"revision":     func(i *GroupSideIntent) { i.Revision += "x" },
		"scope":        func(i *GroupSideIntent) { i.ScopeID += "x" },
		"generation":   func(i *GroupSideIntent) { i.Generation += "x" },
	} {
		t.Run(name, func(t *testing.T) {
			wrong := intent
			edit(&wrong)
			if _, err := StageGroupSide(request, wrong); err == nil {
				t.Fatal("invalid intent staged")
			}
		})
	}
	absent := groupSideIntentFor(t, request, "selection", "1api")
	if _, err := StageGroupSide(request, absent); err == nil {
		t.Fatal("aligned absent zero staged")
	}
	request.Selection.State = "changed_between_reads"
	if _, err := StageGroupSide(request, groupSideIntentFor(t, request, "baseline", "1api")); err == nil {
		t.Fatal("diagnostic pair staged even coherent positive chosen side")
	}
	if _, err := StageGroupSide(ComparisonRequest{}, intent); err == nil {
		t.Fatal("invalid original operands accepted")
	}
}

func TestStageGroupSideRejectsFinalPredicateBudgetAndUnknownJSON(t *testing.T) {
	value := "api"
	a := groupCountCohort(ValueCount{Value: &value, Count: "1"})
	where := Predicate{Field: &Field{Kind: "attribute", Key: "region"}, Op: "eq", Value: "west"}
	for range 4 {
		where = Predicate{All: []Predicate{where}}
	}
	a.Question.Where = &where
	if err := a.Question.Validate(); err != nil {
		t.Fatal("parent must be valid", err)
	}
	request := ComparisonRequest{Baseline: a, Selection: detached(a)}
	if _, err := StageGroupSide(request, groupSideIntentFor(t, request, "baseline", "1api")); err == nil || !strings.Contains(err.Error(), "budget") {
		t.Fatal("invalid child was staged or parent weakened", err)
	}
	valid, _ := json.Marshal(GroupSideIntent{Type: "scope-group-side", Side: "baseline", Key: "1api"})
	var intent GroupSideIntent
	unknown := append(append([]byte(nil), valid[:len(valid)-1]...), []byte(`,"query":{}}`)...)
	if err := json.Unmarshal(unknown, &intent); err == nil {
		t.Fatal("unknown query field accepted")
	}
	if err := intent.UnmarshalJSON(append(valid, valid...)); err == nil {
		t.Fatal("multiple JSON objects accepted")
	}
}
