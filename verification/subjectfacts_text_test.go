package verification

import (
	"github.com/casselc/goo/subjectfacts"
	"hegel.dev/go/hegel"
	"reflect"
	"testing"
	"unicode/utf8"
)

func textPropertyDefinition() *subjectfacts.Definition {
	return &subjectfacts.Definition{APIVersion: "goo/v1alpha1", Name: "text", Source: subjectfacts.Source{Kind: "subject-facts", Fields: map[string]string{"evaluation_id": "string", "id": "string", "label": "string", "include": "bool"}, SubjectKey: []string{"id"}, Bindings: map[string]string{"clickhouse": "facts", "databricks": "catalog.schema.facts"}}, Dimensions: map[string]string{"label": "label"}, Measures: map[string]subjectfacts.Measure{"rows": {Op: "count"}}, Population: &subjectfacts.Predicate{Field: "include", Eq: true}}
}
func TestHegelPreparedTextIntegrity(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		raw := make([]byte, hegel.Draw(tc, hegel.Integers(1, 8)))
		for i := range raw {
			raw[i] = byte(hegel.Draw(tc, hegel.Integers(0, 255)))
		}
		value := string(raw)
		if hegel.Draw(tc, hegel.Booleans()) {
			value = hegel.Draw(tc, hegel.SampledFrom([]string{"", "雪", "λ", "\ufffd", "e\u0301", "é", "\x00"}))
		}
		valid := utf8.ValidString(value)
		field := hegel.Draw(tc, hegel.SampledFrom([]string{"id", "label", "evaluation_id"}))
		include := hegel.Draw(tc, hegel.Booleans())
		d := textPropertyDefinition()
		fact := subjectfacts.Fact{"evaluation_id": "selected", "id": "a", "label": "base", "include": include}
		fact[field] = value
		evaluation := "selected"
		if field == "evaluation_id" {
			evaluation = value
		}
		rows, e := d.Evaluate([]subjectfacts.Fact{fact}, evaluation, []string{"label"})
		accepted := valid && (field == "label" || value != "")
		if (e == nil) != accepted {
			tc.Fatalf("field %s text %x included %v: rows %v error %v", field, []byte(value), include, rows, e)
		}
		if accepted && include && (len(rows) != 1 || rows[0].Dimensions["label"] != fact["label"] || *rows[0].Measures["rows"].Count != 1) {
			tc.Fatal("valid text changed", rows)
		}
		base := subjectfacts.Fact{"evaluation_id": "selected", "id": "a", "label": "base", "include": true}
		want, e := d.Evaluate([]subjectfacts.Fact{base}, "selected", []string{"label"})
		if e != nil {
			tc.Fatal(e)
		}
		other := subjectfacts.Fact{"evaluation_id": "other", "id": value, "label": value}
		got, e := d.Evaluate([]subjectfacts.Fact{base, other}, "selected", []string{"label"})
		if e != nil || !reflect.DeepEqual(got, want) {
			tc.Fatal("foreign evaluation inspected", got, e)
		}
		for _, mode := range []string{"population", "measure", "enum"} {
			authored := textPropertyDefinition()
			switch mode {
			case "population":
				authored.Population = &subjectfacts.Predicate{Field: "label", Eq: value}
			case "measure":
				authored.Measures["rows"] = subjectfacts.Measure{Op: "count", Where: &subjectfacts.Predicate{Field: "label", Eq: value}}
			case "enum":
				authored.Source.Enums = map[string][]string{"label": {value}}
			}
			if (authored.Validate() == nil) != valid {
				tc.Fatal("authored validation mismatch", mode)
			}
			for _, target := range []string{"clickhouse", "databricks"} {
				_, e := authored.Compile(target, []string{"label"})
				if (e == nil) != valid {
					tc.Fatal("compiler text mismatch", mode, target, e)
				}
			}
			if _, e := authored.Evaluate(nil, "selected", nil); (e == nil) != valid {
				tc.Fatal("evaluator authoring mismatch", mode, e)
			}
		}
	}, options(200)...)
}
