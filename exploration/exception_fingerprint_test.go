package exploration

import (
	"strings"
	"testing"
)

func TestExceptionFingerprint416PairsAndSpanGrain(t *testing.T) {
	pair := map[string]string{"exception.type": "TypeError", "exception.stacktrace": "frame\n  at worker:42"}
	base := ReferenceLogExceptionFingerprint(pair)
	if base.State != "classified" || base.Inputs.Type != "TypeError" || ValidateExceptionFingerprint(base.Fingerprint) != nil {
		t.Fatal(base)
	}
	for reason, attrs := range map[string]map[string]string{
		"both-missing": {}, "missing-type": {"exception.stacktrace": "x"}, "missing-stack": {"exception.type": "x"}, "both-empty": {"exception.type": "", "exception.stacktrace": ""}, "empty-type": {"exception.type": "", "exception.stacktrace": "x"}, "empty-stack": {"exception.type": "x", "exception.stacktrace": ""}, "oversize": {"exception.type": "x", "exception.stacktrace": strings.Repeat("x", 8193)}, "invalid-utf8": {"exception.type": "x", "exception.stacktrace": string([]byte{255})},
	} {
		got := ReferenceLogExceptionFingerprint(attrs)
		if got.State != "unclassified" || got.Reason != reason || got.Inputs != nil || ValidateExceptionFingerprint(got.Fingerprint) != nil {
			t.Fatal(reason, got)
		}
	}
	duplicated, e := ReferenceSpanExceptionFingerprint([]string{"exception", "ignored", "exception"}, []map[string]string{pair, {}, pair})
	if e != nil || duplicated.Fingerprint != base.Fingerprint {
		t.Fatal(duplicated, e)
	}
	mixed, e := ReferenceSpanExceptionFingerprint([]string{"exception", "exception"}, []map[string]string{{"exception.type": "TypeError"}, {"exception.stacktrace": "frame\n  at worker:42"}})
	if e != nil || mixed.Reason != "multiple" {
		t.Fatal("cross-event tuple synthesized", mixed, e)
	}
	if _, e = ReferenceSpanExceptionFingerprint([]string{"exception"}, nil); e == nil {
		t.Fatal("misaligned arrays accepted")
	}
	none, e := ReferenceSpanExceptionFingerprint([]string{"other"}, []map[string]string{pair})
	if e != nil || none.Reason != "none" {
		t.Fatal(none, e)
	}
	// Presence and exact bytes, including whitespace, Unicode and delimiter text,
	// remain identity inputs. No normalization or message fallback is performed.
	for _, attrs := range []map[string]string{{"exception.type": "TypeError ", "exception.stacktrace": pair["exception.stacktrace"]}, {"exception.type": "TypeError", "exception.stacktrace": pair["exception.stacktrace"] + " "}, {"exception.type": "TypeError", "exception.message": pair["exception.stacktrace"]}} {
		if ReferenceLogExceptionFingerprint(attrs).Fingerprint == base.Fingerprint {
			t.Fatal("different input collapsed")
		}
	}
	a := ReferenceLogExceptionFingerprint(map[string]string{"exception.type": "a1:b", "exception.stacktrace": "c"})
	b := ReferenceLogExceptionFingerprint(map[string]string{"exception.type": "a", "exception.stacktrace": "1:bc"})
	if a.Fingerprint == b.Fingerprint {
		t.Fatal("ambiguous framing")
	}
}

func TestExceptionFingerprint416CompilerAndExactPredicate(t *testing.T) {
	id := ReferenceLogExceptionFingerprint(map[string]string{"exception.type": "Error", "exception.stacktrace": "frame"}).Fingerprint
	for _, signal := range []string{"logs", "spans"} {
		q := Query{Signal: signal, Environment: "prod", FromMS: 1, ToMS: 60001, Limit: 100, Where: &Predicate{Field: &Field{Kind: "column", Key: "exception_fingerprint_v1"}, Op: "eq", Value: id}}
		p, e := Compile(q)
		if e != nil || !strings.Contains(p.Rows, "AS exception_fingerprint_v1") || !strings.Contains(p.Count, "SHA256") || !strings.Contains(p.Count, "goo.exception-fingerprint/1\\n") {
			t.Fatal(signal, e, p)
		}
		if signal == "spans" && !strings.Contains(p.Rows, "arrayZip(`Events.Name`, `Events.Attributes`)") {
			t.Fatal("event projection absent")
		}
		q.Where.Op = "contains"
		if _, e = Compile(q); e == nil {
			t.Fatal("substring identity allowed")
		}
		q.Where.Op = "eq"
		q.Where.Value = "ef1:c:bad"
		if _, e = Compile(q); e == nil {
			t.Fatal("noncanonical identity allowed")
		}
	}
}
