package exploration

import (
	"fmt"
	"math/big"
	"regexp"
	"unicode"
	"unicode/utf8"
)

// NumericBinding is an explicit interpretation of an exporter string map.
// Unit labels are declarations, never inferred from an OTel attribute name.
type NumericBinding struct {
	Field Field  `json:"field"`
	Type  string `json:"type"`
	Unit  string `json:"unit"`
	Parse string `json:"parse"`
}

type NumericDiagnostic struct {
	Code    string `json:"code"`
	Path    string `json:"path"`
	Message string `json:"message"`
}

type NumericPlan struct {
	Scope           QuestionScope       `json:"scope"`
	Binding         NumericBinding      `json:"binding"`
	Operators       []string            `json:"operators"`
	Executable      bool                `json:"executable"`
	Diagnostics     []NumericDiagnostic `json:"diagnostics"`
	Coverage        []string            `json:"coverage"`
	MeanDenominator string              `json:"meanDenominator"`
}

// PlanNumericBinding is source-free. It deliberately does not add an executable
// expression to Query or existing package/capture schemas.
func PlanNumericBinding(q Query, b NumericBinding) (NumericPlan, error) {
	scope, err := (CaptureQuestion{Kind: "explore", Explore: &q}).Scope()
	if err != nil {
		return NumericPlan{}, err
	}
	p := NumericPlan{Scope: scope, Binding: b, Operators: []string{"sum", "min", "max", "mean"},
		Coverage: []string{"valid", "missing", "empty", "invalid"}, MeanDenominator: "valid"}
	add := func(code, path, message string) {
		p.Diagnostics = append(p.Diagnostics, NumericDiagnostic{code, path, message})
	}
	if _, err := b.Field.Type(q.Signal); err != nil {
		add("unsupported_field", "binding.field", err.Error())
	} else if b.Field.Kind != "attribute" && b.Field.Kind != "resource" {
		add("unsupported_field", "binding.field", "numeric attribute bindings require an attribute or resource map key; use the existing duration summary for duration_ns")
	}
	if b.Type != "int64" && b.Type != "uint64" {
		add("unsupported_type", "binding.type", "supported declared types: int64, uint64; floats and decimals have no parsing contract yet")
	}
	if b.Parse != "canonical-decimal" {
		add("unsupported_parse", "binding.parse", "require canonical-decimal; no whitespace, exponents, fractions, plus sign or leading zeros")
	}
	if !validNumericUnit(b.Unit) {
		add("invalid_unit", "binding.unit", "declare a nonempty unit label of at most 64 UTF-8 bytes without whitespace or control characters; use 1 for dimensionless values")
	}
	add("execution_unavailable", "execution", "reference semantics only; ClickHouse lowering, Explore controls and package/capture persistence are not implemented for numeric attributes")
	return p, nil
}

func validNumericUnit(unit string) bool {
	if len(unit) == 0 || len(unit) > 64 || !utf8.ValidString(unit) {
		return false
	}
	for _, r := range unit {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return false
		}
	}
	return true
}

var canonicalInteger = regexp.MustCompile(`^(0|[1-9][0-9]*|-[1-9][0-9]*)$`)

// NumericSummary preserves exact decimal integers. Mean is the rational Sum/Valid.
// Empty and malformed values never contribute zero or enter the denominator.
type NumericSummary struct {
	State   string `json:"state"`
	Unit    string `json:"unit"`
	Count   string `json:"count"`
	Valid   string `json:"valid"`
	Missing string `json:"missing"`
	Empty   string `json:"empty"`
	Invalid string `json:"invalid"`
	Sum     string `json:"sum"`
	Min     string `json:"min,omitempty"`
	Max     string `json:"max,omitempty"`
}

// EvaluateNumericValues consumes one optional exporter map value per record.
// nil is an absent key; pointer-to-empty is a present empty string. It does not
// establish original OTLP scalar types or OTel convention conformance.
func EvaluateNumericValues(b NumericBinding, values []*string) (NumericSummary, error) {
	if b.Field.Kind != "attribute" && b.Field.Kind != "resource" {
		return NumericSummary{}, fmt.Errorf("require attribute or resource binding")
	}
	if _, err := b.Field.Type("spans"); err != nil {
		return NumericSummary{}, err
	}
	if (b.Type != "int64" && b.Type != "uint64") || b.Parse != "canonical-decimal" || !validNumericUnit(b.Unit) {
		return NumericSummary{}, fmt.Errorf("unsupported numeric binding")
	}
	valid, missing, empty, invalid := int64(0), int64(0), int64(0), int64(0)
	sum := new(big.Int)
	var min, max *big.Int
	for _, raw := range values {
		if raw == nil {
			missing++
			continue
		}
		if *raw == "" {
			empty++
			continue
		}
		if len(*raw) > 20 || !canonicalInteger.MatchString(*raw) {
			invalid++
			continue
		}
		v, ok := new(big.Int).SetString(*raw, 10)
		if !ok || b.Type == "int64" && !v.IsInt64() || b.Type == "uint64" && !v.IsUint64() {
			invalid++
			continue
		}
		valid++
		sum.Add(sum, v)
		if min == nil || v.Cmp(min) < 0 {
			min = new(big.Int).Set(v)
		}
		if max == nil || v.Cmp(max) > 0 {
			max = new(big.Int).Set(v)
		}
	}
	s := NumericSummary{State: "no_valid_values", Unit: b.Unit, Count: fmt.Sprint(len(values)), Valid: fmt.Sprint(valid), Missing: fmt.Sprint(missing), Empty: fmt.Sprint(empty), Invalid: fmt.Sprint(invalid), Sum: sum.String()}
	if len(values) == 0 {
		s.State = "no_data"
	}
	if valid > 0 {
		s.State = "observed"
		s.Min = min.String()
		s.Max = max.String()
	}
	return s, nil
}
