package querycel

import (
	"unicode/utf8"
)

// Diagnostic identifies an authoring refusal independently of its human message.
// A missing Span means there is no trustworthy expression location (for example,
// an invalid signal binding). Path identifies a domain input, not a source span.
type Diagnostic struct {
	Code    string      `json:"code"`
	Phase   string      `json:"phase"`
	Message string      `json:"message"`
	Span    *SourceSpan `json:"span,omitempty"`
	Path    string      `json:"path,omitempty"`
	Cause   error       `json:"-"`
}

// SourceSpan uses zero-based Unicode code-point offsets [Start, End), never byte
// or UTF-16 offsets. Line and Column are one-based. A zero-width span marks a
// known insertion point or an AST anchor whose full extent CEL does not retain.
type SourceSpan struct {
	Start  int `json:"start"`
	End    int `json:"end"`
	Line   int `json:"line"`
	Column int `json:"column"`
}

const (
	CodeInput       = "input"
	CodeSyntax      = "syntax"
	CodeType        = "type"
	CodeUnsupported = "unsupported"
	CodeBudget      = "budget"
	CodeBinding     = "binding"
	CodeTarget      = "target"
	CodeInternal    = "internal"
	PhaseInput      = "input"
	PhaseParse      = "parse"
	PhaseCheck      = "check"
	PhaseLower      = "lower"
	PhaseAdmission  = "admission"
	PhaseCost       = "cost"
	PhaseBinding    = "binding"
	PhaseTarget     = "target"
)

func (d *Diagnostic) Error() string { return d.Message }
func (d *Diagnostic) Unwrap() error { return d.Cause }

func diagnostic(code, phase string, cause error) *Diagnostic {
	return &Diagnostic{Code: code, Phase: phase, Message: cause.Error(), Cause: cause}
}

func sourceSpan(source string, start, end int) *SourceSpan {
	if start < 0 || end < start || end > utf8.RuneCountInString(source) {
		return nil
	}
	line, column := 1, 1
	for i, r := range []rune(source) {
		if i == start {
			break
		}
		if r == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return &SourceSpan{Start: start, End: end, Line: line, Column: column}
}

var _ error = (*Diagnostic)(nil)
