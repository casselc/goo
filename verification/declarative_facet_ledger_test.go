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

type declarativeFacetStep struct {
	Action    string `json:"action"`
	Reads     int    `json:"reads"`
	Pages     int    `json:"pages"`
	Saves     int    `json:"saves"`
	Compares  int    `json:"compares"`
	Lookups   int    `json:"lookups"`
	Encodes   int    `json:"encodes"`
	Decodes   int    `json:"decodes"`
	Preserved bool   `json:"preserved"`
}

var declarativeFacetActions = []string{"raw", "invalidFacet", "configureFacet", "save", "encodeArtifact", "previewArtifact", "importDefinition", "openImported", "runImported", "valuesParent", "externalOrigin", "clearOwned", "missingEmpty", "stageSelected", "runSelected", "stale", "valuesSelected", "destination", "applyDisplay", "valuesDestination"}

func checkDeclarativeFacet(t *testing.T, c *validate.Checker, steps []declarativeFacetStep) bool {
	t.Helper()
	if len(steps) != len(declarativeFacetActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != declarativeFacetActions[i] || !s.Preserved {
			return false
		}
		rec.Record(context.Background(), s.Action, nil, "obs.reads", s.Reads, "obs.pages", s.Pages, "obs.saves", s.Saves, "obs.compares", s.Compares, "obs.lookups", s.Lookups, "obs.encodes", s.Encodes, "obs.decodes", s.Decodes)
	}
	result, e := c.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return result.OK()
}
func validateDeclarativeFacet(t *testing.T, steps []declarativeFacetStep) {
	t.Helper()
	spec, e := filepath.Abs("model/declarative_facet_310.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: declarative_facet_310\nconstants: {MAX_READS: '6'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range declarativeFacetActions {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}} and pages == {{.obs.pages}} and saves == {{.obs.saves}} and compares == {{.obs.compares}} and lookups == {{.obs.lookups}} and encodes == {{.obs.encodes}} and decodes == {{.obs.decodes}}'\n", a)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	checker, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if e != nil {
		t.Fatal(e)
	}
	if !checkDeclarativeFacet(t, checker, steps) {
		t.Fatal("authored cohort journey refused")
	}
	for i := 0; i < 7; i++ {
		bad := append([]declarativeFacetStep(nil), steps...)
		switch i {
		case 0:
			bad[2].Lookups++
		case 1:
			bad[9].Lookups--
		case 2:
			bad[7].Pages++
		case 3:
			bad[13].Reads++
		case 4:
			bad[14].Pages--
		case 5:
			bad[18].Lookups++
		case 6:
			bad[11].Preserved = false
		}

		if checkDeclarativeFacet(t, checker, bad) {
			t.Fatal("mutant accepted", i)
		}
	}

	if checkDeclarativeFacet(t, checker, steps[:len(steps)-1]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintDeclarativeFacetModelSmoke(t *testing.T) {
	raw := `[{"action":"raw","reads":3,"pages":1,"lookups":0,"saves":0,"encodes":0,"decodes":0,"compares":0,"preserved":true},{"action":"invalidFacet","reads":3,"pages":1,"lookups":0,"saves":0,"encodes":0,"decodes":0,"compares":0,"preserved":true},{"action":"configureFacet","reads":3,"pages":1,"lookups":0,"saves":0,"encodes":0,"decodes":0,"compares":0,"preserved":true},{"action":"save","reads":3,"pages":1,"lookups":0,"saves":1,"encodes":0,"decodes":0,"compares":0,"preserved":true},{"action":"encodeArtifact","reads":3,"pages":1,"lookups":0,"saves":1,"encodes":1,"decodes":0,"compares":0,"preserved":true},{"action":"previewArtifact","reads":3,"pages":1,"lookups":0,"saves":1,"encodes":1,"decodes":1,"compares":0,"preserved":true},{"action":"importDefinition","reads":3,"pages":1,"lookups":0,"saves":2,"encodes":1,"decodes":1,"compares":0,"preserved":true},{"action":"openImported","reads":3,"pages":1,"lookups":0,"saves":2,"encodes":1,"decodes":1,"compares":0,"preserved":true},{"action":"runImported","reads":3,"pages":2,"lookups":0,"saves":2,"encodes":1,"decodes":1,"compares":0,"preserved":true},{"action":"valuesParent","reads":3,"pages":2,"lookups":1,"saves":2,"encodes":1,"decodes":1,"compares":0,"preserved":true},{"action":"externalOrigin","reads":3,"pages":2,"lookups":1,"saves":2,"encodes":1,"decodes":1,"compares":0,"preserved":true},{"action":"clearOwned","reads":3,"pages":2,"lookups":1,"saves":2,"encodes":1,"decodes":1,"compares":0,"preserved":true},{"action":"missingEmpty","reads":3,"pages":2,"lookups":1,"saves":2,"encodes":1,"decodes":1,"compares":0,"preserved":true},{"action":"stageSelected","reads":3,"pages":2,"lookups":1,"saves":2,"encodes":1,"decodes":1,"compares":0,"preserved":true},{"action":"runSelected","reads":3,"pages":3,"lookups":1,"saves":2,"encodes":1,"decodes":1,"compares":0,"preserved":true},{"action":"stale","reads":3,"pages":3,"lookups":1,"saves":2,"encodes":1,"decodes":1,"compares":0,"preserved":true},{"action":"valuesSelected","reads":3,"pages":3,"lookups":2,"saves":2,"encodes":1,"decodes":1,"compares":0,"preserved":true},{"action":"destination","reads":6,"pages":4,"lookups":2,"saves":2,"encodes":1,"decodes":1,"compares":0,"preserved":true},{"action":"applyDisplay","reads":6,"pages":4,"lookups":2,"saves":2,"encodes":1,"decodes":1,"compares":0,"preserved":true},{"action":"valuesDestination","reads":6,"pages":4,"lookups":3,"saves":2,"encodes":1,"decodes":1,"compares":0,"preserved":true}]`
	var steps []declarativeFacetStep
	if e := json.Unmarshal([]byte(raw), &steps); e != nil {
		t.Fatal(e)
	}
	validateDeclarativeFacet(t, steps)
}
func TestQuintDeclarativeFacetObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_DECLARATIVE_FACET_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []declarativeFacetStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateDeclarativeFacet(t, steps)
}
