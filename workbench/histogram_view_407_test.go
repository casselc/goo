package workbench

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
)

func histogramQuestion407(signal string) exploration.Query {
	return exploration.Query{Signal: signal, Environment: "source", FromMS: 1000, ToMS: 6000, Limit: 2, Numeric: &exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "amount"}, Type: "decimal64", Scale: 2, Unit: "USD", Parse: "canonical-decimal"}, NumericHistogram: &exploration.NumericHistogramSpec{Binning: "powers-of-two"}}
}
func cloneHistogramQuestion407(q exploration.Query) exploration.Query {
	raw, _ := json.Marshal(q)
	var out exploration.Query
	_ = json.Unmarshal(raw, &out)
	return out
}
func histogramView407(q exploration.Query) ExplorationView {
	return ExplorationView{Version: 1, Title: "Value distribution", Queries: []exploration.Query{q}, Template: &ViewTemplateBinding{Version: "goo.template/1", Group: "none", Profile: "ordinary-histogram"}, Composition: &ViewComposition{Version: "goo.composition/1", Kind: CompositionKind{Name: "stack", Version: 1}, Children: []ViewCompositionChild{{Histogram: &HistogramViewDeclaration{Version: "goo.view/1", Scope: ViewDeclarationScope{Kind: "record-query", Signal: q.Signal}, Kind: HistogramViewKind{Name: "histogram", Version: 1}, Display: HistogramViewDisplay{Category: "from", Value: "count", RequiresComplete: true}}}, {Records: &RecordViewDeclaration{Version: "goo.view/1", Scope: ViewDeclarationScope{Kind: "record-query", Signal: q.Signal}, Kind: RecordViewKind{Name: "record-table", Version: 1}, Display: RecordViewDisplay{Columns: []string{"utc", "summary"}}}}}}}
}
func TestHistogram407StoreTemplateAndDestination(t *testing.T) {
	for _, signal := range []string{"logs", "spans"} {
		t.Run(signal, func(t *testing.T) {
			q := histogramQuestion407(signal)
			v := histogramView407(q)
			if err := v.Validate(); err != nil {
				t.Fatal(err)
			}
			store := Store{Dir: t.TempDir()}
			saved, err := store.CreateExplorationView(v)
			if err != nil {
				t.Fatal(err)
			}
			loaded, err := store.LoadExplorationView(saved.ID)
			if err != nil || !reflect.DeepEqual(saved, loaded) {
				t.Fatal("store changed declaration", err)
			}
			destination := cloneHistogramQuestion407(q)
			destination.Environment = "destination"
			destination.FromMS += 10000
			destination.ToMS += 13000
			destination.Limit = 7
			destination.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "team"}, Op: "eq", Value: "destination"}
			before, _ := json.Marshal(loaded)
			out, err := ExpandViewTemplate(ViewTemplateExpansionRequest{Source: loaded, Query: &destination})
			if err != nil {
				t.Fatal(err)
			}
			after, _ := json.Marshal(loaded)
			if string(before) != string(after) || !reflect.DeepEqual(out.View.Queries, []exploration.Query{destination}) || !reflect.DeepEqual(out.View.Composition, v.Composition) {
				t.Fatal("reuse replaced query or mutated donor")
			}
			for _, mutate := range []func(*exploration.Query){func(q *exploration.Query) { q.Numeric.Field.Kind = "resource" }, func(q *exploration.Query) { q.Numeric.Field.Key = "other" }, func(q *exploration.Query) { q.Numeric.Type = "int64"; q.Numeric.Scale = 0 }, func(q *exploration.Query) { q.Numeric.Scale = 3 }, func(q *exploration.Query) { q.Numeric.Unit = "usd" }, func(q *exploration.Query) { q.Offset = 1 }, func(q *exploration.Query) { q.NumericHistogram = nil }, func(q *exploration.Query) { q.GroupBy = []exploration.Field{{Kind: "column", Key: "service"}} }} {
				bad := cloneHistogramQuestion407(destination)
				mutate(&bad)
				if _, err := ExpandViewTemplate(ViewTemplateExpansionRequest{Source: loaded, Query: &bad}); err == nil {
					t.Fatal("incompatible histogram template accepted")
				}
			}
		})
	}
}
func TestHistogram407ClosedDeclaration(t *testing.T) {
	q := histogramQuestion407("logs")
	v := histogramView407(q)
	raw, _ := json.Marshal(v.Composition.Children[0].Histogram)
	for _, bad := range []string{`{"version":"goo.view/1","scope":{"kind":"record-query","signal":"logs"},"kind":{"name":"histogram","version":1},"display":{"category":"from","value":"count","title":null}}`, `{"version":"goo.view/1","scope":{"kind":"record-query","signal":"logs","query":{}},"kind":{"name":"histogram","version":1},"display":{"category":"from","value":"count"}}`, `{"version":"goo.view/1","scope":{"kind":"record-query","signal":"logs"},"kind":{"name":"histogram","version":1},"display":{"category":"from","value":"count","normalization":"density"}}`} {
		var d HistogramViewDeclaration
		if json.Unmarshal([]byte(bad), &d) == nil {
			t.Fatal("open declaration", bad)
		}
	}
	var d HistogramViewDeclaration
	if err := json.Unmarshal(raw, &d); err != nil || d.Validate(q) != nil {
		t.Fatal(err)
	}
	v.Composition.Children[0].Hidden = true
	if v.Validate() == nil {
		t.Fatal("hidden histogram accepted")
	}
}
