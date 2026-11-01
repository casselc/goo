package workbench

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/casselc/goo/exploration"
)

// CountComparisonViewDeclaration is the count-comparison operand union. A pinned
// kind chooses either an immutable capture ID or one portable fixed baseline
// query. The parent owns the current query; neither arm carries captured results.
type CountComparisonViewDeclaration struct {
	Version         string                     `json:"version"`
	Scope           ViewDeclarationScope       `json:"scope"`
	Kind            CountComparisonViewKind    `json:"kind"`
	BaselineCapture string                     `json:"baselineCapture,omitempty"`
	BaselineQuery   *exploration.Query         `json:"baselineQuery,omitempty"`
	Display         CountComparisonViewDisplay `json:"display"`
}
type CountComparisonViewKind struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
}
type CountComparisonViewDisplay struct {
	Value string `json:"value"`
	View  string `json:"view,omitempty"`
	Order string `json:"order,omitempty"`
	Title string `json:"title,omitempty"`
}

// Keep operand arms exclusive even for explicit null and empty wire values.
func (d *CountComparisonViewDeclaration) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	var kind CountComparisonViewKind
	if err := json.Unmarshal(fields["kind"], &kind); err != nil {
		return err
	}
	_, capture := fields["baselineCapture"]
	query, hasQuery := fields["baselineQuery"]
	switch kind.Name {
	case "captured-count-comparison":
		if !capture || hasQuery {
			return errors.New("captured comparison requires only baselineCapture")
		}
	case "query-count-comparison":
		if capture || !hasQuery || bytes.Equal(bytes.TrimSpace(query), []byte("null")) {
			return errors.New("query comparison requires only a non-null baselineQuery")
		}
	default:
		return errors.New("choose a pinned captured-count-comparison or query-count-comparison kind")
	}
	type wire CountComparisonViewDeclaration
	var value wire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	*d = CountComparisonViewDeclaration(value)
	return nil
}

func (d *CountComparisonViewDisplay) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for _, name := range []string{"title", "view", "order"} {
		if raw, ok := fields[name]; ok && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return errors.New("omit unused comparison display options rather than supplying null")
		}
	}
	type wire CountComparisonViewDisplay
	var value wire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	*d = CountComparisonViewDisplay(value)
	return nil
}

func (d CountComparisonViewDeclaration) Validate(q exploration.Query) error {
	if d.Version != "goo.view/1" || d.Scope.Kind != "record-query" || d.Scope.Signal != q.Signal || (d.Kind.Name != "captured-count-comparison" && d.Kind.Name != "query-count-comparison") || d.Kind.Version != 1 || d.Display.Value != "count" {
		return errors.New("require a pinned count-comparison display for the configured record query")
	}
	if d.Kind.Name == "captured-count-comparison" {
		if !ID.MatchString(d.BaselineCapture) || d.BaselineQuery != nil {
			return errors.New("captured comparison requires only an exact immutable baseline capture ID")
		}
	} else {
		if d.BaselineCapture != "" || d.BaselineQuery == nil {
			return errors.New("query comparison requires only a fixed baseline query")
		}
		baseline := *d.BaselineQuery
		if q.Offset != 0 || baseline.Offset != 0 || q.FromMS != baseline.FromMS || q.ToMS != baseline.ToMS {
			return errors.New("query comparison requires matching fixed windows and page-zero queries")
		}
		if err := exploration.CompatibleCountCohortQueries(q, baseline); err != nil {
			return err
		}
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
