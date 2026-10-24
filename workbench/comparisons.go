package workbench

import (
	"errors"
	"reflect"
	"time"

	"github.com/casselc/goo/exploration"
)

const ComparisonAlgorithm = "goo/evidence-comparison/v3"
const NumericComparisonAlgorithm = "goo/evidence-comparison/v4"
const DistributionComparisonAlgorithm = "goo/evidence-comparison/v5"

var ErrComparisonChanged = errors.New("comparison changed; compare these captures again before saving")

// ComparisonReceipt freezes the server-verified decision and displayed result.
// Findings remain authored document content, separately revisioned from evidence.
type ComparisonReceipt struct {
	ID         string                          `json:"id"`
	Version    int                             `json:"version"`
	Algorithm  string                          `json:"algorithm"`
	RecordedAt string                          `json:"recordedAt"`
	Result     *exploration.EvidenceComparison `json:"result"`
}

func (s Store) CreateComparison(baseline, selection string, expected *exploration.EvidenceComparison, load func(string) (*exploration.SavedCapture, error)) (*ComparisonReceipt, error) {
	if !ID.MatchString(baseline) || !ID.MatchString(selection) || baseline == selection || expected == nil || load == nil {
		return nil, errors.New("two distinct exact capture IDs and the displayed comparison are required")
	}
	a, e := load(baseline)
	if e != nil {
		return nil, e
	}
	b, e := load(selection)
	if e != nil {
		return nil, e
	}
	var result *exploration.EvidenceComparison
	if expected.Distribution {
		result, e = exploration.CompareDistributionEvidence(a, b)
	} else {
		result, e = exploration.CompareEvidence(a, b)
	}
	if e != nil {
		return nil, e
	}
	if !reflect.DeepEqual(result, expected) {
		return nil, ErrComparisonChanged
	}
	algorithm := ComparisonAlgorithm
	if a.Question.Explore != nil && a.Question.Explore.Numeric != nil || b.Question.Explore != nil && b.Question.Explore.Numeric != nil {
		algorithm = NumericComparisonAlgorithm
	}
	if expected.Distribution {
		algorithm = DistributionComparisonAlgorithm
	}
	receipt := ComparisonReceipt{Version: 1, Algorithm: algorithm, RecordedAt: time.Now().UTC().Format(time.RFC3339Nano), Result: result}
	id, e := digest(receipt)
	if e != nil {
		return nil, e
	}
	receipt.ID = id
	if e = s.publish("comparisons", id, receipt); e != nil {
		return nil, e
	}
	return s.LoadComparison(id)
}
func (s Store) LoadComparison(id string) (*ComparisonReceipt, error) {
	var r ComparisonReceipt
	if e := s.read("comparisons", id, &r); e != nil {
		return nil, e
	}
	actual := r.ID
	r.ID = ""
	hash, e := digest(r)
	r.ID = actual
	if e != nil || actual != id || hash != id || r.Version != 1 || (r.Algorithm != ComparisonAlgorithm && r.Algorithm != NumericComparisonAlgorithm && r.Algorithm != DistributionComparisonAlgorithm && r.Algorithm != "goo/evidence-comparison/v1" && r.Algorithm != "goo/evidence-comparison/v2") || !validTime(r.RecordedAt) || r.Result == nil {
		return nil, errors.New("comparison integrity check failed")
	}
	if (r.Algorithm == DistributionComparisonAlgorithm) != r.Result.Distribution {
		return nil, errors.New("distribution intent differs from stored comparison algorithm")
	}
	if r.Algorithm != DistributionComparisonAlgorithm && r.Result.Measures != nil && r.Result.Measures.Numeric != nil && r.Result.Measures.Numeric.Distribution != nil {
		return nil, errors.New("distribution evidence requires distribution comparison algorithm")
	}
	a, b := r.Result.Baseline.Receipt, r.Result.Selection.Receipt
	if !ID.MatchString(a.ID) || !ID.MatchString(b.ID) || a.ID == b.ID || a.Question.Validate() != nil || b.Question.Validate() != nil {
		return nil, errors.New("invalid stored comparison references")
	}
	switch r.Result.State {
	case "comparable", "incompatible", "unavailable", "unsupported":
	case "unknown":
		if r.Algorithm != NumericComparisonAlgorithm && r.Algorithm != DistributionComparisonAlgorithm {
			return nil, errors.New("unknown numeric decision requires numeric comparison algorithm")
		}
	default:
		return nil, errors.New("invalid stored comparison decision")
	}
	// Deliberately do not run a newer comparison engine when reopening this receipt.
	return &r, nil
}
