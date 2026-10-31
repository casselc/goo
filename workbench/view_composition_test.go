package workbench

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
)

func compositionView() ExplorationView {
	q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 1, GroupBy: []exploration.Field{{Kind: "column", Key: "service"}}}
	return ExplorationView{Version: 1, Title: "Records and groups", Queries: []exploration.Query{q}, Composition: &ViewComposition{
		Version: "goo.composition/1", Kind: CompositionKind{Name: "stack", Version: 1}, Children: []ViewCompositionChild{
			{Aggregate: &ViewDeclaration{Version: "goo.view/1", Scope: ViewDeclarationScope{Kind: "record-query", Signal: "spans"}, Kind: ViewKind{Name: "bar", Version: 1}, Display: ViewDisplay{Category: "category", Value: "count", Scale: "population"}}},
			{Records: &RecordViewDeclaration{Version: "goo.view/1", Scope: ViewDeclarationScope{Kind: "record-query", Signal: "spans"}, Kind: RecordViewKind{Name: "record-table", Version: 1}, Display: RecordViewDisplay{Columns: []string{"utc", "summary"}, FieldColumns: []exploration.Field{{Kind: "attribute", Key: "http.route"}}}}},
		},
	}}
}

func cloneCompositionView(t *testing.T, v ExplorationView) ExplorationView {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var result ExplorationView
	if err = json.Unmarshal(raw, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func TestCompositionOrdersAndHidesRolesWithoutChangingSource(t *testing.T) {
	v := compositionView()
	if err := v.Validate(); err != nil {
		t.Fatal(err)
	}
	queries := cloneCompositionView(t, v).Queries
	store := Store{Dir: t.TempDir()}
	first, err := store.CreateExplorationView(v)
	if err != nil {
		t.Fatal(err)
	}
	v.Composition.Children[0], v.Composition.Children[1] = v.Composition.Children[1], v.Composition.Children[0]
	v.Composition.Children[1].Hidden = true
	v.Composition.Children[0].Records.Display.Columns = []string{"utc", "service", "summary"}
	second, err := store.CreateExplorationView(v)
	if err != nil {
		t.Fatal(err)
	}
	old, err := store.LoadExplorationView(first.ID)
	if err != nil || old.Composition.Children[0].Aggregate == nil || old.Composition.Children[0].Hidden || !reflect.DeepEqual(old.Queries, queries) || !reflect.DeepEqual(second.Queries, queries) || first.ID == second.ID {
		t.Fatal("composition mutated source or immutable prior revision", err)
	}
	raw, err := json.Marshal(second)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{`"layout"`, `"fieldColumns":null`, `"displays"`, `"tail"`, `"breakdowns"`} {
		if strings.Contains(string(raw), forbidden) {
			t.Fatal("flat or runtime authority escaped", forbidden, string(raw))
		}
	}
}

func TestCompositionRefusesConflictingOwnersAndUnsupportedChildren(t *testing.T) {
	changes := map[string]func(*ExplorationView){
		"second source": func(v *ExplorationView) { v.Queries = append(v.Queries, v.Queries[0]); v.Queries[1].Signal = "logs" },
		"layout owner":  func(v *ExplorationView) { v.Layout = "records" },
		"columns owner": func(v *ExplorationView) { v.Columns = []string{"utc", "summary"} },
		"field owner":   func(v *ExplorationView) { v.FieldColumns = []exploration.Field{{Kind: "attribute", Key: "x"}} },
		"display owner": func(v *ExplorationView) {
			v.Displays = map[string]ViewDeclaration{"spans": *v.Composition.Children[0].Aggregate}
		},
		"tail owner":       func(v *ExplorationView) { v.Tail = &ExplorationTail{Signal: "spans"} },
		"comparison owner": func(v *ExplorationView) { v.Breakdowns = map[string]ExplorationBreakdown{"spans": {Metric: "count"}} },
		"nested kind":      func(v *ExplorationView) { v.Composition.Kind.Name = "grid" },
		"root version":     func(v *ExplorationView) { v.Composition.Version = "goo.composition/2" },
		"duplicate role":   func(v *ExplorationView) { v.Composition.Children[1] = v.Composition.Children[0] },
		"both roles":       func(v *ExplorationView) { v.Composition.Children[0].Records = v.Composition.Children[1].Records },
		"no role":          func(v *ExplorationView) { v.Composition.Children[0] = ViewCompositionChild{} },
		"third child": func(v *ExplorationView) {
			v.Composition.Children = append(v.Composition.Children, v.Composition.Children[0])
		},
		"all hidden": func(v *ExplorationView) {
			for i := range v.Composition.Children {
				v.Composition.Children[i].Hidden = true
			}
		},
		"record source":  func(v *ExplorationView) { v.Composition.Children[1].Records.Scope.Signal = "logs" },
		"record kind":    func(v *ExplorationView) { v.Composition.Children[1].Records.Kind.Name = "waterfall" },
		"record version": func(v *ExplorationView) { v.Composition.Children[1].Records.Version = "goo.view/2" },
		"missing UTC": func(v *ExplorationView) {
			v.Composition.Children[1].Records.Display.Columns = []string{"service", "summary"}
		},
		"duplicate core": func(v *ExplorationView) {
			v.Composition.Children[1].Records.Display.Columns = []string{"utc", "summary", "utc"}
		},
		"invalid field": func(v *ExplorationView) {
			v.Composition.Children[1].Records.Display.FieldColumns = []exploration.Field{{Kind: "column", Key: "body"}}
		},
		"duplicate field": func(v *ExplorationView) {
			d := &v.Composition.Children[1].Records.Display
			d.FieldColumns = append(d.FieldColumns, d.FieldColumns[0])
		},
		"too many fields": func(v *ExplorationView) {
			for i := 0; i < 9; i++ {
				v.Composition.Children[1].Records.Display.FieldColumns = append(v.Composition.Children[1].Records.Display.FieldColumns, exploration.Field{Kind: "attribute", Key: strings.Repeat("x", i+1)})
			}
		},
		"page denominator": func(v *ExplorationView) { v.Composition.Children[0].Aggregate.Display.Value = "page.count" },
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			v := compositionView()
			change(&v)
			if v.Validate() == nil {
				t.Fatal("unsupported composition accepted")
			}
		})
	}
}

func TestCompositionWireRejectsEvenEmptyDualOptionsAndUnknownPayload(t *testing.T) {
	raw, _ := json.Marshal(compositionView())
	for _, extra := range []string{`"layout":""`, `"columns":[]`, `"displays":null`, `"tail":null`, `"breakdowns":{}`, `"fieldColumns":[]`, `"rows":[]`} {
		var v ExplorationView
		if json.Unmarshal(append(append([]byte(nil), raw[:len(raw)-1]...), []byte(","+extra+"}")...), &v) == nil {
			t.Fatal("dual/unknown owner accepted", extra)
		}
	}
	for _, raw := range []string{`{"records":null,"aggregate":null}`, `{"records":null,"id":"arbitrary"}`, `{"hidden":false}`, `{"records":{},"hidden":null}`} {
		var child ViewCompositionChild
		if json.Unmarshal([]byte(raw), &child) == nil {
			t.Fatal("invalid child keys accepted", raw)
		}
	}
	var display RecordViewDisplay
	if json.Unmarshal([]byte(`{"columns":["utc","summary"],"fieldColumns":null}`), &display) == nil {
		t.Fatal("nullable field columns accepted")
	}
}
