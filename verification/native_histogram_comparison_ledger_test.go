package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type nativeHistogramPairStep struct {
	Action    string `json:"action"`
	Posts     int    `json:"posts"`
	Baseline  string `json:"baseline"`
	Selection string `json:"selection"`
	Scale     string `json:"scale"`
	Ceiling   string `json:"ceiling"`
	Pairs     int    `json:"pairs"`
	Held      bool   `json:"held"`
}

func checkNativeHistogramPairLedger(t *testing.T, checker *validate.Checker, steps []nativeHistogramPairStep) bool {
	t.Helper()
	names := []string{"compare", "count", "share", "saved", "withheld", "complete"}
	if len(steps) != len(names) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != names[i] || s.Posts != 0 {
			return false
		}
		if s.Action == "withheld" || s.Action == "complete" {
			if s.Pairs != 0 || s.Baseline != "" || s.Selection != "" {
				return false
			}
		} else {
			if s.Pairs != 1 || s.Baseline != "4" || s.Selection != "8" || !s.Held {
				return false
			}
			if s.Action == "share" {
				if s.Scale != "observation-share" || s.Ceiling != "100" {
					return false
				}
			} else if s.Scale != "count" || s.Ceiling != "5" {
				return false
			}
		}
		rec.Record(context.Background(), s.Action, nil, "obs.posts", s.Posts)
	}
	r, e := checker.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return r.OK()
}
func validateNativeHistogramPairLedger(t *testing.T, steps []nativeHistogramPairStep) {
	t.Helper()
	spec, e := filepath.Abs("model/native_histogram_comparison_274.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: native_histogram_comparison_274\nconstants: {MAX_POSTS: '0'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range []string{"compare", "count", "share", "saved", "withheld", "complete"} {
		binding += fmt.Sprintf("  %s:\n    expect: 'posts == {{.obs.posts}}'\n", a)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	checker, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if e != nil {
		t.Fatal(e)
	}
	if !checkNativeHistogramPairLedger(t, checker, steps) {
		t.Fatal("native comparison journey refused")
	}
	for mutant := 0; mutant < 8; mutant++ {
		bad := append([]nativeHistogramPairStep(nil), steps...)
		switch mutant {
		case 0:
			bad[0].Baseline = "2"
		case 1:
			bad[1].Selection = "3"
		case 2:
			bad[2].Ceiling = "5"
		case 3:
			bad[2].Scale = "valid-share"
		case 4:
			bad[3].Held = false
		case 5:
			bad[4].Pairs = 1
		case 6:
			bad[5].Posts = 1
		case 7:
			bad[1].Ceiling = "8"
		}
		if checkNativeHistogramPairLedger(t, checker, bad) {
			t.Fatal("mutant passed", mutant)
		}
	}
	if checkNativeHistogramPairLedger(t, checker, steps[:len(steps)-1]) {
		t.Fatal("unfinished journey passed")
	}
}
func nativeHistogramPairSmokeSteps() []nativeHistogramPairStep {
	return []nativeHistogramPairStep{{Action: "compare", Baseline: "4", Selection: "8", Scale: "count", Ceiling: "5", Pairs: 1, Held: true}, {Action: "count", Baseline: "4", Selection: "8", Scale: "count", Ceiling: "5", Pairs: 1, Held: true}, {Action: "share", Baseline: "4", Selection: "8", Scale: "observation-share", Ceiling: "100", Pairs: 1, Held: true}, {Action: "saved", Baseline: "4", Selection: "8", Scale: "count", Ceiling: "5", Pairs: 1, Held: true}, {Action: "withheld"}, {Action: "complete"}}
}
func TestQuintNativeHistogramPairModelSmoke(t *testing.T) {
	validateNativeHistogramPairLedger(t, nativeHistogramPairSmokeSteps())
}
func TestQuintNativeHistogramPairObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_NATIVE_HISTOGRAM_PAIR_LEDGER")
	if file == "" {
		t.Skip("real observed browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []nativeHistogramPairStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateNativeHistogramPairLedger(t, steps)
}
