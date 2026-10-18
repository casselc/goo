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
	a := cohort("8", ValueCount{str("a"), "4"}, ValueCount{nil, "2"}, ValueCount{str(""), "2"})
	b := cohort("4", ValueCount{str("a"), "3"}, ValueCount{str(""), "1"})
	r, e := CompareCohorts(ComparisonRequest{a, b})
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
	again, e := CompareCohorts(ComparisonRequest{a, b})
	if e != nil || r.Basis != again.Basis || !reflect.DeepEqual(r.Rows, again.Rows) {
		t.Fatal("ordering changes comparison", e)
	}
}
func TestCohortComparisonUnknownEmptyAndInvalid(t *testing.T) {
	a := cohort("2", ValueCount{str("x"), "2"})
	b := a
	b.Question.ToMS++
	if _, e := CompareCohorts(ComparisonRequest{a, b}); e == nil {
		t.Fatal("window drift accepted")
	}
	b = a
	b.Groups = []ValueCount{{str("x"), "1"}}
	if _, e := CompareCohorts(ComparisonRequest{a, b}); e == nil {
		t.Fatal("partial groups accepted")
	}
	b.State = "changed_between_reads"
	r, e := CompareCohorts(ComparisonRequest{a, b})
	if e != nil || r.State != "unknown" || len(r.Rows) != 0 {
		t.Fatal(r, e)
	}
	b = cohort("0")
	b.State = "no_data"
	r, e = CompareCohorts(ComparisonRequest{a, b})
	if e != nil || r.State != "no_data" || len(r.Rows) != 0 {
		t.Fatal(r, e)
	}
	b = cohort("2", ValueCount{nil, "1"}, ValueCount{nil, "1"})
	if _, e := CompareCohorts(ComparisonRequest{a, b}); e == nil {
		t.Fatal("duplicate missing bucket accepted")
	}
	a = cohort("18446744073709551615", ValueCount{str("x"), "18446744073709551614"}, ValueCount{str("y"), "1"})
	r, e = CompareCohorts(ComparisonRequest{a, a})
	if e != nil || r.Rows[0].DifferencePP != "0.00" {
		t.Fatal("lost uint64 precision", e)
	}
}
