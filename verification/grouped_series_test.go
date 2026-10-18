package verification

import (
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

type seriesEvent struct {
	label    int
	at       int64
	duration uint64
}
type seriesMachine struct {
	q      exploration.Query
	events []seriesEvent
}

func (m *seriesMachine) RuleAppend(tc hegel.TestCase) {
	m.events = append(m.events, seriesEvent{hegel.Draw(tc, hegel.Integers(0, 2)), int64(hegel.Draw(tc, hegel.Integers(0, 59))) * 1000, uint64(hegel.Draw(tc, hegel.Integers(0, 1000000000)))})
}
func (m *seriesMachine) RuleInterval(tc hegel.TestCase) {
	m.q.BucketMS = []int64{1000, 10000, 60000}[hegel.Draw(tc, hegel.Integers(0, 2))]
}
func (m *seriesMachine) RulePage(tc hegel.TestCase) {
	m.q.Limit = hegel.Draw(tc, hegel.Integers(1, 100))
	m.q.Offset = hegel.Draw(tc, hegel.Integers(0, 1000))
}
func (m *seriesMachine) InvariantAccounting(tc hegel.TestCase) {
	type key struct {
		label int
		at    int64
	}
	cells := map[key][]uint64{}
	for _, e := range m.events {
		k := key{e.label, m.q.FromMS + e.at/m.q.BucketMS*m.q.BucketMS}
		cells[k] = append(cells[k], e.duration)
	}
	raw := []json.RawMessage{}
	expectedSum := new(big.Int)
	for k, values := range cells {
		d := exploration.ReferenceDurationSummary(values)
		var label any
		if k.label == 1 {
			label = ""
		}
		if k.label == 2 {
			label = "/orders"
		}
		b, _ := json.Marshal(map[string]any{"g0": label, "from_ms": fmt.Sprint(k.at), "count": d.Count, "sum_ns": d.SumNS, "min_ns": d.MinNS, "max_ns": d.MaxNS, "quantiles_ns": []string{d.P50NS, d.P95NS, d.P99NS}})
		raw = append(raw, b)
		n, _ := new(big.Int).SetString(d.SumNS, 10)
		expectedSum.Add(expectedSum, n)
	}
	groups, e := exploration.DecodeGroupedTimeSeries(m.q, raw)
	if e != nil {
		panic(e)
	}
	total, sum := new(big.Int), new(big.Int)
	for _, g := range groups {
		for _, p := range g.Points {
			n, _ := new(big.Int).SetString(p.Count, 10)
			total.Add(total, n)
			n, _ = new(big.Int).SetString(p.Duration.SumNS, 10)
			sum.Add(sum, n)
			if p.Count == "0" && (p.State != "no_data" || p.Duration.State != "no_data") {
				panic("invented zero duration")
			}
		}
	}
	if total.String() != fmt.Sprint(len(m.events)) || sum.Cmp(expectedSum) != 0 {
		panic("full-population accounting changed with chart/page")
	}
}
func TestHegelGroupedSeriesStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		m := &seriesMachine{q: exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100, Summary: "span-duration", BucketMS: 1000, GroupTimeSeries: true, GroupBy: []exploration.Field{{Kind: "attribute", Key: "http.route"}}}}
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(25), hegel.WithAlwaysCheckInvariants("InvariantAccounting"))
	}, options(100)...)
}
