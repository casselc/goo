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

func TestQuintInstrumentPresenceLedger(t *testing.T) {
	file := os.Getenv("GOO_INSTRUMENT_PRESENCE_LEDGER")
	if file == "" {
		t.Skip("live ledger opt-in")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	check := func(bad int) bool {
		var l struct {
			Discovery      measurements.DiscoveryResult
			Observations   []measurements.DiscoveryObservation
			Effects        int
			DiscoveryReads int
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		p := l.Discovery.Instruments[0].CapabilityPlan
		switch bad {
		case 0:
			p.Coverage.TotalStreams++
		case 1:
			p.Coverage.Dimensions[0].Present++
		case 2:
			p.Coverage.Streams[0].Instance = "forged"
		case 3:
			delete(p.Coverage.Streams[0].Stream.Attributes, "http.request.method")
			p.Coverage.Streams[0].Stream.Attributes["forged"] = "value"
		case 4:
			p.Eligible = !p.Eligible
		case 5:
			l.Effects++
		}
		expected := measurements.PlanInstrument(l.Discovery.Instruments[0], l.Observations)
		coherent := reflect.DeepEqual(*p, expected) && len(l.Observations) == l.Discovery.Records
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		rec.Record(context.Background(), "discover", nil, "obs.coherent", coherent, "obs.reads", l.DiscoveryReads, "obs.effects", 0)
		reads := l.DiscoveryReads
		if bad == 6 {
			reads++
		}
		rec.Record(context.Background(), "inspect", nil, "obs.coherent", coherent, "obs.reads", reads, "obs.effects", l.Effects)
		checker, e := validate.NewChecker("model/instrument_presence.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		result, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return result.OK()
	}
	if !check(-1) {
		t.Fatal("valid presence trace rejected")
	}
	for bad := 0; bad < 7; bad++ {
		if check(bad) {
			t.Fatal("forged presence trace accepted", bad)
		}
	}
}
