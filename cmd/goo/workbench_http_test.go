package main

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWorkbenchHTTPRejectsUntrustedReceiptsAndUnavailableLinks(t *testing.T) {
	resultSnapshots := &exploration.ResultSnapshots{}
	mux := http.NewServeMux()
	registerWorkbench(mux, t.TempDir(), retentionConnection{}, resultSnapshots, nil)
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

func TestAnnotationHTTPStrictMetadataContract(t *testing.T) {
	resultSnapshots := &exploration.ResultSnapshots{}
	mux := http.NewServeMux()
	registerWorkbench(mux, t.TempDir(), retentionConnection{}, resultSnapshots, nil)
	call := func(method, path, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(body)))
		return w
	}
	body := `{"target":{"kind":"service","source":"clickstack","environment":"prod","service":"checkout","namespace":{"mode":"missing"}},"tags":["incident:latency"],"note":"Review","state":"active","applies":{"fromMs":100,"toMs":200}}`
	w := call("POST", "/annotations", body)
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	var r struct{ ID, RecordedAt, Origin string }
	json.Unmarshal(w.Body.Bytes(), &r)
	if r.RecordedAt == "" || r.Origin != "local-user-unverified" {
		t.Fatal(r)
	}
	for _, tc := range []struct {
		method, path, body string
		code               int
	}{
		{"GET", "/annotations/" + r.ID, "", 200}, {"GET", "/annotations?tag=incident:latency", "", 200},
		{"GET", "/annotations?history=0", "", 400}, {"GET", "/annotations?tag=a&tag=b", "", 400}, {"GET", "/annotations?query=live", "", 400},
		{"GET", "/annotations/" + r.ID + "?history=1", "", 400}, {"GET", "/annotations?target=null", "", 400},
		{"POST", "/annotations", strings.Replace(body, `"note":`, `"recordedAt":"forged","note":`, 1), 400},
		{"POST", "/annotations", strings.Replace(body, `"note":`, `"origin":"trusted-human","note":`, 1), 400},
		{"POST", "/annotations", strings.Replace(body, `"namespace":{"mode":"missing"}`, `"namespace":null`, 1), 400},
		{"POST", "/annotations", body + body, 400},
	} {
		w = call(tc.method, tc.path, tc.body)
		if w.Code != tc.code {
			t.Fatalf("%s %s: %d %s", tc.method, tc.path, w.Code, w.Body.String())
		}
		if w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("cache policy")
		}
	}
}
