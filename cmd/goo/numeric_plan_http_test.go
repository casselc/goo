package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casselc/goo/exploration"
)

func TestNumericPlanHTTPSourceFree(t *testing.T) {
	mux := http.NewServeMux()
	registerTelemetryPackages(mux, t.TempDir())
	input := map[string]any{"question": exploration.Query{Signal: "logs", Environment: "ordinary", FromMS: 1000, ToMS: 61000, Limit: 100}, "binding": exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "db.response.returned_rows"}, Type: "uint64", Unit: "{row}", Parse: "canonical-decimal"}}
	data, _ := json.Marshal(input)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/telemetry-packages/numeric-plan", bytes.NewReader(data)))
	var plan exploration.NumericPlan
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &plan) != nil || plan.Executable || len(plan.Diagnostics) != 1 || plan.Scope.Signal != "logs" {
		t.Fatal(w.Code, w.Body.String())
	}
	for _, data := range []string{`{}`, `{"sql":"SELECT 1"}`, string(data) + ` {}`} {
		w = httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/telemetry-packages/numeric-plan", bytes.NewBufferString(data)))
		if w.Code != 400 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
}
