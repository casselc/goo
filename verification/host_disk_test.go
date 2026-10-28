package verification

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"hegel.dev/go/hegel"
	"strconv"
	"testing"
	"time"
)

func TestHegelDiskByteAdmission(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		const limit int64 = 9007199254740992
		base := int64(hegel.Draw(tc, hegel.Integers(0, 200)))
		if hegel.Draw(tc, hegel.Integers(0, 1)) == 1 {
			base += limit - 100
		}
		delta := int64(hegel.Draw(tc, hegel.Integers(0, 200)))
		fraction := hegel.Draw(tc, hegel.Integers(0, 2)) == 1
		s := metrics.Stream{Resource: map[string]string{"host.id": "host", "deployment.environment.name": "prod"}, Scope: "hostmetrics", SchemaURL: measurements.HostMetricsSchema, Name: "system.disk.io", Unit: "By", Attributes: map[string]string{"system.device": "sda", "disk.io.direction": "read"}}
		q := measurements.Question{Definition: measurements.Definition{APIVersion: measurements.Version, Name: "disk", Title: "Disk", Kind: "counter", CounterDomain: metrics.DecimalCounterDomain, Environment: "prod", LookbackMinutes: 2, MaxGapMS: 60000, Grain: "per-stream", Source: measurements.Binding{Resource: &measurements.ResourceBinding{Kind: "host", ID: "host"}, Services: []string{}, Scope: s.Scope, SchemaURL: s.SchemaURL, Metric: s.Name, Unit: s.Unit, ExactStream: &s}}, FromMS: 1000, ToMS: 3000}
		rows := []json.RawMessage{}
		for i, value := range []int64{base, base + delta} {
			text := strconv.FormatInt(value, 10)
			if fraction {
				text += ".000000000000000001"
			}
			b, _ := json.Marshal(map[string]any{"resource": s.Resource, "attributes": s.Attributes, "resource_schema_url": "", "scope": s.Scope, "version": "", "schema_url": s.SchemaURL, "name": s.Name, "unit": s.Unit, "flags": 0, "temporality": 2, "monotonic": true, "start_ns": "0", "end_ns": []string{"1000000000", "3000000000"}[i], "value": text})
			rows = append(rows, b)
		}
		r, e := measurements.Evaluate(q, rows, time.Unix(3, 0))
		valid := base < limit && base+delta < limit && !fraction
		if valid {
			if e != nil || r.Counters.Series[0].Intervals[0].Delta != fmt.Sprint(delta) || r.Counters.Series[0].Intervals[0].PerSecond == nil || *r.Counters.Series[0].Intervals[0].PerSecond != float64(delta)/2 {
				tc.Errorf("byte delta %v %v", r, e)
			}
		} else if e == nil {
			tc.Errorf("unsupported byte evidence accepted")
		}
	}, options(8)...)
}
