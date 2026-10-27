package measurements

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/casselc/goo/metrics"
)

func containerDiscoveryQuestion() PodMetricDiscoveryQuestion {
	q := podMetricQuestion()
	q.Resource["container.id"] = "containerd://reported-literal"
	q.Resource["k8s.container.name"] = "worker"
	return q
}
func containerReadyMetadata(q PodMetricDiscoveryQuestion) map[string]any {
	row := podPhaseMetadata(q)
	row["metric"] = "k8s.container.ready"
	return row
}
func TestContainerReadinessDiscoveryAndPlan(t *testing.T) {
	q := containerDiscoveryQuestion()
	rows := []json.RawMessage{raw(containerReadyMetadata(q)), raw(podPhaseMetadata(q))}
	d, err := discoverPodMetricRows(q, rows)
	if err != nil {
		t.Fatal(err)
	}
	before, _ := json.Marshal(d)
	plan, err := PlanPodDashboard(PodDashboardRequest{Discovery: *d})
	if err != nil || len(plan.Widgets) != 2 || len(plan.Candidates) != 2 {
		t.Fatal(plan, err)
	}
	after, _ := json.Marshal(d)
	if string(before) != string(after) {
		t.Fatal("planner mutated discovery")
	}
	for _, w := range plan.Widgets {
		if w.Question.Definition.Source.Metric != "k8s.container.ready" {
			continue
		}
		def := w.Question.Definition
		if def.Kind != "state" || def.StateProfile != metrics.ContainerReadyProfile || def.Source.Resource.Kind != "kubernetes-container" || def.Source.Resource.ID != q.Resource["container.id"] || w.Grain != "container-resource-stream" || !reflect.DeepEqual(def.Source.ExactStream.Resource, q.Resource) {
			t.Fatal("container authority broadened", w)
		}
		result, err := Evaluate(w.Question, []json.RawMessage{raw(hostResourceRow(w.Question, "0"))}, time.Unix(200, 0), "", nil)
		if err != nil || ValidateResult(result) != nil || result.State != "observed" || result.States.Series[0].Samples[0].Label != "Not ready" {
			t.Fatal("zero state lost", result, err)
		}
	}
	rows = append(rows, podMetadata(q, "k8s.pod.memory.usage", KubeletStatsScope, KubeletStatsVersion, ""), podMetadata(q, "k8s.pod.memory.working_set", KubeletStatsScope, KubeletStatsVersion, ""))
	mixed, err := discoverPodMetricRows(q, rows)
	if err != nil {
		t.Fatal(err)
	}
	pair, err := PlanPodDashboard(PodDashboardRequest{Discovery: *mixed})
	if err != nil || len(pair.Widgets) != 2 || len(pair.Candidates) != 4 {
		t.Fatal(pair, err)
	}
	for _, w := range pair.Widgets {
		if w.Profile != PodReceiverProfile {
			t.Fatal("readiness added default reads")
		}
	}
	for a, b := 0, len(rows)-1; a < b; a, b = a+1, b-1 {
		rows[a], rows[b] = rows[b], rows[a]
	}
	reversed, _ := discoverPodMetricRows(q, rows)
	again, err := PlanPodDashboard(PodDashboardRequest{Discovery: *reversed})
	if err != nil || !reflect.DeepEqual(pair, again) {
		t.Fatal("order changes plan", err)
	}
}
func TestContainerReadinessDiscoveryRequiresReportedIdentity(t *testing.T) {
	for _, key := range []string{"container.id", "k8s.container.name"} {
		q := containerDiscoveryQuestion()
		delete(q.Resource, key)
		d, err := discoverPodMetricRows(q, []json.RawMessage{raw(containerReadyMetadata(q))})
		if err != nil {
			t.Fatal(err)
		}
		p, err := PlanPodDashboard(PodDashboardRequest{Discovery: *d})
		if err != nil || len(p.Candidates) != 0 || d.Instruments[0].Supported {
			t.Fatal("missing context inferred", key, p, err)
		}
	}
	for key, value := range map[string]any{"scope": KubeletStatsScope, "version": "0.138.0", "schemaURL": PodMetricsSchema, "unit": "1", "temporality": 2, "monotonic": true} {
		q := containerDiscoveryQuestion()
		row := containerReadyMetadata(q)
		row[key] = value
		d, err := discoverPodMetricRows(q, []json.RawMessage{raw(row)})
		if err != nil {
			t.Fatal(err)
		}
		p, err := PlanPodDashboard(PodDashboardRequest{Discovery: *d})
		if err != nil || len(p.Candidates) != 0 {
			t.Fatal("readiness alias admitted", key, p, err)
		}
	}
}
