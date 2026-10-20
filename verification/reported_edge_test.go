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
	"reflect"
	"strings"
	"testing"
)

func TestHegelReportedEdgePlanning(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		ns := []*identity.Namespace{{Mode: "missing"}, {Mode: "value"}, {Mode: "value", Value: "payments"}}
		edge := &exploration.ReportedEdge{Parent: exploration.EdgeService{Service: fmt.Sprintf("parent-%d'", hegel.Draw(tc, hegel.Integers(0, 1000000))), Namespace: ns[hegel.Draw(tc, hegel.Integers(0, 2))]}, Child: exploration.EdgeService{Service: "worker", Namespace: ns[hegel.Draw(tc, hegel.Integers(0, 2))]}}
		from := int64(hegel.Draw(tc, hegel.Integers(0, 1000000)))
		q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: from, ToMS: from + 60000, Limit: 100, ReportedEdge: edge}
		p, e := exploration.Compile(q)
		if e != nil || strings.Contains(p.Count, edge.Parent.Service) || !strings.Contains(p.Count, "GROUP BY TraceId, SpanId HAVING groupUniqArray") {
			tc.Fatal(p, e)
		}
		normalized := (exploration.CaptureQuestion{Kind: "explore", Explore: &q}).Normalized()
		edge.Parent.Service = "mutated"
		if normalized.Explore.ReportedEdge.Parent.Service == "mutated" {
			tc.Fatal("custody selector aliases caller")
		}
	}, options(100)...)
}
func TestQuintReportedEdgeLive(t *testing.T) {
	name := os.Getenv("GOO_EDGE_LEDGER")
	if name == "" {
		t.Skip("live ledger opt-in")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Discovery                                    exploration.ServiceDiscovery
			Q, ScopedQuestion                            exploration.Query
			Result, Scoped, Ambiguous, Applied, Narrowed exploration.Result
			Staged                                       workbench.Step
			Captured                                     exploration.CapturePage
			Doc                                          workbench.Document
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		if mutation == 0 {
			l.Q.ReportedEdge.Parent.Namespace.Value = "foreign"
		}
		if mutation == 1 {
			l.Result.Count = "100"
		}
		if mutation == 2 {
			l.ScopedQuestion.ReportedEdge.DiscoveryNamespace = nil
		}
		if mutation == 3 {
			l.Captured.Receipt.Question.Explore.ReportedEdge = nil
		}
		if mutation == 4 {
			l.Narrowed.Count = "130"
		}
		found := false
		for _, d := range l.Discovery.Dependencies {
			if d.Kind == "edge" && d.ParentNamespace != nil && *d.ParentNamespace == "payments" && d.ChildNamespace != nil && *d.ChildNamespace == "orders" {
				found = d.Count == "130"
			}
		}
		stages := []bool{found && l.Scoped.Count == "1" && l.Ambiguous.Count == "0" && l.ScopedQuestion.ReportedEdge.DiscoveryNamespace != nil, l.Result.Count == "130" && len(l.Result.Rows) == 100 && l.Result.HasMore && l.Result.SpanStatus.Error == "65" && l.Result.Duration.P95NS == "40000000" && reflect.DeepEqual(l.Q.ReportedEdge, l.Result.Question.ReportedEdge) && reflect.DeepEqual(l.Staged.Record.ReportedEdge, l.Q.ReportedEdge) && l.Narrowed.Count == "65" && reflect.DeepEqual(l.Narrowed.Question.ReportedEdge, l.Q.ReportedEdge), l.Captured.Explore.Count == "130" && reflect.DeepEqual(l.Captured.Receipt.Question.Explore.ReportedEdge, l.Q.ReportedEdge) && reflect.DeepEqual(l.Doc.Cards[0].Step.Record.ReportedEdge, l.Q.ReportedEdge)}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for i, a := range []string{"discover", "investigate", "capture"} {
			rec.Record(context.Background(), a, nil, "obs.coherent", stages[i])
		}
		checker, e := validate.NewChecker("model/reported_edge.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("live trace rejected")
	}
	for i := 0; i < 5; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}
