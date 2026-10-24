package verification

import (
	"context"
	"encoding/json"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"os"
	"testing"
)

func TestQuintStreamEntryLedger(t *testing.T) {
	name := os.Getenv("GOO_STREAM_ENTRY_LEDGER")
	if name == "" {
		t.Skip("observed browser ledger required")
	}
	validateStreamEntryLedger(t, name, []int{0, 1, 2, 3, 4})
}

// Navigation holds and pauses; it never resumes, so no resume mutation is claimed.
func TestQuintActivityNavigationLedger(t *testing.T) {
	name := os.Getenv("GOO_ACTIVITY_NAVIGATION_LEDGER")
	if name == "" {
		t.Skip("observed Activity navigation ledger required")
	}
	validateStreamEntryLedger(t, name, []int{0, 2, 3, 4})
}

func validateStreamEntryLedger(t *testing.T, name string, mutations []int) {
	t.Helper()
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Unexpected, Errors []string
			FirstRowY          float64
			Events             []struct {
				Action          string
				Reads, Retained int
				Gap, Coherent   bool
			}
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch mutation {
		case 0:
			for i := range l.Events {
				if l.Events[i].Action == "hold" {
					l.Events[i].Reads++
					break
				}
			}
		case 1:
			for i := range l.Events {
				if l.Events[i].Action == "resume" {
					l.Events[i].Gap = false
					break
				}
			}
		case 2:
			l.Events[0].Reads = 2
		case 3:
			l.Events[0].Retained = 101
		case 4:
			l.Unexpected = []string{"hidden aggregate"}
		}
		coherent := len(l.Unexpected) == 0 && len(l.Errors) == 0 && l.FirstRowY < 844
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, ev := range l.Events {
			rec.Record(context.Background(), ev.Action, nil, "obs.reads", ev.Reads, "obs.gap", ev.Gap, "obs.coherent", coherent && ev.Coherent && ev.Retained <= 100)
		}
		c, e := validate.NewChecker("model/stream_entry.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("observed entry rejected")
	}
	for _, i := range mutations {
		if check(i) {
			t.Fatal("negative control accepted", i)
		}
	}
}
