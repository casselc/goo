package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"os/exec"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

func lpcGeneratedCohort(seed, duration int, counts []int) exploration.Cohort {
	q := exploration.Query{Signal: "logs", Environment: "pattern-pair-property", FromMS: 1800000000000 + int64(seed)*1000, Limit: 100, GroupBy: []exploration.Field{{Kind: "column", Key: "body_pattern_v1"}}, BucketMS: 1000}
	q.ToMS = q.FromMS + int64(duration)
	for _, bucket := range []int64{1000, 10000, 60000, 300000} {
		if (int64(duration)+bucket-1)/bucket <= 240 {
			q.BucketMS = bucket
			break
		}
	}
	q.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "resource", Key: "team"}, Op: "eq", Value: fmt.Sprintf("team-%d", seed%3)}
	ids := []string{lpOracle("common 123").Identity, lpOracle("baseline literal n").Identity, lpOracle("selection ⟪digits⟫").Identity, "lp1:e", "lp1:o"}
	c := exploration.Cohort{Question: q, Basis: fmt.Sprintf("%064x", seed), State: "observed", Groups: []exploration.ValueCount{}}
	total := 0
	for i, count := range counts {
		total += count
		if count > 0 {
			id := ids[i]
			c.Groups = append(c.Groups, exploration.ValueCount{Value: &id, Count: strconv.Itoa(count)})
		}
	}
	c.Count = strconv.Itoa(total)
	if total == 0 {
		c.State = "no_data"
	}
	return c
}
func TestHegelLogPatternComparisonCanonicalBoundary(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		a := hegel.Draw(tc, hegel.Integers(0, 10000))
		z := hegel.Draw(tc, hegel.Integers(0, 10000))
		duration := hegel.Draw(tc, hegel.Integers(1, 7200000))
		mode := hegel.Draw(tc, hegel.Integers(0, 2))
		request := exploration.ComparisonRequest{Baseline: lpcGeneratedCohort(seed, duration, []int{a, 7, 0, 2, 3}), Selection: lpcGeneratedCohort(seed+1, 1+duration/2, []int{z, 0, 11, 1, 4})}
		if mode == 1 {
			request.Baseline = lpcGeneratedCohort(seed, duration, []int{0, 0, 0, 0, 0})
		}
		if mode == 2 {
			request.Baseline.State = "changed_between_reads"
		}
		before, _ := json.Marshal(request)
		result, e := exploration.CompareCohorts(request)
		after, _ := json.Marshal(request)
		if e != nil || string(before) != string(after) {
			tc.Errorf("valid scopes changed/rejected %v", e)
			return
		}
		expected := "ok"
		if mode == 1 {
			expected = "no_data"
		}
		if mode == 2 {
			expected = "unknown"
		}
		if result.State != expected || result.BaselineCount != request.Baseline.Count || result.SelectionCount != request.Selection.Count {
			tc.Errorf("known-empty/unknown contract %+v", result)
		}
		if mode != 0 && len(result.Rows) != 0 {
			tc.Errorf("unavailable share rows exposed")
		}

		for _, id := range []string{"lp1:p:", "lp1:p:nn", "lp1:p:31", "lp1:p:6A", "lp1:p:ff", "lp2:e"} {
			if lpCanonical(id) {
				tc.Errorf("invalid independent ID corpus")
			}
			for _, side := range []string{"baseline", "selection"} {
				bad := request
				c := &bad.Selection
				if side == "baseline" {
					c = &bad.Baseline
				}
				c.Groups = []exploration.ValueCount{{Value: &id, Count: "1"}}
				c.Count = "1"
				if c.State == "no_data" {
					c.State = "observed"
				}
				// The untouched zero/changed opposite side remains valid. The only defect
				// on the mutated side is its ID: positive group count equals population.
				if _, e := exploration.CompareCohorts(bad); e == nil || !strings.Contains(e.Error(), "body_pattern_v1") {
					tc.Errorf("malformed %s pattern not checked before empty/unknown resolution %q: %v", side, id, e)
				}
			}
		}
		// Legacy service distributions retain arbitrary exact string values.
		legacy := request
		legacy.Baseline.Question.GroupBy = []exploration.Field{{Kind: "column", Key: "service"}}
		legacy.Selection.Question.GroupBy = []exploration.Field{{Kind: "column", Key: "service"}}
		value := "lp1:p:31"
		legacy.Baseline.Groups = []exploration.ValueCount{{Value: &value, Count: "1"}}
		legacy.Baseline.Count = "1"
		legacy.Baseline.State = "observed"
		if _, e := exploration.CompareCohorts(legacy); e != nil {
			tc.Errorf("pattern-only guard leaked into legacy service field: %v", e)
		}
	}, options(70)...)
}

