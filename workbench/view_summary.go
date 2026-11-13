package workbench

import (
	"bytes"
	"encoding/json"
	"errors"
	"slices"
	"strings"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
)

// SummaryViewDeclaration chooses existing supplied population facts, not arithmetic
// or another source query. Field order is authored presentation.
type SummaryViewDeclaration struct {
	Version string               `json:"version"`
	Scope   ViewDeclarationScope `json:"scope"`
	Kind    SummaryViewKind      `json:"kind"`
	Display SummaryViewDisplay   `json:"display"`
}
type SummaryViewKind struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
}
type SummaryViewDisplay struct {
	Fields []string `json:"fields"`
	Title  string   `json:"title,omitempty"`
}

func summaryJSONFields(data []byte, allowed ...string) (map[string]json.RawMessage, error) {
	if !identity.WellFormedJSON(data) {
		return nil, errors.New("summary requires valid Unicode JSON")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	for key := range fields {
		if !slices.Contains(allowed, key) {
			return nil, errors.New("unsupported summary field: " + key)
		}
	}
	return fields, nil
}
func (d *SummaryViewDeclaration) UnmarshalJSON(data []byte) error {
	fields, err := summaryJSONFields(data, "version", "scope", "kind", "display")
	if err != nil {
		return err
	}
	for name, keys := range map[string][]string{"scope": {"kind", "signal"}, "kind": {"name", "version"}} {
		if raw, present := fields[name]; present {
			if _, err := summaryJSONFields(raw, keys...); err != nil {
				return err
			}
		}
	}
	type wire SummaryViewDeclaration
	var value wire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	*d = SummaryViewDeclaration(value)
	return nil
}
func (d *SummaryViewDisplay) UnmarshalJSON(data []byte) error {
	fields, err := summaryJSONFields(data, "fields", "title")
	if err != nil {
		return err
	}
	if raw, present := fields["title"]; present && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return errors.New("omit unused summary title rather than supplying null")
	}
	type wire SummaryViewDisplay
	var value wire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	*d = SummaryViewDisplay(value)
	return nil
}

// SummaryViewFields enumerates existing source-provided facts in the finite
// ordinary record profile. It makes no availability or completeness attestation.
func SummaryViewFields(q exploration.Query) []string {
	if _, err := exploration.Compile(q); err != nil {
		return nil
	}
	if q.NumericHistogram != nil || q.GroupTimeSeries || q.ReportedEdge != nil {
		return nil
	}
	for _, field := range q.GroupBy {
		if field.Kind == "column" && (field.Key == "body_pattern_v1" || field.Key == "duration_bucket") {
			return nil
		}
	}
	fields := []string{"count"}
	if exploration.RecordRateSummaryQuery(q) {
		fields = append(fields, exploration.RecordRateKey)
	}
	for _, field := range exploration.ViewMeasureFields(q) {
		fields = append(fields, field.Key)
	}
	return fields
}
func (d SummaryViewDeclaration) Validate(q exploration.Query) error {
	available := SummaryViewFields(q)
	if len(available) == 0 || d.Version != "goo.view/1" || d.Scope.Kind != "record-query" || d.Scope.Signal != q.Signal || d.Kind.Name != "measure-summary" || d.Kind.Version != 1 {
		return errors.New("require a pinned ordinary record population summary")
	}
	if len(d.Display.Fields) < 1 || len(d.Display.Fields) > 6 {
		return errors.New("choose one to six summary facts")
	}
	seen := map[string]bool{}
	for _, field := range d.Display.Fields {
		if seen[field] || !slices.Contains(available, field) {
			return errors.New("choose distinct supplied summary facts")
		}
		seen[field] = true
	}
	if !text(d.Display.Title, 120, d.Display.Title != "", false) {
		return errors.New("summary title must be bounded valid text")
	}
	return nil
}
func ReuseSummaryViewDeclaration(d SummaryViewDeclaration, source, destination exploration.Query) (SummaryViewDeclaration, error) {
	if err := d.Validate(source); err != nil {
		return SummaryViewDeclaration{}, err
	}
	if source.Signal != destination.Signal {
		return SummaryViewDeclaration{}, errors.New("summary requires the same record signal")
	}
	if slices.ContainsFunc(d.Display.Fields, func(field string) bool { return strings.HasPrefix(field, "numeric.") }) && (source.Numeric == nil || destination.Numeric == nil || *source.Numeric != *destination.Numeric) {
		return SummaryViewDeclaration{}, errors.New("numeric summary requires the same field, type, scale, unit and parse policy")
	}
	if err := d.Validate(destination); err != nil {
		return SummaryViewDeclaration{}, err
	}
	d.Display.Fields = slices.Clone(d.Display.Fields)
	return d, nil
}
