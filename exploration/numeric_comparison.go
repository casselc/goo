package exploration

import (
	"fmt"
	"math/big"
	"reflect"
)

// Numeric changes are exact decimal integers or reduced rational strings. Mean
// uses valid values as its denominator; coverage percentages use all records.
type NumericCoverageChange struct {
	Kind           string `json:"kind"`
	BaselineCount  string `json:"baselineCount"`
	SelectionCount string `json:"selectionCount"`
	Difference     string `json:"difference,omitempty"`
	DifferencePP   string `json:"differencePP,omitempty"`
}
type NumericMeasureComparison struct {
	Distribution   *NumericDistributionComparison `json:"distribution,omitempty"`
	Binding        NumericBinding                 `json:"binding"`
	State          string                         `json:"state"`
	Reason         string                         `json:"reason,omitempty"`
	Baseline       *NumericSummary                `json:"baseline,omitempty"`
	Selection      *NumericSummary                `json:"selection,omitempty"`
	SumDifference  string                         `json:"sumDifference,omitempty"`
	MinDifference  string                         `json:"minDifference,omitempty"`
	MaxDifference  string                         `json:"maxDifference,omitempty"`
	MeanDifference string                         `json:"meanDifference,omitempty"`
	Coverage       []NumericCoverageChange        `json:"coverage"`
}

func compatibleNumericBindings(a, b *NumericBinding) bool { return a != nil && b != nil && *a == *b }
func numericPayload(c Cohort) bool {
	if c.Numeric != nil || c.NumericHistogram != nil {
		return true
	}
	for _, g := range c.Groups {
		if g.Numeric != nil {
			return true
		}
	}
	return false
}
func validateNumericCohort(c Cohort) error {
	if c.NumericHistogram != nil {
		if e := validateOptionalNumericHistogram(c.Question, c.NumericHistogram, c.Numeric); e != nil {
			return e
		}
	}
	if c.Numeric == nil {
		for _, g := range c.Groups {
			if g.Numeric != nil {
				return fmt.Errorf("numeric group requires population numeric coverage")
			}
		}
		return nil
	}
	if c.Question.Numeric == nil {
		return fmt.Errorf("numeric comparison payload requires declared binding")
	}
	b := *c.Question.Numeric
	if e := validateOptionalNumeric(&b, c.Numeric, c.Count); e != nil {
		return e
	}
	parts := make([]*NumericSummary, 0, len(c.Groups))
	for _, g := range c.Groups {
		if e := validateOptionalNumeric(&b, g.Numeric, g.Count); e != nil {
			return e
		}
		parts = append(parts, g.Numeric)
	}
	if c.State != "changed_between_reads" && !reflect.DeepEqual(c.Numeric, mergeNumeric(b, parts)) {
		return fmt.Errorf("numeric group aggregates disagree with population")
	}
	return nil
}

func compareNumericMeasures(b NumericBinding, a, z *NumericSummary, unknown bool) *NumericMeasureComparison {
	r := &NumericMeasureComparison{Binding: b, State: "comparable", Baseline: detached(a), Selection: detached(z), Coverage: []NumericCoverageChange{}}
	counts := func(s *NumericSummary) []string {
		if s == nil {
			return []string{"0", "0", "0", "0"}
		}
		return []string{s.Valid, s.Missing, s.Empty, s.Invalid}
	}
	ac, bc := counts(a), counts(z)
	for i, kind := range []string{"valid", "missing", "empty", "invalid"} {
		row := NumericCoverageChange{Kind: kind, BaselineCount: ac[i], SelectionCount: bc[i]}
		if !unknown {
			row.Difference = signedDifference(ac[i], bc[i])
			if a != nil && z != nil && a.Count != "0" && z.Count != "0" {
				an, _ := decimal(ac[i])
				bn, _ := decimal(bc[i])
				ad, _ := decimal(a.Count)
				bd, _ := decimal(z.Count)
				row.DifferencePP = new(big.Rat).Mul(new(big.Rat).Sub(new(big.Rat).SetFrac(bn, bd), new(big.Rat).SetFrac(an, ad)), big.NewRat(100, 1)).RatString()
			}
		}
		r.Coverage = append(r.Coverage, row)
	}
	if unknown {
		r.State = "unknown"
		r.Reason = "source-reads-disagree"
		return r
	}
	if a == nil || z == nil {
		r.State = "unavailable"
		r.Reason = "absent-group"
		return r
	}
	if a.Count == "0" || z.Count == "0" {
		r.State = "unavailable"
		r.Reason = "empty-population"
		return r
	}
	if a.Valid == "0" || z.Valid == "0" {
		r.State = "unavailable"
		r.Reason = "no-valid-values"
		return r
	}
	r.SumDifference = signedDifference(a.Sum, z.Sum)
	r.MinDifference = signedDifference(a.Min, z.Min)
	r.MaxDifference = signedDifference(a.Max, z.Max)
	an, _ := new(big.Int).SetString(a.Sum, 10)
	bn, _ := new(big.Int).SetString(z.Sum, 10)
	ad, _ := decimal(a.Valid)
	bd, _ := decimal(z.Valid)
	r.MeanDifference = new(big.Rat).Sub(new(big.Rat).SetFrac(bn, bd), new(big.Rat).SetFrac(an, ad)).RatString()
	return r
}
func withNumericMeasures(m *MeasureComparison, n *NumericMeasureComparison) *MeasureComparison {
	if m == nil {
		m = &MeasureComparison{}
	}
	m.Numeric = n
	return m
}
