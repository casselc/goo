package verification

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"reflect"
	"testing"

	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
)

type investigationDraftMachine struct{ worker journalMachine }

func (m *investigationDraftMachine) RuleCheckpoint(tc hegel.TestCase) { m.worker.call("checkpoint") }
func (m *investigationDraftMachine) RulePublishUnknown(tc hegel.TestCase) {
	m.worker.call("beforePublish")
}
func (m *investigationDraftMachine) RuleReload(tc hegel.TestCase) { m.worker.call("reload") }
func (m *investigationDraftMachine) RuleLateAcknowledgment(tc hegel.TestCase) {
	m.worker.call("oldAck")
}
func (m *investigationDraftMachine) RuleAcknowledgeCurrent(tc hegel.TestCase) { m.worker.call("ack") }
func (m *investigationDraftMachine) RuleDiscard(tc hegel.TestCase)            { m.worker.call("discard") }
func (m *investigationDraftMachine) RuleStaleDiscard(tc hegel.TestCase) {
	m.worker.call("staleDiscard")
}
func (m *investigationDraftMachine) RuleChangedSlot(tc hegel.TestCase) {
	m.worker.call("foreignWriter")
}
func (m *investigationDraftMachine) RuleRejectPayload(tc hegel.TestCase) { m.worker.call("malformed") }
func (m *investigationDraftMachine) RuleRejectParent(tc hegel.TestCase)  { m.worker.call("wrongParent") }
func (m *investigationDraftMachine) RuleExpiry(tc hegel.TestCase)        { m.worker.call("expired") }
func (m *investigationDraftMachine) RuleOtherPath(tc hegel.TestCase)     { m.worker.call("wrongPath") }
func (m *investigationDraftMachine) RuleDeniedStorage(tc hegel.TestCase) {
	m.worker.call("deniedStorage")
}
func (m *investigationDraftMachine) RuleQuotaFailure(tc hegel.TestCase) { m.worker.call("quota") }
func (m *investigationDraftMachine) InvariantExactAuthoredIntent(tc hegel.TestCase) {
	m.worker.call("inspect")
}

func TestHegelInvestigationDraftStateSwarm(t *testing.T) {
	cmd := exec.Command("node", "bridge/investigation-draft.cjs")
	in, e := cmd.StdinPipe()
	if e != nil {
		t.Fatal(e)
	}
	out, e := cmd.StdoutPipe()
	if e != nil {
		t.Fatal(e)
	}
	cmd.Stderr = os.Stderr
	if e = cmd.Start(); e != nil {
		t.Fatal(e)
	}
	defer func() {
		in.Close()
		if e := cmd.Wait(); e != nil {
			t.Error(e)
		}
	}()
	worker := journalMachine{in: in, out: bufio.NewScanner(out)}
	// Deterministically cover the high-risk transitions even when swarm subsets
	// omit a rule; generated sequences then explore their composition and order.
	for _, action := range []string{"reset", "checkpoint", "beforePublish", "reload", "oldAck", "staleDiscard", "foreignWriter", "malformed", "wrongParent", "expired", "wrongPath", "deniedStorage", "quota", "checkpoint", "ack", "discard"} {
		worker.call(action)
	}
	hegel.Test(t, func(tc *hegel.T) {
		m := &investigationDraftMachine{worker: worker}
		m.worker.call("reset")
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(24), hegel.WithAlwaysCheckInvariants("InvariantExactAuthoredIntent"))
	}, options(60)...)
}

type investigationDraftObservation struct {
	Action  string   `json:"action"`
	Reads   int      `json:"reads"`
	Writes  int      `json:"writes"`
	Parent  string   `json:"parent"`
	Refs    []string `json:"refs"`
	Pending bool     `json:"pending"`
	Unknown bool     `json:"unknown"`
	Dirty   bool     `json:"dirty"`
}
type investigationDraftLedger struct {
	Version        int                             `json:"version"`
	ExpectedParent string                          `json:"expectedParent"`
	ExpectedRefs   []string                        `json:"expectedRefs"`
	Steps          []investigationDraftObservation `json:"steps"`
}

