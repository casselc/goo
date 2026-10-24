package exploration

import (
	"encoding/json"
	"math"
	"math/big"
	"sort"
)

const TracePageComparisonVersion = "goo/trace-page-comparison/v1"

type TraceComparisonSide struct {
	TraceState string              `json:"traceState"`
	TraceID    *string             `json:"traceId"`
	Breakdown  *TracePageBreakdown `json:"breakdown"`
	StartNS    *string             `json:"startNs"`
	EndNS      *string             `json:"endNs"`
	ExtentNS   *string             `json:"extentNs"`
}
type TracePageComparison struct {
	Version   string                 `json:"version"`
	Baseline  TraceComparisonSide    `json:"baseline"`
	Selection TraceComparisonSide    `json:"selection"`
	Groups    []TraceComparisonGroup `json:"groups"`
}
type TraceComparisonGroup struct {
	MatchReason           string                   `json:"matchReason,omitempty"`
	Service               *string                  `json:"service"`
	Operation             *string                  `json:"operation"`
	Baseline              *TracePageBreakdownGroup `json:"baseline"`
	Selection             *TracePageBreakdownGroup `json:"selection"`
	MatchState            string                   `json:"matchState"`
	InclusiveDifferenceNS *string                  `json:"inclusiveDifferenceNs,omitempty"`
	OutsideDifferenceNS   *string                  `json:"outsideDifferenceNs,omitempty"`
	Match                 *TraceRowMatch           `json:"match,omitempty"`
}
type TraceRowMatch struct {
	BaselineIndex             int     `json:"baselineIndex"`
	SelectionIndex            int     `json:"selectionIndex"`
	InclusiveDifferenceNS     *string `json:"inclusiveDifferenceNs,omitempty"`
	OutsideDifferenceNS       *string `json:"outsideDifferenceNs,omitempty"`
	RelativeStartDifferenceNS *string `json:"relativeStartDifferenceNs,omitempty"`
}

func traceSide(rows []json.RawMessage, truncated bool) (TraceComparisonSide, error) {
	b, e := AnalyzeTracePage(rows, truncated)
	s := TraceComparisonSide{Breakdown: b, TraceState: "no-valid-trace"}
	traces := map[string]bool{}
	for _, raw := range rows {
		var v struct {
			Trace string `json:"TraceId"`
		}
		if json.Unmarshal(raw, &v) == nil && exampleID(v.Trace, 32) {
			traces[v.Trace] = true
		}
	}
	if len(traces) == 1 {
		s.TraceState = "single-trace"
		for id := range traces {
			v := id
			s.TraceID = &v
		}
	} else if len(traces) > 1 {
		s.TraceState = "mixed-traces"
	}
	if e != nil {
		return s, e
	}
	var low, high uint64
	known := false
	for _, raw := range rows {
		start, end, ok := traceComparisonInterval(raw)
		if !ok {
			continue
		}
		if !known || start < low {
			low = start
		}
		if !known || end > high {
			high = end
		}
		known = true
	}
	if known {
		a, z, d := new(big.Int).SetUint64(low).String(), new(big.Int).SetUint64(high).String(), new(big.Int).SetUint64(high-low).String()
		s.StartNS, s.EndNS, s.ExtentNS = &a, &z, &d
	}
	return s, nil
}
func traceComparisonInterval(raw json.RawMessage) (uint64, uint64, bool) {
	var r struct {
		Start    string `json:"timestamp_ns"`
		Duration string `json:"duration_ns"`
	}
	if json.Unmarshal(raw, &r) != nil {
		return 0, 0, false
	}
	s, sok := canonicalNS(r.Start)
	d, dok := canonicalNS(r.Duration)
	return s, s + d, sok && dok && d <= math.MaxUint64-s
}
func traceUniqueIdentities(rows []json.RawMessage) map[int]bool {
	keys := map[string][]int{}
	for i, raw := range rows {
		var r struct {
			Trace string `json:"TraceId"`
			Span  string `json:"SpanId"`
		}
		if json.Unmarshal(raw, &r) == nil && exampleID(r.Trace, 32) && exampleID(r.Span, 16) {
			k := r.Trace + ":" + r.Span
			keys[k] = append(keys[k], i)
		}
	}
	out := map[int]bool{}
	for _, indexes := range keys {
		if len(indexes) == 1 {
			out[indexes[0]] = true
		}
	}
	return out
}
func traceDifference(a, b *string) *string {
	if a == nil || b == nil {
		return nil
	}
	x, _ := new(big.Int).SetString(*a, 10)
	y, _ := new(big.Int).SetString(*b, 10)
	s := y.Sub(y, x).String()
	return &s
}
func traceGroupKey(service, operation *string) string {
	raw, _ := json.Marshal([2]*string{service, operation})
	return string(raw)
}

