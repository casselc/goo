package exploration

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func TestDurationPartitionBoundaries(t *testing.T) {
	cases := map[uint64]string{0: "[0,1) ms", 999999: "[0,1) ms", 1000000: "[1,5) ms", 4999999: "[1,5) ms", 5000000: "[5,10) ms", 10000000: "[10,50) ms", 50000000: "[50,100) ms", 100000000: "[100,500) ms", 500000000: "[500,1000) ms", 1000000000: "[1000,+inf) ms", math.MaxUint64: "[1000,+inf) ms"}
	for n, want := range cases {
		if got := DurationBucket(n); got != want {
			t.Fatalf("%d: %s != %s", n, got, want)
		}
	}
	q := basic()
	q.GroupBy = []Field{{"column", "duration_bucket"}}
	q.Where = &Predicate{Field: &q.GroupBy[0], Op: "eq", Value: "[1,5) ms"}
	p, e := Compile(q)
	if e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(p.Groups, "multiIf(Duration < 1000000") || !strings.Contains(p.Groups, "Duration < 1000000000") || strings.Contains(p.Groups, "OFFSET") || !strings.Contains(p.Rows, "AS duration_bucket") {
		t.Fatal("bucket query lost population partition", p)
	}
	q.Signal = "logs"
	if _, e = Compile(q); e == nil {
		t.Fatal("log duration accepted")
	}
}

func TestDurationDistributionContract(t *testing.T) {
	r := Result{Question: Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100, GroupBy: []Field{{Kind: "column", Key: "duration_bucket"}}}, State: "observed", Count: "1", Basis: "unit-fixture", CapturedAt: "2026-10-24T00:00:01Z", Groups: []json.RawMessage{json.RawMessage(`{"g0":"[0,1) ms","count":"1"}`)}}
	if e := ValidateDurationDistribution(&r); e != nil {
		t.Fatal(e)
	}
	for _, mutate := range []func(*Result){func(r *Result) {
		r.Groups = []json.RawMessage{json.RawMessage(`{"g0":"not a duration bucket","count":"1"}`)}
	}, func(r *Result) { r.Groups = []json.RawMessage{json.RawMessage(`{"g0":null,"count":"1"}`)} }, func(r *Result) { r.Count = "2" }, func(r *Result) { r.State = "changed_between_reads" }, func(r *Result) { r.CapturedAt = "" }, func(r *Result) { r.Question.Signal = "logs" }} {
		bad := r
		mutate(&bad)
		if ValidateDurationDistribution(&bad) == nil {
			t.Fatal("unsupported baseline distribution accepted", bad)
		}
	}
	r.State = "no_data"
	r.Count = "0"
	r.Groups = nil
	if e := ValidateDurationDistribution(&r); e != nil {
		t.Fatal("empty baseline must remain explicit no-data", e)
	}
}
