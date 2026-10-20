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

func TestQuintDeclaredServicesLive(t *testing.T) {
	file := os.Getenv("GOO_DECLARED_LEDGER")
	if file == "" {
		t.Skip("live ledger opt-in")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Scope     exploration.InventoryQuestion
			Service   string
			Discovery exploration.ServiceDiscovery
			Registry  struct {
				State    string
				Packages []json.RawMessage
			}
			Explored       exploration.Result
			Captured       exploration.CapturePage
			ScopedQuestion exploration.Query
			ScopedCapture  exploration.CapturePage
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		if mutation == 0 {
			l.Discovery.Signals = append(l.Discovery.Signals, exploration.ServiceSignal{Service: l.Service, Signal: "spans", Count: "1"})
		}
		if mutation == 1 {
			l.Explored.Question.ToMS++
		}
		if mutation == 2 {
			l.Captured.Explore.Count = "1"
		}
		if mutation == 4 {
			l.ScopedCapture.Receipt.Question.Explore.Where.All[1].Value = "foreign"
		}
		if mutation == 3 {
			l.Captured.Receipt.Question.Explore.Where.Value = "foreign"
		}
		phases := []bool{l.Registry.State == "loaded" && len(l.Registry.Packages) > 0 && len(l.Discovery.Signals) == 0 && len(l.Discovery.Dependencies) == 0, l.Explored.Count == "0" && l.Explored.State == "no_data" && l.Explored.Question.Environment == l.Scope.Environment && l.Explored.Question.FromMS == l.Scope.FromMS && l.Explored.Question.ToMS == l.Scope.ToMS && l.Explored.Question.Where.Value == l.Service, l.Captured.Explore.Count == "0" && l.Captured.Explore.State == "no_data" && reflect.DeepEqual(l.Captured.Receipt.Question.Explore, &l.Explored.Question)}
		phases[2] = phases[2] && l.ScopedCapture.Explore.Count == "0" && reflect.DeepEqual(l.ScopedCapture.Receipt.Question.Explore, &l.ScopedQuestion) && l.ScopedQuestion.Where.All[1].Value == "rollout-canary"
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for i, a := range []string{"discover", "open", "capture"} {
			rec.Record(context.Background(), a, nil, "obs.coherent", phases[i])
		}
		checker, e := validate.NewChecker("model/declared_services.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("live trace rejected")
	}
	for i := 0; i < 5; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}
func TestHegelDeclaredCatalogIsolation(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		count := hegel.Draw(tc, hegel.Integers(1, 1000000))
		namespace := hegel.Draw(tc, hegel.Integers(0, 2))
		script := `const {catalogModel}=require('../ui/inventory');const ns=[undefined,'','payments'][Number(process.argv[2])],r={service:'api',signal:'spans',count:process.argv[1],firstMs:1,lastMs:1};if(ns!==undefined)r.namespace=ns;const e={question:{environment:'prod',fromMs:0,toMs:2},signals:[r]},p=(rev,namespace)=>({revision:rev,definition:{name:rev,title:rev,resource:{service:'api',environment:'prod',namespace},instrumentation:{name:'sdk',version:'1'}}}),ps=[p('a',{mode:'value',value:'orders'}),p('b',{mode:'value',value:'orders'})],before=JSON.stringify(e),a=catalogModel(e,ps),b=catalogModel(e,[...ps].reverse());if(JSON.stringify(a)!==JSON.stringify(b)||JSON.stringify(e)!==before||a.length!==2||a.find(x=>x.observed).signals.spans.count!==process.argv[1]||a.find(x=>!x.observed).declarations.length!==2||Object.keys(a.find(x=>!x.observed).signals).length)process.exit(1);`
		if out, e := exec.Command("node", "-e", script, strconv.Itoa(count), strconv.Itoa(namespace)).CombinedOutput(); e != nil {
			tc.Fatal(e, string(out))
		}
	}, options(100)...)
}
