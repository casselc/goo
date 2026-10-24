package verification

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestActivityViewProperties(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		n := hegel.Draw(tc, hegel.Integers(0, 1000))
		v := workbench.ActivityView{Version: 1, Title: "Activity " + strconv.Itoa(n), Selector: exploration.TailSelector{Signal: "logs", Environment: "prod"}, Columns: []string{"utc", "summary"}, FieldColumns: []exploration.Field{{Kind: "resource", Key: "service.namespace"}}}
		s := workbench.Store{Dir: t.TempDir()}
		a, e := s.CreateActivityView(v)
		if e != nil {
			t.Fatal(e)
		}
		b, e := s.CreateActivityView(v)
		if e != nil || a.ID != b.ID {
			t.Fatal("unstable rolling configuration", e)
		}
		v.Selector.Environment = "other"
		changed, e := s.CreateActivityView(v)
		if e != nil || changed.ID == a.ID {
			t.Fatal("scope did not bind identity", e)
		}
		for _, environment := range []string{"prod", "other", "absent"} {
			list, err := s.ActivityViews("", environment)
			if err != nil {
				t.Fatal(err)
			}
			expected := 1
			if environment == "absent" {
				expected = 0
			}
			if len(list.Items) != expected {
				t.Fatal("environment catalog cardinality", environment, len(list.Items))
			}
			for _, item := range list.Items {
				if item.Selector.Environment != environment {
					t.Fatal("cross-environment configuration")
				}
			}
		}
		old, e := s.LoadActivityView(a.ID)
		if e != nil || old.Selector.Environment != "prod" {
			t.Fatal("old identity changed", e)
		}
	}, options(50)...)
}
func TestQuintActivityViewLedger(t *testing.T) {
	file := os.Getenv("GOO_ACTIVITY_VIEW_LEDGER")
	if file == "" {
		t.Skip("observed Activity view browser ledger required")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	check := func(bad int) bool {
		var l struct {
			Saved             workbench.ActivityView
			Reads, StagedRows int
			StartVisible      bool
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch bad {
		case 0:
			l.Reads++
		case 1:
			l.StagedRows++
		case 2:
			l.StartVisible = false
		case 3:
			l.Saved.Columns = []string{"context"}
		}
		coherent := l.Saved.Validate() == nil && l.Reads == 0 && l.StagedRows == 0 && l.StartVisible
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		rec.Record(context.Background(), "inspect", nil, "obs.reads", l.Reads, "obs.published", 0, "obs.coherent", coherent)
		checker, e := validate.NewChecker("model/annotations.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("observed staged view rejected")
	}
	for bad := 0; bad < 4; bad++ {
		if check(bad) {
			t.Fatal("altered staged view accepted", bad)
		}
	}
}

// Validate the observed environment switches as well as source-free staging.
func TestQuintActivityViewMenuLedger(t *testing.T) {
	file := os.Getenv("GOO_ACTIVITY_VIEW_MENU_LEDGER")
	if file == "" {
		t.Skip("observed Activity menu ledger required")
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	check := func(bad int) bool {
		var l struct {
			Saved                              workbench.ActivityView
			Reads, StagedRows                  int
			StartVisible                       bool
			Catalogs                           []string
			Environment, Other, ScopeAfterRace string
		}
		if err := json.Unmarshal(raw, &l); err != nil {
			t.Fatal(err)
		}
		switch bad {
		case 0:
			l.Reads++
		case 1:
			l.StagedRows++
		case 2:
			l.StartVisible = false
		case 3:
			l.Catalogs[0] = l.Other
		case 4:
			l.ScopeAfterRace = "Worker activity"
		case 5:
			l.Saved.Selector.Environment = l.Environment
		}
		coherent := l.Saved.Validate() == nil && l.Reads == 0 && l.StagedRows == 0 && l.StartVisible && len(l.Catalogs) >= 3 && l.Catalogs[0] == l.Environment && l.Catalogs[len(l.Catalogs)-1] == l.Other && l.Saved.Selector.Environment == l.Other && strings.Contains(l.ScopeAfterRace, "Other environment") && !strings.Contains(l.ScopeAfterRace, "Worker activity")
		sink := &qobs.MemorySink{}
		qobs.NewRecorder(sink).Record(context.Background(), "inspect", nil, "obs.reads", l.Reads, "obs.published", 0, "obs.coherent", coherent)
		checker, e := validate.NewChecker("model/annotations.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("observed Activity menu rejected")
	}
	for bad := 0; bad < 6; bad++ {
		if check(bad) {
			t.Fatal("altered menu accepted", bad)
		}
	}
}
