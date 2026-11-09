package exploration

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"time"

	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
)

// Capture has six finite families. The envelope selects exactly one; payload
// validators retain responsibility for each family's evidence and clocks.
var captureArmKeys = [...]string{"traceSearch", "explanation", "measurement", "comparison", "explore", "inventory"}

func captureArmIndex(kind string) int {
	switch kind {
	case "trace-search":
		return 0
	case "explanation":
		return 1
	case "measurement":
		return 2
	case "comparison":
		return 3
	case "explore":
		return 4
	case "inventory":
		return 5
	default:
		return -1
	}
}
func validateCaptureArms(kind string, present [6]bool) error {
	selected := captureArmIndex(kind)
	if selected < 0 {
		return errors.New("capture kind must select a supported family")
	}
	for i, exists := range present {
		if exists != (i == selected) {
			return errors.New("capture requires exactly its selected family arm")
		}
	}
	return nil
}
func (q CaptureQuestion) validateShape() error {
	return validateCaptureArms(q.Kind, [6]bool{q.TraceSearch != nil, q.Explanation != nil, q.Measurement != nil, q.Comparison != nil, q.Explore != nil, q.Inventory != nil})
}
func (c SavedCapture) validateShape() error {
	return validateCaptureArms(c.Question.Kind, [6]bool{c.TraceSearch != nil, c.Explanation != nil, c.Measurement != nil, c.Comparison != nil, c.Explore != nil, c.Inventory != nil})
}
func (p CapturePage) validateShape() error {
	return validateCaptureArms(p.Receipt.Question.Kind, [6]bool{p.TraceSearch != nil, p.Explanation != nil, p.Measurement != nil, p.Comparison != nil, p.Explore != nil, p.Inventory != nil})
}
func validateCaptureReceipt(r CaptureReceipt, requireID bool) error {
	if r.Version != 1 || !validCaptureTitle(r.Title) || (requireID && !regexpRevision.MatchString(r.ID)) || (r.ID != "" && !regexpRevision.MatchString(r.ID)) || (r.Parent != "" && !regexpRevision.MatchString(r.Parent)) {
		return errors.New("invalid capture receipt")
	}
	if _, err := time.Parse(time.RFC3339Nano, r.CapturedAt); err != nil {
		return errors.New("invalid capture time")
	}
	if err := r.Question.Validate(); err != nil {
		return err
	}
	if !reflect.DeepEqual(r.Question, r.Question.Normalized()) {
		return errors.New("capture receipt question must be normalized")
	}
	return nil
}

