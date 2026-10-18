package verification

import (
	"encoding/json"
	"github.com/casselc/goo/measurements"
	"hegel.dev/go/hegel"
	"os"
	"reflect"
	"testing"
	"time"
)

type measurementMachine struct {
	q      measurements.Question
	rows   []json.RawMessage
	result *measurements.Result
}

func (m *measurementMachine) RuleObserve(tc hegel.TestCase) {
	value := hegel.Draw(tc, hegel.Integers(0, 100))
	instance := []string{"a", "b", "c"}[hegel.Draw(tc, hegel.Integers(0, 2))]
	at := hegel.Draw(tc, hegel.Integers(100001, 219999))
	d := m.q.Definition
	b, _ := json.Marshal(map[string]any{"resource": map[string]string{"service.name": d.Source.Services[0], "service.instance.id": instance, "deployment.environment.name": d.Environment}, "scope": d.Source.Scope, "version": d.Source.Version, "schema_url": d.Source.SchemaURL, "name": d.Source.Metric, "unit": d.Source.Unit, "attributes": d.Source.Attributes, "end_ns": fmtNS(int64(at)), "value": fmtInt(value)})
	m.rows = append(m.rows, b)
}
func fmtNS(n int64) string               { return fmtInt(n * 1000000) }
func fmtInt[T ~int | ~int64](n T) string { b, _ := json.Marshal(n); return string(b) }
func (m *measurementMachine) RuleEvaluate(tc hegel.TestCase) {
	r, e := measurements.Evaluate(m.q, m.rows, time.UnixMilli(230000), "", nil)
	if e != nil {
		panic(e)
	}
	if e = measurements.ValidateResult(r); e != nil {
		panic(e)
	}
	m.result = r
}
func (m *measurementMachine) RulePermutation(tc hegel.TestCase) {
	if m.result == nil {
		return
	}
	rows := append([]json.RawMessage{}, m.result.Rows...)
	for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
		rows[i], rows[j] = rows[j], rows[i]
	}
	r, e := measurements.Evaluate(m.q, rows, time.UnixMilli(230000), "", nil)
	if e != nil || !reflect.DeepEqual(r.Gauges, m.result.Gauges) {
		panic("delivery ordering changed gauge facts")
	}
}
func (m *measurementMachine) RuleTamper(tc hegel.TestCase) {
	if m.result == nil {
		return
	}
	b, _ := json.Marshal(m.result)
	var bad measurements.Result
	json.Unmarshal(b, &bad)
	bad.Question.Definition.Environment = "other"
	if measurements.ValidateResult(&bad) == nil {
		panic("tampered scope accepted")
	}
}
func (m *measurementMachine) InvariantQuestion(tc hegel.TestCase) {
	if e := m.q.Validate(); e != nil {
		panic(e)
	}
	if m.result != nil {
		if e := measurements.ValidateResult(m.result); e != nil {
			panic(e)
		}
	}
}
func TestHegelMeasurementStateSwarm(t *testing.T) {
	b, e := os.ReadFile("../config/measurements/queue-lag.yaml")
	if e != nil {
		t.Fatal(e)
	}
	d, e := measurements.Read(string(b))
	if e != nil {
		t.Fatal(e)
	}
	hegel.Test(t, func(tc *hegel.T) {
		m := &measurementMachine{q: measurements.Question{Definition: d, FromMS: 100000, ToMS: 220000}}
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(35), hegel.WithAlwaysCheckInvariants("InvariantQuestion"))
	}, options(120)...)
}
