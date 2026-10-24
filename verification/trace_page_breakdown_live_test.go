package verification

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type tpbLiveRecord struct {
	TraceID              string `json:"traceId"`
	SpanID               string `json:"spanId"`
	ParentSpanID         string `json:"parentSpanId"`
	StartNS              string `json:"startNs"`
	DurationNS           string `json:"durationNs"`
	Service, Operation   string
	Attributes, Resource map[string]string
}
type tpbLiveTrace struct {
	Case    string
	TraceID string `json:"traceId"`
	Records []tpbLiveRecord
}
type tpbLiveFixture struct {
	Version     int
	Environment string
	FromMS      int64 `json:"fromMs"`
	ToMS        int64 `json:"toMs"`
	Traces      []tpbLiveTrace
}
type tpbLiveView struct {
	Input struct {
		Rows      []json.RawMessage
		Truncated bool
	}
	Model              exploration.TracePageBreakdown
	SelectedIndex      int `json:"selectedIndex"`
	SelectedGroupIndex int `json:"selectedGroupIndex"`
	Groups             []struct {
		Key     [2]*string
		Indexes []int
		Values  []struct {
			Measure, Value, Known, Unknown string
			Bar                            *tpbLiveBar `json:"bar,omitempty"`
		}
	}
	Diagnostics []struct {
		RowIndex int `json:"rowIndex"`
		State    string
		Reasons  []string
		Values   []struct{ Measure, Value string }
	}
}
type tpbLiveEvent struct {
	Action             string
	Source             int
	Connected          bool
	SelectedIndex      int `json:"selectedIndex"`
	SelectedGroupIndex int `json:"selectedGroupIndex"`
	Highlights         int
}
type tpbLiveRequest struct {
	Endpoint, Effect string
	Ordinal          int
	Payload          json.RawMessage
}
type tpbLiveResponse struct {
	Question exploration.Query
	Result   exploration.Result
}
type tpbLiveMain struct {
	tpbLiveFixture
	Original, TracePage, SelectedLogs exploration.Result
	View, Highlighted                 tpbLiveView
	DesktopGeometry, MobileGeometry   tpbLiveGeometry
	Saved, Reopened                   exploration.CapturePage
	Events                            []tpbLiveEvent
	Requests                          []tpbLiveRequest
	Responses                         []tpbLiveResponse
	Custody                           struct {
		ID     string
		Bytes  int
		SHA256 string
	}
	SourceBytes, AfterBytes []byte `json:"-"`
}
type tpbLiveGraphs struct {
	tpbLiveFixture
	Cases []struct {
		Case                string
		Original, TracePage exploration.Result
		View                tpbLiveView
		Next                *exploration.Result
		NextView            *tpbLiveView
	}
	Events   []tpbLiveEvent
	Requests []tpbLiveRequest
}
type tpbLiveOwnership struct {
	tpbLiveFixture
	TracePage exploration.Result
	Initial   tpbLiveView
	Hidden    struct {
		Connected          bool
		SelectedIndex      int `json:"selectedIndex"`
		SelectedGroupIndex int `json:"selectedGroupIndex"`
		Highlights         int
	}
	BeforeURL, AfterURL string
	Events              []tpbLiveEvent
	Requests            []tpbLiveRequest
}

