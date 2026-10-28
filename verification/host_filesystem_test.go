package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"hegel.dev/go/hegel"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestHegelFilesystemStockAdmission(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		first := hegel.Draw(tc, hegel.Integers(0, 100000))
		last := hegel.Draw(tc, hegel.Integers(0, 100000))
		bad := hegel.Draw(tc, hegel.Integers(0, 3))
		reverse := hegel.Draw(tc, hegel.Integers(0, 1)) == 1
		s := metrics.Stream{Resource: map[string]string{"host.id": "host", "deployment.environment.name": "prod"}, Scope: "hostmetrics", SchemaURL: measurements.HostMetricsSchema, Name: "system.filesystem.usage", Unit: "By", Attributes: map[string]string{"system.device": "/dev/sda", "system.filesystem.mountpoint": "/data", "system.filesystem.state": "used"}}
		q := measurements.Question{Definition: measurements.Definition{APIVersion: measurements.Version, Name: "filesystem", Title: "Filesystem", Kind: "stock", StockDomain: "nonnegative-integer", Environment: "prod", LookbackMinutes: 2, MaxGapMS: 60000, Grain: "per-stream", Source: measurements.Binding{Resource: &measurements.ResourceBinding{Kind: "host", ID: "host"}, Services: []string{}, Scope: s.Scope, SchemaURL: s.SchemaURL, Metric: s.Name, Unit: s.Unit, ExactStream: &s}}, FromMS: 1000, ToMS: 3000}
		rows := []json.RawMessage{}
		for i, v := range []int{first, last} {
			text := fmt.Sprint(v)
			if bad == 1 {
				text += ".000000000000000001"
			}
			if bad == 2 {
				text = "9007199254740992"
			}
			raw, _ := json.Marshal(map[string]any{"resource": s.Resource, "attributes": s.Attributes, "resource_schema_url": "", "scope": s.Scope, "version": "", "schema_url": s.SchemaURL, "name": s.Name, "unit": s.Unit, "flags": 0, "temporality": 2, "monotonic": false, "start_ns": "0", "end_ns": []string{"1000000000", "2000000000"}[i], "value": text})
			rows = append(rows, raw)
		}
		if reverse {
			rows[0], rows[1] = rows[1], rows[0]
		}
		r, e := measurements.Evaluate(q, rows, time.Unix(3, 0))
		if bad == 1 || bad == 2 {
			if e == nil {
				tc.Errorf("unsafe/fractional stock admitted")
			}
			return
		}
		if e != nil || len(r.Stocks.Series) != 1 || *r.Stocks.Series[0].Samples[0].Value != float64(first) || *r.Stocks.Series[0].Samples[1].Value != float64(last) {
			tc.Errorf("stock order/decrease changed: %v %v", r, e)
		}
	}, options(8)...)
}
func TestHegelFilesystemOwnerSchedules(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		input, _ := json.Marshal(map[string]any{"first": hegel.Draw(tc, hegel.Integers(0, 1)) == 1, "direction": hegel.Draw(tc, hegel.Integers(0, 1)), "mode": hegel.Draw(tc, hegel.Integers(0, 3))})
		cmd := exec.Command("node", "../scripts/check_network_schedule.js")
		cmd.Env = append(os.Environ(), "GOO_SCHEDULE_PROFILE=filesystem")
		cmd.Stdin = bytes.NewReader(input)
		if output, e := cmd.CombinedOutput(); e != nil {
			tc.Errorf("filesystem owner: %v %s", e, output)
		}
	}, options(8)...)
}
