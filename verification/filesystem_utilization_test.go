package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"hegel.dev/go/hegel"
	"os/exec"
	"testing"
	"time"
)

func TestHegelFilesystemFractionAdmission(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		n := hegel.Draw(tc, hegel.Integers(0, 1000))
		bad := hegel.Draw(tc, hegel.Integers(0, 2))
		value := fmt.Sprintf("%.3f", float64(n)/1000)
		if bad == 1 {
			value = "1.00000000000000000001"
		}
		if bad == 2 {
			value = "-0.00000000000000000001"
		}
		s := metrics.Stream{Resource: map[string]string{"host.id": "host", "deployment.environment.name": "prod"}, Scope: "hostmetrics", SchemaURL: measurements.HostMetricsSchema, Name: "system.filesystem.utilization", Unit: "1", Attributes: map[string]string{"system.device": "/dev/sda", "system.filesystem.mountpoint": "/data", "system.filesystem.state": "used"}}
		q := measurements.Question{Definition: measurements.Definition{APIVersion: measurements.Version, Name: "fraction", Title: "Reported fraction", Kind: "gauge", Environment: "prod", LookbackMinutes: 2, MaxGapMS: 60000, Grain: "per-stream", Source: measurements.Binding{Resource: &measurements.ResourceBinding{Kind: "host", ID: "host"}, Services: []string{}, Scope: s.Scope, SchemaURL: s.SchemaURL, Metric: s.Name, Unit: s.Unit, ExactStream: &s}}, FromMS: 1000, ToMS: 3000}
		raw, _ := json.Marshal(map[string]any{"resource": s.Resource, "attributes": s.Attributes, "resource_schema_url": "", "scope": s.Scope, "version": "", "schema_url": s.SchemaURL, "name": s.Name, "unit": s.Unit, "flags": 0, "end_ns": "2000000000", "value": value})
		r, e := measurements.Evaluate(q, []json.RawMessage{raw}, time.Unix(3, 0))
		if bad > 0 {
			if e == nil {
				tc.Errorf("out-of-range original fraction admitted")
			}
			return
		}
		if e != nil || r.Gauges.Series[0].Samples[0].Value == nil || *r.Gauges.Series[0].Samples[0].Value != float64(n)/1000 {
			tc.Errorf("reported fraction changed %v %v", r, e)
		}
	}, options(8)...)
}
func TestHegelFilesystemFractionOwnerSchedules(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		input, _ := json.Marshal(map[string]any{"first": hegel.Draw(tc, hegel.Integers(0, 1)) == 1, "mode": hegel.Draw(tc, hegel.Integers(0, 4))})
		cmd := exec.Command("node", "../scripts/check_filesystem_utilization_schedule.js")
		cmd.Stdin = bytes.NewReader(input)
		if out, e := cmd.CombinedOutput(); e != nil {
			tc.Errorf("fraction owner %s %v %s", input, e, out)
		}
	}, options(8)...)
}
