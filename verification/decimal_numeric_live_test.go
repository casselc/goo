package verification

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"gopkg.in/yaml.v3"
)

type decLiveRecord struct {
	Environment, Signal, ID, TraceID, Service          string
	TimeMS                                             int64 `json:"timeMs"`
	Raw, ResourceRaw, Scale1Raw, Scale9Raw, IntegerRaw *string
}

func decLiveField(f exploration.Field, r decLiveRecord) (*string, error) {
	if f.Kind == "resource" {
		switch f.Key {
		case "declared.amount":
			return r.ResourceRaw, nil
		case "service.name":
			return &r.Service, nil
		case "deployment.environment.name":
			return &r.Environment, nil
		}
	}
	if f.Kind == "attribute" {
		switch f.Key {
		case "charge.amount":
			return r.Raw, nil
		case "amount.scale1":
			return r.Scale1Raw, nil
		case "amount.scale9":
			return r.Scale9Raw, nil
		case "work.items":
			return r.IntegerRaw, nil
		}
	}
	if f.Kind == "column" {
		switch f.Key {
		case "service":
			return &r.Service, nil
		case "span_id":
			return &r.ID, nil
		case "trace_id":
			return &r.TraceID, nil
		}
	}
	return nil, fmt.Errorf("unsupported fixture field %s:%s", f.Kind, f.Key)
}
func decLiveScalar(raw *string, b exploration.NumericBinding) (*big.Int, string) {
	if b.Type == "decimal64" {
		return decCoefficient(raw, b.Scale)
	}
	state := predicateState(raw, b.Type)
	if state != "valid" {
		return nil, state
	}
	n, _ := new(big.Int).SetString(*raw, 10)
	return n, state
}
func decLiveMatch(p *exploration.Predicate, r decLiveRecord) (bool, error) {
	if p == nil {
		return true, nil
	}
	if len(p.All) > 0 {
		for i := range p.All {
			yes, e := decLiveMatch(&p.All[i], r)
			if e != nil || !yes {
				return false, e
			}
		}
		return true, nil
	}
	if len(p.Any) > 0 {
		for i := range p.Any {
			yes, e := decLiveMatch(&p.Any[i], r)
			if e != nil {
				return false, e
			}
			if yes {
				return true, nil
			}
		}
		return false, nil
	}
	if p.Numeric != nil {
		raw, e := decLiveField(p.Numeric.Field, r)
		if e != nil {
			return false, e
		}
		n, state := decLiveScalar(raw, *p.Numeric)
		switch p.Op {
		case "valid", "missing", "empty", "invalid":
			return state == p.Op, nil
		}
		if state != "valid" {
			return false, nil
		}
		z, zstate := decLiveScalar(&p.Value, *p.Numeric)
		if zstate != "valid" {
			return false, fmt.Errorf("noncanonical human threshold")
		}
		c := n.Cmp(z)
		switch p.Op {
		case "eq":
			return c == 0, nil
		case "ne":
			return c != 0, nil
		case "lt":
			return c < 0, nil
		case "lte":
			return c <= 0, nil
		case "gt":
			return c > 0, nil
		case "gte":
			return c >= 0, nil
		}
		return false, fmt.Errorf("unknown numeric operator")
	}
	if p.Field == nil {
		return false, fmt.Errorf("missing field")
	}
	raw, e := decLiveField(*p.Field, r)
	if e != nil {
		return false, e
	}
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
	}
	return false, fmt.Errorf("unknown string operator")
}
func decLiveRows(q exploration.Query, environment string, records []decLiveRecord) ([]decLiveRecord, error) {
	if q.Environment != environment || q.Signal != "spans" && q.Signal != "logs" || q.ToMS <= q.FromMS {
		return nil, fmt.Errorf("invalid source scope")
	}
	rows := []decLiveRecord{}
	for _, r := range records {
		if r.Environment != environment || r.Signal != q.Signal || r.TimeMS < q.FromMS || r.TimeMS >= q.ToMS {
			continue
		}
		yes, e := decLiveMatch(q.Where, r)
		if e != nil {
			return nil, e
		}
		if yes {
			rows = append(rows, r)
		}
	}
	return rows, nil
}
func decLiveSummary(b exploration.NumericBinding, rows []decLiveRecord) exploration.NumericSummary {
	values := []*string{}
	for _, r := range rows {
		raw, e := decLiveField(b.Field, r)
		if e != nil {
			panic(e)
		}
		values = append(values, raw)
	}
	if b.Type == "decimal64" {
		return decSummary(b, values)
	}
	return comparisonOracleSummary(b, values)
}
func decLiveHistogram(b exploration.NumericBinding, rows []decLiveRecord) exploration.NumericHistogram {
	values := []*string{}
	for _, r := range rows {
		raw, e := decLiveField(b.Field, r)
		if e != nil {
			panic(e)
		}
		values = append(values, raw)
	}
	if b.Type == "decimal64" {
		return decHistogram(b, values)
	}
	edges := histogramOracleEdges(b.Type)
	counts := histogramOracleCounts(b, values)
	bins := []exploration.NumericHistogramBin{}
	for i, c := range counts {
		bins = append(bins, exploration.NumericHistogramBin{From: edges[i], To: edges[i+1], Count: c})
	}
	return exploration.NumericHistogram{Binning: "powers-of-two", Bins: bins}
}
func decLiveSource(result exploration.Result, environment string, records []decLiveRecord) bool {
	q := result.Question
	if q.Numeric == nil {
		return false
	}
	rows, e := decLiveRows(q, environment, records)
	if e != nil {
		return false
	}
	expected := decLiveSummary(*q.Numeric, rows)
	if result.Numeric == nil || !reflect.DeepEqual(*result.Numeric, expected) {
		return false
	}
	if q.NumericHistogram != nil {
		if result.NumericHistogram == nil || !reflect.DeepEqual(*result.NumericHistogram, decLiveHistogram(*q.Numeric, rows)) {
			return false
		}
	} else if result.NumericHistogram != nil {
		return false
	}
	plain := []peLiveRecord{}
	byID := map[string]decLiveRecord{}
	for _, r := range rows {
		attrs := map[string]string{}
		for _, f := range []exploration.Field{{Kind: "attribute", Key: "charge.amount"}, {Kind: "attribute", Key: "amount.scale1"}, {Kind: "attribute", Key: "amount.scale9"}, {Kind: "attribute", Key: "work.items"}} {
			raw, _ := decLiveField(f, r)
			if raw != nil {
				attrs[f.Key] = *raw
			}
		}
		resource := map[string]string{}
		if r.ResourceRaw != nil {
			resource["declared.amount"] = *r.ResourceRaw
		}
		plain = append(plain, peLiveRecord{Environment: r.Environment, Signal: r.Signal, ID: r.ID, Service: r.Service, Status: "Ok", Severity: "INFO", TimeMS: r.TimeMS, Attributes: attrs, Resource: resource})
		byID[r.ID] = r
	}
	page := result
	page.Question.Where = nil
	page.Question.Numeric = nil
	page.Question.NumericHistogram = nil
	page.Rows = make([]json.RawMessage, len(result.Rows))
	for i, raw := range result.Rows {
		var record map[string]json.RawMessage
		var id, trace string
		if json.Unmarshal(raw, &record) != nil || json.Unmarshal(record["SpanId"], &id) != nil || json.Unmarshal(record["TraceId"], &trace) != nil {
			return false
		}
		fact, ok := byID[id]
		var resources map[string]string
		expectedResources := map[string]string{"service.name": fact.Service, "deployment.environment.name": fact.Environment}
		if fact.ResourceRaw != nil {
			expectedResources["declared.amount"] = *fact.ResourceRaw
		}
		if !ok || trace != fact.TraceID || json.Unmarshal(record["ResourceAttributes"], &resources) != nil || !reflect.DeepEqual(resources, expectedResources) {
			return false
		}
		if q.Signal == "logs" {
			var body string
			if json.Unmarshal(record["Body"], &body) != nil || body != "declared amount "+id {
				return false
			}
			record["Body"], _ = json.Marshal("work observed " + id)
		}
		page.Rows[i], _ = json.Marshal(record)
	}
	if !peLiveSource(page, environment, plain) {
		return false
	}
	expectedGroups := map[string][]decLiveRecord{}
	if len(q.GroupBy) > 0 {
		for _, r := range rows {
			values := []*string{}
			for _, f := range q.GroupBy {
				value, e := decLiveField(f, r)
				if e != nil {
					return false
				}
				values = append(values, value)
			}
			key, _ := json.Marshal(values)
			expectedGroups[string(key)] = append(expectedGroups[string(key)], r)
		}
	}
	if len(result.Groups) != len(expectedGroups) {
		return false
	}
	seen := map[string]bool{}
	for _, raw := range result.Groups {
		var group map[string]json.RawMessage
		if json.Unmarshal(raw, &group) != nil {
			return false
		}
		values := []*string{}
		for i := range q.GroupBy {
			var value *string
			data, ok := group[fmt.Sprintf("g%d", i)]
			if !ok || json.Unmarshal(data, &value) != nil {
				return false
			}
			values = append(values, value)
		}
		key, _ := json.Marshal(values)
		members, ok := expectedGroups[string(key)]
		if !ok || seen[string(key)] {
			return false
		}
		seen[string(key)] = true
		var count string
		var numeric exploration.NumericSummary
		if json.Unmarshal(group["count"], &count) != nil || count != strconv.Itoa(len(members)) || json.Unmarshal(group["numeric"], &numeric) != nil || !reflect.DeepEqual(numeric, decLiveSummary(*q.Numeric, members)) {
			return false
		}
	}
	if q.BucketMS > 0 {
		n := (q.ToMS - q.FromMS + q.BucketMS - 1) / q.BucketMS
		if int64(len(result.TimeSeries)) != n {
			return false
		}
		for i, p := range result.TimeSeries {
			from := q.FromMS + int64(i)*q.BucketMS
			to := from + q.BucketMS
			if to > q.ToMS {
				to = q.ToMS
			}
			members := []decLiveRecord{}
			for _, r := range rows {
				if r.TimeMS >= from && r.TimeMS < to {
					members = append(members, r)
				}
			}
			state := "observed"
			if len(members) == 0 {
				state = "no_data"
			}
			if p.FromMS != from || p.ToMS != to || p.Count != strconv.Itoa(len(members)) || p.State != state || p.Numeric == nil || !reflect.DeepEqual(*p.Numeric, decLiveSummary(*q.Numeric, members)) {
				return false
			}
		}
	} else if len(result.TimeSeries) > 0 {
		return false
	}
	return result.Duration == nil && result.SpanStatus == nil && len(result.GroupedTimeSeries) == 0
}

