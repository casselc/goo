package verification

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// Contextual plots use identical map keys in two distinct qualified namespaces.
// Reuse the independently checked coefficient arithmetic, not production evaluators.
func cpLiveField(f exploration.Field, r decLiveRecord) (*string, error) {
	if f.Kind == "resource" {
		switch f.Key {
		case "work.amount":
			return r.ResourceRaw, nil
		case "service.name":
			return &r.Service, nil
		case "deployment.environment.name":
			return &r.Environment, nil
		}
	}
	if f.Kind == "attribute" {
		if f.Key == "team" {
			v := "core"
			return &v, nil
		}
		switch f.Key {
		case "work.amount":
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
func cpLiveMatch(p *exploration.Predicate, r decLiveRecord) (bool, error) {
	if p == nil {
		return true, nil
	}
	if len(p.All) > 0 {
		for i := range p.All {
			yes, e := cpLiveMatch(&p.All[i], r)
			if e != nil || !yes {
				return false, e
			}
		}
		return true, nil
	}
	if len(p.Any) > 0 {
		for i := range p.Any {
			yes, e := cpLiveMatch(&p.Any[i], r)
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
		raw, e := cpLiveField(p.Numeric.Field, r)
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
	raw, e := cpLiveField(*p.Field, r)
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
func cpLiveRows(q exploration.Query, environment string, records []decLiveRecord) ([]decLiveRecord, error) {
	if q.Environment != environment || q.Signal != "spans" && q.Signal != "logs" || q.ToMS <= q.FromMS {
		return nil, fmt.Errorf("invalid source scope")
	}
	rows := []decLiveRecord{}
	for _, r := range records {
		if r.Environment != environment || r.Signal != q.Signal || r.TimeMS < q.FromMS || r.TimeMS >= q.ToMS {
			continue
		}
		yes, e := cpLiveMatch(q.Where, r)
		if e != nil {
			return nil, e
		}
		if yes {
			rows = append(rows, r)
		}
	}
	return rows, nil
}
func cpLiveSource(result exploration.Result, environment string, records []decLiveRecord) bool {
	q := result.Question
	if q.Numeric == nil {
		return false
	}
	rows, e := cpLiveRows(q, environment, records)
	if e != nil {
		return false
	}
	expected := cpLiveSummary(*q.Numeric, rows)
	if result.Numeric == nil || !reflect.DeepEqual(*result.Numeric, expected) {
		return false
	}
	if q.NumericHistogram != nil {
		if result.NumericHistogram == nil || !reflect.DeepEqual(*result.NumericHistogram, cpLiveHistogram(*q.Numeric, rows)) {
			return false
		}
	} else if result.NumericHistogram != nil {
		return false
	}
	plain := []peLiveRecord{}
	byID := map[string]decLiveRecord{}
	for _, r := range rows {
		attrs := map[string]string{"team": "core"}
		for _, f := range []exploration.Field{{Kind: "attribute", Key: "work.amount"}, {Kind: "attribute", Key: "amount.scale1"}, {Kind: "attribute", Key: "amount.scale9"}, {Kind: "attribute", Key: "work.items"}} {
			raw, _ := cpLiveField(f, r)
			if raw != nil {
				attrs[f.Key] = *raw
			}
		}
		resource := map[string]string{}
		if r.ResourceRaw != nil {
			resource["work.amount"] = *r.ResourceRaw
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
		expectedAttrs := map[string]string{"team": "core"}
		for key, v := range map[string]*string{"work.amount": fact.Raw, "amount.scale1": fact.Scale1Raw, "amount.scale9": fact.Scale9Raw, "work.items": fact.IntegerRaw} {
			if v != nil {
				expectedAttrs[key] = *v
			}
		}
		var attrs map[string]string
		physical := "SpanAttributes"
		if q.Signal == "logs" {
			physical = "Attributes"
		}
		if json.Unmarshal(record["Attributes"], &attrs) != nil || !reflect.DeepEqual(attrs, expectedAttrs) {
			return false
		}
		attrs = nil
		if json.Unmarshal(record[physical], &attrs) != nil || !reflect.DeepEqual(attrs, expectedAttrs) {
			return false
		}
		var resources map[string]string
		expectedResources := map[string]string{"service.name": fact.Service, "deployment.environment.name": fact.Environment}
		if fact.ResourceRaw != nil {
			expectedResources["work.amount"] = *fact.ResourceRaw
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
				value, e := cpLiveField(f, r)
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
		if json.Unmarshal(group["count"], &count) != nil || count != strconv.Itoa(len(members)) || json.Unmarshal(group["numeric"], &numeric) != nil || !reflect.DeepEqual(numeric, cpLiveSummary(*q.Numeric, members)) {
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
			if p.FromMS != from || p.ToMS != to || p.Count != strconv.Itoa(len(members)) || p.State != state || p.Numeric == nil || !reflect.DeepEqual(*p.Numeric, cpLiveSummary(*q.Numeric, members)) {
				return false
			}
		}
	} else if len(result.TimeSeries) > 0 {
		return false
	}
	return result.Duration == nil && result.SpanStatus == nil && len(result.GroupedTimeSeries) == 0
}

func cpLiveCohort(result exploration.Result) (exploration.Cohort, error) {
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
func cpLivePair(request exploration.ComparisonRequest, response exploration.CohortComparison, baseline, selection exploration.Result, environment string, records []decLiveRecord) bool {
	if !request.Numeric || request.Distribution || !cpLiveSource(baseline, environment, records) || !cpLiveSource(selection, environment, records) || baseline.Question.Numeric == nil || selection.Question.Numeric == nil || *baseline.Question.Numeric != *selection.Question.Numeric {
		return false
	}
	a, e := cpLiveCohort(baseline)
	if e != nil {
		return false
	}
	z, e := cpLiveCohort(selection)
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
func cpLiveSummary(b exploration.NumericBinding, rows []decLiveRecord) exploration.NumericSummary {
	values := []*string{}
	for _, r := range rows {
		raw, e := cpLiveField(b.Field, r)
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
func cpLiveHistogram(b exploration.NumericBinding, rows []decLiveRecord) exploration.NumericHistogram {
	values := []*string{}
	for _, r := range rows {
		raw, e := cpLiveField(b.Field, r)
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

type cpLiveEvent struct {
	Action                                   string
	Reads, Plans, Protocols, Writes, Streams int
}
type cpLiveOrigin struct {
	Kind   string
	Parent exploration.Query
	Field  exploration.Field
	Staged exploration.Query
	Result *exploration.Result
}
type cpLiveLedger struct {
	Version             int
	Environment, Signal string
	Records             []decLiveRecord
	Q                   exploration.Query
	Applied             struct {
		Selector struct {
			Signal, Environment string
			Where               *exploration.Predicate
		}
		ReadWindow struct{ FromMS, ToMS int64 }
		HeldRecord json.RawMessage
		HeldRow    struct {
			Record     json.RawMessage
			ReadWindow struct{ FromMS, ToMS int64 }
		}
	}
	Origins                                                 []cpLiveOrigin
	Result, ResourceResult, ParentAgain, Selected, SavedRun exploration.Result
	Selection                                               decLivePlan
	CountPlan                                               exploration.NumericHistogramSelection
	Comparison                                              struct {
		Request  exploration.ComparisonRequest
		Response exploration.CohortComparison
	}
	Saved, Reopened exploration.CapturePage
	Stored          exploration.SavedCapture
	Custody         decLiveCustody
	StoredBytes     int
	StoredSHA       string
	Events          []cpLiveEvent
	Requests        []struct {
		Effect, Endpoint string
		Body             json.RawMessage
	}
	Observations []struct {
		Reason, BeforeURL, AfterURL string
		DialogOpen                  bool
		Streams, Reads              int
	}
}

func cpLiveTransform(o cpLiveOrigin) bool {
	q, p := o.Staged, o.Parent
	if o.Field.Kind != "attribute" && o.Field.Kind != "resource" || o.Field.Key != "work.amount" {
		return false
	}
	b := exploration.NumericBinding{Field: o.Field, Type: "decimal64", Scale: 2, Unit: "USD", Parse: "canonical-decimal"}
	if q.Numeric == nil || *q.Numeric != b || q.NumericHistogram == nil || q.NumericHistogram.Binning != "powers-of-two" || q.Limit != 100 || q.Offset != 0 || len(q.GroupBy) != 0 || q.GroupTimeSeries || q.BucketMS != 0 || q.Summary != "" || q.SpanStatus {
		return false
	}
	return q.Signal == p.Signal && q.Environment == p.Environment && q.FromMS == p.FromMS && q.ToMS == p.ToMS && reflect.DeepEqual(q.Where, p.Where) && reflect.DeepEqual(q.ReportedEdge, p.ReportedEdge)
}
func cpLiveScope(l cpLiveLedger) bool {
	if l.Version != 1 || len(l.Records) != 400 || len(l.Origins) != 4 || l.Saved.Receipt.Question.Explore == nil {
		return false
	}
	a := l.Applied
	if a.Selector.Signal != l.Signal || a.Selector.Environment != l.Environment || a.ReadWindow.ToMS <= a.ReadWindow.FromMS || a.ReadWindow != a.HeldRow.ReadWindow || !aeJSONEqual(a.HeldRecord, a.HeldRow.Record) {
		return false
	}
	parent := exploration.Query{Signal: a.Selector.Signal, Environment: a.Selector.Environment, FromMS: a.ReadWindow.FromMS, ToMS: a.ReadWindow.ToMS, Where: a.Selector.Where, Limit: 100}
	if !nhLiveSameQuery(parent, l.Origins[0].Parent) || !reflect.DeepEqual(parent.Where, l.Q.Where) || !nhLiveSameQuery(l.Origins[1].Parent, l.Result.Question) || !nhLiveSameQuery(l.Origins[2].Parent, l.ResourceResult.Question) || !nhLiveSameQuery(l.Origins[3].Parent, *l.Saved.Receipt.Question.Explore) {
		return false
	}
	for i, o := range l.Origins {
		if o.Kind != []string{"activity", "fixed", "discovery", "saved"}[i] || !cpLiveTransform(o) {
			return false
		}
		if o.Result != nil && (!nhLiveSameQuery(o.Staged, o.Result.Question) || !cpLiveSource(*o.Result, l.Environment, l.Records)) {
			return false
		}
	}
	if l.Origins[0].Field.Kind != "attribute" || l.Origins[1].Field.Kind != "resource" || l.Origins[2].Field.Kind != "attribute" || l.Origins[3].Field.Kind != "resource" {
		return false
	}
	if !nhLiveSameQuery(l.Origins[0].Staged, l.Result.Question) || !nhLiveSameQuery(l.Origins[1].Staged, l.ResourceResult.Question) || !nhLiveSameQuery(l.Origins[2].Staged, l.ParentAgain.Question) || !nhLiveSameQuery(l.Origins[3].Staged, l.SavedRun.Question) {
		return false
	}
	// The observed record must really belong to this fixed source window. No latest
	// wall-clock inference or row count is accepted as the measurement population.
	var record map[string]json.RawMessage
	var id string
	if json.Unmarshal(a.HeldRecord, &record) != nil || json.Unmarshal(record["SpanId"], &id) != nil {
		return false
	}
	seen := false
	for _, r := range l.Records {
		if r.ID == id {
			seen = r.TimeMS >= parent.FromMS && r.TimeMS < parent.ToMS
		}
	}
	return seen
}
func cpLiveStored(l cpLiveLedger) bool {
	s, p := l.Stored, l.Saved
	if s.Explore == nil || s.Question.Explore == nil || p.Explore == nil || !reflect.DeepEqual(s.CaptureReceipt, p.Receipt) || !nhLiveSameQuery(*s.Question.Explore, l.ParentAgain.Question) || !peLiveCaptureHash(exploration.CapturePage{Receipt: s.CaptureReceipt, Explore: s.Explore}) || !aeJSONEqual(p, l.Reopened) {
		return false
	}
	members, e := cpLiveRows(l.ParentAgain.Question, l.Environment, l.Records)
	if e != nil || len(s.Explore.Rows) != len(members) || s.Explore.HasMore || p.RetainedRecords != len(members) || p.RecordsLimited {
		return false
	}
	n := len(members)
	if n > 100 {
		n = 100
	}
	if !aeJSONEqual(s.Explore.Rows[:n], p.Explore.Rows) || p.Explore.HasMore != (len(members) > n) {
		return false
	}
	all := *s.Explore
	all.Question.Limit = len(members)
	if !cpLiveSource(all, l.Environment, l.Records) || !cpLiveSource(*p.Explore, l.Environment, l.Records) {
		return false
	}
	return decLiveCustodyCheck(decLiveLedger{Custody: l.Custody, StoredBytes: l.StoredBytes, StoredSHA: l.StoredSHA, Stored: l.Stored, Saved: l.Saved})
}
func cpLivePlan(plan exploration.NumericHistogramSelection, parent exploration.Query, from, to int, records []decLiveRecord) bool {
	if from < 0 || to > 128 || from >= to {
		return false
	}
	edges := histogramOracleEdges("int64")
	if plan.FromBin != from || plan.ToBin != to || plan.From != edges[from] || plan.To != edges[to] || !nhLiveSameQuery(plan.Parent, parent) || parent.Numeric == nil {
		return false
	}

	if plan.NumericComparison == nil {
		return false
	}
	for i, q := range []exploration.Query{plan.NumericComparison.Baseline, plan.NumericComparison.Selection} {
		base := []exploration.Query{plan.Rest, plan.Selected}[i]
		base.NumericHistogram = nil
		base.Summary = ""
		base.SpanStatus = false
		base.GroupBy = []exploration.Field{{Kind: "column", Key: "service"}}
		if !nhLiveSameQuery(base, q) {
			return false
		}
	}
	lo, hi := ncLiveInt(plan.From), ncLiveInt(plan.To)
	for _, r := range records {
		in, e := cpLiveMatch(parent.Where, r)
		if e != nil {
			return false
		}
		raw, e := cpLiveField(parent.Numeric.Field, r)
		if e != nil {
			return false
		}
		n, state := decLiveScalar(raw, *parent.Numeric)
		selected := state == "valid" && n.Cmp(lo) >= 0 && n.Cmp(hi) < 0
		yes, e := cpLiveMatch(plan.Selected.Where, r)
		if e != nil || yes != (in && selected) {
			return false
		}
		rest, e := cpLiveMatch(plan.Rest.Where, r)
		if e != nil || rest != (in && !selected) {
			return false
		}
	}
	for _, q := range []exploration.Query{plan.Selected, plan.Rest} {
		if q.FromMS != parent.FromMS || q.ToMS != parent.ToMS || q.Environment != parent.Environment || q.Signal != parent.Signal || !reflect.DeepEqual(q.Numeric, parent.Numeric) || !reflect.DeepEqual(q.ReportedEdge, parent.ReportedEdge) {
			return false
		}
	}
	return true
}
func cpLiveComparison(l cpLiveLedger) bool {
	req, res := l.Comparison.Request, l.Comparison.Response
	if !req.Numeric || req.Distribution || !nhLiveSameQuery(req.Baseline.Question, l.CountPlan.NumericComparison.Baseline) || !nhLiveSameQuery(req.Selection.Question, l.CountPlan.NumericComparison.Selection) {
		return false
	}
	for _, c := range []exploration.Cohort{req.Baseline, req.Selection} {
		rows, e := cpLiveRows(c.Question, l.Environment, l.Records)
		if e != nil || c.Numeric == nil || !reflect.DeepEqual(*c.Numeric, cpLiveSummary(*c.Question.Numeric, rows)) || c.Count != strconv.Itoa(len(rows)) {
			return false
		}
		groups := map[string][]decLiveRecord{}
		if len(c.Question.GroupBy) != 1 {
			return false
		}
		for _, r := range rows {
			v, e := cpLiveField(c.Question.GroupBy[0], r)
			if e != nil {
				return false
			}
			groups[explanationValueKey(v)] = append(groups[explanationValueKey(v)], r)
		}
		if len(groups) != len(c.Groups) {
			return false
		}
		seen := map[string]bool{}
		for _, g := range c.Groups {
			k := explanationValueKey(g.Value)
			rs, ok := groups[k]
			if !ok || seen[k] || g.Count != strconv.Itoa(len(rs)) || g.Numeric == nil || !reflect.DeepEqual(*g.Numeric, cpLiveSummary(*c.Question.Numeric, rs)) {
				return false
			}
			seen[k] = true
		}
	}
	if res.Measures == nil || checkNumericChanges(res.Measures.Numeric, *req.Baseline.Question.Numeric, req.Baseline.Numeric, req.Selection.Numeric, false) != nil || res.BaselineCount != req.Baseline.Count || res.SelectionCount != req.Selection.Count {
		return false
	}
	canonical := req
	canonical.Baseline.Groups = append([]exploration.ValueCount{}, req.Baseline.Groups...)
	canonical.Selection.Groups = append([]exploration.ValueCount{}, req.Selection.Groups...)
	sortNumericGroups(canonical.Baseline.Groups)
	sortNumericGroups(canonical.Selection.Groups)
	raw, _ := json.Marshal(canonical)
	hash := sha256.Sum256(raw)
	if res.Basis != hex.EncodeToString(hash[:]) {
		return false
	}
	return true
}
func cpLiveRequests(l cpLiveLedger) bool {
	expected := []exploration.Query{l.Result.Question, l.ResourceResult.Question, l.ParentAgain.Question, l.Selected.Question, l.Comparison.Request.Baseline.Question, l.Comparison.Request.Selection.Question, l.ParentAgain.Question, l.SavedRun.Question}
	i, plans, protocols, captures, streams := 0, 0, 0, 0, 0
	for _, r := range l.Requests {
		switch r.Effect {
		case "source":
			if r.Endpoint == "/api/captures" {
				var p struct {
					Title    string
					Question exploration.CaptureQuestion
				}
				if captures != 0 || json.Unmarshal(r.Body, &p) != nil || p.Question.Explore == nil || p.Title != l.Saved.Receipt.Title || !nhLiveSameQuery(*p.Question.Explore, l.ParentAgain.Question) {
					return false
				}
				captures++
				continue
			}
			var q exploration.Query
			if r.Endpoint != "/api/explore" || i >= len(expected) || json.Unmarshal(r.Body, &q) != nil {
				return false
			}
			// The two independent comparison reads may complete in either source order.
			if i == 4 || i == 5 {
				a, z := expected[4], expected[5]
				if i == 4 && nhLiveSameQuery(q, z) {
					expected[4], expected[5] = z, a
				}
			}
			if !nhLiveSameQuery(q, expected[i]) {
				return false
			}
			i++
		case "plan":
			var p struct {
				Parent         exploration.Query
				FromBin, ToBin int
			}
			if r.Endpoint != "/api/numeric-histograms/selection" || json.Unmarshal(r.Body, &p) != nil || !nhLiveSameQuery(p.Parent, l.ParentAgain.Question) || p.FromBin != l.Selection.Request.FromBin || p.ToBin != l.Selection.Request.ToBin {
				return false
			}
			plans++
		case "comparison":
			var p exploration.ComparisonRequest
			if r.Endpoint != "/api/compare" || json.Unmarshal(r.Body, &p) != nil || !reflect.DeepEqual(p, l.Comparison.Request) {
				return false
			}
			protocols++
		case "stream":
			var p struct {
				Signal, Environment string
				Where               *exploration.Predicate
			}
			if r.Endpoint != "/api/tail" || json.Unmarshal(r.Body, &p) != nil || p.Environment != l.Environment || p.Signal != "" && p.Signal != l.Signal || !reflect.DeepEqual(p.Where, l.Applied.Selector.Where) {
				return false
			}
			streams++
		default:
			return false
		}
	}
	return i == 8 && captures == 1 && plans == 2 && protocols == 1 && streams == 1
}
func cpLiveSemantic(l cpLiveLedger) bool {
	if !cpLiveScope(l) || !cpLiveStored(l) || !cpLiveRequests(l) {
		return false
	}
	for _, r := range []exploration.Result{l.Result, l.ResourceResult, l.ParentAgain, l.Selected, l.SavedRun} {
		if !cpLiveSource(r, l.Environment, l.Records) {
			return false
		}
	}
	if !nhLiveSameQuery(l.Selection.Request.Parent, l.ParentAgain.Question) || !nhLiveSameQuery(l.Selected.Question, l.Selection.Response.Selected) || !cpLivePlan(l.Selection.Response, l.ParentAgain.Question, l.Selection.Request.FromBin, l.Selection.Request.ToBin, l.Records) || !cpLivePlan(l.CountPlan, l.ParentAgain.Question, l.Selection.Request.FromBin, l.Selection.Request.ToBin, l.Records) || !cpLiveComparison(l) {
		return false
	}
	return true
}
func cpLiveReplay(t *testing.T, l cpLiveLedger, ownership bool) bool {
	t.Helper()
	model := "model/contextual_numeric_plot.yaml"
	if ownership {
		model = "model/contextual_numeric_plot_ownership.yaml"
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for _, e := range l.Events {
		rec.Record(context.Background(), strings.ReplaceAll(e.Action, "-", "_"), nil, "obs.coherent", true, "obs.reads", e.Reads, "obs.plans", e.Plans, "obs.protocols", e.Protocols, "obs.writes", e.Writes, "obs.streams", e.Streams)
	}
	checker, e := validate.NewChecker(model, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	result, e := checker.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return result.OK()
}
func cpLiveOwnership(l cpLiveLedger) bool {
	if l.Version != 1 || len(l.Observations) != 3 {
		return false
	}
	for i, o := range l.Observations {
		if o.Reason != []string{"navigation", "dirty-scope", "resume"}[i] || o.DialogOpen || o.Reads != 0 || o.Streams != []int{1, 1, 3}[i] || o.AfterURL == "" || o.BeforeURL == "" {
			return false
		}
		if i != 0 && o.BeforeURL != o.AfterURL {
			return false
		}
	}
	for _, r := range l.Requests {
		if r.Effect != "stream" || r.Endpoint != "/api/tail" {
			return false
		}
	}
	return len(l.Requests) == 3
}
func TestQuintContextualNumericPlotLiveLedger(t *testing.T) {
	file := os.Getenv("GOO_CONTEXTUAL_NUMERIC_PLOT_LEDGER")
	if file == "" {
		t.Skip("requires actual contextual plot journey ledger")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var base cpLiveLedger
	if e = json.Unmarshal(raw, &base); e != nil {
		t.Fatal(e)
	}
	stored, e := os.ReadFile(filepath.Join(filepath.Dir(file), "scope-capture.json"))
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(stored, &base.Stored); e != nil {
		t.Fatal(e)
	}
	base.StoredBytes = len(stored)
	hash := sha256.Sum256(stored)
	base.StoredSHA = hex.EncodeToString(hash[:])
	custody, e := os.ReadFile(filepath.Join(filepath.Dir(file), "capture-custody.json"))
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(custody, &base.Custody); e != nil {
		t.Fatal(e)
	}
	if !cpLiveSemantic(base) {
		t.Fatalf("independent contextual facts rejected: scope=%v stored=%v requests=%v comparison=%v sources=%v/%v/%v/%v/%v", cpLiveScope(base), cpLiveStored(base), cpLiveRequests(base), cpLiveComparison(base), cpLiveSource(base.Result, base.Environment, base.Records), cpLiveSource(base.ResourceResult, base.Environment, base.Records), cpLiveSource(base.ParentAgain, base.Environment, base.Records), cpLiveSource(base.Selected, base.Environment, base.Records), cpLiveSource(base.SavedRun, base.Environment, base.Records))
	}
	if !cpLiveReplay(t, base, false) {
		t.Fatal("actual lifecycle rejected")
	}

	controls := []struct {
		name   string
		mutate func(*cpLiveLedger)
	}{
		{"scope-window", func(l *cpLiveLedger) { l.Origins[0].Staged.ToMS++ }},
		{"scope-where", func(l *cpLiveLedger) { l.Origins[0].Staged.Where = nil }},
		{"namespace", func(l *cpLiveLedger) { l.Origins[1].Staged.Numeric.Field.Kind = "attribute" }},
		{"scale", func(l *cpLiveLedger) { l.Origins[0].Staged.Numeric.Scale = 9 }},
		{"stale-group", func(l *cpLiveLedger) {
			l.Origins[0].Staged.GroupBy = []exploration.Field{{Kind: "column", Key: "service"}}
		}},
		{"stale-bucket", func(l *cpLiveLedger) { l.Origins[0].Staged.BucketMS = 1000 }},
		{"held-window", func(l *cpLiveLedger) { l.Applied.HeldRow.ReadWindow.ToMS++ }},
		{"held-record", func(l *cpLiveLedger) { l.Applied.HeldRecord = json.RawMessage(`{"SpanId":"wrong"}`) }},
		{"saved-parent", func(l *cpLiveLedger) { l.Origins[3].Parent.FromMS++ }},
		{"duplicate-row", func(l *cpLiveLedger) { l.Result.Rows[1] = l.Result.Rows[0] }},
		{"raw-value", func(l *cpLiveLedger) { cpLiveMutateRaw(&l.Result.Rows[0], "attribute") }},
		{"raw-resource", func(l *cpLiveLedger) { cpLiveMutateRaw(&l.ResourceResult.Rows[0], "resource") }},
		{"summary", func(l *cpLiveLedger) { l.Result.Numeric.Sum += "0" }},
		{"summary-scale", func(l *cpLiveLedger) { l.Result.Numeric.Scale = 9 }},
		{"histogram-count", func(l *cpLiveLedger) { l.Result.NumericHistogram.Bins[0].Count += "0" }},
		{"histogram-boundary", func(l *cpLiveLedger) { l.Result.NumericHistogram.Bins[0].From = "0" }},
		{"selected-membership", func(l *cpLiveLedger) { l.Selected.Question.Where = nil }},
		{"selection-index", func(l *cpLiveLedger) { l.Selection.Response.FromBin++ }},
		{"rest-excluded", func(l *cpLiveLedger) { l.CountPlan.Rest.Where = l.CountPlan.Selected.Where }},
		{"comparison-mean", func(l *cpLiveLedger) { l.Comparison.Response.Measures.Numeric.MeanDifference = "1234567/13" }},
		{"comparison-basis", func(l *cpLiveLedger) { l.Comparison.Response.Basis = "wrong" }},
		{"custody-sha", func(l *cpLiveLedger) { l.Custody.SHA256 = "wrong" }},
		{"stored-id", func(l *cpLiveLedger) { l.Stored.ID = "wrong" }},
		{"stored-row", func(l *cpLiveLedger) { cpLiveMutateRaw(&l.Stored.Explore.Rows[0], "attribute") }},
		{"reopened-custody", func(l *cpLiveLedger) { l.Reopened.Receipt.CapturedAt = "wrong" }},
		{"request-endpoint", func(l *cpLiveLedger) { l.Requests[1].Endpoint = "/api/inventory" }},
	}
	for _, action := range []string{"activity-cancel", "activity-stage", "resource-stage", "discovery-stage", "brush", "comparison-stage", "saved-stage", "return-paused"} {
		a := action
		controls = append(controls, struct {
			name   string
			mutate func(*cpLiveLedger)
		}{"hidden-read-" + a, func(l *cpLiveLedger) {
			for i := range l.Events {
				if l.Events[i].Action == a {
					l.Events[i].Reads++
					return
				}
			}
		}})
	}
	for _, control := range controls {
		t.Run("reject-"+control.name, func(t *testing.T) {
			before, _ := json.Marshal(base)
			var changed cpLiveLedger
			json.Unmarshal(before, &changed)
			control.mutate(&changed)
			after, _ := json.Marshal(changed)
			if string(before) == string(after) {
				t.Fatal("control did not change ledger")
			}
			if cpLiveSemantic(changed) && cpLiveReplay(t, changed, false) {
				t.Fatal("corrupted evidence accepted")
			}
		})
	}
}
func TestQuintContextualNumericPlotOwnershipLedger(t *testing.T) {
	file := os.Getenv("GOO_CONTEXTUAL_NUMERIC_PLOT_OWNERSHIP_LEDGER")
	if file == "" {
		t.Skip("requires actual contextual plot ownership ledger")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var base cpLiveLedger
	if e = json.Unmarshal(raw, &base); e != nil {
		t.Fatal(e)
	}
	if !cpLiveOwnership(base) || !cpLiveReplay(t, base, true) {
		t.Fatal("actual ownership rejected")
	}

	for i := 0; i < 14; i++ {
		n := i
		t.Run(fmt.Sprintf("reject-ownership-%02d", n), func(t *testing.T) {
			before, _ := json.Marshal(base)
			var l cpLiveLedger
			json.Unmarshal(before, &l)
			switch {
			case n < 3:
				l.Observations[n].DialogOpen = true
			case n < 6:
				l.Observations[n-3].Reads++
			case n < 9:
				l.Observations[n-6].Streams++
			case n == 9:
				l.Observations[1].AfterURL += "-stale-stage"
			case n == 10:
				l.Requests[0].Effect = "source"
			case n == 11:
				l.Events[1].Reads++
			case n == 12:
				l.Events[2].Action = l.Events[1].Action
			case n == 13:
				l.Events[3].Streams++
			}
			after, _ := json.Marshal(l)
			if string(before) == string(after) {
				t.Fatal("control did not change ledger")
			}
			if cpLiveOwnership(l) && cpLiveReplay(t, l, true) {
				t.Fatal("stale ownership accepted")
			}
		})
	}
}

func cpLiveMutateRaw(row *json.RawMessage, kind string) {
	var value map[string]json.RawMessage
	json.Unmarshal(*row, &value)
	key := "Attributes"
	if _, ok := value["SpanAttributes"]; ok {
		key = "SpanAttributes"
	}
	if kind == "resource" {
		key = "ResourceAttributes"
	}
	var attrs map[string]string
	json.Unmarshal(value[key], &attrs)
	if attrs == nil {
		attrs = map[string]string{}
	}
	attrs["work.amount"] = "999999.99"
	value[key], _ = json.Marshal(attrs)
	*row, _ = json.Marshal(value)
}

type cpLivePopulation struct {
	Version           int
	Environment       string
	Records           []decLiveRecord
	View              workbench.ExplorationView
	OriginalQueries   []exploration.Query
	ParentResults     []exploration.Result
	Children          []cpLiveOrigin
	CompanionRetained bool
	Events            []cpLiveEvent
	Requests          []struct {
		Effect, Endpoint string
		Body             json.RawMessage
	}
}

func cpLivePlainSource(result exploration.Result, environment string, records []decLiveRecord) bool {
	facts := []peLiveRecord{}
	for _, r := range records {
		attrs := map[string]string{"team": "core"}
		resource := map[string]string{}
		for key, raw := range map[string]*string{"work.amount": r.Raw, "amount.scale1": r.Scale1Raw, "amount.scale9": r.Scale9Raw, "work.items": r.IntegerRaw} {
			if raw != nil {
				attrs[key] = *raw
			}
		}
		if r.ResourceRaw != nil {
			resource["work.amount"] = *r.ResourceRaw
		}
		facts = append(facts, peLiveRecord{Environment: r.Environment, Signal: r.Signal, ID: r.ID, Service: r.Service, TimeMS: r.TimeMS, Status: "Ok", Severity: "INFO", Attributes: attrs, Resource: resource})
	}
	byID := map[string]decLiveRecord{}
	for _, r := range records {
		byID[r.ID] = r
	}
	normalized := result
	normalized.Rows = make([]json.RawMessage, len(result.Rows))
	for i, raw := range result.Rows {
		var row map[string]json.RawMessage
		var id string
		if json.Unmarshal(raw, &row) != nil || json.Unmarshal(row["SpanId"], &id) != nil {
			return false
		}
		fact, ok := byID[id]
		if !ok {
			return false
		}
		expected := map[string]string{"team": "core"}
		for key, v := range map[string]*string{"work.amount": fact.Raw, "amount.scale1": fact.Scale1Raw, "amount.scale9": fact.Scale9Raw, "work.items": fact.IntegerRaw} {
			if v != nil {
				expected[key] = *v
			}
		}
		physical := "SpanAttributes"
		if result.Question.Signal == "logs" {
			physical = "Attributes"
		}
		var attrs map[string]string
		if json.Unmarshal(row[physical], &attrs) != nil || !reflect.DeepEqual(attrs, expected) {
			return false
		}
		resources := map[string]string{"service.name": fact.Service, "deployment.environment.name": fact.Environment}
		if fact.ResourceRaw != nil {
			resources["work.amount"] = *fact.ResourceRaw
		}
		var actual map[string]string
		var trace string
		if json.Unmarshal(row["ResourceAttributes"], &actual) != nil || !reflect.DeepEqual(actual, resources) || json.Unmarshal(row["TraceId"], &trace) != nil || trace != fact.TraceID {
			return false
		}
		if result.Question.Signal == "logs" {
			var body string
			if json.Unmarshal(row["Body"], &body) != nil || body != "declared amount "+id {
				return false
			}
			row["Body"], _ = json.Marshal("work observed " + id)
		}
		normalized.Rows[i], _ = json.Marshal(row)
	}
	return peLiveSource(normalized, environment, facts)
}
func cpLivePopulationSemantic(l cpLivePopulation) bool {
	if l.Version != 1 || len(l.Records) != 800 || len(l.OriginalQueries) != 2 || len(l.ParentResults) != 2 || len(l.Children) != 2 || !l.CompanionRetained {
		return false
	}
	if !reflect.DeepEqual(l.View.Queries, l.OriginalQueries) {
		return false
	}
	parents := map[string]exploration.Query{}
	for _, q := range l.OriginalQueries {
		if q.Environment != l.Environment {
			return false
		}
		parents[q.Signal] = q
	}
	if len(parents) != 2 {
		return false
	}
	for _, r := range l.ParentResults {
		if !nhLiveSameQuery(r.Question, parents[r.Question.Signal]) || !cpLivePlainSource(r, l.Environment, l.Records) {
			return false
		}
	}
	for i, c := range l.Children {
		signal := []string{"spans", "logs"}[i]
		kind := []string{"attribute", "resource"}[i]
		if c.Kind != []string{"population-record", "population-field"}[i] || c.Field.Kind != kind || !nhLiveSameQuery(c.Parent, parents[signal]) || !cpLiveTransform(c) || c.Result == nil || !nhLiveSameQuery(c.Staged, c.Result.Question) || !cpLiveSource(*c.Result, l.Environment, l.Records) {
			return false
		}
	}
	expected := []exploration.Query{l.OriginalQueries[0], l.OriginalQueries[1], l.Children[0].Staged, l.Children[1].Staged}
	if len(l.Requests) != 4 {
		return false
	}
	for i, r := range l.Requests {
		var q exploration.Query
		if r.Effect != "source" || r.Endpoint != "/api/explore" || json.Unmarshal(r.Body, &q) != nil {
			return false
		}
		if i == 0 && nhLiveSameQuery(q, expected[1]) {
			expected[0], expected[1] = expected[1], expected[0]
		}
		if !nhLiveSameQuery(q, expected[i]) {
			return false
		}
	}
	return true
}
func cpLivePopulationReplay(t *testing.T, l cpLivePopulation) bool {
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for _, e := range l.Events {
		rec.Record(context.Background(), strings.ReplaceAll(e.Action, "-", "_"), nil, "obs.coherent", true, "obs.reads", e.Reads, "obs.plans", e.Plans, "obs.protocols", e.Protocols, "obs.writes", e.Writes, "obs.streams", e.Streams)
	}
	checker, e := validate.NewChecker("model/contextual_numeric_plot_population.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	r, e := checker.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return r.OK()
}
func TestQuintContextualNumericPlotPopulationLedger(t *testing.T) {
	file := os.Getenv("GOO_CONTEXTUAL_NUMERIC_PLOT_POPULATION_LEDGER")
	if file == "" {
		t.Skip("requires actual two-signal contextual population journey")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var base cpLivePopulation
	if e = json.Unmarshal(raw, &base); e != nil {
		t.Fatal(e)
	}
	if !cpLivePopulationSemantic(base) || !cpLivePopulationReplay(t, base) {
		t.Fatal("actual two-signal population evidence rejected")
	}
	for i := 0; i < 14; i++ {
		n := i
		t.Run(fmt.Sprintf("reject-population-%02d", n), func(t *testing.T) {
			before, _ := json.Marshal(base)
			var l cpLivePopulation
			json.Unmarshal(before, &l)
			switch n {
			case 0:
				l.CompanionRetained = false
			case 1:
				l.OriginalQueries[1].ToMS++
			case 2:
				l.Children[0].Staged.Where = nil
			case 3:
				l.Children[1].Staged.Numeric.Field.Kind = "attribute"
			case 4:
				l.Children[1].Staged.Signal = "spans"
			case 5:
				l.Children[0].Result.Numeric.Scale = 9
			case 6:
				l.Children[1].Result.Rows[1] = l.Children[1].Result.Rows[0]
			case 7:
				cpLiveMutateRaw(&l.Children[0].Result.Rows[0], "attribute")
			case 8:
				l.Requests[2].Endpoint = "/api/inventory"
			case 9:
				l.Events[0].Reads++
			case 10:
				l.Events[2].Reads++
			case 11:
				l.Events[4].Reads++
			case 12:
				l.Events[5].Reads++
			case 13:
				l.Events[7].Reads++
			}
			after, _ := json.Marshal(l)
			if string(before) == string(after) {
				t.Fatal("control did not change ledger")
			}
			if cpLivePopulationSemantic(l) && cpLivePopulationReplay(t, l) {
				t.Fatal("corrupted population evidence accepted")
			}
		})
	}
}
