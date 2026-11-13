package exploration

import (
	"encoding/json"
	"math/big"
	"reflect"
	"testing"
)

func rateCohort418(signal, count string, from, to int64) Cohort {
	c := Cohort{Question: Query{Signal: signal, Environment: "prod", FromMS: from, ToMS: to, Limit: 100, GroupBy: []Field{{Kind: "column", Key: "service"}}}, Basis: "caller-held-rate", Count: count, State: "observed", Groups: []ValueCount{}}
	if count == "0" {
		c.State = "no_data"
	} else {
		c.Groups = append(c.Groups, ValueCount{Value: viewText("worker"), Count: count})
	}
	return c
}

func TestCapturedRateComparison418ExactUnequalExposure(t *testing.T) {
	for _, signal := range []string{"logs", "spans"} {
		for _, tc := range []struct {
			name, a, b string
			aw, bw     int64
		}{
			{"equal rates different counts", "120", "240", 60000, 120000},
			{"fractional difference", "7", "19", 3001, 9007},
			{"beyond binary integer precision", "9007199254740993", "18014398509481986", 60001, 120002},
			{"negative change", "50", "1", 1000, 1999},
		} {
			t.Run(signal+"/"+tc.name, func(t *testing.T) {
				a, b := rateCohort418(signal, tc.a, 1000, 1000+tc.aw), rateCohort418(signal, tc.b, 300000, 300000+tc.bw)
				a.CapturedAt, b.CapturedAt = "2026-11-13T00:00:00.000000001Z", "2026-11-13T00:00:01.000000002Z"
				before, _ := json.Marshal(ComparisonRequest{Baseline: a, Selection: b})
				result, err := CompareCohorts(ComparisonRequest{Baseline: a, Selection: b})
				if err != nil || result.ViewComparison == nil {
					t.Fatal(result, err)
				}
				comparison := result.ViewComparison
				reference := func(n string, ms int64) *big.Rat {
					count, _ := new(big.Int).SetString(n, 10)
					return new(big.Rat).Mul(new(big.Rat).SetInt(count), big.NewRat(1000, ms))
				}
				expected := new(big.Rat).Sub(reference(tc.b, tc.bw), reference(tc.a, tc.aw)).RatString()
				change := comparison.Changes[RecordRateKey]
				if comparison.State != "observed" || change.State != "derived" || change.Value == nil || *change.Value != expected || change.Unit != "records/s" || change.Aggregation == nil || change.Aggregation.Family != "rate" || change.Aggregation.Kind != "records-per-second" || change.Aggregation.Method != "exact-rational" {
					t.Fatal(comparison, expected)
				}
				for i, side := range []*ViewDataset{comparison.Baseline, comparison.Selection} {
					cohort := []Cohort{a, b}[i]
					var retained Query
					if json.Unmarshal(side.Scope.Selector, &retained) != nil || !reflect.DeepEqual(retained, cohort.Question) || ValidateRecordRateDataset(side) != nil {
						t.Fatal("whole rate operand changed", side)
					}
					cell := side.Rows[0].Cells[RecordRateKey]
					if cell.State != "derived" || cell.Rate == nil || cell.Rate.Count != cohort.Count || cell.Rate.ElapsedMS != new(big.Int).SetInt64(cohort.Question.ToMS-cohort.Question.FromMS).String() || *side.Rows[0].Values[RecordRateKey] != reference(cohort.Count, cohort.Question.ToMS-cohort.Question.FromMS).RatString() {
						t.Fatal("rate operand lacks exact full-count exposure", side)
					}
					if side.Origin != "browser-held" || side.Quality.Extent != "unknown" || side.Quality.Sampling != "unknown" || side.Completeness != nil || side.SourceBasis != cohort.Basis || side.CapturedAt != cohort.CapturedAt {
						t.Fatal("caller claims became guarantees", side)
					}
				}
				if comparison.Baseline.Scope.ID == comparison.Selection.Scope.ID || comparison.Baseline.CapturedAt == comparison.Selection.CapturedAt {
					t.Fatal("independent scope or clocks flattened")
				}
				after, _ := json.Marshal(ComparisonRequest{Baseline: a, Selection: b})
				if string(before) != string(after) {
					t.Fatal("comparison mutated input")
				}
			})
		}
	}
}

