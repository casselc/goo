package exploration

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/casselc/goo/internal/sourcecap"
)

func TestServiceDiscoverySelectedFamilies431(t *testing.T) {
	q := InventoryQuestion{Environment: "prod'\n", FromMS: 1000, ToMS: 2000}
	for _, kind := range []string{sourcecap.Logs, sourcecap.Spans, sourcecap.Gauge, sourcecap.Counter, sourcecap.Histogram, sourcecap.ExponentialHistogram} {
		t.Run(kind, func(t *testing.T) {
			p, err := compileServiceDiscoveryKinds(q, []string{kind})
			if err != nil {
				t.Fatal(err)
			}
			tables := map[string]string{sourcecap.Logs: "otel_logs", sourcecap.Spans: "otel_traces", sourcecap.Gauge: "otel_metrics_gauge", sourcecap.Counter: "otel_metrics_sum", sourcecap.Histogram: "otel_metrics_histogram", sourcecap.ExponentialHistogram: "otel_metrics_exp_histogram"}
			for family, table := range tables {
				if strings.Contains(p.Count, table) != (family == kind) {
					t.Fatal("unavailable family entered union", family, p.Count)
				}
			}
			if (p.Groups != "") != (kind == sourcecap.Spans) {
				t.Fatal("relationships require captured spans", p.Groups)
			}
			if strings.Count(p.Count, "LIMIT 1001") != 1 || strings.Contains(p.Count, q.Environment) {
				t.Fatal("global bound or typed environment lost", p)
			}
		})
	}
	for _, kinds := range [][]string{nil, {"unknown"}} {
		if _, err := compileServiceDiscoveryKinds(q, kinds); err == nil {
			t.Fatal("invalid family set accepted", kinds)
		}
	}
}

func TestServiceDiscoveryPartialFamilies431(t *testing.T) {
	for _, kind := range []string{sourcecap.Logs, sourcecap.Spans, sourcecap.Gauge} {
		t.Run(kind, func(t *testing.T) {
			table := map[string]string{sourcecap.Logs: "otel_logs", sourcecap.Spans: "otel_traces", sourcecap.Gauge: "otel_metrics_gauge"}[kind]
			aggregates, edges := 0, 0
			source := Source{Endpoint: "http://source", Client: &http.Client{Transport: serviceScopeTransport318(func(r *http.Request) (*http.Response, error) {
				raw, _ := io.ReadAll(r.Body)
				sql := string(raw)
				if strings.HasPrefix(sql, "DESCRIBE TABLE ") && !strings.HasPrefix(sql, "DESCRIBE TABLE "+table+" ") {
					if strings.Contains(sql, "otel_metrics_exp_histogram") {
						return serviceDiscoveryReply431(403, "private credential diagnostic", "497"), nil
					}
					return serviceDiscoveryReply431(404, "not exposed", "60"), nil
				}
				if body, handled := sourceCapabilityResponse431(sql); handled {
					return serviceDiscoveryReply431(200, body, ""), nil
				}
				if strings.Contains(sql, "WITH captured") {
					edges++
					return serviceDiscoveryReply431(200, "", ""), nil
				}
				aggregates++
				signal := kind
				if kind == sourcecap.Gauge {
					signal = "metrics"
				}
				return serviceDiscoveryReply431(200, fmt.Sprintf(`{"service":"ordinary","signal":%q,"count":"1","first_ms":"1000","last_ms":"1000"}`, signal), ""), nil
			})}}
			result, err := source.DiscoverServices(t.Context(), InventoryQuestion{Environment: "prod", FromMS: 1000, ToMS: 2000})
			if err != nil || result == nil || len(result.Signals) != 1 || len(result.Sources) != 6 || len(result.Diagnostics) != 5 || aggregates != 1 {
				t.Fatal(result, err, aggregates)
			}
			for _, family := range result.Sources {
				if family.Kind == sourcecap.ExponentialHistogram && family.State != "unreadable" {
					t.Fatal("denied family became absence", family)
				}
			}
			if (edges == 1) != (kind == sourcecap.Spans) {
				t.Fatal("parent read without spans", edges)
			}
		})
	}
}
