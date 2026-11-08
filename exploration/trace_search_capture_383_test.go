package exploration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func traceCapture383Source(t *testing.T, request TraceSearchRequest) (Source, *int) {
	t.Helper()
	full := request
	full.Limit = 100
	items, total, err := ReferenceTraceSearch(full, traceSearchTestRecords())
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	for _, item := range items {
		raw := map[string]any{"trace_id": item.TraceID, "matched_traces": total}
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
		b, _ := json.Marshal(raw)
		lines = append(lines, string(b))
	}
	reads := new(int)
	return Source{Endpoint: "http://unused", Client: &http.Client{Transport: traceSearchTransport(func(*http.Request) (*http.Response, error) {
		*reads++
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(strings.Join(lines, "\n"))), Header: http.Header{}}, nil
	})}}, reads
}
func TestTraceSearchCapture383ReadStoreReopenRefresh(t *testing.T) {
	q := traceSearchTestRequest()
	source, reads := traceCapture383Source(t, q)
	store := CaptureStore{Dir: t.TempDir()}
	saved, err := store.Create(context.Background(), source, "Search evidence", CaptureQuestion{Kind: "trace-search", TraceSearch: &q}, "")
	if err != nil {
		t.Fatal(err)
	}
	if *reads != 1 || saved.TraceSearch.MatchedTraces != "2" || !saved.TraceSearch.HasMore || len(saved.TraceSearch.Items) != 1 || saved.TraceSearch.ReadAt == "" || saved.CapturedAt == "" {
		t.Fatal("search evidence lost", saved, reads)
	}
	first, _ := json.Marshal(saved)
	reopened, err := store.Load(saved.ID)
	if err != nil {
		t.Fatal(err)
	}
	page, err := reopened.Page("")
	if err != nil || page.TraceSearch == nil || page.RetainedRecords != 1 || !page.RecordsLimited || *reads != 1 {
		t.Fatal("offline page changed", page, err)
	}
	summary, err := store.Search(context.Background(), CaptureSearchQuery{Kind: "trace-search", Signal: "spans"}, "")
	if err != nil || summary.Matched != 1 {
		t.Fatal(summary, err)
	}
	child, err := store.Create(context.Background(), source, "Search refresh", saved.Question, saved.ID)
	if err != nil || child.Parent != saved.ID || *reads != 2 {
		t.Fatal(child, err)
	}
	again, err := store.Load(saved.ID)
	after, _ := json.Marshal(again)
	if err != nil || string(first) != string(after) {
		t.Fatal("refresh rewrote immutable evidence", err)
	}
	// Reopening retains an older compiler's checked plan; only refresh recompiles.
	old := detached(*saved)
	old.Parent = ""
	old.TraceSearch.Plan.SQL = "historical checked plan"
	historical, err := store.publish(old)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.Load(historical.ID); err != nil {
		t.Fatal("old evidence not reopenable", err)
	}
	if _, err = store.Create(context.Background(), source, "Changed compiler", historical.Question, historical.ID); err != ErrCapturePlanChanged || *reads != 2 {
		t.Fatal("changed plan reread", err)
	}
}

