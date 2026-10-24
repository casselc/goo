package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type ncLiveRecord struct {
	Raw                             *string
	Service, Population, ID, Signal string
	TimeMS                          int64 `json:"timeMs"`
}
type ncLivePair struct {
	Baseline, Selection exploration.Result
	Request             exploration.ComparisonRequest
	Response            exploration.CohortComparison
	Kind                string
}
type ncLivePartition struct {
	Plan           exploration.NumericHistogramSelection
	Selected, Rest exploration.Result
	Request        exploration.ComparisonRequest
	Response       exploration.CohortComparison
}
type ncLiveLedger struct {
	Version                                   int
	Environment                               string
	Binding                                   exploration.NumericBinding
	Records                                   []ncLiveRecord
	Baseline, Selection, ParentResult         exploration.Result
	Compare, Previous, Logs, Unsigned, Legacy ncLivePair
	Partition, Reused                         ncLivePartition
	EdgeCases                                 []ncLivePair
	Saved, Reopened                           exploration.CapturePage
	View                                      struct {
		Version    int
		ID         string
		Queries    []exploration.Query
		Breakdowns map[string]struct{ Baseline *exploration.Query }
	}
	Refusals []struct {
		Kind   string
		Status int
		Body   json.RawMessage
	}
	Events []struct {
		Action           string
		Reads, Protocols int
	}
	BeforeReopen, AfterReopen int
}

