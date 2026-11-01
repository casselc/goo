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

type httpMethodFacetStep struct {
	Action    string `json:"action"`
	Reads     int    `json:"reads"`
	Pages     int    `json:"pages"`
	Saves     int    `json:"saves"`
	Expands   int    `json:"expands"`
	Lookups   int    `json:"lookups"`
	Encodes   int    `json:"encodes"`
	Preserved bool   `json:"preserved"`
}

var httpMethodFacetActions = []string{"owner", "customize", "runOriginal", "valuesOriginal", "presentation", "saveOriginal", "selectMethods", "stageMethods", "runSelected", "valuesSelected", "saveSelected", "encodeArtifact", "reopen", "rerun", "destination", "applyTemplate", "valuesDestination", "saveDestination"}

func checkHTTPMethodFacet(t *testing.T, c *validate.Checker, steps []httpMethodFacetStep) bool {
	t.Helper()
	if len(steps) != len(httpMethodFacetActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != httpMethodFacetActions[i] || !s.Preserved {
			return false
		}
		rec.Record(context.Background(), s.Action, nil, "obs.reads", s.Reads, "obs.pages", s.Pages, "obs.saves", s.Saves, "obs.expands", s.Expands, "obs.lookups", s.Lookups, "obs.encodes", s.Encodes)
	}
	result, e := c.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return result.OK()
}
func validateHTTPMethodFacet(t *testing.T, steps []httpMethodFacetStep) {
	t.Helper()
	spec, e := filepath.Abs("model/http_method_facet_311.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: http_method_facet_311\nconstants: {MAX_READS: '6'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range httpMethodFacetActions {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}} and pages == {{.obs.pages}} and saves == {{.obs.saves}} and expands == {{.obs.expands}} and lookups == {{.obs.lookups}} and encodes == {{.obs.encodes}}'\n", a)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	checker, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if e != nil {
		t.Fatal(e)
	}
	if !checkHTTPMethodFacet(t, checker, steps) {
		t.Fatal("HTTP method facet journey refused")
	}
	for i := 0; i < 7; i++ {
		bad := append([]httpMethodFacetStep(nil), steps...)
		switch i {
		case 0:
			bad[1].Reads++
		case 1:
			bad[3].Lookups--
		case 2:
			bad[7].Reads++
		case 3:
			bad[8].Reads--
		case 4:
			bad[12].Lookups++
		case 5:
			bad[15].Reads++
		case 6:
			bad[17].Preserved = false
		}

		if checkHTTPMethodFacet(t, checker, bad) {
			t.Fatal("mutant accepted", i)
		}
	}

	if checkHTTPMethodFacet(t, checker, steps[:len(steps)-1]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintHTTPMethodFacetModelSmoke(t *testing.T) {
	raw := `[{"action":"owner","reads":0,"lookups":0,"expands":0,"saves":0,"encodes":0,"pages":0,"preserved":true},{"action":"customize","reads":0,"lookups":0,"expands":1,"saves":0,"encodes":0,"pages":0,"preserved":true},{"action":"runOriginal","reads":1,"lookups":0,"expands":1,"saves":0,"encodes":0,"pages":0,"preserved":true},{"action":"valuesOriginal","reads":1,"lookups":1,"expands":1,"saves":0,"encodes":0,"pages":0,"preserved":true},{"action":"presentation","reads":1,"lookups":1,"expands":1,"saves":0,"encodes":0,"pages":0,"preserved":true},{"action":"saveOriginal","reads":1,"lookups":1,"expands":1,"saves":1,"encodes":0,"pages":0,"preserved":true},{"action":"selectMethods","reads":1,"lookups":1,"expands":1,"saves":1,"encodes":0,"pages":0,"preserved":true},{"action":"stageMethods","reads":1,"lookups":1,"expands":1,"saves":1,"encodes":0,"pages":0,"preserved":true},{"action":"runSelected","reads":2,"lookups":1,"expands":1,"saves":1,"encodes":0,"pages":0,"preserved":true},{"action":"valuesSelected","reads":2,"lookups":2,"expands":1,"saves":1,"encodes":0,"pages":0,"preserved":true},{"action":"saveSelected","reads":2,"lookups":2,"expands":1,"saves":2,"encodes":0,"pages":0,"preserved":true},{"action":"encodeArtifact","reads":2,"lookups":2,"expands":1,"saves":2,"encodes":1,"pages":0,"preserved":true},{"action":"reopen","reads":2,"lookups":2,"expands":1,"saves":2,"encodes":1,"pages":0,"preserved":true},{"action":"rerun","reads":3,"lookups":2,"expands":1,"saves":2,"encodes":1,"pages":0,"preserved":true},{"action":"destination","reads":6,"lookups":2,"expands":1,"saves":2,"encodes":1,"pages":0,"preserved":true},{"action":"applyTemplate","reads":6,"lookups":2,"expands":2,"saves":2,"encodes":1,"pages":0,"preserved":true},{"action":"valuesDestination","reads":6,"lookups":3,"expands":2,"saves":2,"encodes":1,"pages":0,"preserved":true},{"action":"saveDestination","reads":6,"lookups":3,"expands":2,"saves":3,"encodes":1,"pages":0,"preserved":true}]`
	var steps []httpMethodFacetStep
	if e := json.Unmarshal([]byte(raw), &steps); e != nil {
		t.Fatal(e)
	}
	validateHTTPMethodFacet(t, steps)
}
func TestQuintHTTPMethodFacetObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_HTTP_METHOD_FACET_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []httpMethodFacetStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateHTTPMethodFacet(t, steps)
}
