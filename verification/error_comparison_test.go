package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"testing"
)

type errorComparisonMachine struct {
	errors, unset  int
	changed, scope bool
	shift          int64
}

func (m *errorComparisonMachine) RuleCounts(tc hegel.TestCase) {
	m.errors = hegel.Draw(tc, hegel.Integers(0, 100))
	m.unset = hegel.Draw(tc, hegel.Integers(0, 100-m.errors))
}
func (m *errorComparisonMachine) RuleReadDisagreement(tc hegel.TestCase) { m.changed = !m.changed }
func (m *errorComparisonMachine) RuleScope(tc hegel.TestCase)            { m.scope = !m.scope }
func (m *errorComparisonMachine) RuleWindow(tc hegel.TestCase) {
	m.shift = int64(hegel.Draw(tc, hegel.Integers(0, 100000)))
}
func (m *errorComparisonMachine) InvariantMeasures(tc hegel.TestCase) {
	makeC := func(errors, unset int) exploration.Cohort {
		c := cohort(100, "observed")
		c.Question.SpanStatus = true
		c.Question.Summary = "span-duration"
		d := exploration.ReferenceDurationSummary([]uint64{10, 20})
		d.Count = "100"
		d.SumNS = "1500"
		s := exploration.SpanStatusSummary{Count: "100", Error: strconv.Itoa(errors), Unset: strconv.Itoa(unset), OK: strconv.Itoa(100 - errors - unset), Unknown: "0"}
		c.Duration = &d
		c.SpanStatus = &s
		c.Groups[0].Duration = &d
		c.Groups[0].SpanStatus = &s
		return c
	}
	a, b := makeC(10, 20), makeC(m.errors, m.unset)
	b.Question.FromMS += m.shift
	b.Question.ToMS += m.shift
	if m.scope {
		b.Question.Environment = "other"
	}
	if m.changed {
		b.State = "changed_between_reads"
	}
	r, e := exploration.CompareCohorts(exploration.ComparisonRequest{Baseline: a, Selection: b})
	if m.scope {
		if e == nil {
			panic("mixed scope accepted")
		}
		return
	}
	if e != nil {
		panic(e)
	}
	if m.changed {
		if r.State != "unknown" || r.Measures != nil {
			panic("invented comparison")
		}
		return
	}
	if r.State != "ok" || r.Measures.MeanDifferenceNS != "0.000" || r.Measures.P95DifferenceNS != "0" || r.Measures.ErrorDifferencePP != fmt.Sprintf("%d.00", m.errors-10) {
		panic("wrong full-population difference")
	}
	reverse, e := exploration.CompareCohorts(exploration.ComparisonRequest{Baseline: b, Selection: a})
	if e != nil {
		panic(e)
	}
	x, _ := new(big.Rat).SetString(r.Measures.ErrorDifferencePP)
	y, _ := new(big.Rat).SetString(reverse.Measures.ErrorDifferencePP)
	if x.Add(x, y).Sign() != 0 {
		panic("comparison lost antisymmetry")
	}
}
func TestHegelErrorComparisonStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		hegel.RunStateful(tc, &errorComparisonMachine{errors: 10, unset: 20}, hegel.WithStatefulStepCount(25), hegel.WithAlwaysCheckInvariants("InvariantMeasures"))
	}, options(100)...)
}

type errorComparisonLedger struct {
	Result     exploration.Result `json:"result"`
	Comparison struct {
		Baseline  exploration.Cohort           `json:"baseline"`
		Selection exploration.Cohort           `json:"selection"`
		Result    exploration.CohortComparison `json:"result"`
	} `json:"comparison"`
	Saved    exploration.CapturePage `json:"saved"`
	Reopened exploration.CapturePage `json:"reopened"`
}

func TestQuintErrorComparisonLiveLedger(t *testing.T) {
	name := os.Getenv("GOO_ERROR_COMPARISON_LEDGER")
	if name == "" {
		t.Skip("live ledger opt-in")
	}
	data, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	var l errorComparisonLedger
	if json.Unmarshal(data, &l) != nil {
		t.Fatal("ledger")
	}
	check := func(l errorComparisonLedger) (bool, string) {
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink, qobs.WithActor("error-comparison"))
		r := l.Result
		s := r.SpanStatus
		rec.Record(context.Background(), "measure", nil, "obs.coherent", s != nil && s.Validate() == nil && s.Count == "124" && s.Error == "20" && s.Unset == "44" && r.Duration.SumNS == "2600000000" && len(r.Rows) == 100)
		c := l.Comparison
		expected, err := exploration.CompareCohorts(exploration.ComparisonRequest{Baseline: c.Baseline, Selection: c.Selection})
		coherent := err == nil && reflect.DeepEqual(expected, &c.Result) && c.Baseline.Count == "80" && c.Baseline.Duration.SumNS == "260000000" && c.Baseline.SpanStatus.Error == "2" && c.Selection.Count == "124" && c.Result.Measures.ErrorDifferencePP == "13.63" && c.Result.Measures.P95DifferenceNS == "90000000" && c.Baseline.Question.ToMS == c.Selection.Question.FromMS && c.Baseline.Question.ToMS-c.Baseline.Question.FromMS == c.Selection.Question.ToMS-c.Selection.Question.FromMS
		rec.Record(context.Background(), "compare", nil, "obs.coherent", coherent)
		pair := l.Saved.Comparison
		coherent = pair != nil
		if pair != nil {
			want, err := exploration.CompareCohorts(pair.Request)
			coherent = err == nil && reflect.DeepEqual(want, pair.Result) && pair.Result.Measures.ErrorDifferencePP == "13.63" && pair.Result.Measures.P95DifferenceNS == "90000000"
		}
		rec.Record(context.Background(), "save", nil, "obs.coherent", coherent)
		rec.Record(context.Background(), "reopen", nil, "obs.coherent", l.Saved.Receipt.ID == l.Reopened.Receipt.ID && reflect.DeepEqual(l.Saved.Comparison, l.Reopened.Comparison))
		checker, e := validate.NewChecker("model/error_comparison.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		report, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return report.OK(), report.Summary()
	}
	ok, summary := check(l)
	artifact("error-comparison-live.txt", []byte(summary))
	if !ok {
		t.Fatal(summary)
	}
	for i := 0; i < 4; i++ {
		var bad errorComparisonLedger
		json.Unmarshal(data, &bad)
		switch i {
		case 0:
			bad.Result.SpanStatus.Count = "100"
		case 1:
			bad.Comparison.Result.Measures.P95DifferenceNS = "45000000"
		case 2:
			bad.Comparison.Baseline.Question.ToMS--
		case 3:
			bad.Reopened.Comparison.Result.Measures.ErrorDifferencePP = "0.00"
		}
		ok, summary := check(bad)
		artifact(fmt.Sprintf("error-comparison-negative-%d.txt", i), []byte(summary))
		if ok {
			t.Fatal("negative accepted", i)
		}
	}
}
