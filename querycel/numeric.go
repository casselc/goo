package querycel

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/ast"
	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"
	"cel.dev/cel-go/common/types/traits"

	"github.com/casselc/goo/exploration"
)

// The handle is internal to this library. It is neither a number nor a new
// query value: only a directly nested, literal constructor can be lowered.
var numericAttributeType = cel.OpaqueType("goo.numeric_attribute")

const numericPrefix = "goo_numeric_attribute_"

var numericOperations = []string{"eq", "ne", "gt", "gte", "lt", "lte", "valid", "missing", "empty", "invalid"}

func numericState(op string) bool {
	return op == "valid" || op == "missing" || op == "empty" || op == "invalid"
}
func numericID(op string) string { return numericPrefix + op + "_v1" }

type numericAttribute struct {
	attributes traits.Mapper
	binding    exploration.NumericBinding
}

func (v numericAttribute) Type() ref.Type { return numericAttributeType }
func (v numericAttribute) Value() any     { return v }
func (v numericAttribute) ConvertToNative(reflect.Type) (any, error) {
	return nil, fmt.Errorf("numeric attributes do not convert to native values")
}
func (v numericAttribute) ConvertToType(t ref.Type) ref.Val {
	if t == types.TypeType {
		return numericAttributeType
	}
	if t == numericAttributeType {
		return v
	}
	return types.NewErr("numeric attributes do not convert")
}
func (v numericAttribute) Equal(ref.Val) ref.Val {
	return types.NewErr("use an explicit numeric comparison")
}

func numericOptions() []cel.EnvOption {
	var options []cel.EnvOption
	for _, typ := range []string{"int64", "uint64", "decimal64"} {
		signature := []*cel.Type{cel.MapType(cel.StringType, cel.StringType), cel.StringType, cel.StringType}
		if typ == "decimal64" {
			signature = append(signature, cel.IntType)
		}
		options = append(options, cel.Function(typ, cel.MemberOverload(numericID(typ), signature, numericAttributeType, cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			expected := 3
			if typ == "decimal64" {
				expected = 4
			}
			if len(args) != expected {
				return types.NewErr("numeric declaration has the wrong arity")
			}
			attributes, ok := args[0].(traits.Mapper)
			key, keyOK := args[1].(types.String)
			unit, unitOK := args[2].(types.String)
			if !ok || !keyOK || !unitOK {
				return types.NewErr("numeric declaration requires a string map, key and unit")
			}
			b := exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: string(key)}, Type: typ, Unit: string(unit), Parse: "canonical-decimal"}
			if typ == "decimal64" {
				scale, ok := args[3].(types.Int)
				if !ok || scale < 1 || scale > 9 {
					return types.NewErr("decimal64 requires scale 1..9")
				}
				b.Scale = int(scale)
			}
			if err := b.Validate("spans"); err != nil {
				return types.NewErr("%s", err)
			}
			return numericAttribute{attributes, b}
		}))))
	}
	for _, op := range numericOperations {
		signature := []*cel.Type{numericAttributeType}
		if !numericState(op) {
			signature = append(signature, cel.StringType)
		}
		options = append(options, cel.Function(op, cel.MemberOverload(numericID(op), signature, cel.BoolType, cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			expected := 2
			if numericState(op) {
				expected = 1
			}
			if len(args) != expected {
				return types.NewErr("numeric condition has the wrong arity")
			}
			numeric, ok := args[0].(numericAttribute)
			if !ok {
				return types.NewErr("numeric condition requires an explicit declaration")
			}
			p := &exploration.Predicate{Numeric: &numeric.binding, Op: op}
			if !numericState(op) {
				threshold, ok := args[1].(types.String)
				if !ok {
					return types.NewErr("numeric threshold requires an exact string")
				}
				p.Value = string(threshold)
			}
			if err := exploration.ValidatePredicate(p, "spans"); err != nil {
				return types.NewErr("%s", err)
			}
			values := map[exploration.Field]string{}
			if raw, present := numeric.attributes.Find(types.String(numeric.binding.Field.Key)); present {
				value, ok := raw.(types.String)
				if !ok {
					return types.NewErr("numeric source attribute requires string")
				}
				values[numeric.binding.Field] = string(value)
			}
			// The existing grammar owns canonical parsing and comparison. No float or
			// independent CEL numeric parser is introduced.
			return types.Bool(exploration.Match(p, values))
		}))))
	}
	return options
}

