package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"hegel.dev/go/hegel"
)

type discoveryIdentity376 struct {
	Kind, Scope, Version, Schema, Metric, Unit string
	Temporality                                int
	Monotonic                                  bool
}

func identity376(i measurements.Instrument) discoveryIdentity376 {
	return discoveryIdentity376{i.Kind, i.Scope, i.Version, i.SchemaURL, i.Metric, i.Unit, i.Temporality, i.Monotonic}
}
func TestHegelUnsupportedMetricDiscovery376(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		mode := hegel.Draw(tc, hegel.Integers(0, 3))
		nsMode := hegel.Draw(tc, hegel.Integers(0, 3))
		present := hegel.Draw(tc, hegel.Integers(0, 1)) == 1
		q := measurements.DiscoveryQuestion{Service: fmt.Sprintf("worker-雪-%d", seed), FromMS: 1000, ToMS: 61000, EnvironmentSelector: &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: fmt.Sprintf("prod-%d", seed)}}}
		if mode == 0 {
			q.EnvironmentSelector.Policy = "current-only"
		} else if mode == 2 {
			q.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "empty"}
		} else if mode == 3 {
			q.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "unreported"}
		}
		if nsMode == 1 {
			q.Namespace = &identity.Namespace{Mode: "missing"}
		} else if nsMode >= 2 {
			q.Namespace = &identity.Namespace{Mode: "value"}
			if nsMode == 3 {
				q.Namespace.Value = "team-雪"
			}
		}
		before, _ := json.Marshal(q)
		resources := func(instance int) map[string]string {
			r := map[string]string{"service.name": q.Service, "service.instance.id": fmt.Sprint(instance), "unrelated.resource": "literal"}
			if mode != 3 {
				r["deployment.environment.name"] = q.EnvironmentSelector.Scope.Value
			}
			if nsMode >= 2 {
				r["service.namespace"] = q.Namespace.Value
			}
			return r
		}
		expected := map[discoveryIdentity376]int{}
		dimensions := map[discoveryIdentity376]map[string]bool{}
		encoded := []string{}
		add := func(kind, version, schema, unit string, temporality, instance int) {
			raw := map[string]any{"kind": kind, "scope": "otel.discovery", "version": version, "schemaURL": schema, "metric": "http.server.request.duration", "unit": unit, "temporality": temporality, "monotonic": false, "dimensions": []string{fmt.Sprintf("dimension.%d", instance%3), "common"}, "resource": resources(instance), "resourceSchemaURL": "https://opentelemetry.io/schemas/1.30.0", "attributes": map[string]string{"common": "", fmt.Sprintf("dimension.%d", instance%3): fmt.Sprint(instance)}}
			b, _ := json.Marshal(raw)
			encoded = append(encoded, string(b))
			key := discoveryIdentity376{kind, "otel.discovery", version, schema, "http.server.request.duration", unit, temporality, false}
			expected[key]++
			if dimensions[key] == nil {
				dimensions[key] = map[string]bool{}
			}
			dimensions[key][fmt.Sprintf("dimension.%d", instance%3)] = true
			dimensions[key]["common"] = true
		}
		for i := 0; i < 3; i++ {
			add("histogram", "", "", "s", 2, i)
		}
		if present {
			for i := 0; i < 3; i++ {
				add("exponential-histogram", "", "", "s", 2, i)
			}
			add("exponential-histogram", "2", "", "s", 2, 7)
			add("exponential-histogram", "", "https://opentelemetry.io/schemas/1.37.0", "s", 2, 8)
			add("exponential-histogram", "", "", "ms", 2, 9)
			add("exponential-histogram", "", "", "s", 1, 10)
		}
		reads := 0
		source := measurements.Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: snapshotTransport(func(req *http.Request) (*http.Response, error) {
			reads++
			body, _ := io.ReadAll(req.Body)
			sql := string(body)
			text := ""
			if reads == 1 {
				if sql != "EXISTS TABLE otel_metrics_exp_histogram FORMAT JSONEachRow" {
					tc.Fatalf("wrong bounded availability probe %q", sql)
				}
				if present {
					text = `{"result":1}`
				} else {
					text = `{"result":0}`
				}
			} else {
				if reads != 2 || !strings.HasPrefix(sql, "SELECT * FROM (") || !strings.HasSuffix(sql, ") LIMIT 10001 FORMAT JSONEachRow") {
					tc.Fatalf("discovery is not one globally bounded inventory: %q", sql)
				}
				if strings.Contains(sql, "otel_metrics_exp_histogram") != present {
					tc.Fatal("optional table query differs from availability")
				}
				if req.URL.Query().Get("param_service") != q.Service || req.URL.Query().Get("param_from") != "1000" || req.URL.Query().Get("param_to") != "61000" || req.URL.Query().Get("readonly") != "1" || req.URL.Query().Get("max_result_rows") != "10001" {
					tc.Fatal("bounded source scope parameters changed")
				}
				text = strings.Join(encoded, "\n")
			}
			return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(text))}, nil
		})}}
		result, err := source.Discover(t.Context(), q)
		if err != nil {
			tc.Fatal(err)
		}
		after, _ := json.Marshal(q)
		if string(before) != string(after) || reads != 2 || result.Records != len(encoded) || len(result.Instruments) != len(expected) || !reflect.DeepEqual(result.Question, q) {
			tc.Fatal("discovery changed question, read budget, grouping or exported point counts")
		}
		seen := map[discoveryIdentity376]bool{}
		for _, instrument := range result.Instruments {
			key := identity376(instrument)
			if seen[key] || instrument.Records != expected[key] {
				tc.Fatal("instrument identity merged or duplicated", key)
			}
			seen[key] = true
			want := []string{}
			for dimension := range dimensions[key] {
				want = append(want, dimension)
			}
			slices.Sort(want)
			if !reflect.DeepEqual(instrument.Dimensions, want) {
				tc.Fatal("observed dimension union changed")
			}
			if instrument.Kind == "exponential-histogram" {
				if instrument.Supported || instrument.Definition != nil || !strings.Contains(instrument.UnsupportedReason, "zero-threshold") || !strings.Contains(instrument.UnsupportedReason, "geometry") || instrument.CapabilityPlan == nil || instrument.CapabilityPlan.Eligible || instrument.CapabilityPlan.Verification != "not-run" {
					tc.Fatal("unsupported exponential metadata became executable", instrument)
				}
			} else if !instrument.Supported || instrument.Definition == nil {
				tc.Fatal("ordinary explicit histogram was unnecessarily disabled")
			}
		}
		if present {
			if len(result.Diagnostics) != 0 {
				tc.Fatal("available source published absence")
			}
		} else if len(result.Diagnostics) != 1 || result.Diagnostics[0].Kind != "exponential-histogram" || result.Diagnostics[0].Reason != "optional-table-missing" {
			tc.Fatal("missing optional table not explicitly diagnosed")
		}
	}, options(40)...)
}

