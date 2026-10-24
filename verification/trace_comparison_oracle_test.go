package verification

import (
	"encoding/json"
	"math/big"
	"sort"
	"strings"

	"github.com/casselc/goo/exploration"
)

// Independently derive all measures from raw pages. The page-side oracle uses
// endpoint cells; comparison is an exact optional-label set union, never a zip
// of spans or an expected result produced by CompareTracePages.
func tcInteger(v any) *big.Int {
	s, ok := v.(string)
	if !ok || s == "" || (len(s) > 1 && s[0] == '0') {
		return nil
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return nil
		}
	}
	n, ok := new(big.Int).SetString(s, 10)
	if !ok || n.BitLen() > 64 {
		return nil
	}
	return n
}
func tcObject(raw json.RawMessage) map[string]any {
	var out map[string]any
	if json.Unmarshal(raw, &out) != nil {
		return nil
	}
	return out
}
func tcInterval(raw json.RawMessage) (*big.Int, *big.Int) {
	obj := tcObject(raw)
	a, d := tcInteger(obj["timestamp_ns"]), tcInteger(obj["duration_ns"])
	if a == nil || d == nil {
		return nil, nil
	}
	end := new(big.Int).Add(a, d)
	if end.BitLen() > 64 {
		return nil, nil
	}
	return a, end
}
func tcID(s string, length int) bool {
	if len(s) != length || strings.Trim(s, "0") == "" {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func tcUnique(rows []json.RawMessage) map[int]bool {
	counts := map[string]int{}
	keys := make([]string, len(rows))
	for i, raw := range rows {
		obj := tcObject(raw)
		trace, tok := obj["TraceId"].(string)
		span, sok := obj["SpanId"].(string)
		if tok && sok && tcID(trace, 32) && tcID(span, 16) {
			keys[i] = trace + span
			counts[keys[i]]++
		}
	}
	out := map[int]bool{}
	for i, key := range keys {
		out[i] = key != "" && counts[key] == 1
	}
	return out
}
func tcOracleSide(rows []json.RawMessage, truncated bool) exploration.TraceComparisonSide {
	b := tpbOracle(rows, truncated)
	s := exploration.TraceComparisonSide{Breakdown: &b, TraceState: "no-valid-trace"}
	traces := map[string]bool{}
	for _, raw := range rows {
		obj := tcObject(raw)
		trace, _ := obj["TraceId"].(string)
		if tcID(trace, 32) {
			traces[trace] = true
		}
	}
	if len(traces) > 1 {
		s.TraceState = "mixed-traces"
	} else if len(traces) == 1 {
		s.TraceState = "single-trace"
		for trace := range traces {
			copy := trace
			s.TraceID = &copy
		}
	}
	var low, high *big.Int
	for _, raw := range rows {
		a, z := tcInterval(raw)
		if a == nil {
			continue
		}
		if low == nil || a.Cmp(low) < 0 {
			low = new(big.Int).Set(a)
		}
		if high == nil || z.Cmp(high) > 0 {
			high = new(big.Int).Set(z)
		}
	}
	if low != nil {
		a, z, d := low.String(), high.String(), new(big.Int).Sub(high, low).String()
		s.StartNS = &a
		s.EndNS = &z
		s.ExtentNS = &d
	}
	return s
}
func tcDelta(a, b *string) *string {
	if a == nil || b == nil {
		return nil
	}
	x, xok := new(big.Int).SetString(*a, 10)
	y, yok := new(big.Int).SetString(*b, 10)
	if !xok || !yok {
		return nil
	}
	s := new(big.Int).Sub(y, x).String()
	return &s
}
func tcKey(service, operation *string) string {
	raw, _ := json.Marshal([2]*string{service, operation})
	return string(raw)
}
func tcOptionalCompare(a, b *string) int {
	if a == nil {
		if b == nil {
			return 0
		}
		return -1
	}
	if b == nil {
		return 1
	}
	return strings.Compare(*a, *b)
}
func tcOracle(a, b []json.RawMessage, at, bt bool) exploration.TracePageComparison {
	out := exploration.TracePageComparison{Version: "goo/trace-page-comparison/v1", Baseline: tcOracleSide(a, at), Selection: tcOracleSide(b, bt), Groups: []exploration.TraceComparisonGroup{}}
	keys := map[string]bool{}
	left, right := map[string]exploration.TracePageBreakdownGroup{}, map[string]exploration.TracePageBreakdownGroup{}
	for _, g := range out.Baseline.Breakdown.Groups {
		k := tcKey(g.Service, g.Operation)
		keys[k] = true
		left[k] = g
	}
	for _, g := range out.Selection.Breakdown.Groups {
		k := tcKey(g.Service, g.Operation)
		keys[k] = true
		right[k] = g
	}
	au, bu := tcUnique(a), tcUnique(b)
	for key := range keys {
		x, xok := left[key]
		y, yok := right[key]
		g := exploration.TraceComparisonGroup{}
		if xok {
			copy := x
			g.Baseline = &copy
			g.Service = x.Service
			g.Operation = x.Operation
		}
		if yok {
			copy := y
			g.Selection = &copy
			g.Service = y.Service
			g.Operation = y.Operation
		}
		if !xok {
			g.MatchState = "selection-only"
		} else if !yok {
			g.MatchState = "baseline-only"
		} else {
			if x.InclusiveUnknown+y.InclusiveUnknown == 0 {
				g.InclusiveDifferenceNS = tcDelta(x.InclusiveNS, y.InclusiveNS)
			}
			if x.OutsideUnknown+y.OutsideUnknown == 0 {
				g.OutsideDifferenceNS = tcDelta(x.OutsideChildrenNS, y.OutsideChildrenNS)
			}
			if len(x.RowIndexes) != 1 || len(y.RowIndexes) != 1 {
				g.MatchState = "ambiguous"
				g.MatchReason = "repeated-label"
			} else {
				ai, bi := x.RowIndexes[0], y.RowIndexes[0]
				malformed := func(r exploration.TraceSpanBreakdown) bool {
					for _, reason := range r.Reasons {
						if reason == "invalid-row" {
							return true
						}
					}
					return false
				}
				ak, aok := tcObject(a[ai])["SpanKind"].(string)
				bk, bok := tcObject(b[bi])["SpanKind"].(string)
				if g.Service == nil || g.Operation == nil {
					g.MatchState = "unavailable"
					g.MatchReason = "unreported-label"
				} else if out.Baseline.TraceState != "single-trace" || out.Selection.TraceState != "single-trace" {
					g.MatchState = "unavailable"
					g.MatchReason = "mixed-or-unreported-trace"
				} else if !au[ai] || !bu[bi] {
					g.MatchState = "unavailable"
					g.MatchReason = "invalid-or-duplicate-identity"
				} else if malformed(out.Baseline.Breakdown.Rows[ai]) || malformed(out.Selection.Breakdown.Rows[bi]) {
					g.MatchState = "unavailable"
					g.MatchReason = "malformed-row"
				} else if !aok || !bok || ak == "" || ak != bk {
					g.MatchState = "unavailable"
					g.MatchReason = "span-kind-mismatch-or-unreported"
				} else {
					g.MatchState = "unique-label-match"
					ar, br := out.Baseline.Breakdown.Rows[ai], out.Selection.Breakdown.Rows[bi]
					g.Match = &exploration.TraceRowMatch{BaselineIndex: ai, SelectionIndex: bi, InclusiveDifferenceNS: tcDelta(ar.InclusiveNS, br.InclusiveNS), OutsideDifferenceNS: tcDelta(ar.OutsideChildrenNS, br.OutsideChildrenNS)}
					as, _ := tcInterval(a[ai])
					bs, _ := tcInterval(b[bi])
					if as != nil && bs != nil && out.Baseline.StartNS != nil && out.Selection.StartNS != nil {
						al, _ := new(big.Int).SetString(*out.Baseline.StartNS, 10)
						bl, _ := new(big.Int).SetString(*out.Selection.StartNS, 10)
						av, bv := new(big.Int).Sub(as, al).String(), new(big.Int).Sub(bs, bl).String()
						g.Match.RelativeStartDifferenceNS = tcDelta(&av, &bv)
					}
				}
			}
		}
		out.Groups = append(out.Groups, g)
	}
	sort.Slice(out.Groups, func(i, j int) bool {
		x, y := out.Groups[i], out.Groups[j]
		if c := tcOptionalCompare(x.Service, y.Service); c != 0 {
			return c < 0
		}
		return tcOptionalCompare(x.Operation, y.Operation) < 0
	})
	return out
}
func tcOracleEqual(want, got exploration.TracePageComparison) bool {
	for _, side := range []*exploration.TraceComparisonSide{&got.Baseline, &got.Selection} {
		if side.Breakdown == nil {
			return false
		}
		copy := *side.Breakdown
		copy.Rows = append([]exploration.TraceSpanBreakdown{}, copy.Rows...)
		for i := range copy.Rows {
			copy.Rows[i].Reasons = append([]string{}, copy.Rows[i].Reasons...)
			sort.Strings(copy.Rows[i].Reasons)
		}
		side.Breakdown = &copy
	}
	return aeJSONEqual(want, got)
}
