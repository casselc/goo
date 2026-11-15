// Package querycel translates a bounded, checked CEL predicate into Goo's
// ordinary record grammar. It never reads telemetry or executes source SQL.
package querycel

import (
	"strings"

	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/checker"
	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"
	"cel.dev/cel-go/common/types/traits"

	"github.com/casselc/goo/exploration"
)

const (
	LibraryVersion                   = "goo.cel.otel/v1"
	MaxExpressionBytes               = 4096
	MaxFieldBytes                    = 65536
	MaxCost                   uint64 = 1000000
	severityOverload                 = "goo_otel_severity_band_uint_v1"
	attributeEqualOverload           = "goo_attribute_eq_v1"
	attributeNotEqualOverload        = "goo_attribute_ne_v1"
	attributeContainsOverload        = "goo_attribute_contains_v1"
)

// OTelLibrary declares the portable helpers and their bounded pure runtime
// implementations. The query compiler separately admits this resolved overload.
type OTelLibrary struct{}

func (OTelLibrary) LibraryName() string { return LibraryVersion }
func (OTelLibrary) CompileOptions() []cel.EnvOption {
	options := []cel.EnvOption{cel.Function("otel.severityBand", cel.Overload(severityOverload,
		[]*cel.Type{cel.UintType}, cel.StringType,
		cel.UnaryBinding(func(value ref.Val) ref.Val {
			n, ok := value.(types.Uint)
			if !ok {
				return types.NewErr("severityBand requires uint")
			}
			return types.String(exploration.LogSeverityClass(uint64(n)))
		})))}
	for _, operation := range []struct{ name, id string }{{"eq", attributeEqualOverload}, {"ne", attributeNotEqualOverload}, {"contains", attributeContainsOverload}} {
		op := operation
		options = append(options, cel.Function(op.name, cel.MemberOverload(op.id,
			[]*cel.Type{cel.MapType(cel.StringType, cel.StringType), cel.StringType, cel.StringType}, cel.BoolType,
			cel.FunctionBinding(func(args ...ref.Val) ref.Val {
				if len(args) != 3 {
					return types.NewErr("attribute comparison requires map, key and value")
				}
				attributes, ok := args[0].(traits.Mapper)
				if !ok {
					return types.NewErr("attribute comparison requires a string map")
				}
				key, keyOK := args[1].(types.String)
				expected, valueOK := args[2].(types.String)
				if !keyOK || !valueOK {
					return types.NewErr("attribute comparison requires string key and value")
				}
				actual, found := attributes.Find(key)
				if !found {
					return types.False
				}
				value, ok := actual.(types.String)
				if !ok {
					return types.NewErr("attribute value requires string")
				}
				switch op.id {
				case attributeEqualOverload:
					return types.Bool(value == expected)
				case attributeNotEqualOverload:
					return types.Bool(value != expected)
				default:
					return types.Bool(strings.Contains(string(value), string(expected)))
				}
			}))))
	}
	return append(options, numericOptions()...)
}
func (OTelLibrary) ProgramOptions() []cel.ProgramOption { return nil }

type costEstimator struct{}

func (costEstimator) EstimateSize(node checker.AstNode) *checker.SizeEstimate {
	if node.Type() == cel.StringType {
		return &checker.SizeEstimate{Min: 0, Max: MaxFieldBytes}
	}
	return nil
}
func (costEstimator) EstimateCallCost(_ string, id string, _ *checker.AstNode, _ []checker.AstNode) *checker.CallEstimate {
	if numericConstructor(id) {
		// Admission bounds literal keys to 128 bytes and units to 64. A
		// constructor validates only that metadata; it never reads a map value.
		return &checker.CallEstimate{CostEstimate: checker.CostEstimate{Min: 1, Max: 256}}
	}
	if numericCost(id) || id == attributeEqualOverload || id == attributeNotEqualOverload || id == attributeContainsOverload {
		return &checker.CallEstimate{CostEstimate: checker.CostEstimate{Min: 1, Max: MaxFieldBytes}}
	}
	if id == severityOverload {
		return &checker.CallEstimate{CostEstimate: checker.CostEstimate{Min: 1, Max: 1}, ResultSize: &checker.SizeEstimate{Min: 4, Max: 11}}
	}
	return nil
}
