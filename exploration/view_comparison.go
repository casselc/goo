package exploration

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"reflect"
	"sort"
	"time"
)

// ViewComparison retains independently scoped supplied snapshots. It is not a
// merged dataset, signed receipt, statistical effect estimate or pooled quantile.
type ViewComparison struct {
	Version   string                `json:"version"`
	Baseline  *ViewDataset          `json:"baseline"`
	Selection *ViewDataset          `json:"selection"`
	State     string                `json:"state"`
	Reason    string                `json:"reason,omitempty"`
	Changes   map[string]ViewChange `json:"changes"`
}
type ViewChange struct {
	State       string           `json:"state"`
	Value       *string          `json:"value"`
	Unit        string           `json:"unit"`
	Reason      string           `json:"reason,omitempty"`
	Aggregation *ViewAggregation `json:"aggregation,omitempty"`
}

// summaryViewDataset consumes only the whole-population count aggregate and its
// validated measures. Its attestation never comes from a count/page comparison.
func summaryViewDataset(r *Result, exhaustiveSummary bool) *ViewDataset {
	if !supportedViewMeasures(r.Question) {
		return nil
	}
	if _, err := decimal(r.Count); err != nil {
		return nil
	}
	if r.State != "observed" && r.State != "no_data" && r.State != "changed_between_reads" {
		return nil
	}
	if r.State == "observed" && r.Count == "0" || r.State == "no_data" && r.Count != "0" {
		return nil
	}
	if exhaustiveSummary && (r.CapturedAt == "" || !validCohortClock(Cohort{CapturedAt: r.CapturedAt})) {
		return nil
	}
	d := baseViewDataset(r, "summary", "summary")
	if d == nil {
		return nil
	}
	d.Fields = []ViewField{{Key: "count", Label: "Captured records", Type: "number", Unit: "records"}}
	d.Fields = append(d.Fields, viewMeasureFields(r.Question)...)
	values := ViewValues{"count": viewText(r.Count)}
	cells, ok := viewMeasureCells(r.Question, r.Count, r.Numeric, r.Duration, false, values)
	if !ok {
		return nil
	}
	d.Rows = []ViewRow{{Key: "population", Values: values, Cells: cells}}
	if exhaustiveSummary && r.State != "changed_between_reads" {
		attestView(d, "summary")
	}
	return d
}

var errCohortClock = errors.New("cohort capturedAt must be a valid RFC3339 timestamp when supplied")

func validCohortClock(c Cohort) bool {
	if c.CapturedAt == "" {
		return true
	}
	_, err := time.Parse(time.RFC3339Nano, c.CapturedAt)
	return err == nil
}

// CompareCohorts validates existing comparison families before deriving public
// display snapshots. Browser-held claims can never produce source attestations.
func CompareCohorts(request ComparisonRequest) (*CohortComparison, error) {
	if !validCohortClock(request.Baseline) || !validCohortClock(request.Selection) {
		return nil, errCohortClock
	}
	result, err := compareCohorts(request)
	if err != nil {
		return nil, err
	}
	result.ViewComparison = cohortViewComparison(request)
	return result, nil
}

func heldSummaryView(c Cohort) *ViewDataset {
	if !supportedViewMeasures(c.Question) || !validCohortClock(c) {
		return nil
	}
	// Hash the actual immutable held content rather than accepting claimed basis
	// as a version. A separate sourceBasis retains that unverified caller metadata.
	canonical := detached(c)
	sort.Slice(canonical.Groups, func(i, j int) bool { return valueKey(canonical.Groups[i].Value) < valueKey(canonical.Groups[j].Value) })
	raw, err := json.Marshal(canonical)
	if err != nil {
		return nil
	}
	h := sha256.Sum256(raw)
	r := &Result{Question: canonical.Question, Basis: hex.EncodeToString(h[:]), CapturedAt: canonical.CapturedAt, Count: canonical.Count, State: canonical.State, Numeric: canonical.Numeric, Duration: canonical.Duration}
	d := summaryViewDataset(r, false)
	if d == nil {
		return nil
	}
	d.Origin = "browser-held"
	d.SourceBasis = canonical.Basis
	// Origin is descriptive. The trusted constructor, not this string, withholds
	// completeness regardless of what a caller claims about its count or basis.
	d.Quality.Extent = "unknown"
	d.Completeness = nil
	return d
}

func cohortViewComparison(request ComparisonRequest) *ViewComparison {
	a, b := heldSummaryView(request.Baseline), heldSummaryView(request.Selection)
	if a == nil || b == nil {
		return nil
	}
	result := &ViewComparison{Version: "goo.comparison/1", Baseline: a, Selection: b, State: "observed", Changes: map[string]ViewChange{}}
	fields := map[string]ViewField{}
	af, bf := map[string]ViewField{}, map[string]ViewField{}
	for _, f := range a.Fields {
		fields[f.Key] = f
		af[f.Key] = f
	}
	for _, f := range b.Fields {
		fields[f.Key] = f
		bf[f.Key] = f
	}
	compatible := request.Baseline.Question.Signal == request.Selection.Question.Signal && request.Baseline.Question.SameEnvironment(request.Selection.Question)
	for key := range fields {
		other, exists := bf[key]
		baseline, hasBaseline := af[key]
		if !exists || !hasBaseline || baseline.Type != other.Type || baseline.Unit != other.Unit || !reflect.DeepEqual(baseline.Aggregation, other.Aggregation) {
			compatible = false
		}
	}
	if request.Baseline.Question.Numeric != nil || request.Selection.Question.Numeric != nil {
		if !compatibleNumericBindings(request.Baseline.Question.Numeric, request.Selection.Question.Numeric) {
			compatible = false
		}
	}
	switch {
	case !compatible:
		result.State = "incompatible"
		result.Reason = "incompatible-fields"
	case request.Baseline.State == "changed_between_reads" || request.Selection.State == "changed_between_reads":
		result.State = "diagnostic"
		result.Reason = "selection-source-reads-disagree"
		if request.Baseline.State == "changed_between_reads" {
			result.Reason = "baseline-source-reads-disagree"
			if request.Selection.State == "changed_between_reads" {
				result.Reason = "both-source-reads-disagree"
			}
		}
	case request.Baseline.Count == "0" || request.Selection.Count == "0":
		result.State = "no_data"
	}
	ar, br := a.Rows[0], b.Rows[0]
	for key, f := range fields {
		change := ViewChange{State: "unavailable", Unit: f.Unit, Aggregation: detached(f.Aggregation)}
		switch result.State {
		case "incompatible":
			change.Reason = "incompatible-fields"
		case "diagnostic":
			change.Reason = "source-reads-disagree"
		default:
			av, bv := ar.Values[key], br.Values[key]
			exact := key == "count" || ar.Cells[key].State == "derived" && br.Cells[key].State == "derived"
			if av == nil || bv == nil || !exact {
				change.Reason = "unavailable-input"
				if result.State == "no_data" {
					change.Reason = "empty-population"
				} else if ar.Cells[key].State == "no_valid_values" || br.Cells[key].State == "no_valid_values" {
					change.Reason = "no-valid-values"
				}
			} else {
				x, okA := new(big.Rat).SetString(*av)
				y, okB := new(big.Rat).SetString(*bv)
				if !okA || !okB {
					return nil
				}
				change.State = "derived"
				if key == "count" {
					change.Reason = "absolute-count-difference-not-normalized-rate"
				}
				change.Value = viewText(new(big.Rat).Sub(y, x).RatString())
			}
		}
		result.Changes[key] = change
	}
	return result
}
