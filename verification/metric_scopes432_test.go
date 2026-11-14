package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
)

type metricScopeDriver432 struct {
	bank            map[string][]json.RawMessage
	resource        map[string]string
	service, points string
	window          measurements.MetricScopesQuestion
	mutate          func()
	native          bool
	calls           int
}

func metricScopeSchema432(t *testing.T) map[string][]json.RawMessage {
	t.Helper()
	raw, err := os.ReadFile("../internal/sourcecap/testdata/schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var bank map[string][]json.RawMessage
	if err = json.Unmarshal(raw, &bank); err != nil {
		t.Fatal(err)
	}
	return bank
}
func (d *metricScopeDriver432) RoundTrip(r *http.Request) (*http.Response, error) {
	d.calls++
	raw, _ := io.ReadAll(r.Body)
	sql := string(raw)
	status := 200
	head := make(http.Header)
	var rows []json.RawMessage
	if d.native && d.mutate != nil {
		f := d.mutate
		d.mutate = nil
		f()
	}
	switch {
	case strings.HasPrefix(sql, "DESCRIBE TABLE "):
		if strings.Contains(sql, "otel_metrics_gauge ") {
			rows = d.bank["gauge"]
		} else {
			status = 404
			head.Set("X-ClickHouse-Exception-Code", "60")
		}
	case strings.HasSuffix(sql, " LIMIT 0 FORMAT JSONEachRow"):
	case strings.Contains(sql, "sum(toUInt128(n))"):
		environment := identity.ObserveEnvironment(d.resource).Scope
		present := "1"
		if environment.Mode == "unreported" {
			present = "0"
		}
		ns, nsp := "", "0"
		if value, ok := d.resource[identity.NamespaceKey]; ok {
			ns, nsp = value, "1"
		}
		row := map[string]string{"environment_present": present, "environment": environment.Value, "namespace_present": nsp, "namespace": ns, "service": d.service, "points": d.points, "first_ms": fmt.Sprint(d.window.FromMS + 1), "last_ms": fmt.Sprint(d.window.ToMS - 1)}
		encoded, _ := json.Marshal(row)
		rows = []json.RawMessage{encoded}
	case strings.Contains(sql, "SELECT 'gauge' AS kind"):
		resource := map[string]string{"service.name": d.service}
		for k, v := range d.resource {
			resource[k] = v
		}
		row := map[string]any{"kind": "gauge", "scope": "native432", "version": "", "schemaURL": "", "metric": "load", "unit": "1", "temporality": 0, "monotonic": false, "dimensions": []string{}, "resource": resource, "attributes": map[string]string{}, "resourceSchemaURL": ""}
		encoded, _ := json.Marshal(row)
		rows = []json.RawMessage{encoded}
	default:
		return nil, fmt.Errorf("unexpected scope transition source read")
	}
	var body strings.Builder
	for _, row := range rows {
		compact, _ := json.Marshal(row)
		body.Write(compact)
		body.WriteByte('\n')
	}
	return &http.Response{StatusCode: status, Header: head, Body: io.NopCloser(strings.NewReader(body.String()))}, nil
}
func metricScopeContinuation432(t interface{ Fatal(...any) }, driver *metricScopeDriver432) (*measurements.MetricScopesResult, *measurements.DiscoveryResult, *measurements.DiscoveryQuestion, bool) {
	source := measurements.Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: driver}}
	scopes, err := source.DiscoverMetricScopes(context.Background(), driver.window)
	if err != nil || len(scopes.Choices) != 1 {
		t.Fatal("scope discovery", scopes, err)
		return nil, nil, nil, false
	}
	choice := scopes.Choices[0]
	if !choice.Selectable {
		return scopes, nil, nil, false
	}
	wire := *choice.EnvironmentSelector
	original := measurements.DiscoveryQuestion{Service: choice.Service, Namespace: choice.Namespace.Clone(), EnvironmentSelector: &wire, FromMS: driver.window.FromMS, ToMS: driver.window.ToMS}
	copied, _ := json.Marshal(original)
	var expected measurements.DiscoveryQuestion
	_ = json.Unmarshal(copied, &expected)
	// Existing native discovery must freeze the staged caller-owned selector before
	// its first metadata/source read, just as it does for package handoffs.
	driver.native = true
	driver.mutate = func() { original.EnvironmentSelector.Scope.Mode = "broken"; original.Namespace.Mode = "broken" }
	native, err := source.Discover(context.Background(), original)
	if err != nil {
		t.Fatal("native discovery", err)
		return scopes, nil, nil, false
	}
	detached := native.Question.EnvironmentSelector != original.EnvironmentSelector && native.Question.Namespace != original.Namespace
	return scopes, native, &expected, detached
}
func TestHegelMetricScopeContinuation432(t *testing.T) {
	bank := metricScopeSchema432(t)
	hegel.Test(t, func(tc *hegel.T) {
		resource := map[string]string{}
		values := []string{"", "prod", "unreported", "雪", "prod' OR 1=1 --"}
		if hegel.Draw(tc, hegel.Booleans()) {
			resource[identity.EnvironmentKey] = hegel.Draw(tc, hegel.SampledFrom(values))
		}
		if hegel.Draw(tc, hegel.Booleans()) {
			resource[identity.LegacyEnvironmentKey] = hegel.Draw(tc, hegel.SampledFrom(values))
		}
		if hegel.Draw(tc, hegel.Booleans()) {
			resource[identity.NamespaceKey] = hegel.Draw(tc, hegel.SampledFrom(values))
		}
		service := hegel.Draw(tc, hegel.SampledFrom([]string{"worker", "雪-worker", "", strings.Repeat("s", 129)}))
		switch hegel.Draw(tc, hegel.SampledFrom([]int{0, 1, 2})) {
		case 1:
			resource[identity.EnvironmentKey] = strings.Repeat("e", 129)
		case 2:
			resource[identity.NamespaceKey] = strings.Repeat("n", 129)
		}
		points := hegel.Draw(tc, hegel.SampledFrom([]string{"1", "18446744073709551616", "340282366920938463463374607431768211455"}))
		from := hegel.Draw(tc, hegel.SampledFrom([]int64{100000, 200000, 10000000}))
		window := measurements.MetricScopesQuestion{FromMS: from, ToMS: from + 60000}
		driver := &metricScopeDriver432{bank: bank, resource: resource, service: service, points: points, window: window}
		scopes, native, expected, detached := metricScopeContinuation432(tc, driver)
		if scopes == nil {
			return
		}
		choice := scopes.Choices[0]
		environment := identity.ObserveEnvironment(resource).Scope
		namespace := identity.FromValue(nil)
		if value, ok := resource[identity.NamespaceKey]; ok {
			namespace = identity.FromValue(&value)
		}
		valid := environment.Validate() == nil && namespace.Validate() == nil && service != "" && len(service) <= 128
		if choice.Selectable != valid || choice.Points != points || scopes.Question != window || choice.Environment != environment || !choice.Namespace.Equal(namespace) {
			tc.Fatal("observed identity changed", choice, valid)
		}
		if !valid {
			if native != nil || expected != nil || choice.Reason == "" || driver.calls != 6 {
				tc.Fatal("unsupported identity escaped staging", choice, driver.calls)
			}
			return
		}
		if native == nil || !detached || !reflect.DeepEqual(native.Question, *expected) || native.Question.Validate() != nil || len(native.Instruments) != 1 || native.Records != 1 {
			tc.Fatal("native handoff changed window/scope or aliases", native, expected, detached)
		}
		selected, err := native.Question.EnvironmentSelection()
		if err != nil || !selected.Matches(resource) || !native.Question.Namespace.Matches(resource) {
			tc.Fatal("staged native population differs", err, native.Question)
		}
	})
}
func TestQuintMetricScopeContinuation432(t *testing.T) {
	quint, err := exec.LookPath("quint")
	if err != nil {
		t.Fatal("native metric-scope trace validation requires quint on PATH")
	}
	spec, err := filepath.Abs("model/metric_scopes432.qnt")
	if err != nil {
		t.Fatal(err)
	}
	manifest := "spec: " + spec + "\nmodule: metric_scopes432\nconstants: {INITIAL: '0'}\ninvariants: [safety]\nactions:\n  catalog:\n    expect: 'windowOK == {{.obs.windowOK}}'\n  choose:\n    expect: 'scopeOK == {{.obs.scopeOK}}'\n  native_read:\n    expect: 'scopeOK == {{.obs.scopeOK}} and detached == {{.obs.detached}}'\n"
	config := filepath.Join(t.TempDir(), "binding.yaml")
	if err = os.WriteFile(config, []byte(manifest), 0600); err != nil {
		t.Fatal(err)
	}
	bank := metricScopeSchema432(t)
	for _, scope := range []string{"named", "empty", "unreported"} {
		for _, corrupt := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/corrupt%t", scope, corrupt), func(t *testing.T) {
				resource := map[string]string{identity.NamespaceKey: "payments"}
				if scope != "unreported" {
					resource[identity.EnvironmentKey] = ""
					if scope == "named" {
						resource[identity.EnvironmentKey] = "prod"
						resource[identity.LegacyEnvironmentKey] = "other"
					}
				}
				window := measurements.MetricScopesQuestion{FromMS: 100000, ToMS: 160000}
				driver := &metricScopeDriver432{bank: bank, resource: resource, service: "worker", points: "18446744073709551616", window: window}
				scopes, native, expected, detached := metricScopeContinuation432(t, driver)
				if native == nil {
					t.Fatal("expected executable scope")
				}
				sink := &qobs.MemorySink{}
				rec := qobs.NewRecorder(sink)
				rec.Record(context.Background(), "catalog", nil, "obs.windowOK", scopes.Question == window)
				choice := scopes.Choices[0]
				selected, err := choice.EnvironmentSelector.Selection()
				if err != nil {
					t.Fatal(err)
				}
				rec.Record(context.Background(), "choose", nil, "obs.scopeOK", choice.Selectable && selected.Matches(resource) && choice.Namespace.Matches(resource))
				actual := native.Question
				if corrupt {
					actual.EnvironmentSelector.Policy = "current-only"
				}
				rec.Record(context.Background(), "native_read", nil, "obs.scopeOK", reflect.DeepEqual(actual, *expected), "obs.detached", detached)
				checker, err := validate.NewChecker(config, validate.Options{Quint: quint, Backend: "typescript", Dir: t.TempDir()})
				if err != nil {
					t.Fatal(err)
				}
				result, err := checker.Check(context.Background(), sink.Steps())
				if err != nil {
					t.Fatal(err)
				}
				if result.OK() == corrupt {
					t.Fatal("scope handoff trace acceptance changed", corrupt, result.OK())
				}
			})
		}
	}
}
