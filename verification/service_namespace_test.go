package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
	"os"
	"strings"
	"testing"
)

func TestQuintNamespaceLive(t *testing.T) {
	file := os.Getenv("GOO_NAMESPACE_LEDGER")
	if file == "" {
		t.Skip("live ledger opt-in")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Discovery exploration.ServiceDiscovery
			Link      workbench.Step
			Plan      exploration.DashboardPlan
			Result    exploration.Result
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		if mutation == 0 {
			l.Discovery.Dependencies[0].ParentNamespace = nil
		}
		if mutation == 1 {
			l.Link.Record.Where.All[1].Value = "orders"
		}
		if mutation == 2 {
			l.Result.Count = "2"
		}
		if mutation == 3 {
			l.Result.Question.ToMS++
		}
		phases := []bool{false, false, false}
		for _, d := range l.Discovery.Dependencies {
			if d.Kind == "edge" {
				phases[0] = d.Parent == "api" && d.Child == "api" && d.ParentNamespace != nil && *d.ParentNamespace == "payments" && d.ChildNamespace != nil && *d.ChildNamespace == "orders" && d.Count == "1"
			}
		}
		phases[1] = l.Link.Validate() == nil && l.Link.Record.Where.All[1].Value == "payments"
		for _, w := range l.Plan.Widgets {
			if w.Key == "contract-1" {
				a, _ := json.Marshal(w.Query)
				b, _ := json.Marshal(l.Result.Question)
				phases[2] = w.State == "supported" && l.Result.Count == "1" && string(a) == string(b)
			}
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for i, action := range []string{"discover", "open", "contract"} {
			rec.Record(context.Background(), action, nil, "obs.coherent", phases[i])
		}
		checker, e := validate.NewChecker("model/service_namespace.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
	for i := 0; i < 4; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}

func TestHegelNamespacePredicateIsolation(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		value := fmt.Sprintf("tenant-%d'\\雪", hegel.Draw(tc, hegel.Integers(0, 1000000)))
		selectors := []*identity.Namespace{nil, {Mode: "missing"}, {Mode: "value"}, {Mode: "value", Value: value}}
		selected := selectors[hegel.Draw(tc, hegel.Integers(0, 3))]
		actual := selectors[hegel.Draw(tc, hegel.Integers(1, 3))]
		resource := map[string]string{}
		fields := map[exploration.Field]string{}
		if actual.Mode == "value" {
			resource[identity.NamespaceKey] = actual.Value
			fields[exploration.Field{Kind: "resource", Key: identity.NamespaceKey}] = actual.Value
		}
		if selected.Matches(resource) != exploration.Match(exploration.NamespacePredicate(selected), fields) {
			tc.Fatal("namespace predicate disagrees with resource identity")
		}
		plan, e := exploration.CompileInventory(exploration.InventoryQuestion{Namespace: selected, Environment: "prod", FromMS: 0, ToMS: 60000})
		if e != nil || strings.Contains(plan.Count, value) {
			tc.Fatal("namespace SQL escaped parameter binding", e)
		}
	}, options(100)...)
}
