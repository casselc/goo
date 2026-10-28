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

func TestHegelFacetSelection(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		mode := hegel.Draw(tc, hegel.Integers(0, 1))
		choice := hegel.Draw(tc, hegel.Integers(0, 2))
		script := `const {selectionQueries}=require('../ui/population-workspace.mjs');const q=['spans','logs'].map(signal=>({signal,environment:'prod',fromMs:1000,toMs:2000,limit:100,offset:0,where:{any:[{field:{kind:'resource',key:'team'},op:'missing'},{field:{kind:'resource',key:'team'},op:'eq',value:''}]}}));const before=JSON.stringify(q);const out=selectionQueries(q,[{signal:'spans',field:{kind:'attribute',key:'region'},value:[null,'','west'][Number(process.argv[2])],mode:Number(process.argv[1])?'exclude':'include'}]);if(before!==JSON.stringify(q)||JSON.stringify(q[1])!==JSON.stringify(out[1]))throw Error('parent or companion changed');process.stdout.write(JSON.stringify(out));`
		raw, e := exec.Command("node", "-e", script, strconv.Itoa(mode), strconv.Itoa(choice)).Output()
		if e != nil {
			tc.Fatal(e)
		}
		var qs []exploration.Query
		if json.Unmarshal(raw, &qs) != nil || len(qs) != 2 {
			tc.Fatal(string(raw))
		}
		for _, q := range qs {
			if _, e := exploration.Compile(q); e != nil {
				tc.Fatal(e)
			}
		}
		f := exploration.Field{Kind: "attribute", Key: "region"}
		for i, v := range []string{"", "", "west", "east"} {
			record := map[exploration.Field]string{}
			if i > 0 {
				record[f] = v
			}
			want := i == choice
			if mode == 1 {
				want = !want
			}
			if got := exploration.Match(qs[0].Where, record); got != want {
				tc.Fatal("incorrect missing/empty exclusion", mode, choice, i, got, want)
			}
		}
	}, options(60)...)
}

func TestQuintFacetLedger(t *testing.T) {
	name := os.Getenv("GOO_FACET_LEDGER")
	if name == "" {
		t.Skip("requires observed browser ledger")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Initial, Selected                                      []exploration.Query
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
			l.Selected[1].Environment = "other"
		case 3:
			l.Selected[0].Where = nil
		case 4:
			for i := range l.Events {
				if l.Events[i].Action == "back" {
					l.Events[i].Depth++
					break
				}
			}
		}
		coherent := l.BeforePreview == l.AfterPreview && l.BeforeReturn == l.AfterReturn && len(l.Initial) == 2 && len(l.Selected) == 2
		if coherent {
			coherent = reflect.DeepEqual(l.Initial[1], l.Selected[1])
			q := l.Selected[0]
			_, e := exploration.Compile(q)
			coherent = coherent && e == nil && q.Environment == l.Initial[0].Environment && q.FromMS == l.Initial[0].FromMS && q.ToMS == l.Initial[0].ToMS && q.Where != nil
			if coherent {
				f := exploration.Field{Kind: "attribute", Key: "region"}
				coherent = exploration.Match(q.Where, map[exploration.Field]string{}) && exploration.Match(q.Where, map[exploration.Field]string{f: ""}) && !exploration.Match(q.Where, map[exploration.Field]string{f: "west"})
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
		t.Fatal("observed facet trace rejected")
	}
	for i := 0; i < 5; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}
