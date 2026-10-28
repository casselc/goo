package main

import (
	"bytes"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestDistributionComparisonSelectionHTTP(t *testing.T) {
	resultSnapshots := &exploration.ResultSnapshots{}
	binding := exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work.items"}, Type: "uint64", Unit: "items", Parse: "canonical-decimal"}
	cohort := func(service string, from, to int64) exploration.Cohort {
		q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: from, ToMS: to, Limit: 100, Numeric: &binding, NumericHistogram: &exploration.NumericHistogramSpec{Binning: "powers-of-two"}, Where: &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: service}}
		one, two := "1", "2"
		raw := []*string{&one, &two, nil}
		s, e := exploration.EvaluateNumericValues(binding, raw)
		if e != nil {
			t.Fatal(e)
		}
		h, e := exploration.EvaluateNumericHistogram(binding, raw)
		if e != nil {
			t.Fatal(e)
		}
		return exploration.Cohort{Question: q, Basis: "caller-held", Count: "3", State: "observed", Numeric: &s, NumericHistogram: &h, Groups: []exploration.ValueCount{}}
	}
	req := exploration.ComparisonRequest{Numeric: true, Distribution: true, Baseline: cohort("earlier-worker", 1000, 2000), Selection: cohort("later-worker", 4000, 5000)}
	mux := http.NewServeMux()
	registerWorkbench(mux, t.TempDir(), retentionConnection{}, resultSnapshots, nil)
	call := func(body map[string]any) *httptest.ResponseRecorder {
		raw, e := json.Marshal(body)
		if e != nil {
			t.Fatal(e)
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/numeric-histograms/comparison-selection", bytes.NewReader(raw)))
		return w
	}
	for _, side := range []string{"baseline", "selection"} {
		w := call(map[string]any{"comparison": req, "side": side, "fromBin": 1, "toBin": 2})
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		var p exploration.NumericHistogramSelection
		if e := json.Unmarshal(w.Body.Bytes(), &p); e != nil {
			t.Fatal(e)
		}
		want := req.Baseline.Question
		if side == "selection" {
			want = req.Selection.Question
		}
		if !reflect.DeepEqual(p.Parent, want) || p.Selected.FromMS != want.FromMS || p.Selected.ToMS != want.ToMS {
			t.Fatal("planner substituted other side scope", p.Parent)
		}
	}
	for _, body := range []map[string]any{
		{"comparison": req, "side": "baseline", "toBin": 2},
		{"comparison": req, "side": "baseline", "fromBin": 1},
		{"comparison": req, "side": "other", "fromBin": 1, "toBin": 2},
		{"comparison": req, "side": "selection", "fromBin": 0, "toBin": 1},
		{"comparison": req, "side": "selection", "fromBin": 1, "toBin": 2, "ignore": true},
	} {
		if w := call(body); w.Code != 400 {
			t.Fatal("bad selection admitted", w.Code, w.Body.String())
		}
	}
	req.Distribution = false
	if w := call(map[string]any{"comparison": req, "side": "baseline", "fromBin": 1, "toBin": 2}); w.Code != 400 {
		t.Fatal("implicit distribution mode admitted")
	}
}
