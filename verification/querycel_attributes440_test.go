package verification

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/querycel"
	"hegel.dev/go/hegel"
)

func attributeExpression440(scope, key, op, value string) string {
	switch op {
	case "exists":
		return celLiteral(key) + " in " + scope
	case "missing":
		return "!(" + celLiteral(key) + " in " + scope + ")"
	default:
		return scope + "." + op + "(" + celLiteral(key) + ", " + celLiteral(value) + ")"
	}
}

func attributeMeaning440(op, actual, expected string, present bool) bool {
	switch op {
	case "exists":
		return present
	case "missing":
		return !present
	case "eq":
		return present && actual == expected
	case "ne":
		return present && actual != expected
	case "contains":
		return present && strings.Contains(actual, expected)
	default:
		panic("test operation is not defined")
	}
}

// The expected verdict is defined directly from key presence and string facts,
// independently of both the CEL helper and the existing predicate interpreter.
func TestHegelCELAttributes440Parity(t *testing.T) {
	keys := []string{"http.request.method", "deployment.environment.name", "quote\"'\\_雪", "scope.version"}
	values := []string{"", "0", "00", "7", "7.0", "false", "雪Λ", "before\x00after", "_%'[.*]\\"}
	hegel.Test(t, func(tc *hegel.T) {
		signal := hegel.Draw(tc, hegel.SampledFrom([]string{"logs", "spans"}))
		kind := hegel.Draw(tc, hegel.SampledFrom([]string{"resource", "attribute"}))
		scope := "resource.attributes"
		if kind == "attribute" {
			scope = "record.attributes"
		}
		key := hegel.Draw(tc, hegel.SampledFrom(keys)) // AST keys are nonempty.
		actual := hegel.Draw(tc, hegel.SampledFrom(values))
		expected := hegel.Draw(tc, hegel.SampledFrom([]string{"", "0", "00", "7", "7.0", "false", "雪Λ", "_%'[.*]\\"}))
		present := hegel.Draw(tc, hegel.Booleans())
		op := hegel.Draw(tc, hegel.SampledFrom([]string{"exists", "missing", "eq", "ne", "contains"}))
		expression := attributeExpression440(scope, key, op, expected)
		program, err := querycel.Compile(signal, expression)
		if err != nil {
			tc.Fatal(expression, err)
		}
		record := map[exploration.Field]string{}
		if present {
			record[exploration.Field{Kind: kind, Key: key}] = actual
		}
		// Same-key facts in the other namespace must never satisfy the selected map.
		other := "attribute"
		if kind == "attribute" {
			other = "resource"
		}
		record[exploration.Field{Kind: other, Key: key}] = expected
		got, err := program.Eval(record)
		if err != nil {
			tc.Fatal(expression, err)
		}
		want := attributeMeaning440(op, actual, expected, present)
		if got != want || exploration.Match(program.Predicate(), record) != want {
			tc.Fatal("presence/value disagreement", expression, present, actual, expected, got, want)
		}
		printed, err := querycel.PrintPredicate(signal, program.Predicate())
		if err != nil {
			tc.Fatal(err)
		}
		round, err := querycel.Compile(signal, printed)
		if err != nil || !reflect.DeepEqual(program.Predicate(), round.Predicate()) {
			tc.Fatal("exact inverse lowering changed a leaf", printed, err)
		}
		parent := exploration.Query{Signal: signal, Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 7, Where: &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "api"}}
		before, _ := json.Marshal(parent)
		refined, err := program.Apply(parent)
		if err != nil {
			tc.Fatal(err)
		}
		if refined.Environment != parent.Environment || refined.FromMS != parent.FromMS || refined.ToMS != parent.ToMS || refined.Limit != parent.Limit || !reflect.DeepEqual(refined.Where.All[0], *parent.Where) {
			tc.Fatal("attribute helper changed enclosing source question")
		}
		refined.Where.All[0].Value = "changed"
		after, _ := json.Marshal(parent)
		if string(before) != string(after) {
			tc.Fatal("refinement aliases its parent")
		}
	}, options(180)...)
}

// Exhaustive partitions prevent generation from accidentally missing the empty
// inequality case or treating numeric-looking exporter strings as SDK numbers.
func TestCELAttributes440PresencePartitions(t *testing.T) {
	for _, signal := range []string{"logs", "spans"} {
		for _, kind := range []string{"resource", "attribute"} {
			scope := "resource.attributes"
			if kind == "attribute" {
				scope = "record.attributes"
			}
			for _, op := range []string{"exists", "missing", "eq", "ne", "contains"} {
				program, err := querycel.Compile(signal, attributeExpression440(scope, "http.response.status_code", op, ""))
				if err != nil {
					t.Fatal(err)
				}
				for _, value := range []string{"", "0", "00", "7", "7.0", "false"} {
					for _, present := range []bool{false, true} {
						record := map[exploration.Field]string{}
						if present {
							record[exploration.Field{Kind: kind, Key: "http.response.status_code"}] = value
						}
						got, err := program.Eval(record)
						want := attributeMeaning440(op, value, "", present)
						if err != nil || got != want || exploration.Match(program.Predicate(), record) != want {
							t.Fatalf("%s %s %s present=%v value=%q got=%v want=%v err=%v", signal, kind, op, present, value, got, want, err)
						}
					}
				}
			}
		}
	}
	for _, expression := range []string{
		`resource.attributes["k"] == ""`, `resource.attributes[record.service] != "x"`,
		`resource.attributes.eq(record.service, "x")`, `resource.attributes.eq("k", record.service)`,
		`!resource.attributes.eq("k", "x")`, `record.attributes.eq("k", 7)`,
		`int(record.attributes["k"]) == 7`, `resource.attributes["nested"]["k"] == "x"`,
	} {
		if _, err := querycel.Compile("logs", expression); err == nil {
			t.Errorf("unsupported expression admitted: %s", expression)
		}
	}
}

