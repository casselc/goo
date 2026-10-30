package main

import (
	"bytes"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNumericComparisonIntentHTTP(t *testing.T) {
	resultSnapshots := &exploration.ResultSnapshots{}
	binding := exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work.items"}, Type: "int64", Unit: "items", Parse: "canonical-decimal"}
	q := exploration.Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 100, GroupBy: []exploration.Field{{Kind: "column", Key: "service"}}, Numeric: &binding}
	cohort := func(values ...string) exploration.Cohort {
		raw := make([]*string, len(values))
		for i := range values {
			raw[i] = &values[i]
		}
		s, e := exploration.EvaluateNumericValues(binding, raw)
		if e != nil {
			t.Fatal(e)
		}
		group := "worker"
		return exploration.Cohort{Question: q, Basis: "caller-held", Count: s.Count, State: "observed", Numeric: &s, Groups: []exploration.ValueCount{{Value: &group, Count: s.Count, Numeric: &s}}}
	}
	request := exploration.ComparisonRequest{Numeric: true, Baseline: cohort("-1", "0", "0"), Selection: cohort("0", "0", "0")}
	mux := http.NewServeMux()
	registerExploration(mux, exploration.Source{}, resultSnapshots, nil)
	call := func(v exploration.ComparisonRequest) *httptest.ResponseRecorder {
		raw, e := json.Marshal(v)
		if e != nil {
			t.Fatal(e)
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/compare", bytes.NewReader(raw)))
		return w
	}
	w := call(request)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var result exploration.CohortComparison
	if e := json.Unmarshal(w.Body.Bytes(), &result); e != nil {
		t.Fatal(e)
	}
	if result.Measures == nil || result.Measures.Numeric == nil || result.Measures.Numeric.MeanDifference != "1/3" || result.ViewComparison == nil || result.ViewComparison.Changes["numeric.mean"].Value == nil || *result.ViewComparison.Changes["numeric.mean"].Value != "1/3" {
		t.Fatal("exact rational transport lost", w.Body.String())
	}
	request.Numeric = false
	if w = call(request); w.Code != 400 {
		t.Fatal("numeric payload silently accepted without mode", w.Code)
	}
	request.Numeric = true
	request.Selection.Numeric = nil
	if w = call(request); w.Code != 400 {
		t.Fatal("missing numeric evidence accepted", w.Code)
	}
	request.Numeric = false
	request.Baseline.Numeric = nil
	request.Baseline.Groups[0].Numeric = nil
	request.Selection.Groups[0].Numeric = nil
	if w = call(request); w.Code != 400 || !bytes.Contains(w.Body.Bytes(), []byte("explicit numeric comparison")) {
		t.Fatal("declared numeric profile silently dropped", w.Code, w.Body.String())
	}
	// A count-only question is explicit; it cannot retain a numeric definition
	// while omitting the evidence needed to evaluate that family.
	request.Baseline.Question.Numeric = nil
	request.Selection.Question.Numeric = nil
	if w = call(request); w.Code != 200 {
		t.Fatal("explicit count-only question refused", w.Code, w.Body.String())
	}
	var counts exploration.CohortComparison
	if err := json.Unmarshal(w.Body.Bytes(), &counts); err != nil || counts.Measures != nil || counts.ViewComparison == nil || len(counts.ViewComparison.Baseline.Fields) != 1 || counts.ViewComparison.Baseline.Fields[0].Key != "count" {
		t.Fatal("count-only recipe acquired a numeric family", w.Body.String(), err)
	}
}
