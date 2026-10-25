package verification

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"hegel.dev/go/hegel"
)

// Adapter/replay integration, rather than another stock-kernel permutation test.
func TestHegelStockAdapterIdentityAndEpochs(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		shift := hegel.Draw(tc, hegel.Integers(-1000, 1000))
		scope := exploration.InventoryQuestion{Environment: "prod", FromMS: 100000, ToMS: 160000}
		instrument := hdInstrument("stock-level", "ordinary.connections", "stock", "{connection}", scope, "api")
		q := measurements.Question{Definition: *instrument.Definition, FromMS: scope.FromMS, ToMS: scope.ToMS}
		rows := []json.RawMessage{}
		expected := map[string]map[string]struct {
			value float64
			epoch string
		}{}
		for ns := 0; ns < 3; ns++ {
			for instance := 0; instance < 2; instance++ {
				resource := map[string]string{"deployment.environment.name": "prod", "service.name": "api", "service.instance.id": fmt.Sprintf("worker-%d", instance)}
				if ns > 0 {
					resource["service.namespace"] = []string{"", "", "commerce"}[ns]
				}
				key := strconv.Itoa(ns) + ":" + strconv.Itoa(instance)
				expected[key] = map[string]struct {
					value float64
					epoch string
				}{}
				for i := 0; i < 6; i++ {
					at := int64(110000 + i*1000)
					epoch := int64(90000)
					if i >= 3 {
						epoch = 112500
					}
					value := float64(shift-i-instance) / 4
					row := map[string]any{"resource": resource, "scope": instrument.Scope, "version": instrument.Version, "schema_url": instrument.SchemaURL, "name": instrument.Metric, "unit": instrument.Unit, "attributes": map[string]string{"label": "separate"}, "start_ns": strconv.FormatInt(epoch*1000000, 10), "end_ns": strconv.FormatInt(at*1000000, 10), "value": strconv.FormatFloat(value, 'f', -1, 64), "temporality": 2, "monotonic": false, "flags": 0}
					raw, _ := json.Marshal(row)
					rows = append(rows, raw)
					expected[key][strconv.FormatInt(at*1000000, 10)] = struct {
						value float64
						epoch string
					}{value, strconv.FormatInt(epoch*1000000, 10)}
				}
			}
		}
		r, e := measurements.Evaluate(q, rows, time.Unix(200, 0), "", nil)
		if e != nil || r.Stocks == nil || r.Gauges != nil || r.Counters != nil || len(r.Stocks.Series) != 6 {
			tc.Errorf("stock adapter identity %v", e)
			return
		}
		for _, s := range r.Stocks.Series {
			ns := 0
			if v, ok := s.Stream.Resource["service.namespace"]; ok {
				ns = 1
				if v != "" {
					ns = 2
				}
			}
			key := strconv.Itoa(ns) + ":" + s.Stream.Resource["service.instance.id"][7:]
			samples := expected[key]
			if len(samples) != len(s.Samples) {
				tc.Errorf("merged/dropped stream")
				return
			}
			for _, sample := range s.Samples {
				want, ok := samples[sample.AtNS]
				if !ok || sample.State != "observed" || sample.Value == nil || *sample.Value != want.value || sample.StartNS != want.epoch {
					tc.Errorf("physical sample/epoch changed")
					return
				}
			}
		}
		encoded, _ := json.Marshal(r)
		var replay measurements.Result
		json.Unmarshal(encoded, &replay)
		if measurements.ValidateResult(&replay) != nil {
			tc.Errorf("valid stock replay refused")
			return
		}
		replay.Stocks.Series[0].Samples[0].StartNS = "0"
		if measurements.ValidateResult(&replay) == nil {
			tc.Errorf("tampered accumulation epoch admitted")
		}
	}, options(12)...)
}
