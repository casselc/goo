package exploration

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/identity"
	"strconv"
)

// validateSelectedRecord retains presence, raw resource values and half-open time
// semantics at new-selector source/snapshot boundaries, including overflow rows.
func validateSelectedRecord(selection identity.EnvironmentSelection, signal string, from, to int64, wire json.RawMessage) error {
	var record struct {
		Resource  map[string]*string `json:"ResourceAttributes"`
		Body      *string            `json:"Body"`
		SpanName  *string            `json:"SpanName"`
		Timestamp string             `json:"timestamp_ns"`
	}
	if e := json.Unmarshal(wire, &record); e != nil || record.Resource == nil {
		return fmt.Errorf("selected record source returned a different environment scope")
	}
	resource := map[string]string{}
	for key, value := range record.Resource {
		if value == nil {
			return fmt.Errorf("selected record source has null resource attribute")
		}
		resource[key] = *value
	}
	if !selection.Matches(resource) || signal == "logs" && record.Body == nil || signal == "spans" && record.SpanName == nil {
		return fmt.Errorf("selected record source returned a different environment or signal")
	}
	ns, e := strconv.ParseInt(record.Timestamp, 10, 64)
	if e != nil || ns < 0 || strconv.FormatInt(ns, 10) != record.Timestamp || ns/1000000 < from || ns/1000000 >= to {
		return fmt.Errorf("selected record source returned a record outside its window")
	}
	return nil
}
func validateSelectedRows(q Query, rows []json.RawMessage) error {
	if q.EnvironmentSelector == nil && q.ReportedEdge == nil {
		return nil
	}
	selection, err := q.EnvironmentSelection()
	if err != nil {
		return err
	}
	for _, wire := range rows {
		if q.ReportedEdge != nil {
			if err = validateReportedEdgeRecord(q.ReportedEdge, wire); err != nil {
				return err
			}
		}
		if err = validateSelectedRecord(selection, q.Signal, q.FromMS, q.ToMS, wire); err != nil {
			return err
		}
	}
	return nil
}

// Parent ownership is certified by the checked source plan, not reconstructed
// from a bounded record page or refreshed while reopening captured evidence.
func validateReportedEdgeRecord(edge *ReportedEdge, wire json.RawMessage) error {
	if !identity.WellFormedJSON(wire) {
		return fmt.Errorf("reported-edge source has malformed record JSON")
	}
	var record struct {
		Service  *string            `json:"ServiceName"`
		Trace    string             `json:"TraceId"`
		Parent   string             `json:"ParentSpanId"`
		Resource map[string]*string `json:"ResourceAttributes"`
	}
	if err := json.Unmarshal(wire, &record); err != nil || record.Service == nil || *record.Service != edge.Child.Service || !exampleID(record.Trace, 32) || !exampleID(record.Parent, 16) {
		return fmt.Errorf("reported-edge source returned a different child identity or invalid relationship IDs")
	}
	resource := make(map[string]string, len(record.Resource))
	for key, value := range record.Resource {
		if value == nil {
			return fmt.Errorf("reported-edge source has null resource attribute")
		}
		resource[key] = *value
	}
	if !edge.Child.Namespace.Matches(resource) {
		return fmt.Errorf("reported-edge source returned a different child namespace")
	}
	return nil
}
