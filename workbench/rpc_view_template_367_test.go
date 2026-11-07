package workbench

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
)

func rpcCurrentTemplateInput367(t *testing.T) ViewTemplateExpansionRequest {
	t.Helper()
	selection, err := identity.EffectiveEnvironment(identity.EnvironmentScope{Mode: "named", Value: "prod"})
	if err != nil {
		t.Fatal(err)
	}
	selector, err := selection.Selector()
	if err != nil {
		t.Fatal(err)
	}
	plan, err := exploration.PlanServiceDashboard(exploration.DashboardRequest{Profile: "rpc-server-current", RPCServerCurrentObserved: true, Scope: exploration.InventoryQuestion{EnvironmentSelector: &selector, Namespace: &identity.Namespace{Mode: "missing"}, FromMS: 1000, ToMS: 61000}, Service: "rpc-api", Signals: []string{"spans"}, Fields: []exploration.Field{{Kind: "attribute", Key: "rpc.system.name"}}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range plan.Widgets {
		if w.Key == "dimension" {
			return ViewTemplateExpansionRequest{Template: "otel.rpc-server-current-methods", Version: 1, Widget: w}
		}
	}
	t.Fatal("missing current RPC dimension without method metadata")
	return ViewTemplateExpansionRequest{}
}

func TestCurrentRPCViewTemplate367WholeScopeAndOptionalMethod(t *testing.T) {
	in := rpcCurrentTemplateInput367(t)
	out, err := ExpandViewTemplate(in)
	if err != nil {
		t.Fatal(err)
	}
	if out.View.Template == nil || out.View.Template.Profile != "otel.rpc-server-current" || !reflect.DeepEqual(out.View.Queries[0], *in.Widget.Query) {
		t.Fatal("profile or whole source question changed")
	}
	if out.View.Composition.Children[0].Facet.Field != (exploration.Field{Kind: "attribute", Key: "rpc.method"}) {
		t.Fatal("wrong method facet")
	}
	for _, p := range conjuncts(out.View.Queries[0].Where) {
		if p.Field != nil && p.Field.Key == "rpc.method" {
			t.Fatal("optional method became an admission predicate")
		}
	}
	destination := cloneAuthored(t, out.View).Queries[0]
	destination.FromMS = 70000
	destination.ToMS = 130000
	destination.GroupBy = []exploration.Field{{Kind: "attribute", Key: "rpc.service"}}
	destination.Where.All = append(destination.Where.All, exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "tenant"}, Op: "eq", Value: "other"})
	destination.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "empty"}
	original, _ := json.Marshal(destination)
	expanded, err := ExpandViewTemplate(ViewTemplateExpansionRequest{Source: &out.View, Query: &destination})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expanded.View.Queries[0], destination) {
		t.Fatal("destination bounds/policy/namespace/predicates/grouping changed")
	}
	expanded.View.Queries[0].EnvironmentSelector.Scope.Mode = "unreported"
	expanded.View.Queries[0].Where.All[0].Value = "changed"
	after, _ := json.Marshal(destination)
	if string(original) != string(after) {
		t.Fatal("destination aliases expansion")
	}
	store := Store{Dir: t.TempDir()}
	saved, err := store.CreateExplorationView(out.View)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := store.LoadExplorationView(saved.ID)
	if err != nil || loaded.Template.Profile != "otel.rpc-server-current" {
		t.Fatal("RPC template not reusable after save", err)
	}
}

func TestCurrentRPCViewTemplate367RefusesWeakenedRole(t *testing.T) {
	source, err := ExpandViewTemplate(rpcCurrentTemplateInput367(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"kind", "rpc.system.name", "service"} {
		t.Run(key, func(t *testing.T) {
			q := cloneAuthored(t, source.View).Queries[0]
			predicates := conjuncts(q.Where)
			var kept []exploration.Predicate
			for _, p := range predicates {
				if p.Field == nil || p.Field.Key != key {
					kept = append(kept, p)
				}
			}
			q.Where = &exploration.Predicate{All: kept}
			if _, err := ExpandViewTemplate(ViewTemplateExpansionRequest{Source: &source.View, Query: &q}); err == nil {
				t.Fatal("weakened current RPC profile accepted")
			}
		})
	}
	in := rpcCurrentTemplateInput367(t)
	in.Widget.Family = "rpc-server"
	if _, err := ExpandViewTemplate(in); err == nil {
		t.Fatal("legacy widget treated as current")
	}
}
