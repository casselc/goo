package exploration

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/casselc/goo/internal/sourcefailure"
)

func serviceProbeResponse379(r *http.Request) string {
	raw, _ := io.ReadAll(r.Body)
	body, _ := sourceCapabilityResponse431(string(raw))
	return body
}
func serviceDiscoveryReply431(status int, body string, code string) *http.Response {
	h := make(http.Header)
	if code != "" {
		h.Set("X-ClickHouse-Exception-Code", code)
	}
	return &http.Response{StatusCode: status, Header: h, Body: io.NopCloser(strings.NewReader(body))}
}
func TestOptionalMetricServiceDiscovery379(t *testing.T) {
	for _, present := range []bool{false, true} {
		t.Run(fmt.Sprint(present), func(t *testing.T) {
			reads := 0
			source := Source{Endpoint: "http://source", Client: &http.Client{Transport: serviceScopeTransport318(func(r *http.Request) (*http.Response, error) {
				reads++
				raw, _ := io.ReadAll(r.Body)
				sql := string(raw)
				if !present && sql == "DESCRIBE TABLE otel_metrics_exp_histogram FORMAT JSONEachRow" {
					return serviceDiscoveryReply431(404, "private missing-table diagnostic", "60"), nil
				}
				if body, handled := sourceCapabilityResponse431(sql); handled {
					return serviceDiscoveryReply431(200, body, ""), nil
				}
				if strings.Contains(sql, "WITH captured") {
					return serviceDiscoveryReply431(200, "", ""), nil
				}
				if strings.Contains(sql, "otel_metrics_exp_histogram") != present || strings.Count(sql, "LIMIT 1001") != 1 || strings.Contains(sql, "BucketCounts") {
					t.Fatal("global metric union changed", sql)
				}
				return serviceDiscoveryReply431(200, `{"service":"exp-only","signal":"metrics","count":"7","first_ms":"1000","last_ms":"1999"}`, ""), nil
			})}}
			result, err := source.DiscoverServices(t.Context(), InventoryQuestion{Environment: "prod", FromMS: 1000, ToMS: 2000})
			wantReads := 14
			if !present {
				wantReads = 13
			}
			if err != nil || reads != wantReads || len(result.Sources) != 6 || result.Diagnostics == nil || len(result.Signals) != 1 || result.Signals[0].Count != "7" {
				t.Fatal(result, err, reads)
			}
			if present && len(result.Diagnostics) != 0 || !present && (len(result.Diagnostics) != 1 || result.Diagnostics[0] != (ServiceDiscoveryDiagnostic{Kind: "exponential-histogram", Reason: "absent"})) {
				t.Fatal(result.Diagnostics)
			}
		})
	}
}
func TestOptionalMetricServiceFailures379(t *testing.T) {
	for _, phase := range []string{"probe-malformed", "probe-http", "union-race", "global-cap"} {
		t.Run(phase, func(t *testing.T) {
			aggregates := 0
			source := Source{Endpoint: "http://source", Client: &http.Client{Transport: serviceScopeTransport318(func(r *http.Request) (*http.Response, error) {
				raw, _ := io.ReadAll(r.Body)
				sql := string(raw)
				if strings.HasPrefix(sql, "DESCRIBE TABLE ") {
					if phase == "probe-malformed" {
						return serviceDiscoveryReply431(200, `{}`, ""), nil
					}
					if phase == "probe-http" {
						return serviceDiscoveryReply431(503, "private upstream diagnostic", ""), nil
					}
				}
				if body, handled := sourceCapabilityResponse431(sql); handled {
					return serviceDiscoveryReply431(200, body, ""), nil
				}
				aggregates++
				if phase == "union-race" {
					return serviceDiscoveryReply431(503, "private upstream diagnostic", ""), nil
				}
				return serviceDiscoveryReply431(200, strings.Repeat(`{"service":"a","signal":"metrics","count":"1","first_ms":"1000","last_ms":"1000"}`+"\n", 1001), ""), nil
			})}}
			result, err := source.DiscoverServices(t.Context(), InventoryQuestion{Environment: "prod", FromMS: 1000, ToMS: 2000})
			if err == nil || result != nil {
				t.Fatal("failure published", result, err)
			}
			wantAggregates := 0
			if phase == "union-race" || phase == "global-cap" {
				wantAggregates = 1
			}
			if aggregates != wantAggregates {
				t.Fatal("unready source acquired", aggregates)
			}
			want := sourcefailure.Unavailable
			if phase == "global-cap" {
				want = sourcefailure.Budget
			}
			if sourcefailure.Describe(err).Refusal.Kind != want {
				t.Fatal(err, sourcefailure.Describe(err))
			}
		})
	}
}