func decLiveCohort(result exploration.Result) (exploration.Cohort, error) {
	c := exploration.Cohort{Question: result.Question, Basis: result.Basis, State: result.State, Count: result.Count, Numeric: result.Numeric, NumericHistogram: result.NumericHistogram, Groups: []exploration.ValueCount{}}
	if len(result.Question.GroupBy) != 1 {
		return c, fmt.Errorf("one grouped field required")
	}
	for _, raw := range result.Groups {
		var g exploration.ValueCount
		if e := json.Unmarshal(raw, &g); e != nil {
			return c, e
		}
		c.Groups = append(c.Groups, g)
	}
	return c, nil
}
func decLivePair(request exploration.ComparisonRequest, response exploration.CohortComparison, baseline, selection exploration.Result, environment string, records []decLiveRecord) bool {
	if !request.Numeric || request.Distribution || !decLiveSource(baseline, environment, records) || !decLiveSource(selection, environment, records) || baseline.Question.Numeric == nil || selection.Question.Numeric == nil || *baseline.Question.Numeric != *selection.Question.Numeric {
		return false
	}
	a, e := decLiveCohort(baseline)
	if e != nil {
		return false
	}
	z, e := decLiveCohort(selection)
	if e != nil {
		return false
	}
	if !reflect.DeepEqual(a, request.Baseline) || !reflect.DeepEqual(z, request.Selection) || response.State != "ok" || response.Field != a.Question.GroupBy[0] || response.BaselineCount != a.Count || response.SelectionCount != z.Count || response.BaselineBasis != a.Basis || response.SelectionBasis != z.Basis || response.Measures == nil || response.Measures.Numeric == nil || response.Measures.Numeric.Distribution != nil {
		return false
	}
	b := *a.Question.Numeric
	if checkNumericChanges(response.Measures.Numeric, b, a.Numeric, z.Numeric, false) != nil {
		return false
	}
	canonical := request
	canonical.Baseline.Groups = append([]exploration.ValueCount{}, a.Groups...)
	canonical.Selection.Groups = append([]exploration.ValueCount{}, z.Groups...)
	for _, groups := range [][]exploration.ValueCount{canonical.Baseline.Groups, canonical.Selection.Groups} {
		sortNumericGroups(groups)
	}
	encoded, _ := json.Marshal(canonical)
	hash := sha256.Sum256(encoded)
	if response.Basis != hex.EncodeToString(hash[:]) {
		return false
	}
	ag, zg := map[string]exploration.ValueCount{}, map[string]exploration.ValueCount{}
	keys := map[string]bool{}
	for _, g := range a.Groups {
		k := explanationValueKey(g.Value)
		ag[k] = g
		keys[k] = true
	}
	for _, g := range z.Groups {
		k := explanationValueKey(g.Value)
		zg[k] = g
		keys[k] = true
	}
	if len(response.Rows) != len(keys) {
		return false
	}
	seen := map[string]bool{}
	previousScore := new(big.Rat).SetInt64(1000)
	previousKey := ""
	for _, row := range response.Rows {
		k := explanationValueKey(row.Value)
		if !keys[k] || seen[k] {
			return false
		}
		seen[k] = true
		ac, zc := "0", "0"
		as, zs := ag[k], zg[k]
		if as.Count != "" {
			ac = as.Count
		}
		if zs.Count != "" {
			zc = zs.Count
		}
		ap := new(big.Rat).Mul(new(big.Rat).SetFrac(ncLiveInt(ac), ncLiveInt(a.Count)), big.NewRat(100, 1))
		zp := new(big.Rat).Mul(new(big.Rat).SetFrac(ncLiveInt(zc), ncLiveInt(z.Count)), big.NewRat(100, 1))
		delta := new(big.Rat).Sub(zp, ap)
		display := delta.FloatString(2)
		if display == "-0.00" {
			display = "0.00"
		}
		score := new(big.Rat).Abs(delta)
		if row.BaselineCount != ac || row.SelectionCount != zc || row.BaselinePercent != ap.FloatString(2) || row.SelectionPercent != zp.FloatString(2) || row.DifferencePP != display || row.Measures == nil || row.Measures.Numeric == nil || checkNumericChanges(row.Measures.Numeric, b, as.Numeric, zs.Numeric, false) != nil {
			return false
		}
		if previousScore.Cmp(score) < 0 || previousScore.Cmp(score) == 0 && previousKey > k {
			return false
		}
		previousScore = score
		previousKey = k
	}
	return true
}
func sortNumericGroups(groups []exploration.ValueCount) {
	sort.Slice(groups, func(i, j int) bool {
		return explanationValueKey(groups[i].Value) < explanationValueKey(groups[j].Value)
	})
}

