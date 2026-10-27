package measurements

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/metrics"
	"gopkg.in/yaml.v3"
)

func containerReadyQuestion(t *testing.T) Question {
	q := podPhaseQuestion(t)
	d := &q.Definition
	d.StateProfile = metrics.ContainerReadyProfile
	d.Source.Resource = &ResourceBinding{Kind: "kubernetes-container", ID: "reported-container"}
	d.Source.Metric = "k8s.container.ready"
	d.Source.ExactStream.Name = d.Source.Metric
	d.Source.ExactStream.Resource["container.id"] = "reported-container"
	d.Source.ExactStream.Resource["k8s.container.name"] = "worker"
	return q
}

func TestContainerReadinessCompileReplayAndAuthoring(t *testing.T) {
	q := containerReadyQuestion(t)
	plan, err := Compile(q)
	if err != nil || !strings.Contains(plan.SQL, "FROM otel_metrics_gauge") || !strings.Contains(plan.SQL, "ResourceAttributes['container.id']") || !strings.Contains(plan.SQL, "arraySort(arrayZip(mapKeys(ResourceAttributes)") || strings.Contains(plan.SQL, "ServiceName") {
		t.Fatal(plan, err)
	}
	for _, value := range []string{"0", "1", "0.5", "2"} {
		r, err := Evaluate(q, []json.RawMessage{raw(hostResourceRow(q, value))}, time.Unix(200, 0), "", nil)
		if err != nil || ValidateResult(r) != nil {
			t.Fatal("readiness replay", value, err)
		}
		s := r.States.Series[0].Samples[0]
		if value == "0" && (s.Label != "Not ready" || r.State != "observed") {
			t.Fatal("reported false became missing")
		}
		if value == "1" && (s.Label != "Ready" || r.State != "observed") {
			t.Fatal("reported true missing")
		}
		if value == "2" && (s.CodeState != "unsupported" || r.State != "unknown") {
			t.Fatal("unknown code became readiness")
		}
		r.States.Series[0].Samples[0].Label = "Healthy pod"
		if ValidateResult(r) == nil {
			t.Fatal("invented health accepted")
		}
	}
	encoded, _ := yaml.Marshal(q.Definition)
	d, err := Read(string(encoded))
	if err != nil || Hash(d) != Hash(q.Definition) {
		t.Fatal("authoring identity drift", err)
	}
}

func TestContainerReadinessRefusesBroadeningAndPrecisionLoss(t *testing.T) {
	for _, edit := range []func(*Question){
		func(q *Question) { q.Definition.Source.Resource.Kind = "kubernetes-pod" },
		func(q *Question) { q.Definition.Source.Resource.ID = "replacement-id" },
		func(q *Question) { q.Definition.Source.ExactStream = nil },
		func(q *Question) {
			q.Definition.Kind = "gauge"
			// Scalar container gauges are now explicit authority, but a state
			// profile cannot be carried into a scalar package as an alias.
			q.Definition.Source.Unit = "1"
			q.Definition.Source.ExactStream.Unit = "1"
		},
		func(q *Question) { delete(q.Definition.Source.ExactStream.Resource, "k8s.container.name") },
		func(q *Question) { delete(q.Definition.Source.ExactStream.Resource, "k8s.pod.uid") },
		func(q *Question) { q.Definition.StateProfile = metrics.PodPhaseProfile },
	} {
		q := containerReadyQuestion(t)
		edit(&q)
		if q.Validate() == nil {
			t.Fatal("container authority broadened")
		}
	}
	q := containerReadyQuestion(t)
	for _, value := range []string{"1e-400", "1e-10000", "0.9999999999999999999", "1.0000000000000000001"} {
		if _, err := Evaluate(q, []json.RawMessage{raw(hostResourceRow(q, value))}, time.Unix(200, 0), "", nil); err == nil {
			t.Fatal("precision manufactured readiness", value)
		}
	}
	for _, key := range []string{"container.id", "k8s.pod.uid", "k8s.container.name"} {
		row := hostResourceRow(q, "1")
		row.Resource = cloneLabels(row.Resource)
		row.Resource[key] = "different"
		if _, err := Evaluate(q, []json.RawMessage{raw(row)}, time.Unix(200, 0), "", nil); err == nil {
			t.Fatal("raw resource broadened", key)
		}
	}
}

func TestContainerReadinessUsesVerbatimReportedID(t *testing.T) {
	q := containerReadyQuestion(t)
	const reported = "containerd://literal-reported-id"
	q.Definition.Source.Resource.ID = reported
	q.Definition.Source.ExactStream.Resource["container.id"] = reported
	r, err := Evaluate(q, []json.RawMessage{raw(hostResourceRow(q, "1"))}, time.Unix(200, 0), "", nil)
	if err != nil || r.States.Series[0].Stream.Resource["container.id"] != reported {
		t.Fatal("reported ID was normalized", err)
	}
	plan, err := Compile(q)
	found := false
	for _, value := range plan.Parameters {
		if value == reported {
			found = true
		}
	}
	if err != nil || !found {
		t.Fatal("literal container ID not parameterized", err)
	}
}

func TestContainerReadinessRequiresReportedMapsAndFlags(t *testing.T) {
	q := containerReadyQuestion(t)
	for _, key := range []string{"resource", "attributes", "flags"} {
		for _, null := range []bool{false, true} {
			var row map[string]json.RawMessage
			_ = json.Unmarshal(raw(hostResourceRow(q, "0")), &row)
			if null {
				row[key] = json.RawMessage("null")
			} else {
				delete(row, key)
			}
			if _, err := Evaluate(q, []json.RawMessage{raw(row)}, time.Unix(200, 0), "", nil); err == nil {
				t.Fatal("unknown source field became reported false", key)
			}
		}
	}
	for _, flags := range []json.RawMessage{json.RawMessage("1e-400"), json.RawMessage("0.0000000000000000001"), json.RawMessage("0.0")} {
		var row map[string]json.RawMessage
		_ = json.Unmarshal(raw(hostResourceRow(q, "0")), &row)
		row["flags"] = flags
		if _, err := Evaluate(q, []json.RawMessage{raw(row)}, time.Unix(200, 0), "", nil); err == nil {
			t.Fatal("noninteger flags rounded into known flags")
		}
	}
	q.Definition.Source.ExactStream.Attributes = map[string]string{"dimension": ""}
	var row map[string]json.RawMessage
	_ = json.Unmarshal(raw(hostResourceRow(q, "0")), &row)
	row["attributes"] = json.RawMessage(`{"dimension":null}`)
	if _, err := Evaluate(q, []json.RawMessage{raw(row)}, time.Unix(200, 0), "", nil); err == nil {
		t.Fatal("null dimension became known empty")
	}
}
