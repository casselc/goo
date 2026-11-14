package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// Use the checked-in standard exporter metadata without copying the production
// profile definitions or weakening source admission in the HTTP fixtures.
func sourcecapFixture431(t *testing.T, sql string) (string, bool) {
	t.Helper()
	if strings.HasPrefix(sql, "SELECT ") && strings.HasSuffix(sql, " LIMIT 0 FORMAT JSONEachRow") {
		return "", true
	}
	if !strings.HasPrefix(sql, "DESCRIBE TABLE ") {
		return "", false
	}
	table := strings.TrimSuffix(strings.TrimPrefix(sql, "DESCRIBE TABLE "), " FORMAT JSONEachRow")
	kind := map[string]string{"otel_logs": "logs", "otel_traces": "spans", "otel_metrics_gauge": "gauge", "otel_metrics_sum": "counter", "otel_metrics_histogram": "histogram", "otel_metrics_exp_histogram": "exponential-histogram"}[table]
	if kind == "" {
		t.Fatal("unexpected capability table", table)
	}
	raw, err := os.ReadFile("../../internal/sourcecap/testdata/schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var bank map[string][]json.RawMessage
	if err = json.Unmarshal(raw, &bank); err != nil {
		t.Fatal(err)
	}
	var response strings.Builder
	for _, row := range bank[kind] {
		compact, err := json.Marshal(row)
		if err != nil {
			t.Fatal(err)
		}
		response.Write(compact)
		response.WriteByte('\n')
	}
	return response.String(), true
}

func TestSourcecapFixtureNDJSON431(t *testing.T) {
	for _, table := range []string{"otel_logs", "otel_traces", "otel_metrics_gauge", "otel_metrics_sum", "otel_metrics_histogram", "otel_metrics_exp_histogram"} {
		t.Run(table, func(t *testing.T) {
			response, handled := sourcecapFixture431(t, "DESCRIBE TABLE "+table+" FORMAT JSONEachRow")
			if !handled || response == "" {
				t.Fatal("missing schema response")
			}
			seen := map[string]bool{}
			for _, line := range strings.Split(strings.TrimSuffix(response, "\n"), "\n") {
				var column struct {
					Name string `json:"name"`
					Type string `json:"type"`
				}
				if json.Unmarshal([]byte(line), &column) != nil || column.Name == "" || column.Type == "" || seen[column.Name] {
					t.Fatal("invalid NDJSON schema row", line)
				}
				seen[column.Name] = true
			}
		})
	}
}
