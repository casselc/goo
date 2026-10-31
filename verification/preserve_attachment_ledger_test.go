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

type preserveAttachmentStep struct {
	Action string `json:"action"`
	Reads  int    `json:"reads"`
	Writes int    `json:"documentPosts"`
	Cards  int    `json:"cards"`
}

var preserveAttachmentActions = []string{"preserve", "closedRefusal", "hiddenParentRefusal", "newSaveRefusal", "freshSaveAttached", "navigationRefusal", "pendingRefusal", "attach", "saveDocument"}

func checkPreserveAttachment(t *testing.T, c *validate.Checker, steps []preserveAttachmentStep) bool {
	t.Helper()
	if len(steps) != len(preserveAttachmentActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != preserveAttachmentActions[i] {
			return false
		}
		rec.Record(context.Background(), s.Action, nil, "obs.reads", s.Reads, "obs.writes", s.Writes, "obs.cards", s.Cards)
	}
	r, e := c.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return r.OK()
}
func validatePreserveAttachment(t *testing.T, steps []preserveAttachmentStep) {
	t.Helper()
	spec, e := filepath.Abs("model/preserve_attachment_293.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: preserve_attachment_293\nconstants: {MAX_READS: '1'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range preserveAttachmentActions {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}} and writes == {{.obs.writes}} and cards == {{.obs.cards}}'\n", a)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	c, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if e != nil {
		t.Fatal(e)
	}
	if !checkPreserveAttachment(t, c, steps) {
		t.Fatal("preserve attachment journey refused")
	}
	for i := 0; i < 6; i++ {
		bad := append([]preserveAttachmentStep(nil), steps...)
		switch i {
		case 0:
			bad[1].Reads++
		case 1:
			bad[7].Cards = 0
		case 2:
			bad[4].Writes = 0
		case 3:
			bad[5].Reads++
		case 4:
			bad[8].Writes--
		case 5:
			bad[1].Cards = 1
		}
		if checkPreserveAttachment(t, c, bad) {
			t.Fatal("mutant accepted", i)
		}
	}
	if checkPreserveAttachment(t, c, steps[:8]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintPreserveAttachmentModelSmoke(t *testing.T) {
	steps := []preserveAttachmentStep{}
	for i, a := range preserveAttachmentActions {
		writes, cards := 0, 0
		if i >= 4 {
			writes = 1
		}
		if i == 4 || i >= 7 {
			cards = 1
		}
		if i == 8 {
			writes = 2
		}
		steps = append(steps, preserveAttachmentStep{Action: a, Reads: 1, Writes: writes, Cards: cards})
	}
	validatePreserveAttachment(t, steps)
}

func TestQuintPreserveAttachmentObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_PRESERVE_ATTACHMENT_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []preserveAttachmentStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validatePreserveAttachment(t, steps)
}
