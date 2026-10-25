package verification

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
)

type tsLiveRecord struct {
	TraceID                         string `json:"traceId"`
	SpanID                          string `json:"spanId"`
	ParentSpanID                    string `json:"parentSpanId"`
	StartNS                         string `json:"startNs"`
	DurationNS                      string `json:"durationNs"`
	Service, Operation, Environment string
	Status, Kind                    int
	Attributes, Resource            map[string]string
}
type tsLiveTrace struct {
	Role, Environment string
	TraceID           string `json:"traceId"`
	Records           []tsLiveRecord
}
type tsLiveFixture struct {
	Version                int
	Environment            string
	FromMS                 int64 `json:"fromMs"`
	ToMS                   int64 `json:"toMs"`
	ConditionA, ConditionB exploration.Predicate
	Traces                 []tsLiveTrace
	Ready                  exploration.Result
	Ingestion              json.RawMessage
	PhysicalProbes         struct{ Common, Other exploration.Result }
}

func tsLiveRecords(f tsLiveFixture) []tsRecord {
	records := []tsRecord{}
	kinds := []string{"Unspecified", "Internal", "Server", "Client", "Producer", "Consumer"}
	for _, trace := range f.Traces {
		for _, r := range trace.Records {
			values := map[exploration.Field]string{}
			put := func(kind, key, value string) { values[exploration.Field{Kind: kind, Key: key}] = value }
			put("column", "trace_id", r.TraceID)
			put("column", "span_id", r.SpanID)
			put("column", "parent_span_id", r.ParentSpanID)
			put("column", "service", r.Service)
			put("column", "name", r.Operation)
			put("column", "timestamp_ns", r.StartNS)
			put("column", "duration_ns", r.DurationNS)
			put("column", "scope", "goo-journey-186")
			put("column", "scope_version", "")
			put("column", "exception_event", "no")
			put("column", "exception_type_state", "unreported")
			if r.Kind >= 0 && r.Kind < len(kinds) {
				put("column", "kind", kinds[r.Kind])
			}
			status, state := "Unset", "unset"
			if r.Status == 1 {
				status, state = "Ok", "ok"
			} else if r.Status == 2 {
				status, state = "Error", "error"
			}
			put("column", "status", status)
			put("column", "status_class", state)
			for k, v := range r.Attributes {
				put("attribute", k, v)
			}
			for k, v := range r.Resource {
				put("resource", k, v)
			}
			put("resource", "service.name", r.Service)
			put("resource", "deployment.environment.name", r.Environment)
			records = append(records, tsRecord{Environment: r.Environment, Values: values})
		}
	}
	return records
}
func tsLiveFixtureOK(f tsLiveFixture) bool {
	if f.Version != 1 || f.Environment == "" || f.FromMS < 0 || f.ToMS <= f.FromMS || f.ToMS-f.FromMS > 7200000 || len(f.Traces) < 8 {
		return false
	}
	roles := map[string]bool{}
	ids := map[string]bool{}
	for _, trace := range f.Traces {
		if roles[trace.Role] || ids[trace.TraceID] || !tcID(trace.TraceID, 32) || trace.Environment == "" || len(trace.Records) == 0 {
			return false
		}
		roles[trace.Role] = true
		ids[trace.TraceID] = true
		for _, r := range trace.Records {
			if r.TraceID != trace.TraceID || r.Environment != trace.Environment || !tcID(r.SpanID, 16) || tsInteger(r.StartNS, "uint64") == nil || tsInteger(r.DurationNS, "uint64") == nil || r.Status < 0 || r.Status > 2 || r.Kind < 0 || r.Kind > 5 || r.Service == "" || r.Operation == "" || r.Resource["service.namespace"] != "commerce" {
				return false
			}
		}
	}
	for _, role := range []string{"cross-service", "same-span", "repeated-witness", "condition-a-only", "condition-b-only", "fast-database-error", "wrong-environment", "outside-window"} {
		if !roles[role] {
			return false
		}
	}
	return true
}
func tsLiveSearch(request exploration.TraceSearchRequest, result exploration.TraceSearchResult, f tsLiveFixture) bool {
	if !tsLiveFixtureOK(f) || request.Version != exploration.TraceSearchVersion || request.Environment != f.Environment || request.FromMS != f.FromMS || request.ToMS != f.ToMS || len(request.Conditions) < 1 || len(request.Conditions) > 4 || request.Limit < 1 || request.Limit > 100 || !aeJSONEqual(result.Request, request) || !peLiveClock(result.ReadAt) || !peLiveHex(result.Basis) {
		return false
	}
	expected, total := tsOracle(request, tsLiveRecords(f))
	state := "observed"
	if total == "0" {
		state = "no_data"
	}
	n, _ := strconv.Atoi(total)
	if result.State != state || result.MatchedTraces != total || result.HasMore != (n > request.Limit) || !tsItemsEqual(result.Items, expected) {
		return false
	}
	// The source result is one full aggregate read with a bounded ID prefix.
	// Scope parameters stay bound; full count/witness evidence is checked above
	// independently of the SQL's surface spelling and plan presentation.
	if result.Plan.Parameters["environment"] != request.Environment || result.Plan.Parameters["from"] != strconv.FormatInt(request.FromMS, 10) || result.Plan.Parameters["to"] != strconv.FormatInt(request.ToMS, 10) || result.Plan.SQL == "" || len(result.Plan.SQL) > 256*1024 {
		return false
	}
	return true
}
func tsLiveMatchesQuery(q exploration.Query, f tsLiveFixture) []tsLiveRecord {
	out := []tsLiveRecord{}
	if q.Signal != "spans" {
		return out
	}
	for _, trace := range f.Traces {
		for _, r := range trace.Records {
			if r.Environment != q.Environment {
				continue
			}
			start := tsInteger(r.StartNS, "uint64")
			if start == nil {
				continue
			}
			low := new(big.Int).Mul(big.NewInt(q.FromMS), big.NewInt(1000000))
			high := new(big.Int).Mul(big.NewInt(q.ToMS), big.NewInt(1000000))
			if start.Cmp(low) < 0 || start.Cmp(high) >= 0 {
				continue
			}
			record := tsLiveRecords(tsLiveFixture{Traces: []tsLiveTrace{{Records: []tsLiveRecord{r}}}})[0]
			if q.Where == nil || tsPredicate(*q.Where, record.Values) {
				out = append(out, r)
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if c := tsInteger(a.StartNS, "uint64").Cmp(tsInteger(b.StartNS, "uint64")); c != 0 {
			return c < 0
		}
		if a.TraceID != b.TraceID {
			return a.TraceID < b.TraceID
		}
		if a.SpanID != b.SpanID {
			return a.SpanID < b.SpanID
		}
		if a.Service != b.Service {
			return a.Service < b.Service
		}
		return a.Operation < b.Operation
	})
	return out
}
func tsLiveRaw(raw json.RawMessage, r tsLiveRecord) bool {
	var x struct {
		TraceID                                        string `json:"TraceId"`
		SpanID                                         string `json:"SpanId"`
		ParentID                                       string `json:"ParentSpanId"`
		Service                                        string `json:"ServiceName"`
		Operation                                      string `json:"SpanName"`
		Kind                                           string `json:"SpanKind"`
		Start                                          string `json:"timestamp_ns"`
		Duration                                       string `json:"duration_ns"`
		Attributes, SpanAttributes, ResourceAttributes map[string]string
		Status                                         string `json:"StatusCode"`
		Scope                                          string `json:"ScopeName"`
		ScopeVersion, StatusMessage                    string
	}
	if json.Unmarshal(raw, &x) != nil {
		return false
	}
	kinds := []string{"Unspecified", "Internal", "Server", "Client", "Producer", "Consumer"}
	statuses := []string{"Unset", "Ok", "Error"}
	if x.TraceID != r.TraceID || x.SpanID != r.SpanID || x.ParentID != r.ParentSpanID || x.Service != r.Service || x.Operation != r.Operation || x.Kind != kinds[r.Kind] || x.Start != r.StartNS || x.Duration != r.DurationNS || x.Status != statuses[r.Status] || x.Scope != "goo-journey-186" || x.ScopeVersion != "" || x.StatusMessage != "" || !aeJSONEqual(x.Attributes, r.Attributes) || !aeJSONEqual(x.SpanAttributes, r.Attributes) {
		return false
	}
	resource := map[string]string{"service.name": r.Service, "deployment.environment.name": r.Environment}
	for k, v := range r.Resource {
		resource[k] = v
	}
	if !aeJSONEqual(x.ResourceAttributes, resource) {
		return false
	}
	var physical map[string]json.RawMessage
	_ = json.Unmarshal(raw, &physical)
	for _, key := range []string{"event_times_ns", "event_names", "event_attributes", "linked_attributes", "linked_traces", "linked_spans"} {
		var values []json.RawMessage
		if json.Unmarshal(physical[key], &values) != nil || len(values) != 0 {
			return false
		}
	}
	return true
}
func tsLiveExplore(result exploration.Result, f tsLiveFixture) bool {
	q := result.Question
	if q.Environment != f.Environment || (q.Signal != "spans" && q.Signal != "logs") || q.FromMS != f.FromMS || q.ToMS != f.ToMS || q.Limit != 100 || q.Offset < 0 || q.Numeric != nil || q.ReportedEdge != nil || q.BucketMS != 0 || !peLiveClock(result.CapturedAt) || !peLiveHex(result.Basis) {
		return false
	}
	records := tsLiveMatchesQuery(q, f)
	state := "observed"
	if len(records) == 0 {
		state = "no_data"
	}
	if result.Count != strconv.Itoa(len(records)) || result.State != state || len(result.TimeSeries) != 0 {
		return false
	}
	start := q.Offset
	if start > len(records) {
		start = len(records)
	}
	end := start + q.Limit
	if end > len(records) {
		end = len(records)
	}
	if len(result.Rows) != end-start || result.HasMore != (end < len(records)) {
		return false
	}
	for i, raw := range result.Rows {
		if !tsLiveRaw(raw, records[start+i]) {
			return false
		}
	}
	converted := make([]tpbLiveRecord, len(records))
	for i, r := range records {
		converted[i] = tpbLiveRecord{TraceID: r.TraceID, SpanID: r.SpanID, ParentSpanID: r.ParentSpanID, StartNS: r.StartNS, DurationNS: r.DurationNS, Service: r.Service, Operation: r.Operation, Attributes: r.Attributes, Resource: r.Resource}
	}
	if q.SpanStatus {
		counts := map[int]int{}
		for _, r := range records {
			counts[r.Status]++
		}
		want := exploration.SpanStatusSummary{Count: result.Count, Error: strconv.Itoa(counts[2]), OK: strconv.Itoa(counts[1]), Unset: strconv.Itoa(counts[0]), Unknown: "0"}
		if result.SpanStatus == nil || *result.SpanStatus != want {
			return false
		}
	}
	// Aggregate helper's fixture assumes all spans OK; replace only the detached
	// status payload after checking the original raw status counts above.
	detachedStatus := result.SpanStatus
	if q.SpanStatus {
		allOK := tpbLiveStatus(converted)
		detachedStatus = &allOK
	}
	groups := append([]json.RawMessage{}, result.Groups...)
	if len(q.GroupBy) > 0 {
		for i, raw := range groups {
			var object map[string]json.RawMessage
			if json.Unmarshal(raw, &object) != nil {
				return false
			}
			var count string
			if json.Unmarshal(object["count"], &count) != nil {
				return false
			}
			allOK := exploration.SpanStatusSummary{Count: count, OK: count, Error: "0", Unset: "0", Unknown: "0"}
			if q.SpanStatus {
				var g0, g1 string
				_ = json.Unmarshal(object["g0"], &g0)
				_ = json.Unmarshal(object["g1"], &g1)
				counts := map[int]int{}
				total := 0
				for _, r := range records {
					if r.Service == g0 && r.Operation == g1 {
						counts[r.Status]++
						total++
					}
				}
				want := exploration.SpanStatusSummary{Count: strconv.Itoa(total), Error: strconv.Itoa(counts[2]), OK: strconv.Itoa(counts[1]), Unset: strconv.Itoa(counts[0]), Unknown: "0"}
				var original exploration.SpanStatusSummary
				if json.Unmarshal(object["spanStatus"], &original) != nil || original != want {
					return false
				}
				object["spanStatus"], _ = json.Marshal(allOK)
				groups[i], _ = json.Marshal(object)
			}
		}
	}
	return tpbLiveAggregates(q, result.Duration, detachedStatus, groups, converted)
}

func tsLivePhysical(result exploration.Result, f tsLiveFixture, environment string) bool {
	q := exploration.Query{Signal: "spans", Environment: environment, FromMS: f.FromMS - 20000, ToMS: f.ToMS, Limit: 100, GroupBy: []exploration.Field{{Kind: "column", Key: "trace_id"}}}
	if !nhLiveSameQuery(result.Question, q) || !peLiveClock(result.CapturedAt) || !peLiveHex(result.Basis) || result.State != "observed" || result.Duration != nil || result.SpanStatus != nil || len(result.TimeSeries) != 0 {
		return false
	}
	expected := tsLiveMatchesQuery(q, f)
	if result.Count != strconv.Itoa(len(expected)) || len(result.Rows) != min(len(expected), 100) || result.HasMore != (len(expected) > 100) {
		return false
	}
	for i, raw := range result.Rows {
		if !tsLiveRaw(raw, expected[i]) {
			return false
		}
	}
	groups := map[string]int{}
	for _, r := range expected {
		groups[r.TraceID]++
	}
	if len(groups) != len(result.Groups) {
		return false
	}
	last := ""
	for _, raw := range result.Groups {
		var g struct {
			G0    *string
			Count string
		}
		if json.Unmarshal(raw, &g) != nil || g.G0 == nil || *g.G0 <= last || g.Count != strconv.Itoa(groups[*g.G0]) {
			return false
		}
		last = *g.G0
	}
	return true
}

type tsLivePair struct {
	Request exploration.TraceSearchRequest
	Result  exploration.TraceSearchResult
}
type tsLiveDOM struct {
	Status string
	Scope  []string
	Rows   []struct {
		TraceID   string `json:"traceId"`
		Witnesses []struct {
			ConditionIndex       int `json:"conditionIndex"`
			Summary, Text, Title string
			Fields               []struct {
				Key, State string
				Value      *string `json:"value,omitempty"`
			}
		}
	}
}
type tsLiveObservation struct {
	Hidden, Connected bool
	Rows, Conditions  int
	Status            string
	RunDisabled       bool `json:"runDisabled"`
	CancelVisible     bool `json:"cancelVisible"`
}
type tsLiveEvent struct {
	Action   string
	Source   int
	Observed *tsLiveObservation
}
type tsLiveRequest struct {
	Endpoint, Method, Effect string
	Payload                  json.RawMessage
	Source                   int
}
type tsLiveResponse struct {
	Endpoint          string
	Request, Response json.RawMessage
}
type tsLiveNetwork struct {
	Endpoint, Method string
	Payload          json.RawMessage
}
type tsLiveMain struct {
	tsLiveFixture
	Parent                                         exploration.Result
	Observed, Bounded, Empty                       tsLivePair
	DOM                                            tsLiveDOM
	Layout                                         struct{ Left, Top, Width, PanelHeight, ControlsToFirstResult, FirstBottom float64 }
	DirectSourceReads                              int
	Reference, Current, RepeatedPage, RepeatedNext exploration.Result
	Comparison                                     struct {
		Input struct{ Baseline, Selection tcLiveSideInput }
		Model exploration.TracePageComparison
	}
	Events    []tsLiveEvent
	Requests  []tsLiveRequest
	Responses []tsLiveResponse
	Network   []tsLiveNetwork
}
type tsLiveOwner struct {
	tsLiveFixture
	Parent                            exploration.Result
	Initial, ClosedEditorRerun, Fresh tsLivePair
	HiddenURL                         string
	HeldRequest                       exploration.TraceSearchRequest
	HeldResult                        exploration.TraceSearchResult
	DeliveryError                     *string
	Events                            []tsLiveEvent
	Requests                          []tsLiveRequest
	Responses                         []tsLiveResponse
	Network                           []tsLiveNetwork
}

func tsLiveParentQuestion(f tsLiveFixture) exploration.Query {
	return exploration.Query{Signal: "spans", Environment: f.Environment, FromMS: f.FromMS, ToMS: f.ToMS, Limit: 100, Where: &exploration.Predicate{All: []exploration.Predicate{f.ConditionA, tsLeaf("resource", "service.namespace", "eq", "commerce"), tsLeaf("attribute", "team", "eq", "core")}}}
}
func tsFlatConj(p exploration.Predicate) (map[string]bool, bool) {
	out := map[string]bool{}
	var visit func(exploration.Predicate) bool
	visit = func(p exploration.Predicate) bool {
		if len(p.Any) > 0 {
			return false
		}
		if p.All != nil {
			if len(p.All) == 0 {
				return false
			}
			for _, c := range p.All {
				if !visit(c) {
					return false
				}
			}
			return true
		}
		raw, e := json.Marshal(p)
		if e != nil {
			return false
		}
		out[string(raw)] = true
		return true
	}
	return out, visit(p)
}
func tsLiveRequestScope(request exploration.TraceSearchRequest, parent exploration.Result, f tsLiveFixture) bool {
	if len(parent.Rows) == 0 || request.Version != exploration.TraceSearchVersion || request.Environment != f.Environment || request.FromMS != f.FromMS || request.ToMS != f.ToMS || request.Limit != 20 || len(request.Conditions) != 2 || !aeJSONEqual(request.Conditions[1], f.ConditionB) {
		return false
	}
	row := tcObject(parent.Rows[0])
	service, sok := row["ServiceName"].(string)
	operation, ook := row["SpanName"].(string)
	if !sok || !ook || parent.Question.Where == nil {
		return false
	}
	expected := exploration.Predicate{All: []exploration.Predicate{*parent.Question.Where, tsLeaf("column", "service", "eq", service), tsLeaf("column", "name", "eq", operation)}}
	want, wok := tsFlatConj(expected)
	actual, aok := tsFlatConj(request.Conditions[0])
	return wok && aok && aeJSONEqual(want, actual)
}
func tsLiveDOMCheck(dom tsLiveDOM, request exploration.TraceSearchRequest, result exploration.TraceSearchResult) bool {
	if len(dom.Rows) != len(result.Items) || !strings.Contains(dom.Status, result.MatchedTraces+" distinct reported trace IDs") || !strings.Contains(dom.Status, strconv.Itoa(len(result.Items))+" listed") {
		return false
	}
	scope := strings.Join(dom.Scope, "\n")
	format := func(ms int64) string { return time.UnixMilli(ms).UTC().Format("2006-01-02T15:04:05.000Z") }
	for _, value := range []string{request.Environment, format(request.FromMS), format(request.ToMS), "condition 1 only"} {
		if !strings.Contains(scope, value) {
			return false
		}
	}
	for i, row := range dom.Rows {
		item := result.Items[i]
		if row.TraceID != item.TraceID || len(row.Witnesses) != len(item.Witnesses) {
			return false
		}
		for j, shown := range row.Witnesses {
			w := item.Witnesses[j]
			summary := "Condition " + strconv.Itoa(w.ConditionIndex+1) + " · " + w.MatchingRecords + " matching span records · " + w.Service + " / " + w.Operation
			if shown.ConditionIndex != w.ConditionIndex || shown.Summary != summary || shown.Title != "Timestamp "+w.TimestampNS+" ns · duration "+w.DurationNS+" ns (exact)" || !strings.Contains(shown.Text, w.SpanID) || !strings.Contains(shown.Text, w.Kind) || len(shown.Fields) != len(w.FieldValues) {
				return false
			}
			for k, field := range shown.Fields {
				want := w.FieldValues[k]
				state := "present"
				if want.Value == nil {
					state = "missing"
				}
				if field.Key != want.Field.Kind+":"+want.Field.Key || field.State != state || !aeJSONEqual(field.Value, want.Value) {
					return false
				}
			}
		}
	}
	return true
}
func tsLiveEventCheck(events []tsLiveEvent, names []string, counts []int) bool {
	if len(events) != len(names) {
		return false
	}
	for i, e := range events {
		if e.Action != names[i] || e.Source != counts[i] {
			return false
		}
	}
	return true
}
func tsLiveRawResponse(response tsLiveResponse, expectedEndpoint string, expectedRequest any, expectedResult any) bool {
	return response.Endpoint == expectedEndpoint && tsLiveBoundRequest(response.Request, expectedRequest) && aeJSONEqual(response.Response, expectedResult)
}
func tsLiveProtocol(requests []tsLiveRequest, network []tsLiveNetwork, expected []struct {
	endpoint string
	request  any
}, cancels int) bool {
	source := 0
	cancelCount := 0
	for _, r := range requests {
		if r.Method != "POST" {
			return false
		}
		if r.Effect == "cancel" {
			cancelCount++
			if !strings.HasPrefix(r.Endpoint, "/api/queries/") || !strings.HasSuffix(r.Endpoint, "/cancel") || r.Source != source || !aeJSONEqual(r.Payload, map[string]any{}) {
				return false
			}
			continue
		}
		if r.Effect != "source" || source >= len(expected) {
			return false
		}
		want := expected[source]
		source++
		if r.Source != source || r.Endpoint != want.endpoint || !tsLiveBoundRequest(r.Payload, want.request) {
			return false
		}
	}
	if source != len(expected) || cancelCount != cancels {
		return false
	}
	posts := []tsLiveNetwork{}
	for _, n := range network {
		if n.Method == "POST" {
			posts = append(posts, n)
		} else if n.Method != "GET" || n.Endpoint == "/api/trace-search" || n.Endpoint == "/api/explore" || strings.HasPrefix(n.Endpoint, "/api/tail") {
			return false
		}
	}
	if len(posts) != len(requests) {
		return false
	}
	for i, n := range posts {
		if n.Endpoint != requests[i].Endpoint || !aeJSONEqual(n.Payload, requests[i].Payload) {
			return false
		}
	}
	return true
}
func tsLiveTraceQuery(f tsLiveFixture, id string, offset int) exploration.Query {
	return exploration.Query{Signal: "spans", Environment: f.Environment, FromMS: f.FromMS, ToMS: f.ToMS, Limit: 100, Offset: offset, Where: &exploration.Predicate{All: []exploration.Predicate{tsLeaf("column", "trace_id", "eq", id)}}, Summary: "span-duration", SpanStatus: true, GroupBy: []exploration.Field{{Kind: "column", Key: "service"}, {Kind: "column", Key: "name"}}}
}
func tsLiveRole(f tsLiveFixture, role string) *tsLiveTrace {
	for i, t := range f.Traces {
		if t.Role == role {
			return &f.Traces[i]
		}
	}
	return nil
}

func tsLiveSetup(f tsLiveFixture) bool {
	return tsLiveFixtureOK(f) && tsLiveExplore(f.Ready, f) && nhLiveSameQuery(f.Ready.Question, tsLiveParentQuestion(f)) && tsLivePhysical(f.PhysicalProbes.Common, f, f.Environment) && tsLivePhysical(f.PhysicalProbes.Other, f, f.Environment+"-other")
}
func tsLiveMainScope(l tsLiveMain) bool {
	f := l.tsLiveFixture
	if !tsLiveSetup(f) || !tsLiveExplore(l.Parent, f) || !nhLiveSameQuery(l.Parent.Question, tsLiveParentQuestion(f)) || l.Parent.Count != "135" || len(l.Parent.Rows) != 100 || !l.Parent.HasMore || !tsLiveRequestScope(l.Observed.Request, l.Parent, f) || !tsLiveSearch(l.Observed.Request, l.Observed.Result, f) || l.Observed.Result.MatchedTraces != "15" || !tsLiveDOMCheck(l.DOM, l.Observed.Request, l.Observed.Result) {
		return false
	}
	if l.Layout.Left < 0 || l.Layout.Top < 0 || l.Layout.Width <= 0 || l.Layout.FirstBottom > 2600 || l.Layout.FirstBottom <= l.Layout.Top || l.Layout.ControlsToFirstResult < 0 || l.Layout.ControlsToFirstResult > l.Layout.PanelHeight || l.Layout.Left+l.Layout.Width > 1441 || l.Layout.Top+l.Layout.ControlsToFirstResult > l.Layout.FirstBottom {
		return false
	}
	bounded := l.Observed.Request
	bounded.Limit = 2
	if !aeJSONEqual(l.Bounded.Request, bounded) || !tsLiveSearch(l.Bounded.Request, l.Bounded.Result, f) || !l.Bounded.Result.HasMore {
		return false
	}
	empty := l.Observed.Request
	empty.Conditions = append([]exploration.Predicate{}, empty.Conditions...)
	empty.Conditions[1] = tsLeaf("column", "service", "eq", f.Environment+"-absent-service")
	if !aeJSONEqual(l.Empty.Request, empty) || !tsLiveSearch(l.Empty.Request, l.Empty.Result, f) || l.Empty.Result.State != "no_data" || l.DirectSourceReads != 2 {
		return false
	}
	cross, same, repeat := tsLiveRole(f, "cross-service"), tsLiveRole(f, "same-span"), tsLiveRole(f, "repeated-witness")
	if cross == nil || same == nil || repeat == nil {
		return false
	}
	for _, entry := range []struct {
		result exploration.Result
		id     string
		offset int
	}{{l.Reference, cross.TraceID, 0}, {l.Current, same.TraceID, 0}, {l.RepeatedPage, repeat.TraceID, 0}, {l.RepeatedNext, repeat.TraceID, 100}} {
		if !tsLiveExplore(entry.result, f) || !nhLiveSameQuery(entry.result.Question, tsLiveTraceQuery(f, entry.id, entry.offset)) {
			return false
		}
	}
	if l.Reference.Count != "3" || l.Current.Count != "2" || l.RepeatedPage.Count != "241" || l.RepeatedNext.Count != "241" || len(l.RepeatedPage.Rows) != 100 || len(l.RepeatedNext.Rows) != 100 || !l.RepeatedPage.HasMore || !l.RepeatedNext.HasMore {
		return false
	}
	if !aeJSONEqual(l.Comparison.Input.Baseline.Rows, l.Reference.Rows) || !aeJSONEqual(l.Comparison.Input.Selection.Rows, l.Current.Rows) || l.Comparison.Input.Baseline.Truncated || l.Comparison.Input.Selection.Truncated || !tcOracleEqual(tcOracle(l.Reference.Rows, l.Current.Rows, false, false), l.Comparison.Model) {
		return false
	}
	// Witness membership is independent of either trace inspection page.
	var repeatItem *exploration.TraceSearchItem
	for i, item := range l.Observed.Result.Items {
		if item.TraceID == repeat.TraceID {
			repeatItem = &l.Observed.Result.Items[i]
		}
	}
	if repeatItem == nil || len(repeatItem.Witnesses) != 2 || repeatItem.Witnesses[0].MatchingRecords != "120" || repeatItem.Witnesses[1].MatchingRecords != "120" {
		return false
	}
	contains := func(rows []json.RawMessage, id string) bool {
		for _, raw := range rows {
			obj := tcObject(raw)
			if obj["SpanId"] == id {
				return true
			}
		}
		return false
	}
	if !contains(l.RepeatedPage.Rows, repeatItem.Witnesses[0].SpanID) || contains(l.RepeatedPage.Rows, repeatItem.Witnesses[1].SpanID) || !contains(l.RepeatedNext.Rows, repeatItem.Witnesses[1].SpanID) {
		return false
	}
	return tsLiveEventCheck(l.Events, []string{"parent-run", "stage", "search", "witness-details", "open-cross", "pin-cross", "open-same", "compare", "open-repeated", "next-trace-page"}, []int{1, 1, 2, 2, 4, 4, 6, 6, 8, 10})
}
func tsLiveMainProtocol(l tsLiveMain) bool {
	expected := []struct {
		endpoint string
		request  any
	}{{"/api/explore", l.Parent.Question}, {"/api/trace-search", l.Observed.Request}, {"/api/explore", l.Reference.Question}, {"/api/explore", tcLiveLogQuestion(l.Reference.Question)}, {"/api/explore", l.Current.Question}, {"/api/explore", tcLiveLogQuestion(l.Current.Question)}, {"/api/explore", l.RepeatedPage.Question}, {"/api/explore", tcLiveLogQuestion(l.RepeatedPage.Question)}, {"/api/explore", l.RepeatedNext.Question}, {"/api/explore", tcLiveLogQuestion(l.RepeatedNext.Question)}}
	if !tsLiveProtocol(l.Requests, l.Network, expected, 0) || len(l.Responses) != 10 {
		return false
	}
	for i, want := range expected {
		r := l.Responses[i]
		if r.Endpoint != want.endpoint || !tsLiveBoundRequest(r.Request, want.request) {
			return false
		}
		if want.endpoint == "/api/trace-search" {
			if !aeJSONEqual(r.Response, l.Observed.Result) {
				return false
			}
		} else {
			var result exploration.Result
			if json.Unmarshal(r.Response, &result) != nil || !tsLiveExplore(result, l.tsLiveFixture) {
				return false
			}
			switch i {
			case 0:
				if !aeJSONEqual(result, l.Parent) {
					return false
				}
			case 2:
				if !aeJSONEqual(result, l.Reference) {
					return false
				}
			case 4:
				if !aeJSONEqual(result, l.Current) {
					return false
				}
			case 6:
				if !aeJSONEqual(result, l.RepeatedPage) {
					return false
				}
			case 8:
				if !aeJSONEqual(result, l.RepeatedNext) {
					return false
				}
			}
		}
	}
	return true
}
func tsLiveMainOK(l tsLiveMain) bool { return tsLiveMainScope(l) && tsLiveMainProtocol(l) }
func tsLiveOwnerOK(l tsLiveOwner) bool {
	f := l.tsLiveFixture
	if !tsLiveSetup(f) || !tsLiveExplore(l.Parent, f) || !nhLiveSameQuery(l.Parent.Question, tsLiveParentQuestion(f)) || !tsLiveRequestScope(l.Initial.Request, l.Parent, f) || !tsLiveSearch(l.Initial.Request, l.Initial.Result, f) || !tsLiveSearch(l.ClosedEditorRerun.Request, l.ClosedEditorRerun.Result, f) || !tsLiveSearch(l.HeldRequest, l.HeldResult, f) || !tsLiveSearch(l.Fresh.Request, l.Fresh.Result, f) || !strings.HasSuffix(l.HiddenURL, "#definitions") {
		return false
	}
	if !aeJSONEqual(l.Initial.Request, l.ClosedEditorRerun.Request) || !aeJSONEqual(l.Initial.Request, l.HeldRequest) || !aeJSONEqual(l.Initial.Request, l.Fresh.Request) || l.Initial.Result.MatchedTraces != "15" {
		return false
	}
	expected := []struct {
		endpoint string
		request  any
	}{{"/api/explore", l.Parent.Question}, {"/api/trace-search", l.Initial.Request}, {"/api/trace-search", l.ClosedEditorRerun.Request}, {"/api/trace-search", l.HeldRequest}, {"/api/trace-search", l.Fresh.Request}}
	if !tsLiveProtocol(l.Requests, l.Network, expected, 1) || len(l.Responses) != 4 {
		return false
	}
	if !tsLiveRawResponse(l.Responses[0], "/api/explore", l.Parent.Question, l.Parent) || !tsLiveRawResponse(l.Responses[1], "/api/trace-search", l.Initial.Request, l.Initial.Result) || !tsLiveRawResponse(l.Responses[2], "/api/trace-search", l.ClosedEditorRerun.Request, l.ClosedEditorRerun.Result) || !tsLiveRawResponse(l.Responses[3], "/api/trace-search", l.Fresh.Request, l.Fresh.Result) {
		return false
	}
	for i, read := range []string{l.Initial.Result.ReadAt, l.ClosedEditorRerun.Result.ReadAt, l.HeldResult.ReadAt, l.Fresh.Result.ReadAt} {
		if i > 0 {
			previous := []string{l.Initial.Result.ReadAt, l.ClosedEditorRerun.Result.ReadAt, l.HeldResult.ReadAt}[i-1]
			a, ae := time.Parse(time.RFC3339Nano, previous)
			b, be := time.Parse(time.RFC3339Nano, read)
			if ae != nil || be != nil || !b.After(a) {
				return false
			}
		}
	}
	names := []string{"initial", "closed-editor", "closed-editor-rerun", "closed", "hidden", "held", "superseded", "fresh"}
	counts := []int{2, 2, 3, 3, 3, 4, 4, 5}
	if !tsLiveEventCheck(l.Events, names, counts) {
		return false
	}
	for i, e := range l.Events {
		o := e.Observed
		if o == nil || !o.Connected {
			return false
		}
		switch i {
		case 0, 1, 2, 7:
			if o.Hidden || o.Rows != 15 || o.Conditions != 2 || o.RunDisabled || o.CancelVisible || !strings.Contains(o.Status, "15 distinct reported trace IDs") {
				return false
			}
		case 3, 4:
			if !o.Hidden || o.Rows != 0 || o.Conditions != 0 || !o.RunDisabled || o.CancelVisible {
				return false
			}
		case 5:
			if o.Hidden || o.Rows != 0 || o.Conditions != 2 || !o.RunDisabled || !o.CancelVisible {
				return false
			}
		case 6:
			if o.Hidden || o.Rows != 0 || o.Conditions != 2 || o.RunDisabled || o.CancelVisible || !strings.Contains(o.Status, "Run explicitly") {
				return false
			}
		}
	}
	return true
}

func tsLiveBoundRequest(raw json.RawMessage, want any) bool {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	switch expected := want.(type) {
	case exploration.Query:
		var actual exploration.Query
		if decoder.Decode(&actual) != nil {
			return false
		}
		return nhLiveSameQuery(actual, expected)
	case exploration.TraceSearchRequest:
		var actual exploration.TraceSearchRequest
		if decoder.Decode(&actual) != nil {
			return false
		}
		return aeJSONEqual(actual, expected)
	}
	return aeJSONEqual(raw, want)
}

func tsLiveReplay(t *testing.T, events []tsLiveEvent, model string) bool {
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for _, e := range events {
		rec.Record(context.Background(), "observe_"+strings.ReplaceAll(e.Action, "-", "_"), nil, "obs.coherent", true, "obs.source", e.Source)
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
func tsLiveRejectTraces(t *testing.T, events []tsLiveEvent, model string) {
	for _, kind := range []string{"unexpected source effect", "wrong action order"} {
		t.Run("finite trace rejects "+kind, func(t *testing.T) {
			bad := append([]tsLiveEvent{}, events...)
			if kind == "unexpected source effect" {
				bad[0].Source++
			} else {
				bad[0].Action = bad[1].Action
			}
			tpbLiveChanged(t, events, bad)
			if tsLiveReplay(t, bad, model) {
				t.Fatal("finite model accepted changed trace")
			}
		})
	}
}
func tsLiveFindItem(result *exploration.TraceSearchResult, id string) *exploration.TraceSearchItem {
	for i, item := range result.Items {
		if item.TraceID == id {
			return &result.Items[i]
		}
	}
	return nil
}
func TestQuintTraceSearchLiveLedger(t *testing.T) {
	var base tsLiveMain
	tpbLiveRead(t, "GOO_TRACE_SEARCH_LEDGER", &base)
	if !tsLiveMainOK(base) {
		t.Fatalf("independent trace search rejected scope=%v protocol=%v setup=%v parent=%v search=%v dom=%v", tsLiveMainScope(base), tsLiveMainProtocol(base), tsLiveSetup(base.tsLiveFixture), tsLiveExplore(base.Parent, base.tsLiveFixture), tsLiveSearch(base.Observed.Request, base.Observed.Result, base.tsLiveFixture), tsLiveDOMCheck(base.DOM, base.Observed.Request, base.Observed.Result))
	}
	if !tsLiveReplay(t, base.Events, "model/trace_search.yaml") {
		t.Fatal("actual trace search lifecycle rejected")
	}
	tsLiveRejectTraces(t, base.Events, "model/trace_search.yaml")
	controls := []struct {
		name   string
		mutate func(*tsLiveMain)
	}{
		{"request environment", func(l *tsLiveMain) { l.Observed.Request.Environment += "-other" }},
		{"request window", func(l *tsLiveMain) { l.Observed.Request.FromMS++ }},
		{"parent qualifiers dropped", func(l *tsLiveMain) { l.Observed.Request.Conditions[0] = l.ConditionA }},
		{"independent condition dropped", func(l *tsLiveMain) { l.Observed.Request.Conditions = l.Observed.Request.Conditions[:1] }},
		{"conditions forced onto same span", func(l *tsLiveMain) {
			l.Observed.Request.Conditions = []exploration.Predicate{{All: l.Observed.Request.Conditions}}
		}},
		{"trace population total", func(l *tsLiveMain) { l.Observed.Result.MatchedTraces = "999" }},
		{"trace total replaced by bound", func(l *tsLiveMain) { l.Observed.Result.MatchedTraces = "2" }},
		{"false truncation", func(l *tsLiveMain) { l.Observed.Result.HasMore = true }},
		{"zero trace identity", func(l *tsLiveMain) { l.Observed.Result.Items[0].TraceID = strings.Repeat("0", 32) }},
		{"malformed trace identity", func(l *tsLiveMain) { l.Observed.Result.Items[0].TraceID = strings.Repeat("A", 32) }},
		{"duplicate trace identity", func(l *tsLiveMain) { l.Observed.Result.Items[1].TraceID = l.Observed.Result.Items[0].TraceID }},
		{"unordered trace list", func(l *tsLiveMain) { items := l.Observed.Result.Items; items[0], items[1] = items[1], items[0] }},
		{"missing matched trace", func(l *tsLiveMain) { l.Observed.Result.Items = l.Observed.Result.Items[:14] }},
		{"missing condition witness", func(l *tsLiveMain) { l.Observed.Result.Items[0].Witnesses = l.Observed.Result.Items[0].Witnesses[:1] }},
		{"duplicate condition index", func(l *tsLiveMain) { l.Observed.Result.Items[0].Witnesses[1].ConditionIndex = 0 }},
		{"matching records deduplicated", func(l *tsLiveMain) {
			item := tsLiveFindItem(&l.Observed.Result, tsLiveRole(l.tsLiveFixture, "repeated-witness").TraceID)
			item.Witnesses[0].MatchingRecords = "1"
		}},
		{"representative span ID", func(l *tsLiveMain) { l.Observed.Result.Items[0].Witnesses[0].SpanID = "000000000000ffff" }},
		{"representative parent ID", func(l *tsLiveMain) { l.Observed.Result.Items[0].Witnesses[0].ParentSpanID = "000000000000ffff" }},
		{"representative service", func(l *tsLiveMain) { l.Observed.Result.Items[0].Witnesses[0].Service = "other" }},
		{"representative operation", func(l *tsLiveMain) { l.Observed.Result.Items[0].Witnesses[0].Operation = "wrong" }},
		{"representative kind", func(l *tsLiveMain) { l.Observed.Result.Items[0].Witnesses[0].Kind = "Server" }},
		{"timestamp rounded", func(l *tsLiveMain) { l.Observed.Result.Items[0].Witnesses[0].TimestampNS = "9007199254740992" }},
		{"duration rounded", func(l *tsLiveMain) { l.Observed.Result.Items[0].Witnesses[0].DurationNS = "1" }},
		{"qualified witness namespace", func(l *tsLiveMain) { l.Observed.Result.Items[0].Witnesses[0].FieldValues[0].Field.Kind = "resource" }},
		{"qualified witness raw value", func(l *tsLiveMain) {
			l.Observed.Result.Items[0].Witnesses[0].FieldValues[0].Value = lpcPointer("wrong")
		}},
		{"present witness value missing", func(l *tsLiveMain) { l.Observed.Result.Items[0].Witnesses[0].FieldValues[0].Value = nil }},
		{"witness proof leaf removed", func(l *tsLiveMain) {
			w := &l.Observed.Result.Items[0].Witnesses[0]
			w.FieldValues = w.FieldValues[:len(w.FieldValues)-1]
		}},
		{"witness proof leaf duplicated", func(l *tsLiveMain) {
			w := &l.Observed.Result.Items[0].Witnesses[0]
			w.FieldValues = append(w.FieldValues, w.FieldValues[0])
		}},
		{"proof contradicts service metadata", func(l *tsLiveMain) {
			for i, f := range l.Observed.Result.Items[0].Witnesses[0].FieldValues {
				if f.Field == (exploration.Field{Kind: "column", Key: "service"}) {
					l.Observed.Result.Items[0].Witnesses[0].FieldValues[i].Value = lpcPointer("other")
					return
				}
			}
		}},
		{"bounded list loses full total", func(l *tsLiveMain) { l.Bounded.Result.MatchedTraces = "2" }},
		{"bounded request changes condition", func(l *tsLiveMain) { l.Bounded.Request.Conditions[1].Value = "ok" }},
		{"empty state invented observed", func(l *tsLiveMain) { l.Empty.Result.State = "observed" }},
		{"empty count invented nonzero", func(l *tsLiveMain) { l.Empty.Result.MatchedTraces = "1" }},
		{"parent raw row duplicated", func(l *tsLiveMain) { l.Parent.Rows[1] = l.Parent.Rows[0] }},
		{"parent raw attribute", func(l *tsLiveMain) { lpLiveMutateRow(&l.Parent.Rows[0], "Attributes", "changed") }},
		{"parent scope window", func(l *tsLiveMain) { l.Parent.Question.ToMS++ }},
		{"parent filter lost", func(l *tsLiveMain) { l.Parent.Question.Where = &l.ConditionA }},
		{"physical source count", func(l *tsLiveMain) { l.PhysicalProbes.Common.Count = "290" }},
		{"physical excluded trace omitted", func(l *tsLiveMain) { l.PhysicalProbes.Common.Groups = l.PhysicalProbes.Common.Groups[:18] }},
		{"other environment proof misplaced", func(l *tsLiveMain) { l.PhysicalProbes.Other.Question.Environment = l.Environment }},
		{"trace context uses witness count", func(l *tsLiveMain) { l.Reference.Count = "1" }},
		{"trace context raw identity", func(l *tsLiveMain) {
			lpLiveMutateRow(&l.Reference.Rows[0], "TraceId", l.Observed.Result.Items[0].TraceID+"0")
		}},
		{"repeated trace full count truncated", func(l *tsLiveMain) { l.RepeatedPage.Count = "100" }},
		{"witness replaces inspection page", func(l *tsLiveMain) { l.RepeatedPage.Rows = l.RepeatedPage.Rows[:1] }},
		{"next page offset reset", func(l *tsLiveMain) { l.RepeatedNext.Question.Offset = 0 }},
		{"next page repeats earlier member", func(l *tsLiveMain) { l.RepeatedNext.Rows[0] = l.RepeatedPage.Rows[0] }},
		{"comparison delta sign", func(l *tsLiveMain) {
			for i, g := range l.Comparison.Model.Groups {
				if g.Operation != nil && *g.Operation == "GET /checkout" {
					l.Comparison.Model.Groups[i].InclusiveDifferenceNS = lpcPointer("-100000000")
					return
				}
			}
		}},
		{"comparison row paired arbitrarily", func(l *tsLiveMain) {
			for i, g := range l.Comparison.Model.Groups {
				if g.Match != nil {
					l.Comparison.Model.Groups[i].Match.BaselineIndex = 99
					return
				}
			}
		}},
		{"DOM matching count", func(l *tsLiveMain) { l.DOM.Rows[0].Witnesses[0].Summary = "Condition 1 · 1 matching span record" }},
		{"DOM present state", func(l *tsLiveMain) { l.DOM.Rows[0].Witnesses[0].Fields[0].State = "missing" }},
		{"DOM namespace", func(l *tsLiveMain) { l.DOM.Rows[0].Witnesses[0].Fields[0].Key = "resource:team" }},
		{"DOM exact nanoseconds", func(l *tsLiveMain) { l.DOM.Rows[0].Witnesses[0].Title = "Rounded milliseconds" }},
		{"controls clipped", func(l *tsLiveMain) { l.Layout.Top = -1 }},
		{"auxiliary read budget omitted", func(l *tsLiveMain) { l.DirectSourceReads = 0 }},
		{"source plan scope", func(l *tsLiveMain) { l.Observed.Result.Plan.Parameters["environment"] = "other" }},
		{"missing source operation", func(l *tsLiveMain) { l.Requests = l.Requests[:len(l.Requests)-1] }},
	}
	for _, index := range []int{1, 3, 5, 7} {
		i := index
		controls = append(controls, struct {
			name   string
			mutate func(*tsLiveMain)
		}{name: "hidden read on " + base.Events[i].Action, mutate: func(l *tsLiveMain) { l.Events[i].Source++ }})
	}
	for _, control := range controls {
		t.Run(control.name, func(t *testing.T) {
			bad := tpbLiveClone(base)
			control.mutate(&bad)
			tpbLiveChanged(t, base, bad)
			if tsLiveMainOK(bad) {
				t.Fatal("changed evidence accepted")
			}
		})
	}
}
func TestQuintTraceSearchOwnershipLedger(t *testing.T) {
	var base tsLiveOwner
	tpbLiveRead(t, "GOO_TRACE_SEARCH_OWNERSHIP_LEDGER", &base)
	if !tsLiveOwnerOK(base) {
		t.Fatalf("trace search ownership rejected setup=%v parent=%v initial=%v fresh=%v", tsLiveSetup(base.tsLiveFixture), tsLiveExplore(base.Parent, base.tsLiveFixture), tsLiveSearch(base.Initial.Request, base.Initial.Result, base.tsLiveFixture), tsLiveSearch(base.Fresh.Request, base.Fresh.Result, base.tsLiveFixture))
	}
	if !tsLiveReplay(t, base.Events, "model/trace_search_ownership.yaml") {
		t.Fatal("actual owner lifecycle rejected")
	}
	tsLiveRejectTraces(t, base.Events, "model/trace_search_ownership.yaml")
	controls := []struct {
		name   string
		mutate func(*tsLiveOwner)
	}{
		{"closed editor changes rendered result", func(l *tsLiveOwner) { l.Events[1].Observed.Rows = 0 }},
		{"closed editor changes future request", func(l *tsLiveOwner) { l.ClosedEditorRerun.Request.Conditions[1].Value = "ok" }},
		{"closed owner remains visible", func(l *tsLiveOwner) { l.Events[3].Observed.Hidden = false }},
		{"hidden owner publishes rows", func(l *tsLiveOwner) { l.Events[4].Observed.Rows = 15 }},
		{"held owner publishes rows", func(l *tsLiveOwner) { l.Events[5].Observed.Rows = 15 }},
		{"superseded response publishes rows", func(l *tsLiveOwner) { l.Events[6].Observed.Rows = 15 }},
		{"superseded owner stuck disabled", func(l *tsLiveOwner) { l.Events[6].Observed.RunDisabled = true }},
		{"closed action source read", func(l *tsLiveOwner) { l.Events[3].Source++ }},
		{"hidden action source read", func(l *tsLiveOwner) { l.Events[4].Source++ }},
		{"held action extra source read", func(l *tsLiveOwner) { l.Events[5].Source++ }},
		{"superseded action source read", func(l *tsLiveOwner) { l.Events[6].Source++ }},
		{"fresh action skips source", func(l *tsLiveOwner) { l.Events[7].Source = 4 }},
		{"cancellation counted as source", func(l *tsLiveOwner) {
			for i, r := range l.Requests {
				if r.Effect == "cancel" {
					l.Requests[i].Effect = "source"
					return
				}
			}
		}},
		{"held field proof changed", func(l *tsLiveOwner) { l.HeldResult.Items[0].Witnesses[0].FieldValues[0].Value = lpcPointer("wrong") }},
		{"fresh read reuses held clock", func(l *tsLiveOwner) { l.Fresh.Result.ReadAt = l.HeldResult.ReadAt }},
		{"hidden link navigation", func(l *tsLiveOwner) { l.HiddenURL = "http://127.0.0.1:8088/#explore" }},
	}
	for _, control := range controls {
		t.Run(control.name, func(t *testing.T) {
			bad := tpbLiveClone(base)
			control.mutate(&bad)
			tpbLiveChanged(t, base, bad)
			if tsLiveOwnerOK(bad) {
				t.Fatal("changed owner evidence accepted")
			}
		})
	}
}
