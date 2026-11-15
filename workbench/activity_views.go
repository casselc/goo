package workbench

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"slices"
)

// ActivitySelector is configuration, not a physical tail request. "both" is
// admitted only here and projected into two independently validated queries.
type ActivitySelector exploration.TailSelector

// activityObject closes only saved Activity configuration objects. It does not
// change the JSON policy of physical queries or other persisted resources.
func activityObject(raw []byte, allowed ...string) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, err
	}
	if fields == nil {
		return nil, errors.New("Activity configuration requires an object")
	}
	for key := range fields {
		if !slices.Contains(allowed, key) {
			return nil, fmt.Errorf("unknown Activity field %q", key)
		}
	}
	return fields, nil
}

func (s *ActivitySelector) UnmarshalJSON(raw []byte) error {
	fields, err := activityObject(raw, "signal", "environment", "environmentSelector", "where")
	if err != nil {
		return err
	}
	for _, key := range []string{"signal", "environment", "where"} {
		if value, present := fields[key]; present && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("Activity selector %s must not be null", key)
		}
	}
	type wire ActivitySelector
	var decoded wire
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		return err
	}
	*s = ActivitySelector(decoded)
	return nil
}

func (s ActivitySelector) Queries() ([]exploration.Query, error) {
	signals := []string{s.Signal}
	if s.Signal == "both" {
		signals = []string{"logs", "spans"}
	}
	queries := make([]exploration.Query, 0, len(signals))
	for _, signal := range signals {
		physical := exploration.TailSelector(s)
		physical.Signal = signal
		q := physical.Query(60000)
		detached, err := exploration.ClonePredicate(q.Where, q.Signal)
		if err != nil {
			return nil, fmt.Errorf("%s Activity predicate: %w", q.Signal, err)
		}
		q.Where = detached
		if err := q.Validate(); err != nil {
			return nil, fmt.Errorf("%s Activity selector: %w", q.Signal, err)
		}
		queries = append(queries, q)
	}
	return queries, nil
}

type ActivityLayout struct {
	Columns      []string            `json:"columns"`
	FieldColumns []exploration.Field `json:"fieldColumns,omitempty"`
}

func (layout *ActivityLayout) UnmarshalJSON(raw []byte) error {
	fields, err := activityObject(raw, "columns", "fieldColumns")
	if err != nil {
		return err
	}
	type wire ActivityLayout
	var decoded wire
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		return err
	}
	for _, key := range []string{"columns", "fieldColumns"} {
		if value, present := fields[key]; present && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("Activity %s must be an array", key)
		}
	}
	*layout = ActivityLayout(decoded)
	return nil
}

type ActivityLanes struct {
	Logs  ActivityLayout `json:"logs"`
	Spans ActivityLayout `json:"spans"`
}

func (lanes *ActivityLanes) UnmarshalJSON(raw []byte) error {
	if _, err := activityObject(raw, "logs", "spans"); err != nil {
		return err
	}
	type wire ActivityLanes
	var decoded wire
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		return err
	}
	*lanes = ActivityLanes(decoded)
	return nil
}

// ActivityView stores a rolling selector and presentation, without observations,
// fixed windows, membership capabilities or an execution instruction.
type ActivityView struct {
	Version         int                 `json:"version"`
	ID              string              `json:"id,omitempty"`
	Title           string              `json:"title"`
	Selector        ActivitySelector    `json:"selector"`
	Columns         []string            `json:"columns,omitempty"`
	FieldColumns    []exploration.Field `json:"fieldColumns,omitempty"`
	BaselineCapture string              `json:"baselineCapture,omitempty"`
	Lanes           *ActivityLanes      `json:"lanes,omitempty"`
}

// UnmarshalJSON keeps the composition branches exclusive, including explicit
// empty/null fields that a zero Go value alone cannot distinguish from omission.
func (v *ActivityView) UnmarshalJSON(raw []byte) error {
	fields, err := activityObject(raw, "version", "id", "title", "selector", "columns", "fieldColumns", "baselineCapture", "lanes")
	if err != nil {
		return err
	}
	type wire ActivityView
	var decoded wire
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		return err
	}
	for _, key := range []string{"columns", "fieldColumns", "baselineCapture"} {
		if value, present := fields[key]; present && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("Activity %s must not be null", key)
		}
	}
	if decoded.Selector.Signal == "both" {
		for _, key := range []string{"columns", "fieldColumns", "baselineCapture"} {
			if _, present := fields[key]; present {
				return fmt.Errorf("paired Activity omits top-level %s", key)
			}
		}
	} else if _, present := fields["lanes"]; present {
		return errors.New("single Activity omits lanes")
	}
	*v = ActivityView(decoded)
	return nil
}

func validateActivityLayout(title string, layout ActivityLayout, q exploration.Query) error {
	for _, column := range layout.Columns {
		if column == "context" {
			return errors.New("Activity supports UTC, service and summary columns")
		}
	}
	// The synthetic window validates configuration only. It is not persisted or read.
	return (ExplorationView{Version: 1, Title: title, Queries: []exploration.Query{q}, Layout: "records", Columns: layout.Columns, FieldColumns: layout.FieldColumns}).Validate()
}

func (v ActivityView) Validate() error {
	if v.Version != 1 || !text(v.Title, 120, true, false) {
		return errors.New("require version 1 and bounded Activity view title")
	}
	queries, err := v.Selector.Queries()
	if err != nil {
		return err
	}
	if v.Selector.Signal == "both" {
		if v.Lanes == nil || v.Columns != nil || v.FieldColumns != nil || v.BaselineCapture != "" {
			return errors.New("paired Activity requires logs/spans lanes and omits flat presentation and baseline")
		}
		for i, layout := range []ActivityLayout{v.Lanes.Logs, v.Lanes.Spans} {
			if err := validateActivityLayout(v.Title, layout, queries[i]); err != nil {
				return fmt.Errorf("%s lane: %w", queries[i].Signal, err)
			}
		}
		return nil
	}
	if v.Lanes != nil {
		return errors.New("single Activity omits lanes")
	}
	if v.BaselineCapture != "" && (!ID.MatchString(v.BaselineCapture) || v.Selector.Signal != "spans") {
		return errors.New("baselineCapture requires an exact capture ID and span selector")
	}
	return validateActivityLayout(v.Title, ActivityLayout{Columns: v.Columns, FieldColumns: v.FieldColumns}, queries[0])
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
	var selection *identity.EnvironmentSelection
	if environment != "" {
		checked, err := identity.CurrentEnvironment(environment)
		if err != nil {
			return nil, err
		}
		selection = &checked
	}
	return s.ActivityViewsForEnvironment(after, selection)
}

// ActivityViewsForEnvironment filters by checked policy and scope, never by a
// display name alone. A nil selection is an explicit unfiltered catalog request.
func (s Store) ActivityViewsForEnvironment(after string, selection *identity.EnvironmentSelection) (*ActivityViewList, error) {
	if after != "" && !ID.MatchString(after) {
		return nil, errors.New("invalid Activity view cursor")
	}
	if selection != nil {
		if err := selection.Validate(); err != nil {
			return nil, err
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
		if selection != nil {
			queries, err := v.Selector.Queries()
			if err != nil {
				return nil, err
			}
			candidate, err := queries[0].EnvironmentSelection()
			if err != nil {
				return nil, err
			}
			if !selection.Equal(candidate) {
				continue
			}
		}
		if len(out.Items) == 20 {
			out.NextCursor = out.Items[19].ID
			break
		}
		out.Items = append(out.Items, *v)
	}
	return out, nil
}
