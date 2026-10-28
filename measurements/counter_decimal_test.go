package measurements

import (
	"encoding/json"
	"github.com/casselc/goo/metrics"
	"strings"
	"testing"
	"time"
)

func cpuTimeQuestion() Question {
	s := metrics.Stream{Resource: map[string]string{"host.id": "host", "deployment.environment.name": "prod"}, Scope: "hostmetrics", SchemaURL: HostMetricsSchema, Name: "system.cpu.time", Unit: "s", Attributes: map[string]string{"cpu.mode": "user", "cpu.logical_number": "0"}}
	return Question{Definition: Definition{APIVersion: Version, Name: "cpu-time", Title: "CPU time", Kind: "counter", CounterDomain: metrics.DecimalCounterDomain, Environment: "prod", LookbackMinutes: 2, MaxGapMS: 60000, Grain: "per-stream", Source: Binding{Resource: &ResourceBinding{Kind: "host", ID: "host"}, Services: []string{}, Scope: s.Scope, SchemaURL: s.SchemaURL, Metric: s.Name, Unit: s.Unit, ExactStream: &s}}, FromMS: 1000, ToMS: 3000}
}
func cpuTimeRow(q Question, end, value string) json.RawMessage {
	s := q.Definition.Source.ExactStream
	b, _ := json.Marshal(map[string]any{"resource": s.Resource, "attributes": s.Attributes, "resource_schema_url": s.ResourceSchemaURL, "scope": s.Scope, "version": s.Version, "schema_url": s.SchemaURL, "name": s.Name, "unit": s.Unit, "start_ns": "0", "end_ns": end, "value": value, "flags": 0, "temporality": 2, "monotonic": true})
	return b
}
func TestDecimalCounterSourceExactAndOriginalEvidence(t *testing.T) {
	q := cpuTimeQuestion()
	rows := []json.RawMessage{cpuTimeRow(q, "1000000000", "1.25e+0"), cpuTimeRow(q, "3000000000", "1.75")}
	r, e := Evaluate(q, rows, time.Unix(3, 0))
	if e != nil {
		t.Fatal(e)
	}
	p := r.Counters.Series[0].Intervals[0]
	if p.Delta != "0.5" || p.ExactPerSecond != "1/4" || *p.PerSecond != .25 {
		t.Fatal(p)
	}
	for _, key := range []string{"resource_schema_url", "version", "attributes", "flags", "temporality", "monotonic"} {
		for _, mode := range []string{"missing", "null"} {
			var row map[string]json.RawMessage
			_ = json.Unmarshal(rows[0], &row)
			if mode == "missing" {
				delete(row, key)
			} else {
				row[key] = json.RawMessage("null")
			}
			bad, _ := json.Marshal(row)
			if _, e := Evaluate(q, []json.RawMessage{bad, rows[1]}, time.Unix(3, 0)); e == nil {
				t.Fatal("accepted", key, mode)
			}
		}
	}
	for _, bad := range []string{strings.Replace(string(rows[0]), `"flags":0`, `"flags":1e0`, 1), strings.Replace(string(rows[0]), `"flags":0`, `"flags":0.000000000000000001`, 1), strings.Replace(string(rows[0]), `"cpu.mode":"user"`, `"cpu.mode":null`, 1), strings.Replace(string(rows[0]), `"temporality":2`, `"temporality":2.0000000000000000001`, 1), strings.Replace(string(rows[0]), `"temporality":2`, `"temporality":2e0`, 1)} {
		if _, e := Evaluate(q, []json.RawMessage{json.RawMessage(bad)}, time.Unix(3, 0)); e == nil {
			t.Fatal("accepted manufactured field", bad)
		}
	}
	q.Definition.CounterDomain = ""
	if q.Validate() == nil {
		t.Fatal("integer host authority broadened")
	}
}
func TestCPUTimeProfileFallbackAndNarrowBinding(t *testing.T) {
	q := HostDiscoveryQuestion{Environment: "prod", FromMS: 100000, ToMS: 104000}
	rows := []json.RawMessage{hostTestRow("system.cpu.time", "counter", "s", map[string]string{"cpu.mode": "user", "cpu.logical_number": "2"}, nil), hostTestRow("system.cpu.time", "counter", "s", map[string]string{"cpu.mode": "system", "cpu.logical_number": "1"}, nil)}
	inventory, e := discoverHostRows(q, rows)
	if e != nil {
		t.Fatal(e)
	}
	p, e := PlanHostDashboard(HostDashboardRequest{Question: q, Host: inventory.Hosts[0]})
	if e != nil || len(p.Widgets) != 1 || p.Widgets[0].Question.Definition.CounterDomain != metrics.DecimalCounterDomain || p.Widgets[0].Question.Definition.Source.ExactStream.Attributes["cpu.mode"] != "user" {
		t.Fatal(p, e)
	}
	i := Instrument{Kind: "counter", Metric: "system.cpu.time", Unit: "s", SchemaURL: HostMetricsSchema, Temporality: 2, Monotonic: true}
	if !HostCPUTimeInstrument(i) {
		t.Fatal("profile")
	}
	for _, edit := range []func(*Instrument){func(i *Instrument) { i.Unit = "ms" }, func(i *Instrument) { i.Temporality = 1 }, func(i *Instrument) { i.Monotonic = false }, func(i *Instrument) { i.SchemaURL = "" }, func(i *Instrument) { i.Metric = "container.cpu.time" }} {
		x := i
		edit(&x)
		if HostCPUTimeInstrument(x) {
			t.Fatal("broadened profile", x)
		}
	}
}
