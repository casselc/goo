package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/workbench"
)

func TestCurrentRPCViewTemplateHTTP367(t *testing.T) {
	reads := 0
	previous := http.DefaultTransport
	http.DefaultTransport = templateNoSourceTransport{&reads}
	t.Cleanup(func() { http.DefaultTransport = previous })
	mux := http.NewServeMux()
	registerWorkbench(mux, t.TempDir(), retentionConnection{URL: "http://source.invalid"}, &exploration.ResultSnapshots{}, nil)
	type observation struct {
		Name     string            `json:"name"`
		Method   string            `json:"method"`
		Path     string            `json:"path"`
		Status   int               `json:"status"`
		Headers  map[string]string `json:"headers"`
		Body     string            `json:"body"`
		Response string            `json:"response"`
	}
	var cases []observation
	call := func(name, method, path, pattern string, body any, want int) *httptest.ResponseRecorder {
		t.Helper()
		var raw []byte
		var err error
		if body != nil {
			raw, err = json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(string(raw))))
		if w.Code != want {
			t.Fatal(name, w.Code, w.Body.String())
		}
		if w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("cacheable configuration response", name)
		}
		if want < 400 {
			cases = append(cases, observation{name, method, pattern, w.Code, map[string]string{"Content-Type": w.Header().Get("Content-Type"), "Cache-Control": w.Header().Get("Cache-Control")}, string(raw), w.Body.String()})
		}
		return w
	}
	widget := func(service, namespace string, from int64, scope identity.EnvironmentScope) exploration.DashboardWidget {
		t.Helper()
		selection, err := identity.EffectiveEnvironment(scope)
		if err != nil {
			t.Fatal(err)
		}
		selector, err := selection.Selector()
		if err != nil {
			t.Fatal(err)
		}
		plan, err := exploration.PlanServiceDashboard(exploration.DashboardRequest{Profile: "rpc-server-current", RPCServerCurrentObserved: true, Scope: exploration.InventoryQuestion{EnvironmentSelector: &selector, Namespace: &identity.Namespace{Mode: "value", Value: namespace}, FromMS: from, ToMS: from + 60000}, Service: service, Signals: []string{"spans"}, Fields: []exploration.Field{{Kind: "attribute", Key: "rpc.system.name"}}}, nil)
		if err != nil {
			t.Fatal(err)
		}
		for _, w := range plan.Widgets {
			if w.Key == "dimension" {
				return w
			}
		}
		t.Fatal("no conditional method population")
		return exploration.DashboardWidget{}
	}
	in := workbench.ViewTemplateExpansionRequest{Template: "otel.rpc-server-current-methods", Version: 1, Widget: widget("source-rpc", "source-ns", 1000, identity.EnvironmentScope{Mode: "named", Value: "source-env"})}
	w := call("rpc-shipped", "POST", "/view-templates/expand", "/view-templates/expand", in, 200)
	var expanded workbench.ViewTemplateExpansion
	if err := json.Unmarshal(w.Body.Bytes(), &expanded); err != nil {
		t.Fatal(err)
	}
	if expanded.View.Template == nil || expanded.View.Template.Profile != "otel.rpc-server-current" || !reflect.DeepEqual(expanded.View.Queries[0], *in.Widget.Query) {
		t.Fatal("profile/source question changed")
	}
	expanded.View.Title = "Current RPC method review"
	w = call("rpc-save", "POST", "/exploration-views", "/exploration-views", expanded.View, 201)
	var saved workbench.ExplorationView
	if err := json.Unmarshal(w.Body.Bytes(), &saved); err != nil {
		t.Fatal(err)
	}
	w = call("rpc-reopen", "GET", "/exploration-views/"+saved.ID, "/exploration-views/{id}", nil, 200)
	if err := json.Unmarshal(w.Body.Bytes(), &saved); err != nil {
		t.Fatal(err)
	}
	destination := *widget("target-rpc", "target-ns", 80000, identity.EnvironmentScope{Mode: "empty"}).Query
	destination.GroupBy = []exploration.Field{{Kind: "attribute", Key: "rpc.system.name"}}
	destination.Where.All = append(destination.Where.All, exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "tenant"}, Op: "eq", Value: "target-only"})
	w = call("rpc-reuse", "POST", "/view-templates/expand", "/view-templates/expand", workbench.ViewTemplateExpansionRequest{Source: &saved, Query: &destination}, 200)
	if err := json.Unmarshal(w.Body.Bytes(), &expanded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expanded.View.Queries[0], destination) {
		t.Fatal("donor scope replaced destination")
	}
	for _, mutate := range []func(*exploration.Query){
		func(q *exploration.Query) {
			for i := range q.Where.All {
				p := &q.Where.All[i]
				if p.Field != nil && p.Field.Key == "kind" {
					p.Value = "Client"
				}
			}
		},
		func(q *exploration.Query) {
			for i := range q.Where.All {
				p := &q.Where.All[i]
				if p.Field != nil && p.Field.Key == "rpc.system.name" {
					p.Field.Key = "rpc.system"
				}
			}
		},
	} {
		raw, _ := json.Marshal(destination)
		var bad exploration.Query
		if err := json.Unmarshal(raw, &bad); err != nil {
			t.Fatal(err)
		}
		mutate(&bad)
		call("rpc-role-refused", "POST", "/view-templates/expand", "", workbench.ViewTemplateExpansionRequest{Source: &saved, Query: &bad}, 400)
	}
	if reads != 0 {
		t.Fatal("configuration acquired telemetry", reads)
	}
	if file := os.Getenv("GOO_RPC_TEMPLATE_367_OBSERVATIONS"); file != "" {
		raw, err := json.MarshalIndent(struct {
			Cases []observation `json:"cases"`
		}{cases}, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, append(raw, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
}
