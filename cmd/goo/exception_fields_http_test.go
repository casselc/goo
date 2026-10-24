package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casselc/goo/exploration"
)

type exceptionFieldTransport func(*http.Request) (*http.Response, error)

func (f exceptionFieldTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestComparisonHTTPNullableDerivedFieldBoundary(t *testing.T) {
	reads := 0
	mux := http.NewServeMux()
	registerExploration(mux, exploration.Source{Client: &http.Client{Transport: exceptionFieldTransport(func(*http.Request) (*http.Response, error) { reads++; return nil, errors.New("unexpected source read") })}})
	worker, empty := "worker", ""
	for _, tc := range []struct {
		name   string
		field  exploration.Field
		value  *string
		status int
	}{
		{"legacy present fixed column", exploration.Field{Kind: "column", Key: "service"}, &worker, 200},
		{"malformed missing fixed column", exploration.Field{Kind: "column", Key: "service"}, nil, 400},
		{"derived no unique reported type", exploration.Field{Kind: "column", Key: "exception_type"}, nil, 200},
		{"derived present empty type", exploration.Field{Kind: "column", Key: "exception_type"}, &empty, 200},
		{"legacy missing map key", exploration.Field{Kind: "attribute", Key: "exception.type"}, nil, 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100, GroupBy: []exploration.Field{tc.field}}
			cohort := exploration.Cohort{Question: q, Basis: "held", Count: "1", State: "observed", Groups: []exploration.ValueCount{{Value: tc.value, Count: "1"}}}
			body, e := json.Marshal(exploration.ComparisonRequest{Baseline: cohort, Selection: cohort})
			if e != nil {
				t.Fatal(e)
			}
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, httptest.NewRequest("POST", "/compare", bytes.NewReader(body)))
			if response.Code != tc.status {
				t.Fatalf("status %d: %s", response.Code, response.Body.String())
			}
			if tc.status == 200 {
				var result exploration.CohortComparison
				if e = json.Unmarshal(response.Body.Bytes(), &result); e != nil {
					t.Fatal(e)
				}
				if result.Field != tc.field || result.BaselineCount != "1" || len(result.Rows) != 1 || (result.Rows[0].Value == nil) != (tc.value == nil) {
					t.Fatalf("comparison lost field/presence: %+v", result)
				}
			}
		})
	}
	if reads != 0 {
		t.Fatalf("pure comparison performed %d source reads", reads)
	}
}
