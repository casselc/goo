package main

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/capturehttp"
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
	var resources atomic.Value
	resources.Store(`{"deployment.environment.name":"prod"}`)
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
			fmt.Fprintf(w, "{\"timestamp_ns\":\"2000000000\",\"duration_ns\":\"1\",\"SpanName\":\"baseline\",\"ResourceAttributes\":%s}\n", resources.Load().(string))
		}
	}))
	defer source.Close()
	root := t.TempDir()
	mux := http.NewServeMux()
	capturehttp.Register(mux, exploration.Source{Endpoint: source.URL, Client: source.Client()}, root, resultSnapshots)
	registerActivityViews(mux, root)
	call := func(method, path, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(body)))
		return w
	}
	makeCapture := func(env, field string, selector ...string) string {
		resource := map[string]string{"deployment.environment.name": env}
		if len(selector) == 1 {
			var selection struct {
				Scope struct {
					Mode  string `json:"mode"`
					Value string `json:"value"`
				} `json:"scope"`
			}
			if e := json.Unmarshal([]byte(selector[0]), &selection); e != nil {
				t.Fatal(e)
			}
			if selection.Scope.Mode == "unreported" {
				resource = map[string]string{}
			} else {
				resource["deployment.environment.name"] = selection.Scope.Value
			}
		}
		encoded, _ := json.Marshal(resource)
		resources.Store(string(encoded))
		raw := fmt.Sprintf(`{"title":"Duration baseline","question":{"kind":"explore","explore":{"signal":"spans","environment":%q,"fromMs":1000,"toMs":61000,"limit":100,"offset":0,"groupBy":[{"kind":"column","key":%q}]}}}`, env, field)
		if len(selector) == 1 {
			raw = strings.Replace(raw, `"environment":`+fmt.Sprintf("%q", env), `"environment":"","environmentSelector":`+selector[0], 1)
		}
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
	wires := []string{
		`{"version":"goo/environment-selector/v1","policy":"current-only","scope":{"mode":"named","value":"prod"}}`,
		`{"version":"goo/environment-selector/v1","policy":"effective","scope":{"mode":"named","value":"prod"}}`,
		`{"version":"goo/environment-selector/v1","policy":"effective","scope":{"mode":"empty"}}`,
		`{"version":"goo/environment-selector/v1","policy":"effective","scope":{"mode":"unreported"}}`,
	}
	typedCaptures := make([]string, len(wires))
	for i, wire := range wires {
		typedCaptures[i] = makeCapture("", "duration_bucket", wire)
	}
	before := reads.Load()
	source.Close()
	body := func(id, signal string, selector ...string) string {
		raw := fmt.Sprintf(`{"version":1,"title":"With baseline","selector":{"signal":%q,"environment":"prod"},"columns":["utc","summary"],"baselineCapture":%q}`, signal, id)
		if len(selector) == 1 {
			raw = strings.Replace(raw, `"environment":"prod"`, `"environment":"","environmentSelector":`+selector[0], 1)
		}
		return raw
	}
	for i, wire := range wires {
		if got := call("POST", "/activity-views", body(typedCaptures[i], "spans", wire)); got.Code != 201 {
			t.Fatal("matching typed baseline refused", got.Code, got.Body.String())
		}
		if got := call("POST", "/activity-views", body(typedCaptures[(i+1)%len(wires)], "spans", wire)); got.Code != 400 {
			t.Fatal("cross-policy/scope baseline accepted", got.Code, got.Body.String())
		}
	}
	if got := call("POST", "/activity-views", body(baseline, "spans", wires[0])); got.Code != 201 {
		t.Fatal("semantic current-only equivalence refused", got.Code, got.Body.String())
	}
	if got := call("POST", "/activity-views", body(typedCaptures[0], "spans")); got.Code != 201 {
		t.Fatal("typed current-only baseline refused by legacy current", got.Code, got.Body.String())
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
