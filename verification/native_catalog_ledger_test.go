package verification

import (
	"context"
	"encoding/json"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"os"
	"testing"
)

func TestQuintNativeCatalogLedger(t *testing.T) {
	path := os.Getenv("GOO_NATIVE_CATALOG_LEDGER")
	if path == "" {
		t.Skip("observed native catalog ledger required")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	type event struct {
		Action string
		Count  int
	}
	var events []event
	if e = json.Unmarshal(raw, &events); e != nil {
		t.Fatal(e)
	}
	// This binding validates the observed capacity boundary prefix, not campaign
	// completion. Require this concrete task's four observations before replay.
	if len(events) != 4 || events[0].Action != "publish" || events[1].Action != "publish" || events[2].Action != "refuse" || events[3].Action != "duplicate" {
		t.Fatal("missing boundary observations")
	}
	check := func(mutation int) bool {
		sink := &qobs.MemorySink{}
		r := qobs.NewRecorder(sink)
		for i, v := range events {
			if mutation == 0 && i == 2 {
				v.Count++
			}
			if mutation == 1 && i == 2 {
				v.Action = "publish"
				v.Count++
			}
			if mutation == 2 && i == 3 {
				v.Count--
			}
			r.Record(context.Background(), v.Action, nil, "obs.count", v.Count)
		}
		checker, e := validate.NewChecker("model/native_catalog.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("real capacity observations rejected")
	}
	for i := range 3 {
		if check(i) {
			t.Fatal("accepted changed admission observation", i)
		}
	}
}
