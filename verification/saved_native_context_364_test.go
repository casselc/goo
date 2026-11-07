package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"hegel.dev/go/hegel"
)

// The mounted owner must preserve a stored source across suspension without
// allowing an old action lease, an ABA visit or a caller's interval to acquire it.
func TestHegelSavedNativeContext364(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		offset := hegel.Draw(tc, hegel.Integers(0, 999999))
		width := hegel.Draw(tc, hegel.Integers(1, 2000001))
		nsMode := hegel.Draw(tc, hegel.Integers(0, 2))
		var ns *identity.Namespace
		if nsMode > 0 {
			ns = &identity.Namespace{Mode: "value"}
			if nsMode == 2 {
				ns.Value = "team-π"
			}
		}
		actions := make([]int, 24)
		for i := range actions {
			actions[i] = hegel.Draw(tc, hegel.Integers(0, 6))
		}
		pages := []any{}
		details := []any{}
		for i, kind := range []string{"counter", "histogram"} {
			p, s := nrcInputs(ns, kind)
			p.Definition.Environment = ""
			p.Definition.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: []string{"current-only", "effective"}[i], Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}}
			if kind == "histogram" {
				p.Definition.Source.Unit = "s"
				s.Unit = "s"
			}
			from := int64(1790000000000000000) + int64(offset) + int64(i)*3000000
			to := from + int64(width)
			a, b := strconv.FormatInt(from, 10), strconv.FormatInt(to, 10)
			signal := []string{"logs", "spans"}[i]
			plan, err := exploration.PlanNativeRecordContext(p, s, signal, a, b)
			if err != nil {
				tc.Fatal(err)
			}
			// Independent integer membership at both excluded neighbours; no planner
			// predicates or UI detail are used to derive expected membership.
			for _, at := range []int64{from - 1, from, from + 1, to - 1, to, to + 1} {
				values := map[exploration.Field]string{{Kind: "column", Key: "service"}: "api", {Kind: "resource", Key: "service.instance.id"}: "instance-one", {Kind: "column", Key: "timestamp_ns"}: strconv.FormatInt(at, 10)}
				if ns != nil {
					values[exploration.Field{Kind: "resource", Key: "service.namespace"}] = ns.Value
				}
				if nrcPredicate(plan.Question.Where, values) != (at >= from && at < to) {
					tc.Fatal("exact retained interval membership")
				}
				values[exploration.Field{Kind: "resource", Key: "service.instance.id"}] = "foreign"
				if nrcPredicate(plan.Question.Where, values) {
					tc.Fatal("foreign instance admitted")
				}
			}
			basis := map[string]any{"source": "retained", "task": 364}
			readAt := "2026-11-07T10:00:00Z"
			result := map[string]any{"question": p, "basis": basis, "readAt": readAt, kind + "s": map[string]any{"series": []any{map[string]any{"stream": s, "intervals": []any{map[string]any{"fromNS": a, "toNS": b, "state": "ok"}}}}}}
			pages = append(pages, map[string]any{"receipt": map[string]any{"id": fmt.Sprintf("%064x", i+1), "question": map[string]any{"kind": "measurement", "measurement": p}}, "measurement": result})
			details = append(details, map[string]any{"question": plan.Question, "fromNS": a, "toNS": b, "grain": plan.Grain, "measurement": p, "stream": s, "basis": basis, "readAt": readAt})
		}
		encoded, err := json.Marshal(map[string]any{"pages": pages, "details": details, "actions": actions})
		if err != nil {
			tc.Fatal(err)
		}
		cmd := exec.CommandContext(t.Context(), "node", "--experimental-test-module-mocks", "saved_native_context_364_fixture.mjs")
		cmd.Stdin = bytes.NewReader(encoded)
		out, err := cmd.Output()
		if err != nil {
			if e, ok := err.(*exec.ExitError); ok {
				tc.Fatal(string(e.Stderr))
			}
			tc.Fatal(err)
		}
		var observed struct {
			Complete      bool
			Checks, Opens int
		}
		if err = json.Unmarshal(out, &observed); err != nil || !observed.Complete || observed.Checks < 20 || observed.Opens < 3 {
			tc.Fatal("incomplete actual saved owner schedule", string(out), err)
		}
	}, options(40)...)
}
