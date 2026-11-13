package exploration

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestRecordQueryJSON417LiteralCanonicalRoundTrip(t *testing.T) {
	const input = `{"signal":"logs","environment":"production","fromMs":0,"toMs":60000,"limit":100,"offset":0,"spanStatus":false,"groupTimeSeries":false,"bucketMs":0,"summary":"","where":{"all":[{"field":{"kind":"attribute","key":"雪😀"},"op":"eq","value":"literal � 😀 雪"},{"numeric":{"field":{"kind":"attribute","key":"count"},"type":"uint64","unit":"1","parse":"canonical-decimal","scale":0},"op":"eq","value":"0"}]},"groupBy":[]}`
	var q Query
	if err := json.Unmarshal([]byte(input), &q); err != nil {
		t.Fatal(err)
	}
	if err := q.Validate(); err != nil {
		t.Fatal(err)
	}
	if q.FromMS != 0 || q.Offset != 0 || q.SpanStatus || q.GroupTimeSeries || q.BucketMS != 0 || q.Summary != "" || q.Where.All[1].Numeric.Scale != 0 {
		t.Fatal("explicit zero values changed", q)
	}
	encoded, err := json.Marshal(q)
	if err != nil {
		t.Fatal(err)
	}
	var restored Query
	if err := json.Unmarshal(encoded, &restored); err != nil {
		t.Fatal(err)
	}
	// Empty grouping is omitted by the canonical wire representation. It does
	// not alter the selected population or turn an empty value into absence.
	q.GroupBy = nil
	if !reflect.DeepEqual(q, restored) {
		t.Fatal("canonical round trip changed record intent", string(encoded))
	}
	if restored.Where.All[0].Value != "literal � 😀 雪" {
		t.Fatal("literal replacement character or Unicode was repaired")
	}
}

func TestRecordQueryJSON417MissingDefaultsAndDuplicateReplacement(t *testing.T) {
	var q Query
	if err := json.Unmarshal([]byte(`{"signal":"logs","environment":"production","fromMs":0,"toMs":60000,"limit":100}`), &q); err != nil || q.Validate() != nil || q.Offset != 0 {
		t.Fatal("omitted zero defaults refused", q, err)
	}
	if err := json.Unmarshal([]byte(`{}`), &q); err != nil || !reflect.DeepEqual(q, Query{}) {
		t.Fatal("decoding must leave semantic requirements to Validate", q, err)
	}
	if q.Validate() == nil {
		t.Fatal("absent semantic requirements became a runnable query")
	}
	const input = `{"signal":"logs","environment":"production","fromMs":0,"toMs":60000,"limit":100,"where":{"all":[{"field":{"kind":"column","key":"service"},"op":"eq","value":"discarded"}]},"where":{"field":{"kind":"column","key":"service"},"op":"eq","value":"last"}}`
	if err := json.Unmarshal([]byte(input), &q); err != nil || q.Validate() != nil {
		t.Fatal("last-value record intent refused", err)
	}
	if len(q.Where.All) != 0 || q.Where.Field == nil || q.Where.Value != "last" {
		t.Fatal("duplicate members merged rather than replaced", q.Where)
	}
}

