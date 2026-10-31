package exploration

import (
	"math/big"
	"sort"
)

// GroupCountComparison aligns internally accounted supplied groups. It carries
// no source completeness witness or claim of exhaustive upstream telemetry.
type GroupCountComparison struct {
	Version             string          `json:"version"`
	Baseline            *ViewDataset    `json:"baseline"`
	Selection           *ViewDataset    `json:"selection"`
	Field               Field           `json:"field"`
	State               string          `json:"state"`
	Reason              string          `json:"reason,omitempty"`
	BaselineAccounting  string          `json:"baselineAccounting"`
	SelectionAccounting string          `json:"selectionAccounting"`
	Rows                []GroupCountRow `json:"rows"`
}

type GroupCountRow struct {
	Key       string         `json:"key"`
	Value     *string        `json:"value"`
	Baseline  GroupCountCell `json:"baseline"`
	Selection GroupCountCell `json:"selection"`
	Change    ViewChange     `json:"change"`
}

type GroupCountCell struct {
	Presence string  `json:"presence"`
	State    string  `json:"state"`
	Value    *string `json:"value"`
	Reason   string  `json:"reason,omitempty"`
}

// Called after the existing broad comparison has validated both cohorts. This
// optional narrow profile never changes that operation's existing admission.
func cohortGroupCountComparison(request ComparisonRequest, summary *ViewComparison) *GroupCountComparison {
	if summary == nil || summary.State == "incompatible" || request.Numeric || request.Distribution || CompatibleCapturedCountQueries(request.Baseline.Question, request.Selection.Question) != nil {
		return nil
	}
	a, _, err := request.Baseline.validate()
	if err != nil {
		return nil
	}
	b, _, err := request.Selection.validate()
	if err != nil {
		return nil
	}
	accounting := func(c Cohort) string {
		if c.State == "changed_between_reads" {
			return "diagnostic"
		}
		return "coherent"
	}
	result := &GroupCountComparison{Version: "goo.group-count-comparison/1", Baseline: summary.Baseline, Selection: summary.Selection, Field: request.Baseline.Question.GroupBy[0], State: summary.State, Reason: summary.Reason, BaselineAccounting: accounting(request.Baseline), SelectionAccounting: accounting(request.Selection), Rows: []GroupCountRow{}}
	keys := make([]string, 0, len(a)+len(b))
	for key := range a {
		keys = append(keys, key)
	}
	for key := range b {
		if _, exists := a[key]; !exists {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	cell := func(value ValueCount, present bool, side string) GroupCountCell {
		if side == "diagnostic" {
			presence := "unknown"
			if present {
				presence = "present"
			}
			return GroupCountCell{Presence: presence, State: "unavailable", Reason: "source-reads-disagree"}
		}
		if present {
			return GroupCountCell{Presence: "present", State: "derived", Value: viewText(value.Count)}
		}
		return GroupCountCell{Presence: "absent", State: "derived", Value: viewText("0")}
	}
	for _, key := range keys {
		av, aok := a[key]
		bv, bok := b[key]
		category := av.Value
		if !aok {
			category = bv.Value
		}
		if category != nil {
			category = viewText(*category)
		}
		row := GroupCountRow{Key: key, Value: category, Baseline: cell(av, aok, result.BaselineAccounting), Selection: cell(bv, bok, result.SelectionAccounting), Change: ViewChange{State: "unavailable", Unit: "records", Reason: "source-reads-disagree"}}
		if result.BaselineAccounting == "coherent" && result.SelectionAccounting == "coherent" {
			baseline, _ := new(big.Int).SetString(*row.Baseline.Value, 10)
			selection, _ := new(big.Int).SetString(*row.Selection.Value, 10)
			row.Change = ViewChange{State: "derived", Value: viewText(new(big.Int).Sub(selection, baseline).String()), Unit: "records"}
		}
		result.Rows = append(result.Rows, row)
	}
	return result
}
