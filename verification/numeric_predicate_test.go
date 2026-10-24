package verification

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

var predicateRawCorpus = []*string{nil}

func init() {
	for _, raw := range []string{"", "0", "-0", "+0", "+1", "01", "-01", "1", "-1", "7", "-7", "9007199254740993", "9007199254740992", "9223372036854775807", "9223372036854775808", "-9223372036854775808", "-9223372036854775809", "18446744073709551615", "18446744073709551616", "1e3", "1.0", "NaN", " 1", "1\n", "雪", strings.Repeat("9", 100)} {
		value := raw
		predicateRawCorpus = append(predicateRawCorpus, &value)
	}
}

// Independent oracle: standard library parsing plus decimal roundtrip, not the
// production numeric parser or production summary used to predict its result.
func predicateState(raw *string, typ string) string {
	if raw == nil {
		return "missing"
	}
	if *raw == "" {
		return "empty"
	}
	if typ == "int64" {
		n, e := strconv.ParseInt(*raw, 10, 64)
		if e == nil && strconv.FormatInt(n, 10) == *raw {
			return "valid"
		}
	} else {
		n, e := strconv.ParseUint(*raw, 10, 64)
		if e == nil && strconv.FormatUint(n, 10) == *raw {
			return "valid"
		}
	}
	return "invalid"
}
func predicateOracle(raw *string, b exploration.NumericBinding, op, value string) bool {
	state := predicateState(raw, b.Type)
	switch op {
	case "valid", "missing", "empty", "invalid":
		return state == op
	}
	if state != "valid" {
		return false
	}
	cmp := 0
	if b.Type == "int64" {
		a, _ := strconv.ParseInt(*raw, 10, 64)
		v, _ := strconv.ParseInt(value, 10, 64)
		if a < v {
			cmp = -1
		} else if a > v {
			cmp = 1
		}
	} else {
		a, _ := strconv.ParseUint(*raw, 10, 64)
		v, _ := strconv.ParseUint(value, 10, 64)
		if a < v {
			cmp = -1
		} else if a > v {
			cmp = 1
		}
	}
	switch op {
	case "eq":
		return cmp == 0
	case "ne":
		return cmp != 0
	case "gt":
		return cmp > 0
	case "gte":
		return cmp >= 0
	case "lt":
		return cmp < 0
	case "lte":
		return cmp <= 0
	}
	panic("unknown oracle operator")
}
func predicateBinding(typ, kind string) exploration.NumericBinding {
	return exploration.NumericBinding{Field: exploration.Field{Kind: kind, Key: "custom.value'\\"}, Type: typ, Unit: "widgets", Parse: "canonical-decimal"}
}
func predicateLeaf(b exploration.NumericBinding, op, value string) exploration.Predicate {
	return exploration.Predicate{Numeric: &b, Op: op, Value: value}
}
func predicateQuery(p exploration.Predicate) exploration.Query {
	return exploration.Query{Signal: "spans", Environment: "test", FromMS: 1000, ToMS: 61000, Limit: 100, Where: &p}
}
func predicateValues(b exploration.NumericBinding, raw *string) map[exploration.Field]string {
	v := map[exploration.Field]string{}
	if raw != nil {
		v[b.Field] = *raw
	}
	return v
}
func predicateComplement(b exploration.NumericBinding, op, value string) exploration.Predicate {
	states := []string{"valid", "missing", "empty", "invalid"}
	parts := []exploration.Predicate{}
	inverse := map[string]string{"eq": "ne", "ne": "eq", "gt": "lte", "gte": "lt", "lt": "gte", "lte": "gt"}
	if inv, ok := inverse[op]; ok {
		parts = append(parts, predicateLeaf(b, inv, value))
		for _, state := range states[1:] {
			parts = append(parts, predicateLeaf(b, state, ""))
		}
	} else {
		for _, state := range states {
			if state != op {
				parts = append(parts, predicateLeaf(b, state, ""))
			}
		}
	}
	return exploration.Predicate{Any: parts}
}
func checkPredicate(p exploration.Predicate, b exploration.NumericBinding) error {
	q := predicateQuery(p)
	if err := q.Validate(); err != nil {
		return err
	}
	plan, err := exploration.Compile(q)
	if err != nil {
		return err
	}
	if strings.Contains(plan.Count, " LIMIT ") || strings.Contains(plan.Rows, b.Field.Key) {
		return fmt.Errorf("paging affected totals or field escaped typed parameters")
	}
	found := false
	for _, v := range plan.Parameters {
		found = found || v == b.Field.Key
	}
	if !found {
		return fmt.Errorf("numeric field missing from typed parameters")
	}
	other := predicateComplement(b, p.Op, p.Value)
	if err := predicateQuery(other).Validate(); err != nil {
		return err
	}
	for _, raw := range predicateRawCorpus {
		values := predicateValues(b, raw)
		want := predicateOracle(raw, b, p.Op, p.Value)
		got := exploration.Match(&p, values)
		if got != want {
			return fmt.Errorf("oracle mismatch type=%s op=%s raw=%v threshold=%s", b.Type, p.Op, raw, p.Value)
		}
		rest := exploration.Match(&other, values)
		if got == rest {
			return fmt.Errorf("selected/rest not disjoint exhaustive")
		}
		sum := 0
		for _, state := range []string{"valid", "missing", "empty", "invalid"} {
			leaf := predicateLeaf(b, state, "")
			if exploration.Match(&leaf, values) {
				sum++
			}
		}
		if sum != 1 {
			return fmt.Errorf("parse states fail to partition")
		}
		// A nullable/string ne is not the numeric complement: typed ne excludes all
		// nonvalid states. AND/OR composition must preserve that distinction.
		truth := exploration.Predicate{Any: []exploration.Predicate{p, other}}
		falsehood := exploration.Predicate{All: []exploration.Predicate{p, other}}
		if !exploration.Match(&truth, values) || exploration.Match(&falsehood, values) {
			return fmt.Errorf("nested boolean composition diverged")
		}
	}
	return nil
}
func TestHegelTypedNumericPredicateProperties(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		typ := []string{"int64", "uint64"}[hegel.Draw(tc, hegel.Integers(0, 1))]
		kind := []string{"attribute", "resource"}[hegel.Draw(tc, hegel.Integers(0, 1))]
		b := predicateBinding(typ, kind)
		ops := []string{"eq", "ne", "gt", "gte", "lt", "lte", "valid", "missing", "empty", "invalid"}
		op := ops[hegel.Draw(tc, hegel.Integers(0, len(ops)-1))]
		thresholds := []string{"0", "1", "9007199254740993", "9223372036854775807"}
		if typ == "int64" {
			thresholds = append(thresholds, "-1", "-9223372036854775808")
		} else {
			thresholds = append(thresholds, "18446744073709551615")
		}
		value := thresholds[hegel.Draw(tc, hegel.Integers(0, len(thresholds)-1))]
		if hegel.Draw(tc, hegel.Integers(0, 1)) == 1 {
			value = strconv.Itoa(hegel.Draw(tc, hegel.Integers(0, 100000)))
		}
		if op == "valid" || op == "missing" || op == "empty" || op == "invalid" {
			value = ""
		}
		if err := checkPredicate(predicateLeaf(b, op, value), b); err != nil {
			tc.Fatal(err)
		}
	}, options(100)...)
}

