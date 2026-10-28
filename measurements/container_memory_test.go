package measurements

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/metrics"
)

func containerMemoryParent() PodMetricDiscoveryQuestion {
	q := podMetricQuestion()
	q.Resource["container.id"] = "containerd://literal-one"
	q.Resource["k8s.container.name"] = "worker"
	return q
}
func containerMemoryTestResult(t *testing.T, parent PodMetricDiscoveryQuestion, role, value string) *Result {
	t.Helper()
	metric, scope := "container.memory.working_set", KubeletStatsScope
	if role != "working-set" {
		scope = K8sClusterScope
		metric = "k8s.container.memory_" + role
	}
	stream := metrics.Stream{Resource: cloneLabels(parent.Resource), ResourceSchemaURL: *parent.ResourceSchemaURL, Scope: scope, Version: "0.137.0", Name: metric, Unit: "By", Attributes: map[string]string{}}
	d := Definition{APIVersion: Version, Name: "memory-" + role, Title: role, Kind: "gauge", Environment: parent.Environment, LookbackMinutes: 1, MaxGapMS: 60000, Grain: "per-stream", Source: Binding{Resource: &ResourceBinding{Kind: "kubernetes-container", ID: stream.Resource["container.id"]}, Services: []string{}, Scope: scope, Version: stream.Version, Metric: metric, Unit: "By", ExactStream: &stream}}
	q := Question{Definition: d, FromMS: parent.FromMS, ToMS: parent.ToMS}
	row := metricRow{Resource: stream.Resource, ResourceSchemaURL: stream.ResourceSchemaURL, Scope: scope, Version: stream.Version, Name: metric, Unit: "By", Attributes: stream.Attributes, EndNS: ns(120000), Value: value}
	result, e := Evaluate(q, []json.RawMessage{raw(row)}, time.Unix(200, 0))
	if e != nil {
		t.Fatal(e)
	}
	return result
}
func containerMemoryFixture(t *testing.T) ContainerMemoryViewRequest {
	q := containerMemoryParent()
	return ContainerMemoryViewRequest{Parent: q, WorkingSet: containerMemoryTestResult(t, q, "working-set", "50"), Request: containerMemoryTestResult(t, q, "request", "100"), Limit: containerMemoryTestResult(t, q, "limit", "200")}
}
func TestContainerMemoryExactPercentagesAndIndependentEvidence(t *testing.T) {
	input := containerMemoryFixture(t)
	before, _ := json.Marshal(input)
	view, e := ComposeContainerMemoryView(input)
	if e != nil {
		t.Fatal(e)
	}
	if view.Percentages.Request.Value != "50" || view.Percentages.Request.Text != "50.0%" || view.Percentages.Limit.Value != "25" || view.Signals[0].Value != "50" {
		t.Fatal(view)
	}
	view.Parent.Resource["container.id"] = "mutated"
	view.Signals[0].Question.Definition.Source.ExactStream.Resource["k8s.pod.uid"] = "mutated"
	after, _ := json.Marshal(input)
	if string(before) != string(after) {
		t.Fatal("input evidence mutated")
	}
	zero := containerMemoryFixture(t)
	zero.WorkingSet = containerMemoryTestResult(t, zero.Parent, "working-set", "0")
	view, e = ComposeContainerMemoryView(zero)
	if e != nil || view.Signals[0].Value != "0" || view.Percentages.Limit.Value != "0" {
		t.Fatal(view, e)
	}
	if file := os.Getenv("GOO_CONTAINER_MEMORY_FIXTURE"); file != "" {
		sample := containerMemoryFixture(t)
		sample.WorkingSet = containerMemoryTestResult(t, sample.Parent, "working-set", "0")
		result, e := ComposeContainerMemoryView(sample)
		if e != nil {
			t.Fatal(e)
		}
		b, _ := json.MarshalIndent(struct {
			Request  ContainerMemoryViewRequest `json:"request"`
			Response *ContainerMemoryView       `json:"response"`
		}{sample, result}, "", "  ")
		if e = os.WriteFile(file, append(b, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	}
}
func TestContainerMemoryByteDomainAndMissingInputs(t *testing.T) {
	for _, value := range []string{"-1", "0.5", "9007199254740992", "1.00000000000000001", "9007199254740991.1"} {
		input := containerMemoryFixture(t)
		input.WorkingSet = containerMemoryTestResult(t, input.Parent, "working-set", value)
		before, _ := json.Marshal(input.WorkingSet.Rows)
		view, e := ComposeContainerMemoryView(input)
		if e != nil || view.Signals[0].State != "invalid" || view.Signals[0].Value != "" || view.Percentages.Limit.Reason != "invalid-byte-domain" {
			t.Fatal(value, view, e)
		}
		after, _ := json.Marshal(input.WorkingSet.Rows)
		if string(before) != string(after) {
			t.Fatal("invalid rawrows masked")
		}
	}
	input := containerMemoryFixture(t)
	input.Request = nil
	input.Limit = containerMemoryTestResult(t, input.Parent, "limit", "0")
	view, e := ComposeContainerMemoryView(input)
	if e != nil || view.Signals[1].State != "unavailable" || view.Signals[2].Value != "0" || view.Percentages.Limit.Reason != "zero-denominator" {
		t.Fatal(view, e)
	}
	input = containerMemoryFixture(t)
	q := input.Limit.Question
	r, e := Evaluate(q, nil, time.Unix(200, 0))
	if e != nil {
		t.Fatal(e)
	}
	input.Limit = r
	view, e = ComposeContainerMemoryView(input)
	if e != nil || view.Signals[2].State != "no_data" || view.Signals[2].Value != "" {
		t.Fatal(view, e)
	}
}
func TestContainerMemoryAlignmentContextAndNoOlderFallback(t *testing.T) {
	input := containerMemoryFixture(t)
	input.Request.Question.Definition.Source.ExactStream.Resource["cloud.account.id"] = "other"
	replacePoolRows(t, input.Request, func(rows []metricRow) []metricRow {
		rows[0].Resource = cloneLabels(input.Request.Question.Definition.Source.ExactStream.Resource)
		return rows
	})
	view, e := ComposeContainerMemoryView(input)
	if e != nil || view.Signals[1].Value != "100" || view.Percentages.Request.Reason != "resource-or-attribute-mismatch" || view.Percentages.Limit.State != "observed" {
		t.Fatal("independent variant hidden or ratio inferred", view, e)
	}
	input = containerMemoryFixture(t)
	replacePoolRows(t, input.Request, func(rows []metricRow) []metricRow { rows[0].EndNS = ns(120001); return rows })
	view, e = ComposeContainerMemoryView(input)
	if e != nil || view.Percentages.Request.Reason != "observation-time-mismatch" {
		t.Fatal(view, e)
	}
	input = containerMemoryFixture(t)
	input.Request.Question.Definition.MaxGapMS = 1
	replacePoolRows(t, input.Request, func(rows []metricRow) []metricRow { return rows })
	view, e = ComposeContainerMemoryView(input)
	if e != nil || view.Signals[1].State != "stale" || view.Percentages.Request.State != "unavailable" {
		t.Fatal(view, e)
	}
	input = containerMemoryFixture(t)
	replacePoolRows(t, input.WorkingSet, func(rows []metricRow) []metricRow {
		latest := rows[0]
		latest.EndNS = ns(121000)
		latest.Flags = 1
		return append(rows, latest)
	})
	view, e = ComposeContainerMemoryView(input)
	if e != nil || view.Signals[0].State != "unknown" || view.Signals[0].Value != "" || view.Signals[0].AtNS != ns(121000) {
		t.Fatal("older fallback", view, e)
	}
	input = containerMemoryFixture(t)
	input.Request.Question.Definition.Source.ExactStream.Resource["k8s.cluster.uid"] = ""
	replacePoolRows(t, input.Request, func(rows []metricRow) []metricRow {
		rows[0].Resource = cloneLabels(input.Request.Question.Definition.Source.ExactStream.Resource)
		return rows
	})
	if _, e = ComposeContainerMemoryView(input); e == nil {
		t.Fatal("missing/empty cluster qualification merged")
	}
}
func TestContainerMemoryMetadataAndStrictSourceAuthority(t *testing.T) {
	q := containerMemoryParent()
	rows := []json.RawMessage{}
	for _, role := range []string{"working-set", "request", "limit"} {
		r := containerMemoryTestResult(t, q, role, "1")
		d := r.Question.Definition
		rows = append(rows, podMetadata(q, d.Source.Metric, d.Source.Scope, d.Source.Version, d.Source.SchemaURL))
	}
	discovery, e := discoverPodMetricRows(q, rows)
	if e != nil {
		t.Fatal(e)
	}
	plan, e := PlanPodDashboard(PodDashboardRequest{Discovery: *discovery})
	if e != nil || len(plan.Widgets) != 0 || len(plan.Candidates) != 3 {
		t.Fatal("newroles added default reads", plan, e)
	}
	reversed := *discovery
	reversed.Instruments = append([]Instrument{}, discovery.Instruments...)
	for a, b := 0, len(reversed.Instruments)-1; a < b; a, b = a+1, b-1 {
		reversed.Instruments[a], reversed.Instruments[b] = reversed.Instruments[b], reversed.Instruments[a]
	}
	reordered, e := PlanPodDashboard(PodDashboardRequest{Discovery: reversed})
	if e != nil {
		t.Fatal(e)
	}
	a, _ := json.Marshal(plan)
	b, _ := json.Marshal(reordered)
	if string(a) != string(b) {
		t.Fatal("container role plan depends on metadata order")
	}
	roles := map[string]bool{}
	for _, candidate := range plan.Candidates {
		roles[candidate.Role] = true
		if candidate.Question.Definition.Source.Resource.Kind != "kubernetes-container" || candidate.Grain != "container-resource-stream" {
			t.Fatal(candidate)
		}
	}
	if !roles["working-set"] || !roles["request"] || !roles["limit"] {
		t.Fatal(roles)
	}
	missing := q
	missing.Resource = cloneLabels(q.Resource)
	delete(missing.Resource, "container.id")
	rawMissing := []json.RawMessage{podMetadata(missing, "container.memory.working_set", KubeletStatsScope, KubeletStatsVersion, "")}
	inventory, e := discoverPodMetricRows(missing, rawMissing)
	if e != nil || inventory.Instruments[0].Supported || inventory.Instruments[0].Definition != nil {
		t.Fatal("fabricated container ID", inventory, e)
	}
	r := containerMemoryTestResult(t, q, "working-set", "1")
	for _, badFlag := range []string{"null", "0.00000000000000001", "0.0"} {
		bad := strings.Replace(string(r.Rows[0]), `"flags":0`, `"flags":`+badFlag, 1)
		if _, e := Evaluate(r.Question, []json.RawMessage{json.RawMessage(bad)}, time.Unix(200, 0)); e == nil {
			t.Fatal("bad rawflag accepted", badFlag)
		}
	}
	input := containerMemoryFixture(t)
	input.Limit.Gauges.Series[0].Samples[0].Value = nil
	if _, e = ComposeContainerMemoryView(input); e == nil {
		t.Fatal("forged derived result accepted")
	}
}

func TestContainerMemoryBoundsAndDecimalSyntax(t *testing.T) {
	for _, value := range []string{"1_000", "0x10", "1/1", "1\n", "0e999999999999"} {
		if _, ok := exactContainerBytes(value); ok {
			t.Fatal("nondecimal source token certified", value)
		}
	}
	for _, key := range []string{"container.id", "k8s.container.name"} {
		parent := containerMemoryParent()
		parent.Resource[key] = strings.Repeat("x", 513)
		if _, e := ComposeContainerMemoryView(ContainerMemoryViewRequest{Parent: parent}); e == nil {
			t.Fatal("oversize parentcontext accepted", key)
		}
	}
	input := containerMemoryFixture(t)
	input.WorkingSet.Rows = []json.RawMessage{json.RawMessage(strings.Repeat(" ", 8<<20) + "{}")}
	if _, e := ComposeContainerMemoryView(input); e == nil {
		t.Fatal("oversize evidence accepted")
	}
	input = containerMemoryFixture(t)
	input.Request.Question.Definition.MaxGapMS = 60001
	replacePoolRows(t, input.Request, func(rows []metricRow) []metricRow { return rows })
	view, e := ComposeContainerMemoryView(input)
	if e != nil || view.Percentages.Request.Reason != "freshness-policy-mismatch" {
		t.Fatal(view, e)
	}
}
