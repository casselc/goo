package metrics

import (
	"reflect"
	"testing"
)

func TestGaugeIdentityFreshnessAndConflicts(t *testing.T) {
	a := Stream{Scope: "worker", Name: "queue", Unit: "{message}", Resource: map[string]string{"service.name": "consumer", "service.instance.id": "a"}}
	b := cloneStream(a)
	b.Resource["service.instance.id"] = "b"
	ps := []GaugePoint{{Stream: a, EndNS: "10", Value: 4}, {Stream: a, EndNS: "10", Value: 4}, {Stream: b, EndNS: "80", Value: 0}, {Stream: a, EndNS: "90", Value: 8}, {Stream: a, EndNS: "90", Value: 9}}
	r, e := EvaluateGauges("0", "100", "30", ps)
	if e != nil {
		t.Fatal(e)
	}
	if len(r.Series) != 2 || r.Series[0].Samples[0].State != "stale" || r.Series[0].Samples[0].Value == nil || r.Series[0].State != "unknown" || r.Series[0].Samples[1].Value != nil || r.Series[1].State != "observed" {
		t.Fatalf("bad gauge semantics %+v", r)
	}
	for i, j := 0, len(ps)-1; i < j; i, j = i+1, j-1 {
		ps[i], ps[j] = ps[j], ps[i]
	}
	other, e := EvaluateGauges("0", "100", "30", ps)
	if e != nil || !reflect.DeepEqual(r, other) {
		t.Fatal("order affected evidence")
	}
	empty, e := EvaluateGauges("0", "100", "30", nil)
	if e != nil || len(empty.Series) != 0 {
		t.Fatal("absence became zero")
	}
}
