package exploration

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/measurements"
	"gopkg.in/yaml.v3"
)

func TestScopeRoundtripPresentationAndOTelNamespaces(t *testing.T) {
	c := followupFixture()
	q := c.Question
	q.Explore.Where = &Predicate{Any: []Predicate{
		{Field: &Field{Kind: "resource", Key: "service.namespace"}, Op: "eq", Value: "one"},
		{Field: &Field{Kind: "attribute", Key: "service.namespace"}, Op: "eq", Value: "two"},
		{Field: &Field{Kind: "column", Key: "scope_version"}, Op: "eq", Value: "1.2.3"},
		{Field: &Field{Kind: "attribute", Key: "http.route"}, Op: "missing"},
	}}
	before, _ := json.Marshal(q)
	s, e := q.Scope()
	if e != nil {
		t.Fatal(e)
	}
	other := detached(q)
	other.Explore.Limit = 100
	other.Explore.Offset = 100
	other.Explore.Summary = "span-duration"
	other.Explore.GroupBy = []Field{{Kind: "column", Key: "service"}}
	other.Explore.BucketMS = 1000
	next, e := other.Scope()
	if e != nil || !reflect.DeepEqual(s, next) {
		t.Fatal("presentation changed population", e)
	}
	var round CaptureQuestion
	if e = json.Unmarshal(before, &round); e != nil {
		t.Fatal(e)
	}
	again, _ := round.Scope()
	if !reflect.DeepEqual(s, again) {
		t.Fatal("wire roundtrip changed scope")
	}
	p, e := DraftTelemetryPackage(*q.Explore, "otel", "OTel", 1)
	if e != nil {
		t.Fatal(e)
	}
	wire, _ := yaml.Marshal(p)
	pkg, e := ReadTelemetryPackage(string(wire))
	if e != nil {
		t.Fatal(e)
	}
	reused := pkg.Question(q.Explore.FromMS, q.Explore.ToMS)
	rs, e := (CaptureQuestion{Kind: "explore", Explore: &reused}).Scope()
	if e != nil || !reflect.DeepEqual(rs, s) {
		t.Fatal("package scope drift", e)
	}
	if s.Convention != "Unbound; no semantic-convention profile inferred" || !strings.Contains(s.SchemaMetadata, "Unavailable") {
		t.Fatal("inferred convention from library version")
	}
	s.Selector.Any[0].Value = "changed"
	after, _ := json.Marshal(q)
	if string(before) != string(after) {
		t.Fatal("scope aliases origin")
	}
}

func TestCheckedTransformRejectsFalseClaims(t *testing.T) {
	origin := followupFixture().Question
	cases := []struct {
		name     string
		relation string
		edit     func(*CaptureQuestion)
	}{
		{"window", "same-population", func(q *CaptureQuestion) { q.Explore.ToMS++ }},
		{"environment", "same-population", func(q *CaptureQuestion) { q.Explore.Environment = "foreign" }},
		{"selector", "same-population", func(q *CaptureQuestion) { q.Explore.Where = nil }},
		{"signal", "same-population", func(q *CaptureQuestion) { q.Explore.Signal = "logs" }},
		{"definition", "same-question", func(q *CaptureQuestion) { q.Explore.Summary = "span-duration" }},
		{"forged relationship", "causal", func(q *CaptureQuestion) {}},
		{"non ID", "exact-trace", func(q *CaptureQuestion) {}},
		{"zero ID", "exact-trace", func(q *CaptureQuestion) {
			q.Explore.Where = ptrPredicate(followupEqual("trace_id", strings.Repeat("0", 32)))
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			q := detached(origin)
			tc.edit(&q)
			if _, e := DescribeTransform(origin, q, tc.relation); e == nil {
				t.Fatal("false relationship accepted")
			}
		})
	}
	p, e := PlanRecipe("latency", *origin.Explore)
	if e != nil {
		t.Fatal(e)
	}
	if p.Transform.Relation != "same-population" || !reflect.DeepEqual(p.Transform.Origin, p.Transform.Target) {
		t.Fatal("recipe lost semantic scope")
	}
	for _, c := range p.Transform.Changes {
		if c.Field == "selector" && c.Effect != "preserved" {
			t.Fatal("bad selector diff")
		}
	}
}

