package verification

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

// Only independent stdlib parse/roundtrip and big.Int domain construction from
// earlier verification helpers predict these exact bins. No production numeric
// evaluator, histogram evaluator or bin planner is used as the oracle.
func distributionOracleHistogram(b exploration.NumericBinding, values []*string) *exploration.NumericHistogram {
	edges := histogramOracleEdges(b.Type)
	counts := histogramOracleCounts(b, values)
	h := &exploration.NumericHistogram{Binning: "powers-of-two", Bins: []exploration.NumericHistogramBin{}}
	for i, count := range counts {
		h.Bins = append(h.Bins, exploration.NumericHistogramBin{From: edges[i], To: edges[i+1], Count: count})
	}
	return h
}
func distributionOracleCohort(b exploration.NumericBinding, values []*string, side string) exploration.Cohort {
	c := comparisonOracleCohort(b, map[string][]*string{"shared": values})
	c.Question.NumericHistogram = &exploration.NumericHistogramSpec{Binning: "powers-of-two"}
	c.NumericHistogram = distributionOracleHistogram(b, values)
	c.Question.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "population"}, Op: "eq", Value: side}
	if side == "selection" {
		c.Question.FromMS = 2000
		c.Question.ToMS = 32000
	}
	return c
}
func distributionPercent(count, total string) string {
	if total == "0" {
		return ""
	}
	return new(big.Rat).Mul(comparisonOracleRat(count, total), big.NewRat(100, 1)).RatString()
}
func distributionCheck(req exploration.ComparisonRequest) error {
	before, _ := json.Marshal(req)
	r, e := exploration.CompareCohorts(req)
	if e != nil {
		return e
	}
	after, _ := json.Marshal(req)
	if string(before) != string(after) {
		return fmt.Errorf("distribution comparison mutated evidence")
	}
	if r.Measures == nil || r.Measures.Numeric == nil || r.Measures.Numeric.Distribution == nil {
		return fmt.Errorf("explicit distribution omitted")
	}
	aggregate := len(req.Baseline.Question.GroupBy) == 0 && len(req.Selection.Question.GroupBy) == 0
	if r.AggregateOnly != aggregate || aggregate && (len(r.Rows) != 0 || r.Field != (exploration.Field{})) {
		return fmt.Errorf("aggregate-only comparison fabricated a grouped result")
	}
	d := r.Measures.Numeric.Distribution
	b := *req.Baseline.Question.Numeric
	edges := histogramOracleEdges(b.Type)
	if d.Binning != "powers-of-two" || d.Denominator != "population-count" || d.BaselineCount != req.Baseline.Count || d.SelectionCount != req.Selection.Count || len(d.Bins) != len(edges)-1 {
		return fmt.Errorf("distribution identity/denominator/domain mismatch")
	}
	unknown := req.Baseline.State == "changed_between_reads" || req.Selection.State == "changed_between_reads"
	state, reason := "comparable", ""
	if unknown {
		state, reason = "unknown", "source-reads-disagree"
	} else if req.Baseline.Count == "0" || req.Selection.Count == "0" {
		state, reason = "unavailable", "empty-population"
	}
	if d.State != state || d.Reason != reason {
		return fmt.Errorf("distribution state %s/%s differs from %s/%s", d.State, d.Reason, state, reason)
	}
	for i, row := range d.Bins {
		a, z := req.Baseline.NumericHistogram.Bins[i].Count, req.Selection.NumericHistogram.Bins[i].Count
		if row.From != edges[i] || row.To != edges[i+1] || row.BaselineCount != a || row.SelectionCount != z {
			return fmt.Errorf("bin misaligned %d", i)
		}
		diff, ap, zp, pp := "", "", "", ""
		if !unknown {
			diff = comparisonOracleDifference(a, z)
			ap = distributionPercent(a, req.Baseline.Count)
			zp = distributionPercent(z, req.Selection.Count)
			if ap != "" && zp != "" {
				x, _ := new(big.Rat).SetString(ap)
				y, _ := new(big.Rat).SetString(zp)
				pp = y.Sub(y, x).RatString()
			}
		}
		if row.CountDifference != diff || row.BaselinePercent != ap || row.SelectionPercent != zp || row.DifferencePP != pp {
			return fmt.Errorf("inexact bin %d count/share: %+v want=%s,%s,%s,%s", i, row, diff, ap, zp, pp)
		}
	}
	// Bin shares sum to valid/ALL-records *100, not valid/valid *100. Coverage
	// excluded from bins is meaningful evidence and never renormalized away.
	if !unknown {
		for _, side := range []string{"baseline", "selection"} {
			cohort := req.Baseline
			if side == "selection" {
				cohort = req.Selection
			}
			if cohort.Count == "0" {
				continue
			}
			sum := new(big.Rat)
			for _, row := range d.Bins {
				v := row.BaselinePercent
				if side == "selection" {
					v = row.SelectionPercent
				}
				n, ok := new(big.Rat).SetString(v)
				if !ok {
					return fmt.Errorf("noncanonical bin percent")
				}
				sum.Add(sum, n)
			}
			if sum.RatString() != distributionPercent(cohort.Numeric.Valid, cohort.Count) {
				return fmt.Errorf("coverage excluded records lost in distribution")
			}
		}
	}
	frozen, _ := json.Marshal(r)
	req.Baseline.NumericHistogram.Bins[0].Count = "changed"
	req.Selection.NumericHistogram.Bins[0].From = "changed"
	again, _ := json.Marshal(r)
	if string(frozen) != string(again) {
		return fmt.Errorf("output distribution aliases input histogram")
	}
	return nil
}
func distributionSideCheck(req exploration.ComparisonRequest, side string, from, to int) error {
	parent := req.Baseline
	if side == "selection" {
		parent = req.Selection
	}
	n := new(big.Int)
	for i := from; i < to; i++ {
		v, _ := new(big.Int).SetString(parent.NumericHistogram.Bins[i].Count, 10)
		n.Add(n, v)
	}
	plan, e := exploration.PlanNumericDistributionSelection(req, side, from, to)
	unavailable := req.Baseline.State == "changed_between_reads" || req.Selection.State == "changed_between_reads" || parent.Count == "0" || n.Sign() == 0
	if unavailable {
		if e == nil {
			return fmt.Errorf("unknown or zero-count range selectable")
		}
		return nil
	}
	if e != nil {
		return e
	}
	edges := histogramOracleEdges(parent.Question.Numeric.Type)
	if !reflect.DeepEqual(plan.Parent, parent.Question) || plan.From != edges[from] || plan.To != edges[to] || plan.FromBin != from || plan.ToBin != to {
		return fmt.Errorf("side-specific range lost original question/domain")
	}
	for _, child := range []exploration.Query{plan.Selected, plan.Rest} {
		if child.Validate() != nil || child.Environment != parent.Question.Environment || child.FromMS != parent.Question.FromMS || child.ToMS != parent.Question.ToMS {
			return fmt.Errorf("side-specific scope changed")
		}
	}
	lo, _ := new(big.Int).SetString(edges[from], 10)
	hi, _ := new(big.Int).SetString(edges[to], 10)
	binding := *parent.Question.Numeric
	for _, raw := range histogramBoundaryValues(binding.Type) {
		v := predicateValues(binding, raw)
		v[exploration.Field{Kind: "attribute", Key: "population"}] = side
		want := false
		if predicateState(raw, binding.Type) == "valid" {
			x, _ := new(big.Int).SetString(*raw, 10)
			want = x.Cmp(lo) >= 0 && x.Cmp(hi) < 0
		}
		a, z := exploration.Match(plan.Selected.Where, v), exploration.Match(plan.Rest.Where, v)
		if a != want || z == want {
			return fmt.Errorf("side selected/rest partition changed")
		}
		v[exploration.Field{Kind: "attribute", Key: "population"}] = "other"
		if exploration.Match(plan.Selected.Where, v) || exploration.Match(plan.Rest.Where, v) {
			return fmt.Errorf("side range broadened original selector")
		}
	}
	return nil
}
func TestHegelDistributionComparisonExactSharesAndSidePlans(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		b := predicateBinding([]string{"int64", "uint64"}[hegel.Draw(tc, hegel.Integers(0, 1))], []string{"attribute", "resource"}[hegel.Draw(tc, hegel.Integers(0, 1))])
		values := func() []*string {
			out := []*string{}
			for i, n := 0, hegel.Draw(tc, hegel.Integers(0, 30)); i < n; i++ {
				out = append(out, predicateRawCorpus[hegel.Draw(tc, hegel.Integers(0, len(predicateRawCorpus)-1))])
			}
			return out
		}
		req := exploration.ComparisonRequest{Numeric: true, Distribution: true, Baseline: distributionOracleCohort(b, values(), "baseline"), Selection: distributionOracleCohort(b, values(), "selection")}
		if hegel.Draw(tc, hegel.Integers(0, 1)) == 0 {
			req.Baseline.Question.GroupBy = nil
			req.Selection.Question.GroupBy = nil
			req.Baseline.Groups = []exploration.ValueCount{}
			req.Selection.Groups = []exploration.ValueCount{}
		}
		if hegel.Draw(tc, hegel.Integers(0, 6)) == 0 {
			req.Selection.State = "changed_between_reads"
		}
		if e := distributionCheck(comparisonClone(req)); e != nil {
			tc.Fatal(e)
		}
		bins := len(histogramOracleEdges(b.Type)) - 1
		from := hegel.Draw(tc, hegel.Integers(0, bins-1))
		to := hegel.Draw(tc, hegel.Integers(from+1, bins))
		for _, side := range []string{"baseline", "selection"} {
			if e := distributionSideCheck(req, side, from, to); e != nil {
				tc.Fatal(e)
			}
		}
	}, options(100)...)
}
func TestDistributionComparisonCoverageStatesAndDomain(t *testing.T) {
	for _, typ := range []string{"int64", "uint64"} {
		b := predicateBinding(typ, "attribute")
		for _, values := range [][]*string{nil, comparisonValues("bad", ""), histogramBoundaryValues(typ), comparisonValues("0", "9007199254740993", "18446744073709551615", "-9223372036854775808")} {
			req := exploration.ComparisonRequest{Numeric: true, Distribution: true, Baseline: distributionOracleCohort(b, values, "baseline"), Selection: distributionOracleCohort(b, append([]*string{nil}, comparisonValues("0", "1", "bad")...), "selection")}
			if e := distributionCheck(comparisonClone(req)); e != nil {
				t.Fatal(e)
			}
			if e := distributionSideCheck(req, "baseline", 0, len(histogramOracleEdges(typ))-1); e != nil {
				t.Fatal(e)
			}
			unknown := comparisonClone(req)
			unknown.Baseline.State = "changed_between_reads"
			if e := distributionCheck(unknown); e != nil {
				t.Fatal(e)
			}
		}
	}
}
func TestDistributionComparisonIntentLegacyAndRefusals(t *testing.T) {
	b := predicateBinding("uint64", "attribute")
	req := exploration.ComparisonRequest{Numeric: true, Distribution: true, Baseline: distributionOracleCohort(b, comparisonValues("1", "2", "bad"), "baseline"), Selection: distributionOracleCohort(b, comparisonValues("9007199254740993", "0"), "selection")}
	for _, mutate := range []func(*exploration.ComparisonRequest){func(r *exploration.ComparisonRequest) { r.Numeric = false }, func(r *exploration.ComparisonRequest) { r.Baseline.NumericHistogram = nil }, func(r *exploration.ComparisonRequest) { r.Selection.NumericHistogram.Bins[0].From = "-1" }, func(r *exploration.ComparisonRequest) { r.Selection.NumericHistogram.Bins[0].Count = "999999" }, func(r *exploration.ComparisonRequest) { r.Selection.Question.Numeric.Unit = "ms" }, func(r *exploration.ComparisonRequest) { r.Selection.NumericHistogram.Binning = "auto" }} {
		bad := comparisonClone(req)
		mutate(&bad)
		if _, e := exploration.CompareCohorts(bad); e == nil {
			t.Fatal("invalid explicit distribution accepted")
		}
	}
	for _, side := range []string{"other", ""} {
		if _, e := exploration.PlanNumericDistributionSelection(req, side, 0, 1); e == nil {
			t.Fatal("unknown side accepted")
		}
	}
	legacy := comparisonClone(req)
	legacy.Distribution = false
	legacy.Baseline.NumericHistogram = nil
	legacy.Selection.NumericHistogram = nil
	wire, _ := json.Marshal(legacy)
	var envelope map[string]json.RawMessage
	json.Unmarshal(wire, &envelope)
	if envelope["distribution"] != nil {
		t.Fatal("legacy numeric request gained distribution intent")
	}
	out, e := exploration.CompareCohorts(legacy)
	if e != nil || out.Measures == nil || out.Measures.Numeric == nil || out.Measures.Numeric.Distribution != nil {
		t.Fatal("legacy numeric comparison semantics changed", e)
	}
	after, _ := json.Marshal(legacy)
	if string(wire) != string(after) {
		t.Fatal("legacy numeric evidence mutated")
	}
	outAgain, e := exploration.CompareCohorts(comparisonClone(legacy))
	if e != nil || outAgain.Basis != out.Basis {
		t.Fatal("legacy request hash unstable")
	}
	if e := distributionSideCheck(req, "selection", 0, len(histogramOracleEdges(b.Type))-1); e != nil {
		t.Fatal(e)
	}
}

