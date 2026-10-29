package verification

import (
	"context"
	"encoding/json"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"os"
	"testing"
)

func TestQuintCaptureCatalogLedger(t *testing.T) {
	name := os.Getenv("GOO_CAPTURE_CATALOG_LEDGER")
	if name == "" {
		t.Skip("real browser catalog ledger required")
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
	var events []event
	if e = json.Unmarshal(raw, &events); e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		sink := &qobs.MemorySink{}
		r := qobs.NewRecorder(sink)
		for i, v := range events {
			if mutation == 0 && i == len(events)-1 {
				v.Reads++
			}
			if mutation == 1 && i == 0 {
				v.Reads = 0
			}
			if mutation == 2 && i == len(events)-1 {
				v.Coherent = false
			}
			r.Record(context.Background(), v.Action, nil, "obs.reads", v.Reads, "obs.coherent", v.Coherent)
		}
		c, e := validate.NewChecker("model/capture_catalog.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		result, e := c.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return result.OK()
	}
	if !check(-1) {
		t.Fatal("real observed catalog journey refused")
	}
	for i := range 3 {
		if check(i) {
			t.Fatal("accepted mutation", i)
		}
	}
}