type decLivePlan struct {
	Request struct {
		Parent         exploration.Query
		FromBin, ToBin int
	}
	Response exploration.NumericHistogramSelection
}
type decLiveCustody struct {
	File                                string
	Bytes                               int
	SHA256, ReceiptID, CopiedAt, Source string
}
type decLiveLedger struct {
	Custody                                                                                        decLiveCustody
	StoredBytes                                                                                    int
	StoredSHA                                                                                      string
	Version                                                                                        int
	Environment, Signal                                                                            string
	Records                                                                                        []decLiveRecord
	Q, Initial, ParentQuery                                                                        exploration.Query
	Binding                                                                                        exploration.NumericBinding
	Filtered, Negative, ResourceFilter, Result, Selected, Rest, ParentAgain, Reused, LastBinResult exploration.Result
	Coverage                                                                                       []struct {
		Operator string
		Result   exploration.Result
	}
	Selection decLivePlan
	Partition struct {
		Plan           exploration.NumericHistogramSelection
		Selected, Rest exploration.Result
		Request        exploration.ComparisonRequest
		Response       exploration.CohortComparison
	}
	Signed [][]struct{ Value, Exact, Text string }
	View   struct {
		Points             []struct{ Key, From, To, Value, Label string }
		FromLabel, ToLabel string
		Interactive        bool
	}
	Saved, Reopened exploration.CapturePage
	Stored          exploration.SavedCapture
	Draft           exploration.PackagePlan
	Reusable        exploration.SavedPackage
	EdgeResults     []struct {
		Kind   string
		Result exploration.Result
	}
	LastBinPlan exploration.NumericHistogramSelection
	Refusals    []struct {
		Kind     string
		Status   int
		Response json.RawMessage
	}
	Events []struct {
		Action                          string
		Reads, Plans, Protocols, Writes int
	}
	Requests []struct {
		Effect, Endpoint string
		Body             json.RawMessage
	}
}

