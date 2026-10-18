package exploration

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func followupFixture() SavedCapture {
	q := Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 20, Offset: 0, Where: &Predicate{All: []Predicate{followupEqual("service", "edge"), {Field: &Field{Kind: "attribute", Key: "route"}, Op: "eq", Value: "/checkout"}}}}
	row := json.RawMessage(`{"ServiceName":"edge","TraceId":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","SpanId":"bbbbbbbbbbbbbbbb","SpanName":"checkout","Attributes":{"goo.operation.outcome":"ok"}}`)
	return SavedCapture{CaptureReceipt: CaptureReceipt{ID: strings.Repeat("c", 64), Question: CaptureQuestion{Kind: "explore", Explore: &q}}, Explore: &Result{Question: q, Rows: []json.RawMessage{row}}}
}
func TestFollowupScopeOwnershipAndDetachedPlans(t *testing.T) {
	c := followupFixture()
	before, _ := json.Marshal(c)
	plans, e := c.Followups(0)
	if e != nil {
		t.Fatal(e)
	}
	found := map[string]Followup{}
	for _, a := range plans.Actions {
		found[a.ID] = a
	}
	if !reflect.DeepEqual(found["duration-status"].Query.Where, c.Question.Explore.Where) || found["duration-status"].Query.Summary != "span-duration" {
		t.Fatal("population changed")
	}
	for _, id := range []string{"trace", "trace-logs"} {
		q := found[id].Query
		if q == nil || q.Environment != "prod" || q.FromMS != 1000 || q.ToMS != 61000 || q.Where.Field.Key != "trace_id" || q.Where.Value != strings.Repeat("a", 32) {
			t.Fatal("bad exact-ID context", q)
		}
	}
	if len(found["span-logs"].Query.Where.All) != 2 || found["resources-edge"].Discovery.FromMS != 1000 {
		t.Fatal("lost context")
	}
	found["duration-status"].Query.Where.All[0].Value = "changed"
	if found["reported-goo.operation.outcome"].Query.Where.All[0].Value == "changed" {
		t.Fatal("returned plans share mutable predicates")
	}
	after, _ := json.Marshal(c)
	if string(before) != string(after) {
		t.Fatal("mutated stored evidence")
	}
	if _, e = c.Followups(1); e == nil {
		t.Fatal("invented retained row")
	}
	if _, e = c.Followups(100); e == nil {
		t.Fatal("unbounded observation choice")
	}
}
func TestNoInventedTraceOrOutcomeIdentity(t *testing.T) {
	c := followupFixture()
	c.Explore.Rows = []json.RawMessage{json.RawMessage(`{"ServiceName":"edge","TraceId":"00000000000000000000000000000000","SpanId":"bad"}`)}
	p, e := c.Followups(0)
	if e != nil {
		t.Fatal(e)
	}
	for _, a := range p.Actions {
		if strings.Contains(a.ID, "logs") || a.ID == "trace" || strings.HasPrefix(a.ID, "reported-") {
			t.Fatal("invented absent identity", a)
		}
	}
}
