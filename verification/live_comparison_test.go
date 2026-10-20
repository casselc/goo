package verification

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
)

func TestQuintLiveComparisonLedger(t *testing.T) {
	name := os.Getenv("GOO_LIVE_COMPARISON_LEDGER")
	if name == "" {
		t.Skip("live opt-in")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Baseline      exploration.CapturePage
			First, Second struct {
				Comparison exploration.EvidenceComparison
			}
			Frozen, Reopened                   exploration.CapturePage
			SourceCount, AxisBefore, AxisAfter string
			Stale                              bool
			ObsoleteBars                       int
			ObsoletePhase                      string
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch mutation {
		case 0:
			l.First.Comparison.Baseline.Receipt.ID = "changed"
		case 1:
			l.AxisBefore = "7"
		case 2:
			l.Stale = false
		case 3:
			l.Frozen.Explore.Count = l.SourceCount
		case 4:
			l.Reopened.Receipt.ID = "other"
		case 5:
			l.ObsoleteBars = 2
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		record := func(a string, b bool) { rec.Record(context.Background(), a, nil, "obs.coherent", b) }
		first := l.First.Comparison
		record("observe", first.Baseline.Receipt.ID == l.Baseline.Receipt.ID && first.Baseline.Count == "2" && first.Selection.Count == "4" && first.Selection.Receipt.Question.Explore.ToMS-first.Selection.Receipt.Question.Explore.FromMS == 120000)
		record("axis", l.AxisBefore == "4" && l.AxisAfter == "7")
		record("stale", l.Stale)
		record("freeze", l.Frozen.Explore.CapturedAt == l.Second.Comparison.Selection.ReadAt && reflect.DeepEqual(l.Frozen.Receipt.Question, l.Second.Comparison.Selection.Receipt.Question) && l.Frozen.Explore.Count == "7" && l.SourceCount == "8" && l.Frozen.Explore.Question.ToMS-l.Frozen.Explore.Question.FromMS == 120000)
		record("reopen", reflect.DeepEqual(l.Reopened, l.Frozen))
		record("discard", l.ObsoleteBars == 0 && l.ObsoletePhase == "")
		checker, e := validate.NewChecker("model/live_comparison.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("actual live comparison trace rejected")
	}
	for i := 0; i < 6; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}

func TestQuintLiveSpanMeasuresLedger(t *testing.T) {
	name := os.Getenv("GOO_LIVE_SPAN_MEASURES_LEDGER")
	if name == "" {
		t.Skip("live opt-in")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Baseline      exploration.CapturePage
			First, Second struct {
				Comparison exploration.EvidenceComparison
				Basis      string
			}
			Frozen, Reopened                          exploration.CapturePage
			SourceCount, AxisBefore, AxisAfter        string
			Stale                                     bool
			FirstDisplay, SecondDisplay, StaleDisplay struct {
				Basis, ReadAt, FromMS, ToMS string
				Summaries                   []exploration.DurationSummary
				Ceilings                    []string
				Status                      string
			}
			ObsoleteBars  int
			ObsoletePhase string
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch mutation {
		case 0:
			l.First.Comparison.Baseline.Receipt.ID = "changed"
		case 1:
			l.AxisBefore = "7"
		case 2:
			l.Stale = false
		case 3:
			l.Frozen.Explore.Count = l.SourceCount
		case 4:
			l.Reopened.Receipt.ID = "other"
		case 5:
			l.ObsoleteBars = 2
		case 6:
			l.SecondDisplay.Ceilings[1] = "10000000"
		case 7:
			l.Frozen.Explore.Duration.SumNS = "1340000000"
		case 8:
			l.Frozen.Explore.SpanStatus.Error = "5"
		case 9:
			l.StaleDisplay.Basis = "other"
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		record := func(a string, b bool) { rec.Record(context.Background(), a, nil, "obs.coherent", b) }
		first := l.First.Comparison
		record("observe", first.Baseline.Receipt.ID == l.Baseline.Receipt.ID && first.Baseline.Count == "2" && first.Selection.Count == "4" && first.Selection.Receipt.Question.Explore.ToMS-first.Selection.Receipt.Question.Explore.FromMS == 120000)
		record("axis", l.AxisBefore == "4" && l.AxisAfter == "7" && reflect.DeepEqual(l.FirstDisplay.Ceilings, []string{"10000000", "10000000"}) && reflect.DeepEqual(l.SecondDisplay.Ceilings, []string{"100000000", "100000000"}))
		record("stale", l.Stale && reflect.DeepEqual(l.StaleDisplay, l.SecondDisplay))
		record("freeze", l.Frozen.Explore.Basis == l.Second.Basis && reflect.DeepEqual(l.Frozen.Explore.Duration, l.Second.Comparison.Measures.SelectionDuration) && reflect.DeepEqual(l.Frozen.Explore.SpanStatus, l.Second.Comparison.Measures.SelectionStatus) && l.Frozen.Explore.Duration.SumNS == "340000000" && l.Frozen.Explore.SpanStatus.Error == "4" && l.Frozen.Explore.CapturedAt == l.Second.Comparison.Selection.ReadAt && reflect.DeepEqual(l.Frozen.Receipt.Question, l.Second.Comparison.Selection.Receipt.Question) && l.Frozen.Explore.Count == "7" && l.SourceCount == "8" && l.Frozen.Explore.Question.ToMS-l.Frozen.Explore.Question.FromMS == 120000)
		record("reopen", reflect.DeepEqual(l.Reopened, l.Frozen))
		record("discard", l.ObsoleteBars == 0 && l.ObsoletePhase == "")
		checker, e := validate.NewChecker("model/live_comparison.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("actual live comparison trace rejected")
	}
	for i := 0; i < 10; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}

func TestHegelRecordComparisonWindowCompatibility(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		duration := int64(hegel.Draw(tc, hegel.Integers(1, 7200000)))
		offset := int64(hegel.Draw(tc, hegel.Integers(1, 1000000)))
		makeCapture := func(from int64) *exploration.SavedCapture {
			q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: from, ToMS: from + duration, Limit: 100}
			return &exploration.SavedCapture{CaptureReceipt: exploration.CaptureReceipt{Version: 1, Title: "Observed", CapturedAt: "2026-10-20T00:00:00Z", Question: exploration.CaptureQuestion{Kind: "explore", Explore: &q}}, Explore: &exploration.Result{Question: q, Count: "1200", State: "observed", Basis: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}
		}
		a, b := makeCapture(1000), makeCapture(1000+offset)
		before, _ := json.Marshal(a)
		pair, e := exploration.CompareEvidence(a, b)
		if e != nil || pair.State != "comparable" || pair.CountDifference != "0" {
			tc.Fatal("equal-duration shifted population", e)
		}
		after, _ := json.Marshal(a)
		if string(before) != string(after) {
			tc.Fatal("baseline mutated")
		}
		b.Question.Explore.Environment = "other"
		b.Explore.Question.Environment = "other"
		pair, e = exploration.CompareEvidence(a, b)
		if e != nil || pair.State != "incompatible" || pair.CountDifference != "" {
			tc.Fatal("cross-environment numbers published", e)
		}
	}, options(100)...)
}