func decLivePlanCheck(plan exploration.NumericHistogramSelection, parent exploration.Query, from, to int, records []decLiveRecord) bool {
	edges := histogramOracleEdges("int64")
	if parent.Numeric == nil || plan.FromBin != from || plan.ToBin != to || plan.From != edges[from] || plan.To != edges[to] || !nhLiveSameQuery(plan.Parent, parent) {
		return false
	}
	lo, hi := ncLiveInt(edges[from]), ncLiveInt(edges[to])
	for _, r := range records {
		parentMatch, e := decLiveMatch(parent.Where, r)
		if e != nil {
			return false
		}
		raw, e := decLiveField(parent.Numeric.Field, r)
		if e != nil {
			return false
		}
		n, state := decLiveScalar(raw, *parent.Numeric)
		inside := state == "valid" && n.Cmp(lo) >= 0 && n.Cmp(hi) < 0
		selected, e := decLiveMatch(plan.Selected.Where, r)
		if e != nil {
			return false
		}
		rest, e := decLiveMatch(plan.Rest.Where, r)
		if e != nil || selected != (parentMatch && inside) || rest != (parentMatch && !inside) || selected && rest {
			return false
		}
	}
	for _, child := range []exploration.Query{plan.Selected, plan.Rest} {
		expected := parent
		expected.Where = child.Where
		expected.Offset = 0
		if !nhLiveSameQuery(child, expected) {
			return false
		}
	}
	for _, pair := range [][2]exploration.Query{{plan.NumericComparison.Selection, plan.Selected}, {plan.NumericComparison.Baseline, plan.Rest}} {
		expected := pair[1]
		expected.NumericHistogram = nil
		expected.Summary = ""
		expected.SpanStatus = false
		if !nhLiveSameQuery(pair[0], expected) {
			return false
		}
	}
	var walk func(*exploration.Predicate) bool
	walk = func(p *exploration.Predicate) bool {
		if p == nil {
			return true
		}
		if p.Numeric != nil {
			if *p.Numeric != *parent.Numeric {
				return false
			}
			switch p.Op {
			case "eq", "ne", "lt", "lte", "gt", "gte":
				_, state := decCoefficient(&p.Value, parent.Numeric.Scale)
				if state != "valid" {
					return false
				}
			}
		}
		for i := range p.All {
			if !walk(&p.All[i]) {
				return false
			}
		}
		for i := range p.Any {
			if !walk(&p.Any[i]) {
				return false
			}
		}
		return true
	}
	return walk(plan.Selected.Where) && walk(plan.Rest.Where)
}
func decLiveStored(l decLiveLedger) bool {
	s, p := l.Stored, l.Saved
	if s.Explore == nil || s.Question.Explore == nil || s.Question.Kind != "explore" || p.Explore == nil || !reflect.DeepEqual(s.CaptureReceipt, p.Receipt) || !nhLiveSameQuery(*s.Question.Explore, l.ParentQuery) || !peLiveCaptureHash(exploration.CapturePage{Receipt: s.CaptureReceipt, Explore: s.Explore}) || !aeJSONEqual(p, l.Reopened) {
		return false
	}
	rows, e := decLiveRows(l.ParentQuery, l.Environment, l.Records)
	if e != nil || len(rows) > 1000 || s.Explore.HasMore || len(s.Explore.Rows) != len(rows) || p.RetainedRecords != len(rows) || p.RecordsLimited {
		return false
	}
	n := len(rows)
	if n > 100 {
		n = 100
	}
	if !aeJSONEqual(s.Explore.Rows[:n], p.Explore.Rows) || p.Explore.HasMore != (len(rows) > n) {
		return false
	}
	retained := *s.Explore
	retained.Question.Limit = len(rows)
	return decLiveSource(retained, l.Environment, l.Records) && decLiveSource(*p.Explore, l.Environment, l.Records)
}