func TestTraceSearchCapture383ResultControls(t *testing.T) {
	q := traceSearchTestRequest()
	q.Limit = 100
	source, _ := traceCapture383Source(t, q)
	valid, err := source.SearchTraces(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	if err = ValidateTraceSearchResult(valid); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*TraceSearchResult){
		"full-count": func(r *TraceSearchResult) { r.MatchedTraces = "1" },
		"has-more":   func(r *TraceSearchResult) { r.HasMore = true },
		"condition":  func(r *TraceSearchResult) { r.Items[0].Witnesses[0].ConditionIndex = 1 },
		"membership": func(r *TraceSearchResult) {
			for _, v := range r.Items[0].Witnesses[1].FieldValues {
				if v.Field.Key == "status_class" {
					*v.Value = "ok"
				}
			}
		},
		"metadata": func(r *TraceSearchResult) { r.Items[0].Witnesses[0].Service = "wrong" },
		"environment": func(r *TraceSearchResult) {
			for _, v := range r.Items[0].Witnesses[0].FieldValues {
				if v.Field.Key == "deployment.environment.name" {
					*v.Value = "wrong"
				}
			}
		},
		"field-namespace": func(r *TraceSearchResult) { r.Items[0].Witnesses[0].FieldValues[0].Field.Kind = "attribute" },
		"count-overflow":  func(r *TraceSearchResult) { r.Items[0].Witnesses[0].MatchingRecords = "18446744073709551616" },
		"read-clock":      func(r *TraceSearchResult) { r.ReadAt = "not-time" },
		"missing-warning": func(r *TraceSearchResult) { r.Warnings = nil },
	} {
		t.Run(name, func(t *testing.T) {
			r := detached(*valid)
			mutate(&r)
			if ValidateTraceSearchResult(&r) == nil {
				t.Fatal("invalid evidence accepted")
			}
		})
	}
	if valid.Items[1].Witnesses[0].SpanID != valid.Items[1].Witnesses[1].SpanID {
		t.Fatal("shared-span fixture lost")
	}
}
func TestTraceSearchCapture383StrictWire(t *testing.T) {
	request, _ := json.Marshal(traceSearchTestRequest())
	wire := `{"kind":"trace-search","traceSearch":` + string(request) + `}`
	var q CaptureQuestion
	if err := json.Unmarshal([]byte(wire), &q); err != nil || q.Validate() != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{strings.Replace(wire, `"traceSearch":`, `"explore":null,"traceSearch":`, 1), strings.Replace(wire, `"conditions":`, `"Conditions":`, 1), strings.Replace(wire, `"limit":1`, `"limit":null`, 1), strings.Replace(wire, `"field":`, `"field":null,"field":`, 1), strings.Replace(wire, `"fromMs":1000,`, "", 1)} {
		// Earlier duplicate null is replaced by a valid last field under JSON semantics.
		if strings.Contains(bad, `"field":null,"field":`) {
			if err := json.Unmarshal([]byte(bad), &q); err != nil {
				t.Fatal("last-value semantics lost", err)
			}
			continue
		}
		if err := json.Unmarshal([]byte(bad), &q); err == nil && q.Validate() == nil {
			t.Fatal("bad wire admitted", bad)
		}
	}
	duplicate := strings.Replace(wire, `"traceSearch":`, `"traceSearch":{"environment":"wrong"},"traceSearch":`, 1)
	if err := json.Unmarshal([]byte(duplicate), &q); err != nil || !reflect.DeepEqual(q.TraceSearch, func() *TraceSearchRequest { r := traceSearchTestRequest(); return &r }()) {
		t.Fatal("duplicate object merge", q, err)
	}
}

func TestTraceSearchCapture383MissingVersusEmpty(t *testing.T) {
	q := traceSearchTestRequest()
	q.Limit = 100
	empty, missing := Field{Kind: "resource", Key: "test.empty"}, Field{Kind: "attribute", Key: "test.missing"}
	q.Conditions = []Predicate{{All: []Predicate{{Field: &empty, Op: "eq", Value: ""}, {Field: &missing, Op: "missing"}}}}
	records := traceSearchTestRecords()
	for _, r := range records {
		r.Fields[empty] = ""
	}
	items, total, err := ReferenceTraceSearch(q, records)
	if err != nil {
		t.Fatal(err)
	}
	plan, _ := CompileTraceSearch(q)
	result := newTraceSearchResult(q, plan)
	result.Items = items
	result.MatchedTraces = total
	result.State = "observed"
	result.ReadAt = "2026-11-08T22:00:00Z"
	result.Basis = strings.Repeat("a", 64)
	if err = ValidateTraceSearchResult(result); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"test.empty", "test.missing"} {
		r := detached(*result)
		for i := range r.Items[0].Witnesses[0].FieldValues {
			fv := &r.Items[0].Witnesses[0].FieldValues[i]
			if fv.Field.Key == key {
				if fv.Value == nil {
					value := ""
					fv.Value = &value
				} else {
					fv.Value = nil
				}
			}
		}
		if ValidateTraceSearchResult(&r) == nil {
			t.Fatal("missing/empty conflated", key)
		}
	}
}
