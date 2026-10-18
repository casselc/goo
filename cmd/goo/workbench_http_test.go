package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWorkbenchHTTPRejectsUntrustedReceiptsAndUnavailableLinks(t *testing.T) {
	mux := http.NewServeMux()
	registerWorkbench(mux, t.TempDir())
	call := func(method, path, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(body)))
		return w
	}
	body := `{"environment":"prod","title":"Deploy","revision":"v2","occurredAtMs":100}`
	w := call("POST", "/deployments", body)
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	var m map[string]any
	json.Unmarshal(w.Body.Bytes(), &m)
	id := m["id"].(string)
	for _, tc := range []struct {
		method, path, body string
		code               int
	}{{"POST", "/deployments", strings.Replace(body, `"title":`, `"recordedAt":"yesterday","title":`, 1), 400}, {"GET", "/deployments?environment=prod&fromMs=100&toMs=200&fromMs=0", "", 400}, {"GET", "/deployments?environment=prod&fromMs=100&toMs=200", "", 200}, {"POST", "/documents", `{"document":{"title":"Notes","cards":[{"kind":"capture","reference":"` + strings.Repeat("a", 64) + `"}]}}`, 400}, {"POST", "/documents", `{"document":{"title":"Notes","cards":[{"kind":"deployment","reference":"` + id + `"}]}}`, 201}} {
		w := call(tc.method, tc.path, tc.body)
		if w.Code != tc.code {
			t.Fatal(tc.path, w.Code, w.Body.String())
		}
		if w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("missing local response cache policy")
		}
	}
}
