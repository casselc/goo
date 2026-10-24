package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
	"os"
	"reflect"
	"testing"
)

func TestHegelSavedActivityBaselineIdentity(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		n := hegel.Draw(tc, hegel.Integers(1, 1000000))
		s := workbench.Store{Dir: t.TempDir()}
		v := workbench.ActivityView{Version: 1, Title: fmt.Sprintf("Span view %d", n), Selector: exploration.TailSelector{Signal: "spans", Environment: "prod", Where: &exploration.Predicate{Any: []exploration.Predicate{{Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "missing"}, {Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "eq", Value: ""}}}}, Columns: []string{"utc", "summary"}}
		a, e := s.CreateActivityView(v)
		if e != nil {
			tc.Fatal(e)
		}
		v.BaselineCapture = fmt.Sprintf("%064x", n)
		b, e := s.CreateActivityView(v)
		if e != nil || a.ID == b.ID {
			tc.Fatal("reference did not bind identity", e)
		}
		same, e := s.CreateActivityView(v)
		if e != nil || same.ID != b.ID {
			tc.Fatal("unstable reference", e)
		}
		v.BaselineCapture = fmt.Sprintf("%064x", n+1)
		c, e := s.CreateActivityView(v)
		if e != nil || c.ID == b.ID {
			tc.Fatal("replacement did not change identity", e)
		}
		v.BaselineCapture = ""
		cleared, e := s.CreateActivityView(v)
		if e != nil || cleared.ID != a.ID {
			tc.Fatal("clear changed legacy projection", e)
		}
		old, e := s.LoadActivityView(b.ID)
		if e != nil || old.BaselineCapture != b.BaselineCapture {
			tc.Fatal("old reference mutated", e)
		}
	}, options(50)...)
}

func TestQuintSavedActivityBaselineLedger(t *testing.T) {
	file := os.Getenv("GOO_SAVED_ACTIVITY_BASELINE_LEDGER")
	if file == "" {
		t.Skip("observed browser ledger")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Saved, Cleared, Old       workbench.ActivityView
			Captured                  string
			BeforeReopen, AfterReopen struct{ Reads, Captures, Comparisons int }
			RowsAfterReopen           int
			StartExplicit             bool
			BeforeRetry, AfterRetry   int
		}
		if e = json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch mutation {
		case 0:
			l.AfterReopen.Reads++
		case 1:
			l.AfterReopen.Captures++
		case 2:
			l.AfterReopen.Comparisons++
		case 3:
			l.Saved.BaselineCapture = fmt.Sprintf("%064x", 1)
		case 4:
			l.Cleared.BaselineCapture = l.Captured
		case 5:
			l.StartExplicit = false
		}
		v := l.Saved
		v.ID = ""
		stored, err := (workbench.Store{Dir: t.TempDir()}).CreateActivityView(v)
		coherent := err == nil && stored.ID == l.Saved.ID && l.Saved.Validate() == nil && l.Cleared.Validate() == nil && l.Saved.BaselineCapture == l.Captured && l.Cleared.BaselineCapture == "" && l.Saved.ID != l.Cleared.ID && reflect.DeepEqual(l.Old, l.Saved) && l.BeforeReopen == l.AfterReopen && l.RowsAfterReopen == 0 && l.StartExplicit && l.BeforeRetry == l.AfterRetry
		sink := &qobs.MemorySink{}
		qobs.NewRecorder(sink).Record(context.Background(), "inspect", nil, "obs.reads", l.AfterReopen.Reads-l.BeforeReopen.Reads, "obs.published", 0, "obs.coherent", coherent)
		checker, e := validate.NewChecker("model/annotations.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("valid saved reference ledger rejected")
	}
	for i := 0; i < 6; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}
