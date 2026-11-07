package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"hegel.dev/go/hegel"
)

type retainedSchedule361 struct {
	mask    int
	actions []map[string]any
}

func (m *retainedSchedule361) RulePin(tc hegel.TestCase) {
	if m.mask&1 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "pin"})
	}
}
func (m *retainedSchedule361) RuleCompare(tc hegel.TestCase) {
	if m.mask&2 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "compare"})
	}
}
func (m *retainedSchedule361) RuleCurrent(tc hegel.TestCase) {
	if m.mask&4 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "current", "frame": hegel.Draw(tc, hegel.Integers(0, 4)), "location": hegel.Draw(tc, hegel.Integers(0, 3))})
	}
}
func (m *retainedSchedule361) RuleSuspend(tc hegel.TestCase) {
	if m.mask&8 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "suspend"})
	}
}
func (m *retainedSchedule361) RuleResume(tc hegel.TestCase) {
	if m.mask&16 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "resume"})
	}
}
func (m *retainedSchedule361) RuleClear(tc hegel.TestCase) {
	if m.mask&32 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "clear"})
	}
}
func (m *retainedSchedule361) RuleOpen(tc hegel.TestCase) {
	if m.mask&64 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "open", "side": hegel.Draw(tc, hegel.Integers(0, 1))})
	}
}
func (m *retainedSchedule361) RuleOld(tc hegel.TestCase) {
	if m.mask&128 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "old"})
	}
}
func (m *retainedSchedule361) RuleHidden(tc hegel.TestCase) {
	if m.mask&256 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "hidden"})
	}
}

type retainedFrame361 struct {
	Capture  exploration.CapturePage `json:"capture"`
	Location struct {
		ID     string `json:"id"`
		Cursor string `json:"cursor"`
		Offset int    `json:"offset"`
	} `json:"location"`
	Truncated bool `json:"truncated"`
}

