package workbench

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/casselc/goo/exploration"
)

// HistogramViewDeclaration selects a supplied numeric distribution. It carries
// neither observed bins nor support facts or selection authority.
type HistogramViewDeclaration struct {
	Version string               `json:"version"`
	Scope   ViewDeclarationScope `json:"scope"`
	Kind    HistogramViewKind    `json:"kind"`
	Display HistogramViewDisplay `json:"display"`
}
type HistogramViewKind struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
}
type HistogramViewDisplay struct {
	Category         string `json:"category"`
	Value            string `json:"value"`
	Title            string `json:"title,omitempty"`
	RequiresComplete bool   `json:"requiresComplete,omitempty"`
}

func (d *HistogramViewDeclaration) UnmarshalJSON(data []byte) error {
	fields, err := summaryJSONFields(data, "version", "scope", "kind", "display")
	if err != nil {
		return err
	}
	for name, keys := range map[string][]string{"scope": {"kind", "signal"}, "kind": {"name", "version"}} {
		if raw, ok := fields[name]; ok {
			if _, err := summaryJSONFields(raw, keys...); err != nil {
				return err
			}
		}
	}
	type wire HistogramViewDeclaration
	var value wire
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&value); err != nil {
		return err
	}
	*d = HistogramViewDeclaration(value)
	return nil
}
func (d *HistogramViewDisplay) UnmarshalJSON(data []byte) error {
	fields, err := summaryJSONFields(data, "category", "value", "title", "requiresComplete")
	if err != nil {
		return err
	}
	for key, raw := range fields {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return errors.New("histogram display " + key + " must not be null")
		}
	}
	type wire HistogramViewDisplay
	var value wire
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&value); err != nil {
		return err
	}
	*d = HistogramViewDisplay(value)
	return nil
}

// HistogramViewEligible admits a page-zero ordinary numeric histogram question.
func HistogramViewEligible(q exploration.Query) bool {
	if _, err := exploration.Compile(q); err != nil {
		return false
	}
	return (q.Signal == "logs" || q.Signal == "spans") && q.Offset == 0 && len(q.GroupBy) == 0 && !q.GroupTimeSeries && q.ReportedEdge == nil && q.Numeric != nil && q.NumericHistogram != nil && q.NumericHistogram.Binning == "powers-of-two"
}
func (d HistogramViewDeclaration) Validate(q exploration.Query) error {
	if !HistogramViewEligible(q) || d.Version != "goo.view/1" || d.Scope.Kind != "record-query" || d.Scope.Signal != q.Signal || d.Kind.Version != 1 || (d.Kind.Name != "histogram" && d.Kind.Name != "table") {
		return errors.New("require a pinned histogram or table for an ungrouped page-zero numeric histogram question")
	}
	if d.Display.Category != "from" || d.Display.Value != "count" || !text(d.Display.Title, 120, d.Display.Title != "", false) {
		return errors.New("histogram display requires bin start, captured count and bounded title")
	}
	return nil
}
func ReuseHistogramViewDeclaration(d HistogramViewDeclaration, source, destination exploration.Query) (HistogramViewDeclaration, error) {
	if err := d.Validate(source); err != nil {
		return HistogramViewDeclaration{}, err
	}
	if !HistogramViewEligible(destination) || source.Signal != destination.Signal || *source.Numeric != *destination.Numeric || *source.NumericHistogram != *destination.NumericHistogram {
		return HistogramViewDeclaration{}, errors.New("histogram reuse requires the same signal, qualified numeric field, type, scale, unit, parse policy and bin specification")
	}
	if err := d.Validate(destination); err != nil {
		return HistogramViewDeclaration{}, err
	}
	return d, nil
}
