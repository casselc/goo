package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/capturehttp"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestPopulationExplanationHTTPObservedPromotionAndFilters(t *testing.T) {
	resultSnapshots := &exploration.ResultSnapshots{}
	sourceReads := 0
	badGroup := false
	client := &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		sourceReads++
		deadline, ok := r.Context().Deadline()
		if !ok || time.Until(deadline) > 3100*time.Millisecond {
			t.Fatal("missing bounded child source deadline")
		}
		sql, _ := io.ReadAll(r.Body)
		text := string(sql)
		if strings.Contains(text, "OFFSET") || strings.Contains(text, "arrayJoin") || strings.Contains(text, "numeric_histogram") {
			t.Fatal("explanation read inspection/measures instead of count/groups", text)
		}
		body := `{"count":"4"}`
		if strings.Contains(text, "GROUP BY") {
			side := "baseline-only"
			for _, v := range r.URL.Query() {
				for _, value := range v {
					if value == "selected" {
						side = "selected-only"
					}
				}
			}
			b, _ := json.Marshal(map[string]string{"g0": side, "count": "4"})
			body = string(b)
			if badGroup {
				body = `{"count":"4"}`
			}
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}
	source := exploration.Source{Endpoint: "http://explanation-source.invalid", Client: client}
	mux := http.NewServeMux()
	registerExploration(mux, source, resultSnapshots, nil)
	capturehttp.Register(mux, source, t.TempDir(), resultSnapshots)
	q := exploration.Query{Signal: "spans", Environment: "test", FromMS: 1000, ToMS: 2000, Limit: 100, Where: &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "baseline"}}
	selected := q
	selected.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "selected"}
	req := exploration.ExplanationRequest{Version: exploration.ExplanationVersion, Baseline: q, Selection: selected, Fields: []exploration.Field{{Kind: "column", Key: "service"}}}
	call := func(path string, body any) *httptest.ResponseRecorder {
		raw, e := json.Marshal(body)
		if e != nil {
			t.Fatal(e)
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", path, bytes.NewReader(raw)))
		return w
	}
	if w := call("/population/explain/plan", req); w.Code != 200 || sourceReads != 0 {
		t.Fatal("preflight read source", w.Code, w.Body.String())
	}
	w := call("/population/explain", req)
	if w.Code != 200 || sourceReads != 4 {
		t.Fatal(w.Code, w.Body.String(), sourceReads)
	}
	var observed exploration.ExplanationResult
	if e := json.Unmarshal(w.Body.Bytes(), &observed); e != nil {
		t.Fatal(e)
	}
	if observed.ResultToken == "" || observed.State != "complete" || len(observed.Ranked) != 2 {
		t.Fatal("observed result absent", w.Body.String())
	}
	for _, value := range []any{nil, ""} {
		w = call("/population/explain/filter", map[string]any{"request": req, "field": req.Fields[0], "value": value, "side": "selection", "mode": "exclude"})
		if w.Code != 200 || sourceReads != 4 {
			t.Fatal("filter planning read source", w.Code, w.Body.String())
		}
	}
	if w = call("/population/explain/filter", map[string]any{"request": req, "field": req.Fields[0], "side": "selection", "mode": "include"}); w.Code != 400 {
		t.Fatal("missing/null intent conflated")
	}
	promotion := map[string]any{"title": "Exact observed difference", "resultToken": observed.ResultToken}
	w = call("/captures", promotion)
	if w.Code != 201 || sourceReads != 4 {
		t.Fatal("save reread telemetry", w.Code, w.Body.String(), sourceReads)
	}
	var saved exploration.CapturePage
	if e := json.Unmarshal(w.Body.Bytes(), &saved); e != nil {
		t.Fatal(e)
	}
	observed.ResultToken = ""
	if saved.Explanation == nil || !reflect.DeepEqual(*saved.Explanation, observed) {
		t.Fatal("save changed observed data")
	}
	retry := call("/captures", promotion)
	if retry.Code != 201 || retry.Body.String() != w.Body.String() || sourceReads != 4 {
		t.Fatal("promotion retry changed evidence", retry.Body.String())
	}
	get := httptest.NewRecorder()
	mux.ServeHTTP(get, httptest.NewRequest("GET", "/captures/"+saved.Receipt.ID, nil))
	if get.Code != 200 || get.Body.String() != w.Body.String() || sourceReads != 4 {
		t.Fatal("reopen changed evidence/read source")
	}
	if invalid := call("/captures", map[string]any{"title": "No reread", "question": map[string]any{"kind": "explanation", "explanation": req}}); invalid.Code != 400 || sourceReads != 4 {
		t.Fatal("explanation capture silently reran")
	}
	if invalid := call("/captures", map[string]any{"title": "Expired", "resultToken": strings.Repeat("0", 64)}); invalid.Code != 409 || sourceReads != 4 {
		t.Fatal("missing promotion silently reran")
	}
	regular := &exploration.Result{Question: q, Count: "0", State: "no_data", CapturedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	token, e := resultSnapshots.Freeze(regular)
	if e != nil {
		t.Fatal(e)
	}
	if invalid := call("/captures", map[string]any{"title": "Wrong token type", "resultToken": token}); invalid.Code != 400 || sourceReads != 4 {
		t.Fatal("record page token promoted as explanation")
	}
	badGroup = true
	w = call("/population/explain", req)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	var failed exploration.ExplanationResult
	json.Unmarshal(w.Body.Bytes(), &failed)
	if failed.State != "unavailable" || len(failed.Ranked) != 0 || len(failed.Fields[0].Rows) != 0 {
		t.Fatal("malformed group became a missing-key lead", w.Body.String())
	}
}

func TestPopulationExplanationAdmissionCancelBeforeRead(t *testing.T) {
	resultSnapshots := &exploration.ResultSnapshots{}
	mux := http.NewServeMux()
	registerExploration(mux, exploration.Source{Client: &http.Client{Transport: nativeDependencyTransport(func(*http.Request) (*http.Response, error) {
		t.Fatal("cancelled request reached source")
		return nil, context.Canceled
	})}}, resultSnapshots, nil)
	handler := newQueryBudget(mux, 1)
	id := "11111111-1111-4111-8111-111111111111"
	cancel := httptest.NewRecorder()
	handler.ServeHTTP(cancel, httptest.NewRequest("POST", "/queries/"+id+"/cancel", strings.NewReader("{}")))
	if cancel.Code != 202 {
		t.Fatal(cancel.Code)
	}
	request := httptest.NewRequest("POST", "/population/explain", strings.NewReader("{}"))
	request.Header.Set("X-Goo-Query-ID", id)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, request)
	if w.Code != 409 {
		t.Fatal("explanation bypassed cancellation admission", w.Code, w.Body.String())
	}
}