type predicateEditMachine struct {
	b             exploration.NumericBinding
	op, value     string
	page          int
	frozen        [][]byte
	frozenQueries []exploration.Query
}

func (m *predicateEditMachine) RuleType(tc hegel.TestCase) {
	m.b.Type = []string{"int64", "uint64"}[hegel.Draw(tc, hegel.Integers(0, 1))]
}
func (m *predicateEditMachine) RuleNamespace(tc hegel.TestCase) {
	m.b.Field.Kind = []string{"attribute", "resource"}[hegel.Draw(tc, hegel.Integers(0, 1))]
}
func (m *predicateEditMachine) RuleThreshold(tc hegel.TestCase) {
	m.value = []string{"0", "1", "9007199254740993", "9223372036854775807"}[hegel.Draw(tc, hegel.Integers(0, 3))]
}
func (m *predicateEditMachine) RuleOperator(tc hegel.TestCase) {
	m.op = []string{"eq", "ne", "gt", "gte", "lt", "lte", "valid", "missing", "empty", "invalid"}[hegel.Draw(tc, hegel.Integers(0, 9))]
}
func (m *predicateEditMachine) RuleUnit(tc hegel.TestCase) {
	m.b.Unit = []string{"1", "ms", "widgets"}[hegel.Draw(tc, hegel.Integers(0, 2))]
}
func (m *predicateEditMachine) RulePage(tc hegel.TestCase) {
	m.page = hegel.Draw(tc, hegel.Integers(0, 1000))
}
func (m *predicateEditMachine) leaf() exploration.Predicate {
	value := m.value
	if m.op == "valid" || m.op == "missing" || m.op == "empty" || m.op == "invalid" {
		value = ""
	}
	return predicateLeaf(m.b, m.op, value)
}
func (m *predicateEditMachine) RuleFreeze(tc hegel.TestCase) {
	q := predicateQuery(m.leaf())
	q.Offset = m.page
	raw, e := json.Marshal(q)
	if e != nil {
		panic(e)
	}
	var detached exploration.Query
	if json.Unmarshal(raw, &detached) != nil {
		panic("freeze decode failed")
	}
	m.frozen = append(m.frozen, raw)
	m.frozenQueries = append(m.frozenQueries, detached)
}
func (m *predicateEditMachine) RuleRejectMalformed(tc hegel.TestCase) {
	p := m.leaf()
	switch hegel.Draw(tc, hegel.Integers(0, 5)) {
	case 0:
		p.Field = &exploration.Field{Kind: "attribute", Key: "other"}
	case 1:
		p.All = []exploration.Predicate{m.leaf()}
	case 2:
		p.Numeric.Parse = "coerce"
	case 3:
		p.Numeric.Type = "float64"
	case 4:
		p.Op = "contains"
	case 5:
		p.Op = "eq"
		p.Value = "01"
	}
	if predicateQuery(p).Validate() == nil {
		panic("malformed numeric predicate accepted")
	}
}
func (m *predicateEditMachine) InvariantPredicateMeaning(tc hegel.TestCase) {
	p := m.leaf()
	if err := checkPredicate(p, m.b); err != nil {
		panic(err)
	}
	q := predicateQuery(p)
	q.Offset = m.page
	q.Signal = "logs"
	if _, e := exploration.Compile(q); e != nil {
		panic(e)
	}
	for i, q := range m.frozenQueries {
		raw, e := json.Marshal(q)
		if e != nil || string(raw) != string(m.frozen[i]) || q.Validate() != nil {
			panic("frozen predicate changed")
		}
	}
}
func TestHegelTypedNumericPredicateStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		m := &predicateEditMachine{b: predicateBinding("int64", "attribute"), op: "gte", value: "0"}
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(20), hegel.WithAlwaysCheckInvariants("InvariantPredicateMeaning"))
	}, options(60)...)
}

