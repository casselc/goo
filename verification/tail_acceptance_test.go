package verification

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/retention"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"os"
	"testing"
)

func TestQuintTailAcceptanceLedger(t *testing.T) {
	file := os.Getenv("GOO_TAIL_ACCEPTANCE_LEDGER")
	if file == "" {
		t.Skip("observed Activity acceptance ledger required")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	check := func(bad int) bool {
		var l struct {
			Original                                 retention.TailIntent
			Receipt                                  retention.TailReceipt
			Status, Reads                            int
			LocalExpired, PendingState, RetryEnabled bool
			Events                                   []struct {
				Action           string
				Reads, Published int
			}
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch bad {
		case 0:
			l.Status = 410
		case 1:
			l.Reads++
		case 2:
			l.RetryEnabled = false
		case 3:
			l.Original.Refs[0].Key = "forged"
		case 4:
			l.Events[0].Published++
		}
		coherent := l.Status == 502 && l.Reads == 0 && l.LocalExpired && l.PendingState && l.RetryEnabled && l.Original.Validate() == nil && l.Receipt.Manifest.Validate() == nil && l.Original.Connection == l.Receipt.Manifest.Selection.Connection && len(l.Original.Refs) == 1 && len(l.Receipt.Manifest.Selection.Members) == 1 && l.Original.Refs[0] == l.Receipt.Manifest.Selection.Members[0].Ref
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
		t.Fatal("observed acceptance rejected")
	}
	for bad := 0; bad < 5; bad++ {
		if check(bad) {
			t.Fatal("altered acceptance accepted", bad)
		}
	}
}
