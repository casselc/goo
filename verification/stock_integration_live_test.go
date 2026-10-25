package verification

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
)

type siPoint struct {
	Metric, Unit, Namespace, Instance, StartNS, EndNS, Value string
	Attributes                                               map[string]string
	Flags, Temporality                                       int
	Monotonic                                                bool
}
type siLedger struct {
	Facts struct {
		Environment, FromNS, ToNS string
		Points                    []siPoint
	}
	Question                  measurements.Question
	Result                    measurements.Result
	Saved, Reopened           exploration.CapturePage
	StoredBefore, StoredAfter struct{ SHA256 string }
	Events                    []struct {
		Action string
		Source int
	}
}

func siKey(v any) string { b, _ := json.Marshal(v); return string(b) }
func siExpected(l siLedger) map[string]map[string]metrics.StockSample {
	grouped := map[string]map[string][]siPoint{}
	for _, p := range l.Facts.Points {
		if p.Metric != l.Question.Definition.Source.Metric {
			continue
		}
		stream := metrics.Stream{Resource: map[string]string{"deployment.environment.name": l.Facts.Environment, "service.name": "stock-worker", "service.namespace": p.Namespace, "service.instance.id": p.Instance, "telemetry.sdk.language": "python"}, Scope: "ordinary.stock", Version: "1", Name: p.Metric, Unit: p.Unit, Attributes: p.Attributes}
		key := siKey(stream)
		if grouped[key] == nil {
			grouped[key] = map[string][]siPoint{}
		}
		grouped[key][p.EndNS] = append(grouped[key][p.EndNS], p)
	}
	expected := map[string]map[string]metrics.StockSample{}
	to, _ := strconv.ParseUint(l.Facts.ToNS, 10, 64)
	age := uint64(l.Question.Definition.MaxGapMS) * 1000000
	for key, times := range grouped {
		expected[key] = map[string]metrics.StockSample{}
		for at, points := range times {
			p := points[0]
			s := metrics.StockSample{AtNS: at, StartNS: p.StartNS, State: "observed"}
			conflict := false
			for _, other := range points[1:] {
				if other.Value != p.Value || other.StartNS != p.StartNS || other.Flags != p.Flags {
					conflict = true
				}
			}
			switch {
			case conflict:
				s.StartNS = ""
				s.State = "unknown"
				s.Reason = "conflicting observations at one timestamp"
			case p.Flags == 1:
				s.State = "unknown"
				s.Reason = "no recorded value"
			default:
				v, _ := strconv.ParseFloat(p.Value, 64)
				s.Value = &v
				n, _ := strconv.ParseUint(at, 10, 64)
				if to-n > age {
					s.State = "stale"
					s.Reason = "older than declared freshness"
				}
			}
			expected[key][at] = s
		}
	}
	return expected
}
func siResultOK(l siLedger, r *measurements.Result) bool {
	if r == nil || r.Stocks == nil || r.Gauges != nil || r.Counters != nil || r.Histograms != nil || siKey(r.Question) != siKey(l.Question) || r.Stocks.FromNS != l.Facts.FromNS || r.Stocks.ToNS != l.Facts.ToNS {
		return false
	}
	expected := siExpected(l)
	if len(r.Stocks.Series) != len(expected) {
		return false
	}
	seen := map[string]bool{}
	for _, series := range r.Stocks.Series {
		key := siKey(series.Stream)
		want, ok := expected[key]
		if !ok || seen[key] || len(series.Samples) != len(want) {
			return false
		}
		seen[key] = true
		last := uint64(0)
		lastState := ""
		for _, s := range series.Samples {
			at, e := strconv.ParseUint(s.AtNS, 10, 64)
			if e != nil || at <= last || !reflect.DeepEqual(s, want[s.AtNS]) {
				return false
			}
			last = at
			lastState = s.State
		}
		if series.State != lastState {
			return false
		}
	}
	// Independently check the physical source rows as a multiset, including both
	// conflicting observations, rather than accepting only the derived sample.
	remaining := map[string]int{}
	n := 0
	for _, p := range l.Facts.Points {
		if p.Metric != l.Question.Definition.Source.Metric {
			continue
		}
		remaining[siKey([]any{p.Instance, p.Attributes, p.StartNS, p.EndNS, p.Value, p.Flags, p.Temporality, p.Monotonic})]++
		n++
	}
	if r.Count != strconv.Itoa(n) || len(r.Rows) != n {
		return false
	}
	for _, raw := range r.Rows {
		var row struct {
			Resource                   map[string]string
			Attributes                 map[string]string
			StartNS                    string `json:"start_ns"`
			EndNS                      string `json:"end_ns"`
			Value                      string
			Flags, Temporality         int
			Monotonic                  bool
			Name, Unit, Scope, Version string
			Schema                     string `json:"schema_url"`
		}
		if json.Unmarshal(raw, &row) != nil {
			return false
		}
		if row.Resource["deployment.environment.name"] != l.Facts.Environment || row.Resource["service.name"] != "stock-worker" || row.Resource["service.namespace"] != "inventory" || row.Name != l.Question.Definition.Source.Metric || row.Unit != "{connection}" || row.Scope != "ordinary.stock" || row.Version != "1" || row.Schema != "" {
			return false
		}
		key := siKey([]any{row.Resource["service.instance.id"], row.Attributes, row.StartNS, row.EndNS, row.Value, row.Flags, row.Temporality, row.Monotonic})
		if remaining[key] < 1 {
			return false
		}
		remaining[key]--
	}
	return true
}
func TestStockIntegrationObserved(t *testing.T) {
	path := os.Getenv("GOO_STOCK_INTEGRATION_LEDGER")
	if path == "" {
		t.Skip("actual195 stock ledger required")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var l siLedger
	if json.Unmarshal(raw, &l) != nil || !siResultOK(l, &l.Result) || l.Saved.Measurement == nil || !siResultOK(l, l.Saved.Measurement) || !reflect.DeepEqual(l.Saved, l.Reopened) {
		t.Fatal("actual raw stock/epoch/identity/reopen mismatch")
	}
	for i, event := range l.Events {
		if event.Action == "observation-selected" || event.Action == "frozen-reopened" {
			if i == 0 || event.Source != l.Events[i-1].Source {
				t.Fatal("local/reopen added read")
			}
		}
	}
	before, e := os.ReadFile(filepath.Join(filepath.Dir(path), "capture-before.json"))
	if e != nil {
		t.Fatal(e)
	}
	after, e := os.ReadFile(filepath.Join(filepath.Dir(path), "capture-after.json"))
	if e != nil {
		t.Fatal(e)
	}
	sum := sha256.Sum256(before)
	var stored exploration.SavedCapture
	if !reflect.DeepEqual(before, after) || hex.EncodeToString(sum[:]) != l.StoredBefore.SHA256 || l.StoredAfter.SHA256 != l.StoredBefore.SHA256 || json.Unmarshal(before, &stored) != nil || stored.ID != l.Saved.Receipt.ID || !siResultOK(l, stored.Measurement) || !peLiveCaptureHash(exploration.CapturePage{Receipt: stored.CaptureReceipt, Measurement: stored.Measurement}) {
		t.Fatal("full stock custody/hash")
	}
	for _, name := range []string{"value", "epoch", "identity", "state", "raw-row", "reopen"} {
		var m siLedger
		json.Unmarshal(raw, &m)
		switch name {
		case "value":
			for i := range m.Result.Stocks.Series {
				for j := range m.Result.Stocks.Series[i].Samples {
					if m.Result.Stocks.Series[i].Samples[j].Value != nil {
						*m.Result.Stocks.Series[i].Samples[j].Value += 1
					}
				}
			}
		case "epoch":
			m.Result.Stocks.Series[0].Samples[0].StartNS = "0"
		case "identity":
			m.Result.Stocks.Series[0].Stream.Resource["service.instance.id"] = "other"
		case "state":
			m.Result.Stocks.Series[0].Samples[0].State = "fabricated"
		case "raw-row":
			m.Result.Rows = m.Result.Rows[1:]
		case "reopen":
			m.Reopened.Receipt.Title += "changed"
		}
		if siResultOK(m, &m.Result) && reflect.DeepEqual(m.Saved, m.Reopened) {
			t.Fatal("changed control accepted", name)
		}
	}
	t.Log("actual physical stock rows, independent levels/epochs/identity/freshness and full immutable capture passed;6changed controls rejected;no source calls or new model")
}
