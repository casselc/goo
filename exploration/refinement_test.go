package exploration

import (
	"github.com/casselc/goo/identity"
	"reflect"
	"testing"
)

func refinementParent() Query {
	return Query{Signal: "logs", Environment: "test", FromMS: 1, ToMS: 60001, Limit: 30, Offset: 7, GroupBy: []Field{{Kind: "column", Key: "service"}}, Where: &Predicate{Any: []Predicate{{Field: &Field{Kind: "column", Key: "service"}, Op: "eq", Value: "a"}, {Field: &Field{Kind: "column", Key: "service"}, Op: "eq", Value: "b"}}}}
}

func TestRefinePreservesCompleteParentAndIsolation(t *testing.T) {
	parent := refinementParent()
	before, err := CloneQuery(parent)
	if err != nil {
		t.Fatal(err)
	}
	leaf := &Predicate{Field: &Field{Kind: "column", Key: "body"}, Op: "contains", Value: "timeout"}
	out, err := Refine(parent, leaf)
	if err != nil {
		t.Fatal(err)
	}
	if out.Offset != parent.Offset || out.FromMS != parent.FromMS || out.ToMS != parent.ToMS || out.Environment != parent.Environment || !reflect.DeepEqual(out.GroupBy, parent.GroupBy) {
		t.Fatal("question policy changed")
	}
	if len(out.Where.All) != 2 || !reflect.DeepEqual(out.Where.All[0], *parent.Where) {
		t.Fatal("whole parent not retained")
	}
	out.Where.All[0].Any[0].Field.Key = "body"
	out.Where.All[1].Field.Key = "severity"
	out.GroupBy[0].Key = "body"
	if !reflect.DeepEqual(parent, before) || leaf.Field.Key != "body" {
		t.Fatal("input aliased")
	}
}

func TestPredicateAdmissionAndCloneBounded(t *testing.T) {
	if ValidatePredicate(nil, "bad") == nil {
		t.Fatal("unknown signal admitted")
	}
	cycle := &Predicate{All: make([]Predicate, 1)}
	cycle.All[0].All = cycle.All
	if _, err := ClonePredicate(cycle, "logs"); err == nil {
		t.Fatal("cycle admitted")
	}
	if _, err := CloneQuery(Query{}); err == nil {
		t.Fatal("zero query admitted")
	}
	wide := &Predicate{All: make([]Predicate, 33)}
	for i := range wide.All {
		wide.All[i] = Predicate{Field: &Field{Kind: "column", Key: "service"}, Op: "eq", Value: "a"}
	}
	if _, err := ClonePredicate(wide, "logs"); err == nil {
		t.Fatal("oversize admitted")
	}
	leaf := &Predicate{Field: &Field{Kind: "column", Key: "severity_number"}, Op: "eq", Value: "01"}
	if ValidatePredicate(leaf, "logs") != nil {
		t.Fatal("ordinary numeric grammar changed")
	}
	if out, err := ClonePredicate(nil, "logs"); err != nil || out != nil {
		t.Fatal("nil clone")
	}
}

func TestRefineCombinedBudget(t *testing.T) {
	parent := refinementParent()
	leaf := Predicate{Field: &Field{Kind: "column", Key: "body"}, Op: "eq", Value: "x"}
	parent.Where = &leaf
	for i := 0; i < 4; i++ {
		parent.Where = &Predicate{All: []Predicate{*parent.Where}}
	}
	if err := parent.Validate(); err != nil {
		t.Fatal(err)
	}
	if _, err := Refine(parent, &leaf); err == nil {
		t.Fatal("combined depth admitted")
	}
	parent = refinementParent()
	parent.Where = &Predicate{All: make([]Predicate, 30)}
	for i := range parent.Where.All {
		parent.Where.All[i] = leaf
	}
	if err := parent.Validate(); err != nil {
		t.Fatal(err)
	}
	if _, err := Refine(parent, &leaf); err == nil {
		t.Fatal("combined nodes admitted")
	}
}

func TestCloneQueryNestedIsolation(t *testing.T) {
	q := refinementParent()
	q.Environment = ""
	q.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "test"}}
	q.Numeric = &NumericBinding{Field: Field{Kind: "attribute", Key: "amount"}, Type: "int64", Unit: "1", Parse: "canonical-decimal"}
	q.NumericHistogram = &NumericHistogramSpec{Binning: "powers-of-two"}
	q.Where = &Predicate{Numeric: q.Numeric, Op: "gte", Value: "1"}
	copy, err := CloneQuery(q)
	if err != nil {
		t.Fatal(err)
	}
	copy.EnvironmentSelector.Scope.Value = "changed"
	copy.Numeric.Field.Key = "changed"
	copy.NumericHistogram.Binning = "changed"
	copy.Where.Numeric.Field.Key = "changed"
	if q.EnvironmentSelector.Scope.Value != "test" || q.Numeric.Field.Key != "amount" || q.NumericHistogram.Binning != "powers-of-two" || q.Where.Numeric.Field.Key != "amount" {
		t.Fatal("nested clone aliases original")
	}
	edge := Query{Signal: "spans", Environment: "test", FromMS: 1, ToMS: 60001, Limit: 10, ReportedEdge: &ReportedEdge{Parent: EdgeService{Service: "api", Namespace: &identity.Namespace{Mode: "value", Value: "ns"}}, Child: EdgeService{Service: "worker", Namespace: &identity.Namespace{Mode: "value", Value: "ns"}}, DiscoveryNamespace: &identity.Namespace{Mode: "value", Value: "ns"}}}
	copied, err := CloneQuery(edge)
	if err != nil {
		t.Fatal(err)
	}
	copied.ReportedEdge.Parent.Namespace.Value = "changed"
	copied.ReportedEdge.Child.Namespace.Value = "changed"
	copied.ReportedEdge.DiscoveryNamespace.Value = "changed"
	if edge.ReportedEdge.Parent.Namespace.Value != "ns" || edge.ReportedEdge.Child.Namespace.Value != "ns" || edge.ReportedEdge.DiscoveryNamespace.Value != "ns" {
		t.Fatal("edge namespace alias")
	}
}

func TestRefineNilIdentityAndEmptyBranchRefusal(t *testing.T) {
	q := refinementParent()
	out, err := Refine(q, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(q, out) {
		t.Fatal("nil refinement changed question")
	}
	out.Where.Any[0].Value = "changed"
	if q.Where.Any[0].Value == "changed" {
		t.Fatal("identity clone aliases original")
	}
	for _, p := range []*Predicate{{All: []Predicate{}}, {Any: []Predicate{}}} {
		if _, err := ClonePredicate(p, "logs"); err == nil {
			t.Fatal("empty branch admitted")
		}
	}
}
