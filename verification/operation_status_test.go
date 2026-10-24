package verification

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/annotations"
	"github.com/casselc/goo/investigations"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
	"os"
	"os/exec"
	"reflect"
	"testing"
)

func TestHegelOperationStatus(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		outcome := hegel.Draw(tc, hegel.Integers(0, 2))
		status := hegel.Draw(tc, hegel.Integers(0, 5))
		args, _ := json.Marshal([]int{outcome, status})
		script := `const a=require('node:assert/strict'),{waterfallRows,waterfallFacts}=require('../ui/operations');const[o,s]=JSON.parse(process.argv[1]),raw=[undefined,'','Ok','Error','Unset','Vendor'][s],span={durationNS:'1',outcome:['ok','error','panic'][o],...(raw===undefined?{}:{reportedStatus:raw})},before=JSON.stringify(span),row=waterfallRows([span],'worker')[0],facts=waterfallFacts([span])[0];a.equal(row.StatusCode,raw);a.equal(Object.hasOwn(row,'StatusCode'),raw!==undefined);a.equal(facts[0].value,span.outcome);a.equal(facts[1].value,raw===undefined?'Not retained':raw===''?'Unknown (empty)':raw==='Vendor'?'Unknown (Vendor)':raw);a.equal(JSON.stringify(span),before);`
		if out, e := exec.Command("node", "-e", script, string(args)).CombinedOutput(); e != nil {
			tc.Fatalf("projection: %s %v", out, e)
		}
	}, options(60)...)
}

func TestQuintOperationStatusLedger(t *testing.T) {
	name := os.Getenv("GOO_OPERATION_STATUS_LEDGER")
	if name == "" {
		t.Skip("live ledger opt-in")
	}
	data, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Old, Fresh, AfterOld investigations.Capture
			Oracle               []struct{ TraceID, ReportedStatus, Outcome string }
			Colors               []struct {
				TraceID string
				Error   bool
			}
			Events []struct {
				Action   string
				Reads    int
				Coherent bool
			}
			Errors []string
		}
		if e := json.Unmarshal(data, &l); e != nil {
			t.Fatal(e)
		}
		var old, fresh annotations.LiveOperations
		if e := json.Unmarshal(l.Old.Evidence.Operations, &old); e != nil {
			t.Fatal(e)
		}
		if e := json.Unmarshal(l.Fresh.Evidence.Operations, &fresh); e != nil {
			t.Fatal(e)
		}
		switch mutation {
		case 0:
			l.Events[0].Reads++
		case 1:
			raw := "Error"
			old.Spans[0].ReportedStatus = &raw
		case 2:
			raw := "Wrong"
			fresh.Spans[0].ReportedStatus = &raw
		case 3:
			l.Colors[0].Error = !l.Colors[0].Error
		case 4:
			l.Fresh.Question.ToMS++
		}
		coherent := reflect.DeepEqual(l.Old, l.AfterOld) && l.Fresh.Parent == l.Old.ID && reflect.DeepEqual(l.Old.Question, l.Fresh.Question) && reflect.DeepEqual(old.Conformance.Operations, fresh.Conformance.Operations) && old.Conformance.Basis != fresh.Conformance.Basis && len(l.Errors) == 0 && len(old.Spans) == 6 && len(fresh.Spans) == 6 && len(l.Oracle) == 6 && len(l.Colors) == 6
		for _, s := range old.Spans {
			coherent = coherent && s.ReportedStatus == nil
		}
		for _, s := range fresh.Spans {
			found := false
			for _, o := range l.Oracle {
				if o.TraceID == s.TraceID {
					found = true
					coherent = coherent && s.ReportedStatus != nil && *s.ReportedStatus == o.ReportedStatus && s.Outcome == o.Outcome
				}
			}
			coherent = coherent && found
		}
		for _, c := range l.Colors {
			found := false
			for _, o := range l.Oracle {
				if o.TraceID == c.TraceID {
					found = true
					coherent = coherent && c.Error == (o.ReportedStatus == "Error")
				}
			}
			coherent = coherent && found
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, ev := range l.Events {
			rec.Record(context.Background(), ev.Action, nil, "obs.reads", ev.Reads, "obs.coherent", coherent && ev.Coherent)
		}
		checker, e := validate.NewChecker("model/question_links.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("observed status trace rejected")
	}
	for i := 0; i < 5; i++ {
		if check(i) {
			t.Fatalf("mutation %d accepted", i)
		}
	}
}
