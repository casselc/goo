package verification

import (
	"context"
	"encoding/json"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"os"
	"testing"
)

func TestQuintTailLiveLedger(t *testing.T) {
	name := os.Getenv("GOO_TAIL_LEDGER")
	if name == "" {
		t.Skip("live opt-in")
	}
	data, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	type state struct {
		Phase                       string
		Retained, Evicted, Observed int
		Gap                         bool
		Keys                        []string
		Environment                 *string
	}
	type ledger struct {
		Environment                                          string
		Initial, Appended, Retained, Paused, Resumed, Scoped state
		PausedStable, FocusStable, StaticStable              bool
		Late, Duplicates                                     int
	}
	check := func(mutation int) bool {
		var l ledger
		if json.Unmarshal(data, &l) != nil {
			t.Fatal("ledger")
		}
		switch mutation {
		case 0:
			l.Initial.Environment = nil
		case 1:
			l.Duplicates = 1
		case 2:
			l.Retained.Retained = 101
		case 3:
			l.FocusStable = false
		case 4:
			l.Resumed.Gap = false
		case 5:
			l.Scoped.Retained = 1
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		record := func(a string, v bool) { rec.Record(context.Background(), a, nil, "obs.coherent", v) }
		valid := func(s state) bool {
			if len(s.Keys) != s.Retained || s.Retained > 100 || s.Observed != s.Retained+s.Evicted {
				return false
			}
			m := map[string]bool{}
			for _, k := range s.Keys {
				if m[k] {
					return false
				}
				m[k] = true
			}
			return true
		}
		record("connect", valid(l.Initial) && l.Initial.Environment != nil && *l.Initial.Environment == l.Environment && l.Initial.Retained == 3)
		record("receive_batch", valid(l.Appended) && l.Appended.Retained == 5 && l.Duplicates == 2 && l.Late >= 1 && l.StaticStable)
		record("retain", valid(l.Retained) && l.Retained.Retained == 100 && l.Retained.Evicted == 25)
		record("pause", valid(l.Paused) && l.Paused.Phase == "paused" && l.PausedStable && l.FocusStable)
		record("resume", valid(l.Resumed) && l.Resumed.Gap && l.Resumed.Environment != nil && *l.Resumed.Environment == l.Environment)
		record("scope", l.Scoped.Phase == "stopped" && l.Scoped.Retained == 0 && l.Scoped.Environment == nil)
		c, e := validate.NewChecker("model/tail.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("actual tail trace rejected")
	}
	for i := 0; i < 6; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}
