package exploration

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
)

// Cohort is a browser-held capture of complete group/count results, not a signed
// source receipt. Comparison validates its internal accounting, not provenance.
type Cohort struct {
	NumericHistogram *NumericHistogram  `json:"numericHistogram,omitempty"`
	Numeric          *NumericSummary    `json:"numeric,omitempty"`
	Duration         *DurationSummary   `json:"duration,omitempty"`
	SpanStatus       *SpanStatusSummary `json:"spanStatus,omitempty"`
	Question         Query              `json:"question"`
	Basis            string             `json:"basis"`
	Count            string             `json:"count"`
	State            string             `json:"state"`
	Groups           []ValueCount       `json:"groups"`
}
type ValueCount struct {
	Numeric    *NumericSummary    `json:"numeric,omitempty"`
	Duration   *DurationSummary   `json:"duration,omitempty"`
	SpanStatus *SpanStatusSummary `json:"spanStatus,omitempty"`
	Value      *string            `json:"g0"`
	Count      string             `json:"count"`
}
type ComparisonRequest struct {
	Distribution bool   `json:"distribution,omitempty"`
	Numeric      bool   `json:"numeric,omitempty"`
	Baseline     Cohort `json:"baseline"`
	Selection    Cohort `json:"selection"`
}
type ValueComparison struct {
	Measures         *MeasureComparison `json:"measures,omitempty"`
	Value            *string            `json:"value"`
	BaselineCount    string             `json:"baselineCount"`
	SelectionCount   string             `json:"selectionCount"`
	BaselinePercent  string             `json:"baselinePercent"`
	SelectionPercent string             `json:"selectionPercent"`
	DifferencePP     string             `json:"differencePP"`
}
type CohortComparison struct {
	AggregateOnly  bool               `json:"aggregateOnly,omitempty"`
	Measures       *MeasureComparison `json:"measures,omitempty"`
	State          string             `json:"state"`
	Reason         string             `json:"reason,omitempty"`
	Field          Field              `json:"field"`
	BaselineCount  string             `json:"baselineCount"`
	SelectionCount string             `json:"selectionCount"`
	BaselineBasis  string             `json:"baselineBasis"`
	SelectionBasis string             `json:"selectionBasis"`
	Basis          string             `json:"basis"`
	Rows           []ValueComparison  `json:"rows"`
	Warnings       []string           `json:"warnings"`
}

