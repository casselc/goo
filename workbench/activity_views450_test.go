package workbench

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"reflect"
	"testing"
)

func laneCondition450(key string) *exploration.Predicate {
	return &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: key}, Op: "eq", Value: "error"}
}
func TestActivityLaneConditions450(t *testing.T) {
	v := pairedActivityFixture()
	v.Lanes.Logs.Where = laneCondition450("body")
	v.Lanes.Spans.Where = laneCondition450("name")
	q, err := v.Queries()
	if err != nil {
		t.Fatal(err)
	}
	for i, w := range []*exploration.Predicate{v.Lanes.Logs.Where, v.Lanes.Spans.Where} {
		if len(q[i].Where.All) != 2 || !reflect.DeepEqual(q[i].Where.All[1], *w) {
			t.Fatal("lost lane", q[i])
		}
	}
	q[0].Where.All[1].Value = "mutated"
	if v.Lanes.Logs.Where.Value != "error" || q[1].Where.All[1].Value != "error" {
		t.Fatal("projection aliases")
	}
	store := Store{Dir: t.TempDir()}
	saved, err := store.CreateActivityView(v)
	if err != nil {
		t.Fatal(err)
	}
	held, err := store.LoadActivityView(saved.ID)
	if err != nil || !reflect.DeepEqual(held.Lanes, v.Lanes) {
		t.Fatal("round trip", err)
	}
	v.Lanes.Logs.Where.Value = "different"
	next, err := store.CreateActivityView(v)
	if err != nil || next.ID == saved.ID {
		t.Fatal("lane not bound to content identity", err)
	}
	if held.Lanes.Logs.Where.Value != "error" {
		t.Fatal("store aliases input")
	}
	v.Selector.Where = nil
	if err := v.Validate(); err != nil {
		t.Fatal("lane without shared", err)
	}
}
func TestActivityLaneConditionRefusals450(t *testing.T) {
	for _, key := range []string{"name", "duration_ns"} {
		v := pairedActivityFixture()
		v.Lanes.Logs.Where = laneCondition450(key)
		if v.Validate() == nil {
			t.Fatal("wrong-family lane admitted", key)
		}
	}
	v := pairedActivityFixture()
	v.Lanes.Spans.Where = laneCondition450("body")
	if v.Validate() == nil {
		t.Fatal("log field in spans")
	}
	v = pairedActivityFixture()
	many := make([]exploration.Predicate, 30)
	for i := range many {
		many[i] = *laneCondition450("service")
	}
	v.Selector.Where = &exploration.Predicate{All: many}
	v.Lanes.Logs.Where = laneCondition450("body")
	if v.Validate() == nil {
		t.Fatal("combined budget bypass")
	}
	for _, raw := range []string{`{"columns":["utc","summary"],"where":null}`, `{"columns":["utc","summary"],"where":{"op":"eq"}}`, `{"columns":["utc","summary"],"condition":{}}`} {
		var layout ActivityLayout
		err := json.Unmarshal([]byte(raw), &layout)
		if err == nil {
			v := pairedActivityFixture()
			v.Lanes.Logs = layout
			err = v.Validate()
		}
		if err == nil {
			t.Fatal("malformed lane admitted", raw)
		}
	}
}
