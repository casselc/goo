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

type viewCLIStep struct {
	Action      string `json:"action"`
	Validations int    `json:"validations"`
	Refusals    int    `json:"refusals"`
	Imports     int    `json:"imports"`
	Reads       int    `json:"reads"`
	Saves       int    `json:"saves"`
	Expands     int    `json:"expands"`
	Preserved   bool   `json:"preserved"`
}

var viewCLIActions = []string{"validateChecked", "refuseInvalid", "validateScoped", "cliImport", "catalog", "open", "read", "destination", "apply", "save"}

func checkViewCLI(t *testing.T, c *validate.Checker, steps []viewCLIStep) bool {
	t.Helper()
	if len(steps) != len(viewCLIActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != viewCLIActions[i] || !s.Preserved {
			return false
		}
		rec.Record(context.Background(), s.Action, nil, "obs.validations", s.Validations, "obs.refusals", s.Refusals, "obs.imports", s.Imports, "obs.reads", s.Reads, "obs.saves", s.Saves, "obs.expands", s.Expands)
	}
	result, e := c.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return result.OK()
}
func validateViewCLI(t *testing.T, steps []viewCLIStep) {
	t.Helper()
	spec, e := filepath.Abs("model/view_cli_306.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: view_cli_306\nconstants: {MAX_READS: '5'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range viewCLIActions {
		binding += fmt.Sprintf("  %s:\n    expect: 'validations == {{.obs.validations}} and refusals == {{.obs.refusals}} and imports == {{.obs.imports}} and reads == {{.obs.reads}} and saves == {{.obs.saves}} and expands == {{.obs.expands}}'\n", a)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	checker, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if e != nil {
		t.Fatal(e)
	}
	if !checkViewCLI(t, checker, steps) {
		t.Fatal("view CLI journey refused")
	}
	for i := 0; i < 7; i++ {
		bad := append([]viewCLIStep(nil), steps...)
		switch i {
		case 0:
			bad[0].Reads++
		case 1:
			bad[1].Refusals--
		case 2:
			bad[3].Imports--
		case 3:
			bad[5].Reads++
		case 4:
			bad[8].Reads++
		case 5:
			bad[9].Saves--
		case 6:
			bad[7].Preserved = false
		}
		if checkViewCLI(t, checker, bad) {
			t.Fatal("mutant accepted", i)
		}
	}

	if checkViewCLI(t, checker, steps[:len(steps)-1]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintViewCLIModelSmoke(t *testing.T) {
	raw := `[{"action":"validateChecked","validations":1,"refusals":0,"imports":0,"reads":0,"saves":0,"expands":0,"preserved":true},{"action":"refuseInvalid","validations":1,"refusals":1,"imports":0,"reads":0,"saves":0,"expands":0,"preserved":true},{"action":"validateScoped","validations":2,"refusals":1,"imports":0,"reads":0,"saves":0,"expands":0,"preserved":true},{"action":"cliImport","validations":2,"refusals":1,"imports":1,"reads":0,"saves":0,"expands":0,"preserved":true},{"action":"catalog","validations":2,"refusals":1,"imports":1,"reads":0,"saves":0,"expands":0,"preserved":true},{"action":"open","validations":2,"refusals":1,"imports":1,"reads":0,"saves":0,"expands":0,"preserved":true},{"action":"read","validations":2,"refusals":1,"imports":1,"reads":1,"saves":0,"expands":0,"preserved":true},{"action":"destination","validations":2,"refusals":1,"imports":1,"reads":5,"saves":0,"expands":0,"preserved":true},{"action":"apply","validations":2,"refusals":1,"imports":1,"reads":5,"saves":0,"expands":1,"preserved":true},{"action":"save","validations":2,"refusals":1,"imports":1,"reads":5,"saves":1,"expands":1,"preserved":true}]`
	var steps []viewCLIStep
	if e := json.Unmarshal([]byte(raw), &steps); e != nil {
		t.Fatal(e)
	}
	validateViewCLI(t, steps)
}
func TestQuintViewCLIObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_VIEW_CLI_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []viewCLIStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateViewCLI(t, steps)
}
