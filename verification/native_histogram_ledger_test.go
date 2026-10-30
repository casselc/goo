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

type nativeHistogramStep struct {
	Action      string `json:"action"`
	Reads       int    `json:"reads"`
	Total       string `json:"total"`
	Bins        int    `json:"bins"`
	Lower       string `json:"lower"`
	Upper       string `json:"upper"`
	Unit        string `json:"unit"`
	Temporality string `json:"temporality"`
}

func checkNativeHistogramLedger(t *testing.T, checker *validate.Checker, steps []nativeHistogramStep) bool {
	t.Helper()
	names := []string{"source", "table", "zero", "unknown", "cumulative", "saved", "complete"}
	if len(steps) != len(names) {
		return false
	}
	sink := &qobs.MemorySink{}
	recorder := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != names[i] {
			return false
		}
		switch s.Action {
		case "source", "table":
			if s.Total != "4" || s.Bins != 3 || s.Lower != "-inf" || s.Upper != "+inf" || s.Unit != "s" || s.Temporality != "delta" {
				return false
			}
		case "zero":
			if s.Total != "0" || s.Bins != 3 || s.Lower != "-inf" || s.Upper != "+inf" {
				return false
			}
		case "unknown":
			if s.Bins != 0 || s.Total != "" {
				return false
			}
		default:
			if s.Total != "5" || s.Bins != 3 || s.Lower != "-inf" || s.Upper != "+inf" || s.Unit != "ms" || s.Temporality != "cumulative" {
				return false
			}
		}
		recorder.Record(context.Background(), s.Action, nil, "obs.reads", s.Reads)
	}
	r, err := checker.Check(context.Background(), sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	return r.OK()
}
func validateNativeHistogramLedger(t *testing.T, steps []nativeHistogramStep) {
	t.Helper()
	spec, err := filepath.Abs("model/native_histogram_272.qnt")
	if err != nil {
		t.Fatal(err)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: native_histogram_272\nconstants: {MAX_READS: '2'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range []string{"source", "table", "zero", "unknown", "cumulative", "saved", "complete"} {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}}'\n", a)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if err = os.WriteFile(file, []byte(binding), 0600); err != nil {
		t.Fatal(err)
	}
	checker, err := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if err != nil {
		t.Fatal(err)
	}
	if !checkNativeHistogramLedger(t, checker, steps) {
		t.Fatal("actual native interval journey rejected")
	}
	for mutant := 0; mutant < 7; mutant++ {
		bad := append([]nativeHistogramStep(nil), steps...)
		switch mutant {
		case 0:
			bad[0].Lower = "0"
		case 1:
			bad[1].Reads++
		case 2:
			bad[2].Total = "4"
		case 3:
			bad[3].Bins = 3
		case 4:
			bad[4].Total = "7"
		case 5:
			bad[5].Unit = "s"
		case 6:
			bad[6].Reads++
		}
		if checkNativeHistogramLedger(t, checker, bad) {
			t.Fatalf("incorrect native histogram trace accepted %d", mutant)
		}
	}
	if checkNativeHistogramLedger(t, checker, steps[:len(steps)-1]) {
		t.Fatal("unfinished prefix passed")
	}
}
func TestQuintNativeHistogramObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_NATIVE_HISTOGRAM_LEDGER")
	if file == "" {
		t.Skip("actual native histogram journey ledger required")
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var steps []nativeHistogramStep
	if err = json.Unmarshal(raw, &steps); err != nil {
		t.Fatal(err)
	}
	validateNativeHistogramLedger(t, steps)
}
func TestQuintNativeHistogramModelSmoke(t *testing.T) {
	steps := []nativeHistogramStep{{"source", 1, "4", 3, "-inf", "+inf", "s", "delta"}, {"table", 1, "4", 3, "-inf", "+inf", "s", "delta"}, {"zero", 1, "0", 3, "-inf", "+inf", "s", "delta"}, {"unknown", 1, "", 0, "", "", "", ""}, {"cumulative", 2, "5", 3, "-inf", "+inf", "ms", "cumulative"}, {"saved", 2, "5", 3, "-inf", "+inf", "ms", "cumulative"}, {"complete", 2, "5", 3, "-inf", "+inf", "ms", "cumulative"}}
	validateNativeHistogramLedger(t, steps)
}