func TestCapturedRateComparison418ZeroAndDiagnostic(t *testing.T) {
	for _, tc := range []struct{ a, b, delta string }{{"0", "0", "0"}, {"0", "120", "1"}, {"120", "0", "-2"}} {
		a, b := rateCohort418("logs", tc.a, 1000, 61000), rateCohort418("logs", tc.b, 100000, 220000)
		result, err := CompareCohorts(ComparisonRequest{Baseline: a, Selection: b})
		if err != nil || result.ViewComparison == nil {
			t.Fatal(result, err)
		}
		c := result.ViewComparison
		v := c.Changes[RecordRateKey]
		if c.State != "no_data" || v.State != "derived" || v.Value == nil || *v.Value != tc.delta {
			t.Fatal("known zero became unavailable", c)
		}
	}
	for _, side := range []string{"baseline", "selection", "both"} {
		a, b := rateCohort418("logs", "120", 1000, 61000), rateCohort418("logs", "240", 100000, 220000)
		if side == "baseline" || side == "both" {
			a.State = "changed_between_reads"
		}
		if side == "selection" || side == "both" {
			b.State = "changed_between_reads"
		}
		result, err := CompareCohorts(ComparisonRequest{Baseline: a, Selection: b})
		if err != nil || result.ViewComparison == nil {
			t.Fatal(result, err)
		}
		v := result.ViewComparison.Changes[RecordRateKey]
		if result.ViewComparison.State != "diagnostic" || v.State != "unavailable" || v.Value != nil || v.Reason != "source-reads-disagree" {
			t.Fatal("diagnostic rate delta invented", result.ViewComparison)
		}
	}
}

func TestRecordRateSummary418DoesNotExpandIntervalsOrSpecializedProfiles(t *testing.T) {
	q := rateCohort418("spans", "1", 1000, 61000).Question
	if !RecordRateSummaryQuery(q) || RecordRateQuery(q) {
		t.Fatal("summary grouping changed interval admission")
	}
	q.Summary = "span-duration"
	if !RecordRateSummaryQuery(q) {
		t.Fatal("ordinary duration context lost population rate")
	}
	for _, mutate := range []func(*Query){
		func(q *Query) { q.GroupBy = append(q.GroupBy, Field{Kind: "attribute", Key: "operation"}) },
		func(q *Query) { q.SpanStatus = true },
		func(q *Query) { b := numericBinding(); q.Numeric = &b },
		func(q *Query) { q.GroupBy = []Field{{Kind: "column", Key: "duration_bucket"}} },
		func(q *Query) { q.GroupTimeSeries = true; q.BucketMS = 1000 },
	} {
		bad := detached(q)
		mutate(&bad)
		if RecordRateSummaryQuery(bad) {
			t.Fatal("unsupported summary rate profile", bad)
		}
	}
	q.Summary = ""
	q.GroupBy = []Field{{Kind: "column", Key: "body_pattern_v1"}}
	q.Signal = "logs"
	if RecordRateSummaryQuery(q) {
		t.Fatal("pattern profile silently generalized")
	}
}

func TestCapturedRateSummary418RefusesDetachedOrIntervalSupport(t *testing.T) {
	a, b := rateCohort418("logs", "120", 1000, 61000), rateCohort418("logs", "240", 100000, 220000)
	result, err := CompareCohorts(ComparisonRequest{Baseline: a, Selection: b})
	if err != nil || result.ViewComparison == nil {
		t.Fatal(result, err)
	}
	original := result.ViewComparison.Baseline
	for _, mutate := range []func(*ViewDataset){
		func(d *ViewDataset) { d.Rows[0].Cells[RecordRateKey].Rate.Count = "119" },
		func(d *ViewDataset) { d.Rows[0].Cells[RecordRateKey].Rate.ElapsedMS = "120000" },
		func(d *ViewDataset) { d.Rows[0].Values[RecordRateKey] = viewText("4/2") },
		func(d *ViewDataset) {
			for i := range d.Fields {
				if d.Fields[i].Key == RecordRateKey {
					d.Fields[i].Unit = "requests/s"
				}
			}
		},
		func(d *ViewDataset) {
			d.Shape = "time-series"
			d.Rows[0].Values["fromMs"] = viewText("1000")
			d.Rows[0].Values["toMs"] = viewText("61000")
		},
	} {
		bad := detached(*original)
		mutate(&bad)
		if ValidateRecordRateDataset(&bad) == nil {
			t.Fatal("detached rate proof or widened interval profile admitted")
		}
	}
}
