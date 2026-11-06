package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"hegel.dev/go/hegel"
)

type fixedEditorSchedule358 struct {
	mask    int
	actions []map[string]any
}

func (m *fixedEditorSchedule358) RuleToggle(tc hegel.TestCase) {
	if m.mask&1 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "toggle"})
	}
}
func (m *fixedEditorSchedule358) RuleEdit(tc hegel.TestCase) {
	if m.mask&2 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "edit", "invalid": hegel.Draw(tc, hegel.Booleans())})
	}
}
func (m *fixedEditorSchedule358) RuleRun(tc hegel.TestCase) {
	if m.mask&4 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "run"})
	}
}
func (m *fixedEditorSchedule358) RuleComplete(tc hegel.TestCase) {
	if m.mask&8 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "complete", "stale": hegel.Draw(tc, hegel.Booleans()), "failed": hegel.Draw(tc, hegel.Booleans())})
	}
}
func (m *fixedEditorSchedule358) RuleStage(tc hegel.TestCase) {
	if m.mask&16 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "stage"})
	}
}
func (m *fixedEditorSchedule358) RuleCancel(tc hegel.TestCase) {
	if m.mask&32 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "cancel"})
	}
}
func (m *fixedEditorSchedule358) RuleVisibility(tc hegel.TestCase) {
	if m.mask&64 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "visibility", "kind": []string{"covered", "hidden", "detached"}[hegel.Draw(tc, hegel.Integers(0, 2))]})
	}
}
func (m *fixedEditorSchedule358) RuleBack(tc hegel.TestCase) {
	if m.mask&128 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "back"})
	}
}
func (m *fixedEditorSchedule358) RuleOld(tc hegel.TestCase) {
	if m.mask&256 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "old"})
	}
}
func (m *fixedEditorSchedule358) RuleUpdate(tc hegel.TestCase) {
	if m.mask&512 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "update"})
	}
}
func fixedEditorQuestion358(seed, mode, profile int) exploration.Query {
	q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1001 + int64(seed), ToMS: 61001 + int64(seed), Limit: 100, Offset: 0, Where: &exploration.Predicate{All: []exploration.Predicate{{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: fmt.Sprintf("api'π-%d", seed)}, {Any: []exploration.Predicate{{Field: &exploration.Field{Kind: "attribute", Key: "release"}, Op: "missing"}, {Field: &exploration.Field{Kind: "attribute", Key: "release"}, Op: "eq", Value: ""}}}}}, GroupBy: []exploration.Field{{Kind: "column", Key: "name"}}}
	if mode > 0 {
		s := identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}}
		if mode == 1 {
			s.Policy = "current-only"
		}
		if mode == 3 {
			s.Scope = identity.EnvironmentScope{Mode: "empty"}
		}
		if mode == 4 {
			s.Scope = identity.EnvironmentScope{Mode: "unreported"}
		}
		q.Environment = ""
		q.EnvironmentSelector = &s
	}
	switch profile {
	case 0:
		q.Signal = "logs"
		q.GroupBy = []exploration.Field{{Kind: "column", Key: "service"}}
	case 1:
		q.Summary = "span-duration"
		q.SpanStatus = true
		q.BucketMS = 1000
	case 2:
		q.Numeric = &exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "bytes"}, Type: "int64", Parse: "canonical-decimal", Unit: "By"}
	case 3:
		q.Numeric = &exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "bytes"}, Type: "uint64", Parse: "canonical-decimal", Unit: "By"}
		q.NumericHistogram = &exploration.NumericHistogramSpec{Binning: "powers-of-two"}
		q.GroupBy = nil
	case 4:
		q.GroupTimeSeries = true
		q.BucketMS = 1000
	case 5:
		q.ReportedEdge = &exploration.ReportedEdge{Parent: exploration.EdgeService{Service: "parent", Namespace: &identity.Namespace{Mode: "missing"}}, Child: exploration.EdgeService{Service: fmt.Sprintf("api'π-%d", seed), Namespace: &identity.Namespace{Mode: "missing"}}}
	}
	return q
}
func TestHegelFixedQuestionEditor358(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		m := &fixedEditorSchedule358{mask: hegel.Draw(tc, hegel.Integers(1, 1023))}
		q := fixedEditorQuestion358(hegel.Draw(tc, hegel.Integers(1, 10000)), hegel.Draw(tc, hegel.Integers(0, 4)), hegel.Draw(tc, hegel.Integers(0, 5)))
		if err := q.Validate(); err != nil {
			tc.Fatal("generated independent question invalid", err)
		}
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(25))
		raw, _ := json.Marshal(map[string]any{"question": q, "actions": m.actions})
		cmd := exec.CommandContext(t.Context(), "node", "fixed_question_358_fixture.mjs")
		cmd.Stdin = bytes.NewReader(raw)
		out, err := cmd.Output()
		if err != nil {
			if failure, ok := err.(*exec.ExitError); ok {
				tc.Fatal("actual presentation owner/session", err, string(failure.Stderr))
			} else {
				tc.Fatal(err)
			}
			return
		}
		var verdict struct {
			Complete, Leases bool
			Checks, Reads    int
		}
		if json.Unmarshal(out, &verdict) != nil || !verdict.Complete || !verdict.Leases || verdict.Checks < 20 || verdict.Reads < 3 {
			tc.Fatal("incomplete actual presentation schedule", string(out))
		}
	}, options(40)...)
}
