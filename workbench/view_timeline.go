package workbench

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
)

// TimelineViewDeclaration consumes the supplied unsplit time dataset. Bucket
// width belongs to the parent query; this declaration never requests data or
// promotes captured counts to rates, population shares or complete telemetry.
// Duration measures require raw-span summaries from the checked parent query.
type TimelineViewDeclaration struct {
	Version string               `json:"version"`
	Scope   ViewDeclarationScope `json:"scope"`
	Kind    TimelineViewKind     `json:"kind"`
	Display TimelineViewDisplay  `json:"display"`
}

type TimelineViewKind struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
}

type TimelineViewDisplay struct {
	Profile          string `json:"profile,omitempty"`
	Query            string `json:"query,omitempty"`
	Order            string `json:"order,omitempty"`
	Value            string `json:"value"`
	Title            string `json:"title,omitempty"`
	RequiresComplete bool   `json:"requiresComplete,omitempty"`
}

// New local-control fields are optional strings, not nullable reset commands.
func (d *TimelineViewDisplay) UnmarshalJSON(data []byte) error {
	if !identity.WellFormedJSON(data) {
		return errors.New("timeline text must be valid Unicode")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for _, name := range []string{"query", "order", "title"} {
		if raw, present := fields[name]; present {
			var value string
			if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) || json.Unmarshal(raw, &value) != nil {
				return errors.New("timeline " + name + " must be a non-null string")
			}
		}
	}
	if raw, present := fields["requiresComplete"]; present && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return errors.New("timeline complete-only must not be null")
	}

	type wire TimelineViewDisplay
	var value wire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	*d = TimelineViewDisplay(value)
	return nil
}

func (d TimelineViewDeclaration) Validate(q exploration.Query) error {
	if q.GroupTimeSeries || d.Display.Profile != "" {
		return d.validateGrouped(q)
	}
	if d.Display.Query != "" || d.Display.Order != "" {
		return errors.New("unsplit timelines do not define grouped local controls")
	}
	if d.Version != "goo.view/1" || d.Scope.Kind != "record-query" || d.Scope.Signal != q.Signal || d.Kind.Version != 1 || (d.Kind.Name != "time-bar" && d.Kind.Name != "table") {
		return errors.New("require a pinned time-bar or table timeline for the configured record query")
	}
	if q.BucketMS <= 0 || q.GroupTimeSeries || q.NumericHistogram != nil {
		return errors.New("timeline requires explicit bounded unsplit time buckets")
	}
	if err := q.Validate(); err != nil {
		return err
	}
	available := d.Display.Value == "count" || d.Display.Value == exploration.RecordRateKey && exploration.RecordRateQuery(q)
	if q.Signal == "spans" && q.Summary == "span-duration" && (d.Display.Value == "duration.mean" || d.Display.Value == "duration.p95") {
		for _, field := range exploration.ViewMeasureFields(q) {
			if field.Key == d.Display.Value && field.Aggregation != nil && field.Aggregation.Family == "duration" {
				available = true
				break
			}
		}
	}
	if !available {
		return errors.New("choose count, captured record rate or a mean/p95 duration supplied by the checked span query")
	}
	if !text(d.Display.Title, 120, d.Display.Title != "", false) {
		return errors.New("timeline title must be bounded valid text")
	}
	return nil
}

func (d TimelineViewDeclaration) validateGrouped(q exploration.Query) error {
	count := d.Display.Profile == "grouped-count-v1" && exploration.GroupedCountTimelineQuery(q) && d.Display.Value == "count"
	mean := d.Display.Profile == "grouped-duration-mean-v1" && exploration.GroupedMeanDurationTimelineQuery(q) && d.Display.Value == "duration.mean"
	if d.Version != "goo.view/1" || d.Scope.Kind != "record-query" || d.Scope.Signal != q.Signal || d.Kind.Version != 1 || !(count || mean) {
		return errors.New("grouped timelines require the pinned count or raw-span mean-duration profile")
	}
	if d.Kind.Name != "small-multiples" && d.Kind.Name != "table" && d.Kind.Name != "heatmap" {
		return errors.New("choose small-multiples, heatmap or table for grouped timelines")
	}
	if d.Kind.Name == "heatmap" && 1+(q.ToMS-q.FromMS-1)/q.BucketMS > 32 {
		return errors.New("grouped heatmaps require at most 32 fixed intervals")
	}
	if d.Display.Order != "" && d.Display.Order != "source" && d.Display.Order != "label" {
		return errors.New("choose a supported local group order")
	}
	if !text(d.Display.Query, 256, false, false) || !text(d.Display.Title, 120, d.Display.Title != "", false) {
		return errors.New("grouped timeline title and Find text must be bounded valid text")
	}
	return nil
}

// ReuseTimelineViewDeclaration transfers display intent only. Group identity and
// fixed width must be compatible; destination selectors and bounds never travel
// from the donor configuration.
func ReuseTimelineViewDeclaration(d TimelineViewDeclaration, source, destination exploration.Query) (TimelineViewDeclaration, error) {
	if e := d.Validate(source); e != nil {
		return TimelineViewDeclaration{}, e
	}
	if source.Signal != destination.Signal || source.BucketMS != destination.BucketMS {
		return TimelineViewDeclaration{}, errors.New("timeline requires the same signal and explicit bucket width")
	}
	if (d.Display.Profile == "grouped-count-v1" || d.Display.Profile == "grouped-duration-mean-v1") && (len(destination.GroupBy) != 1 || source.GroupBy[0] != destination.GroupBy[0]) {
		return TimelineViewDeclaration{}, errors.New("grouped timeline requires the same qualified grouping field")
	}
	if e := d.Validate(destination); e != nil {
		return TimelineViewDeclaration{}, e
	}
	return d, nil
}
