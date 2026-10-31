package workbench

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/casselc/goo/exploration"
)

// CapturedCountViewDeclaration retains an immutable baseline reference, never
// captured rows or a comparison result. The parent owns the applied current query.
type CapturedCountViewDeclaration struct {
	Version         string                   `json:"version"`
	Scope           ViewDeclarationScope     `json:"scope"`
	Kind            CapturedCountViewKind    `json:"kind"`
	BaselineCapture string                   `json:"baselineCapture"`
	Display         CapturedCountViewDisplay `json:"display"`
}
type CapturedCountViewKind struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
}
type CapturedCountViewDisplay struct {
	Value string `json:"value"`
	View  string `json:"view,omitempty"`
	Order string `json:"order,omitempty"`
	Title string `json:"title,omitempty"`
}

func (d *CapturedCountViewDisplay) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for _, name := range []string{"title", "view", "order"} {
		if raw, ok := fields[name]; ok && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return errors.New("omit unused comparison display options rather than supplying null")
		}
	}
	type wire CapturedCountViewDisplay
	var value wire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	*d = CapturedCountViewDisplay(value)
	return nil
}

func (d CapturedCountViewDeclaration) Validate(q exploration.Query) error {
	if d.Version != "goo.view/1" || d.Scope.Kind != "record-query" || d.Scope.Signal != q.Signal || d.Kind.Name != "captured-count-comparison" || d.Kind.Version != 1 || d.Display.Value != "count" {
		return errors.New("require a pinned captured-count-comparison display for the configured record query")
	}
	if !ID.MatchString(d.BaselineCapture) {
		return errors.New("comparison requires an exact immutable baseline capture ID")
	}
	if !text(d.Display.Title, 120, d.Display.Title != "", false) {
		return errors.New("comparison title must be bounded valid text")
	}
	if d.Display.View != "" && d.Display.View != "summary" && d.Display.View != "groups" {
		return errors.New("comparison view requires summary or groups")
	}
	if d.Display.Order != "" && d.Display.Order != "change" && d.Display.Order != "label" {
		return errors.New("comparison order requires change or label")
	}
	return exploration.ValidateCapturedCountQuery(q)
}
