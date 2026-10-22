package verification

import (
	"context"
	"encoding/json"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"os"
	"testing"
)

// Reuse the observed read/history protocol: only named context actions add reads.
func TestQuintSavedOwnerLedger(t *testing.T) {
	name := os.Getenv("GOO_SAVED_OWNER_LEDGER")
	if name == "" {
		t.Skip("live ledger opt-in")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	type event struct {
		Action   string
		Reads    int
		Coherent bool
	}
	var l struct {
		Events         []event
		Reads          []json.RawMessage
		Writes, Errors []string
	}
	if e = json.Unmarshal(raw, &l); e != nil {
		t.Fatal(e)
	}
	if len(l.Reads) != 4 || len(l.Writes) != 0 || len(l.Errors) != 0 {
		t.Fatal("unexpected source/publication activity")
	}
	check := func(mutation int) bool {
		events := append([]event(nil), l.Events...)
		switch mutation {
		case 0:
			events[0].Reads++
		case 1:
			events[len(events)-1].Reads++
		case 2:
			events[len(events)-1].Coherent = false
		case 3:
			for i := range events {
				if events[i].Action == "read" {
					events[i].Reads--
					break
				}
			}
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, event := range events {
			rec.Record(context.Background(), event.Action, nil, "obs.reads", event.Reads, "obs.coherent", event.Coherent)
		}
		checker, e := validate.NewChecker("model/question_links.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		report, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return report.OK()
	}
	if !check(-1) {
		t.Fatal("observed saved/context ownership rejected")
	}
	for i := 0; i < 4; i++ {
		if check(i) {
			t.Fatalf("rejecting control %d accepted", i)
		}
	}
}
