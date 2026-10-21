package verification

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"os"
	"reflect"
	"testing"
)

type evidenceSummaryFact struct{ Label, Value string }
type evidenceSummaryDisplay struct {
	Mode, State, Notice string
	Facts, Details      []evidenceSummaryFact
}

func summaryValue(rows []evidenceSummaryFact, label string) string {
	for _, r := range rows {
		if r.Label == label {
			return r.Value
		}
	}
	return ""
}
func TestQuintEvidenceSummaryLive(t *testing.T) {
	file := os.Getenv("GOO_EVIDENCE_SUMMARY_LEDGER")
	if file == "" {
		t.Skip("real capture ledger opt-in")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Capture, Reopened          exploration.CapturePage
			Metadata                   struct{ ReadAt string }
			FullSummary, LinkedSummary evidenceSummaryDisplay
			Before, After              int
		}
		if json.Unmarshal(raw, &l) != nil {
			t.Fatal("invalid ledger")
		}
		switch mutation {
		case 0:
			l.FullSummary.Mode = "live"
		case 1:
			l.FullSummary.Details[1].Value = l.Capture.Receipt.CapturedAt
		case 2:
			l.FullSummary.Details[2].Value = "wrong"
		case 3:
			l.After++
		case 4:
			l.LinkedSummary.State = "healthy"
		case 5:
			l.LinkedSummary.Details[1].Value = "wrong"
		case 6:
			l.FullSummary.Facts[3].Value = "1"
		case 7:
			l.Reopened.Receipt.Title = "changed"
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		record := func(action string, ok bool) { rec.Record(context.Background(), action, nil, "obs.coherent", ok) }
		r := l.Capture.Measurement
		record("open", r != nil && measurements.ValidateResult(r) == nil && l.FullSummary.Mode == "frozen" && l.FullSummary.State == r.State && summaryValue(l.FullSummary.Facts, "Source records") == r.Count)
		record("inspect", l.Before == l.After && summaryValue(l.FullSummary.Details, "Source read") == r.ReadAt && summaryValue(l.FullSummary.Details, "Stored custody time") == l.Capture.Receipt.CapturedAt && summaryValue(l.FullSummary.Details, "Evidence basis") == r.Basis)
		record("linked", l.LinkedSummary.Mode == "frozen" && l.LinkedSummary.State == r.State && summaryValue(l.LinkedSummary.Details, "Source read") == l.Metadata.ReadAt && summaryValue(l.LinkedSummary.Details, "Stored custody time") == l.Capture.Receipt.CapturedAt && summaryValue(l.LinkedSummary.Facts, "Source records") == r.Count)
		record("reopen", reflect.DeepEqual(l.Capture, l.Reopened))
		checker, e := validate.NewChecker("model/evidence_summary.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("actual summary ledger rejected")
	}
	for i := 0; i < 8; i++ {
		if check(i) {
			t.Fatal("corrupt summary ledger accepted", i)
		}
	}
}
