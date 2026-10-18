package exploration

import (
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
