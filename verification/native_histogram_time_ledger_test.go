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

type nativeHistogramTimeStep struct {
	Action      string `json:"action"`
	Reads       int    `json:"reads"`
	Total       string `json:"total"`
	Bins        int    `json:"bins"`
	Lower       string `json:"lower"`
	Upper       string `json:"upper"`
	Unit        string `json:"unit"`
	Temporality string `json:"temporality"`
	WidthNS     string `json:"widthNS"`
	SpanNS      string `json:"spanNS"`
	Index       int    `json:"index"`
}

func checkNativeHistogramTimeLedger(t *testing.T, checker *validate.Checker, steps []nativeHistogramTimeStep) bool {
	t.Helper()
	names := []string{"source", "table", "zero", "unknown", "final", "saved", "complete"}
	if len(steps) != len(names) {
		return false
	}
	sink := &qobs.MemorySink{}
	recorder := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != names[i] {
			return false
		}
		if s.Action == "unknown" {
			if s.WidthNS != "" || s.SpanNS != "" || s.Index != -1 {
				return false
			}
		} else {
			if s.WidthNS != "4000000000" {
				return false
			}
			switch s.Action {
			case "source", "table":
				if s.SpanNS != "4000000000" || s.Index != 0 {
					return false
				}
			case "zero":
				if s.SpanNS != "4000000000" || s.Index != 2 {
					return false
				}
			default:
				if s.SpanNS != "3000000000" || s.Index != 4 {
					return false
				}
			}
		}
		switch s.Action {
		case "source", "table":
			if s.Total != "7" || s.Bins != 3 || s.Lower != "-inf" || s.Upper != "+inf" || s.Unit != "s" || s.Temporality != "delta" {
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
			if s.Total != "4" || s.Bins != 3 || s.Lower != "-inf" || s.Upper != "+inf" || s.Unit != "s" || s.Temporality != "delta" {
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
func validateNativeHistogramTimeLedger(t *testing.T, steps []nativeHistogramTimeStep) {
	t.Helper()
	spec, err := filepath.Abs("model/native_histogram_time_273.qnt")
	if err != nil {
		t.Fatal(err)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: native_histogram_time_273\nconstants: {MAX_READS: '1'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range []string{"source", "table", "zero", "unknown", "final", "saved", "complete"} {
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
	if !checkNativeHistogramTimeLedger(t, checker, steps) {
		t.Fatal("actual native interval journey rejected")
	}
	for mutant := 0; mutant < 10; mutant++ {
		bad := append([]nativeHistogramTimeStep(nil), steps...)
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
			bad[4].Total = "8"
		case 5:
			bad[5].Unit = "ms"
		case 6:
			bad[6].Reads++
		case 7:
			bad[0].WidthNS = "3000000000"
		case 8:
			bad[4].SpanNS = "4000000000"
		case 9:
			bad[2].Index = 1
		}
		if checkNativeHistogramTimeLedger(t, checker, bad) {
			t.Fatalf("incorrect native histogram trace accepted %d", mutant)
		}
	}
	if checkNativeHistogramTimeLedger(t, checker, steps[:len(steps)-1]) {
		t.Fatal("unfinished prefix passed")
	}
}
func TestQuintNativeHistogramTimeObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_NATIVE_HISTOGRAM_TIME_LEDGER")
	if file == "" {
		t.Skip("actual native histogram journey ledger required")
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var steps []nativeHistogramTimeStep
	if err = json.Unmarshal(raw, &steps); err != nil {
		t.Fatal(err)
	}
	validateNativeHistogramTimeLedger(t, steps)
}
func TestQuintNativeHistogramTimeModelSmoke(t *testing.T) {
	steps := []nativeHistogramTimeStep{{"source", 1, "7", 3, "-inf", "+inf", "s", "delta", "4000000000", "4000000000", 0}, {"table", 1, "7", 3, "-inf", "+inf", "s", "delta", "4000000000", "4000000000", 0}, {"zero", 1, "0", 3, "-inf", "+inf", "s", "delta", "4000000000", "4000000000", 2}, {"unknown", 1, "", 0, "", "", "", "", "", "", -1}, {"final", 1, "4", 3, "-inf", "+inf", "s", "delta", "4000000000", "3000000000", 4}, {"saved", 1, "4", 3, "-inf", "+inf", "s", "delta", "4000000000", "3000000000", 4}, {"complete", 1, "4", 3, "-inf", "+inf", "s", "delta", "4000000000", "3000000000", 4}}
	validateNativeHistogramTimeLedger(t, steps)
}
