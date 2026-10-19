package verification

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

func TestQuintNumericLiveLedger(t *testing.T) {
	name := os.Getenv("GOO_NUMERIC_LEDGER")
	if name == "" {
		t.Skip("live ledger opt-in")
	}
	data, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	type ledger struct {
		Plan                                           exploration.PackagePlan
		Result, Reused, Unsigned, Logs, ResourceResult exploration.Result
		Saved, Reopened                                exploration.CapturePage
		Values                                         []*string
		Times                                          []string
		Before, AfterLocal, BeforeReopen, AfterReopen  int
	}
	check := func(mutation int) bool {
		var l ledger
		if json.Unmarshal(data, &l) != nil {
			t.Fatal("invalid ledger")
		}
		switch mutation {
		case 0:
			l.Plan.MeasurePlan.Numeric.Unit = "s"
		case 1:
			l.Result.Numeric.Valid = "128"
		case 2:
			l.AfterLocal++
		case 3:
			l.Saved.Explore.Numeric.Invalid = "0"
		case 4:
			l.Reused.Question.Numeric.Parse = "or-zero"
		case 5:
			l.AfterReopen++
		case 6:
			l.Reopened.Explore.Numeric.Sum = "0"
		case 7:
			l.Result.GroupedTimeSeries[0].Points[0].ToMS++
		case 8:
			l.Unsigned.Numeric.Sum = "0"
		case 9:
			l.Logs.Numeric.Empty = "0"
		case 10:
			l.ResourceResult.Numeric.Unit = "ms"
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		record := func(a string, v bool) { rec.Record(context.Background(), a, nil, "obs.coherent", v) }
		expected, e := l.Plan.Question.RecordMeasures()
		record("plan", e == nil && reflect.DeepEqual(expected, l.Plan.MeasurePlan) && l.Plan.Question.Numeric != nil)
		valid := func(r *exploration.Result) bool {
			if r == nil || r.Question.Numeric == nil || r.Question.Validate() != nil || exploration.ValidateNumericResult(r) != nil || exploration.ValidateTimeSeries(r.Question, r.TimeSeries) != nil || exploration.ValidateGroupedTimeSeries(r.Question, r.GroupedTimeSeries) != nil || r.Count != "128" || r.State != "observed" || len(l.Values) != 128 || len(l.Times) != 128 {
				return false
			}
			b := *r.Question.Numeric
			values := l.Values
			if b.Field.Kind == "resource" {
				values = make([]*string, len(l.Values))
				for i := range values {
					v := "-7"
					values[i] = &v
				}
			}
			want, e := exploration.EvaluateNumericValues(b, values)
			if e != nil || r.Numeric == nil || *r.Numeric != want {
				return false
			}
			reference := func(group string, from, to int64) exploration.NumericSummary {
				selected := []*string{}
				for i, v := range values {
					label := "/edge"
					if i >= 18 {
						label = "/batch"
					}
					at, e := strconv.ParseInt(l.Times[i], 10, 64)
					if e != nil {
						return exploration.NumericSummary{}
					}
					at /= 1000000
					if (group == "" || group == label) && at >= from && at < to {
						selected = append(selected, v)
					}
				}
				s, _ := exploration.EvaluateNumericValues(b, selected)
				return s
			}
			for _, p := range r.TimeSeries {
				s := reference("", p.FromMS, p.ToMS)
				if p.Numeric == nil || *p.Numeric != s {
					return false
				}
			}
			if len(r.Groups) != 2 || len(r.GroupedTimeSeries) != 2 {
				return false
			}
			for _, raw := range r.Groups {
				var g struct {
					G0      string                      `json:"g0"`
					Numeric *exploration.NumericSummary `json:"numeric"`
				}
				if json.Unmarshal(raw, &g) != nil || g.Numeric == nil {
					return false
				}
				s := reference(g.G0, r.Question.FromMS, r.Question.ToMS)
				if *g.Numeric != s {
					return false
				}
			}
			for _, g := range r.GroupedTimeSeries {
				if len(g.Values) != 1 || g.Values[0] == nil {
					return false
				}
				for _, p := range g.Points {
					s := reference(*g.Values[0], p.FromMS, p.ToMS)
					if p.Numeric == nil || *p.Numeric != s {
						return false
					}
				}
			}
			return true
		}
		record("execute", valid(&l.Result) && valid(&l.Unsigned) && valid(&l.Logs) && valid(&l.ResourceResult) && reflect.DeepEqual(l.Result.Question, l.Plan.Question))
		record("inspect", l.Before == l.AfterLocal)
		record("capture", valid(l.Saved.Explore) && reflect.DeepEqual(l.Saved.Receipt.Question.Explore, &l.Result.Question))
		reuseQ := l.Reused.Question
		reuseQ.FromMS = l.Result.Question.FromMS
		reuseQ.ToMS = l.Result.Question.ToMS
		record("reuse", valid(&l.Reused) && reflect.DeepEqual(reuseQ, l.Result.Question) && l.Reused.Question.ToMS >= l.Result.Question.ToMS)
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
		t.Fatal("actual numeric ledger rejected")
	}
	for i := 0; i < 11; i++ {
		if check(i) {
			t.Fatal("numeric mutation accepted", i)
		}
	}
}
