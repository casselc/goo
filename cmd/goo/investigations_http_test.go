package main

import (
	"bytes"
	"encoding/json"
	"github.com/casselc/goo/annotations"
	"github.com/casselc/goo/investigations"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestCaptureHTTPTrustedSourcesAndOfflineReopen(t *testing.T) {
	var reads atomic.Int64
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reads.Add(1)
		if reads.Load() > 1 {
			http.Error(w, "optional source unavailable", 503)
			return
		}
	}))
	defer source.Close()
	t.Setenv("CLICKHOUSE_URL", source.URL)
	plan, e := annotations.LoadOperationPackage("../../config/operations/checkout.yaml")
	if e != nil {
		t.Fatal(e)
	}
	mux := http.NewServeMux()
	registerInvestigations(mux, map[string]*annotations.OperationPlan{"checkout": plan}, source.Client(), t.TempDir())
	call := func(method, path string, body any) *httptest.ResponseRecorder {
		b, _ := json.Marshal(body)
		r := httptest.NewRequest(method, path, bytes.NewReader(b))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		return w
	}
	q := investigations.Question{Package: "checkout", Revision: plan.Revision, FromMS: 60000, ToMS: 120000}
	w := call("POST", "/investigations", map[string]any{"question": q, "selection": map[string]any{"filter": map[string]string{"outcome": "error"}, "ratio": "ratio-key", "comparison": "member-key"}})
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	var first investigations.Capture
	if e = json.Unmarshal(w.Body.Bytes(), &first); e != nil {
		t.Fatal(e)
	}
	if len(first.Evidence.Errors) != 2 {
		t.Fatal("optional errors not captured", first.Evidence.Errors)
	}
	before := reads.Load()
	source.Close()
	w = call("GET", "/investigations/"+first.ID, nil)
	if w.Code != 200 || reads.Load() != before {
		t.Fatal("reopen read source", w.Code)
	}
	var got investigations.Capture
	json.Unmarshal(w.Body.Bytes(), &got)
	if got.ID != first.ID || string(got.Selection) != string(first.Selection) {
		t.Fatal("lost receipt selection")
	}
	q.Revision = "changed"
	w = call("POST", "/investigations", map[string]any{"question": q})
	if w.Code != 409 || reads.Load() != before {
		t.Fatal("changed revision queried source", w.Code)
	}
	q = first.Question
	q.ToMS++
	w = call("POST", "/investigations", map[string]any{"parent": first.ID, "question": q})
	if w.Code != 409 {
		t.Fatal("changed question accepted", w.Code)
	}
	w = call("POST", "/investigations", map[string]any{"question": first.Question, "selection": map[string]any{"filter": map[string]string{"fromNS": "broken"}}})
	if w.Code != 400 || reads.Load() != before {
		t.Fatal("invalid selection reached source", w.Code)
	}
	w = call("POST", "/investigations", map[string]any{"parent": first.ID})
	if w.Code != 502 {
		t.Fatal("source failure invented refresh", w.Code)
	}
	w = call("GET", "/investigations/"+first.ID, nil)
	if w.Code != 200 {
		t.Fatal("failed refresh lost original")
	}
}
