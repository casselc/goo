package exploration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func tracePage403Records() []TraceSearchRecord {
	template := traceSearchTestRecords()[:2]
	rows := []TraceSearchRecord{}
	for i := 1; i <= 7; i++ {
		for _, record := range template {
			fields := maps.Clone(record.Fields)
			fields[Field{Kind: "column", Key: "trace_id"}] = fmt.Sprintf("%032x", i)
			rows = append(rows, TraceSearchRecord{Fields: fields})
		}
	}
	return rows
}
func tracePage403Body(t *testing.T, r TraceSearchPageRequest) string {
	t.Helper()
	all := r.Request
	all.Limit = 100
	items, total, err := ReferenceTraceSearch(all, tracePage403Records())
	if err != nil {
		t.Fatal(err)
	}
	after := []TraceSearchItem{}
	for _, item := range items {
		if item.TraceID > r.Position.LastTraceID {
			after = append(after, item)
		}
	}
	remaining := fmt.Sprint(len(after))
	if len(after) > r.Request.Limit+1 {
		after = after[:r.Request.Limit+1]
	}
	lines := []string{}
	if len(after) == 0 {
		raw := map[string]any{"trace_id": "", "matched_traces": total, "remaining_traces": remaining, "page_present": 0}
		for i := range r.Request.Conditions {
			raw[fmt.Sprint("c", i)] = ""
			raw[fmt.Sprint("w", i)] = []any{"", "", "", "", "", "", "", []any{}}
		}
		encoded, _ := json.Marshal(raw)
		return string(encoded)
	}
	for _, item := range after {
		raw := map[string]any{"trace_id": item.TraceID, "matched_traces": total, "remaining_traces": remaining, "page_present": 1}
		for _, w := range item.Witnesses {
			values := []any{}
			for _, f := range w.FieldValues {
				present, value := 0, ""
				if f.Value != nil {
					present, value = 1, *f.Value
				}
				values = append(values, []any{present, value})
			}
			raw[fmt.Sprint("c", w.ConditionIndex)] = w.MatchingRecords
			raw[fmt.Sprint("w", w.ConditionIndex)] = []any{w.TimestampNS, w.SpanID, w.ParentSpanID, w.Service, w.Operation, w.Kind, w.DurationNS, values}
		}
		encoded, _ := json.Marshal(raw)
		lines = append(lines, string(encoded))
	}
	return strings.Join(lines, "\n")
}
func tracePage403Source(body *string) (Source, *int) {
	reads := new(int)
	return Source{Endpoint: "http://unused", Client: &http.Client{Transport: traceSearchTransport(func(*http.Request) (*http.Response, error) {
		*reads++
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(*body)), Header: http.Header{}}, nil
	})}}, reads
}
func tracePage403Request(t *testing.T, last int) TraceSearchPageRequest {
	t.Helper()
	r := traceSearchTestRequest()
	r.Limit = 2
	p, err := NextTraceSearchPosition(r, fmt.Sprintf("%032x", last))
	if err != nil {
		t.Fatal(err)
	}
	return TraceSearchPageRequest{Request: r, Position: *p}
}
func TestTraceSearchPage403FullRemainingAndEmpty(t *testing.T) {
	request := tracePage403Request(t, 2)
	body := tracePage403Body(t, request)
	source, reads := tracePage403Source(&body)
	got, err := source.SearchTracePage(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if *reads != 1 || got.MatchedTraces != "7" || got.RemainingTraces != "5" || len(got.Items) != 2 || got.Items[0].TraceID != fmt.Sprintf("%032x", 3) || !got.HasMore || got.NextPosition.LastTraceID != fmt.Sprintf("%032x", 4) {
		t.Fatalf("wrong continuation %#v", got)
	}
	if err = ValidateTraceSearchResult(got); err != nil {
		t.Fatal(err)
	}
	reference, total, remaining, err := ReferenceTraceSearchPage(request, tracePage403Records())
	if err != nil || total != "7" || remaining != "5" || !reflect.DeepEqual(reference, got.Items) {
		t.Fatal("reference differs", err)
	}
	if !strings.Contains(got.Plan.SQL, "countIf(trace_id > {trace_after:String})") || !strings.Contains(got.Plan.SQL, "LEFT JOIN page ON 1") || got.Plan.Parameters["trace_after"] != request.Position.LastTraceID {
		t.Fatal("missing bound page plan")
	}
	end := tracePage403Request(t, 7)
	body = tracePage403Body(t, end)
	empty, err := source.SearchTracePage(context.Background(), end)
	if err != nil || empty.MatchedTraces != "7" || empty.RemainingTraces != "0" || empty.State != "no_data" || empty.HasMore || empty.NextPosition != nil || len(empty.Items) != 0 {
		t.Fatal("empty page lost full total", empty, err)
	}
	if err = ValidateTraceSearchResult(empty); err != nil {
		t.Fatal(err)
	}
	body = strings.Replace(body, `"matched_traces":"7"`, `"matched_traces":"8"`, 1)
	changed, err := source.SearchTracePage(context.Background(), end)
	if err != nil || changed.Basis == empty.Basis {
		t.Fatal("metadata change did not advance basis", err)
	}
}
func TestTraceSearchPage403StrictSourceAndPosition(t *testing.T) {
	request := tracePage403Request(t, 7)
	validBody := tracePage403Body(t, request)
	controls := []string{"", strings.Replace(validBody, `"page_present":0`, `"page_present":1`, 1), strings.Replace(validBody, `"remaining_traces":"0"`, `"remaining_traces":"1"`, 1), strings.Replace(validBody, `"c0":""`, `"c0":"1"`, 1), strings.Replace(validBody, `"trace_id":""`, `"trace_id":"00000000000000000000000000000007"`, 1), strings.Replace(validBody, `"matched_traces":"7"`, `"matched_traces":"01"`, 1)}
	for _, body := range controls {
		source, _ := tracePage403Source(&body)
		if got, err := source.SearchTracePage(context.Background(), request); err == nil {
			t.Fatalf("invalid source accepted %s %#v", body, got)
		}
	}
	for _, mutate := range []func(*TraceSearchPageRequest){func(r *TraceSearchPageRequest) { r.Request.FromMS++ }, func(r *TraceSearchPageRequest) { r.Request.Limit++ }, func(r *TraceSearchPageRequest) { r.Request.Conditions[0].All[0].Value = "other" }, func(r *TraceSearchPageRequest) { r.Position.LastTraceID = strings.Repeat("0", 32) }, func(r *TraceSearchPageRequest) { r.Position.RequestHash = strings.Repeat("f", 64) }} {
		bad := detached(request)
		mutate(&bad)
		if bad.Validate() == nil {
			t.Fatal("foreign position accepted")
		}
	}
	encoded, _ := json.Marshal(request)
	for _, raw := range []string{strings.Replace(string(encoded), `"lastTraceId":`, `"LastTraceId":`, 1), strings.Replace(string(encoded), `"position":{`, `"position":{"extra":true,`, 1), strings.Replace(string(encoded), `"request":{`, `"request":null,"ignored":{`, 1)} {
		var bad TraceSearchPageRequest
		if json.Unmarshal([]byte(raw), &bad) == nil {
			t.Fatal("malformed page accepted", raw)
		}
	}
}
func TestTraceSearchPage403CaptureCustodyAndRefresh(t *testing.T) {
	request := tracePage403Request(t, 2)
	body := tracePage403Body(t, request)
	source, reads := tracePage403Source(&body)
	store := CaptureStore{Dir: t.TempDir()}
	q := CaptureQuestion{Kind: "trace-search", TraceSearch: &request.Request, TraceSearchPosition: &request.Position}
	saved, err := store.Create(context.Background(), source, "Second page", q, "")
	if err != nil {
		t.Fatal(err)
	}
	if *reads != 1 || !reflect.DeepEqual(saved.TraceSearch.Position, &request.Position) || saved.TraceSearch.Items[0].TraceID != fmt.Sprintf("%032x", 3) {
		t.Fatal("capture read different page", saved)
	}
	before, _ := json.Marshal(saved)
	reopened, err := store.Load(saved.ID)
	if err != nil {
		t.Fatal(err)
	}
	page, err := reopened.Page("")
	if err != nil || *reads != 1 || !reflect.DeepEqual(page.Receipt.Question.TraceSearchPosition, &request.Position) {
		t.Fatal("offline reopen changed position", err)
	}
	child, err := store.Create(context.Background(), source, "Refresh page", saved.Question, saved.ID)
	if err != nil || *reads != 2 || child.Parent != saved.ID || !reflect.DeepEqual(child.Question.TraceSearchPosition, &request.Position) {
		t.Fatal("refresh changed position", err)
	}
	again, err := store.Load(saved.ID)
	after, _ := json.Marshal(again)
	if err != nil || string(before) != string(after) {
		t.Fatal("refresh rewrote original", err)
	}
	bad := detached(*page)
	bad.Receipt.Question.TraceSearchPosition.LastTraceID = fmt.Sprintf("%032x", 1)
	if validateCapturePage(bad) == nil {
		t.Fatal("wrong page receipt accepted")
	}
	foreign := CaptureQuestion{Kind: "explore", Explore: &Query{}, TraceSearchPosition: &request.Position}
	if foreign.Validate() == nil {
		t.Fatal("wrong capture family admits position")
	}
}
func TestTraceSearchPage403CapturePositionWire(t *testing.T) {
	request := tracePage403Request(t, 2)
	q := CaptureQuestion{Kind: "trace-search", TraceSearch: &request.Request, TraceSearchPosition: &request.Position}
	raw, _ := json.Marshal(q)
	var decoded CaptureQuestion
	if err := json.Unmarshal(raw, &decoded); err != nil || decoded.Validate() != nil || !reflect.DeepEqual(q, decoded) {
		t.Fatal("valid page question not admitted", err)
	}
	var object map[string]any
	json.Unmarshal(raw, &object)
	object["traceSearchPosition"] = nil
	nullRaw, _ := json.Marshal(object)
	if json.Unmarshal(nullRaw, &decoded) == nil {
		t.Fatal("explicit null position accepted")
	}
	for _, kind := range []string{"explore", "inventory", "measurement", "explanation", "comparison"} {
		wrong := map[string]any{"kind": kind, "traceSearchPosition": request.Position, kind: map[string]any{}}
		encoded, _ := json.Marshal(wrong)
		if json.Unmarshal(encoded, &decoded) == nil {
			t.Fatal("other family admits trace position", kind)
		}
	}
}
