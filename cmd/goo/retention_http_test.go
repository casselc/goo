package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRetentionHTTPOptionalAndServerOwned(t *testing.T) {
	t.Setenv("GOO_XTDB_URL", "")
	mux := http.NewServeMux()
	registerRetention(mux, t.TempDir())
	call := func(method, path, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(body)))
		return w
	}
	w := call("GET", "/retained-evidence/status", "")
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"configured":false`) {
		t.Fatal(w.Code, w.Body.String())
	}
	if w = call("POST", "/retained-evidence", `{}`); w.Code != 503 {
		t.Fatal(w.Code)
	}
	t.Setenv("GOO_XTDB_URL", "postgres://xtdb@127.0.0.1:1/xtdb?sslmode=disable")
	mux = http.NewServeMux()
	registerRetention(mux, t.TempDir())
	for _, body := range []string{`{"captureID":"x","rows":[{"forged":true}]}`, `{"captureID":"x"} {}`, `{"captureID":"../escape","indexes":[0],"tags":["review"]}`, `{"captureID":"x","origin":"trusted-user"}`} {
		w = call("POST", "/retained-evidence", body)
		if w.Code != 400 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
}
