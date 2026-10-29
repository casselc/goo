package subjectfacts

import (
	"reflect"
	"testing"
)

func textDefinition() *Definition {
	return &Definition{APIVersion: "goo/v1alpha1", Name: "text", Source: Source{Kind: "subject-facts", Fields: map[string]string{"evaluation_id": "string", "id": "string", "label": "string", "include": "bool"}, SubjectKey: []string{"id"}, Bindings: map[string]string{"clickhouse": "facts", "databricks": "catalog.schema.facts"}}, Dimensions: map[string]string{"label": "label"}, Measures: map[string]Measure{"rows": {Op: "count"}}}
}
func textFact(id, label string) Fact {
	return Fact{"evaluation_id": "selected", "id": id, "label": label, "include": true}
}

func TestMalformedSelectedTextRejectedBeforeIdentityAndPopulation(t *testing.T) {
	for _, field := range []string{"evaluation_id", "id", "label"} {
		for _, bad := range []string{"\xff", "\xfe", "\xc0\xaf", "\xed\xa0\x80", "\xe2\x82"} {
			for _, included := range []bool{true, false} {
				d := textDefinition()
				d.Population = &Predicate{Field: "include", Eq: true}
				f := textFact("a", "valid")
				f[field] = bad
				f["include"] = included
				if r, e := d.Evaluate([]Fact{f}, "selected", []string{"label"}); e == nil || r != nil {
					t.Fatalf("accepted malformed %s (included %v)", field, included)
				}
			}
		}
	}
	if _, e := textDefinition().Evaluate(nil, "\xff", nil); e == nil {
		t.Fatal("invalid execution ID accepted")
	}

}

func TestDistinctMalformedLabelsCannotMerge(t *testing.T) {
	// Original public-Go bug: distinct invalid bytes collapsed to one JSON group.
	if r, e := textDefinition().Evaluate([]Fact{textFact("a", "\xff"), textFact("b", "\xfe")}, "selected", []string{"label"}); e == nil || r != nil {
		t.Fatal("invalid groups merged", r, e)
	}
}

func TestValidTextAndOtherEvaluationVisibility(t *testing.T) {
	d := textDefinition()
	facts := []Fact{textFact("a", "\ufffd"), textFact("b", "λ"), textFact("c", ""), textFact("d", "e\u0301"), textFact("e", "é")}
	before := []Fact{}
	for _, f := range facts {
		copy := Fact{}
		for k, v := range f {
			copy[k] = v
		}
		before = append(before, copy)
	}
	got, e := d.Evaluate(facts, "selected", []string{"label"})
	if e != nil || len(got) != 5 {
		t.Fatal(got, e)
	}
	for _, r := range got {
		if *r.Measures["rows"].Count != 1 {
			t.Fatal("text normalized or merged", r)
		}
	}
	if !reflect.DeepEqual(before, facts) {
		t.Fatal("evaluation mutated input")
	}
	// Once a valid different evaluation ID establishes exclusion, historical fields
	// need not satisfy this definition: do not inspect their values or shape.
	other := Fact{"evaluation_id": "other", "id": "\xff", "label": "\xfe"}
	withOther, e := d.Evaluate(append(facts, other), "selected", []string{"label"})
	if e != nil || !reflect.DeepEqual(got, withOther) {
		t.Fatal("unrelated evaluation leaked", withOther, e)
	}
	// Evaluation routing itself must always be trustworthy.
	other["evaluation_id"] = "\xff"
	if _, e = d.Evaluate([]Fact{other}, "selected", nil); e == nil {
		t.Fatal("invalid routing ID ignored")
	}
}

func TestAuthoredTextValidationAndCompilerBoundary(t *testing.T) {
	for _, where := range []string{"population", "measure", "enum"} {
		d := textDefinition()
		switch where {
		case "population":
			d.Population = &Predicate{Field: "label", Eq: "\xff"}
		case "measure":
			d.Measures["rows"] = Measure{Op: "count", Where: &Predicate{Field: "label", Eq: "\xfe"}}
		case "enum":
			d.Source.Enums = map[string][]string{"label": {"\xff", "\xfe"}}
		}
		if e := d.Validate(); e == nil {
			t.Fatal("malformed authoring accepted", where)
		}
		for _, target := range []string{"clickhouse", "databricks"} {
			if _, e := d.Compile(target, []string{"label"}); e == nil {
				t.Fatal("compiler accepted malformed authoring", where, target)
			}
		}
		if _, e := d.Evaluate(nil, "selected", nil); e == nil {
			t.Fatal("evaluator accepted malformed authoring", where)
		}
	}
	for _, text := range []string{"", "\ufffd", "雪", "e\u0301", "é"} {
		d := textDefinition()
		d.Source.Enums = map[string][]string{"label": {text}}
		d.Population = &Predicate{Field: "label", Eq: text}
		for _, target := range []string{"clickhouse", "databricks"} {
			if _, e := d.Compile(target, []string{"label"}); e != nil {
				t.Fatal(text, target, e)
			}
		}
		r, e := d.Evaluate([]Fact{textFact("a", text)}, "selected", []string{"label"})
		if e != nil || len(r) != 1 || r[0].Dimensions["label"] != text {
			t.Fatal(text, r, e)
		}
	}
}
