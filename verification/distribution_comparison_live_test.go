package verification

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type dcLiveRecord struct {
	Raw, ResourceRaw                *string
	Service, Population, ID, Signal string
	TimeMS                          int64 `json:"timeMs"`
}
type dcLivePair struct {
	Kind                string
	Baseline, Selection exploration.Result
	Request             exploration.ComparisonRequest
	Response            exploration.CohortComparison
}
type dcLivePlan struct {
	Side           string
	FromBin, ToBin int
	Request        struct {
		Comparison     exploration.ComparisonRequest
		Side           string
		FromBin, ToBin int
	}
	Response exploration.NumericHistogramSelection
	Result   exploration.Result
}
type dcLiveView struct {
	Version    int
	ID         string
	Queries    []exploration.Query
	Breakdowns map[string]struct {
		Baseline     *exploration.Query
		Distribution bool
	}
}
type dcLiveViewRun struct {
	Results  []exploration.Result
	Request  exploration.ComparisonRequest
	Response exploration.CohortComparison
}
type dcLiveLedger struct {
	Version             int
	Environment         string
	Binding             exploration.NumericBinding
	Records             []dcLiveRecord
	Baseline, Selection exploration.Result
	Compare, Logs       dcLivePair
	Extremes, EdgeCases []dcLivePair
	Aggregate           struct {
		dcLivePair
		Plans []dcLivePlan
	}
	Plans                                                    []dcLivePlan
	Saved, Reopened, OldNumeric, OldNumericReopened          exploration.CapturePage
	SeedView, View                                           dcLiveView
	InitialView, Reused                                      dcLiveViewRun
	Decision, OldDecision                                    exploration.EvidenceComparison
	Receipt, ReopenedReceipt, OldReceipt, OldReceiptReopened workbench.ComparisonReceipt
	Refusals                                                 []struct {
		Kind   string
		Status int
		Body   json.RawMessage
	}
	Events []struct {
		Action                     string
		Reads, Planners, Protocols int
	}
	DomModel struct{ Count, Share, Saved dcLiveDOM }
}
type dcLiveDOMBin struct {
	Key, From, To  string
	Counts, Shares []string
}
type dcLiveDOM struct {
	Data struct {
		Distribution     exploration.NumericDistributionComparison
		Type, Unit, Mode string
		Valid            []string
		Interactive      bool
	}
	Model struct {
		State, Mode, Ceiling string
		All, Points          []dcLiveDOMBin
		Totals               []string
		Empty, Interactive   bool
	}
	Sides []struct {
		Ceiling, ValueFormat string
		Interactive          bool
		Points               []struct {
			Key   string
			Value *string
			State string
		}
	}
}