func decLiveSemantic(l decLiveLedger) bool {
	if !decLiveCustodyCheck(l) || l.Version != 1 || l.Binding != decBinding(2) || l.Q.Signal != l.Signal || l.Q.Environment != l.Environment {
		return false
	}
	parent := l.Q
	parent.Numeric = &l.Binding
	parent.NumericHistogram = &exploration.NumericHistogramSpec{Binning: "powers-of-two"}
	if !nhLiveSameQuery(parent, l.ParentQuery) || !nhLiveSameQuery(l.Result.Question, parent) || !nhLiveSameQuery(l.ParentAgain.Question, parent) {
		return false
	}
	initial := l.Q
	initial.Where = &exploration.Predicate{Field: &l.Binding.Field, Op: "exists"}
	if !nhLiveSameQuery(initial, l.Initial) {
		return false
	}
	filter := parent
	filter.Where = &exploration.Predicate{Numeric: &l.Binding, Op: "gte", Value: "1.20"}
	if !nhLiveSameQuery(filter, l.Filtered.Question) {
		return false
	}
	filter.Where = &exploration.Predicate{Numeric: &l.Binding, Op: "lt", Value: "-0.01"}
	if !nhLiveSameQuery(filter, l.Negative.Question) {
		return false
	}
	resource := l.Binding
	resource.Field = exploration.Field{Kind: "resource", Key: "declared.amount"}
	filter.Where = &exploration.Predicate{Numeric: &resource, Op: "gte", Value: "1.20"}
	if !nhLiveSameQuery(filter, l.ResourceFilter.Question) {
		return false
	}
	if len(l.Coverage) != 4 {
		return false
	}
	for i, kind := range []string{"valid", "missing", "empty", "invalid"} {
		item := l.Coverage[i]
		q := parent
		q.Where = &exploration.Predicate{Numeric: &l.Binding, Op: kind}
		if item.Operator != kind || !nhLiveSameQuery(item.Result.Question, q) || !decLiveSource(item.Result, l.Environment, l.Records) {
			return false
		}
	}
	for _, result := range []exploration.Result{l.Filtered, l.Negative, l.ResourceFilter, l.Result, l.Selected, l.Rest, l.Partition.Selected, l.Partition.Rest, l.ParentAgain, l.Reused, l.LastBinResult} {
		if !decLiveSource(result, l.Environment, l.Records) {
			return false
		}
	}
	if l.Selection.Request.FromBin != 72 || l.Selection.Request.ToBin != 74 || !nhLiveSameQuery(l.Selection.Request.Parent, parent) || !decLivePlanCheck(l.Selection.Response, parent, 72, 74, l.Records) || !decLivePlanCheck(l.Partition.Plan, parent, 72, 74, l.Records) || !aeJSONEqual(l.Selection.Response, l.Partition.Plan) || !nhLiveSameQuery(l.Selected.Question, l.Selection.Response.Selected) || !nhLiveSameQuery(l.Rest.Question, l.Selection.Response.Rest) || !nhLiveSameQuery(l.Partition.Selected.Question, l.Partition.Plan.NumericComparison.Selection) || !nhLiveSameQuery(l.Partition.Rest.Question, l.Partition.Plan.NumericComparison.Baseline) {
		return false
	}
	if !decLivePair(l.Partition.Request, l.Partition.Response, l.Partition.Rest, l.Partition.Selected, l.Environment, l.Records) || !decLiveStored(l) {
		return false
	}
	if len(l.Signed) != 4 {
		return false
	}
	for i, metric := range []string{"mean", "sum", "min", "max"} {
		if len(l.Signed[i]) != 2 {
			return false
		}
		for side, s := range []*exploration.NumericSummary{l.Partition.Rest.Numeric, l.Partition.Selected.Numeric} {
			if s == nil {
				return false
			}
			numerator := s.Sum
			switch metric {
			case "min":
				numerator = s.Min
			case "max":
				numerator = s.Max
			}
			denominator := big.NewInt(100)
			if metric == "mean" {
				denominator.Mul(denominator, ncLiveInt(s.Valid))
			}
			expected := new(big.Rat).SetFrac(ncLiveInt(numerator), denominator)
			cell := l.Signed[i][side]
			actual, ok := new(big.Rat).SetString(cell.Value)
			if !ok || actual.Cmp(expected) != 0 || cell.Exact != cell.Value+" USD (exact)" || cell.Text == "" {
				return false
			}
		}
	}
	if l.Result.NumericHistogram == nil || len(l.View.Points) != 128 || !l.View.Interactive {
		return false
	}
	for i, p := range l.View.Points {
		b := l.Result.NumericHistogram.Bins[i]
		label := "[" + decText(b.From, 2) + ", " + decText(b.To, 2) + ") USD · " + b.Count + " records"
		if p.Key != strconv.Itoa(i) || p.From != b.From || p.To != b.To || p.Value != b.Count || p.Label != label {
			return false
		}
	}
	if !decLivePackage(l) {
		return false
	}
	if len(l.EdgeResults) != 4 {
		return false
	}
	for i, kind := range []string{"scale1", "scale9", "resource", "integer"} {
		item := l.EdgeResults[i]
		b := l.Binding
		switch kind {
		case "scale1":
			b.Scale = 1
			b.Field.Key = "amount.scale1"
		case "scale9":
			b.Scale = 9
			b.Field.Key = "amount.scale9"
		case "resource":
			b = resource
		case "integer":
			b = exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work.items"}, Type: "int64", Unit: "items", Parse: "canonical-decimal"}
		}
		q := parent
		q.Numeric = &b
		if item.Kind != kind || !nhLiveSameQuery(item.Result.Question, q) || !decLiveSource(item.Result, l.Environment, l.Records) {
			return false
		}
		if kind == "integer" {
			encoded, _ := json.Marshal(item.Result.Numeric)
			var fields map[string]json.RawMessage
			json.Unmarshal(encoded, &fields)
			if _, found := fields["scale"]; found {
				return false
			}
		}
	}
	if !decLivePlanCheck(l.LastBinPlan, parent, 127, 128, l.Records) || !nhLiveSameQuery(l.LastBinResult.Question, l.LastBinPlan.Selected) {
		return false
	}
	if len(l.Refusals) != 5 {
		return false
	}
	for i, kind := range []string{"missing-scale", "scale-zero", "scale-ten", "integer-scale", "scale-mismatch"} {
		if l.Refusals[i].Kind != kind || l.Refusals[i].Status != 400 {
			return false
		}
	}
	return true
}
func decLivePackage(l decLiveLedger) bool {
	p := l.Reusable.Definition
	if p.Numeric == nil || *p.Numeric != l.Binding || p.NumericHistogram == nil || p.NumericHistogram.Binning != "powers-of-two" || p.Source.Signal != l.Signal || p.Source.Environment != l.Environment || p.Source.Adapter != "clickhouse.otel/v1" || p.Measure != "captured-record-count" || p.LookbackMinutes != 5 || !reflect.DeepEqual(p.Where, l.ParentQuery.Where) || !reflect.DeepEqual(p.GroupBy, l.ParentQuery.GroupBy) || p.BucketMS != l.ParentQuery.BucketMS || !reflect.DeepEqual(p, l.Draft.Definition) || l.Draft.Revision != l.Reusable.Revision {
		return false
	}
	var decoded exploration.TelemetryPackage
	if yaml.Unmarshal([]byte(l.Draft.YAML), &decoded) != nil || !reflect.DeepEqual(decoded, p) || l.Draft.Question.Numeric == nil || *l.Draft.Question.Numeric != l.Binding {
		return false
	}
	raw, _ := json.Marshal(p)
	hash := sha256.Sum256(raw)
	if hex.EncodeToString(hash[:]) != l.Reusable.Revision {
		return false
	}
	reused := l.ParentQuery
	reused.FromMS = l.Reused.Question.FromMS
	reused.ToMS = l.Reused.Question.ToMS
	return nhLiveSameQuery(reused, l.Reused.Question) && reused.ToMS-reused.FromMS == 300000 && reused.ToMS >= l.ParentQuery.ToMS && reflect.DeepEqual(l.Reused.Numeric, l.Result.Numeric) && reflect.DeepEqual(l.Reused.NumericHistogram, l.Result.NumericHistogram)
}
func decLiveRequests(l decLiveLedger) bool {
	expected := []exploration.Result{l.Filtered, l.Negative, l.ResourceFilter}
	for _, item := range l.Coverage {
		expected = append(expected, item.Result)
	}
	expected = append(expected, l.Result, l.Selected)
	explores, plans, reads, writes, protocols := 0, 0, 0, 0, 0
	captureSeen := false
	comparisonResults := []exploration.Result{l.Partition.Selected, l.Partition.Rest}
	comparisonSeen := map[string]bool{}
	for _, r := range l.Requests {
		switch r.Effect {
		case "source":
			reads++
			if r.Endpoint == "/api/captures" {
				var in struct {
					Title    string
					Question exploration.CaptureQuestion
				}
				if json.Unmarshal(r.Body, &in) != nil || in.Title != l.Saved.Receipt.Title || in.Question.Explore == nil || !nhLiveSameQuery(*in.Question.Explore, l.ParentQuery) || captureSeen {
					return false
				}
				captureSeen = true
				continue
			}
			if r.Endpoint != "/api/explore" {
				return false
			}
			var q exploration.Query
			if json.Unmarshal(r.Body, &q) != nil {
				return false
			}
			if explores < len(expected) {
				if !nhLiveSameQuery(q, expected[explores].Question) {
					return false
				}
			} else if explores < len(expected)+2 {
				found := false
				for _, result := range comparisonResults {
					if nhLiveSameQuery(q, result.Question) {
						encoded, _ := json.Marshal(q)
						key := string(encoded)
						if comparisonSeen[key] {
							return false
						}
						comparisonSeen[key] = true
						found = true
						break
					}
				}
				if !found {
					return false
				}
			} else {
				tail := []exploration.Result{l.ParentAgain, l.Reused}
				i := explores - len(expected) - 2
				if i >= len(tail) || !nhLiveSameQuery(q, tail[i].Question) {
					return false
				}
			}
			explores++
		case "plan":
			if r.Endpoint != "/api/numeric-histograms/selection" {
				return false
			}
			var in struct {
				Parent         exploration.Query
				FromBin, ToBin int
			}
			if json.Unmarshal(r.Body, &in) != nil || !nhLiveSameQuery(in.Parent, l.ParentQuery) || in.FromBin != 72 || in.ToBin != 74 {
				return false
			}
			plans++
		case "comparison":
			var in exploration.ComparisonRequest
			if r.Endpoint != "/api/compare" || json.Unmarshal(r.Body, &in) != nil || !reflect.DeepEqual(in, l.Partition.Request) {
				return false
			}
			protocols++
		case "publication":
			var in struct{ YAML string }
			if r.Endpoint != "/api/telemetry-packages" || json.Unmarshal(r.Body, &in) != nil || in.YAML != l.Draft.YAML {
				return false
			}
			writes++
		default:
			return false
		}
	}
	return reads == 14 && explores == 13 && plans == 2 && protocols == 1 && writes == 1 && captureSeen
}
func decLiveReplay(t *testing.T, l decLiveLedger) bool {
	t.Helper()
	if len(l.Events) != 21 {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for _, e := range l.Events {
		rec.Record(context.Background(), strings.ReplaceAll(e.Action, "-", "_"), nil, "obs.coherent", true, "obs.reads", e.Reads, "obs.plans", e.Plans, "obs.protocols", e.Protocols, "obs.writes", e.Writes)
	}
	checker, e := validate.NewChecker("model/numeric_authoring.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	result, e := checker.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return result.OK()
}
func TestQuintDecimalNumericLiveLedger(t *testing.T) {
	file := os.Getenv("GOO_DECIMAL_NUMERIC_LEDGER")
	if file == "" {
		t.Skip("requires actual decimal numeric journey source ledger")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	stored, e := os.ReadFile(filepath.Join(filepath.Dir(file), "scope-capture.json"))
	if e != nil {
		t.Fatal(e)
	}
	custodyRaw, e := os.ReadFile(filepath.Join(filepath.Dir(file), "capture-custody.json"))
	if e != nil {
		t.Fatal(e)
	}
	digest := sha256.Sum256(stored)
	check := func(control int) bool {
		var l decLiveLedger
		if e = json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		if e = json.Unmarshal(stored, &l.Stored); e != nil {
			t.Fatal(e)
		}
		if e = json.Unmarshal(custodyRaw, &l.Custody); e != nil {
			t.Fatal(e)
		}
		l.StoredBytes = len(stored)
		l.StoredSHA = hex.EncodeToString(digest[:])
		before, _ := json.Marshal(l)
		switch control {
		case 0:
			l.Result.Rows[1] = append(json.RawMessage{}, l.Result.Rows[0]...)
		case 1:
			var row map[string]any
			json.Unmarshal(l.Result.Rows[0], &row)
			row["Attributes"].(map[string]any)["charge.amount"] = "0.00"
			l.Result.Rows[0], _ = json.Marshal(row)
		case 2:
			var row map[string]any
			json.Unmarshal(l.Result.Rows[0], &row)
			row["ResourceAttributes"].(map[string]any)["declared.amount"] = "0.00"
			l.Result.Rows[0], _ = json.Marshal(row)
		case 3:
			l.Result.Numeric.Scale = 3
		case 4:
			l.Result.Numeric.Sum += "1"
		case 5:
			l.Result.NumericHistogram.Bins[72].From = "129"
		case 6:
			l.Result.NumericHistogram.Bins[72].Count += "1"
		case 7:
			l.Selection.Response.Selected.Where = &exploration.Predicate{Numeric: &l.Binding, Op: "gte", Value: "128"}
		case 8:
			l.Selection.Response.Rest.Where = nil
		case 9:
			l.Partition.Selected.Numeric.Valid += "1"
		case 10:
			l.Partition.Response.Measures.Numeric.MeanDifference += "1"
		case 11:
			l.Signed[0][1].Value = "983/3"
		case 12:
			l.Signed[1][0].Exact = "rounded USD"
		case 13:
			l.Stored.ID = "changed"
		case 14:
			l.Stored.Explore.Rows[0] = json.RawMessage(`{"SpanId":"changed"}`)
		case 15:
			l.Reopened.Receipt.CapturedAt = "changed"
		case 16:
			l.Reused.Question.Numeric.Scale = 3
		case 17:
			l.Reusable.Definition.Numeric.Scale = 3
		case 18:
			l.EdgeResults[0].Result.Numeric.Scale = 9
		case 19:
			l.LastBinPlan.Selected.Where = &exploration.Predicate{Numeric: &l.Binding, Op: "lt", Value: decText("9223372036854775808", 2)}
		case 20:
			l.Result.TimeSeries[0].Numeric.Invalid += "1"
		case 21:
			l.Coverage[1].Result.Question.Where = &exploration.Predicate{Numeric: &l.Binding, Op: "empty"}
		case 22:
			l.Events[1].Reads++
		case 23:
			l.Events[3].Reads++
		case 24:
			l.Events[12].Reads++
		case 25:
			l.Events[15].Reads++
		case 26:
			l.Events[18].Reads++
		case 27:
			l.Events[20].Writes++
		case 28:
			l.Events[0], l.Events[1] = l.Events[1], l.Events[0]
		case 29:
			l.Requests[0].Endpoint = "/api/captures"
		case 30:
			l.Filtered.Question.Where.Value = "1.200"
		case 31:
			l.ResourceFilter.Question.Where.Numeric.Field.Kind = "attribute"
		case 32:
			l.Result.Question.ToMS++
		case 33:
			l.View.Points[72].Label = "[128, 256) USD · 10 records"
		case 34:
			l.Refusals[0].Status = 200
		case 35:
			l.Custody.SHA256 = "changed"
		case 36:
			l.Custody.ReceiptID = "changed"
		case 37:
			l.Custody.CopiedAt = "invalid"
		}
		if control >= 0 {
			after, _ := json.Marshal(l)
			if string(after) == string(before) {
				t.Fatalf("decimal control %d did not change evidence", control)
			}
		}
		semantic, requests := decLiveSemantic(l), decLiveRequests(l)
		if control < 0 {
			t.Logf("independent decimal source/scope=%v requests=%v result=%v capture=%v comparison=%v package=%v", semantic, requests, decLiveSource(l.Result, l.Environment, l.Records), decLiveStored(l), decLivePair(l.Partition.Request, l.Partition.Response, l.Partition.Rest, l.Partition.Selected, l.Environment, l.Records), decLivePackage(l))
		}
		return semantic && requests && decLiveReplay(t, l)
	}
	if !check(-1) {
		t.Fatal("actual decimal numeric source/lifecycle ledger rejected")
	}
	for i := 0; i < 38; i++ {
		if check(i) {
			t.Fatalf("decimal control %d accepted", i)
		}
		t.Logf("decimal control %d rejected", i)
	}
}

func decLiveCustodyCheck(l decLiveLedger) bool {
	c := l.Custody
	if c.File != "scope-capture.json" || c.Bytes != l.StoredBytes || c.Bytes <= 0 || c.SHA256 != l.StoredSHA || c.ReceiptID != l.Stored.ID || !strings.HasPrefix(c.Source, "read-only docker cp ") {
		return false
	}
	copied, e := time.Parse(time.RFC3339Nano, c.CopiedAt)
	if e != nil || l.Stored.Explore == nil {
		return false
	}
	for _, clock := range []string{l.Stored.CapturedAt, l.Stored.Explore.CapturedAt} {
		read, e := time.Parse(time.RFC3339Nano, clock)
		if e != nil || copied.Before(read) {
			return false
		}
	}
	return true
}
