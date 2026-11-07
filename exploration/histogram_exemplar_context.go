package exploration

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
)

// HistogramExemplarContext follows a reported trace reference. Export identity,
// sample time and the original question window remain separate facts; this does
// not claim that a span contributed to a bucket or that a trace is complete.
type HistogramExemplarContext struct {
	Relation      string `json:"relation"`
	Question      Query  `json:"question"`
	FromNS        string `json:"fromNS"`
	ToNS          string `json:"toNS"`
	SourceRow     int    `json:"sourceRow"`
	ExportStartNS string `json:"exportStartNS"`
	ExportEndNS   string `json:"exportEndNS"`
	ExemplarIndex int    `json:"exemplarIndex"`
	TimeNS        string `json:"timeNS"`
	TraceID       string `json:"traceId"`
	SpanID        string `json:"spanId,omitempty"`
}

// PlanHistogramExemplarContext checks retained export membership without source
// I/O. Selection verifies the entire result and all retained raw rows, including
// rows omitted from the bounded metadata report, before authorizing a reference.
func PlanHistogramExemplarContext(result *measurements.Result, stream metrics.Stream, endNS string, ordinal int) (*HistogramExemplarContext, error) {
	if result == nil {
		return nil, fmt.Errorf("reported exemplar requires a retained measurement result")
	}
	parent := result.Question
	d := parent.Definition
	if d.Kind != "histogram" || d.Grain != "per-stream" || d.Source.Resource != nil || d.HistogramBucketMS != nil || (d.Source.Unit != "s" && d.Source.Unit != "ms") {
		return nil, fmt.Errorf("reported exemplar trace context requires an actual service histogram export")
	}
	if old := d.Source.ExactStream; old != nil && !nativeContextSameStream(*old, stream) {
		return nil, fmt.Errorf("reported exemplar stream differs from pinned native stream")
	}
	checked := parent
	checked.Definition.Source.ExactStream = &stream
	if err := checked.Validate(); err != nil {
		return nil, err
	}
	export, reference, err := measurements.SelectHistogramExemplar(result, stream, endNS, ordinal)
	if err != nil {
		return nil, err
	}
	if !hexTrace.MatchString(reference.TraceID) || strings.Trim(reference.TraceID, "0") == "" {
		return nil, fmt.Errorf("reported exemplar has no valid nonzero trace ID")
	}
	fromNS, toNS := parent.BoundsNS()
	from, fromErr := strconv.ParseUint(fromNS, 10, 64)
	to, err := strconv.ParseUint(toNS, 10, 64)
	end, endErr := strconv.ParseUint(endNS, 10, 64)
	if err != nil || fromErr != nil || endErr != nil || to == 0 || end < from || end > to {
		return nil, fmt.Errorf("reported exemplar requires exact original native bounds")
	}
	namespace := d.Source.Namespace.Clone()
	if namespace == nil {
		namespace = identity.FromValue(nil)
		if value, present := stream.Resource[identity.NamespaceKey]; present {
			namespace = identity.FromValue(&value)
		}
	}
	q := Query{Signal: "spans", Environment: d.Environment, EnvironmentSelector: copyEnvironmentSelector(d.EnvironmentSelector), FromMS: parent.FromMS, ToMS: parent.ToMS, Limit: 100, Where: &Predicate{All: []Predicate{
		{Field: &Field{Kind: "column", Key: "trace_id"}, Op: "eq", Value: reference.TraceID},
		*NamespacePredicate(namespace),
		{Field: &Field{Kind: "column", Key: "timestamp_ns"}, Op: "gte", Value: fromNS},
		{Field: &Field{Kind: "column", Key: "timestamp_ns"}, Op: "lte", Value: strconv.FormatUint(to-1, 10)},
	}}}
	if err := q.Validate(); err != nil {
		return nil, err
	}
	context := &HistogramExemplarContext{Relation: "reported-exemplar", Question: q, FromNS: fromNS, ToNS: toNS, SourceRow: export.SourceRow, ExportStartNS: export.StartNS, ExportEndNS: export.EndNS, ExemplarIndex: reference.Index, TimeNS: reference.TimeNS, TraceID: reference.TraceID}
	if hexSpan.MatchString(reference.SpanID) && strings.Trim(reference.SpanID, "0") != "" {
		context.SpanID = reference.SpanID
	}
	return context, nil
}