func checkInvestigationDraftLedger(t *testing.T, l investigationDraftLedger) bool {
	t.Helper()
	if l.Version != 1 || len(l.ExpectedParent) != 64 || len(l.ExpectedRefs) == 0 || len(l.Steps) == 0 || l.Steps[0].Action != "checkpoint" {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	previous := l.Steps[0]
	for i, s := range l.Steps {
		coherent := s.Reads >= 0 && s.Writes >= 0 && s.Reads >= previous.Reads && s.Writes >= previous.Writes
		switch s.Action {
		case "checkpoint", "reload", "restore", "restoreUnknown", "editUnknown", "lateAck", "discard", "invalid":
			if i > 0 {
				coherent = coherent && s.Reads == previous.Reads && s.Writes == previous.Writes
			}
		case "uncertain", "saveSuccess":
			coherent = coherent && s.Writes == previous.Writes+1
		default:
			return false
		}
		if s.Pending {
			coherent = coherent && s.Parent == l.ExpectedParent && reflect.DeepEqual(s.Refs, l.ExpectedRefs)
		}
		if s.Action == "restore" || s.Action == "restoreUnknown" || s.Action == "editUnknown" || s.Action == "lateAck" {
			coherent = coherent && s.Dirty
		}
		if s.Action == "reload" {
			coherent = coherent && !s.Dirty // pending storage must not auto-stage the editor
		}
		rec.Record(context.Background(), s.Action, nil, "obs.pending", s.Pending, "obs.unknown", s.Unknown, "obs.coherent", coherent)
		previous = s
	}
	checker, e := validate.NewChecker("model/investigation_draft.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	result, e := checker.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return result.OK()
}

func TestQuintInvestigationDraftRejectsMutations(t *testing.T) {
	parent := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	ref := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	mk := func(a string, r, w int, p, u, d bool) investigationDraftObservation {
		return investigationDraftObservation{Action: a, Reads: r, Writes: w, Parent: parent, Refs: []string{ref}, Pending: p, Unknown: u, Dirty: d}
	}
	original := investigationDraftLedger{Version: 1, ExpectedParent: parent, ExpectedRefs: []string{ref}, Steps: []investigationDraftObservation{
		mk("checkpoint", 2, 0, true, false, true), mk("reload", 2, 0, true, false, false), mk("restore", 2, 0, true, false, true), mk("uncertain", 2, 1, true, true, true), mk("reload", 2, 1, true, true, false), mk("restoreUnknown", 2, 1, true, true, true), mk("editUnknown", 2, 1, true, true, true), mk("lateAck", 2, 1, true, true, true), mk("saveSuccess", 3, 2, false, false, false), mk("discard", 3, 2, false, false, false),
	}}
	if !checkInvestigationDraftLedger(t, original) {
		t.Fatal("labelled protocol fixture rejected")
	}
	for mutation := 0; mutation < 8; mutation++ {
		raw, _ := json.Marshal(original)
		var changed investigationDraftLedger
		_ = json.Unmarshal(raw, &changed)
		switch mutation {
		case 0:
			changed.Steps[2].Reads++ // automatic reference read on Restore
		case 1:
			changed.Steps[1].Writes++ // automatic publication during reload
		case 2:
			changed.Steps[2].Parent = ref
		case 3:
			changed.Steps[5].Refs = []string{parent}
		case 4:
			changed.Steps[7].Pending = false // stale save ack cleared newer draft
		case 5:
			changed.Steps[6].Unknown = false // local edit falsely cleared uncertainty
		case 6:
			changed.Steps[2].Dirty = false // falsely restored with no staged editor
		case 7:
			changed.Steps[1].Dirty = true // automatic restoration during reload
		}
		if checkInvestigationDraftLedger(t, changed) {
			t.Fatal("forged draft protocol accepted", mutation)
		}
	}
}

func TestQuintInvestigationDraftLedger(t *testing.T) {
	name := os.Getenv("GOO_INVESTIGATION_DRAFT_LEDGER")
	if name == "" {
		t.Skip("real draft recovery browser ledger opt-in")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	var l investigationDraftLedger
	if e = json.Unmarshal(raw, &l); e != nil {
		t.Fatal(e)
	}
	if !checkInvestigationDraftLedger(t, l) {
		t.Fatal("actual browser draft recovery ledger rejected")
	}
	// Mutate actual browser facts rather than substituting a model-only trace.
	for mutation, target := range []string{"restore", "reload", "reload", "restore", "restoreUnknown", "editUnknown", "editUnknown"} {
		found := false
		for i, s := range l.Steps {
			if s.Action != target {
				continue
			}
			copyRaw, _ := json.Marshal(l)
			var forged investigationDraftLedger
			_ = json.Unmarshal(copyRaw, &forged)
			found = true
			switch mutation {
			case 0:
				forged.Steps[i].Reads++
			case 1:
				forged.Steps[i].Writes++
			case 2:
				forged.Steps[i].Dirty = true
			case 3:
				forged.Steps[i].Parent = "wrong-parent"
			case 4:
				forged.Steps[i].Refs = []string{"wrong-reference"}
			case 5:
				forged.Steps[i].Pending = false
			case 6:
				forged.Steps[i].Unknown = false
			}
			if checkInvestigationDraftLedger(t, forged) {
				t.Fatal("forged observed recovery accepted", mutation, target)
			}
			break
		}
		if !found {
			t.Fatal("observed recovery ledger omits required action", target)
		}
	}
}
