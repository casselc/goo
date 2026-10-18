package exploration

import (
	"encoding/json"
	"strings"
	"testing"
)

func basic() Query {
	return Query{Signal: "spans", Environment: "local", FromMS: 0, ToMS: 60000, Limit: 2}
}
func TestPresenceAndTypedPredicates(t *testing.T) {
	f := Field{"attribute", "tenant"}
	p := &Predicate{Field: &f, Op: "eq", Value: ""}
	if Match(p, map[Field]string{}) || !Match(p, map[Field]string{f: ""}) {
		t.Fatal("missing collapsed into empty")
	}
	p.Op = "ne"
	p.Value = "a"
	if Match(p, map[Field]string{}) {
		t.Fatal("missing satisfies inequality")
	}
	d := Field{"column", "duration_ns"}
	p = &Predicate{Field: &d, Op: "gte", Value: "18446744073709551614"}
	if !Match(p, map[Field]string{d: "18446744073709551615"}) || Match(p, map[Field]string{d: "9007199254740992"}) {
		t.Fatal("lost uint64 precision")
	}
	p.Op = "eq"
	p.Value = "01"
	if !Match(p, map[Field]string{d: "1"}) {
		t.Fatal("numeric equality became string equality")
	}
}
func TestSafeCompilerAndPageIndependentCounts(t *testing.T) {
	q := basic()
	field := Field{"attribute", "x'] OR 1=1 --"}
	q.Where = &Predicate{Field: &field, Op: "eq", Value: "'; DROP TABLE otel_traces --"}
	q.GroupBy = []Field{field}
	p, e := Compile(q)
	if e != nil {
		t.Fatal(e)
	}
	b, _ := json.Marshal(p.Parameters)
	if !strings.Contains(string(b), "DROP TABLE") || strings.Contains(p.Rows, "DROP TABLE") || strings.Contains(p.Rows, "x']") {
		t.Fatal("unsafe interpolation")
	}
	if !strings.Contains(p.Rows, "mapContains") || !strings.Contains(p.Groups, "Nullable(String)") {
		t.Fatal("missing key semantics lost")
	}
	q.Offset = 100
	q.Limit = 100
	p2, e := Compile(q)
	if e != nil || p.Count != p2.Count || p.Groups != p2.Groups || p.Fields != p2.Fields {
		t.Fatal("page changes totals/discovery", e)
	}
	q.Where = &Predicate{All: []Predicate{}}
	if q.Validate() == nil {
		t.Fatal("empty branch accepted")
	}
	q = basic()
	q.Where = &Predicate{Field: &Field{"column", "body"}, Op: "contains", Value: "a"}
	if q.Validate() == nil {
		t.Fatal("wrong signal column accepted")
	}
}
func TestTraceContextDropsSearchRestrictions(t *testing.T) {
	q, e := TraceQuestion("prod", 10, 60010, "0123456789abcdef0123456789abcdef", "", "spans", 0)
	if e != nil {
		t.Fatal(e)
	}
	p, e := Compile(q)
	if e != nil {
		t.Fatal(e)
	}
	if len(q.Where.All) != 1 || q.Where.All[0].Field.Key != "trace_id" || !strings.Contains(p.Rows, "environment:String") || strings.Contains(p.Rows, "ServiceName =") {
		t.Fatal("expansion inherited search restrictions")
	}
	if _, e = TraceQuestion("prod", 10, 60010, strings.Repeat("0", 32), "", "logs", 0); e == nil {
		t.Fatal("zero trace accepted")
	}
}
