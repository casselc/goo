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

type recordPageStep struct {
	Action   string `json:"action"`
	Full     int    `json:"full"`
	Pages    int    `json:"pages"`
	Keeps    int    `json:"keeps"`
	Captures int    `json:"captures"`
	Held     int    `json:"held"`
	Fresh    int    `json:"fresh"`
}

var recordPageActions = []string{"open", "readPopulation", "arrival", "page", "keep", "rejectPromotion", "retire", "ordinary", "capture"}

func checkRecordPage(t *testing.T, c *validate.Checker, steps []recordPageStep) bool {
	t.Helper()
	if len(steps) != len(recordPageActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != recordPageActions[i] {
			return false
		}
		rec.Record(context.Background(), s.Action, nil, "obs.full", s.Full, "obs.pages", s.Pages, "obs.keeps", s.Keeps, "obs.captures", s.Captures, "obs.held", s.Held, "obs.fresh", s.Fresh)
	}
	r, e := c.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return r.OK()
}
func validateRecordPage(t *testing.T, steps []recordPageStep) {
	t.Helper()
	spec, e := filepath.Abs("model/record_page_289.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: record_page_289\nconstants: {MAX_READS: '3'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range recordPageActions {
		binding += fmt.Sprintf("  %s:\n    expect: 'full == {{.obs.full}} and pages == {{.obs.pages}} and keeps == {{.obs.keeps}} and captures == {{.obs.captures}} and held == {{.obs.held}} and fresh == {{.obs.fresh}}'\n", a)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	c, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if e != nil {
		t.Fatal(e)
	}
	if !checkRecordPage(t, c, steps) {
		t.Fatal("record-page journey refused")
	}
	for i := 0; i < 6; i++ {
		bad := append([]recordPageStep(nil), steps...)
		switch i {
		case 0:
			bad[1].Full++
		case 1:
			bad[3].Pages--
		case 2:
			bad[3].Held++
		case 3:
			bad[4].Keeps--
		case 4:
			bad[5].Captures++
		case 5:
			bad[7].Full--
		}
		if checkRecordPage(t, c, bad) {
			t.Fatal("mutant accepted", i)
		}
	}
	if checkRecordPage(t, c, steps[:8]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintRecordPageModelSmoke(t *testing.T) {
	steps := []recordPageStep{}
	values := [][6]int{{0, 0, 0, 0, 0, 0}, {1, 0, 0, 0, 3, 3}, {1, 0, 0, 0, 3, 3}, {1, 1, 0, 0, 3, 4}, {1, 1, 1, 0, 3, 4}, {1, 1, 1, 0, 3, 4}, {1, 2, 1, 0, 0, 0}, {3, 2, 1, 0, 0, 4}, {3, 2, 1, 1, 0, 4}}
	for i, a := range recordPageActions {
		v := values[i]
		steps = append(steps, recordPageStep{Action: a, Full: v[0], Pages: v[1], Keeps: v[2], Captures: v[3], Held: v[4], Fresh: v[5]})
	}
	validateRecordPage(t, steps)
}

func TestQuintRecordPageObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_RECORD_PAGE_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []recordPageStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateRecordPage(t, steps)
}
