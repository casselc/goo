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
		case p.Rows:
			text = `{"timestamp_ns":"1000000001","duration_ns":"10"}` + "\n" + `{"timestamp_ns":"1000000002","duration_ns":"20"}`
		case p.Fields:
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

// Fleet identity is the full tuple: the same name in a missing namespace and a
// present-empty namespace must never become one measured service.
func TestRecordAggregatesQualifiedServices(t *testing.T) {
	q := Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100, GroupBy: []Field{{Kind: "column", Key: "service"}, {Kind: "resource", Key: "service.namespace"}}, Summary: "span-duration", SpanStatus: true, Where: &Predicate{Field: &Field{Kind: "column", Key: "kind"}, Op: "eq", Value: "Server"}}
	count := `{"count":"2","sum_ns":"30","min_ns":"10","max_ns":"20","quantiles_ns":["20","20","20"],"status_error":"1","status_ok":"0","status_unset":"1","status_unknown":"0"}`
	a := `{"g0":"api","g1":null,"count":"1","sum_ns":"10","min_ns":"10","max_ns":"10","quantiles_ns":["10","10","10"],"status_error":"1","status_ok":"0","status_unset":"0","status_unknown":"0"}`
	b := `{"g0":"api","g1":"","count":"1","sum_ns":"20","min_ns":"20","max_ns":"20","quantiles_ns":["20","20","20"],"status_error":"0","status_ok":"0","status_unset":"1","status_unknown":"0"}`
	run := func(t *testing.T, total, groups string) (*RecordAggregatesResult, error) {
		t.Helper()
		p, err := Compile(q)
		if err != nil {
			t.Fatal(err)
		}
		reads := []string{}
		transport := recordWireTransport(func(r *http.Request) (*http.Response, error) {
			raw, _ := io.ReadAll(r.Body)
			sql := string(raw)
			reads = append(reads, sql)
			text := ""
			switch sql {
			case p.Count:
				text = total
			case p.Groups:
				text = groups
			default:
				t.Fatalf("unexpected inspection read: %s", sql)
			}
			return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(text))}, nil
		})
		result, err := (Source{Endpoint: "http://fixture.invalid", Client: &http.Client{Transport: transport}}).FetchRecordAggregates(context.Background(), q)
		if len(reads) != 2 {
			t.Fatalf("got %d reads, want exactly count+groups", len(reads))
		}
		return result, err
	}
	t.Run("missing and empty stay distinct", func(t *testing.T) {
		r, err := run(t, count, a+"\n"+b)
		if err != nil || r.State != "observed" || len(r.Groups) != 2 {
			t.Fatal(r, err)
		}
		var groups []map[string]any
		for _, raw := range r.Groups {
			var g map[string]any
			if json.Unmarshal(raw, &g) != nil {
				t.Fatal("invalid normalized group")
			}
			groups = append(groups, g)
		}
		if groups[0]["g1"] != nil || groups[1]["g1"] != "" {
			t.Fatal(groups)
		}
		if r.Reads.Count.Basis == "" || r.Reads.Groups.Basis == "" || r.Basis == "" {
			t.Fatal("missing independent read basis")
		}
	})
	for name, total := range map[string]string{"duration": strings.Replace(count, `"sum_ns":"30"`, `"sum_ns":"31"`, 1), "status": strings.Replace(strings.Replace(count, `"status_error":"1"`, `"status_error":"0"`, 1), `"status_ok":"0"`, `"status_ok":"1"`, 1)} {
		t.Run(name+" disagreement", func(t *testing.T) {
			r, err := run(t, total, a+"\n"+b)
			if err != nil || r.State != "changed_between_reads" {
				t.Fatal(r, err)
			}
		})
	}
	for name, groups := range map[string]string{"duplicate tuple": a + "\n" + a, "missing second dimension": strings.Replace(a, `"g1":null,`, "", 1) + "\n" + b, "invalid second dimension": strings.Replace(a, `"g1":null`, `"g1":3`, 1) + "\n" + b, "group sentinel": strings.Repeat(a+"\n", 1001)} {
		t.Run(name, func(t *testing.T) {
			if r, err := run(t, count, groups); err == nil {
				t.Fatal("invalid groups admitted", r)
			}
		})
	}
	t.Run("count disagreement", func(t *testing.T) {
		q.Summary = ""
		q.SpanStatus = false
		r, err := run(t, `{"count":"3"}`, `{"g0":"api","g1":null,"count":"1"}`+"\n"+`{"g0":"api","g1":"","count":"1"}`)
		if err != nil || r.State != "changed_between_reads" {
			t.Fatal(r, err)
		}
	})
	t.Run("second field type validated", func(t *testing.T) {
		q.GroupBy[1] = Field{Kind: "column", Key: "duration_ns"}
		if ValidateRecordAggregatesQuery(q) == nil {
			t.Fatal("numeric second dimension admitted")
		}
	})
}
