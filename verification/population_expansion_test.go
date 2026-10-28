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
	"reflect"
	"strconv"
	"testing"
)

func TestHegelPopulationExpansion(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		width := hegel.Draw(tc, hegel.Integers(1, 110))
		before := hegel.Draw(tc, hegel.Integers(1, 5))
		after := hegel.Draw(tc, hegel.Integers(0, 5))
		script := `const assert=require('node:assert/strict'),{widenQueries}=require('../ui/population-workspace.mjs');const [width,before,after]=process.argv.slice(1).map(Number),qs=['spans','logs'].map((signal,i)=>({signal,environment:'test',fromMs:7200000-width*60000,toMs:7200000,limit:1,offset:10,where:{field:{kind:i?'resource':'attribute',key:'region'},op:i?'missing':'eq',...(i?{}:{value:''})}})),saved=JSON.stringify(qs),out=widenQueries(qs,{fromMs:qs[0].fromMs-before*60000,toMs:qs[0].toMs+after*60000});assert.equal(JSON.stringify(qs),saved);for(let i=0;i<2;i++)assert.deepEqual(out[i],{...qs[i],fromMs:out[0].fromMs,toMs:out[0].toMs,offset:0});process.stdout.write(JSON.stringify(out));`
		raw, err := exec.Command("node", "-e", script, strconv.Itoa(width), strconv.Itoa(before), strconv.Itoa(after)).Output()
		if err != nil {
			tc.Fatal(err)
		}
		var qs []exploration.Query
		if json.Unmarshal(raw, &qs) != nil || len(qs) != 2 {
			tc.Fatal(string(raw))
		}
		for _, q := range qs {
			if _, err := exploration.Compile(q); err != nil {
				tc.Fatal(err)
			}
		}
	}, options(40)...)
}

func TestQuintPopulationExpansionLedger(t *testing.T) {
	name := os.Getenv("GOO_EXPANSION_LEDGER")
	if name == "" {
		t.Skip("requires observed browser ledger")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Initial, Expanded                                      []exploration.Query
			BeforePreview, AfterPreview, BeforeReturn, AfterReturn int
			Events                                                 []struct {
				Action       string
				Depth, Reads int
			}
		}
		if json.Unmarshal(raw, &l) != nil {
			t.Fatal("invalid ledger")
		}
		switch mutation {
		case 0:
			l.AfterPreview++
		case 1:
			l.AfterReturn++
		case 2:
			l.Expanded[1].Environment = "other"
		case 3:
			l.Expanded[0].Where = nil
		case 5:
			l.Expanded[0].FromMS = l.Initial[0].FromMS + 1
		case 4:
			for i := range l.Events {
				if l.Events[i].Action == "back" {
					l.Events[i].Depth++
					break
				}
			}
		}
		coherent := l.BeforePreview == l.AfterPreview && l.BeforeReturn == l.AfterReturn && len(l.Initial) == 2 && len(l.Expanded) == 2
		if coherent {
			for i, q := range l.Expanded {
				expected := l.Initial[i]
				expected.FromMS = l.Expanded[0].FromMS
				expected.ToMS = l.Expanded[0].ToMS
				expected.Offset = 0
				_, err := exploration.Compile(q)
				coherent = coherent && err == nil && reflect.DeepEqual(expected, q) && q.FromMS <= l.Initial[i].FromMS && q.ToMS >= l.Initial[i].ToMS && (q.FromMS < l.Initial[i].FromMS || q.ToMS > l.Initial[i].ToMS)
			}
		}

		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, event := range l.Events {
			rec.Record(context.Background(), event.Action, nil, "obs.reads", event.Reads, "obs.depth", event.Depth, "obs.coherent", coherent)
		}
		c, e := validate.NewChecker("model/population_drilldown.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		r, e := c.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return r.OK()
	}
	if !check(-1) {
		t.Fatal("observed expansion trace rejected")
	}
	for i := 0; i < 6; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}
