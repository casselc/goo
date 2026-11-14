package workbench

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/casselc/goo/exploration"
)

// FacetViewDeclaration describes one categorical control over the applied root
// scope. Values, counts, draft selections and source requests are not persisted.
type FacetViewDeclaration struct {
	Version string               `json:"version"`
	Scope   ViewDeclarationScope `json:"scope"`
	Kind    FacetViewKind        `json:"kind"`
	Field   exploration.Field    `json:"field"`
	Display FacetViewDisplay     `json:"display"`
}
type FacetViewKind struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
}
type FacetViewDisplay struct {
	Basis string `json:"basis"`
	Title string `json:"title,omitempty"`
}

func (d *FacetViewDisplay) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	if raw, ok := fields["title"]; ok && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return errors.New("omit an unused facet title rather than supplying null")
	}
	type wire FacetViewDisplay
	var value wire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	*d = FacetViewDisplay(value)
	return nil
}

func (d FacetViewDeclaration) Validate(q exploration.Query) error {
	if d.Version != "goo.view/1" || d.Scope.Kind != "record-query" || d.Scope.Signal != q.Signal || d.Kind.Name != "categorical-facet" || d.Kind.Version != 1 || d.Display.Basis != "applied" {
		return errors.New("require a pinned categorical facet over the applied record scope")
	}
	if !text(d.Display.Title, 120, d.Display.Title != "", false) {
		return errors.New("facet title must be bounded valid text")
	}
	if _, err := exploration.Compile(q); err != nil {
		return err
	}
	if q.Numeric != nil || q.NumericHistogram != nil || q.ReportedEdge != nil || q.GroupTimeSeries {
		return errors.New("facets require ordinary record queries without numeric, histogram, reported-edge or grouped-time profiles")
	}
	// Match the existing counted-values adapter's count projection. Scope,
	// predicates and environment remain intact; display measures are not a facet.
	projected := q
	projected.GroupBy = []exploration.Field{d.Field}
	projected.Summary = ""
	projected.BucketMS = 0
	projected.SpanStatus = false
	projected.LogSeverity = false
	projected.Limit = 1
	projected.Offset = 0
	return exploration.ValidateCountedValuesQuery(projected)
}
