package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
)

func TestCountCohortHTTPAlignmentIsSourceFree307(t *testing.T) {
	reads := 0
	source := exploration.Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: nativeDependencyTransport(func(*http.Request) (*http.Response, error) {
		reads++
		return nil, errors.New("comparison cannot read source")
	})}}
	mux := http.NewServeMux()
	registerExploration(mux, source, &exploration.ResultSnapshots{}, nil)
	q := exploration.Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100, GroupBy: []exploration.Field{{Kind: "attribute", Key: "http.route"}}, Where: &exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "cohort"}, Op: "eq", Value: "baseline"}}
	label := "/checkout"
	baseline := exploration.Cohort{Question: q, CapturedAt: "2026-11-01T00:00:00Z", Basis: "untrusted-a", Count: "2", State: "observed", Groups: []exploration.ValueCount{{Value: &label, Count: "2"}}}
	raw, _ := json.Marshal(baseline)
	var selection exploration.Cohort
	json.Unmarshal(raw, &selection)
	selection.Question.Where.Value = "selection"
	selection.Basis = "untrusted-b"
	selection.Count = "5"
	selection.Groups[0].Count = "5"
	observations := []apiObservation{}
	for _, state := range []string{"observed", "no_data", "changed_between_reads"} {
		t.Run(state, func(t *testing.T) {
			candidate := selection
			candidate.State = state
			if state == "no_data" {
				candidate.Count = "0"
				candidate.Groups = []exploration.ValueCount{}
			}
			if state == "changed_between_reads" {
				candidate.Count = "1"
			}
			request := exploration.ComparisonRequest{Baseline: baseline, Selection: candidate}
			body, _ := json.Marshal(request)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, httptest.NewRequest("POST", "/compare", strings.NewReader(string(body))))
			if w.Code != 200 {
				t.Fatal(w.Code, w.Body.String())
			}
			var result exploration.CohortComparison
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			group := result.GroupComparison
			if group == nil {
				t.Fatal("different predicates lost aligned frame")
			}
			var qa, qb exploration.Query
			json.Unmarshal(group.Baseline.Scope.Selector, &qa)
			json.Unmarshal(group.Selection.Scope.Selector, &qb)
			if !reflect.DeepEqual(qa, baseline.Question) || !reflect.DeepEqual(qb, candidate.Question) || group.Baseline.Completeness != nil || group.Selection.Completeness != nil {
				t.Fatal("independent scopes/source authority changed")
			}
			if state == "changed_between_reads" && group.Rows[0].Change.Value != nil {
				t.Fatal("diagnostic supplied arithmetic")
			}
			if state == "observed" && *group.Rows[0].Change.Value != "3" {
				t.Fatal("wrong independent-cohort delta")
			}
			if state == "no_data" && *group.Rows[0].Change.Value != "-2" {
				t.Fatal("empty cohort lost known zero")
			}
			observations = append(observations, apiObservation{Method: "POST", Path: "/compare", Status: 200, Request: body, Response: append(json.RawMessage(nil), w.Body.Bytes()...)})
		})
	}

	t.Run("span duration summary", func(t *testing.T) {
		duration := func(count, sum string) *exploration.DurationSummary {
			return &exploration.DurationSummary{State: "observed", Count: count, SumNS: sum, MinNS: "10", MaxNS: "10", P50NS: "10", P95NS: "10", P99NS: "10"}
		}
		a, b := baseline, selection
		a.Question.Signal = "spans"
		a.Question.Summary = "span-duration"
		b.Question.Signal = "spans"
		b.Question.Summary = "span-duration"
		a.Duration = duration("2", "20")
		b.Duration = duration("5", "50")
		a.Groups = []exploration.ValueCount{{Value: &label, Count: "2", Duration: duration("2", "20")}}
		b.Groups = []exploration.ValueCount{{Value: &label, Count: "5", Duration: duration("5", "50")}}
		request := exploration.ComparisonRequest{Baseline: a, Selection: b}
		direct, err := exploration.CompareCohorts(request)
		if err != nil || direct.GroupComparison == nil || *direct.GroupComparison.Rows[0].Change.Value != "3" {
			t.Fatal("complete duration count cohorts refused", err)
		}
		body, _ := json.Marshal(request)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/compare", strings.NewReader(string(body))))
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		var result exploration.CohortComparison
		if err = json.Unmarshal(w.Body.Bytes(), &result); err != nil || result.GroupComparison == nil || *result.GroupComparison.Rows[0].Change.Value != "3" {
			t.Fatal("HTTP duration count alignment lost", err)
		}
		if result.GroupComparison.Baseline.Completeness != nil || result.GroupComparison.Selection.Completeness != nil {
			t.Fatal("duration summary gained source authority")
		}
		observations = append(observations, apiObservation{Method: "POST", Path: "/compare", Status: 200, Request: body, Response: append(json.RawMessage(nil), w.Body.Bytes()...)})
		// Complete summary payloads are required independently at both levels;
		// do not silently drop duration merely because the chart displays count.
		for _, missing := range []string{"top-level", "group"} {
			candidate := request
			if missing == "top-level" {
				candidate.Baseline.Duration = nil
			} else {
				candidate.Baseline.Groups = []exploration.ValueCount{{Value: &label, Count: "2"}}
			}
			body, _ = json.Marshal(candidate)
			w = httptest.NewRecorder()
			mux.ServeHTTP(w, httptest.NewRequest("POST", "/compare", strings.NewReader(string(body))))
			if w.Code != 400 {
				t.Fatal("missing duration admitted", missing, w.Code, w.Body.String())
			}
		}
	})
	if reads != 0 {
		t.Fatal("comparison acquired source", reads)
	}
	if file := os.Getenv("GOO_COUNT_COHORT_OBSERVATIONS"); file != "" {
		raw, _ := json.MarshalIndent(observations, "", "  ")
		if err := os.WriteFile(file, append(raw, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
}
