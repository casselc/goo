package annotations

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLiveOperationSource(t *testing.T) {
	m, err := ScanDir("../examples/operation-app")
	if err != nil {
		t.Fatal(err)
	}
	source := OperationSource{Service: "lab-go-checkout", Environment: "local", Instance: "woven-demo", DeclaredSchema: m.SchemaURL}
	row := OperationRow{TimestampNS: "1000000", DurationNS: "9007199254740993", TraceID: strings.Repeat("1", 32), SpanID: strings.Repeat("2", 16), Kind: "Server", Scope: m.Scope, Version: m.ScopeVersion, Resource: map[string]string{"service.name": source.Service, "deployment.environment.name": source.Environment, "service.instance.id": source.Instance}, Attributes: map[string]string{"goo.operation.name": "checkout.reserve", "goo.operation.role": "server", "goo.operation.outcome": "error"}}
	sql, params, err := m.OperationQuery(source, 1, 2)
	if err != nil || !strings.Contains(sql, "LIMIT 1001") || strings.Contains(sql, source.Service) || params["service"] != source.Service {
		t.Fatalf("unsafe query: %s %v %v", sql, params, err)
	}
	report, n, err := m.ObserveRows(source, 1, 2, []OperationRow{row, row})
	if err != nil || n != 1 || report.Operations[1].Calls != 1 {
		t.Fatalf("%+v %d %v", report, n, err)
	}
	for _, change := range []func(*OperationRow){func(r *OperationRow) { r.TimestampNS = "2000000" }, func(r *OperationRow) { r.DurationNS = "-1" }, func(r *OperationRow) { r.Kind = "Client" }, func(r *OperationRow) { r.Scope = "other" }, func(r *OperationRow) { r.ParentSpanID = "bad" }} {
		bad := row
		change(&bad)
		if _, _, e := m.ObserveRows(source, 1, 2, []OperationRow{bad}); e == nil {
			t.Fatal("accepted invalid source")
		}
	}
	changed := row
	changed.TimestampNS = "1000001"
	if _, _, e := m.ObserveRows(source, 1, 2, []OperationRow{row, changed}); e == nil {
		t.Fatal("accepted conflicting raw copy")
	}
	if _, _, e := m.ObserveRows(source, 1, 2, make([]OperationRow, 1001)); e == nil {
		t.Fatal("accepted partial population")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("param_service") != source.Service || r.URL.Query().Get("readonly") != "1" {
			t.Error("lost source controls")
		}
		json.NewEncoder(w).Encode(row)
		json.NewEncoder(w).Encode(row)
	}))
	defer server.Close()
	live, err := FetchOperations(context.Background(), server.Client(), server.URL, "u", "p", *m, source, 1, 2)
	if err != nil || live.Conformance.Operations[1].Calls != 1 || live.DuplicateRecords != 1 || len(live.Spans) != 1 || live.Spans[0].DurationNS != row.DurationNS || live.Spans[0].StartNS != row.TimestampNS {
		t.Fatalf("%+v %v", live, err)
	}
}
