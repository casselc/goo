package instrumentation

// This small synchronous transport is for the runnable examples. Production
// applications should configure an OTel logging SDK with batching and retry.
import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"go.opentelemetry.io/otel/trace"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// EmitExampleLog sends a bounded, trace-correlated OTLP/HTTP JSON log. It does
// not assert delivery durability and callers must not change business outcomes
// when telemetry export fails.
func EmitExampleLog(ctx context.Context, service, instance, body, subject, outcome string) error {
	scope, ok := ctx.Value(subjectScopeKey{}).(subjectScope)
	if !ok || scope.namespace == "" || scope.tenant == "" || subject == "" || len(subject) > 256 || len(body) > 128 || len(outcome) > 128 {
		return fmt.Errorf("invalid example log identity")
	}
	now := fmt.Sprint(time.Now().UnixNano())
	attrs := map[string]string{"goo.event.id": rand.Text(), "goo.event.time_ns": now, "goo.event.name": body, "goo.subject.id": subject, "goo.subject.namespace": scope.namespace, "goo.subject.tenant": scope.tenant, "goo.subject.outcome": outcome, "app.outcome": outcome, "app.release": "example-v1"}
	values := func(m map[string]string) []any {
		out := []any{}
		for k, v := range m {
			out = append(out, map[string]any{"key": k, "value": map[string]string{"stringValue": v}})
		}
		return out
	}
	record := map[string]any{"timeUnixNano": now, "observedTimeUnixNano": now, "severityNumber": 9, "severityText": "INFO", "body": map[string]string{"stringValue": body}, "attributes": values(attrs)}
	sc := trace.SpanContextFromContext(ctx)
	if sc.IsValid() {
		record["traceId"] = sc.TraceID().String()
		record["spanId"] = sc.SpanID().String()
		record["flags"] = uint32(sc.TraceFlags())
	}
	payload := map[string]any{"resourceLogs": []any{map[string]any{"resource": map[string]any{"attributes": values(map[string]string{"service.name": service, "service.instance.id": instance, "deployment.environment.name": "local"})}, "scopeLogs": []any{map[string]any{"scope": map[string]string{"name": "goo.examples.logging", "version": "0.1.0"}, "logRecords": []any{record}}}}}}
	data, e := json.Marshal(payload)
	if e != nil {
		return e
	}
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT")
	if endpoint == "" {
		endpoint = strings.TrimRight(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"), "/")
		if endpoint == "" {
			endpoint = "http://localhost:4318"
		}
		endpoint += "/v1/logs"
	}
	exportCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	req, e := http.NewRequestWithContext(exportCtx, http.MethodPost, endpoint, bytes.NewReader(data))
	if e != nil {
		return e
	}
	req.Header.Set("Content-Type", "application/json")
	response, e := http.DefaultClient.Do(req)
	if e != nil {
		return e
	}
	defer response.Body.Close()
	io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
	if response.StatusCode/100 != 2 {
		return fmt.Errorf("OTLP logs status %d", response.StatusCode)
	}
	return nil
}
