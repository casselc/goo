package exploration

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
)

// Shares use every captured record, including absent/empty/invalid numeric values.
// The fixed aligned bins contain valid parsed values only; no rebinning occurs.
type NumericDistributionBin struct {
	From             string `json:"from"`
	To               string `json:"to"`
	BaselineCount    string `json:"baselineCount"`
	SelectionCount   string `json:"selectionCount"`
	CountDifference  string `json:"countDifference,omitempty"`
	BaselinePercent  string `json:"baselinePercent,omitempty"`
	SelectionPercent string `json:"selectionPercent,omitempty"`
	DifferencePP     string `json:"differencePP,omitempty"`
}
type NumericDistributionComparison struct {
	State          string                   `json:"state"`
	Reason         string                   `json:"reason,omitempty"`
	Binning        string                   `json:"binning"`
	Denominator    string                   `json:"denominator"`
	BaselineCount  string                   `json:"baselineCount"`
	SelectionCount string                   `json:"selectionCount"`
	Bins           []NumericDistributionBin `json:"bins"`
}

func compareNumericDistributions(a, b Cohort, unknown bool) *NumericDistributionComparison {
	r := &NumericDistributionComparison{State: "comparable", Binning: "powers-of-two", Denominator: "population-count", BaselineCount: a.Count, SelectionCount: b.Count, Bins: make([]NumericDistributionBin, len(a.NumericHistogram.Bins))}
	if unknown {
		r.State = "unknown"
		r.Reason = "source-reads-disagree"
	} else if a.Count == "0" || b.Count == "0" {
		r.State = "unavailable"
		r.Reason = "empty-population"
	}
	percent := func(count, total string) string {
		if total == "0" {
			return ""
		}
		n, _ := decimal(count)
		d, _ := decimal(total)
		return new(big.Rat).Mul(new(big.Rat).SetFrac(n, d), big.NewRat(100, 1)).RatString()
	}
	for i, x := range a.NumericHistogram.Bins {
		y := b.NumericHistogram.Bins[i]
		row := NumericDistributionBin{From: x.From, To: x.To, BaselineCount: x.Count, SelectionCount: y.Count}
		if !unknown {
			row.CountDifference = signedDifference(x.Count, y.Count)
			row.BaselinePercent = percent(x.Count, a.Count)
			row.SelectionPercent = percent(y.Count, b.Count)
			if row.BaselinePercent != "" && row.SelectionPercent != "" {
				ap, _ := new(big.Rat).SetString(row.BaselinePercent)
				bp, _ := new(big.Rat).SetString(row.SelectionPercent)
				row.DifferencePP = new(big.Rat).Sub(bp, ap).RatString()
			}
		}
		r.Bins[i] = row
	}
	return r
}

// CohortFromResultDistribution is explicit: the older numeric snapshot operation
// deliberately continues to omit histograms and preserve its hashes and wire form.
func CohortFromResultDistribution(r *Result) (Cohort, error) {
	if r != nil && len(r.Question.GroupBy) == 0 {
		c := Cohort{Question: detached(r.Question), Basis: r.Basis, Count: r.Count, State: r.State, Numeric: detached(r.Numeric), NumericHistogram: detached(r.NumericHistogram), Duration: detached(r.Duration), SpanStatus: detached(r.SpanStatus), Groups: []ValueCount{}}
		if len(r.Groups) != 0 {
			return c, fmt.Errorf("aggregate-only distribution cannot contain categorical groups")
		}
		if e := ValidateNumericResult(r); e != nil {
			return c, e
		}
		return c, validateAggregateDistributionCohort(c)
	}
	c, e := CohortFromResult(r)
	if e != nil {
		return c, e
	}
	if r.NumericHistogram == nil || r.Question.NumericHistogram == nil {
		return c, fmt.Errorf("distribution comparison requires an explicitly captured numeric histogram")
	}
	if e := validateOptionalNumericHistogram(r.Question, r.NumericHistogram, r.Numeric); e != nil {
		return c, e
	}
	c.NumericHistogram = detached(r.NumericHistogram)
	return c, nil
}

// PlanNumericDistributionSelection permits only nonempty retained populations.
// It does not turn an empty/uncertain observation into a new asserted selection.
func PlanNumericDistributionSelection(request ComparisonRequest, side string, fromBin, toBin int) (*NumericHistogramSelection, error) {
	if !request.Distribution {
		return nil, fmt.Errorf("explicit distribution comparison required")
	}
	result, e := CompareCohorts(request)
	if e != nil {
		return nil, e
	}
	if result.State == "unknown" {
		return nil, fmt.Errorf("distribution selection unavailable while source reads disagree")
	}
	var c Cohort
	switch side {
	case "baseline":
		c = request.Baseline
	case "selection":
		c = request.Selection
	default:
		return nil, fmt.Errorf("distribution side must be baseline or selection")
	}
	if c.Count == "0" {
		return nil, fmt.Errorf("distribution side has no captured records")
	}
	if fromBin < 0 || toBin <= fromBin || toBin > len(c.NumericHistogram.Bins) {
		return nil, fmt.Errorf("require nonempty half-open distribution bin indexes")
	}
	count := new(big.Int)
	for _, bin := range c.NumericHistogram.Bins[fromBin:toBin] {
		n, _ := decimal(bin.Count)
		count.Add(count, n)
	}
	if count.Sign() == 0 {
		return nil, fmt.Errorf("distribution range has no captured valid values")
	}
	return PlanNumericHistogramSelection(c.Question, fromBin, toBin)
}

