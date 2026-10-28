package verification

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
	"os"
	"os/exec"
	"testing"
)

func TestHegelQuestionLinkRoundtrip(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		from := hegel.Draw(tc, hegel.Integers(0, 1000000))
		width := hegel.Draw(tc, hegel.Integers(1, 7200000))
		variant := hegel.Draw(tc, hegel.Integers(0, 3))
		args, _ := json.Marshal([]int{from, width, variant})
		script := `const assert=require('node:assert/strict');const {questionLink,readQuestionLink}=require('../ui/workspace.mjs');const {Session}=require('../ui/question-session.mjs');const [from,width,v]=JSON.parse(process.argv[1]);const q={signal:v%2?'logs':'spans',environment:'prod 雪',fromMs:from,toMs:from+width,limit:100,offset:0,where:{field:{kind:'attribute',key:'region'},op:v===3?'missing':'eq',...(v===3?{}:{value:['',"literal'\\n",'雪 &?#'][v]})}};const events=[],s=new Session(e=>events.push(e));const restored=readQuestionLink(questionLink(q,v%2?'Distribution':'Records'));assert.deepEqual(restored.question,q);s.stage(restored.question);assert.equal(s.snapshot().phase,'draft');assert.deepEqual(s.snapshot().bounds,{fromMs:from,toMs:from+width});assert.equal(events.some(e=>e.action==='start'),false);const g=s.start(q);s.stage(q);assert.equal(s.publish(g,{question:q}),false);`
		if out, e := exec.Command("node", "-e", script, string(args)).CombinedOutput(); e != nil {
			tc.Fatalf("actual navigation JavaScript: %s %v", out, e)
		}
	}, options(100)...)
}

func TestQuintQuestionLinksLive(t *testing.T) {
	name := os.Getenv("GOO_QUESTION_LINKS_LEDGER")
	if name == "" {
		t.Skip("live ledger opt-in")
	}
	data, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Question exploration.Query
			Ledger   []struct {
				Event    string
				Reads    int
				Question *exploration.Query
			}
		}
		if e := json.Unmarshal(data, &l); e != nil {
			t.Fatal(e)
		}
		if mutation == 0 {
			l.Ledger[0].Reads++
		}
		if mutation == 1 {
			for i := range l.Ledger {
				if l.Ledger[i].Event == "read" {
					l.Ledger[i].Question.ToMS++
				}
			}
		}
		if mutation == 2 {
			l.Ledger[len(l.Ledger)-1].Reads++
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		reads := 0
		for _, event := range l.Ledger {
			coherent := l.Question.Validate() == nil
			if event.Event == "read" {
				reads++
				coherent = coherent && event.Question != nil && event.Question.Validate() == nil && event.Question.FromMS == l.Question.FromMS && event.Question.ToMS == l.Question.ToMS && event.Question.Environment == l.Question.Environment
			} else {
				reads = event.Reads
			}
			rec.Record(context.Background(), event.Event, nil, "obs.reads", reads, "obs.coherent", coherent)
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
		t.Fatal("live navigation trace rejected")
	}
	for i := 0; i < 3; i++ {
		if check(i) {
			t.Fatal("navigation mutation accepted", i)
		}
	}
}
