package exploration

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

const severityRow435 = `"count":"8","severity_trace":"1","severity_debug":"1","severity_info":"1","severity_warn":"1","severity_error":"1","severity_fatal":"1","severity_unspecified":"1","severity_invalid":"1"`

func severityQuestion435() Query {
	return Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100, GroupBy: []Field{{Kind: "column", Key: "severity"}}, BucketMS: 10000, LogSeverity: true}
}

func severitySource435(t *testing.T, q Query, mode string) (Source, *[]string) {
	t.Helper()
	plan, err := Compile(q)
	if err != nil {
		t.Fatal(err)
	}
	reads := []string{}
	transport := recordWireTransport(func(request *http.Request) (*http.Response, error) {
		raw, _ := io.ReadAll(request.Body)
		sql := string(raw)
		reads = append(reads, sql)
		text := ""
		measures := severityRow435
		if mode == "zero" {
			measures = strings.ReplaceAll(measures, `"1"`, `"0"`)
			measures = strings.ReplaceAll(measures, `"8"`, `"0"`)
		}
		changed := strings.ReplaceAll(strings.ReplaceAll(measures, `"severity_error":"1"`, `"severity_error":"0"`), `"severity_fatal":"1"`, `"severity_fatal":"2"`)
		switch sql {
		case plan.Count:
			text = "{" + measures + "}"
		case plan.Groups:
			if mode != "zero" {
				if mode == "group-drift" {
					measures = changed
				}
				if mode == "missing" {
					measures = strings.ReplaceAll(measures, `,"severity_invalid":"1"`, "")
				}
				text = `{"g0":"ERROR",` + measures + "}"
			}
		case plan.TimeSeries:
			if mode != "zero" {
				if mode == "time-drift" {
					measures = changed
				}
				text = `{"from_ms":"1000",` + measures + "}"
			}
		case plan.Rows:
			// Ordinary exploration support is independent of the aggregate-only path.
			if mode != "zero" {
				text = `{"timestamp_ns":"1000000001","Body":"sample","SeverityText":"ERROR","ResourceAttributes":{"deployment.environment.name":"prod"}}`
			}
		case plan.Fields:
		default:
			t.Fatalf("unexpected read: %s", sql)
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(text))}, nil
	})
	return Source{Endpoint: "http://fixture.invalid", Client: &http.Client{Transport: transport}}, &reads
}

func TestLogSeverityClassAndStrictAccounting435(t *testing.T) {
	for _, tc := range []struct {
		number uint64
		band   string
	}{{0, "unspecified"}, {1, "trace"}, {4, "trace"}, {5, "debug"}, {8, "debug"}, {9, "info"}, {12, "info"}, {13, "warn"}, {16, "warn"}, {17, "error"}, {20, "error"}, {21, "fatal"}, {24, "fatal"}, {25, "invalid"}, {255, "invalid"}, {^uint64(0), "invalid"}} {
		if got := LogSeverityClass(tc.number); got != tc.band {
			t.Fatalf("%d: %s != %s", tc.number, got, tc.band)
		}
	}
	value, err := DecodeLogSeverity([]byte("{" + severityRow435 + "}"))
	if err != nil || value.Unspecified != "1" || value.Invalid != "1" {
		t.Fatal(value, err)
	}
	for _, bad := range []string{
		strings.ReplaceAll(severityRow435, `"severity_info":"1"`, `"severity_info":"2"`),
		strings.ReplaceAll(severityRow435, `,"severity_invalid":"1"`, ""),
		strings.ReplaceAll(severityRow435, `"severity_info":"1"`, `"severity_info":1`),
		strings.ReplaceAll(severityRow435, `"severity_info":"1"`, `"severity_info":"01"`),
		strings.ReplaceAll(severityRow435, `"count":"8"`, `"count":"18446744073709551616"`),
	} {
		if _, err := DecodeLogSeverity([]byte("{" + bad + "}")); err == nil {
			t.Fatal("invalid severity admitted", bad)
		}
	}
	large := LogSeveritySummary{Count: "18446744073709551615", Trace: "0", Debug: "0", Info: "0", Warn: "0", Error: "18446744073709551615", Fatal: "0", Unspecified: "0", Invalid: "0"}
	if large.Validate() != nil {
		t.Fatal("valid maximum count rejected")
	}
}

