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
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type httpCounters385 struct{ Calls, Source, Plans, SQL, Documents, Captures, StoreReads int }
type httpFact385 struct {
	Name          string
	Before, After httpCounters385
}
type httpRead385 struct {
	Key      string
	Request  exploration.Query
	Response json.RawMessage
	Status   int
}
type httpCase385 struct {
	Scope                                   exploration.InventoryQuestion
	Automatic, Plan                         exploration.DashboardPlan
	Widgets                                 []httpRead385
	Initial, Retry, RecordsRun, Trace, Logs httpRead385
	ParentBefore, ParentAfter               json.RawMessage
	Facts                                   []httpFact385
}
type httpLedger385 struct {
	OriginalDocument, LinkedDocument workbench.Document
	Version                          string
	Errors                           []string
	Fixture                          struct {
		Rows     []httpRaw385
		Scope    exploration.InventoryQuestion
		Workload struct{ Service string }
	}
	Cases                               []httpCase385
	DestinationRun                      httpRead385
	CapturePage, Reopened               exploration.CapturePage
	CapturedRequest                     struct{ Question exploration.CaptureQuestion }
	StoredBytesBefore, StoredBytesAfter string
	SourceRefusalProbe                  struct{ Status int }
	EmptyCurrent                        struct {
		Scope    exploration.InventoryQuestion
		Plan     exploration.DashboardPlan
		Response json.RawMessage
	}
	ProxyCalls []struct {
		SQL    string
		Status int
		Fault  *string
	}
	Calls []struct {
		Path, Method   string
		Body, Response json.RawMessage
		Status         int
	}
}

