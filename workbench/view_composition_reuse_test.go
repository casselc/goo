package workbench

import (
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
)

func TestReuseCompositionCopiesOnlyDetachedIntent(t *testing.T) {
	v := compositionView()
	v.Composition.Children = slices.Clone(v.Composition.Children)
	slices.Reverse(v.Composition.Children)
	v.Composition.Children[1].Hidden = true
	v.Composition.Children[1].Aggregate.Display.RequiresComplete = true
	source := v.Queries[0]
	destination := source
	destination.Environment = ""
	selection, _ := identity.EffectiveEnvironment(identity.EnvironmentScope{Mode: "unreported"})
	selector, _ := selection.Selector()
	destination.EnvironmentSelector = &selector
	destination.FromMS, destination.ToMS, destination.Limit, destination.Offset = 3000, 5000, 3, 2
	destination.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "team"}, Op: "eq", Value: "other"}
	before, _ := json.Marshal([]exploration.Query{source, destination})
	intent, _ := json.Marshal(v.Composition)
	next, err := ReuseViewComposition(*v.Composition, source, destination)
	if err != nil || !reflect.DeepEqual(next, *v.Composition) {
		t.Fatal("reuse changed presentation intent", err, next)
	}
	// Mutation reaches through every mutable level of the returned composition.
	next.Children[0].Records.Display.Columns[0] = "context"
	next.Children[0].Records.Display.FieldColumns[0].Key = "changed"
	next.Children[0].Records.Kind.Name = "changed"
	next.Children[1].Aggregate.Display.Title = "changed"
	next.Children[1].Hidden = false
	slices.Reverse(next.Children)
	afterIntent, _ := json.Marshal(v.Composition)
	afterQueries, _ := json.Marshal([]exploration.Query{source, destination})
	if string(intent) != string(afterIntent) || string(before) != string(afterQueries) {
		t.Fatal("reuse aliased source intent or changed either query")
	}
}

func TestReuseCompositionChecksAggregateQuantityAndRecordProfile(t *testing.T) {
	base := compositionView()
	source := base.Queries[0]
	source.Numeric = &exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work.delta"}, Type: "int64", Unit: "items", Parse: "canonical-decimal"}
	source.Summary = "span-duration"
	for _, measure := range []string{"count", "numeric.mean", "numeric.sum", "numeric.min", "numeric.max", "duration.mean", "duration.min", "duration.max", "duration.p50", "duration.p95", "duration.p99"} {
		t.Run(measure, func(t *testing.T) {
			v := cloneCompositionView(t, base)
			v.Composition.Children[0].Aggregate.Display.Value = measure
			v.Composition.Children[0].Aggregate.Display.Scale = "max"
			destination := source
			destination.Environment, destination.FromMS, destination.ToMS = "staging", 3000, 4000
			if _, err := ReuseViewComposition(*v.Composition, source, destination); err != nil {
				t.Fatal(err)
			}
			binding := *source.Numeric
			binding.Unit = "bytes"
			destination.Numeric = &binding
			_, err := ReuseViewComposition(*v.Composition, source, destination)
			if strings.HasPrefix(measure, "numeric.") != (err != nil) {
				t.Fatal("wrong quantity compatibility", err)
			}
		})
	}
	for name, mutate := range map[string]func(*ViewComposition, *exploration.Query){
		"signal": func(_ *ViewComposition, q *exploration.Query) { q.Signal = "logs"; q.Summary = ""; q.Numeric = nil },
		"qualified group": func(_ *ViewComposition, q *exploration.Query) {
			q.GroupBy = []exploration.Field{{Kind: "attribute", Key: "service"}}
		},
		"invalid query": func(_ *ViewComposition, q *exploration.Query) { q.ToMS = q.FromMS },
		"both hidden": func(c *ViewComposition, _ *exploration.Query) {
			for i := range c.Children {
				c.Children[i].Hidden = true
			}
		},
		"record field": func(c *ViewComposition, _ *exploration.Query) {
			c.Children[1].Records.Display.FieldColumns = []exploration.Field{{Kind: "column", Key: "body"}}
		},
		"record scope": func(c *ViewComposition, _ *exploration.Query) { c.Children[1].Records.Scope.Signal = "logs" },
	} {
		t.Run(name, func(t *testing.T) {
			v := cloneCompositionView(t, base)
			destination := source
			mutate(v.Composition, &destination)
			if _, err := ReuseViewComposition(*v.Composition, source, destination); err == nil {
				t.Fatal("incompatible composition was reused")
			}
		})
	}
	invalidSource := source
	invalidSource.ToMS = invalidSource.FromMS
	if _, err := ReuseViewComposition(*base.Composition, invalidSource, source); err == nil {
		t.Fatal("invalid source query accepted")
	}
}

func TestReuseCompositionSavesDestinationWithoutChangingOriginal(t *testing.T) {
	store := Store{Dir: t.TempDir()}
	source := compositionView()
	saved, err := store.CreateExplorationView(source)
	if err != nil {
		t.Fatal(err)
	}
	destination := source.Queries[0]
	destination.Environment, destination.FromMS, destination.ToMS, destination.Limit = "staging", 3000, 4000, 7
	destination.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "team.unreported"}, Op: "exists"}
	c, err := ReuseViewComposition(*saved.Composition, saved.Queries[0], destination)
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.CreateExplorationView(ExplorationView{Version: 1, Title: "Current population", Queries: []exploration.Query{destination}, Composition: &c})
	if err != nil || second.ID == saved.ID || !reflect.DeepEqual(second.Queries, []exploration.Query{destination}) {
		t.Fatal("reused configuration changed its destination", err)
	}
	old, err := store.LoadExplorationView(saved.ID)
	if err != nil || !reflect.DeepEqual(old, saved) {
		t.Fatal("source revision changed", err)
	}
}
