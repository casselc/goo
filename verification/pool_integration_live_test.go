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

type piPoint struct {
	siPoint
	SchemaURL string
}
type piFacts struct {
	Environment, Service, Namespace string
	Points                          []piPoint
}
type piState struct {
	Name, State, AtNS, StartNS, Reason string
	Value                              *float64
	Stream                             metrics.Stream
	Index                              int
}
type piGroup struct {
	Key, Pool, Instance string
	Stream              metrics.Stream
	States              []piState
}
type piModel struct {
	Groups     []piGroup
	Ineligible []struct{ Index int }
	Max, State string
}
type piLedger struct {
	Facts                piFacts
	Plan                 exploration.DashboardPlan
	WidgetResult, Native measurements.Result
	PoolModel            piModel
	DomRows              []struct {
		Pool, Instance string
		Cells          []struct{ PoolState, State, Value string }
	}
	Saved, Reopened           exploration.CapturePage
	StoredBefore, StoredAfter struct{ SHA256 string }
	InvalidDomainProbe        struct {
		GenericQuestion, DomainQuestion measurements.Question
		Generic                         measurements.Result
		Status                          int
	}
	MetadataControls []struct {
		Request  exploration.DashboardRequest
		Response exploration.DashboardPlan
	}
	Events []struct {
		Action        string
		Source, Plans int
	}
}

