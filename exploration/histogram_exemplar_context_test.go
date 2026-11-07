package exploration

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
)

func exemplarContextFixture377(t *testing.T, trace, span, sample string) (*measurements.Result, metrics.Stream, string) {
	t.Helper()
	d := dashboardTestInstrument("exemplar", "custom.duration", "histogram", "s").Definition
	d.Grain = "per-stream"
	d.Source.Attributes = map[string]string{"metric.only": "never-a-span-predicate"}
	parent := measurements.Question{Definition: *d, FromMS: 1790000000000, ToMS: 1790000006000, FromNS: "1790000000000000123", ToNS: "1790000005999999877"}
	stream := metrics.Stream{Resource: map[string]string{"service.name": d.Source.Services[0], "service.instance.id": "instance-one", "deployment.environment.name": d.Environment}, Scope: d.Source.Scope, Version: d.Source.Version, SchemaURL: d.Source.SchemaURL, Name: d.Source.Metric, Unit: "s", Attributes: map[string]string{"metric.only": "never-a-span-predicate"}}
	end := "1790000003000000000"
	raw := exemplarContextRow377(stream, end, trace, span, sample)
	result, err := measurements.Evaluate(parent, []json.RawMessage{raw}, time.Date(2026, 11, 7, 0, 0, 0, 123, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	return result, stream, end
}
func exemplarContextRow377(stream metrics.Stream, end, trace, span, sample string) json.RawMessage {
	raw, _ := json.Marshal(map[string]any{"resource_schema_url": stream.ResourceSchemaURL, "resource": stream.Resource, "scope": stream.Scope, "version": stream.Version, "schema_url": stream.SchemaURL, "name": stream.Name, "unit": stream.Unit, "attributes": stream.Attributes, "start_ns": "1790000001000000000", "end_ns": end, "temporality": 1, "count": "1", "bucket_counts": []string{"0", "1", "0", "0"}, "bounds": []float64{0, .1, 1}, "flags": 0, "exemplar_total": "1", "exemplar_span_total": "1", "exemplar_time_total": "1", "exemplar_value_total": "1", "exemplar_attribute_total": "1", "exemplar_times": []string{sample}, "exemplar_values": []string{"0.05"}, "exemplar_trace_ids": []string{trace}, "exemplar_span_ids": []string{span}})
	return raw
}
func TestHistogramExemplarContext377(t *testing.T) {
	trace, span := strings.Repeat("a", 32), strings.Repeat("b", 16)
	for _, reportedSpan := range []string{span, "", strings.Repeat("0", 16), "invalid"} {
		t.Run("span-"+reportedSpan, func(t *testing.T) {
			result, stream, end := exemplarContextFixture377(t, trace, reportedSpan, "0")
			before, _ := json.Marshal(result)
			got, err := PlanHistogramExemplarContext(result, stream, end, 0)
			if err != nil {
				t.Fatal(err)
			}
			after, _ := json.Marshal(result)
			if string(before) != string(after) {
				t.Fatal("projection mutated retained evidence")
			}
			if got.Relation != "reported-exemplar" || got.TraceID != trace || got.TimeNS != "0" || got.FromNS != result.Question.FromNS || got.ToNS != result.Question.ToNS || got.ExportEndNS != end || got.ExemplarIndex != 0 || got.SourceRow != 0 || got.Question.FromMS != result.Question.FromMS || got.Question.ToMS != result.Question.ToMS {
				t.Fatal("origin, export, sample and original question were conflated", got)
			}
			if (got.SpanID == span) != (reportedSpan == span) {
				t.Fatal("invalid span identity became a focus hint")
			}
			if len(got.Question.Where.All) != 4 || got.Question.Where.All[0].Field.Key != "trace_id" || got.Question.Where.All[1].Field.Key != identity.NamespaceKey || got.Question.Where.All[1].Op != "missing" || got.Question.Where.All[2].Value != got.FromNS || got.Question.Where.All[3].Value != "1790000005999999876" {
				t.Fatal("trace query copied metric selectors or lost exact scope")
			}
			if _, err = Compile(got.Question); err != nil {
				t.Fatal(err)
			}
		})
	}
	for _, trace := range []string{"", strings.Repeat("0", 32), strings.Repeat("A", 32), "not-a-trace"} {
		result, stream, end := exemplarContextFixture377(t, trace, span, "1790000002000000000")
		if _, err := PlanHistogramExemplarContext(result, stream, end, 0); err == nil {
			t.Fatal("invalid reported trace authorized staging", trace)
		}
	}
	result, stream, end := exemplarContextFixture377(t, trace, span, "1790000002000000000")
	for _, ordinal := range []int{-1, 1, 999} {
		if _, err := PlanHistogramExemplarContext(result, stream, end, ordinal); err == nil {
			t.Fatal("unreported ordinal authorized staging")
		}
	}
	if _, err := PlanHistogramExemplarContext(result, stream, "1790000003000000001", 0); err == nil {
		t.Fatal("wrong export authorized staging")
	}
}
func TestHistogramExemplarContextOmittedDuplicate377(t *testing.T) {
	result, stream, end := exemplarContextFixture377(t, strings.Repeat("a", 32), strings.Repeat("b", 16), "1790000002000000000")
	rows := append([]json.RawMessage{}, result.Rows...)
	for i := 0; i < 120; i++ {
		noise := stream
		noise.Resource = map[string]string{}
		for key, value := range stream.Resource {
			noise.Resource[key] = value
		}
		noise.Resource["service.instance.id"] = fmt.Sprint(i)
		noise.Resource["large.literal"] = strings.Repeat("x", 1024)
		rows = append(rows, exemplarContextRow377(noise, end, strings.Repeat("a", 32), "", strconv.Itoa(i)))
	}
	rows = append(rows, result.Rows[0])
	duplicate, err := measurements.Evaluate(result.Question, rows, time.Date(2026, 11, 7, 0, 0, 0, 123, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if duplicate.HistogramExemplars == nil || !duplicate.HistogramExemplars.Limited || duplicate.HistogramExemplars.OmittedExports == 0 {
		t.Fatal("fixture did not cross the real metadata budget")
	}
	if _, err = PlanHistogramExemplarContext(duplicate, stream, end, 0); err == nil {
		t.Fatal("budget-omitted duplicate became unique by filtering the report")
	}
	copy := *result
	copy.HistogramExemplars = &measurements.HistogramExemplarReport{}
	if _, err = PlanHistogramExemplarContext(&copy, stream, end, 0); err == nil {
		t.Fatal("invented report bypassed raw-result replay")
	}
	if reflect.DeepEqual(copy.HistogramExemplars, result.HistogramExemplars) {
		t.Fatal("tamper fixture did not differ")
	}
}

func TestHistogramExemplarContextUnavailable377(t *testing.T) {
	for _, mode := range []string{"metadata-missing", "metadata-invalid", "known-empty", "no-recorded-value", "warmup-export"} {
		t.Run(mode, func(t *testing.T) {
			result, stream, end := exemplarContextFixture377(t, strings.Repeat("a", 32), strings.Repeat("b", 16), "1790000002000000000")
			var row map[string]any
			if err := json.Unmarshal(result.Rows[0], &row); err != nil {
				t.Fatal(err)
			}
			switch mode {
			case "metadata-missing":
				for key := range row {
					if strings.HasPrefix(key, "exemplar_") {
						delete(row, key)
					}
				}
			case "metadata-invalid":
				row["exemplar_span_ids"] = []string{}
			case "known-empty":
				for key := range row {
					if strings.HasPrefix(key, "exemplar_") {
						if strings.HasSuffix(key, "_total") {
							row[key] = "0"
						} else {
							row[key] = []string{}
						}
					}
				}
			case "no-recorded-value":
				row["flags"] = 1
			case "warmup-export":
				end = "1790000000000000000"
				row["end_ns"] = end
				row["start_ns"] = "1789999999000000000"
			}
			raw, err := json.Marshal(row)
			if err != nil {
				t.Fatal(err)
			}
			changed, err := measurements.Evaluate(result.Question, []json.RawMessage{raw}, time.Date(2026, 11, 7, 0, 0, 0, 123, time.UTC))
			if err != nil {
				t.Fatal(err)
			}
			if _, err = PlanHistogramExemplarContext(changed, stream, end, 0); err == nil {
				t.Fatal("unavailable reference authorized trace staging")
			}
		})
	}
}
