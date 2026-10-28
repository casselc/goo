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

func TestHegelLinkedTimelineScope(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		from := hegel.Draw(tc, hegel.Integers(1000, 1000000))
		width := hegel.Draw(tc, hegel.Integers(2, 10000))
		script := `const a=require('node:assert/strict'),{narrowed}=require('../ui/service-timeline.mjs');const from=+process.argv[1],width=+process.argv[2],input={scope:{environment:'prod',namespace:{mode:'missing'},fromMs:from,toMs:from+width},service:'edge',nativeSelection:['exact'],instruments:[{unit:'ms',temporality:2}]},copy=structuredClone(input);const out=narrowed(input,{fromMs:from+1,toMs:from+width});a.deepEqual(input,copy);a.deepEqual({...out.scope,fromMs:from},input.scope);a.deepEqual(out.instruments,input.instruments);a.deepEqual(out.nativeSelection,input.nativeSelection);a.throws(()=>narrowed(input,{fromMs:from-1,toMs:from+width}));`
		if out, e := exec.Command("node", "-e", script, strconv.Itoa(from), strconv.Itoa(width)).CombinedOutput(); e != nil {
			tc.Fatalf("%s %v", out, e)
		}
	}, options(100)...)
}
func TestQuintLinkedTimelineLedger(t *testing.T) {
	name := os.Getenv("GOO_LINKED_TIMELINE_LEDGER")
	if name == "" {
		t.Skip("live opt-in")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Original, Narrowed, Reset                      exploration.DashboardPlan
			Selected                                       struct{ FromMS, ToMS int64 }
			Before, AfterPreview, AppliedReads, ResetReads int
			Domains                                        [][]string
		}
		if json.Unmarshal(raw, &l) != nil {
			t.Fatal("invalid ledger")
		}
		switch mutation {
		case 0:
			l.Narrowed.Scope.Environment = "foreign"
		case 1:
			l.Narrowed.Scope.FromMS--
		case 2:
			l.AfterPreview++
		case 3:
			l.ResetReads++
		case 4:
			l.Narrowed.Widgets[0].Query.Where = nil
		}
		coherent := l.Narrowed.Scope.Environment == l.Original.Scope.Environment && l.Narrowed.Scope.FromMS == l.Selected.FromMS && l.Narrowed.Scope.ToMS == l.Selected.ToMS && l.Selected.FromMS >= l.Original.Scope.FromMS && l.Selected.ToMS <= l.Original.Scope.ToMS && l.Reset.Scope.FromMS == l.Original.Scope.FromMS && l.Reset.Scope.ToMS == l.Original.Scope.ToMS
		for i, w := range l.Narrowed.Widgets {
			old := l.Original.Widgets[i]
			if w.Query != nil {
				a, _ := json.Marshal(w.Query.Where)
				b, _ := json.Marshal(old.Query.Where)
				coherent = coherent && string(a) == string(b) && w.Query.FromMS == l.Selected.FromMS && w.Query.ToMS == l.Selected.ToMS
			}
			if w.Measurement != nil {
				a, _ := json.Marshal(w.Measurement.Definition)
				b, _ := json.Marshal(old.Measurement.Definition)
				coherent = coherent && string(a) == string(b) && w.Measurement.FromMS == l.Selected.FromMS && w.Measurement.ToMS == l.Selected.ToMS
			}
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		rec.Record(context.Background(), "restore", nil, "obs.reads", 0, "obs.coherent", coherent)
		rec.Record(context.Background(), "history", nil, "obs.reads", l.AfterPreview-l.Before, "obs.coherent", true)
		for i := 0; i < l.AppliedReads+l.ResetReads; i++ {
			rec.Record(context.Background(), "read", nil, "obs.reads", i+1, "obs.coherent", true)
		}
		rec.Record(context.Background(), "history", nil, "obs.reads", 8, "obs.coherent", true)
		c, e := validate.NewChecker("model/question_links.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("observed window rejected")
	}
	for i := 0; i < 5; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}
