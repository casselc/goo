package workbench

import (
	"errors"

	"github.com/casselc/goo/exploration"
)

// TimelineViewDeclaration consumes the supplied unsplit time dataset. Bucket
// width belongs to the parent query; this declaration never requests data or
// promotes captured counts to rates, population shares or complete telemetry.
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
	Value            string `json:"value"`
	Title            string `json:"title,omitempty"`
	RequiresComplete bool   `json:"requiresComplete,omitempty"`
}

func (d TimelineViewDeclaration) Validate(q exploration.Query) error {
	if d.Version != "goo.view/1" || d.Scope.Kind != "record-query" || d.Scope.Signal != q.Signal || d.Kind.Version != 1 || (d.Kind.Name != "time-bar" && d.Kind.Name != "table") {
		return errors.New("require a pinned time-bar or table timeline for the configured record query")
	}
	if q.BucketMS <= 0 || q.GroupTimeSeries || q.NumericHistogram != nil {
		return errors.New("timeline requires explicit bounded unsplit time buckets")
	}
	if err := q.Validate(); err != nil {
		return err
	}
	if d.Display.Value != "count" {
		return errors.New("timeline supports captured record counts only")
	}
	if !text(d.Display.Title, 120, d.Display.Title != "", false) {
		return errors.New("timeline title must be bounded valid text")
	}
	return nil
}
