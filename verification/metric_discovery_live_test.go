package verification

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"os"
	"reflect"
	"testing"
)

func TestQuintMetricDiscoveryLedger(t *testing.T) {
	path := os.Getenv("GOO_METRIC_DISCOVERY_LEDGER")
	if path == "" {
		t.Skip("live ledger opt-in")
	}
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var l struct {
		Discovery measurements.DiscoveryResult
		Result    measurements.Result
	}
	if e = json.Unmarshal(b, &l); e != nil {
		t.Fatal(e)
	}
	check := func(bad int) bool {
		// Mutate the actual ledger, then recompute observations through product validators.
		original, _ := json.Marshal(l)
		defer json.Unmarshal(original, &l)
		switch bad {
		case 0:
			l.Discovery.Question.Service = "different-service"
		case 1:
			for i := range l.Discovery.Instruments {
				l.Discovery.Instruments[i].Definition = nil
			}
		case 2:
			l.Result.Basis = "tampered"
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		record := func(a string, v bool) { rec.Record(context.Background(), a, nil, "obs.coherent", v) }
		record("discover", l.Discovery.Records > 0 && l.Discovery.Question.Service == l.Result.Question.Definition.Source.Services[0])
		found := false
		for _, i := range l.Discovery.Instruments {
			found = found || reflect.DeepEqual(i.Definition, &l.Result.Question.Definition)
		}
		record("build", found)
		record("execute", measurements.ValidateResult(&l.Result) == nil && l.Result.Count != "0")
		c, e := validate.NewChecker("model/metric_discovery.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		r, e := c.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return r.OK()
	}
	if !check(-1) {
		t.Fatal("valid live trace rejected")
	}
	for i := 0; i < 3; i++ {
		if check(i) {
			t.Fatal("incoherent trace accepted", i)
		}
	}
}