// Only the physical relation is replaced with a read-only generated fixture.
// Values and keys still cross CEL, the ordinary predicate compiler, ClickHouse
// map-presence lowering, and typed parameter transport without SQL substitution.
func TestCELAttributes440ClickHouseParity(t *testing.T) {
	endpoint := os.Getenv("GOO_CEL_CLICKHOUSE_URL")
	if endpoint == "" {
		t.Skip("requires GOO_CEL_CLICKHOUSE_URL for live attribute parity")
	}
	const key = "http.response.status_code"
	fixtureValues := []string{"", "0", "00", "7", "7.0", "雪_%'\\"}
	sqlValues := make([]string, len(fixtureValues))
	for i, v := range fixtureValues {
		sqlValues[i] = "unhex('" + hex.EncodeToString([]byte(v)) + "')"
	}
	for _, signal := range []string{"logs", "spans"} {
		for _, kind := range []string{"resource", "attribute"} {
			scope := "resource.attributes"
			column := "ResourceAttributes"
			if kind == "attribute" {
				scope = "record.attributes"
				column = "LogAttributes"
				if signal == "spans" {
					column = "SpanAttributes"
				}
			}
			reported := "if(number=0,map(),map('" + key + "',arrayElement([" + strings.Join(sqlValues, ",") + "],number)))"
			resource := "map('deployment.environment.name','cel-attrs440')"
			attribute := "map()"
			if column == "ResourceAttributes" {
				resource = "mapConcat(" + resource + "," + reported + ")"
			} else {
				attribute = reported
			}
			attrColumn := "LogAttributes"
			table := "otel_logs"
			if signal == "spans" {
				attrColumn = "SpanAttributes"
				table = "otel_traces"
			}
			relation := "(SELECT fromUnixTimestamp64Milli(1000) AS Timestamp," + resource + " AS ResourceAttributes," + attribute + " AS " + attrColumn + " FROM numbers(7))"
			for _, tc := range []struct{ op, value string }{{"exists", ""}, {"missing", ""}, {"eq", ""}, {"ne", ""}, {"eq", "0"}, {"ne", "7"}, {"contains", "7"}, {"contains", ""}} {
				expression := attributeExpression440(scope, key, tc.op, tc.value)
				t.Run(signal+"/"+kind+"/"+tc.op+"/"+tc.value, func(t *testing.T) {
					program, err := querycel.Compile(signal, expression)
					if err != nil {
						t.Fatal(err)
					}
					q, err := program.Apply(exploration.Query{Signal: signal, Environment: "cel-attrs440", FromMS: 1000, ToMS: 61000, Limit: 1})
					if err != nil {
						t.Fatal(err)
					}
					plan, err := exploration.Compile(q)
					if err != nil {
						t.Fatal(err)
					}
					boundary := " FROM " + table + " WHERE "
					if strings.Count(plan.Count, boundary) != 1 {
						t.Fatal("physical source boundary changed")
					}
					rows, err := reportedCH353(t.Context(), endpoint, strings.Replace(plan.Count, boundary, " FROM "+relation+" WHERE ", 1), plan.Parameters)
					if err != nil || len(rows) != 1 {
						t.Fatal("live attribute count failed", err, len(rows))
					}
					var count string
					if err = json.Unmarshal(rows[0]["count"], &count); err != nil {
						t.Fatal(err)
					}
					want := 0
					for i := 0; i < 7; i++ {
						record := map[exploration.Field]string{}
						actual := ""
						if i > 0 {
							actual = fixtureValues[i-1]
							record[exploration.Field{Kind: kind, Key: key}] = actual
						}
						reference := attributeMeaning440(tc.op, actual, tc.value, i > 0)
						cel, err := program.Eval(record)
						if err != nil || cel != reference || exploration.Match(program.Predicate(), record) != reference {
							t.Fatal("reference disagreement", i, err)
						}
						if reference {
							want++
						}
					}
					if count != strconv.Itoa(want) {
						t.Fatal(fmt.Sprintf("ClickHouse count=%s CEL/AST/independent=%d", count, want))
					}
				})
			}
		}
	}
}
