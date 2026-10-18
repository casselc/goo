package exploration

import (
	"fmt"
	"strings"
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
