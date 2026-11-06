package exploration

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/casselc/goo/internal/sourcefailure"
)

func TestOrdinarySourceRefusalBoundaries344(t *testing.T) {
	q := Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 10, GroupBy: []Field{{Kind: "attribute", Key: "cohort"}}}
	readers := map[string]func(Source, context.Context, Query) error{
		"records":    func(s Source, c context.Context, q Query) error { _, e := s.Fetch(c, q); return e },
		"page":       func(s Source, c context.Context, q Query) error { _, e := s.FetchRecordPage(c, q); return e },
		"aggregates": func(s Source, c context.Context, q Query) error { _, e := s.FetchRecordAggregates(c, q); return e },
		"values":     func(s Source, c context.Context, q Query) error { _, e := s.FetchCountedValues(c, q); return e },
	}
	for name, read := range readers {
		t.Run(name, func(t *testing.T) {
			cases := []struct {
				name, body string
				status     int
				kind       sourcefailure.Kind
			}{
				{"unavailable", "DO NOT PUBLISH credentials or SQL", 503, sourcefailure.Unavailable},
				{"invalid row", "{", 200, sourcefailure.Malformed},
				{"invalid count", `{"count":"invalid"}`, 200, sourcefailure.Malformed},
				{"row budget", strings.Repeat("{}\n", 1002), 200, sourcefailure.Budget},
			}
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					attempts := 0
					s := Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: recordWireTransport(func(*http.Request) (*http.Response, error) {
						attempts++
						return &http.Response{StatusCode: tc.status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(tc.body))}, nil
					})}}
					err := read(s, context.Background(), q)
					if err == nil || sourcefailure.Describe(err).Refusal.Kind != tc.kind || attempts != 1 {
						t.Fatal(err, attempts)
					}
					invalid := q
					invalid.Signal = "invalid"
					err = read(s, context.Background(), invalid)
					if _, typed := errors.AsType[*sourcefailure.Error](err); err == nil || typed || attempts != 1 {
						t.Fatal("request validation became source failure", err, attempts)
					}
				})
			}
		})
	}
}

func TestSourceFailureContextIdentity344(t *testing.T) {
	for _, cause := range []error{context.Canceled, context.DeadlineExceeded} {
		s := Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: recordWireTransport(func(*http.Request) (*http.Response, error) { return nil, cause })}}
		_, err := s.read(context.Background(), "SELECT", nil)
		if !errors.Is(err, cause) {
			t.Fatal("context cause lost", err)
		}
		want := sourcefailure.Cancelled
		if cause == context.DeadlineExceeded {
			want = sourcefailure.Timeout
		}
		if sourcefailure.Describe(err).Refusal.Kind != want {
			t.Fatal("context hidden by transport marker", err)
		}
	}
}