func TestMetricScopePreservesDeclaredIdentityAndUnknownConvention(t *testing.T) {
	d := measurements.Definition{APIVersion: measurements.Version, Name: "custom", Title: "Custom metric", Kind: "gauge", Environment: "prod", LookbackMinutes: 1, MaxGapMS: 60000, Grain: "per-stream", Source: measurements.Binding{Services: []string{"worker"}, Scope: "custom.scope", Version: "9.9.9", SchemaURL: "https://opentelemetry.io/schemas/1.30.0", Metric: "custom.depth", Unit: "{job}", Attributes: map[string]string{"queue": "one"}}}
	mq := measurements.Question{Definition: d, FromMS: 1000, ToMS: 61000}
	q := CaptureQuestion{Kind: "measurement", Measurement: &mq}
	s, e := q.Scope()
	if e != nil {
		t.Fatal(e)
	}
	if s.Signal != "metrics" || s.MeasurementKind != "gauge" || !reflect.DeepEqual(s.Binding, &d.Source) || s.DefinitionRevision != measurements.Hash(d) || !strings.HasPrefix(s.Convention, "Unbound") {
		t.Fatal("lost binding or inferred convention")
	}
	s.Binding.Attributes["queue"] = "mutated"
	if d.Source.Attributes["queue"] != "one" {
		t.Fatal("aliased metric binding")
	}
	tr, e := DescribeTransform(q, q, "same-question")
	if e != nil || len(tr.Replaces) != 0 {
		t.Fatal("identical metric scope changed", e)
	}
	disc, e := DescribeDiscoveryTransform(q, measurements.DiscoveryQuestion{Environment: "prod", Service: "worker", FromMS: 1000, ToMS: 61000})
	if e != nil || disc.Relation != "service-time" || disc.Target.Binding.Scope != "" || disc.Target.DefinitionRevision != "" {
		t.Fatal("discovery retained false metric identity", e)
	}
	if _, e = DescribeDiscoveryTransform(q, measurements.DiscoveryQuestion{Environment: "prod", Service: "worker", FromMS: 1000, ToMS: 61001}); e == nil {
		t.Fatal("window drift accepted")
	}
}

func TestCapturePlanningDoesNotRewriteReceipt(t *testing.T) {
	c := followupFixture()
	before, _ := json.Marshal(c)
	plans, e := c.Followups(0)
	if e != nil {
		t.Fatal(e)
	}
	for _, a := range plans.Actions {
		if a.Transform == nil {
			t.Fatal("missing checked transform")
		}
		if a.Query != nil {
			a.Transform.Target.Selector.Value = "edited metadata"
		}
	}
	after, _ := json.Marshal(c)
	if string(before) != string(after) {
		t.Fatal("capture wire changed")
	}
}

func TestInventoryAndComparisonScopeAdapters(t *testing.T) {
	i := InventoryQuestion{Environment: "prod", FromMS: 1000, ToMS: 61000}
	inv, e := (CaptureQuestion{Kind: "inventory", Inventory: &i}).Scope()
	if e != nil || inv.Grain != "span-record" || inv.Selector != nil {
		t.Fatal("inventory changed population grain", e)
	}
	a := followupFixture().Question.Explore
	b := detached(*a)
	a.GroupBy = []Field{{Kind: "column", Key: "service"}}
	b.GroupBy = a.GroupBy
	b.FromMS = 61000
	b.ToMS = 121000
	pair := CaptureQuestion{Kind: "comparison", Comparison: &ComparisonQuestion{Baseline: *a, Selection: b}}
	scope, e := pair.Scope()
	if e != nil || len(scope.Members) != 2 || scope.Members[0].FromMS == scope.Members[1].FromMS {
		t.Fatal("collapsed comparison windows", e)
	}
	if _, e = DescribeTransform(pair, pair, "same-question"); e == nil {
		t.Fatal("implicit comparison transform")
	}
}
