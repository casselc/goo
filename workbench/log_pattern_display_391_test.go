package workbench

import (
	"github.com/casselc/goo/exploration"
	"testing"
)

func patternDisplay391() (ViewDeclaration, exploration.Query) {
	v := authoredView()
	q := v.Queries[0]
	q.Signal = "logs"
	q.GroupBy = []exploration.Field{{Kind: "column", Key: "body_pattern_v1"}}
	d := v.Displays["spans"]
	d.Scope.Signal = "logs"
	d.Display.Profile = "log-pattern-count-v1"
	return d, q
}
func TestLogPatternDisplay391ProfileAndReuse(t *testing.T) {
	d, q := patternDisplay391()
	for _, kind := range []string{"bar", "table"} {
		d.Kind.Name = kind
		if err := d.Validate(q); err != nil {
			t.Fatal(err)
		}
	}
	destination := q
	destination.Environment = "stage"
	destination.FromMS = 3000
	destination.ToMS = 4000
	got, err := ReuseViewDeclaration(d, q, destination)
	if err != nil || got != d {
		t.Fatal(got, err)
	}
	for _, profile := range []string{"", "log-pattern-v2"} {
		bad := d
		bad.Display.Profile = profile
		if bad.Validate(q) == nil {
			t.Fatal("accepted profile", profile)
		}
	}
	ordinary := q
	ordinary.GroupBy = []exploration.Field{{Kind: "column", Key: "service"}}
	if d.Validate(ordinary) == nil {
		t.Fatal("pattern profile on ordinary query")
	}
	for _, value := range []string{"numeric.mean", "duration.mean"} {
		bad := d
		bad.Display.Value = value
		if bad.Validate(q) == nil {
			t.Fatal("accepted measure", value)
		}
	}
	for _, mutate := range []func(*exploration.Query){func(q *exploration.Query) { q.Summary = "span-duration" }, func(q *exploration.Query) { q.SpanStatus = true }, func(q *exploration.Query) { q.GroupTimeSeries = true }, func(q *exploration.Query) { q.Numeric = &exploration.NumericBinding{} }, func(q *exploration.Query) { q.NumericHistogram = &exploration.NumericHistogramSpec{} }, func(q *exploration.Query) { q.ReportedEdge = &exploration.ReportedEdge{} }} {
		bad := q
		mutate(&bad)
		if d.Validate(bad) == nil {
			t.Fatal("accepted non-count query")
		}
	}
}
func TestLogPatternDisplay391NotTemplateExpansion(t *testing.T) {
	d, q := patternDisplay391()
	c := *compositionView().Composition
	c.Children[0].Aggregate = &d
	c.Children[1].Records.Scope.Signal = "logs"
	c.Children[1].Records.Display.FieldColumns = nil
	if err := c.Validate(q); err != nil {
		t.Fatal("test requires otherwise valid composition", err)
	}
	b := ViewTemplateBinding{Version: "goo.template/1", Group: "destination", Profile: "ordinary"}
	if b.Validate(q, c) == nil {
		t.Fatal("pattern declaration widened template admission")
	}
}
