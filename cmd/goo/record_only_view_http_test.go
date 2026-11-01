package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
)

func TestRecordsOnlySavedHTTPConfigurationDoesNotReadSource302(t *testing.T) {
	reads := 0
	previous := http.DefaultTransport
	http.DefaultTransport = templateNoSourceTransport{&reads}
	t.Cleanup(func() { http.DefaultTransport = previous })
	mux := http.NewServeMux()
	registerWorkbench(mux, t.TempDir(), retentionConnection{URL: "http://source.invalid"}, &exploration.ResultSnapshots{}, nil)
	call := func(method, path string, body any) *httptest.ResponseRecorder {
		t.Helper()
		raw, _ := json.Marshal(body)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(string(raw))))
		return w
	}
	for _, signal := range []string{"logs", "spans"} {
		kind := "record-list"
		if signal == "spans" {
			kind = "record-span-waterfall"
		}
		q := exploration.Query{Signal: signal, Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100}
		v := workbench.ExplorationView{Version: 1, Title: "Raw " + signal, Queries: []exploration.Query{q}, Composition: &workbench.ViewComposition{Version: "goo.composition/1", Kind: workbench.CompositionKind{Name: "stack", Version: 1}, Children: []workbench.ViewCompositionChild{{Records: &workbench.RecordViewDeclaration{Version: "goo.view/1", Scope: workbench.ViewDeclarationScope{Kind: "record-query", Signal: signal}, Kind: workbench.RecordViewKind{Name: kind, Version: 1}, Display: workbench.RecordViewDisplay{Columns: []string{"utc", "summary"}}}}}}}
		w := call("POST", "/exploration-views", v)
		if w.Code != 201 {
			t.Fatal(w.Code, w.Body.String())
		}
		var saved, loaded workbench.ExplorationView
		if err := json.Unmarshal(w.Body.Bytes(), &saved); err != nil {
			t.Fatal(err)
		}
		w = call("GET", "/exploration-views/"+saved.ID, nil)
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		if err := json.Unmarshal(w.Body.Bytes(), &loaded); err != nil || !reflect.DeepEqual(saved, loaded) || !reflect.DeepEqual(loaded.Queries, v.Queries) {
			t.Fatal("raw HTTP configuration changed", err)
		}
		// Keep the category-template boundary even when a caller supplies a valid
		// grouping: the new raw composition is not an authored aggregate template.
		v.Queries[0].GroupBy = []exploration.Field{{Kind: "column", Key: "service"}}
		v.Template = &workbench.ViewTemplateBinding{Version: "goo.template/1", Group: "destination", Profile: "ordinary"}
		w = call("POST", "/exploration-views", v)
		if w.Code != 400 {
			t.Fatal("raw category template published", w.Code, w.Body.String())
		}
		target := v.Queries[0]
		w = call("POST", "/view-templates/expand", workbench.ViewTemplateExpansionRequest{Source: &v, Query: &target})
		if w.Code != 400 {
			t.Fatal("raw category template expanded", w.Code, w.Body.String())
		}
		v.Template = nil
		v.Composition.Children[0].Hidden = true
		w = call("POST", "/exploration-views", v)
		if w.Code != 400 {
			t.Fatal("hidden-only records published", w.Code, w.Body.String())
		}
	}
	if reads != 0 {
		t.Fatal("saved configuration acquired telemetry", reads)
	}
}
