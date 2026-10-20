package exploration

import (
	"encoding/json"
	"github.com/casselc/goo/identity"
	"reflect"
	"strings"
	"testing"
)

func edgeQuery() Query {
	return Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100, Summary: "span-duration", SpanStatus: true, ReportedEdge: &ReportedEdge{Parent: EdgeService{Service: "api'", Namespace: &identity.Namespace{Mode: "value", Value: "payments"}}, Child: EdgeService{Service: "worker", Namespace: &identity.Namespace{Mode: "missing"}}}}
}
func TestReportedEdgeBoundedPlanAndCustody(t *testing.T) {
	q := edgeQuery()
	q.Where = &Predicate{Field: &Field{Kind: "attribute", Key: "http.route"}, Op: "eq", Value: "/slow"}
	p, e := Compile(q)
	if e != nil || strings.Contains(p.Count, "payments") || strings.Contains(p.Count, "api'") || strings.Count(p.Count, "http.route") > 0 || !strings.Contains(p.Count, "groupUniqArray") || !strings.Contains(p.Count, "(TraceId,ParentSpanId) IN") {
		t.Fatal(p, e)
	}
	for _, sql := range []string{p.Count, p.Rows, p.Fields} {
		if !strings.Contains(sql, "groupUniqArray") || !strings.Contains(sql, "{environment:String}") {
			t.Fatal("edge scope lost", sql)
		}
	}
	pkg, e := DraftTelemetryPackage(q, "edge", "Edge", 1)
	if e != nil || !reflect.DeepEqual(pkg.ReportedEdge, q.ReportedEdge) {
		t.Fatal(pkg, e)
	}
	planned, e := PlanTelemetryPackage(pkg, q.FromMS, q.ToMS, nil)
	serialized := ""
	if planned != nil {
		serialized = planned.YAML
	}
	if e != nil {
		t.Fatal(e)
	}
	decoded, e := ReadTelemetryPackage(serialized)
	if e != nil || !reflect.DeepEqual(decoded.ReportedEdge, q.ReportedEdge) {
		t.Fatal(e)
	}
	saved := (CaptureQuestion{Kind: "explore", Explore: &q}).Normalized()
	q.ReportedEdge.Parent.Service = "changed"
	if saved.Explore.ReportedEdge.Parent.Service != "api'" {
		t.Fatal("edge custody aliases caller")
	}
	wider := *saved.Explore
	wider.ReportedEdge = nil
	if _, e = DescribeTransform(saved, CaptureQuestion{Kind: "explore", Explore: &wider}, "same-population"); e == nil {
		t.Fatal("edge removal claimed same population")
	}
	b, _ := json.Marshal(wider)
	if strings.Contains(string(b), "reportedEdge") {
		t.Fatal("legacy nil edge changed wire bytes")
	}
}
func TestReportedEdgeRejectsUnsafeBindings(t *testing.T) {
	for _, mutate := range []func(*Query){func(q *Query) { q.Signal = "logs"; q.Summary = ""; q.SpanStatus = false }, func(q *Query) { q.ReportedEdge.Parent.Namespace = nil }, func(q *Query) { q.ReportedEdge.Child = q.ReportedEdge.Parent }, func(q *Query) { q.ReportedEdge.DiscoveryNamespace = &identity.Namespace{Mode: "missing"} }} {
		q := edgeQuery()
		mutate(&q)
		if q.Validate() == nil {
			t.Fatal("invalid edge accepted", q)
		}
	}
}
