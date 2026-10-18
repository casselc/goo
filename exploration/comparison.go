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
	Question Query        `json:"question"`
	Basis    string       `json:"basis"`
	Count    string       `json:"count"`
	State    string       `json:"state"`
	Groups   []ValueCount `json:"groups"`
}
type ValueCount struct {
	Value *string `json:"g0"`
	Count string  `json:"count"`
}
type ComparisonRequest struct {
	Baseline  Cohort `json:"baseline"`
	Selection Cohort `json:"selection"`
}
type ValueComparison struct {
	Value            *string `json:"value"`
	BaselineCount    string  `json:"baselineCount"`
	SelectionCount   string  `json:"selectionCount"`
	BaselinePercent  string  `json:"baselinePercent"`
	SelectionPercent string  `json:"selectionPercent"`
	DifferencePP     string  `json:"differencePP"`
}
type CohortComparison struct {
	State          string            `json:"state"`
	Reason         string            `json:"reason,omitempty"`
	Field          Field             `json:"field"`
	BaselineCount  string            `json:"baselineCount"`
	SelectionCount string            `json:"selectionCount"`
	BaselineBasis  string            `json:"baselineBasis"`
	SelectionBasis string            `json:"selectionBasis"`
	Basis          string            `json:"basis"`
	Rows           []ValueComparison `json:"rows"`
	Warnings       []string          `json:"warnings"`
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
	return m, total, nil
}
func CompareCohorts(request ComparisonRequest) (*CohortComparison, error) {
	a, totalA, e := request.Baseline.validate()
	if e != nil {
		return nil, e
	}
	b, totalB, e := request.Selection.validate()
	if e != nil {
		return nil, e
	}
	qa, qb := request.Baseline.Question, request.Selection.Question
	if qa.Signal != qb.Signal || qa.Environment != qb.Environment || qa.FromMS != qb.FromMS || qa.ToMS != qb.ToMS || qa.GroupBy[0] != qb.GroupBy[0] {
		return nil, fmt.Errorf("comparison requires identical signal, environment, fixed window and distribution field")
	}
	result := &CohortComparison{State: "ok", Field: qa.GroupBy[0], BaselineCount: request.Baseline.Count, SelectionCount: request.Selection.Count, BaselineBasis: request.Baseline.Basis, SelectionBasis: request.Selection.Basis, Rows: []ValueComparison{}, Warnings: []string{"Captures were read separately; late data, sampling and source changes may affect differences.", "Populations may overlap. Percentage-point differences are descriptive leads, not causal attribution or statistical significance.", "These are captured record counts, not unique requests or business outcomes. Source bases are caller-supplied metadata, not authenticated provenance."}}
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
			rows = append(rows, ranked{ValueComparison{value, ac.String(), bc.String(), ap.FloatString(2), bp.FloatString(2), displayDifference}, new(big.Rat).Abs(difference), key})
		}
		sort.Slice(rows, func(i, j int) bool {
			if order := rows[i].score.Cmp(rows[j].score); order != 0 {
				return order > 0
			}
			return rows[i].key < rows[j].key
		})
		for _, row := range rows {
			result.Rows = append(result.Rows, row.row)
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
