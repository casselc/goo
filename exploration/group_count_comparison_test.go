package exploration

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"testing"
)

func groupCountCohort(groups ...ValueCount) Cohort {
	total := new(big.Int)
	for _, group := range groups {
		n, _ := new(big.Int).SetString(group.Count, 10)
		total.Add(total, n)
	}
	state := "observed"
	if total.Sign() == 0 {
		state = "no_data"
	}
	q := capturedCountQuestion()
	q.GroupBy = []Field{{Kind: "attribute", Key: "cohort"}}
	return Cohort{Question: q, Basis: "held-groups", CapturedAt: "2026-10-31T00:00:00Z", Count: total.String(), State: state, Groups: groups}
}
func groupCountRowMap(result *GroupCountComparison) map[string]GroupCountRow {
	out := map[string]GroupCountRow{}
	for _, row := range result.Rows {
		out[row.Key] = row
	}
	return out
}
func TestGroupCountComparisonAlignsTypedIdentitiesAndKeepsOperands(t *testing.T) {
	empty, literal, shared, old, newer := "", "(missing)", "shared", "old", "new"
	a := groupCountCohort(ValueCount{Count: "2"}, ValueCount{Value: &empty, Count: "1"}, ValueCount{Value: &literal, Count: "1"}, ValueCount{Value: &shared, Count: "5"}, ValueCount{Value: &old, Count: "2"})
	b := groupCountCohort(ValueCount{Count: "1"}, ValueCount{Value: &empty, Count: "2"}, ValueCount{Value: &shared, Count: "7"}, ValueCount{Value: &newer, Count: "3"})
	beforeA, beforeB := detached(a), detached(b)
	result, err := CompareCohorts(ComparisonRequest{Baseline: a, Selection: b})
	if err != nil || result.GroupComparison == nil {
		t.Fatal("group comparison missing", err)
	}
	g := result.GroupComparison
	if g.Baseline != result.ViewComparison.Baseline || g.Selection != result.ViewComparison.Selection {
		t.Fatal("new summary operand identity")
	}
	if g.State != "observed" || g.BaselineAccounting != "coherent" || len(g.Rows) != 6 {
		t.Fatalf("wrong profile: %+v", g)
	}
	rows := groupCountRowMap(g)
	for key, want := range map[string]string{"0": "-1", "1": "1", "1(missing)": "-1", "1shared": "2", "1old": "-2", "1new": "3"} {
		row, ok := rows[key]
		if !ok || row.Change.State != "derived" || row.Change.Value == nil || *row.Change.Value != want || row.Change.Unit != "records" {
			t.Fatal("wrong exact group change", key, row)
		}
	}
	if rows["0"].Value != nil || rows["1"].Value == nil || *rows["1"].Value != "" || rows["1old"].Selection.Presence != "absent" || *rows["1old"].Selection.Value != "0" || rows["1new"].Baseline.Presence != "absent" {
		t.Fatal("category or operand absence conflated", rows)
	}
	sum := new(big.Int)
	for _, row := range g.Rows {
		v, _ := new(big.Int).SetString(*row.Change.Value, 10)
		sum.Add(sum, v)
	}
	if sum.String() != *result.ViewComparison.Changes["count"].Value {
		t.Fatal("group changes do not account for whole change")
	}
	if g.Baseline.Completeness != nil || g.Selection.Completeness != nil || g.Baseline.Quality.Extent != "unknown" || g.Selection.Origin != "browser-held" {
		t.Fatal("supplied accounting gained source witness")
	}
	if !reflect.DeepEqual(a, beforeA) || !reflect.DeepEqual(b, beforeB) {
		t.Fatal("source operands mutated")
	}
	*g.Rows[1].Value = "mutated"
	*g.Rows[1].Baseline.Value = "999"
	if !reflect.DeepEqual(a, beforeA) || !reflect.DeepEqual(b, beforeB) {
		t.Fatal("aligned row aliases input")
	}
}
func TestGroupCountComparisonZeroAndSignedUint64Extremes(t *testing.T) {
	label := "api"
	max := "18446744073709551615"
	for _, tc := range []struct {
		a, b Cohort
		want string
		rows int
	}{{groupCountCohort(), groupCountCohort(ValueCount{Value: &label, Count: max}), max, 1}, {groupCountCohort(ValueCount{Value: &label, Count: max}), groupCountCohort(), "-" + max, 1}, {groupCountCohort(), groupCountCohort(), "0", 0}} {
		result, err := CompareCohorts(ComparisonRequest{Baseline: tc.a, Selection: tc.b})
		if err != nil || result.GroupComparison == nil {
			t.Fatal(err)
		}
		g := result.GroupComparison
		if g.State != "no_data" || len(g.Rows) != tc.rows || *result.ViewComparison.Changes["count"].Value != tc.want {
			t.Fatal("zero population lost known count", g)
		}
		if tc.rows > 0 && (*g.Rows[0].Change.Value != tc.want || g.Rows[0].Change.State != "derived") {
			t.Fatal("signed uint64 range lost", g.Rows[0])
		}
		if raw, _ := json.Marshal(g); strings.Contains(string(raw), `"rows":null`) {
			t.Fatal("empty union must be supplied array")
		}
	}
}
func TestGroupCountComparisonDiagnosticDoesNotCompleteUnknownGroups(t *testing.T) {
	shared, old, newer := "shared", "old", "new"
	a := groupCountCohort(ValueCount{Value: &shared, Count: "2"}, ValueCount{Value: &old, Count: "1"})
	b := groupCountCohort(ValueCount{Value: &shared, Count: "7"}, ValueCount{Value: &newer, Count: "1"})
	b.State = "changed_between_reads"
	b.Count = "1"
	result, err := CompareCohorts(ComparisonRequest{Baseline: a, Selection: b})
	if err != nil || result.GroupComparison == nil {
		t.Fatal(err)
	}
	g := result.GroupComparison
	if g.State != "diagnostic" || g.SelectionAccounting != "diagnostic" || g.BaselineAccounting != "coherent" {
		t.Fatal(g)
	}
	rows := groupCountRowMap(g)
	for _, row := range g.Rows {
		if row.Selection.State != "unavailable" || row.Selection.Value != nil || row.Selection.Reason == "" || row.Change.State != "unavailable" || row.Change.Value != nil || row.Change.Reason == "" {
			t.Fatal("diagnostic gained arithmetic", row)
		}
	}
	if rows["1old"].Selection.Presence != "unknown" || rows["1shared"].Selection.Presence != "present" || rows["1new"].Baseline.Presence != "absent" || *rows["1new"].Baseline.Value != "0" {
		t.Fatal("independent membership lost", rows)
	}
}
func TestGroupCountComparisonFullUnionAndCategoryBounds(t *testing.T) {
	var a, b []ValueCount
	for i := range 1000 {
		av, bv := fmt.Sprintf("a%04d", i), fmt.Sprintf("b%04d", i)
		a = append(a, ValueCount{Value: &av, Count: "1"})
		b = append(b, ValueCount{Value: &bv, Count: "1"})
	}
	result, err := CompareCohorts(ComparisonRequest{Baseline: groupCountCohort(a...), Selection: groupCountCohort(b...)})
	if err != nil || result.GroupComparison == nil || len(result.GroupComparison.Rows) != 2000 {
		t.Fatal("full union truncated", err)
	}
	label := strings.Repeat("x", 4096)
	c := groupCountCohort(ValueCount{Value: &label, Count: "1"})
	result, err = CompareCohorts(ComparisonRequest{Baseline: c, Selection: c})
	if err != nil || result.GroupComparison == nil || len(result.GroupComparison.Rows[0].Key) != 4097 {
		t.Fatal("existing category bound narrowed", err)
	}
}
func TestGroupCountComparisonRemainsOptionalWithoutNarrowingBroadCompare(t *testing.T) {
	label := "api"
	a := groupCountCohort(ValueCount{Value: &label, Count: "2"})
	b := detached(a)
	b.Question.Where.Value = "different"
	result, err := CompareCohorts(ComparisonRequest{Baseline: a, Selection: b})
	if err != nil || result.ViewComparison == nil || result.GroupComparison != nil {
		t.Fatal("optional profile changed broad comparison", err)
	}
	malformed := detached(a)
	malformed.Groups = append(malformed.Groups, malformed.Groups[0])
	if _, err := CompareCohorts(ComparisonRequest{Baseline: a, Selection: malformed}); err == nil {
		t.Fatal("duplicate group admitted")
	}
	malformed = detached(a)
	malformed.Count = "3"
	if _, err := CompareCohorts(ComparisonRequest{Baseline: a, Selection: malformed}); err == nil {
		t.Fatal("incoherent claimed complete accounting admitted")
	}
}