// Fixed boundary matrix complements the generated campaigns so every operator
// sees both endpoints on every run, independently of a random schedule.
func TestTypedNumericPredicateBoundaryMatrix(t *testing.T) {
	for _, typ := range []string{"int64", "uint64"} {
		thresholds := []string{"0", "1", "9007199254740993", "9223372036854775807"}
		if typ == "int64" {
			thresholds = append(thresholds, "-1", "-9223372036854775808")
		} else {
			thresholds = append(thresholds, "18446744073709551615")
		}
		for _, kind := range []string{"attribute", "resource"} {
			b := predicateBinding(typ, kind)
			for _, op := range []string{"eq", "ne", "gt", "gte", "lt", "lte"} {
				for _, value := range thresholds {
					if e := checkPredicate(predicateLeaf(b, op, value), b); e != nil {
						t.Fatalf("%s %s %s %s: %v", typ, kind, op, value, e)
					}
				}
			}
			for _, op := range []string{"valid", "missing", "empty", "invalid"} {
				if e := checkPredicate(predicateLeaf(b, op, ""), b); e != nil {
					t.Fatal(e)
				}
				if predicateQuery(predicateLeaf(b, op, "1")).Validate() == nil {
					t.Fatal("state accepted a value")
				}
			}
			for _, value := range []string{"", "01", "-0", "+1", "1e3", "1.0", " 1", "18446744073709551616", "-9223372036854775809"} {
				if predicateQuery(predicateLeaf(b, "eq", value)).Validate() == nil {
					t.Fatalf("invalid threshold accepted: %q", value)
				}
			}
		}
	}
}
func TestTypedNumericPredicateBindingIndependence(t *testing.T) {
	b := predicateBinding("int64", "attribute")
	raw := "9007199254740993"
	p := predicateLeaf(b, "eq", raw)
	q := predicateQuery(p)
	// A summary may use another field, type and unit. It cannot rewrite or infer
	// the predicate's interpretation from a shared numeric-looking field name.
	other := predicateBinding("uint64", "resource")
	other.Field.Key = "http.response.status_code"
	other.Unit = "ms"
	q.Numeric = &other
	if q.Validate() != nil {
		t.Fatal("independent summary binding rejected")
	}
	if _, e := exploration.Compile(q); e != nil {
		t.Fatal(e)
	}
	if !exploration.Match(q.Where, predicateValues(b, &raw)) {
		t.Fatal("summary binding changed predicate")
	}
	for _, unit := range []string{"1", "ms", "widgets"} {
		changed := b
		changed.Unit = unit
		leaf := predicateLeaf(changed, "eq", raw)
		if !exploration.Match(&leaf, predicateValues(b, &raw)) {
			t.Fatal("declared unit coerced value")
		}
	}
}
