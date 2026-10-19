package verification

import (
	"encoding/json"
	"math/big"
	"strconv"
	"strings"
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

type numericMachine struct {
	q           exploration.Query
	values      []*string
	frozen      []exploration.NumericSummary
	frozenBytes []string
}

func (m *numericMachine) RuleAppend(tc hegel.TestCase) {
	choices := []*string{nil}
	for _, v := range []string{"", "-9", "0", "11", "01", "NaN", "9223372036854775807", "18446744073709551615"} {
		s := v
		choices = append(choices, &s)
	}
	m.values = append(m.values, choices[hegel.Draw(tc, hegel.Integers(0, len(choices)-1))])
}
func (m *numericMachine) RuleType(tc hegel.TestCase) {
	m.q.Numeric.Type = []string{"int64", "uint64"}[hegel.Draw(tc, hegel.Integers(0, 1))]
}
func (m *numericMachine) RulePage(tc hegel.TestCase) {
	m.q.Limit = hegel.Draw(tc, hegel.Integers(1, 100))
	m.q.Offset = hegel.Draw(tc, hegel.Integers(0, 1000))
}
func (m *numericMachine) RuleFreeze(tc hegel.TestCase) {
	s, e := exploration.EvaluateNumericValues(*m.q.Numeric, m.values)
	if e != nil {
		panic(e)
	}
	m.frozen = append(m.frozen, s)
	b, _ := json.Marshal(s)
	m.frozenBytes = append(m.frozenBytes, string(b))
}
func (m *numericMachine) InvariantExactCoverageAndFrozen(tc hegel.TestCase) {
	s, e := exploration.EvaluateNumericValues(*m.q.Numeric, m.values)
	if e != nil || s.Validate(*m.q.Numeric) != nil {
		panic("invalid reference state")
	}
	total := new(big.Int)
	for _, v := range []string{s.Valid, s.Missing, s.Empty, s.Invalid} {
		n, _ := new(big.Int).SetString(v, 10)
		total.Add(total, n)
	}
	if total.String() != strconv.Itoa(len(m.values)) {
		panic("coverage loss")
	}
	p, e := exploration.Compile(m.q)
	if e != nil || strings.Contains(p.Count, " LIMIT ") {
		panic("paging changed measurement")
	}
	plan, e := m.q.RecordMeasures()
	if e != nil || plan.Numeric == nil || *plan.Numeric != *m.q.Numeric {
		panic("binding lost")
	}
	for i, s := range m.frozen {
		b, _ := json.Marshal(s)
		if string(b) != m.frozenBytes[i] {
			panic("frozen evidence changed")
		}
	}
}
func TestHegelNumericBindingStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		m := &numericMachine{q: exploration.Query{Signal: "spans", Environment: "test", FromMS: 1000, ToMS: 61000, Limit: 100, Numeric: &exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "custom.value"}, Type: "int64", Unit: "1", Parse: "canonical-decimal"}}}
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(25), hegel.WithAlwaysCheckInvariants("InvariantExactCoverageAndFrozen"))
	}, options(100)...)
}
