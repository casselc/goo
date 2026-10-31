package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
)

func TestCompositionSavedViewHTTPPreservesOneQueryAndSinglePresentationOwner(t *testing.T) {
	mux := http.NewServeMux()
	registerWorkbench(mux, t.TempDir(), retentionConnection{URL: "http://unused.invalid"}, &exploration.ResultSnapshots{}, nil)
	in := workbench.ExplorationView{Version: 1, Title: "One query, two supplied aspects", Queries: []exploration.Query{{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 1, GroupBy: []exploration.Field{{Kind: "column", Key: "service"}}}}, Composition: &workbench.ViewComposition{
		Version: "goo.composition/1", Kind: workbench.CompositionKind{Name: "stack", Version: 1}, Children: []workbench.ViewCompositionChild{
			{Aggregate: &workbench.ViewDeclaration{Version: "goo.view/1", Scope: workbench.ViewDeclarationScope{Kind: "record-query", Signal: "spans"}, Kind: workbench.ViewKind{Name: "bar", Version: 1}, Display: workbench.ViewDisplay{Category: "category", Value: "count", Scale: "population"}}},
			{Records: &workbench.RecordViewDeclaration{Version: "goo.view/1", Scope: workbench.ViewDeclarationScope{Kind: "record-query", Signal: "spans"}, Kind: workbench.RecordViewKind{Name: "record-table", Version: 1}, Display: workbench.RecordViewDisplay{Columns: []string{"utc", "summary"}}}},
		},
	}}
	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	post := func(body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/exploration-views", strings.NewReader(body)))
		return w
	}
	w := post(string(raw))
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	var saved workbench.ExplorationView
	if err = json.Unmarshal(w.Body.Bytes(), &saved); err != nil || saved.Composition == nil || len(saved.Queries) != 1 || saved.Queries[0].Limit != 1 {
		t.Fatal("composition not saved", err, w.Body.String())
	}
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/exploration-views/"+saved.ID, nil))
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var reopened map[string]json.RawMessage
	if err = json.Unmarshal(w.Body.Bytes(), &reopened); err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"layout", "columns", "fieldColumns", "displays", "breakdowns", "tail", "rows"} {
		if _, present := reopened[forbidden]; present {
			t.Fatal("multiple presentation owners or runtime payload", forbidden)
		}
	}
	for _, extra := range []string{`"columns":[]`, `"layout":""`, `"displays":null`, `"tail":null`, `"rows":[]`} {
		body := string(raw[:len(raw)-1]) + "," + extra + "}"
		if got := post(body); got.Code != 400 {
			t.Fatal("invalid composition accepted", extra, got.Code, got.Body.String())
		}
	}
}
