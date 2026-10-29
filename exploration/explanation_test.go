package exploration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

const explanationTestTime = "2026-10-24T12:00:00Z"

func explanationRequest() ExplanationRequest {
	a := histogramQuestion("int64")
	a.Where = numericPredicateFixture("int64", "gte", "9007199254740993")
	a.GroupBy = []Field{{Kind: "column", Key: "service"}}
	a.BucketMS = 10000
	a.GroupTimeSeries = true
	a.Summary = "span-duration"
	a.SpanStatus = true
	a.Offset = 50
	b := detached(a)
	b.FromMS = 60000
	b.ToMS = 120000
	return ExplanationRequest{Version: ExplanationVersion, Baseline: a, Selection: b, Fields: []Field{{Kind: "attribute", Key: "route"}, {Kind: "resource", Key: "version"}}}
}
func explanationPopulation(count string) ExplanationPopulation {
	return ExplanationPopulation{Count: count, ReadAt: explanationTestTime, Basis: strings.Repeat("a", 64)}
}
func explanationRead(values ...ExplanationValueCount) ExplanationFieldRead {
	return ExplanationFieldRead{State: "complete", ReadAt: explanationTestTime, Basis: strings.Repeat("b", 64), Values: values}
}
func explanationFixture(t *testing.T) *ExplanationResult {
	t.Helper()
	request := explanationRequest()
	fields := []ExplanationFieldResult{{Field: request.Fields[0], Baseline: explanationRead(ExplanationValueCount{Value: nil, Count: "1"}, ExplanationValueCount{Value: str(""), Count: "1"}, ExplanationValueCount{Value: str("a"), Count: "2"}), Selection: explanationRead(ExplanationValueCount{Value: str(""), Count: "1"}, ExplanationValueCount{Value: str("a"), Count: "1"}, ExplanationValueCount{Value: str("b"), Count: "1"})}, {Field: request.Fields[1], Baseline: explanationRead(ExplanationValueCount{Value: str("v1"), Count: "4"}), Selection: explanationRead(ExplanationValueCount{Value: str("v2"), Count: "3"})}}
	r, e := BuildExplanation(request, explanationPopulation("4"), explanationPopulation("3"), fields, explanationTestTime)
	if e != nil {
		t.Fatal(e)
	}
	return r
}

func TestExplanationPlanProjectionAndExactRank(t *testing.T) {
	r := explanationFixture(t)
	p, e := PlanExplanation(r.Request)
	if e != nil {
		t.Fatal(e)
	}
	for i, q := range []Query{p.Baseline, p.Selection} {
		original := r.Request.Baseline
		if i == 1 {
			original = r.Request.Selection
		}
		if q.Numeric != nil || q.NumericHistogram != nil || q.Summary != "" || q.SpanStatus || q.GroupTimeSeries || q.BucketMS != 0 || len(q.GroupBy) != 0 || q.Offset != 0 || !reflect.DeepEqual(q.Where, original.Where) || q.FromMS != original.FromMS || q.ToMS != original.ToMS {
			t.Fatal("count projection changed source selector", q)
		}
	}
	if r.State != "complete" || ValidateExplanationResult(r) != nil || len(r.Ranked) != 6 {
		t.Fatal(r)
	}
	first := r.Ranked[0]
	if first.FieldIndex != 1 || *r.Fields[first.FieldIndex].Rows[first.RowIndex].Value != "v1" {
		t.Fatal("exact score/tie ranking changed", r.Ranked)
	}
	route := r.Fields[0]
	if route.Rows[0].Value != nil || route.Rows[0].BaselinePercent != "25" || route.Rows[0].DifferencePP != "-25" || route.Rows[1].Value == nil || *route.Rows[1].Value != "" || route.Rows[1].DifferencePP != "25/3" {
		t.Fatal("missing/empty or exact rational lost", route.Rows)
	}
	changed := detached(*r)
	changed.Fields[0].Rows[0].DifferencePP = "0"
	if ValidateExplanationResult(&changed) == nil {
		t.Fatal("forged arithmetic accepted")
	}
	changed = detached(*r)
	changed.Ranked[0].RowIndex = 999
	if ValidateExplanationResult(&changed) == nil {
		t.Fatal("forged lead reference accepted")
	}
	r.Request.Baseline.Where.Numeric.Unit = "mutated"
	if p.Baseline.Where.Numeric.Unit == "mutated" {
		t.Fatal("plan aliases caller")
	}
}

