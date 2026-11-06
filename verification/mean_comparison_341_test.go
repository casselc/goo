package verification

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

// The oracle sums individual records, rather than averaging group means or
// trusting the comparison's rounded legacy meanDifferenceNs field.
func TestHegelPreviousWindowMean341(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		a := []uint64{0, uint64(seed), 9007199254740993}
		b := []uint64{uint64(seed + 1)}
		for range seed % 13 {
			b = append(b, 18446744073709551615)
		}
		cohort := func(values []uint64, from int64) exploration.Cohort {
			d := exploration.ReferenceDurationSummary(values)
			c := exploration.Cohort{Question: exploration.Query{Signal: "spans", Environment: "prod", Summary: "span-duration", FromMS: from, ToMS: from + 60000, Limit: 100, GroupBy: []exploration.Field{{Kind: "column", Key: "service"}}}, CapturedAt: fmt.Sprintf("2026-11-06T00:00:%02dZ", from/60000), Basis: "held-source", Count: d.Count, State: d.State, Duration: &d, Groups: []exploration.ValueCount{}}
			if len(values) > 0 {
				name := "service"
				c.Groups = append(c.Groups, exploration.ValueCount{Value: &name, Count: d.Count, Duration: &d})
			}
			return c
		}
		mean := func(values []uint64) *big.Rat {
			sum := new(big.Int)
			for _, v := range values {
				sum.Add(sum, new(big.Int).SetUint64(v))
			}
			return new(big.Rat).SetFrac(sum, big.NewInt(int64(len(values))))
		}
		left, right := cohort(a, 60000), cohort(b, 120000)
		compare := func(x, y exploration.Cohort) *exploration.ViewComparison {
			r, e := exploration.CompareCohorts(exploration.ComparisonRequest{Baseline: x, Selection: y})
			if e != nil || r.ViewComparison == nil {
				tc.Fatal(e, r)
				return nil
			}
			return r.ViewComparison
		}
		v := compare(left, right)
		if v == nil {
			return
		}
		want := new(big.Rat).Sub(mean(b), mean(a)).RatString()
		delta := v.Changes["duration.mean"]
		if delta.State != "derived" || delta.Value == nil || *delta.Value != want || delta.Unit != "ns" {
			tc.Fatalf("exact unequal-weight change: got %+v want %s", delta, want)
		}
		for i, side := range []*exploration.ViewDataset{v.Baseline, v.Selection} {
			values := [][]uint64{a, b}[i]
			if side.Rows[0].Values["duration.mean"] == nil || *side.Rows[0].Values["duration.mean"] != mean(values).RatString() || side.Rows[0].Values["count"] == nil || *side.Rows[0].Values["count"] != fmt.Sprint(len(values)) {
				tc.Fatal("mean/count authority lost", side)
			}
		}
		zero := compare(cohort([]uint64{0, 0}, 60000), cohort([]uint64{0}, 120000))
		if zero == nil {
			return
		}
		if zero.State != "observed" || zero.Changes["duration.mean"].Value == nil || *zero.Changes["duration.mean"].Value != "0" {
			tc.Fatal("measured zero lost", zero)
		}
		empty := compare(cohort(nil, 60000), right)
		if empty == nil {
			return
		}
		if empty.State != "no_data" || empty.Changes["duration.mean"].State != "unavailable" || empty.Changes["duration.mean"].Value != nil {
			tc.Fatal("empty became a mean", empty)
		}
		left.State = "changed_between_reads"
		diagnostic := compare(left, right)
		if diagnostic == nil {
			return
		}
		if diagnostic.State != "diagnostic" || diagnostic.Changes["duration.mean"].Value != nil {
			tc.Fatal("disagreement became a derived effect", diagnostic)
		}
	}, options(70)...)
}

// Large counts cannot be materialized as records. This coherent constant-valued
// population checks the wire boundary independently using integer products.
func TestPreviousWindowMeanHugeContributors341(t *testing.T) {
	makeCohort := func(count, value string, from int64) exploration.Cohort {
		n, _ := new(big.Int).SetString(count, 10)
		v, _ := new(big.Int).SetString(value, 10)
		d := exploration.DurationSummary{State: "observed", Count: count, SumNS: new(big.Int).Mul(n, v).String(), MinNS: value, MaxNS: value, P50NS: value, P95NS: value, P99NS: value}
		name := "api"
		return exploration.Cohort{Question: exploration.Query{Signal: "spans", Environment: "prod", Summary: "span-duration", FromMS: from, ToMS: from + 60000, Limit: 100, GroupBy: []exploration.Field{{Kind: "column", Key: "service"}}}, Basis: "aggregate-source", CapturedAt: "2026-11-06T00:00:01Z", Count: count, State: "observed", Duration: &d, Groups: []exploration.ValueCount{{Value: &name, Count: count, Duration: &d}}}
	}
	a := makeCohort("9007199254740993", "9007199254740993", 60000)
	b := makeCohort("18446744073709551615", "9007199254740994", 120000)
	r, e := exploration.CompareCohorts(exploration.ComparisonRequest{Baseline: a, Selection: b})
	if e != nil || r.ViewComparison == nil {
		t.Fatal(e, r)
	}
	v := r.ViewComparison
	if got := v.Changes["duration.mean"]; got.Value == nil || *got.Value != "1" {
		t.Fatal("large contributor count rounded", got)
	}
	if *v.Baseline.Rows[0].Values["count"] != a.Count || *v.Selection.Rows[0].Values["count"] != b.Count {
		t.Fatal("contributing counts lost", v)
	}
	// Missing measure authority must be refused, never guessed from count/records.
	a.Duration = nil
	if _, e = exploration.CompareCohorts(exploration.ComparisonRequest{Baseline: a, Selection: b}); e == nil {
		t.Fatal("missing duration accepted")
	}
}
