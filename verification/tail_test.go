package verification

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
	"testing"
)

func TestHegelTailOverlapMultiplicityState(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		tracker := exploration.TailTracker{}
		max := 0
		total := 0
		stamp := int64(100)
		seen := map[string]bool{}
		rounds := hegel.Draw(tc, hegel.Integers(2, 30))
		for i := 0; i < rounds; i++ {
			n := hegel.Draw(tc, hegel.Integers(0, 20))
			raw := json.RawMessage(`{"timestamp_ns":"100","Body":"duplicate"}`)
			rows := make([]json.RawMessage, n)
			for j := range rows {
				rows[j] = raw
			}
			out, e := tracker.Observe(rows)
			if e != nil {
				tc.Fatal(e)
			}
			want := 0
			if n > max {
				want = n - max
				max = n
			}
			if len(out) != want {
				tc.Fatalf("overlap invented/lost copies: %d want %d", len(out), want)
			}
			for _, r := range out {
				if seen[r.Key] {
					tc.Fatal("replayed key")
				}
				seen[r.Key] = true
				total++
			}
			stamp--
			out, e = tracker.Observe([]json.RawMessage{json.RawMessage(fmt.Sprintf(`{"timestamp_ns":"%d","Body":"late-%d"}`, stamp, i))})
			if e != nil || len(out) != 1 || out[0].Late != (i > 0 || max > 0) {
				tc.Fatal("late event withheld", e)
			}
		}
		if total != max {
			tc.Fatal("multiplicity conservation")
		}
	}, options(100)...)
}
