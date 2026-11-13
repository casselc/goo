package status419

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

type modelStep struct {
	Name              string
	Reads, SQL, Pairs int
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
		t.Fatalf("Quint required, not skipped: %v", err)
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink, qobs.WithActor(t.Name()))
	for _, s := range steps {
		rec.Record(context.Background(), s.Name, nil, "obs.reads", s.Reads, "obs.sql", s.SQL, "obs.pairs", s.Pairs)
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
func TestStatusVersionReadStageRunProtocol(t *testing.T) {
	good := []modelStep{{"local_intent", 0, 0, 0}, {"discover", 1, 2, 0}, {"local_intent", 1, 2, 0}, {"stage_pair", 1, 2, 0}, {"run_pair", 3, 8, 1}, {"return_service", 3, 8, 1}, {"failed_discovery", 4, 9, 1}, {"discover", 5, 11, 1}, {"retire", 5, 11, 1}}
	if !checkModel(t, good) {
		t.Fatal("valid explicit discovery/comparison flow refused")
	}
	supplied := []modelStep{{"discover", 1, 2, 0}, {"stage_pair", 1, 2, 0}, {"run_pair", 3, 8, 1}, {"supplied_compare", 3, 8, 2}, {"return_service", 3, 8, 2}}
	if !checkModel(t, supplied) {
		t.Fatal("pure supplied comparison refused")
	}
	badSupplied := append([]modelStep(nil), supplied...)
	badSupplied[3].Reads++
	if checkModel(t, badSupplied) {
		t.Fatal("supplied comparison acquired telemetry")
	}
	for _, index := range []int{0, 2, 3, 5} {
		bad := append([]modelStep(nil), good...)
		bad[index].Reads++
		bad[index].SQL++
		if checkModel(t, bad) {
			t.Fatal("implicit local source acquisition accepted")
		}
	}
	if checkModel(t, []modelStep{{"stage_pair", 0, 0, 0}, {"run_pair", 2, 6, 1}}) {
		t.Fatal("comparison before discovery accepted")
	}
	if checkModel(t, []modelStep{{"supplied_compare", 0, 0, 1}}) {
		t.Fatal("supplied compare before applied operands accepted")
	}
	bad := append([]modelStep(nil), good...)
	bad[4].SQL++
	if checkModel(t, bad) {
		t.Fatal("extra comparison source fanout accepted")
	}
}

type observedCounters struct{ Source, Compare, SQL int }
type observedFact struct {
	Name          string
	Before, After observedCounters
}
type versionReport struct {
	Version, Status string
	Facts           []observedFact
	Errors          []string
}

func TestObservedStatusVersionReadLedger(t *testing.T) {
	path := os.Getenv("GOO_STATUS419_REPORT")
	if path == "" {
		t.Skip("actual browser ledger supplied by GOO_STATUS419_REPORT")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var report versionReport
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if report.Version != "goo.serviceVersions419/1" || report.Status != "passed" || len(report.Errors) > 0 || len(report.Facts) < 10 {
		t.Fatal("incomplete or failed actual browser report")
	}
	steps := []modelStep{}
	state := modelStep{}
	seenDiscover, seenRun, seenStage, seenReturn := false, false, false, false
	seenStatus := map[string]bool{}
	for i, f := range report.Facts {
		if i > 0 && f.Before != report.Facts[i-1].After {
			t.Fatalf("unrecorded effect between facts %s", f.Name)
		}
		switch {
		case f.Name == "explicit-version-discovery" || f.Name == "recover-version-discovery":
			state.Name = "discover"
			seenDiscover = true
		case f.Name == "failed-discovery-withdraws-choice":
			state.Name = "failed_discovery"
		case strings.HasPrefix(f.Name, "supplied-"):
			state.Name = "supplied_compare"
		case strings.HasPrefix(f.Name, "stage-"):
			state.Name = "stage_pair"
			seenStage = true
		case strings.HasPrefix(f.Name, "explicit-") && strings.HasSuffix(f.Name, "-run"):
			state.Name = "run_pair"
			seenRun = true
		case strings.HasPrefix(f.Name, "held-service-return"):
			state.Name = "return_service"
			seenReturn = true
		case f.Name == "pending-version-discovery" || f.Name == "pending-discovery":
			state.Name = "pending_discovery"
		case f.Name == "departure-retires-discovery":
			state.Name = "retire"
		default:
			state.Name = "local_intent"
		}
		if strings.HasPrefix(f.Name, "explicit-error-") && strings.HasSuffix(f.Name, "-run") {
			seenStatus[f.Name] = true
		}
		state.Reads += f.After.Source - f.Before.Source
		state.SQL += f.After.SQL - f.Before.SQL
		state.Pairs += f.After.Compare - f.Before.Compare
		steps = append(steps, state)
	}
	if !seenDiscover || !seenRun || !seenStage || !seenReturn || !seenStatus["explicit-error-count-run"] || !seenStatus["explicit-error-fraction-run"] {
		t.Fatal("required actual journey path absent")
	}
	if !checkModel(t, steps) {
		t.Fatal("actual recorded read/stage/run budgets violate the model")
	}
	// Mutating one actual local action must be rejected by the same validator.
	for i, s := range steps {
		if s.Name == "local_intent" {
			bad := append([]modelStep(nil), steps...)
			bad[i].Reads++
			bad[i].SQL++
			if checkModel(t, bad) {
				t.Fatal("mutated actual local acquisition accepted")
			}
			break
		}
	}
}
