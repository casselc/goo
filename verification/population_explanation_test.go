package verification

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

const explanationClock = "2026-10-24T12:00:00Z"

func explanationValueKey(v *string) string {
	if v == nil {
		return "0"
	}
	return "1" + *v
}
func explanationFixture(fields int) (exploration.ExplanationRequest, exploration.ExplanationPopulation, exploration.ExplanationPopulation) {
	q := exploration.Query{Signal: "spans", Environment: "ordinary-otel", FromMS: 1000, ToMS: 61000, Limit: 100, Offset: 100, Summary: "span-duration", SpanStatus: true, BucketMS: 10000, GroupBy: []exploration.Field{{Kind: "attribute", Key: "old-group"}}, GroupTimeSeries: true, Where: &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "baseline"}}
	b := predicateBinding("uint64", "attribute")
	q.Numeric = &b
	q.NumericHistogram = &exploration.NumericHistogramSpec{Binning: "powers-of-two"}
	other := comparisonClone(q)
	other.Where.Value = "selection"
	other.FromMS = 2000
	other.ToMS = 32000
	req := exploration.ExplanationRequest{Version: "goo/explanation/v1", Baseline: q, Selection: other, Fields: []exploration.Field{}}
	for i := 0; i < fields; i++ {
		kind := "attribute"
		if i%2 == 1 {
			kind = "resource"
		}
		req.Fields = append(req.Fields, exploration.Field{Kind: kind, Key: fmt.Sprintf("field.%d", i/2)})
	}
	return req, exploration.ExplanationPopulation{Count: "0", ReadAt: explanationClock, Basis: strings.Repeat("a", 64)}, exploration.ExplanationPopulation{Count: "0", ReadAt: explanationClock, Basis: strings.Repeat("b", 64)}
}
func explanationRead(values []*string, weights []*big.Int, seed string) exploration.ExplanationFieldRead {
	counts := map[string]*big.Int{}
	labels := map[string]*string{}
	for i, value := range values {
		key := explanationValueKey(value)
		labels[key] = value
		if counts[key] == nil {
			counts[key] = new(big.Int)
		}
		counts[key].Add(counts[key], weights[i])
	}
	keys := []string{}
	for key, n := range counts {
		if n.Sign() > 0 {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	out := exploration.ExplanationFieldRead{State: "complete", ReadAt: explanationClock, Basis: strings.Repeat(seed, 64), Values: []exploration.ExplanationValueCount{}}
	for _, key := range keys {
		out.Values = append(out.Values, exploration.ExplanationValueCount{Value: labels[key], Count: counts[key].String()})
	}
	return out
}
func explanationCount(weights []*big.Int) string {
	n := new(big.Int)
	for _, w := range weights {
		n.Add(n, w)
	}
	return n.String()
}
func explanationExpectedRow(a, z exploration.ExplanationFieldRead, key string, base, selected string) (string, string, string, string, string) {
	ac, zc := "0", "0"
	for _, v := range a.Values {
		if explanationValueKey(v.Value) == key {
			ac = v.Count
		}
	}
	for _, v := range z.Values {
		if explanationValueKey(v.Value) == key {
			zc = v.Count
		}
	}
	ap, zp := distributionPercent(ac, base), distributionPercent(zc, selected)
	pp := ""
	if ap != "" && zp != "" {
		x, _ := new(big.Rat).SetString(ap)
		y, _ := new(big.Rat).SetString(zp)
		pp = y.Sub(y, x).RatString()
	}
	return ac, zc, ap, zp, pp
}
func explanationCheck(req exploration.ExplanationRequest, base, selected exploration.ExplanationPopulation, reads []exploration.ExplanationFieldResult) error {
	before, _ := json.Marshal(reads)
	r, e := exploration.BuildExplanation(req, base, selected, reads, explanationClock)
	if e != nil {
		return e
	}
	after, _ := json.Marshal(reads)
	if string(before) != string(after) {
		return fmt.Errorf("explanation mutated caller-held source reads")
	}
	if e := exploration.ValidateExplanationResult(r); e != nil {
		return e
	}
	if !reflect.DeepEqual(r.Request, req) || len(r.Fields) != len(reads) {
		return fmt.Errorf("original explanation scope/fields changed")
	}
	type lead struct {
		field, row int
		score      *big.Rat
		key        string
	}
	expected := []lead{}
	for i, f := range r.Fields {
		input := reads[i]
		if f.Field != req.Fields[i] {
			return fmt.Errorf("field namespace/key changed")
		}
		if f.State != "complete" && f.State != "no_data" {
			if len(f.Rows) != 0 {
				return fmt.Errorf("unavailable/changed field fabricated rows")
			}
			continue
		}
		union := map[string]bool{}
		for _, v := range input.Baseline.Values {
			union[explanationValueKey(v.Value)] = true
		}
		for _, v := range input.Selection.Values {
			union[explanationValueKey(v.Value)] = true
		}
		if len(f.Rows) != len(union) {
			return fmt.Errorf("complete field union truncated/missing-empty collapsed")
		}
		seen := map[string]bool{}
		for j, row := range f.Rows {
			key := explanationValueKey(row.Value)
			if !union[key] || seen[key] {
				return fmt.Errorf("duplicate/unexpected field value")
			}
			seen[key] = true
			ac, zc, ap, zp, pp := explanationExpectedRow(input.Baseline, input.Selection, key, base.Count, selected.Count)
			if row.BaselineCount != ac || row.SelectionCount != zc || row.BaselinePercent != ap || row.SelectionPercent != zp || row.DifferencePP != pp {
				return fmt.Errorf("inexact complete field share/delta")
			}
			if f.State == "complete" && pp != "" && pp != "0" {
				n, _ := new(big.Rat).SetString(pp)
				expected = append(expected, lead{i, j, new(big.Rat).Abs(n), f.Field.Kind + "\x00" + f.Field.Key + "\x00" + key})
			}
		}
	}
	sort.Slice(expected, func(i, j int) bool {
		if n := expected[i].score.Cmp(expected[j].score); n != 0 {
			return n > 0
		}
		return expected[i].key < expected[j].key
	})
	if len(r.Ranked) != len(expected) {
		return fmt.Errorf("rank includes unknown/zero/missing leads or truncates complete leads")
	}
	for i, lead := range r.Ranked {
		if lead.FieldIndex != expected[i].field || lead.RowIndex != expected[i].row {
			return fmt.Errorf("ranking used rounded shares or nondeterministic tie break")
		}
	}
	frozen, _ := json.Marshal(r)
	if len(reads) > 0 && len(reads[0].Baseline.Values) > 0 {
		reads[0].Baseline.Values[0].Count = "changed"
	}
	req.Baseline.Where.Value = "changed"
	again, _ := json.Marshal(r)
	if string(frozen) != string(again) {
		return fmt.Errorf("explanation aliases caller scopes/reads")
	}
	return nil
}
func TestHegelPopulationExplanationExactRanksAndCoverage(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		n := hegel.Draw(tc, hegel.Integers(1, 6))
		req, base, selected := explanationFixture(n)
		choices := []*string{nil}
		for _, value := range []string{"", "a", "b", "雪"} {
			v := value
			choices = append(choices, &v)
		}
		weights := func() []*big.Int {
			out := []*big.Int{}
			for i, n := 0, hegel.Draw(tc, hegel.Integers(0, 12)); i < n; i++ {
				scale := big.NewInt(1)
				if hegel.Draw(tc, hegel.Integers(0, 1)) == 1 {
					scale.Lsh(scale, 53)
				}
				scale.Mul(scale, big.NewInt(int64(hegel.Draw(tc, hegel.Integers(1, 7)))))
				scale.Add(scale, big.NewInt(int64(hegel.Draw(tc, hegel.Integers(0, 2)))))
				out = append(out, scale)
			}
			return out
		}
		aw, zw := weights(), weights()
		base.Count = explanationCount(aw)
		selected.Count = explanationCount(zw)
		reads := []exploration.ExplanationFieldResult{}
		for _, f := range req.Fields {
			values := func(weights []*big.Int) []*string {
				out := []*string{}
				for range weights {
					out = append(out, choices[hegel.Draw(tc, hegel.Integers(0, len(choices)-1))])
				}
				return out
			}
			reads = append(reads, exploration.ExplanationFieldResult{Field: f, Baseline: explanationRead(values(aw), aw, "c"), Selection: explanationRead(values(zw), zw, "d")})
		}
		if e := explanationCheck(req, base, selected, reads); e != nil {
			tc.Fatal(e)
		}
	}, options(100)...)
}
func TestPopulationExplanationFailuresAndLimits(t *testing.T) {
	req, base, selected := explanationFixture(1)
	base.Count = "2"
	selected.Count = "2"
	values := comparisonValues("", "a")
	weights := []*big.Int{big.NewInt(1), big.NewInt(1)}
	read := exploration.ExplanationFieldResult{Field: req.Fields[0], Baseline: explanationRead(values, weights, "c"), Selection: explanationRead(values, weights, "d")}
	for _, state := range []string{"partial", "unavailable"} {
		reads := []exploration.ExplanationFieldResult{comparisonClone(read)}
		reads[0].Baseline.State = state
		reads[0].Baseline.Reason = "source-read-failed"
		if state == "partial" {
			reads[0].Baseline.Reason = "cardinality-limit"
		}
		reads[0].Baseline.Values = []exploration.ExplanationValueCount{}
		r, e := exploration.BuildExplanation(req, base, selected, reads, explanationClock)
		if e != nil || len(r.Fields[0].Rows) != 0 || len(r.Ranked) != 0 {
			t.Fatal("partial source ranked/fabricated rows", e)
		}
	}
	changed := comparisonClone(read)
	changed.Baseline.Values[0].Count = "2"
	r, e := exploration.BuildExplanation(req, base, selected, []exploration.ExplanationFieldResult{changed}, explanationClock)
	if e != nil || r.Fields[0].State != "changed_between_reads" || len(r.Fields[0].Rows) != 0 || len(r.Ranked) != 0 {
		t.Fatal("cross-read change ranked", e)
	}
	for _, mutate := range []func(*exploration.ExplanationRequest){func(r *exploration.ExplanationRequest) { r.Version = "old" }, func(r *exploration.ExplanationRequest) { r.Fields = nil }, func(r *exploration.ExplanationRequest) { r.Fields = append(r.Fields, r.Fields[0]) }, func(r *exploration.ExplanationRequest) { r.Selection.Environment = "other" }, func(r *exploration.ExplanationRequest) { r.Selection.Signal = "logs" }, func(r *exploration.ExplanationRequest) { r.Fields[0].Key = strings.Repeat("x", 129) }} {
		bad := comparisonClone(req)
		mutate(&bad)
		if _, e := exploration.PlanExplanation(bad); e == nil {
			t.Fatal("invalid explanation request accepted")
		}
	}
	many, _, _ := explanationFixture(7)
	if _, e := exploration.PlanExplanation(many); e == nil {
		t.Fatal("field budget exceeded")
	}
	overflow := comparisonClone(read)
	overflow.Baseline.Values = nil
	for i := 0; i < 1001; i++ {
		v := fmt.Sprintf("%04d", i)
		overflow.Baseline.Values = append(overflow.Baseline.Values, exploration.ExplanationValueCount{Value: &v, Count: "1"})
	}
	if _, e := exploration.BuildExplanation(req, base, selected, []exploration.ExplanationFieldResult{overflow}, explanationClock); e == nil {
		t.Fatal("overflow complete grouping accepted")
	}
	oversized := comparisonClone(read)
	oversized.Baseline.Values = nil
	for i := 0; i < 40; i++ {
		v := fmt.Sprintf("%04d", i) + strings.Repeat("x", 4000)
		oversized.Baseline.Values = append(oversized.Baseline.Values, exploration.ExplanationValueCount{Value: &v, Count: "1"})
	}
	base.Count = "40"
	r, e = exploration.BuildExplanation(req, base, selected, []exploration.ExplanationFieldResult{oversized}, explanationClock)
	if e != nil || r.Fields[0].State != "unavailable" || len(r.Fields[0].Rows) != 0 || len(r.Fields[0].Baseline.Values) != 0 || len(r.Ranked) != 0 {
		t.Fatal("field byte bound did not withhold ranking/payload", e)
	}
}
func TestPopulationExplanationProjectionAndFilterPartition(t *testing.T) {
	req, _, _ := explanationFixture(2)
	if e := json.Unmarshal([]byte(`{"parent":{"service":"api","namespace":{"mode":"value","value":"ns"}},"child":{"service":"worker","namespace":{"mode":"value","value":"ns"}}}`), &req.Baseline.ReportedEdge); e != nil {
		t.Fatal(e)
	}
	req.Selection.ReportedEdge = comparisonClone(req.Baseline.ReportedEdge)
	for _, side := range []string{"baseline", "selection"} {
		parent := req.Baseline
		if side == "selection" {
			parent = req.Selection
		}
		for _, value := range []*string{nil, comparisonValues("")[0], comparisonValues("a")[0]} {
			include, e := exploration.PlanExplanationFilter(req, req.Fields[0], value, side, "include")
			if e != nil {
				t.Fatal(e)
			}
			exclude, e := exploration.PlanExplanationFilter(req, req.Fields[0], value, side, "exclude")
			if e != nil {
				t.Fatal(e)
			}
			for _, child := range []exploration.Query{include, exclude} {
				if child.Environment != parent.Environment || child.Signal != parent.Signal || child.FromMS != parent.FromMS || child.ToMS != parent.ToMS || !reflect.DeepEqual(child.ReportedEdge, parent.ReportedEdge) {
					t.Fatal("staged filter changed source scope")
				}
			}
			for _, raw := range []*string{nil, comparisonValues("")[0], comparisonValues("a")[0], comparisonValues("b")[0]} {
				v := map[exploration.Field]string{{Kind: "column", Key: "service"}: side}
				if raw != nil {
					v[req.Fields[0]] = *raw
				}
				want := raw == nil && value == nil || raw != nil && value != nil && *raw == *value
				if exploration.Match(include.Where, v) != want || exploration.Match(exclude.Where, v) == want {
					t.Fatal("include/exclude does not partition original population")
				}
				v[exploration.Field{Kind: "column", Key: "service"}] = "other"
				if exploration.Match(include.Where, v) || exploration.Match(exclude.Where, v) {
					t.Fatal("filter broadened parent selector")
				}
			}
		}
	}
	projectionReq := comparisonClone(req)
	for _, q := range []*exploration.Query{&projectionReq.Baseline, &projectionReq.Selection} {
		b := predicateBinding("uint64", "attribute")
		q.Where = &exploration.Predicate{All: []exploration.Predicate{*q.Where, predicateLeaf(b, "gte", "9007199254740993")}}
	}
	plan, e := exploration.PlanExplanation(projectionReq)
	if e != nil {
		t.Fatal(e)
	}
	for i, q := range []exploration.Query{plan.Baseline, plan.Selection} {
		parent := projectionReq.Baseline
		if i == 1 {
			parent = projectionReq.Selection
		}
		if q.Numeric != nil || q.NumericHistogram != nil || q.Summary != "" || q.SpanStatus || q.BucketMS != 0 || q.GroupTimeSeries || len(q.GroupBy) != 0 || q.Offset != 0 || !reflect.DeepEqual(q.Where, parent.Where) || !reflect.DeepEqual(q.ReportedEdge, parent.ReportedEdge) || q.FromMS != parent.FromMS || q.ToMS != parent.ToMS {
			t.Fatal("count projection changed selectors or retained measure/row metadata")
		}
	}
}

