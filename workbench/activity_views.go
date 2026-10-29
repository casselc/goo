package workbench

import (
	"encoding/json"
	"errors"
	"github.com/casselc/goo/exploration"
)

// ActivityView stores a rolling selector and presentation, without observations,
// fixed windows, membership capabilities or an execution instruction.
type ActivityView struct {
	Version         int                      `json:"version"`
	ID              string                   `json:"id,omitempty"`
	Title           string                   `json:"title"`
	Selector        exploration.TailSelector `json:"selector"`
	Columns         []string                 `json:"columns"`
	FieldColumns    []exploration.Field      `json:"fieldColumns,omitempty"`
	BaselineCapture string                   `json:"baselineCapture,omitempty"`
}

func (v ActivityView) Validate() error {
	if v.Selector.EnvironmentSelector != nil {
		return errors.New("saved Activity views do not support environmentSelector yet")
	}
	if v.Version != 1 || !text(v.Title, 120, true, false) {
		return errors.New("require version 1 and bounded Activity view title")
	}
	if v.BaselineCapture != "" && (!ID.MatchString(v.BaselineCapture) || v.Selector.Signal != "spans") {
		return errors.New("baselineCapture requires an exact capture ID and span selector")
	}
	if err := v.Selector.Query(60000).Validate(); err != nil {
		return err
	}
	// Reuse fixed-view presentation validation with a synthetic validation window
	// only. That question is never stored, executed or represented as source evidence.
	for _, c := range v.Columns {
		if c == "context" {
			return errors.New("Activity supports UTC, service and summary columns")
		}
	}
	return (ExplorationView{Version: 1, Title: v.Title, Queries: []exploration.Query{v.Selector.Query(60000)}, Layout: "records", Columns: v.Columns, FieldColumns: v.FieldColumns}).Validate()
}
func (s Store) CreateActivityView(v ActivityView) (*ActivityView, error) {
	if v.ID != "" {
		return nil, errors.New("new Activity view must omit ID")
	}
	if e := v.Validate(); e != nil {
		return nil, e
	}
	raw, _ := json.Marshal(v)
	var detached ActivityView
	if e := json.Unmarshal(raw, &detached); e != nil {
		return nil, e
	}
	v = detached
	id, e := digest(v)
	if e != nil {
		return nil, e
	}
	v.ID = id
	if e = s.publish("activity-views", id, v); e != nil {
		return nil, e
	}
	return s.LoadActivityView(id)
}
func (s Store) LoadActivityView(id string) (*ActivityView, error) {
	var v ActivityView
	if e := s.read("activity-views", id, &v); e != nil {
		return nil, e
	}
	actual := v.ID
	v.ID = ""
	hash, e := digest(v)
	v.ID = actual
	if e != nil || hash != id || actual != id || v.Validate() != nil {
		return nil, errors.New("Activity view integrity check failed")
	}
	return &v, nil
}

type ActivityViewList struct {
	Items      []ActivityView `json:"items"`
	NextCursor string         `json:"nextCursor,omitempty"`
}

func (s Store) ActivityViews(after, environment string) (*ActivityViewList, error) {
	if after != "" && !ID.MatchString(after) {
		return nil, errors.New("invalid Activity view cursor")
	}
	if environment != "" {
		if e := (exploration.TailSelector{Environment: environment}).Query(60000).Validate(); e != nil {
			return nil, e
		}
	}
	ids, e := s.ids("activity-views")
	if e != nil {
		return nil, e
	}
	out := &ActivityViewList{Items: []ActivityView{}}
	for _, id := range ids {
		if id <= after {
			continue
		}
		v, e := s.LoadActivityView(id)
		if e != nil {
			return nil, e
		}
		if environment != "" && v.Selector.Environment != environment {
			continue
		}
		if len(out.Items) == 20 {
			out.NextCursor = out.Items[19].ID
			break
		}
		out.Items = append(out.Items, *v)
	}
	return out, nil
}