func TestRecordQueryJSON417RejectsLossyOrUnknownWireIntent(t *testing.T) {
	const base = `{"signal":"logs","environment":"production","fromMs":0,"toMs":60000,"limit":100}`
	appendField := func(field string) string { return strings.TrimSuffix(base, "}") + "," + field + "}" }
	cases := map[string]string{
		"missing parent service":   appendField(`"reportedEdge":{"parent":{"namespace":{"mode":"missing"}},"child":{"service":"b","namespace":{"mode":"missing"}}}`),
		"missing child service":    appendField(`"reportedEdge":{"parent":{"service":"a","namespace":{"mode":"missing"}},"child":{"namespace":{"mode":"missing"}}}`),
		"top null":                 `null`,
		"top array":                `[]`,
		"trailing object":          base + `{}`,
		"trailing garbage":         base + `x`,
		"unknown query":            appendField(`"querySecret":"not supported"`),
		"case-alias query":         appendField(`"Signal":"spans"`),
		"case-alias predicate":     appendField(`"where":{"Field":{"kind":"column","key":"service"},"op":"eq","value":"api"}`),
		"branch empty op":          appendField(`"where":{"all":[{"field":{"kind":"column","key":"service"},"op":"eq","value":"api"}],"op":""}`),
		"branch empty value":       appendField(`"where":{"any":[{"field":{"kind":"column","key":"service"},"op":"eq","value":"api"}],"value":""}`),
		"branch empty field":       appendField(`"where":{"all":[],"field":{}}`),
		"both empty branches":      appendField(`"where":{"all":[],"any":[]}`),
		"both leaf arms":           appendField(`"where":{"field":{},"numeric":{},"op":"valid","value":""}`),
		"nested branch empty op":   appendField(`"where":{"all":[{"any":[{"field":{"kind":"column","key":"service"},"op":"exists"}],"op":""}]}`),
		"unknown leaf":             appendField(`"where":{"field":{"kind":"column","key":"service"},"op":"eq","value":"api","extra":true}`),
		"unknown field":            appendField(`"groupBy":[{"kind":"column","key":"service","extra":true}]`),
		"unknown numeric":          appendField(`"numeric":{"field":{"kind":"attribute","key":"count"},"type":"uint64","unit":"1","parse":"canonical-decimal","extra":true}`),
		"unknown histogram":        appendField(`"numericHistogram":{"binning":"powers-of-two","extra":true}`),
		"unknown edge":             appendField(`"reportedEdge":{"parent":{"service":"a","namespace":{"mode":"missing"}},"child":{"service":"b","namespace":{"mode":"missing"}},"extra":true}`),
		"case-alias namespace":     appendField(`"reportedEdge":{"parent":{"service":"a","namespace":{"Mode":"missing"}},"child":{"service":"b","namespace":{"mode":"missing"}}}`),
		"null where":               appendField(`"where":null`),
		"null groupBy":             appendField(`"groupBy":null`),
		"null group field":         appendField(`"groupBy":[null]`),
		"null predicate child":     appendField(`"where":{"all":[null]}`),
		"null numeric":             appendField(`"numeric":null`),
		"null numeric field":       appendField(`"numeric":{"field":null}`),
		"null numeric scale":       appendField(`"numeric":{"scale":null}`),
		"null histogram":           appendField(`"numericHistogram":null`),
		"null edge":                appendField(`"reportedEdge":null`),
		"null selector":            appendField(`"environmentSelector":null`),
		"null namespace":           appendField(`"reportedEdge":{"parent":{"service":"a","namespace":null},"child":{"service":"b","namespace":{"mode":"missing"}}}`),
		"null scalar":              appendField(`"offset":null`),
		"null boolean":             appendField(`"spanStatus":null`),
		"null predicate value":     appendField(`"where":{"field":{"kind":"column","key":"service"},"op":"eq","value":null}`),
		"wrong scalar type":        appendField(`"offset":"0"`),
		"unpaired key surrogate":   appendField(`"where":{"field":{"kind":"attribute","key":"\ud800"},"op":"eq","value":"v"}`),
		"unpaired value surrogate": appendField(`"where":{"field":{"kind":"attribute","key":"key"},"op":"eq","value":"\udfff"}`),
		"malformed raw UTF8":       appendField("\"where\":{\"field\":{\"kind\":\"attribute\",\"key\":\"key\"},\"op\":\"eq\",\"value\":\"\xff\"}"),
	}
	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			q := Query{Signal: "logs", Environment: "stale", FromMS: 1, ToMS: 2, Limit: 100}
			if err := q.UnmarshalJSON([]byte(input)); err == nil {
				t.Fatal("unsupported wire intent admitted", input)
			}
			if !reflect.DeepEqual(q, Query{}) {
				t.Fatal("failed decode retained prior query authority", q)
			}
			var wire Query
			if err := json.Unmarshal([]byte(input), &wire); err == nil {
				t.Fatal("encoding/json bypassed strict query admission", input)
			}
		})
	}
}
