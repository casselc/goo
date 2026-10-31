package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/casselc/goo/workbench"
)

func TestActivityViewsTypedScopeCatalogHTTP(t *testing.T) {
	mux := http.NewServeMux()
	registerActivityViews(mux, t.TempDir())
	call := func(method, path, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(body)))
		return w
	}
	wires := []string{
		`{"version":"goo/environment-selector/v1","policy":"current-only","scope":{"mode":"named","value":"prod"}}`,
		`{"version":"goo/environment-selector/v1","policy":"effective","scope":{"mode":"named","value":"prod"}}`,
		`{"version":"goo/environment-selector/v1","policy":"effective","scope":{"mode":"empty"}}`,
		`{"version":"goo/environment-selector/v1","policy":"effective","scope":{"mode":"unreported"}}`,
	}
	for _, wire := range wires {
		body := `{"version":1,"title":"Exact scope","selector":{"signal":"logs","environment":"","environmentSelector":` + wire + `,"where":{"field":{"kind":"attribute","key":"level"},"op":"eq","value":"warn"}},"columns":["utc","summary"]}`
		w := call("POST", "/activity-views", body)
		if w.Code != 201 {
			t.Fatal(w.Code, w.Body.String())
		}
		var saved workbench.ActivityView
		if e := json.Unmarshal(w.Body.Bytes(), &saved); e != nil {
			t.Fatal(e)
		}
		w = call("GET", "/activity-views/"+saved.ID, "")
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		w = call("GET", "/activity-views?environmentSelector="+url.QueryEscape(wire), "")
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		var page workbench.ActivityViewList
		if e := json.Unmarshal(w.Body.Bytes(), &page); e != nil || len(page.Items) != 1 || page.Items[0].ID != saved.ID {
			t.Fatal(page, e)
		}
	}
	w := call("GET", "/activity-views?environment=prod", "")
	var current workbench.ActivityViewList
	json.Unmarshal(w.Body.Bytes(), &current)
	if w.Code != 200 || len(current.Items) != 1 || current.Items[0].Selector.EnvironmentSelector.Policy != "current-only" {
		t.Fatal(w.Code, w.Body.String())
	}
	for _, raw := range []string{"", "null", "{}", wires[0] + "{}", strings.Replace(wires[0], `"current-only"`, `"unknown"`, 1), strings.Replace(wires[0], `"scope":`, `"extra":true,"scope":`, 1), strings.Repeat("x", 1025)} {
		w = call("GET", "/activity-views?environmentSelector="+url.QueryEscape(raw), "")
		if w.Code != 400 {
			t.Fatal("bad filter accepted", raw, w.Code)
		}
	}
	for _, q := range []string{"environment=prod&environmentSelector=" + url.QueryEscape(wires[0]), "environment=&environmentSelector=" + url.QueryEscape(wires[0]), "environment=", "environmentSelector=" + url.QueryEscape(wires[0]) + "&environmentSelector=" + url.QueryEscape(wires[1])} {
		if w = call("GET", "/activity-views?"+q, ""); w.Code != 400 {
			t.Fatal("ambiguous/repeated filter accepted", q, w.Code)
		}
	}
}
