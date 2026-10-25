package exploration

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

type countedValuesTransport func(*http.Request) (*http.Response, error)

func (f countedValuesTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func countedValuesTestSource(t *testing.T, count, groups string) (Source, *int) {
	t.Helper()
	reads := 0
	client := &http.Client{Transport: countedValuesTransport(func(r *http.Request) (*http.Response, error) {
		reads++
		body, _ := io.ReadAll(r.Body)
		sql := string(body)
		if !strings.Contains(sql, "count()") || strings.Contains(sql, " LIMIT 101 ") || strings.Contains(sql, "SpanAttributes) AS") {
			t.Fatalf("unexpected nonaggregate read %s", sql)
		}
		text := count
		if strings.Contains(sql, "GROUP BY") {
			text = groups
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(text)), Header: http.Header{}}, nil
	})}
	return Source{Endpoint: "http://source.invalid", Client: client}, &reads
}
func countedValuesQuestion() Query {
	return Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100, GroupBy: []Field{{Kind: "attribute", Key: "tag"}}, Where: &Predicate{Numeric: &NumericBinding{Field: Field{Kind: "attribute", Key: "amount"}, Type: "int64", Unit: "1", Parse: "canonical-decimal"}, Op: "gte", Value: "1"}}
}
func TestCountedValuesAggregateReadsAndStates(t *testing.T) {
	for _, c := range []struct{ count, groups, state string }{{`{"count":"3"}`, "{\"g0\":null,\"count\":\"1\"}\n{\"g0\":\"\",\"count\":\"2\"}", "observed"}, {`{"count":"0"}`, "", "no_data"}, {`{"count":"4"}`, `{"g0":"x","count":"3"}`, "changed_between_reads"}, {`{"count":"0"}`, `{"g0":"x","count":"1"}`, "changed_between_reads"}} {
		source, reads := countedValuesTestSource(t, c.count, c.groups)
		q := countedValuesQuestion()
		r, e := source.FetchCountedValues(context.Background(), q)
		if e != nil || r.State != c.state || *reads != 2 {
			t.Fatal(r, e, *reads)
		}
		if r.Question.Where.Numeric == nil || r.Question.Where.Numeric.Field.Key != "amount" || len(r.Basis) != 64 || len(r.Reads.Count.Basis) != 64 || r.Reads.Count.ReadAt == "" || r.Reads.Groups.ReadAt == "" {
			t.Fatal(r)
		}
		q.Where.Value = "2"
		if r.Question.Where.Value != "1" {
			t.Fatal("caller mutation changed receipt")
		}
	}
}
func TestCountedValuesMalformedAndBounded(t *testing.T) {
	cases := []struct{ count, groups string }{{`{"count":"01"}`, ""}, {`{}`, ""}, {`{"count":1}`, ""}, {`{"count":"1"}`, `{"count":"1"}`}, {`{"count":"1"}`, `{"g0":1,"count":"1"}`}, {`{"count":"1"}`, `{"g0":"x","count":"0"}`}, {`{"count":"1"}`, `{"g0":"x","count":"01"}`}, {`{"count":"2"}`, "{\"g0\":null,\"count\":\"1\"}\n{\"g0\":null,\"count\":\"1\"}"}}
	var many strings.Builder
	for i := 0; i < 1001; i++ {
		fmt.Fprintf(&many, "{\"g0\":\"%d\",\"count\":\"1\"}\n", i)
	}
	cases = append(cases, struct{ count, groups string }{`{"count":"1001"}`, many.String()})
	for _, c := range cases {
		source, _ := countedValuesTestSource(t, c.count, c.groups)
		if _, e := source.FetchCountedValues(context.Background(), countedValuesQuestion()); e == nil {
			t.Fatal("malformed response admitted", c.count, c.groups)
		}
	}
}
func TestCountedValuesValidationBeforeSourceAndCancellation(t *testing.T) {
	base := countedValuesQuestion()
	for _, edit := range []func(*Query){func(q *Query) { q.Offset = 1 }, func(q *Query) { q.GroupBy = nil }, func(q *Query) { q.GroupBy = []Field{{Kind: "column", Key: "duration_ns"}} }, func(q *Query) { q.SpanStatus = true }, func(q *Query) { q.BucketMS = 1000 }, func(q *Query) { q.Summary = "span-duration" }, func(q *Query) {
		q.Numeric = &NumericBinding{Field: Field{Kind: "attribute", Key: "amount"}, Type: "int64", Unit: "1", Parse: "canonical-decimal"}
	}} {
		q := base
		edit(&q)
		source, n := countedValuesTestSource(t, `{"count":"0"}`, "")
		if _, e := source.FetchCountedValues(context.Background(), q); e == nil || *n != 0 {
			t.Fatal("validation read source", e, *n)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	source := Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: countedValuesTransport(func(r *http.Request) (*http.Response, error) { return nil, r.Context().Err() })}}
	if _, e := source.FetchCountedValues(ctx, base); e == nil {
		t.Fatal("cancellation swallowed")
	}
}

func TestCountedValuesCancellationAfterCount(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reads := 0
	source := Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: countedValuesTransport(func(r *http.Request) (*http.Response, error) {
		reads++
		cancel()
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"count":"1"}`))}, nil
	})}}
	if _, e := source.FetchCountedValues(ctx, countedValuesQuestion()); e == nil || reads != 1 {
		t.Fatal("cancelled count triggered group read", e, reads)
	}
}
func TestCountedValuesDerivedFieldPolicies(t *testing.T) {
	for _, c := range []struct {
		field   string
		group   string
		refused bool
	}{{"service", `{"g0":null,"count":"1"}`, true}, {"exception_type", `{"g0":null,"count":"1"}`, false}, {"body_pattern_v1", `{"g0":"not-canonical","count":"1"}`, true}} {
		q := countedValuesQuestion()
		q.GroupBy = []Field{{Kind: "column", Key: c.field}}
		if c.field == "body_pattern_v1" {
			q.Signal = "logs"
		}
		source, _ := countedValuesTestSource(t, `{"count":"1"}`, c.group)
		_, e := source.FetchCountedValues(context.Background(), q)
		if (e != nil) != c.refused {
			t.Fatal(c, e)
		}
	}
}
