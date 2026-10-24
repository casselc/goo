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
	Scale int    `json:"scale,omitempty" yaml:"scale,omitempty"`
	Field Field  `json:"field" yaml:"field"`
	Type  string `json:"type" yaml:"type"`
	Unit  string `json:"unit" yaml:"unit"`
	Parse string `json:"parse" yaml:"parse"`
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

// PlanNumericBinding reports source-free capabilities for the explicit binding.
// Executable means supported by this adapter, not that any data was observed.
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
	if b.Type != "int64" && b.Type != "uint64" && b.Type != "decimal64" {
		add("unsupported_type", "binding.type", "supported declared types: int64, uint64, decimal64; floats and arbitrary precision decimals are unsupported")
	}
	if b.Type == "decimal64" && (b.Scale < 1 || b.Scale > 9) || b.Type != "decimal64" && b.Scale != 0 {
		add("invalid_scale", "binding.scale", "decimal64 requires scale 1..9; integer bindings require scale 0")
	}
	if b.Parse != "canonical-decimal" {
		message := "require canonical-decimal integers; no whitespace, exponents, fractions, plus sign or leading zeros"
		if b.Type == "decimal64" {
			message = "require canonical-decimal with exactly the declared scale of fractional digits; no whitespace, exponent, plus sign, leading zeros, negative zero or rounding"
		}
		add("unsupported_parse", "binding.parse", message)
	}
	if !validNumericUnit(b.Unit) {
		add("invalid_unit", "binding.unit", "declare a nonempty unit label of at most 64 UTF-8 bytes without whitespace or control characters; use 1 for dimensionless values")
	}
	p.Executable = len(p.Diagnostics) == 0
	if p.Diagnostics == nil {
		p.Diagnostics = []NumericDiagnostic{}
	}
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

// parseCanonicalInteger is shared by numeric coverage and typed predicates.
// Exporter strings and thresholds obey the same exact, bounded parsing contract.
func parseCanonicalInteger(raw, typ string) (*big.Int, bool) {
	if len(raw) > 20 || !canonicalInteger.MatchString(raw) {
		return nil, false
	}
	v, ok := new(big.Int).SetString(raw, 10)
	if !ok || typ != "int64" && typ != "uint64" || typ == "int64" && !v.IsInt64() || typ == "uint64" && !v.IsUint64() {
		return nil, false
	}
	return v, true
}

// NumericSummary preserves exact integer coefficients. Decimal64 human values
// divide coefficients by 10^Scale; the human mean is Sum/(Valid*10^Scale).
// Integers have Scale=0. Empty/invalid values never enter the valid denominator.
type NumericSummary struct {
	Scale   int    `json:"scale,omitempty"`
	State   string `json:"state"`
	Unit    string `json:"unit" yaml:"unit"`
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
	if b.Validate("spans") != nil {
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
		v, ok := parseNumericCoefficient(*raw, b)
		if !ok {
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
	s := NumericSummary{Scale: b.Scale, State: "no_valid_values", Unit: b.Unit, Count: fmt.Sprint(len(values)), Valid: fmt.Sprint(valid), Missing: fmt.Sprint(missing), Empty: fmt.Sprint(empty), Invalid: fmt.Sprint(invalid), Sum: sum.String()}
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
