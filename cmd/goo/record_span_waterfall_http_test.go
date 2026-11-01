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

func TestRecordSpanWaterfallHTTPConfigurationHasNoSourceRead(t *testing.T) {
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
	q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100, GroupBy: []exploration.Field{{Kind: "column", Key: "service"}}}
	source := workbench.ExplorationView{Version: 1, Title: "Supplied span timing", Queries: []exploration.Query{q}, Template: &workbench.ViewTemplateBinding{Version: "goo.template/1", Group: "destination", Profile: "ordinary"}, Composition: &workbench.ViewComposition{
		Version: "goo.composition/1", Kind: workbench.CompositionKind{Name: "stack", Version: 1}, Children: []workbench.ViewCompositionChild{
			{Aggregate: &workbench.ViewDeclaration{Version: "goo.view/1", Scope: workbench.ViewDeclarationScope{Kind: "record-query", Signal: "spans"}, Kind: workbench.ViewKind{Name: "bar", Version: 1}, Display: workbench.ViewDisplay{Category: "category", Value: "count"}}},
			{Records: &workbench.RecordViewDeclaration{Version: "goo.view/1", Scope: workbench.ViewDeclarationScope{Kind: "record-query", Signal: "spans"}, Kind: workbench.RecordViewKind{Name: "record-span-waterfall", Version: 1}, Display: workbench.RecordViewDisplay{Columns: []string{"utc", "summary", "context"}, FieldColumns: []exploration.Field{{Kind: "attribute", Key: "zone"}, {Kind: "resource", Key: "zone"}}}}, Hidden: true},
		},
	}}
	response := call("POST", "/exploration-views", source)
	if response.Code != 201 {
		t.Fatal(response.Code, response.Body.String())
	}
	var saved, loaded workbench.ExplorationView
	if err := json.Unmarshal(response.Body.Bytes(), &saved); err != nil {
		t.Fatal(err)
	}
	response = call("GET", "/exploration-views/"+saved.ID, nil)
	if response.Code != 200 {
		t.Fatal(response.Code, response.Body.String())
	}
	if err := json.Unmarshal(response.Body.Bytes(), &loaded); err != nil || !reflect.DeepEqual(saved, loaded) {
		t.Fatal("waterfall changed on reload", err)
	}
	target := q
	target.Environment = "staging"
	target.FromMS = 70000
	target.ToMS = 130000
	target.Offset = 100
	target.GroupBy = []exploration.Field{{Kind: "attribute", Key: "zone"}}
	response = call("POST", "/view-templates/expand", workbench.ViewTemplateExpansionRequest{Source: &loaded, Query: &target})
	if response.Code != 200 {
		t.Fatal(response.Code, response.Body.String())
	}
	var expanded workbench.ViewTemplateExpansion
	if err := json.Unmarshal(response.Body.Bytes(), &expanded); err != nil {
		t.Fatal(err)
	}
	target.Offset = 0
	if !reflect.DeepEqual(expanded.View.Queries[0], target) || !reflect.DeepEqual(expanded.View.Composition, source.Composition) {
		t.Fatal("waterfall imported query or changed presentation")
	}
	// The same otherwise valid ordinary logs view remains forbidden, including a
	// hidden waterfall; a hidden role does not weaken its input contract.
	for _, hidden := range []bool{false, true} {
		source.Composition.Children[1].Hidden = hidden
		source.Queries[0].Signal = "logs"
		source.Composition.Children[0].Aggregate.Scope.Signal = "logs"
		source.Composition.Children[1].Records.Scope.Signal = "logs"
		response = call("POST", "/exploration-views", source)
		if response.Code != 400 {
			t.Fatal("logs waterfall published", hidden, response.Code, response.Body.String())
		}
		logs := source.Queries[0]
		response = call("POST", "/view-templates/expand", workbench.ViewTemplateExpansionRequest{Source: &source, Query: &logs})
		if response.Code != 400 {
			t.Fatal("logs waterfall expanded", hidden, response.Code, response.Body.String())
		}
	}
	if reads != 0 {
		t.Fatal("configuration accessed telemetry", reads)
	}
}