func httpAdmitResult385(f httpFixture385, read httpRead385, records bool) error {
	if read.Status != 200 {
		return fmt.Errorf("successful read absent")
	}
	var r exploration.Result
	if json.Unmarshal(read.Response, &r) != nil {
		return fmt.Errorf("invalid result")
	}

	if read.Request.FromMS != f.Scope.FromMS || read.Request.ToMS != f.Scope.ToMS || !reflect.DeepEqual(read.Request.EnvironmentSelector, f.Scope.EnvironmentSelector) || !reflect.DeepEqual(r.Question, read.Request) {
		return fmt.Errorf("request scope/response echo changed")
	}
	sum := new(big.Int)
	max := new(big.Int)
	statuses := map[string]int{"Error": 0, "Ok": 0, "Unset": 0, "unknown": 0}
	for _, row := range f.Rows {
		if !httpMember385(f, row) {
			continue
		}
		n, ok := new(big.Int).SetString(row.DurationNS, 10)
		if !ok {
			return fmt.Errorf("invalid raw duration")
		}
		sum.Add(sum, n)
		if n.Cmp(max) > 0 {
			max.Set(n)
		}
		key := row.Status
		if key != "Error" && key != "Ok" && key != "Unset" {
			key = "unknown"
		}
		statuses[key]++
	}
	if read.Request.Summary == "span-duration" && (r.Duration == nil || r.Duration.SumNS != sum.String() || r.Duration.MaxNS != max.String()) {
		return fmt.Errorf("duration changed raw population")
	}
	if read.Request.SpanStatus && (r.SpanStatus == nil || r.SpanStatus.Error != strconv.Itoa(statuses["Error"]) || r.SpanStatus.OK != strconv.Itoa(statuses["Ok"]) || r.SpanStatus.Unset != strconv.Itoa(statuses["Unset"]) || r.SpanStatus.Unknown != strconv.Itoa(statuses["unknown"])) {
		return fmt.Errorf("recorded status changed or inferred from HTTP")
	}
	want := httpIDs385(f)
	if r.Count != strconv.Itoa(len(want)) {
		return fmt.Errorf("full count differs from raw Client records")
	}
	// Predicate membership is independently interpreted over raw source rows.
	for _, row := range f.Rows {
		if hdPredicate(read.Request.Where, httpValues385(row)) != (row.Service == f.Service && row.Kind == "Client" && row.Attributes["http.request.method"] != "" && continuationNamespace352(f.Scope.Namespace, row.Resource)) {
			return fmt.Errorf("query changed literal Client population")
		}
	}
	if records {
		got, err := reportedResultIDs353(r)
		if err != nil {
			return err
		}
		if len(got) != min(read.Request.Limit, len(want)) || r.HasMore != (len(want) > len(got)) {
			return fmt.Errorf("page/full count boundary changed")
		}
		for _, id := range got {
			if !slices.Contains(want, id) {
				return fmt.Errorf("foreign row retained")
			}
		}
		if len(got) == len(want) {
			slices.Sort(got)
			if !slices.Equal(got, want) {
				return fmt.Errorf("full record membership changed")
			}
		}
	}
	return nil
}
func admitHTTP385(r httpLedger385) error {
	fail := func(s string) error { return fmt.Errorf("http385: %s", s) }
	if r.Version != "goo.httpClient385/1" || len(r.Errors) > 0 || len(r.Cases) != 1 || len(r.Fixture.Rows) == 0 {
		return fail("missing actual journey")
	}
	c := r.Cases[0]
	f := httpFixture385{Scope: c.Scope, Service: r.Fixture.Workload.Service, Rows: r.Fixture.Rows}
	if len(httpIDs385(f)) <= 100 || c.Plan.Profile != "http-client" || c.Automatic.Profile == "http-client" {
		return fail("explicit lens/page boundary absent")
	}
	if !sameTransport384(c.ParentBefore, c.ParentAfter) {
		return fail("held parent changed")
	}
	keys := map[string]string{"dimension": "server.address", "http-client-methods": "http.request.method", "http-client-ports": "server.port", "http-client-responses": "http.response.status_code"}
	if !reflect.DeepEqual(c.Plan.Scope, c.Scope) || c.Plan.Service != f.Service {
		return fail("plan scope changed")
	}
	if len(c.Widgets) != 5 {
		return fail("widget portfolio changed")
	}
	seenWidgets := map[string]bool{}
	for _, w := range c.Widgets {
		if seenWidgets[w.Key] {
			return fail("duplicate widget")
		}
		seenWidgets[w.Key] = true
		found := false
		for _, planned := range c.Plan.Widgets {
			if planned.Key == w.Key && planned.Query != nil && reflect.DeepEqual(*planned.Query, w.Request) {
				found = true
			}
		}
		if !found {
			return fail("read departed from explicit plan")
		}
		if w.Key == "spans" {
			if w.Status != 502 {
				return fail("initial source refusal absent")
			}
			continue
		}
		if w.Request.FromMS != f.Scope.FromMS || w.Request.ToMS != f.Scope.ToMS || !reflect.DeepEqual(w.Request.EnvironmentSelector, f.Scope.EnvironmentSelector) {
			return fail("scope drift")
		}
		if e := httpAdmitResult385(f, w, false); e != nil {
			return e
		}
		var result struct {
			Groups []struct {
				G0    *string
				Count string
			}
		}
		if json.Unmarshal(w.Response, &result) != nil {
			return fail("groups malformed")
		}
		got := map[string]int{}
		for _, g := range result.Groups {

			n, e := strconv.Atoi(g.Count)
			if e != nil {
				return e
			}
			got[continuationJSON352(g.G0)] = n
		}
		if !reflect.DeepEqual(got, httpGroupCounts385(f, keys[w.Key])) {
			return fail("missing/empty/literal grouping changed")
		}
	}
	if !reflect.DeepEqual(c.Retry.Request, c.RecordsRun.Request) {
		return fail("retry/records question changed")
	}
	if e := httpAdmitResult385(f, c.Retry, false); e != nil {
		return e
	}
	if e := httpAdmitResult385(f, c.RecordsRun, true); e != nil {
		return e
	}
	var dest exploration.Result
	if json.Unmarshal(r.DestinationRun.Response, &dest) != nil || r.DestinationRun.Status != 200 {
		return fail("destination read absent")
	}
	wantDest := []string{}
	for _, row := range f.Rows {
		if httpMember385(f, row) {
			if _, ok := row.Attributes["server.address"]; !ok {
				wantDest = append(wantDest, row.ID)
			}
		}
	}
	gotDest, err := reportedResultIDs353(dest)
	if err != nil {
		return err
	}
	slices.Sort(gotDest)
	slices.Sort(wantDest)
	if !slices.Equal(gotDest, wantDest) || dest.Count != strconv.Itoa(len(wantDest)) {
		return fail("missing destination conflated with empty")
	}
	// Whole-trace context preserves env/window while explicitly dropping service/namespace.
	if c.Trace.Status != 200 || c.Logs.Status != 200 || c.Trace.Request.FromMS != f.Scope.FromMS || c.Trace.Request.ToMS != f.Scope.ToMS || !reflect.DeepEqual(c.Trace.Request.EnvironmentSelector, f.Scope.EnvironmentSelector) {
		return fail("whole trace scope drift")
	}
	var trace exploration.Result
	if json.Unmarshal(c.Trace.Response, &trace) != nil {
		return fail("trace malformed")
	}
	var actualTrace string
	gotTrace, err := reportedResultIDs353(trace)
	if err != nil {
		return err
	}
	if len(gotTrace) != 3 {
		return fail("whole trace context absent")
	}
	for _, row := range f.Rows {
		if row.ID == gotTrace[0] {
			actualTrace = row.TraceID
		}
	}
	wantTrace := []string{}
	for _, row := range f.Rows {
		if actualTrace != "" && row.TraceID == actualTrace {
			wantTrace = append(wantTrace, row.ID)
		}
	}
	slices.Sort(gotTrace)
	slices.Sort(wantTrace)
	if !reflect.DeepEqual(c.Logs.Request.Where, c.Trace.Request.Where) || c.Logs.Request.FromMS != c.Trace.Request.FromMS || c.Logs.Request.ToMS != c.Trace.Request.ToMS || !reflect.DeepEqual(c.Logs.Request.EnvironmentSelector, c.Trace.Request.EnvironmentSelector) {
		return fail("logs context scope changed")
	}
	wantTracePredicate := &exploration.Predicate{All: []exploration.Predicate{{Field: &exploration.Field{Kind: "column", Key: "trace_id"}, Op: "eq", Value: actualTrace}}}
	if !reflect.DeepEqual(c.Trace.Request.Where, wantTracePredicate) {
		return fail("whole trace query retained population filters")
	}

	if !slices.Equal(gotTrace, wantTrace) {
		return fail("trace membership changed")
	}
	if r.EmptyCurrent.Plan.Profile != "http-client" {
		return fail("empty scope fallback")
	}
	var empty struct{ Count string }
	if json.Unmarshal(r.EmptyCurrent.Response, &empty) != nil || empty.Count != "0" {
		return fail("empty source invented records")
	}
	// Captured evidence is a separate, complete source reread, with physical custody.
	p := r.CapturePage
	if p.Explore == nil || p.Receipt.Question.Kind != "explore" || p.Receipt.Question.Explore == nil || !reflect.DeepEqual(r.CapturedRequest.Question, p.Receipt.Question) || !reflect.DeepEqual(*p.Receipt.Question.Explore, c.RecordsRun.Request) {
		return fail("capture question changed")
	}
	if r.OriginalDocument.ID == "" || len(r.OriginalDocument.Cards) != 1 || r.OriginalDocument.Cards[0].Step == nil || r.OriginalDocument.Cards[0].Step.Record == nil || !reflect.DeepEqual(*r.OriginalDocument.Cards[0].Step.Record, c.RecordsRun.Request) || r.LinkedDocument.Parent != r.OriginalDocument.ID || len(r.LinkedDocument.Cards) != 1 || !slices.Equal(r.LinkedDocument.Cards[0].Evidence, []string{p.Receipt.ID}) {
		return fail("saved question/evidence linkage changed")
	}
	var displayed exploration.Result
	if json.Unmarshal(c.RecordsRun.Response, &displayed) != nil {
		return fail("displayed records malformed")
	}
	sourceClock, se := time.Parse(time.RFC3339Nano, p.Explore.CapturedAt)
	custodyClock, ce := time.Parse(time.RFC3339Nano, p.Receipt.CapturedAt)
	if se != nil || ce != nil || sourceClock.After(custodyClock) || p.Explore.CapturedAt == displayed.CapturedAt || p.Explore.Basis == "" {
		return fail("capture source/custody clocks not independent")
	}
	raw, _ := json.Marshal(p.Explore)
	if e := httpAdmitResult385(f, httpRead385{Request: *p.Receipt.Question.Explore, Response: raw, Status: 200}, false); e != nil {
		return e
	}
	// API pages stay bounded; physical custody retains the complete bounded answer.
	if p.Explore.Count != strconv.Itoa(len(httpIDs385(f))) || len(p.Explore.Rows) != 100 || !p.Explore.HasMore {
		return fail("capture completeness changed")
	}
	if !sameTransport384(mustJSON385(p), mustJSON385(r.Reopened)) || r.StoredBytesBefore == "" || r.StoredBytesBefore != r.StoredBytesAfter || r.SourceRefusalProbe.Status != 502 {
		return fail("offline custody changed")
	}
	var stored exploration.SavedCapture
	if json.Unmarshal([]byte(r.StoredBytesBefore), &stored) != nil || stored.ID != p.Receipt.ID {
		return fail("physical capture absent")
	}
	id := stored.ID
	stored.ID = ""
	wire, _ := json.Marshal(stored)
	hash := sha256.Sum256(wire)
	if hex.EncodeToString(hash[:]) != id {
		return fail("capture digest changed")
	}
	if stored.Explore == nil || stored.Explore.Count != strconv.Itoa(len(httpIDs385(f))) || len(stored.Explore.Rows) != len(httpIDs385(f)) {
		return fail("physical full capture incomplete")
	}
	physicalIDs, e := reportedResultIDs353(*stored.Explore)
	if e != nil || !slices.Equal(physicalIDs, httpIDs385(f)) {
		return fail("physical capture membership changed")
	}
	stored.ID = id
	storedPage, e := stored.Page("")
	if e != nil {
		return e
	}
	if !sameTransport384(mustJSON385(storedPage), mustJSON385(p)) {
		return fail("stored result changed")
	}
	allowed := map[string]int{"profile-apply": 5, "disclosure": 0, "retry": 1, "destination-stage": 0, "destination-run": 1, "destination-return": 0, "records-stage": 0, "records-run": 1, "expand-trace": 2, "return": 0, "save-question": 0, "capture": 0, "reopen-offline": 0}
	seen := map[string]bool{}
	for _, fact := range c.Facts {

		if fact.Before.Calls < 0 || fact.After.Calls < fact.Before.Calls || fact.After.Calls > len(r.Calls) {
			return fail("invalid call interval")
		}
		source, plans, documents, captures, store := 0, 0, 0, 0, 0
		for _, call := range r.Calls[fact.Before.Calls:fact.After.Calls] {
			if call.Method == "POST" && slices.Contains([]string{"/api/discovery", "/api/explore", "/api/explore/page", "/api/inventory", "/api/metric-discovery", "/api/record-aggregates", "/api/measurements"}, call.Path) {
				source++
			}
			if call.Method == "POST" && call.Path == "/api/dashboards/plan" {
				plans++
			}
			if call.Method == "POST" && call.Path == "/api/documents" {
				documents++
			}
			if call.Method == "POST" && call.Path == "/api/captures" {
				captures++
			}
			if call.Method == "GET" && len(call.Path) == len("/api/captures/")+64 && strings.HasPrefix(call.Path, "/api/captures/") {
				store++
			}
		}
		if source != fact.After.Source-fact.Before.Source || plans != fact.After.Plans-fact.Before.Plans || documents != fact.After.Documents-fact.Before.Documents || captures != fact.After.Captures-fact.Before.Captures || store != fact.After.StoreReads-fact.Before.StoreReads {
			return fail("measured counters disagree with actual call slice")
		}
		n, ok := allowed[fact.Name]
		if !ok || seen[fact.Name] {
			return fail("unknown/duplicate gesture")
		}
		seen[fact.Name] = true
		if fact.After.Source-fact.Before.Source != n {
			return fail("implicit or missing source read: " + fact.Name)
		}
		capture := 0
		if fact.Name == "capture" {
			capture = 1
		}
		if fact.After.Captures-fact.Before.Captures != capture {
			return fail("implicit capture")
		}
		if fact.Before.SQL < 0 || fact.After.SQL < fact.Before.SQL || fact.After.SQL > len(r.ProxyCalls) {
			return fail("invalid physical SQL interval")
		}
		if (n > 0 || capture > 0) && fact.After.SQL == fact.Before.SQL {
			return fail("source action lacks physical reads")
		}
		for _, call := range r.ProxyCalls[fact.Before.SQL:fact.After.SQL] {
			if call.SQL == "" || call.Status < 200 {
				return fail("physical source call absent")
			}
		}
		if n == 0 && capture == 0 && fact.After.SQL != fact.Before.SQL {
			return fail("passive source SQL")
		}
		expectedStore := 0
		if fact.Name == "reopen-offline" {
			expectedStore = 1
		}
		if fact.After.StoreReads-fact.Before.StoreReads != expectedStore {
			return fail("implicit capture-store read")
		}
	}
	if len(seen) != len(allowed) {
		return fail("missing gesture")
	}
	return nil
}
func mustJSON385(v any) json.RawMessage { b, _ := json.Marshal(v); return b }
func replayHTTP385(t *testing.T, label string, facts []httpFact385, mutate bool) bool {
	t.Helper()
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	reads, captures := 0, 0
	for _, f := range facts {
		delta := f.After.Source - f.Before.Source
		for range delta {
			reads++
			rec.Record(context.Background(), "read", nil, "obs.reads", reads, "obs.captures", captures)
		}
		action := "history"
		if f.After.Captures > f.Before.Captures {
			action = "capture"
			captures++
		}
		observed := reads
		if mutate && f.Name == "disclosure" {
			observed++
		}
		rec.Record(context.Background(), action, nil, "obs.reads", observed, "obs.captures", captures)
	}
	wire, _ := json.Marshal(sink.Steps())
	if e := artifact("http385-"+label+"-steps.json", wire); e != nil {
		t.Fatal(e)
	}
	checker, e := validate.NewChecker("model/captured_trace_search_383.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	result, e := checker.Check(t.Context(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	if e = artifact("http385-"+label+"-model.txt", []byte(result.Summary())); e != nil {
		t.Fatal(e)
	}
	return result.OK()
}
func TestQuintHTTPClientObserved385(t *testing.T) {
	path := os.Getenv("GOO_HTTP_CLIENT_385_REPORT")
	if path == "" {
		t.Skip("actual HTTP client journey report required")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var r httpLedger385
	if json.Unmarshal(raw, &r) != nil {
		t.Fatal("invalid report")
	}
	if e = admitHTTP385(r); e != nil {
		t.Fatal(e)
	}
	if !replayHTTP385(t, "actual", r.Cases[0].Facts, false) {
		t.Fatal("actual observed facts violate model")
	}
	mutations := []func(*httpLedger385){
		func(r *httpLedger385) { r.StoredBytesAfter += " " },
		func(r *httpLedger385) { r.Cases[0].ParentAfter = json.RawMessage(`{"changed":true}`) },
		func(r *httpLedger385) { r.Cases[0].RecordsRun.Request.FromMS++ },
		func(r *httpLedger385) { r.Cases[0].Trace.Request.Where = r.Cases[0].RecordsRun.Request.Where },
		func(r *httpLedger385) { r.Cases[0].Facts[1].After.Source++ },
		func(r *httpLedger385) { r.Cases[0].Widgets[1].Request.ToMS++ },

		func(r *httpLedger385) { r.CapturePage.Explore.Count = "100" },
		func(r *httpLedger385) { r.LinkedDocument.Cards[0].Evidence = []string{"wrong-capture"} },
		func(r *httpLedger385) {
			for i, row := range r.Fixture.Rows {
				f := httpFixture385{Scope: r.Cases[0].Scope, Service: r.Fixture.Workload.Service}
				if httpMember385(f, row) {
					r.Fixture.Rows[i].DurationNS = "9007199254740993"
					break
				}
			}
		},
		func(r *httpLedger385) {
			var value map[string]any
			json.Unmarshal(r.Cases[0].Retry.Response, &value)
			value["spanStatus"] = map[string]string{"count": "117", "error": "0", "ok": "117", "unset": "0", "unknown": "0"}
			r.Cases[0].Retry.Response = mustJSON385(value)
		},
	}
	for i, mutate := range mutations {
		var bad httpLedger385
		json.Unmarshal(raw, &bad)
		mutate(&bad)
		if admitHTTP385(bad) == nil {
			t.Fatalf("semantic mutant%d admitted", i)
		}
	}
	if replayHTTP385(t, "passive-read-mutant", r.Cases[0].Facts, true) {
		t.Fatal("passive source read escaped model")
	}
}