func TestPopulationExplanationMixedFailureAndExactNearTie(t *testing.T) {
	req, base, selected := explanationFixture(2)
	base.Count = "9223372036854775808"
	selected.Count = base.Count
	half := new(big.Int).Lsh(big.NewInt(1), 62)
	reads := []exploration.ExplanationFieldResult{}
	for i, f := range req.Fields {
		a, b := "a", "b"
		delta := big.NewInt(int64(i + 1))
		reads = append(reads, exploration.ExplanationFieldResult{Field: f, Baseline: explanationRead([]*string{&a, &b}, []*big.Int{new(big.Int).Set(half), new(big.Int).Set(half)}, "c"), Selection: explanationRead([]*string{&a, &b}, []*big.Int{new(big.Int).Add(half, delta), new(big.Int).Sub(half, delta)}, "d")})
	}
	if e := explanationCheck(comparisonClone(req), base, selected, comparisonClone(reads)); e != nil {
		t.Fatal(e)
	}
	r, e := exploration.BuildExplanation(req, base, selected, reads, explanationClock)
	if e != nil || len(r.Ranked) != 4 || r.Ranked[0].FieldIndex != 1 {
		t.Fatal("tiny exact lead lost to rounded percentage tie", e)
	}
	reads[0].Baseline.State = "partial"
	reads[0].Baseline.Reason = "cardinality-limit"
	reads[0].Baseline.Values = []exploration.ExplanationValueCount{}
	if e := explanationCheck(req, base, selected, reads); e != nil {
		t.Fatal("one failed field suppressed complete field leads", e)
	}
}

