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

func TestQuintGroupedSeriesLiveLedger(t *testing.T) {
	name := os.Getenv("GOO_SERIES_LEDGER")
	if name == "" {
		t.Skip("live ledger opt-in")
	}
	b, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	type ledger struct {
		Plan                                          exploration.PackagePlan
		Result                                        exploration.Result
		Saved, Reopened                               exploration.CapturePage
		Reused                                        exploration.Result
		Before, AfterLocal, BeforeReopen, AfterReopen int
	}
	check := func(mutation int) bool {
		var l ledger
		if e = json.Unmarshal(b, &l); e != nil {
			t.Fatal(e)
		}
		switch mutation {
		case 0:
			l.Plan.MeasurePlan.Measures[1].Unit = "s"
		case 1:
			l.Result.GroupedTimeSeries[0].Points[0].ToMS++
		case 2:
			l.AfterLocal++
		case 3:
			l.Saved.Explore.GroupedTimeSeries[0].Values[0] = nil
		case 4:
			l.Reused.Question.Environment = "foreign"
		case 5:
			l.AfterReopen++
		case 6:
			l.Reopened.Explore.GroupedTimeSeries[0].Points[0].Count = "900"
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		record := func(a string, v bool) { rec.Record(context.Background(), a, nil, "obs.coherent", v) }
		expected, e := l.Plan.Question.RecordMeasures()
		record("plan", e == nil && reflect.DeepEqual(expected, l.Plan.MeasurePlan) && l.Plan.Question.GroupTimeSeries)
		valid := func(r *exploration.Result) bool {
			if r == nil || r.Question.Validate() != nil || !r.Question.GroupTimeSeries || exploration.ValidateGroupedTimeSeries(r.Question, r.GroupedTimeSeries) != nil {
				return false
			}
			if r.Count != "124" || len(r.GroupedTimeSeries) != 4 || r.Duration == nil || r.Duration.SumNS != "2500000000" || r.Duration.Count != r.Count || r.Duration.Validate() != nil || exploration.ValidateTimeSeries(r.Question, r.TimeSeries) != nil || r.State != "observed" {
				return false
			}
			sums := map[string]string{}
			for _, g := range r.GroupedTimeSeries {
				key, _ := json.Marshal(g.Values)
				n := new(big.Int)
				for _, p := range g.Points {
					count, ok := new(big.Int).SetString(p.Count, 10)
					if !ok {
						return false
					}
					n.Add(n, count)
				}
				sums[string(key)] = n.String()
			}
			return sums[`["/fast"]`] == "100" && sums[`["/slow"]`] == "20" && sums[`[null]`] == "2" && sums[`[""]`] == "2"
		}
		record("execute", valid(&l.Result) && reflect.DeepEqual(l.Result.Question, l.Plan.Question))
		record("inspect", l.Before == l.AfterLocal)
		record("capture", valid(l.Saved.Explore) && reflect.DeepEqual(l.Saved.Receipt.Question.Explore, &l.Result.Question))
		a, _ := (exploration.CaptureQuestion{Kind: "explore", Explore: &l.Result.Question}).Scope()
		next := l.Reused.Question
		next.FromMS = a.FromMS
		next.ToMS = a.ToMS
		b, _ := (exploration.CaptureQuestion{Kind: "explore", Explore: &next}).Scope()
		record("reuse", valid(&l.Reused) && reflect.DeepEqual(a, b) && l.Reused.Question.ToMS >= l.Result.Question.ToMS)
		record("reopen", reflect.DeepEqual(l.Saved, l.Reopened) && l.BeforeReopen == l.AfterReopen)
		checker, e := validate.NewChecker("model/grouped_series.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		r, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return r.OK()
	}
	if !check(-1) {
		t.Fatal("actual series ledger rejected")
	}
	for i := 0; i < 7; i++ {
		if check(i) {
			t.Fatal("mutated series ledger accepted", i)
		}
	}
}
