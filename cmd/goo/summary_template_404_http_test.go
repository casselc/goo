package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/workbench"
)

func TestSummaryTemplate404HTTPStoreReopenExpandWithoutSource(t *testing.T) {
	reads := 0
	previous := http.DefaultTransport
	http.DefaultTransport = templateNoSourceTransport{&reads}
	t.Cleanup(func() { http.DefaultTransport = previous })
	mux := http.NewServeMux()
	registerWorkbench(mux, t.TempDir(), retentionConnection{URL: "http://source.invalid"}, &exploration.ResultSnapshots{}, nil)
	call := func(method, path string, body any, want int) []byte {
		t.Helper()
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(string(raw))))
		if w.Code != want {
			t.Fatalf("%s %s: %d want %d: %s", method, path, w.Code, want, w.Body.String())
		}
		return append([]byte(nil), w.Body.Bytes()...)
	}
	for _, signal := range []string{"logs", "spans"} {
		q := exploration.Query{Signal: signal, Environment: "source", FromMS: 1000, ToMS: 61000, Limit: 7}
		fields := []string{"count"}
		if signal == "spans" {
			q.Summary = "span-duration"
			q.SpanStatus = true
			fields = []string{"status.error", "count", "duration.mean"}
		}
		scope := workbench.ViewDeclarationScope{Kind: "record-query", Signal: signal}
		source := workbench.ExplorationView{Version: 1, Title: "Reusable " + signal + " facts", Queries: []exploration.Query{q},
			Template: &workbench.ViewTemplateBinding{Version: "goo.template/1", Group: "none", Profile: "ordinary-summary"},
			Composition: &workbench.ViewComposition{Version: "goo.composition/1", Kind: workbench.CompositionKind{Name: "stack", Version: 1}, Children: []workbench.ViewCompositionChild{
				{Summary: &workbench.SummaryViewDeclaration{Version: "goo.view/1", Scope: scope, Kind: workbench.SummaryViewKind{Name: "measure-summary", Version: 1}, Display: workbench.SummaryViewDisplay{Fields: fields}}},
				{Records: &workbench.RecordViewDeclaration{Version: "goo.view/1", Scope: scope, Kind: workbench.RecordViewKind{Name: "record-list", Version: 1}, Display: workbench.RecordViewDisplay{Columns: []string{"utc", "summary"}}}},
			}}}
		stored := call("POST", "/exploration-views", source, 201)
		var saved, loaded workbench.ExplorationView
		if err := json.Unmarshal(stored, &saved); err != nil {
			t.Fatal(err)
		}
		reopened := call("GET", "/exploration-views/"+saved.ID, nil, 200)
		if err := json.Unmarshal(reopened, &loaded); err != nil || !reflect.DeepEqual(saved, loaded) {
			t.Fatal("stored template changed on reopen", err)
		}
		destination := q
		destination.Environment = ""
		destination.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "empty"}}
		destination.FromMS = 62000
		destination.ToMS = 122000
		destination.Limit = 3
		service := exploration.Field{Kind: "column", Key: "service"}
		destination.Where = &exploration.Predicate{Field: &service, Op: "eq", Value: "destination-api"}
		before, _ := json.Marshal(saved)
		raw := call("POST", "/view-templates/expand", workbench.ViewTemplateExpansionRequest{Source: &saved, Query: &destination}, 200)
		var expanded workbench.ViewTemplateExpansion
		if err := json.Unmarshal(raw, &expanded); err != nil {
			t.Fatal(err)
		}
		after, _ := json.Marshal(saved)
		if expanded.View.ID != "" || !reflect.DeepEqual(expanded.View.Queries, []exploration.Query{destination}) || !reflect.DeepEqual(expanded.View.Composition, saved.Composition) || string(before) != string(after) {
			t.Fatal("expansion changed destination scope, presentation or donor")
		}
		call("POST", "/exploration-views", expanded.View, 201)
		for _, mutate := range []func(*exploration.Query){func(q *exploration.Query) { q.Offset = 1 }, func(q *exploration.Query) { q.GroupBy = []exploration.Field{{Kind: "column", Key: "service"}} }, func(q *exploration.Query) {
			q.Numeric = &exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work.amount"}, Type: "int64", Unit: "{item}", Parse: "canonical-decimal"}
		}} {
			bad := destination
			mutate(&bad)
			call("POST", "/view-templates/expand", workbench.ViewTemplateExpansionRequest{Source: &saved, Query: &bad}, 400)
		}
		baseline := strings.Repeat("a", 64)
		call("POST", "/view-templates/expand", workbench.ViewTemplateExpansionRequest{Source: &saved, Query: &destination, BaselineCapture: &baseline}, 400)
		bad := source
		bad.Template = &workbench.ViewTemplateBinding{Version: "goo.template/1", Group: "destination", Profile: "ordinary"}
		call("POST", "/exploration-views", bad, 400)
		bad = source
		bad.Queries = []exploration.Query{q}
		bad.Queries[0].Offset = 1
		call("POST", "/exploration-views", bad, 400)
	}
	if reads != 0 {
		t.Fatalf("configuration lifecycle read telemetry: %d", reads)
	}
}
