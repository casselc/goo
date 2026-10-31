package workbench

import (
	"errors"

	"github.com/casselc/goo/exploration"
)

// ViewDeclaration stores reusable display intent, never captured values, runtime
// row IDs or completeness witnesses. Scope refers to a checked query in its parent.
type ViewDeclaration struct {
	Version string               `json:"version"`
	Scope   ViewDeclarationScope `json:"scope"`
	Kind    ViewKind             `json:"kind"`
	Display ViewDisplay          `json:"display"`
}

type ViewDeclarationScope struct {
	Kind   string `json:"kind"`
	Signal string `json:"signal"`
}

type ViewKind struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
}

// ViewDisplay is the bounded ordinary record-measure profile of the public display
// contract. Empty optional strings mean the renderer's defaults, not new queries.
type ViewDisplay struct {
	Category         string `json:"category"`
	Value            string `json:"value"`
	Title            string `json:"title,omitempty"`
	Scale            string `json:"scale,omitempty"`
	Order            string `json:"order,omitempty"`
	Query            string `json:"query,omitempty"`
	RequiresComplete bool   `json:"requiresComplete,omitempty"`
}

func (d ViewDeclaration) Validate(q exploration.Query) error {
	if d.Version != "goo.view/1" || d.Scope.Kind != "record-query" || d.Scope.Signal != q.Signal || d.Kind.Version != 1 || (d.Kind.Name != "bar" && d.Kind.Name != "table") {
		return errors.New("require a versioned record-query display and a pinned bar or table kind")
	}
	if len(q.GroupBy) != 1 || q.NumericHistogram != nil || q.GroupTimeSeries || q.GroupBy[0].Kind == "column" && (q.GroupBy[0].Key == "body_pattern_v1" || q.GroupBy[0].Key == "duration_bucket") {
		return errors.New("authored displays require an ordinary single-field record breakdown")
	}
	available := d.Display.Value == "count"
	for _, field := range exploration.ViewMeasureFields(q) {
		if field.Key == d.Display.Value && field.Aggregation != nil && (field.Aggregation.Family == "numeric" || field.Aggregation.Family == "duration") {
			available = true
			break
		}
	}
	if d.Display.Category != "category" || !available {
		return errors.New("choose a count, numeric or duration measure supplied by the checked query")
	}
	if d.Display.Scale != "" && d.Display.Scale != "max" && d.Display.Scale != "population" {
		return errors.New("choose max or population display scale")
	}
	if d.Display.Value != "count" && d.Display.Scale == "population" {
		return errors.New("population share is available only for captured record counts")
	}
	if d.Display.Order != "" && d.Display.Order != "source" && d.Display.Order != "label" && d.Display.Order != "value-desc" && d.Display.Order != "value-asc" {
		return errors.New("choose a supported local display order")
	}
	if !text(d.Display.Query, 256, false, false) || !text(d.Display.Title, 120, d.Display.Title != "", false) {
		return errors.New("display Find text and title must be bounded valid text")
	}
	return nil
}
