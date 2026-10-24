package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNumericHistogramSelectionHTTP(t *testing.T) {
	mux := http.NewServeMux()
	registerWorkbench(mux, t.TempDir(), retentionConnection{})
	parent := `{"signal":"spans","environment":"prod","fromMs":1000,"toMs":2000,"limit":100,"numeric":{"field":{"kind":"attribute","key":"work.items"},"type":"uint64","unit":"items","parse":"canonical-decimal"},"numericHistogram":{"binning":"powers-of-two"}}`
	for _, tc := range []struct {
		body string
		code int
	}{
		{`{"parent":` + parent + `,"fromBin":64,"toBin":65}`, 200},
		{`{"parent":` + parent + `,"fromBin":0,"toBin":65}`, 200},
		{`{"parent":` + parent + `,"toBin":65}`, 400},
		{`{"parent":` + parent + `,"fromBin":0}`, 400},
		{`{"parent":` + parent + `,"fromBin":64,"toBin":66}`, 400},
		{`{"parent":` + parent + `,"fromBin":1,"toBin":1}`, 400},
		{`{"parent":` + parent + `,"fromBin":0,"toBin":1,"where":"ignored"}`, 400},
		{`{"parent":` + parent + `,"fromBin":0,"toBin":1} {}`, 400},
	} {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/numeric-histograms/selection", strings.NewReader(tc.body)))
		if w.Code != tc.code {
			t.Fatalf("HTTP %d want %d: %s", w.Code, tc.code, w.Body.String())
		}
		if tc.code == 200 {
			var out map[string]json.RawMessage
			if e := json.Unmarshal(w.Body.Bytes(), &out); e != nil {
				t.Fatal(e)
			}
			for _, key := range []string{"parent", "selected", "rest", "comparison"} {
				if len(out[key]) == 0 {
					t.Fatalf("missing %s", key)
				}
			}
		}
	}
}
