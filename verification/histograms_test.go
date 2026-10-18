package verification

import (
	"context"
	"fmt"
	"github.com/casselc/goo/metrics"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
	"os"
	"strconv"
	"testing"
)

func histogramPoint(start, end int64, a, b uint64) metrics.HistogramPoint {
	s := counterStream()
	s.Name = "duration"
	s.Unit = "s"
	return metrics.HistogramPoint{Stream: s, Temporality: "cumulative", StartNS: fmt.Sprint(start), EndNS: fmt.Sprint(end), Bounds: []float64{.01}, BucketCounts: []string{fmt.Sprint(a), fmt.Sprint(b)}, Count: fmt.Sprint(a + b)}
}

type histogramMachine struct {
	points           []metrics.HistogramPoint
	start, end, from int64
	a, b             uint64
	tainted          bool
	expected         string
}

func (m *histogramMachine) add(start, end int64, a, b uint64) {
	m.from = m.end
	m.expected = "ok"
	if start != m.start {
		m.expected = "unknown"
		m.tainted = false
	} else if a < m.a || b < m.b {
		m.expected = "unknown"
		m.tainted = true
	} else if m.tainted || end-m.end > 2 {
		m.expected = "unknown"
	}
	m.start, m.end, m.a, m.b = start, end, a, b
	m.points = append(m.points, histogramPoint(start, end, a, b))
}
func (m *histogramMachine) RuleAdd(tc hegel.TestCase) {
	m.add(m.start, m.end+1, m.a+uint64(hegel.Draw(tc, hegel.Integers(0, 5))), m.b+uint64(hegel.Draw(tc, hegel.Integers(0, 5))))
}
func (m *histogramMachine) RuleRestart(tc hegel.TestCase) { m.add(m.end+1, m.end+2, 0, 0) }
func (m *histogramMachine) RuleGap(tc hegel.TestCase)     { m.add(m.start, m.end+5, m.a+1, m.b+1) }
func (m *histogramMachine) RuleDecreaseBucket(tc hegel.TestCase) {
	if m.a > 0 {
		m.add(m.start, m.end+1, m.a-1, m.b+2)
	}
}
func (m *histogramMachine) RuleDuplicate(tc hegel.TestCase) {
	m.points = append(m.points, m.points[len(m.points)-1])
}
func (m *histogramMachine) InvariantNoInventedPopulation(tc hegel.TestCase) {
	r, e := metrics.EvaluateHistograms(metrics.HistogramRequest{FromNS: fmt.Sprint(m.from), ToNS: fmt.Sprint(m.end), MaxGapNS: "2", Points: m.points})
	if e != nil {
		panic(e)
	}
	s := r.Series[0]
	if s.State != m.expected {
		panic(fmt.Sprintf("observed %s expected %s", s.State, m.expected))
	}
	for _, interval := range s.Intervals {
		if interval.State != "ok" && interval.Distribution != nil {
			panic("unknown histogram has population")
		}
		if d := interval.Distribution; d != nil {
			sum := uint64(0)
			for _, b := range d.BucketCounts {
				n, _ := strconv.ParseUint(b, 10, 64)
				sum += n
			}
			if fmt.Sprint(sum) != d.Count {
				panic("lost bucket population")
			}
		}
	}
}
func TestHegelHistogramStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		m := &histogramMachine{points: []metrics.HistogramPoint{histogramPoint(1, 10, 0, 0)}, start: 1, end: 10, from: 9, expected: "no_data"}
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(30), hegel.WithAlwaysCheckInvariants("InvariantNoInventedPopulation"))
	}, options(100)...)
}
func TestHegelHistogramMerge(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		a := uint64(hegel.Draw(tc, hegel.Integers(0, 10000)))
		b := uint64(hegel.Draw(tc, hegel.Integers(0, 10000)))
		p := histogramPoint(10, 20, a, b)
		p.Temporality = "delta"
		r, e := metrics.EvaluateHistograms(metrics.HistogramRequest{FromNS: "10", ToNS: "20", MaxGapNS: "10", Points: []metrics.HistogramPoint{p}})
		if e != nil {
			tc.Fatalf("%v", e)
		}
		one := metrics.HistogramMember{Stream: p.Stream, Interval: r.Series[0].Intervals[0]}
		two := one
		two.Stream = counterStream()
		two.Stream.Name = "duration"
		two.Stream.Unit = "s"
		two.Stream.Resource["service.instance.id"] = "two"
		merged, e := metrics.MergeAlignedHistograms([]metrics.HistogramMember{one, two})
		if e != nil || merged.State != "ok" || merged.Distribution.Count != fmt.Sprint(2*(a+b)) {
			tc.Fatalf("merge lost population: %v", e)
		}
		back, e := metrics.MergeAlignedHistograms([]metrics.HistogramMember{two, one})
		if e != nil || back.Basis != merged.Basis {
			tc.Fatalf("merge not commutative")
		}
		two.Interval.ToNS = "21"
		bad, e := metrics.MergeAlignedHistograms([]metrics.HistogramMember{one, two})
		if e != nil || bad.State != "unknown" || bad.Distribution != nil {
			tc.Fatalf("merged unaligned data")
		}
	}, options(150)...)
}
func TestQuintObservedHistogramSemantics(t *testing.T) {
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink, qobs.WithActor("histogram"), qobs.WithProcess("fixture"))
	points := []metrics.HistogramPoint{histogramPoint(1, 10, 1, 1)}
	observe := func(action string, from, to int64) *metrics.HistogramResult {
		r, e := metrics.EvaluateHistograms(metrics.HistogramRequest{FromNS: fmt.Sprint(from), ToNS: fmt.Sprint(to), MaxGapNS: "10", Points: points})
		if e != nil {
			t.Fatal(e)
		}
		known := r.Series[0].State == "ok"
		count := int64(-1)
		if known {
			count = 0
			for _, i := range r.Series[0].Intervals {
				n, _ := strconv.ParseInt(i.Distribution.Count, 10, 64)
				count += n
			}
		}
		rec.Record(context.Background(), action, nil, "obs.known", known, "obs.count", count)
		return r
	}
	points = append(points, histogramPoint(1, 20, 2, 3))
	observe("increase", 10, 20)
	points = append(points, histogramPoint(21, 30, 1, 0))
	observe("restart", 20, 30)
	points = append(points, histogramPoint(21, 40, 2, 2))
	valid := observe("increase", 30, 40)
	points = append(points, histogramPoint(21, 70, 3, 4))
	observe("gap", 40, 70)
	points = append(points, histogramPoint(21, 50, 3, 2), histogramPoint(21, 60, 3, 3))
	observe("late_repair", 40, 70)
	one := metrics.HistogramMember{Stream: valid.Series[0].Stream, Interval: valid.Series[0].Intervals[0]}
	two := one
	two.Stream = counterStream()
	two.Stream.Name = "duration"
	two.Stream.Unit = "s"
	two.Stream.Resource["service.instance.id"] = "two"
	merged, e := metrics.MergeAlignedHistograms([]metrics.HistogramMember{one, two})
	if e != nil {
		t.Fatal(e)
	}
	n, _ := strconv.ParseInt(merged.Distribution.Count, 10, 64)
	rec.Record(context.Background(), "merge", nil, "obs.known", merged.State == "ok", "obs.count", n)
	two.Interval.ToNS = "41"
	bad, e := metrics.MergeAlignedHistograms([]metrics.HistogramMember{one, two})
	if e != nil {
		t.Fatal(e)
	}
	rec.Record(context.Background(), "bad_merge", nil, "obs.known", bad.State == "ok", "obs.count", int64(-1))
	checker, e := validate.NewChecker("model/histogram.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	steps := sink.Steps()
	report, e := checker.Check(context.Background(), steps)
	if e != nil {
		t.Fatal(e)
	}
	if !report.OK() {
		t.Fatal(report.Summary())
	}
	if e = artifact("histogram-validation.txt", []byte(report.Summary())); e != nil {
		t.Fatal(e)
	}
	for _, actor := range report.Actors {
		if actor.ITF != "" {
			data, e := os.ReadFile(actor.ITF)
			if e != nil {
				t.Fatal(e)
			}
			if e = artifact("histogram-observed.itf.json", data); e != nil {
				t.Fatal(e)
			}
		}
	}
	steps[1].Obs["known"] = true
	steps[1].Obs["count"] = int64(1)
	negative, e := checker.Check(context.Background(), steps)
	if e != nil {
		t.Fatal(e)
	}
	if negative.OK() {
		t.Fatal("accepted histogram crossing reset")
	}
	if e = artifact("histogram-reset-negative.txt", []byte(negative.Summary())); e != nil {
		t.Fatal(e)
	}
	steps[1].Obs["known"] = false
	steps[1].Obs["count"] = int64(-1)
	steps[6].Obs["known"] = true
	steps[6].Obs["count"] = int64(6)
	negative, e = checker.Check(context.Background(), steps)
	if e != nil {
		t.Fatal(e)
	}
	if negative.OK() {
		t.Fatal("accepted incompatible merge")
	}
	if e = artifact("histogram-merge-negative.txt", []byte(negative.Summary())); e != nil {
		t.Fatal(e)
	}
}
