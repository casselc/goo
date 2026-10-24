package verification

import (
	"encoding/json"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

// Independent fixed-scale grammar: inspect ASCII digits and integer/fraction
// boundaries before standard-library integer parsing. No production parser,
// Decimal/float conversion or production formatter predicts an expected value.
func decCoefficient(raw *string, scale int) (*big.Int, string) {
	if raw == nil {
		return nil, "missing"
	}
	text := *raw
	if text == "" {
		return nil, "empty"
	}
	if scale < 1 || scale > 9 {
		return nil, "invalid"
	}
	negative := false
	if text[0] == '-' {
		negative = true
		text = text[1:]
	}
	dot := strings.IndexByte(text, '.')
	if dot < 1 || strings.LastIndexByte(text, '.') != dot || len(text)-dot-1 != scale {
		return nil, "invalid"
	}
	integer, fraction := text[:dot], text[dot+1:]
	if len(integer) > 1 && integer[0] == '0' {
		return nil, "invalid"
	}
	for _, part := range []string{integer, fraction} {
		for i := 0; i < len(part); i++ {
			if part[i] < '0' || part[i] > '9' {
				return nil, "invalid"
			}
		}
	}
	n, ok := new(big.Int).SetString(integer+fraction, 10)
	if !ok {
		return nil, "invalid"
	}
	if negative {
		if n.Sign() == 0 {
			return nil, "invalid"
		}
		n.Neg(n)
	}
	if !n.IsInt64() {
		return nil, "invalid"
	}
	return n, "valid"
}
func decText(coefficient string, scale int) string {
	n, ok := new(big.Int).SetString(coefficient, 10)
	if !ok || scale < 1 || scale > 9 {
		panic("invalid independent decimal formatting input")
	}
	negative := n.Sign() < 0
	digits := new(big.Int).Abs(n).String()
	if len(digits) <= scale {
		digits = strings.Repeat("0", scale+1-len(digits)) + digits
	}
	at := len(digits) - scale
	out := digits[:at] + "." + digits[at:]
	if negative {
		out = "-" + out
	}
	return out
}
func decBinding(scale int) exploration.NumericBinding {
	return exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "charge.amount"}, Type: "decimal64", Scale: scale, Unit: "USD", Parse: "canonical-decimal"}
}
func decSummary(b exploration.NumericBinding, values []*string) exploration.NumericSummary {
	s := exploration.NumericSummary{State: "no_valid_values", Unit: b.Unit, Scale: b.Scale, Count: strconv.Itoa(len(values)), Sum: "0"}
	counts := map[string]int{}
	sum := new(big.Int)
	var lo, hi *big.Int
	for _, raw := range values {
		n, state := decCoefficient(raw, b.Scale)
		counts[state]++
		if state != "valid" {
			continue
		}
		sum.Add(sum, n)
		if lo == nil || n.Cmp(lo) < 0 {
			lo = new(big.Int).Set(n)
		}
		if hi == nil || n.Cmp(hi) > 0 {
			hi = new(big.Int).Set(n)
		}
	}
	s.Valid = strconv.Itoa(counts["valid"])
	s.Missing = strconv.Itoa(counts["missing"])
	s.Empty = strconv.Itoa(counts["empty"])
	s.Invalid = strconv.Itoa(counts["invalid"])
	s.Sum = sum.String()
	if len(values) == 0 {
		s.State = "no_data"
	}
	if lo != nil {
		s.State = "observed"
		s.Min = lo.String()
		s.Max = hi.String()
	}
	return s
}
func decCorpus(scale int) []*string {
	values := []*string{nil}
	add := func(raw string) { s := raw; values = append(values, &s) }
	for _, coefficient := range []string{"0", "1", "-1", "10", "-10", "100", "-100", "9007199254740991", "9007199254740992", "9007199254740993", "-9007199254740993", "9223372036854775807", "-9223372036854775808", "9223372036854775808", "-9223372036854775809"} {
		add(decText(coefficient, scale))
	}
	zero := strings.Repeat("0", scale)
	for _, raw := range []string{"", "0", "1", "-1", "+1." + zero, "-0." + zero, "00." + zero, "-00." + zero, "01." + zero, ".0", "1.", "1." + zero + "0", "1." + zero + "1", "1." + zero + "e0", "1e0." + zero, " 1." + zero, "1." + zero + " ", "1." + zero + "\n", "1," + zero, "NaN", "Infinity", "1_0." + zero, "0x1." + zero, "١." + zero, "1.." + zero, strings.Repeat("9", 100) + "." + zero} {
		add(raw)
	}
	if scale > 1 {
		add("1." + strings.Repeat("0", scale-1))
	}
	return values
}
func decHistogram(b exploration.NumericBinding, values []*string) exploration.NumericHistogram {
	edges := histogramOracleEdges("int64")
	bins := make([]exploration.NumericHistogramBin, len(edges)-1)
	counts := make([]int, len(bins))
	for _, raw := range values {
		n, state := decCoefficient(raw, b.Scale)
		if state != "valid" {
			continue
		}
		found := false
		for i := range bins {
			lo, _ := new(big.Int).SetString(edges[i], 10)
			hi, _ := new(big.Int).SetString(edges[i+1], 10)
			if n.Cmp(lo) >= 0 && n.Cmp(hi) < 0 {
				counts[i]++
				found = true
				break
			}
		}
		if !found {
			panic("valid coefficient outside signed64 bins")
		}
	}
	for i := range bins {
		bins[i] = exploration.NumericHistogramBin{From: edges[i], To: edges[i+1], Count: strconv.Itoa(counts[i])}
	}
	return exploration.NumericHistogram{Binning: "powers-of-two", Bins: bins}
}
func decBoundaryCorpus(scale int) []*string {
	values := decCorpus(scale)
	for _, edge := range histogramOracleEdges("int64") {
		n, _ := new(big.Int).SetString(edge, 10)
		for _, delta := range []int64{-1, 0, 1} {
			raw := decText(new(big.Int).Add(n, big.NewInt(delta)).String(), scale)
			values = append(values, &raw)
		}
	}
	return values
}
func TestHegelDecimal64CoverageAndPartition(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		scale := hegel.Draw(tc, hegel.Integers(1, 9))
		seed := hegel.Draw(tc, hegel.Integers(-1000000, 1000000))
		b := decBinding(scale)
		values := decCorpus(scale)
		for i := 0; i < 10; i++ {
			raw := decText(strconv.FormatInt(int64(seed)*(int64(i)+1), 10), scale)
			values = append(values, &raw)
		}
		before, _ := json.Marshal(values)
		expected := decSummary(b, values)
		actual, e := exploration.EvaluateNumericValues(b, values)
		if e != nil || !reflect.DeepEqual(actual, expected) || actual.Validate(b) != nil {
			tc.Fatal("decimal coefficient/coverage mismatch", scale, e, actual, expected)
		}
		cut := hegel.Draw(tc, hegel.Integers(0, len(values)))
		a, e := exploration.EvaluateNumericValues(b, values[:cut])
		if e != nil {
			tc.Fatal(e)
		}
		z, e := exploration.EvaluateNumericValues(b, values[cut:])
		if e != nil {
			tc.Fatal(e)
		}
		if !reflect.DeepEqual(a, decSummary(b, values[:cut])) || !reflect.DeepEqual(z, decSummary(b, values[cut:])) {
			tc.Fatal("partition oracle mismatch")
		}
		for _, triple := range [][3]string{{actual.Count, a.Count, z.Count}, {actual.Valid, a.Valid, z.Valid}, {actual.Missing, a.Missing, z.Missing}, {actual.Empty, a.Empty, z.Empty}, {actual.Invalid, a.Invalid, z.Invalid}, {actual.Sum, a.Sum, z.Sum}} {
			x, _ := new(big.Int).SetString(triple[1], 10)
			y, _ := new(big.Int).SetString(triple[2], 10)
			if x.Add(x, y).String() != triple[0] {
				tc.Fatal("decimal partition lost coefficient or coverage")
			}
		}
		wrong := actual
		wrong.Scale = scale%9 + 1
		if wrong.Validate(b) == nil {
			tc.Fatal("summary scale mismatch accepted")
		}
		if after, _ := json.Marshal(values); string(before) != string(after) {
			tc.Fatal("numeric evaluation mutated exporter text")
		}
		for _, badScale := range []int{-1, 0, 10} {
			invalid := b
			invalid.Scale = badScale
			if _, e := exploration.EvaluateNumericValues(invalid, values); e == nil {
				tc.Fatal("invalid decimal scale accepted", badScale)
			}
		}
	}, options(90)...)
}
func TestHegelDecimal64HistogramAndRangePartition(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		scale := hegel.Draw(tc, hegel.Integers(1, 9))
		b := decBinding(scale)
		values := decBoundaryCorpus(scale)
		expected := decHistogram(b, values)
		actual, e := exploration.EvaluateNumericHistogram(b, values)
		if e != nil || !reflect.DeepEqual(actual, expected) {
			tc.Fatal("decimal histogram not coefficient-based", e)
		}
		from := hegel.Draw(tc, hegel.Integers(0, len(expected.Bins)-1))
		to := hegel.Draw(tc, hegel.Integers(from+1, len(expected.Bins)))
		parent := exploration.Query{Signal: "spans", Environment: "decimal-property", FromMS: 1000, ToMS: 61000, Limit: 1, Offset: 17, Numeric: &b, NumericHistogram: &exploration.NumericHistogramSpec{Binning: "powers-of-two"}, Where: &exploration.Predicate{Field: &exploration.Field{Kind: "resource", Key: "team"}, Op: "eq", Value: "core"}}
		frozen, _ := json.Marshal(parent)
		plan, e := exploration.PlanNumericHistogramSelection(parent, from, to)
		if e != nil {
			tc.Fatal(e)
		}
		lo, _ := new(big.Int).SetString(expected.Bins[from].From, 10)
		hi, _ := new(big.Int).SetString(expected.Bins[to-1].To, 10)
		if plan.From != lo.String() || plan.To != hi.String() || plan.FromBin != from || plan.ToBin != to || !nhLiveSameQuery(plan.Parent, parent) {
			tc.Fatal("range geometry/scope changed")
		}
		for _, raw := range values {
			n, state := decCoefficient(raw, scale)
			inside := state == "valid" && n.Cmp(lo) >= 0 && n.Cmp(hi) < 0
			for _, team := range []string{"core", "outside"} {
				fields := map[exploration.Field]string{{Kind: "resource", Key: "team"}: team}
				if raw != nil {
					fields[b.Field] = *raw
				}
				selected := exploration.Match(plan.Selected.Where, fields)
				rest := exploration.Match(plan.Rest.Where, fields)
				if selected != (team == "core" && inside) || rest != (team == "core" && !inside) || selected && rest {
					tc.Fatal("decimal selected/rest partition mismatch", scale, raw, team)
				}
			}
		}
		// Every finite typed predicate uses human fixed-scale text; result edges stay
		// integer coefficients. Domain sentinel boundaries must remain omitted.
		var inspect func(*exploration.Predicate)
		inspect = func(p *exploration.Predicate) {
			if p == nil {
				return
			}
			if p.Numeric != nil && (p.Op == "gte" || p.Op == "lt" || p.Op == "lte" || p.Op == "gt") {
				n, state := decCoefficient(&p.Value, scale)
				if state != "valid" || p.Numeric.Scale != scale || n == nil {
					tc.Fatal("range planner emitted coefficient text as human threshold", p.Value)
				}
			}
			for i := range p.All {
				inspect(&p.All[i])
			}
			for i := range p.Any {
				inspect(&p.Any[i])
			}
		}
		inspect(plan.Selected.Where)
		inspect(plan.Rest.Where)
		for _, q := range []exploration.Query{plan.Selected, plan.Rest} {
			if q.Numeric == nil || *q.Numeric != b || q.FromMS != parent.FromMS || q.ToMS != parent.ToMS || q.Environment != parent.Environment {
				tc.Fatal("decimal range lost binding or fixed scope")
			}
			compiled, e := exploration.Compile(q)
			if e != nil || strings.Contains(strings.ToLower(compiled.Count), "tofloat") || strings.Contains(strings.ToUpper(compiled.Count), " LIMIT ") {
				tc.Fatal("decimal range compile failed or used floating evaluation", e)
			}
		}
		if after, _ := json.Marshal(parent); string(after) != string(frozen) {
			tc.Fatal("range planner mutated original question")
		}
	}, options(70)...)
}
func decCohort(b exploration.NumericBinding, values []*string, histogram bool) exploration.Cohort {
	s := decSummary(b, values)
	label := "worker"
	c := exploration.Cohort{Question: exploration.Query{Signal: "spans", Environment: "decimal-property", FromMS: 1000, ToMS: 61000, Limit: 100, Numeric: &b, GroupBy: []exploration.Field{{Kind: "column", Key: "service"}}}, Basis: "independent-decimal-facts", State: "observed", Count: strconv.Itoa(len(values)), Numeric: &s, Groups: []exploration.ValueCount{}}
	if len(values) == 0 {
		c.State = "no_data"
	} else {
		c.Groups = append(c.Groups, exploration.ValueCount{Value: &label, Count: c.Count, Numeric: &s})
	}
	if histogram {
		h := decHistogram(b, values)
		c.NumericHistogram = &h
		c.Question.NumericHistogram = &exploration.NumericHistogramSpec{Binning: "powers-of-two"}
	}
	return c
}
func TestHegelDecimal64ComparisonScaleAndIntegerWire(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		scale := hegel.Draw(tc, hegel.Integers(1, 9))
		b := decBinding(scale)
		values := decCorpus(scale)
		cut := hegel.Draw(tc, hegel.Integers(0, len(values)))
		a, z := decCohort(b, values[:cut], true), decCohort(b, values[cut:], true)
		for _, unknown := range []bool{false, true} {
			request := exploration.ComparisonRequest{Numeric: true, Distribution: true, Baseline: a, Selection: z}
			if unknown {
				request.Baseline.State = "changed_between_reads"
			}
			before, _ := json.Marshal(request)
			result, e := exploration.CompareCohorts(request)
			if e != nil {
				tc.Fatal(e)
			}
			if result.Measures == nil || result.Measures.Numeric == nil {
				tc.Fatal("decimal comparison missing")
			}
			if e := checkNumericChanges(result.Measures.Numeric, b, a.Numeric, z.Numeric, unknown); e != nil {
				tc.Fatal(e)
			}
			expected := dcLiveDistribution(request.Baseline, request.Selection)
			if result.Measures.Numeric.Distribution == nil || !reflect.DeepEqual(*result.Measures.Numeric.Distribution, expected) {
				tc.Fatal("decimal distribution changed coefficient bins or all-record shares")
			}
			if after, _ := json.Marshal(request); string(after) != string(before) {
				tc.Fatal("comparison mutated decimal evidence")
			}
		}
		otherScale := scale%9 + 1
		otherBinding := decBinding(otherScale)
		mismatch := exploration.ComparisonRequest{Numeric: true, Baseline: a, Selection: decCohort(otherBinding, decCorpus(otherScale), true)}
		if _, e := exploration.CompareCohorts(mismatch); e == nil {
			tc.Fatal("different decimal scales compared without refusal")
		}
		for _, typ := range []string{"int64", "uint64"} {
			integer := predicateBinding(typ, "attribute")
			one := "1"
			s, e := exploration.EvaluateNumericValues(integer, []*string{&one})
			if e != nil {
				tc.Fatal(e)
			}
			for _, v := range []any{integer, s} {
				raw, _ := json.Marshal(v)
				var wire map[string]json.RawMessage
				json.Unmarshal(raw, &wire)
				if _, present := wire["scale"]; present {
					tc.Fatal("integer wire acquired scale metadata")
				}
			}
			integer.Scale = scale
			if _, e := exploration.EvaluateNumericValues(integer, nil); e == nil {
				tc.Fatal("integer accepted decimal scale")
			}
		}
	}, options(70)...)
}

