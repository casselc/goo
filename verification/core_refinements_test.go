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
	"strconv"
	"testing"
)

func TestHegelPopulationSelection(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		low := hegel.Draw(tc, hegel.Integers(0, 1000000))
		width := hegel.Draw(tc, hegel.Integers(1, 1000000))
		script := `const {question}=require('../ui/population-selection.mjs');const low=process.argv[1],high=process.argv[2];const q={signal:'spans',environment:'prod',fromMs:1000,toMs:10000,limit:100,offset:0,where:{any:[{field:{kind:'attribute',key:'route'},op:'missing'},{field:{kind:'attribute',key:'route'},op:'eq',value:''}]}};process.stdout.write(JSON.stringify(question(q,{fromMs:2000,toMs:9000,minMS:low,maxMS:high})))`
		out, e := exec.Command("node", "-e", script, strconv.Itoa(low), strconv.Itoa(low+width)).Output()
		if e != nil {
			tc.Fatal(e)
		}
		var q exploration.Query
		if json.Unmarshal(out, &q) != nil || q.Validate() != nil || q.FromMS != 2000 || q.ToMS != 9000 || len(q.Where.All) != 3 || len(q.Where.All[0].Any) != 2 {
			tc.Fatal(string(out))
		}
		a, b := q.Where.All[1], q.Where.All[2]
		if a.Value != strconv.FormatInt(int64(low)*1000000, 10) || b.Value != strconv.FormatInt(int64(low+width)*1000000-1, 10) || a.Op != "gte" || b.Op != "lte" {
			tc.Fatal("half-open integer duration bounds widened")
		}
	}, options(80)...)
}

func TestQuintCoreRefinementsLedger(t *testing.T) {
	name := os.Getenv("GOO_CORE_REFINEMENTS_LEDGER")
	if name == "" {
		t.Skip("live ledger opt-in")
	}
	data, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	type ledger struct {
		Q, Selected                                            exploration.Query
		BeforeReturn, AfterReturn, BeforeService, AfterService int
	}
	check := func(mutation int) bool {
		var l ledger
		if json.Unmarshal(data, &l) != nil {
			t.Fatal("invalid ledger")
		}
		switch mutation {
		case 0:
			l.Selected.Environment = "foreign"
		case 1:
			l.Selected.FromMS = l.Q.FromMS - 1
		case 2:
			l.AfterReturn++
		case 3:
			l.AfterService++
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		coherent := l.Q.Validate() == nil && l.Selected.Validate() == nil && l.Selected.Environment == l.Q.Environment && l.Selected.FromMS >= l.Q.FromMS && l.Selected.ToMS <= l.Q.ToMS && len(l.Selected.Where.All) == 2 && l.Selected.Where.All[0].Value == "40000000" && l.Selected.Where.All[1].Value == "79999999"
		// Existing question-navigation model: selection stages a definition; retained history reads nothing.
		rec.Record(context.Background(), "restore", nil, "obs.reads", 0, "obs.coherent", coherent)
		rec.Record(context.Background(), "read", nil, "obs.reads", 1, "obs.coherent", true)
		rec.Record(context.Background(), "history", nil, "obs.reads", 1+l.AfterReturn-l.BeforeReturn, "obs.coherent", true)
		rec.Record(context.Background(), "restore", nil, "obs.reads", 1+l.AfterService-l.BeforeService, "obs.coherent", true)
		checker, e := validate.NewChecker("model/question_links.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("observed selection/navigation rejected")
	}
	for i := 0; i < 4; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}
