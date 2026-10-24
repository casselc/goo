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
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type peLiveRecord struct {
	Environment, Signal, ID, Service, Status, Severity string
	TimeMS                                             int64 `json:"timeMs"`
	Attributes                                         map[string]string
	Resource                                           map[string]string
}
type peLiveCase struct {
	Environment string
	Records     []peLiveRecord
	Request     exploration.ExplanationRequest
	Result      exploration.ExplanationResult
}
type peLiveValue struct {
	Missing bool
	Text    string
}

func peLiveValueOf(v *string) peLiveValue {
	if v == nil {
		return peLiveValue{Missing: true}
	}
	return peLiveValue{Text: *v}
}
func peLivePointer(v peLiveValue) *string {
	if v.Missing {
		return nil
	}
	s := v.Text
	return &s
}
func peLiveLess(a, z peLiveValue) bool {
	if a.Missing != z.Missing {
		return a.Missing
	}
	return a.Text < z.Text
}
func peLiveField(f exploration.Field, r peLiveRecord) (*string, error) {
	var text string
	switch f.Kind {
	case "attribute":
		v, ok := r.Attributes[f.Key]
		if !ok {
			return nil, nil
		}
		return &v, nil
	case "resource":
		if f.Key == "service.name" {
			return &r.Service, nil
		}
		if f.Key == "deployment.environment.name" {
			return &r.Environment, nil
		}
		v, ok := r.Resource[f.Key]
		if !ok {
			return nil, nil
		}
		return &v, nil
	case "column":
		switch f.Key {
		case "service":
			text = r.Service
		case "status":
			if r.Signal != "spans" {
				return nil, fmt.Errorf("status requires spans")
			}
			text = r.Status
		case "severity":
			if r.Signal != "logs" {
				return nil, fmt.Errorf("severity requires logs")
			}
			text = r.Severity
		case "span_id":
			text = r.ID
		case "timestamp_ns":
			text = strconv.FormatInt(r.TimeMS*1000000, 10)
		default:
			return nil, fmt.Errorf("fixture does not define column %s", f.Key)
		}
	default:
		return nil, fmt.Errorf("unsupported fixture namespace")
	}
	return &text, nil
}
func peLiveMatch(p *exploration.Predicate, row peLiveRecord) (bool, error) {
	if p == nil {
		return true, nil
	}
	if len(p.All) > 0 {
		for i := range p.All {
			ok, e := peLiveMatch(&p.All[i], row)
			if e != nil || !ok {
				return false, e
			}
		}
		return true, nil
	}
	if len(p.Any) > 0 {
		for i := range p.Any {
			ok, e := peLiveMatch(&p.Any[i], row)
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
		return false, fmt.Errorf("missing fixture predicate field")
	}
	raw, e := peLiveField(*field, row)
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
			return false, fmt.Errorf("invalid numeric threshold")
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
	return false, fmt.Errorf("unsupported fixture operator %s", p.Op)
}
func peLiveRows(q exploration.Query, environment string, records []peLiveRecord) ([]peLiveRecord, error) {
	if q.Environment != environment || q.Signal != "spans" && q.Signal != "logs" || q.FromMS < 0 || q.ToMS <= q.FromMS {
		return nil, fmt.Errorf("wrong source scope")
	}
	out := []peLiveRecord{}
	for _, row := range records {
		if row.Environment != "" && row.Environment != environment || row.Signal != q.Signal || row.TimeMS < q.FromMS || row.TimeMS >= q.ToMS {
			continue
		}
		ok, e := peLiveMatch(q.Where, row)
		if e != nil {
			return nil, e
		}
		if ok {
			out = append(out, row)
		}
	}
	return out, nil
}
func peLiveGroups(field exploration.Field, rows []peLiveRecord) ([]exploration.ExplanationValueCount, error) {
	groups := map[peLiveValue]int{}
	for _, row := range rows {
		raw, e := peLiveField(field, row)
		if e != nil {
			return nil, e
		}
		groups[peLiveValueOf(raw)]++
	}
	keys := []peLiveValue{}
	for value := range groups {
		keys = append(keys, value)
	}
	sort.Slice(keys, func(i, j int) bool { return peLiveLess(keys[i], keys[j]) })
	out := []exploration.ExplanationValueCount{}
	for _, value := range keys {
		out = append(out, exploration.ExplanationValueCount{Value: peLivePointer(value), Count: strconv.Itoa(groups[value])})
	}
	return out, nil
}
func peLivePercent(count, total string) string {
	if total == "0" {
		return ""
	}
	return new(big.Rat).Mul(new(big.Rat).SetFrac(ncLiveInt(count), ncLiveInt(total)), big.NewRat(100, 1)).RatString()
}
func peLiveRowsOfGroups(a, z []exploration.ExplanationValueCount, ac, zc string) []exploration.ExplanationRow {
	aa, zz := map[peLiveValue]string{}, map[peLiveValue]string{}
	values := map[peLiveValue]bool{}
	for _, v := range a {
		key := peLiveValueOf(v.Value)
		aa[key] = v.Count
		values[key] = true
	}
	for _, v := range z {
		key := peLiveValueOf(v.Value)
		zz[key] = v.Count
		values[key] = true
	}
	keys := []peLiveValue{}
	for key := range values {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return peLiveLess(keys[i], keys[j]) })
	rows := []exploration.ExplanationRow{}
	for _, key := range keys {
		x, y := aa[key], zz[key]
		if x == "" {
			x = "0"
		}
		if y == "" {
			y = "0"
		}
		r := exploration.ExplanationRow{Value: peLivePointer(key), BaselineCount: x, SelectionCount: y, BaselinePercent: peLivePercent(x, ac), SelectionPercent: peLivePercent(y, zc)}
		if r.BaselinePercent != "" && r.SelectionPercent != "" {
			ap, _ := new(big.Rat).SetString(r.BaselinePercent)
			zp, _ := new(big.Rat).SetString(r.SelectionPercent)
			r.DifferencePP = new(big.Rat).Sub(zp, ap).RatString()
		}
		rows = append(rows, r)
	}
	return rows
}
func peLiveClock(s string) bool { _, e := time.Parse(time.RFC3339Nano, s); return e == nil }
func peLiveHex(s string) bool {
	b, e := hex.DecodeString(s)
	return e == nil && len(b) == 32 && hex.EncodeToString(b) == s
}
func peLiveResultBasis(r exploration.ExplanationResult) string {
	r.Basis = ""
	r.ResultToken = ""
	raw, _ := json.Marshal(r)
	hash := sha256.Sum256(raw)
	return hex.EncodeToString(hash[:])
}
func peLiveFieldExpected(actual exploration.ExplanationFieldResult, base, selection []peLiveRecord, ac, zc string) (exploration.ExplanationFieldResult, bool) {
	a, e := peLiveGroups(actual.Field, base)
	if e != nil {
		return exploration.ExplanationFieldResult{}, false
	}
	z, e := peLiveGroups(actual.Field, selection)
	if e != nil {
		return exploration.ExplanationFieldResult{}, false
	}
	out := exploration.ExplanationFieldResult{Field: actual.Field, State: "complete", Baseline: actual.Baseline, Selection: actual.Selection, Rows: []exploration.ExplanationRow{}}
	for i, p := range []*exploration.ExplanationFieldRead{&out.Baseline, &out.Selection} {
		values := a
		if i == 1 {
			values = z
		}
		p.State = "complete"
		p.Reason = ""
		p.Values = values
		if !peLiveClock(p.ReadAt) || !peLiveHex(p.Basis) {
			return out, false
		}
		if len(values) > 1000 {
			p.State = "partial"
			p.Reason = "cardinality-limit"
			p.Values = []exploration.ExplanationValueCount{}
		}
	}
	if out.Baseline.State == "partial" || out.Selection.State == "partial" {
		out.State = "partial"
		out.Reason = "cardinality-limit"
	} else {
		out.Rows = peLiveRowsOfGroups(a, z, ac, zc)
		if ac == "0" || zc == "0" {
			out.State = "no_data"
			out.Reason = "empty-population"
		}
	}
	// The byte budget covers the full derived field, not just retained rows. Use
	// independent source facts with the actual read metadata to measure that wire.
	raw, _ := json.Marshal(out)
	if len(raw) > 128<<10 {
		out.State = "unavailable"
		out.Reason = "field-bytes-limit"
		out.Rows = []exploration.ExplanationRow{}
		for _, p := range []*exploration.ExplanationFieldRead{&out.Baseline, &out.Selection} {
			p.State = "unavailable"
			p.Reason = "field-bytes-limit"
			p.Values = []exploration.ExplanationValueCount{}
		}
	}
	return out, true
}
func peLiveResult(request exploration.ExplanationRequest, result exploration.ExplanationResult, environment string, records []peLiveRecord) bool {
	if request.Version != "goo/explanation/v1" || result.Version != request.Version || !reflect.DeepEqual(request, result.Request) || len(request.Fields) < 1 || len(request.Fields) > 6 || len(result.Fields) != len(request.Fields) || request.Baseline.Signal != request.Selection.Signal || request.Baseline.Environment != request.Selection.Environment || !peLiveClock(result.ReadAt) || !peLiveHex(result.Basis) || result.Basis != peLiveResultBasis(result) {
		return false
	}
	base, e := peLiveRows(request.Baseline, environment, records)
	if e != nil {
		return false
	}
	selection, e := peLiveRows(request.Selection, environment, records)
	if e != nil {
		return false
	}
	ac, zc := strconv.Itoa(len(base)), strconv.Itoa(len(selection))
	if result.Baseline.Count != ac || result.Selection.Count != zc || !peLiveClock(result.Baseline.ReadAt) || !peLiveClock(result.Selection.ReadAt) || !peLiveHex(result.Baseline.Basis) || !peLiveHex(result.Selection.Basis) {
		return false
	}
	finished, _ := time.Parse(time.RFC3339Nano, result.ReadAt)
	clocks := map[string]bool{}
	for _, population := range []exploration.ExplanationPopulation{result.Baseline, result.Selection} {
		at, _ := time.Parse(time.RFC3339Nano, population.ReadAt)
		if at.After(finished) {
			return false
		}
		clocks[population.ReadAt] = true
	}
	for _, field := range result.Fields {
		for _, side := range []exploration.ExplanationFieldRead{field.Baseline, field.Selection} {
			at, e := time.Parse(time.RFC3339Nano, side.ReadAt)
			if e != nil || at.After(finished) {
				return false
			}
			clocks[side.ReadAt] = true
		}
	}
	if len(clocks) < 3 {
		return false
	}
	expectedFields := make([]exploration.ExplanationFieldResult, len(request.Fields))
	kinds := map[string]int{}
	seen := map[exploration.Field]bool{}
	rank := []exploration.ExplanationLead{}
	for i, field := range result.Fields {
		if field.Field != request.Fields[i] || seen[field.Field] {
			return false
		}
		seen[field.Field] = true
		expected, ok := peLiveFieldExpected(field, base, selection, ac, zc)
		if !ok || !reflect.DeepEqual(field, expected) {
			return false
		}
		expectedFields[i] = expected
		kinds[expected.State]++
		if expected.State == "complete" {
			for j, row := range expected.Rows {
				if row.DifferencePP != "0" {
					rank = append(rank, exploration.ExplanationLead{FieldIndex: i, RowIndex: j})
				}
			}
		}
	}
	sort.Slice(rank, func(i, j int) bool {
		x, y := rank[i], rank[j]
		a, z := expectedFields[x.FieldIndex], expectedFields[y.FieldIndex]
		ar, zr := a.Rows[x.RowIndex], z.Rows[y.RowIndex]
		an, _ := new(big.Rat).SetString(ar.DifferencePP)
		zn, _ := new(big.Rat).SetString(zr.DifferencePP)
		if cmp := new(big.Rat).Abs(an).Cmp(new(big.Rat).Abs(zn)); cmp != 0 {
			return cmp > 0
		}
		if a.Field.Kind != z.Field.Kind {
			return a.Field.Kind < z.Field.Kind
		}
		if a.Field.Key != z.Field.Key {
			return a.Field.Key < z.Field.Key
		}
		return peLiveLess(peLiveValueOf(ar.Value), peLiveValueOf(zr.Value))
	})
	state := "complete"
	if kinds["complete"] != len(request.Fields) {
		switch {
		case kinds["no_data"] == len(request.Fields):
			state = "no_data"
		case kinds["unavailable"] == len(request.Fields):
			state = "unavailable"
		case kinds["changed_between_reads"] == len(request.Fields):
			state = "changed_between_reads"
		default:
			state = "partial"
		}
	}
	return result.State == state && reflect.DeepEqual(result.Ranked, rank)
}
func peLiveCaptureHash(page exploration.CapturePage) bool {
	c := exploration.SavedCapture{Explanation: page.Explanation, CaptureReceipt: page.Receipt, Measurement: page.Measurement, Comparison: page.Comparison, Explore: page.Explore, Inventory: page.Inventory}
	id := c.ID
	c.ID = ""
	raw, e := json.Marshal(c)
	if e != nil {
		return false
	}
	hash := sha256.Sum256(raw)
	return id == hex.EncodeToString(hash[:])
}