// This oracle shares only the independent stdlib integer/roundtrip and power-of-2
// construction from the histogram ledger checker; it never calls goo evaluation,
// comparison, parser, predicate, SQL compiler or merge functions for expectations.
func ncLiveRaw(field exploration.Field, row ncLiveRecord) (*string, error) {
	switch {
	case field.Kind == "column" && field.Key == "service":
		return &row.Service, nil
	case field.Kind == "attribute" && field.Key == "work.population":
		return &row.Population, nil
	case field.Kind == "attribute" && field.Key == "work.items":
		return row.Raw, nil
	}
	return nil, fmt.Errorf("fixture does not define field %+v", field)
}
func ncLiveMatch(p *exploration.Predicate, row ncLiveRecord) (bool, error) {
	if p == nil {
		return true, nil
	}
	if len(p.All) > 0 {
		for i := range p.All {
			ok, e := ncLiveMatch(&p.All[i], row)
			if e != nil || !ok {
				return false, e
			}
		}
		return true, nil
	}
	if len(p.Any) > 0 {
		for i := range p.Any {
			ok, e := ncLiveMatch(&p.Any[i], row)
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
	raw, e := ncLiveRaw(*f, row)
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
	return false, fmt.Errorf("unsupported predicate %s", p.Op)
}
func ncLiveRows(q exploration.Query, l ncLiveLedger) ([]ncLiveRecord, error) {
	if q.Environment != l.Environment || q.Signal != "spans" && q.Signal != "logs" || q.ToMS <= q.FromMS {
		return nil, fmt.Errorf("wrong fixture scope")
	}
	out := []ncLiveRecord{}
	for _, r := range l.Records {
		if r.Signal != q.Signal || r.TimeMS < q.FromMS || r.TimeMS >= q.ToMS {
			continue
		}
		ok, e := ncLiveMatch(q.Where, r)
		if e != nil {
			return nil, e
		}
		if ok {
			out = append(out, r)
		}
	}
	return out, nil
}
func ncLiveSummary(b exploration.NumericBinding, rows []ncLiveRecord) (exploration.NumericSummary, []exploration.NumericHistogramBin) {
	fixture := make([]numericHistogramLiveRecord, len(rows))
	for i, r := range rows {
		fixture[i] = numericHistogramLiveRecord{Raw: r.Raw, TimeMS: r.TimeMS, ID: r.ID}
	}
	return nhLiveSummary(b, fixture)
}
func ncLiveBinding(b *exploration.NumericBinding, l ncLiveLedger) bool {
	if b == nil {
		return false
	}
	want := l.Binding
	if b.Type == "uint64" {
		want.Type = "uint64"
	}
	return *b == want
}
func ncLiveCohort(c exploration.Cohort, l ncLiveLedger, numeric bool) bool {
	rows, e := ncLiveRows(c.Question, l)
	if e != nil || c.Count != strconv.Itoa(len(rows)) || c.Basis == "" || len(c.Question.GroupBy) != 1 || c.Question.GroupBy[0] != (exploration.Field{Kind: "column", Key: "service"}) {
		return false
	}
	if c.State != "changed_between_reads" && (len(rows) > 0 && c.State != "observed" || len(rows) == 0 && c.State != "no_data") {
		return false
	}
	if numeric {
		if !ncLiveBinding(c.Question.Numeric, l) {
			return false
		}
		summary, _ := ncLiveSummary(*c.Question.Numeric, rows)
		if c.Numeric == nil || *c.Numeric != summary {
			return false
		}
	} else if c.Numeric != nil {
		return false
	}
	groups := map[string][]ncLiveRecord{}
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
		if numeric {
			summary, _ := ncLiveSummary(*c.Question.Numeric, members)
			if g.Numeric == nil || *g.Numeric != summary {
				return false
			}
		} else if g.Numeric != nil {
			return false
		}
	}
	return true
}
func ncLiveCohortOf(r exploration.Result) exploration.Cohort {
	c := exploration.Cohort{Question: r.Question, Basis: r.Basis, Count: r.Count, State: r.State, Numeric: r.Numeric, Duration: r.Duration, SpanStatus: r.SpanStatus, Groups: []exploration.ValueCount{}}
	for _, raw := range r.Groups {
		var g exploration.ValueCount
		if json.Unmarshal(raw, &g) != nil {
			return exploration.Cohort{}
		}
		c.Groups = append(c.Groups, g)
	}
	return c
}
func ncLiveResult(r exploration.Result, l ncLiveLedger) bool {
	if r.CapturedAt == "" || !ncLiveCohort(ncLiveCohortOf(r), l, true) {
		return false
	}
	rows, e := ncLiveRows(r.Question, l)
	if e != nil {
		return false
	}
	_, bins := ncLiveSummary(*r.Question.Numeric, rows)
	if r.Question.NumericHistogram != nil {
		if r.NumericHistogram == nil || r.NumericHistogram.Binning != "powers-of-two" || !reflect.DeepEqual(r.NumericHistogram.Bins, bins) {
			return false
		}
	} else if r.NumericHistogram != nil {
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
	fixture := map[int64]ncLiveRecord{}
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
		if !ok || seen[source.TimeMS] || ns != source.TimeMS*1000000 || row.Service != source.Service || r.Question.Signal == "spans" && row.ID != source.ID {
			return false
		}
		seen[source.TimeMS] = true
		value, present := row.Attributes["work.items"]
		if source.Raw == nil && present || source.Raw != nil && (!present || value != *source.Raw) || row.Attributes["work.population"] != source.Population || row.Resource["deployment.environment.name"] != l.Environment {
			return false
		}
	}
	return r.Plan.Count != "" && !strings.Contains(strings.ToUpper(r.Plan.Count), " LIMIT ") && r.Plan.Parameters["environment"] == l.Environment && r.Plan.Parameters["from"] == strconv.FormatInt(r.Question.FromMS, 10) && r.Plan.Parameters["to"] == strconv.FormatInt(r.Question.ToMS, 10)
}
func ncLiveInt(s string) *big.Int { v, _ := new(big.Int).SetString(s, 10); return v }
func ncLiveDifference(a, z string) string {
	return new(big.Int).Sub(ncLiveInt(z), ncLiveInt(a)).String()
}
func ncLiveMeasure(b exploration.NumericBinding, a, z *exploration.NumericSummary, unknown bool) exploration.NumericMeasureComparison {
	r := exploration.NumericMeasureComparison{Binding: b, State: "comparable", Baseline: a, Selection: z, Coverage: []exploration.NumericCoverageChange{}}
	counts := func(s *exploration.NumericSummary) []string {
		if s == nil {
			return []string{"0", "0", "0", "0"}
		}
		return []string{s.Valid, s.Missing, s.Empty, s.Invalid}
	}
	aa, zz := counts(a), counts(z)
	for i, k := range []string{"valid", "missing", "empty", "invalid"} {
		c := exploration.NumericCoverageChange{Kind: k, BaselineCount: aa[i], SelectionCount: zz[i]}
		if !unknown {
			c.Difference = ncLiveDifference(aa[i], zz[i])
			if a != nil && z != nil && a.Count != "0" && z.Count != "0" {
				x := new(big.Rat).SetFrac(ncLiveInt(aa[i]), ncLiveInt(a.Count))
				y := new(big.Rat).SetFrac(ncLiveInt(zz[i]), ncLiveInt(z.Count))
				c.DifferencePP = new(big.Rat).Mul(new(big.Rat).Sub(y, x), big.NewRat(100, 1)).RatString()
			}
		}
		r.Coverage = append(r.Coverage, c)
	}
	switch {
	case unknown:
		r.State = "unknown"
		r.Reason = "source-reads-disagree"
	case a == nil || z == nil:
		r.State = "unavailable"
		r.Reason = "absent-group"
	case a.Count == "0" || z.Count == "0":
		r.State = "unavailable"
		r.Reason = "empty-population"
	case a.Valid == "0" || z.Valid == "0":
		r.State = "unavailable"
		r.Reason = "no-valid-values"
	default:
		r.SumDifference = ncLiveDifference(a.Sum, z.Sum)
		r.MinDifference = ncLiveDifference(a.Min, z.Min)
		r.MaxDifference = ncLiveDifference(a.Max, z.Max)
		am := new(big.Rat).SetFrac(ncLiveInt(a.Sum), ncLiveInt(a.Valid))
		zm := new(big.Rat).SetFrac(ncLiveInt(z.Sum), ncLiveInt(z.Valid))
		r.MeanDifference = new(big.Rat).Sub(zm, am).RatString()
	}
	return r
}
func ncLivePairCheck(request exploration.ComparisonRequest, response exploration.CohortComparison, l ncLiveLedger) bool {
	a, z := request.Baseline, request.Selection
	if !ncLiveCohort(a, l, request.Numeric) || !ncLiveCohort(z, l, request.Numeric) || a.Question.Signal != z.Question.Signal || a.Question.Environment != z.Question.Environment || !reflect.DeepEqual(a.Question.GroupBy, z.Question.GroupBy) || request.Numeric && (*a.Question.Numeric != *z.Question.Numeric) {
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
	if response.State != state || response.Reason != reason || response.Field != a.Question.GroupBy[0] || response.BaselineCount != a.Count || response.SelectionCount != z.Count || response.BaselineBasis != a.Basis || response.SelectionBasis != z.Basis || len(response.Basis) != 64 {
		return false
	}
	if request.Numeric {
		want := ncLiveMeasure(*a.Question.Numeric, a.Numeric, z.Numeric, unknown)
		if response.Measures == nil || response.Measures.Numeric == nil || !reflect.DeepEqual(*response.Measures.Numeric, want) {
			return false
		}
	} else if response.Measures != nil && response.Measures.Numeric != nil {
		return false
	}
	if state != "ok" {
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
		d := new(big.Rat).Sub(zp, ap).FloatString(2)
		if d == "-0.00" {
			d = "0.00"
		}
		if row.BaselineCount != ac || row.SelectionCount != zc || row.BaselinePercent != ap.FloatString(2) || row.SelectionPercent != zp.FloatString(2) || row.DifferencePP != d {
			return false
		}
		if request.Numeric {
			want := ncLiveMeasure(*a.Question.Numeric, x.Numeric, y.Numeric, false)
			if row.Measures == nil || row.Measures.Numeric == nil || !reflect.DeepEqual(*row.Measures.Numeric, want) {
				return false
			}
		} else if row.Measures != nil && row.Measures.Numeric != nil {
			return false
		}
	}
	return true
}
func ncLiveActualPair(pair ncLivePair, l ncLiveLedger) bool {
	return ncLiveResult(pair.Baseline, l) && ncLiveResult(pair.Selection, l) && reflect.DeepEqual(pair.Request.Baseline, ncLiveCohortOf(pair.Baseline)) && reflect.DeepEqual(pair.Request.Selection, ncLiveCohortOf(pair.Selection)) && pair.Request.Numeric && ncLivePairCheck(pair.Request, pair.Response, l)
}
func ncLivePartitionCheck(l ncLiveLedger) bool {
	p := l.Partition
	s := p.Plan
	parent := l.ParentResult.Question
	edges := nhLiveEdges(l.Binding.Type)
	if !ncLiveResult(l.ParentResult, l) || !nhLiveSameQuery(parent, s.Parent) || s.NumericComparison == nil || s.FromBin < 0 || s.ToBin <= s.FromBin || s.ToBin >= len(edges) || s.From != edges[s.FromBin] || s.To != edges[s.ToBin] || s.Binning != "powers-of-two" {
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
		if !nhLiveSameQuery(want, q) || !ncLiveBinding(q.Numeric, l) {
			return false
		}
	}
	lo, hi := ncLiveInt(s.From), ncLiveInt(s.To)
	rows := append([]ncLiveRecord{}, l.Records...)
	for _, raw := range []string{"", "01", "-1", "9007199254740993"} {
		v := raw
		rows = append(rows, ncLiveRecord{Raw: &v})
	}
	rows = append(rows, ncLiveRecord{})
	for _, raw := range edges {
		v := raw
		rows = append(rows, ncLiveRecord{Raw: &v})
	}
	for _, r := range rows {
		original, e := ncLiveMatch(parent.Where, r)
		if e != nil {
			return false
		}
		selected, e := ncLiveMatch(s.NumericComparison.Selection.Where, r)
		if e != nil {
			return false
		}
		rest, e := ncLiveMatch(s.NumericComparison.Baseline.Where, r)
		if e != nil {
			return false
		}
		value, state := nhLiveInteger(r.Raw, l.Binding.Type)
		want := original && state == "valid" && value.Cmp(lo) >= 0 && value.Cmp(hi) < 0
		if selected != want || rest != (original && !want) || selected && rest {
			return false
		}
	}
	if !nhLiveSameQuery(p.Selected.Question, s.NumericComparison.Selection) || !nhLiveSameQuery(p.Rest.Question, s.NumericComparison.Baseline) || !ncLiveResult(p.Selected, l) || !ncLiveResult(p.Rest, l) || !reflect.DeepEqual(p.Request.Baseline, ncLiveCohortOf(p.Rest)) || !reflect.DeepEqual(p.Request.Selection, ncLiveCohortOf(p.Selected)) || !p.Request.Numeric || !ncLivePairCheck(p.Request, p.Response, l) {
		return false
	}
	return true
}
func ncLiveEventCheck(l ncLiveLedger) bool {
	expected := []struct {
		action           string
		reads, protocols int
	}{{"stage", 0, 0}, {"baseline", 1, 0}, {"pin", 1, 1}, {"edit", 1, 1}, {"compare", 2, 2}, {"previous", 3, 3}, {"capture", 4, 3}, {"reopen", 4, 3}, {"brush", 5, 3}, {"partition-stage", 5, 4}, {"partition-run", 7, 5}, {"view-reopen", 7, 5}, {"view-run", 9, 6}}
	if len(expected) != len(l.Events) {
		return false
	}
	for i, want := range expected {
		got := l.Events[i]
		if got.Action != want.action || got.Reads != want.reads || got.Protocols != want.protocols {
			return false
		}
	}
	return l.BeforeReopen == l.AfterReopen
}

func TestQuintNumericComparisonLiveLedger(t *testing.T) {
	file := os.Getenv("GOO_NUMERIC_COMPARISON_LEDGER")
	if file == "" {
		t.Skip("actual journey175 numeric comparison ledger opt-in")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l ncLiveLedger
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		beforeMutation, _ := json.Marshal(l)
		switch mutation {
		case 0:
			l.Compare.Response.Measures.Numeric.MeanDifference = "-1/5"
		case 1:
			l.Compare.Response.Measures.Numeric.SumDifference = "0"
		case 2:
			l.Compare.Response.Measures.Numeric.MeanDifference = "1/4"
		case 3:
			l.Partition.Plan.NumericComparison.Baseline.Where = l.Partition.Plan.NumericComparison.Selection.Where
		case 4:
			l.Compare.Request.Selection.Question.Environment += "-foreign"
		case 5:
			l.Compare.Request.Selection.Question.Numeric.Unit = "seconds"
		case 6:
			l.Events[3].Reads++
		case 7:
			l.Events[9].Reads++
		case 8:
			l.Saved.Comparison.Result.Measures.Numeric.MeanDifference = "0"
		case 9:
			l.Reopened.Comparison.Request.Selection.Numeric.Sum = "0"
		case 10:
			l.Reused.Selected.Question.FromMS++
		case 11:
			for i := range l.Records {
				if l.Records[i].Raw != nil && *l.Records[i].Raw == "9007199254740993" {
					v := "9007199254740992"
					l.Records[i].Raw = &v
					break
				}
			}
		case 12:
			value := &l.Compare.Response.Measures.Numeric.Coverage[0].DifferencePP
			if *value == "0" {
				*value = "1"
			} else {
				*value = "0"
			}
		case 13:
			for i := range l.Compare.Response.Rows {
				n := l.Compare.Response.Rows[i].Measures.Numeric
				if n.State == "unavailable" {
					n.State = "comparable"
					break
				}
			}
		case 14:
			l.Previous.Baseline.Question.ToMS++
		case 15:
			l.View.Breakdowns["spans"].Baseline.Numeric.Parse = "or-zero"
		case 16:
			l.Events[7].Protocols++
		case 17:
			l.Partition.Response.Measures.Numeric.MeanDifference = "0"
		case 18:
			for i := range l.EdgeCases {
				if l.EdgeCases[i].Kind == "unknown" {
					l.EdgeCases[i].Response.Measures.Numeric.SumDifference = "0"
				}
			}
		case 19:
			for i := range l.EdgeCases {
				if l.EdgeCases[i].Kind == "unknown" {
					l.EdgeCases[i].Response.Measures.Numeric.Coverage[0].Difference = "0"
				}
			}
		case 20:
			l.Legacy.Response.Measures = &exploration.MeasureComparison{Numeric: l.Compare.Response.Measures.Numeric}
		case 21:
			l.Compare.Request.Numeric = false
		}
		afterMutation, _ := json.Marshal(l)
		if mutation >= 0 && string(beforeMutation) == string(afterMutation) {
			t.Fatalf("negative control %d did not change the actual ledger", mutation)
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		record := func(action string, ok bool) {
			if !ok {
				t.Logf("numeric comparison ledger rejected phase %s mutation %d", action, mutation)
			}
			rec.Record(context.Background(), action, nil, "obs.coherent", ok)
		}
		record("plan", l.Version == 1 && len(l.Records) > 100 && l.Binding == (exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work.items"}, Type: "int64", Unit: "items", Parse: "canonical-decimal"}) && ncLiveBinding(l.Baseline.Question.Numeric, l) && ncLiveBinding(l.Selection.Question.Numeric, l))
		base, selection := l.Baseline, l.Selection
		direct := base.Question.FromMS == selection.Question.FromMS && base.Question.ToMS == selection.Question.ToMS && ncLiveResult(base, l) && ncLiveResult(selection, l) && nhLiveSameQuery(base.Question, l.Compare.Request.Baseline.Question) && reflect.DeepEqual(l.Compare.Request.Baseline, ncLiveCohortOf(base)) && reflect.DeepEqual(l.Compare.Request.Selection, ncLiveCohortOf(selection)) && l.Compare.Request.Numeric && ncLivePairCheck(l.Compare.Request, l.Compare.Response, l)
		prior := l.Previous
		aq, zq := prior.Baseline.Question, prior.Selection.Question
		normalized := aq
		normalized.FromMS = zq.FromMS
		normalized.ToMS = zq.ToMS
		previous := ncLiveActualPair(prior, l) && aq.ToMS == zq.FromMS && aq.ToMS-aq.FromMS == zq.ToMS-zq.FromMS && nhLiveSameQuery(normalized, zq) && nhLiveSameQuery(zq, selection.Question)
		cases := ncLiveActualPair(l.Logs, l) && l.Logs.Baseline.Question.Signal == "logs" && ncLiveActualPair(l.Unsigned, l) && l.Unsigned.Baseline.Question.Numeric.Type == "uint64"
		edgeKinds := map[string]bool{}
		edgesOK := len(l.EdgeCases) == 3
		for _, pair := range l.EdgeCases {
			expectedBase := ncLiveCohortOf(pair.Baseline)
			if pair.Kind == "unknown" {
				expectedBase.State = "changed_between_reads"
			}
			edgesOK = edgesOK && ncLiveResult(pair.Baseline, l) && ncLiveResult(pair.Selection, l) && reflect.DeepEqual(pair.Request.Baseline, expectedBase) && reflect.DeepEqual(pair.Request.Selection, ncLiveCohortOf(pair.Selection)) && pair.Request.Numeric && ncLivePairCheck(pair.Request, pair.Response, l)
			edgeKinds[pair.Kind] = true
		}
		edgesOK = edgesOK && edgeKinds["unknown"] && edgeKinds["invalid-only"] && edgeKinds["empty"]
		legacy := !l.Legacy.Request.Numeric && ncLivePairCheck(l.Legacy.Request, l.Legacy.Response, l)
		executed := true
		for label, ok := range map[string]bool{"current populations": direct, "previous-window populations": previous, "logs and unsigned": cases, "selected/rest partition": ncLivePartitionCheck(l), "unknown/unavailable": edgesOK, "legacy count mode": legacy} {
			if !ok {
				executed = false
				t.Logf("numeric comparison oracle rejected %s mutation %d", label, mutation)
			}
		}
		record("execute", executed)
		refusals := map[string]bool{}
		for _, r := range l.Refusals {
			var body struct{ Error string }
			if r.Status == 400 && json.Unmarshal(r.Body, &body) == nil && body.Error != "" {
				refusals[r.Kind] = true
			}
		}
		rejected := true
		for _, kind := range []string{"scope", "numeric-false", "unit", "type", "missing-summary", "missing-group-summary"} {
			rejected = rejected && refusals[kind]
		}
		record("inspect", ncLiveEventCheck(l) && rejected)
		saved := l.Saved
		capture := saved.Comparison != nil && saved.Receipt.Question.Kind == "comparison" && saved.Receipt.Question.Comparison != nil && saved.Comparison.Request.Numeric && ncLivePairCheck(saved.Comparison.Request, *saved.Comparison.Result, l) && nhLiveSameQuery(saved.Comparison.Request.Baseline.Question, prior.Baseline.Question) && nhLiveSameQuery(saved.Comparison.Request.Selection.Question, prior.Selection.Question) && nhLiveSameQuery(saved.Receipt.Question.Comparison.Baseline, prior.Baseline.Question) && nhLiveSameQuery(saved.Receipt.Question.Comparison.Selection, prior.Selection.Question) && len(saved.Receipt.ID) == 64
		record("capture", capture)
		reused := l.Reused
		views := l.View
		viewBase := views.Breakdowns["spans"].Baseline
		reuse := len(views.ID) == 64 && len(views.Queries) == 1 && nhLiveSameQuery(views.Queries[0], l.Partition.Selected.Question) && viewBase != nil && nhLiveSameQuery(*viewBase, l.Partition.Rest.Question) && ncLiveResult(reused.Selected, l) && ncLiveResult(reused.Rest, l) && nhLiveSameQuery(reused.Selected.Question, l.Partition.Selected.Question) && nhLiveSameQuery(reused.Rest.Question, l.Partition.Rest.Question) && reflect.DeepEqual(reused.Request.Selection, ncLiveCohortOf(reused.Selected)) && reflect.DeepEqual(reused.Request.Baseline, ncLiveCohortOf(reused.Rest)) && reused.Request.Numeric && ncLivePairCheck(reused.Request, reused.Response, l)
		record("reuse", reuse)
		record("reopen", reflect.DeepEqual(l.Saved, l.Reopened) && l.BeforeReopen == l.AfterReopen)
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
		t.Fatal("actual numeric comparison ledger rejected")
	}
	for i := 0; i < 22; i++ {
		if check(i) {
			t.Fatalf("numeric comparison mutation accepted: %d", i)
		}
	}
}
