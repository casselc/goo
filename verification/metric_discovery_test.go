package verification

import (
	"github.com/casselc/goo/metrics"
	"hegel.dev/go/hegel"
	"testing"
)

func TestHegelOptionalMetricIdentity(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		version := []string{"", "1", "2"}[hegel.Draw(tc, hegel.Integers(0, 2))]
		schema := []string{"", "https://opentelemetry.io/schemas/1.27.0"}[hegel.Draw(tc, hegel.Integers(0, 1))]
		count := hegel.Draw(tc, hegel.Integers(0, 1000))
		stream := metrics.Stream{Resource: map[string]string{"service.name": "ordinary", "service.instance.id": "one"}, Scope: "ordinary", Version: version, SchemaURL: schema, Name: "jobs", Unit: "{job}", Attributes: map[string]string{"outcome": "ok"}}
		point := metrics.CounterPoint{Stream: stream, Temporality: "delta", Monotonic: true, StartNS: "1000000000", EndNS: "2000000000", Value: fmtInt(count)}
		r, e := metrics.EvaluateCounters(metrics.CounterRequest{FromNS: "1000000000", ToNS: "2000000000", MaxGapNS: "1000000000", Points: []metrics.CounterPoint{point}})
		if e != nil || len(r.Series) != 1 || r.Series[0].Stream.Version != version || r.Series[0].Stream.SchemaURL != schema {
			t.Fatal("optional identity changed", r, e)
		}
		other := point
		other.Stream.Version = version + "different"
		r, e = metrics.EvaluateCounters(metrics.CounterRequest{FromNS: "1000000000", ToNS: "2000000000", MaxGapNS: "1000000000", Points: []metrics.CounterPoint{point, other}})
		if e != nil || len(r.Series) != 2 {
			t.Fatal("distinct versions merged", r, e)
		}
	}, options(100)...)
}
