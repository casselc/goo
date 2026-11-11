package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/scopedview"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type histogramCounters407 struct{ Source, Writes, Expand, Plans, SQL int }
type histogramFrame407 struct {
	Snapshot scopedview.Dataset
	Display  json.RawMessage
	Text     string
}
type histogramCase407 struct {
	Kind, Title                                        string
	SourceQuestion, DestinationQuestion                exploration.Query
	Source, Child, Destination, Reopened, SelectedRead customizeRead394
	Selected, Table, Complete, DestinationSnapshot     histogramFrame407
	Saved, DestinationSetup                            customizeSaved394
	Expansion                                          templateExpansion404
	ParentBefore, ParentAfter                          struct {
		Text, HTML string
		Same       bool
	}
	Plan struct {
		Request struct {
			Parent         exploration.Query
			FromBin, ToBin int
		}
		Response exploration.NumericHistogramSelection
		Status   int
	}
}
type histogramLedger407 struct {
	Version, Status string
	Fixture         struct {
		Token   string
		Records []struct {
			Signal, ID, TraceID, Environment, Namespace, Team, Service string
			TimeMS                                                     int64
			Raw                                                        *string `json:"raw"`
			LegacyOnly                                                 bool
		}
	}
	Raw      struct{ Logs, Spans []summaryRaw400 }
	Cases    []histogramCase407
	Controls []struct {
		Name     string
		Request  json.RawMessage
		Response json.RawMessage
		Status   int
	}
	Facts []struct {
		Name          string
		Before, After histogramCounters407
	}
	Counters   histogramCounters407
	ProxyCalls []literalWire393
	Errors     []string
}

var histogramNames407 = []string{"source-read", "customize-stage", "authored-read", "range-local", "table-local", "retired-range-handler", "complete-local", "save-template", "held-back", "reopen-definition", "reopened-read", "inspect-destination", "apply-destination", "retired-apply", "destination-read", "histogram-local", "destination-range-local", "selected-stage", "selected-read", "selected-held-back", "logs-read"}

func histogramEffect407(name string) histogramCounters407 {
	switch name {
	case "source-read", "authored-read", "reopened-read", "destination-read", "selected-read", "logs-read":
		return histogramCounters407{Source: 1, SQL: 3}
	case "save-template":
		return histogramCounters407{Writes: 1}
	case "inspect-destination":
		return histogramCounters407{Expand: 1}
	case "selected-stage":
		return histogramCounters407{Plans: 1}
	}
	return histogramCounters407{}
}
func histogramFacts407(r histogramLedger407) error {
	if len(r.Facts) != len(histogramNames407) {
		return fmt.Errorf("407 unfinished interaction prefix")
	}
	last := histogramCounters407{}
	for i, f := range r.Facts {
		if f.Name != histogramNames407[i] || f.Before != last {
			return fmt.Errorf("407 boundary sequence changed")
		}
		e := histogramEffect407(f.Name)
		want := histogramCounters407{Source: last.Source + e.Source, Writes: last.Writes + e.Writes, Expand: last.Expand + e.Expand, Plans: last.Plans + e.Plans, SQL: last.SQL + e.SQL}
		if f.After != want {
			return fmt.Errorf("407 unauthorized effects at %s", f.Name)
		}
		last = f.After
	}
	if r.Counters != last {
		return fmt.Errorf("407 final effect counters differ")
	}
	return nil
}

