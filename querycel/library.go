// Package querycel translates a bounded, checked CEL predicate into Goo's
// ordinary record grammar. It never reads telemetry or executes source SQL.
package querycel

import (
	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/checker"
	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"

	"github.com/casselc/goo/exploration"
)

const (
	LibraryVersion            = "goo.cel.otel/v1"
	MaxExpressionBytes        = 4096
	MaxFieldBytes             = 65536
	MaxCost            uint64 = 1000000
	severityOverload          = "goo_otel_severity_band_uint_v1"
)

// OTelLibrary declares the portable helper and its pure, constant-work runtime
// implementation. The query compiler separately admits this resolved overload.
type OTelLibrary struct{}

func (OTelLibrary) LibraryName() string { return LibraryVersion }
func (OTelLibrary) CompileOptions() []cel.EnvOption {
	return []cel.EnvOption{cel.Function("otel.severityBand", cel.Overload(severityOverload,
		[]*cel.Type{cel.UintType}, cel.StringType,
		cel.UnaryBinding(func(value ref.Val) ref.Val {
			n, ok := value.(types.Uint)
			if !ok {
				return types.NewErr("severityBand requires uint")
			}
			return types.String(exploration.LogSeverityClass(uint64(n)))
		})))}
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
	if id == severityOverload {
		return &checker.CallEstimate{CostEstimate: checker.CostEstimate{Min: 1, Max: 1}, ResultSize: &checker.SizeEstimate{Min: 4, Max: 11}}
	}
	return nil
}