func retainedFrames361(seed, width, n, mode, policy int, partial bool) []retainedFrame361 {
	a, b := tcPages(seed, width, n, mode)
	c := []json.RawMessage{tcSet(a[0], "timestamp_ns", "not-reported"), tcRow(fmt.Sprintf("%032x", seed+2), "0000000000000020", "", "selection-only", "zero", "Internal", 9007199255741093, 0)}
	data := [][]json.RawMessage{a, b, c, {}, {}}
	out := []retainedFrame361{}
	for index, rows := range data {
		q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 9007199254 + int64(index), ToMS: 9007259254 + int64(index), Limit: 100, Offset: 0}
		scopePolicy := (policy + index) % 5
		if scopePolicy > 0 {
			s := identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: fmt.Sprintf("prod-%d", index)}}
			if scopePolicy == 1 {
				s.Policy = "current-only"
			}
			if scopePolicy == 3 {
				s.Scope = identity.EnvironmentScope{Mode: "empty"}
			}
			if scopePolicy == 4 {
				s.Scope = identity.EnvironmentScope{Mode: "unreported"}
			}
			q.Environment = ""
			q.EnvironmentSelector = &s
		}
		ns := ""
		if index%2 == 0 {
			q.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "missing"}
		} else {
			q.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "eq", Value: ns}
		}
		if index == 2 {
			q = out[1].Capture.Explore.Question
			q.Offset = 100
		}
		if index == 4 {
			q.Signal = "logs"
		}
		if rows == nil {
			rows = []json.RawMessage{}
		}
		enriched := []json.RawMessage{}
		for position, raw := range rows {
			var value map[string]any
			json.Unmarshal(raw, &value)
			value["Attributes"] = map[string]any{"original.position": strconv.Itoa(position), "frame": index}
			resource := map[string]string{}
			if index%2 == 1 || index == 2 {
				resource["service.namespace"] = ns
			}
			value["ResourceAttributes"] = resource
			v, _ := json.Marshal(value)
			enriched = append(enriched, v)
		}
		id := strings.Repeat(string(rune('a'+index)), 64)
		if index == 2 {
			id = out[1].Capture.Receipt.ID
		}
		rowPartial := partial && index == 0 || index == 1 || index == 2
		count := strconv.Itoa(len(enriched))
		retained := len(enriched)
		if rowPartial {
			count = "900"
			retained = 200
		}
		state := "observed"
		if len(enriched) == 0 {
			state = "no_data"
			count = "0"
			retained = 0
			rowPartial = false
		}
		receiptQ := q
		receiptQ.Offset = 0
		if index == 2 {
			receiptQ = *out[1].Capture.Receipt.Question.Explore
		}
		f := retainedFrame361{Capture: exploration.CapturePage{Receipt: exploration.CaptureReceipt{Version: 1, ID: id, Title: fmt.Sprintf("Page fixture%d", index), CapturedAt: fmt.Sprintf("2026-11-06T00:00:%02dZ", index+10), Question: exploration.CaptureQuestion{Kind: "explore", Explore: &receiptQ}}, Explore: &exploration.Result{Question: q, Basis: strings.Repeat(string(rune('1'+index)), 64), CapturedAt: fmt.Sprintf("2026-11-06T00:00:%02dZ", index), Count: count, State: state, Rows: enriched, Groups: []json.RawMessage{}, Fields: []exploration.Field{}, Warnings: []string{}, HasMore: rowPartial}, RetainedRecords: retained, RecordsLimited: rowPartial}, Truncated: rowPartial || q.Offset > 0}
		if rowPartial {
			f.Capture.NextCursor = "known-next"
		}
		if index == 2 {
			f.Capture.Receipt = out[1].Capture.Receipt
			f.Capture.Explore.Basis = out[1].Capture.Explore.Basis
			f.Capture.Explore.CapturedAt = out[1].Capture.Explore.CapturedAt
		}
		f.Location.ID = id
		f.Location.Offset = q.Offset
		f.Location.Cursor = ""
		if index == 1 {
			f.Location.Cursor = "b3BhcXVlLXplcm8tcGFnZS1jdXJzb3I"
		}
		if index == 2 {
			f.Location.Cursor = "b3BhcXVlLWN1cnJlbnQtcGFnZS1jdXJzb3I"
		}
		out = append(out, f)
	}
	return out
}
func TestHegelRetainedComparison361(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		frames := retainedFrames361(hegel.Draw(tc, hegel.Integers(1, 10000)), hegel.Draw(tc, hegel.Integers(1, 80)), hegel.Draw(tc, hegel.Integers(1, 7)), hegel.Draw(tc, hegel.Integers(0, 14)), hegel.Draw(tc, hegel.Integers(0, 4)), hegel.Draw(tc, hegel.Booleans()))
		expected := map[string]*exploration.TracePageComparison{}
		before := savedJSON360(frames)
		for i := range 4 {
			for j := range 4 {
				a, b := frames[i], frames[j]
				actual, err := exploration.CompareTracePages(a.Capture.Explore.Rows, b.Capture.Explore.Rows, a.Truncated, b.Truncated)
				if err != nil || !tcOracleEqual(tcOracle(a.Capture.Explore.Rows, b.Capture.Explore.Rows, a.Truncated, b.Truncated), *actual) {
					tc.Fatal("independent raw endpoint/group/index comparison oracle", i, j, err)
				}
				expected[fmt.Sprintf("%d:%d", i, j)] = actual
			}
		}
		if savedJSON360(frames) != before {
			tc.Fatal("Go comparison mutated supplied operands")
		}
		m := &retainedSchedule361{mask: hegel.Draw(tc, hegel.Integers(1, 511))}
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(24))
		raw, _ := json.Marshal(map[string]any{"frames": frames, "expected": expected, "actions": m.actions})
		cmd := exec.CommandContext(t.Context(), "node", "--experimental-test-module-mocks", "retained_comparison_361_fixture.mjs")
		cmd.Stdin = bytes.NewReader(raw)
		out, err := cmd.Output()
		if err != nil {
			if failure, ok := err.(*exec.ExitError); ok {
				tc.Fatal("actual retained comparison owner", err, string(failure.Stderr))
			} else {
				tc.Fatal(err)
			}
			return
		}
		var verdict struct {
			Complete                              bool
			Checks, Telemetry, Comparisons, Opens int
		}
		if json.Unmarshal(out, &verdict) != nil || !verdict.Complete || verdict.Checks < 20 || verdict.Telemetry != 0 || verdict.Comparisons < 2 || verdict.Opens < 2 {
			tc.Fatal("incomplete retained-operand schedule", string(out))
		}
	}, options(40)...)
}
