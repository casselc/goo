package measurements

import (
	"encoding/json"
	"github.com/casselc/goo/metrics"
	"strings"
	"testing"
	"time"
)

func networkQuestion() Question {
	q := cpuTimeQuestion()
	q.Definition.Name = "network"
	q.Definition.Title = "Network"
	s := *q.Definition.Source.ExactStream
	s.Name = "system.network.io"
	s.Unit = "By"
	s.Attributes = map[string]string{"network.interface.name": "eth0", "network.io.direction": "receive"}
	q.Definition.Source.Metric = s.Name
	q.Definition.Source.Unit = s.Unit
	q.Definition.Source.ExactStream = &s
	return q
}
func TestNetworkByteProfileStrictValuesAndTemporalDrift(t *testing.T) {
	q := networkQuestion()
	rows := []json.RawMessage{cpuTimeRow(q, "1000000000", "100"), cpuTimeRow(q, "3000000000", "612")}
	r, e := Evaluate(q, rows, time.Unix(3, 0), "", nil)
	if e != nil || r.Counters.Series[0].Intervals[0].Delta != "512" || r.Counters.Series[0].Intervals[0].ExactPerSecond != "256" {
		t.Fatal(r, e)
	}
	for _, bad := range []string{"-1", "1.00000000000000000001", "9007199254740992", "NaN"} {
		raw := cpuTimeRow(q, "1000000000", bad)
		if _, e := Evaluate(q, []json.RawMessage{raw, rows[1]}, time.Unix(3, 0), "", nil); e == nil {
			t.Fatal("accepted byte value", bad)
		}
	}
	missing := strings.Replace(string(rows[0]), `"flags":0`, `"flags":1`, 1)
	missing = strings.Replace(missing, `"value":"100"`, `"value":"-1"`, 1)
	if _, e := Evaluate(q, []json.RawMessage{json.RawMessage(missing), rows[1]}, time.Unix(3, 0), "", nil); e != nil {
		t.Fatal("missing became numerical evidence", e)
	}
	for _, replacement := range []string{`"temporality":1`, `"temporality":2.0000000000000000001`} {
		raw := strings.Replace(string(rows[0]), `"temporality":2`, replacement, 1)
		if _, e := Evaluate(q, []json.RawMessage{json.RawMessage(raw)}, time.Unix(3, 0), "", nil); e == nil {
			t.Fatal("accepted temporal drift", replacement)
		}
	}
}
func TestNetworkCandidatesNeverBecomeAutomaticDefaults(t *testing.T) {
	q := HostDiscoveryQuestion{Environment: "prod", FromMS: 100000, ToMS: 104000}
	rows := []json.RawMessage{hostTestRow("system.cpu.utilization", "gauge", "1", map[string]string{"cpu.mode": "user"}, nil), hostTestRow("system.memory.usage", "counter", "By", map[string]string{"system.memory.state": "used"}, nil)}
	for _, direction := range []string{"receive", "transmit"} {
		row := hostTestRow("system.cpu.time", "counter", "s", map[string]string{"network.interface.name": "eth0", "network.io.direction": direction}, nil)
		row = json.RawMessage(strings.ReplaceAll(strings.ReplaceAll(string(row), `"system.cpu.time"`, `"system.network.io"`), `"s"`, `"By"`))
		rows = append(rows, row)
	}
	inventory, e := discoverHostRows(q, rows)
	if e != nil {
		t.Fatal(e)
	}
	p, e := PlanHostDashboard(HostDashboardRequest{Question: q, Host: inventory.Hosts[0]})
	if e != nil || len(p.Widgets) != 2 {
		t.Fatal(p, e)
	}
	network := 0
	for _, w := range p.Widgets {
		if w.Question.Definition.Source.Metric == "system.network.io" {
			t.Fatal("network became hidden default")
		}
	}
	for _, w := range p.Candidates {
		if HostNetworkDefinition(w.Question.Definition) {
			network++
			if w.Question.Definition.CounterDomain != metrics.DecimalCounterDomain {
				t.Fatal(w)
			}
		}
	}
	if network != 2 {
		t.Fatal("network candidates missing", network)
	}
	i := Instrument{Kind: "counter", Metric: "system.network.io", Unit: "By", SchemaURL: HostMetricsSchema, Temporality: 2, Monotonic: true}
	if !HostNetworkInstrument(i) {
		t.Fatal(i)
	}
	for _, edit := range []func(*Instrument){func(i *Instrument) { i.Unit = "bytes" }, func(i *Instrument) { i.Temporality = 1 }, func(i *Instrument) { i.Monotonic = false }, func(i *Instrument) { i.SchemaURL = "" }} {
		x := i
		edit(&x)
		if HostNetworkInstrument(x) {
			t.Fatal("broadened profile", x)
		}
	}
}
