package querycel

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
)

func TestNumeric443ExactDeclarationsAndRepresentation(t *testing.T) {
	for _, signal := range []string{"logs", "spans"} {
		for _, target := range []struct{ name, kind string }{{"resource.attributes", "resource"}, {"record.attributes", "attribute"}} {
			for _, declaration := range []struct {
				typ, unit, threshold string
				scale                int
			}{
				{"int64", "1", "-9223372036854775808", 0}, {"uint64", "By", "18446744073709551615", 0}, {"decimal64", "USD", "92233720368547758.07", 2},
			} {
				for _, op := range numericOperations {
					constructor := fmt.Sprintf("%s.%s(%q, %q", target.name, declaration.typ, "example.amount", declaration.unit)
					if declaration.scale != 0 {
						constructor += fmt.Sprintf(", %d", declaration.scale)
					}
					constructor += ")"
					threshold := fmt.Sprintf("%q", declaration.threshold)
					if numericState(op) {
						threshold = ""
					}
					expression := constructor + "." + op + "(" + threshold + ")"
					program, err := Compile(signal, expression)
					if err != nil {
						t.Fatalf("%s: %v", expression, err)
					}
					want := &exploration.Predicate{Numeric: &exploration.NumericBinding{Field: exploration.Field{Kind: target.kind, Key: "example.amount"}, Type: declaration.typ, Unit: declaration.unit, Parse: "canonical-decimal", Scale: declaration.scale}, Op: op}
					if !numericState(op) {
						want.Value = declaration.threshold
					}
					if !reflect.DeepEqual(program.Predicate(), want) {
						t.Fatalf("%s: got %+v want %+v", expression, program.Predicate(), want)
					}
					printed, err := PrintPredicate(signal, program.Predicate())
					if err != nil {
						t.Fatalf("%s print: %v", expression, err)
					}
					again, err := Compile(signal, printed)
					if err != nil || !reflect.DeepEqual(again.Predicate(), want) {
						t.Fatalf("roundtrip %s: %v", printed, err)
					}
					for _, raw := range []*string{nil, new(""), new("0"), new("-1"), new("01"), new("+1"), new("0.00"), new("-0.00"), new("1.20"), new("1.2"), new("1e3"), new("1\n"), new(declaration.threshold), new(strings.Repeat("9", 50))} {
						values := map[exploration.Field]string{}
						if raw != nil {
							values[want.Numeric.Field] = *raw
						}
						got, err := program.Eval(values)
						if err != nil {
							t.Fatalf("%s/%v: %v", expression, raw, err)
						}
						if got != exploration.Match(want, values) {
							t.Fatalf("reference mismatch %s/%v", expression, raw)
						}
					}
				}
			}
		}
	}
}

func TestNumeric443ValidOnlyComparisonsAndIndependentDeclarations(t *testing.T) {
	field := exploration.Field{Kind: "attribute", Key: "retry.count"}
	for _, raw := range []*string{nil, new(""), new("invalid"), new("-1"), new("01"), new("+2"), new("2.0")} {
		values := map[exploration.Field]string{}
		if raw != nil {
			values[field] = *raw
		}
		program, err := Compile("logs", `record.attributes.uint64("retry.count", "1").ne("2")`)
		if err != nil {
			t.Fatal(err)
		}
		got, err := program.Eval(values)
		if err != nil || got {
			t.Fatalf("invalid/missing matched inequality: %v %t %v", raw, got, err)
		}
	}
	program, err := Compile("logs", `record.attributes.int64("retry.count", "1").valid() && record.attributes.uint64("retry.count", "1").invalid()`)
	if err != nil {
		t.Fatal(err)
	}
	got, err := program.Eval(map[exploration.Field]string{field: "-1"})
	if err != nil || !got {
		t.Fatalf("shared field declarations conflict: %t %v", got, err)
	}
	printed, err := PrintPredicate("logs", program.Predicate())
	if err != nil {
		t.Fatal(err)
	}
	again, err := Compile("logs", printed)
	if err != nil || !reflect.DeepEqual(program.Predicate(), again.Predicate()) {
		t.Fatalf("compound representation changed: %s %v", printed, err)
	}
	q := exploration.Query{Signal: "logs", Environment: "prod", FromMS: 1, ToMS: 60001, Limit: 100, Where: &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "worker"}}
	refined, err := program.Apply(q)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(refined.Where.All[0], *q.Where) || refined.Environment != q.Environment || refined.FromMS != q.FromMS || refined.ToMS != q.ToMS {
		t.Fatal("numeric refinement lost parent scope")
	}
}

