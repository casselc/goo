package main

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func TestActivityViewsHTTPSourceFreeAndStrict(t *testing.T) {
	mux := http.NewServeMux()
	registerActivityViews(mux, t.TempDir())
	body := `{"version":1,"title":"Recent errors","selector":{"environment":"prod","signal":"logs"},"columns":["utc","summary"]}`
	var observations []apiObservation
	defer func() {
		if path := os.Getenv("GOO_ACTIVITY_VIEW_OBSERVATIONS"); path != "" {
			raw, e := json.Marshal(observations)
			if e != nil {
				t.Fatal(e)
			}
			if e = os.WriteFile(path, raw, 0600); e != nil {
				t.Fatal(e)
			}
		}
	}()
	call := func(method, path, raw string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(raw)))
		if w.Code < 400 {
			pattern := strings.Split(path, "?")[0]
			if strings.HasPrefix(pattern, "/activity-views/") {
				pattern = "/activity-views/{id}"
			}
			observations = append(observations, apiObservation{Method: method, Path: pattern, Status: w.Code, Request: json.RawMessage(raw), Response: json.RawMessage(w.Body.Bytes())})
		}
		return w
	}
	w := call("POST", "/activity-views", body)
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	var v struct {
		ID string `json:"id"`
	}
	json.Unmarshal(w.Body.Bytes(), &v)
	for _, path := range []string{"/activity-views/" + v.ID, "/activity-views?environment=prod"} {
		w = call("GET", path, "")
		if w.Code != 200 || w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	for _, raw := range []string{strings.Replace(body, `"columns":`, `"rows":[],"columns":`, 1), strings.Replace(body, `"environment":"prod"`, `"fromMs":0,"environment":"prod"`, 1), body + "{}"} {
		if w = call("POST", "/activity-views", raw); w.Code != 400 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	for _, path := range []string{"/activity-views?environment=a&environment=b", "/activity-views/" + v.ID + "?run=1"} {
		if w = call("GET", path, ""); w.Code != 400 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
}

func TestActivityBaselineHTTPStoredReferenceOnly(t *testing.T) {
	resultSnapshots := &exploration.ResultSnapshots{}
	var reads atomic.Int64
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reads.Add(1)
		raw, _ := io.ReadAll(r.Body)
		sql := string(raw)
		switch {
		case strings.Contains(sql, "SELECT DISTINCT"):
		case strings.Contains(sql, "GROUP BY"):
			fmt.Fprintln(w, `{"g0":"[0,1) ms","count":"1"}`)
		case strings.Contains(sql, "count()"):
			fmt.Fprintln(w, `{"count":"1"}`)
		default:
			fmt.Fprintln(w, `{"timestamp_ns":"2000000000","duration_ns":"1"}`)
		}
	}))
	defer source.Close()
	root := t.TempDir()
	mux := http.NewServeMux()
	registerCaptures(mux, exploration.Source{Endpoint: source.URL, Client: source.Client()}, root, resultSnapshots)
	registerActivityViews(mux, root)
	call := func(method, path, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(body)))
		return w
	}
	makeCapture := func(env, field string) string {
		raw := fmt.Sprintf(`{"title":"Duration baseline","question":{"kind":"explore","explore":{"signal":"spans","environment":%q,"fromMs":1000,"toMs":61000,"limit":100,"offset":0,"groupBy":[{"kind":"column","key":%q}]}}}`, env, field)
		w := call("POST", "/captures", raw)
		if w.Code != 201 {
			t.Fatal(w.Code, w.Body.String())
		}
		var page exploration.CapturePage
		if e := json.Unmarshal(w.Body.Bytes(), &page); e != nil {
			t.Fatal(e)
		}
		return page.Receipt.ID
	}
	baseline := makeCapture("prod", "duration_bucket")
	other := makeCapture("other", "duration_bucket")
	wrong := makeCapture("prod", "service")
	before := reads.Load()
	source.Close()
	body := func(id, signal string) string {
		return fmt.Sprintf(`{"version":1,"title":"With baseline","selector":{"signal":%q,"environment":"prod"},"columns":["utc","summary"],"baselineCapture":%q}`, signal, id)
	}
	w := call("POST", "/activity-views", body(baseline, "spans"))
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	var v workbench.ActivityView
	json.Unmarshal(w.Body.Bytes(), &v)
	if v.BaselineCapture != baseline {
		t.Fatal("lost baseline")
	}
	for _, id := range []string{other, wrong, strings.Repeat("f", 64), "bad"} {
		if w = call("POST", "/activity-views", body(id, "spans")); w.Code != 400 {
			t.Fatal("accepted unsupported reference", w.Code, w.Body.String())
		}
	}
	if w = call("POST", "/activity-views", body(baseline, "logs")); w.Code != 400 {
		t.Fatal("logs accepted baseline")
	}
	// Removing evidence does not corrupt the immutable configuration or its catalog.
	if e := os.Remove(filepath.Join(root, "exploration", baseline+".json")); e != nil {
		t.Fatal(e)
	}
	for _, path := range []string{"/activity-views/" + v.ID, "/activity-views?environment=prod"} {
		if w = call("GET", path, ""); w.Code != 200 {
			t.Fatal("dangling evidence hid configuration", w.Code, w.Body.String())
		}
	}
	if reads.Load() != before {
		t.Fatal("configuration operation read telemetry")
	}
}
