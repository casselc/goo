package shadow

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chucklehead-dev/gotogether/validate/model"
)

const pairedDesign = "design/paired-filters.binding.yaml"

func designContext(t *testing.T) context.Context {
	t.Helper()
	if _, err := exec.LookPath("quint"); err != nil {
		t.Fatal("design checks require quint; missing tooling is not a pass", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)
	return ctx
}

func TestPairedFilterDesignMatrix(t *testing.T) {
	// Check before Load/Matrix can regenerate the sibling model.
	tangleCommand(t, 0, "-check", "design/paired-filters.md")
	report, err := model.Matrix(designContext(t), pairedDesign, model.SimulateOptions{Graph: true, States: 10000})
	if err != nil {
		t.Fatal(err)
	}
	if report.Code != 0 {
		t.Fatalf("design matrix did not pass:\n%s", report.Text())
	}
	if len(report.Variants) != 5 || len(report.Rows) != 4 {
		t.Fatalf("missing design coverage: %d variants, %d properties", len(report.Variants), len(report.Rows))
	}
	for _, row := range report.Rows {
		for _, cell := range row.Cells {
			if !cell.OK || (cell.Expect == "holds" && cell.Verdict != "holds") || (cell.Expect == "allows" && (cell.Verdict != "breaks" || cell.Steps < 1 || cell.Steps > 4)) {
				t.Fatalf("unjudged, truncated or incorrect cell: %+v", cell)
			}
		}
	}
	raw, err := os.ReadFile("design/paired-filters.md")
	if err != nil {
		t.Fatal(err)
	}
	_, rest, ok := strings.Cut(string(raw), "<!-- gotogether:matrix -->")
	if !ok {
		t.Fatal("missing matrix start marker")
	}
	table, _, ok := strings.Cut(rest, "<!-- /gotogether:matrix -->")
	if !ok {
		t.Fatal("missing matrix end marker")
	}
	if strings.TrimSpace(table) != strings.TrimSpace(report.Text()) {
		t.Fatalf("design table stale; regenerate with the pinned simulate -matrix -write-matrix command\n%s", report.Text())
	}
	t.Log(report.Text())
}

func TestPairedFilterDesignRejectsCrossLaneEdit(t *testing.T) {
	// The shared-filter alternative must fail when claimed as the chosen design.
	report, err := model.Simulate(designContext(t), pairedDesign, model.SimulateOptions{Graph: true, States: 10000, Constants: []string{"SHARED_FILTER=true"}})
	if err != nil {
		t.Fatal(err)
	}
	if report.Code != 1 {
		t.Fatalf("coupled edit incorrectly accepted: %s", report.Text)
	}
	for _, verdict := range report.Invariants {
		if verdict.Name == "independentDrafts" && verdict.Violated && verdict.Counterexample != nil {
			return
		}
	}
	t.Fatalf("control failed without its intended counterexample: %s", report.Text)
}

func TestPairedFilterDesignRequiresReadableLane(t *testing.T) {
	// A design doing nothing must not pass merely because its safety flags hold.
	dir := t.TempDir()
	for _, name := range []string{"paired-filters.md", "paired-filters.binding.yaml"} {
		raw, err := os.ReadFile(filepath.Join("design", name))
		if err != nil {
			t.Fatal(err)
		}
		if strings.HasSuffix(name, ".md") {
			old := "action apply(lane: int): bool = all {\n    turns < 4,"
			if strings.Count(string(raw), old) != 1 {
				t.Fatal("cannot plant blocked-read model control")
			}
			raw = []byte(strings.Replace(string(raw), old, "action apply(lane: int): bool = all {\n    false,\n    turns < 4,", 1))
		}
		if err := os.WriteFile(filepath.Join(dir, name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	report, err := model.Simulate(designContext(t), filepath.Join(dir, "paired-filters.binding.yaml"), model.SimulateOptions{Graph: true, States: 10000})
	if err != nil {
		t.Fatal(err)
	}
	if report.Code != 1 {
		t.Fatalf("unreadable design incorrectly accepted: %s", report.Text)
	}
	for _, possible := range report.Possibles {
		if possible.Possible.Name == "edited lane can read and finish" && !possible.Found && possible.Impossible {
			return
		}
	}
	t.Fatalf("missing-read control lacked its intended diagnosis: %s", report.Text)
}
