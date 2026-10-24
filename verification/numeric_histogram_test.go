package verification

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

// Independent domain construction: do not obtain expected bounds or memberships
// from the production histogram/range planner.
func histogramOracleEdges(typ string) []string {
	edges := []string{}
	if typ == "int64" {
		for power := 63; power >= 0; power-- {
			v := new(big.Int).Lsh(big.NewInt(1), uint(power))
			edges = append(edges, v.Neg(v).String())
		}
	}
	edges = append(edges, "0")
	last := 64
	if typ == "int64" {
		last = 63
	}
	for power := 0; power <= last; power++ {
		edges = append(edges, new(big.Int).Lsh(big.NewInt(1), uint(power)).String())
	}
	return edges
}
func histogramBoundaryValues(typ string) []*string {
	values := append([]*string{}, predicateRawCorpus...)
	for _, edge := range histogramOracleEdges(typ) {
		n, _ := new(big.Int).SetString(edge, 10)
		for _, delta := range []int64{-1, 0, 1} {
			value := new(big.Int).Add(n, big.NewInt(delta)).String()
			values = append(values, &value)
		}
	}
	return values
}
func histogramOracleCounts(b exploration.NumericBinding, values []*string) []string {
	edges := histogramOracleEdges(b.Type)
	counts := make([]int, len(edges)-1)
	for _, raw := range values {
		if predicateState(raw, b.Type) != "valid" {
			continue
		}
		v, _ := new(big.Int).SetString(*raw, 10)
		found := false
		for i := range counts {
			lo, _ := new(big.Int).SetString(edges[i], 10)
			hi, _ := new(big.Int).SetString(edges[i+1], 10)
			if v.Cmp(lo) >= 0 && v.Cmp(hi) < 0 {
				counts[i]++
				found = true
				break
			}
		}
		if !found {
			panic("oracle valid integer outside domain")
		}
	}
	out := make([]string, len(counts))
	for i, n := range counts {
		out[i] = strconv.Itoa(n)
	}
	return out
}
func histogramParent(b exploration.NumericBinding) exploration.Query {
	return exploration.Query{Signal: "spans", Environment: "ordinary-otel", FromMS: 1000, ToMS: 61000, Limit: 100, Numeric: &b, NumericHistogram: &exploration.NumericHistogramSpec{Binning: "powers-of-two"}, Where: &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "allowed"}}
}
func histogramCheck(b exploration.NumericBinding, values []*string) error {
	expected := histogramOracleEdges(b.Type)
	edges, e := exploration.NumericHistogramEdges(b.Type)
	if e != nil || !reflect.DeepEqual(expected, edges) {
		return fmt.Errorf("derived domain bounds differ: %s", b.Type)
	}
	h, e := exploration.EvaluateNumericHistogram(b, values)
	if e != nil {
		return e
	}
	counts := histogramOracleCounts(b, values)
	if h.Binning != "powers-of-two" || len(h.Bins) != len(counts) {
		return fmt.Errorf("wrong histogram shape")
	}
	total := 0
	for i, bin := range h.Bins {
		if bin.From != expected[i] || bin.To != expected[i+1] || bin.Count != counts[i] {
			return fmt.Errorf("wrong exact bin %d: %+v", i, bin)
		}
		n, _ := strconv.Atoi(bin.Count)
		total += n
	}
	valid := 0
	for _, raw := range values {
		if predicateState(raw, b.Type) == "valid" {
			valid++
		}
	}
	if total != valid {
		return fmt.Errorf("valid population lost")
	}
	summary, e := exploration.EvaluateNumericValues(b, values)
	if e != nil || h.Validate(b, summary) != nil {
		return fmt.Errorf("histogram and coverage disagree")
	}
	q := histogramParent(b)
	p, e := exploration.Compile(q)
	if e != nil {
		return e
	}
	if strings.Contains(p.Count, " LIMIT ") {
		return fmt.Errorf("histogram count bounded by displayed page")
	}
	if len(p.Count) > 262144 {
		return fmt.Errorf("histogram exceeds baseline ClickHouse query text budget")
	}
	return nil
}
func histogramRangeCheck(b exploration.NumericBinding, from, to int) error {
	q := histogramParent(b)
	before, _ := json.Marshal(q)
	plan, e := exploration.PlanNumericHistogramSelection(q, from, to)
	if e != nil {
		return e
	}
	edges := histogramOracleEdges(b.Type)
	if plan.FromBin != from || plan.ToBin != to || plan.From != edges[from] || plan.To != edges[to] || plan.Binning != "powers-of-two" || !reflect.DeepEqual(plan.Parent, q) {
		return fmt.Errorf("range plan changed parent or derived edges")
	}
	for _, child := range []exploration.Query{plan.Selected, plan.Rest, plan.Comparison.Baseline, plan.Comparison.Selection} {
		if child.Validate() != nil || child.Signal != q.Signal || child.Environment != q.Environment || child.FromMS != q.FromMS || child.ToMS != q.ToMS {
			return fmt.Errorf("range plan lost checked fixed scope")
		}
		if _, e := exploration.Compile(child); e != nil {
			return e
		}
	}
	if plan.Comparison.Baseline.Numeric != nil || plan.Comparison.Selection.Numeric != nil || plan.Comparison.Baseline.NumericHistogram != nil || plan.Comparison.Selection.NumericHistogram != nil || !reflect.DeepEqual(plan.Comparison.Baseline.Where, plan.Rest.Where) || !reflect.DeepEqual(plan.Comparison.Selection.Where, plan.Selected.Where) {
		return fmt.Errorf("comparison changed cohort or retained measure instead of record counts")
	}
	lo, _ := new(big.Int).SetString(edges[from], 10)
	hi, _ := new(big.Int).SetString(edges[to], 10)
	values := histogramBoundaryValues(b.Type)
	for _, raw := range values {
		for _, inParent := range []bool{false, true} {
			v := predicateValues(b, raw)
			v[exploration.Field{Kind: "column", Key: "service"}] = "other"
			if inParent {
				v[exploration.Field{Kind: "column", Key: "service"}] = "allowed"
			}
			wanted := false
			if inParent && predicateState(raw, b.Type) == "valid" {
				n, _ := new(big.Int).SetString(*raw, 10)
				wanted = n.Cmp(lo) >= 0 && n.Cmp(hi) < 0
			}
			selected := exploration.Match(plan.Selected.Where, v)
			rest := exploration.Match(plan.Rest.Where, v)
			if selected != wanted || rest != (inParent && !wanted) || selected && rest {
				return fmt.Errorf("selected/rest partition failure %s bins[%d,%d)", b.Type, from, to)
			}
		}
	}
	// All numeric thresholds are ordinary in-range decimals. Sentinel max+1 is
	// permitted as histogram metadata only, never as a compiled predicate value.
	var checkedThresholds func(*exploration.Predicate) bool
	checkedThresholds = func(p *exploration.Predicate) bool {
		if p == nil {
			return true
		}
		if p.Numeric != nil && p.Value != "" && predicateState(&p.Value, p.Numeric.Type) != "valid" {
			return false
		}
		for i := range p.All {
			if !checkedThresholds(&p.All[i]) {
				return false
			}
		}
		for i := range p.Any {
			if !checkedThresholds(&p.Any[i]) {
				return false
			}
		}
		return true
	}
	if !checkedThresholds(plan.Selected.Where) || !checkedThresholds(plan.Rest.Where) {
		return fmt.Errorf("histogram sentinel leaked into threshold")
	}
	after, _ := json.Marshal(q)
	if string(before) != string(after) {
		return fmt.Errorf("planning mutated input")
	}
	frozen, _ := json.Marshal(plan)
	q.Numeric.Unit = "changed"
	q.Where.Value = "changed"
	again, _ := json.Marshal(plan)
	if string(frozen) != string(again) {
		return fmt.Errorf("range plan aliases caller-owned binding or selectors")
	}
	return nil
}
func TestHegelNumericHistogramExactPartition(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		typ := []string{"int64", "uint64"}[hegel.Draw(tc, hegel.Integers(0, 1))]
		kind := []string{"attribute", "resource"}[hegel.Draw(tc, hegel.Integers(0, 1))]
		b := predicateBinding(typ, kind)
		values := append([]*string{}, predicateRawCorpus...)
		for i, n := 0, hegel.Draw(tc, hegel.Integers(0, 80)); i < n; i++ {
			raw := strconv.Itoa(hegel.Draw(tc, hegel.Integers(-1000000, 1000000)))
			if hegel.Draw(tc, hegel.Integers(0, 4)) == 0 {
				raw = "0" + raw
			}
			values = append(values, &raw)
		}
		if e := histogramCheck(b, values); e != nil {
			tc.Fatal(e)
		}
	}, options(100)...)
}
func TestHegelNumericHistogramRangePartition(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		typ := []string{"int64", "uint64"}[hegel.Draw(tc, hegel.Integers(0, 1))]
		b := predicateBinding(typ, []string{"attribute", "resource"}[hegel.Draw(tc, hegel.Integers(0, 1))])
		count := len(histogramOracleEdges(typ)) - 1
		from := hegel.Draw(tc, hegel.Integers(0, count-1))
		to := hegel.Draw(tc, hegel.Integers(from+1, count))
		if e := histogramRangeCheck(b, from, to); e != nil {
			tc.Fatal(e)
		}
	}, options(100)...)
}
func TestNumericHistogramDomainEdgesAndRefusal(t *testing.T) {
	for _, typ := range []string{"int64", "uint64"} {
		for _, kind := range []string{"attribute", "resource"} {
			b := predicateBinding(typ, kind)
			if e := histogramCheck(b, histogramBoundaryValues(typ)); e != nil {
				t.Fatal(e)
			}
			edges := histogramOracleEdges(typ)
			count := len(edges) - 1
			for _, pair := range [][2]int{{0, count}, {0, 1}, {count - 1, count}, {count / 2, count/2 + 1}} {
				if e := histogramRangeCheck(b, pair[0], pair[1]); e != nil {
					t.Fatal(e)
				}
			}
			for _, pair := range [][2]int{{-1, 1}, {0, 0}, {1, 0}, {0, count + 1}} {
				if _, e := exploration.PlanNumericHistogramSelection(histogramParent(b), pair[0], pair[1]); e == nil {
					t.Fatal("invalid bin index accepted", pair)
				}
			}
			h, e := exploration.EvaluateNumericHistogram(b, predicateRawCorpus)
			if e != nil {
				t.Fatal(e)
			}
			summary, _ := exploration.EvaluateNumericValues(b, predicateRawCorpus)
			for _, mutate := range []func(*exploration.NumericHistogram){func(h *exploration.NumericHistogram) { h.Bins[0].Count = "01" }, func(h *exploration.NumericHistogram) { h.Bins[0].From = "-999999999999999999999" }, func(h *exploration.NumericHistogram) { h.Bins = h.Bins[1:] }, func(h *exploration.NumericHistogram) { h.Binning = "min-max" }, func(h *exploration.NumericHistogram) { h.Bins[0].Count = "18446744073709551615" }} {
				raw, _ := json.Marshal(h)
				var bad exploration.NumericHistogram
				json.Unmarshal(raw, &bad)
				mutate(&bad)
				if bad.Validate(b, summary) == nil {
					t.Fatal("invalid histogram accepted")
				}
			}
		}
	}
	if _, e := exploration.NumericHistogramEdges("float64"); e == nil {
		t.Fatal("unsupported numeric domain accepted")
	}
	q := histogramParent(predicateBinding("uint64", "attribute"))
	q.Numeric = nil
	if q.Validate() == nil {
		t.Fatal("histogram without numeric binding accepted")
	}
	q = histogramParent(predicateBinding("int64", "resource"))
	q.NumericHistogram.Binning = "auto"
	if q.Validate() == nil {
		t.Fatal("unapproved bin strategy accepted")
	}
}
