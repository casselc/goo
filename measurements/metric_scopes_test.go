package measurements

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/internal/sourcecap"
	"github.com/casselc/goo/internal/sourcefailure"
)

func metricScopeRow432(changes map[string]string) json.RawMessage {
	row := map[string]string{"environment_present": "1", "environment": "prod", "namespace_present": "0", "namespace": "", "service": "worker", "points": "3", "first_ms": "100001", "last_ms": "159999"}
	for k, v := range changes {
		row[k] = v
	}
	encoded, _ := json.Marshal(row)
	return encoded
}

var metricScopeQuestion432 = MetricScopesQuestion{FromMS: 100000, ToMS: 160000}

func TestMetricScopeCompiler432(t *testing.T) {
	p, err := compileMetricScopes(metricScopeQuestion432, []string{sourcecap.Gauge, sourcecap.Counter})
	if err != nil {
		t.Fatal(err)
	}
	for _, part := range []string{"FROM otel_metrics_gauge", "FROM otel_metrics_sum", "sum(toUInt128(n))", "mapContains(ResourceAttributes,'deployment.environment.name')", "mapContains(ResourceAttributes,'deployment.environment')", "mapContains(ResourceAttributes,'service.namespace')", "if(mapContains(ResourceAttributes,'deployment.environment.name'),ResourceAttributes['deployment.environment.name'],ResourceAttributes['deployment.environment'])", "GROUP BY env_present,env_value,ns_present,ns_value,service", "LIMIT 129 FORMAT JSONEachRow"} {
		if !strings.Contains(p.SQL, part) {
			t.Fatalf("missing fixed profile clause %q", part)
		}
	}
	if len(p.Parameters) != 2 || p.Parameters["from"] != "100000" || p.Parameters["to"] != "160000" || strings.Contains(p.SQL, "WHERE ServiceName=") || strings.Contains(p.SQL, "otel_logs") || strings.Contains(p.SQL, "otel_traces") || strings.Contains(p.SQL, "otel_metrics_histogram") {
		t.Fatal("guessed identity or unrelated source", p)
	}
	for _, kinds := range [][]string{nil, {sourcecap.Logs}, {sourcecap.Gauge, sourcecap.Gauge}, {"otel_metrics_gauge; DROP TABLE secret"}} {
		if _, err := compileMetricScopes(metricScopeQuestion432, kinds); err == nil {
			t.Fatal("unsafe family admitted", kinds)
		}
	}
	for _, q := range []MetricScopesQuestion{{}, {FromMS: -1, ToMS: 1}, {FromMS: 1, ToMS: 7200002}, {FromMS: 1, ToMS: 4102444800001}} {
		if q.Validate() == nil {
			t.Fatal("invalid window admitted", q)
		}
	}
}
func TestMetricScopeRowsPreserveTypedIdentity432(t *testing.T) {
	rows := []json.RawMessage{
		metricScopeRow432(map[string]string{"environment_present": "0", "environment": ""}),
		metricScopeRow432(map[string]string{"environment": "", "namespace_present": "1", "namespace": ""}),
		metricScopeRow432(map[string]string{"namespace_present": "1", "namespace": "payments", "points": "18446744073709551616"}),
	}
	choices, err := metricScopeRows(metricScopeQuestion432, rows)
	if err != nil || len(choices) != 3 {
		t.Fatal(choices, err)
	}
	for i, c := range choices {
		if !c.Selectable || c.EnvironmentSelector == nil || c.EnvironmentSelector.Policy != "effective" || c.EnvironmentSelector.Scope != c.Environment || c.Namespace.Validate() != nil || c.Reason != "" {
			t.Fatal("unchecked executable identity", c)
		}
		q := DiscoveryQuestion{EnvironmentSelector: c.EnvironmentSelector, Namespace: c.Namespace, Service: c.Service, FromMS: metricScopeQuestion432.FromMS, ToMS: metricScopeQuestion432.ToMS}
		if q.Validate() != nil {
			t.Fatal("choice cannot become existing discovery", i, q.Validate())
		}
	}
	if choices[0].Environment.Mode != "unreported" || choices[0].Namespace.Mode != "missing" || choices[1].Environment.Mode != "empty" || choices[1].Namespace.Mode != "value" || choices[1].Namespace.Value != "" || choices[2].Namespace.Value != "payments" || choices[2].Points != "18446744073709551616" {
		t.Fatal("presence, namespace or exact count collapsed", choices)
	}
}
func TestMetricScopeUnselectableIdentity432(t *testing.T) {
	cases := []struct {
		changes  map[string]string
		reason   string
		selector bool
	}{
		{map[string]string{"service": ""}, "Service name is empty or not reported.", true},
		{map[string]string{"service": strings.Repeat("s", 129)}, "Service name exceeds selector limits.", true},
		{map[string]string{"environment": strings.Repeat("e", 129)}, "Environment label exceeds selector limits.", false},
		{map[string]string{"namespace_present": "1", "namespace": strings.Repeat("n", 129)}, "Namespace exceeds selector limits.", true},
		{map[string]string{"environment": "with\x00zero"}, "Environment label exceeds selector limits.", false},
	}
	for _, tc := range cases {
		choices, err := metricScopeRows(metricScopeQuestion432, []json.RawMessage{metricScopeRow432(tc.changes)})
		if err != nil {
			t.Fatal(err)
		}
		c := choices[0]
		if c.Selectable || c.Reason != tc.reason || (c.EnvironmentSelector != nil) != tc.selector {
			t.Fatal("unsafe identity offered", c)
		}
	}
}
func TestMetricScopeRowsRejectIncompleteSource432(t *testing.T) {
	for _, changes := range []map[string]string{{"points": "0"}, {"points": "01"}, {"points": "-1"}, {"points": "340282366920938463463374607431768211456"}, {"first_ms": "100000.0"}, {"first_ms": "99999"}, {"last_ms": "160000"}, {"first_ms": "150000", "last_ms": "140000"}, {"environment_present": "false"}, {"environment_present": "0", "environment": "present"}, {"namespace_present": "0", "namespace": "present"}, {"service": strings.Repeat("x", 4097)}} {
		if _, err := metricScopeRows(metricScopeQuestion432, []json.RawMessage{metricScopeRow432(changes)}); err == nil {
			t.Fatal("invalid source identity admitted", changes)
		}
	}
	for _, rows := range [][]json.RawMessage{{json.RawMessage(`{}`)}, {metricScopeRow432(nil), metricScopeRow432(nil)}, make([]json.RawMessage, 129)} {
		if _, err := metricScopeRows(metricScopeQuestion432, rows); err == nil {
			t.Fatal("partial, duplicate or oversized groups admitted")
		}
	}
}

