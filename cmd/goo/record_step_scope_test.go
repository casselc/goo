package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestAuthoredRecordStepSelectorHTTPRoundTrip(t *testing.T) {
	mux := http.NewServeMux()
	registerWorkbench(mux, t.TempDir(), retentionConnection{}, nil, nil)
	call := func(method, path, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(body)))
		return w
	}
	for _, scope := range []string{`{"mode":"named","value":"prod"}`, `{"mode":"empty"}`, `{"mode":"unreported"}`} {
		selector := `{"version":"goo/environment-selector/v1","policy":"effective","scope":` + scope + `}`
		question := `{"signal":"logs","environment":"","environmentSelector":` + selector + `,"fromMs":1000,"toMs":2000,"limit":10,"offset":0,"where":{"field":{"kind":"attribute","key":"cohort"},"op":"eq","value":"canary"}}`
		step := `{"version":1,"kind":"record","scope":{"environment":"","environmentSelector":` + selector + `,"fromMs":1000,"toMs":2000},"view":"Records","record":` + question + `}`
		body := `{"document":{"title":"Typed fixed question","cards":[{"kind":"question","step":` + step + `}]}}`
		created := call("POST", "/documents", body)
		if created.Code != 201 {
			t.Fatal(created.Code, created.Body.String())
		}
		var doc struct {
			ID    string `json:"id"`
			Cards []struct {
				Step json.RawMessage `json:"step"`
			} `json:"cards"`
		}
		if err := json.Unmarshal(created.Body.Bytes(), &doc); err != nil {
			t.Fatal(err)
		}
		var expected, actual any
		json.Unmarshal([]byte(step), &expected)
		json.Unmarshal(doc.Cards[0].Step, &actual)
		if !reflect.DeepEqual(expected, actual) {
			t.Fatal("saved scope or filters changed", string(doc.Cards[0].Step))
		}
		reopened := call("GET", "/documents/"+doc.ID, "")
		if reopened.Code != 200 || reopened.Body.String() != created.Body.String() {
			t.Fatal("immutable reopen differs", reopened.Code)
		}
		bad := strings.Replace(body, `"scope":{"environment":"","environmentSelector":`, `"scope":{"environment":"foreign","environmentSelector":`, 1)
		if w := call("POST", "/documents", bad); w.Code != 400 {
			t.Fatal("ambiguous scope published", w.Code, w.Body.String())
		}
	}
}