func decimal(value string) (*big.Int, error) {
	n, ok := new(big.Int).SetString(value, 10)
	if !ok || n.Sign() < 0 || n.BitLen() > 64 || n.String() != value {
		return nil, fmt.Errorf("require canonical uint64 count")
	}
	return n, nil
}
func valueKey(value *string) string {
	if value == nil {
		return "0"
	}
	return "1" + *value
}
func (c Cohort) validate() (map[string]ValueCount, *big.Int, error) {
	if e := c.Question.Validate(); e != nil {
		return nil, nil, e
	}
	if len(c.Question.GroupBy) != 1 {
		return nil, nil, fmt.Errorf("comparison requires one distribution field")
	}
	if c.Basis == "" || len(c.Basis) > 128 {
		return nil, nil, fmt.Errorf("captured source basis required")
	}
	if c.State != "observed" && c.State != "no_data" && c.State != "changed_between_reads" {
		return nil, nil, fmt.Errorf("unsupported capture state")
	}
	if len(c.Groups) > 1000 {
		return nil, nil, fmt.Errorf("distribution exceeds 1000 groups")
	}
	total, e := decimal(c.Count)
	if e != nil {
		return nil, nil, e
	}
	m := map[string]ValueCount{}
	sum := new(big.Int)
	for _, group := range c.Groups {
		if e := validateGroupDimension(c.Question.GroupBy[0], group.Value); e != nil {
			return nil, nil, e
		}
		if group.Value != nil && len(*group.Value) > 4096 {
			return nil, nil, fmt.Errorf("distribution value exceeds 4096 bytes")
		}
		key := valueKey(group.Value)
		if _, exists := m[key]; exists {
			return nil, nil, fmt.Errorf("duplicate distribution value")
		}
		n, e := decimal(group.Count)
		if e != nil || n.Sign() == 0 {
			return nil, nil, fmt.Errorf("group counts must be positive uint64")
		}
		sum.Add(sum, n)
		m[key] = group
	}
	if c.State != "changed_between_reads" && (sum.Cmp(total) != 0 || c.State == "no_data" && total.Sign() != 0 || c.State == "observed" && total.Sign() == 0) {
		return nil, nil, fmt.Errorf("distribution counts disagree with population")
	}
	if e := c.validateMeasures(); e != nil {
		return nil, nil, e
	}
	return m, total, nil
}
func CompareCohorts(request ComparisonRequest) (*CohortComparison, error) {
	if request.Distribution && !request.Numeric {
		return nil, fmt.Errorf("distribution comparison requires explicit numeric mode")
	}
	if !request.Distribution && (request.Baseline.NumericHistogram != nil || request.Selection.NumericHistogram != nil) {
		return nil, fmt.Errorf("histogram payload requires explicit distribution comparison")
	}
	if request.Distribution && (request.Baseline.NumericHistogram == nil || request.Selection.NumericHistogram == nil) {
		return nil, fmt.Errorf("distribution comparison requires complete histograms on both sides")
	}
	if !request.Numeric && (numericPayload(request.Baseline) || numericPayload(request.Selection)) {
		return nil, fmt.Errorf("numeric payload requires explicit numeric comparison")
	}
	if request.Numeric && (!compatibleNumericBindings(request.Baseline.Question.Numeric, request.Selection.Question.Numeric) || request.Baseline.Numeric == nil || request.Selection.Numeric == nil) {
		return nil, fmt.Errorf("numeric comparison requires identical declared bindings and complete numeric coverage")
	}
	if request.Distribution && len(request.Baseline.Question.GroupBy) == 0 && len(request.Selection.Question.GroupBy) == 0 {
		return compareAggregateDistributions(request)
	}
	a, totalA, e := request.Baseline.validate()
	if e != nil {
		return nil, e
	}
	b, totalB, e := request.Selection.validate()
	if e != nil {
		return nil, e
	}
	qa, qb := request.Baseline.Question, request.Selection.Question
	if qa.Signal != qb.Signal || qa.Environment != qb.Environment || qa.GroupBy[0] != qb.GroupBy[0] {
		return nil, fmt.Errorf("comparison requires identical signal, environment and distribution field")
	}
	result := &CohortComparison{State: "ok", Field: qa.GroupBy[0], BaselineCount: request.Baseline.Count, SelectionCount: request.Selection.Count, BaselineBasis: request.Baseline.Basis, SelectionBasis: request.Selection.Basis, Rows: []ValueComparison{}, Warnings: []string{"Captures were read separately; late data, sampling and source changes may affect differences.", "Populations may overlap. Percentage-point differences are descriptive leads, not causal attribution or statistical significance.", "These are captured record counts, not unique requests or business outcomes. Source bases are caller-supplied metadata, not authenticated provenance."}}
	if request.Baseline.Duration != nil && request.Selection.Duration != nil || request.Baseline.SpanStatus != nil && request.Selection.SpanStatus != nil {
		result.Measures = compareMeasures(request.Baseline.Duration, request.Selection.Duration, request.Baseline.SpanStatus, request.Selection.SpanStatus)
	}
	if request.Baseline.State == "changed_between_reads" || request.Selection.State == "changed_between_reads" {
		result.State = "unknown"
		result.Reason = "source-reads-disagree"
	} else if totalA.Sign() == 0 || totalB.Sign() == 0 {
		result.State = "no_data"
		result.Reason = "empty-population"
	} else {
		keys := []string{}
		for key := range a {
			keys = append(keys, key)
		}
		for key := range b {
			if _, exists := a[key]; !exists {
				keys = append(keys, key)
			}
		}
		type ranked struct {
			row   ValueComparison
			score *big.Rat
			key   string
		}
		rows := []ranked{}
		for _, key := range keys {
			av, aok := a[key]
			bv, bok := b[key]
			value := av.Value
			if !aok {
				value = bv.Value
			}
			ac, bc := new(big.Int), new(big.Int)
			if aok {
				ac, _ = decimal(av.Count)
			}
			if bok {
				bc, _ = decimal(bv.Count)
			}
			ap := new(big.Rat).Mul(new(big.Rat).SetFrac(ac, totalA), big.NewRat(100, 1))
			bp := new(big.Rat).Mul(new(big.Rat).SetFrac(bc, totalB), big.NewRat(100, 1))
			difference := new(big.Rat).Sub(bp, ap)
			displayDifference := difference.FloatString(2)
			if displayDifference == "-0.00" {
				displayDifference = "0.00"
			}
			rows = append(rows, ranked{ValueComparison{Value: value, BaselineCount: ac.String(), SelectionCount: bc.String(), BaselinePercent: ap.FloatString(2), SelectionPercent: bp.FloatString(2), DifferencePP: displayDifference, Measures: compareMeasures(av.Duration, bv.Duration, av.SpanStatus, bv.SpanStatus)}, new(big.Rat).Abs(difference), key})
		}
		sort.Slice(rows, func(i, j int) bool {
			if order := rows[i].score.Cmp(rows[j].score); order != 0 {
				return order > 0
			}
			return rows[i].key < rows[j].key
		})
		for _, row := range rows {
			if request.Numeric {
				av := a[row.key]
				bv := b[row.key]
				row.row.Measures = withNumericMeasures(row.row.Measures, compareNumericMeasures(*qa.Numeric, av.Numeric, bv.Numeric, false))
			}
			result.Rows = append(result.Rows, row.row)
		}
	}
	if result.State != "ok" {
		result.Measures = nil
	}
	if request.Numeric {
		result.Measures = withNumericMeasures(result.Measures, compareNumericMeasures(*qa.Numeric, request.Baseline.Numeric, request.Selection.Numeric, result.State == "unknown"))
		if request.Distribution {
			result.Measures.Numeric.Distribution = compareNumericDistributions(request.Baseline, request.Selection, result.State == "unknown")
		}
	}
	// Hash the actual caller-held group evidence, not only claimed source bases.
	canonical := request
	canonical.Baseline.Groups = append([]ValueCount(nil), request.Baseline.Groups...)
	canonical.Selection.Groups = append([]ValueCount(nil), request.Selection.Groups...)
	sort.Slice(canonical.Baseline.Groups, func(i, j int) bool {
		return valueKey(canonical.Baseline.Groups[i].Value) < valueKey(canonical.Baseline.Groups[j].Value)
	})
	sort.Slice(canonical.Selection.Groups, func(i, j int) bool {
		return valueKey(canonical.Selection.Groups[i].Value) < valueKey(canonical.Selection.Groups[j].Value)
	})
	raw, e := json.Marshal(canonical)
	if e != nil {
		return nil, e
	}
	h := sha256.Sum256(raw)
	result.Basis = hex.EncodeToString(h[:])
	return result, nil
}
