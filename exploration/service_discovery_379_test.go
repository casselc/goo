package exploration

import (
	"fmt"
	"github.com/casselc/goo/internal/optionalmetric"
	"github.com/casselc/goo/internal/sourcefailure"
	"io"
	"net/http"
	"strings"
	"testing"
)

func serviceProbeResponse379(r *http.Request) string {
	raw, _ := io.ReadAll(r.Body)
	if string(raw) == optionalmetric.ExponentialHistogramProbeSQL {
		return `{"result":1}`
	}
	return ""
}
func TestOptionalMetricServiceDiscovery379(t *testing.T) {
	for _, present := range []bool{false, true} {
		t.Run(fmt.Sprint(present), func(t *testing.T) {
			reads := 0
			source := Source{Endpoint: "http://source", Client: &http.Client{Transport: serviceScopeTransport318(func(r *http.Request) (*http.Response, error) {
				reads++
				raw, _ := io.ReadAll(r.Body)
				sql := string(raw)
				body := ""
				switch reads {
				case 1:
					if sql != optionalmetric.ExponentialHistogramProbeSQL {
						t.Fatal(sql)
					}
					body = fmt.Sprintf(`{"result":%d}`, map[bool]int{false: 0, true: 1}[present])
				case 2:
					if strings.Contains(sql, "otel_metrics_exp_histogram") != present || strings.Count(sql, "LIMIT 1001") != 1 || strings.Contains(sql, "BucketCounts") {
						t.Fatal("global metric union changed", sql)
					}
					body = `{"service":"exp-only","signal":"metrics","count":"7","first_ms":"1000","last_ms":"1999"}`
				case 3:
					if !strings.Contains(sql, "WITH captured") {
						t.Fatal("edges changed")
					}
				default:
					t.Fatal("extra source read")
				}
				return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
			})}}
			result, err := source.DiscoverServices(t.Context(), InventoryQuestion{Environment: "prod", FromMS: 1000, ToMS: 2000})
			if err != nil || reads != 3 || result.Diagnostics == nil || len(result.Signals) != 1 || result.Signals[0].Count != "7" {
				t.Fatal(result, err, reads)
			}
			if present && len(result.Diagnostics) != 0 || !present && (len(result.Diagnostics) != 1 || result.Diagnostics[0] != (ServiceDiscoveryDiagnostic{Kind: "exponential-histogram", Reason: "optional-table-missing"})) {
				t.Fatal(result.Diagnostics)
			}
		})
	}
}
func TestOptionalMetricServiceFailures379(t *testing.T) {
	for _, phase := range []string{"probe-malformed", "probe-http", "union-race", "global-cap"} {
		t.Run(phase, func(t *testing.T) {
			reads := 0
			source := Source{Endpoint: "http://source", Client: &http.Client{Transport: serviceScopeTransport318(func(r *http.Request) (*http.Response, error) {
				reads++
				status := 200
				body := `{"result":1}`
				if reads == 1 && phase == "probe-malformed" {
					body = `{"result":1,"result":0}`
				}
				if reads == 1 && phase == "probe-http" || reads == 2 && phase == "union-race" {
					status = 503
					body = "private upstream failure"
				}
				if reads == 2 && phase == "global-cap" {
					body = strings.Repeat(`{"service":"a","signal":"metrics","count":"1","first_ms":"1000","last_ms":"1000"}`+"\n", 1001)
				}
				return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
			})}}
			result, err := source.DiscoverServices(t.Context(), InventoryQuestion{Environment: "prod", FromMS: 1000, ToMS: 2000})
			if err == nil || result != nil {
				t.Fatal("failure published", result, err)
			}
			wantReads := 1
			if phase == "union-race" || phase == "global-cap" {
				wantReads = 2
			}
			if reads != wantReads {
				t.Fatal(reads)
			}
			want := sourcefailure.Unavailable
			if phase == "probe-malformed" {
				want = sourcefailure.Malformed
			}
			if phase == "global-cap" {
				want = sourcefailure.Budget
			}
			if sourcefailure.Describe(err).Refusal.Kind != want {
				t.Fatal(err, sourcefailure.Describe(err))
			}
		})
	}
}
