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

func TestQuintKeptRecordsLedger(t *testing.T) {
	file := os.Getenv("GOO_KEPT_RECORDS_LEDGER")
	if file == "" {
		t.Skip("real record keep ledger required")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutant int) bool {
		type segment struct {
			Intent  exploration.KeepRecordsIntent
			Receipt exploration.KeptRecords
		}
		var l struct {
			Result, Saved segment
			Requests      []string
		}
		if e = json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch mutant {
		case 0:
			l.Result.Receipt.Question.Offset = 0
		case 1:
			l.Saved.Intent.Indexes = []int{0, 1}
		case 2:
			l.Result.Receipt.Members[0].Record = json.RawMessage(`{"Body":"forged"}`)
		case 3:
			l.Saved.Receipt.CaptureAt = ""
		case 4:
			l.Requests = append(l.Requests, "/api/explore")
		}
		coherent := len(l.Requests) == 1 && l.Requests[0] == "/api/kept-observations/records"
		for _, s := range []segment{l.Result, l.Saved} {
			coherent = coherent && s.Intent.Validate() == nil && s.Receipt.Validate() == nil && s.Intent.RequestID == s.Receipt.ID && len(s.Intent.Indexes) == len(s.Receipt.Members)
			for i, index := range s.Intent.Indexes {
				coherent = coherent && i < len(s.Receipt.Members) && s.Receipt.Members[i].Index == index
			}
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		rec.Record(context.Background(), "inspect", nil, "obs.reads", 0, "obs.published", 0, "obs.coherent", coherent)
		rec.Record(context.Background(), "publish", nil, "obs.reads", 0, "obs.published", 1, "obs.coherent", coherent)
		rec.Record(context.Background(), "retry", nil, "obs.reads", 0, "obs.published", 1, "obs.coherent", coherent)
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
		t.Fatal("original observed keep rejected")
	}
	for m := 0; m < 5; m++ {
		if check(m) {
			t.Fatal("mutated keep accepted", m)
		}
	}
}
