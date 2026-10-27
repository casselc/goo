package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"hegel.dev/go/hegel"
	"math/big"
	"os/exec"
	"testing"
)

// Reuse the existing state machine with a fractional affine transform. The
// large offset would erase small increments under Float64 subtraction.
type decimalCounterMachine struct {
	*counterMachine
	scale int64
}

func (m *decimalCounterMachine) InvariantNoInventedRate(tc hegel.TestCase) {
	points := append([]metrics.CounterPoint(nil), m.points...)
	base := big.NewRat(10000000000000000, 1)
	for i := range points {
		value, _ := new(big.Int).SetString(points[i].Value, 10)
		points[i].Value = metrics.CounterDecimal(new(big.Rat).Add(base, new(big.Rat).SetFrac(value, big.NewInt(m.scale))))
	}
	before, _ := json.Marshal(points)
	r, e := metrics.EvaluateCounters(metrics.CounterRequest{ValueDomain: metrics.DecimalCounterDomain, FromNS: fmt.Sprint(m.from), ToNS: fmt.Sprint(m.end), MaxGapNS: "2", Points: points})
	after, _ := json.Marshal(points)
	if e != nil || string(before) != string(after) || r.Series[0].State != m.expected {
		tc.Errorf("fractional epoch state %v %v", r, e)
		return
	}
	for _, p := range r.Series[0].Intervals {
		if p.State != "ok" && (p.PerSecond != nil || p.Delta != "" || p.ExactPerSecond != "") {
			tc.Errorf("unknown acquired arithmetic: %+v", p)
		}
	}
	if m.expected == "ok" {
		previous, latestEnd := int64(-1), int64(-1)
		for _, p := range m.points {
			end, _ := new(big.Int).SetString(p.EndNS, 10)
			if end.Int64() < m.end && end.Int64() > latestEnd {
				latestEnd = end.Int64()
				v, _ := new(big.Int).SetString(p.Value, 10)
				previous = v.Int64()
			}
		}
		want := big.NewRat((m.value-previous)*1000000000, m.scale*(m.end-m.from))
		if previous < 0 || r.Series[0].Intervals[0].ExactPerSecond != want.RatString() {
			tc.Errorf("exact fractional rate differs: %v want %s", r, want)
		}
	}
}
func TestHegelDecimalCounterStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		scale := int64(10)
		for i := 0; i < hegel.Draw(tc, hegel.Integers(0, 5)); i++ {
			scale *= 10
		}
		m := &decimalCounterMachine{counterMachine: &counterMachine{points: []metrics.CounterPoint{cp(1, 10, 0)}, start: 1, end: 10, from: 9, expected: "no_data"}, scale: scale}
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(20), hegel.WithAlwaysCheckInvariants("InvariantNoInventedRate"))
	}, options(8)...)
}
func TestHegelCPUTimeRecordContextParity(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		offset := int64(hegel.Draw(tc, hegel.Integers(1, 999999)))
		s := metrics.Stream{Resource: map[string]string{"host.id": fmt.Sprint("host-", offset), "deployment.environment.name": "prod", "telemetry.sdk.language": "not-a-record-filter"}, Scope: "hostmetrics", SchemaURL: measurements.HostMetricsSchema, Name: "system.cpu.time", Unit: "s", Attributes: map[string]string{"cpu.mode": "user"}}
		q := measurements.Question{Definition: measurements.Definition{APIVersion: measurements.Version, Name: "cpu-time", Title: "CPU", Kind: "counter", CounterDomain: metrics.DecimalCounterDomain, Environment: "prod", LookbackMinutes: 2, MaxGapMS: 60000, Grain: "per-stream", Source: measurements.Binding{Resource: &measurements.ResourceBinding{Kind: "host", ID: s.Resource["host.id"]}, Services: []string{}, Scope: s.Scope, SchemaURL: s.SchemaURL, Metric: s.Name, Unit: s.Unit, ExactStream: &s}}, FromMS: 1000, ToMS: 2000}
		lo, hi := fmt.Sprint(1000000000+offset), fmt.Sprint(1999999999-offset)
		signal := []string{"logs", "spans"}[hegel.Draw(tc, hegel.Integers(0, 1))]
		want, e := exploration.PlanNativeRecordContext(q, s, signal, lo, hi)
		if e != nil {
			tc.Errorf("Go: %v", e)
			return
		}
		input, _ := json.Marshal(map[string]any{"q": q, "s": s, "signal": signal, "lo": lo, "hi": hi})
		cmd := exec.Command("node", "-e", `const x=JSON.parse(require('fs').readFileSync(0,'utf8'));process.stdout.write(JSON.stringify(require('../ui/native-record-context').plan(x.q,x.s,x.signal,x.lo,x.hi)))`)
		cmd.Stdin = bytes.NewReader(input)
		raw, e := cmd.CombinedOutput()
		var got exploration.NativeRecordContext
		if e != nil || json.Unmarshal(raw, &got) != nil || !aeJSONEqual(want, got) {
			tc.Errorf("JS/Go context differs: %v %s", e, raw)
			return
		}
		if _, e := exploration.Compile(got.Question); e != nil {
			tc.Errorf("source compilation: %v", e)
		}
	}, options(8)...)
}