type lpcModelRow struct {
	ID                                                     string
	BaselineCount, CurrentCount                            *string
	BaselineRate, CurrentRate, BaselineShare, CurrentShare *string
	CountDifference, RateDifference, ShareDifference       *string
	Presence                                               string
}
type lpcModel struct {
	State, Metric, BaselineCount, CurrentCount, BaselineWidthMS, CurrentWidthMS string
	Rows                                                                        []lpcModelRow
}

func lpcText(v *big.Rat) *string {
	if v == nil {
		return nil
	}
	s := v.RatString()
	return &s
}
func lpcDifference(a, z *string) *string {
	if a == nil || z == nil {
		return nil
	}
	x, ok := new(big.Rat).SetString(*a)
	if !ok {
		return nil
	}
	y, ok := new(big.Rat).SetString(*z)
	if !ok {
		return nil
	}
	return lpcText(new(big.Rat).Sub(y, x))
}
func lpcExpected(request exploration.ComparisonRequest, metric string) lpcModel {
	a, z := request.Baseline, request.Selection
	wa, wz := a.Question.ToMS-a.Question.FromMS, z.Question.ToMS-z.Question.FromMS
	out := lpcModel{State: "known", Metric: metric, BaselineCount: a.Count, CurrentCount: z.Count, BaselineWidthMS: strconv.FormatInt(wa, 10), CurrentWidthMS: strconv.FormatInt(wz, 10), Rows: []lpcModelRow{}}
	unknown := a.State == "changed_between_reads" || z.State == "changed_between_reads"
	if unknown {
		out.State = "unknown"
	}
	ag, zg := map[string]string{}, map[string]string{}
	union := map[string]bool{}
	for _, g := range a.Groups {
		ag[*g.Value] = g.Count
		union[*g.Value] = true
	}
	for _, g := range z.Groups {
		zg[*g.Value] = g.Count
		union[*g.Value] = true
	}
	for id := range union {
		ac, aok := ag[id]
		zc, zok := zg[id]
		row := lpcModelRow{ID: id, Presence: "unknown"}
		if aok {
			v := ac
			row.BaselineCount = &v
		}
		if zok {
			v := zc
			row.CurrentCount = &v
		}
		if !unknown {
			if !aok {
				ac = "0"
			}
			if !zok {
				zc = "0"
			}
			row.BaselineCount = &ac
			row.CurrentCount = &zc
			row.BaselineRate = lpcText(new(big.Rat).Mul(new(big.Rat).SetFrac(ncLiveInt(ac), big.NewInt(wa)), big.NewRat(1000, 1)))
			row.CurrentRate = lpcText(new(big.Rat).Mul(new(big.Rat).SetFrac(ncLiveInt(zc), big.NewInt(wz)), big.NewRat(1000, 1)))
			if a.Count != "0" {
				row.BaselineShare = lpcText(new(big.Rat).Mul(new(big.Rat).SetFrac(ncLiveInt(ac), ncLiveInt(a.Count)), big.NewRat(100, 1)))
			}
			if z.Count != "0" {
				row.CurrentShare = lpcText(new(big.Rat).Mul(new(big.Rat).SetFrac(ncLiveInt(zc), ncLiveInt(z.Count)), big.NewRat(100, 1)))
			}
			diff := new(big.Int).Sub(ncLiveInt(zc), ncLiveInt(ac)).String()
			row.CountDifference = &diff
			row.RateDifference = lpcDifference(row.BaselineRate, row.CurrentRate)
			row.ShareDifference = lpcDifference(row.BaselineShare, row.CurrentShare)
			row.Presence = "observed-in-both"
			if ac == "0" {
				row.Presence = "not-observed-in-baseline"
			} else if zc == "0" {
				row.Presence = "not-observed-in-current"
			}
		}
		out.Rows = append(out.Rows, row)
	}
	score := func(r lpcModelRow) *big.Rat {
		v := r.RateDifference
		if metric == "count" {
			v = r.CountDifference
		}
		if metric == "share" {
			v = r.ShareDifference
		}
		if v == nil {
			return nil
		}
		q, _ := new(big.Rat).SetString(*v)
		return new(big.Rat).Abs(q)
	}
	sort.Slice(out.Rows, func(i, j int) bool {
		x, y := score(out.Rows[i]), score(out.Rows[j])
		if x != nil && y == nil {
			return true
		}
		if x == nil && y != nil {
			return false
		}
		if x != nil && y != nil && x.Cmp(y) != 0 {
			return x.Cmp(y) > 0
		}
		return out.Rows[i].ID < out.Rows[j].ID
	})
	return out
}
func TestHegelLogPatternComparisonExactClientModel(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		base := hegel.Draw(tc, hegel.SampledFrom([]int{1, 7, 9007199254740993, 4503599627370497}))
		width := hegel.Draw(tc, hegel.Integers(1, 3600000))
		mode := hegel.Draw(tc, hegel.Integers(0, 4))
		request := exploration.ComparisonRequest{Baseline: lpcGeneratedCohort(seed, 2*width, []int{2 * base, 3 + seed%7, 0, 2, 3}), Selection: lpcGeneratedCohort(seed+1, width, []int{base, 0, 4 + seed%9, 1, 4})}
		switch mode {
		case 1:
			request.Baseline = lpcGeneratedCohort(seed, 2*width, []int{0, 0, 0, 0, 0})
		case 2:
			request.Selection = lpcGeneratedCohort(seed+1, width, []int{0, 0, 0, 0, 0})
		case 3:
			request.Baseline.State = "changed_between_reads"
		case 4:
			request.Selection.State = "changed_between_reads"
			request.Selection.Count = new(big.Int).Add(ncLiveInt(request.Selection.Count), big.NewInt(1)).String()
		}
		result, e := exploration.CompareCohorts(request)
		if e != nil {
			tc.Errorf("generated valid pair rejected %v", e)
			return
		}
		input := struct {
			Request exploration.ComparisonRequest `json:"request"`
			Result  *exploration.CohortComparison `json:"result"`
		}{request, result}
		raw, _ := json.Marshal(input)
		const script = `const fs=require('fs'),m=require('./ui/log-pattern-comparison.mjs'),x=JSON.parse(fs.readFileSync(0,'utf8')),before=JSON.stringify(x),models=['rate','count','share'].map(metric=>m.model(x.request,x.result,{metric})),frozen=JSON.stringify(models),previous=m.previousQuestion(x.request.selection.question),pair=m.plannedPair(x.request.selection.question,x.request.baseline.question),immutable=JSON.stringify(x)===before;
const mutations=[y=>y.result.baselineBasis+='wrong',y=>y.result.selectionCount+='0',y=>y.result.state=y.result.state==='ok'?'unknown':'ok',y=>{if(y.result.rows.length)y.result.rows.pop();else y.result.rows.push({value:'lp1:e',baselineCount:'0',selectionCount:'0'})},y=>y.request.selection.question.toMs=y.request.selection.question.fromMs,y=>y.request.selection.question.environment+='other',y=>{const c=y.request.selection.groups.length?y.request.selection:y.request.baseline;c.groups[0].g0='lp1:p:31'},y=>y.result.field.key='service'];
const refusals=mutations.map(change=>{const y=structuredClone(x);change(y);return m.model(y.request,y.result)===null});x.request.baseline.question.environment='edited';x.request.baseline.groups.length=0;console.log(JSON.stringify({models,previous,pair,immutable,stable:JSON.stringify(models)===frozen,refusals}));`
		cmd := exec.Command("node", "-e", script)
		cmd.Dir = ".."
		cmd.Stdin = bytes.NewReader(raw)
		output, e := cmd.CombinedOutput()
		if e != nil {
			tc.Errorf("actual model invocation %v %s", e, output)
			return
		}
		var got struct {
			Models            []*lpcModel
			Previous          exploration.Query
			Pair              struct{ Baseline, Selection exploration.Query }
			Immutable, Stable bool
			Refusals          []bool
		}
		if e = json.Unmarshal(output, &got); e != nil {
			tc.Errorf("model wire %v", e)
			return
		}
		if len(got.Refusals) != 8 {
			tc.Errorf("missing binding controls")
		}
		for i, refused := range got.Refusals {
			if !refused {
				tc.Errorf("malformed response binding control %d accepted", i)
			}
		}
		if !got.Immutable || !got.Stable || len(got.Models) != 3 {
			tc.Errorf("model mutated source/evidence")
		}
		for i, metric := range []string{"rate", "count", "share"} {
			expected := lpcExpected(request, metric)
			if got.Models[i] == nil || !reflect.DeepEqual(*got.Models[i], expected) {
				actual, _ := json.Marshal(got.Models[i])
				want, _ := json.Marshal(expected)
				tc.Errorf("independent exact model mode%d metric%s actual%s expected%s", mode, metric, actual, want)
			}
		}
		previous := request.Selection.Question
		previous.ToMS = previous.FromMS
		previous.FromMS -= int64(width)
		if !nhLiveSameQuery(previous, got.Previous) || !nhLiveSameQuery(request.Baseline.Question, got.Pair.Baseline) || !nhLiveSameQuery(request.Selection.Question, got.Pair.Selection) {
			tc.Errorf("source-free previous/planned query changed original scopes")
		}
	}, options(80)...)
}

