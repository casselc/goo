package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
)

func TestHegelStoredComparison(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		n := hegel.Draw(tc, hegel.Integers(1, 1000000))
		delta := hegel.Draw(tc, hegel.Integers(1, 1000000))
		makeCapture := func(v int) *exploration.SavedCapture {
			q := exploration.Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 100}
			count := new(big.Int).Add(big.NewInt(9007199254740992), big.NewInt(int64(v))).String()
			return &exploration.SavedCapture{CaptureReceipt: exploration.CaptureReceipt{Version: 1, Title: "Stored evidence", CapturedAt: "2026-10-19T00:00:00Z", Question: exploration.CaptureQuestion{Kind: "explore", Explore: &q}}, Explore: &exploration.Result{Question: q, Count: count, State: "observed", Basis: strings.Repeat("a", 64)}}
		}
		a, b := makeCapture(n), makeCapture(n+delta)
		before, _ := json.Marshal([]*exploration.SavedCapture{a, b})
		ab, e := exploration.CompareEvidence(a, b)
		if e != nil {
			tc.Fatal(e)
		}
		ba, e := exploration.CompareEvidence(b, a)
		if e != nil {
			tc.Fatal(e)
		}
		if ab.CountDifference != fmt.Sprint(delta) || ba.CountDifference != fmt.Sprint(-delta) {
			tc.Fatal("count arithmetic lost precision")
		}
		after, _ := json.Marshal([]*exploration.SavedCapture{a, b})
		if string(before) != string(after) {
			tc.Fatal("comparison mutated receipts")
		}
		b.Question.Explore.Environment = "other"
		b.Explore.Question.Environment = "other"
		r, e := exploration.CompareEvidence(a, b)
		if e != nil || r.State != "incompatible" || r.CountDifference != "" {
			tc.Fatal("cross-environment delta")
		}
	}, options(100)...)
}
func TestQuintStoredComparisonLive(t *testing.T) {
	name := os.Getenv("GOO_EVIDENCE_PAIR_LEDGER")
	if name == "" {
		t.Skip("live ledger opt-in")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Events []struct {
				Action     string
				Reads      int
				Compatible bool
				Delta      bool
			}
			Before, After   int
			Original, Final json.RawMessage
			Pair            exploration.EvidenceComparison
		}
		if json.Unmarshal(raw, &l) != nil {
			t.Fatal("invalid ledger")
		}
		if mutation == 0 {
			l.Events[1].Reads++
		}
		if mutation == 1 {
			l.Events[2].Delta = true
		}
		if mutation == 2 {
			l.Pair.CountDifference = "999"
		}
		if mutation == 3 {
			l.Final = json.RawMessage(`{}`)
		}
		coherent := l.Before == l.After && string(l.Original) == string(l.Final) && l.Pair.CountDifference == "2" && l.Pair.Measures != nil && l.Pair.Measures.P95DifferenceNS == "9000000"
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, event := range l.Events {
			rec.Record(context.Background(), event.Action, nil, "obs.reads", event.Reads, "obs.compatible", event.Compatible, "obs.delta", event.Delta, "obs.coherent", coherent)
		}
		checker, e := validate.NewChecker("model/evidence_comparison.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("live comparison rejected")
	}
	for i := 0; i < 4; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}
