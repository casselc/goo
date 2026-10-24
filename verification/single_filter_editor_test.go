package verification

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"os"
	"testing"
)

// Replay actual browser staging observations; reject hidden reads and mismatched scope.
func TestQuintSingleFilterEditorLedger(t *testing.T) {
	file := os.Getenv("GOO_SINGLE_FILTER_LEDGER")
	if file == "" {
		t.Skip("real browser ledger required")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	check := func(bad int) bool {
		var l struct {
			Original, Requested, Published exploration.Query
			ReadsBeforeRun                 int
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch bad {
		case 0:
			l.ReadsBeforeRun++
		case 1:
			l.Published.Where = nil
		case 2:
			l.Requested.FromMS++
		case 3:
			l.Requested.Environment = "other"
		}
		requestedJSON, _ := json.Marshal(l.Requested)
		publishedJSON, _ := json.Marshal(l.Published)
		coherent := l.Requested.Validate() == nil && string(requestedJSON) == string(publishedJSON) && l.Requested.FromMS == l.Original.FromMS && l.Requested.ToMS == l.Original.ToMS && l.Requested.Environment == l.Original.Environment && l.ReadsBeforeRun == 0
		sink := &qobs.MemorySink{}
		qobs.NewRecorder(sink).Record(context.Background(), "inspect", nil, "obs.reads", l.ReadsBeforeRun, "obs.published", 0, "obs.coherent", coherent)
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
		t.Fatal("actual editor staging trace refused")
	}
	for bad := 0; bad < 4; bad++ {
		if check(bad) {
			t.Fatalf("incorrect trace %d admitted", bad)
		}
	}
}
