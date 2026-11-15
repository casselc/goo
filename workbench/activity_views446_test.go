package workbench

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
)

func pairedActivityFixture() ActivityView {
	return ActivityView{Version: 1, Title: "Service logs and spans", Selector: ActivitySelector{Signal: "both", Environment: "prod", Where: &exploration.Predicate{Field: &exploration.Field{Kind: "resource", Key: "service.name"}, Op: "eq", Value: "api"}}, Lanes: &ActivityLanes{Logs: ActivityLayout{Columns: []string{"summary", "utc"}, FieldColumns: []exploration.Field{{Kind: "column", Key: "body"}}}, Spans: ActivityLayout{Columns: []string{"utc", "service", "summary"}, FieldColumns: []exploration.Field{{Kind: "column", Key: "name"}}}}}
}

func TestPairedActivityIdentityAndDetachedPresentation(t *testing.T) {
	store := Store{Dir: t.TempDir()}
	view := pairedActivityFixture()
	saved, err := store.CreateActivityView(view)
	if err != nil {
		t.Fatal(err)
	}
	same, err := store.CreateActivityView(view)
	if err != nil || same.ID != saved.ID {
		t.Fatal("unstable paired ID", err)
	}
	view.Lanes.Logs.Columns[0] = "service"
	view.Lanes.Spans.FieldColumns[0].Key = "duration_ns"
	view.Selector.Where.Value = "changed"
	held, err := store.LoadActivityView(saved.ID)
	if err != nil || !reflect.DeepEqual(*held.Lanes, *saved.Lanes) || held.Selector.Where.Value != "api" {
		t.Fatal("caller mutation leaked", held, err)
	}
	changed := pairedActivityFixture()
	changed.Lanes.Spans.Columns = []string{"summary", "utc"}
	next, err := store.CreateActivityView(changed)
	if err != nil || next.ID == saved.ID {
		t.Fatal("spans layout not bound to identity", err)
	}
	queries, err := held.Selector.Queries()
	if err != nil || len(queries) != 2 || queries[0].Signal != "logs" || queries[1].Signal != "spans" || !queries[0].SameEnvironment(queries[1]) {
		t.Fatal("bad physical projections", queries, err)
	}
	queries[0].Where.Value = "changed projection"
	if held.Selector.Where.Value != "api" || queries[1].Where.Value != "api" {
		t.Fatal("query projection leaked mutation")
	}
	if exploration.TailSelector(held.Selector).Query(60000).Validate() == nil {
		t.Fatal("paired selector became executable physical query")
	}
	raw, err := json.Marshal(held)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"columns", "fieldColumns", "baselineCapture", "readWindow", "connection", "rows"} {
		if _, ok := fields[key]; ok {
			t.Fatal("paired view has flat/runtime field", key)
		}
	}
}

func TestPairedActivityWholePredicateAndLayoutAdmission(t *testing.T) {
	for name, mutate := range map[string]func(*ActivityView){
		"missing lanes":      func(v *ActivityView) { v.Lanes = nil },
		"missing spans":      func(v *ActivityView) { v.Lanes.Spans = ActivityLayout{} },
		"flat empty columns": func(v *ActivityView) { v.Columns = []string{} },
		"flat empty fields":  func(v *ActivityView) { v.FieldColumns = []exploration.Field{} },
		"baseline": func(v *ActivityView) {
			v.BaselineCapture = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		},
		"log-only predicate":  func(v *ActivityView) { v.Selector.Where.Field = &exploration.Field{Kind: "column", Key: "body"} },
		"span-only predicate": func(v *ActivityView) { v.Selector.Where.Field = &exploration.Field{Kind: "column", Key: "name"} },
		"wrong log field": func(v *ActivityView) {
			v.Lanes.Logs.FieldColumns = []exploration.Field{{Kind: "column", Key: "name"}}
		},
		"context":      func(v *ActivityView) { v.Lanes.Spans.Columns = append(v.Lanes.Spans.Columns, "context") },
		"single lanes": func(v *ActivityView) { v.Selector.Signal = "logs" },
		"all environments": func(v *ActivityView) {
			v.Selector.Environment = ""
			v.Selector.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "current-only", Scope: identity.EnvironmentScope{Mode: "all"}}
		},
	} {
		t.Run(name, func(t *testing.T) {
			view := pairedActivityFixture()
			mutate(&view)
			if err := view.Validate(); err == nil {
				t.Fatal("accepted incompatible composition")
			}
		})
	}
}

func TestPairedActivityCatalogKeepsExactEnvironmentIdentity(t *testing.T) {
	store := Store{Dir: t.TempDir()}
	for _, scope := range []identity.EnvironmentScope{{Mode: "named", Value: "prod"}, {Mode: "empty"}, {Mode: "unreported"}} {
		view := pairedActivityFixture()
		view.Selector.Environment = ""
		wire := identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: scope}
		view.Selector.EnvironmentSelector = &wire
		saved, err := store.CreateActivityView(view)
		if err != nil {
			t.Fatal(err)
		}
		selected, err := wire.Selection()
		if err != nil {
			t.Fatal(err)
		}
		page, err := store.ActivityViewsForEnvironment("", &selected)
		if err != nil || len(page.Items) != 1 || page.Items[0].ID != saved.ID {
			t.Fatal("paired catalog widened scope", page, err)
		}
	}
	current, err := store.ActivityViews("", "prod")
	if err != nil || len(current.Items) != 0 {
		t.Fatal("effective policy aliased current", current, err)
	}
}
