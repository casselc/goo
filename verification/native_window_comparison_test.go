package verification

import (
	"context"
	"encoding/json"
	"math/big"
	"os"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/metrics"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
)

func TestHegelNativeWindowComparison(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		a := hegel.Draw(tc, hegel.Integers(0, 1000000))
		b := hegel.Draw(tc, hegel.Integers(0, 1000000))
		width := hegel.Draw(tc, hegel.Integers(1, 1000000))
		base := int64(1790000000000000000)
		s := metrics.Stream{Resource: map[string]string{"service.name": "http", "service.instance.id": "one"}, Scope: "otel.http", Name: "calls", Unit: "{call}"}
		makeResult := func(n int) *metrics.CounterResult {
			r, e := metrics.EvaluateCounters(metrics.CounterRequest{FromNS: strconv.FormatInt(base, 10), ToNS: strconv.FormatInt(base+int64(width), 10), MaxGapNS: "1000000000", Points: []metrics.CounterPoint{{Stream: s, Temporality: "delta", Monotonic: true, StartNS: strconv.FormatInt(base, 10), EndNS: strconv.FormatInt(base+int64(width), 10), Value: strconv.Itoa(n)}}})
			if e != nil {
				tc.Fatal(e)
			}
			return r
		}
		x, y := makeResult(a), makeResult(b)
		before, _ := json.Marshal([]*metrics.CounterResult{x, y})
		ab := metrics.CompareCounterWindows(x, y)
		ba := metrics.CompareCounterWindows(y, x)
		u, _ := new(big.Rat).SetString(ab[0].RateDifference)
		v, _ := new(big.Rat).SetString(ba[0].RateDifference)
		want := new(big.Rat).SetFrac(big.NewInt(int64(b-a)*1000000000), big.NewInt(int64(width)))
		if u.Cmp(want) != 0 || new(big.Rat).Add(u, v).Sign() != 0 {
			tc.Fatal("rate precision or reversal")
		}
		after, _ := json.Marshal([]*metrics.CounterResult{x, y})
		if string(before) != string(after) {
			tc.Fatal("mutated inputs")
		}
		makeHistogram := func(n int) *metrics.HistogramResult {
			stream := s
			stream.Unit = "s"
			stream.Name = "latency"
			r, e := metrics.EvaluateHistograms(metrics.HistogramRequest{FromNS: strconv.FormatInt(base, 10), ToNS: strconv.FormatInt(base+int64(width), 10), MaxGapNS: "1000000000", Points: []metrics.HistogramPoint{{Stream: stream, Temporality: "delta", StartNS: strconv.FormatInt(base, 10), EndNS: strconv.FormatInt(base+int64(width), 10), Bounds: []float64{1}, BucketCounts: []string{strconv.Itoa(n), strconv.Itoa(1000000 - n)}, Count: "1000000"}}})
			if e != nil {
				tc.Fatal(e)
			}
			return r
		}
		hx, hy := makeHistogram(a), makeHistogram(b)
		hr := metrics.CompareHistogramWindows(hx, hy)
		sum := new(big.Rat)
		for _, bucket := range hr[0].Buckets {
			v, _ := new(big.Rat).SetString(bucket.DifferencePP)
			sum.Add(sum, v)
		}
		first, _ := new(big.Rat).SetString(hr[0].Buckets[0].DifferencePP)
		if sum.Sign() != 0 || first.Cmp(new(big.Rat).SetFrac(big.NewInt(int64(b-a)), big.NewInt(10000))) != 0 {
			tc.Fatal("histogram population/share conservation")
		}
		hy.Series[0].Intervals[0].Distribution.Bounds[0] = 2
		hr = metrics.CompareHistogramWindows(hx, hy)
		if hr[0].State != "incompatible" || len(hr[0].Buckets) != 0 {
			tc.Fatal("incompatible histogram layout was rebinned")
		}
		y.Series[0].Intervals[0].State = "unknown"
		y.Series[0].Intervals[0].Reason = "counter-epoch-change"
		y.Series[0].State = "unknown"
		r := metrics.CompareCounterWindows(x, y)
		if r[0].RateDifference != "" || r[0].State != "unavailable" {
			tc.Fatal("reset accepted")
		}
	}, options(100)...)
}
func TestQuintNativeWindowsLive(t *testing.T) {
	file := os.Getenv("GOO_NATIVE_WINDOWS_LEDGER")
	if file == "" {
		t.Skip("live ledger opt-in")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Events []struct {
				Action  string
				Reads   int
				Allowed bool
			}
			Counter, Histogram, Reset, Layout, Gap exploration.EvidenceComparison
			Receipt                                workbench.ComparisonReceipt
			Before, After                          int
		}
		if json.Unmarshal(raw, &l) != nil {
			t.Fatal("bad ledger")
		}
		if mutation == 0 {
			l.Events[1].Reads++
		}
		if mutation == 1 {
			l.Reset.NativeWindows[0].RateDifference = "1"
		}
		if mutation == 2 {
			l.Histogram.NativeWindows[0].Buckets[0].DifferencePP = "0"
		}
		if mutation == 3 {
			l.Layout.State = "comparable"
		}
		if mutation == 4 {
			l.Receipt.Algorithm = "goo/evidence-comparison/v1"
		}
		if mutation == 5 {
			l.Gap.NativeWindows[0].RateDifference = "1"
		}
		coherent := l.Before == l.After && l.Counter.State == "comparable" && l.Counter.NativeWindows[0].RateDifference == "2" && l.Histogram.State == "comparable" && l.Histogram.NativeWindows[0].Buckets[0].DifferencePP == "-100" && l.Histogram.NativeWindows[0].P95Relation == "selection-range-higher" && l.Reset.State == "unavailable" && l.Reset.NativeWindows[0].RateDifference == "" && l.Gap.State == "unavailable" && l.Gap.NativeWindows[0].RateDifference == "" && l.Layout.State == "incompatible" && len(l.Layout.NativeWindows[0].Buckets) == 0 && l.Receipt.Algorithm == "goo/evidence-comparison/v2" && l.Receipt.Result.NativeWindows[0].RateDifference == "2"
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, event := range l.Events {
			rec.Record(context.Background(), event.Action, nil, "obs.reads", event.Reads, "obs.allowed", event.Allowed, "obs.coherent", coherent)
		}
		checker, e := validate.NewChecker("model/native_windows.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("observed trace rejected")
	}
	for i := 0; i < 6; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}
