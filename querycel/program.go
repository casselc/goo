package querycel

import (
	"fmt"
	"strconv"
	"unicode/utf8"

	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/types"

	"github.com/casselc/goo/exploration"
)

type fieldDeclaration struct {
	name  string
	field exploration.Field
	typ   *cel.Type
}

func declarations(signal string) ([]fieldDeclaration, error) {
	if signal != "logs" && signal != "spans" {
		return nil, fmt.Errorf("CEL predicate requires logs or spans")
	}
	fields := []fieldDeclaration{}
	add := func(name, key string, typ *cel.Type) {
		fields = append(fields, fieldDeclaration{name, exploration.Field{Kind: "column", Key: key}, typ})
	}
	for _, key := range []string{"service", "trace_id", "span_id", "scope", "scope_version"} {
		add("record."+key, key, cel.StringType)
	}
	add("record.timestamp_ns", "timestamp_ns", cel.UintType)
	if signal == "logs" {
		add("log.body", "body", cel.StringType)
		add("log.severity_text", "severity", cel.StringType)
		add("log.severity_number", "severity_number", cel.UintType)
	} else {
		for _, key := range []string{"name", "kind", "status"} {
			add("span."+key, key, cel.StringType)
		}
		add("span.duration_ns", "duration_ns", cel.UintType)
	}
	return fields, nil
}

// Program retains a checked CEL evaluator and a detached normalized predicate.
// Its exported methods never expose mutable compiler state.
type Program struct {
	signal    string
	predicate *exploration.Predicate
	evaluator cel.Program
	fields    map[string]fieldDeclaration
}

func (p *Program) Predicate() *exploration.Predicate {
	if p == nil || p.evaluator == nil || p.predicate == nil {
		return nil
	}
	// The compiler admits this private predicate before constructing Program.
	clone, err := exploration.ClonePredicate(p.predicate, p.signal)
	if err != nil {
		return nil
	}
	return clone
}

// Apply adds the compiled predicate with AND; it never replaces the enclosing
// selector, changes a clock, removes projections or resets delivery limits.
func (p *Program) Apply(parent exploration.Query) (exploration.Query, error) {
	if p == nil || p.evaluator == nil || p.predicate == nil {
		return exploration.Query{}, fmt.Errorf("CEL program is unavailable")
	}
	if parent.Signal != p.signal {
		return exploration.Query{}, fmt.Errorf("CEL predicate signal differs from parent question")
	}
	return exploration.Refine(parent, p.predicate)
}

// Eval executes the checked CEL reference program over already normalized fixed
// columns. Missing fields and malformed values are errors, never default values.
// Only columns referenced by this expression are required.
func (p *Program) Eval(values map[exploration.Field]string) (bool, error) {
	if p == nil || p.evaluator == nil || p.predicate == nil {
		return false, fmt.Errorf("CEL program is unavailable")
	}
	activation := make(map[string]any, len(p.fields))
	for name, declaration := range p.fields {
		value, ok := values[declaration.field]
		if !ok {
			return false, fmt.Errorf("missing fixed CEL field %s", name)
		}
		if declaration.typ == cel.UintType {
			n, err := strconv.ParseUint(value, 10, 64)
			if err != nil || strconv.FormatUint(n, 10) != value {
				return false, fmt.Errorf("CEL field %s requires canonical uint64", name)
			}
			activation[name] = n
		} else {
			if len(value) > MaxFieldBytes || !utf8.ValidString(value) {
				return false, fmt.Errorf("CEL field %s requires bounded well-formed text", name)
			}
			activation[name] = value
		}
	}
	value, _, err := p.evaluator.Eval(activation)
	if err != nil {
		return false, fmt.Errorf("CEL evaluation failed: %w", err)
	}
	result, ok := value.(types.Bool)
	if !ok {
		return false, fmt.Errorf("CEL predicate did not return bool")
	}
	return bool(result), nil
}
