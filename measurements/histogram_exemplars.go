package measurements

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/casselc/goo/metrics"
)

const HistogramExemplarReportVersion = "goo/histogram-exemplar-references/1"
const HistogramExemplarLimit = 32
const HistogramExemplarReportMaxBytes = 64 << 10

type HistogramExemplarReference struct {
	Index       int    `json:"index"`
	TimeNS      string `json:"timeNS"`
	TraceID     string `json:"traceId"`
	SpanID      string `json:"spanId"`
	StoredValue string `json:"storedValue"`
}
type HistogramExemplarExport struct {
	SourceRow   int                          `json:"sourceRow"`
	Stream      metrics.Stream               `json:"stream"`
	StartNS     string                       `json:"startNS"`
	EndNS       string                       `json:"endNS"`
	Temporality string                       `json:"temporality"`
	State       string                       `json:"state"`
	Items       []HistogramExemplarReference `json:"items"`
	Total       string                       `json:"total"`
	Omitted     string                       `json:"omitted"`
	Diagnostics []string                     `json:"diagnostics"`
}
type HistogramExemplarReport struct {
	Version           string                    `json:"version"`
	ValueEncoding     string                    `json:"valueEncoding"`
	OriginalValueType string                    `json:"originalValueType"`
	Exports           []HistogramExemplarExport `json:"exports"`
	Limited           bool                      `json:"limited"`
	OmittedExports    int                       `json:"omittedExports"`
}

var histogramExemplarFields = []string{"exemplar_total", "exemplar_span_total", "exemplar_time_total", "exemplar_value_total", "exemplar_attribute_total", "exemplar_times", "exemplar_values", "exemplar_trace_ids", "exemplar_span_ids"}

func hasHistogramExemplars(fields map[string]json.RawMessage) bool {
	for _, key := range histogramExemplarFields {
		if _, ok := fields[key]; ok {
			return true
		}
	}
	return false
}
func exemplarCount(raw json.RawMessage) (uint64, bool) {
	var value string
	if json.Unmarshal(raw, &value) != nil || value == "" {
		return 0, false
	}
	n, e := strconv.ParseUint(value, 10, 64)
	return n, e == nil && strconv.FormatUint(n, 10) == value
}
func exemplarStrings(raw json.RawMessage, n int) ([]string, bool) {
	var values []json.RawMessage
	if json.Unmarshal(raw, &values) != nil || values == nil || len(values) != n {
		return nil, false
	}
	out := make([]string, n)
	for i, v := range values {
		var s *string
		if json.Unmarshal(v, &s) != nil || s == nil || !text(*s, 128) {
			return nil, false
		}
		out[i] = *s
	}
	return out, true
}
func exemplarTime(value string) bool {
	n, e := strconv.ParseInt(value, 10, 64)
	return e == nil && n >= 0 && strconv.FormatInt(n, 10) == value
}
func exemplarValue(value string) bool {
	if value == "" || strings.TrimSpace(value) != value || strings.ContainsAny(value, "xXpP") {
		return false
	}
	switch strings.ToLower(value) {
	case "nan", "inf", "+inf", "-inf", "infinity", "+infinity", "-infinity":
		return true
	}
	_, e := strconv.ParseFloat(value, 64)
	return e == nil
}
func histogramSourceStream(row metricRow) metrics.Stream {
	return metrics.Stream{ResourceSchemaURL: row.ResourceSchemaURL, Resource: row.Resource, Scope: row.Scope, Version: row.Version, SchemaURL: row.SchemaURL, Name: row.Name, Unit: row.Unit, Attributes: row.Attributes}
}
func histogramExemplarExport(raw json.RawMessage, index int) HistogramExemplarExport {
	var row metricRow
	_ = json.Unmarshal(raw, &row)
	out := HistogramExemplarExport{SourceRow: index, Stream: histogramSourceStream(row), StartNS: row.StartNS, EndNS: row.EndNS, Temporality: map[int]string{1: "delta", 2: "cumulative"}[row.Temporality], State: "unavailable", Items: []HistogramExemplarReference{}, Total: "", Omitted: "", Diagnostics: []string{}}
	var fields map[string]json.RawMessage
	_ = json.Unmarshal(raw, &fields)
	refuse := func(reason string) HistogramExemplarExport { out.Diagnostics = []string{reason}; return out }
	if !hasHistogramExemplars(fields) {
		return refuse("metadata-not-retained")
	}
	total, ok := exemplarCount(fields["exemplar_total"])
	if !ok {
		return refuse("metadata-invalid")
	}
	out.Total = strconv.FormatUint(total, 10)
	out.Omitted = out.Total
	for _, key := range histogramExemplarFields[1:5] {
		value, ok := exemplarCount(fields[key])
		if !ok || value != total {
			return refuse("metadata-invalid")
		}
	}
	n := int(min(total, uint64(HistogramExemplarLimit)))
	times, tok := exemplarStrings(fields["exemplar_times"], n)
	values, vok := exemplarStrings(fields["exemplar_values"], n)
	traces, rok := exemplarStrings(fields["exemplar_trace_ids"], n)
	spans, sok := exemplarStrings(fields["exemplar_span_ids"], n)
	if !tok || !vok || !rok || !sok {
		return refuse("metadata-invalid")
	}
	items := make([]HistogramExemplarReference, n)
	for i := range n {
		if !exemplarTime(times[i]) || !exemplarValue(values[i]) {
			return refuse("metadata-invalid")
		}
		items[i] = HistogramExemplarReference{Index: i, TimeNS: times[i], StoredValue: values[i], TraceID: traces[i], SpanID: spans[i]}
	}
	if row.Flags&1 != 0 {
		return refuse("no-recorded-value")
	}
	out.State = "available"
	out.Items = items
	out.Omitted = strconv.FormatUint(total-uint64(n), 10)
	if total > uint64(n) {
		out.Diagnostics = append(out.Diagnostics, "sample-limit")
	}
	return out
}