func TestDistributionComparisonExplicitAggregateOnly(t *testing.T) {
	for _, typ := range []string{"int64", "uint64"} {
		b := predicateBinding(typ, "resource")
		for _, values := range [][]*string{nil, comparisonValues("bad", ""), histogramBoundaryValues(typ)} {
			req := exploration.ComparisonRequest{Numeric: true, Distribution: true, Baseline: distributionOracleCohort(b, values, "baseline"), Selection: distributionOracleCohort(b, comparisonValues("0", "1", "9007199254740993"), "selection")}
			req.Baseline.Question.GroupBy = nil
			req.Selection.Question.GroupBy = nil
			req.Baseline.Groups = []exploration.ValueCount{}
			req.Selection.Groups = []exploration.ValueCount{}
			if e := distributionCheck(comparisonClone(req)); e != nil {
				t.Fatal(e)
			}
			for _, side := range []string{"baseline", "selection"} {
				if e := distributionSideCheck(req, side, 0, len(histogramOracleEdges(typ))-1); e != nil {
					t.Fatal(e)
				}
			}
			for _, mutate := range []func(*exploration.ComparisonRequest){
				func(r *exploration.ComparisonRequest) {
					r.Distribution = false
					r.Baseline.NumericHistogram = nil
					r.Selection.NumericHistogram = nil
				},
				func(r *exploration.ComparisonRequest) {
					r.Numeric = false
					r.Distribution = false
					r.Baseline.NumericHistogram = nil
					r.Selection.NumericHistogram = nil
					r.Baseline.Numeric = nil
					r.Selection.Numeric = nil
				},
				func(r *exploration.ComparisonRequest) {
					r.Selection.Question.GroupBy = []exploration.Field{{Kind: "attribute", Key: "route"}}
				},
				func(r *exploration.ComparisonRequest) {
					name := "fabricated"
					summary := comparisonOracleSummary(b, comparisonValues("1"))
					r.Selection.Groups = []exploration.ValueCount{{Value: &name, Count: "1", Numeric: &summary}}
				},
				func(r *exploration.ComparisonRequest) { r.Selection.NumericHistogram = nil },
			} {
				bad := comparisonClone(req)
				mutate(&bad)
				if _, e := exploration.CompareCohorts(bad); e == nil {
					t.Fatal("aggregate-only comparison escaped explicit intent or complete histogram constraints")
				}
			}
		}
	}
}
