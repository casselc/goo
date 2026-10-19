package exploration

import (
	"encoding/json"
	"github.com/casselc/goo/measurements"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func comparisonFixture(count, state string) *SavedCapture {
	q := Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 100}
	return &SavedCapture{CaptureReceipt: CaptureReceipt{Version: 1, Title: "Stored spans", CapturedAt: "2026-10-19T00:00:00Z", Question: CaptureQuestion{Kind: "explore", Explore: &q}}, Explore: &Result{Question: q, Count: count, State: state, Basis: strings.Repeat("a", 64)}}
}
func TestEvidenceComparisonCompatibilityAndExactCounts(t *testing.T) {
	a, b := comparisonFixture("18446744073709551614", "observed"), comparisonFixture("18446744073709551615", "observed")
	before, _ := json.Marshal([]*SavedCapture{a, b})
	r, e := CompareEvidence(a, b)
	after, _ := json.Marshal([]*SavedCapture{a, b})
	if e != nil || r.CountDifference != "1" || r.State != "comparable" || string(before) != string(after) {
		t.Fatal(r, e)
	}
	b.Question.Explore.ToMS++
	b.Explore.Question.ToMS++
	r, e = CompareEvidence(a, b)
	if e != nil || r.State != "comparable" || !reflect.DeepEqual(r.Differences, []string{"Fixed window differs"}) {
		t.Fatal(r, e)
	}
	b.Question.Explore.Environment = "staging"
	b.Explore.Question.Environment = "staging"
	r, e = CompareEvidence(a, b)
	if e != nil || r.State != "incompatible" || r.CountDifference != "" || r.Measures != nil {
		t.Fatal(r, e)
	}
	b = comparisonFixture("0", "no_data")
	r, e = CompareEvidence(a, b)
	if e != nil || r.State != "unavailable" || r.CountDifference != "" {
		t.Fatal(r, e)
	}
	b = comparisonFixture("9", "changed_between_reads")
	r, e = CompareEvidence(a, b)
	if e != nil || r.State != "unavailable" || r.CountDifference != "" {
		t.Fatal(r, e)
	}
}
func TestEvidenceComparisonUsesFullMeasuresNeverRows(t *testing.T) {
	a, b := comparisonFixture("2", "observed"), comparisonFixture("4", "observed")
	for _, c := range []*SavedCapture{a, b} {
		c.Question.Explore.Summary = "span-duration"
		c.Question.Explore.SpanStatus = true
		c.Explore.Question = *c.Question.Explore
		c.Explore.Duration = &DurationSummary{State: "observed", Count: c.Explore.Count, SumNS: "200", MinNS: "0", MaxNS: "200", P50NS: "0", P95NS: "200", P99NS: "200"}
		c.Explore.SpanStatus = &SpanStatusSummary{Count: c.Explore.Count, Error: "1", OK: "0", Unset: "0", Unknown: signedDifference("1", c.Explore.Count)}
		c.Explore.Rows = []json.RawMessage{json.RawMessage(`{"duration_ns":"9999999999"}`)}
	}
	r, e := CompareEvidence(a, b)
	if e != nil || r.Measures == nil || r.Measures.MeanDifferenceNS != "-50.000" || r.Measures.ErrorDifferencePP != "-25.00" {
		t.Fatal(r, e)
	}
}

func TestGaugeEvidencePreservesStreamIdentityUnitsAndUnknownEndpoints(t *testing.T) {
	raw, e := os.ReadFile("../config/measurements/worker-active.yaml")
	if e != nil {
		t.Fatal(e)
	}
	d, e := measurements.Read(string(raw))
	if e != nil {
		t.Fatal(e)
	}
	makeCapture := func(value, instance string, flags int, unit string) *SavedCapture {
		def := d
		def.Source.Unit = unit
		q := measurements.Question{Definition: def, FromMS: 100000, ToMS: 220000}
		row, _ := json.Marshal(map[string]any{"resource": map[string]string{"service.name": def.Source.Services[0], "service.instance.id": instance, "deployment.environment.name": def.Environment}, "scope": def.Source.Scope, "version": def.Source.Version, "schema_url": def.Source.SchemaURL, "attributes": def.Source.Attributes, "name": def.Source.Metric, "unit": unit, "end_ns": "210000000000", "value": value, "flags": flags})
		result, err := measurements.Evaluate(q, []json.RawMessage{row}, time.Unix(300, 0), "", nil)
		if err != nil {
			t.Fatal(err)
		}
		return &SavedCapture{CaptureReceipt: CaptureReceipt{Version: 1, Title: "Native evidence", CapturedAt: "2026-10-19T00:00:00Z", Question: CaptureQuestion{Kind: "measurement", Measurement: &q}}, Measurement: result}
	}
	a, b := makeCapture("0.1", "worker", 0, d.Source.Unit), makeCapture("0.2", "worker", 0, d.Source.Unit)
	r, e := CompareEvidence(a, b)
	if e != nil || len(r.Gauges) != 1 || r.Gauges[0].Difference != "0.1" {
		t.Fatal(r, e)
	}
	other := makeCapture("0.2", "other", 0, d.Source.Unit)
	r, e = CompareEvidence(a, other)
	if e != nil || len(r.Gauges) != 2 || r.Gauges[0].Difference != "" || r.Gauges[1].Difference != "" {
		t.Fatal(r, e)
	}
	bad := makeCapture("0.2", "worker", 0, "ms")
	r, e = CompareEvidence(a, bad)
	if e != nil || r.State != "incompatible" || len(r.Gauges) != 0 {
		t.Fatal(r, e)
	}
	unknown := makeCapture("0.2", "worker", 1, d.Source.Unit)
	r, e = CompareEvidence(a, unknown)
	if e != nil || r.State != "unavailable" || len(r.Gauges) != 0 {
		t.Fatal(r, e)
	}
}
