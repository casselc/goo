package verification

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"hegel.dev/go/hegel"
)

// Availability varies independently of exact environment/namespace identity and
// exported-point counts. No histogram mass enters the service count contract.
func TestHegelServiceDiscovery379(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		present := hegel.Draw(tc, hegel.Integers(0, 1)) == 1
		mode := hegel.Draw(tc, hegel.Integers(0, 3))
		nsMode := hegel.Draw(tc, hegel.Integers(0, 3))
		n := hegel.Draw(tc, hegel.Integers(1, 10000))
		selector := identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: fmt.Sprintf("prod-'雪-%d", n)}}
		if mode == 0 {
			selector.Policy = "current-only"
		}
		if mode == 2 {
			selector.Scope = identity.EnvironmentScope{Mode: "empty"}
		}
		if mode == 3 {
			selector.Scope = identity.EnvironmentScope{Mode: "unreported"}
		}
		q := exploration.InventoryQuestion{EnvironmentSelector: &selector, FromMS: 1000, ToMS: 61000}
		if nsMode == 1 {
			q.Namespace = &identity.Namespace{Mode: "missing"}
		}
		if nsMode >= 2 {
			q.Namespace = &identity.Namespace{Mode: "value"}
			if nsMode == 3 {
				q.Namespace.Value = "team-'雪"
			}
		}
		before, _ := json.Marshal(q)
		count := fmt.Sprint(uint64(9007199254740992) + uint64(n))
		reads := 0
		source := exploration.Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: snapshotTransport(func(req *http.Request) (*http.Response, error) {
			reads++
			body, _ := io.ReadAll(req.Body)
			sql := string(body)
			text := ""
			switch reads {
			case 1:
				if sql != "EXISTS TABLE otel_metrics_exp_histogram FORMAT JSONEachRow" {
					tc.Fatal("unexpected availability query")
				}
				if present {
					text = `{"result":1}`
				} else {
					text = `{"result":0}`
				}
			case 2:
				if strings.Contains(sql, "otel_metrics_exp_histogram") != present || !strings.HasSuffix(sql, "LIMIT 1001 FORMAT JSONEachRow") || strings.Contains(sql, "BucketCounts") || strings.Contains(sql, "sum(Count)") {
					tc.Fatal("family/bound/count semantics changed")
				}
				if req.URL.Query().Get("param_from") != "1000" || req.URL.Query().Get("param_to") != "61000" || req.URL.Query().Get("readonly") != "1" {
					tc.Fatal("source scope changed")
				}
				var namespace any
				if q.Namespace != nil && q.Namespace.Mode == "value" {
					namespace = q.Namespace.Value
				}
				row, _ := json.Marshal(map[string]any{"namespace": namespace, "service": "worker-'雪", "signal": "metrics", "count": count, "first_ms": "1000", "last_ms": "61000"})
				// Upper bounds are exclusive: preserve a valid last event in the source row.
				text = strings.ReplaceAll(string(row), `"last_ms":"61000"`, `"last_ms":"60999"`)
			case 3:
				if !strings.Contains(sql, "WITH captured AS") || strings.Contains(sql, "otel_metrics_") {
					tc.Fatal("metric presence fabricated parent edges")
				}
			default:
				tc.Fatal("unbounded discovery reads")
			}
			return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(text))}, nil
		})}}
		result, err := source.DiscoverServices(t.Context(), q)
		if err != nil {
			tc.Fatal(err)
		}
		after, _ := json.Marshal(q)
		if string(before) != string(after) || !reflect.DeepEqual(result.Question, q) || reads != 3 || len(result.Signals) != 1 || result.Signals[0].Count != count || len(result.Dependencies) != 0 {
			tc.Fatal("discovery lost exact independent question/count/edge semantics")
		}
		if result.Diagnostics == nil || len(result.Diagnostics) != map[bool]int{true: 0, false: 1}[present] {
			tc.Fatal("missing source and empty result conflated")
		}
	}, options(100)...)
}