// These expected answers use only emitted fixture records, stdlib canonical
// decimal roundtrip, arbitrary-precision integers and exact rational arithmetic.
// The shared nhLive*/ncLiveMeasure helpers are independent test oracles; no goo
// evaluator, parser, compiler, comparison or selection planner predicts a result.
func dcLiveRaw(f exploration.Field, row dcLiveRecord) (*string, error) {
	switch {
	case f.Kind == "attribute" && f.Key == "work.items":
		return row.Raw, nil
	case f.Kind == "resource" && f.Key == "work.extreme":
		return row.ResourceRaw, nil
	case f.Kind == "attribute" && f.Key == "work.population":
		return &row.Population, nil
	case f.Kind == "column" && f.Key == "service":
		return &row.Service, nil
	}
	return nil, fmt.Errorf("fixture does not define %+v", f)
}
func dcLiveMatch(p *exploration.Predicate, row dcLiveRecord) (bool, error) {
	if p == nil {
		return true, nil
	}
	if len(p.All) > 0 {
		for i := range p.All {
			ok, e := dcLiveMatch(&p.All[i], row)
			if e != nil || !ok {
				return false, e
			}
		}
		return true, nil
	}
	if len(p.Any) > 0 {
		for i := range p.Any {
			ok, e := dcLiveMatch(&p.Any[i], row)
			if e != nil {
				return false, e
			}
			if ok {
				return true, nil
			}
		}
		return false, nil
	}
	f := p.Field
	if p.Numeric != nil {
		f = &p.Numeric.Field
	}
	if f == nil {
		return false, fmt.Errorf("missing field")
	}
	raw, e := dcLiveRaw(*f, row)
	if e != nil {
		return false, e
	}
	if p.Numeric != nil {
		n, state := nhLiveInteger(raw, p.Numeric.Type)
		switch p.Op {
		case "valid", "missing", "empty", "invalid":
			return state == p.Op, nil
		}
		if state != "valid" {
			return false, nil
		}
		threshold, s := nhLiveInteger(&p.Value, p.Numeric.Type)
		if s != "valid" {
			return false, fmt.Errorf("invalid threshold")
		}
		cmp := n.Cmp(threshold)
		switch p.Op {
		case "eq":
			return cmp == 0, nil
		case "ne":
			return cmp != 0, nil
		case "lt":
			return cmp < 0, nil
		case "lte":
			return cmp <= 0, nil
		case "gt":
			return cmp > 0, nil
		case "gte":
			return cmp >= 0, nil
		}
	} else {
		switch p.Op {
		case "exists":
			return raw != nil, nil
		case "missing":
			return raw == nil, nil
		}
		if raw == nil {
			return false, nil
		}
		switch p.Op {
		case "eq":
			return *raw == p.Value, nil
		case "ne":
			return *raw != p.Value, nil
		case "contains":
			return strings.Contains(*raw, p.Value), nil
		}
	}
	return false, fmt.Errorf("unsupported fixture predicate %s", p.Op)
}
func dcLiveRows(q exploration.Query, l dcLiveLedger) ([]dcLiveRecord, error) {
	if q.Environment != l.Environment || q.Signal != "spans" && q.Signal != "logs" || q.ToMS <= q.FromMS {
		return nil, fmt.Errorf("wrong fixture scope")
	}
	rows := []dcLiveRecord{}
	for _, r := range l.Records {
		if r.Signal != q.Signal || r.TimeMS < q.FromMS || r.TimeMS >= q.ToMS {
			continue
		}
		ok, e := dcLiveMatch(q.Where, r)
		if e != nil {
			return nil, e
		}
		if ok {
			rows = append(rows, r)
		}
	}
	return rows, nil
}
func dcLiveSummary(b exploration.NumericBinding, rows []dcLiveRecord) (exploration.NumericSummary, []exploration.NumericHistogramBin) {
	input := make([]numericHistogramLiveRecord, len(rows))
	for i, r := range rows {
		input[i] = numericHistogramLiveRecord{Raw: r.Raw, ResourceRaw: r.ResourceRaw, TimeMS: r.TimeMS, ID: r.ID}
	}
	return nhLiveSummary(b, input)
}
func dcLiveBinding(b *exploration.NumericBinding, l dcLiveLedger) bool {
	return b != nil && (b.Field == (exploration.Field{Kind: "attribute", Key: "work.items"}) || b.Field == (exploration.Field{Kind: "resource", Key: "work.extreme"})) && (b.Type == "int64" || b.Type == "uint64") && b.Parse == "canonical-decimal" && b.Unit == l.Binding.Unit
}
func dcLiveCohortOf(r exploration.Result) exploration.Cohort {
	c := ncLiveCohortOf(r)
	c.NumericHistogram = r.NumericHistogram
	return c
}
func dcLiveCohort(c exploration.Cohort, l dcLiveLedger) bool {
	rows, e := dcLiveRows(c.Question, l)
	if e != nil || !dcLiveBinding(c.Question.Numeric, l) || c.Question.NumericHistogram == nil || c.Question.NumericHistogram.Binning != "powers-of-two" || c.Count != strconv.Itoa(len(rows)) || c.Basis == "" || len(c.Question.GroupBy) > 1 || len(c.Question.GroupBy) == 1 && c.Question.GroupBy[0] != (exploration.Field{Kind: "column", Key: "service"}) {
		return false
	}
	if c.State != "changed_between_reads" && (len(rows) > 0 && c.State != "observed" || len(rows) == 0 && c.State != "no_data") {
		return false
	}
	summary, bins := dcLiveSummary(*c.Question.Numeric, rows)
	if c.Numeric == nil || *c.Numeric != summary || c.NumericHistogram == nil || c.NumericHistogram.Binning != "powers-of-two" || !reflect.DeepEqual(c.NumericHistogram.Bins, bins) {
		return false
	}
	if len(c.Question.GroupBy) == 0 {
		return len(c.Groups) == 0
	}
	groups := map[string][]dcLiveRecord{}
	for _, r := range rows {
		groups[r.Service] = append(groups[r.Service], r)
	}
	if len(groups) != len(c.Groups) {
		return false
	}
	seen := map[string]bool{}
	for _, g := range c.Groups {
		if g.Value == nil || seen[*g.Value] {
			return false
		}
		seen[*g.Value] = true
		members, ok := groups[*g.Value]
		if !ok || g.Count != strconv.Itoa(len(members)) {
			return false
		}
		summary, _ := dcLiveSummary(*c.Question.Numeric, members)
		if g.Numeric == nil || *g.Numeric != summary {
			return false
		}
	}
	return true
}
func dcLiveResult(r exploration.Result, l dcLiveLedger) bool {
	if r.CapturedAt == "" || !dcLiveCohort(dcLiveCohortOf(r), l) {
		return false
	}
	rows, e := dcLiveRows(r.Question, l)
	if e != nil {
		return false
	}
	n := len(rows) - r.Question.Offset
	if n < 0 {
		n = 0
	}
	if n > r.Question.Limit {
		n = r.Question.Limit
	}
	if len(r.Rows) != n || r.HasMore != (len(rows) > r.Question.Offset+n) {
		return false
	}
	fixture := map[int64]dcLiveRecord{}
	for _, row := range rows {
		if _, ok := fixture[row.TimeMS]; ok {
			return false
		}
		fixture[row.TimeMS] = row
	}
	seen := map[int64]bool{}
	for _, raw := range r.Rows {
		var row struct {
			Timestamp  string            `json:"timestamp_ns"`
			Service    string            `json:"ServiceName"`
			ID         string            `json:"SpanId"`
			Attributes map[string]string `json:"Attributes"`
			Resource   map[string]string `json:"ResourceAttributes"`
		}
		if json.Unmarshal(raw, &row) != nil {
			return false
		}
		ns, e := strconv.ParseInt(row.Timestamp, 10, 64)
		if e != nil {
			return false
		}
		source, ok := fixture[ns/1000000]
		if !ok || seen[source.TimeMS] || ns != source.TimeMS*1000000 || source.Service != row.Service || r.Question.Signal == "spans" && source.ID != row.ID {
			return false
		}
		seen[source.TimeMS] = true
		for _, pair := range []struct {
			key    string
			values map[string]string
			raw    *string
		}{{"work.items", row.Attributes, source.Raw}, {"work.extreme", row.Resource, source.ResourceRaw}} {
			v, present := pair.values[pair.key]
			if pair.raw == nil && present || pair.raw != nil && (!present || v != *pair.raw) {
				return false
			}
		}
		if row.Attributes["work.population"] != source.Population || row.Resource["deployment.environment.name"] != l.Environment {
			return false
		}
	}
	return r.Plan.Count != "" && !strings.Contains(strings.ToUpper(r.Plan.Count), " LIMIT ") && r.Plan.Parameters["environment"] == l.Environment && r.Plan.Parameters["from"] == strconv.FormatInt(r.Question.FromMS, 10) && r.Plan.Parameters["to"] == strconv.FormatInt(r.Question.ToMS, 10)
}
func dcLiveDistribution(a, z exploration.Cohort) exploration.NumericDistributionComparison {
	unknown := a.State == "changed_between_reads" || z.State == "changed_between_reads"
	d := exploration.NumericDistributionComparison{State: "comparable", Binning: "powers-of-two", Denominator: "population-count", BaselineCount: a.Count, SelectionCount: z.Count, Bins: []exploration.NumericDistributionBin{}}
	if unknown {
		d.State = "unknown"
		d.Reason = "source-reads-disagree"
	} else if a.Count == "0" || z.Count == "0" {
		d.State = "unavailable"
		d.Reason = "empty-population"
	}
	for i, x := range a.NumericHistogram.Bins {
		y := z.NumericHistogram.Bins[i]
		bin := exploration.NumericDistributionBin{From: x.From, To: x.To, BaselineCount: x.Count, SelectionCount: y.Count}
		if !unknown {
			bin.CountDifference = ncLiveDifference(x.Count, y.Count)
			if a.Count != "0" {
				bin.BaselinePercent = new(big.Rat).Mul(new(big.Rat).SetFrac(ncLiveInt(x.Count), ncLiveInt(a.Count)), big.NewRat(100, 1)).RatString()
			}
			if z.Count != "0" {
				bin.SelectionPercent = new(big.Rat).Mul(new(big.Rat).SetFrac(ncLiveInt(y.Count), ncLiveInt(z.Count)), big.NewRat(100, 1)).RatString()
			}
			if a.Count != "0" && z.Count != "0" {
				ap, _ := new(big.Rat).SetString(bin.BaselinePercent)
				zp, _ := new(big.Rat).SetString(bin.SelectionPercent)
				bin.DifferencePP = new(big.Rat).Sub(zp, ap).RatString()
			}
		}
		d.Bins = append(d.Bins, bin)
	}
	return d
}
func dcLivePairCheck(request exploration.ComparisonRequest, response exploration.CohortComparison, l dcLiveLedger) bool {
	a, z := request.Baseline, request.Selection
	if !request.Numeric || !request.Distribution || !dcLiveCohort(a, l) || !dcLiveCohort(z, l) || *a.Question.Numeric != *z.Question.Numeric || a.Question.Signal != z.Question.Signal || a.Question.Environment != z.Question.Environment || !reflect.DeepEqual(a.Question.GroupBy, z.Question.GroupBy) {
		return false
	}
	unknown := a.State == "changed_between_reads" || z.State == "changed_between_reads"
	state, reason := "ok", ""
	if unknown {
		state = "unknown"
		reason = "source-reads-disagree"
	} else if a.Count == "0" || z.Count == "0" {
		state = "no_data"
		reason = "empty-population"
	}
	field := exploration.Field{}
	aggregate := len(a.Question.GroupBy) == 0
	if !aggregate {
		field = a.Question.GroupBy[0]
	}
	if response.AggregateOnly != aggregate || response.State != state || response.Reason != reason || response.Field != field || response.BaselineCount != a.Count || response.SelectionCount != z.Count || response.BaselineBasis != a.Basis || response.SelectionBasis != z.Basis || len(response.Basis) != 64 || response.Measures == nil || response.Measures.Numeric == nil {
		return false
	}
	want := ncLiveMeasure(*a.Question.Numeric, a.Numeric, z.Numeric, unknown)
	d := dcLiveDistribution(a, z)
	want.Distribution = &d
	if !reflect.DeepEqual(*response.Measures.Numeric, want) {
		return false
	}
	if state != "ok" || aggregate {
		return len(response.Rows) == 0
	}
	ag, zg := map[string]exploration.ValueCount{}, map[string]exploration.ValueCount{}
	keys := map[string]bool{}
	for _, g := range a.Groups {
		ag[*g.Value] = g
		keys[*g.Value] = true
	}
	for _, g := range z.Groups {
		zg[*g.Value] = g
		keys[*g.Value] = true
	}
	if len(response.Rows) != len(keys) {
		return false
	}
	seen := map[string]bool{}
	for _, row := range response.Rows {
		if row.Value == nil || !keys[*row.Value] || seen[*row.Value] {
			return false
		}
		seen[*row.Value] = true
		x, y := ag[*row.Value], zg[*row.Value]
		ac, zc := x.Count, y.Count
		if ac == "" {
			ac = "0"
		}
		if zc == "" {
			zc = "0"
		}
		ap := new(big.Rat).Mul(new(big.Rat).SetFrac(ncLiveInt(ac), ncLiveInt(a.Count)), big.NewRat(100, 1))
		zp := new(big.Rat).Mul(new(big.Rat).SetFrac(ncLiveInt(zc), ncLiveInt(z.Count)), big.NewRat(100, 1))
		diff := new(big.Rat).Sub(zp, ap).FloatString(2)
		if diff == "-0.00" {
			diff = "0.00"
		}
		if row.BaselineCount != ac || row.SelectionCount != zc || row.BaselinePercent != ap.FloatString(2) || row.SelectionPercent != zp.FloatString(2) || row.DifferencePP != diff || row.Measures == nil || row.Measures.Numeric == nil {
			return false
		}
		want := ncLiveMeasure(*a.Question.Numeric, x.Numeric, y.Numeric, false)
		if !reflect.DeepEqual(*row.Measures.Numeric, want) {
			return false
		}
	}
	return true
}
func dcLiveActualPair(p dcLivePair, l dcLiveLedger) bool {
	a := dcLiveCohortOf(p.Baseline)
	if p.Kind == "unknown" {
		a.State = "changed_between_reads"
	}
	return dcLiveResult(p.Baseline, l) && dcLiveResult(p.Selection, l) && reflect.DeepEqual(p.Request.Baseline, a) && reflect.DeepEqual(p.Request.Selection, dcLiveCohortOf(p.Selection)) && dcLivePairCheck(p.Request, p.Response, l)
}
func dcLivePlanScope(p dcLivePlan, l dcLiveLedger) bool {
	in, s := p.Request, p.Response
	if in.Side != p.Side || in.FromBin != p.FromBin || in.ToBin != p.ToBin || !dcLiveCohort(in.Comparison.Baseline, l) || !dcLiveCohort(in.Comparison.Selection, l) || !in.Comparison.Distribution || !in.Comparison.Numeric {
		return false
	}
	c := in.Comparison.Baseline
	if p.Side == "selection" {
		c = in.Comparison.Selection
	} else if p.Side != "baseline" {
		return false
	}
	if c.State == "changed_between_reads" || in.Comparison.Baseline.State == "changed_between_reads" || in.Comparison.Selection.State == "changed_between_reads" {
		return false
	}
	edges := nhLiveEdges(c.Question.Numeric.Type)
	if p.FromBin < 0 || p.ToBin <= p.FromBin || p.ToBin >= len(edges) || s.FromBin != p.FromBin || s.ToBin != p.ToBin || s.From != edges[p.FromBin] || s.To != edges[p.ToBin] || s.Binning != "powers-of-two" || !nhLiveSameQuery(s.Parent, c.Question) {
		return false
	}
	n := new(big.Int)
	for _, bin := range c.NumericHistogram.Bins[p.FromBin:p.ToBin] {
		n.Add(n, ncLiveInt(bin.Count))
	}
	if n.Sign() == 0 {
		return false
	}
	for _, q := range []exploration.Query{s.Selected, s.Rest} {
		want := c.Question
		want.Where = q.Where
		want.Offset = 0
		if !nhLiveSameQuery(want, q) {
			return false
		}
	}
	for i, q := range []exploration.Query{s.Comparison.Selection, s.Comparison.Baseline} {
		want := s.Selected
		if i == 1 {
			want = s.Rest
		}
		want.Numeric = nil
		want.NumericHistogram = nil
		want.Summary = ""
		want.SpanStatus = false
		want.GroupBy = []exploration.Field{{Kind: "column", Key: "service"}}
		if len(c.Question.GroupBy) > 0 {
			want.GroupBy = []exploration.Field{c.Question.GroupBy[0]}
		}
		if !nhLiveSameQuery(want, q) {
			return false
		}
	}
	if s.NumericComparison == nil {
		return false
	}
	for i, q := range []exploration.Query{s.NumericComparison.Selection, s.NumericComparison.Baseline} {
		want := s.Selected
		if i == 1 {
			want = s.Rest
		}
		want.NumericHistogram = nil
		want.Summary = ""
		want.SpanStatus = false
		want.GroupBy = []exploration.Field{{Kind: "column", Key: "service"}}
		if len(c.Question.GroupBy) > 0 {
			want.GroupBy = []exploration.Field{c.Question.GroupBy[0]}
		}
		if !nhLiveSameQuery(want, q) {
			return false
		}
	}
	lo, hi := ncLiveInt(s.From), ncLiveInt(s.To)
	rows := append([]dcLiveRecord{}, l.Records...)
	for _, v := range append(edges, "", "01", "9007199254740993") {
		raw := v
		rows = append(rows, dcLiveRecord{Raw: &raw, ResourceRaw: &raw, Population: "baseline"}, dcLiveRecord{Raw: &raw, ResourceRaw: &raw, Population: "selection"})
	}
	rows = append(rows, dcLiveRecord{Population: "baseline"}, dcLiveRecord{Population: "selection"})
	for _, r := range rows {
		parent, e := dcLiveMatch(c.Question.Where, r)
		if e != nil {
			return false
		}
		a, e := dcLiveMatch(s.Selected.Where, r)
		if e != nil {
			return false
		}
		z, e := dcLiveMatch(s.Rest.Where, r)
		if e != nil {
			return false
		}
		raw, e := dcLiveRaw(c.Question.Numeric.Field, r)
		if e != nil {
			return false
		}
		value, state := nhLiveInteger(raw, c.Question.Numeric.Type)
		want := parent && state == "valid" && value.Cmp(lo) >= 0 && value.Cmp(hi) < 0
		if a != want || z != (parent && !want) || a && z {
			return false
		}
	}
	return true
}
func dcLivePlanCheck(p dcLivePlan, l dcLiveLedger) bool {
	return dcLivePlanScope(p, l) && dcLiveResult(p.Result, l) && nhLiveSameQuery(p.Result.Question, p.Response.Selected)
}

