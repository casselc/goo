package measurements

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func hostTestQuestion() HostDiscoveryQuestion {
	return HostDiscoveryQuestion{Environment: "prod", FromMS: 100000, ToMS: 104000}
}
func hostTestRow(metric, kind, unit string, attrs map[string]string, resource map[string]string) json.RawMessage {
	if resource == nil {
		resource = map[string]string{"host.id": "host-a", "host.name": "display-only", "deployment.environment.name": "prod"}
	}
	row := map[string]any{"kind": kind, "scope": "host-sdk", "version": "", "schemaURL": HostMetricsSchema, "metric": metric, "unit": unit, "temporality": 0, "monotonic": false, "resource": resource, "resourceSchemaURL": "resource-schema", "attributes": attrs}
	if kind == "counter" {
		row["temporality"] = 2
	}
	if metric == "system.cpu.time" {
		row["monotonic"] = true
	}
	raw, _ := json.Marshal(row)
	return raw
}
func hostTestRows() []json.RawMessage {
	return []json.RawMessage{
		hostTestRow("system.cpu.utilization", "gauge", "1", map[string]string{"cpu.mode": "system", "cpu.logical_number": "10"}, nil),
		hostTestRow("system.cpu.utilization", "gauge", "1", map[string]string{"cpu.mode": "user", "cpu.logical_number": "2"}, nil),
		hostTestRow("system.cpu.utilization", "gauge", "1", map[string]string{"cpu.mode": "user", "cpu.logical_number": "10"}, nil),
		hostTestRow("system.memory.usage", "counter", "By", map[string]string{"system.memory.state": "free"}, nil),
		hostTestRow("system.memory.usage", "counter", "By", map[string]string{"system.memory.state": "used"}, nil),
		hostTestRow("system.cpu.time", "counter", "s", map[string]string{"cpu.mode": "user"}, nil),
	}
}
func TestHostDiscoveryRawIdentityAndCandidateDefaults(t *testing.T) {
	result, e := discoverHostRows(hostTestQuestion(), hostTestRows())
	if e != nil || len(result.Hosts) != 1 {
		t.Fatal(result, e)
	}
	host := result.Hosts[0]
	if _, ok := host.Resource["service.name"]; ok {
		t.Fatal("manufactured service")
	}
	if _, ok := host.Resource["service.instance.id"]; ok {
		t.Fatal("manufactured instance")
	}
	before, _ := json.Marshal(result)
	plan, e := PlanHostDashboard(HostDashboardRequest{Question: result.Question, Host: host})
	if e != nil || len(plan.Widgets) != 2 || len(plan.Candidates) != 6 {
		t.Fatal(plan, e)
	}
	haveCPU, haveMemory := false, false
	for _, w := range plan.Widgets {
		d := w.Question.Definition
		s := d.Source.ExactStream
		if d.Source.Resource == nil || d.Source.Resource.ID != "host-a" || len(d.Source.Services) != 0 || d.Source.Namespace != nil || s == nil {
			t.Fatal("host question service fallback", d)
		}
		if e := w.Question.Validate(); e != nil {
			t.Fatal(e)
		}
		if d.Source.Metric == "system.cpu.utilization" {
			haveCPU = true
			if s.Attributes["cpu.mode"] != "user" || s.Attributes["cpu.logical_number"] != "2" {
				t.Fatal("CPU rank is not numeric user/core", s)
			}
		}
		if d.Source.Metric == "system.memory.usage" {
			haveMemory = true
			if d.Kind != "stock" || d.StockDomain != "nonnegative-integer" || s.Attributes["system.memory.state"] != "used" {
				t.Fatal("memory not exact used stock", d)
			}
		}
	}
	if !haveCPU || !haveMemory {
		t.Fatal(plan.Widgets)
	}
	after, _ := json.Marshal(result)
	if string(before) != string(after) {
		t.Fatal("planning mutated retained inventory")
	}
	for _, i := range host.Instruments {
		if i.Metric == "system.cpu.time" && (!i.Supported || i.Definition == nil || i.Definition.CounterDomain != "nonnegative-decimal") {
			t.Fatal("pinned CPU-time domain missing", i)
		}
	}
}
func TestHostDiscoveryResourceVariantsAndMissingID(t *testing.T) {
	rows := []json.RawMessage{}
	for _, name := range []string{"one", "two"} {
		rows = append(rows, hostTestRow("system.cpu.utilization", "gauge", "1", nil, map[string]string{"host.id": "same-id", "host.name": name, "deployment.environment.name": "prod"}))
	}
	rows = append(rows, hostTestRow("system.cpu.utilization", "gauge", "1", nil, map[string]string{"host.name": "not-an-id", "deployment.environment.name": "prod"}))
	result, e := discoverHostRows(hostTestQuestion(), rows)
	if e != nil || len(result.Hosts) != 2 || result.MissingHostIDRecords != 1 || result.Hosts[0].Key == result.Hosts[1].Key {
		t.Fatal(result, e)
	}
	for _, h := range result.Hosts {
		if h.Resource["host.id"] != "same-id" {
			t.Fatal(h)
		}
	}
}
func TestHostDiscoverySingleGlobalReadAndFailClosed(t *testing.T) {
	reads := 0
	payload := string(hostTestRows()[0])
	source := Source{Endpoint: "http://source", Client: &http.Client{Transport: countedValuesHostTransport(func(r *http.Request) (*http.Response, error) {
		reads++
		sql, _ := io.ReadAll(r.Body)
		s := string(sql)
		if strings.Count(s, "UNION ALL") != 2 || strings.Count(s, "LIMIT 10001") != 1 || strings.Contains(s, "ServiceName") || strings.Contains(s, "mapUpdate") || !strings.Contains(s, "ResourceAttributes AS resource") || !strings.Contains(s, "mapContains(ResourceAttributes,'host.id')") {
			t.Fatal("host inventory source scope", s)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(payload)), Header: make(http.Header)}, nil
	})}}
	result, e := source.DiscoverHosts(context.Background(), hostTestQuestion())
	if e != nil || reads != 1 || len(result.Hosts) != 1 {
		t.Fatal(result, e, reads)
	}
	payload = strings.Repeat(string(hostTestRows()[0])+"\n", RowLimit+1)
	if _, e := source.DiscoverHosts(context.Background(), hostTestQuestion()); e == nil {
		t.Fatal("partial host inventory published")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	before := reads
	if _, e := source.DiscoverHosts(ctx, hostTestQuestion()); e == nil || reads != before {
		t.Fatal("cancelled inventory source read")
	}
}

