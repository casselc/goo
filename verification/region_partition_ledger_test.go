package verification

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"os"
	"reflect"
	"strconv"
	"testing"
)

func TestQuintRegionPartitionLedger(t *testing.T) {
	name := os.Getenv("GOO_REGION_PARTITION_LEDGER")
	if name == "" {
		t.Skip("observed browser ledger")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(bad int) bool {
		var l struct {
			Parent                             exploration.Query
			Pair                               struct{ Selection, Baseline exploration.Result }
			Initial, BeforeReturn, AfterReturn int
			Events                             []struct {
				Action       string
				Depth, Reads int
			}
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch bad {
		case 0:
			l.AfterReturn++
		case 1:
			l.Pair.Baseline.Question.Where = l.Pair.Selection.Question.Where
		case 2:
			l.Pair.Selection.Question.ToMS--
		case 3:
			l.Events[1].Reads++
		case 4:
			l.Pair.Baseline.Question.Environment = "other"
		}
		coherent := l.BeforeReturn == l.AfterReturn && l.Pair.Selection.Count == "2" && l.Pair.Baseline.Count == "3"
		for _, q := range []exploration.Query{l.Pair.Selection.Question, l.Pair.Baseline.Question} {
			want := l.Parent
			want.Where = q.Where
			want.Offset = 0
			_, e := exploration.Compile(q)
			coherent = coherent && e == nil && reflect.DeepEqual(want, q)
		}
		for _, ms := range []int64{0, 1000, 9999, 10000, 29999} {
			for _, d := range []uint64{0, 4999999, 5000000, 9999999, 10000000, 18446744073709551615} {
				for _, owner := range []string{"core", "outside"} {
					vals := map[exploration.Field]string{{Kind: "column", Key: "timestamp_ns"}: strconv.FormatInt((l.Parent.FromMS+ms)*1000000, 10), {Kind: "column", Key: "duration_ns"}: strconv.FormatUint(d, 10), {Kind: "attribute", Key: "owner"}: owner}
					p := exploration.Match(l.Parent.Where, vals)
					a, b := exploration.Match(l.Pair.Selection.Question.Where, vals), exploration.Match(l.Pair.Baseline.Question.Where, vals)
					coherent = coherent && !(a && b) && (a || b) == p && a == (p && ms < 10000 && d >= 5000000 && d < 10000000)
				}
			}
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		// Normalize the preceding fixed-question read into the existing drilldown protocol.
		rec.Record(context.Background(), "read", nil, "obs.reads", 1, "obs.depth", 0, "obs.coherent", coherent)
		for _, event := range l.Events {
			rec.Record(context.Background(), event.Action, nil, "obs.reads", event.Reads-l.Initial+1, "obs.depth", event.Depth, "obs.coherent", coherent)
		}
		checker, e := validate.NewChecker("model/population_drilldown.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		result, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return result.OK()
	}
	if !check(-1) {
		t.Fatal("valid region trace rejected")
	}
	for bad := 0; bad < 5; bad++ {
		if check(bad) {
			t.Fatal("forged region trace accepted", bad)
		}
	}
}
