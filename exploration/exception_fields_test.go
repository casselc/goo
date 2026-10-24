package exploration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

type exceptionFixture struct {
	names        []string
	attrs        []map[string]string
	event, state string
	typ          *string
}

func exceptionFixtures() []exceptionFixture {
	return []exceptionFixture{
		{[]string{}, []map[string]string{}, "no", "unreported", nil},
		{[]string{"other"}, []map[string]string{{"exception.type": "Ignored"}}, "no", "unreported", nil},
		{[]string{"exception"}, []map[string]string{{}}, "yes", "unreported", nil},
		{[]string{"exception"}, []map[string]string{{"exception.type": ""}}, "yes", "empty", str("")},
		{[]string{"exception"}, []map[string]string{{"exception.type": "Foo"}}, "yes", "single", str("Foo")},
		{[]string{"exception", "exception", "exception"}, []map[string]string{{"exception.type": "Foo"}, {"exception.type": "Foo"}, {"exception.type": "Foo"}}, "yes", "single", str("Foo")},
		{[]string{"exception", "exception"}, []map[string]string{{"exception.type": "Foo"}, {"exception.type": "Bar"}}, "yes", "multiple", nil},
		{[]string{"exception", "exception"}, []map[string]string{{}, {"exception.type": "Foo"}}, "yes", "single", str("Foo")},
		{[]string{"exception", "exception"}, []map[string]string{{"exception.type": ""}, {"exception.type": "Foo"}}, "yes", "multiple", nil},
		{[]string{"other", "exception"}, []map[string]string{{"exception.type": "Foo"}, {"exception.type": "Bar'雪\\"}}, "yes", "single", str("Bar'雪\\")},
	}
}
func TestExceptionReferenceReportingStatesAndNoExplosion(t *testing.T) {
	event, typeField, state := Field{Kind: "column", Key: "exception_event"}, Field{Kind: "column", Key: "exception_type"}, Field{Kind: "column", Key: "exception_type_state"}
	for _, tc := range exceptionFixtures() {
		values, e := ReferenceExceptionFields(tc.names, tc.attrs)
		if e != nil || values[event] != tc.event || values[state] != tc.state {
			t.Fatal(tc, values, e)
		}
		v, present := values[typeField]
		if present != (tc.typ != nil) || present && v != *tc.typ {
			t.Fatal("nullable type meaning changed", tc, values)
		}
		if ExceptionEventPresent(tc.names) != (tc.event == "yes") {
			t.Fatal("event presence changed")
		}
		p := Predicate{Field: &typeField, Op: "missing"}
		if Match(&p, values) != (tc.typ == nil) {
			t.Fatal("missing collapsed empty/ambiguous", tc)
		}
		p = Predicate{Field: &typeField, Op: "eq", Value: ""}
		if Match(&p, values) != (tc.typ != nil && *tc.typ == "") {
			t.Fatal("empty type lost", tc)
		}
	}
	if _, e := ReferenceExceptionFields([]string{"exception"}, nil); e == nil {
		t.Fatal("unaligned events guessed a type")
	}
	if _, e := ReferenceExceptionFields(nil, []map[string]string{{"exception.type": "Foo"}}); e == nil {
		t.Fatal("extra attribute map ignored malformed alignment")
	}
}

