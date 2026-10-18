package verification

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestQuintFollowupsLiveLedger(t *testing.T) {
	name := os.Getenv("GOO_FOLLOWUPS_LEDGER")
	if name == "" {
		t.Skip("live ledger opt-in")
	}
	data, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	type ledger struct {
		Origin, Resource, LogCapture                           exploration.CapturePage
		Plans                                                  exploration.FollowupSet
		Duration, Trace, Logs                                  exploration.Result
		Discovery                                              measurements.DiscoveryResult
		First, Second, Final, Reopened                         workbench.Document
		BeforePreview, AfterPreview, BeforeReopen, AfterReopen int
	}
	check := func(mutation int) bool {
		var l ledger
		if e = json.Unmarshal(data, &l); e != nil {
			t.Fatal(e)
		}
		switch mutation {
		case 0:
			l.AfterPreview++
		case 1:
			l.Duration.Question.Environment = "foreign"
		case 2:
			l.Trace.Question.Where.Value = strings.Repeat("f", 32)
		case 3:
			l.Logs.Question.ToMS++
		case 4:
			l.Discovery.Question.Service = "foreign"
		case 5:
			l.Final.OpenQuestions = "no citation"
		case 6:
			l.Reopened.Finding = "changed"
		case 7:
			l.Plans.Actions[0].Transform.Target.Grain = "root-request"
		case 8:
			l.Plans.Actions[0].Transform.Relation = "service-time"
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		record := func(a string, v bool) { rec.Record(context.Background(), a, nil, "obs.coherent", v) }
		actions := map[string]exploration.Followup{}
		for _, a := range l.Plans.Actions {
			actions[a.ID] = a
		}
		q := l.Origin.Receipt.Question.Explore
		expected, e := (exploration.SavedCapture{CaptureReceipt: l.Origin.Receipt, Explore: l.Origin.Explore}).Followups(0)
		record("preview", e == nil && reflect.DeepEqual(expected, &l.Plans) && l.Plans.Origin == l.Origin.Receipt.ID && l.BeforePreview == l.AfterPreview && q != nil)
		record("duration", reflect.DeepEqual(actions["duration-status"].Query, &l.Duration.Question) && reflect.DeepEqual(l.Duration.Question.Where, q.Where) && l.Duration.Count == l.Origin.Explore.Count && l.Duration.SpanStatus != nil)
		record("trace", reflect.DeepEqual(actions["trace"].Query, &l.Trace.Question) && l.Trace.Question.Where != nil && l.Trace.Question.Where.Value == l.Plans.Observations[0].TraceID)
		record("logs", reflect.DeepEqual(actions["trace-logs"].Query, &l.Logs.Question) && l.Logs.Count != "0" && reflect.DeepEqual(l.LogCapture.Receipt.Question.Explore, &l.Logs.Question))
		resourceAction := actions["resources-"+l.Plans.Observations[0].Service]
		m := l.Resource.Measurement
		record("resources", reflect.DeepEqual(resourceAction.Discovery, &l.Discovery.Question) && m != nil && m.Question.FromMS == q.FromMS && m.Question.ToMS == q.ToMS && m.Question.Definition.Environment == q.Environment && measurements.ValidateResult(m) == nil)
		has := func(id string) bool {
			for _, c := range l.Final.Cards {
				if c.Reference == id {
					return true
				}
			}
			return false
		}
		record("gap", l.Final.Parent == l.Second.ID && l.Second.Parent == l.First.ID && l.Final.Question == l.First.Question && l.Final.Finding != "" && strings.Contains(l.Final.OpenQuestions, l.Origin.Receipt.ID) && has(l.Origin.Receipt.ID) && has(l.LogCapture.Receipt.ID) && has(l.Resource.Receipt.ID))
		record("reopen", reflect.DeepEqual(l.Final, l.Reopened) && l.BeforeReopen == l.AfterReopen)
		checker, e := validate.NewChecker("model/followups.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("actual follow-up ledger rejected")
	}
	for i := 0; i < 9; i++ {
		if check(i) {
			t.Fatal("mutated follow-up ledger accepted", i)
		}
	}
}
