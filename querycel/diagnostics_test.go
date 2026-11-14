package querycel

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestCompileDiagnostics(t *testing.T) {
	tests := []struct {
		name, signal, source, code, phase string
		span                              bool
	}{
		{"empty", "logs", "", CodeInput, PhaseInput, false},
		{"invalid utf8", "logs", string([]byte{0xff}), CodeInput, PhaseInput, false},
		{"byte limit", "logs", strings.Repeat("é", MaxExpressionBytes/2+1), CodeBudget, PhaseInput, false},
		{"signal binding", "metrics", `log.body == "x"`, CodeBinding, PhaseBinding, false},
		{"parser recursion budget", "logs", strings.Repeat("(", 80) + `log.body == "x"` + strings.Repeat(")", 80), CodeSyntax, PhaseParse, false},
		{"parse", "logs", `log.body ==`, CodeSyntax, PhaseParse, true},
		{"source cannot forge budget", "logs", `log.body == "exceeded recursion limit" && unknown == "x"`, CodeType, PhaseCheck, true},
		{"type", "logs", `log.severity_number >= 17`, CodeType, PhaseCheck, true},
		{"result type", "logs", `log.body`, CodeType, PhaseCheck, true},
		{"unsupported", "logs", `log.body.matches("x")`, CodeUnsupported, PhaseLower, true},
		{"unsupported dead branch", "logs", `log.body == "never" && false && log.body.matches("x")`, CodeUnsupported, PhaseLower, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := Compile(tt.signal, tt.source)
			if program != nil || err == nil {
				t.Fatalf("got program %v, error %v", program, err)
			}
			var d *Diagnostic
			if !errors.As(fmt.Errorf("authoring: %w", err), &d) {
				t.Fatalf("untyped error %T", err)
			}
			if d.Code != tt.code || d.Phase != tt.phase {
				t.Fatalf("got %s/%s: %v", d.Code, d.Phase, err)
			}
			if (d.Span != nil) != tt.span {
				t.Fatalf("span = %+v", d.Span)
			}
			if d.Span != nil && (d.Span.Start < 0 || d.Span.End < d.Span.Start || d.Span.End > utf8.RuneCountInString(tt.source)) {
				t.Fatalf("invalid span %+v", d.Span)
			}
			if d.Message == "" || errors.Unwrap(d) == nil {
				t.Fatal("missing message/cause")
			}
		})
	}
}

func TestCompileDiagnosticUnicodePositions(t *testing.T) {
	source := "log.body == \"é🦀\" &&\n  unknown == \"x\""
	_, err := Compile("logs", source)
	var d *Diagnostic
	if !errors.As(err, &d) || d.Span == nil {
		t.Fatalf("missing diagnostic span: %v", err)
	}
	want := utf8.RuneCountInString(strings.Split(source, "unknown")[0])
	if d.Code != CodeType || d.Span.Start != want || d.Span.Line != 2 || d.Span.Column != 3 {
		t.Fatalf("got %+v, want rune offset %d, line 2 column 3", d, want)
	}
	source = `log.body == "é🦀" && log.body.matches("x")`
	_, err = Compile("logs", source)
	if !errors.As(err, &d) || d.Span == nil {
		t.Fatalf("missing lowering span: %v", err)
	}
	// CEL anchors member-call errors at the opening parenthesis, not the entire call.
	if d.Code != CodeUnsupported || d.Span.Start != utf8.RuneCountInString(strings.Split(source, "(")[0]) {
		t.Fatalf("got span %+v", d.Span)
	}
}

func TestCompileDiagnosticExpandedBudget(t *testing.T) {
	leaf := `otel.severityBand(log.severity_number) == "error"`
	_, err := Compile("logs", strings.Join([]string{leaf, leaf, leaf, leaf, leaf, leaf, leaf, leaf, leaf, leaf, leaf}, " || "))
	var d *Diagnostic
	if !errors.As(err, &d) || d.Code != CodeBudget || d.Phase != PhaseAdmission {
		t.Fatalf("expected ordinary grammar budget diagnostic, got %v", err)
	}
	if d.Path != "where" || d.Span != nil {
		t.Fatalf("domain admission must not invent source extent: %+v", d)
	}
}