func TestExceptionFieldsValidationProjectionAndNullableConsumers(t *testing.T) {
	for _, key := range []string{"exception_event", "exception_type", "exception_type_state"} {
		f := Field{Kind: "column", Key: key}
		if typ, e := f.Type("spans"); e != nil || typ != "string" {
			t.Fatal(key, e)
		}
		if _, e := f.Type("logs"); e == nil {
			t.Fatal("span event column leaked to logs")
		}
	}
	severity := Field{Kind: "column", Key: "severity_number"}
	if typ, e := severity.Type("logs"); e != nil || typ != "uint64" {
		t.Fatal(e)
	}
	if _, e := severity.Type("spans"); e == nil {
		t.Fatal("log severity leaked to spans")
	}
	q := basic()
	plain, e := Compile(q)
	if e != nil {
		t.Fatal(e)
	}
	for _, key := range []string{"exception_event", "exception_type", "exception_type_state"} {
		if strings.Contains(plain.Rows, " AS "+key) {
			t.Fatal("old row projection acquired derived field")
		}
		next := q
		f := Field{Kind: "column", Key: key}
		next.Where = &Predicate{Field: &f, Op: "exists"}
		p, e := Compile(next)
		if e != nil || !strings.Contains(p.Rows, " AS "+key) {
			t.Fatal("requested exception column absent", e)
		}
		if key != "exception_event" && !strings.Contains(p.Rows, "arrayZip") {
			t.Fatal("type source no longer checks alignment")
		}
	}
	q.Signal = "logs"
	plain, e = Compile(q)
	if e != nil || strings.Contains(plain.Rows, "SeverityNumber") {
		t.Fatal("old logs row projection changed", e)
	}
	q.Where = &Predicate{All: []Predicate{{Field: &severity, Op: "gte", Value: "17"}, {Field: &severity, Op: "lte", Value: "24"}}}
	p, e := Compile(q)
	if e != nil || !strings.Contains(p.Rows, "toString(toUInt64(SeverityNumber)) AS severity_number") {
		t.Fatal("severity projection missing", e)
	}
	for _, v := range []string{"0", "16", "17", "20", "21", "24", "25", "255"} {
		want := v == "17" || v == "20" || v == "21" || v == "24"
		if Match(q.Where, map[Field]string{severity: v}) != want {
			t.Fatal("severity numeric range mismatch", v)
		}
	}
	f := Field{Kind: "column", Key: "exception_type"}
	if !f.CanBeMissing() || (Field{Kind: "column", Key: "service"}).CanBeMissing() {
		t.Fatal("nullable fixed columns too broad")
	}
	request := explanationRequest()
	request.Fields = []Field{f}
	side := explanationRead(ExplanationValueCount{Value: nil, Count: "1"}, ExplanationValueCount{Value: str(""), Count: "1"})
	result, e := BuildExplanation(request, explanationPopulation("2"), explanationPopulation("2"), []ExplanationFieldResult{{Field: f, Baseline: side, Selection: side}}, explanationTestTime)
	if e != nil || result.Fields[0].State != "complete" || result.Fields[0].Rows[0].Value != nil || *result.Fields[0].Rows[1].Value != "" {
		t.Fatal("explanation rejected nullable derived field", e)
	}
	for _, value := range []*string{nil, str("")} {
		include, e := PlanExplanationFilter(request, f, value, "baseline", "include")
		if e != nil {
			t.Fatal(e)
		}
		exclude, e := PlanExplanationFilter(request, f, value, "baseline", "exclude")
		if e != nil {
			t.Fatal(e)
		}
		for _, tc := range exceptionFixtures() {
			row, _ := ReferenceExceptionFields(tc.names, tc.attrs)
			row[request.Baseline.Where.Numeric.Field] = "9007199254740993"
			if Match(include.Where, row) == Match(exclude.Where, row) {
				t.Fatal("derived type include/exclude not complementary")
			}
		}
	}
	request.Fields = []Field{{Kind: "column", Key: "service"}}
	if _, e := BuildExplanation(request, explanationPopulation("2"), explanationPopulation("2"), []ExplanationFieldResult{{Field: request.Fields[0], Baseline: side, Selection: side}}, explanationTestTime); e == nil {
		t.Fatal("arbitrary fixed columns became nullable")
	}
	span := basic()
	span.GroupBy = []Field{{Kind: "column", Key: "service"}, f}
	span.Where = &Predicate{Field: &Field{Kind: "column", Key: "exception_event"}, Op: "eq", Value: "yes"}
	pkg, e := DraftTelemetryPackage(span, "exceptions", "Reported exception events", 1)
	if e != nil {
		t.Fatal(e)
	}
	plan, e := PlanTelemetryPackage(pkg, 0, 60000, nil)
	if e != nil {
		t.Fatal(e)
	}
	loaded, e := ReadTelemetryPackage(plan.YAML)
	if e != nil || !reflect.DeepEqual(pkg, loaded) {
		t.Fatal("derived field package lost semantics", e)
	}
	if _, e := groupValues(span, json.RawMessage(`{"g0":"svc","g1":null}`)); e != nil {
		t.Fatal("nullable grouped/captured dimension refused", e)
	}
	if _, e := groupValues(span, json.RawMessage(`{"g0":null,"g1":null}`)); e == nil {
		t.Fatal("nonnullable service accepted null")
	}
}

