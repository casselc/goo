package verification

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/casselc/goo/metrics"
	"hegel.dev/go/hegel"
)

// Each slot is an application export with independent ten-nanosecond support.
// The oracle reasons about source supports/populations, not helper output or
// the aggregation implementation's coverage function.
type timeExport struct {
	a, b    uint64
	present bool
	layout  bool
	unknown string
}
type histogramTimeMachine struct {
	origin  int64
	width   int64
	exports [6]timeExport
	reverse bool
}

func newHistogramTimeMachine(tc hegel.TestCase) *histogramTimeMachine {
	m := &histogramTimeMachine{origin: int64(hegel.Draw(tc, hegel.Integers(100, 1000000))), width: 20}
	for i := range m.exports {
		m.exports[i] = timeExport{a: uint64(hegel.Draw(tc, hegel.Integers(0, 1000))), b: uint64(hegel.Draw(tc, hegel.Integers(0, 1000))), present: true}
		if hegel.Draw(tc, hegel.Integers(0, 3)) == 0 {
			m.exports[i].a, m.exports[i].b = 0, 0
		}
	}
	return m
}
func (m *histogramTimeMachine) slot(tc hegel.TestCase) int {
	return hegel.Draw(tc, hegel.Integers(0, len(m.exports)-1))
}
func (m *histogramTimeMachine) RuleWithdraw(tc hegel.TestCase) { m.exports[m.slot(tc)].present = false }
func (m *histogramTimeMachine) RuleRestore(tc hegel.TestCase)  { m.exports[m.slot(tc)].present = true }
func (m *histogramTimeMachine) RuleKnownZero(tc hegel.TestCase) {
	i := m.slot(tc)
	m.exports[i].a = 0
	m.exports[i].b = 0
	m.exports[i].present = true
	m.exports[i].unknown = ""
}
func (m *histogramTimeMachine) RuleLayoutChange(tc hegel.TestCase) {
	i := m.slot(tc)
	m.exports[i].layout = !m.exports[i].layout
}
func (m *histogramTimeMachine) RuleUnknown(tc hegel.TestCase) {
	i := m.slot(tc)
	m.exports[i].present = true
	m.exports[i].unknown = "no-recorded-value"
}
func (m *histogramTimeMachine) RuleResetDiagnostic(tc hegel.TestCase) {
	i := m.slot(tc)
	m.exports[i].present = true
	m.exports[i].unknown = "histogram-epoch-change"
}
func (m *histogramTimeMachine) RuleRepair(tc hegel.TestCase) {
	i := m.slot(tc)
	m.exports[i].present = true
	m.exports[i].unknown = ""
	m.exports[i].layout = false
}
func (m *histogramTimeMachine) RuleIncrease(tc hegel.TestCase) {
	i := m.slot(tc)
	m.exports[i].a += uint64(hegel.Draw(tc, hegel.Integers(0, 100)))
	m.exports[i].b += uint64(hegel.Draw(tc, hegel.Integers(0, 100)))
}
func (m *histogramTimeMachine) RuleWidth(tc hegel.TestCase) {
	m.width = int64(hegel.Draw(tc, hegel.SampledFrom([]int{10, 15, 20, 30, 60})))
}
func (m *histogramTimeMachine) RuleArrivalOrder(tc hegel.TestCase) { m.reverse = !m.reverse }

func (m *histogramTimeMachine) evaluated() (*metrics.HistogramResult, error) {
	// An out-of-window prior export retains instrument identity even if all
	// in-window exports are withdrawn. It contributes no measured population.
	before := histogramPoint(m.origin-10, m.origin, 0, 0)
	before.Temporality = "delta"
	points := []metrics.HistogramPoint{before}
	for i, e := range m.exports {
		if !e.present {
			continue
		}
		p := histogramPoint(m.origin+int64(i)*10, m.origin+int64(i+1)*10, e.a, e.b)
		p.Temporality = "delta"
		if e.layout {
			p.Bounds = []float64{.02}
		}
		if e.unknown != "" {
			p.NoRecordedValue = true
		}
		points = append(points, p)
	}
	if m.reverse {
		for i, j := 0, len(points)-1; i < j; i, j = i+1, j-1 {
			points[i], points[j] = points[j], points[i]
		}
	}
	r, err := metrics.EvaluateHistograms(metrics.HistogramRequest{FromNS: fmt.Sprint(m.origin), ToNS: fmt.Sprint(m.origin + 60), MaxGapNS: "10", Points: points})
	if err != nil {
		return nil, err
	}
	// The bucketer consumes evaluated unknown intervals, including diagnostics
	// from cumulative epoch resets. Inject only that diagnostic: the real source
	// evaluator has already withheld this missing-value interval's distribution.
	for i := range r.Series[0].Intervals {
		interval := &r.Series[0].Intervals[i]
		from, _ := strconv.ParseInt(interval.FromNS, 10, 64)
		slot := int((from - m.origin) / 10)
		if m.exports[slot].unknown == "histogram-epoch-change" {
			interval.Reason = "histogram-epoch-change"
		}
	}
	return r, nil
}

