package workbench

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
)

func httpRouteTemplateInput(t *testing.T) ViewTemplateExpansionRequest {
	t.Helper()
	plan, err := exploration.PlanServiceDashboard(exploration.DashboardRequest{Profile: "http-server", HTTPServerObserved: true, Scope: exploration.InventoryQuestion{Environment: "prod", FromMS: 1000, ToMS: 61000}, Service: "api", Signals: []string{"spans"}, Fields: []exploration.Field{{Kind: "attribute", Key: "http.request.method"}}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, widget := range plan.Widgets {
		if widget.Key == "dimension" {
			return ViewTemplateExpansionRequest{Template: "otel.http-server-routes", Version: 1, Widget: widget}
		}
	}
	t.Fatal("no HTTP route widget")
	return ViewTemplateExpansionRequest{}
}

func TestViewTemplatePreservesCheckedHTTPQueryAndDetaches(t *testing.T) {
	in := httpRouteTemplateInput(t)
	selection, err := identity.EffectiveEnvironment(identity.EnvironmentScope{Mode: "unreported"})
	if err != nil {
		t.Fatal(err)
	}
	wire, err := selection.Selector()
	if err != nil {
		t.Fatal(err)
	}
	in.Widget.Query.Environment = ""
	in.Widget.Query.EnvironmentSelector = &wire
	in.Widget.Query.Numeric = &exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work"}, Type: "int64", Unit: "items", Parse: "canonical-decimal"}
	extra := exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "team"}, Op: "eq", Value: "platform"}
	in.Widget.Query.Where.All = append(in.Widget.Query.Where.All, extra)
	original, _ := json.Marshal(in)
	out, err := ExpandViewTemplate(in)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(out.View.Queries[0], *in.Widget.Query) || out.Template.ID != in.Template || out.Template.Version != 1 || out.View.ID != "" || out.View.Tail != nil {
		t.Fatal("template changed query or published configuration", out)
	}
	d := *out.View.Composition.Children[0].Aggregate
	if d.Kind.Name != "bar" || d.Display.Value != "count" || d.Display.Scale != "population" {
		t.Fatal("wrong shipped display", d)
	}
	out.View.Queries[0].Where.All[0].Value = "changed"
	out.View.Queries[0].GroupBy[0].Key = "changed"
	out.View.Queries[0].Numeric.Unit = "changed"
	out.View.Queries[0].EnvironmentSelector.Scope.Mode = "empty"
	out.View.Composition.Children[1].Records.Display.Columns[0] = "changed"
	d.Display.Title = "changed"
	*out.View.Composition.Children[0].Aggregate = d
	after, _ := json.Marshal(in)
	if string(original) != string(after) {
		t.Fatal("expansion aliases widget query")
	}
	next, err := ExpandViewTemplate(in)
	if err != nil || next.View.Composition.Children[1].Records.Display.Columns[0] != "utc" || next.View.Composition.Children[0].Aggregate.Display.Title != "HTTP server routes" {
		t.Fatal("expansion aliases shipped defaults", err)
	}
	store := Store{Dir: t.TempDir()}
	saved, err := store.CreateExplorationView(next.View)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := store.LoadExplorationView(saved.ID)
	if err != nil || !reflect.DeepEqual(loaded.Queries, next.View.Queries) {
		t.Fatal("template does not round trip existing catalog", err)
	}
}

func TestViewTemplateRefusesUnsupportedAndWeakenedPopulations(t *testing.T) {
	mutations := map[string]func(*ViewTemplateExpansionRequest){
		"mixed native payload": func(in *ViewTemplateExpansionRequest) { in.Widget.Measurement = &measurements.Question{} },
		"template":             func(in *ViewTemplateExpansionRequest) { in.Template = "arbitrary" },
		"version":              func(in *ViewTemplateExpansionRequest) { in.Version = 2 },
		"family label":         func(in *ViewTemplateExpansionRequest) { in.Widget.Family = "all-spans" },
		"state":                func(in *ViewTemplateExpansionRequest) { in.Widget.State = "not_observed" },
		"key":                  func(in *ViewTemplateExpansionRequest) { in.Widget.Key = "spans" },
		"grain":                func(in *ViewTemplateExpansionRequest) { in.Widget.Grain = "span-record" },
		"missing query":        func(in *ViewTemplateExpansionRequest) { in.Widget.Query = nil },
		"route namespace":      func(in *ViewTemplateExpansionRequest) { in.Widget.Query.GroupBy[0].Kind = "resource" },
		"page":                 func(in *ViewTemplateExpansionRequest) { in.Widget.Query.Offset = 100 },
		"no service":           func(in *ViewTemplateExpansionRequest) { in.Widget.Query.Where.All = in.Widget.Query.Where.All[1:] },
		"empty service":        func(in *ViewTemplateExpansionRequest) { in.Widget.Query.Where.All[0].Value = "" },
		"client":               func(in *ViewTemplateExpansionRequest) { in.Widget.Query.Where.All[1].Value = "Client" },
		"legacy method":        func(in *ViewTemplateExpansionRequest) { in.Widget.Query.Where.All[2].Field.Key = "http.method" },
		"empty method allowed": func(in *ViewTemplateExpansionRequest) { in.Widget.Query.Where.All[2].Op = "present" },
		"OR weakening": func(in *ViewTemplateExpansionRequest) {
			in.Widget.Query.Where = &exploration.Predicate{Any: []exploration.Predicate{*in.Widget.Query.Where, {Field: &exploration.Field{Kind: "column", Key: "kind"}, Op: "eq", Value: "Client"}}}
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			in := httpRouteTemplateInput(t)
			mutate(&in)
			if _, err := ExpandViewTemplate(in); err == nil {
				t.Fatal("unsupported template accepted")
			}
		})
	}
	// Additional OR filters can narrow a population once mandatory predicates
	// remain outside Any; nested All guards retain exactly the same query.
	in := httpRouteTemplateInput(t)
	in.Widget.Query.Where = &exploration.Predicate{All: []exploration.Predicate{*in.Widget.Query.Where, {Any: []exploration.Predicate{{Field: &exploration.Field{Kind: "attribute", Key: "tenant"}, Op: "eq", Value: "a"}, {Field: &exploration.Field{Kind: "attribute", Key: "tenant"}, Op: "eq", Value: "b"}}}}}
	if _, err := ExpandViewTemplate(in); err != nil {
		t.Fatal("safe additional filter rejected", err)
	}
}
