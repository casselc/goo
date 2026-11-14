package querycel

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/ast"
	"cel.dev/cel-go/common/overloads"
	"cel.dev/cel-go/common/types"

	"github.com/casselc/goo/exploration"
)

// Compile typechecks CEL, admits a source-compilable subset by resolved overload
// identity, and validates the expanded result against the ordinary Goo grammar.
func Compile(signal, expression string) (*Program, error) {
	if len(expression) == 0 || len(expression) > MaxExpressionBytes || !utf8.ValidString(expression) || strings.ContainsRune(expression, 0) {
		return nil, fmt.Errorf("CEL expression requires 1–%d well-formed bytes without NUL", MaxExpressionBytes)
	}
	fields, err := declarations(signal)
	if err != nil {
		return nil, err
	}
	options := []cel.EnvOption{cel.ClearMacros(), cel.ParserRecursionLimit(32), cel.ParserExpressionSizeLimit(MaxExpressionBytes), cel.ExpressionNodeLimit(128), cel.Lib(OTelLibrary{})}
	byName := make(map[string]fieldDeclaration, len(fields))
	for _, field := range fields {
		options = append(options, cel.Variable(field.name, field.typ))
		byName[field.name] = field
	}
	env, err := cel.NewEnv(options...)
	if err != nil {
		return nil, err
	}
	checked, issues := env.Compile(expression)
	if issues.Err() != nil {
		return nil, fmt.Errorf("CEL typecheck: %w", issues.Err())
	}
	if checked.OutputType() != cel.BoolType {
		return nil, fmt.Errorf("CEL predicate must return bool")
	}
	lower := lowerer{tree: checked.NativeRep(), fields: byName, used: map[string]fieldDeclaration{}}
	predicate, err := lower.predicate(checked.NativeRep().Expr(), 0)
	if err != nil {
		return nil, err
	}
	q := exploration.Query{Signal: signal, Environment: "cel-validation", FromMS: 0, ToMS: 60000, Limit: 100, Where: predicate}
	if err = q.Validate(); err != nil {
		return nil, fmt.Errorf("lowered CEL predicate: %w", err)
	}
	cost, err := env.EstimateCost(checked, costEstimator{})
	if err != nil {
		return nil, fmt.Errorf("CEL cost estimation: %w", err)
	}
	if cost.Max > MaxCost {
		return nil, fmt.Errorf("CEL expression exceeds work budget")
	}
	evaluator, err := env.Program(checked, cel.CostLimit(MaxCost))
	if err != nil {
		return nil, err
	}
	return &Program{signal: signal, predicate: predicate, evaluator: evaluator, fields: lower.used}, nil
}

type lowerer struct {
	tree   *ast.AST
	fields map[string]fieldDeclaration
	used   map[string]fieldDeclaration
	nodes  int
}