func dcLiveDOMCheck(view dcLiveDOM, request exploration.ComparisonRequest, mode string, interactive bool) bool {
	a, z := request.Baseline, request.Selection
	want := dcLiveDistribution(a, z)
	data, model := view.Data, view.Model
	if !reflect.DeepEqual(data.Distribution, want) || data.Type != a.Question.Numeric.Type || data.Unit != a.Question.Numeric.Unit || data.Mode != mode || data.Interactive != interactive || !reflect.DeepEqual(data.Valid, []string{a.Numeric.Valid, z.Numeric.Valid}) || model.State != want.State || model.Mode != mode || !reflect.DeepEqual(model.Totals, []string{a.Count, z.Count}) || model.Interactive != (interactive && want.State != "unknown") {
		return false
	}
	all := make([]dcLiveDOMBin, len(want.Bins))
	first, last := -1, -1
	max := new(big.Int)
	for i, b := range want.Bins {
		all[i] = dcLiveDOMBin{Key: strconv.Itoa(i), From: b.From, To: b.To, Counts: []string{b.BaselineCount, b.SelectionCount}, Shares: []string{b.BaselinePercent, b.SelectionPercent}}
		if b.BaselineCount != "0" || b.SelectionCount != "0" {
			if first < 0 {
				first = i
			}
			last = i
		}
		for _, n := range all[i].Counts {
			v := ncLiveInt(n)
			if v.Cmp(max) > 0 {
				max.Set(v)
			}
		}
	}
	points := []dcLiveDOMBin{}
	if first >= 0 {
		points = all[first : last+1]
	}
	ceiling := max.String()
	if mode == "share" {
		ceiling = "100"
	}
	if !reflect.DeepEqual(model.All, all) || !reflect.DeepEqual(model.Points, points) || model.Ceiling != ceiling || model.Empty != (first < 0) || len(view.Sides) != 2 {
		return false
	}
	for side, chart := range view.Sides {
		if chart.Ceiling != ceiling || chart.Interactive != model.Interactive || len(chart.Points) != len(points) || mode == "share" && chart.ValueFormat != "rational" {
			return false
		}
		for i, p := range chart.Points {
			v := points[i].Counts[side]
			if mode == "share" {
				v = points[i].Shares[side]
			}
			if p.Key != points[i].Key || p.Value == nil || *p.Value != v || p.State != "observed" {
				return false
			}
		}
	}
	return true
}
func dcLiveViewRunCheck(v dcLiveViewRun, definition dcLiveView, l dcLiveLedger) bool {
	base := definition.Breakdowns["spans"].Baseline
	if len(definition.ID) != 64 || len(definition.Queries) != 1 || base == nil || !definition.Breakdowns["spans"].Distribution || !nhLiveSameQuery(v.Request.Baseline.Question, *base) || !nhLiveSameQuery(v.Request.Selection.Question, definition.Queries[0]) || len(v.Results) != 2 {
		return false
	}
	found := map[string]bool{}
	for _, r := range v.Results {
		if !dcLiveResult(r, l) {
			return false
		}
		switch {
		case nhLiveSameQuery(r.Question, v.Request.Baseline.Question):
			if found["baseline"] || !reflect.DeepEqual(dcLiveCohortOf(r), v.Request.Baseline) {
				return false
			}
			found["baseline"] = true
		case nhLiveSameQuery(r.Question, v.Request.Selection.Question):
			if found["selection"] || !reflect.DeepEqual(dcLiveCohortOf(r), v.Request.Selection) {
				return false
			}
			found["selection"] = true
		default:
			return false
		}
	}
	return found["baseline"] && found["selection"] && dcLivePairCheck(v.Request, v.Response, l)
}
func dcLiveNCFixture(l dcLiveLedger) ncLiveLedger {
	out := ncLiveLedger{Environment: l.Environment, Binding: l.Binding, Records: []ncLiveRecord{}}
	for _, r := range l.Records {
		out.Records = append(out.Records, ncLiveRecord{Raw: r.Raw, Service: r.Service, Population: r.Population, ID: r.ID, Signal: r.Signal, TimeMS: r.TimeMS})
	}
	return out
}
func dcLiveOldNumeric(saved exploration.CapturePage, l dcLiveLedger) bool {
	if saved.Receipt.Question.Comparison == nil || saved.Receipt.Question.Comparison.Distribution || saved.Comparison == nil || saved.Comparison.Request.Distribution || saved.Comparison.Request.Baseline.NumericHistogram != nil || saved.Comparison.Request.Selection.NumericHistogram != nil || saved.Comparison.Result == nil || saved.Comparison.Result.Measures == nil || saved.Comparison.Result.Measures.Numeric == nil || saved.Comparison.Result.Measures.Numeric.Distribution != nil {
		return false
	}
	return ncLivePairCheck(saved.Comparison.Request, *saved.Comparison.Result, dcLiveNCFixture(l))
}
func dcLiveReceiptHash(r workbench.ComparisonReceipt) bool {
	id := r.ID
	r.ID = ""
	raw, e := json.Marshal(r)
	if e != nil {
		return false
	}
	hash := sha256.Sum256(raw)
	return id == hex.EncodeToString(hash[:])
}
func dcLiveStoredEvidence(l dcLiveLedger) bool {
	saved := l.Saved
	if saved.Comparison == nil || saved.Comparison.Result == nil || saved.Receipt.Question.Comparison == nil || !saved.Receipt.Question.Comparison.Distribution || !saved.Comparison.Request.Distribution || !dcLivePairCheck(saved.Comparison.Request, *saved.Comparison.Result, l) || !nhLiveSameQuery(saved.Receipt.Question.Comparison.Baseline, saved.Comparison.Request.Baseline.Question) || !nhLiveSameQuery(saved.Receipt.Question.Comparison.Selection, saved.Comparison.Request.Selection.Question) || !nhLiveSameQuery(saved.Comparison.Request.Baseline.Question, l.Baseline.Question) || !nhLiveSameQuery(saved.Comparison.Request.Selection.Question, l.Selection.Question) {
		return false
	}
	a, z := saved.Comparison.Request.Baseline, saved.Comparison.Request.Selection
	decision := l.Decision
	want := ncLiveMeasure(*a.Question.Numeric, a.Numeric, z.Numeric, false)
	dist := dcLiveDistribution(a, z)
	want.Distribution = &dist
	if !decision.Distribution || decision.State != "comparable" || decision.Measures == nil || decision.Measures.Numeric == nil || !reflect.DeepEqual(*decision.Measures.Numeric, want) || decision.CountDifference != ncLiveDifference(a.Count, z.Count) || decision.Baseline.Receipt.ID != saved.Comparison.BaselineID || decision.Selection.Receipt.ID != saved.Comparison.SelectionID || !nhLiveSameQuery(*decision.Baseline.Receipt.Question.Explore, a.Question) || !nhLiveSameQuery(*decision.Selection.Receipt.Question.Explore, z.Question) || decision.Baseline.Count != a.Count || decision.Selection.Count != z.Count {
		return false
	}
	r := l.Receipt
	if r.Version != 1 || r.Algorithm != "goo/evidence-comparison/v5" || r.RecordedAt == "" || !dcLiveReceiptHash(r) || r.Result == nil || !reflect.DeepEqual(*r.Result, decision) || !reflect.DeepEqual(r, l.ReopenedReceipt) {
		return false
	}
	old := l.OldDecision
	want.Distribution = nil
	if old.Distribution || old.State != "comparable" || old.Measures == nil || old.Measures.Numeric == nil || !reflect.DeepEqual(*old.Measures.Numeric, want) || old.Baseline.Receipt.ID != decision.Baseline.Receipt.ID || old.Selection.Receipt.ID != decision.Selection.Receipt.ID {
		return false
	}
	previous := l.OldReceipt
	return previous.Version == 1 && previous.Algorithm == "goo/evidence-comparison/v4" && previous.Result != nil && reflect.DeepEqual(*previous.Result, old) && dcLiveReceiptHash(previous) && reflect.DeepEqual(previous, l.OldReceiptReopened) && dcLiveOldNumeric(l.OldNumeric, l) && reflect.DeepEqual(l.OldNumeric, l.OldNumericReopened)
}
func dcLiveEvents(l dcLiveLedger) bool {
	expected := []struct {
		action                     string
		reads, protocols, planners int
	}{{"stage", 0, 0, 0}, {"baseline", 1, 0, 0}, {"pin", 1, 1, 0}, {"edit", 1, 1, 0}, {"selection", 2, 2, 0}, {"distribution", 2, 3, 0}, {"toggle", 2, 3, 0}, {"capture", 3, 3, 0}, {"reopen", 3, 3, 0}, {"view-stage", 3, 3, 0}, {"view-run", 5, 4, 0}, {"view-save", 5, 4, 0}, {"view-reopen", 5, 4, 0}, {"view-reuse", 7, 5, 0}, {"baseline-brush", 7, 5, 0}, {"baseline-stage", 7, 5, 1}, {"baseline-run", 8, 5, 1}, {"back", 8, 5, 1}, {"selection-brush", 8, 5, 1}, {"selection-stage", 8, 5, 2}, {"selection-run", 9, 5, 2}}
	if len(l.Events) != len(expected) {
		return false
	}
	for i, want := range expected {
		got := l.Events[i]
		if got.Action != want.action || got.Reads != want.reads || got.Protocols != want.protocols || got.Planners != want.planners {
			return false
		}
	}
	return true
}
func TestQuintDistributionComparisonLiveLedger(t *testing.T) {
	file := os.Getenv("GOO_DISTRIBUTION_COMPARISON_LEDGER")
	if file == "" {
		t.Skip("actual journey176 distribution comparison ledger opt-in")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l dcLiveLedger
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		before, _ := json.Marshal(l)
		switch mutation {
		case 0:
			l.Compare.Response.Measures.Numeric.Distribution.Denominator = "valid-count"
		case 1:
			l.Compare.Response.Measures.Numeric.Distribution.Bins[0].From = "rounded"
		case 2:
			l.Compare.Response.Measures.Numeric.Distribution.Bins[0].CountDifference = "1"
		case 3:
			l.Compare.Response.Measures.Numeric.Distribution.Bins[0].BaselinePercent = "1"
		case 4:
			l.Compare.Request.Selection.Question.Numeric.Unit = "seconds"
		case 5:
			l.Plans[0].Response.Parent = l.Plans[0].Request.Comparison.Selection.Question
		case 6:
			l.Plans[0].Response.Selected.FromMS++
		case 7:
			l.Plans[1].Response.Selected.Where = l.Plans[1].Response.Rest.Where
		case 8:
			l.Events[6].Reads++
		case 9:
			l.Events[14].Planners++
		case 10:
			l.Events[17].Reads++
		case 11:
			l.Saved.Comparison.Result.Measures.Numeric.Distribution.Bins[0].BaselineCount = "999"
		case 12:
			l.Reopened.Receipt.Question.Comparison.Distribution = false
		case 13:
			l.Reused.Request.Distribution = false
		case 14:
			l.View.Breakdowns["spans"].Baseline.ToMS++
		case 15:
			l.DomModel.Share.Sides[0].Ceiling = "1"
		case 16:
			l.DomModel.Count.Model.Points[0].Key = "0"
		case 17:
			for i := range l.EdgeCases {
				if l.EdgeCases[i].Kind == "unknown" {
					l.EdgeCases[i].Response.Measures.Numeric.Distribution.Bins[0].DifferencePP = "0"
				}
			}
		case 18:
			for i := range l.EdgeCases {
				if l.EdgeCases[i].Kind == "empty" {
					l.EdgeCases[i].Response.Measures.Numeric.Distribution.Bins[0].SelectionPercent = "0"
				}
			}
		case 19:
			for i := range l.EdgeCases {
				if l.EdgeCases[i].Kind == "all-invalid" {
					l.EdgeCases[i].Response.Measures.Numeric.Distribution.Bins[0].SelectionPercent = ""
				}
			}
		case 20:
			for i := range l.Records {
				if l.Records[i].ResourceRaw != nil && *l.Records[i].ResourceRaw == "9007199254740993" {
					v := "9007199254740992"
					l.Records[i].ResourceRaw = &v
					break
				}
			}
		case 21:
			l.Receipt.Algorithm = "goo/evidence-comparison/v4"
		case 22:
			l.OldReceipt.Algorithm = "goo/evidence-comparison/v5"
		case 23:
			l.Aggregate.Response.AggregateOnly = false
		case 24:
			l.Aggregate.Plans[0].Response.Rest.Where = l.Aggregate.Plans[0].Response.Selected.Where
		}
		after, _ := json.Marshal(l)
		if mutation >= 0 && string(before) == string(after) {
			t.Fatalf("negative control %d made no change", mutation)
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		record := func(action string, ok bool) {
			if !ok {
				t.Logf("distribution ledger rejected phase %s mutation %d", action, mutation)
			}
			rec.Record(context.Background(), action, nil, "obs.coherent", ok)
		}
		record("plan", l.Version == 1 && len(l.Records) > 100 && l.Binding == (exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work.items"}, Type: "int64", Unit: "items", Parse: "canonical-decimal"}))
		main := dcLiveResult(l.Baseline, l) && dcLiveResult(l.Selection, l) && reflect.DeepEqual(l.Compare.Request.Baseline, dcLiveCohortOf(l.Baseline)) && reflect.DeepEqual(l.Compare.Request.Selection, dcLiveCohortOf(l.Selection)) && dcLivePairCheck(l.Compare.Request, l.Compare.Response, l)
		extras := dcLiveActualPair(l.Logs, l) && l.Logs.Request.Baseline.Question.Signal == "logs" && len(l.Extremes) == 2
		types := map[string]bool{}
		for _, pair := range l.Extremes {
			extras = extras && dcLiveActualPair(pair, l) && pair.Request.Baseline.Question.Numeric.Field == (exploration.Field{Kind: "resource", Key: "work.extreme"})
			types[pair.Request.Baseline.Question.Numeric.Type] = true
		}
		extras = extras && types["int64"] && types["uint64"]
		kinds := map[string]bool{}
		edgeOK := len(l.EdgeCases) == 3
		for _, pair := range l.EdgeCases {
			edgeOK = edgeOK && dcLiveActualPair(pair, l)
			kinds[pair.Kind] = true
		}
		edgeOK = edgeOK && kinds["all-invalid"] && kinds["empty"] && kinds["unknown"]
		aggregate := dcLiveActualPair(l.Aggregate.dcLivePair, l) && l.Aggregate.Response.AggregateOnly && len(l.Aggregate.Plans) == 2
		for _, plan := range l.Aggregate.Plans {
			aggregate = aggregate && dcLivePlanCheck(plan, l) && reflect.DeepEqual(plan.Request.Comparison, l.Aggregate.Request)
		}
		plans := len(l.Plans) == 2
		sides := map[string]bool{}
		for _, plan := range l.Plans {
			plans = plans && dcLivePlanCheck(plan, l) && reflect.DeepEqual(plan.Request.Comparison, l.Reused.Request)
			sides[plan.Side] = true
		}
		plans = plans && sides["baseline"] && sides["selection"]
		executed := true
		for label, ok := range map[string]bool{"full population": main, "logs/resource integer domains": extras, "empty/invalid/unknown": edgeOK, "aggregate-only": aggregate, "original-side continuations": plans} {
			if !ok {
				executed = false
				t.Logf("distribution oracle rejected %s mutation %d", label, mutation)
			}
		}
		record("execute", executed)
		refused := map[string]bool{}
		for _, r := range l.Refusals {
			var body struct{ Error string }
			if r.Status == 400 && json.Unmarshal(r.Body, &body) == nil && body.Error != "" {
				refused[r.Kind] = true
			}
		}
		refusals := true
		for _, k := range []string{"missing-histogram", "bin-count", "binding", "missing-optin", "numeric-false", "legacy-zero-groups", "empty-range", "unknown-range"} {
			refusals = refusals && refused[k]
		}
		record("inspect", dcLiveEvents(l) && refusals && dcLiveDOMCheck(l.DomModel.Count, l.Compare.Request, "count", true) && dcLiveDOMCheck(l.DomModel.Share, l.Compare.Request, "share", true))
		record("capture", dcLiveStoredEvidence(l))
		seed := l.SeedView
		base := seed.Breakdowns["spans"].Baseline
		validScope := len(seed.Queries) == 1 && base != nil
		if validScope {
			a, z := *base, seed.Queries[0]
			validScope = a.ToMS == z.FromMS && a.FromMS == l.Baseline.Question.FromMS && z.ToMS == l.Selection.Question.ToMS
			a.FromMS = l.Baseline.Question.FromMS
			a.ToMS = l.Baseline.Question.ToMS
			z.FromMS = l.Selection.Question.FromMS
			z.ToMS = l.Selection.Question.ToMS
			validScope = validScope && nhLiveSameQuery(a, l.Baseline.Question) && nhLiveSameQuery(z, l.Selection.Question) && len(l.View.Queries) == 1 && nhLiveSameQuery(seed.Queries[0], l.View.Queries[0]) && l.View.Breakdowns["spans"].Baseline != nil && nhLiveSameQuery(*base, *l.View.Breakdowns["spans"].Baseline)
		}
		record("reuse", validScope && dcLiveViewRunCheck(l.InitialView, seed, l) && dcLiveViewRunCheck(l.Reused, l.View, l))
		record("reopen", reflect.DeepEqual(l.Saved, l.Reopened) && dcLiveDOMCheck(l.DomModel.Saved, l.Saved.Comparison.Request, "share", true))
		checker, e := validate.NewChecker("model/grouped_series.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		r, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return r.OK()
	}
	if !check(-1) {
		t.Fatal("actual distribution ledger rejected")
	}
	for i := 0; i < 25; i++ {
		if check(i) {
			t.Fatalf("distribution corruption accepted: %d", i)
		}
	}
}

type dcLiveLate struct {
	Version     int
	Environment string
	Binding     exploration.NumericBinding
	Records     []dcLiveRecord
	View        dcLiveView
	Applied     dcLiveViewRun
	Events      []struct {
		Action                     string
		Reads, Protocols, Planners int
	}
	OriginalScope               struct{ Baseline, Selection exploration.Query }
	Late, Resumed               dcLivePlan
	Before, AfterLate, AfterRun int
}

func TestQuintDistributionComparisonLateLedger(t *testing.T) {
	file := os.Getenv("GOO_DISTRIBUTION_COMPARISON_LATE_LEDGER")
	if file == "" {
		t.Skip("actual journey176 held real planner ledger opt-in")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var observed dcLiveLate
		if e := json.Unmarshal(raw, &observed); e != nil {
			t.Fatal(e)
		}
		before, _ := json.Marshal(observed)
		switch mutation {
		case 0:
			observed.AfterLate++
		case 1:
			observed.Events[4].Reads++
		case 2:
			observed.Late.Response.Parent = observed.OriginalScope.Selection
		case 3:
			observed.Resumed.Response.Selected.FromMS++
		case 4:
			observed.Resumed.Response.Selected.Where = observed.Late.Response.Selected.Where
		case 5:
			observed.Resumed.Result.NumericHistogram.Bins[0].Count = "999"
		case 6:
			observed.Events[2].Planners++
		case 7:
			observed.Events[7].Reads++
		case 8:
			observed.OriginalScope.Baseline.ToMS++
		case 9:
			observed.Resumed.Request.Comparison.Distribution = false
		case 10:
			observed.Events[6].Reads++
		}
		after, _ := json.Marshal(observed)
		if mutation >= 0 && string(before) == string(after) {
			t.Fatalf("late control %d did not change ledger", mutation)
		}
		fixture := dcLiveLedger{Version: observed.Version, Environment: observed.Environment, Binding: observed.Binding, Records: observed.Records}
		held, resumed := observed.Late, observed.Resumed
		for _, p := range []*dcLivePlan{&held, &resumed} {
			p.Side = p.Request.Side
			p.FromBin = p.Request.FromBin
			p.ToBin = p.Request.ToBin
		}
		stage := observed.Version == 1 && len(observed.Records) > 100 && len(observed.View.Queries) == 1 && observed.View.Breakdowns["spans"].Baseline != nil && nhLiveSameQuery(*observed.View.Breakdowns["spans"].Baseline, observed.OriginalScope.Baseline) && nhLiveSameQuery(observed.View.Queries[0], observed.OriginalScope.Selection) && observed.OriginalScope.Baseline.ToMS == observed.OriginalScope.Selection.FromMS
		applied := dcLiveViewRunCheck(observed.Applied, observed.View, fixture)
		validHeld := held.Side == "baseline" && reflect.DeepEqual(held.Request.Comparison, observed.Applied.Request) && dcLivePlanScope(held, fixture) && nhLiveSameQuery(held.Response.Parent, observed.OriginalScope.Baseline)
		returned := observed.Before == 2 && observed.AfterLate == observed.Before
		freshPlan := resumed.Side == "selection" && reflect.DeepEqual(resumed.Request.Comparison, observed.Applied.Request) && dcLivePlanScope(resumed, fixture) && nhLiveSameQuery(resumed.Response.Parent, observed.OriginalScope.Selection)
		validResume := freshPlan && resumed.Side == "selection" && reflect.DeepEqual(resumed.Request.Comparison, observed.Applied.Request) && dcLivePlanCheck(resumed, fixture) && nhLiveSameQuery(resumed.Response.Parent, observed.OriginalScope.Selection) && observed.AfterRun == observed.Before+1
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		checks := map[string]bool{"stage": stage, "apply": applied, "brush": applied, "held": validHeld, "return": returned, "fresh-brush": returned, "fresh-stage": freshPlan && returned, "fresh-run": validResume}
		if len(observed.Events) != 8 {
			return false
		}
		for _, event := range observed.Events {
			ok := checks[event.Action]
			if !ok {
				t.Logf("late distribution oracle rejected %s mutation %d", event.Action, mutation)
			}
			rec.Record(context.Background(), strings.ReplaceAll(event.Action, "-", "_"), nil, "obs.coherent", ok, "obs.reads", event.Reads, "obs.protocols", event.Protocols, "obs.planners", event.Planners)
		}
		checker, e := validate.NewChecker("model/distribution_lifecycle.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("actual late distribution trace rejected")
	}
	for i := 0; i < 11; i++ {
		if check(i) {
			t.Fatalf("late distribution control accepted: %d", i)
		}
	}
}
