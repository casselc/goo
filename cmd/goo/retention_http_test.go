package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRetentionHTTPOptionalAndServerOwned(t *testing.T) {
	connection := retentionConnection{}
	mux := http.NewServeMux()
	registerRetention(mux, t.TempDir(), connection)
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
	connection = retentionConnection{URL: "postgres://xtdb@127.0.0.1:1/xtdb?sslmode=disable"}
	mux = http.NewServeMux()
	registerRetention(mux, t.TempDir(), connection)
	for _, body := range []string{`{"requestID":"` + strings.Repeat("a", 64) + `","tailRefs":[{"token":"server-only","key":"untrusted"}]}`, `{"captureID":"x","rows":[{"forged":true}]}`, `{"captureID":"x"} {}`, `{"captureID":"../escape","indexes":[0],"tags":["review"]}`, `{"captureID":"x","origin":"trusted-user"}`} {
		w = call("POST", "/retained-evidence", body)
		if w.Code != 400 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
}

func TestRetentionRecoveryDoesNotTreatUnavailableStoreAsAbsentCommit(t *testing.T) {
	connection := retentionConnection{URL: "postgres://xtdb@127.0.0.1:1/xtdb?sslmode=disable"}
	mux := http.NewServeMux()
	registerRetention(mux, t.TempDir(), connection)
	w := httptest.NewRecorder()
	body := `{"requestID":"` + strings.Repeat("a", 64) + `","resultToken":"` + strings.Repeat("b", 64) + `","indexes":[0],"tags":["review"],"note":"original draft"}`
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/retained-evidence", strings.NewReader(body)))
	if w.Code != 502 || !strings.Contains(w.Body.String(), "retry unchanged") || strings.Contains(w.Body.String(), "run the question") {
		t.Fatal(w.Code, w.Body.String())
	}
	for _, path := range []string{"/retained-evidence/" + strings.Repeat("a", 64) + "?awaitToken=x&awaitToken=y", "/retained-evidence/" + strings.Repeat("a", 64) + "?untrusted=x"} {
		w = httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 400 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
}