func tpbLiveFixtureOK(f tpbLiveFixture) bool {
	if f.Version != 1 || f.Environment == "" || f.FromMS >= f.ToMS || len(f.Traces) != 7 {
		return false
	}
	names := []string{"overlap", "missing-parent", "outside-parent", "grandchild-outside-child", "duplicate-identity", "parent-cycle", "bounded-page"}
	sizes := []int{5, 1, 2, 3, 3, 2, 126}
	ids := map[string]bool{}
	for i, tr := range f.Traces {
		if tr.Case != names[i] || len(tr.Records) != sizes[i] || len(tr.TraceID) != 32 || ids[tr.TraceID] {
			return false
		}
		ids[tr.TraceID] = true
		for _, r := range tr.Records {
			start, ok := new(big.Int).SetString(r.StartNS, 10)
			duration, dok := new(big.Int).SetString(r.DurationNS, 10)
			if !ok || !dok || start.Sign() < 0 || duration.Sign() < 0 || start.BitLen() > 64 || duration.BitLen() > 64 || r.TraceID != tr.TraceID || r.Service == "" || r.Operation == "" || r.Resource["service.namespace"] != "commerce" || r.Attributes["work.case"] != tr.Case {
				return false
			}
			ms := new(big.Int).Quo(start, big.NewInt(1000000))
			if !ms.IsInt64() || ms.Int64() < f.FromMS || ms.Int64() >= f.ToMS {
				return false
			}
		}
	}
	return true
}
func tpbLiveMatch(p *exploration.Predicate, r tpbLiveRecord) bool {
	if p == nil {
		return true
	}
	if len(p.All) > 0 {
		for _, x := range p.All {
			if !tpbLiveMatch(&x, r) {
				return false
			}
		}
		return true
	}
	if len(p.Any) > 0 {
		for _, x := range p.Any {
			if tpbLiveMatch(&x, r) {
				return true
			}
		}
		return false
	}
	if p.Field == nil || p.Field.Kind != "column" || p.Op != "eq" {
		return false
	}
	if p.Field.Key == "trace_id" {
		return r.TraceID == p.Value
	}
	if p.Field.Key == "span_id" {
		return r.SpanID == p.Value
	}
	return false
}
func tpbLiveMatching(q exploration.Query, f tpbLiveFixture) []tpbLiveRecord {
	rows := []tpbLiveRecord{}
	if q.Signal != "spans" {
		return rows
	}
	for _, tr := range f.Traces {
		for _, r := range tr.Records {
			n, _ := new(big.Int).SetString(r.StartNS, 10)
			ms := new(big.Int).Quo(n, big.NewInt(1000000)).Int64()
			if ms >= q.FromMS && ms < q.ToMS && tpbLiveMatch(q.Where, r) {
				rows = append(rows, r)
			}
		}
	}
	// The source orders timestamp first, then trace/span/service/attributes/name/duration.
	// Equal timestamp rows in these fixtures have distinct span IDs except identical
	// repeats whose full records are checked by a multiset below.
	sort.SliceStable(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		x, _ := new(big.Int).SetString(a.StartNS, 10)
		y, _ := new(big.Int).SetString(b.StartNS, 10)
		if c := x.Cmp(y); c != 0 {
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
func tpbLiveRaw(raw json.RawMessage, r tpbLiveRecord, env string) bool {
	var x struct {
		TraceID                                        string `json:"TraceId"`
		SpanID                                         string `json:"SpanId"`
		ParentID                                       string `json:"ParentSpanId"`
		Service                                        string `json:"ServiceName"`
		Operation                                      string `json:"SpanName"`
		Timestamp                                      string `json:"timestamp_ns"`
		Duration                                       string `json:"duration_ns"`
		Attributes, SpanAttributes, ResourceAttributes map[string]string
		Status                                         string `json:"StatusCode"`
	}
	if json.Unmarshal(raw, &x) != nil {
		return false
	}
	var physical map[string]json.RawMessage
	if json.Unmarshal(raw, &physical) != nil {
		return false
	}
	for key, want := range map[string]string{"ScopeName": "goo-journey-184", "ScopeVersion": "", "SpanKind": "Internal", "StatusMessage": ""} {
		var value string
		if json.Unmarshal(physical[key], &value) != nil || value != want {
			return false
		}
	}
	for _, key := range []string{"event_times_ns", "event_names", "event_attributes", "linked_attributes", "linked_traces", "linked_spans"} {
		var value []json.RawMessage
		if json.Unmarshal(physical[key], &value) != nil || len(value) != 0 {
			return false
		}
	}
	if x.TraceID != r.TraceID || x.SpanID != r.SpanID || x.ParentID != r.ParentSpanID || x.Service != r.Service || x.Operation != r.Operation || x.Timestamp != r.StartNS || x.Duration != r.DurationNS || x.Status != "Ok" || !aeJSONEqual(x.Attributes, r.Attributes) || !aeJSONEqual(x.SpanAttributes, r.Attributes) {
		return false
	}
	want := map[string]string{"service.name": r.Service, "deployment.environment.name": env}
	for k, v := range r.Resource {
		want[k] = v
	}
	return aeJSONEqual(x.ResourceAttributes, want)
}
func tpbLiveSource(result exploration.Result, f tpbLiveFixture) bool {
	q := result.Question
	if q.Environment != f.Environment || (q.Signal != "spans" && q.Signal != "logs") || q.Limit != 100 || q.Offset < 0 || q.FromMS != f.FromMS || q.ToMS != f.ToMS || q.Numeric != nil || q.ReportedEdge != nil || q.BucketMS != 0 || !peLiveHex(result.Basis) || !peLiveClock(result.CapturedAt) {
		return false
	}
	expected := tpbLiveMatching(q, f)
	state := "observed"
	if len(expected) == 0 {
		state = "no_data"
	}
	if result.Count != strconv.Itoa(len(expected)) || result.State != state || len(result.TimeSeries) != 0 {
		return false
	}
	start := q.Offset
	if start > len(expected) {
		start = len(expected)
	}
	end := start + q.Limit
	if end > len(expected) {
		end = len(expected)
	}
	if len(result.Rows) != end-start || result.HasMore != (end < len(expected)) {
		return false
	}
	for i, raw := range result.Rows {
		if !tpbLiveRaw(raw, expected[start+i], f.Environment) {
			return false
		}
	}
	if !tpbLiveAggregates(q, result.Duration, result.SpanStatus, result.Groups, expected) {
		return false
	}
	return true
}
func tpbLiveViewOK(v tpbLiveView, result exploration.Result) bool {
	truncated := result.HasMore || result.Question.Offset > 0
	if v.Input.Truncated != truncated || !tpbLiveEnrichedRows(v.Input.Rows, result.Rows) || !tpbOracleEqual(result.Rows, truncated, v.Model) || len(v.Groups) != len(v.Model.Groups) || len(v.Diagnostics) != len(v.Model.Rows) {
		return false
	}
	for i, g := range v.Groups {
		want := v.Model.Groups[i]
		if !aeJSONEqual(g.Key, [2]*string{want.Service, want.Operation}) || !aeJSONEqual(g.Indexes, want.RowIndexes) || len(g.Values) != 2 {
			return false
		}
		for j, value := range g.Values {
			measure := "inclusiveNs"
			n := want.InclusiveNS
			k, u := want.InclusiveKnown, want.InclusiveUnknown
			if j == 1 {
				measure = "outsideChildrenNs"
				n = want.OutsideChildrenNS
				k, u = want.OutsideKnown, want.OutsideUnknown
			}
			text := "unavailable"
			if n != nil {
				text = *n
			}
			if value.Measure != measure || value.Value != text || value.Known != strconv.Itoa(k) || value.Unknown != strconv.Itoa(u) || value.Bar != nil && !tpbLiveBarOK(*value.Bar, n, k, j, v.Model) {
				return false
			}
		}
	}
	for i, d := range v.Diagnostics {
		want := v.Model.Rows[i]
		if d.RowIndex != i || d.State != want.State || !aeJSONEqual(d.Reasons, want.Reasons) || len(d.Values) != 3 {
			return false
		}
		for j, value := range d.Values {
			measure := []string{"inclusiveNs", "childUnionNs", "outsideChildrenNs"}[j]
			n := []*string{want.InclusiveNS, want.ChildUnionNS, want.OutsideChildrenNS}[j]
			text := "unavailable"
			if n != nil {
				text = *n
			}
			if value.Measure != measure || value.Value != text {
				return false
			}
		}
	}
	return true
}
func tpbLiveRequestOK(requests []tpbLiveRequest, f tpbLiveFixture, want int) bool {
	if len(requests) != want {
		return false
	}
	for i, r := range requests {
		if r.Effect != "source" || r.Ordinal != i+1 {
			return false
		}
		if r.Endpoint == "/api/explore" {
			var q exploration.Query
			if json.Unmarshal(r.Payload, &q) != nil || q.Environment != f.Environment || q.FromMS != f.FromMS || q.ToMS != f.ToMS || q.Where == nil || q.Offset < 0 || q.Limit != 100 {
				return false
			}
			if q.Signal != "spans" && q.Signal != "logs" {
				return false
			}
		} else if r.Endpoint == "/api/captures" {
			var c struct {
				Title    string
				Question exploration.CaptureQuestion
			}
			if json.Unmarshal(r.Payload, &c) != nil || c.Title == "" || c.Question.Kind != "explore" || c.Question.Explore == nil {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

func tpbLiveMainScope(l tpbLiveMain) bool {
	if !tpbLiveFixtureOK(l.tpbLiveFixture) || !tpbLiveSource(l.Original, l.tpbLiveFixture) || !tpbLiveSource(l.TracePage, l.tpbLiveFixture) || !tpbLiveSource(l.SelectedLogs, l.tpbLiveFixture) {
		return false
	}
	if !nhLiveSameQuery(l.Original.Question, tpbLiveQuestion(l.tpbLiveFixture, l.Traces[0])) {
		return false
	}
	if l.Original.Count != "5" || l.TracePage.Count != "5" || l.Original.Question.Offset != 0 || !tpbLiveExpanded(l.Original.Question, l.TracePage.Question) || l.SelectedLogs.Question.Signal != "logs" || l.SelectedLogs.Count != "0" {
		return false
	}
	if !tpbLiveGeometryOK(l.DesktopGeometry, l.View.Model) || !tpbLiveGeometryOK(l.MobileGeometry, l.View.Model) {
		return false
	}
	if !tpbLiveBarsPresent(l.View) || !tpbLiveBarsPresent(l.Highlighted) {
		return false
	}
	if !tpbLiveViewOK(l.View, l.TracePage) || !tpbLiveViewOK(l.Highlighted, l.TracePage) || l.View.SelectedGroupIndex != -1 || l.Highlighted.SelectedGroupIndex < 0 || l.Highlighted.SelectedGroupIndex >= len(l.View.Model.Groups) {
		return false
	}
	if !aeJSONEqual(l.View.Input, l.Highlighted.Input) || !aeJSONEqual(l.View.Model, l.Highlighted.Model) {
		return false
	}
	logs := l.Original.Question
	logs.Signal = "logs"
	selected := logs
	selected.Where = &exploration.Predicate{All: append(append([]exploration.Predicate{}, logs.Where.All...), exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "span_id"}, Op: "eq", Value: l.Traces[0].Records[0].SpanID})}
	if !nhLiveSameQuery(l.SelectedLogs.Question, selected) {
		return false
	}
	if len(l.Responses) != 4 {
		return false
	}
	if !aeJSONEqual(l.Responses[0].Result, l.Original) || !aeJSONEqual(l.Responses[1].Result, l.TracePage) || !nhLiveSameQuery(l.Responses[2].Question, logs) || !aeJSONEqual(l.Responses[3].Result, l.SelectedLogs) {
		return false
	}
	for _, r := range l.Responses {
		if !nhLiveSameQuery(r.Question, r.Result.Question) || !tpbLiveSource(r.Result, l.tpbLiveFixture) {
			return false
		}
	}
	// Request order is the existing source path, not an extra breakdown query.
	if !tpbLiveRequestOK(l.Requests, l.tpbLiveFixture, 5) {
		return false
	}
	for i := 0; i < 4; i++ {
		var q exploration.Query
		if json.Unmarshal(l.Requests[i].Payload, &q) != nil || !nhLiveSameQuery(q, l.Responses[i].Question) {
			return false
		}
	}
	return true
}
func tpbLiveCustody(l tpbLiveMain) bool {
	if len(l.SourceBytes) != l.Custody.Bytes || !bytes.Equal(l.SourceBytes, l.AfterBytes) || l.Saved.Receipt.ID != l.Custody.ID || !aeJSONEqual(l.Saved, l.Reopened) || l.Saved.Explore == nil || l.Saved.RetainedRecords != 5 || l.Saved.RecordsLimited {
		return false
	}
	hash := sha256.Sum256(l.SourceBytes)
	if hex.EncodeToString(hash[:]) != l.Custody.SHA256 {
		return false
	}
	var stored exploration.SavedCapture
	if json.Unmarshal(l.SourceBytes, &stored) != nil || stored.Explore == nil || stored.ID != l.Custody.ID || !aeJSONEqual(stored.Explore, l.Saved.Explore) || !aeJSONEqual(stored.CaptureReceipt, l.Saved.Receipt) {
		return false
	}
	page := exploration.CapturePage{Receipt: stored.CaptureReceipt, Explore: stored.Explore}
	if !peLiveCaptureHash(page) || !tpbLiveSource(*stored.Explore, l.tpbLiveFixture) || !nhLiveSameQuery(stored.Explore.Question, l.Original.Question) || stored.Question.Explore == nil || !nhLiveSameQuery(*stored.Question.Explore, l.Original.Question) {
		return false
	}
	// Custody follows the explicit fresh source read, not the original page clock.
	custody, cerr := time.Parse(time.RFC3339Nano, stored.CapturedAt)
	source, serr := time.Parse(time.RFC3339Nano, stored.Explore.CapturedAt)
	original, oerr := time.Parse(time.RFC3339Nano, l.Original.CapturedAt)
	if cerr != nil || serr != nil || oerr != nil || custody.Before(source) || !source.After(original) {
		return false
	}
	var capture struct {
		Title    string
		Question exploration.CaptureQuestion
	}
	if json.Unmarshal(l.Requests[4].Payload, &capture) != nil || capture.Title != stored.Title || !aeJSONEqual(capture.Question, stored.Question) {
		return false
	}
	return true
}
func tpbLiveEvents(events []tpbLiveEvent, names []string, counts []int) bool {
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
func tpbLiveMainOK(l tpbLiveMain) bool {
	return tpbLiveMainScope(l) && tpbLiveCustody(l) && tpbLiveEvents(l.Events, []string{"run", "expand", "group-highlight", "group-clear", "row-select", "capture", "reopen"}, []int{1, 3, 3, 3, 4, 5, 5})
}
func tpbLiveGraphsOK(l tpbLiveGraphs) bool {
	if !tpbLiveFixtureOK(l.tpbLiveFixture) || len(l.Cases) != 6 || len(l.Events) != 6 {
		return false
	}
	count := 0
	expectedRequests := []exploration.Query{}
	for i, c := range l.Cases {
		tr := l.Traces[i+1]
		if !nhLiveSameQuery(c.Original.Question, tpbLiveQuestion(l.tpbLiveFixture, tr)) {
			return false
		}
		if c.Case != tr.Case || !tpbLiveSource(c.Original, l.tpbLiveFixture) || !tpbLiveSource(c.TracePage, l.tpbLiveFixture) || !tpbLiveViewOK(c.View, c.TracePage) || !tpbLiveExpanded(c.Original.Question, c.TracePage.Question) || c.Original.Count != strconv.Itoa(len(tr.Records)) {
			return false
		}
		count += 3
		logs := c.Original.Question
		logs.Signal = "logs"
		expectedRequests = append(expectedRequests, c.Original.Question, c.TracePage.Question, logs)
		if c.Case == "bounded-page" {
			if c.Next == nil || c.NextView == nil || !tpbLiveSource(*c.Next, l.tpbLiveFixture) || !tpbLiveViewOK(*c.NextView, *c.Next) || c.Next.Question.Offset != 100 || c.Next.Count != "126" || len(c.Next.Rows) != 26 || !c.NextView.Model.Truncated {
				return false
			}
			parent := c.TracePage.Question
			parent.Offset = 100
			if !nhLiveSameQuery(parent, c.Next.Question) {
				return false
			}
			count += 2
			expectedRequests = append(expectedRequests, c.Next.Question, logs)
		} else if c.Next != nil || c.NextView != nil {
			return false
		}
		if l.Events[i].Action != c.Case || l.Events[i].Source != count {
			return false
		}
	}
	return tpbLiveRequestOK(l.Requests, l.tpbLiveFixture, count) && tpbLiveQueries(l.Requests, expectedRequests)
}
func tpbLiveOwnershipOK(l tpbLiveOwnership) bool {
	if !tpbLiveFixtureOK(l.tpbLiveFixture) || !tpbLiveSource(l.TracePage, l.tpbLiveFixture) || !tpbLiveViewOK(l.Initial, l.TracePage) || !tpbLiveRequestOK(l.Requests, l.tpbLiveFixture, 3) || !tpbLiveEvents(l.Events, []string{"closed-controls", "hidden-controls", "return-current"}, []int{3, 3, 3}) {
		return false
	}
	if l.BeforeURL != l.AfterURL || !strings.HasSuffix(l.BeforeURL, "#definitions") || !l.Hidden.Connected || l.Hidden.Highlights != 0 || l.Hidden.SelectedIndex != l.Initial.SelectedIndex || l.Hidden.SelectedGroupIndex != l.Initial.SelectedGroupIndex {
		return false
	}
	original := tpbLiveQuestion(l.tpbLiveFixture, l.Traces[0])
	expanded := original
	expanded.GroupBy = []exploration.Field{{Kind: "column", Key: "service"}, {Kind: "column", Key: "name"}}
	expanded.Summary = "span-duration"
	expanded.SpanStatus = true
	logs := original
	logs.Signal = "logs"
	if !nhLiveSameQuery(l.TracePage.Question, expanded) || !tpbLiveQueries(l.Requests, []exploration.Query{original, expanded, logs}) {
		return false
	}
	e := l.Events[1]
	return e.Connected == l.Hidden.Connected && e.Highlights == 0 && e.SelectedIndex == l.Hidden.SelectedIndex && e.SelectedGroupIndex == l.Hidden.SelectedGroupIndex
}
func tpbLiveReplay(t *testing.T, events []tpbLiveEvent, model string) bool {
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
func tpbLiveRead(t *testing.T, env string, out any) string {
	file := os.Getenv(env)
	if file == "" {
		t.Skip("requires actual ordinary-OTLP trace-page ledger")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(raw, out); e != nil {
		t.Fatal(e)
	}
	return file
}
func tpbLiveClone[T any](l T) T {
	raw, _ := json.Marshal(l)
	var out T
	_ = json.Unmarshal(raw, &out)
	return out
}
func tpbLiveChanged[T any](t *testing.T, base, bad T) {
	a, _ := json.Marshal(base)
	b, _ := json.Marshal(bad)
	if bytes.Equal(a, b) {
		t.Fatal("negative control did not change serialized evidence")
	}
}

func TestQuintTracePageBreakdownLiveLedger(t *testing.T) {
	var base tpbLiveMain
	file := tpbLiveRead(t, "GOO_TRACE_PAGE_LEDGER", &base)
	var e error
	base.SourceBytes, e = os.ReadFile(filepath.Join(filepath.Dir(file), "scope-capture.json"))
	if e != nil {
		t.Fatal(e)
	}
	base.AfterBytes, e = os.ReadFile(filepath.Join(filepath.Dir(file), "scope-capture.after.json"))
	if e != nil {
		t.Fatal(e)
	}
	if !tpbLiveMainOK(base) {
		t.Fatalf("trace page rejected scope=%v custody=%v effects=%v", tpbLiveMainScope(base), tpbLiveCustody(base), tpbLiveEvents(base.Events, []string{"run", "expand", "group-highlight", "group-clear", "row-select", "capture", "reopen"}, []int{1, 3, 3, 3, 4, 5, 5}))
	}
	if !tpbLiveReplay(t, base.Events, "model/trace_page.yaml") {
		t.Fatal("actual lifecycle rejected")
	}
	tpbLiveRejectTraces(t, base.Events, "model/trace_page.yaml")
	controls := []struct {
		name   string
		mutate func(*tpbLiveMain)
	}{
		{"overlap double counted", func(l *tpbLiveMain) { s := "90000000"; l.View.Model.Rows[0].ChildUnionNS = &s }},
		{"outside residual", func(l *tpbLiveMain) { s := "10000000"; l.View.Model.Rows[0].OutsideChildrenNS = &s }},
		{"inclusive duration rewritten", func(l *tpbLiveMain) { s := "30"; l.View.Model.Rows[0].InclusiveNS = &s }},
		{"group sum", func(l *tpbLiveMain) { s := "1"; l.View.Model.Groups[0].InclusiveNS = &s }},
		{"group eligibility", func(l *tpbLiveMain) { l.View.Model.Groups[0].OutsideKnown++ }},
		{"group row index", func(l *tpbLiveMain) { l.View.Model.Groups[0].RowIndexes[0] = 99 }},
		{"DOM exact value", func(l *tpbLiveMain) { l.View.Groups[0].Values[0].Value = "1" }},
		{"bar shared max", func(l *tpbLiveMain) { l.View.Groups[0].Values[0].Bar.ScaleMax = "1" }},
		{"bar wrong side", func(l *tpbLiveMain) { l.View.Groups[0].Values[0].Bar.Side = "selection" }},
		{"bar quantized width", func(l *tpbLiveMain) { s := "99.99%"; l.View.Groups[0].Values[0].Bar.Width = &s }},
		{"bar rounded title", func(l *tpbLiveMain) { l.View.Groups[0].Values[0].Bar.Title = "rounded" }},
		{"mobile label clipped", func(l *tpbLiveMain) { l.MobileGeometry.Labels[0].Left = -1 }},
		{"separate physical track scales", func(l *tpbLiveMain) { l.DesktopGeometry.Tracks[0][1] += 10 }},

		{"DOM unknown count", func(l *tpbLiveMain) { l.View.Groups[0].Values[0].Unknown = "1" }},
		{"DOM reason", func(l *tpbLiveMain) { l.View.Diagnostics[0].Reasons = []string{"parent-cycle"} }},
		{"page index", func(l *tpbLiveMain) { l.View.Model.Rows[0].RowIndex = 99 }},
		{"raw duration", func(l *tpbLiveMain) { lpLiveMutateRow(&l.TracePage.Rows[0], "duration_ns", "100000001") }},
		{"raw timestamp rounded", func(l *tpbLiveMain) { lpLiveMutateRow(&l.TracePage.Rows[1], "timestamp_ns", "9007199254740992") }},
		{"raw parent edge", func(l *tpbLiveMain) { lpLiveMutateRow(&l.TracePage.Rows[1], "ParentSpanId", "0000000000000099") }},
		{"raw attribute", func(l *tpbLiveMain) { lpLiveMutateRow(&l.TracePage.Rows[0], "Attributes", "changed") }},
		{"duplicate observed row", func(l *tpbLiveMain) { l.TracePage.Rows[1] = l.TracePage.Rows[0] }},
		{"page scope", func(l *tpbLiveMain) { l.TracePage.Question.ToMS++ }},
		{"page offset", func(l *tpbLiveMain) { l.TracePage.Question.Offset = 100 }},
		{"false truncation", func(l *tpbLiveMain) { l.View.Input.Truncated = true; l.View.Model.Truncated = true }},
		{"source count", func(l *tpbLiveMain) { l.TracePage.Count = "100" }},
		{"source basis", func(l *tpbLiveMain) { l.TracePage.Basis = "bad" }},
		{"highlight changes model", func(l *tpbLiveMain) { l.Highlighted.Model.State = "partial" }},
		{"highlight outside group", func(l *tpbLiveMain) { l.Highlighted.SelectedGroupIndex = 999 }},
		{"hidden group read", func(l *tpbLiveMain) { l.Events[2].Source++ }},
		{"hidden reopen read", func(l *tpbLiveMain) { l.Events[6].Source++ }},
		{"missing selected linked-log read", func(l *tpbLiveMain) { l.Requests = l.Requests[:4] }},
		{"source response mismatch", func(l *tpbLiveMain) { l.Responses[0].Question.Environment = "other" }},
		{"capture count", func(l *tpbLiveMain) { l.Saved.Explore.Count = "999" }},
		{"capture custody", func(l *tpbLiveMain) { l.Custody.SHA256 = strings.Repeat("0", 64) }},
		{"capture title", func(l *tpbLiveMain) { l.Saved.Receipt.Title = "changed" }},
		{"reopened capture", func(l *tpbLiveMain) { l.Reopened.Receipt.Parent = "changed" }},
	}
	for _, control := range controls {
		t.Run(control.name, func(t *testing.T) {
			bad := tpbLiveClone(base)
			control.mutate(&bad)
			tpbLiveChanged(t, base, bad)
			bad.SourceBytes = base.SourceBytes
			bad.AfterBytes = base.AfterBytes
			if tpbLiveMainOK(bad) {
				t.Fatal("changed evidence accepted")
			}
		})
	}
}
func TestQuintTracePageGraphsLedger(t *testing.T) {
	var base tpbLiveGraphs
	tpbLiveRead(t, "GOO_TRACE_PAGE_GRAPHS_LEDGER", &base)
	if !tpbLiveGraphsOK(base) {
		t.Fatal("independent graph/page oracle rejected actual evidence")
	}
	if !tpbLiveReplay(t, base.Events, "model/trace_page_graphs.yaml") {
		t.Fatal("actual graph lifecycle rejected")
	}
	tpbLiveRejectTraces(t, base.Events, "model/trace_page_graphs.yaml")
	for index := range base.Cases {
		for _, kind := range []string{"row arithmetic", "row authority", "raw edge"} {
			t.Run(fmt.Sprintf("%s %s", base.Cases[index].Case, kind), func(t *testing.T) {
				bad := tpbLiveClone(base)
				c := &bad.Cases[index]
				switch kind {
				case "row arithmetic":
					s := "1"
					c.View.Model.Rows[0].ChildUnionNS = &s
				case "row authority":
					c.View.Model.Rows[0].State = "fabricated"
				case "raw edge":
					lpLiveMutateRow(&c.TracePage.Rows[0], "ParentSpanId", "000000000000ffff")
				}
				tpbLiveChanged(t, base, bad)
				if tpbLiveGraphsOK(bad) {
					t.Fatal("changed graph accepted")
				}
			})
		}
	}
	controls := []struct {
		name   string
		mutate func(*tpbLiveGraphs)
	}{
		{"page offset", func(l *tpbLiveGraphs) { l.Cases[5].Next.Question.Offset = 0 }},
		{"page truncation", func(l *tpbLiveGraphs) {
			l.Cases[5].NextView.Input.Truncated = false
			l.Cases[5].NextView.Model.Truncated = false
		}},
		{"page membership repeated", func(l *tpbLiveGraphs) { l.Cases[5].Next.Rows[0] = l.Cases[5].TracePage.Rows[0] }},
		{"page full aggregate changed", func(l *tpbLiveGraphs) { l.Cases[5].Next.Duration.SumNS = "0" }},
		{"page local sum takes full query", func(l *tpbLiveGraphs) { s := "1250000000"; l.Cases[5].NextView.Model.Groups[0].InclusiveNS = &s }},
		{"page omitted source read", func(l *tpbLiveGraphs) { l.Events[5].Source-- }},
	}
	for _, control := range controls {
		t.Run(control.name, func(t *testing.T) {
			bad := tpbLiveClone(base)
			control.mutate(&bad)
			tpbLiveChanged(t, base, bad)
			if tpbLiveGraphsOK(bad) {
				t.Fatal("changed page accepted")
			}
		})
	}
}
func TestQuintTracePageOwnershipLedger(t *testing.T) {
	var base tpbLiveOwnership
	tpbLiveRead(t, "GOO_TRACE_PAGE_OWNERSHIP_LEDGER", &base)
	if !tpbLiveOwnershipOK(base) {
		t.Fatal("independent ownership evidence rejected")
	}
	if !tpbLiveReplay(t, base.Events, "model/trace_page_ownership.yaml") {
		t.Fatal("actual ownership lifecycle rejected")
	}
	tpbLiveRejectTraces(t, base.Events, "model/trace_page_ownership.yaml")
	controls := []struct {
		name   string
		mutate func(*tpbLiveOwnership)
	}{
		{"closed group read", func(l *tpbLiveOwnership) { l.Events[0].Source++ }},
		{"hidden row read", func(l *tpbLiveOwnership) { l.Events[1].Source++ }},
		{"return read", func(l *tpbLiveOwnership) { l.Events[2].Source++ }},
		{"stale navigation", func(l *tpbLiveOwnership) { l.AfterURL += "changed" }},
		{"detached instead of hidden owner", func(l *tpbLiveOwnership) { l.Hidden.Connected = false }},
		{"stale row selection", func(l *tpbLiveOwnership) { l.Hidden.SelectedIndex = 999 }},
		{"stale group selection", func(l *tpbLiveOwnership) { l.Hidden.SelectedGroupIndex = 999 }},
		{"stale highlights", func(l *tpbLiveOwnership) { l.Hidden.Highlights = 1 }},
		{"event hides selection", func(l *tpbLiveOwnership) { l.Events[1].SelectedIndex = 999 }},
		{"event hides highlights", func(l *tpbLiveOwnership) { l.Events[1].Highlights = 1 }},
		{"extra source request", func(l *tpbLiveOwnership) { r := l.Requests[0]; r.Ordinal = 4; l.Requests = append(l.Requests, r) }},
		{"changed model", func(l *tpbLiveOwnership) { s := "1"; l.Initial.Model.Rows[0].OutsideChildrenNS = &s }},
		{"changed raw row", func(l *tpbLiveOwnership) { lpLiveMutateRow(&l.TracePage.Rows[0], "duration_ns", "1") }},
		{"changed window", func(l *tpbLiveOwnership) { l.TracePage.Question.FromMS-- }},
	}
	for _, control := range controls {
		t.Run(control.name, func(t *testing.T) {
			bad := tpbLiveClone(base)
			control.mutate(&bad)
			tpbLiveChanged(t, base, bad)
			if tpbLiveOwnershipOK(bad) {
				t.Fatal("changed ownership accepted")
			}
		})
	}
}

func tpbLiveExpanded(original, expanded exploration.Query) bool {
	original.GroupBy = []exploration.Field{{Kind: "column", Key: "service"}, {Kind: "column", Key: "name"}}
	original.Summary = "span-duration"
	original.SpanStatus = true
	return nhLiveSameQuery(original, expanded)
}
func tpbLiveDuration(rows []tpbLiveRecord) exploration.DurationSummary {
	d := exploration.DurationSummary{State: "no_data", Count: strconv.Itoa(len(rows)), SumNS: "0"}
	if len(rows) == 0 {
		return d
	}
	values := make([]*big.Int, len(rows))
	sum := new(big.Int)
	for i, r := range rows {
		values[i], _ = new(big.Int).SetString(r.DurationNS, 10)
		sum.Add(sum, values[i])
	}
	sort.Slice(values, func(i, j int) bool { return values[i].Cmp(values[j]) < 0 })
	d.State = "observed"
	d.SumNS = sum.String()
	d.MinNS = values[0].String()
	d.MaxNS = values[len(values)-1].String()
	quantile := func(percent int) string {
		index := len(values) * percent / 100
		if index >= len(values) {
			index = len(values) - 1
		}
		return values[index].String()
	}
	d.P50NS = quantile(50)
	d.P95NS = quantile(95)
	d.P99NS = quantile(99)
	return d
}
func tpbLiveStatus(rows []tpbLiveRecord) exploration.SpanStatusSummary {
	return exploration.SpanStatusSummary{Count: strconv.Itoa(len(rows)), OK: strconv.Itoa(len(rows)), Error: "0", Unset: "0", Unknown: "0"}
}
func tpbLiveAggregates(q exploration.Query, duration *exploration.DurationSummary, status *exploration.SpanStatusSummary, rawGroups []json.RawMessage, rows []tpbLiveRecord) bool {
	if q.Summary == "span-duration" {
		want := tpbLiveDuration(rows)
		if duration == nil || *duration != want {
			return false
		}
	} else if q.Summary != "" || duration != nil {
		return false
	}
	if q.SpanStatus {
		want := tpbLiveStatus(rows)
		if status == nil || *status != want {
			return false
		}
	} else if status != nil {
		return false
	}
	if len(q.GroupBy) == 0 {
		return len(rawGroups) == 0
	}
	if !aeJSONEqual(q.GroupBy, []exploration.Field{{Kind: "column", Key: "service"}, {Kind: "column", Key: "name"}}) {
		return false
	}
	groups := map[string][]tpbLiveRecord{}
	for _, r := range rows {
		k, _ := json.Marshal([2]string{r.Service, r.Operation})
		groups[string(k)] = append(groups[string(k)], r)
	}
	if len(groups) != len(rawGroups) {
		return false
	}
	seen := map[string]bool{}
	for _, raw := range rawGroups {
		var g struct {
			G0, G1     string
			Count      string
			Duration   *exploration.DurationSummary
			SpanStatus *exploration.SpanStatusSummary
		}
		if json.Unmarshal(raw, &g) != nil {
			return false
		}
		key, _ := json.Marshal([2]string{g.G0, g.G1})
		records, ok := groups[string(key)]
		if !ok || seen[string(key)] || g.Count != strconv.Itoa(len(records)) {
			return false
		}
		seen[string(key)] = true
		wantDuration, wantStatus := tpbLiveDuration(records), tpbLiveStatus(records)
		if g.Duration == nil || *g.Duration != wantDuration || g.SpanStatus == nil || *g.SpanStatus != wantStatus {
			return false
		}
	}
	return true
}

func tpbLiveQuestion(f tpbLiveFixture, tr tpbLiveTrace) exploration.Query {
	return exploration.Query{Signal: "spans", Environment: f.Environment, FromMS: f.FromMS, ToMS: f.ToMS, Limit: 100, Where: &exploration.Predicate{All: []exploration.Predicate{{Field: &exploration.Field{Kind: "column", Key: "trace_id"}, Op: "eq", Value: tr.TraceID}}}}
}

func tpbLiveQueries(requests []tpbLiveRequest, expected []exploration.Query) bool {
	if len(requests) != len(expected) {
		return false
	}
	for i, r := range requests {
		var q exploration.Query
		if r.Endpoint != "/api/explore" || json.Unmarshal(r.Payload, &q) != nil || !nhLiveSameQuery(q, expected[i]) {
			return false
		}
	}
	return true
}

func tpbLiveRejectTraces(t *testing.T, events []tpbLiveEvent, model string) {
	for _, kind := range []string{"unexpected source effect", "wrong action order"} {
		t.Run("finite trace rejects "+kind, func(t *testing.T) {
			bad := append([]tpbLiveEvent{}, events...)
			if kind == "unexpected source effect" {
				bad[0].Source++
			} else {
				bad[0].Action = bad[1].Action
			}
			tpbLiveChanged(t, events, bad)
			if tpbLiveReplay(t, bad, model) {
				t.Fatal("finite model accepted changed trace")
			}
		})
	}
}

type tpbLiveBar struct {
	Value, ScaleMax, Side, Title string
	Width                        *string
}

func tpbLiveBarsPresent(v tpbLiveView) bool {
	for _, g := range v.Groups {
		for _, value := range g.Values {
			if value.Bar == nil {
				return false
			}
		}
	}
	return true
}
func tpbLiveBarOK(bar tpbLiveBar, value *string, known, index int, model exploration.TracePageBreakdown) bool {
	max := new(big.Int)
	for _, g := range model.Groups {
		for _, v := range []*string{g.InclusiveNS, g.OutsideChildrenNS} {
			if v != nil {
				n, ok := new(big.Int).SetString(*v, 10)
				if !ok {
					return false
				}
				if n.Cmp(max) > 0 {
					max = n
				}
			}
		}
	}
	side := "baseline"
	if index == 1 {
		side = "selection"
	}
	if bar.Side != side || bar.ScaleMax != max.String() {
		return false
	}
	if value == nil {
		return bar.Value == "unavailable" && bar.Width == nil && bar.Title == "No known supplied durations"
	}
	if bar.Value != *value || bar.Width == nil || bar.Title != *value+" ns (exact sum of "+strconv.Itoa(known)+" known supplied rows)" {
		return false
	}
	n, ok := new(big.Int).SetString(*value, 10)
	if !ok {
		return false
	}
	percent := new(big.Int)
	if max.Sign() > 0 {
		percent.Quo(new(big.Int).Mul(n, big.NewInt(10000)), max)
	}
	actual, e := strconv.ParseFloat(strings.TrimSuffix(*bar.Width, "%"), 64)
	return e == nil && strings.HasSuffix(*bar.Width, "%") && math.Abs(actual-float64(percent.Int64())/100) < 0.00001
}

type tpbLiveGeometry struct {
	Width  float64
	Tracks [][]float64
	Labels []struct {
		Text                          string
		Left, Right, Width, TextWidth float64
	}
	Diagnostics []struct{ Width, TextWidth float64 }
}

func tpbLiveGeometryOK(g tpbLiveGeometry, model exploration.TracePageBreakdown) bool {
	if g.Width <= 0 || len(g.Tracks) != len(model.Groups) || len(g.Labels) != len(model.Groups) || len(g.Diagnostics) != len(model.Rows) {
		return false
	}
	label := func(s *string, kind string) string {
		if s == nil {
			return kind + " not reported"
		}
		if *s == "" {
			return "Present empty " + strings.ToLower(kind)
		}
		raw, _ := json.Marshal(*s)
		return string(raw)
	}
	for i, tracks := range g.Tracks {
		if len(tracks) != 2 || tracks[0] <= 0 || tracks[1] <= 0 || math.Abs(tracks[0]-tracks[1]) > 1 {
			return false
		}
		v := g.Labels[i]
		want := model.Groups[i]
		if v.Text != label(want.Service, "Service")+" / "+label(want.Operation, "Operation") || v.Left < 0 || v.Right > g.Width+1 || v.Right < v.Left || v.Width <= 0 || v.TextWidth > v.Width+1 {
			return false
		}
	}
	for _, v := range g.Diagnostics {
		if v.Width < 160 || v.TextWidth > v.Width+1 {
			return false
		}
	}
	return true
}

// The existing waterfall supplies one non-authoritative numeric display alias.
// Validate its binary64 conversion separately, then require every raw field.
func tpbLiveEnrichedRows(input, raw []json.RawMessage) bool {
	if len(input) != len(raw) {
		return false
	}
	for i := range raw {
		var shown, original map[string]json.RawMessage
		if json.Unmarshal(input[i], &shown) != nil || json.Unmarshal(raw[i], &original) != nil {
			return false
		}
		if _, exists := original["duration_ms"]; exists {
			return false
		}
		var text string
		var ms float64
		if json.Unmarshal(original["duration_ns"], &text) != nil || json.Unmarshal(shown["duration_ms"], &ms) != nil {
			return false
		}
		ns, e := strconv.ParseFloat(text, 64)
		if e != nil || ms != ns/1000000 {
			return false
		}
		delete(shown, "duration_ms")
		if !aeJSONEqual(shown, original) {
			return false
		}
	}
	return true
}
