package measurements

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/casselc/goo/metrics"
)

func podPhaseMetadata(q PodMetricDiscoveryQuestion) map[string]any {
	return map[string]any{"kind": "gauge", "scope": "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver", "version": "0.137.0", "schemaURL": "", "metric": "k8s.pod.phase", "unit": "", "temporality": 0, "monotonic": false, "resource": q.Resource, "resourceSchemaURL": *q.ResourceSchemaURL, "attributes": map[string]string{}}
}

func TestPodPhaseDiscoveryPlanningAndNativeExecution(t *testing.T) {
	q := podMetricQuestion()
	discovery, err := discoverPodMetricRows(q, []json.RawMessage{raw(podPhaseMetadata(q))})
	if err != nil || len(discovery.Instruments) != 1 {
		t.Fatal("phase discovery", err)
	}
	i := discovery.Instruments[0]
	if i.Kind != "gauge" || !i.Supported || i.Definition == nil || i.Definition.Kind != "state" || i.Definition.StateProfile != metrics.PodPhaseProfile || !i.CapabilityPlan.Eligible {
		t.Fatal("physical gauge confused with typed state", i)
	}
	plan, err := PlanPodDashboard(PodDashboardRequest{Discovery: *discovery})
	if err != nil || len(plan.Widgets) != 1 || len(plan.Candidates) != 1 || plan.Widgets[0].Profile != PodClusterProfile || plan.Widgets[0].Title != "Reported pod phase" {
		t.Fatal("phase-only defaults", plan, err)
	}
	native := plan.Widgets[0].Question
	if !reflect.DeepEqual(native.Definition.Source.ExactStream.Resource, q.Resource) || native.Definition.Source.ExactStream.ResourceSchemaURL != *q.ResourceSchemaURL || native.FromMS != q.FromMS || native.ToMS != q.ToMS {
		t.Fatal("phase authority broadened")
	}
	r, err := Evaluate(native, []json.RawMessage{raw(hostResourceRow(native, "5"))}, time.Unix(200, 0))
	if err != nil || ValidateResult(r) != nil || r.States == nil || r.States.Series[0].Samples[0].Label != "Unknown" || r.State != "observed" {
		t.Fatal("discovered phase execution", r, err)
	}
	// Memory remains the same independent default pair; state is explicit.
	rows := []json.RawMessage{raw(podPhaseMetadata(q)), podMetadata(q, "k8s.pod.memory.usage", KubeletStatsScope, KubeletStatsVersion, ""), podMetadata(q, "k8s.pod.memory.working_set", KubeletStatsScope, KubeletStatsVersion, "")}
	mixed, err := discoverPodMetricRows(q, rows)
	if err != nil {
		t.Fatal(err)
	}
	paired, err := PlanPodDashboard(PodDashboardRequest{Discovery: *mixed})
	if err != nil || len(paired.Widgets) != 2 || len(paired.Candidates) != 3 {
		t.Fatal("mixed defaults", paired, err)
	}
	for _, w := range paired.Widgets {
		if w.Question.Definition.Kind != "gauge" || w.Profile != PodReceiverProfile {
			t.Fatal("phase added a default source read")
		}
	}
	for a, b := 0, len(rows)-1; a < b; a, b = a+1, b-1 {
		rows[a], rows[b] = rows[b], rows[a]
	}
	reversed, _ := discoverPodMetricRows(q, rows)
	again, err := PlanPodDashboard(PodDashboardRequest{Discovery: *reversed})
	if err != nil || !reflect.DeepEqual(paired, again) {
		t.Fatal("metadata order changes plan", err)
	}
}

func TestPodPhaseDiscoveryRefusesAliasesAndWrongShape(t *testing.T) {
	q := podMetricQuestion()
	for key, value := range map[string]any{"scope": KubeletStatsScope, "version": "0.138.0", "schemaURL": PodMetricsSchema, "unit": "1", "metric": "k8s.container.ready", "temporality": 2, "monotonic": true} {
		row := podPhaseMetadata(q)
		row[key] = value
		discovery, err := discoverPodMetricRows(q, []json.RawMessage{raw(row)})
		if err != nil {
			t.Fatal(err)
		}
		plan, err := PlanPodDashboard(PodDashboardRequest{Discovery: *discovery})
		if err != nil || len(plan.Candidates) != 0 || len(plan.Widgets) != 0 {
			t.Fatalf("%s alias admitted: %v", key, err)
		}
	}
	for _, key := range []string{"schemaURL", "unit", "resourceSchemaURL"} {
		row := podPhaseMetadata(q)
		row[key] = nil
		if _, err := discoverPodMetricRows(q, []json.RawMessage{raw(row)}); err == nil {
			t.Fatal("unknown became known empty", key)
		}
	}
}

func TestPodPhaseCandidateCapDoesNotReplaceMemoryDefaults(t *testing.T) {
	q := podMetricQuestion()
	rows := []json.RawMessage{raw(podPhaseMetadata(q)), podMetadata(q, "k8s.pod.memory.working_set", KubeletStatsScope, KubeletStatsVersion, "")}
	for n := 0; n < 20; n++ {
		var row map[string]any
		_ = json.Unmarshal(podMetadata(q, "k8s.pod.memory.usage", KubeletStatsScope, KubeletStatsVersion, ""), &row)
		row["attributes"] = map[string]string{"reported.dimension": fmt.Sprint(n)}
		rows = append(rows, raw(row))
	}
	discovery, err := discoverPodMetricRows(q, rows)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := PlanPodDashboard(PodDashboardRequest{Discovery: *discovery})
	if err != nil || len(plan.Widgets) != 2 || len(plan.Candidates) != 16 || plan.CandidatesOmitted != 6 {
		t.Fatal("phase cap or defaults", plan, err)
	}
	for _, w := range plan.Candidates {
		if w.Question.Definition.Kind == "state" {
			t.Fatal("candidate cap policy unexpectedly reordered memory")
		}
	}
	for _, w := range plan.Widgets {
		if w.Profile != PodReceiverProfile {
			t.Fatal("phase replaced a memory default")
		}
	}
}
