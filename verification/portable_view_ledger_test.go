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

type portableViewStep struct {
	Action    string `json:"action"`
	Reads     int    `json:"reads"`
	Pages     int    `json:"pages"`
	Saves     int    `json:"saves"`
	Encodes   int    `json:"encodes"`
	Decodes   int    `json:"decodes"`
	Preserved bool   `json:"preserved"`
}

var portableViewActions = []string{"raw", "rawRead", "add", "fullRead", "saveSource", "export", "isolatedStore", "preview", "import", "open", "runImported", "destinationRead", "apply", "saveDestination"}

// The browser verbs "import" and "export" are Quint language keywords.
func portableModelAction(action string) string {
	switch action {
	case "export":
		return "encodeArtifact"
	case "import":
		return "saveArtifact"
	}
	return action
}

func checkPortableView(t *testing.T, c *validate.Checker, steps []portableViewStep) bool {
	t.Helper()
	if len(steps) != len(portableViewActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != portableViewActions[i] || !s.Preserved {
			return false
		}
		rec.Record(context.Background(), portableModelAction(s.Action), nil, "obs.reads", s.Reads, "obs.pages", s.Pages, "obs.saves", s.Saves, "obs.encodes", s.Encodes, "obs.decodes", s.Decodes)
	}
	result, e := c.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return result.OK()
}
func validatePortableView(t *testing.T, steps []portableViewStep) {
	t.Helper()
	spec, e := filepath.Abs("model/portable_view_305.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: portable_view_305\nconstants: {MAX_READS: '9'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range portableViewActions {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}} and pages == {{.obs.pages}} and saves == {{.obs.saves}} and encodes == {{.obs.encodes}} and decodes == {{.obs.decodes}}'\n", portableModelAction(a))
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	checker, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if e != nil {
		t.Fatal(e)
	}
	if !checkPortableView(t, checker, steps) {
		t.Fatal("portable view journey refused")
	}
	for i := 0; i < 7; i++ {
		bad := append([]portableViewStep(nil), steps...)
		switch i {
		case 0:
			bad[2].Reads++
		case 1:
			bad[5].Encodes--
		case 2:
			bad[7].Reads++
		case 3:
			bad[8].Saves--
		case 4:
			bad[9].Reads++
		case 5:
			bad[12].Pages++
		case 6:
			bad[11].Preserved = false
		}
		if checkPortableView(t, checker, bad) {
			t.Fatal("mutant accepted", i)
		}
	}

	if checkPortableView(t, checker, steps[:len(steps)-1]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintPortableViewModelSmoke(t *testing.T) {
	raw := `[{"action":"raw","reads":3,"pages":0,"saves":0,"encodes":0,"decodes":0,"preserved":true},{"action":"rawRead","reads":3,"pages":1,"saves":0,"encodes":0,"decodes":0,"preserved":true},{"action":"add","reads":3,"pages":1,"saves":0,"encodes":0,"decodes":0,"preserved":true},{"action":"fullRead","reads":4,"pages":1,"saves":0,"encodes":0,"decodes":0,"preserved":true},{"action":"saveSource","reads":4,"pages":1,"saves":1,"encodes":0,"decodes":0,"preserved":true},{"action":"export","reads":4,"pages":1,"saves":1,"encodes":1,"decodes":0,"preserved":true},{"action":"isolatedStore","reads":4,"pages":1,"saves":1,"encodes":1,"decodes":0,"preserved":true},{"action":"preview","reads":4,"pages":1,"saves":1,"encodes":1,"decodes":1,"preserved":true},{"action":"import","reads":4,"pages":1,"saves":2,"encodes":1,"decodes":1,"preserved":true},{"action":"open","reads":4,"pages":1,"saves":2,"encodes":1,"decodes":1,"preserved":true},{"action":"runImported","reads":5,"pages":1,"saves":2,"encodes":1,"decodes":1,"preserved":true},{"action":"destinationRead","reads":9,"pages":1,"saves":2,"encodes":1,"decodes":1,"preserved":true},{"action":"apply","reads":9,"pages":1,"saves":2,"encodes":1,"decodes":1,"preserved":true},{"action":"saveDestination","reads":9,"pages":1,"saves":3,"encodes":1,"decodes":1,"preserved":true}]`
	var steps []portableViewStep
	if e := json.Unmarshal([]byte(raw), &steps); e != nil {
		t.Fatal(e)
	}
	validatePortableView(t, steps)
}
func TestQuintPortableViewObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_PORTABLE_VIEW_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []portableViewStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validatePortableView(t, steps)
}
