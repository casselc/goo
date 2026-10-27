package measurements

import (
	"encoding/json"
	"github.com/casselc/goo/metrics"
	"os"
	"strings"
	"testing"
	"time"
)

func containerCPUParent() PodMetricDiscoveryQuestion {
	q := podMetricQuestion()
	q.Resource["container.id"] = "containerd://literal-one"
	q.Resource["k8s.container.name"] = "worker"
	return q
}
func containerCPUTestResult(t *testing.T, parent PodMetricDiscoveryQuestion, role, value string) *Result {
	t.Helper()
	metric, scope := "container.cpu.usage", KubeletStatsScope
	if role != "usage" {
		scope = K8sClusterScope
		metric = "k8s.container.cpu_" + role
	}
	stream := metrics.Stream{Resource: cloneLabels(parent.Resource), ResourceSchemaURL: *parent.ResourceSchemaURL, Scope: scope, Version: "0.137.0", Name: metric, Unit: "{cpu}", Attributes: map[string]string{}}
	d := Definition{APIVersion: Version, Name: "cpu-" + role, Title: role, Kind: "gauge", Environment: parent.Environment, LookbackMinutes: 1, MaxGapMS: 60000, Grain: "per-stream", Source: Binding{Resource: &ResourceBinding{Kind: "kubernetes-container", ID: stream.Resource["container.id"]}, Services: []string{}, Scope: scope, Version: stream.Version, Metric: metric, Unit: "{cpu}", ExactStream: &stream}}
	q := Question{Definition: d, FromMS: parent.FromMS, ToMS: parent.ToMS}
	row := metricRow{Resource: stream.Resource, ResourceSchemaURL: stream.ResourceSchemaURL, Scope: scope, Version: stream.Version, Name: metric, Unit: "{cpu}", Attributes: stream.Attributes, EndNS: ns(120000), Value: value}
	result, e := Evaluate(q, []json.RawMessage{raw(row)}, time.Unix(200, 0), "", nil)
	if e != nil {
		t.Fatal(e)
	}
	return result
}
func containerCPUFixture(t *testing.T) ContainerCPUViewRequest {
	q := containerCPUParent()
	return ContainerCPUViewRequest{Parent: q, Usage: containerCPUTestResult(t, q, "usage", "0.125"), Request: containerCPUTestResult(t, q, "request", "0.25"), Limit: containerCPUTestResult(t, q, "limit", "0.5")}
}
func TestContainerCPUExactPercentagesAndIndependentEvidence(t *testing.T) {
	input := containerCPUFixture(t)
	before, _ := json.Marshal(input)
	view, e := ComposeContainerCPUView(input)
	if e != nil {
		t.Fatal(e)
	}
	if view.Percentages.Request.Value != "50" || view.Percentages.Request.Text != "50.0%" || view.Percentages.Limit.Value != "25" || view.Signals[0].Value != "0.125" {
		t.Fatal(view)
	}
	view.Parent.Resource["container.id"] = "mutated"
	view.Signals[0].Question.Definition.Source.ExactStream.Resource["k8s.pod.uid"] = "mutated"
	after, _ := json.Marshal(input)
	if string(before) != string(after) {
		t.Fatal("input evidence mutated")
	}
	zero := containerCPUFixture(t)
	zero.Usage = containerCPUTestResult(t, zero.Parent, "usage", "0")
	view, e = ComposeContainerCPUView(zero)
	if e != nil || view.Signals[0].Value != "0" || view.Percentages.Limit.Value != "0" {
		t.Fatal(view, e)
	}
	if file := os.Getenv("GOO_CONTAINER_CPU_FIXTURE"); file != "" {
		sample := containerCPUFixture(t)
		sample.Usage = containerCPUTestResult(t, sample.Parent, "usage", "0")
		result, e := ComposeContainerCPUView(sample)
		if e != nil {
			t.Fatal(e)
		}
		b, _ := json.MarshalIndent(struct {
			Request  ContainerCPUViewRequest `json:"request"`
			Response *ContainerCPUView       `json:"response"`
		}{sample, result}, "", "  ")
		if e = os.WriteFile(file, append(b, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	}
}
func TestContainerCPUAlignmentContextAndNoOlderFallback(t *testing.T) {
	input := containerCPUFixture(t)
	input.Request.Question.Definition.Source.ExactStream.Resource["cloud.account.id"] = "other"
	replacePoolRows(t, input.Request, func(rows []metricRow) []metricRow {
		rows[0].Resource = cloneLabels(input.Request.Question.Definition.Source.ExactStream.Resource)
		return rows
	})
	view, e := ComposeContainerCPUView(input)
	if e != nil || view.Signals[1].Value != "0.25" || view.Percentages.Request.Reason != "resource-or-attribute-mismatch" || view.Percentages.Limit.State != "observed" {
		t.Fatal("independent variant hidden or ratio inferred", view, e)
	}
	input = containerCPUFixture(t)
	replacePoolRows(t, input.Request, func(rows []metricRow) []metricRow { rows[0].EndNS = ns(120001); return rows })
	view, e = ComposeContainerCPUView(input)
	if e != nil || view.Percentages.Request.Reason != "observation-time-mismatch" {
		t.Fatal(view, e)
	}
	input = containerCPUFixture(t)
	input.Request.Question.Definition.MaxGapMS = 1
	replacePoolRows(t, input.Request, func(rows []metricRow) []metricRow { return rows })
	view, e = ComposeContainerCPUView(input)
	if e != nil || view.Signals[1].State != "stale" || view.Percentages.Request.State != "unavailable" {
		t.Fatal(view, e)
	}
	input = containerCPUFixture(t)
	replacePoolRows(t, input.Usage, func(rows []metricRow) []metricRow {
		latest := rows[0]
		latest.EndNS = ns(121000)
		latest.Flags = 1
		return append(rows, latest)
	})
	view, e = ComposeContainerCPUView(input)
	if e != nil || view.Signals[0].State != "unknown" || view.Signals[0].Value != "" || view.Signals[0].AtNS != ns(121000) {
		t.Fatal("older fallback", view, e)
	}
	input = containerCPUFixture(t)
	input.Request.Question.Definition.Source.ExactStream.Resource["k8s.cluster.uid"] = ""
	replacePoolRows(t, input.Request, func(rows []metricRow) []metricRow {
		rows[0].Resource = cloneLabels(input.Request.Question.Definition.Source.ExactStream.Resource)
		return rows
	})
	if _, e = ComposeContainerCPUView(input); e == nil {
		t.Fatal("missing/empty cluster qualification merged")
	}
}
func TestContainerCPUReportedDecimalDomain(t *testing.T) {
	for _, tc := range []struct{ raw, want string }{{"0.125", "0.125"}, {"1.00000000000000001", "1.00000000000000001"}, {"1e-300", "0." + strings.Repeat("0", 299) + "1"}, {"-0.000", "0"}, {"1.2500e+1", "12.5"}, {"1e300", "1" + strings.Repeat("0", 300)}} {
		value, ok := exactContainerCPU(tc.raw)
		if !ok || containerCPUDecimal(value) != tc.want {
			t.Fatal(tc.raw, value, ok)
		}
	}
	for _, raw := range []string{"-0.01", "NaN", "Inf", "1e400", "1e-400", "0x1", "1/2", "1_000", "1\n", strings.Repeat("1", 129), "1e-401"} {
		if _, ok := exactContainerCPU(raw); ok {
			t.Fatal("invalid decimal admitted", raw)
		}
	}
	input := containerCPUFixture(t)
	input.Usage = containerCPUTestResult(t, input.Parent, "usage", "1.00000000000000001")
	view, e := ComposeContainerCPUView(input)
	if e != nil || view.Signals[0].Value != "1.00000000000000001" || view.Percentages.Request.Value != "100000000000000001/250000000000000" {
		t.Fatal(view, e)
	}
	input.Usage = containerCPUTestResult(t, input.Parent, "usage", "-0.25")
	view, e = ComposeContainerCPUView(input)
	if e != nil || view.Signals[0].State != "invalid" || view.Percentages.Request.Reason != "invalid-core-domain" {
		t.Fatal(view, e)
	}
	input.Usage = containerCPUTestResult(t, input.Parent, "usage", "0")
	input.Limit = containerCPUTestResult(t, input.Parent, "limit", "0")
	view, e = ComposeContainerCPUView(input)
	if e != nil || view.Signals[0].Value != "0" || view.Percentages.Limit.Reason != "zero-denominator" {
		t.Fatal(view, e)
	}
}
func TestContainerCPUExplicitMetadataAndNoAliases(t *testing.T) {
	q := containerCPUParent()
	rows := []json.RawMessage{}
	for _, role := range []string{"usage", "request", "limit"} {
		result := containerCPUTestResult(t, q, role, "0.125")
		d := result.Question.Definition
		var meta map[string]any
		json.Unmarshal(podMetadata(q, d.Source.Metric, d.Source.Scope, d.Source.Version, d.Source.SchemaURL), &meta)
		meta["unit"] = "{cpu}"
		b, _ := json.Marshal(meta)
		rows = append(rows, b)
	}
	discovery, e := discoverPodMetricRows(q, rows)
	if e != nil {
		t.Fatal(e)
	}
	plan, e := PlanPodDashboard(PodDashboardRequest{Discovery: *discovery})
	if e != nil || len(plan.Candidates) != 3 || len(plan.Widgets) != 0 {
		t.Fatal(plan, e)
	}
	for _, candidate := range plan.Candidates {
		if candidate.Role == "" || candidate.Question.Definition.Source.Resource.Kind != "kubernetes-container" {
			t.Fatal(candidate)
		}
	}
	reversed := *discovery
	reversed.Instruments = append([]Instrument{}, discovery.Instruments...)
	for a, b := 0, len(reversed.Instruments)-1; a < b; a, b = a+1, b-1 {
		reversed.Instruments[a], reversed.Instruments[b] = reversed.Instruments[b], reversed.Instruments[a]
	}
	reverse, e := PlanPodDashboard(PodDashboardRequest{Discovery: reversed})
	a, _ := json.Marshal(plan)
	b, _ := json.Marshal(reverse)
	if e != nil || string(a) != string(b) {
		t.Fatal("order leaked", e)
	}
	i := Instrument{Kind: "gauge", Metric: "container.cpu.usage", Unit: "{cpu}", Scope: KubeletStatsScope, Version: "0.137.0"}
	if ContainerCPURole(i) != "usage" {
		t.Fatal(i)
	}
	for _, edit := range []func(*Instrument){func(i *Instrument) { i.Unit = "s" }, func(i *Instrument) { i.Version = "v0.137.0" }, func(i *Instrument) { i.SchemaURL = "other" }, func(i *Instrument) { i.Metric = "container.cpu.time" }, func(i *Instrument) { i.Kind = "counter" }} {
		bad := i
		edit(&bad)
		if ContainerCPURole(bad) != "" {
			t.Fatal("alias admitted", bad)
		}
	}
	delete(q.Resource, "container.id")
	discovery, e = discoverPodMetricRows(q, rows)
	if e == nil && len(discovery.Instruments) > 0 {
		for _, i := range discovery.Instruments {
			if i.Supported {
				t.Fatal("container ID invented")
			}
		}
	}
}

func TestContainerCPUCompactExtremePercentagesAndRawQuality(t *testing.T) {
	input := containerCPUFixture(t)
	input.Usage = containerCPUTestResult(t, input.Parent, "usage", "1e-300")
	view, e := ComposeContainerCPUView(input)
	if e != nil || view.Signals[0].Value == "0" || view.Percentages.Request.Text != "<0.1%" {
		t.Fatal(view, e)
	}
	input.Usage = containerCPUTestResult(t, input.Parent, "usage", "1e300")
	input.Request = containerCPUTestResult(t, input.Parent, "request", "1e-300")
	view, e = ComposeContainerCPUView(input)
	if e != nil || view.Percentages.Request.State != "observed" || len(view.Percentages.Request.Text) > 96 || !strings.Contains(view.Percentages.Request.Text, "e+") {
		t.Fatal(view, e)
	}
	input = containerCPUFixture(t)
	replacePoolRows(t, input.Usage, func(rows []metricRow) []metricRow { row := rows[0]; row.Value = "0.25"; return append(rows, row) })
	view, e = ComposeContainerCPUView(input)
	if e != nil || view.Signals[0].State != "unknown" || view.Signals[0].Value != "" {
		t.Fatal("conflict hidden", view, e)
	}
	input = containerCPUFixture(t)
	input.Usage.Question.FromNS = "100000000001"
	input.Usage.Question.ToNS = ns(input.Parent.ToMS)
	replacePoolRows(t, input.Usage, func(rows []metricRow) []metricRow { return rows })
	if _, e = ComposeContainerCPUView(input); e == nil {
		t.Fatal("different nanosecond window admitted")
	}
}