func histogramTimeFail(format string, args ...any) { panic(fmt.Sprintf(format, args...)) }

func (m *histogramTimeMachine) InvariantWholeExportPopulation(tc hegel.TestCase) {
	r, err := m.evaluated()
	if err != nil {
		histogramTimeFail("source: %v", err)
	}
	before, _ := json.Marshal(r)
	out, err := metrics.BucketHistograms(r, fmt.Sprint(m.width))
	if err != nil {
		histogramTimeFail("bucket: %v", err)
	}
	after, _ := json.Marshal(r)
	if string(before) != string(after) {
		panic("time aggregation mutated retained input")
	}
	if len(out.Series) != 1 || !reflect.DeepEqual(out.Series[0].Stream, r.Series[0].Stream) || out.SourceBasis != r.Basis {
		panic("changed source identity/basis")
	}
	bins := out.Series[0].Buckets
	wantBins := int((60 + m.width - 1) / m.width)
	if len(bins) != wantBins {
		histogramTimeFail("invented/omitted temporal bucket: %d != %d", len(bins), wantBins)
	}
	for j, bin := range bins {
		lo := int64(j) * m.width
		hi := lo + m.width
		if hi > 60 {
			hi = 60
		}
		if bin.FromNS != fmt.Sprint(m.origin+lo) || bin.ToNS != fmt.Sprint(m.origin+hi) {
			panic("changed origin/half-open support")
		}
		known := true
		haveLayout := false
		layout := false
		sumA, sumB := uint64(0), uint64(0)
		for i, e := range m.exports {
			a, b := int64(i)*10, int64(i+1)*10
			if b <= lo || a >= hi {
				continue
			}
			if !e.present || e.unknown != "" || a < lo || b > hi {
				known = false
				continue
			}
			if haveLayout && layout != e.layout {
				known = false
			}
			haveLayout = true
			layout = e.layout
			sumA += e.a
			sumB += e.b
		}
		if !known {
			if bin.State != "unavailable" || bin.Distribution != nil {
				histogramTimeFail("partial/unknown export acquired a value in bucket %d: %+v", j, bin)
			}
			continue
		}
		wantState := "observed"
		if sumA+sumB == 0 {
			wantState = "no_data"
		}
		d := bin.Distribution
		if bin.State != wantState || d == nil || d.Count != fmt.Sprint(sumA+sumB) || !reflect.DeepEqual(d.BucketCounts, []string{fmt.Sprint(sumA), fmt.Sprint(sumB)}) {
			histogramTimeFail("whole-export population mismatch in bucket %d: %+v expected %d/%d", j, bin, sumA, sumB)
		}
		bound := .01
		if layout {
			bound = .02
		}
		if len(d.Bounds) != 1 || d.Bounds[0] != bound || d.Unit != "s" {
			panic("wrong supplied layout/unit")
		}
		if sumA+sumB == 0 && d.P95.State != "no_data" {
			panic("zero population gained a percentile")
		}
	}
	// Arrival order must not change counts or the derived policy/source basis.
	m.reverse = !m.reverse
	reordered, err := m.evaluated()
	m.reverse = !m.reverse
	if err != nil {
		panic(err)
	}
	again, err := metrics.BucketHistograms(reordered, fmt.Sprint(m.width))
	if err != nil || !reflect.DeepEqual(out, again) {
		histogramTimeFail("arrival order changed result: %v", err)
	}
}

func TestHegelHistogramTimeAlignedPopulation(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) { m := newHistogramTimeMachine(tc); m.InvariantWholeExportPopulation(tc) }, options(100)...)
}
func TestHegelHistogramTimeStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		m := newHistogramTimeMachine(tc)
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(20), hegel.WithAlwaysCheckInvariants("InvariantWholeExportPopulation"))
	}, options(100)...)
}