// Canonicalize the finite envelope before decoding: later duplicate arm
// objects replace earlier objects, without rewriting nested family payloads.
func captureWire(data []byte, allowed []string) (map[string]json.RawMessage, []byte, error) {
	if !identity.WellFormedJSON(data) {
		return nil, nil, errors.New("capture contains malformed JSON or Unicode")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, nil, err
	}
	if fields == nil {
		return nil, nil, errors.New("capture requires an object")
	}
	canonical, err := json.Marshal(fields)
	if err != nil {
		return nil, nil, err
	}
	for key, raw := range fields {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return nil, nil, fmt.Errorf("capture field %q must be non-null", key)
		}
		if !slices.Contains(allowed, key) {
			return nil, nil, fmt.Errorf("unknown capture field %q", key)
		}
	}
	return fields, canonical, nil
}
func captureWireArm(fields map[string]json.RawMessage, kind string) error {
	var present [6]bool
	for i, key := range captureArmKeys {
		raw, exists := fields[key]
		present[i] = exists
		if exists {
			var object map[string]json.RawMessage
			if json.Unmarshal(raw, &object) != nil || object == nil {
				return errors.New("capture family arm must be a non-null object")
			}
		}
	}
	return validateCaptureArms(kind, present)
}
func decodeCaptureWire(raw []byte, out any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	return decoder.Decode(out)
}
func (q *CaptureQuestion) UnmarshalJSON(data []byte) error {
	*q = CaptureQuestion{}
	fields, canonical, err := captureWire(data, append([]string{"kind"}, captureArmKeys[:]...))
	if err != nil {
		return err
	}
	var kind string
	if err = json.Unmarshal(fields["kind"], &kind); err != nil {
		return err
	}
	if err = captureWireArm(fields, kind); err != nil {
		return err
	}
	type wire CaptureQuestion
	var value wire
	if err = decodeCaptureWire(canonical, &value); err != nil {
		return err
	}
	*q = CaptureQuestion(value)
	return nil
}
func (c *SavedCapture) UnmarshalJSON(data []byte) error {
	*c = SavedCapture{}
	fields, canonical, err := captureWire(data, append([]string{"version", "id", "parent", "title", "capturedAt", "question"}, captureArmKeys[:]...))
	if err != nil {
		return err
	}
	var question CaptureQuestion
	if err = json.Unmarshal(fields["question"], &question); err != nil {
		return err
	}
	if err = captureWireArm(fields, question.Kind); err != nil {
		return err
	}
	type wire SavedCapture
	var value wire
	if err = decodeCaptureWire(canonical, &value); err != nil {
		return err
	}
	candidate := SavedCapture(value)
	if err = validateCaptureReceipt(candidate.CaptureReceipt, true); err != nil {
		return err
	}
	*c = candidate
	return nil
}
func (p *CapturePage) UnmarshalJSON(data []byte) error {
	*p = CapturePage{}
	fields, canonical, err := captureWire(data, append([]string{"receipt", "retainedRecords", "recordsLimited", "nextCursor", "previousCursor"}, captureArmKeys[:]...))
	if err != nil {
		return err
	}
	_, receiptWire, err := captureWire(fields["receipt"], []string{"version", "id", "parent", "title", "capturedAt", "question"})
	if err != nil {
		return err
	}
	var receipt CaptureReceipt
	if err = json.Unmarshal(receiptWire, &receipt); err != nil {
		return err
	}
	for _, key := range []string{"nextCursor", "previousCursor"} {
		if raw, present := fields[key]; present {
			var cursor string
			if receipt.Question.Kind != "explore" || json.Unmarshal(raw, &cursor) != nil || cursor == "" {
				return errors.New("capture cursor fields require an explored record page and a nonempty cursor")
			}
		}
	}
	if err = captureWireArm(fields, receipt.Question.Kind); err != nil {
		return err
	}
	for _, key := range []string{"retainedRecords", "recordsLimited"} {
		raw, ok := fields[key]
		if !ok || bytes.Equal(raw, []byte("null")) {
			return errors.New("capture page requires explicit retained records and limiting policy")
		}
	}
	fields["receipt"] = receiptWire
	canonical, err = json.Marshal(fields)
	if err != nil {
		return err
	}
	type wire CapturePage
	var value wire
	if err = decodeCaptureWire(canonical, &value); err != nil {
		return err
	}
	candidate := CapturePage(value)
	if err = validateCapturePage(candidate); err != nil {
		return err
	}
	*p = candidate
	return nil
}
func validateCapturePage(p CapturePage) error {
	if err := validateCaptureReceipt(p.Receipt, true); err != nil {
		return err
	}
	if err := p.validateShape(); err != nil {
		return err
	}
	limit := CaptureRowLimit
	if p.Receipt.Question.Kind == "measurement" {
		limit = measurements.RowLimit
	}
	if p.RetainedRecords < 0 || p.RetainedRecords > limit {
		return errors.New("invalid retained capture record count")
	}
	noCursors := p.NextCursor == "" && p.PreviousCursor == ""
	switch p.Receipt.Question.Kind {
	case "explore":
		q := p.Explore.Question
		offset := q.Offset
		if q.Limit != CapturePageSize || offset < 0 || offset >= CaptureRowLimit || offset%CapturePageSize != 0 || (offset >= p.RetainedRecords && offset != 0) || len(p.Explore.Rows) != min(CapturePageSize, p.RetainedRecords-offset) {
			return errors.New("capture exploration page does not match retained records")
		}
		normalized := (CaptureQuestion{Kind: "explore", Explore: &q}).Normalized()
		if !reflect.DeepEqual(normalized.Explore, p.Receipt.Question.Explore) {
			return errors.New("capture page question differs from receipt")
		}
		end := offset + len(p.Explore.Rows)
		hasMore := end < p.RetainedRecords
		if p.Explore.HasMore != hasMore || (p.NextCursor != "") != hasMore || (p.PreviousCursor != "") != (offset > 0) {
			return errors.New("capture page navigation disagrees with retained records")
		}
		if p.NextCursor != "" {
			if n, err := cursorOffset(p.Receipt.ID, p.NextCursor); err != nil || n != end {
				return errors.New("invalid next capture cursor")
			}
		}
		if p.PreviousCursor != "" {
			if n, err := cursorOffset(p.Receipt.ID, p.PreviousCursor); err != nil || n != offset-CapturePageSize {
				return errors.New("invalid previous capture cursor")
			}
		}
	case "measurement":
		if !reflect.DeepEqual(p.Measurement.Question, *p.Receipt.Question.Measurement) {
			return errors.New("capture page measurement differs from receipt")
		}
		if !noCursors || p.RecordsLimited || p.RetainedRecords != len(p.Measurement.Rows) {
			return errors.New("invalid measurement capture page")
		}
	case "trace-search":
		if !reflect.DeepEqual(p.TraceSearch.Request, *p.Receipt.Question.TraceSearch) {
			return errors.New("capture page trace search differs from receipt")
		}
		if !noCursors || p.RetainedRecords > p.Receipt.Question.TraceSearch.Limit || p.RetainedRecords != len(p.TraceSearch.Items) || p.RecordsLimited != p.TraceSearch.HasMore {
			return errors.New("invalid trace-search capture page")
		}
	case "explanation":
		if !reflect.DeepEqual(p.Explanation.Request, *p.Receipt.Question.Explanation) {
			return errors.New("capture page explanation differs from receipt")
		}
		if !noCursors || p.RecordsLimited || p.RetainedRecords != 0 {
			return errors.New("capture family has no record pages")
		}
	case "inventory":
		if !reflect.DeepEqual(p.Inventory.Question, *p.Receipt.Question.Inventory) {
			return errors.New("capture page inventory differs from receipt")
		}
		if !noCursors || p.RecordsLimited || p.RetainedRecords != 0 {
			return errors.New("capture family has no record pages")
		}
	case "comparison":
		q := p.Receipt.Question.Comparison
		if p.Comparison.Request.Distribution != q.Distribution || !reflect.DeepEqual(p.Comparison.Request.Baseline.Question, q.Baseline) || !reflect.DeepEqual(p.Comparison.Request.Selection.Question, q.Selection) {
			return errors.New("capture page comparison differs from receipt")
		}
		if !noCursors || p.RecordsLimited || p.RetainedRecords != 0 {
			return errors.New("capture family has no record pages")
		}
	}
	return nil
}