func piStream(f piFacts, p piPoint) metrics.Stream {
	return metrics.Stream{Resource: map[string]string{"service.name": f.Service, "service.namespace": p.Namespace, "service.instance.id": p.Instance, "deployment.environment.name": f.Environment, "telemetry.sdk.language": "python"}, Scope: "ordinary.pool", Version: "1", SchemaURL: p.SchemaURL, Name: p.Metric, Unit: p.Unit, Attributes: p.Attributes}
}
func piIdentity(s metrics.Stream) string {
	b, _ := json.Marshal(s)
	var obj map[string]any
	json.Unmarshal(b, &obj)
	attributes := map[string]string{}
	for k, v := range s.Attributes {
		if k != "db.client.connection.state" {
			attributes[k] = v
		}
	}
	obj["attributes"] = attributes
	return siKey(obj)
}
func piResult(f piFacts, r *measurements.Result, domain bool) bool {
	if r == nil || r.Stocks == nil || r.Question.Definition.Kind != "stock" || (r.Question.Definition.StockDomain == "nonnegative-integer") != domain {
		return false
	}
	binding := r.Question.Definition.Source
	if r.Question.Definition.Environment != f.Environment || len(binding.Services) != 1 || binding.Services[0] != f.Service || binding.Scope != "ordinary.pool" || binding.Version != "1" {
		return false
	}
	if domain && (binding.SchemaURL != "https://opentelemetry.io/schemas/1.37.0" || binding.Metric != "db.client.connection.count" || binding.Unit != "{connection}") {
		return false
	}
	groups := map[string]map[string][]piPoint{}
	physical := map[string]int{}
	count := 0
	from, to := r.Question.BoundsNS()
	lo, _ := strconv.ParseUint(from, 10, 64)
	hi, _ := strconv.ParseUint(to, 10, 64)
	for _, p := range f.Points {
		s := piStream(f, p)
		at, _ := strconv.ParseUint(p.EndNS, 10, 64)
		binding := r.Question.Definition.Source
		if p.Metric != binding.Metric || p.Unit != binding.Unit || p.SchemaURL != binding.SchemaURL || at < lo || at >= hi {
			continue
		}
		if binding.ExactStream != nil && siKey(s) != siKey(*binding.ExactStream) {
			continue
		}
		key := siKey(s)
		if groups[key] == nil {
			groups[key] = map[string][]piPoint{}
		}
		groups[key][p.EndNS] = append(groups[key][p.EndNS], p)
		physical[siKey([]any{s, p.StartNS, p.EndNS, p.Value, p.Flags, p.Temporality, p.Monotonic})]++
		count++
	}
	if r.Count != strconv.Itoa(count) || len(r.Rows) != count || len(r.Stocks.Series) != len(groups) || r.Stocks.FromNS != from || r.Stocks.ToNS != to {
		return false
	}
	for _, raw := range r.Rows {
		var row struct {
			Resource                          map[string]string
			Attributes                        map[string]string
			Scope, Version, Name, Unit, Value string
			Schema                            string `json:"schema_url"`
			StartNS                           string `json:"start_ns"`
			EndNS                             string `json:"end_ns"`
			Flags, Temporality                int
			Monotonic                         bool
		}
		if json.Unmarshal(raw, &row) != nil {
			return false
		}
		s := metrics.Stream{Resource: row.Resource, Attributes: row.Attributes, Scope: row.Scope, Version: row.Version, Name: row.Name, Unit: row.Unit, SchemaURL: row.Schema}
		key := siKey([]any{s, row.StartNS, row.EndNS, row.Value, row.Flags, row.Temporality, row.Monotonic})
		if physical[key] < 1 {
			return false
		}
		physical[key]--
	}
	seen := map[string]bool{}
	for _, series := range r.Stocks.Series {
		key := siKey(series.Stream)
		points, ok := groups[key]
		if !ok || seen[key] || len(points) != len(series.Samples) {
			return false
		}
		seen[key] = true
		last := uint64(0)
		latest := ""
		for _, sample := range series.Samples {
			set := points[sample.AtNS]
			if len(set) == 0 {
				return false
			}
			p := set[0]
			expected := metrics.StockSample{AtNS: p.EndNS, StartNS: p.StartNS, State: "observed"}
			conflict := false
			for _, v := range set[1:] {
				if v.Value != p.Value || v.StartNS != p.StartNS || v.Flags != p.Flags {
					conflict = true
				}
			}
			switch {
			case conflict:
				expected.StartNS = ""
				expected.State = "unknown"
				expected.Reason = "conflicting observations at one timestamp"
			case p.Flags == 1:
				expected.State = "unknown"
				expected.Reason = "no recorded value"
			default:
				v, e := strconv.ParseFloat(p.Value, 64)
				if e != nil {
					return false
				}
				expected.Value = &v
				at, _ := strconv.ParseUint(p.EndNS, 10, 64)
				if hi-at > uint64(r.Question.Definition.MaxGapMS)*1000000 {
					expected.State = "stale"
					expected.Reason = "older than declared freshness"
				}
			}
			at, _ := strconv.ParseUint(sample.AtNS, 10, 64)
			if at <= last || !reflect.DeepEqual(sample, expected) {
				return false
			}
			last = at
			latest = sample.State
		}
		if series.State != latest {
			return false
		}
	}
	return true
}
func piModelOK(l piLedger) bool {
	expected := map[string]map[string]piState{}
	ineligible := map[int]bool{}
	max := float64(0)
	for i, series := range l.WidgetResult.Stocks.Series {
		s := series.Stream
		pool, state := s.Attributes["db.client.connection.pool.name"], s.Attributes["db.client.connection.state"]
		if pool == "" || state == "" {
			ineligible[i] = true
			continue
		}
		key := piIdentity(s)
		if expected[key] == nil {
			expected[key] = map[string]piState{}
		}
		sample := series.Samples[len(series.Samples)-1]
		value := sample.Value
		if sample.State == "unknown" {
			value = nil
		}
		if value != nil && *value > max {
			max = *value
		}
		expected[key][state] = piState{Name: state, State: sample.State, AtNS: sample.AtNS, StartNS: sample.StartNS, Reason: sample.Reason, Value: value, Stream: s, Index: i}
	}
	if l.PoolModel.State != "observed" || len(l.PoolModel.Groups) != len(expected) || len(l.PoolModel.Ineligible) != len(ineligible) || l.PoolModel.Max != strconv.FormatFloat(max, 'f', -1, 64) {
		return false
	}
	for _, item := range l.PoolModel.Ineligible {
		if !ineligible[item.Index] {
			return false
		}
		delete(ineligible, item.Index)
	}
	display := map[string]map[string]piState{}
	for _, g := range l.PoolModel.Groups {
		states, ok := expected[g.Key]
		if !ok || g.Key != piIdentity(g.Stream) || g.Pool != g.Stream.Attributes["db.client.connection.pool.name"] || g.Instance != g.Stream.Resource["service.instance.id"] || len(states) != len(g.States) {
			return false
		}
		display[g.Pool+":"+g.Instance] = states
		for _, s := range g.States {
			if !reflect.DeepEqual(s, states[s.Name]) {
				return false
			}
		}
	}
	if len(l.DomRows) != len(expected) {
		return false
	}
	for _, row := range l.DomRows {
		states, ok := display[row.Pool+":"+row.Instance]
		if !ok {
			return false
		}
		for _, cell := range row.Cells {
			state, reported := states[cell.PoolState]
			value := ""
			wantState := "unavailable"
			if reported {
				wantState = state.State
				if state.Value != nil {
					value = strconv.FormatFloat(*state.Value, 'f', -1, 64)
				}
			}
			if cell.State != wantState || cell.Value != value {
				return false
			}
		}
	}
	return true
}
func piOK(l piLedger) bool {
	if !piResult(l.Facts, &l.WidgetResult, true) || !piResult(l.Facts, &l.Native, true) || !piModelOK(l) || !reflect.DeepEqual(l.Saved, l.Reopened) || !piResult(l.Facts, l.Saved.Measurement, true) {
		return false
	}
	found := false
	for _, w := range l.Plan.Widgets {
		if w.Family == "database-pool-connections" && w.Measurement != nil {
			found = siKey(*w.Measurement) == siKey(l.WidgetResult.Question)
		}
	}
	if !found || siKey(l.WidgetResult.Question) != siKey(l.Native.Question) || l.Saved.Measurement == nil || siKey(l.Saved.Measurement.Question) != siKey(l.Native.Question) {
		return false
	}
	if l.InvalidDomainProbe.Status != 502 || l.InvalidDomainProbe.DomainQuestion.Definition.StockDomain != "nonnegative-integer" || !piResult(l.Facts, &l.InvalidDomainProbe.Generic, false) {
		return false
	}
	if len(l.MetadataControls) != 2 {
		return false
	}
	for _, c := range l.MetadataControls {
		if len(c.Request.Instruments) != 1 {
			return false
		}
		coverage := c.Request.Instruments[0].CapabilityPlan
		if coverage == nil {
			return false
		}
		streams := coverage.Coverage.Streams
		if len(streams) > 25 {
			streams = streams[:25]
		}
		for _, observed := range streams {
			s := observed.Stream
			if s.Resource["service.instance.id"] != "" && s.Attributes["db.client.connection.pool.name"] != "" && s.Attributes["db.client.connection.state"] != "" {
				return false
			}
		}
		generic := false
		for _, w := range c.Response.Widgets {
			if w.Family == "database-pool-connections" {
				return false
			}
			if w.Measurement != nil && w.Measurement.Definition.Kind == "stock" {
				generic = true
			}
		}
		if !generic {
			return false
		}
	}
	for i, e := range l.Events {
		switch e.Action {
		case "local-filter-disclosure", "native-staged", "frozen-reopened":
			if i == 0 || e.Source != l.Events[i-1].Source {
				return false
			}
		}
	}
	return true
}
func TestPoolIntegrationObserved(t *testing.T) {
	path := os.Getenv("GOO_POOL_INTEGRATION_LEDGER")
	if path == "" {
		t.Skip("actual196 ledger required")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var l piLedger
	if json.Unmarshal(raw, &l) != nil || !piOK(l) {
		t.Fatal("actual pool raw/domain/model/DOM/custody rejected")
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
	if json.Unmarshal(before, &stored) != nil || stored.ID != l.Saved.Receipt.ID || !reflect.DeepEqual(before, after) || hex.EncodeToString(sum[:]) != l.StoredBefore.SHA256 || l.StoredAfter.SHA256 != l.StoredBefore.SHA256 || !piResult(l.Facts, stored.Measurement, true) || !reflect.DeepEqual(stored.CaptureReceipt, l.Saved.Receipt) || !peLiveCaptureHash(exploration.CapturePage{Receipt: stored.CaptureReceipt, Measurement: stored.Measurement}) {
		t.Fatal("full opted-domain capture bytes/hash")
	}
	for _, name := range []string{"domain", "capture-question", "drop-ineligible", "merge-instance", "zero-missing", "receipt"} {
		var m piLedger
		json.Unmarshal(raw, &m)
		switch name {
		case "domain":
			m.WidgetResult.Question.Definition.StockDomain = ""
		case "capture-question":
			m.Saved.Measurement.Question.Definition.Title += "changed"
			m.Reopened.Measurement.Question.Definition.Title += "changed"
		case "drop-ineligible":
			m.PoolModel.Ineligible = nil
		case "merge-instance":
			m.PoolModel.Groups[0].Instance = "merged"
		case "zero-missing":
			for i := range m.PoolModel.Groups {
				for j := range m.PoolModel.Groups[i].States {
					if m.PoolModel.Groups[i].States[j].Value == nil {
						z := float64(0)
						m.PoolModel.Groups[i].States[j].Value = &z
					}
				}
			}
		case "receipt":
			m.Reopened.Receipt.Title += "changed"
		}
		if piOK(m) {
			t.Fatal("changed pool control accepted", name)
		}
	}
	t.Log("actual raw pool streams/domain/state model/DOM/epochs and full capture passed;6changed controls rejected;no source or new model")
}
