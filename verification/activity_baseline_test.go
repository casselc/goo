package verification

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"testing"
)

func TestHegelActivityBaselineAccounting(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		script := `const {capture,request}=require('../ui/activity-baseline.mjs');const n=9007199254740993n+BigInt(process.argv[1]),q={signal:'spans',environment:'test',fromMs:1700000000000,toMs:1700000060000,limit:100,offset:0,groupBy:[{kind:'column',key:'duration_bucket'}],where:{field:{kind:'attribute',key:'owner'},op:'missing'}},a={question:q,state:'observed',count:String(n+1n),basis:'baseline',capturedAt:'2026-10-24T00:00:01Z',groups:[{g0:'[0,1) ms',count:String(n)},{g0:'[5,10) ms',count:'1'}]},id='a'.repeat(64),frozen=capture({receipt:{id,title:'Prior duration',capturedAt:'2026-10-24T00:00:02Z',question:{kind:'explore',explore:q}},explore:a},id),current={...a,question:{...q,fromMs:q.toMs,toMs:q.toMs+60000,limit:1,where:{field:{kind:'attribute',key:'owner'},op:'eq',value:''}},basis:'activity',count:String(n+2n),groups:[{g0:'[5,10) ms',count:String(n)},{g0:'[10,50) ms',count:'2'}]};process.stdout.write(JSON.stringify(request(frozen,current)));`
		raw, e := exec.Command("node", "-e", script, strconv.Itoa(seed)).CombinedOutput()
		if e != nil {
			tc.Fatal(e, string(raw))
		}
		var pair exploration.ComparisonRequest
		if e = json.Unmarshal(raw, &pair); e != nil {
			tc.Fatal(e)
		}
		value, e := exploration.CompareCohorts(pair)
		if e != nil || value.State != "ok" {
			tc.Fatal(e, value)
		}
		body, e := json.Marshal(map[string]any{"pair": pair, "value": value})
		if e != nil {
			tc.Fatal(e)
		}
		validate := exec.Command("node", "-e", `const fs=require('node:fs'),assert=require('node:assert/strict'),{result}=require('../ui/activity-baseline.mjs'),{pair,value}=JSON.parse(fs.readFileSync(0,'utf8'));assert.deepEqual(result(pair,value),value);assert.throws(()=>result(pair,{...value,baselineBasis:'other'}));assert.throws(()=>result(pair,{...value,rows:value.rows.slice(1)}));`)
		validate.Stdin = bytes.NewReader(body)
		if out, e := validate.CombinedOutput(); e != nil {
			tc.Fatal(e, string(out))
		}
	}, options(40)...)
}

func TestQuintActivityBaselineLedger(t *testing.T) {
	name := os.Getenv("GOO_ACTIVITY_BASELINE_LEDGER")
	if name == "" {
		t.Skip("observed browser ledger")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Saved, Unchanged                             exploration.CapturePage
			First, Last                                  exploration.ComparisonRequest
			InitialReads, BeforeEmpty, AfterClear, Reads int
			Comparison                                   exploration.CohortComparison
		}
		if e = json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch mutation {
		case 0:
			l.BeforeEmpty++
		case 1:
			l.Last.Baseline.Count = "5"
		case 2:
			l.Last.Baseline.Question.ToMS = l.Last.Selection.Question.ToMS
		case 3:
			l.Last.Baseline.Basis = "other"
		case 4:
			l.Unchanged.Receipt.ID = "changed"
		}
		coherent := reflect.DeepEqual(l.Saved, l.Unchanged) && l.InitialReads == 1 && l.BeforeEmpty == l.AfterClear && l.Reads == l.BeforeEmpty && l.First.Baseline.Count == "4" && l.First.Selection.Count == "5" && l.Last.Selection.Count == "6" && reflect.DeepEqual(l.First.Baseline, l.Last.Baseline) && l.Last.Selection.Question.ToMS > l.First.Selection.Question.ToMS && l.Comparison.BaselineCount == "4" && l.Comparison.SelectionCount == "6"
		for _, pair := range []exploration.ComparisonRequest{l.First, l.Last} {
			value, e := exploration.CompareCohorts(pair)
			coherent = coherent && e == nil && value.State == "ok"
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for i := 1; i <= l.InitialReads; i++ {
			rec.Record(context.Background(), "read", nil, "obs.reads", i, "obs.coherent", coherent, "obs.chosen", false)
		}
		rec.Record(context.Background(), "choose", nil, "obs.reads", l.InitialReads, "obs.coherent", coherent, "obs.chosen", true)
		for i := l.InitialReads + 1; i <= l.AfterClear; i++ {
			rec.Record(context.Background(), "read", nil, "obs.reads", i, "obs.coherent", coherent, "obs.chosen", true)
		}
		rec.Record(context.Background(), "clear", nil, "obs.reads", l.BeforeEmpty, "obs.coherent", coherent, "obs.chosen", false)
		rec.Record(context.Background(), "choose", nil, "obs.reads", l.Reads, "obs.coherent", coherent, "obs.chosen", true)
		checker, e := validate.NewChecker("model/activity_baseline.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		report, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return report.OK()
	}
	if !check(-1) {
		t.Fatal("valid observed baseline ledger rejected")
	}
	for i := 0; i < 5; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}
