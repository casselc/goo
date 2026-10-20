package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/metrics"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
)

func TestQuintLiveCounterLedger(t *testing.T) {
	name := os.Getenv("GOO_LIVE_COUNTER_LEDGER")
	if name == "" {
		t.Skip("live opt-in")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Baseline, Frozen, Reopened exploration.CapturePage
			First, Gap, Repaired       struct {
				Comparison exploration.EvidenceComparison
				Basis      string
			}
			Initial, Displayed struct {
				Basis, ReadAt, FromNS, ToNS, Scale string
				Values                             []string
			}
			SourceLatest string
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch mutation {
		case 0:
			l.First.Comparison.NativeWindows[0].Selection.Rate = "10"
		case 1:
			l.Gap.Comparison.State = "comparable"
		case 2:
			l.Displayed.Scale = "5"
		case 3:
			l.Frozen.Measurement.Basis = "different"
		case 4:
			l.Reopened.Receipt.ID = "other"
		case 5:
			l.Frozen.Measurement.Question.ToNS = l.SourceLatest
		case 6:
			l.Repaired.Comparison.NativeWindows[0].Stream.Resource["service.instance.id"] = "other"
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		record := func(a string, b bool) { rec.Record(context.Background(), a, nil, "obs.coherent", b) }
		first := l.First.Comparison.NativeWindows[0]
		fixed := l.Repaired.Comparison.NativeWindows[0]
		record("observe", first.State == "comparable" && first.Baseline.Rate == "3" && first.Selection.Rate == "5" && first.Stream.Resource["service.instance.id"] == "one")
		record("axis", l.Gap.Comparison.State == "unavailable" && l.Displayed.Scale == "15" && reflect.DeepEqual(l.Displayed.Values, []string{"3", "15"}) && fixed.Stream.Resource["service.instance.id"] == first.Stream.Resource["service.instance.id"])
		record("stale", l.Displayed.Basis == l.Repaired.Basis)
		record("freeze", l.Frozen.Measurement.Basis == l.Repaired.Basis && l.Frozen.Measurement.Question.ToNS == l.Repaired.Comparison.Selection.Receipt.Question.Measurement.ToNS && l.SourceLatest != l.Frozen.Measurement.Question.ToNS)
		record("reopen", reflect.DeepEqual(l.Reopened, l.Frozen))
		record("discard", l.Baseline.Receipt.ID == l.First.Comparison.Baseline.Receipt.ID)
		c, e := validate.NewChecker("model/live_comparison.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		result, e := c.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return result.OK()
	}
	if !check(-1) {
		t.Fatal("actual native ledger rejected")
	}
	for i := 0; i < 7; i++ {
		if check(i) {
			t.Fatal("corrupt native trace accepted", i)
		}
	}
}
func TestHegelCounterWindowRateAndCoverageState(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		width := int64(hegel.Draw(tc, hegel.Integers(1, 1000000)))
		delta := hegel.Draw(tc, hegel.Integers(0, 100000))
		from := int64(1000000000)
		stream := metrics.Stream{Resource: map[string]string{"service.name": "api", "service.instance.id": "one"}, Scope: "otel.http", Name: "requests", Unit: "{request}"}
		evaluate := func(a, b int64) *metrics.CounterResult {
			r, e := metrics.EvaluateCounters(metrics.CounterRequest{FromNS: fmt.Sprint(from), ToNS: fmt.Sprint(from + width), MaxGapNS: "10000000000", Points: []metrics.CounterPoint{{Stream: stream, Temporality: "delta", Monotonic: true, StartNS: fmt.Sprint(a), EndNS: fmt.Sprint(b), Value: fmt.Sprint(delta)}}})
			if e != nil {
				tc.Fatal(e)
			}
			return r
		}
		whole := evaluate(from, from+width)
		same := metrics.CompareCounterWindows(whole, whole)
		if len(same) != 1 || same[0].State != "comparable" || same[0].RateDifference != "0" {
			tc.Fatal("complete exact window refused")
		}
		broken := evaluate(from-1, from+width)
		comparison := metrics.CompareCounterWindows(whole, broken)
		if comparison[0].State == "comparable" || comparison[0].RateDifference != "" {
			tc.Fatal("boundary-crossing interval asserted rate")
		}
	}, options(100)...)
}