func TestPopulationExplanationExactSnapshotPromotion(t *testing.T) {
	req, base, selected := explanationFixture(1)
	base.Count = "2"
	selected.Count = "2"
	weights := []*big.Int{big.NewInt(1), big.NewInt(1)}
	read := exploration.ExplanationFieldResult{Field: req.Fields[0], Baseline: explanationRead(comparisonValues("", "a"), weights, "c"), Selection: explanationRead([]*string{nil, comparisonValues("")[0]}, weights, "d")}
	result, e := exploration.BuildExplanation(req, base, selected, []exploration.ExplanationFieldResult{read}, explanationClock)
	if e != nil {
		t.Fatal(e)
	}
	snapshots := &exploration.ResultSnapshots{}
	plan, e := exploration.PlanExplanation(req)
	if e != nil {
		t.Fatal(e)
	}
	legacy := &exploration.Result{Question: plan.Baseline, CapturedAt: explanationClock, Basis: strings.Repeat("f", 64), Count: "0", State: "no_data", Rows: []json.RawMessage{}, Groups: []json.RawMessage{}, Fields: []exploration.Field{}, Warnings: []string{}}
	oldToken, e := snapshots.Freeze(legacy)
	if e != nil {
		t.Fatal(e)
	}
	token, e := snapshots.FreezeExplanation(result)
	if e != nil {
		t.Fatal(e)
	}
	held, e := snapshots.Load(token)
	if e != nil || held.Question.Kind != "explanation" || held.Explanation == nil || held.Explore != nil {
		t.Fatal("typed snapshot confused with record page", e)
	}
	expected := comparisonClone(result)
	result.Fields[0].Rows[0].BaselineCount = "changed"
	if !reflect.DeepEqual(held.Explanation, expected) {
		t.Fatal("held explanation aliases caller output")
	}
	store := exploration.CaptureStore{Dir: t.TempDir()}
	saved, e := snapshots.PromoteExplanation(token, "Observed explanation", store)
	if e != nil {
		t.Fatal(e)
	}
	again, e := snapshots.PromoteExplanation(token, "Observed explanation", store)
	if e != nil || again.ID != saved.ID {
		t.Fatal("held exact promotion duplicated the same title/store", e)
	}
	reopened, e := store.Load(saved.ID)
	if e != nil || reopened.Question.Kind != "explanation" || !reflect.DeepEqual(reopened.Explanation, expected) || reopened.Explanation.ReadAt != explanationClock {
		t.Fatal("stored explanation reread or changed original evidence", e)
	}
	page, e := reopened.Page("")
	if e != nil || page.Explore != nil || !reflect.DeepEqual(page.Explanation, expected) {
		t.Fatal("explanation reopening manufactured record evidence", e)
	}
	oldHeld, e := snapshots.Load(oldToken)
	if e != nil || oldHeld.Explanation != nil || !reflect.DeepEqual(oldHeld.Explore, legacy) {
		t.Fatal("new typed explanation snapshot changed legacy result token", e)
	}
	if _, e := snapshots.PromoteExplanation(oldToken, "Wrong kind", store); e == nil {
		t.Fatal("legacy result token promoted as explanation")
	}
	corrupt := comparisonClone(expected)
	corrupt.Ranked = nil
	if len(expected.Ranked) > 0 {
		if _, e := snapshots.FreezeExplanation(corrupt); e == nil {
			t.Fatal("changed ranking accepted as exact source evidence")
		}
	}
}
