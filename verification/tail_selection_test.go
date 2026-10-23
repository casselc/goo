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

func TestQuintTailSelectionLedger(t *testing.T) {
	file := os.Getenv("GOO_TAIL_SELECTION_LEDGER")
	if file == "" {
		t.Skip("observed Activity selection browser ledger required")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	check := func(bad int) bool {
		var l struct {
			Receipt, Recovered retention.TailReceipt
			Intent             retention.TailIntent
			Events             []struct {
				Action           string
				Reads, Published int
				Coherent         bool
			}
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch bad {
		case 0:
			l.Events[0].Reads++
		case 1:
			l.Events[1].Published++
		case 2:
			l.Recovered.RetainedAt = l.Recovered.RetainedAt.Add(1)
		case 3:
			l.Intent.Connection = "forged"
		case 4:
			l.Recovered.Manifest.Selection.Members[0].FromMS++
		}
		coherent := l.Intent.Validate() == nil && l.Receipt.Manifest.Validate() == nil && l.Recovered.Manifest.Validate() == nil && equalTailJSON(l.Receipt.Manifest, l.Recovered.Manifest) && l.Receipt.RetainedAt.Equal(l.Recovered.RetainedAt) && l.Receipt.Annotation.ID == l.Recovered.Annotation.ID && l.Intent.Connection == l.Receipt.Manifest.Selection.Connection && len(l.Intent.Refs) == 2 && len(l.Receipt.Manifest.Selection.Members) == 2
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, ev := range l.Events {
			rec.Record(context.Background(), ev.Action, nil, "obs.reads", ev.Reads, "obs.published", ev.Published, "obs.coherent", coherent && ev.Coherent)
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
		t.Fatal("observed Activity selection rejected")
	}
	for bad := 0; bad < 5; bad++ {
		if check(bad) {
			t.Fatal("altered Activity selection accepted", bad)
		}
	}
}
