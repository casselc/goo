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

func TestQuintRecordPopulationLedger(t *testing.T) {
	name := os.Getenv("GOO_RECORD_LEDGER")
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
			Provenance                                             struct {
				Question     exploration.Query
				Basis        string
				Observations []struct {
					PageIndex       int
					TimestampNS     string `json:"timestamp_ns"`
					TraceID, SpanID string
				}
			}
			Events []struct {
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
				coherent = !exploration.Match(q.Where, map[exploration.Field]string{}) && !exploration.Match(q.Where, map[exploration.Field]string{f: ""}) && exploration.Match(q.Where, map[exploration.Field]string{f: "west"}) && !exploration.Match(q.Where, map[exploration.Field]string{f: "east"})
			}
		}
		if mutation == 5 {
			l.Provenance.Question.Environment = "changed"
		}
		if mutation == 7 {
			l.Provenance.Observations[0].TimestampNS = "0"
		}
		if mutation == 6 {
			l.Provenance.Observations = append(l.Provenance.Observations, l.Provenance.Observations...)
		}
		coherent = coherent && reflect.DeepEqual(l.Provenance.Question, l.Initial[0]) && len(l.Provenance.Basis) == 64 && len(l.Provenance.Observations) == 2
		seen := map[int]bool{}
		for _, o := range l.Provenance.Observations {
			ns, err := strconv.ParseUint(o.TimestampNS, 10, 64)
			if err != nil || ns < uint64(l.Initial[0].FromMS)*1000000 || ns >= uint64(l.Initial[0].ToMS)*1000000 {
				coherent = false
			}
			if seen[o.PageIndex] || o.PageIndex < 0 || o.PageIndex >= l.Initial[0].Limit || o.TimestampNS == "" {
				coherent = false
			}
			seen[o.PageIndex] = true
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
		t.Fatal("observed record-population trace rejected")
	}
	for i := 0; i < 8; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}

func TestHegelSharedRecordFields(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		choice := hegel.Draw(tc, hegel.Integers(0, 4))
		same := hegel.Draw(tc, hegel.Integers(0, 1))
		script := `const {commonGroups}=require('../ui/attributes'),{selectionQueries}=require('../ui/population-workspace');const value=[undefined,'','west',null,false][Number(process.argv[1])],same=Number(process.argv[2]);const records=[{ServiceName:'api',Attributes:value===undefined?{}:{region:value}},{ServiceName:'worker',Attributes:same?(value===undefined?{}:{region:value}):{region:'other'}}];const before=JSON.stringify(records),fields=commonGroups(records).flatMap(g=>g.entries),items=fields.filter(e=>e.field.kind==='attribute'&&e.field.key==='region').map(e=>({signal:'spans',field:e.field,value:e.value,mode:'include'})),qs=selectionQueries([{signal:'spans',environment:'prod',fromMs:1000,toMs:2000,limit:100,offset:0}],items);if(JSON.stringify(records)!==before)throw Error('records changed');process.stdout.write(JSON.stringify({qs,items}));`
		raw, e := exec.Command("node", "-e", script, strconv.Itoa(choice), strconv.Itoa(same)).Output()
		if e != nil {
			tc.Fatal(e)
		}
		var result struct {
			Qs    []exploration.Query
			Items []json.RawMessage
		}
		if json.Unmarshal(raw, &result) != nil || len(result.Qs) != 1 {
			tc.Fatal(string(raw))
		}
		q := result.Qs[0]
		if _, e := exploration.Compile(q); e != nil {
			tc.Fatal(e)
		}
		want := same == 1 && (choice == 1 || choice == 2)
		if (len(result.Items) == 1) != want {
			tc.Fatal("inferred non-shared or non-string value", string(raw))
		}
		if want {
			value := "west"
			if choice == 1 {
				value = ""
			}
			if !exploration.Match(q.Where, map[exploration.Field]string{{Kind: "attribute", Key: "region"}: value}) || exploration.Match(q.Where, map[exploration.Field]string{}) {
				tc.Fatal("incorrect shared-field predicate")
			}
		}
	}, options(60)...)
}
