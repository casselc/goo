package verification

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

func TestQuintKeptObservationLedger(t *testing.T) {
	file := os.Getenv("GOO_KEPT_LEDGER")
	if file == "" {
		t.Skip("observed browser keep ledger required")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutant int) bool {
		var l struct {
			Original   exploration.KeepTailIntent
			Receipt    exploration.KeptObservations
			RetryEqual bool
			Requests   []string
			Events     []struct {
				Action           string
				Reads, Published int
			}
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch mutant {
		case 0:
			l.Events[1].Reads++
		case 1:
			l.RetryEqual = false
		case 2:
			l.Receipt.Members[0].Record = json.RawMessage(`{"Body":"forged"}`)
		case 3:
			l.Events[2].Published++
		case 4:
			l.Original.Connection = "foreign"
		}
		coherent := l.Original.Validate() == nil && l.Receipt.Validate() == nil && l.RetryEqual && l.Original.Connection == l.Receipt.Connection && l.Original.RequestID == l.Receipt.ID && len(l.Original.Refs) == len(l.Receipt.Members) && len(l.Requests) == 2
		keys := map[string]bool{}
		for _, m := range l.Receipt.Members {
			keys[m.Key] = true
		}
		for _, ref := range l.Original.Refs {
			coherent = coherent && keys[ref.Key]
		}
		for _, path := range l.Requests {
			coherent = coherent && path == "/api/kept-observations/tail"
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, ev := range l.Events {
			rec.Record(context.Background(), ev.Action, nil, "obs.reads", ev.Reads, "obs.published", ev.Published, "obs.coherent", coherent)
		}
		checker, e := validate.NewChecker("model/annotations.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("observed keep rejected")
	}
	for mutant := 0; mutant < 5; mutant++ {
		if check(mutant) {
			t.Fatal("altered keep accepted", mutant)
		}
	}
}
