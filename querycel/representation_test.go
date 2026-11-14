package querycel

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
)

func TestPrintPredicateExactRoundTrip(t *testing.T) {
	for _, tc := range []struct{ signal, source string }{
		{"logs", `log.body == ""`},
		{"logs", `log.body.contains("é🦀_\\\"\n\t\x01")`},
		{"logs", `record.service != "x' OR 1=1 --"`},
		{"logs", `log.severity_number >= 18446744073709551615u`},
		{"logs", `log.severity_number <= 0u`},
		{"logs", `(log.body == "a" || record.service == "b") && log.severity_number >= 17u`},
		{"logs", `log.body == "a" && (record.service == "b" || log.body != "c")`},
		{"logs", `(log.body == "a" && record.service == "b") && log.body != "c"`},
		{"logs", `log.body == "a" && (record.service == "b" && log.body != "c")`},
		{"logs", `otel.severityBand(log.severity_number) == "error"`},
		{"spans", `span.duration_ns != 1000u && span.name.contains("work")`},
		{"spans", `span.status == "Error" || span.kind != "client"`},
	} {
		t.Run(tc.signal+tc.source, func(t *testing.T) {
			original, err := Compile(tc.signal, tc.source)
			if err != nil {
				t.Fatal(err)
			}
			predicate := original.Predicate()
			before, _ := json.Marshal(predicate)
			text, err := PrintPredicate(tc.signal, predicate)
			if err != nil {
				t.Fatal(err)
			}
			checked, err := Compile(tc.signal, text)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(checked.Predicate(), predicate) {
				t.Fatalf("structure changed: %s", text)
			}
			second, err := PrintPredicate(tc.signal, checked.Predicate())
			if err != nil || second != text {
				t.Fatalf("unstable canonical text: %q %v", second, err)
			}
			after, _ := json.Marshal(predicate)
			if string(before) != string(after) {
				t.Fatal("printer mutates input")
			}
		})
	}
}

func TestPrintPredicateRefusesCompleteUnsupportedStructures(t *testing.T) {
	leaf := func(key, op, value string) exploration.Predicate {
		return exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: key}, Op: op, Value: value}
	}
	attr := exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "team"}, Op: "eq", Value: "edge"}
	cycle := &exploration.Predicate{}
	cycle.All = []exploration.Predicate{*cycle}
	cycle.All[0].All = cycle.All
	for _, tc := range []struct {
		name, signal string
		predicate    *exploration.Predicate
	}{
		{"nil", "logs", nil},
		{"attribute", "logs", &attr},
		{"supported sibling cannot hide unsupported", "logs", &exploration.Predicate{All: []exploration.Predicate{leaf("body", "eq", "x"), attr}}},
		{"presence", "logs", ptrPredicate(leaf("body", "exists", ""))},
		{"single branch", "logs", &exploration.Predicate{All: []exploration.Predicate{leaf("body", "eq", "x")}}},
		{"flat ternary changes inverse structure", "logs", &exploration.Predicate{All: []exploration.Predicate{leaf("body", "eq", "a"), leaf("body", "eq", "b"), leaf("body", "eq", "c")}}},
		{"noncanonical uint spelling", "logs", ptrPredicate(leaf("severity_number", "eq", "01"))},
		{"derived field", "logs", ptrPredicate(leaf("body_pattern_v1", "exists", ""))},
		{"invalid signal", "metrics", ptrPredicate(leaf("body", "eq", "x"))},
		{"cycle", "logs", cycle},
		{"numeric binding", "logs", &exploration.Predicate{Numeric: &exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "duration"}, Type: "uint64", Unit: "ms", Parse: "canonical-uint"}, Op: "eq", Value: "1"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			text, err := PrintPredicate(tc.signal, tc.predicate)
			var d *Diagnostic
			if text != "" || !errors.As(err, &d) || d.Phase != PhaseRepresentation {
				t.Fatalf("partial or untyped result: %q %v", text, err)
			}
		})
	}
}
func ptrPredicate(p exploration.Predicate) *exploration.Predicate { return &p }

func TestPrintPredicateRefusesOversizedGeneratedSource(t *testing.T) {
	leaf := exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "body"}, Op: "eq", Value: strings.Repeat("x", 1024)}
	p := &exploration.Predicate{All: []exploration.Predicate{leaf, leaf, leaf, leaf}}
	text, err := PrintPredicate("logs", p)
	var d *Diagnostic
	if text != "" || !errors.As(err, &d) || d.Code != CodeBudget {
		t.Fatalf("expected bounded refusal: %q %v", text, err)
	}
}