type decMachine struct {
	Binding exploration.NumericBinding
	Values  []*string
	Frozen  []struct {
		Binding exploration.NumericBinding
		Values  []*string
		Summary exploration.NumericSummary
		Wire    string
	}
}

func (m *decMachine) RuleAppend(tc hegel.TestCase) {
	corpus := decCorpus(m.Binding.Scale)
	m.Values = append(m.Values, corpus[hegel.Draw(tc, hegel.Integers(0, len(corpus)-1))])
}
func (m *decMachine) RuleScale(tc hegel.TestCase) {
	m.Binding.Scale = hegel.Draw(tc, hegel.Integers(1, 9))
}
func (m *decMachine) RuleReplace(tc hegel.TestCase) {
	if len(m.Values) == 0 {
		return
	}
	i := hegel.Draw(tc, hegel.Integers(0, len(m.Values)-1))
	corpus := decCorpus(m.Binding.Scale)
	m.Values[i] = corpus[hegel.Draw(tc, hegel.Integers(0, len(corpus)-1))]
}
func (m *decMachine) RuleFreeze(tc hegel.TestCase) {
	s, e := exploration.EvaluateNumericValues(m.Binding, m.Values)
	if e != nil {
		tc.Errorf("decimal freeze failed: %v", e)
	}
	raw, _ := json.Marshal(s)
	m.Frozen = append(m.Frozen, struct {
		Binding exploration.NumericBinding
		Values  []*string
		Summary exploration.NumericSummary
		Wire    string
	}{m.Binding, comparisonClone(m.Values), s, string(raw)})
}
func (m *decMachine) InvariantCoverageAndFrozenScale(tc hegel.TestCase) {
	s, e := exploration.EvaluateNumericValues(m.Binding, m.Values)
	if e != nil || !reflect.DeepEqual(s, decSummary(m.Binding, m.Values)) {
		tc.Errorf("scale-change coverage/coefficient drift scale=%d: %v", m.Binding.Scale, e)
	}
	for _, saved := range m.Frozen {
		raw, _ := json.Marshal(saved.Summary)
		if string(raw) != saved.Wire || !reflect.DeepEqual(saved.Summary, decSummary(saved.Binding, saved.Values)) {
			tc.Errorf("frozen decimal observation reinterpreted at another scale=%d", saved.Binding.Scale)
		}
	}
}
func TestHegelDecimal64ScaleStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		m := &decMachine{Binding: decBinding(hegel.Draw(tc, hegel.Integers(1, 9)))}
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(30), hegel.WithAlwaysCheckInvariants("InvariantCoverageAndFrozenScale"))
	}, options(70)...)
}