// CompareTracePages aligns supplied observations, not unique invocations or
// complete traces. Repeated service/operation rows are never zipped or deduped.
// Group differences compare only complete known metric sums on both pages;
// absent groups and unknown values never become invented zero durations.
func CompareTracePages(baselineRows, selectionRows []json.RawMessage, baselineTruncated, selectionTruncated bool) (*TracePageComparison, error) {
	a, e := traceSide(baselineRows, baselineTruncated)
	if e != nil {
		return nil, e
	}
	b, e := traceSide(selectionRows, selectionTruncated)
	if e != nil {
		return nil, e
	}
	out := &TracePageComparison{Version: TracePageComparisonVersion, Baseline: a, Selection: b, Groups: []TraceComparisonGroup{}}
	groups := map[string]*TraceComparisonGroup{}
	for _, g := range a.Breakdown.Groups {
		v := g
		k := traceGroupKey(g.Service, g.Operation)
		groups[k] = &TraceComparisonGroup{Service: g.Service, Operation: g.Operation, Baseline: &v, MatchState: "baseline-only"}
	}
	for _, g := range b.Breakdown.Groups {
		v := g
		k := traceGroupKey(g.Service, g.Operation)
		r := groups[k]
		if r == nil {
			r = &TraceComparisonGroup{Service: g.Service, Operation: g.Operation, MatchState: "selection-only"}
			groups[k] = r
		}
		r.Selection = &v
	}
	aIDs, bIDs := traceUniqueIdentities(baselineRows), traceUniqueIdentities(selectionRows)
	for _, g := range groups {
		if g.Baseline == nil || g.Selection == nil {
			continue
		}
		x, y := g.Baseline, g.Selection
		if x.InclusiveUnknown == 0 && y.InclusiveUnknown == 0 {
			g.InclusiveDifferenceNS = traceDifference(x.InclusiveNS, y.InclusiveNS)
		}
		if x.OutsideUnknown == 0 && y.OutsideUnknown == 0 {
			g.OutsideDifferenceNS = traceDifference(x.OutsideChildrenNS, y.OutsideChildrenNS)
		}
		if len(x.RowIndexes) != 1 || len(y.RowIndexes) != 1 {
			g.MatchState = "ambiguous"
			g.MatchReason = "repeated-label"
			continue
		}
		if g.Service == nil || g.Operation == nil {
			g.MatchState = "unavailable"
			g.MatchReason = "unreported-label"
			continue
		}
		if a.TraceState != "single-trace" || b.TraceState != "single-trace" {
			g.MatchState = "unavailable"
			g.MatchReason = "mixed-or-unreported-trace"
			continue
		}
		ai, bi := x.RowIndexes[0], y.RowIndexes[0]
		if !aIDs[ai] || !bIDs[bi] {
			g.MatchState = "unavailable"
			g.MatchReason = "invalid-or-duplicate-identity"
			continue
		}
		if traceMalformedShape(a.Breakdown.Rows[ai]) || traceMalformedShape(b.Breakdown.Rows[bi]) {
			g.MatchState = "unavailable"
			g.MatchReason = "malformed-row"
			continue
		}
		if !traceSameReportedKind(baselineRows[ai], selectionRows[bi]) {
			g.MatchState = "unavailable"
			g.MatchReason = "span-kind-mismatch-or-unreported"
			continue
		}
		g.MatchState = "unique-label-match"
		ar, br := a.Breakdown.Rows[ai], b.Breakdown.Rows[bi]
		g.Match = &TraceRowMatch{BaselineIndex: ai, SelectionIndex: bi, InclusiveDifferenceNS: traceDifference(ar.InclusiveNS, br.InclusiveNS), OutsideDifferenceNS: traceDifference(ar.OutsideChildrenNS, br.OutsideChildrenNS)}
		as, _, aok := traceComparisonInterval(baselineRows[ai])
		bs, _, bok := traceComparisonInterval(selectionRows[bi])
		if aok && bok && a.StartNS != nil && b.StartNS != nil {
			al, _ := new(big.Int).SetString(*a.StartNS, 10)
			bl, _ := new(big.Int).SetString(*b.StartNS, 10)
			av := new(big.Int).Sub(new(big.Int).SetUint64(as), al).String()
			bv := new(big.Int).Sub(new(big.Int).SetUint64(bs), bl).String()
			g.Match.RelativeStartDifferenceNS = traceDifference(&av, &bv)
		}
	}
	for _, g := range groups {
		out.Groups = append(out.Groups, *g)
	}
	sort.Slice(out.Groups, func(i, j int) bool {
		x, y := out.Groups[i], out.Groups[j]
		if c := compareOptionalTraceLabel(x.Service, y.Service); c != 0 {
			return c < 0
		}
		return compareOptionalTraceLabel(x.Operation, y.Operation) < 0
	})
	return out, nil
}

func traceMalformedShape(r TraceSpanBreakdown) bool {
	for _, reason := range r.Reasons {
		if reason == "invalid-row" {
			return true
		}
	}
	return false
}

func traceSameReportedKind(a, b json.RawMessage) bool {
	var x, y struct {
		Kind string `json:"SpanKind"`
	}
	return json.Unmarshal(a, &x) == nil && json.Unmarshal(b, &y) == nil && x.Kind != "" && x.Kind == y.Kind
}
