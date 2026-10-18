package exploration

import (
	"encoding/json"
	"fmt"
	"math/big"
)

type MeasureComparison struct {
	BaselineDuration  *DurationSummary   `json:"baselineDuration,omitempty"`
	SelectionDuration *DurationSummary   `json:"selectionDuration,omitempty"`
	BaselineStatus    *SpanStatusSummary `json:"baselineStatus,omitempty"`
	SelectionStatus   *SpanStatusSummary `json:"selectionStatus,omitempty"`
	MeanDifferenceNS  string             `json:"meanDifferenceNs,omitempty"`
	P95DifferenceNS   string             `json:"p95DifferenceNs,omitempty"`
	P99DifferenceNS   string             `json:"p99DifferenceNs,omitempty"`
	ErrorDifferencePP string             `json:"errorDifferencePP,omitempty"`
}

func signedDifference(a, b string) string {
	x, _ := new(big.Int).SetString(a, 10)
	y, _ := new(big.Int).SetString(b, 10)
	return y.Sub(y, x).String()
}
func cleanDecimal(r *big.Rat, digits int) string {
	s := r.FloatString(digits)
	if s == "-0.00" || s == "-0.000" {
		return s[1:]
	}
	return s
}
func compareMeasures(a, b *DurationSummary, x, y *SpanStatusSummary) *MeasureComparison {
	if a == nil && b == nil && x == nil && y == nil {
		return nil
	}
	r := &MeasureComparison{BaselineDuration: a, SelectionDuration: b, BaselineStatus: x, SelectionStatus: y}
	if a != nil && b != nil && a.State == "observed" && b.State == "observed" {
		an, _ := durationDecimal(a.SumNS, 128)
		ad, _ := decimal(a.Count)
		bn, _ := durationDecimal(b.SumNS, 128)
		bd, _ := decimal(b.Count)
		r.MeanDifferenceNS = cleanDecimal(new(big.Rat).Sub(new(big.Rat).SetFrac(bn, bd), new(big.Rat).SetFrac(an, ad)), 3)
		r.P95DifferenceNS = signedDifference(a.P95NS, b.P95NS)
		r.P99DifferenceNS = signedDifference(a.P99NS, b.P99NS)
	}
	if x != nil && y != nil && x.Count != "0" && y.Count != "0" {
		an, _ := decimal(x.Error)
		ad, _ := decimal(x.Count)
		bn, _ := decimal(y.Error)
		bd, _ := decimal(y.Count)
		r.ErrorDifferencePP = cleanDecimal(new(big.Rat).Mul(new(big.Rat).Sub(new(big.Rat).SetFrac(bn, bd), new(big.Rat).SetFrac(an, ad)), big.NewRat(100, 1)), 2)
	}
	return r
}
func (c Cohort) validateMeasures() error {
	if c.Duration != nil && (c.Question.Summary != "span-duration" || c.Duration.Count != c.Count || c.Duration.Validate() != nil) {
		return fmt.Errorf("invalid cohort duration")
	}
	if c.SpanStatus != nil && (!c.Question.SpanStatus || c.SpanStatus.Count != c.Count || c.SpanStatus.Validate() != nil) {
		return fmt.Errorf("invalid cohort span status")
	}
	status := emptySpanStatus()
	sum := new(big.Int)
	var lo, hi *big.Int
	for _, g := range c.Groups {
		if (g.Duration == nil) != (c.Duration == nil) || (g.SpanStatus == nil) != (c.SpanStatus == nil) {
			return fmt.Errorf("incomplete group measures")
		}
		if g.Duration != nil {
			d := g.Duration
			if d.Count != g.Count || d.State != "observed" || d.Validate() != nil {
				return fmt.Errorf("invalid group duration")
			}
			v, _ := durationDecimal(d.SumNS, 128)
			sum.Add(sum, v)
			l, _ := decimal(d.MinNS)
			h, _ := decimal(d.MaxNS)
			if lo == nil || l.Cmp(lo) < 0 {
				lo = l
			}
			if hi == nil || h.Cmp(hi) > 0 {
				hi = h
			}
		}
		if g.SpanStatus != nil {
			if g.SpanStatus.Count != g.Count || g.SpanStatus.Validate() != nil {
				return fmt.Errorf("invalid group span status")
			}
			status.add(*g.SpanStatus)
		}
	}
	if c.State != "changed_between_reads" {
		if c.Duration != nil {
			if sum.String() != c.Duration.SumNS || lo != nil && (lo.String() != c.Duration.MinNS || hi.String() != c.Duration.MaxNS) {
				return fmt.Errorf("group durations disagree with population")
			}
		}
		if c.SpanStatus != nil && status != *c.SpanStatus {
			return fmt.Errorf("group status disagrees with population")
		}
	}
	return nil
}

// CohortFromResult retains full aggregates, never computes measures from rows.
func CohortFromResult(r *Result) (Cohort, error) {
	c := Cohort{Question: r.Question, Basis: r.Basis, Count: r.Count, State: r.State, Duration: r.Duration, SpanStatus: r.SpanStatus, Groups: []ValueCount{}}
	for _, raw := range r.Groups {
		var g ValueCount
		if json.Unmarshal(raw, &g) != nil {
			return c, fmt.Errorf("invalid group")
		}
		c.Groups = append(c.Groups, g)
	}
	_, _, e := c.validate()
	return c, e
}
