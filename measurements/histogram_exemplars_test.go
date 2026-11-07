package measurements

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

func histogramExemplarRow377(t *testing.T, total int) (Question, map[string]any) {
	t.Helper()
	base := histogramViewFixture(t)
	var row map[string]any
	if err := json.Unmarshal(base.Rows[0], &row); err != nil {
		t.Fatal(err)
	}
	for _, key := range histogramExemplarFields[:5] {
		row[key] = strconv.Itoa(total)
	}
	times, values, traces, spans := []string{}, []string{}, []string{}, []string{}
	for range min(total, HistogramExemplarLimit) {
		times = append(times, "100500000123")
		values = append(values, "0.125")
		traces = append(traces, strings.Repeat("1", 32))
		spans = append(spans, strings.Repeat("2", 16))
	}
	row["exemplar_times"], row["exemplar_values"], row["exemplar_trace_ids"], row["exemplar_span_ids"] = times, values, traces, spans
	return base.Question, row
}
func histogramExemplarResult377(t *testing.T, q Question, rows ...map[string]any) *Result {
	t.Helper()
	wire := make([]json.RawMessage, len(rows))
	for i, row := range rows {
		wire[i] = raw(row)
	}
	r, e := Evaluate(q, wire, time.Unix(200, 0))
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func TestHistogramExemplar377RetentionAndCaps(t *testing.T) {
	base := histogramViewFixture(t)
	if base.HistogramExemplars != nil {
		t.Fatal("absent metadata became known empty")
	}
	for _, total := range []int{0, 1, 40} {
		q, row := histogramExemplarRow377(t, total)
		r := histogramExemplarResult377(t, q, row)
		report := r.HistogramExemplars
		if report == nil || report.Version != HistogramExemplarReportVersion || report.ValueEncoding != "source-float64" || report.OriginalValueType != "unknown" || len(report.Exports) != 1 {
			t.Fatal("report envelope", report)
		}
		export := report.Exports[0]
		if export.State != "available" || export.SourceRow != 0 || export.Total != strconv.Itoa(total) || len(export.Items) != min(total, 32) || export.Omitted != strconv.Itoa(max(0, total-32)) || report.Limited != (total > 32) {
			t.Fatal("sample retention/cap", export)
		}
		if total == 0 {
			if len(export.Items) != 0 || export.Diagnostics == nil {
				t.Fatal("empty samples unavailable")
			}
			continue
		}
		selected, item, e := SelectHistogramExemplar(r, export.Stream, export.EndNS, 0)
		if e != nil || item.TraceID != strings.Repeat("1", 32) || item.TimeNS != "100500000123" || item.StoredValue != "0.125" {
			t.Fatal("exact retained selection", selected, item, e)
		}
		selected.Stream.Resource["service.instance.id"] = "mutated"
		item.TraceID = "mutated"
		if report.Exports[0].Stream.Resource["service.instance.id"] == "mutated" || report.Exports[0].Items[0].TraceID == "mutated" {
			t.Fatal("selector aliases retained evidence")
		}
		r.HistogramExemplars.Exports[0].Items[0].StoredValue = "1"
		if _, _, e = SelectHistogramExemplar(r, export.Stream, export.EndNS, 0); e == nil {
			t.Fatal("forged projection bypassed result replay")
		}
	}
	q, row := histogramExemplarRow377(t, 40)
	r := histogramExemplarResult377(t, q, row)
	e := r.HistogramExemplars.Exports[0]
	if _, _, err := SelectHistogramExemplar(r, e.Stream, e.EndNS, 32); err == nil {
		t.Fatal("omitted reference selected")
	}
	if !strings.Contains(r.Plan.SQL, "arraySlice(`Exemplars.TimeUnix`,1,32)") || strings.Contains(r.Plan.SQL, " AS exemplar_filtered_attributes") || !strings.Contains(r.Plan.SQL, "length(`Exemplars.FilteredAttributes`)") {
		t.Fatal("source projection not bounded essentials")
	}
}
func TestHistogramExemplar377OptionalFailureIndependent(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"null-total", func(r map[string]any) { r["exemplar_total"] = nil }},
		{"partial", func(r map[string]any) { delete(r, "exemplar_span_ids") }},
		{"lengths", func(r map[string]any) { r["exemplar_span_total"] = "2" }},
		{"short-array", func(r map[string]any) { r["exemplar_times"] = []string{} }},
		{"noncanonical-time", func(r map[string]any) { r["exemplar_times"] = []string{"01"} }},
		{"overflow-time", func(r map[string]any) { r["exemplar_times"] = []string{"9223372036854775808"} }},
		{"invalid-value", func(r map[string]any) { r["exemplar_values"] = []string{"0x1p0"} }},
	} {
		t.Run(test.name, func(t *testing.T) {
			q, row := histogramExemplarRow377(t, 1)
			baseline := histogramExemplarResult377(t, q, row)
			test.mutate(row)
			r := histogramExemplarResult377(t, q, row)
			if r.Histograms.Basis != baseline.Histograms.Basis || r.HistogramExemplars.Exports[0].State != "unavailable" || !slices.Contains(r.HistogramExemplars.Exports[0].Diagnostics, "metadata-invalid") {
				t.Fatal("optional metadata poisoned numeric evaluation", r.HistogramExemplars)
			}
			if test.name == "null-total" && (r.HistogramExemplars.Exports[0].Total != "" || r.HistogramExemplars.Exports[0].Omitted != "" || r.HistogramExemplars.Limited) {
				t.Fatal("unknown metadata invented zero/limit")
			}
		})
	}
	for _, value := range []string{"nan", "inf", "-inf", "0"} {
		q, row := histogramExemplarRow377(t, 1)
		row["exemplar_values"] = []string{value}
		row["exemplar_times"] = []string{"0"}
		row["exemplar_span_ids"] = []string{"invalid-span"}
		r := histogramExemplarResult377(t, q, row)
		if r.HistogramExemplars.Exports[0].State != "available" || r.HistogramExemplars.Exports[0].Items[0].StoredValue != value || r.HistogramExemplars.Exports[0].Items[0].SpanID != "invalid-span" {
			t.Fatal("stored metadata gained original scalar/ID validity claims")
		}
	}
}
func TestHistogramExemplar377ReportBudgetAndAmbiguity(t *testing.T) {
	q, row := histogramExemplarRow377(t, 32)
	rows := make([]map[string]any, 150)
	for i := range rows {
		wire := raw(row)
		if json.Unmarshal(wire, &rows[i]) != nil {
			t.Fatal("fixture")
		}
		if i != 0 && i != len(rows)-1 {
			rows[i]["resource"].(map[string]any)["service.instance.id"] = fmt.Sprintf("filler-%d", i)
		}
	}
	r := histogramExemplarResult377(t, q, rows...)
	report := r.HistogramExemplars
	encoded, e := json.Marshal(report)
	if e != nil || len(encoded) > HistogramExemplarReportMaxBytes || !report.Limited || report.OmittedExports == 0 || len(report.Exports)+report.OmittedExports != len(rows) || report.Exports[0].SourceRow != 149 {
		t.Fatal("typed report budget/ordering", len(encoded), e)
	}
	if slices.ContainsFunc(report.Exports, func(e HistogramExemplarExport) bool { return e.SourceRow == 0 }) {
		t.Fatal("fixture did not omit the older duplicate")
	}
	export := report.Exports[0]
	if _, _, e = SelectHistogramExemplar(r, export.Stream, export.EndNS, 0); e == nil {
		t.Fatal("omitted duplicate made association falsely unique")
	}
}
