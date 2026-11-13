package actions420

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"hegel.dev/go/hegel"
)

type condition struct {
	Connected, Hidden, Inert, Disabled, Fieldset, Legend, Closed, OuterClosed, ARIA, DocumentHidden, CSSHidden, Current bool
}

type expected struct {
	Invoke, Complete bool
	Focus            string
	Opened           int
}

func oracle(c condition, reveal bool) expected {
	// This oracle knows the generated tree's state; it does not reproduce selectors
	// or call the implementation's readiness functions.
	available := c.Connected && !c.Hidden && !c.Inert && !c.ARIA && !c.DocumentHidden && !c.CSSHidden && (!c.Fieldset || c.Legend)
	visible := available && !c.Closed && !c.OuterClosed
	out := expected{Invoke: c.Current && visible && !c.Disabled, Complete: c.Current && visible, Focus: "fallback"}
	if c.DocumentHidden {
		out.Focus = "none"
	}
	if available && !c.Disabled && (reveal || !c.Closed && !c.OuterClosed) {
		out.Focus = "origin"
		if reveal {
			if c.Closed {
				out.Opened++
			}
			if c.OuterClosed {
				out.Opened++
			}
		}
	}
	return out
}

type sample struct {
	Condition condition
	Reveal    bool
	Expected  expected
}

func TestGeneratedNativeReadinessAndExplicitReturn(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 4095))
		cases := []sample{}
		// Each generated schedule repeatedly withdraws/reinstates one generated
		// ancestor or owner state. Explicit controls cover available pending-disabled
		// completion and the native first-legend exception on every schedule.
		for step := range 64 {
			mask := (seed + step*137) & 4095
			c := condition{Connected: mask&1 == 0, Hidden: mask&2 != 0, Inert: mask&4 != 0, Disabled: mask&8 != 0, Fieldset: mask&16 != 0, Legend: mask&32 != 0, Closed: mask&64 != 0, OuterClosed: mask&128 != 0, ARIA: mask&256 != 0, DocumentHidden: mask&512 != 0, CSSHidden: mask&1024 != 0, Current: mask&2048 == 0}
			for _, reveal := range []bool{false, true} {
				cases = append(cases, sample{c, reveal, oracle(c, reveal)})
			}
		}
		for _, c := range []condition{{Connected: true, Current: true}, {Connected: true, Current: true, Disabled: true}, {Connected: true, Current: true, Fieldset: true, Legend: true}, {Connected: true, Current: true, Closed: true, OuterClosed: true}, {Connected: true, Current: false}} {
			for _, reveal := range []bool{false, true} {
				cases = append(cases, sample{c, reveal, oracle(c, reveal)})
			}
		}
		raw, err := json.Marshal(cases)
		if err != nil {
			tc.Fatalf("marshal: %v", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "node", "bridge.mjs")
		cmd.Stdin = bytes.NewReader(raw)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		out, err := cmd.Output()
		if err != nil {
			tc.Fatalf("actual native helpers: %v %s", err, stderr.String())
		}
		var got int
		if err = json.Unmarshal(out, &got); err != nil || got != len(cases) {
			tc.Fatalf("bridge lost cases: %s %v", out, err)
		}
	}, hegel.WithTestCases(32), hegel.WithSeed(420), hegel.WithDatabase(filepath.Join(t.TempDir(), "hegel")))
}
