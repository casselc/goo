package verification

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/casselc/goo/lifecycle"
	"hegel.dev/go/hegel"
)

func options(n int) []hegel.Option {
	if scale, err := strconv.Atoi(os.Getenv("PBT_SCALE")); err == nil && scale > 0 {
		n *= scale
	}
	return []hegel.Option{hegel.WithTestCases(n), hegel.WithDatabase(filepath.Join(os.TempDir(), "goo-hegel"))}
}

func TestHegelExactDeadlineAndDuplicateDeliveries(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		p := policy(100)
		delta := hegel.Draw(tc, hegel.Integers(-10, 20))
		copies := hegel.Draw(tc, hegel.Integers(1, 20))
		epoch := hegel.Draw(tc, hegel.SampledFrom([]int64{0, 1700000000000000000, 9223372036854775707}))
		p.EvaluationTimeNS, p.EvidenceCutoffNS = strconv.FormatInt(epoch+100, 10), strconv.FormatInt(epoch+100, 10)
		a, term := event("accepted"), event("terminal")
		a.EventTimeNS, a.ReceivedTimeNS = strconv.FormatInt(epoch+60, 10), strconv.FormatInt(epoch+61, 10)
		term.EventTimeNS = strconv.FormatInt(epoch+70+int64(delta), 10)
		term.ReceivedTimeNS = strconv.FormatInt(epoch+95, 10)
		input := []lifecycle.Event{a}
		for i := 0; i < copies; i++ {
			input = append(input, term)
		}
		prepared, err := p.Prepare(input)
		if err != nil {
			tc.Fatalf("prepare: %v", err)
		}
		wantGood := 1
		if delta > 0 {
			wantGood = 0
		}
		if prepared.Receipt.Counts["eligible"] != 1 || prepared.Receipt.Counts["good"] != wantGood {
			tc.Fatalf("deadline/dedup mismatch: %+v", prepared.Receipt.Counts)
		}
	}, options(300)...)
}

func TestHegelInvisibleAppendPreservesHistoricalEvidence(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		p := policy(80)
		original, err := p.Prepare([]lifecycle.Event{event("accepted")})
		if err != nil {
			tc.Fatal(err)
		}
		e := event("terminal")
		e.ReceivedTimeNS = strconv.Itoa(hegel.Draw(tc, hegel.Integers(80, 200)))
		updated, err := p.Prepare([]lifecycle.Event{event("accepted"), e})
		if err != nil {
			tc.Fatal(err)
		}
		before, _ := json.Marshal(original.Facts)
		after, _ := json.Marshal(updated.Facts)
		if string(before) != string(after) || original.Receipt.EvidenceDigest != updated.Receipt.EvidenceDigest || original.Receipt.SnapshotDigest != updated.Receipt.SnapshotDigest {
			tc.Fatal("invisible evidence changed the historical answer")
		}
		if original.Receipt.InputManifestDigest == updated.Receipt.InputManifestDigest {
			tc.Fatal("input manifest failed to distinguish the appended source input")
		}
	}, options(200)...)
}

type swarmMachine struct{ w *world }

func (m *swarmMachine) step(tc hegel.TestCase, action string) {
	if _, err := m.w.apply(action); err != nil {
		panic(err)
	}
}
func (m *swarmMachine) RuleAccept(tc hegel.TestCase)    { m.step(tc, "accept") }
func (m *swarmMachine) RuleTerminal(tc hegel.TestCase)  { m.step(tc, "terminal") }
func (m *swarmMachine) RuleDuplicate(tc hegel.TestCase) { m.step(tc, "duplicate") }
func (m *swarmMachine) RuleAdvance(tc hegel.TestCase) {
	m.step(tc, hegel.Draw(tc, hegel.SampledFrom([]string{"advance95", "advance96", "advance100"})))
}
func (m *swarmMachine) RuleSnapshot(tc hegel.TestCase) { m.step(tc, "snapshot") }
func (m *swarmMachine) InvariantProjection(tc hegel.TestCase) {
	if err := m.w.check(); err != nil {
		panic(err)
	}
}

func TestHegelLifecycleStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		m := &swarmMachine{newWorld()}
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(30), hegel.WithAlwaysCheckInvariants("InvariantProjection"))
	}, options(150)...)
}
