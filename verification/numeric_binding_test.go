package verification

import (
	"math/big"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

// Partition conservation is independent of input order and page boundaries.
// Invalid/missing records must never affect the valid-value mean denominator.
func TestHegelNumericBindingPartitionConservation(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		b := exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "custom.value"}, Type: "int64", Unit: "1", Parse: "canonical-decimal"}
		n := hegel.Draw(tc, hegel.Integers(0, 80))
		values := make([]*string, 0, n)
		want := new(big.Int)
		valid := 0
		for i := 0; i < n; i++ {
			kind := hegel.Draw(tc, hegel.Integers(0, 4))
			if kind == 0 {
				values = append(values, nil)
				continue
			}
			s := ""
			if kind == 1 {
				s = "broken"
			} else if kind >= 3 {
				v := hegel.Draw(tc, hegel.Integers(-1000000, 1000000))
				s = strconv.Itoa(v)
				want.Add(want, big.NewInt(int64(v)))
				valid++
			}
			values = append(values, &s)
		}
		cut := hegel.Draw(tc, hegel.Integers(0, n))
		all, e := exploration.EvaluateNumericValues(b, values)
		left, e1 := exploration.EvaluateNumericValues(b, values[:cut])
		right, e2 := exploration.EvaluateNumericValues(b, values[cut:])
		if e != nil || e1 != nil || e2 != nil || all.Sum != want.String() || all.Valid != strconv.Itoa(valid) {
			t.Fatal(all, e, e1, e2)
		}
		for _, pair := range [][3]string{{all.Count, left.Count, right.Count}, {all.Valid, left.Valid, right.Valid}, {all.Missing, left.Missing, right.Missing}, {all.Empty, left.Empty, right.Empty}, {all.Invalid, left.Invalid, right.Invalid}, {all.Sum, left.Sum, right.Sum}} {
			a, _ := new(big.Int).SetString(pair[1], 10)
			c, _ := new(big.Int).SetString(pair[2], 10)
			if a.Add(a, c).String() != pair[0] {
				t.Fatal("partition loss", pair)
			}
		}
	}, options(100)...)
}
