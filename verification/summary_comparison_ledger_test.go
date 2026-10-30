package verification

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type summaryComparisonEvent struct {
	Action      string `json:"action"`
	Reads       int    `json:"reads"`
	Comparisons int    `json:"comparisons"`
	Current     int    `json:"current"`
	Baseline    int    `json:"baseline"`
}

// The browser records settled source and comparison requests plus population
// ordinals. Pinning the baseline cannot read the source or follow a later Run.
// This bounded projection is not a claim about security or whole-product liveness.
func TestQuintSummaryComparisonLedger(t *testing.T) {
	path := os.Getenv("GOO_SUMMARY_COMPARISON_LEDGER")
	if path == "" {
		t.Skip("observed summary/comparison browser ledger required")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var events []summaryComparisonEvent
	if e = json.Unmarshal(raw, &events); e != nil {
		t.Fatal(e)
	}
	if len(events) < 6 || events[0] != (summaryComparisonEvent{Action: "run", Reads: 1, Current: 1}) {
		t.Fatal("missing initial observed population read")
	}
	first := map[string]int{}
	secondRun := -1
	for i, event := range events {
		if _, seen := first[event.Action]; !seen {
			first[event.Action] = i
		}
		if event.Action == "run" && event.Current > 1 {
			secondRun = i
		}
	}
	for _, action := range []string{"run", "pin", "stage", "inspect", "reject"} {
		if _, seen := first[action]; !seen {
			t.Fatalf("missing observed %s phase", action)
		}
	}
	last := events[len(events)-1]
	if secondRun < 0 || last.Current <= last.Baseline || last.Baseline < 1 || last.Comparisons < 2 {
		t.Fatal("missing rerun with held independent baseline")
	}
	spec, e := os.ReadFile("model/summary_comparison_264.qnt")
	if e != nil {
		t.Fatal(e)
	}
	dir := t.TempDir()
	if e = os.WriteFile(filepath.Join(dir, "summary_comparison_264.qnt"), spec, 0600); e != nil {
		t.Fatal(e)
	}
	binding := `spec: summary_comparison_264.qnt
module: summary_comparison_264
constants: {LIMIT: '8'}
invariants: [safety]
actions:
  applyRun:
    expect: 'reads == {{.obs.reads}} and comparisons == {{.obs.comparisons}} and current == {{.obs.current}} and baseline == {{.obs.baseline}}'
  pin:
    expect: 'reads == {{.obs.reads}} and comparisons == {{.obs.comparisons}} and current == {{.obs.current}} and baseline == {{.obs.baseline}}'
  stage:
    expect: 'reads == {{.obs.reads}} and comparisons == {{.obs.comparisons}} and current == {{.obs.current}} and baseline == {{.obs.baseline}}'
  inspect:
    expect: 'reads == {{.obs.reads}} and comparisons == {{.obs.comparisons}} and current == {{.obs.current}} and baseline == {{.obs.baseline}}'
  reject:
    expect: 'reads == {{.obs.reads}} and comparisons == {{.obs.comparisons}} and current == {{.obs.current}} and baseline == {{.obs.baseline}}'
`
	bindingPath := filepath.Join(dir, "binding.yaml")
	if e = os.WriteFile(bindingPath, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	check := func(trace []summaryComparisonEvent) bool {
		t.Helper()
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, event := range trace {
			action := event.Action
			if action == "run" {
				action = "applyRun"
			}
			rec.Record(context.Background(), action, nil, "obs.reads", event.Reads, "obs.comparisons", event.Comparisons, "obs.current", event.Current, "obs.baseline", event.Baseline)
		}
		checker, e := validate.NewChecker(bindingPath, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		result, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return result.OK()
	}
	if !check(events) {
		t.Fatal("observed summary/comparison ownership rejected")
	}
	faults := []struct {
		name   string
		index  int
		change func(*summaryComparisonEvent)
	}{
		{"pin illicit source read", first["pin"], func(e *summaryComparisonEvent) { e.Reads++ }},
		{"stage illicit source read", first["stage"], func(e *summaryComparisonEvent) { e.Reads++ }},
		{"inspect illicit source read", first["inspect"], func(e *summaryComparisonEvent) { e.Reads++ }},
		{"reject illicit source read", first["reject"], func(e *summaryComparisonEvent) { e.Reads++ }},
		{"rerun silently retargets baseline", secondRun, func(e *summaryComparisonEvent) { e.Baseline = e.Current }},
		{"pin omits comparison", first["pin"], func(e *summaryComparisonEvent) { e.Comparisons-- }},
		{"rerun omits comparison", secondRun, func(e *summaryComparisonEvent) { e.Comparisons-- }},
		{"stage triggers comparison", first["stage"], func(e *summaryComparisonEvent) { e.Comparisons++ }},
		{"local current changes", first["inspect"], func(e *summaryComparisonEvent) { e.Current++ }},
		{"local equality-preserving read and current change", first["inspect"], func(e *summaryComparisonEvent) { e.Reads++; e.Current++ }},
		{"pin breaks read/current equality", first["pin"], func(e *summaryComparisonEvent) { e.Current++ }},
	}
	for _, fault := range faults {
		changed := append([]summaryComparisonEvent(nil), events...)
		fault.change(&changed[fault.index])
		if check(changed) {
			t.Fatal("accepted forged ownership:", fault.name)
		}
	}
}
