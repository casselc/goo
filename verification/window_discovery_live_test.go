package verification

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"os"
	"testing"
)

func TestQuintWindowDiscoveryLive(t *testing.T) {
	file := os.Getenv("GOO_WINDOW_DISCOVERY_LEDGER")
	if file == "" {
		t.Skip("live ledger opt-in")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Events []struct {
				Action       string
				Reads, Phase int
			}
			Cases []struct {
				Discoveries []exploration.CapturedWindowDiscovery
				Captures    []exploration.CapturePage
				OriginalIDs []string
				Expected    []measurements.Question
			}
			After int
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch mutation {
		case 0:
			l.Events[0].Reads++
		case 1:
			l.Events[1].Phase = 3
		case 2:
			l.Cases[0].Captures[0].Receipt.Question.Measurement.FromNS = ""
		case 3:
			l.Cases[0].OriginalIDs[0] = l.Cases[0].Captures[0].Receipt.ID
		case 4:
			l.Cases[0].Discoveries[0].Support.Windows[0].Members = nil
		case 5:
			l.After++
		}
		coherent := len(l.Cases) == 2 && l.After == 4
		for _, c := range l.Cases {
			coherent = coherent && len(c.Captures) == 2 && len(c.Discoveries) == 2 && len(c.Expected) == 2
			for i, d := range c.Discoveries {
				w := d.Support.Windows[0]
				q := c.Captures[i].Receipt.Question.Measurement
				want := c.Expected[i]
				coherent = coherent && d.Receipt.ID == c.OriginalIDs[i] && c.Captures[i].Receipt.ID != c.OriginalIDs[i] && q.FromNS == want.FromNS && q.ToNS == want.ToNS && len(w.Members) == d.Support.Streams && len(w.Members) > 0 && w.FromNS == want.FromNS && w.ToNS == want.ToNS && c.Captures[i].Measurement.State == "observed"
			}
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, event := range l.Events {
			rec.Record(context.Background(), event.Action, nil, "obs.reads", event.Reads, "obs.phase", event.Phase, "obs.coherent", coherent)
		}
		checker, e := validate.NewChecker("model/window_discovery.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("observed trace rejected")
	}
	for i := 0; i < 6; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}
