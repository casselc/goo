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

type pairedFacetStep337 struct {
	Action    string `json:"action"`
	Reads     int    `json:"reads"`
	Lookups   int    `json:"lookups"`
	Compares  int    `json:"compares"`
	Saves     int    `json:"saves"`
	Encodes   int    `json:"encodes"`
	Preserved bool   `json:"preserved"`
}

var pairedFacetActions337 = []string{"owner", "runParent", "valuesParent", "stageExclude", "applyExclude", "runExcluded", "valuesExcluded", "stageInclude", "applyInclude", "runIncluded", "openGroup", "readGroup", "backGroup", "save", "encodeArtifact", "reopen", "runReopened"}

func checkPairedFacet337(t *testing.T, c *validate.Checker, steps []pairedFacetStep337) bool {
	t.Helper()
	if len(steps) != len(pairedFacetActions337) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != pairedFacetActions337[i] || !s.Preserved {
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
func validatePairedFacet337(t *testing.T, steps []pairedFacetStep337) {
	t.Helper()
	spec, err := filepath.Abs("model/paired_facet_337.qnt")
	if err != nil {
		t.Fatal(err)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: paired_facet_337\nconstants: {MAX_READS: '9'}\ninvariants: [safety]\nactions:\n", spec)
	for _, name := range pairedFacetActions337 {
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
	if !checkPairedFacet337(t, c, steps) {
		t.Fatal("completed paired-facet journey refused")
	}
	for mutant := 0; mutant < 8; mutant++ {
		bad := append([]pairedFacetStep337(nil), steps...)
		switch mutant {
		case 0:
			bad[3].Reads++
		case 1:
			bad[5].Reads--
		case 2:
			bad[4].Lookups++
		case 3:
			bad[6].Lookups--
		case 4:
			bad[11].Reads--
		case 5:
			bad[15].Reads++
		case 6:
			bad[16].Compares--
		case 7:
			bad[10].Preserved = false
		}
		if checkPairedFacet337(t, c, bad) {
			t.Fatal("paired-effect mutant accepted", mutant)
		}
	}
	if checkPairedFacet337(t, c, steps[:len(steps)-1]) {
		t.Fatal("unfinished paired-facet prefix accepted")
	}
}
func TestQuintPairedFacetModelSmoke337(t *testing.T) {
	counts := [][5]int{{0, 0, 0, 0, 0}, {2, 0, 1, 0, 0}, {2, 1, 1, 0, 0}, {2, 1, 1, 0, 0}, {2, 1, 1, 0, 0}, {4, 1, 2, 0, 0}, {4, 2, 2, 0, 0}, {4, 2, 2, 0, 0}, {4, 2, 2, 0, 0}, {6, 2, 3, 0, 0}, {6, 2, 3, 0, 0}, {7, 2, 3, 0, 0}, {7, 2, 3, 0, 0}, {7, 2, 3, 1, 0}, {7, 2, 3, 1, 1}, {7, 2, 3, 1, 1}, {9, 2, 4, 1, 1}}
	steps := make([]pairedFacetStep337, len(counts))
	for i, c := range counts {
		steps[i] = pairedFacetStep337{Action: pairedFacetActions337[i], Reads: c[0], Lookups: c[1], Compares: c[2], Saves: c[3], Encodes: c[4], Preserved: true}
	}
	validatePairedFacet337(t, steps)
}
func TestQuintPairedFacetObservedBrowserLedger337(t *testing.T) {
	file := os.Getenv("GOO_PAIRED_FACET_LEDGER")
	if file == "" {
		t.Skip("real 337 browser ledger not supplied")
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var steps []pairedFacetStep337
	if err = json.Unmarshal(raw, &steps); err != nil {
		t.Fatal(err)
	}
	validatePairedFacet337(t, steps)
}
