package exploration

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func numericBinding() NumericBinding {
	return NumericBinding{Field: Field{Kind: "attribute", Key: "db.response.returned_rows"}, Type: "int64", Unit: "{row}", Parse: "canonical-decimal"}
}
func numericValues(values ...string) []*string {
	out := make([]*string, len(values))
	for i := range values {
		out[i] = &values[i]
	}
	return out
}

func TestNumericCoverageAndExactArithmetic(t *testing.T) {
	b := numericBinding()
	values := append(numericValues("9223372036854775807", "9223372036854775807", "-9223372036854775808", "0", "", "-0", "01", "+1", " 1", "1.0", "1e2", "9223372036854775808"), nil)
	s, err := EvaluateNumericValues(b, values)
	want := NumericSummary{State: "observed", Unit: "{row}", Count: "13", Valid: "4", Missing: "1", Empty: "1", Invalid: "7", Sum: "9223372036854775806", Min: "-9223372036854775808", Max: "9223372036854775807"}
	if err != nil || s != want {
		t.Fatal(s, err)
	}
	b.Type = "uint64"
	s, err = EvaluateNumericValues(b, numericValues("18446744073709551615", "18446744073709551615", "-1", "18446744073709551616"))
	if err != nil || s.Sum != "36893488147419103230" || s.Valid != "2" || s.Invalid != "2" {
		t.Fatal(s, err)
	}
}

func TestNumericUnavailableIsNotZero(t *testing.T) {
	for _, tc := range []struct {
		v                              []*string
		state, missing, empty, invalid string
	}{
		{nil, "no_data", "0", "0", "0"},
		{[]*string{nil}, "no_valid_values", "1", "0", "0"},
		{numericValues(""), "no_valid_values", "0", "1", "0"},
		{numericValues("NaN"), "no_valid_values", "0", "0", "1"},
	} {
		s, e := EvaluateNumericValues(numericBinding(), tc.v)
		if e != nil || s.State != tc.state || s.Valid != "0" || s.Missing != tc.missing || s.Empty != tc.empty || s.Invalid != tc.invalid || s.Min != "" || s.Max != "" {
			t.Fatal(s, e)
		}
	}
	s, _ := EvaluateNumericValues(numericBinding(), numericValues("0"))
	if s.State != "observed" || s.Min != "0" || s.Max != "0" || s.Valid != "1" {
		t.Fatal(s)
	}
}

func TestNumericPlanDiagnosticsAndIsolation(t *testing.T) {
	q := Query{Signal: "logs", Environment: "test", FromMS: 1000, ToMS: 61000, Limit: 100, Where: &Predicate{Field: &Field{Kind: "resource", Key: "service.name"}, Op: "eq", Value: "ordinary"}}
	before, _ := json.Marshal(q)
	p, e := PlanNumericBinding(q, numericBinding())
	if e != nil || p.Executable || len(p.Diagnostics) != 1 || p.Diagnostics[0].Code != "execution_unavailable" || p.MeanDenominator != "valid" {
		t.Fatal(p, e)
	}
	p.Scope.Selector.Value = "changed"
	after, _ := json.Marshal(q)
	if string(before) != string(after) {
		t.Fatal("planner mutated question")
	}
	b := numericBinding()
	b.Field = Field{Kind: "column", Key: "body"}
	b.Type = "float64"
	b.Parse = "or-zero"
	b.Unit = "ms to s"
	p, e = PlanNumericBinding(q, b)
	if e != nil {
		t.Fatal(e)
	}
	codes := []string{}
	for _, d := range p.Diagnostics {
		codes = append(codes, d.Code)
		if d.Path == "" || d.Message == "" {
			t.Fatal(d)
		}
	}
	if !reflect.DeepEqual(codes, []string{"unsupported_field", "unsupported_type", "unsupported_parse", "invalid_unit", "execution_unavailable"}) {
		t.Fatal(codes)
	}
	if _, e = EvaluateNumericValues(b, nil); e == nil {
		t.Fatal("unsupported binding evaluated")
	}
	q.Signal = "metrics"
	if _, e = PlanNumericBinding(q, numericBinding()); e == nil {
		t.Fatal("invalid scope accepted")
	}
}

func TestNumericBindingsFailClosed(t *testing.T) {
	for _, unit := range []string{"", "a\x00b", "ms\n", "a\u2003b", strings.Repeat("a", 65), string([]byte{255})} {
		b := numericBinding()
		b.Unit = unit
		if _, e := EvaluateNumericValues(b, nil); e == nil {
			t.Fatalf("unit accepted %q", unit)
		}
	}
	b := numericBinding()
	b.Field.Key = ""
	if _, e := EvaluateNumericValues(b, nil); e == nil {
		t.Fatal("empty key accepted")
	}
}
