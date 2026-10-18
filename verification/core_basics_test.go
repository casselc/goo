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

type durationMachine struct{ values []uint64 }

func (m *durationMachine) RuleAppend(tc hegel.TestCase) {
	v := uint64(hegel.Draw(tc, hegel.Integers(0, 1000000000)))
	if hegel.Draw(tc, hegel.Integers(0, 10)) == 0 {
		v = 18446744073709551615
	}
	m.values = append(m.values, v)
}
func (m *durationMachine) RuleReset(tc hegel.TestCase) { m.values = nil }
func (m *durationMachine) RuleReverse(tc hegel.TestCase) {
	for i, j := 0, len(m.values)-1; i < j; i, j = i+1, j-1 {
		m.values[i], m.values[j] = m.values[j], m.values[i]
	}
}
func (m *durationMachine) InvariantSummary(tc hegel.TestCase) {
	d := exploration.ReferenceDurationSummary(m.values)
	if d.Validate() != nil {
		panic("invalid reference summary")
	}
	sum := new(big.Int)
	for _, v := range m.values {
		sum.Add(sum, new(big.Int).SetUint64(v))
	}
	if d.SumNS != sum.String() {
		panic("sum overflow")
	}
	values := append([]uint64(nil), m.values...)
	for i, j := 0, len(values)-1; i < j; i, j = i+1, j-1 {
		values[i], values[j] = values[j], values[i]
	}
	if !reflect.DeepEqual(d, exploration.ReferenceDurationSummary(values)) {
		panic("order changed summary")
	}
	q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61001, Limit: 1, Summary: "span-duration", BucketMS: 1000}
	p, e := exploration.Compile(q)
	if e != nil {
		panic(e)
	}
	q.Limit = 100
	q.Offset = 1000
	next, e := exploration.Compile(q)
	if e != nil || p.Count != next.Count || p.TimeSeries != next.TimeSeries {
		panic("inspection page changed measures")
	}
}
func TestHegelDurationSummaryStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		hegel.RunStateful(tc, &durationMachine{}, hegel.WithStatefulStepCount(30), hegel.WithAlwaysCheckInvariants("InvariantSummary"))
	}, options(100)...)
}

type basicsLedger struct {
	Result    exploration.Result           `json:"result"`
	Saved     exploration.CapturePage      `json:"saved"`
	Reused    exploration.Result           `json:"reused"`
	Discovery exploration.ServiceDiscovery `json:"discovery"`
	Cancelled bool                         `json:"cancelled"`
}

func TestQuintCoreBasicsLiveLedger(t *testing.T) {
	name := os.Getenv("GOO_CORE_BASICS_LEDGER")
	if name == "" {
		t.Skip("live ledger opt-in")
	}
	b, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	var l basicsLedger
	if e = json.Unmarshal(b, &l); e != nil {
		t.Fatal(e)
	}
	check := func(l basicsLedger) (bool, string) {
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink, qobs.WithActor("core-basics"))
		a := l.Result
		d := a.Duration
		measure := d != nil && d.Validate() == nil && d.Count == a.Count && a.Count == "205" && d.SumNS == "21115000000" && d.P95NS == "195000000" && len(a.Rows) == 100 && exploration.ValidateTimeSeries(a.Question, a.TimeSeries) == nil
		sum := new(big.Int)
		for _, p := range a.TimeSeries {
			n, _ := new(big.Int).SetString(p.Count, 10)
			if n == nil {
				measure = false
				continue
			}
			sum.Add(sum, n)
		}
		measure = measure && sum.String() == a.Count
		rec.Record(context.Background(), "measure", nil, "obs.coherent", measure)
		save := l.Saved.Explore != nil && l.Saved.Explore.Duration != nil && reflect.DeepEqual(l.Saved.Explore.Duration, d) && l.Saved.Receipt.Question.Explore.Summary == "span-duration" && l.Saved.Receipt.Question.Explore.BucketMS == 60000
		rec.Record(context.Background(), "save", nil, "obs.coherent", save)
		expectedQuestion := *l.Saved.Receipt.Question.Explore
		expectedQuestion.FromMS = l.Reused.Question.FromMS
		expectedQuestion.ToMS = l.Reused.Question.ToMS
		reuse := reflect.DeepEqual(l.Reused.Question, expectedQuestion) && l.Reused.Duration != nil && reflect.DeepEqual(l.Reused.Duration, d) && l.Reused.Question.Summary == "span-duration" && l.Reused.Question.BucketMS == 60000 && l.Reused.Question.FromMS != l.Saved.Receipt.Question.Explore.FromMS
		rec.Record(context.Background(), "reuse", nil, "obs.coherent", reuse)
		edge, ambiguous, unresolved, logOnly := 0, 0, 0, false
		for _, s := range l.Discovery.Signals {
			if s.Service == "log-only" && s.Signal == "logs" && s.Count == "1" {
				logOnly = true
			}
		}
		valid := true
		for _, x := range l.Discovery.Dependencies {
			n, _ := strconv.Atoi(x.Count)
			switch x.Kind {
			case "edge":
				if x.Parent != "api" || x.Child != "worker" {
					valid = false
				}
				edge += n
			case "ambiguous":
				ambiguous += n
			case "unresolved":
				unresolved += n
			}
		}
		rec.Record(context.Background(), "discover", nil, "obs.coherent", valid && edge == 1 && ambiguous == 1 && unresolved == 1 && logOnly)
		rec.Record(context.Background(), "cancel", nil, "obs.coherent", l.Cancelled)
		checker, e := validate.NewChecker("model/core_basics.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
	artifact("core-basics-live.txt", []byte(summary))
	if !ok {
		t.Fatal(summary)
	}
	for i := 0; i < 5; i++ {
		var bad basicsLedger
		json.Unmarshal(b, &bad)
		switch i {
		case 0:
			bad.Result.Duration.P95NS = "196000000"
		case 1:
			bad.Saved.Explore.Duration.SumNS = "21000000000"
		case 2:
			for j := range bad.Discovery.Dependencies {
				if bad.Discovery.Dependencies[j].Kind == "ambiguous" {
					bad.Discovery.Dependencies[j].Kind = "edge"
				}
			}
		case 3:
			bad.Cancelled = false
		case 4:
			bad.Reused.Question.Environment = "other-environment"
		}
		ok, summary := check(bad)
		artifact(fmt.Sprintf("core-basics-negative-%d.txt", i), []byte(summary))
		if ok {
			t.Fatal("negative control accepted", i)
		}
	}
}
