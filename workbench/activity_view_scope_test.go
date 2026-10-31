package workbench

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
)

func TestActivityViewExactEnvironmentScopeRoundTripAndCatalog(t *testing.T) {
	store := Store{Dir: t.TempDir()}
	selections := []identity.EnvironmentSelector{
		{Version: identity.EnvironmentSelectorVersion, Policy: "current-only", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "empty"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "unreported"}},
	}
	legacy, err := store.CreateActivityView(activityFixture())
	if err != nil {
		t.Fatal(err)
	}
	for _, wire := range selections {
		v := activityFixture()
		v.Selector.Environment = ""
		v.Selector.EnvironmentSelector = &wire
		v.Selector.Where = &exploration.Predicate{Any: []exploration.Predicate{{Field: &exploration.Field{Kind: "resource", Key: "team"}, Op: "missing"}, {Field: &exploration.Field{Kind: "attribute", Key: "status"}, Op: "eq", Value: ""}}}
		before, _ := json.Marshal(v)
		saved, err := store.CreateActivityView(v)
		if err != nil {
			t.Fatal(err)
		}
		again, err := store.LoadActivityView(saved.ID)
		if err != nil {
			t.Fatal(err)
		}
		again.ID = ""
		after, _ := json.Marshal(again)
		if string(before) != string(after) {
			t.Fatal("saved selector/predicate representation changed", string(before), string(after))
		}
		selection, _ := wire.Selection()
		page, err := store.ActivityViewsForEnvironment("", &selection)
		if err != nil {
			t.Fatal(err)
		}
		want := 1
		if wire.Policy == "current-only" {
			want = 2
		}
		if len(page.Items) != want {
			t.Fatal(wire, len(page.Items), want)
		}
		for _, candidate := range page.Items {
			got, e := candidate.Selector.Query(60000).EnvironmentSelection()
			if e != nil || !selection.Equal(got) {
				t.Fatal("catalog mixed policies/scopes", candidate, e)
			}
		}
		saved.Selector.EnvironmentSelector.Scope.Mode = "empty"
		stable, err := store.LoadActivityView(saved.ID)
		if err != nil || !reflect.DeepEqual(stable.Selector, v.Selector) {
			t.Fatal("store retained mutable selector", err)
		}
	}
	page, err := store.ActivityViews("", "prod")
	if err != nil || len(page.Items) != 2 {
		t.Fatal("legacy filter widened to effective scope", page, err)
	}
	for _, v := range page.Items {
		if !v.Selector.Query(60000).SameEnvironment(legacy.Selector.Query(60000)) {
			t.Fatal("current-only changed")
		}
	}
	all, err := store.ActivityViews("", "")
	if err != nil || len(all.Items) != 5 {
		t.Fatal(all, err)
	}
	invalid := identity.EnvironmentSelection{}
	if _, err = store.ActivityViewsForEnvironment("", &invalid); err == nil {
		t.Fatal("zero checked selection accepted")
	}
}
