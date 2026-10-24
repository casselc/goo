package exploration

import (
	"fmt"
	"strings"
	"time"
)

// DurationBucket is a fixed half-open partition over captured span durations.
// These are record counts, not native histogram intervals or trace latencies.
var durationBounds = []uint64{1_000_000, 5_000_000, 10_000_000, 50_000_000, 100_000_000, 500_000_000, 1_000_000_000}
var durationLabels = []string{"[0,1) ms", "[1,5) ms", "[5,10) ms", "[10,50) ms", "[50,100) ms", "[100,500) ms", "[500,1000) ms", "[1000,+inf) ms"}

func DurationBucket(ns uint64) string {
	for i, bound := range durationBounds {
		if ns < bound {
			return durationLabels[i]
		}
	}
	return durationLabels[len(durationLabels)-1]
}
func durationBucketSQL() string {
	parts := []string{}
	for i, bound := range durationBounds {
		parts = append(parts, fmt.Sprintf("Duration < %d, '%s'", bound, durationLabels[i]))
	}
	return "multiIf(" + strings.Join(parts, ", ") + ", '" + durationLabels[len(durationLabels)-1] + "')"
}

// ValidateDurationDistribution checks the full raw-span bucket contract without
// executing a source or inferring groups from retained inspection records.
func ValidateDurationDistribution(r *Result) error {
	if r == nil || r.Question.Signal != "spans" || r.Question.Numeric != nil || len(r.Question.GroupBy) != 1 || r.Question.GroupBy[0] != (Field{Kind: "column", Key: "duration_bucket"}) || (r.State != "observed" && r.State != "no_data") {
		return fmt.Errorf("require a consistent raw-span duration distribution")
	}
	if _, e := time.Parse(time.RFC3339Nano, r.CapturedAt); e != nil {
		return fmt.Errorf("duration source-read time unavailable")
	}
	c, e := CohortFromResult(r)
	if e != nil {
		return e
	}
	for _, g := range c.Groups {
		known := false
		if g.Value != nil {
			for _, label := range durationLabels {
				if *g.Value == label {
					known = true
					break
				}
			}
		}
		if !known {
			return fmt.Errorf("unknown duration bucket")
		}
	}
	return nil
}
