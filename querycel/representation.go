package querycel

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"cel.dev/cel-go/cel"

	"github.com/casselc/goo/exploration"
)

const PhaseRepresentation = "representation"

// PrintPredicate returns CEL only when checked inverse lowering preserves the
// complete admitted predicate's identity, including branch order and nesting.
// It covers the fixed-field intersection, not every ordinary visual predicate.
// A refusal returns no partial expression; it never drops an unsupported leaf.
// Nil, presence, attributes, numeric bindings and noncanonical uint spellings
// are not convertible. No clock, source, store or query policy is involved.
func PrintPredicate(signal string, predicate *exploration.Predicate) (string, error) {
	declarations, err := declarations(signal)
	if err != nil {
		return "", representationError(CodeBinding, "signal", err)
	}
	if err := exploration.ValidatePredicate(predicate, signal); err != nil {
		code := CodeUnsupported
		if errors.Is(err, exploration.ErrPredicateBudget) {
			code = CodeBudget
		}
		return "", representationError(code, "where", err)
	}
	if predicate == nil {
		return "", representationError(CodeUnsupported, "where", fmt.Errorf("an unrestricted predicate has no supported CEL representation"))
	}
	byField := make(map[exploration.Field]fieldDeclaration, len(declarations))
	for _, field := range declarations {
		byField[field.field] = field
	}
	expression, err := printPredicate(predicate, byField, "where")
	if err != nil {
		return "", err
	}
	if len(expression) > MaxExpressionBytes {
		return "", representationError(CodeBudget, "where", fmt.Errorf("CEL representation exceeds %d bytes", MaxExpressionBytes))
	}
	program, err := Compile(signal, expression)
	if err != nil {
		return "", representationError(CodeUnsupported, "where", fmt.Errorf("CEL representation cannot preserve this predicate: %w", err))
	}
	// Typed JSON omits only the same optional empty leaf value/numeric members as
	// QuestionSession's predicateIdentity. Object-key order is irrelevant here:
	// both sides are the same Go shape. Branch order and nesting remain exact.
	original, err := json.Marshal(predicate)
	if err != nil {
		return "", representationError(CodeInternal, "where", err)
	}
	lowered, err := json.Marshal(program.Predicate())
	if err != nil {
		return "", representationError(CodeInternal, "where", err)
	}
	if !bytes.Equal(original, lowered) {
		return "", representationError(CodeUnsupported, "where", fmt.Errorf("CEL inverse lowering changes predicate structure or literal spelling"))
	}
	return expression, nil
}

func representationError(code, path string, cause error) *Diagnostic {
	d := diagnostic(code, PhaseRepresentation, cause)
	d.Path = path
	return d
}

func printPredicate(p *exploration.Predicate, fields map[exploration.Field]fieldDeclaration, path string) (string, error) {
	if p.All != nil || p.Any != nil {
		children, operator, branch := p.All, " && ", "all"
		if p.Any != nil {
			children, operator, branch = p.Any, " || ", "any"
		}
		if len(children) < 2 {
			return "", representationError(CodeUnsupported, path, fmt.Errorf("a one-child branch cannot retain its structure in CEL"))
		}
		parts := make([]string, len(children))
		for i := range children {
			text, err := printPredicate(&children[i], fields, fmt.Sprintf("%s.%s[%d]", path, branch, i))
			if err != nil {
				return "", err
			}
			parts[i] = text
		}
		return "(" + strings.Join(parts, operator) + ")", nil
	}
	if p.Numeric != nil || p.Field == nil {
		return "", representationError(CodeUnsupported, path, fmt.Errorf("numeric bindings have no fixed-field CEL representation"))
	}
	field, ok := fields[*p.Field]
	if !ok {
		return "", representationError(CodeUnsupported, path, fmt.Errorf("field has no fixed-field CEL representation"))
	}
	if p.Op == "contains" && field.typ == cel.StringType {
		return field.name + ".contains(" + strconv.Quote(p.Value) + ")", nil
	}
	operators := map[string]string{"eq": " == ", "ne": " != ", "gte": " >= ", "lte": " <= "}
	operator, ok := operators[p.Op]
	if !ok || (p.Op == "gte" || p.Op == "lte") && field.typ != cel.UintType {
		return "", representationError(CodeUnsupported, path, fmt.Errorf("operator has no fixed-field CEL representation"))
	}
	value := strconv.Quote(p.Value)
	if field.typ == cel.UintType {
		value = p.Value + "u"
	}
	return field.name + operator + value, nil
}
