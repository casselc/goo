package verification

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"os"
	"reflect"
	"testing"
)

func TestQuintComparisonInspectionLedger(t *testing.T) {
	name := os.Getenv("GOO_COMPARISON_INSPECTION_LEDGER")
	if name == "" {
		t.Skip("requires observed browser ledger")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Initial                                                []exploration.Query
			Selected, Baseline                                     exploration.Query
			BeforePreview, AfterPreview, BeforeReturn, AfterReturn int
			Events                                                 []struct {
				Action       string
				Depth, Reads int
			}
		}

		if json.Unmarshal(raw, &l) != nil {
			t.Fatal("invalid ledger")
		}
		switch mutation {
		case 0:
			l.AfterPreview++
		case 1:
			l.AfterReturn++
		case 2:
			l.Selected.Environment = "other"
		case 3:
			l.Baseline.Where = nil
		case 4:
			for i := range l.Events {
				if l.Events[i].Action == "back" {
					l.Events[i].Depth++
					break
				}
			}
		case 5:
			l.Baseline.FromMS = l.Selected.FromMS
		}
		coherent := l.BeforePreview == l.AfterPreview && l.BeforeReturn == l.AfterReturn && len(l.Initial) == 3 && l.BeforePreview == 3 && l.BeforeReturn == 4 && len(l.Events) > 0 && l.Events[len(l.Events)-1].Reads == 5
		if coherent {
			for i, q := range []exploration.Query{l.Selected, l.Baseline} {
				source := l.Initial[0]
				if i == 1 {
					source = l.Initial[2]
				}
				expected := source
				expected.Offset = 0
				op := "missing"
				if i == 1 {
					op = "eq"
				}
				field := exploration.Field{Kind: "attribute", Key: "region"}
				leaf := exploration.Predicate{Field: &field, Op: op}
				expected.Where = &exploration.Predicate{All: []exploration.Predicate{*source.Where, leaf}}
				_, err := exploration.Compile(q)
				coherent = coherent && err == nil && reflect.DeepEqual(expected, q)
			}
		}

		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, event := range l.Events {
			rec.Record(context.Background(), event.Action, nil, "obs.reads", event.Reads, "obs.depth", event.Depth, "obs.coherent", coherent)
		}
		c, e := validate.NewChecker("model/population_drilldown.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		r, e := c.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return r.OK()
	}
	if !check(-1) {
		t.Fatal("observed comparison inspection trace rejected")
	}
	for i := 0; i < 6; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}