func histogramValues407(q exploration.Query, row summaryRaw400) map[exploration.Field]string {
	values := map[exploration.Field]string{{Kind: "column", Key: "service"}: row.ServiceName, {Kind: "column", Key: "span_id"}: row.SpanID, {Kind: "column", Key: "trace_id"}: row.TraceID}
	attrs := row.LogAttributes
	if q.Signal == "spans" {
		attrs = row.SpanAttributes
	}
	for k, v := range attrs {
		values[exploration.Field{Kind: "attribute", Key: k}] = v
	}
	for k, v := range row.ResourceAttributes {
		values[exploration.Field{Kind: "resource", Key: k}] = v
	}
	return values
}
func histogramPredicate407(p *exploration.Predicate, values map[exploration.Field]string, b exploration.NumericBinding) bool {
	if p == nil {
		return true
	}
	if len(p.All) > 0 {
		for i := range p.All {
			if !histogramPredicate407(&p.All[i], values, b) {
				return false
			}
		}
		return true
	}
	if len(p.Any) > 0 {
		for i := range p.Any {
			if histogramPredicate407(&p.Any[i], values, b) {
				return true
			}
		}
		return false
	}
	if p.Numeric == nil {
		return hdPredicate(p, values)
	}
	raw := numericHistogramLiveRecord{}
	if v, ok := values[b.Field]; ok {
		if b.Field.Kind == "attribute" {
			raw.Raw = &v
		} else {
			raw.ResourceRaw = &v
		}
	}
	ok, e := nhLiveMatch(p, raw, b)
	return e == nil && ok
}
func histogramMembers407(q exploration.Query, raw []summaryRaw400) ([]summaryRaw400, error) {
	base := q
	base.Where = nil
	rows, e := summaryMembers400(base, raw)
	if e != nil {
		return nil, e
	}
	out := []summaryRaw400{}
	if q.Numeric == nil || q.Numeric.Type != "int64" {
		return nil, fmt.Errorf("actual407 oracle requires int64 binding")
	}
	for _, row := range rows {
		if histogramPredicate407(q.Where, histogramValues407(q, row), *q.Numeric) {
			out = append(out, row)
		}
	}
	return out, nil
}
func histogramRead407(read customizeRead394, raw []summaryRaw400) error {
	q, r := read.Request, read.Response
	fail := func(s string) error { return fmt.Errorf("histogram407 read: %s", s) }
	if read.Status != 200 || !sameTransport384(q, r.Question) || r.Basis == "" || r.CapturedAt == "" || r.HistogramDataset == nil {
		return fail("missing accepted whole source carrier")
	}
	rows, e := histogramMembers407(q, raw)
	if e != nil {
		return e
	}
	if r.Count != strconv.Itoa(len(rows)) {
		return fail("raw full population differs")
	}
	values := []numericHistogramLiveRecord{}
	ids := map[string]bool{}
	for _, row := range rows {
		ids[row.SpanID] = true
		v := numericHistogramLiveRecord{}
		if text, ok := histogramValues407(q, row)[q.Numeric.Field]; ok {
			v.Raw = &text
			v.ResourceRaw = &text
		}
		values = append(values, v)
	}
	want, bins := nhLiveSummary(*q.Numeric, values)
	if r.Numeric == nil || !reflect.DeepEqual(*r.Numeric, want) || r.NumericHistogram == nil || r.NumericHistogram.Binning != "powers-of-two" || !sameTransport384(r.NumericHistogram.Bins, bins) {
		return fail("independent support/bins differ")
	}
	expected := min(q.Limit, max(0, len(rows)-q.Offset))
	if len(r.Rows) != expected || r.HasMore != (q.Offset+expected < len(rows)) {
		return fail("page count or continuation differs")
	}
	seen := map[string]bool{}
	for _, wire := range r.Rows {
		var row struct {
			SpanID string `json:"SpanId"`
		}
		if json.Unmarshal(wire, &row) != nil || !ids[row.SpanID] || seen[row.SpanID] {
			return fail("foreign or duplicate page row")
		}
		seen[row.SpanID] = true
	}
	d := r.HistogramDataset
	if d.Shape != "histogram" || d.ID != d.Scope.ID+":bins" || d.Revision != r.Basis || d.Scope.Generation != r.Basis || d.CapturedAt != r.CapturedAt || !sameTransport384(d.Scope.Selector, q) || d.Scope.Signal != q.Signal || d.Histogram == nil || !sameTransport384(d.Histogram.Binding, q.Numeric) || d.Histogram.Closure != "left-closed-right-open" || d.Histogram.Binning != "powers-of-two" || d.Quality.Extent != "complete" || d.Quality.Sampling != "unknown" || d.Completeness == nil || !reflect.DeepEqual(d.Completeness.Aspects, []string{"bins"}) || d.Completeness.ScopeID != d.Scope.ID || d.Completeness.Revision != r.Basis || len(d.Rows) != len(bins) {
		return fail("histogram identity, clocks or support guarantee differs")
	}
	s := d.Histogram.Support
	if s.Total != want.Count || s.Valid != want.Valid || s.Missing != want.Missing || s.Empty != want.Empty || s.Invalid != want.Invalid {
		return fail("supplied support differs")
	}
	for i, bin := range bins {
		row := d.Rows[i]
		state := "observed"
		if bin.Count == "0" {
			state = "no_data"
		}
		if row.Key != "bin:"+strconv.Itoa(i) || row.State != state || row.Values["from"] == nil || *row.Values["from"] != bin.From || row.Values["to"] == nil || *row.Values["to"] != bin.To || row.Values["count"] == nil || *row.Values["count"] != bin.Count {
			return fail("supplied exact row geometry/count differs")
		}
	}
	return nil
}

