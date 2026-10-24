package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Capability diagnostics are deliberately different from malformed transport:
// an unsupported scale is a source-free plan, not an executable query.
func TestDecimalPlanHTTPScaleBoundary(t *testing.T) {
	mux := http.NewServeMux()
	registerTelemetryPackages(mux, t.TempDir())
	for _, tc := range []struct {
		name, typ  string
		scale      any
		status     int
		executable bool
	}{
		{"minimum", "decimal64", 1, 200, true},
		{"maximum", "decimal64", 9, 200, true},
		{"missing", "decimal64", nil, 200, false},
		{"zero", "decimal64", 0, 200, false},
		{"negative", "decimal64", -1, 200, false},
		{"too large", "decimal64", 10, 200, false},
		{"integer scaled", "int64", 2, 200, false},
		{"fractional transport", "decimal64", 2.5, 400, false},
		{"string transport", "decimal64", "2", 400, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			binding := map[string]any{"field": map[string]string{"kind": "attribute", "key": "example.cost"}, "type": tc.typ, "unit": "USD", "parse": "canonical-decimal"}
			if tc.scale != nil {
				binding["scale"] = tc.scale
			}
			body, err := json.Marshal(map[string]any{"question": map[string]any{"signal": "logs", "environment": "decimal-plan", "fromMs": 1000, "toMs": 61000, "limit": 100}, "binding": binding})
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, httptest.NewRequest("POST", "/telemetry-packages/numeric-plan", bytes.NewReader(body)))
			if w.Code != tc.status {
				t.Fatalf("HTTP %d, want %d: %s", w.Code, tc.status, w.Body.String())
			}
			if tc.status == 200 {
				var out struct {
					Executable  bool              `json:"executable"`
					Diagnostics []json.RawMessage `json:"diagnostics"`
				}
				if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
					t.Fatal(err)
				}
				if out.Executable != tc.executable || (len(out.Diagnostics) == 0) != tc.executable {
					t.Fatalf("unexpected plan: %s", w.Body.String())
				}
			}
		})
	}
}
