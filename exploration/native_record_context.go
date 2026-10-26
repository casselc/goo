package exploration

import (
	"fmt"
	"strconv"

	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
)

// NativeRecordContext projects only established shared resource identity. It is
// contextual coincidence, not metric-to-record causality or interval coverage.
type NativeRecordContext struct {
	Question Query  `json:"question"`
	FromNS   string `json:"fromNS"`
	ToNS     string `json:"toNS"`
	Grain    string `json:"grain"`
}

// PlanNativeRecordContext preserves exact half-open nanosecond membership inside
// an enclosing millisecond source window. Gauge points must be supplied as
// [atNS,atNS+1); no lookaround or current-time fallback is inferred.
func PlanNativeRecordContext(parent measurements.Question, stream metrics.Stream, signal, fromNS, toNS string) (*NativeRecordContext, error) {
	if parent.Definition.Source.Resource != nil {
		return nil, fmt.Errorf("service record context does not support resource-bound measurements")
	}
	if signal != "logs" && signal != "spans" {
		return nil, fmt.Errorf("record context requires logs or spans")
	}
	if err := parent.Validate(); err != nil {
		return nil, err
	}
	if parent.Definition.Kind != "gauge" && parent.Definition.Kind != "counter" && parent.Definition.Kind != "histogram" {
		return nil, fmt.Errorf("record context requires a native metric")
	}
	if old := parent.Definition.Source.ExactStream; old != nil {
		// Validation below cannot replace a previously pinned stream identity.
		if !nativeContextSameStream(*old, stream) {
			return nil, fmt.Errorf("record context stream differs from pinned native stream")
		}
	}
	copy := parent
	copy.Definition.Source.ExactStream = &stream
	if err := copy.Validate(); err != nil {
		return nil, err
	}
	parse := func(v string) (uint64, error) {
		n, e := strconv.ParseUint(v, 10, 64)
		if e != nil || strconv.FormatUint(n, 10) != v {
			return 0, fmt.Errorf("record context requires canonical nanosecond bounds")
		}
		return n, nil
	}
	from, e := parse(fromNS)
	if e != nil {
		return nil, e
	}
	to, e := parse(toNS)
	if e != nil {
		return nil, e
	}
	a, b := parent.BoundsNS()
	lo, _ := strconv.ParseUint(a, 10, 64)
	hi, _ := strconv.ParseUint(b, 10, 64)
	if to <= from || from < lo || to > hi {
		return nil, fmt.Errorf("record context must be a nonempty interval inside the native evaluation window")
	}
	grain := "interval-context"
	if parent.Definition.Kind == "gauge" {
		if to-from != 1 {
			return nil, fmt.Errorf("gauge record context requires a single nanosecond point")
		}
		grain = "point-context"
	}
	fromMS, toMS := from/1000000, (to+999999)/1000000
	if toMS-fromMS > 7200000 {
		return nil, fmt.Errorf("record context enclosing millisecond window exceeds 120 minutes; narrow the interval")
	}
	resource := stream.Resource
	namespace := identity.FromValue(nil)
	if v, ok := resource[identity.NamespaceKey]; ok {
		namespace = identity.FromValue(&v)
	}
	q := Query{Signal: signal, Environment: parent.Definition.Environment, FromMS: int64(fromMS), ToMS: int64(toMS), Limit: 100, Where: &Predicate{All: []Predicate{
		{Field: &Field{Kind: "column", Key: "service"}, Op: "eq", Value: resource["service.name"]},
		*NamespacePredicate(namespace),
		{Field: &Field{Kind: "resource", Key: "service.instance.id"}, Op: "eq", Value: resource["service.instance.id"]},
		{Field: &Field{Kind: "column", Key: "timestamp_ns"}, Op: "gte", Value: fromNS},
		{Field: &Field{Kind: "column", Key: "timestamp_ns"}, Op: "lte", Value: strconv.FormatUint(to-1, 10)},
	}}}
	if e := q.Validate(); e != nil {
		return nil, e
	}
	return &NativeRecordContext{Question: q, FromNS: fromNS, ToNS: toNS, Grain: grain}, nil
}
func nativeContextSameStream(a, b metrics.Stream) bool {
	if a.ResourceSchemaURL != b.ResourceSchemaURL || a.Scope != b.Scope || a.Version != b.Version || a.SchemaURL != b.SchemaURL || a.Name != b.Name || a.Unit != b.Unit || len(a.Resource) != len(b.Resource) || len(a.Attributes) != len(b.Attributes) {
		return false
	}
	for k, v := range a.Resource {
		w, ok := b.Resource[k]
		if !ok || w != v {
			return false
		}
	}
	for k, v := range a.Attributes {
		w, ok := b.Attributes[k]
		if !ok || w != v {
			return false
		}
	}
	return true
}
