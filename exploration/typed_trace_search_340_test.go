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

	"github.com/casselc/goo/identity"
)

func typedTraceRecords340() []TraceSearchRecord {
	records := []TraceSearchRecord{}
	appendRecord := func(id int, service string, resources map[string]string) {
		fields := map[Field]string{{Kind: "column", Key: "trace_id"}: fmt.Sprintf("%032x", id), {Kind: "column", Key: "span_id"}: fmt.Sprintf("%016x", len(records)+1), {Kind: "column", Key: "service"}: service, {Kind: "column", Key: "name"}: "ordinary operation", {Kind: "column", Key: "kind"}: "Server", {Kind: "column", Key: "timestamp_ns"}: "1000000000", {Kind: "column", Key: "duration_ns"}: "0"}
		for key, value := range resources {
			fields[Field{Kind: "resource", Key: key}] = value
		}
		records = append(records, TraceSearchRecord{Fields: fields})
	}
	for i, resources := range []map[string]string{
		{identity.EnvironmentKey: "prod", identity.LegacyEnvironmentKey: "other"},
		{identity.LegacyEnvironmentKey: "prod"},
		{identity.EnvironmentKey: "other", identity.LegacyEnvironmentKey: "prod"},
		{identity.EnvironmentKey: "", identity.LegacyEnvironmentKey: "prod"},
		{},
		{identity.LegacyEnvironmentKey: ""},
		{identity.EnvironmentKey: "prod", identity.LegacyEnvironmentKey: ""},
	} {
		appendRecord(i+1, "db", resources)
		appendRecord(i+1, "worker", resources)
	}
	// Different spans and services may satisfy independent EXISTS conditions while
	// each record must still obey the requested environment policy.
	appendRecord(8, "db", map[string]string{identity.EnvironmentKey: "prod"})
	appendRecord(8, "worker", map[string]string{identity.LegacyEnvironmentKey: "prod"})
	return records
}
func typedTraceRequest340(policy, mode, value string) TraceSearchRequest {
	return TraceSearchRequest{Version: TraceSearchVersion, EnvironmentSelector: &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: policy, Scope: identity.EnvironmentScope{Mode: mode, Value: value}}, FromMS: 1000, ToMS: 61000, Limit: 100, Conditions: []Predicate{{Field: &Field{Kind: "column", Key: "service"}, Op: "eq", Value: "db"}, {Field: &Field{Kind: "column", Key: "service"}, Op: "eq", Value: "worker"}}}
}
func traceSearchWire340(items []TraceSearchItem, total string) string {
	var lines strings.Builder
	for _, item := range items {
		row := map[string]any{"trace_id": item.TraceID, "matched_traces": total}
		for _, w := range item.Witnesses {
			fields := make([]any, 0, len(w.FieldValues))
			for _, f := range w.FieldValues {
				if f.Value == nil {
					fields = append(fields, []any{0, ""})
				} else {
					fields = append(fields, []any{1, *f.Value})
				}
			}
			row[fmt.Sprintf("c%d", w.ConditionIndex)] = w.MatchingRecords
			row[fmt.Sprintf("w%d", w.ConditionIndex)] = []any{w.TimestampNS, w.SpanID, w.ParentSpanID, w.Service, w.Operation, w.Kind, w.DurationNS, fields}
		}
		raw, _ := json.Marshal(row)
		lines.Write(raw)
		lines.WriteByte('\n')
	}
	return lines.String()
}
func TestTraceSearchTypedEnvironmentPartitions340(t *testing.T) {
	for _, tc := range []struct {
		policy, mode, value, total string
		ids                        []int
	}{
		{"current-only", "named", "prod", "2", []int{1, 7}},
		{"effective", "named", "prod", "4", []int{1, 2, 7, 8}},
		{"effective", "empty", "", "2", []int{4, 6}},
		{"effective", "unreported", "", "1", []int{5}},
	} {
		t.Run(tc.policy+"/"+tc.mode, func(t *testing.T) {
			request := typedTraceRequest340(tc.policy, tc.mode, tc.value)
			items, total, err := ReferenceTraceSearch(request, typedTraceRecords340())
			if err != nil || total != tc.total {
				t.Fatal(items, total, err)
			}
			ids := []string{}
			for _, item := range items {
				ids = append(ids, item.TraceID)
				for _, w := range item.Witnesses {
					if len(w.FieldValues) != 3 {
						t.Fatal("condition plus both environment facts required", w)
					}
				}
			}
			expected := []string{}
			for _, id := range tc.ids {
				expected = append(expected, fmt.Sprintf("%032x", id))
			}
			if !reflect.DeepEqual(ids, expected) {
				t.Fatal(ids, expected)
			}
			reads := 0
			source := Source{Endpoint: "http://unused", Client: &http.Client{Transport: traceSearchTransport(func(*http.Request) (*http.Response, error) {
				reads++
				return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(traceSearchWire340(items, total)))}, nil
			})}}
			result, err := source.SearchTraces(context.Background(), request)
			if err != nil || reads != 1 || !reflect.DeepEqual(result.Request, request) || !reflect.DeepEqual(result.Items, items) {
				t.Fatal(result, reads, err)
			}
			if request.EnvironmentSelector.Scope.Mode != "named" {
				if _, exists := result.Plan.Parameters["environment"]; exists {
					t.Fatal("nameless selector acquired scalar binding")
				}
			}
			encoded, _ := json.Marshal(result.Request)
			var wire map[string]any
			_ = json.Unmarshal(encoded, &wire)
			if _, exists := wire["environment"]; exists {
				t.Fatal("typed echo projected into scalar environment")
			}
		})
	}
}
func TestTraceSearchWitnessEnvironmentProof340(t *testing.T) {
	request := typedTraceRequest340("effective", "named", "prod")
	items, total, err := ReferenceTraceSearch(request, typedTraceRecords340())
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"wrong-current-shadows-legacy", "missing-current-legacy-other", "missing-environment-proof"} {
		t.Run(name, func(t *testing.T) {
			bad := detached(items)
			for i := range bad[0].Witnesses {
				values := bad[0].Witnesses[i].FieldValues
				for j := range values {
					switch values[j].Field.Key {
					case identity.EnvironmentKey:
						if name == "wrong-current-shadows-legacy" {
							value := "wrong"
							values[j].Value = &value
						} else {
							values[j].Value = nil
						}
					case identity.LegacyEnvironmentKey:
						if name == "wrong-current-shadows-legacy" {
							value := "prod"
							values[j].Value = &value
						} else {
							value := "other"
							values[j].Value = &value
						}
					}
				}
				if name == "missing-environment-proof" {
					bad[0].Witnesses[i].FieldValues = values[:1]
				}
			}
			source := Source{Endpoint: "http://unused", Client: &http.Client{Transport: traceSearchTransport(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(traceSearchWire340(bad, total)))}, nil
			})}}
			if _, err := source.SearchTraces(context.Background(), request); err == nil {
				t.Fatal("source witness weakened requested environment policy")
			}
		})
	}
}
func TestTraceSearchDetachesTypedRequestBeforeRead340(t *testing.T) {
	request := typedTraceRequest340("effective", "named", "prod")
	original := detached(request)
	records := typedTraceRecords340()
	items, total, err := ReferenceTraceSearch(request, records)
	if err != nil {
		t.Fatal(err)
	}
	source := Source{Endpoint: "http://unused", Client: &http.Client{Transport: traceSearchTransport(func(*http.Request) (*http.Response, error) {
		request.EnvironmentSelector.Scope.Value = "mutated"
		request.Conditions[0].Value = "mutated"
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(traceSearchWire340(items, total)))}, nil
	})}}
	result, err := source.SearchTraces(context.Background(), request)
	if err != nil || !reflect.DeepEqual(result.Request, original) {
		t.Fatal(result, err)
	}
}
