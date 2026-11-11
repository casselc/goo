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

func TestHistogramTemplate407HTTPStoreReopenExpandWithoutSource(t *testing.T) {
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
		for _, typ := range []string{"int64", "uint64", "decimal64"} {
			binding := exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work.amount"}, Type: typ, Unit: "{item}", Parse: "canonical-decimal"}
			if typ == "decimal64" {
				binding.Scale = 2
			}
			q := exploration.Query{Signal: signal, Environment: "source", FromMS: 1000, ToMS: 61000, Limit: 7, Numeric: &binding, NumericHistogram: &exploration.NumericHistogramSpec{Binning: "powers-of-two"}}
			scope := workbench.ViewDeclarationScope{Kind: "record-query", Signal: signal}
			for _, kind := range []string{"histogram", "table"} {
				source := workbench.ExplorationView{Version: 1, Title: "Reusable " + typ + " " + signal,
					Queries: []exploration.Query{q}, Template: &workbench.ViewTemplateBinding{Version: "goo.template/1", Group: "none", Profile: "ordinary-histogram"},
					Composition: &workbench.ViewComposition{Version: "goo.composition/1", Kind: workbench.CompositionKind{Name: "stack", Version: 1}, Children: []workbench.ViewCompositionChild{
						{Histogram: &workbench.HistogramViewDeclaration{Version: "goo.view/1", Scope: scope, Kind: workbench.HistogramViewKind{Name: kind, Version: 1}, Display: workbench.HistogramViewDisplay{Category: "from", Value: "count", Title: "Captured amounts", RequiresComplete: true}}},
						{Records: &workbench.RecordViewDeclaration{Version: "goo.view/1", Scope: scope, Kind: workbench.RecordViewKind{Name: "record-list", Version: 1}, Display: workbench.RecordViewDisplay{Columns: []string{"utc", "summary"}}}},
					}}}
				stored := call("POST", "/exploration-views", source, 201)
				var saved, loaded workbench.ExplorationView
				if err := json.Unmarshal(stored, &saved); err != nil {
					t.Fatal(err)
				}
				reopened := call("GET", "/exploration-views/"+saved.ID, nil, 200)
				if err := json.Unmarshal(reopened, &loaded); err != nil || !reflect.DeepEqual(saved, loaded) {
					t.Fatal("stored histogram configuration changed on reopen", err)
				}
				destination := q
				destination.Environment = ""
				destination.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "empty"}}
				destination.FromMS, destination.ToMS, destination.Limit = 62000, 122000, 3
				destination.BucketMS = 10000 // An independent valid directive remains destination-owned.
				if signal == "spans" {
					destination.Summary, destination.SpanStatus = "span-duration", true
				}
				service := exploration.Field{Kind: "column", Key: "service"}
				destination.Where = &exploration.Predicate{All: []exploration.Predicate{{Field: &service, Op: "eq", Value: "destination-api"}, {Numeric: &binding, Op: "gte", Value: "0"}}}
				if typ == "decimal64" {
					destination.Where.All[1].Value = "0.00"
				}
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
				for _, mutate := range []func(*exploration.Query){
					func(q *exploration.Query) { q.Offset = 1 },
					func(q *exploration.Query) { q.GroupBy = []exploration.Field{{Kind: "column", Key: "service"}} },
					func(q *exploration.Query) { q.Numeric = nil },
					func(q *exploration.Query) { q.NumericHistogram = nil },
					func(q *exploration.Query) { q.Numeric.Unit = "items" },
					func(q *exploration.Query) { q.Numeric.Field.Kind = "resource" },
					func(q *exploration.Query) { q.Numeric.Field.Key = "other.amount" },
					func(q *exploration.Query) { q.Numeric.Parse = "permissive" },
					func(q *exploration.Query) { q.NumericHistogram.Binning = "adaptive" },
				} {
					wire, _ := json.Marshal(destination)
					var bad exploration.Query
					if err := json.Unmarshal(wire, &bad); err != nil {
						t.Fatal(err)
					}
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
		}
	}
	if reads != 0 {
		t.Fatalf("configuration lifecycle read telemetry: %d", reads)
	}
}
