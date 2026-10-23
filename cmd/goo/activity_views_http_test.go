package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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
