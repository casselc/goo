package verification

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
)

type tcLiveRecord struct {
	TraceID              string `json:"traceId"`
	SpanID               string `json:"spanId"`
	ParentSpanID         string `json:"parentSpanId"`
	StartNS              string `json:"startNs"`
	DurationNS           string `json:"durationNs"`
	Service, Operation   string
	Kind                 int
	Attributes, Resource map[string]string
}
type tcLiveTrace struct {
	Role    string
	TraceID string `json:"traceId"`
	FromMS  int64  `json:"fromMs"`
	ToMS    int64  `json:"toMs"`
	Records []tcLiveRecord
}
type tcLiveFixture struct {
	Version     int
	Environment string
	Traces      []tcLiveTrace
}

func tcLiveFixtureOK(f tcLiveFixture) bool {
	if f.Version != 1 || f.Environment == "" || len(f.Traces) != 5 {
		return false
	}
	roles := []string{"reference", "current", "problem-reference", "problem-current", "partial-current"}
	sizes := []int{7, 7, 3, 3, 126}
	ids := map[string]bool{}
	for i, tr := range f.Traces {
		if tr.Role != roles[i] || len(tr.Records) != sizes[i] || !tcID(tr.TraceID, 32) || ids[tr.TraceID] || tr.FromMS >= tr.ToMS {
			return false
		}
		ids[tr.TraceID] = true
		for _, r := range tr.Records {
			start, duration := tcInteger(r.StartNS), tcInteger(r.DurationNS)
			if start == nil || duration == nil || r.TraceID != tr.TraceID || !tcID(r.SpanID, 16) || r.Kind < 1 || r.Kind > 5 || r.Attributes["work.role"] != tr.Role || r.Resource["service.namespace"] != "commerce" {
				return false
			}
			ms := new(big.Int).Quo(start, big.NewInt(1000000))
			if !ms.IsInt64() || ms.Int64() < tr.FromMS || ms.Int64() >= tr.ToMS {
				return false
			}
		}
	}
	return true
}
func tcLiveMatch(p *exploration.Predicate, r tcLiveRecord) bool {
	if p == nil {
		return true
	}
	if len(p.All) > 0 {
		for _, c := range p.All {
			if !tcLiveMatch(&c, r) {
				return false
			}
		}
		return true
	}
	if len(p.Any) > 0 {
		for _, c := range p.Any {
			if tcLiveMatch(&c, r) {
				return true
			}
		}
		return false
	}
	if p.Field == nil || p.Field.Kind != "column" || p.Op != "eq" {
		return false
	}
	switch p.Field.Key {
	case "trace_id":
		return r.TraceID == p.Value
	case "span_id":
		return r.SpanID == p.Value
	}
	return false
}
func tcLiveRows(q exploration.Query, f tcLiveFixture) []tcLiveRecord {
	rows := []tcLiveRecord{}
	if q.Signal != "spans" {
		return rows
	}
	for _, tr := range f.Traces {
		for _, r := range tr.Records {
			start := tcInteger(r.StartNS)
			if start == nil {
				continue
			}
			ms := new(big.Int).Quo(start, big.NewInt(1000000)).Int64()
			if q.FromMS <= ms && ms < q.ToMS && tcLiveMatch(q.Where, r) {
				rows = append(rows, r)
			}
		}
	}
	sort.SliceStable(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if c := tcInteger(a.StartNS).Cmp(tcInteger(b.StartNS)); c != 0 {
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
	return rows
}
func tcLiveRaw(raw json.RawMessage, r tcLiveRecord, env string) bool {
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
	kinds := map[int]string{1: "Internal", 2: "Server", 3: "Client", 4: "Producer", 5: "Consumer"}
	if x.TraceID != r.TraceID || x.SpanID != r.SpanID || x.ParentID != r.ParentSpanID || x.Service != r.Service || x.Operation != r.Operation || x.Kind != kinds[r.Kind] || x.Start != r.StartNS || x.Duration != r.DurationNS || x.Status != "Ok" || x.Scope != "goo-journey-185" || x.ScopeVersion != "" || x.StatusMessage != "" || !aeJSONEqual(x.Attributes, r.Attributes) || !aeJSONEqual(x.SpanAttributes, r.Attributes) {
		return false
	}
	resource := map[string]string{"service.name": r.Service, "deployment.environment.name": env}
	for k, v := range r.Resource {
		resource[k] = v
	}
	if !aeJSONEqual(resource, x.ResourceAttributes) {
		return false
	}
	var physical map[string]json.RawMessage
	if json.Unmarshal(raw, &physical) != nil {
		return false
	}
	for _, key := range []string{"event_times_ns", "event_names", "event_attributes", "linked_attributes", "linked_traces", "linked_spans"} {
		var values []json.RawMessage
		if json.Unmarshal(physical[key], &values) != nil || len(values) != 0 {
			return false
		}
	}
	return true
}
func tcLiveSource(result exploration.Result, f tcLiveFixture) bool {
	q := result.Question
	if q.Environment != f.Environment || (q.Signal != "spans" && q.Signal != "logs") || q.FromMS >= q.ToMS || q.ToMS-q.FromMS > 7200000 || q.Limit != 100 || q.Offset < 0 || q.Numeric != nil || q.ReportedEdge != nil || q.Where == nil || q.BucketMS != 0 || !peLiveHex(result.Basis) || !peLiveClock(result.CapturedAt) {
		return false
	}
	expected := tcLiveRows(q, f)
	state := "observed"
	if len(expected) == 0 {
		state = "no_data"
	}
	if result.Count != strconv.Itoa(len(expected)) || result.State != state || len(result.TimeSeries) != 0 {
		return false
	}
	from := q.Offset
	if from > len(expected) {
		from = len(expected)
	}
	to := from + q.Limit
	if to > len(expected) {
		to = len(expected)
	}
	if len(result.Rows) != to-from || result.HasMore != (to < len(expected)) {
		return false
	}
	for i, raw := range result.Rows {
		if !tcLiveRaw(raw, expected[from+i], f.Environment) {
			return false
		}
	}
	converted := make([]tpbLiveRecord, len(expected))
	for i, r := range expected {
		converted[i] = tpbLiveRecord{TraceID: r.TraceID, SpanID: r.SpanID, ParentSpanID: r.ParentSpanID, StartNS: r.StartNS, DurationNS: r.DurationNS, Service: r.Service, Operation: r.Operation, Attributes: r.Attributes, Resource: r.Resource}
	}
	return tpbLiveAggregates(q, result.Duration, result.SpanStatus, result.Groups, converted)
}

type tcLiveSideInput struct {
	Rows      []json.RawMessage
	Truncated bool
}
type tcLiveField struct {
	State string
	Value *string `json:"value,omitempty"`
}
type tcLiveView struct {
	Input struct{ Baseline, Selection tcLiveSideInput }
	Model exploration.TracePageComparison
	Cards []struct {
		Index                   int
		MatchState, Label, Text string
		Sides                   []struct {
			Side               string
			RowIndex           *string `json:"rowIndex"`
			TrackWidth         float64 `json:"trackWidth"`
			Left, Width, Title *string
			Text               string
		}
	}
	Differences []struct {
		Field  string
		Values []string
		Fields []tcLiveField
	}
}
type tcLiveObservation struct {
	Connected, Visible     bool
	ComparisonCount        int    `json:"comparisonCount"`
	DifferenceRows         int    `json:"differenceRows"`
	PinDisabled            bool   `json:"pinDisabled"`
	CompareDisabled        bool   `json:"compareDisabled"`
	ReferenceText          string `json:"referenceText"`
	RetainedPairConnected  bool   `json:"retainedPairConnected"`
	RetainedDifferenceRows int    `json:"retainedDifferenceRows"`
}
type tcLiveEvent struct {
	Action string
	Source int
	tcLiveObservation
}
type tcLiveNetwork struct {
	Endpoint, Method string
	Payload          json.RawMessage
}
type tcLiveMain struct {
	tcLiveFixture
	CaptureRequest struct {
		Title    string
		Question exploration.CaptureQuestion
	}
	Saved, Reopened exploration.CapturePage
	Custody         struct {
		ID     string
		Bytes  int
		SHA256 string
	}
	ReferenceSearch, Reference, CurrentSearch, Current exploration.Result
	ReferenceContext, ComparisonContext                string
	View, MobileView                                   tcLiveView
	Events                                             []tcLiveEvent
	Requests                                           []tpbLiveRequest
	Responses                                          []tpbLiveResponse
	Network                                            []tcLiveNetwork
	SourceBytes, AfterBytes                            []byte `json:"-"`
}
type tcLivePartial struct {
	tcLiveFixture
	ReferenceSearch, Reference, CurrentSearch, Current, PartialSearch, Partial exploration.Result
	Problem, PartialView                                                       tcLiveView
	Events                                                                     []tcLiveEvent
	Requests                                                                   []tpbLiveRequest
	Responses                                                                  []tpbLiveResponse
	Network                                                                    []tcLiveNetwork
}
type tcLiveOwnership struct {
	tcLiveFixture
	Reference, Current, ReturnedPage, Partial, HeldResponse, FreshSearch, Fresh exploration.Result
	View, Returned, Recovered                                                   tcLiveView
	BeforeURL, AfterURL, StagedURL, DirtyEnvironment                            string
	HeldRequest                                                                 exploration.Query
	Observations                                                                struct{ ClosedPair, Hidden, Pending, Stale tcLiveObservation }
	Events                                                                      []tcLiveEvent
	Requests                                                                    []tpbLiveRequest
	Responses                                                                   []tpbLiveResponse
	Network                                                                     []tcLiveNetwork
}

func tcLiveSearch(q exploration.Query, tr tcLiveTrace, f tcLiveFixture, exactWindow bool) bool {
	if exactWindow && (q.FromMS != tr.FromMS || q.ToMS != tr.ToMS) {
		return false
	}
	if !exactWindow && q.ToMS-q.FromMS != 60000 {
		return false
	}
	want := exploration.Query{Signal: "spans", Environment: f.Environment, FromMS: q.FromMS, ToMS: q.ToMS, Limit: 100, Where: q.Where}
	return tcLiveOnlyTrace(q.Where, tr.TraceID) && nhLiveSameQuery(q, want)
}
func tcLiveViewOK(view tcLiveView, a, b exploration.Result) bool {
	at, bt := a.HasMore || a.Question.Offset > 0, b.HasMore || b.Question.Offset > 0
	if view.Input.Baseline.Truncated != at || view.Input.Selection.Truncated != bt || !aeJSONEqual(view.Input.Baseline.Rows, a.Rows) || !aeJSONEqual(view.Input.Selection.Rows, b.Rows) || !tcOracleEqual(tcOracle(a.Rows, b.Rows, at, bt), view.Model) || len(view.Cards) != len(view.Model.Groups) {
		return false
	}
	// Derive a common elapsed page extent. This is placement geometry, not
	// a sum of group work and not inferred wall time for a complete trace.
	extent := new(big.Int)
	for _, side := range []exploration.TraceComparisonSide{view.Model.Baseline, view.Model.Selection} {
		if side.ExtentNS != nil {
			n, _ := new(big.Int).SetString(*side.ExtentNS, 10)
			if n.Cmp(extent) > 0 {
				extent = n
			}
		}
	}
	label := func(v *string, kind string) string {
		if v == nil {
			return kind + " not reported"
		}
		if *v == "" {
			return "Present empty " + strings.ToLower(kind)
		}
		raw, _ := json.Marshal(*v)
		return string(raw)
	}
	for i, card := range view.Cards {
		g := view.Model.Groups[i]
		if card.Index != i || card.MatchState != g.MatchState || card.Label != label(g.Service, "Service")+" / "+label(g.Operation, "Operation") || len(card.Sides) != 2 {
			return false
		}
		if math.Abs(card.Sides[0].TrackWidth-card.Sides[1].TrackWidth) > 1 {
			return false
		}
		for j, shown := range card.Sides {
			sideName := "baseline"
			page := a
			side := view.Model.Baseline
			if j == 1 {
				sideName = "selection"
				page = b
				side = view.Model.Selection
			}
			if shown.Side != sideName || shown.TrackWidth <= 0 {
				return false
			}
			var start, end *big.Int
			index := -1
			if g.Match != nil {
				index = g.Match.BaselineIndex
				if j == 1 {
					index = g.Match.SelectionIndex
				}
				if index < 0 || index >= len(page.Rows) {
					return false
				}
				start, end = tcInterval(page.Rows[index])
			}
			if g.Match == nil || start == nil || side.StartNS == nil {
				if shown.RowIndex != nil || shown.Left != nil || shown.Width != nil || shown.Title != nil {
					return false
				}
				continue
			}
			origin, _ := new(big.Int).SetString(*side.StartNS, 10)
			relative := new(big.Int).Sub(start, origin)
			duration := new(big.Int).Sub(end, start)
			if relative.Sign() < 0 || new(big.Int).Add(relative, duration).Cmp(extent) > 0 {
				return false
			}
			if shown.RowIndex == nil || *shown.RowIndex != strconv.Itoa(index) || shown.Left == nil || shown.Width == nil || shown.Title == nil || *shown.Title != "Original page row "+strconv.Itoa(index+1)+" · "+duration.String()+" ns (exact)" {
				return false
			}
			geometry := func(value *big.Int) float64 {
				if extent.Sign() == 0 {
					return 0
				}
				hundredths := new(big.Int).Quo(new(big.Int).Mul(value, big.NewInt(1000000)), extent)
				return float64(hundredths.Int64()) / 10000
			}
			for k, text := range []*string{shown.Left, shown.Width} {
				actual, e := strconv.ParseFloat(strings.TrimSuffix(*text, "%"), 64)
				expected := geometry(relative)
				if k == 1 {
					expected = geometry(duration)
				}
				if e != nil || !strings.HasSuffix(*text, "%") || math.Abs(actual-expected) > 0.00001 {
					return false
				}
			}
		}
	}
	return true
}
func tcLiveDifferences(view tcLiveView, operation string) bool {
	var group *exploration.TraceComparisonGroup
	for i, g := range view.Model.Groups {
		if g.Operation != nil && *g.Operation == operation {
			group = &view.Model.Groups[i]
			break
		}
	}
	if group == nil || group.Match == nil {
		return false
	}
	a, b := view.Input.Baseline.Rows[group.Match.BaselineIndex], view.Input.Selection.Rows[group.Match.SelectionIndex]
	extract := func(raw json.RawMessage) map[string]string {
		var obj struct{ Attributes, ResourceAttributes map[string]string }
		_ = json.Unmarshal(raw, &obj)
		out := map[string]string{}
		for kind, values := range map[string]map[string]string{"attribute": obj.Attributes, "resource": obj.ResourceAttributes} {
			for k, v := range values {
				out[kind+":"+k] = v
			}
		}
		return out
	}
	left, right := extract(a), extract(b)
	keys := map[string]bool{}
	for k := range left {
		keys[k] = true
	}
	for k := range right {
		keys[k] = true
	}
	ordered := []string{}
	for k := range keys {
		av, aok := left[k]
		bv, bok := right[k]
		if aok != bok || av != bv {
			ordered = append(ordered, k)
		}
	}
	sort.Strings(ordered)
	if len(view.Differences) != len(ordered) {
		return false
	}
	for i, key := range ordered {
		row := view.Differences[i]
		if row.Field != key || len(row.Fields) != 2 || len(row.Values) != 2 {
			return false
		}
		for j, values := range []map[string]string{left, right} {
			value, present := values[key]
			field := row.Fields[j]
			if present {
				if field.State != "present" || field.Value == nil || *field.Value != value {
					return false
				}
				if value == "" {
					if !strings.Contains(row.Values[j], "Empty string") {
						return false
					}
				} else if !strings.Contains(row.Values[j], value) {
					return false
				}
			} else if field.State != "missing" || field.Value != nil || !strings.Contains(row.Values[j], "Missing value") {
				return false
			}
		}
	}
	return true
}

func tcLiveOnlyTrace(p *exploration.Predicate, id string) bool {
	if p == nil || p.Numeric != nil || len(p.Any) > 0 {
		return false
	}
	if len(p.All) > 0 {
		return len(p.All) == 1 && p.Field == nil && p.Op == "" && p.Value == "" && tcLiveOnlyTrace(&p.All[0], id)
	}
	return p.Field != nil && *p.Field == (exploration.Field{Kind: "column", Key: "trace_id"}) && p.Op == "eq" && p.Value == id
}
func tcLiveExpanded(search, page exploration.Query, trace string) bool {
	want := exploration.Query{Signal: "spans", Environment: search.Environment, FromMS: search.FromMS, ToMS: search.ToMS, Limit: 100, Offset: page.Offset, Where: &exploration.Predicate{All: []exploration.Predicate{{Field: &exploration.Field{Kind: "column", Key: "trace_id"}, Op: "eq", Value: trace}}}, GroupBy: []exploration.Field{{Kind: "column", Key: "service"}, {Kind: "column", Key: "name"}}, Summary: "span-duration", SpanStatus: true}
	return nhLiveSameQuery(want, page)
}

func tcLiveLogQuestion(page exploration.Query) exploration.Query {
	return exploration.Query{Signal: "logs", Environment: page.Environment, FromMS: page.FromMS, ToMS: page.ToMS, Limit: 100, Where: page.Where}
}
func tcLiveProtocol(f tcLiveFixture, requests []tpbLiveRequest, responses []tpbLiveResponse, network []tcLiveNetwork, expected []exploration.Query) bool {
	if len(requests) != len(expected) || len(responses) != len(expected) {
		return false
	}
	for i, r := range requests {
		var q exploration.Query
		if r.Endpoint != "/api/explore" || r.Effect != "source" || r.Ordinal != i+1 || json.Unmarshal(r.Payload, &q) != nil || !nhLiveSameQuery(q, expected[i]) || !nhLiveSameQuery(responses[i].Question, q) || !nhLiveSameQuery(responses[i].Result.Question, q) || !tcLiveSource(responses[i].Result, f) {
			return false
		}
	}
	posts := []tcLiveNetwork{}
	for _, n := range network {
		if n.Method == "POST" {
			posts = append(posts, n)
		} else if n.Method != "GET" || strings.HasPrefix(n.Endpoint, "/api/tail") || n.Endpoint == "/api/explore" {
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
func tcLiveSourcePair(f tcLiveFixture, search, page exploration.Result, traceIndex int, exactWindow bool) bool {
	return traceIndex >= 0 && traceIndex < len(f.Traces) && tcLiveSearch(search.Question, f.Traces[traceIndex], f, exactWindow) && tcLiveSource(search, f) && tcLiveSource(page, f) && tcLiveExpanded(search.Question, page.Question, f.Traces[traceIndex].TraceID) && page.Question.Offset == 0
}
func tcLiveEvents(events []tcLiveEvent, names []string, counts []int) bool {
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
func tcLiveProvenance(text string, page exploration.Result) bool {
	if len(page.Rows) == 0 {
		return false
	}
	obj := tcObject(page.Rows[0])
	trace, _ := obj["TraceId"].(string)
	format := func(ms int64) string { return time.UnixMilli(ms).UTC().Format("2006-01-02T15:04:05.000Z") }
	for _, value := range []string{trace, page.Question.Environment, format(page.Question.FromMS), format(page.Question.ToMS), page.CapturedAt, page.Basis, strconv.Itoa(len(page.Rows)) + " supplied rows"} {
		if !strings.Contains(text, value) {
			return false
		}
	}
	return true
}
func tcLiveMainScope(l tcLiveMain) bool {
	f := l.tcLiveFixture
	if !tcLiveFixtureOK(f) || !tcLiveSourcePair(f, l.ReferenceSearch, l.Reference, 0, true) || !tcLiveSourcePair(f, l.CurrentSearch, l.Current, 1, false) || l.Reference.Count != "7" || l.Current.Count != "7" {
		return false
	}
	if !tcLiveViewOK(l.View, l.Reference, l.Current) || !tcLiveViewOK(l.MobileView, l.Reference, l.Current) || !tcLiveDifferences(l.View, "GET /checkout") || len(l.MobileView.Differences) != 0 || !aeJSONEqual(l.View.Input, l.MobileView.Input) || !aeJSONEqual(l.View.Model, l.MobileView.Model) {
		return false
	}
	if !tcLiveProvenance(l.ReferenceContext, l.Reference) || !tcLiveProvenance(l.ComparisonContext, l.Reference) || !tcLiveProvenance(l.ComparisonContext, l.Current) {
		return false
	}
	expected := []exploration.Query{l.ReferenceSearch.Question, l.Reference.Question, tcLiveLogQuestion(l.Reference.Question), l.CurrentSearch.Question, l.Current.Question, tcLiveLogQuestion(l.Current.Question)}
	if !tcLiveProtocol(f, l.Requests, l.Responses, l.Network, expected) {
		return false
	}
	for _, pair := range []struct{ a, b exploration.Result }{{l.Responses[0].Result, l.ReferenceSearch}, {l.Responses[1].Result, l.Reference}, {l.Responses[3].Result, l.CurrentSearch}, {l.Responses[4].Result, l.Current}} {
		if !aeJSONEqual(pair.a, pair.b) {
			return false
		}
	}
	return tcLiveEvents(l.Events, []string{"reference-search", "reference-expand", "pin", "current-search", "current-expand", "compare", "pair-attributes", "capture-reopen"}, []int{1, 3, 3, 4, 6, 6, 6, 6})
}
func tcLiveCustody(l tcLiveMain) bool {
	if len(l.SourceBytes) != l.Custody.Bytes || !bytes.Equal(l.SourceBytes, l.AfterBytes) || l.Saved.Receipt.ID != l.Custody.ID || !aeJSONEqual(l.Saved, l.Reopened) || l.Saved.Explore == nil || l.Saved.RetainedRecords != 7 || l.Saved.RecordsLimited {
		return false
	}
	hash := sha256.Sum256(l.SourceBytes)
	if hex.EncodeToString(hash[:]) != l.Custody.SHA256 {
		return false
	}
	var stored exploration.SavedCapture
	if json.Unmarshal(l.SourceBytes, &stored) != nil || stored.Explore == nil || stored.ID != l.Custody.ID || !aeJSONEqual(stored.CaptureReceipt, l.Saved.Receipt) || !aeJSONEqual(stored.Explore, l.Saved.Explore) {
		return false
	}
	page := exploration.CapturePage{Receipt: stored.CaptureReceipt, Explore: stored.Explore}
	if !peLiveCaptureHash(page) || !tcLiveSource(*stored.Explore, l.tcLiveFixture) || stored.Question.Explore == nil || !nhLiveSameQuery(*stored.Question.Explore, l.ReferenceSearch.Question) || !nhLiveSameQuery(stored.Explore.Question, l.ReferenceSearch.Question) {
		return false
	}
	if l.CaptureRequest.Title != stored.Title || !aeJSONEqual(l.CaptureRequest.Question, stored.Question) {
		return false
	}
	custody, ce := time.Parse(time.RFC3339Nano, stored.CapturedAt)
	source, se := time.Parse(time.RFC3339Nano, stored.Explore.CapturedAt)
	reference, re := time.Parse(time.RFC3339Nano, l.ReferenceSearch.CapturedAt)
	return ce == nil && se == nil && re == nil && !custody.Before(source) && reference.After(source)
}
func tcLiveMainOK(l tcLiveMain) bool { return tcLiveMainScope(l) && tcLiveCustody(l) }
func tcLivePartialOK(l tcLivePartial) bool {
	f := l.tcLiveFixture
	if !tcLiveFixtureOK(f) || !tcLiveSourcePair(f, l.ReferenceSearch, l.Reference, 2, true) || !tcLiveSourcePair(f, l.CurrentSearch, l.Current, 3, false) || !tcLiveSourcePair(f, l.PartialSearch, l.Partial, 4, false) || l.Partial.Count != "126" || len(l.Partial.Rows) != 100 || !l.Partial.HasMore {
		return false
	}
	if !tcLiveViewOK(l.Problem, l.Reference, l.Current) || !tcLiveViewOK(l.PartialView, l.Reference, l.Partial) || len(l.Problem.Differences) != 0 || len(l.PartialView.Differences) != 0 {
		return false
	}
	expected := []exploration.Query{}
	for _, pair := range []struct{ search, page exploration.Result }{{l.ReferenceSearch, l.Reference}, {l.CurrentSearch, l.Current}, {l.PartialSearch, l.Partial}} {
		expected = append(expected, pair.search.Question, pair.page.Question, tcLiveLogQuestion(pair.page.Question))
	}
	if !tcLiveProtocol(f, l.Requests, l.Responses, l.Network, expected) {
		return false
	}
	return tcLiveEvents(l.Events, []string{"problem-pin", "problem-compare", "partial-compare"}, []int{3, 6, 9})
}

func tcLiveObservationOK(o tcLiveObservation, reference exploration.Result, closed, hidden bool) bool {
	if !o.Connected || o.DifferenceRows != 0 || o.RetainedDifferenceRows != 0 || !tcLiveProvenance(o.ReferenceText, reference) {
		return false
	}
	if closed {
		return o.Visible && o.ComparisonCount == 1 && !o.PinDisabled && !o.CompareDisabled && o.RetainedPairConnected
	}
	if hidden {
		return !o.Visible && o.ComparisonCount == 1 && !o.PinDisabled && !o.CompareDisabled && o.RetainedPairConnected
	}
	return o.ComparisonCount == 0 && o.PinDisabled && o.CompareDisabled && !o.RetainedPairConnected
}
func tcLiveOwnershipOK(l tcLiveOwnership) bool {
	f := l.tcLiveFixture
	if !tcLiveFixtureOK(f) || !tcLiveSource(l.Reference, f) || !tcLiveSource(l.Current, f) || !tcLiveSource(l.ReturnedPage, f) || !tcLiveSource(l.Partial, f) || !tcLiveSource(l.HeldResponse, f) || !tcLiveSourcePair(f, l.FreshSearch, l.Fresh, 4, false) {
		return false
	}
	if l.BeforeURL != l.AfterURL || !strings.HasSuffix(l.BeforeURL, "#definitions") || l.DirtyEnvironment != f.Environment+"-draft" || !nhLiveSameQuery(l.Current.Question, l.ReturnedPage.Question) || !nhLiveSameQuery(l.Partial.Question, l.Fresh.Question) {
		return false
	}
	held := l.Partial.Question
	held.Offset = 100
	if !nhLiveSameQuery(held, l.HeldRequest) || !nhLiveSameQuery(l.HeldRequest, l.HeldResponse.Question) || l.HeldResponse.Count != "126" || len(l.HeldResponse.Rows) != 26 {
		return false
	}
	if !tcLiveViewOK(l.View, l.Reference, l.Current) || !tcLiveViewOK(l.Returned, l.Reference, l.ReturnedPage) || !tcLiveViewOK(l.Recovered, l.Reference, l.Fresh) || len(l.View.Differences) != 0 || len(l.Returned.Differences) != 0 || len(l.Recovered.Differences) != 0 {
		return false
	}
	if !tcLiveObservationOK(l.Observations.ClosedPair, l.Reference, true, false) || !tcLiveObservationOK(l.Observations.Hidden, l.Reference, false, true) || !tcLiveObservationOK(l.Observations.Pending, l.Reference, false, false) || !tcLiveObservationOK(l.Observations.Stale, l.Reference, false, false) {
		return false
	}
	if len(l.Requests) != 15 || len(l.Responses) != 15 {
		return false
	}
	// First search and natural scope-switch queries are retained in the actual
	// request stream. Verify their declared scopes independently, then bind every
	// returned page/read to this precise stream (including the held real reply).
	queryAt := func(index int) exploration.Query {
		var q exploration.Query
		_ = json.Unmarshal(l.Requests[index].Payload, &q)
		return q
	}
	refSearch, curSearch, partialSearch := queryAt(0), queryAt(3), queryAt(8)
	if !tcLiveSearch(refSearch, f.Traces[0], f, true) || !tcLiveSearch(curSearch, f.Traces[1], f, false) || !tcLiveSearch(partialSearch, f.Traces[4], f, false) || !tcLiveExpanded(refSearch, l.Reference.Question, f.Traces[0].TraceID) || !tcLiveExpanded(curSearch, l.Current.Question, f.Traces[1].TraceID) || !tcLiveExpanded(partialSearch, l.Partial.Question, f.Traces[4].TraceID) || !nhLiveSameQuery(partialSearch, l.FreshSearch.Question) {
		return false
	}
	expected := []exploration.Query{refSearch, l.Reference.Question, tcLiveLogQuestion(l.Reference.Question), curSearch, l.Current.Question, tcLiveLogQuestion(l.Current.Question), l.ReturnedPage.Question, tcLiveLogQuestion(l.ReturnedPage.Question), partialSearch, l.Partial.Question, tcLiveLogQuestion(l.Partial.Question), l.HeldRequest, l.FreshSearch.Question, l.Fresh.Question, tcLiveLogQuestion(l.Fresh.Question)}
	if !tcLiveProtocol(f, l.Requests, l.Responses, l.Network, expected) {
		return false
	}
	for _, pair := range []struct {
		index  int
		result exploration.Result
	}{{1, l.Reference}, {4, l.Current}, {6, l.ReturnedPage}, {9, l.Partial}, {11, l.HeldResponse}, {12, l.FreshSearch}, {13, l.Fresh}} {
		if !aeJSONEqual(l.Responses[pair.index].Result, pair.result) {
			return false
		}
	}
	times := []string{l.Current.CapturedAt, l.ReturnedPage.CapturedAt, l.Partial.CapturedAt, l.HeldResponse.CapturedAt, l.FreshSearch.CapturedAt, l.Fresh.CapturedAt}
	for i := 1; i < len(times); i++ {
		a, ae := time.Parse(time.RFC3339Nano, times[i-1])
		b, be := time.Parse(time.RFC3339Nano, times[i])
		if ae != nil || be != nil || !b.After(a) {
			return false
		}
	}
	names := []string{"pair", "closed-pair", "hidden-actions", "return-stage", "returned-compare", "pending-actions", "stale-response", "fresh-accepted"}
	counts := []int{6, 6, 6, 6, 8, 12, 12, 15}
	if !tcLiveEvents(l.Events, names, counts) {
		return false
	}
	for _, pair := range []struct {
		index       int
		observation tcLiveObservation
	}{{1, l.Observations.ClosedPair}, {2, l.Observations.Hidden}, {5, l.Observations.Pending}, {6, l.Observations.Stale}} {
		if !aeJSONEqual(l.Events[pair.index].tcLiveObservation, pair.observation) {
			return false
		}
	}
	return true
}

func tcLiveReplay(t *testing.T, events []tcLiveEvent, model string) bool {
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
func tcLiveRejectTraces(t *testing.T, events []tcLiveEvent, model string) {
	for _, kind := range []string{"unexpected source effect", "wrong action order"} {
		t.Run("finite trace rejects "+kind, func(t *testing.T) {
			bad := append([]tcLiveEvent{}, events...)
			if kind == "unexpected source effect" {
				bad[0].Source++
			} else {
				bad[0].Action = bad[1].Action
			}
			tpbLiveChanged(t, events, bad)
			if tcLiveReplay(t, bad, model) {
				t.Fatal("finite model accepted changed trace")
			}
		})
	}
}
func tcLiveOperation(view *tcLiveView, operation string) *exploration.TraceComparisonGroup {
	for i, g := range view.Model.Groups {
		if g.Operation != nil && *g.Operation == operation {
			return &view.Model.Groups[i]
		}
	}
	return nil
}
func tcLiveCard(view *tcLiveView, operation string) int {
	for i, g := range view.Model.Groups {
		if g.Operation != nil && *g.Operation == operation {
			return i
		}
	}
	return -1
}
func TestQuintTraceComparisonLiveLedger(t *testing.T) {
	var base tcLiveMain
	file := tpbLiveRead(t, "GOO_TRACE_COMPARISON_LEDGER", &base)
	var e error
	base.SourceBytes, e = os.ReadFile(filepath.Join(filepath.Dir(file), "scope-capture.json"))
	if e != nil {
		t.Fatal(e)
	}
	base.AfterBytes, e = os.ReadFile(filepath.Join(filepath.Dir(file), "scope-capture.after.json"))
	if e != nil {
		t.Fatal(e)
	}
	if !tcLiveMainOK(base) {
		t.Fatalf("independent trace comparison rejected scope=%v custody=%v reference=%v current=%v modelDOM=%v differences=%v", tcLiveMainScope(base), tcLiveCustody(base), tcLiveSource(base.Reference, base.tcLiveFixture), tcLiveSource(base.Current, base.tcLiveFixture), tcLiveViewOK(base.View, base.Reference, base.Current), tcLiveDifferences(base.View, "GET /checkout"))
	}
	if !tcLiveReplay(t, base.Events, "model/trace_comparison.yaml") {
		t.Fatal("actual trace comparison lifecycle rejected")
	}
	tcLiveRejectTraces(t, base.Events, "model/trace_comparison.yaml")
	controls := []struct {
		name   string
		mutate func(*tcLiveMain)
	}{
		{"signed root difference", func(l *tcLiveMain) {
			tcLiveOperation(&l.View, "GET /checkout").InclusiveDifferenceNS = lpcPointer("-30000000")
		}},
		{"negative difference sign lost", func(l *tcLiveMain) { tcLiveOperation(&l.View, "lookup").InclusiveDifferenceNS = lpcPointer("10000000") }},
		{"outside union delta", func(l *tcLiveMain) {
			tcLiveOperation(&l.View, "GET /checkout").OutsideDifferenceNS = lpcPointer("30000000")
		}},
		{"arbitrary repeated pairing", func(l *tcLiveMain) {
			g := tcLiveOperation(&l.View, "prepare")
			g.MatchState = "unique-label-match"
			g.Match = &exploration.TraceRowMatch{BaselineIndex: 2, SelectionIndex: 2}
		}},
		{"different kind paired", func(l *tcLiveMain) {
			g := tcLiveOperation(&l.View, "dispatch")
			g.MatchState = "unique-label-match"
			g.Match = &exploration.TraceRowMatch{BaselineIndex: 6, SelectionIndex: 6}
		}},
		{"absent group invented zero", func(l *tcLiveMain) {
			tcLiveOperation(&l.View, "cache lookup").InclusiveDifferenceNS = lpcPointer("-10000000")
		}},
		{"known zero made unavailable", func(l *tcLiveMain) { tcLiveOperation(&l.View, "marker").Match.InclusiveDifferenceNS = nil }},
		{"source count mismatch", func(l *tcLiveMain) { l.Current.Count = "100" }},
		{"raw duration", func(l *tcLiveMain) { lpLiveMutateRow(&l.Current.Rows[0], "duration_ns", "130000001") }},
		{"raw timestamp rounded", func(l *tcLiveMain) { lpLiveMutateRow(&l.Reference.Rows[0], "timestamp_ns", "9007199254740992") }},
		{"attribute namespace raw", func(l *tcLiveMain) {
			var o map[string]any
			_ = json.Unmarshal(l.Current.Rows[0], &o)
			o["Attributes"].(map[string]any)["work.namespace"] = "resource-new"
			l.Current.Rows[0], _ = json.Marshal(o)
		}},
		{"resource namespace raw", func(l *tcLiveMain) {
			var o map[string]any
			_ = json.Unmarshal(l.Reference.Rows[0], &o)
			o["ResourceAttributes"].(map[string]any)["work.namespace"] = "attribute-old"
			l.Reference.Rows[0], _ = json.Marshal(o)
		}},
		{"duplicate raw page row", func(l *tcLiveMain) { l.Reference.Rows[1] = l.Reference.Rows[0] }},
		{"scope window changed", func(l *tcLiveMain) { l.Current.Question.FromMS++ }},
		{"elapsed extent rounded", func(l *tcLiveMain) { l.View.Model.Selection.ExtentNS = lpcPointer("129999999") }},
		{"group knownness", func(l *tcLiveMain) { l.View.Model.Baseline.Breakdown.Groups[0].OutsideUnknown++ }},
		{"row index changed", func(l *tcLiveMain) { tcLiveOperation(&l.View, "GET /checkout").Match.BaselineIndex = 1 }},
		{"DOM match state", func(l *tcLiveMain) { l.View.Cards[0].MatchState = "unique-label-match" }},
		{"shared elapsed bar width", func(l *tcLiveMain) {
			i := tcLiveCard(&l.View, "GET /checkout")
			l.View.Cards[i].Sides[0].Width = lpcPointer("100%")
		}},
		{"relative bar origin", func(l *tcLiveMain) {
			i := tcLiveCard(&l.View, "GET /checkout")
			l.View.Cards[i].Sides[0].Left = lpcPointer("1%")
		}},
		{"exact row title", func(l *tcLiveMain) {
			i := tcLiveCard(&l.View, "GET /checkout")
			l.View.Cards[i].Sides[0].Title = lpcPointer("100 ms approximate")
		}},
		{"separate physical tracks", func(l *tcLiveMain) { l.View.Cards[0].Sides[0].TrackWidth += 10 }},
		{"qualified field namespace", func(l *tcLiveMain) { l.View.Differences[0].Field = "resource:work.added" }},
		{"empty confused with missing", func(l *tcLiveMain) {
			for i, row := range l.View.Differences {
				if row.Field == "attribute:work.empty" {
					l.View.Differences[i].Fields[0].State = "missing"
					l.View.Differences[i].Fields[0].Value = nil
					return
				}
			}
		}},
		{"attribute difference value", func(l *tcLiveMain) { l.View.Differences[0].Fields[1].Value = lpcPointer("wrong") }},
		{"unchanged field reported", func(l *tcLiveMain) {
			row := l.View.Differences[0]
			row.Field = "attribute:work.same"
			l.View.Differences = append(l.View.Differences, row)
		}},
		{"mobile model mutation", func(l *tcLiveMain) { l.MobileView.Model.Selection.StartNS = lpcPointer("0") }},
		{"pin hidden read", func(l *tcLiveMain) { l.Events[2].Source++ }},
		{"comparison hidden read", func(l *tcLiveMain) { l.Events[5].Source++ }},
		{"pair inspection hidden read", func(l *tcLiveMain) { l.Events[6].Source++ }},
		{"reopen hidden read", func(l *tcLiveMain) { l.Events[7].Source++ }},
		{"unexpected comparison API", func(l *tcLiveMain) { l.Requests[0].Endpoint = "/api/compare" }},
		{"response full count changed", func(l *tcLiveMain) { l.Responses[1].Result.Count = "100" }},
		{"custody hash", func(l *tcLiveMain) { l.Custody.SHA256 = strings.Repeat("0", 64) }},
		{"capture mutation", func(l *tcLiveMain) { l.Saved.Explore.Count = "999" }},
		{"reopened receipt mutation", func(l *tcLiveMain) { l.Reopened.Receipt.Title = "changed" }},
	}
	for _, control := range controls {
		t.Run(control.name, func(t *testing.T) {
			bad := tpbLiveClone(base)
			control.mutate(&bad)
			tpbLiveChanged(t, base, bad)
			bad.SourceBytes = base.SourceBytes
			bad.AfterBytes = base.AfterBytes
			if tcLiveMainOK(bad) {
				t.Fatal("changed evidence accepted")
			}
		})
	}
}
func TestQuintTraceComparisonPartialLedger(t *testing.T) {
	var base tcLivePartial
	tpbLiveRead(t, "GOO_TRACE_COMPARISON_PARTIAL_LEDGER", &base)
	if !tcLivePartialOK(base) {
		t.Fatalf("independent partial comparison rejected source=%v problem=%v partial=%v", tcLiveSource(base.Partial, base.tcLiveFixture), tcLiveViewOK(base.Problem, base.Reference, base.Current), tcLiveViewOK(base.PartialView, base.Reference, base.Partial))
	}
	if !tcLiveReplay(t, base.Events, "model/trace_comparison_partial.yaml") {
		t.Fatal("actual partial lifecycle rejected")
	}
	tcLiveRejectTraces(t, base.Events, "model/trace_comparison_partial.yaml")
	controls := []struct {
		name   string
		mutate func(*tcLivePartial)
	}{
		{"unknown parent outside made zero", func(l *tcLivePartial) {
			tcLiveOperation(&l.Problem, "GET /checkout").OutsideDifferenceNS = lpcPointer("0")
		}},
		{"unknown matched outside made zero", func(l *tcLivePartial) {
			tcLiveOperation(&l.Problem, "GET /checkout").Match.OutsideDifferenceNS = lpcPointer("0")
		}},
		{"child clipped to parent", func(l *tcLivePartial) {
			tcLiveOperation(&l.Problem, "lookup").Match.InclusiveDifferenceNS = lpcPointer("-10000000")
		}},
		{"parent unknown propagated to child", func(l *tcLivePartial) { tcLiveOperation(&l.Problem, "lookup").Selection.OutsideChildrenNS = nil }},
		{"missing parent flag erased", func(l *tcLivePartial) {
			for i, row := range l.Problem.Model.Selection.Breakdown.Rows {
				if row.Operation != nil && *row.Operation == "orphan" {
					l.Problem.Model.Selection.Breakdown.Rows[i].Reasons = []string{}
					return
				}
			}
		}},
		{"full count replaced by page count", func(l *tcLivePartial) { l.Partial.Count = "100" }},
		{"full group count replaced by rows", func(l *tcLivePartial) {
			var g map[string]any
			_ = json.Unmarshal(l.Partial.Groups[0], &g)
			g["count"] = "99"
			l.Partial.Groups[0], _ = json.Marshal(g)
		}},
		{"truncation erased", func(l *tcLivePartial) {
			l.PartialView.Input.Selection.Truncated = false
			l.PartialView.Model.Selection.Breakdown.Truncated = false
		}},
		{"raw input order changed", func(l *tcLivePartial) { r := l.PartialView.Input.Selection.Rows; r[0], r[1] = r[1], r[0] }},
		{"repeated fanout arbitrarily paired", func(l *tcLivePartial) {
			tcLiveOperation(&l.PartialView, "fanout").Match = &exploration.TraceRowMatch{BaselineIndex: 1, SelectionIndex: 1}
		}},
		{"stale baseline replaced", func(l *tcLivePartial) { l.PartialView.Input.Baseline.Rows = l.Current.Rows }},
		{"absent group invented delta", func(l *tcLivePartial) {
			tcLiveOperation(&l.PartialView, "fanout").InclusiveDifferenceNS = lpcPointer("198000000")
		}},
		{"duplicate raw member", func(l *tcLivePartial) { l.Partial.Rows[1] = l.Partial.Rows[0] }},
		{"raw timestamp changed", func(l *tcLivePartial) { lpLiveMutateRow(&l.Current.Rows[0], "timestamp_ns", "0") }},
		{"source read omitted", func(l *tcLivePartial) { l.Events[2].Source-- }},
		{"linked logs read omitted", func(l *tcLivePartial) { l.Requests = l.Requests[:8] }},
		{"scope window changed", func(l *tcLivePartial) { l.Partial.Question.ToMS++ }},
		{"false row alignment bar", func(l *tcLivePartial) {
			i := tcLiveCard(&l.PartialView, "fanout")
			l.PartialView.Cards[i].Sides[1].Width = lpcPointer("100%")
		}},
		{"page extent changed", func(l *tcLivePartial) { l.PartialView.Model.Selection.ExtentNS = lpcPointer("1") }},
		{"source response count", func(l *tcLivePartial) { l.Responses[7].Result.Count = "100" }},
	}
	for _, control := range controls {
		t.Run(control.name, func(t *testing.T) {
			bad := tpbLiveClone(base)
			control.mutate(&bad)
			tpbLiveChanged(t, base, bad)
			if tcLivePartialOK(bad) {
				t.Fatal("changed partial evidence accepted")
			}
		})
	}
}
func TestQuintTraceComparisonOwnershipLedger(t *testing.T) {
	var base tcLiveOwnership
	tpbLiveRead(t, "GOO_TRACE_COMPARISON_OWNERSHIP_LEDGER", &base)
	if !tcLiveOwnershipOK(base) {
		t.Fatal("independent trace comparison owner rejected")
	}
	if !tcLiveReplay(t, base.Events, "model/trace_comparison_ownership.yaml") {
		t.Fatal("actual owner lifecycle rejected")
	}
	tcLiveRejectTraces(t, base.Events, "model/trace_comparison_ownership.yaml")
	controls := []struct {
		name   string
		mutate func(*tcLiveOwnership)
	}{
		{"closed pair created differences", func(l *tcLiveOwnership) { l.Observations.ClosedPair.DifferenceRows = 1 }},
		{"hidden retained comparison changed", func(l *tcLiveOwnership) { l.Observations.Hidden.ComparisonCount = 0 }},
		{"hidden owner visible", func(l *tcLiveOwnership) { l.Observations.Hidden.Visible = true }},
		{"pending pin enabled", func(l *tcLiveOwnership) { l.Observations.Pending.PinDisabled = false }},
		{"pending compare enabled", func(l *tcLiveOwnership) { l.Observations.Pending.CompareDisabled = false }},
		{"stale comparison displayed", func(l *tcLiveOwnership) { l.Observations.Stale.ComparisonCount = 1 }},
		{"retained pair changed", func(l *tcLiveOwnership) { l.Observations.Hidden.RetainedDifferenceRows = 1 }},
		{"stale reference metadata", func(l *tcLiveOwnership) { l.Observations.Stale.ReferenceText = "Reference changed" }},
		{"stale navigation", func(l *tcLiveOwnership) { l.AfterURL += "changed" }},
		{"dirty scope not represented", func(l *tcLiveOwnership) { l.DirtyEnvironment = l.Environment }},
		{"pending source effect", func(l *tcLiveOwnership) { l.Events[5].Source++ }},
		{"stale hidden source effect", func(l *tcLiveOwnership) { l.Events[6].Source++ }},
		{"return reused stale authority", func(l *tcLiveOwnership) { l.Events[4].Source = 6 }},
		{"fresh recovery skipped reads", func(l *tcLiveOwnership) { l.Events[7].Source = 12 }},
		{"held page offset", func(l *tcLiveOwnership) { l.HeldRequest.Offset = 0 }},
		{"held source count", func(l *tcLiveOwnership) { l.HeldResponse.Count = "26" }},
		{"old source clock reused", func(l *tcLiveOwnership) { l.FreshSearch.CapturedAt = l.HeldResponse.CapturedAt }},
		{"recovered baseline mutated", func(l *tcLiveOwnership) { l.Recovered.Input.Baseline.Rows = l.Current.Rows }},
	}
	for _, control := range controls {
		t.Run(control.name, func(t *testing.T) {
			bad := tpbLiveClone(base)
			control.mutate(&bad)
			tpbLiveChanged(t, base, bad)
			if tcLiveOwnershipOK(bad) {
				t.Fatal("changed owner evidence accepted")
			}
		})
	}
}