type metricScopesTransport432 struct {
	mode  string
	calls []string
	rows  []json.RawMessage
}

func (c *metricScopesTransport432) RoundTrip(r *http.Request) (*http.Response, error) {
	body, _ := io.ReadAll(r.Body)
	sql := string(body)
	c.calls = append(c.calls, sql)
	if r.URL.Query().Get("readonly") != "1" {
		return nil, errors.New("readonly source policy missing")
	}
	status, out, headers := 200, "", make(http.Header)
	switch {
	case strings.HasPrefix(sql, "DESCRIBE TABLE "):
		if c.mode == "absent" || c.mode == "denied" || !strings.Contains(sql, "otel_metrics_gauge ") {
			status = 404
			headers.Set("X-ClickHouse-Exception-Code", "60")
			if c.mode == "denied" {
				status = 403
				headers.Set("X-ClickHouse-Exception-Code", "497")
			}
		} else {
			out, _ = sourceCapabilityResponse431(sql)
		}
	case strings.HasSuffix(sql, " LIMIT 0 FORMAT JSONEachRow"):
	case strings.Contains(sql, "sum(toUInt128(n))"):
		if c.mode == "race" {
			status = 500
		} else {
			for _, row := range c.rows {
				out += string(row) + "\n"
			}
		}
	default:
		return nil, errors.New("unexpected source SQL")
	}
	return &http.Response{StatusCode: status, Header: headers, Body: io.NopCloser(strings.NewReader(out))}, nil
}
func TestDiscoverMetricScopesPartialFamilies432(t *testing.T) {
	transport := &metricScopesTransport432{rows: []json.RawMessage{metricScopeRow432(nil)}}
	s := Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: transport}}
	result, err := s.DiscoverMetricScopes(context.Background(), metricScopeQuestion432)
	if err != nil {
		t.Fatal(err)
	}
	if len(transport.calls) != 6 || result.Version != "goo/metric-scopes/v1" || result.Question != metricScopeQuestion432 || len(result.Sources) != 4 || len(result.Diagnostics) != 3 || len(result.Choices) != 1 || result.Unit != "exported-points" || result.Choices[0].Points != "3" {
		t.Fatal("partial source identity changed", result, transport.calls)
	}
	if result.Sources[0].State != "compatible" || result.Sources[1].State != "absent" || !strings.Contains(result.Warning, "not metric values") {
		t.Fatal("availability became coverage", result)
	}
	transport.rows = nil
	empty, err := s.DiscoverMetricScopes(context.Background(), metricScopeQuestion432)
	if err != nil || empty.Choices == nil || len(empty.Choices) != 0 {
		t.Fatal("successful empty discovery differs from absent", empty, err)
	}
}
func TestMetricScopeDiscoveryRefusals432(t *testing.T) {
	for _, mode := range []string{"absent", "denied", "race"} {
		t.Run(mode, func(t *testing.T) {
			transport := &metricScopesTransport432{mode: mode}
			s := Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: transport}}
			result, err := s.DiscoverMetricScopes(context.Background(), metricScopeQuestion432)
			if err == nil || result != nil || sourcefailure.Describe(err).Refusal.Kind != sourcefailure.Unavailable {
				t.Fatal("failure became empty", result, err)
			}
			expected := 4
			if mode == "race" {
				expected = 6
			}
			if len(transport.calls) != expected {
				t.Fatal("unexpected source budget", transport.calls)
			}
		})
	}
	transport := &metricScopesTransport432{}
	s := Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: transport}}
	if _, err := s.DiscoverMetricScopes(context.Background(), MetricScopesQuestion{}); err == nil || len(transport.calls) != 0 {
		t.Fatal("invalid window reached source")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if r, err := s.DiscoverMetricScopes(ctx, metricScopeQuestion432); r != nil || !errors.Is(err, context.Canceled) || len(transport.calls) != 0 {
		t.Fatal("cancelled acquisition admitted", r, err)
	}
}
func TestMetricScopePolicyMatchesEffectiveEnvironment432(t *testing.T) {
	for _, resource := range []map[string]string{{}, {identity.EnvironmentKey: ""}, {identity.LegacyEnvironmentKey: "prod"}, {identity.EnvironmentKey: "", identity.LegacyEnvironmentKey: "prod"}, {identity.EnvironmentKey: "prod", identity.LegacyEnvironmentKey: "other"}} {
		observed := identity.ObserveEnvironment(resource)
		present, value := "0", ""
		if observed.Scope.Mode != "unreported" {
			present = "1"
			value = observed.Scope.Value
		}
		choices, err := metricScopeRows(metricScopeQuestion432, []json.RawMessage{metricScopeRow432(map[string]string{"environment_present": present, "environment": value})})
		if err != nil || choices[0].Environment != observed.Scope {
			t.Fatal("current/legacy precedence changed", choices, err)
		}
	}
}