func TestLogSeverityAggregatesConserveEachBand435(t *testing.T) {
	q := severityQuestion435()
	for _, mode := range []string{"valid", "zero", "group-drift", "time-drift", "missing"} {
		t.Run(mode, func(t *testing.T) {
			source, reads := severitySource435(t, q, mode)
			result, err := source.FetchRecordAggregates(context.Background(), q)
			if mode == "missing" {
				if err == nil {
					t.Fatal("missing band admitted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(*reads) != 3 {
				t.Fatal("aggregate path read rows or discovery", *reads)
			}
			want := "observed"
			if mode == "zero" {
				want = "no_data"
			}
			if strings.Contains(mode, "drift") {
				want = "changed_between_reads"
			}
			if result.State != want || result.LogSeverity == nil || len(result.TimeSeries) != 6 {
				t.Fatal(result)
			}
			if result.TimeSeries[5].LogSeverity == nil || *result.TimeSeries[5].LogSeverity != emptyLogSeverity() {
				t.Fatal("empty interval lost full zero partition")
			}
			if mode == "valid" {
				var g map[string]json.RawMessage
				_ = json.Unmarshal(result.Groups[0], &g)
				if g["logSeverity"] == nil || g["severity_info"] != nil {
					t.Fatal("raw aliases leaked", g)
				}
			}
		})
	}
	source, _ := severitySource435(t, q, "valid")
	result, err := source.Fetch(context.Background(), q)
	if err != nil || result.LogSeverity == nil || result.LogSeverity.Error != "1" || validateLogSeverityResult(result) != nil {
		t.Fatal(result, err)
	}
	if result.ViewDataset != nil || result.SummaryDataset != nil || result.TimeDataset != nil {
		t.Fatal("unsupported generic dataset silently attested severity")
	}
	result.LogSeverity.Error = "0"
	if validateLogSeverityResult(result) == nil {
		t.Fatal("tampered saved partition admitted")
	}
}

func TestLogSeverityGrammarAndUnsupportedProjections435(t *testing.T) {
	q := severityQuestion435()
	plan, err := Compile(q)
	if err != nil {
		t.Fatal(err)
	}
	for _, sql := range []string{plan.Count, plan.Groups, plan.TimeSeries} {
		for _, fragment := range []string{"SeverityNumber BETWEEN 1 AND 4", "SeverityNumber BETWEEN 17 AND 20", "SeverityNumber = 0", "SeverityNumber > 24"} {
			if !strings.Contains(sql, fragment) {
				t.Fatal("missing numeric binding", sql)
			}
		}
	}
	if strings.Contains(plan.Rows, "severity_error") || strings.Contains(plan.Fields, "severity_error") {
		t.Fatal("inspection became aggregate work")
	}
	for _, edit := range []func(*Query){func(q *Query) { q.Signal = "spans" }, func(q *Query) { q.GroupTimeSeries = true }, func(q *Query) {
		q.Numeric = &NumericBinding{Field: Field{Kind: "attribute", Key: "value"}, Type: "int64", Unit: "1", Parse: "canonical-decimal"}
	}} {
		bad := q
		edit(&bad)
		if bad.Validate() == nil {
			t.Fatal("unsupported combined projection admitted", bad)
		}
	}
	for _, validate := range []func(Query) error{ValidateCountedValuesQuery, ValidateCapturedCountQuery, ValidateCountCohortQuery, ValidateExceptionGroupsQuery} {
		if validate(q) == nil {
			t.Fatal("unsupported projection dropped severity")
		}
	}
	definition, err := DraftTelemetryPackage(q, "numeric-severity", "Numeric severity", 1)
	if err != nil {
		t.Fatal(err)
	}
	packagePlan, err := PlanTelemetryPackage(definition, q.FromMS, q.ToMS, nil)
	if err != nil || !packagePlan.Question.LogSeverity || !strings.Contains(packagePlan.YAML, "logSeverity: true") {
		t.Fatal(packagePlan, err)
	}
	decoded, err := ReadTelemetryPackage(packagePlan.YAML)
	if err != nil || !decoded.LogSeverity {
		t.Fatal(decoded, err)
	}
	measures, err := q.RecordMeasures()
	if err != nil {
		t.Fatal(err)
	}
	last := measures.Measures[len(measures.Measures)-1]
	if last.ID != "numeric_log_severity" || last.Input == nil || !reflect.DeepEqual(*last.Input, Field{Kind: "column", Key: "severity_number"}) {
		t.Fatal(last)
	}
}
