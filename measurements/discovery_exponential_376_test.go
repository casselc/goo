package measurements

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/casselc/goo/identity"
)

type discoveryTransport376 func(*http.Request) (*http.Response, error)

func (f discoveryTransport376) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func discoveryReply376(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
}
func discoveryRow376(kind string) string {
	raw, _ := json.Marshal(map[string]any{"kind": kind, "scope": "sdk", "version": "", "schemaURL": "", "metric": "http.server.request.duration", "unit": "s", "temporality": 1, "monotonic": false, "dimensions": []string{"route"}, "resource": map[string]string{"service.name": "worker", "service.instance.id": "one", "service.namespace": "team", "deployment.environment": "prod"}, "attributes": map[string]string{"route": "/work"}})
	return string(raw) + "\n"
}
func TestExponentialDiscoveryObservedOnly376(t *testing.T) {
	for _, typed := range []bool{false, true} {
		q := DiscoveryQuestion{Environment: "prod", Service: "worker", FromMS: 1000, ToMS: 2000}
		if typed {
			q.Environment = ""
			q.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}}
		}
		r, e := discoverRows(q, []json.RawMessage{json.RawMessage(discoveryRow376("histogram")), json.RawMessage(discoveryRow376(exponentialHistogramKind))})
		if e != nil || len(r.Instruments) != 2 || r.Records != 2 || r.Diagnostics == nil {
			t.Fatal("observed catalog", r, e)
		}
		for _, i := range r.Instruments {
			if i.Kind == exponentialHistogramKind {
				if i.Supported || i.Definition != nil || i.UnsupportedReason != exponentialHistogramUnsupportedReason || i.CapabilityPlan == nil || i.CapabilityPlan.Eligible || !slices.Contains(i.CapabilityPlan.Reasons, exponentialHistogramUnsupportedReason) || i.Records != 1 {
					t.Fatal("exponential metadata acquired execution authority", i)
				}
			} else if !i.Supported || i.Definition == nil {
				t.Fatal("explicit histogram regressed", i)
			}
		}
	}
}
func TestExponentialDiscoveryOptionalTable376(t *testing.T) {
	for _, present := range []bool{false, true} {
		t.Run(map[bool]string{false: "missing", true: "present"}[present], func(t *testing.T) {
			calls := 0
			q := DiscoveryQuestion{Service: "worker", FromMS: 1000, ToMS: 2000, Namespace: &identity.Namespace{Mode: "value", Value: "team"}, EnvironmentSelector: &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}}}
			before := q.clone()
			s := Source{Endpoint: "http://source", Client: &http.Client{Transport: discoveryTransport376(func(r *http.Request) (*http.Response, error) {
				calls++
				body, _ := io.ReadAll(r.Body)
				sql := string(body)
				if r.URL.Query().Get("readonly") != "1" {
					t.Fatal("source not readonly")
				}
				if calls == 1 {
					if sql != "EXISTS TABLE otel_metrics_exp_histogram FORMAT JSONEachRow" {
						t.Fatal("unbounded optional probe", sql)
					}
					q.EnvironmentSelector.Policy = "current-only"
					q.Namespace.Value = "changed"
					if present {
						return discoveryReply376(200, "{\"result\":1}\n"), nil
					}
					return discoveryReply376(200, "{\"result\":0}\n"), nil
				}
				unions := 2
				if present {
					unions = 3
				}
				if strings.Count(sql, " UNION ALL ") != unions || strings.Count(sql, "LIMIT 10001") != 1 || strings.Contains(sql, "otel_metrics_exp_histogram") != present || strings.Contains(sql, "PositiveBucketCounts") || strings.Contains(sql, "ZeroThreshold") || r.URL.Query().Get("param_namespace") != "team" || r.URL.Query().Get("param_environment") != "prod" || r.URL.Query().Get("param_from") != "1000" || r.URL.Query().Get("param_to") != "2000" {
					t.Fatal("metadata scope/global budget changed", sql, r.URL.RawQuery)
				}
				rows := discoveryRow376("histogram")
				if present {
					rows += discoveryRow376(exponentialHistogramKind)
				}
				return discoveryReply376(200, rows), nil
			})}}
			result, e := s.Discover(t.Context(), q)
			if e != nil || result == nil || calls != 2 || !reflect.DeepEqual(result.Question, before) {
				t.Fatal("discovery source/ownership", result, e, calls)
			}
			if present {
				if len(result.Diagnostics) != 0 || result.Records != 2 {
					t.Fatal("available source labeled missing", result)
				}
			} else if !reflect.DeepEqual(result.Diagnostics, []DiscoveryDiagnostic{{Kind: exponentialHistogramKind, Reason: "optional-table-missing"}}) || result.Records != 1 || !result.Instruments[0].Supported {
				t.Fatal("optional absence broke supported discovery", result)
			}
		})
	}
}
func TestExponentialDiscoveryProbeRefusesMalformed376(t *testing.T) {
	for _, body := range []string{"", "{}\n", "{\"result\":-0}\n", "{\"result\":null}\n", "{\"result\":2}\n", "{\"result\":true}\n", "{\"result\":\"0\"}\n", "{\"result\":0,\"other\":1}\n", "{\"result\":0}\n{\"result\":1}\n"} {
		calls := 0
		s := Source{Endpoint: "http://source", Client: &http.Client{Transport: discoveryTransport376(func(*http.Request) (*http.Response, error) { calls++; return discoveryReply376(200, body), nil })}}
		result, e := s.Discover(t.Context(), DiscoveryQuestion{Environment: "prod", Service: "worker", FromMS: 1000, ToMS: 2000})
		if e == nil || result != nil || calls != 1 {
			t.Fatal("malformed availability became absence", body, result, e, calls)
		}
	}
}
func TestExponentialDiscoverySourceFailures376(t *testing.T) {
	for _, failAt := range []int{1, 2} {
		calls := 0
		s := Source{Endpoint: "http://source", Client: &http.Client{Transport: discoveryTransport376(func(*http.Request) (*http.Response, error) {
			calls++
			if calls == failAt {
				return discoveryReply376(500, "unavailable"), nil
			}
			return discoveryReply376(200, "{\"result\":1}\n"), nil
		})}}
		result, e := s.Discover(t.Context(), DiscoveryQuestion{Environment: "prod", Service: "worker", FromMS: 1000, ToMS: 2000})
		if e == nil || result != nil || calls != failAt {
			t.Fatal("source error became partial discovery", result, e, calls)
		}
	}
	calls := 0
	s := Source{Endpoint: "http://source", Client: &http.Client{Transport: discoveryTransport376(func(*http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			return discoveryReply376(200, "{\"result\":1}\n"), nil
		}
		return discoveryReply376(200, strings.Repeat(discoveryRow376("histogram"), RowLimit)+discoveryRow376(exponentialHistogramKind)), nil
	})}}
	result, e := s.Discover(t.Context(), DiscoveryQuestion{Environment: "prod", Service: "worker", FromMS: 1000, ToMS: 2000})
	if e == nil || result != nil || calls != 2 {
		t.Fatal("combined supported/exponential limit truncated", result, e, calls)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	s.Client = &http.Client{Transport: discoveryTransport376(func(r *http.Request) (*http.Response, error) { return nil, r.Context().Err() })}
	if result, e = s.Discover(ctx, DiscoveryQuestion{Environment: "prod", Service: "worker", FromMS: 1000, ToMS: 2000}); e == nil || result != nil {
		t.Fatal("cancellation became source absence", result, e)
	}
}
