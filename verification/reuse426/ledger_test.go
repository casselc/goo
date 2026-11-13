package reuse426

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type counters struct{ Source, SQL, Counted, Compare, ViewPosts, ViewGets, CapturePosts int }
type modelStep struct {
	Name string
	counters
}

func checkModel(t *testing.T, steps []modelStep) bool {
	t.Helper()
	quint := os.Getenv("QUINT_BIN")
	if quint == "" {
		quint = "../../node_modules/.bin/quint"
	}
	quint, err := filepath.Abs(quint)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(quint); err != nil {
		t.Fatalf("native Quint mandatory, not skipped: %v", err)
	}
	sink := &qobs.MemorySink{}
	recorder := qobs.NewRecorder(sink, qobs.WithActor(t.Name()))
	for _, s := range steps {
		recorder.Record(context.Background(), s.Name, nil, "obs.reads", s.Source, "obs.sql", s.SQL, "obs.counted", s.Counted, "obs.compares", s.Compare, "obs.viewPosts", s.ViewPosts, "obs.viewGets", s.ViewGets, "obs.captures", s.CapturePosts)
	}
	checker, err := validate.NewChecker("binding.yaml", validate.Options{Quint: quint, Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	report, err := checker.Check(t.Context(), sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	return report.OK()
}
func protocol() []modelStep {
	c := counters{}
	steps := []modelStep{{"inspect", c}, {"facet", c}}
	c.Source++
	c.SQL += 2
	c.Counted++
	steps = append(steps, modelStep{"load_values", c}, modelStep{"choose_values", c}, modelStep{"stage_comparison", c})
	c.Source += 2
	c.SQL += 6
	c.Compare++
	steps = append(steps, modelStep{"execute_comparison", c}, modelStep{"local", c}, modelStep{"local", c})
	c.ViewPosts++
	steps = append(steps, modelStep{"preserve_late", c})
	c.ViewPosts++
	steps = append(steps, modelStep{"preserve_view", c}, modelStep{"local", c})
	c.ViewGets++
	steps = append(steps, modelStep{"open_definition", c}, modelStep{"local", c})
	c.Source += 2
	c.SQL += 6
	c.Compare++
	steps = append(steps, modelStep{"execute_comparison", c}, modelStep{"local", c}, modelStep{"return_held", c}, modelStep{"retire_context", c})
	return steps
}
func TestPreserveReopenAndReturnProtocol(t *testing.T) {
	good := protocol()
	if !checkModel(t, good) {
		t.Fatal("valid complete preservation protocol refused")
	}
	for _, mutation := range []struct {
		at    int
		field string
	}{{7, "write"}, {10, "get"}, {11, "source"}, {15, "source"}, {16, "write"}, {8, "capture"}} {
		bad := append([]modelStep(nil), good...)
		switch mutation.field {
		case "write":
			bad[mutation.at].ViewPosts++
		case "get":
			bad[mutation.at].ViewGets++
		case "source":
			bad[mutation.at].Source++
		case "capture":
			bad[mutation.at].CapturePosts++
		}
		if checkModel(t, bad) {
			t.Fatal("withdrawn/reopen/return gesture gained an effect", mutation)
		}
	}
	if checkModel(t, []modelStep{{"preserve_view", counters{ViewPosts: 1}}}) {
		t.Fatal("context without a configured comparison could preserve")
	}
	bad := append([]modelStep(nil), good[:9]...)
	c := bad[len(bad)-1].counters
	c.ViewGets++
	bad = append(bad, modelStep{"open_definition", c})
	if checkModel(t, bad) {
		t.Fatal("late stored response became an adopted receipt")
	}
}

type fact struct {
	Name          string
	Before, After counters
}
type journeyReport struct {
	Version, Status string
	Facts           []fact
	Errors          []string
}

var factActions = []struct{ name, action string }{
	{"inspect", "inspect"}, {"facet", "facet"}, {"load_values", "load_values"}, {"select_values", "choose_values"}, {"stage_comparison", "stage_comparison"}, {"run_comparison", "execute_comparison"}, {"open_preserve", "local"}, {"guard_preserve", "local"}, {"late_save", "preserve_late"}, {"save_view", "preserve_view"}, {"guard_reopen", "local"}, {"open_saved", "open_definition"}, {"staged_frames", "local"}, {"run_saved", "execute_comparison"}, {"applied_frames", "local"}, {"return_activity", "return_held"}, {"retire_owner", "retire_context"},
}

func observedSteps(t *testing.T, report journeyReport) []modelStep {
	t.Helper()
	if report.Version != "goo.investigationConsolidation426/1" || report.Status != "passed" || len(report.Errors) > 0 || len(report.Facts) != len(factActions) {
		t.Fatal("incomplete real investigation consolidation report")
	}
	var steps []modelStep
	for i, f := range report.Facts {
		if f.Name != factActions[i].name {
			t.Fatal("missing or reordered required journey fact", i, f.Name)
		}
		if i > 0 && f.Before != report.Facts[i-1].After {
			t.Fatal("unrecorded source/store effect", f.Name)
		}
		delta := counters{f.After.Source - f.Before.Source, f.After.SQL - f.Before.SQL, f.After.Counted - f.Before.Counted, f.After.Compare - f.Before.Compare, f.After.ViewPosts - f.Before.ViewPosts, f.After.ViewGets - f.Before.ViewGets, f.After.CapturePosts - f.Before.CapturePosts}
		for _, n := range []int{delta.Source, delta.SQL, delta.Counted, delta.Compare, delta.ViewPosts, delta.ViewGets, delta.CapturePosts} {
			if n < 0 {
				t.Fatal("regressed effect counters", f.Name)
			}
		}
		c := delta
		if i > 0 {
			p := steps[i-1].counters
			c = counters{p.Source + delta.Source, p.SQL + delta.SQL, p.Counted + delta.Counted, p.Compare + delta.Compare, p.ViewPosts + delta.ViewPosts, p.ViewGets + delta.ViewGets, p.CapturePosts + delta.CapturePosts}
		}
		steps = append(steps, modelStep{factActions[i].action, c})
	}
	return steps
}
func TestObservedInvestigationReuseLedger(t *testing.T) {
	path := os.Getenv("GOO_REUSE426_REPORT")
	if path == "" {
		t.Skip("actual completed browser ledger supplied by GOO_REUSE426_REPORT")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var report journeyReport
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	steps := observedSteps(t, report)
	if !checkModel(t, steps) {
		t.Fatal("actual preservation/reopen/return source protocol failed")
	}
	bad := append([]modelStep(nil), steps...)
	bad[11].Source++
	if checkModel(t, bad) {
		t.Fatal("reopening definition silently gained a source read")
	}
}
