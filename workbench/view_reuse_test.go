package workbench

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
)

func TestReuseViewDeclarationKeepsDestinationScope(t *testing.T) {
	source := authoredView().Queries[0]
	d := authoredView().Displays["spans"]
	d.Display.Title = "Requests by service"
	d.Display.Query = "api"
	d.Display.RequiresComplete = true
	destination := source
	destination.Environment = ""
	selection, _ := identity.EffectiveEnvironment(identity.EnvironmentScope{Mode: "unreported"})
	wire, _ := selection.Selector()
	destination.EnvironmentSelector = &wire
	destination.FromMS, destination.ToMS, destination.Limit = 3000, 5000, 1
	field := exploration.Field{Kind: "column", Key: "service"}
	destination.Where = &exploration.Predicate{Field: &field, Op: "eq", Value: "worker"}
	before, _ := json.Marshal([]exploration.Query{source, destination})
	reused, err := ReuseViewDeclaration(d, source, destination)
	if err != nil {
		t.Fatal(err)
	}
	if reused != d {
		t.Fatal("reuse changed display intent")
	}
	reused.Display.Title = "Independent draft"
	if d.Display.Title != "Requests by service" {
		t.Fatal("reuse aliased original")
	}
	after, _ := json.Marshal([]exploration.Query{source, destination})
	if string(after) != string(before) {
		t.Fatal("reuse changed query scope")
	}
}

func TestReuseViewDeclarationCompatibility(t *testing.T) {
	source := authoredView().Queries[0]
	binding := exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work.delta"}, Type: "int64", Unit: "items", Parse: "canonical-decimal"}
	source.Numeric = &binding
	source.Summary = "span-duration"
	d := authoredView().Displays["spans"]
	for _, value := range []string{"count", "numeric.mean", "numeric.sum", "numeric.min", "numeric.max", "duration.mean", "duration.min", "duration.max", "duration.p50", "duration.p95", "duration.p99"} {
		t.Run(value, func(t *testing.T) {
			declaration := d
			declaration.Display.Value = value
			declaration.Display.Scale = "max"
			destination := source
			copyBinding := binding
			destination.Numeric = &copyBinding
			destination.Environment, destination.FromMS, destination.ToMS = "staging", 5000, 6000
			if _, err := ReuseViewDeclaration(declaration, source, destination); err != nil {
				t.Fatal(err)
			}
			destination.Numeric.Unit = "seconds"
			_, err := ReuseViewDeclaration(declaration, source, destination)
			if strings.HasPrefix(value, "numeric.") != (err != nil) {
				t.Fatal("wrong numeric quantity compatibility", err)
			}
		})
	}
	for name, change := range map[string]func(*exploration.Query){
		"signal":          func(q *exploration.Query) { q.Signal = "logs"; q.Summary = "" },
		"group namespace": func(q *exploration.Query) { q.GroupBy = []exploration.Field{{Kind: "attribute", Key: "service"}} },
		"group key":       func(q *exploration.Query) { q.GroupBy = []exploration.Field{{Kind: "column", Key: "name"}} },
		"missing group":   func(q *exploration.Query) { q.GroupBy = nil },
		"invalid window":  func(q *exploration.Query) { q.ToMS = q.FromMS },
		"missing numeric": func(q *exploration.Query) { q.Numeric = nil },
		"numeric field":   func(q *exploration.Query) { q.Numeric.Field.Kind = "resource" },
		"numeric type":    func(q *exploration.Query) { q.Numeric.Type = "uint64" },
		"numeric unit":    func(q *exploration.Query) { q.Numeric.Unit = "bytes" },
		"numeric parse":   func(q *exploration.Query) { q.Numeric.Parse = "lenient" },
		"numeric scale":   func(q *exploration.Query) { q.Numeric.Type = "decimal64"; q.Numeric.Scale = 2 },
	} {
		t.Run(name, func(t *testing.T) {
			destination := source
			copyBinding := binding
			destination.Numeric = &copyBinding
			change(&destination)
			declaration := d
			declaration.Display.Value = "numeric.mean"
			if _, err := ReuseViewDeclaration(declaration, source, destination); err == nil {
				t.Fatal("incompatible display accepted")
			}
		})
	}
	duration := d
	duration.Display.Value = "duration.p95"
	destination := source
	destination.Summary = ""
	if _, err := ReuseViewDeclaration(duration, source, destination); err == nil {
		t.Fatal("missing duration summary accepted")
	}
	invalidSource := source
	invalidSource.ToMS = invalidSource.FromMS
	if _, err := ReuseViewDeclaration(d, invalidSource, source); err == nil {
		t.Fatal("invalid source accepted")
	}
}
