package verification

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"testing"

	"hegel.dev/go/hegel"
)

type widgetRecoverySchedule354 struct {
	mask    int
	actions []map[string]any
}

func (m *widgetRecoverySchedule354) RuleSettle(tc hegel.TestCase) {
	if m.mask&1 == 0 {
		return
	}
	outcomes := []string{"success", "timeout", "unavailable", "budget", "malformed", "cancelled", "network", "client", "mismatch", "aggregate"}
	m.actions = append(m.actions, map[string]any{"type": "settle", "key": hegel.Draw(tc, hegel.Integers(0, 3)), "outcome": outcomes[hegel.Draw(tc, hegel.Integers(0, len(outcomes)-1))]})
}
func (m *widgetRecoverySchedule354) RuleRetry(tc hegel.TestCase) {
	if m.mask&2 == 0 {
		return
	}
	m.actions = append(m.actions, map[string]any{"type": "retry", "key": hegel.Draw(tc, hegel.Integers(0, 3)), "double": hegel.Draw(tc, hegel.Booleans())})
}
func (m *widgetRecoverySchedule354) RuleBlocked(tc hegel.TestCase) {
	if m.mask&4 == 0 {
		return
	}
	modes := []string{"hidden", "inert", "detached"}
	m.actions = append(m.actions, map[string]any{"type": "blocked", "key": hegel.Draw(tc, hegel.Integers(0, 3)), "mode": modes[hegel.Draw(tc, hegel.Integers(0, 2))]})
}
func (m *widgetRecoverySchedule354) RuleReplace(tc hegel.TestCase) {
	if m.mask&8 == 0 {
		return
	}
	reasons := []string{"scope", "profile", "window"}
	m.actions = append(m.actions, map[string]any{"type": "replace", "reason": reasons[hegel.Draw(tc, hegel.Integers(0, 2))]})
}
func (m *widgetRecoverySchedule354) RuleRetire(tc hegel.TestCase) {
	if m.mask&16 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "retire"})
	}
}
func (m *widgetRecoverySchedule354) RuleLate(tc hegel.TestCase) {
	if m.mask&32 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "late"})
	}
}
func (m *widgetRecoverySchedule354) RuleOld(tc hegel.TestCase) {
	if m.mask&64 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "old"})
	}
}
func (m *widgetRecoverySchedule354) RuleAuthority(tc hegel.TestCase) {
	if m.mask&128 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "authority"})
	}
}

// Hegel's stateful rule selection and per-case rule mask form varying action
// swarms. Every transition runs against the actual mounted dashboard owner;
// the JavaScript driver checks its independent per-widget attempt/DOM oracle
// after each action rather than merely checking a generated expected predicate.
func TestHegelWidgetRecovery354(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		m := &widgetRecoverySchedule354{mask: hegel.Draw(tc, hegel.Integers(1, 255))}
		scope := hegel.Draw(tc, hegel.Integers(0, 4))
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(24))
		input, _ := json.Marshal(map[string]any{"operations": m.actions, "scopeMode": scope})
		cmd := exec.CommandContext(t.Context(), "node", "--experimental-test-module-mocks", "widget_recovery_354_fixture.mjs")
		cmd.Stdin = bytes.NewReader(input)
		raw, err := cmd.Output()
		if err != nil {
			if exit, ok := err.(*exec.ExitError); ok {
				tc.Fatal("actual mounted owner", err, string(exit.Stderr))
				return
			}
			tc.Fatal(err)
			return
		}
		var observed struct {
			Checks, MaxActive, Owners, Requests int
			Complete                            bool
			Critical                            struct{ Record, Native, Healthy int }
		}
		if json.Unmarshal(raw, &observed) != nil || !observed.Complete || observed.Checks < 15 || observed.MaxActive != 2 || observed.Critical.Record != 3 || observed.Critical.Native != 2 || observed.Critical.Healthy != 1 {
			tc.Fatal("incomplete or invalid actual recovery schedule", string(raw))
		}
	}, options(40)...)
}