func TestExplanationPartialUnknownByteBudgetAndNoData(t *testing.T) {
	good := explanationFixture(t)
	for _, state := range []string{"partial", "unavailable", "changed"} {
		fields := detached(good.Fields)
		if state == "partial" {
			fields[0].Baseline = ExplanationFieldRead{State: "partial", Reason: "cardinality-limit", ReadAt: explanationTestTime, Basis: strings.Repeat("c", 64), Values: []ExplanationValueCount{}}
		} else if state == "unavailable" {
			fields[0].Selection = ExplanationFieldRead{State: "unavailable", Reason: "source-read-failed", ReadAt: explanationTestTime, Values: []ExplanationValueCount{}}
		} else {
			fields[0].Selection.Values[0].Count = "2"
		}
		r, e := BuildExplanation(good.Request, good.Baseline, good.Selection, fields, explanationTestTime)
		if e != nil || r.State != "partial" || len(r.Fields[0].Rows) != 0 || ValidateExplanationResult(r) != nil {
			t.Fatal("partial/changed field claimed complete", state, r, e)
		}
		for _, lead := range r.Ranked {
			if lead.FieldIndex == 0 {
				t.Fatal("ranked incomplete field")
			}
		}
	}
	request := good.Request
	request.Fields = request.Fields[:1]
	large := explanationRead()
	for i := 0; i < 100; i++ {
		v := fmt.Sprintf("%03d", i) + strings.Repeat("x", 2000)
		large.Values = append(large.Values, ExplanationValueCount{Value: &v, Count: "1"})
	}
	r, e := BuildExplanation(request, explanationPopulation("100"), explanationPopulation("100"), []ExplanationFieldResult{{Field: request.Fields[0], Baseline: large, Selection: large}}, explanationTestTime)
	if e != nil || r.Fields[0].Reason != "field-bytes-limit" || len(r.Fields[0].Rows) != 0 || len(r.Fields[0].Baseline.Values) != 0 || len(r.Ranked) != 0 || ValidateExplanationResult(r) != nil {
		t.Fatal("large field not refused whole", r, e)
	}
	wire, _ := json.Marshal(r.Fields[0])
	if len(wire) > ExplanationFieldMaxBytes {
		t.Fatal("field budget not bounded")
	}
	empty := explanationRead()
	r, e = BuildExplanation(request, explanationPopulation("0"), explanationPopulation("3"), []ExplanationFieldResult{{Field: request.Fields[0], Baseline: empty, Selection: good.Fields[0].Selection}}, explanationTestTime)
	if e != nil || r.State != "no_data" || len(r.Ranked) != 0 {
		t.Fatal(r, e)
	}
	for _, row := range r.Fields[0].Rows {
		if row.BaselinePercent != "" || row.DifferencePP != "" || row.SelectionPercent == "" {
			t.Fatal("zero denominator inferred zero share", row)
		}
	}
}

func TestExplanationValidationAndFilterComplements(t *testing.T) {
	request := explanationRequest()
	for _, mutate := range []func(*ExplanationRequest){func(r *ExplanationRequest) { r.Version = "other" }, func(r *ExplanationRequest) { r.Fields = nil }, func(r *ExplanationRequest) { r.Fields = append(r.Fields, r.Fields[0]) }, func(r *ExplanationRequest) { r.Selection.Environment = "other" }, func(r *ExplanationRequest) { r.Fields = []Field{{Kind: "column", Key: "duration_bucket"}} }, func(r *ExplanationRequest) { r.Fields = []Field{{Kind: "column", Key: "body"}} }} {
		bad := detached(request)
		mutate(&bad)
		if _, e := PlanExplanation(bad); e == nil {
			t.Fatal("invalid request planned")
		}
	}
	field := request.Fields[0]
	for _, value := range []*string{nil, str(""), str("a")} {
		include, e := PlanExplanationFilter(request, field, value, "selection", "include")
		if e != nil {
			t.Fatal(e)
		}
		exclude, e := PlanExplanationFilter(request, field, value, "selection", "exclude")
		if e != nil {
			t.Fatal(e)
		}
		if include.FromMS != request.Selection.FromMS || include.Numeric == nil || include.NumericHistogram == nil {
			t.Fatal("filter dropped original side intent")
		}
		for _, actual := range []*string{nil, str(""), str("a"), str("b")} {
			row := map[Field]string{request.Selection.Where.Numeric.Field: "9007199254740993"}
			if actual != nil {
				row[field] = *actual
			}
			a, b := Match(include.Where, row), Match(exclude.Where, row)
			if a == b {
				t.Fatal("include/exclude not parent complement", value, actual, a, b)
			}
			row[request.Selection.Where.Numeric.Field] = "bad"
			if Match(include.Where, row) || Match(exclude.Where, row) {
				t.Fatal("filter broadened typed parent")
			}
		}
	}
	for _, tc := range []struct {
		field      Field
		side, mode string
	}{{Field{Kind: "attribute", Key: "not-requested"}, "baseline", "include"}, {field, "other", "include"}, {field, "baseline", "other"}} {
		if _, e := PlanExplanationFilter(request, tc.field, str("a"), tc.side, tc.mode); e == nil {
			t.Fatal("invalid filter staged")
		}
	}
}

