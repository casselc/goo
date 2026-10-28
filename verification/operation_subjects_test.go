package verification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/instrumentation"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"hegel.dev/go/hegel"
	"os/exec"
	"testing"
)

func TestRuntimeSubjectScope(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)
	defer otel.SetTracerProvider(previous)
	defer provider.Shutdown(context.Background())
	ctx, span := provider.Tracer("test").Start(context.Background(), "parent")
	ctx = instrumentation.WithSubjectScope(ctx, "reservation", "tenant-a")
	instrumentation.Event(ctx, "reservation.accepted", func() string { return "id" }, func() string { return "" })
	instrumentation.Event(instrumentation.WithSubjectScope(ctx, "", "tenant-a"), "reservation.accepted", func() string { return "id" }, func() string { return "" })
	span.End()
	events := exporter.GetSpans()[0].Events
	if len(events) != 1 {
		t.Fatal("invalid scope was emitted")
	}
	attrs := map[string]string{}
	for _, a := range events[0].Attributes {
		attrs[string(a.Key)] = a.Value.AsString()
	}
	if attrs["goo.subject.namespace"] != "reservation" || attrs["goo.subject.tenant"] != "tenant-a" || attrs["goo.subject.id"] != "id" {
		t.Fatal(attrs)
	}
}
func TestHegelSubjectScopeIsolation(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		count := hegel.Draw(tc, hegel.Integers(1, 30))
		target := hegel.Draw(tc, hegel.Integers(0, 4))
		base := map[string]string{"environment": "local", "service": "checkout", "namespace": "reservation", "tenant": "demo", "id": "same"}
		rows := []map[string]any{}
		expected := 0
		keys := []string{"environment", "service", "namespace", "tenant", "id"}
		for i := 0; i < count; i++ {
			subject := map[string]string{}
			for k, v := range base {
				subject[k] = v
			}
			change := hegel.Draw(tc, hegel.Integers(0, 5))
			if change < 5 {
				subject[keys[change]] = fmt.Sprint(target, "/different")
			} else {
				expected++
			}
			rows = append(rows, map[string]any{"spanID": fmt.Sprint(i), "traceID": fmt.Sprint(i), "events": []map[string]any{{"subject": subject, "timeNS": "1792209000000000000", "name": "accepted"}}})
		}
		input, _ := json.Marshal(map[string]any{"spans": rows, "subject": base})
		cmd := exec.Command("node", "-e", `const fs=require('fs'),o=require('../ui/operations.mjs'),x=JSON.parse(fs.readFileSync(0,'utf8'));console.log(o.subjectSpans(x.spans,x.subject).length)`)
		cmd.Stdin = bytes.NewReader(input)
		out, err := cmd.CombinedOutput()
		if err != nil {
			tc.Fatalf("%s %v", out, err)
		}
		if string(bytes.TrimSpace(out)) != fmt.Sprint(expected) {
			tc.Fatalf("cross-scope join: %s expected %d", out, expected)
		}
	}, options(100)...)
}