func TestNumeric443RefusesAmbiguousOrComputedDeclarations(t *testing.T) {
	for _, expression := range []string{
		`record.attributes.uint64("x", "1")`,
		`record.attributes.uint64("x", "1") > 2u`,
		`record.attributes.uint64("x", "1") == record.attributes.uint64("x", "1")`,
		`record.attributes.uint64("x", "1").gte(2u)`,
		`record.attributes.uint64("x", "1").gte(2.0)`,
		`record.attributes.uint64("x", "1").gte("02")`,
		`record.attributes.uint64("x", "1").gte("18446744073709551616")`,
		`record.attributes.uint64("x", "1").gte("-1")`,
		`record.attributes.int64("x", "1").eq("-0")`,
		`record.attributes.int64("x", "1").eq("9223372036854775808")`,
		`record.attributes.decimal64("x", "USD", 2).eq("1.2")`,
		`record.attributes.decimal64("x", "USD", 2).eq("1.200")`,
		`record.attributes.decimal64("x", "USD", 2).eq("-0.00")`,
		`record.attributes.decimal64("x", "USD", 2).eq("92233720368547758.08")`,
		`record.attributes.decimal64("x", "USD", 0).valid()`,
		`record.attributes.decimal64("x", "USD", 10).valid()`,
		`record.attributes.decimal64("x", "USD", 2u).valid()`,
		`record.attributes.decimal64("x", "USD", 1 + 1).valid()`,
		`record.attributes.uint64(record.service, "1").valid()`,
		`record.attributes.uint64("x", "B" + "y").valid()`,
		`record.attributes.uint64("x", "").valid()`,
		`record.attributes.uint64("x", "some unit").valid()`,
		`record.attributes.uint64("x", "1").gte(record.service)`,
		`record.attributes.uint64("x", "1").valid("extra")`,
		`!record.attributes.uint64("x", "1").eq("1")`,
		`record.service.uint64("x", "1").valid()`,
	} {
		program, err := Compile("logs", expression)
		var d *Diagnostic
		if program != nil || !errors.As(err, &d) || d.Span == nil {
			t.Fatalf("missing located refusal for %s: %v", expression, err)
		}
	}
}

func TestNumeric443MetadataCostDoesNotChargeAnotherValueScan(t *testing.T) {
	estimator := costEstimator{}
	for _, typ := range []string{"int64", "uint64", "decimal64"} {
		cost := estimator.EstimateCallCost("", numericID(typ), nil, nil)
		if cost == nil || cost.Max != 256 {
			t.Fatalf("constructor %s cost: %+v", typ, cost)
		}
	}
	terminal := estimator.EstimateCallCost("", numericID("gte"), nil, nil)
	if terminal == nil || terminal.Max != MaxFieldBytes {
		t.Fatalf("terminal lost bounded raw-value allowance: %+v", terminal)
	}
	// These twelve leaves fit both the ordinary predicate budget and the CEL
	// budget. Charging the raw-value bound twice per leaf falsely refused them.
	expressions := make([]string, 12)
	values := make(map[exploration.Field]string, 12)
	for i := range expressions {
		key := fmt.Sprintf("example.count.%d", i)
		expressions[i] = fmt.Sprintf(`record.attributes.uint64(%q, "1").gte("0")`, key)
		values[exploration.Field{Kind: "attribute", Key: key}] = "1"
	}
	program, err := Compile("logs", strings.Join(expressions, " && "))
	if err != nil {
		t.Fatal(err)
	}
	got, err := program.Eval(values)
	if err != nil || !got {
		t.Fatalf("bounded multi-clause evaluation: %t %v", got, err)
	}
	values[exploration.Field{Kind: "attribute", Key: "example.count.5"}] = strings.Repeat("9", MaxFieldBytes)
	got, err = program.Eval(values)
	if err != nil || got {
		t.Fatalf("long in-envelope numeric text must remain invalid: %t %v", got, err)
	}
}
