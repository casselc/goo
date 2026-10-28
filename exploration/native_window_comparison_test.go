package exploration

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/casselc/goo/measurements"
)

func TestNativeWindowPairWithholdsAllChangesForOneIncompleteStream(t *testing.T) {
	raw, e := os.ReadFile("../config/measurements/checkout-calls.yaml")
	if e != nil {
		t.Fatal(e)
	}
	d, e := measurements.Read(string(raw))
	if e != nil {
		t.Fatal(e)
	}
	makeCapture := func(gap bool) *SavedCapture {
		q := measurements.Question{Definition: d, FromMS: 100000, ToMS: 102000}
		rows := []json.RawMessage{}
		for _, instance := range []string{"one", "two"} {
			start := "100000000000"
			if gap && instance == "two" {
				start = "100500000000"
			}
			row, _ := json.Marshal(map[string]any{"resource": map[string]string{"service.name": d.Source.Services[0], "service.instance.id": instance, "deployment.environment.name": d.Environment}, "scope": d.Source.Scope, "version": d.Source.Version, "schema_url": d.Source.SchemaURL, "attributes": d.Source.Attributes, "name": d.Source.Metric, "unit": d.Source.Unit, "start_ns": start, "end_ns": "102000000000", "value": "10", "temporality": 1, "monotonic": true})
			rows = append(rows, row)
		}
		result, e := measurements.Evaluate(q, rows, time.Unix(300, 0))
		if e != nil {
			t.Fatal(e)
		}
		return &SavedCapture{CaptureReceipt: CaptureReceipt{Version: 1, Title: "Complete pair", CapturedAt: "2026-10-19T00:00:00Z", Question: CaptureQuestion{Kind: "measurement", Measurement: &q}}, Measurement: result}
	}
	a, b := makeCapture(false), makeCapture(true)
	r, e := CompareEvidence(a, b)
	if e != nil || r.State != "unavailable" || r.CountDifference != "" || len(r.NativeWindows) != 2 {
		t.Fatal(r, e)
	}
	for _, w := range r.NativeWindows {
		if w.RateDifference != "" || w.State == "comparable" {
			t.Fatal("healthy subset leaked a comparison", r)
		}
	}
}