type peLiveRun struct {
	Request exploration.ExplanationRequest
	Result  exploration.ExplanationResult
}
type peLiveRange struct {
	Request struct {
		Parent         exploration.Query
		FromBin, ToBin int
	}
	Response exploration.NumericHistogramSelection
}
type peLiveFilter struct {
	Request struct {
		Request    exploration.ExplanationRequest
		Field      exploration.Field
		Value      *string
		Side, Mode string
	}
	Response struct{ Question exploration.Query }
	Result   exploration.Result
}
type peLiveLedger struct {
	Version                       int
	Environment                   string
	Records, LateRecords          []peLiveRecord
	Fields                        []exploration.Field
	Parent, ReadyProbe            exploration.Result
	Range                         peLiveRange
	Observed, Reused, Logs, Equal peLiveRun
	Plan                          struct {
		Request  exploration.ExplanationRequest
		Response exploration.ExplanationPlan
	}
	Promotions []struct {
		Request      struct{ Title, ResultToken string }
		Response     exploration.CapturePage
		Acknowledged bool
	}
	Saved, Reopened exploration.CapturePage
	Filters         []peLiveFilter
	HighCard        struct {
		Environment string
		Records     []peLiveRecord
		Parent      exploration.Result
		Range       peLiveRange
		Request     exploration.ExplanationRequest
		Result      exploration.ExplanationResult
	}
	Refusals []struct {
		Kind, URL string
		Status    int
		Response  json.RawMessage
	}
	Events []struct {
		Action                           string
		SourceOps, Promotions, PurePlans int
	}
	Requests []struct {
		Effect, URL string
		Body        json.RawMessage
	}
}