func numericOperation(id string) (string, bool) {
	for _, op := range numericOperations {
		if id == numericID(op) {
			return op, true
		}
	}
	return "", false
}
func (l *lowerer) numericPredicate(expression ast.Expr, op string) (*exploration.Predicate, error) {
	call := expression.AsCall()
	args := call.Args()
	expected := 1
	if numericState(op) {
		expected = 0
	}
	if !call.IsMemberFunction() || len(args) != expected || call.Target().Kind() != ast.CallKind {
		return nil, fmt.Errorf("numeric condition requires a directly nested literal declaration")
	}
	constructor := call.Target()
	id, err := l.resolved(constructor)
	if err != nil {
		return nil, err
	}
	typ := ""
	for _, candidate := range []string{"int64", "uint64", "decimal64"} {
		if id == numericID(candidate) {
			typ = candidate
		}
	}
	decl := constructor.AsCall()
	parts := decl.Args()
	expected = 2
	if typ == "decimal64" {
		expected = 3
	}
	if typ == "" || !decl.IsMemberFunction() || len(parts) != expected {
		return nil, fmt.Errorf("numeric condition requires int64, uint64 or decimal64 attribute declaration")
	}
	field, err := l.attribute(decl.Target(), parts[0])
	if err != nil {
		return nil, err
	}
	unit, err := literal(parts[1], cel.StringType)
	if err != nil {
		return nil, fmt.Errorf("numeric unit requires a literal string")
	}
	binding := exploration.NumericBinding{Field: field.field, Type: typ, Unit: unit, Parse: "canonical-decimal"}
	if typ == "decimal64" {
		if parts[2].Kind() != ast.LiteralKind {
			return nil, fmt.Errorf("decimal scale requires a literal integer from 1 to 9")
		}
		scale, ok := parts[2].AsLiteral().(types.Int)
		if !ok || scale < 1 || scale > 9 {
			return nil, fmt.Errorf("decimal scale requires a literal integer from 1 to 9")
		}
		binding.Scale = int(scale)
	}
	if err = binding.Validate(l.signal); err != nil {
		return nil, err
	}
	p := &exploration.Predicate{Numeric: &binding, Op: op}
	if !numericState(op) {
		p.Value, err = literal(args[0], cel.StringType)
		if err != nil {
			return nil, fmt.Errorf("numeric threshold requires an exact canonical string literal")
		}
	}
	if err = exploration.ValidatePredicate(p, l.signal); err != nil {
		return nil, err
	}
	return p, nil
}

func printNumericPredicate(p *exploration.Predicate, path string) (string, error) {
	b := p.Numeric
	name := "resource.attributes"
	if b.Field.Kind == "attribute" {
		name = "record.attributes"
	}
	declaration := name + "." + b.Type + "(" + strconv.Quote(b.Field.Key) + ", " + strconv.Quote(b.Unit)
	if b.Type == "decimal64" {
		declaration += ", " + strconv.Itoa(b.Scale)
	}
	declaration += ")"
	if _, ok := numericOperation(numericID(p.Op)); !ok {
		return "", representationError(CodeUnsupported, path, fmt.Errorf("numeric operator has no CEL representation"))
	}
	threshold := ""
	if !numericState(p.Op) {
		threshold = strconv.Quote(p.Value)
	}
	return declaration + "." + p.Op + "(" + threshold + ")", nil
}

func numericCost(id string) bool { return strings.HasPrefix(id, numericPrefix) }

func numericConstructor(id string) bool {
	return id == numericID("int64") || id == numericID("uint64") || id == numericID("decimal64")
}
