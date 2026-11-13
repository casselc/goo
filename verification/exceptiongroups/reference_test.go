package exceptiongroups

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand/v2"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

type tuple struct {
	HasType  bool
	Type     string
	HasStack bool
	Stack    string
}
type expectedFingerprint struct{ ID, Reason, Type, Stack string }

func exactID(typ, stack string) string {
	wire := "goo.exception-fingerprint/1\n" + fmt.Sprint(len(typ)) + ":" + typ + fmt.Sprint(len(stack)) + ":" + stack
	sum := sha256.Sum256([]byte(wire))
	return "ef1:c:" + hex.EncodeToString(sum[:])
}
func oracle(values []tuple) expectedFingerprint {
	unclassified := func(reason string) expectedFingerprint {
		return expectedFingerprint{ID: "ef1:u:" + reason, Reason: reason}
	}
	if len(values) == 0 {
		return unclassified("none")
	}
	size := 0
	for _, v := range values {
		if len(v.Type) > 256 || len(v.Stack) > 8192 {
			return unclassified("oversize")
		}
		size += len(v.Type) + len(v.Stack)
	}
	if len(values) > 32 || size > 65536 {
		return unclassified("oversize")
	}
	for _, v := range values {
		if !utf8.ValidString(v.Type) || !utf8.ValidString(v.Stack) {
			return unclassified("invalid-utf8")
		}
	}
	unique := map[tuple]bool{}
	for _, v := range values {
		unique[v] = true
	}
	if len(unique) > 1 {
		return unclassified("multiple")
	}
	v := values[0]
	switch {
	case !v.HasType && !v.HasStack:
		return unclassified("both-missing")
	case !v.HasType:
		return unclassified("missing-type")
	case !v.HasStack:
		return unclassified("missing-stack")
	case v.Type == "" && v.Stack == "":
		return unclassified("both-empty")
	case v.Type == "":
		return unclassified("empty-type")
	case v.Stack == "":
		return unclassified("empty-stack")
	}
	return expectedFingerprint{ID: exactID(v.Type, v.Stack), Type: v.Type, Stack: v.Stack}
}
func attrs(v tuple) map[string]string {
	m := map[string]string{}
	if v.HasType {
		m["exception.type"] = v.Type
	}
	if v.HasStack {
		m["exception.stacktrace"] = v.Stack
	}
	return m
}
func checkFingerprint(t interface {
	Helper()
	Fatalf(string, ...any)
}, got exploration.ExceptionFingerprint, want expectedFingerprint) {
	t.Helper()
	state := "classified"
	if want.Reason != "" {
		state = "unclassified"
	}
	if got.State != state {
		t.Fatalf("classification state: got %q want %q", got.State, state)
	}
	if got.Fingerprint != want.ID || got.Reason != want.Reason {
		t.Fatalf("fingerprint identity/reason: got %+v want %+v", got, want)
	}
	if want.Reason == "" {
		if got.Inputs == nil || got.Inputs.Type != want.Type || got.Inputs.Stacktrace != want.Stack {
			t.Fatalf("classified inputs were not exact: %+v", got)
		}
	} else if got.Inputs != nil {
		t.Fatalf("unclassified result exposed fabricated/truncated inputs: %+v", got)
	}
	if err := exploration.ValidateExceptionFingerprint(got.Fingerprint); err != nil {
		t.Fatalf("own identity refused: %v", err)
	}
}
func TestExactExceptionFingerprintStateAndFraming(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		rng := rand.New(rand.NewPCG(uint64(seed), 416))
		complete := tuple{true, fmt.Sprintf("Error-%d-雪", seed), true, "frame\n at worker.go:42\twith literal 123 and \x00"}
		values := []tuple{{}, {false, "", true, ""}, {true, "", false, ""}, {true, "", true, ""}, {true, "Error", true, ""}, {true, "", true, "frame"}, complete, {true, "A", true, "BC"}, {true, "AB", true, "C"}, {true, strings.Repeat("t", 256), true, "s"}, {true, strings.Repeat("t", 257), true, "s"}, {true, "T", true, strings.Repeat("é", 4096)}, {true, "T", true, strings.Repeat("é", 4096) + "x"}, {true, strings.Repeat("雪", 85), true, "f"}, {true, strings.Repeat("雪", 86), true, "f"}, {true, "T", true, string([]byte{0xff})}, {true, string([]byte{0xff}), true, "s"}}
		for range 20 {
			v := complete
			v.Type += strings.Repeat("x", rng.IntN(4))
			v.Stack += strings.Repeat("\n0", rng.IntN(20))
			values = append(values, v)
		}
		for _, v := range values {
			input := attrs(v)
			before := map[string]string{}
			for k, x := range input {
				before[k] = x
			}
			got := exploration.ReferenceLogExceptionFingerprint(input)
			checkFingerprint(tc, got, oracle([]tuple{v}))
			if !reflect.DeepEqual(input, before) {
				tc.Fatalf("log projection mutated input")
			}
		}
		spans := [][]tuple{nil, {complete}, {complete, complete}, {{true, "A", true, "S1"}, {true, "B", true, "S2"}}, {{true, "A", false, ""}, {false, "", true, "S1"}}, {complete, {}}, {complete, {true, "", true, ""}}, {{true, " T ", true, " S "}}, {{true, "T", true, "S\n"}}, {{true, "T", true, "S\r\n"}}}
		for _, n := range []int{32, 33} {
			events := make([]tuple, n)
			for i := range events {
				events[i] = complete
			}
			spans = append(spans, events)
		}
		events := make([]tuple, 8)
		for i := range events {
			events[i] = tuple{true, "Type", true, strings.Repeat("s", 8188)}
		}
		spans = append(spans, events)
		over := append([]tuple(nil), events...)
		over[0].Stack += "x"
		spans = append(spans, over)
		for _, events := range spans {
			names := []string{"not-exception"}
			data := []map[string]string{{"exception.type": string([]byte{0xff}), "exception.stacktrace": strings.Repeat("x", 9000)}}
			for _, event := range events {
				names = append(names, "exception")
				data = append(data, attrs(event))
			}
			beforeNames := append([]string(nil), names...)
			beforeData := fmt.Sprint(data)
			got, err := exploration.ReferenceSpanExceptionFingerprint(names, data)
			if err != nil {
				tc.Fatalf("valid aligned event source refused: %v", err)
			}
			checkFingerprint(tc, got, oracle(events))
			if !reflect.DeepEqual(names, beforeNames) || fmt.Sprint(data) != beforeData {
				tc.Fatalf("span projection mutated source")
			}
		}
		if _, err := exploration.ReferenceSpanExceptionFingerprint([]string{"exception"}, nil); err == nil {
			tc.Fatalf("unaligned event arrays accepted")
		}
		if exactID("A", "BC") == exactID("AB", "C") || exactID("T", "S\n") == exactID("T", "S\r\n") {
			tc.Fatalf("framing or raw stack identity collapsed")
		}
		for _, bad := range []string{"", "ef1:c:" + strings.Repeat("A", 64), "ef1:c:" + strings.Repeat("0", 63), "ef1:u:unknown", exactID("T", "S") + "\n"} {
			if exploration.ValidateExceptionFingerprint(bad) == nil {
				tc.Fatalf("noncanonical fingerprint accepted: %q", bad)
			}
		}
	}, hegel.WithTestCases(32), hegel.WithSeed(416), hegel.WithDatabase(filepath.Join(t.TempDir(), "hegel")))
}