type countedValuesHostTransport func(*http.Request) (*http.Response, error)

func (f countedValuesHostTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestHostInventoryAndPlanOmissionsStayExplicit(t *testing.T) {
	rows := []json.RawMessage{}
	for i := 0; i < 102; i++ {
		rows = append(rows, hostTestRow("system.cpu.utilization", "gauge", "1", nil, map[string]string{"host.id": fmt.Sprint(i), "deployment.environment.name": "prod"}))
	}
	result, e := discoverHostRows(hostTestQuestion(), rows)
	if e != nil || len(result.Hosts) != 100 || result.HostsOmitted != 2 {
		t.Fatal(result, e)
	}
	rows = nil
	for core := 0; core < 30; core++ {
		rows = append(rows, hostTestRow("system.cpu.utilization", "gauge", "1", map[string]string{"cpu.mode": "user", "cpu.logical_number": fmt.Sprint(core)}, nil))
	}
	result, e = discoverHostRows(hostTestQuestion(), rows)
	if e != nil {
		t.Fatal(e)
	}
	host := result.Hosts[0]
	plan, e := PlanHostDashboard(HostDashboardRequest{Question: result.Question, Host: host})
	if e != nil || len(plan.Candidates) != 16 || len(plan.Widgets) != 1 || plan.StreamsOmitted != 5 || plan.CandidatesOmitted != 9 {
		t.Fatal(plan, e)
	}
	for _, bad := range []int{-1, RowLimit + 1} {
		copy := host
		raw, _ := json.Marshal(copy)
		_ = json.Unmarshal(raw, &copy)
		copy.Instruments[0].CapabilityPlan.Coverage.OmittedStreams = bad
		if _, e := PlanHostDashboard(HostDashboardRequest{Question: result.Question, Host: copy}); e == nil {
			t.Fatal("bad preview count accepted")
		}
	}
}
func TestHostDashboardMemoryFractionFallbackAndGaugeUsageRefusal(t *testing.T) {
	rows := []json.RawMessage{hostTestRow("system.memory.usage", "gauge", "By", map[string]string{"system.memory.state": "used"}, nil), hostTestRow("system.memory.utilization", "gauge", "1", map[string]string{"system.memory.state": "used"}, nil)}
	discovered, e := discoverHostRows(hostTestQuestion(), rows)
	if e != nil {
		t.Fatal(e)
	}
	plan, e := PlanHostDashboard(HostDashboardRequest{Question: discovered.Question, Host: discovered.Hosts[0]})
	if e != nil || len(plan.Widgets) != 1 || plan.Widgets[0].Question.Definition.Source.Metric != "system.memory.utilization" {
		t.Fatal(plan, e)
	}
	old := plan.Widgets[0].Question.Definition.Source.ExactStream.Resource
	if !reflect.DeepEqual(old, discovered.Hosts[0].Resource) {
		t.Fatal("resource identity broadened")
	}
}

func TestHostDashboardOpaqueCoreLabelsUseLiteralLexicalOrder(t *testing.T) {
	rows := []json.RawMessage{hostTestRow("system.cpu.utilization", "gauge", "1", map[string]string{"cpu.mode": "user", "cpu.logical_number": "core-z"}, nil), hostTestRow("system.cpu.utilization", "gauge", "1", map[string]string{"cpu.mode": "user", "cpu.logical_number": "core-a"}, nil)}
	discovered, e := discoverHostRows(hostTestQuestion(), rows)
	if e != nil {
		t.Fatal(e)
	}
	plan, e := PlanHostDashboard(HostDashboardRequest{Question: discovered.Question, Host: discovered.Hosts[0]})
	if e != nil || len(plan.Widgets) != 1 || plan.Widgets[0].Question.Definition.Source.ExactStream.Attributes["cpu.logical_number"] != "core-a" {
		t.Fatal(plan, e)
	}
}
