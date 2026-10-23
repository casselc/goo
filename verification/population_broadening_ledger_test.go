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

func TestQuintPopulationBroadeningLedger(t *testing.T) {
	name := os.Getenv("GOO_BROADENING_LEDGER")
	if name == "" {
		t.Skip("requires observed browser ledger")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Initial, Broadened                                     []exploration.Query
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
			l.Broadened[1].Environment = "other"
		case 3:
			l.Broadened[0].Where = l.Initial[0].Where
		case 5:
			l.Broadened[0].FromMS = l.Initial[0].FromMS + 1
		case 4:
			for i := range l.Events {
				if l.Events[i].Action == "back" {
					l.Events[i].Depth++
					break
				}
			}
		}
		coherent := l.BeforePreview == l.AfterPreview && l.BeforeReturn == l.AfterReturn && len(l.Initial) == 2 && len(l.Broadened) == 2
		if coherent {
			for i, q := range l.Broadened {
				expected := l.Initial[i]
				if i == 0 {
					expected.Where = &l.Initial[0].Where.All[0]
					expected.Offset = 0
				}
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
		t.Fatal("observed broadening trace rejected")
	}
	for i := 0; i < 6; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}
