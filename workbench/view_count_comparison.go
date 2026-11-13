package workbench

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/casselc/goo/exploration"
)

// CountComparisonViewDeclaration is the ordinary comparison operand union. A pinned
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
	if d.Version != "goo.view/1" || d.Scope.Kind != "record-query" || d.Scope.Signal != q.Signal || (d.Kind.Name != "captured-count-comparison" && d.Kind.Name != "query-count-comparison") || d.Kind.Version != 1 || (d.Display.Value != "count" && d.Display.Value != "duration.mean" && d.Display.Value != exploration.RecordRateKey && d.Display.Value != "status.error" && d.Display.Value != "status.error_fraction") {
		return errors.New("require a pinned ordinary count, captured-rate, mean-duration or recorded-status comparison display for the configured record query")
	}
	if q.SpanStatus || d.Display.Value == "status.error" || d.Display.Value == "status.error_fraction" {
		if d.Kind.Name != "query-count-comparison" || q.Signal != "spans" || !q.SpanStatus || d.Display.View != "" && d.Display.View != "summary" || d.Display.Order != "" {
			return errors.New("recorded-status comparison requires fresh span-status queries, summary view and no group ordering")
		}
	}
	if d.Display.Value == exploration.RecordRateKey {
		if d.Kind.Name != "captured-count-comparison" || !exploration.RecordRateSummaryQuery(q) || d.Display.View != "" && d.Display.View != "summary" || d.Display.Order != "" {
			return errors.New("captured-rate comparison requires an immutable capture, summary view and no group ordering")
		}
	}
	if d.Display.Value == "duration.mean" {
		// Summary means retain the whole population's exact duration sum/count.
		// Count groups and captured operands are separate authoring contracts;
		// neither becomes a per-category or captured mean through this display.
		if d.Kind.Name != "query-count-comparison" || q.Signal != "spans" || q.Summary != "span-duration" || d.Display.View != "" && d.Display.View != "summary" || d.Display.Order != "" {
			return errors.New("mean-duration comparison requires fresh span-duration queries, summary view and no group ordering")
		}
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
		// Validate both questions before doing window arithmetic. Compatible
		// cohorts retain their independently checked predicates and profile.
		if err := exploration.CompatibleCountCohortQueries(q, baseline); err != nil {
			return err
		}
		width := q.ToMS - q.FromMS
		sameWindow := q.FromMS == baseline.FromMS && q.ToMS == baseline.ToMS
		previousWindow := q.FromMS >= width && baseline.FromMS == q.FromMS-width && baseline.ToMS == q.FromMS
		if q.Offset != 0 || baseline.Offset != 0 || !sameWindow && !previousWindow {
			return errors.New("query comparison requires page-zero queries in the same fixed window or an adjacent equal-duration previous window")
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
	if d.Kind.Name == "query-count-comparison" {
		return exploration.ValidateCountCohortQuery(q)
	}
	return exploration.ValidateCapturedCountQuery(q)
}
