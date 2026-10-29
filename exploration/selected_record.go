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
	if q.EnvironmentSelector == nil {
		return nil
	}
	selection, err := q.EnvironmentSelection()
	if err != nil {
		return err
	}
	for _, wire := range rows {
		if err = validateSelectedRecord(selection, q.Signal, q.FromMS, q.ToMS, wire); err != nil {
			return err
		}
	}
	return nil
}
