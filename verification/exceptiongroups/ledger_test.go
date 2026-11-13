package exceptiongroups

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type step struct {
	Name       string
	Reads, SQL int
}

func checkModel(t *testing.T, steps []step) bool {
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
		t.Fatalf("Quint required, not skipped: %v", err)
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink, qobs.WithActor(t.Name()))
	for _, s := range steps {
		rec.Record(context.Background(), s.Name, nil, "obs.reads", s.Reads, "obs.sql", s.SQL)
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
func TestExplicitExceptionReadStageRunProtocol(t *testing.T) {
	good := []step{{"scope", 0, 0}, {"local_intent", 0, 0}, {"read_groups", 1, 2}, {"stage_group", 1, 2}, {"run_group", 2, 5}, {"held_back", 2, 5}, {"failed_read", 3, 6}, {"read_groups", 4, 8}, {"pending_read", 5, 9}, {"retire", 5, 9}}
	if !checkModel(t, good) {
		t.Fatal("explicit exception journey refused")
	}
	localRead := append([]step(nil), good...)
	localRead[1].Reads++
	localRead[1].SQL++
	if checkModel(t, localRead) {
		t.Fatal("source acquisition from local gesture accepted")
	}
	if checkModel(t, []step{{"stage_group", 0, 0}, {"run_group", 1, 3}}) {
		t.Fatal("stage/run without admitted groups accepted")
	}
	fanout := append([]step(nil), good...)
	fanout[4].SQL++
	if checkModel(t, fanout) {
		t.Fatal("extra drilldown source fanout accepted")
	}
}

type counters struct{ Posts, SQL int }
type fact struct {
	Name          string
	Before, After counters
}
type browserReport struct {
	Version, Status string
	Errors          []string
	Facts           []fact
}

func TestObservedExceptionReadLedger(t *testing.T) {
	path := os.Getenv("GOO_EXCEPTION_GROUPS_REPORT")
	if path == "" {
		t.Skip("actual browser ledger supplied by GOO_EXCEPTION_GROUPS_REPORT")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var report browserReport
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if report.Version != "goo.exceptionGroups416/1" || report.Status != "passed" || len(report.Errors) > 0 || len(report.Facts) < 10 {
		t.Fatal("actual completed browser ledger required")
	}
	steps := []step{}
	state := step{}
	seen := map[string]bool{}
	signal := ""
	for _, f := range report.Facts {
		next := strings.SplitN(f.Name, "-", 2)[0]
		if next != "logs" && next != "spans" {
			t.Fatalf("fact lacks exact signal owner: %s", f.Name)
		}
		if next != signal {
			state.Name = "scope"
			steps = append(steps, state)
			signal = next
		}
		switch {
		case strings.HasSuffix(f.Name, "-read-groups") || strings.HasSuffix(f.Name, "-recover-groups"):
			state.Name = "read_groups"
			seen[next+"read"] = true
		case strings.HasSuffix(f.Name, "-failed-read"):
			state.Name = "failed_read"
		case strings.HasSuffix(f.Name, "-refused-groups"):
			state.Name = "refused_groups"
		case strings.HasSuffix(f.Name, "-stage-group"):
			state.Name = "stage_group"
			seen[next+"stage"] = true
		case strings.HasSuffix(f.Name, "-run-group"):
			state.Name = "run_group"
			seen[next+"run"] = true
		case strings.HasSuffix(f.Name, "-held-back"):
			state.Name = "held_back"
			seen[next+"back"] = true
		case strings.HasSuffix(f.Name, "-pending-read"):
			state.Name = "pending_read"
		case strings.HasSuffix(f.Name, "-retire"):
			state.Name = "retire"
			seen[next+"retire"] = true
		default:
			state.Name = "local_intent"
		}
		state.Reads += f.After.Posts - f.Before.Posts
		state.SQL += f.After.SQL - f.Before.SQL
		steps = append(steps, state)
	}
	for _, sig := range []string{"logs", "spans"} {
		for _, action := range []string{"read", "stage", "run", "back", "retire"} {
			if !seen[sig+action] {
				t.Fatalf("actual %s %s journey absent", sig, action)
			}
		}
	}
	if len(steps) == 0 || steps[len(steps)-1].Name != "retire" {
		t.Fatal("actual run ended without retirement")
	}
	if !checkModel(t, steps) {
		t.Fatal("actual exception read ledger violates model")
	}
	for i, s := range steps {
		if s.Name == "local_intent" {
			bad := append([]step(nil), steps...)
			bad[i].Reads++
			bad[i].SQL++
			if checkModel(t, bad) {
				t.Fatal("mutated actual local acquisition accepted")
			}
			break
		}
	}
}
