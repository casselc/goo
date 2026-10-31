package workbench

import (
	"bytes"
	"encoding/json"
	"errors"
	"slices"

	"github.com/casselc/goo/exploration"
)

// UnmarshalJSON retains the composition's single presentation owner even when a
// wire supplies a forbidden flat field as null or an otherwise empty value.
func (v *ExplorationView) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	if composition, present := fields["composition"]; present {
		if bytes.Equal(bytes.TrimSpace(composition), []byte("null")) {
			return errors.New("omit an unused composition rather than supplying null")
		}
		for _, name := range []string{"layout", "columns", "fieldColumns", "displays", "breakdowns", "tail"} {
			if _, exists := fields[name]; exists {
				return errors.New("composition owns presentation; omit flat layout, columns, displays, comparisons and tail")
			}
		}
	}
	type wire ExplorationView
	var value wire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	*v = ExplorationView(value)
	return nil
}

// ViewComposition orders supplied aspects of one checked record query. Children
// cannot own queries, subscriptions, captured data or completeness claims.
type ViewComposition struct {
	Version  string                 `json:"version"`
	Kind     CompositionKind        `json:"kind"`
	Children []ViewCompositionChild `json:"children"`
}

type CompositionKind struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
}

// A child chooses exactly one typed role. Its role is its stable identity;
// duplicate roles, arbitrary IDs and recursive layout are deliberately absent.
type ViewCompositionChild struct {
	Aggregate *ViewDeclaration       `json:"aggregate,omitempty"`
	Records   *RecordViewDeclaration `json:"records,omitempty"`
	Hidden    bool                   `json:"hidden,omitempty"`
}

func (c *ViewCompositionChild) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	_, aggregate := fields["aggregate"]
	_, records := fields["records"]
	if aggregate == records {
		return errors.New("composition children require exactly one aggregate or records role")
	}
	if hidden, present := fields["hidden"]; present && bytes.Equal(bytes.TrimSpace(hidden), []byte("null")) {
		return errors.New("child visibility requires a boolean")
	}
	type wire ViewCompositionChild
	var value wire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	*c = ViewCompositionChild(value)
	return nil
}

// RecordViewDeclaration consumes a bounded admitted record page. It does not
// turn that page into a population denominator or a complete ranked dataset.
type RecordViewDeclaration struct {
	Version string               `json:"version"`
	Scope   ViewDeclarationScope `json:"scope"`
	Kind    RecordViewKind       `json:"kind"`
	Display RecordViewDisplay    `json:"display"`
}

type RecordViewKind struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
}

type RecordViewDisplay struct {
	Columns      []string            `json:"columns"`
	FieldColumns []exploration.Field `json:"fieldColumns,omitempty"`
}

func (d *RecordViewDisplay) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	if selected, present := fields["fieldColumns"]; present && bytes.Equal(bytes.TrimSpace(selected), []byte("null")) {
		return errors.New("omit unused record field columns rather than supplying null")
	}
	type wire RecordViewDisplay
	var value wire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	*d = RecordViewDisplay(value)
	return nil
}

func (d RecordViewDeclaration) Validate(q exploration.Query) error {
	if d.Version != "goo.view/1" || d.Scope.Kind != "record-query" || d.Scope.Signal != q.Signal || d.Kind.Name != "record-table" || d.Kind.Version != 1 {
		return errors.New("require a pinned record-table display for the configured record query")
	}
	return validateRecordColumns(d.Display.Columns, d.Display.FieldColumns, []exploration.Query{q})
}

func validateRecordColumns(columns []string, fields []exploration.Field, queries []exploration.Query) error {
	if len(columns) < 2 || len(columns) > 4 {
		return errors.New("require 2..4 record columns")
	}
	seen := map[string]bool{}
	for _, c := range columns {
		if seen[c] || (c != "utc" && c != "service" && c != "summary" && c != "context") {
			return errors.New("invalid or duplicate view column")
		}
		seen[c] = true
	}
	if !seen["utc"] || !seen["summary"] {
		return errors.New("UTC and summary columns required")
	}
	if len(fields) > 8 {
		return errors.New("at most eight record field columns")
	}
	selected := map[exploration.Field]bool{}
	for _, field := range fields {
		if selected[field] {
			return errors.New("duplicate record field column")
		}
		selected[field] = true
		valid := false
		for _, q := range queries {
			if _, err := field.Type(q.Signal); err == nil {
				valid = true
			}
		}
		if !valid {
			return errors.New("unsupported record field column")
		}
	}
	return nil
}

func (c ViewComposition) Validate(q exploration.Query) error {
	if c.Version != "goo.composition/1" || c.Kind.Name != "stack" || c.Kind.Version != 1 || len(c.Children) != 2 {
		return errors.New("require a pinned two-child record composition")
	}
	aggregate, records, visible := false, false, false
	for _, child := range c.Children {
		if (child.Aggregate == nil) == (child.Records == nil) {
			return errors.New("composition children require exactly one aggregate or records role")
		}
		visible = visible || !child.Hidden
		if child.Aggregate != nil {
			if aggregate {
				return errors.New("composition requires exactly one aggregate child")
			}
			aggregate = true
			if err := child.Aggregate.Validate(q); err != nil {
				return err
			}
		} else {
			if records {
				return errors.New("composition requires exactly one records child")
			}
			records = true
			if err := child.Records.Validate(q); err != nil {
				return err
			}
		}
	}
	if !aggregate || !records || !visible {
		return errors.New("composition requires both roles and at least one visible child")
	}
	return nil
}

// ReuseViewComposition copies only presentation intent into another checked
// population. Neither query nor a source page, runtime guarantee or action owner
// is part of the returned value. The caller retains its destination query.
func ReuseViewComposition(c ViewComposition, source, destination exploration.Query) (ViewComposition, error) {
	if err := c.Validate(source); err != nil {
		return ViewComposition{}, err
	}
	out := c
	out.Children = slices.Clone(c.Children)
	for i, child := range c.Children {
		if child.Aggregate != nil {
			display, err := ReuseViewDeclaration(*child.Aggregate, source, destination)
			if err != nil {
				return ViewComposition{}, err
			}
			out.Children[i].Aggregate = &display
		} else {
			if err := child.Records.Validate(destination); err != nil {
				return ViewComposition{}, err
			}
			display := *child.Records
			display.Display.Columns = slices.Clone(child.Records.Display.Columns)
			display.Display.FieldColumns = slices.Clone(child.Records.Display.FieldColumns)
			out.Children[i].Records = &display
		}
	}
	if err := out.Validate(destination); err != nil {
		return ViewComposition{}, err
	}
	return out, nil
}
