package verification

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

// Independent summaries use the standard-library parse/decimal-roundtrip oracle
// from predicate properties, not EvaluateNumericValues or mergeNumeric.
func comparisonOracleSummary(b exploration.NumericBinding, values []*string) exploration.NumericSummary {
	out := exploration.NumericSummary{Unit: b.Unit, Count: strconv.Itoa(len(values)), Sum: "0", State: "no_valid_values"}
	counts := map[string]int{}
	sum := new(big.Int)
	var lo, hi *big.Int
	for _, raw := range values {
		state := predicateState(raw, b.Type)
		counts[state]++
		if state != "valid" {
			continue
		}
		n, _ := new(big.Int).SetString(*raw, 10)
		sum.Add(sum, n)
		if lo == nil || n.Cmp(lo) < 0 {
			lo = new(big.Int).Set(n)
		}
		if hi == nil || n.Cmp(hi) > 0 {
			hi = new(big.Int).Set(n)
		}
	}
	out.Valid = strconv.Itoa(counts["valid"])
	out.Missing = strconv.Itoa(counts["missing"])
	out.Empty = strconv.Itoa(counts["empty"])
	out.Invalid = strconv.Itoa(counts["invalid"])
	out.Sum = sum.String()
	if len(values) == 0 {
		out.State = "no_data"
	}
	if lo != nil {
		out.State = "observed"
		out.Min = lo.String()
		out.Max = hi.String()
	}
	return out
}
func comparisonOracleCohort(b exploration.NumericBinding, groups map[string][]*string) exploration.Cohort {
	q := exploration.Query{Signal: "spans", Environment: "ordinary-otel", FromMS: 1000, ToMS: 61000, Limit: 100, GroupBy: []exploration.Field{{Kind: "attribute", Key: "route"}}, Numeric: &b}
	c := exploration.Cohort{Question: q, Basis: "independent-fixture", State: "observed", Groups: []exploration.ValueCount{}}
	all := []*string{}
	for key, values := range groups {
		if len(values) == 0 {
			continue
		}
		s := comparisonOracleSummary(b, values)
		name := key
		c.Groups = append(c.Groups, exploration.ValueCount{Value: &name, Count: strconv.Itoa(len(values)), Numeric: &s})
		all = append(all, values...)
	}
	s := comparisonOracleSummary(b, all)
	c.Numeric = &s
	c.Count = s.Count
	if len(all) == 0 {
		c.State = "no_data"
	}
	return c
}
func comparisonClone[T any](v T) T {
	raw, _ := json.Marshal(v)
	var out T
	if e := json.Unmarshal(raw, &out); e != nil {
		panic(e)
	}
	return out
}
func comparisonOracleRat(sum, count string) *big.Rat {
	n, _ := new(big.Int).SetString(sum, 10)
	d, _ := new(big.Int).SetString(count, 10)
	return new(big.Rat).SetFrac(n, d)
}
func comparisonOracleDifference(a, z string) string {
	x, _ := new(big.Int).SetString(a, 10)
	y, _ := new(big.Int).SetString(z, 10)
	return y.Sub(y, x).String()
}
func checkNumericChanges(n *exploration.NumericMeasureComparison, b exploration.NumericBinding, a, z *exploration.NumericSummary, unknown bool) error {
	if n == nil || n.Binding != b || !reflect.DeepEqual(n.Baseline, a) || !reflect.DeepEqual(n.Selection, z) || len(n.Coverage) != 4 {
		return fmt.Errorf("numeric evidence/binding dropped")
	}
	state, reason := "comparable", ""
	switch {
	case unknown:
		state, reason = "unknown", "source-reads-disagree"
	case a == nil || z == nil:
		state, reason = "unavailable", "absent-group"
	case a.Count == "0" || z.Count == "0":
		state, reason = "unavailable", "empty-population"
	case a.Valid == "0" || z.Valid == "0":
		state, reason = "unavailable", "no-valid-values"
	}
	if n.State != state || n.Reason != reason {
		return fmt.Errorf("numeric state mismatch got=%s/%s want=%s/%s", n.State, n.Reason, state, reason)
	}
	ac, zc := []string{"0", "0", "0", "0"}, []string{"0", "0", "0", "0"}
	if a != nil {
		ac = []string{a.Valid, a.Missing, a.Empty, a.Invalid}
	}
	if z != nil {
		zc = []string{z.Valid, z.Missing, z.Empty, z.Invalid}
	}
	for i, kind := range []string{"valid", "missing", "empty", "invalid"} {
		r := n.Coverage[i]
		if r.Kind != kind || r.BaselineCount != ac[i] || r.SelectionCount != zc[i] {
			return fmt.Errorf("coverage identity/count mismatch")
		}
		diff, pp := "", ""
		if !unknown {
			diff = comparisonOracleDifference(ac[i], zc[i])
			if a != nil && z != nil && a.Count != "0" && z.Count != "0" {
				pp = new(big.Rat).Mul(new(big.Rat).Sub(comparisonOracleRat(zc[i], z.Count), comparisonOracleRat(ac[i], a.Count)), big.NewRat(100, 1)).RatString()
			}
		}
		if r.Difference != diff || r.DifferencePP != pp {
			return fmt.Errorf("inexact coverage change %+v want=%s/%s", r, diff, pp)
		}
	}
	if state != "comparable" {
		if n.SumDifference != "" || n.MinDifference != "" || n.MaxDifference != "" || n.MeanDifference != "" {
			return fmt.Errorf("unavailable/unknown numeric arithmetic exposed")
		}
		return nil
	}
	mean := new(big.Rat).Sub(comparisonOracleRat(z.Sum, z.Valid), comparisonOracleRat(a.Sum, a.Valid)).RatString()
	if n.SumDifference != comparisonOracleDifference(a.Sum, z.Sum) || n.MinDifference != comparisonOracleDifference(a.Min, z.Min) || n.MaxDifference != comparisonOracleDifference(a.Max, z.Max) || n.MeanDifference != mean {
		return fmt.Errorf("inexact integer/rational numeric delta")
	}
	parsed, ok := new(big.Rat).SetString(n.MeanDifference)
	if !ok || parsed.RatString() != n.MeanDifference {
		return fmt.Errorf("mean delta not canonical reduced rational")
	}
	return nil
}
func compareNumericProperty(a, z exploration.Cohort, b exploration.NumericBinding) error {
	request := exploration.ComparisonRequest{Numeric: true, Baseline: a, Selection: z}
	before, _ := json.Marshal(request)
	result, e := exploration.CompareCohorts(request)
	if e != nil {
		return e
	}
	after, _ := json.Marshal(request)
	if string(before) != string(after) {
		return fmt.Errorf("comparison mutated evidence")
	}
	unknown := a.State == "changed_between_reads" || z.State == "changed_between_reads"
	if result.Measures == nil {
		return fmt.Errorf("numeric measure missing")
	}
	if e := checkNumericChanges(result.Measures.Numeric, b, a.Numeric, z.Numeric, unknown); e != nil {
		return e
	}
	am, zm := map[string]*exploration.NumericSummary{}, map[string]*exploration.NumericSummary{}
	for _, g := range a.Groups {
		am[*g.Value] = g.Numeric
	}
	for _, g := range z.Groups {
		zm[*g.Value] = g.Numeric
	}
	for _, row := range result.Rows {
		if row.Value == nil || row.Measures == nil {
			return fmt.Errorf("group identity/measure missing")
		}
		if e := checkNumericChanges(row.Measures.Numeric, b, am[*row.Value], zm[*row.Value], false); e != nil {
			return e
		}
	}
	// New output summaries must own their evidence; caller edits cannot alter the
	// already-produced comparison, even though requests remain browser-held data.
	frozen, _ := json.Marshal(result)
	a.Numeric.Sum = "corrupted"
	if len(a.Groups) > 0 {
		a.Groups[0].Numeric.Sum = "corrupted"
	}
	again, _ := json.Marshal(result)
	if string(frozen) != string(again) {
		return fmt.Errorf("numeric comparison aliases input")
	}
	return nil
}
func TestHegelNumericComparisonExactRationalsAndCoverage(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		b := predicateBinding([]string{"int64", "uint64"}[hegel.Draw(tc, hegel.Integers(0, 1))], []string{"attribute", "resource"}[hegel.Draw(tc, hegel.Integers(0, 1))])
		makeGroups := func() map[string][]*string {
			groups := map[string][]*string{}
			for _, key := range []string{"shared", "one-sided"} {
				n := hegel.Draw(tc, hegel.Integers(0, 14))
				for i := 0; i < n; i++ {
					raw := predicateRawCorpus[hegel.Draw(tc, hegel.Integers(0, len(predicateRawCorpus)-1))]
					groups[key] = append(groups[key], raw)
				}
			}
			return groups
		}
		a := comparisonOracleCohort(b, makeGroups())
		z := comparisonOracleCohort(b, makeGroups())
		z.Question.FromMS = 2000
		z.Question.ToMS = 32000
		z.Question.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "population"}, Op: "eq", Value: "selection"}
		if hegel.Draw(tc, hegel.Integers(0, 5)) == 0 {
			a.State = "changed_between_reads"
		}
		if e := compareNumericProperty(a, z, b); e != nil {
			tc.Fatal(e)
		}
	}, options(100)...)
}
func comparisonValues(raw ...string) []*string {
	out := []*string{}
	for _, value := range raw {
		v := value
		out = append(out, &v)
	}
	return out
}
func TestNumericComparisonBoundariesAndUnavailable(t *testing.T) {
	for _, typ := range []string{"int64", "uint64"} {
		b := predicateBinding(typ, "attribute")
		for _, pair := range [][2][]*string{{comparisonValues("0", "1", "0"), comparisonValues("0", "1")}, {comparisonValues("9007199254740992"), comparisonValues("9007199254740993")}, {comparisonValues("-9223372036854775808"), comparisonValues("9223372036854775807")}, {comparisonValues("18446744073709551615"), comparisonValues("0")}, {nil, comparisonValues("1")}, {comparisonValues("bad", ""), comparisonValues("1")}} {
			a := comparisonOracleCohort(b, map[string][]*string{"shared": pair[0]})
			z := comparisonOracleCohort(b, map[string][]*string{"shared": pair[1], "selection-only": comparisonValues("7")})
			if e := compareNumericProperty(a, z, b); e != nil {
				t.Fatal(typ, e)
			}
		}
	}
}
func TestNumericComparisonRejectsBindingsPayloadAndPreservesLegacy(t *testing.T) {
	b := predicateBinding("int64", "attribute")
	a := comparisonOracleCohort(b, map[string][]*string{"shared": comparisonValues("1", "2")})
	z := comparisonClone(a)
	for _, mutate := range []func(*exploration.Cohort){func(c *exploration.Cohort) { c.Question.Numeric.Field.Kind = "resource" }, func(c *exploration.Cohort) { c.Question.Numeric.Field.Key = "other" }, func(c *exploration.Cohort) { c.Question.Numeric.Type = "uint64" }, func(c *exploration.Cohort) { c.Question.Numeric.Unit = "ms" }, func(c *exploration.Cohort) { c.Question.Numeric.Parse = "coerce" }, func(c *exploration.Cohort) { c.Numeric = nil }, func(c *exploration.Cohort) { c.Groups[0].Numeric = nil }, func(c *exploration.Cohort) { c.Numeric.Valid = "3" }, func(c *exploration.Cohort) { c.Groups[0].Numeric.Sum = "4" }} {
		bad := comparisonClone(z)
		mutate(&bad)
		if _, e := exploration.CompareCohorts(exploration.ComparisonRequest{Numeric: true, Baseline: a, Selection: bad}); e == nil {
			t.Fatal("incompatible/malformed numeric comparison accepted")
		}
	}
	// Same numeric binding with an optional histogram presentation remains numeric
	// comparable; that display setting does not retype a measure.
	z.Question.NumericHistogram = &exploration.NumericHistogramSpec{Binning: "powers-of-two"}
	if e := compareNumericProperty(comparisonClone(a), z, b); e != nil {
		t.Fatal(e)
	}
	legacy := comparisonClone(a)
	legacy.Numeric = nil
	for i := range legacy.Groups {
		legacy.Groups[i].Numeric = nil
	}
	request := exploration.ComparisonRequest{Baseline: legacy, Selection: legacy}
	wire, _ := json.Marshal(request)
	var envelope map[string]json.RawMessage
	json.Unmarshal(wire, &envelope)
	if len(envelope) != 2 || envelope["numeric"] != nil {
		t.Fatal("legacy request gained numeric mode")
	}
	result, e := exploration.CompareCohorts(request)
	if e != nil || result.Measures != nil {
		t.Fatal("legacy count-only comparison changed", e)
	}
	again, _ := json.Marshal(request)
	if string(wire) != string(again) {
		t.Fatal("legacy immutable request changed")
	}
	if _, e := exploration.CompareCohorts(exploration.ComparisonRequest{Baseline: a, Selection: a}); e == nil {
		t.Fatal("numeric payload silently treated as count-only")
	}
}
