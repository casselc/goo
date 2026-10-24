package exploration

import (
	"reflect"
	"strings"
	"testing"
)

func str(s string) *string { return &s }
func cohort(total string, groups ...ValueCount) Cohort {
	q := basic()
	q.GroupBy = []Field{{"attribute", "region"}}
	return Cohort{Question: q, Basis: strings.Repeat("a", 64), Count: total, State: "observed", Groups: groups}
}
func TestCohortComparisonExactCountsAndMissing(t *testing.T) {
	a := cohort("8", ValueCount{Value: str("a"), Count: "4"}, ValueCount{Value: nil, Count: "2"}, ValueCount{Value: str(""), Count: "2"})
	b := cohort("4", ValueCount{Value: str("a"), Count: "3"}, ValueCount{Value: str(""), Count: "1"})
	r, e := CompareCohorts(ComparisonRequest{Baseline: a, Selection: b})
	if e != nil {
		t.Fatal(e)
	}
	if r.State != "ok" || r.Rows[0].Value != nil && *r.Rows[0].Value != "a" {
		t.Fatal(r)
	}
	var positive, missing, empty bool
	for _, row := range r.Rows {
		if row.Value == nil {
			missing = row.SelectionCount == "0" && row.DifferencePP == "-25.00"
		} else if *row.Value == "a" {
			positive = row.BaselinePercent == "50.00" && row.SelectionPercent == "75.00" && row.DifferencePP == "25.00"
		} else if *row.Value == "" {
			empty = row.DifferencePP == "0.00"
		}
	}
	if !positive || !missing || !empty {
		t.Fatal(r)
	}
	a.Groups[0], a.Groups[2] = a.Groups[2], a.Groups[0]
	again, e := CompareCohorts(ComparisonRequest{Baseline: a, Selection: b})
	if e != nil || r.Basis != again.Basis || !reflect.DeepEqual(r.Rows, again.Rows) {
		t.Fatal("ordering changes comparison", e)
	}
}
func TestCohortComparisonUnknownEmptyAndInvalid(t *testing.T) {
	a := cohort("2", ValueCount{Value: str("x"), Count: "2"})
	b := a
	b.Question.ToMS++
	if _, e := CompareCohorts(ComparisonRequest{Baseline: a, Selection: b}); e != nil {
		t.Fatal("different fixed windows rejected")
	}
	b = a
	b.Groups = []ValueCount{{Value: str("x"), Count: "1"}}
	if _, e := CompareCohorts(ComparisonRequest{Baseline: a, Selection: b}); e == nil {
		t.Fatal("partial groups accepted")
	}
	b.State = "changed_between_reads"
	r, e := CompareCohorts(ComparisonRequest{Baseline: a, Selection: b})
	if e != nil || r.State != "unknown" || len(r.Rows) != 0 {
		t.Fatal(r, e)
	}
	b = cohort("0")
	b.State = "no_data"
	r, e = CompareCohorts(ComparisonRequest{Baseline: a, Selection: b})
	if e != nil || r.State != "no_data" || len(r.Rows) != 0 {
		t.Fatal(r, e)
	}
	b = cohort("2", ValueCount{Value: nil, Count: "1"}, ValueCount{Value: nil, Count: "1"})
	if _, e := CompareCohorts(ComparisonRequest{Baseline: a, Selection: b}); e == nil {
		t.Fatal("duplicate missing bucket accepted")
	}
	a = cohort("18446744073709551615", ValueCount{Value: str("x"), Count: "18446744073709551614"}, ValueCount{Value: str("y"), Count: "1"})
	r, e = CompareCohorts(ComparisonRequest{Baseline: a, Selection: a})
	if e != nil || r.Rows[0].DifferencePP != "0.00" {
		t.Fatal("lost uint64 precision", e)
	}
}
