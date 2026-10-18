package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecipeHTTPIsSourceFreeAndStrict(t *testing.T) {
	mux := http.NewServeMux()
	registerRecipes(mux)
	for _, tc := range []struct {
		method, body string
		code         int
	}{
		{"GET", "", 200},
		{"POST", `{"id":"errors","question":{"signal":"spans","environment":"local","fromMs":0,"toMs":60000,"limit":100,"offset":0}}`, 200},
		{"POST", `{"id":"errors","question":{"signal":"logs","environment":"local","fromMs":0,"toMs":60000,"limit":100,"offset":0}}`, 400},
		{"POST", `{"id":"errors","question":{},"sql":"SELECT 1"}`, 400},
		{"POST", `{} {}`, 400},
	} {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(tc.method, "/recipes", strings.NewReader(tc.body)))
		if w.Code != tc.code {
			t.Fatalf("%s: %d %s", tc.body, w.Code, w.Body.String())
		}
	}
}