func (l *lowerer) resolved(expression ast.Expr) (string, error) {
	ids := l.tree.GetOverloadIDs(expression.ID())
	if len(ids) != 1 {
		return "", fmt.Errorf("CEL operation has no unique supported overload")
	}
	return ids[0], nil
}
func (l *lowerer) field(expression ast.Expr) (fieldDeclaration, bool) {
	if expression.Kind() != ast.IdentKind {
		return fieldDeclaration{}, false
	}
	field, ok := l.fields[expression.AsIdent()]
	if ok {
		l.used[field.name] = field
	}
	return field, ok
}
func (l *lowerer) predicate(expression ast.Expr, depth int) (*exploration.Predicate, error) {
	l.nodes++
	if l.nodes > 64 || depth > 16 {
		return nil, fmt.Errorf("CEL lowering exceeds expression budget")
	}
	if expression.Kind() != ast.CallKind {
		return nil, fmt.Errorf("CEL predicate requires a supported field comparison")
	}
	id, err := l.resolved(expression)
	if err != nil {
		return nil, err
	}
	call := expression.AsCall()
	args := call.Args()
	if id == overloads.LogicalAnd || id == overloads.LogicalOr {
		if call.IsMemberFunction() || len(args) < 2 {
			return nil, fmt.Errorf("unsupported CEL boolean operation")
		}
		children := make([]exploration.Predicate, len(args))
		for i, arg := range args {
			child, err := l.predicate(arg, depth+1)
			if err != nil {
				return nil, err
			}
			children[i] = *child
		}
		if id == overloads.LogicalAnd {
			return &exploration.Predicate{All: children}, nil
		}
		return &exploration.Predicate{Any: children}, nil
	}
	if id == overloads.ContainsString {
		if !call.IsMemberFunction() || len(args) != 1 {
			return nil, fmt.Errorf("unsupported CEL contains operation")
		}
		field, ok := l.field(call.Target())
		if !ok || field.typ != cel.StringType {
			return nil, fmt.Errorf("CEL contains requires a fixed string field")
		}
		value, err := literal(args[0], cel.StringType)
		if err != nil {
			return nil, err
		}
		return &exploration.Predicate{Field: &field.field, Op: "contains", Value: value}, nil
	}
	op := ""
	switch id {
	case overloads.Equals:
		op = "eq"
	case overloads.NotEquals:
		op = "ne"
	case overloads.GreaterEqualsUint64:
		op = "gte"
	case overloads.LessEqualsUint64:
		op = "lte"
	default:
		return nil, fmt.Errorf("CEL overload %s has no source lowering", id)
	}
	if call.IsMemberFunction() || len(args) != 2 {
		return nil, fmt.Errorf("unsupported CEL comparison")
	}
	if op == "eq" {
		if value, ok, err := l.severity(args[0], args[1]); ok || err != nil {
			return value, err
		}
		if value, ok, err := l.severity(args[1], args[0]); ok || err != nil {
			return value, err
		}
	}
	field, ok := l.field(args[0])
	constant := args[1]
	if !ok {
		field, ok = l.field(args[1])
		constant = args[0]
		if op == "gte" {
			op = "lte"
		} else if op == "lte" {
			op = "gte"
		}
	}
	if !ok {
		return nil, fmt.Errorf("CEL comparison requires one fixed field and one literal")
	}
	if op != "eq" && op != "ne" && field.typ != cel.UintType {
		return nil, fmt.Errorf("CEL ordering requires uint64")
	}
	value, err := literal(constant, field.typ)
	if err != nil {
		return nil, err
	}
	return &exploration.Predicate{Field: &field.field, Op: op, Value: value}, nil
}
func literal(expression ast.Expr, typ *cel.Type) (string, error) {
	if expression.Kind() != ast.LiteralKind {
		return "", fmt.Errorf("CEL source comparison requires a literal value")
	}
	switch value := expression.AsLiteral().(type) {
	case types.String:
		if typ == cel.StringType {
			return string(value), nil
		}
	case types.Uint:
		if typ == cel.UintType {
			return strconv.FormatUint(uint64(value), 10), nil
		}
	}
	return "", fmt.Errorf("CEL literal type differs from its field")
}
func (l *lowerer) severity(expression, constant ast.Expr) (*exploration.Predicate, bool, error) {
	if expression.Kind() != ast.CallKind {
		return nil, false, nil
	}
	id, err := l.resolved(expression)
	if err != nil {
		return nil, false, err
	}
	if id != severityOverload {
		return nil, false, nil
	}
	call := expression.AsCall()
	if call.IsMemberFunction() || len(call.Args()) != 1 {
		return nil, true, fmt.Errorf("unsupported severityBand call")
	}
	field, ok := l.field(call.Args()[0])
	if !ok || field.name != "log.severity_number" {
		return nil, true, fmt.Errorf("severityBand lowering requires log.severity_number")
	}
	band, err := literal(constant, cel.StringType)
	if err != nil {
		return nil, true, err
	}
	leaf := func(op, value string) exploration.Predicate {
		return exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "severity_number"}, Op: op, Value: value}
	}
	if band == "unspecified" {
		value := leaf("eq", "0")
		return &value, true, nil
	}
	if band == "invalid" {
		value := leaf("gte", "25")
		return &value, true, nil
	}
	for i, name := range []string{"trace", "debug", "info", "warn", "error", "fatal"} {
		if band == name {
			return &exploration.Predicate{All: []exploration.Predicate{leaf("gte", strconv.Itoa(i*4+1)), leaf("lte", strconv.Itoa(i*4+4))}}, true, nil
		}
	}
	return nil, true, fmt.Errorf("unknown numeric OTel severity band")
}