func TestExplanationSourceReadsOnlyCountsGroupsAndRetainsFailures(t *testing.T) {
	request := explanationRequest()
	reads := 0
	queries := []string{}
	cardinality := false
	failed := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reads++
		body, _ := io.ReadAll(r.Body)
		sql := string(body)
		queries = append(queries, sql)
		side := r.URL.Query().Get("param_from")
		if !strings.Contains(sql, "GROUP BY") {
			count := "4"
			if side == "60000" {
				count = "3"
			}
			fmt.Fprintf(w, "{\"count\":%q}\n", count)
			return
		}
		version := false
		for k, values := range r.URL.Query() {
			if strings.HasPrefix(k, "param_") && len(values) > 0 && values[0] == "version" {
				version = true
			}
		}
		if version && failed {
			http.Error(w, "do not expose secret endpoint", 500)
			return
		}
		if version && cardinality {
			for i := 0; i < 1001; i++ {
				fmt.Fprintf(w, "{\"g0\":%q,\"count\":\"1\"}\n", fmt.Sprint(i))
			}
			return
		}
		if version {
			value, count := "v1", "4"
			if side == "60000" {
				value, count = "v2", "3"
			}
			fmt.Fprintf(w, "{\"g0\":%q,\"count\":%q}\n", value, count)
			return
		}
		if side == "0" {
			fmt.Fprintln(w, `{"g0":null,"count":"1"}`)
			fmt.Fprintln(w, `{"g0":"","count":"1"}`)
			fmt.Fprintln(w, `{"g0":"a","count":"2"}`)
		} else {
			fmt.Fprintln(w, `{"g0":"","count":"1"}`)
			fmt.Fprintln(w, `{"g0":"a","count":"1"}`)
			fmt.Fprintln(w, `{"g0":"b","count":"1"}`)
		}
	}))
	defer server.Close()
	source := Source{Endpoint: server.URL}
	r, e := source.Explain(context.Background(), request)
	if e != nil || reads != 6 || r.State != "complete" || ValidateExplanationResult(r) != nil {
		t.Fatal(r, reads, e)
	}
	for _, sql := range queries {
		if strings.Contains(sql, " AS timestamp_ns") || strings.Contains(sql, "SELECT DISTINCT") || strings.Contains(sql, " AS numeric") || strings.Contains(sql, "from_ms") || strings.Contains(sql, "quantilesExact") {
			t.Fatal("explanation executed extra record/measure read", sql)
		}
		if !strings.Contains(sql, "toInt64OrNull") {
			t.Fatal("typed parent dropped")
		}
	}
	cardinality = true
	r, e = source.Explain(context.Background(), request)
	if e != nil || r.Fields[1].State != "partial" || len(r.Fields[1].Rows) != 0 || r.Fields[0].State != "complete" {
		t.Fatal("cardinality looked complete", r, e)
	}
	cardinality = false
	failed = true
	r, e = source.Explain(context.Background(), request)
	wire, _ := json.Marshal(r)
	if e != nil || r.Fields[1].State != "unavailable" || strings.Contains(string(wire), "secret endpoint") {
		t.Fatal("field failure erased others or exposed error", e)
	}
	before := reads
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := source.Explain(ctx, request); e != context.Canceled || reads != before {
		t.Fatal("cancelled run performed work", e, reads, before)
	}
}

