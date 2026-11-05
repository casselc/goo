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

type categoryPairStep338 struct {
	Action    string `json:"action"`
	Reads     int    `json:"reads"`
	Source    int    `json:"source"`
	Lookups   int    `json:"lookups"`
	Compares  int    `json:"compares"`
	Saves     int    `json:"saves"`
	Encodes   int    `json:"encodes"`
	Preserved bool   `json:"preserved"`
}

var categoryPairActions338 = []string{"owner", "runParent", "valuesParent", "stageBaseline", "applyBaseline", "runBaseline", "backBaseline", "stageExclude", "applyExclude", "runExcluded", "valuesExcluded", "stageInclude", "applyInclude", "runIncluded", "openGroup", "readGroup", "backGroup", "save", "encodeArtifact", "reopen", "runReopened"}

func checkCategoryPair338(t *testing.T, c *validate.Checker, steps []categoryPairStep338) bool {
	t.Helper()
	if len(steps) != len(categoryPairActions338) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != categoryPairActions338[i] || !s.Preserved || s.Source != s.Reads+s.Lookups {
			return false
		}
		rec.Record(context.Background(), s.Action, nil, "obs.reads", s.Reads, "obs.lookups", s.Lookups, "obs.compares", s.Compares, "obs.saves", s.Saves, "obs.encodes", s.Encodes)
	}
	result, err := c.Check(context.Background(), sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	return result.OK()
}
func validateCategoryPair338(t *testing.T, steps []categoryPairStep338) {
	t.Helper()
	spec, err := filepath.Abs("model/category_pair_338.qnt")
	if err != nil {
		t.Fatal(err)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: category_pair_338\nconstants: {MAX_READS: '11'}\ninvariants: [safety]\nactions:\n", spec)
	for _, name := range categoryPairActions338 {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}} and lookups == {{.obs.lookups}} and compares == {{.obs.compares}} and saves == {{.obs.saves}} and encodes == {{.obs.encodes}}'\n", name)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if err = os.WriteFile(file, []byte(binding), 0600); err != nil {
		t.Fatal(err)
	}
	c, err := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if err != nil {
		t.Fatal(err)
	}
	if !checkCategoryPair338(t, c, steps) {
		t.Fatal("completed category-pair journey refused")
	}
	for mutant := 0; mutant < 11; mutant++ {
		bad := append([]categoryPairStep338(nil), steps...)
		switch mutant {
		case 0:
			bad[3].Reads++
		case 1:
			bad[5].Reads--
		case 2:
			bad[4].Lookups++
		case 3:
			bad[10].Lookups--
		case 4:
			bad[15].Reads--
		case 5:
			bad[19].Reads++
		case 6:
			bad[20].Compares--
		case 7:
			bad[14].Preserved = false
		case 8:
			bad[2].Lookups--
		case 9:
			bad[6].Reads++
		}
		// Keep endpoint accounting coherent so effect mutations reach Quint.
		for i := range bad {
			bad[i].Source = bad[i].Reads + bad[i].Lookups
		}
		if mutant == 10 {
			bad[len(bad)-1].Source++
		}
		if checkCategoryPair338(t, c, bad) {
			t.Fatal("paired-effect mutant accepted", mutant)
		}
	}
	if checkCategoryPair338(t, c, steps[:len(steps)-1]) {
		t.Fatal("unfinished category-pair prefix accepted")
	}
}
func TestQuintCategoryPairModelSmoke338(t *testing.T) {
	counts := [][5]int{{0, 0, 0, 0, 0}, {2, 0, 1, 0, 0}, {2, 2, 1, 0, 0}, {2, 2, 1, 0, 0}, {2, 2, 1, 0, 0}, {4, 2, 2, 0, 0}, {4, 2, 2, 0, 0}, {4, 2, 2, 0, 0}, {4, 2, 2, 0, 0}, {6, 2, 3, 0, 0}, {6, 4, 3, 0, 0}, {6, 4, 3, 0, 0}, {6, 4, 3, 0, 0}, {8, 4, 4, 0, 0}, {8, 4, 4, 0, 0}, {9, 4, 4, 0, 0}, {9, 4, 4, 0, 0}, {9, 4, 4, 1, 0}, {9, 4, 4, 1, 1}, {9, 4, 4, 1, 1}, {11, 4, 5, 1, 1}}
	steps := make([]categoryPairStep338, len(counts))
	for i, c := range counts {
		steps[i] = categoryPairStep338{Action: categoryPairActions338[i], Reads: c[0], Lookups: c[1], Compares: c[2], Saves: c[3], Encodes: c[4], Source: c[0] + c[1], Preserved: true}
	}
	validateCategoryPair338(t, steps)
}
func TestQuintCategoryPairObservedBrowserLedger338(t *testing.T) {
	file := os.Getenv("GOO_CATEGORY_PAIR_LEDGER")
	if file == "" {
		t.Skip("real 338 browser ledger not supplied")
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var steps []categoryPairStep338
	if err = json.Unmarshal(raw, &steps); err != nil {
		t.Fatal(err)
	}
	validateCategoryPair338(t, steps)
}