// The list of observed boundaries and their permitted effects is defined beside
// the journey once its concrete gestures are settled; no reported budget becomes
// its own oracle. Exact SQL is matched to existing compiled source plans.
func admitHistogram407(r histogramLedger407) error {
	fail := func(s string) error { return fmt.Errorf("histogram407: %s", s) }
	if r.Version != "goo.histogram407/1" || r.Status != "passed" || len(r.Errors) > 0 || len(r.Cases) != 1 || r.Fixture.Token == "" {
		return fail("incomplete journey")
	}
	if len(r.Fixture.Records) != 46 || len(r.Raw.Logs) != 23 || len(r.Raw.Spans) != 23 {
		return fail("incomplete physical source bank")
	}
	for signal, raw := range map[string][]summaryRaw400{"logs": r.Raw.Logs, "spans": r.Raw.Spans} {
		seen := map[string]bool{}
		for _, row := range raw {
			found := false
			for _, f := range r.Fixture.Records {
				if f.Signal != signal || f.ID != row.SpanID {
					continue
				}
				if seen[f.ID] {
					return fail("duplicate raw fixture")
				}
				seen[f.ID] = true
				attrs := row.LogAttributes
				if signal == "spans" {
					attrs = row.SpanAttributes
				}
				key := "deployment.environment.name"
				if f.LegacyOnly {
					key = "deployment.environment"
				}
				v, exists := attrs["work.items"]
				if row.TimestampNS != strconv.FormatInt(f.TimeMS*1000000, 10) || row.TraceID != f.TraceID || row.ServiceName != f.Service || row.ResourceAttributes["service.namespace"] != f.Namespace || row.ResourceAttributes[key] != f.Environment || attrs["team"] != f.Team || exists != (f.Raw != nil) || f.Raw != nil && v != *f.Raw {
					return fail("raw numeric/identity/source membership differs")
				}
				found = true
			}
			if !found {
				return fail("unplanned raw row")
			}
		}
	}
	c := r.Cases[0]
	if c.Kind != "histogram" || !sameTransport384(c.Source.Request, c.SourceQuestion) || !sameTransport384(c.Child.Request, c.SourceQuestion) || !sameTransport384(c.Destination.Request, c.DestinationQuestion) {
		return fail("whole question changed")
	}
	reads := []customizeRead394{c.Source, c.Child, c.Destination, c.Reopened, c.SelectedRead}
	for _, read := range reads {
		raw := r.Raw.Spans
		if read.Request.Signal == "logs" {
			raw = r.Raw.Logs
		}
		if e := histogramRead407(read, raw); e != nil {
			return e
		}
	}
	for _, control := range r.Controls {
		if control.Name == "incompatible-unit" && control.Status == 400 {
			continue
		}
		read := customizeRead394{Status: control.Status}
		if json.Unmarshal(control.Request, &read.Request) != nil || json.Unmarshal(control.Response, &read.Response) != nil {
			return fail("malformed source control")
		}
		raw := r.Raw.Spans
		if read.Request.Signal == "logs" {
			raw = r.Raw.Logs
		}
		if e := histogramRead407(read, raw); e != nil {
			return e
		}
		reads = append(reads, read)
	}
	if !sameTransport384(c.Selected.Snapshot, c.Child.Response.HistogramDataset) || !sameTransport384(c.Table.Snapshot, c.Selected.Snapshot) || !sameTransport384(c.Complete.Snapshot, c.Selected.Snapshot) || !sameTransport384(c.DestinationSnapshot.Snapshot, c.Destination.Response.HistogramDataset) {
		return fail("presentation replaced source dataset")
	}
	if e := templateStored404(c.Saved); e != nil {
		return e
	}
	if c.Saved.Response.Template == nil || c.Saved.Response.Template.Profile != "ordinary-histogram" || !sameTransport384(c.Saved.Response.Queries, []exploration.Query{c.SourceQuestion}) || c.Expansion.Status != 200 || !sameTransport384(c.Expansion.Request.Query, c.DestinationQuestion) || !sameTransport384(c.Expansion.Response.View.Queries, []exploration.Query{c.DestinationQuestion}) || !sameTransport384(c.Expansion.Response.View.Composition, c.Saved.Response.Composition) {
		return fail("saved/expanded presentation changed scope or profile")
	}
	if !c.ParentAfter.Same || c.ParentAfter.HTML != c.ParentBefore.HTML || c.ParentAfter.Text != c.ParentBefore.Text {
		return fail("held parent changed")
	}
	if !sameTransport384(c.Reopened.Request, c.SourceQuestion) {
		return fail("reopened whole source differs")
	}
	p := c.Plan.Response
	if c.Plan.Status != 200 || !sameTransport384(p.Parent, c.DestinationQuestion) || !sameTransport384(c.Plan.Request.Parent, c.DestinationQuestion) || c.Plan.Request.FromBin != p.FromBin || c.Plan.Request.ToBin != p.ToBin || !sameTransport384(c.SelectedRead.Request, p.Selected) {
		return fail("planner or selected read changed destination")
	}
	edges := nhLiveEdges(c.DestinationQuestion.Numeric.Type)
	if p.Binning != "powers-of-two" || p.FromBin < 0 || p.ToBin <= p.FromBin || p.ToBin >= len(edges) || p.From != edges[p.FromBin] || p.To != edges[p.ToBin] {
		return fail("planner exact bin geometry differs")
	}
	leaf := func(op, value string) exploration.Predicate {
		b := *c.DestinationQuestion.Numeric
		return exploration.Predicate{Numeric: &b, Op: op, Value: value}
	}
	parts := []exploration.Predicate{}
	if p.FromBin > 0 {
		parts = append(parts, leaf("gte", p.From))
	}
	if p.ToBin < len(edges)-1 {
		parts = append(parts, leaf("lt", p.To))
	}
	selected := leaf("valid", "")
	if len(parts) == 1 {
		selected = parts[0]
	} else if len(parts) > 1 {
		selected = exploration.Predicate{All: parts}
	}
	want := c.DestinationQuestion
	want.Offset = 0
	if want.Where != nil {
		want.Where = &exploration.Predicate{All: []exploration.Predicate{*want.Where, selected}}
	} else {
		want.Where = &selected
	}
	if !sameTransport384(want, p.Selected) {
		return fail("selected scope differs from exact independent bin predicate")
	}
	if len(r.Controls) != 2 || len(reads) != 6 {
		return fail("source/control count differs")
	}
	if e := histogramFacts407(r); e != nil {
		return e
	}
	if e := summaryPlans400(reads, r.ProxyCalls); e != nil {
		return e
	}
	return nil
}