func TestExplanationSnapshotPromotionCustodyAndLegacyPool(t *testing.T) {
	now := time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC)
	bank := ResultSnapshots{now: func() time.Time { return now }}
	result := explanationFixture(t)
	token, e := bank.FreezeExplanation(result)
	if e != nil {
		t.Fatal(e)
	}
	result.Fields[0].Rows[0].BaselineCount = "mutated"
	held, e := bank.Load(token)
	if e != nil || held.Question.Kind != "explanation" || held.Explanation.Fields[0].Rows[0].BaselineCount == "mutated" {
		t.Fatal("held aggregates mutable", e)
	}
	store := CaptureStore{Dir: t.TempDir()}
	var wg sync.WaitGroup
	ids := make(chan string, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			saved, e := bank.PromoteExplanation(token, "Exact explanation", store)
			if e != nil {
				t.Error(e)
				return
			}
			ids <- saved.ID
		}()
	}
	wg.Wait()
	close(ids)
	first := ""
	for id := range ids {
		if first == "" {
			first = id
		} else if id != first {
			t.Fatal("lost-ACK retry created another custody receipt")
		}
	}
	saved, e := store.Load(first)
	if e != nil || saved.Explanation.ReadAt != explanationTestTime || saved.Explanation.ResultToken != "" || saved.Question.Explanation.Baseline.Offset != 50 {
		t.Fatal("save altered observed request/clocks", e)
	}
	page, e := saved.Page("")
	if e != nil || page.Explanation == nil || page.Explore != nil || page.RetainedRecords != 0 {
		t.Fatal("explanation masqueraded as record capture", e)
	}
	if _, e := store.Create(context.Background(), Source{}, "Must not reread", saved.Question, ""); e == nil {
		t.Fatal("generic capture path reran explanation")
	}
	bad := detached(*saved)
	bad.Explanation.Fields[0].Rows[0].DifferencePP = "0"
	if _, e := store.publish(bad); e == nil {
		t.Fatal("forged captured arithmetic published")
	}
	scope, e := saved.Question.Scope()
	if e != nil || scope.Family != "explanation" || len(scope.Members) != 2 || !reflect.DeepEqual(scope.ExplanationFields, saved.Question.Explanation.Fields) {
		t.Fatal("explanation scope lost pair/fields", e)
	}
	old := &Result{Question: basic()}
	oldToken, e := bank.Freeze(old)
	if e != nil {
		t.Fatal(e)
	}
	oldHeld, e := bank.Load(oldToken)
	if e != nil || oldHeld.Explore == nil || oldHeld.Explanation != nil {
		t.Fatal("old token path changed", e)
	}
	if _, e := bank.PromoteExplanation(oldToken, "Wrong kind", store); e == nil {
		t.Fatal("record token promoted as explanation")
	}
	now = now.Add(ResultSnapshotTTL)
	if _, e := bank.PromoteExplanation(token, "Exact explanation", store); e == nil {
		t.Fatal("expired token silently reused saved receipt")
	}
	if _, e := store.Load(first); e != nil {
		t.Fatal("expiry deleted immutable evidence", e)
	}
	// Both typed snapshots consume the original shared 64-entry budget.
	for i := 0; i < 64; i++ {
		if i%2 == 0 {
			_, e = bank.Freeze(old)
		} else {
			_, e = bank.FreezeExplanation(explanationFixture(t))
		}
		if e != nil {
			t.Fatal(e)
		}
	}
	if _, e := bank.FreezeExplanation(explanationFixture(t)); e == nil {
		t.Fatal("separate hidden explanation capacity")
	}
}

func TestExplanationCancellationStopsInflightFieldAndLaterReads(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reads := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reads++
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(string(body), "GROUP BY") {
			cancel()
			<-r.Context().Done()
			return
		}
		fmt.Fprintln(w, `{"count":"1"}`)
	}))
	defer server.Close()
	if _, e := (Source{Endpoint: server.URL}).Explain(ctx, explanationRequest()); e != context.Canceled || reads != 3 {
		t.Fatal("cancelled field run continued", e, reads)
	}
}

