package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
	"math/big"
	"os"
	"reflect"
	"testing"
)

type groupedLedger struct {
	Result exploration.Result      `json:"result"`
	Saved  exploration.CapturePage `json:"saved"`
	Second exploration.CapturePage `json:"second"`
}
type durationGroup struct {
	Value    *string                     `json:"g0"`
	Count    string                      `json:"count"`
	Duration exploration.DurationSummary `json:"duration"`
}

func groupedCoherent(r exploration.Result) bool {
	if r.Count != "124" || r.Duration == nil || r.Duration.P95NS != "100000000" || r.Duration.SumNS != "2500000000" || len(r.Groups) != 4 {
		return false
	}
	count, sum := new(big.Int), new(big.Int)
	seen := map[string]bool{}
	for _, raw := range r.Groups {
		var g durationGroup
		if json.Unmarshal(raw, &g) != nil || g.Duration.Validate() != nil || g.Duration.Count != g.Count {
			return false
		}
		key := "missing"
		if g.Value != nil {
			key = "value:" + *g.Value
		}
		if seen[key] {
			return false
		}
		seen[key] = true
		n, ok := new(big.Int).SetString(g.Count, 10)
		if !ok {
			return false
		}
		count.Add(count, n)
		s, ok := new(big.Int).SetString(g.Duration.SumNS, 10)
		if !ok {
			return false
		}
		sum.Add(sum, s)
		expected := map[string][2]string{"missing": {"2", "200000000"}, "value:": {"2", "0"}, "value:/fast": {"100", "1000000"}, "value:/slow": {"20", "100000000"}}
		want, ok := expected[key]
		if !ok || g.Count != want[0] || g.Duration.P95NS != want[1] {
			return false
		}
	}
	return count.String() == r.Count && sum.String() == r.Duration.SumNS
}
func sameDurationGroups(a, b []json.RawMessage) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		var x, y durationGroup
		if json.Unmarshal(a[i], &x) != nil || json.Unmarshal(b[i], &y) != nil || !reflect.DeepEqual(x, y) {
			return false
		}
	}
	return true
}
func TestQuintGroupedMeasuresLiveLedger(t *testing.T) {
	name := os.Getenv("GOO_GROUPS_LEDGER")
	if name == "" {
		t.Skip("live ledger opt-in")
	}
	b, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	var l groupedLedger
	if json.Unmarshal(b, &l) != nil {
		t.Fatal("ledger")
	}
	check := func(l groupedLedger) (bool, string) {
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink, qobs.WithActor("grouped-measures"))
		rec.Record(context.Background(), "measure", nil, "obs.coherent", groupedCoherent(l.Result) && len(l.Result.Rows) == 100)
		rec.Record(context.Background(), "save", nil, "obs.coherent", l.Saved.Explore != nil && groupedCoherent(*l.Saved.Explore) && sameDurationGroups(l.Saved.Explore.Groups, l.Result.Groups))
		rec.Record(context.Background(), "page", nil, "obs.coherent", l.Second.Explore != nil && l.Second.Receipt.ID == l.Saved.Receipt.ID && l.Second.Explore.Basis == l.Saved.Explore.Basis && l.Second.Explore.Question.Offset == 100 && len(l.Second.Explore.Rows) == 24 && sameDurationGroups(l.Second.Explore.Groups, l.Saved.Explore.Groups))
		checker, e := validate.NewChecker("model/grouped_measures.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		report, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return report.OK(), report.Summary()
	}
	ok, summary := check(l)
	artifact("grouped-live.txt", []byte(summary))
	if !ok {
		t.Fatal(summary)
	}
	for i := 0; i < 3; i++ {
		var bad groupedLedger
		json.Unmarshal(b, &bad)
		switch i {
		case 0:
			bad.Result.Duration.P95NS = "50000000"
		case 1:
			bad.Saved.Explore.Groups = bad.Saved.Explore.Groups[:3]
		case 2:
			bad.Second.Explore.Groups[0] = json.RawMessage(`{"g0":"invented","count":"2"}`)
		}
		ok, summary := check(bad)
		artifact(fmt.Sprintf("grouped-negative-%d.txt", i), []byte(summary))
		if ok {
			t.Fatal("negative accepted", i)
		}
	}
}
func TestHegelGroupedDurationConservation(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		groups := make([][]uint64, 4)
		total := []uint64{}
		n := hegel.Draw(tc, hegel.Integers(0, 200))
		for i := 0; i < n; i++ {
			v := uint64(hegel.Draw(tc, hegel.Integers(0, 1000000000)))
			if hegel.Draw(tc, hegel.Integers(0, 10)) == 0 {
				v = 18446744073709551615
			}
			index := hegel.Draw(tc, hegel.Integers(0, 3))
			groups[index] = append(groups[index], v)
			total = append(total, v)
		}
		want := exploration.ReferenceDurationSummary(total)
		sum, count := new(big.Int), 0
		for _, values := range groups {
			d := exploration.ReferenceDurationSummary(values)
			if d.Validate() != nil {
				panic("invalid group")
			}
			v, _ := new(big.Int).SetString(d.SumNS, 10)
			sum.Add(sum, v)
			count += len(values)
		}
		if sum.String() != want.SumNS || count != len(total) {
			panic("group accounting diverged")
		}
	}, options(100)...)
}
