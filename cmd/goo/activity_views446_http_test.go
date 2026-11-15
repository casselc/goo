package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPairedActivityViewHTTPStrictConfiguration(t *testing.T) {
	mux := http.NewServeMux()
	registerActivityViews(mux, t.TempDir())
	const valid = `{"version":1,"title":"Paired service","selector":{"signal":"both","environment":"prod","where":{"field":{"kind":"resource","key":"team"},"op":"eq","value":"api"}},"lanes":{"logs":{"columns":["utc","summary"],"fieldColumns":[{"kind":"column","key":"body"}]},"spans":{"columns":["summary","service","utc"],"fieldColumns":[{"kind":"column","key":"name"}]}}}`
	call := func(method, path, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(body)))
		return w
	}
	response := call("POST", "/activity-views", valid)
	if response.Code != 201 {
		t.Fatal(response.Code, response.Body.String())
	}
	var saved struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &saved); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/activity-views/" + saved.ID, "/activity-views?environment=prod"} {
		w := call("GET", path, "")
		if w.Code != 200 || !strings.Contains(w.Body.String(), saved.ID) || w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal(path, w.Code, w.Body.String())
		}
	}
	for name, body := range map[string]string{
		"empty flat columns":  strings.Replace(valid, `"lanes":`, `"columns":[],"lanes":`, 1),
		"null flat fields":    strings.Replace(valid, `"lanes":`, `"fieldColumns":null,"lanes":`, 1),
		"empty baseline":      strings.Replace(valid, `"lanes":`, `"baselineCapture":"","lanes":`, 1),
		"null baseline":       strings.Replace(valid, `"lanes":`, `"baselineCapture":null,"lanes":`, 1),
		"single composition":  strings.Replace(valid, `"signal":"both"`, `"signal":"logs"`, 1),
		"extra lane":          strings.Replace(valid, `"lanes":{`, `"lanes":{"metrics":{"columns":["utc","summary"]},`, 1),
		"missing spans":       strings.Replace(valid, `,"spans":{"columns":["summary","service","utc"],"fieldColumns":[{"kind":"column","key":"name"}]}`, ``, 1),
		"null lane fields":    strings.Replace(valid, `"fieldColumns":[{"kind":"column","key":"body"}]`, `"fieldColumns":null`, 1),
		"null lane columns":   strings.Replace(valid, `"columns":["utc","summary"]`, `"columns":null`, 1),
		"runtime owner":       strings.Replace(valid, `"columns":["utc","summary"]`, `"connection":"live","columns":["utc","summary"]`, 1),
		"fixed window":        strings.Replace(valid, `"environment":"prod"`, `"environment":"prod","fromMs":0`, 1),
		"log-only predicate":  strings.Replace(valid, `"kind":"resource","key":"team"`, `"kind":"column","key":"body"`, 1),
		"wrong-case view":     strings.Replace(valid, `"title":`, `"Title":`, 1),
		"wrong-case selector": strings.Replace(valid, `"signal":`, `"Signal":`, 1),
		"wrong-case lanes":    strings.Replace(valid, `"logs":`, `"Logs":`, 1),
		"wrong-case layout":   strings.Replace(valid, `"columns":`, `"Columns":`, 1),
		"null signal":         strings.Replace(valid, `"signal":"both"`, `"signal":null`, 1),
		"null where":          strings.Replace(valid, `"where":{"field":{"kind":"resource","key":"team"},"op":"eq","value":"api"}`, `"where":null`, 1),
		"trailing":            valid + `{}`,
	} {
		t.Run(name, func(t *testing.T) {
			w := call("POST", "/activity-views", body)
			if w.Code != 400 {
				t.Fatal("accepted invalid paired configuration", w.Code, w.Body.String())
			}
		})
	}
	const single = `{"version":1,"title":"Single service","selector":{"signal":"logs","environment":"prod"},"columns":["utc","summary"]}`
	for _, raw := range []string{
		strings.Replace(single, `"signal":"logs"`, `"signal":null`, 1),
		strings.Replace(single, `"environment":"prod"`, `"environment":null`, 1),
		strings.Replace(single, `"columns":["utc","summary"]`, `"columns":null`, 1),
		strings.Replace(single, `"columns":["utc","summary"]`, `"columns":["utc","summary"],"fieldColumns":null`, 1),
		strings.Replace(single, `"environment":"prod"`, `"environment":"prod","where":null`, 1),
	} {
		if w := call("POST", "/activity-views", raw); w.Code != 400 {
			t.Fatal("single null admitted", w.Code, w.Body.String())
		}
	}

}
