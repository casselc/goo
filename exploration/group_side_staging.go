package exploration

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// GroupSideIntent describes one supplied group and side. Its context is not a
// runtime authorization lease; an application must separately bind its owner.
type GroupSideIntent struct {
	Type       string `json:"type"`
	Side       string `json:"side"`
	Key        string `json:"key"`
	DatasetID  string `json:"datasetId"`
	Revision   string `json:"revision"`
	ScopeID    string `json:"scopeId"`
	Generation string `json:"generation"`
}

func (i *GroupSideIntent) UnmarshalJSON(data []byte) error {
	type wire GroupSideIntent
	var value wire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return errors.New("group-side intent requires one JSON object")
	}
	*i = GroupSideIntent(value)
	return nil
}

// StageGroupSide derives question intent without reading or retaining telemetry.
// It reconstructs the existing validated comparison, not a second group validator.
// The caller still owns active/visible/draft/depth and stale-action authorization.
func StageGroupSide(request ComparisonRequest, intent GroupSideIntent) (Query, error) {
	fail := func() (Query, error) {
		return Query{}, errors.New("choose a current reported positive group and exact side context")
	}
	if intent.Type != "scope-group-side" || intent.Side != "baseline" && intent.Side != "selection" {
		return fail()
	}
	result, err := CompareCohorts(request)
	if err != nil {
		return Query{}, err
	}
	comparison := result.GroupComparison
	if comparison == nil || comparison.State == "diagnostic" || comparison.BaselineAccounting != "coherent" || comparison.SelectionAccounting != "coherent" {
		return fail()
	}
	snapshot := comparison.Baseline
	if intent.Side == "selection" {
		snapshot = comparison.Selection
	}
	if snapshot == nil || intent.DatasetID != snapshot.ID || intent.Revision != snapshot.Revision || intent.ScopeID != snapshot.Scope.ID || intent.Generation != snapshot.Scope.Generation {
		return fail()
	}
	for _, row := range comparison.Rows {
		if row.Key != intent.Key {
			continue
		}
		cell := row.Baseline
		if intent.Side == "selection" {
			cell = row.Selection
		}
		if cell.Presence != "present" || cell.State != "derived" || cell.Value == nil {
			return fail()
		}
		count, err := decimal(*cell.Value)
		if err != nil || count.Sign() <= 0 {
			return fail()
		}
		var q Query
		if err := json.Unmarshal(snapshot.Scope.Selector, &q); err != nil {
			return Query{}, err
		}
		q.Offset = 0
		field := comparison.Field
		leaf := Predicate{Field: &field, Op: "missing"}
		if row.Value != nil {
			leaf.Op, leaf.Value = "eq", *row.Value
		}
		if q.Where == nil {
			q.Where = &leaf
		} else {
			q.Where = &Predicate{All: []Predicate{*q.Where, leaf}}
		}
		if err := q.Validate(); err != nil {
			return Query{}, fmt.Errorf("group question cannot be staged: %w", err)
		}
		return q, nil
	}
	return fail()
}
