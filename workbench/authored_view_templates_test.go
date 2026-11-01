package workbench

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
)

func dbTemplateInput(t *testing.T) ViewTemplateExpansionRequest {
	t.Helper()
	plan, err := exploration.PlanServiceDashboard(exploration.DashboardRequest{Profile: "db-client", DBClientObserved: true, Scope: exploration.InventoryQuestion{Environment: "prod", FromMS: 1000, ToMS: 61000}, Service: "db-api", Signals: []string{"spans"}, Fields: []exploration.Field{{Kind: "attribute", Key: "db.system.name"}, {Kind: "attribute", Key: "db.operation.name"}}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range plan.Widgets {
		if w.Key == "dimension" {
			return ViewTemplateExpansionRequest{Template: "otel.db-client-operations", Version: 1, Widget: w}
		}
	}
	t.Fatal("missing DB dimension")
	return ViewTemplateExpansionRequest{}
}
func authoredDBTemplate(t *testing.T) ExplorationView {
	t.Helper()
	out, err := ExpandViewTemplate(dbTemplateInput(t))
	if err != nil {
		t.Fatal(err)
	}
	if out.View.Template == nil || out.View.Template.Profile != "otel.db-client" {
		t.Fatal("DB preset did not use shared template contract")
	}
	return out.View
}
func cloneAuthored(t *testing.T, v ExplorationView) ExplorationView {
	t.Helper()
	out, err := cloneTemplateView(v)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestAuthoredTemplateBindsDestinationWithoutScopeTransfer(t *testing.T) {
	source := authoredDBTemplate(t)
	source.Composition.Children[1].Records.Display.FieldColumns = []exploration.Field{{Kind: "attribute", Key: "db.namespace"}}
	saved, err := (Store{Dir: t.TempDir()}).CreateExplorationView(source)
	if err != nil {
		t.Fatal(err)
	}
	destination := cloneAuthored(t, source).Queries[0]
	destination.Environment = "staging"
	destination.FromMS = 70000
	destination.ToMS = 130000
	destination.Offset = 100
	destination.GroupBy = []exploration.Field{{Kind: "attribute", Key: "db.namespace"}}
	destination.Where.All[0].Value = "other-db-api"
	beforeSource, _ := json.Marshal(saved)
	beforeTarget, _ := json.Marshal(destination)
	out, err := ExpandViewTemplate(ViewTemplateExpansionRequest{Source: saved, Query: &destination})
	if err != nil {
		t.Fatal(err)
	}
	expected := destination
	expected.Offset = 0
	if !reflect.DeepEqual(out.View.Queries[0], expected) || out.Template.ID != saved.ID || out.View.ID != "" || out.View.Template == nil {
		t.Fatal("scope/reference changed", out)
	}
	out.View.Queries[0].Where.All[0].Value = "mutated"
	out.View.Composition.Children[1].Records.Display.FieldColumns[0].Key = "mutated"
	out.View.Template.Profile = "mutated"
	afterSource, _ := json.Marshal(saved)
	afterTarget, _ := json.Marshal(destination)
	if string(beforeSource) != string(afterSource) || string(beforeTarget) != string(afterTarget) {
		t.Fatal("expansion aliases caller configuration")
	}
}
func TestAuthoredTemplateRefusesAmbiguousOrNonportableInputs(t *testing.T) {
	mutations := map[string]func(*ExplorationView){
		"missing optin":   func(v *ExplorationView) { v.Template = nil },
		"unknown profile": func(v *ExplorationView) { v.Template.Profile = "anything" },
		"binding":         func(v *ExplorationView) { v.Template.Group = "source" },
		"version":         func(v *ExplorationView) { v.Template.Version = "goo.template/2" },
		"local Find":      func(v *ExplorationView) { v.Composition.Children[0].Aggregate.Display.Query = "SELECT" },
		"numeric": func(v *ExplorationView) {
			v.Queries[0].Numeric = &exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "bytes"}, Type: "uint64", Parse: "canonical-decimal", Unit: "By"}
		},
		"comparison": func(v *ExplorationView) {
			v.Composition.Children = append(v.Composition.Children, ViewCompositionChild{Comparison: &CountComparisonViewDeclaration{}})
		},
		"tail":     func(v *ExplorationView) { v.Tail = &ExplorationTail{Signal: "spans"} },
		"wrong ID": func(v *ExplorationView) { v.ID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" },
		"OR weakening": func(v *ExplorationView) {
			v.Queries[0].Where = &exploration.Predicate{Any: []exploration.Predicate{*v.Queries[0].Where, {Field: &exploration.Field{Kind: "column", Key: "kind"}, Op: "eq", Value: "Server"}}}
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			source := authoredDBTemplate(t)
			target := cloneAuthored(t, source).Queries[0]
			mutate(&source)
			if _, err := ExpandViewTemplate(ViewTemplateExpansionRequest{Source: &source, Query: &target}); err == nil {
				t.Fatal("bad source accepted")
			}
		})
	}
	for name, mutate := range map[string]func(*exploration.Query){
		"OR": func(q *exploration.Query) {
			q.Where = &exploration.Predicate{Any: []exploration.Predicate{*q.Where, {Field: &exploration.Field{Kind: "column", Key: "kind"}, Op: "eq", Value: "Server"}}}
		},
		"wrong namespace":  func(q *exploration.Query) { q.Where.All[2].Field.Kind = "resource" },
		"missing duration": func(q *exploration.Query) { q.Summary = "" },
		"invalid offset":   func(q *exploration.Query) { q.Offset = -1 },
	} {
		t.Run(name, func(t *testing.T) {
			source := authoredDBTemplate(t)
			target := cloneAuthored(t, source).Queries[0]
			source.Composition.Children[0].Aggregate.Display.Value = "duration.p95"
			source.Composition.Children[0].Aggregate.Display.Scale = "max"
			mutate(&target)
			if _, err := ExpandViewTemplate(ViewTemplateExpansionRequest{Source: &source, Query: &target}); err == nil {
				t.Fatal("bad target accepted")
			}
		})
	}
}
func TestTemplateExpansionWireArmsAreExclusive(t *testing.T) {
	source := authoredDBTemplate(t)
	q := source.Queries[0]
	valid, _ := json.Marshal(ViewTemplateExpansionRequest{Source: &source, Query: &q})
	var in ViewTemplateExpansionRequest
	if err := json.Unmarshal(valid, &in); err != nil {
		t.Fatal(err, string(valid))
	}
	for _, bad := range []string{string(valid[:len(valid)-1]) + `,"widget":null}`, string(valid[:len(valid)-1]) + `,"template":""}`, `{"source":null,"query":{}}`, `{"template":"otel.http-server-routes","version":1,"widget":null}`, `{"source":{},"query":{},"extra":1}`} {
		if json.Unmarshal([]byte(bad), &in) == nil {
			t.Fatal("mixed/null arm admitted", bad)
		}
	}
	raw, _ := json.Marshal(source)
	bad := string(raw[:len(raw)-1]) + `,"extra":true}`
	var v ExplorationView
	if json.Unmarshal([]byte(bad), &v) == nil {
		t.Fatal("unknown view field accepted")
	}
}