func TestQuintHistogramViewObserved407(t *testing.T) {
	path := os.Getenv("GOO_HISTOGRAM_407_REPORT")
	if path == "" {
		t.Skip("actual407 histogram journey required")
	}
	wire, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var r histogramLedger407
	if e = json.Unmarshal(wire, &r); e != nil {
		t.Fatal(e)
	}
	if e = admitHistogram407(r); e != nil {
		t.Fatal(e)
	}
	for i, mutate := range []func(*histogramLedger407){func(r *histogramLedger407) { r.Cases[0].DestinationQuestion.Limit++ }, func(r *histogramLedger407) { r.Cases[0].Destination.Response.NumericHistogram.Bins[0].Count = "999" }, func(r *histogramLedger407) { r.Cases[0].Selected.Snapshot.Histogram.Support.Valid = "0" }, func(r *histogramLedger407) { r.Cases[0].ParentAfter.Same = false }, func(r *histogramLedger407) { r.ProxyCalls[0].Parameters["param_from"] = "0" }} {
		var bad histogramLedger407
		_ = json.Unmarshal(wire, &bad)
		mutate(&bad)
		if admitHistogram407(bad) == nil {
			t.Fatalf("semantic mutant %d escaped", i)
		}
	}
	if !replayHistogram407(t, r, false) || replayHistogram407(t, r, true) {
		t.Fatal("observed effect trace or planted local-read control incorrect")
	}
}

func replayHistogram407(t *testing.T, r histogramLedger407, mutate bool) bool {
	t.Helper()
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	costs := []int{}
	plans := []int{}
	for i, f := range r.Facts {
		effect := histogramEffect407(f.Name)
		costs = append(costs, effect.Source)
		plans = append(plans, effect.Plans)
		reads := f.After.Source
		if mutate && i == 3 {
			reads++
		}
		rec.Record(context.Background(), "step", nil, "obs.phase", i+1, "obs.reads", reads, "obs.plans", f.After.Plans)
	}
	spec, _ := filepath.Abs("model/histogram_view407.qnt")
	a, _ := json.Marshal(costs)
	b, _ := json.Marshal(plans)
	binding := fmt.Sprintf("spec: %q\nmodule: histogram_view407\nconstants: {READS: '%s', PLANS: '%s', LENGTH: '%d'}\ninvariants: [safety]\nactions:\n  step: {expect: 'phase == {{.obs.phase}} and reads == {{.obs.reads}} and plans == {{.obs.plans}}'}\n", spec, a, b, len(r.Facts))
	dir := t.TempDir()
	file := filepath.Join(dir, "binding.yaml")
	if e := os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	checker, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: dir})
	if e != nil {
		t.Fatal(e)
	}
	result, e := checker.Check(t.Context(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return result.OK()
}
