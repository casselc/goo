package main

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPopulationConfigurationHTTP(t *testing.T) {
	resultSnapshots := &exploration.ResultSnapshots{}
	mux := http.NewServeMux()
	registerWorkbench(mux, t.TempDir(), retentionConnection{}, resultSnapshots, nil)
	call := func(method, url, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, url, strings.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		return w
	}
	q := `{"signal":"spans","environment":"prod","fromMs":1000,"toMs":2000,"limit":100,"offset":0}`
	p := call("POST", "/population/plan", `{"question":`+q+`}`)
	if p.Code != 200 {
		t.Fatal(p.Code, p.Body.String())
	}
	if w := call("POST", "/population/plan", `{"question":`+q+`,"dropFilters":true}`); w.Code != 400 {
		t.Fatal(w.Code)
	}
	view := `{"version":1,"title":"Generic population","queries":[` + q + `],"layout":"records","columns":["utc","summary"]}`
	created := call("POST", "/exploration-views", view)
	if created.Code != 201 {
		t.Fatal(created.Code, created.Body.String())
	}
	var saved workbench.ExplorationView
	json.Unmarshal(created.Body.Bytes(), &saved)
	if !workbench.ID.MatchString(saved.ID) {
		t.Fatal(saved)
	}
	if w := call("GET", "/exploration-views/"+saved.ID, ""); w.Code != 200 {
		t.Fatal(w.Code)
	}
	if w := call("GET", "/exploration-views?cursor=bad", ""); w.Code != 400 {
		t.Fatal(w.Code)
	}
	bad := strings.Replace(view, `["utc","summary"]`, `["utc"]`, 1)
	if w := call("POST", "/exploration-views", bad); w.Code != 400 {
		t.Fatal(w.Code)
	}
	// Authoring persists only the checked declaration; unknown runtime data is refused.
	authored := strings.Replace(view, `"limit":100`, `"groupBy":[{"kind":"column","key":"service"}],"limit":100`, 1)
	authored = strings.TrimSuffix(authored, "}") + `,"displays":{"spans":{"version":"goo.view/1","scope":{"kind":"record-query","signal":"spans"},"kind":{"name":"bar","version":1},"display":{"category":"category","value":"count","requiresComplete":true}}}}`
	if w := call("POST", "/exploration-views", authored); w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	for _, replacement := range []string{`"value":"p95"`, `"value":"count","rows":[]`, `"value":"count","complete":true`} {
		bad := strings.Replace(authored, `"value":"count"`, replacement, 1)
		if w := call("POST", "/exploration-views", bad); w.Code != 400 {
			t.Fatal("invalid declaration admitted", w.Code, w.Body.String())
		}
	}

}
