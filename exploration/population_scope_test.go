package exploration

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/casselc/goo/identity"
)

func populationSelectors() []identity.EnvironmentSelector {
	return []identity.EnvironmentSelector{
		{Version: identity.EnvironmentSelectorVersion, Policy: "current-only", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "empty"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "unreported"}},
	}
}

func TestPopulationPlanPreservesTypedScopeAndPortablePredicates(t *testing.T) {
	for _, selector := range populationSelectors() {
		t.Run(selector.Policy+"/"+selector.Scope.Mode, func(t *testing.T) {
			q := basic()
			q.Environment = ""
			q.EnvironmentSelector = &selector
			q.Offset = 7
			q.Where = &Predicate{Any: []Predicate{{Field: &Field{Kind: "attribute", Key: "region"}, Op: "missing"}, {Field: &Field{Kind: "attribute", Key: "region"}, Op: "eq", Value: ""}}}
			before, _ := json.Marshal(q)
			plan, err := PlanPopulation(PopulationRequest{Question: q})
			if err != nil || len(plan.Queries) != 2 || plan.Version != "goo/population/v1" {
				t.Fatal(plan, err)
			}
			for _, planned := range plan.Queries {
				if !planned.SameEnvironment(q) || !reflect.DeepEqual(planned.EnvironmentSelector, q.EnvironmentSelector) || !reflect.DeepEqual(planned.Where, q.Where) || planned.FromMS != q.FromMS || planned.ToMS != q.ToMS || planned.Offset != 0 || planned.BucketMS == 0 {
					t.Fatal("planner changed checked scope or predicates", planned)
				}
				if _, err = Compile(planned); err != nil {
					t.Fatal(err)
				}
			}
			// Both returned questions must be detached from the caller's mutable input.
			plan.Queries[0].EnvironmentSelector.Scope.Mode = "invalid"
			plan.Queries[0].Where.Any[0].Field.Key = "changed"
			after, _ := json.Marshal(q)
			if string(before) != string(after) {
				t.Fatal("planner mutated its input")
			}
		})
	}
}

func TestPopulationPlanChecksCompanionPolicyAndWindow(t *testing.T) {
	selectors := populationSelectors()
	for i, selector := range selectors {
		q := basic()
		q.Environment = ""
		q.EnvironmentSelector = &selector
		for j, otherSelector := range selectors {
			other := q
			other.Signal = "logs"
			other.EnvironmentSelector = &otherSelector
			_, err := PlanPopulation(PopulationRequest{Question: q, Companion: &other})
			if (err == nil) != (i == j) {
				t.Fatalf("companion policy %d/%d accepted=%v", i, j, err == nil)
			}
		}
		other := q
		other.Signal = "logs"
		other.ToMS++
		if _, err := PlanPopulation(PopulationRequest{Question: q, Companion: &other}); err == nil {
			t.Fatal("companion changed fixed window")
		}
		q.Where = &Predicate{Field: &Field{Kind: "column", Key: "status"}, Op: "eq", Value: "Error"}
		plan, err := PlanPopulation(PopulationRequest{Question: q})
		if err != nil || len(plan.Queries) != 1 || len(plan.Notices) < 2 || !reflect.DeepEqual(plan.Queries[0].Where, q.Where) {
			t.Fatal("incompatible companion must be withheld without weakening predicates", plan, err)
		}
	}
	// Historical current-only names and their explicit typed equivalent are semantically equal.
	q := basic()
	q.Environment = "prod"
	other := q
	other.Signal = "logs"
	other.Environment = ""
	other.EnvironmentSelector = &selectors[0]
	if _, err := PlanPopulation(PopulationRequest{Question: q, Companion: &other}); err != nil {
		t.Fatal(err)
	}
}