func TestUnsupportedMetricDiscoveryAvailabilityAndCap376(t *testing.T) {
	for _, body := range []string{"", `null`, `{}`, `{"result":null}`, `{"result":2}`, `{"result":false}`, `{"result":-0}`, `{"result":0.0}`, `{"result":1e0}`, `{"result":"0"}`, `{"result":0,"unknown":true}`, "{\"result\":0}\n{\"result\":1}"} {
		t.Run(body, func(t *testing.T) {
			reads := 0
			s := measurements.Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: snapshotTransport(func(*http.Request) (*http.Response, error) {
				reads++
				return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
			})}}
			r, err := s.Discover(t.Context(), measurements.DiscoveryQuestion{Environment: "prod", Service: "worker", FromMS: 1000, ToMS: 61000})
			if err == nil || r != nil || reads != 1 {
				t.Fatal("ambiguous availability became absence or source reads", reads, err)
			}
		})
	}
	for _, mode := range []string{"probe-http-error", "union-http-error", "cancelled", "global-cap"} {
		t.Run(mode, func(t *testing.T) {
			reads := 0
			ctx := t.Context()
			s := measurements.Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: snapshotTransport(func(req *http.Request) (*http.Response, error) {
				reads++
				if mode == "cancelled" {
					return nil, context.Canceled
				}
				status, text := 200, `{"result":1}`
				if reads == 1 && mode == "probe-http-error" || reads == 2 && mode == "union-http-error" {
					status = 500
					text = "unavailable"
				} else if reads == 2 && mode == "global-cap" {
					text = strings.Repeat("{}\n", 10001)
				}
				return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(text))}, nil
			})}}
			r, err := s.Discover(ctx, measurements.DiscoveryQuestion{Environment: "prod", Service: "worker", FromMS: 1000, ToMS: 61000})
			if err == nil || r != nil || reads > 2 {
				t.Fatal("failed/capped discovery published partial inventory", reads, err)
			}
			if mode == "global-cap" && !strings.Contains(err.Error(), "10000") {
				t.Fatal("global row cap was not the rejecting boundary", err)
			}
			if (mode == "probe-http-error" || mode == "cancelled") && reads != 1 {
				t.Fatal("availability failure continued to telemetry")
			}
		})
	}
}
