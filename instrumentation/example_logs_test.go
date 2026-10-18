package instrumentation

import (
	"context"
	"encoding/json"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExampleOTLPTraceCorrelation(t *testing.T) {
	var payload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("content type")
		}
		json.NewDecoder(r.Body).Decode(&payload)
		w.Write([]byte("{}"))
	}))
	defer server.Close()
	t.Setenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", server.URL)
	tid, _ := trace.TraceIDFromHex("0123456789abcdef0123456789abcdef")
	sid, _ := trace.SpanIDFromHex("0123456789abcdef")
	ctx := trace.ContextWithSpanContext(WithSubjectScope(context.Background(), "report", "tenant-a"), trace.NewSpanContext(trace.SpanContextConfig{TraceID: tid, SpanID: sid}))
	if e := EmitExampleLog(ctx, "reports", "instance-a", "report.ready", "s1", "ready"); e != nil {
		t.Fatal(e)
	}
	resources := payload["resourceLogs"].([]any)
	scope := resources[0].(map[string]any)["scopeLogs"].([]any)[0].(map[string]any)
	record := scope["logRecords"].([]any)[0].(map[string]any)
	if record["traceId"] != tid.String() || record["spanId"] != sid.String() || record["timeUnixNano"] == "" {
		t.Fatal("lost correlation")
	}
	if EmitExampleLog(context.Background(), "reports", "i", "x", "s", "") == nil {
		t.Fatal("unscoped log accepted")
	}
}