// CompareDistributionEvidence is a separate explicit operation. Existing stored
// numeric comparisons keep the older result shape and never acquire bins on load.
func CompareDistributionEvidence(a, b *SavedCapture) (*EvidenceComparison, error) {
	if a == nil || b == nil || a.Explore == nil || b.Explore == nil || a.Explore.Question.NumericHistogram == nil || b.Explore.Question.NumericHistogram == nil || a.Explore.NumericHistogram == nil || b.Explore.NumericHistogram == nil {
		return nil, fmt.Errorf("distribution comparison requires two numeric histogram record captures")
	}
	r, e := CompareEvidence(a, b)
	if e != nil {
		return nil, e
	}
	r.Distribution = true
	if r.State == "incompatible" || r.State == "unsupported" {
		return r, nil
	}
	ac := Cohort{Question: a.Explore.Question, Count: a.Explore.Count, Numeric: a.Explore.Numeric, NumericHistogram: a.Explore.NumericHistogram}
	bc := Cohort{Question: b.Explore.Question, Count: b.Explore.Count, Numeric: b.Explore.Numeric, NumericHistogram: b.Explore.NumericHistogram}
	if r.Measures == nil || r.Measures.Numeric == nil {
		return nil, fmt.Errorf("numeric distribution measures unavailable")
	}
	r.Measures.Numeric.Distribution = compareNumericDistributions(ac, bc, r.State == "unknown")
	return r, nil
}

// Aggregate-only support is isolated behind the new explicit distribution mode.
// Existing count/numeric cohort validation still requires one complete grouping.
func validateAggregateDistributionCohort(c Cohort) error {
	if e := c.Question.Validate(); e != nil {
		return e
	}
	if len(c.Question.GroupBy) != 0 || len(c.Groups) != 0 {
		return fmt.Errorf("aggregate-only distribution must not contain categorical groups")
	}
	if c.Basis == "" || len(c.Basis) > 128 {
		return fmt.Errorf("captured source basis required")
	}
	n, e := decimal(c.Count)
	if e != nil {
		return e
	}
	if c.State != "observed" && c.State != "no_data" && c.State != "changed_between_reads" {
		return fmt.Errorf("unsupported capture state")
	}
	if c.State == "no_data" && n.Sign() != 0 || c.State == "observed" && n.Sign() == 0 {
		return fmt.Errorf("aggregate-only state disagrees with population count")
	}
	if c.Numeric == nil || c.NumericHistogram == nil {
		return fmt.Errorf("aggregate-only distribution requires numeric coverage and complete histogram")
	}
	if e := validateOptionalNumeric(c.Question.Numeric, c.Numeric, c.Count); e != nil {
		return e
	}
	if e := validateOptionalNumericHistogram(c.Question, c.NumericHistogram, c.Numeric); e != nil {
		return e
	}
	if c.Duration != nil && (c.Question.Summary != "span-duration" || c.Duration.Count != c.Count || c.Duration.Validate() != nil) {
		return fmt.Errorf("invalid aggregate-only duration")
	}
	if c.SpanStatus != nil && (!c.Question.SpanStatus || c.SpanStatus.Count != c.Count || c.SpanStatus.Validate() != nil) {
		return fmt.Errorf("invalid aggregate-only span status")
	}
	return nil
}

func compareAggregateDistributions(request ComparisonRequest) (*CohortComparison, error) {
	a, b := request.Baseline, request.Selection
	if e := validateAggregateDistributionCohort(a); e != nil {
		return nil, e
	}
	if e := validateAggregateDistributionCohort(b); e != nil {
		return nil, e
	}
	if a.Question.Signal != b.Question.Signal || a.Question.Environment != b.Question.Environment {
		return nil, fmt.Errorf("distribution comparison requires identical signal and environment")
	}
	r := &CohortComparison{AggregateOnly: true, State: "ok", BaselineCount: a.Count, SelectionCount: b.Count, BaselineBasis: a.Basis, SelectionBasis: b.Basis, Rows: []ValueComparison{}, Warnings: []string{"Complete captured histogram aggregates, without a categorical breakdown. Bin shares use all records; missing, empty and invalid parse states remain excluded from valid-value bins.", "Separate source reads; populations may overlap. Differences are descriptive, not causal attribution or evidence of complete telemetry."}}
	if a.State == "changed_between_reads" || b.State == "changed_between_reads" {
		r.State = "unknown"
		r.Reason = "source-reads-disagree"
	} else if a.Count == "0" || b.Count == "0" {
		r.State = "no_data"
		r.Reason = "empty-population"
	} else {
		r.Measures = compareMeasures(a.Duration, b.Duration, a.SpanStatus, b.SpanStatus)
	}
	r.Measures = withNumericMeasures(r.Measures, compareNumericMeasures(*a.Question.Numeric, a.Numeric, b.Numeric, r.State == "unknown"))
	r.Measures.Numeric.Distribution = compareNumericDistributions(a, b, r.State == "unknown")
	canonical := request
	canonical.Baseline.Groups = nil
	canonical.Selection.Groups = nil
	wire, e := json.Marshal(canonical)
	if e != nil {
		return nil, e
	}
	hash := sha256.Sum256(wire)
	r.Basis = hex.EncodeToString(hash[:])
	return r, nil
}
