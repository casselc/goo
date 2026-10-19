package verification

import (
	"encoding/json"
	"github.com/casselc/goo/metrics"
	"hegel.dev/go/hegel"
	"strconv"
	"testing"
)

func TestHegelWindowDiscoverySupport(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		offset := hegel.Draw(tc, hegel.Integers(1, 999999))
		missing := hegel.Draw(tc, hegel.Integers(0, 7))
		streams := hegel.Draw(tc, hegel.Integers(1, 4))
		points := []metrics.CounterPoint{}
		base := int64(1790000000000000000)
		ns := func(n int64) string { return strconv.FormatInt(n, 10) }
		for s := 0; s < streams; s++ {
			for i := 0; i < 8; i++ {
				points = append(points, metrics.CounterPoint{Stream: metrics.Stream{Resource: map[string]string{"service.name": "search", "service.instance.id": strconv.Itoa(s)}, Scope: "otel.http", Name: "calls", Unit: "{request}"}, Temporality: "delta", Monotonic: true, StartNS: ns(base + int64(offset) + int64(i)*1000000000), EndNS: ns(base + int64(offset) + int64(i+1)*1000000000), Value: "1", NoRecordedValue: s == streams-1 && i == missing})
			}
		}
		r, e := metrics.EvaluateCounters(metrics.CounterRequest{FromNS: ns(base), ToNS: ns(base + 9000000000), MaxGapNS: "2000000000", Points: points})
		if e != nil {
			tc.Fatal(e)
		}
		before, _ := json.Marshal(r)
		d := metrics.DiscoverCounterWindows(r)
		after, _ := json.Marshal(r)
		if string(before) != string(after) || len(d.Windows) == 0 {
			tc.Fatal("mutation or lost supported intervals")
		}
		for _, w := range d.Windows {
			if len(w.Members) != streams {
				tc.Fatal("member omitted")
			}
			check, e := metrics.EvaluateCounters(metrics.CounterRequest{FromNS: w.FromNS, ToNS: w.ToNS, MaxGapNS: "2000000000", Points: points})
			if e != nil {
				tc.Fatal(e)
			}
			for _, s := range check.Series {
				if s.State != "ok" {
					tc.Fatal("offered unsupported interval")
				}
			}
		}
		// A never-supported member must withhold every candidate, not get dropped.
		extra := r.Series[0]
		extra.Stream = metrics.Stream{Resource: map[string]string{"service.name": "search", "service.instance.id": "missing"}, Scope: "otel.http", Name: "calls", Unit: "{request}"}
		extra.Intervals = nil
		r.Series = append(r.Series, extra)
		if len(metrics.DiscoverCounterWindows(r).Windows) != 0 {
			tc.Fatal("silently dropped unsupported stream")
		}
	}, options(100)...)
}
