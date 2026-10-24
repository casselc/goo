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

type numericHistogramLiveView struct {
	Bins    []exploration.NumericHistogramBin `json:"bins"`
	Binding exploration.NumericBinding        `json:"binding"`
}
type numericHistogramLiveRecord struct {
	Raw, ResourceRaw *string
	TimeMS           int64  `json:"timeMs"`
	ID               string `json:"id"`
}
type numericHistogramLiveLedger struct {
	Version                                                            int
	Binding                                                            exploration.NumericBinding
	Records                                                            []numericHistogramLiveRecord
	Original                                                           exploration.Query
	Plan                                                               exploration.PackagePlan
	Result, Selected, Rest, ComparisonSelected, ComparisonRest, Reused exploration.Result
	InvalidOnly, NoData                                                exploration.Result
	Selection                                                          exploration.NumericHistogramSelection
	CountPlan                                                          exploration.NumericHistogramSelection
	Events                                                             []struct {
		Action                       string
		Reads, Plans, FromBin, ToBin int
	}
	Saved, Reopened                                           exploration.CapturePage
	View, SavedView                                           numericHistogramLiveView
	Before, AfterStage, AfterBrush, BeforeReopen, AfterReopen int
}

// Independent expected answers: stdlib parsing and canonical decimal roundtrip,
// not production parsing, SQL or evaluation. Sentinels retain arbitrary precision.
func nhLiveInteger(raw *string, typ string) (*big.Int, string) {
	if raw == nil {
		return nil, "missing"
	}
	if *raw == "" {
		return nil, "empty"
	}
	if typ == "int64" {
		n, e := strconv.ParseInt(*raw, 10, 64)
		if e == nil && strconv.FormatInt(n, 10) == *raw {
			return big.NewInt(n), "valid"
		}
	} else if typ == "uint64" {
		n, e := strconv.ParseUint(*raw, 10, 64)
		if e == nil && strconv.FormatUint(n, 10) == *raw {
			return new(big.Int).SetUint64(n), "valid"
		}
	}
	return nil, "invalid"
}
func nhLiveEdges(typ string) []string {
	edges := []string{}
	if typ == "int64" {
		for i := 63; i >= 0; i-- {
			edges = append(edges, new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), uint(i))).String())
		}
	}
	edges = append(edges, "0")
	max := 64
	if typ == "int64" {
		max = 63
	}
	for i := 0; i <= max; i++ {
		edges = append(edges, new(big.Int).Lsh(big.NewInt(1), uint(i)).String())
	}
	return edges
}
func nhLiveMatch(p *exploration.Predicate, row numericHistogramLiveRecord, binding exploration.NumericBinding) (bool, error) {
	if p == nil {
		return true, nil
	}
	if len(p.All) > 0 {
		for i := range p.All {
			ok, e := nhLiveMatch(&p.All[i], row, binding)
			if e != nil || !ok {
				return false, e
			}
		}
		return true, nil
	}
	if len(p.Any) > 0 {
		for i := range p.Any {
			ok, e := nhLiveMatch(&p.Any[i], row, binding)
			if e != nil {
				return false, e
			}
			if ok {
				return true, nil
			}
		}
		return false, nil
	}
	field := p.Field
	if p.Numeric != nil {
		field = &p.Numeric.Field
	}
	if field == nil {
		return false, fmt.Errorf("missing predicate field")
	}
	var raw *string
	if field.Key != binding.Field.Key {
		return false, fmt.Errorf("fixture does not define %v", field)
	}
	switch field.Kind {
	case "attribute":
		raw = row.Raw
	case "resource":
		raw = row.ResourceRaw
	default:
		return false, fmt.Errorf("unsupported fixture field kind")
	}
	if p.Numeric != nil {
		if *p.Numeric != binding {
			return false, fmt.Errorf("numeric predicate changed binding")
		}
		v, state := nhLiveInteger(raw, p.Numeric.Type)
		switch p.Op {
		case "valid", "missing", "empty", "invalid":
			return state == p.Op, nil
		}
		if state != "valid" {
			return false, nil
		}
		threshold, s := nhLiveInteger(&p.Value, p.Numeric.Type)
		if s != "valid" {
			return false, fmt.Errorf("noncanonical threshold")
		}
		cmp := v.Cmp(threshold)
		switch p.Op {
		case "eq":
			return cmp == 0, nil
		case "ne":
			return cmp != 0, nil
		case "gt":
			return cmp > 0, nil
		case "gte":
			return cmp >= 0, nil
		case "lt":
			return cmp < 0, nil
		case "lte":
			return cmp <= 0, nil
		}
	} else {
		switch p.Op {
		case "missing":
			return raw == nil, nil
		case "exists":
			return raw != nil, nil
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
	return false, fmt.Errorf("unsupported fixture predicate %q", p.Op)
}
func nhLiveMembers(q exploration.Query, l numericHistogramLiveLedger) ([]numericHistogramLiveRecord, error) {
	if q.Environment != l.Original.Environment || q.Signal != l.Original.Signal {
		return nil, fmt.Errorf("changed source scope")
	}
	rows := []numericHistogramLiveRecord{}
	for _, row := range l.Records {
		if row.TimeMS < q.FromMS || row.TimeMS >= q.ToMS {
			continue
		}
		ok, e := nhLiveMatch(q.Where, row, l.Binding)
		if e != nil {
			return nil, e
		}
		if ok {
			rows = append(rows, row)
		}
	}
	return rows, nil
}
func nhLiveSummary(b exploration.NumericBinding, rows []numericHistogramLiveRecord) (exploration.NumericSummary, []exploration.NumericHistogramBin) {
	edges := nhLiveEdges(b.Type)
	bins := make([]exploration.NumericHistogramBin, len(edges)-1)
	counts := make([]int, len(bins))
	coverage := map[string]int{}
	sum := new(big.Int)
	var low, high *big.Int
	for _, row := range rows {
		raw := row.Raw
		if b.Field.Kind == "resource" {
			raw = row.ResourceRaw
		}
		value, state := nhLiveInteger(raw, b.Type)
		coverage[state]++
		if state != "valid" {
			continue
		}
		sum.Add(sum, value)
		if low == nil || value.Cmp(low) < 0 {
			low = new(big.Int).Set(value)
		}
		if high == nil || value.Cmp(high) > 0 {
			high = new(big.Int).Set(value)
		}
		for i := range counts {
			a, _ := new(big.Int).SetString(edges[i], 10)
			z, _ := new(big.Int).SetString(edges[i+1], 10)
			if value.Cmp(a) >= 0 && value.Cmp(z) < 0 {
				counts[i]++
				break
			}
		}
	}
	for i := range bins {
		bins[i] = exploration.NumericHistogramBin{From: edges[i], To: edges[i+1], Count: strconv.Itoa(counts[i])}
	}
	s := exploration.NumericSummary{State: "no_valid_values", Unit: b.Unit, Count: strconv.Itoa(len(rows)), Valid: strconv.Itoa(coverage["valid"]), Missing: strconv.Itoa(coverage["missing"]), Empty: strconv.Itoa(coverage["empty"]), Invalid: strconv.Itoa(coverage["invalid"]), Sum: sum.String()}
	if len(rows) == 0 {
		s.State = "no_data"
	}
	if low != nil {
		s.State = "observed"
		s.Min = low.String()
		s.Max = high.String()
	}
	return s, bins
}
func nhLiveSameQuery(a, b exploration.Query) bool {
	x, e := json.Marshal(a)
	y, f := json.Marshal(b)
	return e == nil && f == nil && string(x) == string(y)
}
func nhLiveResult(r *exploration.Result, q exploration.Query, l numericHistogramLiveLedger) bool {
	if r == nil || !nhLiveSameQuery(r.Question, q) || r.Basis == "" || r.CapturedAt == "" {
		return false
	}
	rows, e := nhLiveMembers(q, l)
	if e != nil || r.Count != strconv.Itoa(len(rows)) {
		return false
	}
	if len(rows) > 0 && r.State != "observed" || len(rows) == 0 && r.State != "no_data" {
		return false
	}
	if q.Numeric != nil {
		if *q.Numeric != l.Binding {
			return false
		}
		summary, bins := nhLiveSummary(*q.Numeric, rows)
		if r.Numeric == nil || *r.Numeric != summary {
			return false
		}
		if q.NumericHistogram != nil {
			if r.NumericHistogram == nil || r.NumericHistogram.Binning != "powers-of-two" || !reflect.DeepEqual(r.NumericHistogram.Bins, bins) {
				return false
			}
		} else if r.NumericHistogram != nil {
			return false
		}
	} else if r.Numeric != nil || r.NumericHistogram != nil {
		return false
	}
	// Pagination never defines histogram totals. Each visible row must match an
	// independently emitted record and its exact optional attribute/resource value.
	expected := len(rows) - q.Offset
	if expected < 0 {
		expected = 0
	}
	if expected > q.Limit {
		expected = q.Limit
	}
	if len(r.Rows) != expected || r.HasMore != (q.Offset+len(r.Rows) < len(rows)) {
		return false
	}
	fixture := map[int64]numericHistogramLiveRecord{}
	for _, row := range rows {
		if _, exists := fixture[row.TimeMS]; exists {
			return false
		}
		fixture[row.TimeMS] = row
	}
	seen := map[int64]bool{}
	for _, raw := range r.Rows {
		var row struct {
			Timestamp  string            `json:"timestamp_ns"`
			SpanID     string            `json:"SpanId"`
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
		emitted, ok := fixture[ns/1000000]
		if !ok || seen[ns/1000000] || ns != emitted.TimeMS*1000000 || row.SpanID != emitted.ID {
			return false
		}
		seen[ns/1000000] = true
		for _, pair := range []struct {
			values map[string]string
			raw    *string
		}{{row.Attributes, emitted.Raw}, {row.Resource, emitted.ResourceRaw}} {
			v, present := pair.values[l.Binding.Field.Key]
			if pair.raw == nil && present || pair.raw != nil && (!present || v != *pair.raw) {
				return false
			}
		}
	}
	return !strings.Contains(strings.ToUpper(r.Plan.Count), " LIMIT ") && r.Plan.Count != ""
}
func nhLiveSelection(l numericHistogramLiveLedger) bool {
	s := l.Selection
	edges := nhLiveEdges(l.Binding.Type)
	if s.FromBin < 0 || s.ToBin <= s.FromBin || s.ToBin >= len(edges) || s.Binning != "powers-of-two" || s.From != edges[s.FromBin] || s.To != edges[s.ToBin] || !nhLiveSameQuery(s.Parent, l.Original) {
		return false
	}
	for _, q := range []exploration.Query{s.Selected, s.Rest} {
		want := l.Original
		want.Offset = 0
		want.Where = q.Where
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
		if len(l.Original.GroupBy) > 0 {
			want.GroupBy = append([]exploration.Field{}, l.Original.GroupBy[0])
		} else {
			want.GroupBy = []exploration.Field{{Kind: "column", Key: "service"}}
		}
		if !nhLiveSameQuery(want, q) {
			return false
		}
	}
	low, _ := new(big.Int).SetString(s.From, 10)
	high, _ := new(big.Int).SetString(s.To, 10)
	// Test all excluded states and every domain edge, even when the captured sample
	// did not happen to include a particular boundary value.
	records := append([]numericHistogramLiveRecord{}, l.Records...)
	for _, raw := range []string{"", "01", "9007199254740991", "9007199254740992", "9007199254740993"} {
		v := raw
		records = append(records, numericHistogramLiveRecord{Raw: &v, ResourceRaw: &v})
	}
	records = append(records, numericHistogramLiveRecord{})
	for _, edge := range edges {
		v := edge
		records = append(records, numericHistogramLiveRecord{Raw: &v, ResourceRaw: &v})
	}
	for _, row := range records {
		parent, e := nhLiveMatch(l.Original.Where, row, l.Binding)
		if e != nil {
			return false
		}
		a, e := nhLiveMatch(s.Selected.Where, row, l.Binding)
		if e != nil {
			return false
		}
		z, e := nhLiveMatch(s.Rest.Where, row, l.Binding)
		if e != nil {
			return false
		}
		raw := row.Raw
		if l.Binding.Field.Kind == "resource" {
			raw = row.ResourceRaw
		}
		v, state := nhLiveInteger(raw, l.Binding.Type)
		selected := parent && state == "valid" && v.Cmp(low) >= 0 && v.Cmp(high) < 0
		if a != selected || z != (parent && !selected) || a && z {
			return false
		}
	}
	return reflect.DeepEqual(l.CountPlan, l.Selection) && nhLiveResult(&l.Selected, s.Selected, l) && nhLiveResult(&l.Rest, s.Rest, l) && nhLiveResult(&l.ComparisonSelected, s.Comparison.Selection, l) && nhLiveResult(&l.ComparisonRest, s.Comparison.Baseline, l)
}

// Actual browser/OTLP ledger only; no production evaluator predicts expectations.
func TestQuintNumericHistogramLiveLedger(t *testing.T) {
	file := os.Getenv("GOO_NUMERIC_HISTOGRAM_LEDGER")
	if file == "" {
		t.Skip("actual numeric histogram browser/OTLP ledger opt-in")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l numericHistogramLiveLedger
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch mutation {
		case 0:
			l.Result.NumericHistogram.Bins[0].To = "0"
		case 1:
			l.Result.Numeric.Sum = "9007199254740992"
		case 2:
			l.AfterBrush++
		case 3:
			l.Selection.Rest.Where = l.Selection.Selected.Where
		case 4:
			l.Selection.Selected.Environment = "foreign"
		case 5:
			l.Selection.Selected.FromMS++
		case 6:
			l.Saved.Explore.NumericHistogram.Bins[0].Count = "900"
		case 7:
			l.Reopened.Explore.NumericHistogram.Bins[0].From = "rounded"
		case 8:
			l.Reused.Question.NumericHistogram.Binning = "linear"
		case 9:
			l.AfterReopen++
		case 10:
			l.View.Bins[0].Count = "42"
		case 11:
			l.AfterStage++
		case 12:
			l.Selection.Comparison.Baseline.Where = l.Selection.Selected.Where
		case 14:
			l.InvalidOnly.Numeric.Invalid = "0"
		case 15:
			l.NoData.NumericHistogram.Bins[0].Count = "1"
		case 16:
			l.CountPlan.Comparison.Selection.ToMS++
		case 17:
			l.Events[0].Plans++
		case 13:
			for i := range l.Records {
				if l.Records[i].Raw != nil && *l.Records[i].Raw == "9007199254740993" {
					v := "9007199254740992"
					l.Records[i].Raw = &v
					break
				}
			}
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		record := func(action string, ok bool) {
			if !ok {
				t.Logf("histogram ledger rejected phase %s (mutation %d)", action, mutation)
			}
			rec.Record(context.Background(), action, nil, "obs.coherent", ok)
		}
		planned := l.Version == 1 && len(l.Records) > 100 && l.Original.Numeric != nil && *l.Original.Numeric == l.Binding && l.Original.NumericHistogram != nil && l.Original.NumericHistogram.Binning == "powers-of-two" && nhLiveSameQuery(l.Plan.Question, l.Original) && l.Plan.MeasurePlan.NumericHistogram != nil && l.Plan.MeasurePlan.NumericHistogram.Binning == "powers-of-two" && reflect.DeepEqual(l.Plan.MeasurePlan.Numeric, &l.Binding)
		hasLarge := false
		for _, row := range l.Records {
			hasLarge = hasLarge || row.Raw != nil && *row.Raw == "9007199254740993"
		}
		record("plan", planned && hasLarge)
		empty := l
		empty.Original.Environment = l.NoData.Question.Environment
		empty.Records = nil
		onlyInvalid := l.Original
		onlyInvalid.Where = &exploration.Predicate{Numeric: &l.Binding, Op: "invalid"}
		noDataQ := l.Original
		noDataQ.Environment += "-absent"
		if !reflect.DeepEqual(l.CountPlan, l.Selection) {
			t.Logf("selected and count-comparison plans differ: selected bins [%d,%d), count bins [%d,%d)", l.Selection.FromBin, l.Selection.ToBin, l.CountPlan.FromBin, l.CountPlan.ToBin)
		}
		executed := true
		for label, ok := range map[string]bool{
			"full population":         nhLiveResult(&l.Result, l.Original, l),
			"selected/rest partition": nhLiveSelection(l),
			"invalid-only population": nhLiveResult(&l.InvalidOnly, onlyInvalid, l),
			"no-data scope":           nhLiveSameQuery(noDataQ, l.NoData.Question),
			"no-data population":      nhLiveResult(&l.NoData, noDataQ, empty),
		} {
			if !ok {
				executed = false
				t.Logf("histogram oracle rejected %s (mutation %d)", label, mutation)
			}
		}
		record("execute", executed)
		eventsOK := len(l.Events) == 4 &&
			l.Events[0].Action == "brush" && l.Events[0].Reads == l.Before && l.Events[0].Plans == 0 && l.Events[0].FromBin == l.Selection.FromBin && l.Events[0].ToBin == l.Selection.ToBin &&
			l.Events[1].Action == "run-selected" && l.Events[1].Reads > l.Events[0].Reads && l.Events[1].Plans == 1 &&
			l.Events[2].Action == "back-and-brush" && l.Events[2].Reads == l.Events[1].Reads && l.Events[2].Plans == l.Events[1].Plans &&
			l.Events[3].Action == "run-comparison" && l.Events[3].Reads == l.Events[1].Reads+2 && l.Events[3].Plans == 2
		record("inspect", eventsOK && l.Before == l.AfterStage && l.Before == l.AfterBrush && l.Before >= 0 && reflect.DeepEqual(l.View.Binding, l.Binding) && l.Result.NumericHistogram != nil && reflect.DeepEqual(l.View.Bins, l.Result.NumericHistogram.Bins))
		capture := l.Saved.Explore != nil && l.Saved.Receipt.Question.Explore != nil && nhLiveSameQuery(*l.Saved.Receipt.Question.Explore, l.Original) && nhLiveResult(l.Saved.Explore, l.Original, l) && reflect.DeepEqual(l.Saved.Explore.NumericHistogram, l.Result.NumericHistogram) && reflect.DeepEqual(l.Saved.Explore.Numeric, l.Result.Numeric) && len(l.Saved.Receipt.ID) == 64
		record("capture", capture)
		reuse := l.Reused.Question
		reuse.FromMS = l.Original.FromMS
		reuse.ToMS = l.Original.ToMS
		record("reuse", nhLiveSameQuery(reuse, l.Original) && nhLiveResult(&l.Reused, l.Reused.Question, l))
		record("reopen", l.BeforeReopen == l.AfterReopen && reflect.DeepEqual(l.Saved, l.Reopened) && reflect.DeepEqual(l.View, l.SavedView))
		checker, e := validate.NewChecker("model/grouped_series.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("actual numeric histogram ledger rejected")
	}
	for mutation := 0; mutation < 18; mutation++ {
		if check(mutation) {
			t.Fatalf("corrupt numeric histogram ledger accepted: mutation %d", mutation)
		}
	}
}
