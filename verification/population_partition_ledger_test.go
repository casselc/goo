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

func TestQuintPopulationPartitionLedger(t *testing.T) {
	name := os.Getenv("GOO_PARTITION_LEDGER")
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
			Selected, Rest                                         exploration.Query
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
			l.Rest.Environment = "other"
		case 3:
			l.Rest.Where = nil
		case 6:
			l.Rest.Where = l.Selected.Where
		case 5:
			l.Selected.FromMS++
		case 4:
			for i := range l.Events {
				if l.Events[i].Action == "back" {
					l.Events[i].Depth++
					break
				}
			}
		}
		coherent := l.BeforePreview == l.AfterPreview && l.BeforeReturn == l.AfterReturn && len(l.Initial) == 2
		if coherent {
			parent := l.Initial[0]
			for _, q := range []exploration.Query{l.Selected, l.Rest} {
				expected := parent
				expected.Offset = 0
				expected.Where = q.Where
				_, err := exploration.Compile(q)
				coherent = coherent && err == nil && reflect.DeepEqual(expected, q) && q.Where != nil
			}
			if coherent {
				for owner := 0; owner < 2; owner++ {
					for region := 0; region < 4; region++ {
						record := map[exploration.Field]string{{Kind: "column", Key: "service"}: "api", {Kind: "attribute", Key: "owner"}: []string{"core", "outside"}[owner]}
						if region > 0 {
							record[exploration.Field{Kind: "attribute", Key: "region"}] = []string{"", "", "west", "east"}[region]
						}
						p := exploration.Match(parent.Where, record)
						a := exploration.Match(l.Selected.Where, record)
						b := exploration.Match(l.Rest.Where, record)
						coherent = coherent && !(a && b) && (a || b) == p && a == (p && region < 2)
					}
				}
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
		t.Fatal("observed partition trace rejected")
	}
	for i := 0; i < 7; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}