func TestDecimal64FixedScaleBoundaryContract(t *testing.T) {
	cases := []struct {
		Scale            int
		Raw, Coefficient string
	}{
		{1, "0.0", "0"}, {1, "-0.0", ""}, {1, "0.1", "1"}, {2, "0.01", "1"}, {3, "0.001", "1"}, {4, "0.0001", "1"}, {5, "0.00001", "1"}, {6, "0.000001", "1"}, {7, "0.0000001", "1"}, {8, "0.00000001", "1"}, {9, "0.000000001", "1"},
		{2, "90071992547409.93", "9007199254740993"}, {2, "-90071992547409.93", "-9007199254740993"}, {2, "92233720368547758.07", "9223372036854775807"}, {2, "-92233720368547758.08", "-9223372036854775808"},
		{9, "9223372036.854775807", "9223372036854775807"}, {9, "-9223372036.854775808", "-9223372036854775808"}, {9, "9223372036.854775808", ""}, {9, "-9223372036.854775809", ""},
		{2, "1.00", "100"}, {2, "1.0", ""}, {2, "1.000", ""}, {2, "-0.01", "-1"}, {2, "-0.00", ""}, {2, "-00.01", ""}, {2, "00.00", ""}, {2, "+1.00", ""}, {2, "1.00 ", ""}, {2, "1e0", ""}, {9, "0.1", ""},
	}
	for _, item := range cases {
		b := decBinding(item.Scale)
		raw := item.Raw
		s, e := exploration.EvaluateNumericValues(b, []*string{&raw})
		if e != nil {
			t.Fatal(e)
		}
		n, state := decCoefficient(&raw, item.Scale)
		if item.Coefficient == "" {
			if state != "invalid" || s.Valid != "0" || s.Invalid != "1" || s.Sum != "0" || s.Min != "" || s.Max != "" || s.Scale != item.Scale {
				t.Fatalf("invalid fixed-scale literal accepted scale%d %q", item.Scale, item.Raw)
			}
		} else if state != "valid" || n.String() != item.Coefficient || s.Valid != "1" || s.Invalid != "0" || s.Sum != item.Coefficient || s.Min != item.Coefficient || s.Max != item.Coefficient || s.Scale != item.Scale {
			t.Fatalf("coefficient boundary mismatch scale%d %q %+v", item.Scale, item.Raw, s)
		}
		p := exploration.Predicate{Numeric: &b, Op: "gte", Value: raw}
		q := predicateQuery(p)
		_, e = exploration.Compile(q)
		if (item.Coefficient == "") != (e != nil) {
			t.Fatalf("threshold/raw parsing disagrees scale%d %q: %v", item.Scale, item.Raw, e)
		}
	}
}