func TestLogPatternComparisonFullTwoThousandUnion(t *testing.T) {
	alphabet := func(value int) string {
		out := ""
		for {
			out = string(rune('a'+value%26)) + out
			value = value/26 - 1
			if value < 0 {
				return out
			}
		}
	}
	request := exploration.ComparisonRequest{Baseline: lpcGeneratedCohort(1, 120000, nil), Selection: lpcGeneratedCohort(2, 60000, nil)}
	for i, c := range []*exploration.Cohort{&request.Baseline, &request.Selection} {
		c.State = "observed"
		c.Count = "1000"
		c.Groups = []exploration.ValueCount{}
		for j := 0; j < 1000; j++ {
			id := lpOracle([]string{"baseline ", "current "}[i] + alphabet(j)).Identity
			c.Groups = append(c.Groups, exploration.ValueCount{Value: &id, Count: "1"})
		}
	}
	response, e := exploration.CompareCohorts(request)
	if e != nil {
		t.Fatal(e)
	}
	input := struct {
		Request exploration.ComparisonRequest `json:"request"`
		Result  *exploration.CohortComparison `json:"result"`
	}{request, response}
	raw, _ := json.Marshal(input)
	cmd := exec.Command("node", "-e", `const fs=require('fs'),m=require('./ui/log-pattern-comparison.mjs'),x=JSON.parse(fs.readFileSync(0,'utf8'));console.log(JSON.stringify(m.model(x.request,x.result)));`)
	cmd.Dir = ".."
	cmd.Stdin = bytes.NewReader(raw)
	output, e := cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("actual full-union model %v %s", e, output)
	}
	var actual *lpcModel
	if e = json.Unmarshal(output, &actual); e != nil {
		t.Fatal(e)
	}
	if actual == nil || len(actual.Rows) != 2000 || !reflect.DeepEqual(*actual, lpcExpected(request, "rate")) {
		t.Fatal("per-side1000 bound incorrectly clipped full2000 union or arithmetic")
	}
}
