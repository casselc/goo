package main

import (
	"encoding/json"
	"github.com/casselc/goo/workbench"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPairedActivityLanePredicatesHTTP450(t *testing.T) {
	mux := http.NewServeMux()
	registerActivityViews(mux, t.TempDir())
	const valid = `{"version":1,"title":"Independent lanes","selector":{"signal":"both","environment":"prod"},"lanes":{"logs":{"columns":["utc","summary"],"where":{"field":{"kind":"column","key":"body"},"op":"eq","value":""}},"spans":{"columns":["utc","summary"],"where":{"field":{"kind":"column","key":"duration_ns"},"op":"gte","value":"1000000"}}}}`
	call := func(method, path, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(body)))
		return w
	}
	response := call("POST", "/activity-views", valid)
	if response.Code != 201 {
		t.Fatal(response.Code, response.Body.String())
	}
	var saved workbench.ActivityView
	if err := json.Unmarshal(response.Body.Bytes(), &saved); err != nil {
		t.Fatal(err)
	}
	read := call("GET", "/activity-views/"+saved.ID, "")
	if read.Code != 200 {
		t.Fatal(read.Code, read.Body.String())
	}
	var held workbench.ActivityView
	if err := json.Unmarshal(read.Body.Bytes(), &held); err != nil {
		t.Fatal(err)
	}
	q, err := held.Queries()
	if err != nil || len(q) != 2 || q[0].Where.Field.Key != "body" || q[0].Where.Value != "" || q[1].Where.Field.Key != "duration_ns" {
		t.Fatal("lost lane predicates", q, err)
	}
	for _, bad := range []string{strings.Replace(valid, `"key":"body"`, `"key":"name"`, 1), strings.Replace(valid, `"where":{"field":{"kind":"column","key":"body"},"op":"eq","value":""}`, `"where":null`, 1)} {
		w := call("POST", "/activity-views", bad)
		if w.Code != 400 {
			t.Fatal("malformed configuration published", w.Code, w.Body.String())
		}
	}
}
