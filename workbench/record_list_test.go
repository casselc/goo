package workbench

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
)

func TestRecordListSavedTemplateAndReusePreserveKindAndScope(t *testing.T) {
	source := authoredDBTemplate(t)
	records := source.Composition.Children[1].Records
	records.Kind.Name = "record-list"
	records.Display.Columns = []string{"utc", "service", "summary", "context"}
	records.Display.FieldColumns = []exploration.Field{{Kind: "attribute", Key: "region"}, {Kind: "resource", Key: "region"}}
	store := Store{Dir: t.TempDir()}
	saved, err := store.CreateExplorationView(source)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := store.CreateExplorationView(source)
	if err != nil || loaded.ID != saved.ID {
		t.Fatal("kind is not content-stable", err)
	}
	loaded, err = store.LoadExplorationView(saved.ID)
	if err != nil || !reflect.DeepEqual(loaded.Composition.Children[1].Records, records) {
		t.Fatal("list declaration changed on reload", err)
	}
	target := cloneAuthored(t, source).Queries[0]
	target.Where.All[0].Value = "another-db-service"
	target.FromMS, target.ToMS, target.Offset = 70000, 130000, 0
	target.GroupBy = []exploration.Field{{Kind: "resource", Key: "db.namespace"}}
	before, _ := json.Marshal(loaded)
	expanded, err := ExpandViewTemplate(ViewTemplateExpansionRequest{Source: loaded, Query: &target})
	if err != nil {
		t.Fatal(err)
	}
	expected := target
	if !reflect.DeepEqual(expanded.View.Queries[0], expected) || !reflect.DeepEqual(expanded.View.Composition.Children[1].Records, records) {
		t.Fatal("list expansion rewrote scope or display")
	}
	expanded.View.Composition.Children[1].Records.Display.FieldColumns[0].Key = "mutated"
	after, _ := json.Marshal(loaded)
	if string(before) != string(after) {
		t.Fatal("expansion aliased saved list")
	}
	reuseTarget := cloneAuthored(t, source).Queries[0]
	reuseTarget.Where.All[0].Value = "reuse-service"
	reused, err := ReuseViewComposition(*source.Composition, source.Queries[0], reuseTarget)
	if err != nil || reused.Children[1].Records.Kind.Name != "record-list" {
		t.Fatal("composition reuse lost list kind", err)
	}
}

func TestRecordListUsesExistingBoundedRecordDeclarationValidation(t *testing.T) {
	for name, change := range map[string]func(*RecordViewDeclaration){
		"unknown kind":         func(d *RecordViewDeclaration) { d.Kind.Name = "record-cards" },
		"unknown kind version": func(d *RecordViewDeclaration) { d.Kind.Version = 2 },
		"missing UTC":          func(d *RecordViewDeclaration) { d.Display.Columns = []string{"service", "summary"} },
		"missing summary":      func(d *RecordViewDeclaration) { d.Display.Columns = []string{"utc", "service"} },
		"duplicate core":       func(d *RecordViewDeclaration) { d.Display.Columns = []string{"utc", "summary", "utc"} },
		"duplicate qualified": func(d *RecordViewDeclaration) {
			d.Display.FieldColumns = []exploration.Field{{Kind: "attribute", Key: "x"}, {Kind: "attribute", Key: "x"}}
		},
		"invalid qualified": func(d *RecordViewDeclaration) {
			d.Display.FieldColumns = []exploration.Field{{Kind: "physical", Key: "SpanAttributes"}}
		},
	} {
		t.Run(name, func(t *testing.T) {
			v := compositionView()
			v.Composition.Children[1].Records.Kind.Name = "record-list"
			change(v.Composition.Children[1].Records)
			if v.Validate() == nil {
				t.Fatal("invalid list declaration accepted")
			}
		})
	}
	for _, signal := range []string{"logs", "spans"} {
		v := compositionView()
		v.Queries[0].Signal = signal
		v.Composition.Children[0].Aggregate.Scope.Signal = signal
		v.Composition.Children[1].Records.Scope.Signal = signal
		v.Composition.Children[1].Records.Kind.Name = "record-list"
		if err := v.Validate(); err != nil {
			t.Fatal(signal, err)
		}
	}
}
