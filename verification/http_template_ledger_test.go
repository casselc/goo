package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"os"
	"path/filepath"
	"testing"
)

type httpTemplateStep struct {
	Action    string `json:"action"`
	Reads     int    `json:"reads"`
	Staged    bool   `json:"staged"`
	Preserved bool   `json:"preserved"`
}

var httpTemplateActions = []string{"service", "customize", "read", "edit", "save", "reopen", "rerun"}

func checkHTTPTemplate(t *testing.T, c *validate.Checker, steps []httpTemplateStep) bool {
	t.Helper()
	if len(steps) != len(httpTemplateActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != httpTemplateActions[i] || !s.Preserved {
			return false
		}
		rec.Record(context.Background(), s.Action, nil, "obs.reads", s.Reads, "obs.staged", s.Staged)
	}
	r, e := c.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return r.OK()
}
func validateHTTPTemplate(t *testing.T, steps []httpTemplateStep) {
	t.Helper()
	spec, e := filepath.Abs("model/http_template_278.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: http_template_278\nconstants: {BASE_READS: '%d', MAX_READS: '%d'}\ninvariants: [safety]\nactions:\n", spec, steps[0].Reads, steps[0].Reads+2)
	for _, a := range httpTemplateActions {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}} and staged == {{.obs.staged}}'\n", a)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	c, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if e != nil {
		t.Fatal(e)
	}
	if !checkHTTPTemplate(t, c, steps) {
		t.Fatal("HTTP template journey refused")
	}
	for i := 0; i < 5; i++ {
		bad := append([]httpTemplateStep(nil), steps...)
		switch i {
		case 0:
			bad[1].Reads++
		case 1:
			bad[1].Staged = false
		case 2:
			bad[4].Preserved = false
		case 3:
			bad[5].Reads++
		case 4:
			bad[6].Reads--
		}
		if checkHTTPTemplate(t, c, bad) {
			t.Fatal("mutant accepted", i)
		}
	}
	if checkHTTPTemplate(t, c, steps[:6]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintHTTPTemplateModelSmoke(t *testing.T) {
	steps := []httpTemplateStep{}
	for i, a := range httpTemplateActions {
		reads := 5
		if i >= 2 {
			reads = 6
		}
		if i == 6 {
			reads = 7
		}
		steps = append(steps, httpTemplateStep{Action: a, Reads: reads, Staged: i >= 1, Preserved: true})
	}
	validateHTTPTemplate(t, steps)
}
func TestQuintHTTPTemplateObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_HTTP_TEMPLATE_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []httpTemplateStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateHTTPTemplate(t, steps)
}