// Optional reference metadata never changes numeric histogram admission. The
// source-row index remains stable even when the bounded report omits an export.
func projectHistogramExemplars(rows []json.RawMessage) *HistogramExemplarReport {
	present := false
	for _, raw := range rows {
		var fields map[string]json.RawMessage
		_ = json.Unmarshal(raw, &fields)
		if hasHistogramExemplars(fields) {
			present = true
			break
		}
	}
	if !present {
		return nil
	}
	report := &HistogramExemplarReport{Version: HistogramExemplarReportVersion, ValueEncoding: "source-float64", OriginalValueType: "unknown", Exports: []HistogramExemplarExport{}}
	// Reserve the largest envelope counters before retaining entries, avoiding
	// repeated whole-report encodes and keeping the final representation <=64KiB.
	reserved := *report
	reserved.OmittedExports = RowLimit
	envelope, _ := json.Marshal(reserved)
	used := len(envelope)
	for i := len(rows) - 1; i >= 0; i-- {
		export := histogramExemplarExport(rows[i], i)
		encoded, err := json.Marshal(export)
		cost := len(encoded)
		if len(report.Exports) > 0 {
			cost++
		}
		if err != nil || used+cost > HistogramExemplarReportMaxBytes {
			report.OmittedExports++
			report.Limited = true
			continue
		}
		report.Exports = append(report.Exports, export)
		used += cost
		if export.Omitted != "" && export.Omitted != "0" {
			report.Limited = true
		}
	}
	return report
}

// SelectHistogramExemplar replays the supplied result, checks association against
// ALL raw exports (including report-budget omissions), and returns detached data.
// A reported trace ID's validity and navigation scope are the context builder's
// separate responsibility; this selector does not fetch or invent a reference.
func SelectHistogramExemplar(r *Result, stream metrics.Stream, endNS string, ordinal int) (*HistogramExemplarExport, *HistogramExemplarReference, error) {
	refuse := func(reason string) (*HistogramExemplarExport, *HistogramExemplarReference, error) {
		return nil, nil, fmt.Errorf("histogram exemplar: %s", reason)
	}
	if r == nil || r.Question.Definition.Kind != "histogram" || !exemplarTime(endNS) || ordinal < 0 || ordinal >= HistogramExemplarLimit {
		return refuse("require a retained histogram export and bounded ordinal")
	}
	if err := ValidateResult(r); err != nil {
		return refuse("result replay refused")
	}
	sourceRow := -1
	matches := 0
	var selected metricRow
	for i, raw := range r.Rows {
		var row metricRow
		if json.Unmarshal(raw, &row) != nil {
			return refuse("invalid source row")
		}
		if row.EndNS == endNS && sameExactStream(histogramSourceStream(row), stream) {
			sourceRow = i
			matches++
			selected = row
		}
	}
	if matches != 1 {
		return refuse("export association is missing or ambiguous")
	}
	if r.HistogramExemplars == nil {
		return refuse("reference metadata was not retained")
	}
	for _, export := range r.HistogramExemplars.Exports {
		if export.SourceRow != sourceRow {
			continue
		}
		if export.State != "available" || !sameExactStream(export.Stream, stream) || export.StartNS != selected.StartNS || export.EndNS != selected.EndNS || export.Temporality != map[int]string{1: "delta", 2: "cumulative"}[selected.Temporality] || ordinal >= len(export.Items) || export.Items[ordinal].Index != ordinal {
			return refuse("reference is unavailable or omitted")
		}
		export.Stream.Resource = maps.Clone(export.Stream.Resource)
		export.Stream.Attributes = maps.Clone(export.Stream.Attributes)
		export.Items = slices.Clone(export.Items)
		export.Diagnostics = slices.Clone(export.Diagnostics)
		return &export, &export.Items[ordinal], nil
	}
	return refuse("export omitted by metadata budget")
}
