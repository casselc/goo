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

func celLiteral(value string) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

func celRecord(number uint64, body, service string) map[exploration.Field]string {
	return map[exploration.Field]string{
		{Kind: "column", Key: "body"}:            body,
		{Kind: "column", Key: "severity_number"}: strconv.FormatUint(number, 10),
		{Kind: "column", Key: "severity"}:        "FATAL",
		{Kind: "column", Key: "service"}:         service,
		{Kind: "column", Key: "trace_id"}:        strings.Repeat("a", 32),
		{Kind: "column", Key: "span_id"}:         strings.Repeat("b", 16),
		{Kind: "column", Key: "scope"}:           "ordinary.cel",
		{Kind: "column", Key: "scope_version"}:   "",
		{Kind: "column", Key: "timestamp_ns"}:    "1000000000",
	}
}

func TestHegelCELPredicateParity(t *testing.T) {
	terms := []string{"timeout", " Error ", "雪Λ", "_%'[.*]\\", "line\nbreak", "", "SELECT 'value'"}
	hegel.Test(t, func(tc *hegel.T) {
		term := hegel.Draw(tc, hegel.SampledFrom(terms))
		number := uint64(hegel.Draw(tc, hegel.Integers(0, 255)))
		if hegel.Draw(tc, hegel.Booleans()) {
			number = hegel.Draw(tc, hegel.SampledFrom([]uint64{0, 1, 24, 25, 255, 9007199254740993, ^uint64(0)}))
		}
		body := hegel.Draw(tc, hegel.SampledFrom([]string{term, "before" + term + "after", "before\x00" + term + "after", strings.ToLower(term), "unrelated"}))
		service := hegel.Draw(tc, hegel.SampledFrom([]string{"api", "worker", "other"}))
		expressions := []string{
			"log.body.contains(" + celLiteral(term) + ")",
			"log.body != " + celLiteral(term),
			"log.severity_number >= 17u && log.severity_number <= 24u",
			"log.severity_number == " + strconv.FormatUint(number, 10) + "u",
			"(otel.severityBand(log.severity_number) == \"error\" && log.body.contains(" + celLiteral(term) + ")) || record.service == \"worker\"",
			"otel.severityBand(log.severity_number) == " + celLiteral(exploration.LogSeverityClass(number)),
		}
		expression := hegel.Draw(tc, hegel.SampledFrom(expressions))
		program, err := querycel.Compile("logs", expression)
		if err != nil {
			tc.Fatal(err)
		}
		record := celRecord(number, body, service)
		got, err := program.Eval(record)
		if err != nil {
			tc.Fatal(err)
		}
		if want := exploration.Match(program.Predicate(), record); got != want {
			tc.Fatal("CEL and lowered predicate disagree", expression, number, got, want)
		}
	}, options(220)...)
}

func TestCELAllPhysicalSeverityCodes(t *testing.T) {
	for _, band := range []string{"trace", "debug", "info", "warn", "error", "fatal", "unspecified", "invalid"} {
		program, err := querycel.Compile("logs", "otel.severityBand(log.severity_number) == "+celLiteral(band))
		if err != nil {
			t.Fatal(err)
		}
		for number := uint64(0); number < 256; number++ {
			record := celRecord(number, "ordinary", "api")
			got, err := program.Eval(record)
			if err != nil {
				t.Fatal(err)
			}
			if want := exploration.LogSeverityClass(number) == band; got != want || exploration.Match(program.Predicate(), record) != want {
				t.Fatalf("code %d band %s failed CEL/Goo/OTel partition parity", number, band)
			}
		}
	}
}

func TestHegelCELParentConservation(t *testing.T) {
	program, err := querycel.Compile("logs", `otel.severityBand(log.severity_number) == "error" && log.body.contains("timeout")`)
	if err != nil {
		t.Fatal(err)
	}
	hegel.Test(t, func(tc *hegel.T) {
		parent := exploration.Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 7, Offset: 0, LogSeverity: true,
			Where: &exploration.Predicate{Any: []exploration.Predicate{
				{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "api"},
				{Field: &exploration.Field{Kind: "resource", Key: "region"}, Op: "ne", Value: "excluded"},
			}}}
		before, _ := json.Marshal(parent)
		applied, err := program.Apply(parent)
		if err != nil {
			tc.Fatal(err)
		}
		if applied.Environment != parent.Environment || applied.FromMS != parent.FromMS || applied.ToMS != parent.ToMS || applied.Limit != parent.Limit || applied.Offset != parent.Offset || applied.LogSeverity != parent.LogSeverity || !reflect.DeepEqual(applied.Where.All[0], *parent.Where) {
			tc.Fatal("application changed original question or OR subtree", applied)
		}
		number := uint64(hegel.Draw(tc, hegel.Integers(0, 255)))
		record := celRecord(number, hegel.Draw(tc, hegel.SampledFrom([]string{"timeout", "TIMEOUT", "ordinary"})), hegel.Draw(tc, hegel.SampledFrom([]string{"api", "worker"})))
		if hegel.Draw(tc, hegel.Booleans()) {
			record[exploration.Field{Kind: "resource", Key: "region"}] = hegel.Draw(tc, hegel.SampledFrom([]string{"", "west", "excluded"}))
		}
		result, err := program.Eval(record)
		if err != nil {
			tc.Fatal(err)
		}
		if got, want := exploration.Match(applied.Where, record), exploration.Match(parent.Where, record) && result; got != want {
			tc.Fatal("application broadened or reinterpreted original membership")
		}
		applied.Where.All[0].Any[0].Value = "changed"
		after, _ := json.Marshal(parent)
		if string(before) != string(after) {
			tc.Fatal("applied predicate aliases parent")
		}
	}, options(120)...)
}

