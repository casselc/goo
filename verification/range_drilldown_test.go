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

func TestHegelBrushPopulationScope(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		a := hegel.Draw(tc, hegel.Integers(0, 98))
		b := hegel.Draw(tc, hegel.Integers(a+1, 100))
		script := `const {brushRange}=require('../ui/value-timeline.mjs'),{rangeQueries}=require('../ui/population-workspace');const qs=['spans','logs'].map(signal=>({signal,environment:'prod',fromMs:1000000,toMs:1300000,limit:100,offset:0,where:{any:[{field:{kind:'attribute',key:'region'},op:'missing'},{field:{kind:'attribute',key:'region'},op:'eq',value:''}]}}));const before=JSON.stringify(qs),r=brushRange({fromNS:'1000000000000',toNS:'1300000000000'},Number(process.argv[1])/100,Number(process.argv[2])/100);const out=rangeQueries(qs,{fromMs:Number(BigInt(r.fromNS)/1000000n),toMs:Number((BigInt(r.toNS)+999999n)/1000000n)});if(JSON.stringify(qs)!==before)throw Error('parent mutated');process.stdout.write(JSON.stringify(out));`
		out, e := exec.Command("node", "-e", script, strconv.Itoa(a), strconv.Itoa(b)).Output()
		if e != nil {
			tc.Fatal(e)
		}
		var qs []exploration.Query
		if json.Unmarshal(out, &qs) != nil || len(qs) != 2 {
			tc.Fatal(string(out))
		}
		for _, q := range qs {
			if _, e := exploration.Compile(q); e != nil || q.Environment != "prod" || q.FromMS < 1000000 || q.ToMS > 1300000 || q.FromMS >= q.ToMS || len(q.Where.Any) != 2 {
				tc.Fatal(q, e)
			}
		}
		if !reflect.DeepEqual(qs[0].Where, qs[1].Where) || qs[0].FromMS != qs[1].FromMS || qs[0].ToMS != qs[1].ToMS {
			tc.Fatal("signals diverged")
		}
	}, options(80)...)
}
func TestQuintRangeDrilldownLedger(t *testing.T) {
	name := os.Getenv("GOO_RANGE_LEDGER")
	if name == "" {
		t.Skip("requires observed browser ledger")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Initial, ChildQueries                                         []exploration.Query
			Selected                                                      struct{ FromMS, ToMS int64 }
			BeforePreview, AfterPreview, BeforeReturn, AfterReturn, Tails int
			Baseline, AfterBaseline                                       string
			TailKeys, AfterTailKeys                                       []string
			Events                                                        []struct {
				Action       string
				Depth, Reads int
			}
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch mutation {
		case 0:
			l.AfterPreview++
		case 1:
			l.AfterReturn++
		case 2:
			l.ChildQueries[0].FromMS = l.Initial[0].FromMS - 1
		case 3:
			l.AfterBaseline = "changed"
		case 4:
			l.AfterTailKeys = nil
		case 5:
			l.Tails++
		case 6:
			for i := range l.Events {
				if l.Events[i].Action == "back" {
					l.Events[i].Depth++
					break
				}
			}
		}
		coherent := len(l.Initial) == 2 && len(l.ChildQueries) == 2 && l.BeforePreview == l.AfterPreview && l.BeforeReturn == l.AfterReturn && l.Baseline == l.AfterBaseline && len(l.TailKeys) > 0 && reflect.DeepEqual(l.TailKeys, l.AfterTailKeys) && l.Tails == 1
		for i, q := range l.ChildQueries {
			if i >= len(l.Initial) {
				coherent = false
				break
			}
			p := l.Initial[i]
			_, e := exploration.Compile(q)
			coherent = coherent && e == nil && q.Signal == p.Signal && q.Environment == p.Environment && q.FromMS >= p.FromMS && q.ToMS <= p.ToMS && q.FromMS == l.Selected.FromMS && q.ToMS == l.Selected.ToMS && reflect.DeepEqual(q.Where, p.Where)
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
		t.Fatal("observed drilldown rejected")
	}
	for i := 0; i < 7; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}
