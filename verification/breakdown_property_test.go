package verification

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
	"reflect"
	"testing"
)

func TestHegelBreakdownPartitions(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		x := hegel.Draw(tc, hegel.Integers(1, 1000000))
		y := hegel.Draw(tc, hegel.Integers(1, 1000000))
		q := exploration.Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 100, GroupBy: []exploration.Field{{Kind: "attribute", Key: "category"}}}
		a := exploration.Cohort{Question: q, Basis: "a", State: "observed", Count: fmt.Sprint(x + y), Groups: []exploration.ValueCount{{Count: fmt.Sprint(x)}, {Value: ptrBreakdown(""), Count: fmt.Sprint(y)}}}
		b := a
		b.Groups = []exploration.ValueCount{a.Groups[1], a.Groups[0]}
		before, _ := json.Marshal(a)
		r, e := exploration.CompareCohorts(exploration.ComparisonRequest{Baseline: a, Selection: b})
		after, _ := json.Marshal(a)
		if e != nil || len(r.Rows) != 2 || !reflect.DeepEqual(before, after) {
			tc.Fatal(r, e)
		}
		for _, row := range r.Rows {
			if row.DifferencePP != "0.00" || row.BaselineCount != row.SelectionCount {
				tc.Fatal(row)
			}
		}
		b.Groups[0].Count = fmt.Sprint(y + 1)
		if _, e = exploration.CompareCohorts(exploration.ComparisonRequest{Baseline: a, Selection: b}); e == nil {
			tc.Fatal("broken partition accepted")
		}
	}, options(100)...)
}
func ptrBreakdown(s string) *string { return &s }
