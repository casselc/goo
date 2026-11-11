package workbench

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
)

func TestRecordSpanWaterfallPersistsAndReusesExistingComposition(t *testing.T) {
	for _, hidden := range []bool{false, true} {
		source := authoredComparisonTemplate(t, true)
		record := source.Composition.Children[1].Records
		record.Kind.Name = "record-span-waterfall"
		record.Display.Columns = []string{"utc", "service", "summary", "context"}
		// All eight existing qualified fields remain legal; presentation does not
		// narrow the supplied page into a measurement or complete trace.
		record.Display.FieldColumns = []exploration.Field{{Kind: "column", Key: "timestamp_ns"}, {Kind: "column", Key: "duration_ns"}, {Kind: "attribute", Key: "db.namespace"}, {Kind: "resource", Key: "db.namespace"}, {Kind: "attribute", Key: "empty"}, {Kind: "resource", Key: "empty"}, {Kind: "attribute", Key: "missing"}, {Kind: "resource", Key: "missing"}}
		source.Composition.Children[1].Hidden = hidden
		source.Composition.Children = []ViewCompositionChild{source.Composition.Children[3], source.Composition.Children[1], source.Composition.Children[2], source.Composition.Children[0]}
		store := Store{Dir: t.TempDir()}
		saved, err := store.CreateExplorationView(source)
		if err != nil {
			t.Fatal(err)
		}
		loaded, err := store.LoadExplorationView(saved.ID)
		if err != nil || !reflect.DeepEqual(loaded.Composition, source.Composition) {
			t.Fatal("saved waterfall changed", err)
		}
		target := cloneAuthored(t, source).Queries[0]
		target.Environment = "staging"
		target.FromMS = 70000
		target.ToMS = 135001
		target.Offset = 0
		target.GroupBy = []exploration.Field{{Kind: "attribute", Key: "db.namespace"}}
		target.Where.All[0].Value = "other-client"
		baseline := strings.Repeat("b", 64)
		beforeSource, _ := json.Marshal(loaded)
		beforeTarget, _ := json.Marshal(target)
		expanded, err := ExpandViewTemplate(ViewTemplateExpansionRequest{Source: loaded, Query: &target, BaselineCapture: &baseline})
		if err != nil {
			t.Fatal(err)
		}
		expected := cloneAuthored(t, source)
		expected.Composition.Children[2].Comparison.BaselineCapture = baseline
		expectedQuery := target
		if !reflect.DeepEqual(expanded.View.Composition, expected.Composition) || !reflect.DeepEqual(expanded.View.Queries, []exploration.Query{expectedQuery}) {
			t.Fatal("waterfall expansion changed presentation or target query")
		}
		expanded.View.Composition.Children[1].Records.Display.FieldColumns[0].Key = "mutated"
		afterSource, _ := json.Marshal(loaded)
		afterTarget, _ := json.Marshal(target)
		if string(beforeSource) != string(afterSource) || string(beforeTarget) != string(afterTarget) {
			t.Fatal("waterfall expansion aliased input")
		}
		// Ordinary reuse preserves exact root compatibility, including the existing
		// baseline and interval policies, without a second waterfall query.
		reuseTarget := cloneAuthored(t, source).Queries[0]
		reuseTarget.FromMS = 80000
		reuseTarget.ToMS = 140000
		reuseTarget.Offset = 100
		reused, err := ReuseViewComposition(*source.Composition, source.Queries[0], reuseTarget)
		if err != nil || !reflect.DeepEqual(reused, *source.Composition) {
			t.Fatal("waterfall reuse changed composition", err)
		}
	}
}

func TestRecordSpanWaterfallRefusesOtherSignalsAndInvalidRecordOptions(t *testing.T) {
	for _, hidden := range []bool{false, true} {
		for _, signal := range []string{"logs", "metrics"} {
			v := compositionView()
			v.Queries[0].Signal = signal
			v.Composition.Children[0].Aggregate.Scope.Signal = signal
			v.Composition.Children[1].Records.Scope.Signal = signal
			v.Composition.Children[1].Records.Kind.Name = "record-span-waterfall"
			v.Composition.Children[1].Hidden = hidden
			if v.Composition.Children[1].Records.Validate(v.Queries[0]) == nil {
				t.Fatal("non-span record declaration accepted", signal, hidden)
			}
			if v.Validate() == nil {
				t.Fatal("non-span waterfall accepted", signal, hidden)
			}
			if _, err := (Store{Dir: t.TempDir()}).CreateExplorationView(v); err == nil {
				t.Fatal("invalid waterfall published", signal, hidden)
			}
			v.Template = &ViewTemplateBinding{Version: "goo.template/1", Group: "destination", Profile: "ordinary"}
			if v.Validate() == nil {
				t.Fatal("ordinary template enabled non-span waterfall", signal, hidden)
			}
		}
	}
	for name, mutate := range map[string]func(*RecordViewDeclaration){
		"version":          func(d *RecordViewDeclaration) { d.Kind.Version = 2 },
		"scope":            func(d *RecordViewDeclaration) { d.Scope.Signal = "logs" },
		"required UTC":     func(d *RecordViewDeclaration) { d.Display.Columns = []string{"service", "summary"} },
		"required summary": func(d *RecordViewDeclaration) { d.Display.Columns = []string{"utc", "context"} },
		"duplicate field": func(d *RecordViewDeclaration) {
			d.Display.FieldColumns = []exploration.Field{{Kind: "attribute", Key: "x"}, {Kind: "attribute", Key: "x"}}
		},
		"nine fields": func(d *RecordViewDeclaration) {
			d.Display.FieldColumns = nil
			for _, key := range []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"} {
				d.Display.FieldColumns = append(d.Display.FieldColumns, exploration.Field{Kind: "attribute", Key: key})
			}
		},
		"physical field": func(d *RecordViewDeclaration) {
			d.Display.FieldColumns = []exploration.Field{{Kind: "physical", Key: "SpanAttributes"}}
		},
	} {
		t.Run(name, func(t *testing.T) {
			v := compositionView()
			v.Composition.Children[1].Records.Kind.Name = "record-span-waterfall"
			mutate(v.Composition.Children[1].Records)
			if v.Validate() == nil {
				t.Fatal("invalid waterfall accepted")
			}
		})
	}
}
