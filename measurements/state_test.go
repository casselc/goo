package measurements

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/metrics"
	"gopkg.in/yaml.v3"
)

func podPhaseQuestion(t *testing.T) Question {
	q := hostResourceQuestion(t, "gauge")
	d := &q.Definition
	d.Kind = "state"
	d.MaxGapMS = 60000
	d.StateProfile = metrics.PodPhaseProfile
	d.Source.Resource = &ResourceBinding{Kind: "kubernetes-pod", ID: "pod-a"}
	d.Source.Scope = "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver"
	d.Source.Version = "0.137.0"
	d.Source.SchemaURL = ""
	d.Source.Metric = "k8s.pod.phase"
	d.Source.Unit = ""
	d.Source.ExactStream = &metrics.Stream{Resource: map[string]string{"k8s.pod.uid": "pod-a", "deployment.environment.name": d.Environment}, Attributes: map[string]string{}, Scope: d.Source.Scope, Version: d.Source.Version, Name: d.Source.Metric}
	return q
}

func TestPodPhaseDefinitionSourceAndFrozenReplay(t *testing.T) {
	q := podPhaseQuestion(t)
	plan, err := Compile(q)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(plan.SQL, "FROM otel_metrics_gauge") || !strings.Contains(plan.SQL, "ResourceAttributes AS resource") || strings.Contains(plan.SQL, "ServiceName") || !strings.Contains(plan.SQL, "ResourceSchemaUrl=") || !strings.Contains(plan.SQL, "arraySort(arrayZip(mapKeys(ResourceAttributes)") || !strings.Contains(plan.SQL, "MetricUnit=") {
		t.Fatal(plan.SQL)
	}
	for _, code := range []string{"1", "2", "3", "4", "5", "0", "1.5", "6"} {
		r, err := Evaluate(q, []json.RawMessage{raw(hostResourceRow(q, code))}, time.Unix(200, 0))
		if err != nil || ValidateResult(r) != nil {
			t.Fatalf("code %s: %v", code, err)
		}
		if r.States == nil || r.Gauges != nil || r.Stocks != nil || r.Counters != nil || r.Histograms != nil {
			t.Fatal("wrong kernel")
		}
		s := r.States.Series[0].Samples[0]
		if code == "5" && (s.State != "observed" || s.Label != "Unknown" || r.State != "observed") {
			t.Fatal("domain unknown became missing")
		}
		if code == "6" && (s.CodeState != "unsupported" || s.Label != "" || r.State != "unknown") {
			t.Fatal("unsupported code became phase")
		}
		saved, _ := json.Marshal(r)
		var reopened Result
		if json.Unmarshal(saved, &reopened) != nil || ValidateResult(&reopened) != nil {
			t.Fatal("offline replay failed")
		}
		reopened.States.Series[0].Samples[0].Label = "Healthy"
		if ValidateResult(&reopened) == nil {
			t.Fatal("invented health label accepted")
		}
	}
	encoded, _ := yaml.Marshal(q.Definition)
	d, err := Read(string(encoded))
	if err != nil || Hash(d) != Hash(q.Definition) {
		t.Fatal("YAML phase identity drift", err)
	}
}

func TestPodPhaseDefinitionAndRawRefusals(t *testing.T) {
	for _, edit := range []func(*Definition){
		func(d *Definition) { d.StateProfile = "" }, func(d *Definition) { d.StateProfile = "future" },
		func(d *Definition) { d.Kind = "gauge" }, func(d *Definition) { d.Source.Resource = nil },
		func(d *Definition) { d.Source.Resource.Kind = "host" }, func(d *Definition) { d.Source.ExactStream = nil },
		func(d *Definition) { d.Source.Services = []string{"fake"} }, func(d *Definition) { d.Source.Unit = "1" },
		func(d *Definition) { d.Source.ExactStream.Version = "future" }, func(d *Definition) { d.Source.Metric = "k8s.container.ready" },
	} {
		q := podPhaseQuestion(t)
		edit(&q.Definition)
		if q.Validate() == nil {
			t.Fatal("invalid state definition admitted")
		}
	}
	q := podPhaseQuestion(t)
	for _, key := range []string{"resource_schema_url", "scope", "version", "schema_url", "name", "unit"} {
		for _, value := range []json.RawMessage{nil, json.RawMessage("null")} {
			var row map[string]json.RawMessage
			_ = json.Unmarshal(raw(hostResourceRow(q, "2")), &row)
			if value == nil {
				delete(row, key)
			} else {
				row[key] = value
			}
			b, _ := json.Marshal(row)
			if _, err := Evaluate(q, []json.RawMessage{b}, time.Unix(200, 0)); err == nil {
				t.Fatalf("unknown %s admitted", key)
			}
		}
	}
	for _, value := range []string{"1.0000000000000000001", "4.9999999999999999999", "0x1p0"} {
		if _, err := Evaluate(q, []json.RawMessage{raw(hostResourceRow(q, value))}, time.Unix(200, 0)); err == nil {
			t.Fatal("raw value rounded into a phase", value)
		}
	}
	row := hostResourceRow(q, "2")
	row.ResourceSchemaURL = "other"
	if _, err := Evaluate(q, []json.RawMessage{raw(row)}, time.Unix(200, 0)); err == nil {
		t.Fatal("resource schema broadened")
	}
	row = hostResourceRow(q, "2")
	row.Flags = 1
	r, err := Evaluate(q, []json.RawMessage{raw(row)}, time.Unix(200, 0))
	if err != nil || r.States.Series[0].Samples[0].Code != nil || r.State != "unknown" {
		t.Fatal("missing became phase", err)
	}
}