func TestCELUnsupportedAndMissingRefuse(t *testing.T) {
	for _, expression := range []string{
		`!log.body.contains("a")`, `log.body.matches(".*")`, `log.body.startsWith("a")`,
		`log.severity_number > 0u`, `log.severity_number < 24u`, `log.severity_number >= 17`,
		`uint(17) == log.severity_number`, `log.severity_number + 1u == 18u`,
		`resource["service.name"] != "worker"`, `[1u,2u].exists(x, x == log.severity_number)`,
		`otel.severityBand(log.severity_number) == "ERROR"`, `span.name == "request"`,
	} {
		if _, err := querycel.Compile("logs", expression); err == nil {
			t.Errorf("unsupported expression admitted: %s", expression)
		}
	}
	program, err := querycel.Compile("logs", `log.body == "" && log.severity_number >= 0u`)
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range []map[exploration.Field]string{
		{}, {{Kind: "column", Key: "body"}: ""},
		{{Kind: "column", Key: "body"}: "", {Kind: "column", Key: "severity_number"}: "not-a-number"},
		{{Kind: "column", Key: "body"}: "", {Kind: "column", Key: "severity_number"}: "18446744073709551616"},
	} {
		if _, err := program.Eval(record); err == nil {
			t.Error("missing or malformed required value became a boolean verdict")
		}
	}
}

// The test replaces only the fixed compiler-owned physical source with a readonly
// numbers fixture. Expressions and values still travel through checked CEL, the
// ordinary query compiler and ClickHouse's typed parameter transport unchanged.
func TestCELClickHouseParity(t *testing.T) {
	endpoint := os.Getenv("GOO_CEL_CLICKHOUSE_URL")
	if endpoint == "" {
		t.Skip("set GOO_CEL_CLICKHOUSE_URL for required live CEL/ClickHouse parity acceptance")
	}
	fixtureBody := "timeout \x00_%'[.*]\\雪"
	relation := "(SELECT fromUnixTimestamp64Milli(1000) AS Timestamp, map('deployment.environment.name','cel-pilot') AS ResourceAttributes, toUInt8(number) AS SeverityNumber, if(number % 3 = 0, unhex('" + hex.EncodeToString([]byte(fixtureBody)) + "'), 'ordinary') AS Body, if(number % 2 = 0, 'api', 'worker') AS ServiceName FROM numbers(256))"
	expressions := []string{
		`otel.severityBand(log.severity_number) == "error"`,
		`otel.severityBand(log.severity_number) == "unspecified" || otel.severityBand(log.severity_number) == "invalid"`,
		`log.body.contains(` + celLiteral("_%'[.*]\\雪") + `) && record.service == "api"`,
		`log.severity_number != 17u && (record.service == "worker" || log.body.contains("timeout"))`,
	}
	for i, expression := range expressions {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			program, err := querycel.Compile("logs", expression)
			if err != nil {
				t.Fatal(err)
			}
			query, err := program.Apply(exploration.Query{Signal: "logs", Environment: "cel-pilot", FromMS: 1000, ToMS: 61000, Limit: 1})
			if err != nil {
				t.Fatal(err)
			}
			plan, err := exploration.Compile(query)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(plan.Count, " FROM otel_logs WHERE ") != 1 {
				t.Fatal("compiler source boundary changed")
			}
			countSQL := strings.Replace(plan.Count, " FROM otel_logs WHERE ", " FROM "+relation+" WHERE ", 1)
			rows, err := reportedCH353(t.Context(), endpoint, countSQL, plan.Parameters)
			if err != nil || len(rows) != 1 {
				t.Fatal("live count failed", err, len(rows))
			}
			var actual string
			if err := json.Unmarshal(rows[0]["count"], &actual); err != nil {
				t.Fatal(err)
			}
			want := 0
			for number := uint64(0); number < 256; number++ {
				body := "ordinary"
				if number%3 == 0 {
					body = fixtureBody
				}
				service := "api"
				if number%2 != 0 {
					service = "worker"
				}
				record := celRecord(number, body, service)
				got, err := program.Eval(record)
				if err != nil {
					t.Fatal(err)
				}
				if got != exploration.Match(query.Where, record) {
					t.Fatal("CEL/Goo reference mismatch")
				}
				if got {
					want++
				}
			}
			if actual != strconv.Itoa(want) {
				t.Fatalf("ClickHouse count=%s, CEL/Goo count=%d", actual, want)
			}
			t.Logf("256 typed records: CEL, Goo and ClickHouse agree on %s matches", actual)
		})
	}
}
