package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
)

type templateNoSourceTransport struct{ requests *int }

func (x templateNoSourceTransport) RoundTrip(*http.Request) (*http.Response, error) {
	*x.requests++
	return nil, errors.New("template expansion must not access a source")
}

func TestViewTemplateExpansionHTTPIsConfigurationOnly(t *testing.T) {
	reads := 0
	previous := http.DefaultTransport
	http.DefaultTransport = templateNoSourceTransport{&reads}
	t.Cleanup(func() { http.DefaultTransport = previous })
	dir := t.TempDir()
	mux := http.NewServeMux()
	registerWorkbench(mux, dir, retentionConnection{URL: "http://source.invalid"}, &exploration.ResultSnapshots{}, nil)
	plan, e := exploration.PlanServiceDashboard(exploration.DashboardRequest{Profile: "http-server", HTTPServerObserved: true, Scope: exploration.InventoryQuestion{Environment: "prod", FromMS: 1000, ToMS: 61000}, Service: "api", Signals: []string{"spans"}, Fields: []exploration.Field{{Kind: "attribute", Key: "http.request.method"}}}, nil)
	if e != nil {
		t.Fatal(e)
	}
	var in workbench.ViewTemplateExpansionRequest
	for _, widget := range plan.Widgets {
		if widget.Key == "dimension" {
			in = workbench.ViewTemplateExpansionRequest{Template: "otel.http-server-routes", Version: 1, Widget: widget}
		}
	}
	post := func(raw string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/view-templates/expand", strings.NewReader(raw)))
		return w
	}
	raw, _ := json.Marshal(in)
	w := post(string(raw))
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var out workbench.ViewTemplateExpansion
	if e := json.Unmarshal(w.Body.Bytes(), &out); e != nil || out.View.ID != "" || out.View.Validate() != nil {
		t.Fatal("invalid staged view", e, w.Body.String())
	}
	if reads != 0 {
		t.Fatal("template expansion read telemetry", reads)
	}
	if _, e := os.Stat(filepath.Join(dir, "workbench")); !os.IsNotExist(e) {
		t.Fatal("template expansion published configuration", e)
	}
	for _, bad := range []string{strings.Replace(string(raw), `"version":1`, `"version":2`, 1), string(raw[:len(raw)-1]) + `,"rows":[]}`, string(raw) + ` {}`, strings.Replace(string(raw), `"Server"`, `"Client"`, 1)} {
		if got := post(bad); got.Code != 400 {
			t.Fatal("invalid template input accepted", got.Code, got.Body.String())
		}
	}
	if reads != 0 {
		t.Fatal("refusal read telemetry", reads)
	}
}