func TestClickHouseExceptionFieldParity(t *testing.T) {
	endpoint := os.Getenv("GOO_CLICKHOUSE_TEST_URL")
	if endpoint == "" {
		t.Skip("GOO_CLICKHOUSE_TEST_URL is not configured")
	}
	source := Source{Endpoint: endpoint, User: os.Getenv("GOO_CLICKHOUSE_TEST_USER"), Password: os.Getenv("GOO_CLICKHOUSE_TEST_PASSWORD")}
	fixtures := exceptionFixtures()
	names := [][]string{}
	attrs := [][]map[string]string{}
	for _, tc := range fixtures {
		names = append(names, tc.names)
		attrs = append(attrs, tc.attrs)
	}
	namesJSON, _ := json.Marshal(names)
	attrsJSON, _ := json.Marshal(attrs)
	inline := " FROM (SELECT number AS n, JSONExtract({names:String}, 'Array(Array(String))')[number+1] AS `Events.Name`, JSONExtract({attrs:String}, 'Array(Array(Map(String, String)))')[number+1] AS `Events.Attributes` FROM numbers(" + fmt.Sprint(len(fixtures)) + "))"
	params := map[string]string{"names": string(namesJSON), "attrs": string(attrsJSON)}
	for _, key := range []string{"exception_event", "exception_type", "exception_type_state"} {
		c := compiler{signal: "spans", params: map[string]string{}}
		value, present := c.field(Field{Kind: "column", Key: key})
		sql := "SELECT toString(n) AS n, if((" + present + "), CAST(" + value + " AS Nullable(String)), NULL) AS value" + inline + " ORDER BY n FORMAT JSONEachRow"
		rows, e := source.read(context.Background(), sql, params)
		if e != nil || len(rows) != len(fixtures) {
			t.Fatal(key, e, len(rows))
		}
		for i, row := range rows {
			var v struct {
				N     string
				Value *string
			}
			if e := json.Unmarshal(row, &v); e != nil {
				t.Fatal(e)
			}
			want, _ := ReferenceExceptionFields(fixtures[i].names, fixtures[i].attrs)
			expected, exists := want[Field{Kind: "column", Key: key}]
			if (v.Value != nil) != exists || v.Value != nil && *v.Value != expected {
				t.Fatal("ClickHouse exception projection differs from reference", key, i, string(row), want)
			}
		}
	}
	c := compiler{signal: "spans", params: map[string]string{}}
	event := &Predicate{Field: &Field{Kind: "column", Key: "exception_event"}, Op: "eq", Value: "yes"}
	where := c.predicate(event)
	for k, v := range c.params {
		params[k] = v
	}
	typeValue, typePresent := c.field(Field{Kind: "column", Key: "exception_type"})
	rows, e := source.read(context.Background(), "SELECT if(("+typePresent+"), CAST("+typeValue+" AS Nullable(String)), NULL) AS value, toString(count()) AS count"+inline+" WHERE "+where+" GROUP BY value FORMAT JSONEachRow", params)
	if e != nil {
		t.Fatal(e)
	}
	sum := 0
	for _, row := range rows {
		var v struct {
			Value *string
			Count string
		}
		json.Unmarshal(row, &v)
		var n int
		fmt.Sscan(v.Count, &n)
		sum += n
		if v.Value != nil && *v.Value == "Foo" && n != 3 {
			t.Fatal("events exploded or first type guessed", string(row))
		}
	}
	if sum != 8 {
		t.Fatal("grouping changed record grain", sum)
	}
	// Type/state fail closed if physical source arrays do not align. Presence by
	// names alone remains independently meaningful and does not guess a type.
	for _, key := range []string{"exception_type", "exception_type_state"} {
		value, _, _ := exceptionFieldSQL(key)
		_, e := source.read(context.Background(), "SELECT "+value+" AS value FROM (SELECT ['exception'] AS `Events.Name`, CAST([], 'Array(Map(String, String))') AS `Events.Attributes`) FORMAT JSONEachRow", nil)
		if e == nil {
			t.Fatal("malformed array alignment accepted", key)
		}
	}
	severity := Field{Kind: "column", Key: "severity_number"}
	c = compiler{signal: "logs", params: map[string]string{}}
	expr := c.predicate(&Predicate{All: []Predicate{{Field: &severity, Op: "gte", Value: "17"}, {Field: &severity, Op: "lte", Value: "24"}}})
	rows, e = source.read(context.Background(), "SELECT toString(SeverityNumber) AS value FROM (SELECT arrayJoin([toUInt8(0),16,17,20,21,24,25,255]) AS SeverityNumber, 'noncanonical text' AS SeverityText) WHERE "+expr+" ORDER BY SeverityNumber FORMAT JSONEachRow", c.params)
	if e != nil || len(rows) != 4 {
		t.Fatal("numeric severity interpretation failed", len(rows), e)
	}
}
