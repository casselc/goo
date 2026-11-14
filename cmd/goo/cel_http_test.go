package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/querycel"
)

type celRejectTransport struct{ calls int }

func (transport *celRejectTransport) RoundTrip(*http.Request) (*http.Response, error) {
	transport.calls++
	return nil, fmt.Errorf("source access forbidden during planning")
}

func TestCELPlanningGenericRuntimeNoSourceOrStore(t *testing.T) {
	// No ClickHouse can be contacted; no store directory may be created by planning.
	t.Setenv("CLICKHOUSE_URL", "http://127.0.0.1:1")
	source := &celRejectTransport{}
	previous := http.DefaultTransport
	http.DefaultTransport = source
	t.Cleanup(func() { http.DefaultTransport = previous })
	store := filepath.Join(t.TempDir(), "uncreated")
	server, err := configuredHTTPServer([]string{"--profile", "generic", "--investigations-dir", store})
	if err != nil {
		t.Fatal(err)
	}
	request, err := querycel.NewPlanRequest(exploration.Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 20}, `log.body == "timeout"`, "condition", "mount-1", "1")
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(request)
	for _, path := range []string{"/cel/plan", "/api/cel/plan"} {
		response := httptest.NewRecorder()
		newQueryBudget(server.Handler, 0).ServeHTTP(response, httptest.NewRequest("POST", path, bytes.NewReader(data)))
		if response.Code != 200 {
			t.Fatal(path, response.Code, response.Body.String())
		}
		var plan querycel.RecordPlan
		if err := json.Unmarshal(response.Body.Bytes(), &plan); err != nil || plan.Binding != request.Binding {
			t.Fatal(err, response.Body.String())
		}
	}
	if source.calls != 0 {
		t.Fatalf("planning made %d source calls", source.calls)
	}
	if _, err := os.Stat(store); !os.IsNotExist(err) {
		t.Fatalf("planning touched store: %v", err)
	}
}
