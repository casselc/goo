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

func TestQuintComparisonSeriesLedger(t *testing.T) {
	name := os.Getenv("GOO_COMPARISON_SERIES_LEDGER")
	if name == "" {
		t.Skip("requires observed browser ledger")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			InitialQueries []exploration.Query
			Selected       exploration.Query
			Initial, After int
			Events         []struct {
				Action       string
				Depth, Reads int
			}
		}
		if json.Unmarshal(raw, &l) != nil {
			t.Fatal("invalid ledger")
		}
		switch mutation {
		case 0:
			l.After++
		case 1:
			l.Selected.Environment = "other"
		case 2:
			l.Selected.FromMS = l.InitialQueries[0].FromMS
		case 3:
			for i := range l.Events {
				if l.Events[i].Action == "back" {
					l.Events[i].Reads++
					break
				}
			}
		case 4:
			l.Selected.BucketMS = 0
		}
		coherent := len(l.InitialQueries) == 2 && l.Initial == 2 && l.After == 3
		if coherent {
			expected := l.InitialQueries[1]
			expected.FromMS += 2 * expected.BucketMS
			expected.ToMS = expected.FromMS + expected.BucketMS
			expected.Offset = 0
			_, err := exploration.Compile(l.Selected)
			coherent = err == nil && reflect.DeepEqual(expected, l.Selected)
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
		t.Fatal("observed interval inspection rejected")
	}
	for i := 0; i < 5; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}
