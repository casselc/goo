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

func aggregateFixture(t *testing.T, q Query, seriesChanged bool) (Source, *[]string) {
	t.Helper()
	p, e := Compile(q)
	if e != nil {
		t.Fatal(e)
	}
	reads := []string{}
	measures := `"count":"2","sum_ns":"30","min_ns":"10","max_ns":"20","quantiles_ns":["20","20","20"],"status_error":"1","status_ok":"0","status_unset":"1","status_unknown":"0"`
	if q.Signal == "logs" {
		measures = `"count":"2"`
	}
	transport := recordWireTransport(func(r *http.Request) (*http.Response, error) {
		b, _ := io.ReadAll(r.Body)
		sql := string(b)
		reads = append(reads, sql)
		text := ""
		switch sql {
		case p.Count:
			text = "{" + measures + "}"
		case p.Groups:
			text = `{"g0":null,` + measures + `}`
		case p.TimeSeries:
			text = `{"from_ms":"1000",` + measures + `}`
			if seriesChanged {
				text = strings.ReplaceAll(text, `"sum_ns":"30","min_ns":"10"`, `"sum_ns":"40","min_ns":"20"`)
			}
		case p.Rows, p.Fields:
		default:
			t.Fatalf("unexpected SQL %s", sql)
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(text))}, nil
	})
	return Source{Endpoint: "http://fixture.invalid", Client: &http.Client{Transport: transport}}, &reads
}
func recordAggregateQuestion() Query {
	q := countedValuesQuestion()
	q.Summary = "span-duration"
	q.SpanStatus = true
	q.BucketMS = 1000
	return q
}
func TestRecordAggregatesReuseFullSemanticsWithoutInspection(t *testing.T) {
	for _, signal := range []string{"spans", "logs"} {
		q := recordAggregateQuestion()
		q.Signal = signal
		if signal == "logs" {
			q.Summary = ""
			q.SpanStatus = false
		}
		source, reads := aggregateFixture(t, q, false)
		aggregate, e := source.FetchRecordAggregates(context.Background(), q)
		if e != nil {
			t.Fatal(e)
		}
		if len(*reads) != 3 || aggregate.State != "observed" || aggregate.Reads.TimeSeries == nil {
			t.Fatal(aggregate, len(*reads))
		}
		full, e := source.Fetch(context.Background(), q)
		if e != nil {
			t.Fatal(e)
		}
		if full.Count != aggregate.Count || !reflect.DeepEqual(full.Groups, aggregate.Groups) || !reflect.DeepEqual(full.Duration, aggregate.Duration) || !reflect.DeepEqual(full.SpanStatus, aggregate.SpanStatus) || !reflect.DeepEqual(full.TimeSeries, aggregate.TimeSeries) {
			t.Fatal("projection semantics diverged")
		}
		encoded, _ := json.Marshal(aggregate)
		for _, key := range []string{`"rows":`, `"fields":`, `"plan":`, `"resultToken":`, `"hasMore":`} {
			if strings.Contains(string(encoded), key) {
				t.Fatal("inspection contract leaked", key)
			}
		}
		q.Where.Value = "2"
		if aggregate.Question.Where.Value != "1" {
			t.Fatal("question aliased")
		}
	}
}
func TestRecordAggregatesSeriesDurationAgreementAndLegacyPolicy(t *testing.T) {
	q := recordAggregateQuestion()
	source, _ := aggregateFixture(t, q, true)
	aggregate, e := source.FetchRecordAggregates(context.Background(), q)
	if e != nil || aggregate.State != "changed_between_reads" {
		t.Fatal(aggregate, e)
	}
	full, e := source.Fetch(context.Background(), q)
	if e != nil || full.State != "observed" {
		t.Fatal("legacy series agreement changed", full, e)
	}
}
func TestRecordAggregatesTwoReadDimensionAndValidation(t *testing.T) {
	q := recordAggregateQuestion()
	q.BucketMS = 0
	source, reads := aggregateFixture(t, q, false)
	r, e := source.FetchRecordAggregates(context.Background(), q)
	if e != nil || len(*reads) != 2 || r.Reads.TimeSeries != nil || r.TimeSeries != nil {
		t.Fatal(r, e)
	}
	for _, edit := range []func(*Query){func(q *Query) { q.Offset = 1 }, func(q *Query) { q.GroupTimeSeries = true; q.BucketMS = 1000 }, func(q *Query) { q.Numeric = q.Where.Numeric }, func(q *Query) { q.GroupBy = nil }, func(q *Query) { q.GroupBy = []Field{{Kind: "column", Key: "duration_ns"}} }} {
		bad := q
		edit(&bad)
		if ValidateRecordAggregatesQuery(bad) == nil {
			t.Fatal("unsupported projection accepted", bad)
		}
	}
}
