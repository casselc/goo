package exploration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/internal/sourcecap"
	"github.com/casselc/goo/internal/sourcefailure"
)

// SpanLinkTarget is an exact reported OTel link target, not a causal edge or an
// invocation identity. LinksTo selects span records, never individual link entries.
type SpanLinkTarget struct {
	TraceID string `json:"traceId" yaml:"traceId"`
	SpanID  string `json:"spanId" yaml:"spanId"`
}

func (t *SpanLinkTarget) Validate() error {
	if t == nil || !exampleID(t.TraceID, 32) || !exampleID(t.SpanID, 16) {
		return fmt.Errorf("span link target requires canonical nonzero lowercase trace and span IDs")
	}
	return nil
}

func (t *SpanLinkTarget) UnmarshalJSON(data []byte) error {
	if !identity.WellFormedJSON(data) {
		return fmt.Errorf("span link target has malformed Unicode")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	if decoder.Decode(new(any)) != io.EOF {
		return fmt.Errorf("expected one span link target")
	}
	type wire SpanLinkTarget
	if !strictTraceSearchValue(value, reflect.TypeFor[wire](), 0) {
		return fmt.Errorf("span link target requires known non-null fields")
	}
	object, ok := value.(map[string]any)
	if !ok || len(object) != 2 {
		return fmt.Errorf("span link target requires both IDs")
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return err
	}
	var out wire
	if err = json.Unmarshal(canonical, &out); err != nil {
		return err
	}
	*t = SpanLinkTarget(out)
	return t.Validate()
}

// MatchSpanLinkTarget independently interprets tuple membership. All three arrays
// must align. Only a matching entry's attributes must be a supplied string map;
// unrelated malformed IDs/maps do not hide a different valid matching entry.
// Duplicate links return one boolean; duplicate records remain distinct records.
func MatchSpanLinkTarget(target *SpanLinkTarget, traces, spans []string, attributes []map[string]string) bool {
	if target.Validate() != nil || len(traces) != len(spans) || len(traces) != len(attributes) {
		return false
	}
	for i, trace := range traces {
		if trace == target.TraceID && spans[i] == target.SpanID && attributes[i] != nil {
			return true
		}
	}
	return false
}

func (c *compiler) spanLinksTo(target *SpanLinkTarget) string {
	trace, span := c.param(target.TraceID, "String"), c.param(target.SpanID, "String")
	// Enumerating one array and indexing the others is safe even when exporter
	// arrays are malformed. Multi-array lambda functions can throw before AND's
	// equal-length guard is evaluated when ClickHouse short circuiting is disabled.
	return "(length(`Links.TraceId`)=length(`Links.SpanId`) AND length(`Links.TraceId`)=length(`Links.Attributes`) AND arrayExists(i -> (`Links.TraceId`[i]=" + trace + " AND `Links.SpanId`[i]=" + span + "), arrayEnumerate(`Links.TraceId`)))"
}

func validateSpanLinkRecord(target *SpanLinkTarget, wire json.RawMessage) error {
	if !identity.WellFormedJSON(wire) {
		return fmt.Errorf("reported-link source has malformed record JSON")
	}
	var row struct {
		Traces     []json.RawMessage `json:"linked_traces"`
		Spans      []json.RawMessage `json:"linked_spans"`
		Attributes []json.RawMessage `json:"linked_attributes"`
	}
	if err := json.Unmarshal(wire, &row); err != nil || len(row.Traces) != len(row.Spans) || len(row.Traces) != len(row.Attributes) {
		return fmt.Errorf("reported-link source has malformed parallel link arrays")
	}
	for i := range row.Traces {
		var trace, span string
		if json.Unmarshal(row.Traces[i], &trace) != nil || json.Unmarshal(row.Spans[i], &span) != nil || trace != target.TraceID || span != target.SpanID {
			continue
		}
		var attrs map[string]*string
		if json.Unmarshal(row.Attributes[i], &attrs) != nil || attrs == nil {
			continue
		}
		valid := true
		for _, v := range attrs {
			if v == nil {
				valid = false
				break
			}
		}
		if valid {
			return nil
		}
	}
	return fmt.Errorf("source span does not report the selected link tuple")
}

// The standard adapter's non-null map/string array types are a prerequisite for
// full-population counts. A page check cannot certify omitted matching records.
func (s Source) requireSpanLinkSource(ctx context.Context, q Query) error {
	if q.LinksTo == nil {
		return nil
	}
	families, err := sourcecap.Inspect(ctx, s.read, sourcecap.Spans)
	if err != nil {
		return err
	}
	if len(families) != 1 || !families[0].Compatible() {
		kind := sourcefailure.Unavailable
		if len(families) == 1 && families[0].State == "incompatible" {
			kind = sourcefailure.Malformed
		}
		return sourcefailure.Wrap(kind, fmt.Errorf("reported-link queries require compatible standard OTel span columns"))
	}
	return nil
}
