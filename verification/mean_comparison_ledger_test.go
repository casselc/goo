package verification

import (
	"context"
	"encoding/json"
	"math/big"
	"os"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

// Reuse the population lifecycle model: arithmetic authority remains the actual
// exact duration sums/counts, independently checked here with math/big.
func TestQuintMeanComparisonObservedLedger(t *testing.T) {
	name := os.Getenv("GOO_MEAN_COMPARISON_LEDGER")
	if name == "" {
		t.Skip("requires observed browser ledger")
	}
	raw, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	type operand struct {
		Question          exploration.Query
		Count             string
		CapturedAt, Basis string
		Duration          struct{ Count, SumNs string }
	}
	type pair struct {
		Selected, Baseline operand
		Comparison         struct {
			ViewComparison struct {
				Baseline, Selection struct {
					SourceBasis, CapturedAt string
					Rows                    []struct{ Values map[string]*string }
				}
				Changes map[string]struct {
					State string
					Value *string
				}
			}
		}
	}
	type ledger struct {
		Pairs  []pair
		Events []struct {
			Action       string
			Depth, Reads int
			Coherent     bool
		}
	}
	check := func(mutation int) bool {
		var l ledger
		if json.Unmarshal(raw, &l) != nil {
			t.Fatal("invalid mean ledger")
		}
		if len(l.Pairs) != 3 || len(l.Events) == 0 {
			return false
		}
		switch mutation {
		case 0:
			l.Pairs[0].Selected.Duration.SumNs = "9000001"
		case 1:
			l.Pairs[0].Baseline.Question.Where = nil
		case 2:
			for i := range l.Events {
				if l.Events[i].Action == "back" {
					l.Events[i].Reads++
					break
				}
			}
		case 3:
			l.Events[0].Coherent = false
		case 4:
			l.Pairs[0].Baseline.Duration.Count = "3"
		case 5:
			l.Events = l.Events[:1]
		}
		last := l.Events[len(l.Events)-1]
		coherent := last.Reads == 7 && last.Depth == 1 && last.Action == "history" && len(l.Events) >= 11
		for _, p := range l.Pairs {
			a, b := p.Selected, p.Baseline
			expected := a.Question
			expected.FromMS -= a.Question.ToMS - a.Question.FromMS
			expected.ToMS = a.Question.FromMS
			expected.Offset = 0
			coherent = coherent && reflect.DeepEqual(expected, b.Question) && a.Question.Summary == "span-duration" && a.Count == a.Duration.Count && b.Count == b.Duration.Count
			view := p.Comparison.ViewComparison
			for i, o := range []operand{b, a} {
				snapshot := view.Baseline
				if i == 1 {
					snapshot = view.Selection
				}
				if len(snapshot.Rows) != 1 || snapshot.SourceBasis == "" || snapshot.CapturedAt == "" || snapshot.SourceBasis != o.Basis || snapshot.CapturedAt != o.CapturedAt {
					coherent = false
					continue
				}
				count := snapshot.Rows[0].Values["count"]
				coherent = coherent && count != nil && *count == o.Count
				value := snapshot.Rows[0].Values["duration.mean"]
				if o.Count == "0" {
					coherent = coherent && value == nil && o.Duration.SumNs == "0"
					continue
				}
				n, ok := new(big.Int).SetString(o.Duration.SumNs, 10)
				d, good := new(big.Int).SetString(o.Duration.Count, 10)
				if !ok || !good || d.Sign() <= 0 || value == nil {
					coherent = false
					continue
				}
				coherent = coherent && new(big.Rat).SetFrac(n, d).RatString() == *value
			}
			change := view.Changes["duration.mean"]
			if a.Count == "0" || b.Count == "0" {
				coherent = coherent && change.State == "unavailable" && change.Value == nil
			} else if len(view.Selection.Rows) == 1 && len(view.Baseline.Rows) == 1 {
				av, bv := view.Selection.Rows[0].Values["duration.mean"], view.Baseline.Rows[0].Values["duration.mean"]
				if av == nil || bv == nil || change.Value == nil {
					coherent = false
					continue
				}
				x, ok := new(big.Rat).SetString(*av)
				y, good := new(big.Rat).SetString(*bv)
				coherent = coherent && ok && good
				if ok && good {
					coherent = coherent && change.State == "derived" && new(big.Rat).Sub(x, y).RatString() == *change.Value
				}
			}
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, e := range l.Events {
			rec.Record(context.Background(), e.Action, nil, "obs.reads", e.Reads, "obs.depth", e.Depth, "obs.coherent", coherent && e.Coherent)
		}
		checker, err := validate.NewChecker("model/population_drilldown.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if err != nil {
			t.Fatal(err)
		}
		result, err := checker.Check(context.Background(), sink.Steps())
		if err != nil {
			t.Fatal(err)
		}
		return result.OK()
	}
	if !check(-1) {
		t.Fatal("observed mean comparison lifecycle rejected")
	}
	for mutation := 0; mutation < 6; mutation++ {
		if check(mutation) {
			t.Fatalf("mutation %d accepted", mutation)
		}
	}
}