func peLiveCountQuestion(q exploration.Query) exploration.Query {
	q.Numeric = nil
	q.NumericHistogram = nil
	q.Summary = ""
	q.SpanStatus = false
	q.BucketMS = 0
	q.GroupTimeSeries = false
	q.GroupBy = nil
	q.Offset = 0
	q.Limit = 100
	return q
}
func peLiveSource(r exploration.Result, environment string, records []peLiveRecord) bool {
	rows, e := peLiveRows(r.Question, environment, records)
	if e != nil || r.Count != strconv.Itoa(len(rows)) || !peLiveClock(r.CapturedAt) || !peLiveHex(r.Basis) {
		return false
	}
	state := "observed"
	if len(rows) == 0 {
		state = "no_data"
	}
	if r.State != state {
		return false
	}
	if r.Question.Numeric != nil {
		b := *r.Question.Numeric
		input := make([]numericHistogramLiveRecord, len(rows))
		for i, row := range rows {
			raw, e := peLiveField(b.Field, row)
			if e != nil {
				return false
			}
			input[i] = numericHistogramLiveRecord{Raw: raw, ResourceRaw: raw}
		}
		summary, bins := nhLiveSummary(b, input)
		if r.Numeric == nil || *r.Numeric != summary {
			return false
		}
		if r.Question.NumericHistogram != nil && (r.NumericHistogram == nil || r.NumericHistogram.Binning != "powers-of-two" || !reflect.DeepEqual(r.NumericHistogram.Bins, bins)) {
			return false
		}
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
	fixture := map[string]peLiveRecord{}
	for _, row := range rows {
		fixture[row.ID] = row
	}
	seen := map[string]bool{}
	for _, raw := range r.Rows {
		var row struct {
			ID         string `json:"SpanId"`
			Time       string `json:"timestamp_ns"`
			Service    string `json:"ServiceName"`
			Status     string `json:"StatusCode"`
			Severity   string `json:"SeverityText"`
			Body       string
			Attributes map[string]string
			Resource   map[string]string `json:"ResourceAttributes"`
		}
		if json.Unmarshal(raw, &row) != nil {
			return false
		}
		id := row.ID
		if r.Question.Signal == "logs" {
			id = strings.TrimPrefix(row.Body, "work observed ")
		}
		fact, ok := fixture[id]
		if !ok || seen[id] || row.Service != fact.Service || row.Time != strconv.FormatInt(fact.TimeMS*1000000, 10) || r.Question.Signal == "spans" && row.Status != fact.Status || r.Question.Signal == "logs" && row.Severity != fact.Severity {
			return false
		}
		seen[id] = true
		if !reflect.DeepEqual(row.Attributes, fact.Attributes) {
			return false
		}
		for key, value := range fact.Resource {
			if row.Resource[key] != value {
				return false
			}
		}
		if row.Resource["deployment.environment.name"] != environment || row.Resource["service.name"] != fact.Service {
			return false
		}
	}
	return r.Plan.Count != "" && !strings.Contains(strings.ToUpper(r.Plan.Count), " LIMIT ") && r.Plan.Parameters["environment"] == environment && r.Plan.Parameters["from"] == strconv.FormatInt(r.Question.FromMS, 10) && r.Plan.Parameters["to"] == strconv.FormatInt(r.Question.ToMS, 10)
}
func peLiveRangeCheck(plan peLiveRange, parent exploration.Result, environment string, records []peLiveRecord) bool {
	in, out := plan.Request, plan.Response
	if !peLiveSource(parent, environment, records) || !nhLiveSameQuery(in.Parent, parent.Question) || !nhLiveSameQuery(out.Parent, parent.Question) || parent.Question.Numeric == nil || parent.NumericHistogram == nil {
		return false
	}
	edges := nhLiveEdges(parent.Question.Numeric.Type)
	if in.FromBin < 0 || in.ToBin <= in.FromBin || in.ToBin >= len(edges) || out.FromBin != in.FromBin || out.ToBin != in.ToBin || out.From != edges[in.FromBin] || out.To != edges[in.ToBin] || out.Binning != "powers-of-two" {
		return false
	}
	for _, q := range []exploration.Query{out.Selected, out.Rest} {
		want := parent.Question
		want.Offset = 0
		want.Where = q.Where
		if !nhLiveSameQuery(want, q) {
			return false
		}
	}
	lo, hi := ncLiveInt(out.From), ncLiveInt(out.To)
	for _, row := range records {
		base, e := peLiveMatch(parent.Question.Where, row)
		if e != nil {
			return false
		}
		a, e := peLiveMatch(out.Selected.Where, row)
		if e != nil {
			return false
		}
		z, e := peLiveMatch(out.Rest.Where, row)
		if e != nil {
			return false
		}
		raw, e := peLiveField(parent.Question.Numeric.Field, row)
		if e != nil {
			return false
		}
		v, state := nhLiveInteger(raw, parent.Question.Numeric.Type)
		selected := base && state == "valid" && v.Cmp(lo) >= 0 && v.Cmp(hi) < 0
		if a != selected || z != (base && !selected) || a && z {
			return false
		}
	}
	return true
}
func peLiveFilterScope(p peLiveFilter, request exploration.ExplanationRequest, environment string, records []peLiveRecord) bool {
	in, q := p.Request, p.Response.Question
	if !reflect.DeepEqual(in.Request, request) {
		return false
	}
	found := false
	for _, f := range request.Fields {
		found = found || f == in.Field
	}
	if !found || in.Side != "baseline" && in.Side != "selection" || in.Mode != "include" && in.Mode != "exclude" {
		return false
	}
	base := request.Baseline
	if in.Side == "selection" {
		base = request.Selection
	}
	want := base
	want.Where = q.Where
	want.Offset = 0
	if !nhLiveSameQuery(want, q) {
		return false
	}
	candidates := append([]peLiveRecord{}, records...)
	// Exhaustively separate missing, present-empty and other values, preserving an
	// outside-parent team case for each value so exclusion never negates the parent.
	for _, value := range []*string{nil, new(string), func() *string { v := "elsewhere"; return &v }()} {
		for _, team := range []string{"core", "other"} {
			row := peLiveRecord{Environment: environment, Signal: base.Signal, Attributes: map[string]string{"work.items": "16"}, Resource: map[string]string{"team": team}}
			if value != nil {
				if in.Field.Kind == "attribute" {
					row.Attributes[in.Field.Key] = *value
				} else if in.Field.Kind == "resource" {
					row.Resource[in.Field.Key] = *value
				}
			}
			candidates = append(candidates, row)
		}
	}
	for _, row := range candidates {
		parent, e := peLiveMatch(base.Where, row)
		if e != nil {
			return false
		}
		selected, e := peLiveMatch(q.Where, row)
		if e != nil {
			return false
		}
		raw, e := peLiveField(in.Field, row)
		if e != nil {
			return false
		}
		equal := reflect.DeepEqual(raw, in.Value)
		if in.Mode == "exclude" {
			equal = !equal
		}
		if selected != (parent && equal) {
			return false
		}
	}
	return true
}
func peLivePromotions(l peLiveLedger) bool {
	if len(l.Promotions) != 2 || l.Promotions[0].Acknowledged || !l.Promotions[1].Acknowledged || l.Saved.Explanation == nil || l.Saved.Explore != nil || l.Saved.Measurement != nil || l.Saved.Inventory != nil || l.Saved.Comparison != nil || l.Saved.Receipt.Version != 1 || l.Saved.Receipt.Question.Kind != "explanation" || l.Saved.Receipt.Question.Explanation == nil || !reflect.DeepEqual(*l.Saved.Receipt.Question.Explanation, l.Observed.Request) || !peLiveCaptureHash(l.Saved) {
		return false
	}
	observed := l.Observed.Result
	observed.ResultToken = ""
	if !reflect.DeepEqual(*l.Saved.Explanation, observed) || !peLiveResult(l.Observed.Request, *l.Saved.Explanation, l.Environment, l.Records) {
		return false
	}
	for _, promotion := range l.Promotions {
		if promotion.Request.ResultToken != l.Observed.Result.ResultToken || promotion.Request.Title != l.Saved.Receipt.Title || !reflect.DeepEqual(promotion.Response, l.Saved) {
			return false
		}
	}
	source, e := time.Parse(time.RFC3339Nano, l.Observed.Result.ReadAt)
	if e != nil {
		return false
	}
	probe, e := time.Parse(time.RFC3339Nano, l.ReadyProbe.CapturedAt)
	if e != nil {
		return false
	}
	custody, e := time.Parse(time.RFC3339Nano, l.Saved.Receipt.CapturedAt)
	if e != nil {
		return false
	}
	return !probe.Before(source) && !custody.Before(probe) && custody.After(source) && l.Saved.Explanation.ReadAt != l.Saved.Receipt.CapturedAt
}
func peLivePlan(l peLiveLedger) bool {
	p := l.Plan
	return reflect.DeepEqual(p.Request, l.Observed.Request) && reflect.DeepEqual(p.Response.Request, p.Request) && nhLiveSameQuery(p.Response.Baseline, peLiveCountQuestion(p.Request.Baseline)) && nhLiveSameQuery(p.Response.Selection, peLiveCountQuestion(p.Request.Selection))
}
func peLiveRequests(l peLiveLedger) bool {
	counts := map[string]int{}
	for _, req := range l.Requests {
		counts[req.Effect]++
		if req.Effect == "promotion" {
			var body map[string]json.RawMessage
			if req.URL != "/api/captures" || json.Unmarshal(req.Body, &body) != nil || len(body) != 2 || body["title"] == nil || body["resultToken"] == nil {
				return false
			}
		}
	}
	return counts["source"] == 8 && counts["promotion"] == 2 && counts["plan"] == 5 && len(l.Requests) == 15
}
func TestQuintPopulationExplanationLiveLedger(t *testing.T) {
	file := os.Getenv("GOO_POPULATION_EXPLANATION_LEDGER")
	if file == "" {
		t.Skip("actual journey177 population explanation ledger opt-in")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l peLiveLedger
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		before, _ := json.Marshal(l)
		switch mutation {
		case 0:
			l.Observed.Result.Baseline.Count = "100"
		case 1:
			l.Observed.Result.Fields[0].Rows[0].DifferencePP = "1/7"
		case 2:
			for i := range l.Observed.Result.Fields {
				if l.Observed.Result.Fields[i].Field == (exploration.Field{Kind: "attribute", Key: "tenant"}) {
					l.Observed.Result.Fields[i].Rows[0].Value = new(string)
				}
			}
		case 3:
			l.Observed.Result.Ranked[0], l.Observed.Result.Ranked[1] = l.Observed.Result.Ranked[1], l.Observed.Result.Ranked[0]
		case 4:
			l.HighCard.Result.Fields[1].Rows = []exploration.ExplanationRow{{BaselineCount: "1", SelectionCount: "1", DifferencePP: "0"}}
		case 5:
			fresh := l.Reused.Result
			fresh.ResultToken = ""
			l.Saved.Explanation = &fresh
		case 6:
			l.Reopened.Explanation.ReadAt = l.Saved.Receipt.CapturedAt
		case 7:
			l.Events[6].SourceOps++
		case 8:
			l.Filters[0].Response.Question.Environment += "-foreign"
		case 9:
			l.Filters[2].Response.Question.Where = l.Filters[1].Response.Question.Where
		case 10:
			l.Reused.Result.Selection.Count = "160"
		case 11:
			l.Promotions[1].Response.Receipt.ID = strings.Repeat("f", 64)
		case 12:
			l.Saved.Explanation.ReadAt = l.Saved.Receipt.CapturedAt
		case 13:
			l.Observed.Result.Fields[0].Field.Kind = "resource"
		case 14:
			l.Plan.Response.Selection.Numeric = l.Observed.Request.Selection.Numeric
		case 15:
			l.Filters[1].Request.Value = nil
		case 16:
			l.Logs.Request.Fields[4] = exploration.Field{Kind: "column", Key: "status"}
		case 17:
			l.Promotions[1].Request.ResultToken = strings.Repeat("0", 64)
		case 18:
			l.Events[8].PurePlans++
		case 19:
			l.Filters[0].Response.Question.ToMS++
		case 20:
			l.Observed.Result.Baseline.ReadAt = l.Observed.Result.ReadAt
			l.Observed.Result.Selection.ReadAt = l.Observed.Result.ReadAt
			for i := range l.Observed.Result.Fields {
				l.Observed.Result.Fields[i].Baseline.ReadAt = l.Observed.Result.ReadAt
				l.Observed.Result.Fields[i].Selection.ReadAt = l.Observed.Result.ReadAt
			}
		}
		// Rehash mathematical corruptions: rejection must come from the independent
		// source/ranking oracle, not merely a stale checksum left by the mutation.
		if mutation >= 0 && mutation <= 3 || mutation == 13 || mutation == 20 {
			l.Observed.Result.Basis = peLiveResultBasis(l.Observed.Result)
		}
		if mutation == 4 {
			l.HighCard.Result.Basis = peLiveResultBasis(l.HighCard.Result)
		}
		if mutation == 10 {
			l.Reused.Result.Basis = peLiveResultBasis(l.Reused.Result)
		}
		after, _ := json.Marshal(l)
		if mutation >= 0 && string(before) == string(after) {
			t.Fatalf("explanation negative control %d made no change", mutation)
		}
		facts := append(append([]peLiveRecord{}, l.Records...), l.LateRecords...)
		observed := peLiveResult(l.Observed.Request, l.Observed.Result, l.Environment, l.Records) && l.Observed.Result.State == "complete" && peLiveHex(l.Observed.Result.ResultToken) && reflect.DeepEqual(l.Fields, l.Observed.Request.Fields) && len(l.Fields) == 6 && nhLiveSameQuery(l.Observed.Request.Baseline, l.Range.Response.Rest) && nhLiveSameQuery(l.Observed.Request.Selection, l.Range.Response.Selected)
		parent := l.Version == 1 && peLiveSource(l.Parent, l.Environment, l.Records) && peLiveRangeCheck(l.Range, l.Parent, l.Environment, l.Records)
		late := len(l.LateRecords) > 0 && peLiveSource(l.ReadyProbe, l.Environment, facts) && nhLiveSameQuery(l.ReadyProbe.Question, l.Observed.Request.Selection) && l.ReadyProbe.Count == "168"
		reused := reflect.DeepEqual(l.Reused.Request, l.Observed.Request) && peLiveResult(l.Reused.Request, l.Reused.Result, l.Environment, facts) && l.Reused.Result.Selection.Count == "168" && l.Reused.Result.Baseline.Count == "160" && l.Reused.Result.Basis != l.Observed.Result.Basis && peLiveHex(l.Reused.Result.ResultToken) && l.Reused.Result.ResultToken != l.Observed.Result.ResultToken
		filters := len(l.Filters) == 3
		for _, p := range l.Filters {
			filters = filters && peLiveFilterScope(p, l.Reused.Request, l.Environment, facts) && peLiveSource(p.Result, l.Environment, facts) && nhLiveSameQuery(p.Result.Question, p.Response.Question)
		}
		high := peLiveSource(l.HighCard.Parent, l.HighCard.Environment, l.HighCard.Records) && peLiveRangeCheck(l.HighCard.Range, l.HighCard.Parent, l.HighCard.Environment, l.HighCard.Records) && peLiveResult(l.HighCard.Request, l.HighCard.Result, l.HighCard.Environment, l.HighCard.Records) && l.HighCard.Result.State == "partial" && len(l.HighCard.Result.Fields) == 2 && l.HighCard.Result.Fields[0].State == "complete" && l.HighCard.Result.Fields[1].State == "partial" && len(l.HighCard.Result.Ranked) > 0
		aux := peLiveResult(l.Logs.Request, l.Logs.Result, l.Environment, facts) && l.Logs.Request.Baseline.Signal == "logs" && peLiveResult(l.Equal.Request, l.Equal.Result, l.Environment, facts) && nhLiveSameQuery(l.Equal.Request.Baseline, l.Equal.Request.Selection) && len(l.Equal.Result.Ranked) == 0
		refused := map[string]bool{}
		for _, r := range l.Refusals {
			var body struct{ Error string }
			expected := 400
			if r.Kind == "unknown-token" {
				expected = 409
			}
			if r.Status == expected && json.Unmarshal(r.Response, &body) == nil && body.Error != "" {
				refused[r.Kind] = true
			}
		}
		refusals := true
		for _, k := range []string{"unknown-token", "too-many-fields", "duplicate-field", "mixed-signals", "unrequested-filter"} {
			refusals = refusals && refused[k]
		}
		semantic := map[string]bool{"stage": parent, "parent": parent, "range": parent, "fields": observed, "explain": observed && peLivePlan(l) && aux && refusals, "late-arrival": late, "save-lost-ack": peLivePromotions(l), "save-retry": peLivePromotions(l), "reopen": reflect.DeepEqual(l.Saved, l.Reopened), "reuse-stage": reused, "reuse-run": reused, "filter-stage": filters, "filter-open": filters, "filter-run": filters, "filter-return": filters, "high-parent": high, "high-explain": high && peLiveRequests(l)}
		if len(l.Events) != 25 {
			return false
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, event := range l.Events {
			ok := semantic[event.Action]
			if !ok {
				t.Logf("explanation oracle rejected %s mutation %d", event.Action, mutation)
			}
			action := strings.ReplaceAll(event.Action, "-", "_")
			if action == "range" {
				action = "select_range"
			}
			if action == "fields" {
				action = "edit_fields"
			}
			rec.Record(context.Background(), action, nil, "obs.coherent", ok, "obs.sources", event.SourceOps, "obs.promotions", event.Promotions, "obs.plans", event.PurePlans)
		}
		checker, e := validate.NewChecker("model/population_explanation.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("actual population explanation ledger rejected")
	}
	for i := 0; i < 21; i++ {
		if check(i) {
			t.Fatalf("population explanation control accepted: %d", i)
		}
	}
}

type peLiveOwnership struct {
	Version            int
	Environment        string
	Records            []peLiveRecord
	Parent             exploration.Result
	HeldRun, Fresh     peLiveRun
	HeldFilter, Filter peLiveFilter
	Result             exploration.Result
	Events             []struct {
		Action               string
		SourceOps, PurePlans int
	}
}

func peLivePartition(parent exploration.Query, request exploration.ExplanationRequest, records []peLiveRecord) bool {
	if parent.Numeric == nil {
		return false
	}
	for _, q := range []exploration.Query{request.Baseline, request.Selection} {
		want := parent
		want.Offset = 0
		want.Where = q.Where
		if !nhLiveSameQuery(want, q) {
			return false
		}
	}
	for _, row := range records {
		p, e := peLiveMatch(parent.Where, row)
		if e != nil {
			return false
		}
		a, e := peLiveMatch(request.Baseline.Where, row)
		if e != nil {
			return false
		}
		z, e := peLiveMatch(request.Selection.Where, row)
		if e != nil {
			return false
		}
		raw, e := peLiveField(parent.Numeric.Field, row)
		if e != nil {
			return false
		}
		v, state := nhLiveInteger(raw, parent.Numeric.Type)
		selected := p && state == "valid" && v.Cmp(big.NewInt(8)) >= 0 && v.Cmp(big.NewInt(32)) < 0
		if z != selected || a != (p && !selected) || a && z {
			return false
		}
	}
	return true
}
func TestQuintPopulationExplanationOwnershipLedger(t *testing.T) {
	file := os.Getenv("GOO_POPULATION_EXPLANATION_OWNERSHIP_LEDGER")
	if file == "" {
		t.Skip("actual journey177 held explanation/filter ledger opt-in")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l peLiveOwnership
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		before, _ := json.Marshal(l)
		switch mutation {
		case 0:
			l.Events[3].SourceOps++
		case 1:
			l.Events[7].PurePlans++
		case 2:
			l.Fresh.Request.Baseline.Where = l.Fresh.Request.Selection.Where
		case 3:
			l.HeldRun.Result.Fields[0].Rows[0].DifferencePP = "1/17"
			l.HeldRun.Result.Basis = peLiveResultBasis(l.HeldRun.Result)
		case 4:
			l.Filter.Response.Question.Where = l.Parent.Question.Where
		case 5:
			l.Result.Count = "100"
		case 6:
			l.Filter.Response.Question.FromMS++
		case 7:
			l.HeldFilter.Response.Question.Environment += "-foreign"
		case 8:
			l.Fresh.Result.ResultToken = l.HeldRun.Result.ResultToken
		case 9:
			l.Events[9].SourceOps++
		case 10:
			l.Events[1].PurePlans++
		}
		after, _ := json.Marshal(l)
		if mutation >= 0 && string(before) == string(after) {
			t.Fatalf("ownership control %d made no change", mutation)
		}
		parent := l.Version == 1 && peLiveSource(l.Parent, l.Environment, l.Records)
		held := peLiveResult(l.HeldRun.Request, l.HeldRun.Result, l.Environment, l.Records) && peLivePartition(l.Parent.Question, l.HeldRun.Request, l.Records)
		fresh := peLiveResult(l.Fresh.Request, l.Fresh.Result, l.Environment, l.Records) && reflect.DeepEqual(l.Fresh.Request, l.HeldRun.Request) && peLiveHex(l.Fresh.Result.ResultToken) && l.Fresh.Result.ResultToken != l.HeldRun.Result.ResultToken && l.Fresh.Result.Basis != l.HeldRun.Result.Basis
		heldFilter := peLiveFilterScope(l.HeldFilter, l.Fresh.Request, l.Environment, l.Records)
		filter := peLiveFilterScope(l.Filter, l.Fresh.Request, l.Environment, l.Records) && reflect.DeepEqual(l.Filter.Request, l.HeldFilter.Request) && nhLiveSameQuery(l.Filter.Response.Question, l.HeldFilter.Response.Question)
		result := filter && peLiveSource(l.Result, l.Environment, l.Records) && nhLiveSameQuery(l.Result.Question, l.Filter.Response.Question)
		checks := map[string]bool{"parent": parent, "fields": parent, "run-held": held, "cancel": held, "return": held, "fresh-run": fresh, "filter-held": heldFilter, "filter-return": heldFilter, "filter-fresh": filter, "stage-side": filter, "run-side": result}
		if len(l.Events) != 11 {
			return false
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, event := range l.Events {
			ok := checks[event.Action]
			if !ok {
				t.Logf("explanation ownership rejected %s mutation %d", event.Action, mutation)
			}
			action := strings.ReplaceAll(event.Action, "-", "_")
			if action == "fields" {
				action = "edit_fields"
			}
			rec.Record(context.Background(), action, nil, "obs.coherent", ok, "obs.sources", event.SourceOps, "obs.plans", event.PurePlans)
		}
		checker, e := validate.NewChecker("model/population_explanation_ownership.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		resultTrace, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return resultTrace.OK()
	}
	if !check(-1) {
		t.Fatal("actual explanation ownership ledger rejected")
	}
	for i := 0; i < 11; i++ {
		if check(i) {
			t.Fatalf("explanation ownership mutation accepted: %d", i)
		}
	}
}