func TestExplanationPromotionReservesCustodyBeforeUncertainPublication(t *testing.T) {
	now := time.Date(2026, 10, 24, 12, 0, 0, 123, time.UTC)
	bank := ResultSnapshots{now: func() time.Time { return now }}
	result := explanationFixture(t)
	token, e := bank.FreezeExplanation(result)
	if e != nil {
		t.Fatal(e)
	}
	store := CaptureStore{Dir: t.TempDir()}
	attempts := 0
	candidateID, custody := "", ""
	bank.publishExplanation = func(store CaptureStore, result *ExplanationResult, title, at string) (*SavedCapture, error) {
		attempts++
		saved, e := store.saveExplanationAt(result, title, at)
		if e != nil {
			return nil, e
		}
		if attempts == 1 {
			candidateID, custody = saved.ID, saved.CapturedAt
			return nil, errors.New("injected acknowledgement/load failure after immutable final publication")
		}
		if saved.ID != candidateID || saved.CapturedAt != custody {
			t.Error("retry constructed a different custody candidate")
		}
		return saved, nil
	}
	if _, e := bank.PromoteExplanation(token, "Uncertain observation", store); e == nil {
		t.Fatal("injected publication uncertainty not exposed")
	}
	entries, e := os.ReadDir(store.Dir)
	if e != nil || len(entries) != 2 || entries[0].Name() != ".catalog-v2" || !entries[0].IsDir() || entries[1].Name() != candidateID+".json" {
		t.Fatal("first attempt did not actually publish one final file", e)
	}
	now = now.Add(30 * time.Second)
	ids := make(chan string, 8)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			saved, e := bank.PromoteExplanation(token, "Uncertain observation", store)
			if e != nil {
				t.Error(e)
				return
			}
			ids <- saved.ID
		}()
	}
	wg.Wait()
	close(ids)
	for id := range ids {
		if id != candidateID {
			t.Fatal("uncertain retry produced another ID", id, candidateID)
		}
	}
	if attempts != 2 {
		t.Fatal("successful retries republished", attempts)
	}
	entries, e = os.ReadDir(store.Dir)
	if e != nil || len(entries) != 2 || entries[0].Name() != ".catalog-v2" || !entries[0].IsDir() || entries[1].Name() != candidateID+".json" {
		t.Fatal("uncertain retry created another file or leaked scratch", e, len(entries))
	}
	saved, e := store.Load(candidateID)
	if e != nil || saved.CapturedAt != custody || saved.Explanation.Baseline.Count != result.Baseline.Count || saved.Explanation.ReadAt != result.ReadAt {
		t.Fatal("retry changed observed aggregates or custody", e)
	}
}

func TestExplanationUncertainReservationsHaveBoundedCapacity(t *testing.T) {
	bank := ResultSnapshots{}
	token, e := bank.FreezeExplanation(explanationFixture(t))
	if e != nil {
		t.Fatal(e)
	}
	store := CaptureStore{Dir: t.TempDir()}
	attempts := 0
	bank.publishExplanation = func(CaptureStore, *ExplanationResult, string, string) (*SavedCapture, error) {
		attempts++
		return nil, errors.New("injected uncertain failure")
	}
	for i := 0; i < 16; i++ {
		if _, e := bank.PromoteExplanation(token, fmt.Sprintf("Reserved %d", i), store); e == nil {
			t.Fatal("fault not injected")
		}
	}
	if len(bank.entries[token].promoted) != 16 {
		t.Fatal("failed attempts not reserved")
	}
	if _, e := bank.PromoteExplanation(token, "One too many", store); e == nil || attempts != 16 {
		t.Fatal("unbounded reservation metadata", attempts, e)
	}
	bank.publishExplanation = nil
	if _, e := bank.PromoteExplanation(token, "Reserved 0", store); e != nil {
		t.Fatal("full reservation capacity prevented same-candidate retry", e)
	}
}

func TestExplanationSourceRejectsAbsentOrMalformedGroupKeys(t *testing.T) {
	request := explanationRequest()
	request.Fields = request.Fields[:1]
	for _, group := range []string{`{"count":"1"}`, `{"g0":null}`, `{"g0":null,"count":1}`, `{"g0":3,"count":"1"}`} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			if strings.Contains(string(body), "GROUP BY") {
				fmt.Fprintln(w, group)
			} else {
				fmt.Fprintln(w, `{"count":"1"}`)
			}
		}))
		r, e := (Source{Endpoint: server.URL}).Explain(context.Background(), request)
		server.Close()
		if e != nil || r.Fields[0].State != "unavailable" || r.Fields[0].Reason != "invalid-source-groups" || len(r.Fields[0].Rows) != 0 || len(r.Ranked) != 0 {
			t.Fatal("malformed group interpreted as missing/empty", group, r, e)
		}
	}
}
