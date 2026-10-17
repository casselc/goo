package metrics

import (
	"fmt"
	"testing"
)

func ratioFixture(ok, err, pan string) CounterResult {
	r := CounterResult{Basis: "counter-basis", FromNS: "10", ToNS: "20"}
	for i, o := range []string{"ok", "error", "panic"} {
		s := Stream{Resource: map[string]string{"service.name": "test", "service.instance.id": "one"}, Scope: "goo", Version: "1", SchemaURL: "goo://v1", Name: "calls", Unit: "{invocation}", Attributes: map[string]string{"goo.operation.name": "checkout", "goo.operation.role": "server", "goo.operation.outcome": o}}
		r.Series = append(r.Series, CounterSeries{Stream: s, Temporality: "cumulative", Intervals: []CounterInterval{{FromNS: "10", ToNS: "20", State: "ok", Delta: []string{ok, err, pan}[i]}}})
	}
	return r
}
func ratioPolicy() RatioPolicy {
	return RatioPolicy{[]string{"ok", "error", "panic"}, []string{"error", "panic"}}
}
func TestErrorRatioAlignmentAndExactness(t *testing.T) {
	for _, tt := range []struct {
		ok, err, pan, calls, errors string
		state                       string
	}{{"8", "1", "1", "10", "2", "ok"}, {"0", "0", "0", "0", "0", "no_data"}, {"18446744073709551615", "18446744073709551615", "0", "36893488147419103230", "18446744073709551615", "ok"}} {
		r, e := EvaluateErrorRatios(ratioFixture(tt.ok, tt.err, tt.pan), ratioPolicy())
		if e != nil {
			t.Fatal(e)
		}
		i := r.Series[0].Intervals[0]
		if i.State != tt.state || i.Calls != tt.calls || i.Errors != tt.errors {
			t.Fatal(i)
		}
		if i.State == "no_data" && i.Percent != nil {
			t.Fatal("zero traffic divided")
		}
	}
	for _, mode := range []string{"missing", "unknown", "unaligned", "identity", "unit", "temporality", "duplicate", "overlap", "invalid-count"} {
		t.Run(mode, func(t *testing.T) {
			r := ratioFixture("8", "1", "1")
			reject := false
			switch mode {
			case "missing":
				r.Series = r.Series[:2]
			case "unknown":
				r.Series[1].Intervals[0] = CounterInterval{FromNS: "10", ToNS: "20", State: "unknown", Reason: "counter-reset"}
			case "unaligned":
				r.Series[1].Intervals[0].FromNS = "11"
			case "identity":
				r.Series[1].Stream.Resource["service.instance.id"] = "two"
			case "unit":
				r.Series[1].Stream.Unit = "s"
				reject = true
			case "temporality":
				r.Series[1].Temporality = "delta"
				reject = true
			case "duplicate":
				r.Series = append(r.Series, r.Series[0])
				reject = true
			case "overlap":
				r.Series[0].Intervals = append(r.Series[0].Intervals, r.Series[0].Intervals[0])
				reject = true
			case "invalid-count":
				r.Series[0].Intervals[0].Delta = "01"
				reject = true
			}
			out, e := EvaluateErrorRatios(r, ratioPolicy())
			if reject {
				if e == nil {
					t.Fatal("accepted invalid input")
				}
				return
			}
			if e != nil {
				t.Fatal(e)
			}
			for _, s := range out.Series {
				for _, i := range s.Intervals {
					if i.State != "unknown" || i.Percent != nil || i.Calls != "" {
						t.Fatalf("invented ratio: %+v", i)
					}
				}
			}
		})
	}
}
func TestErrorRatioBasisAndPolicy(t *testing.T) {
	r := ratioFixture("8", "1", "1")
	a, e := EvaluateErrorRatios(r, ratioPolicy())
	if e != nil {
		t.Fatal(e)
	}
	r.Series[0], r.Series[2] = r.Series[2], r.Series[0]
	b, e := EvaluateErrorRatios(r, ratioPolicy())
	if e != nil || a.Basis != b.Basis {
		t.Fatal("order-dependent", e)
	}
	r.Series[0].Intervals[0].Delta = "2"
	c, e := EvaluateErrorRatios(r, ratioPolicy())
	if e != nil || c.Basis == a.Basis {
		t.Fatal("basis omitted evidence", e)
	}
	if a.Series[0].Intervals[0].Fraction != "2/10" {
		t.Fatal("earlier result mutated")
	}
	for _, p := range []RatioPolicy{{nil, []string{"error"}}, {[]string{"ok"}, []string{"error"}}, {[]string{"ok", "error"}, []string{"error", "error"}}} {
		if e := ValidateRatioPolicy(p); e == nil {
			t.Fatal(fmt.Sprint(p))
		}
	}
}

func TestErrorRatioIdleAndTrafficWindow(t *testing.T) {
	input := ratioFixture("8", "1", "1")
	for n := range input.Series {
		i := input.Series[n].Intervals[0]
		i.FromNS = "15"
		input.Series[n].Intervals = []CounterInterval{{FromNS: "10", ToNS: "15", State: "ok", Delta: "0"}, i}
	}
	r, e := EvaluateErrorRatios(input, ratioPolicy())
	if e != nil {
		t.Fatal(e)
	}
	s := r.Series[0]
	if s.State != "ok" || s.Intervals[0].State != "no_data" || s.Intervals[0].Percent != nil || s.Intervals[1].Fraction != "2/10" {
		t.Fatal(s)
	}
}

func TestErrorRatioPolicyCanExcludePanics(t *testing.T) {
	r, e := EvaluateErrorRatios(ratioFixture("8", "1", "1"), RatioPolicy{Outcomes: []string{"ok", "error", "panic"}, Errors: []string{"error"}})
	if e != nil {
		t.Fatal(e)
	}
	i := r.Series[0].Intervals[0]
	if i.Fraction != "1/10" || i.Percent == nil || *i.Percent != 10 {
		t.Fatal(i)
	}
	if e := ValidateRatioPolicy(RatioPolicy{Outcomes: []string{"ok", "error", "panic"}, Errors: []string{"ok"}}); e == nil {
		t.Fatal("ok classified as error")
	}
}
