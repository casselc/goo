package measurements

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func poolViewFixture(t *testing.T) PoolViewRequest {
	t.Helper()
	req := poolSignalsFixture(t)
	plan, e := PlanPoolSignals(req)
	if e != nil {
		t.Fatal(e)
	}
	makeStock := func(q Question, attrs map[string]string, value string) Result {
		row := stockRow(q.Definition, 101000, value)
		row.Attributes = attrs
		row.Resource = req.Stream.Resource
		r, e := Evaluate(q, []json.RawMessage{raw(row)}, time.Unix(200, 0), "", nil)
		if e != nil {
			t.Fatal(e)
		}
		return *r
	}
	count := makeStock(req.Parent, req.Stream.Attributes, "3")
	capacity := makeStock(plan.Capacity[0].Question, plan.Capacity[0].Question.Definition.Source.ExactStream.Attributes, "10")
	wq := plan.Wait[0].Question
	wr := metricRow{Resource: req.Stream.Resource, Scope: wq.Definition.Source.Scope, Version: wq.Definition.Source.Version, SchemaURL: wq.Definition.Source.SchemaURL, Name: wq.Definition.Source.Metric, Unit: wq.Definition.Source.Unit, Attributes: wq.Definition.Source.ExactStream.Attributes, StartNS: wq.FromNS, EndNS: wq.ToNS, Temporality: 1, Bounds: []float64{0.01, 0.1}, BucketCounts: []string{"1", "1", "0"}, Count: "2"}
	wait, e := Evaluate(wq, []json.RawMessage{raw(wr)}, time.Unix(201, 0), "", nil)
	if e != nil {
		t.Fatal(e)
	}
	return PoolViewRequest{Count: count, Stream: req.Stream, Capacity: &capacity, Wait: wait}
}
func replacePoolRows(t *testing.T, r *Result, edit func([]metricRow) []metricRow) {
	t.Helper()
	rows := []metricRow{}
	for _, v := range r.Rows {
		var row metricRow
		if e := json.Unmarshal(v, &row); e != nil {
			t.Fatal(e)
		}
		rows = append(rows, row)
	}
	rows = edit(rows)
	encoded := []json.RawMessage{}
	for _, row := range rows {
		encoded = append(encoded, raw(row))
	}
	result, e := Evaluate(r.Question, encoded, time.Unix(202, 0), "", nil)
	if e != nil {
		t.Fatal(e)
	}
	*r = *result
}
func TestComposePoolViewExactRatioAndIndependentWait(t *testing.T) {
	req := poolViewFixture(t)
	before, _ := json.Marshal(req)
	view, e := ComposePoolView(req)
	if e != nil {
		t.Fatal(e)
	}
	if view.Utilization.State != "observed" || view.Utilization.Numerator != "3" || view.Utilization.Denominator != "10" || view.Wait == nil || view.Wait.Latest == nil || view.Wait.Latest.Distribution == nil {
		t.Fatal(view)
	}
	if !reflect.DeepEqual(view.CountQuestion, req.Count.Question) || view.CountEvidence.Basis != req.Count.Basis || view.Wait.Evidence.Basis != req.Wait.Basis {
		t.Fatal("provenance altered")
	}
	view.Stream.Resource["service.instance.id"] = "mutated"
	view.States[0].Stream.Attributes["db.client.connection.state"] = "mutated"
	after, _ := json.Marshal(req)
	if string(before) != string(after) {
		t.Fatal("aliased input")
	}
}
func TestComposePoolViewLatestAlignmentReasons(t *testing.T) {
	cases := []struct {
		name, reason string
		edit         func(*PoolViewRequest)
	}{
		{"time", "observation-time-mismatch", func(r *PoolViewRequest) {
			replacePoolRows(t, r.Capacity, func(rows []metricRow) []metricRow { rows[0].EndNS = ns(101001); return rows })
		}},
		{"epoch", "epoch-mismatch", func(r *PoolViewRequest) {
			replacePoolRows(t, r.Capacity, func(rows []metricRow) []metricRow { rows[0].StartNS = ns(95000); return rows })
		}},
		{"zero", "zero-capacity", func(r *PoolViewRequest) {
			replacePoolRows(t, r.Capacity, func(rows []metricRow) []metricRow { rows[0].Value = "0"; return rows })
		}},
		{"small", "used-exceeds-capacity", func(r *PoolViewRequest) {
			replacePoolRows(t, r.Capacity, func(rows []metricRow) []metricRow { rows[0].Value = "2"; return rows })
		}},
		{"stale", "capacity-unavailable", func(r *PoolViewRequest) {
			r.Capacity.Question.Definition.MaxGapMS = 1
			replacePoolRows(t, r.Capacity, func(rows []metricRow) []metricRow { return rows })
		}},
		{"unknown", "capacity-unavailable", func(r *PoolViewRequest) {
			replacePoolRows(t, r.Capacity, func(rows []metricRow) []metricRow { rows[0].Flags = 1; return rows })
		}},
		{"conflict", "capacity-unavailable", func(r *PoolViewRequest) {
			replacePoolRows(t, r.Capacity, func(rows []metricRow) []metricRow { other := rows[0]; other.Value = "11"; return append(rows, other) })
		}},
		{"no-data", "capacity-no-data", func(r *PoolViewRequest) {
			replacePoolRows(t, r.Capacity, func(rows []metricRow) []metricRow { return nil })
		}},
		{"not-supplied", "capacity-not-supplied", func(r *PoolViewRequest) { r.Capacity = nil }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := poolViewFixture(t)
			c.edit(&req)
			view, e := ComposePoolView(req)
			if e != nil || view.Utilization.Reason != c.reason || view.Utilization.Numerator != "" || view.Utilization.Denominator != "" {
				t.Fatal(view, e)
			}
		})
	}
}
func TestComposePoolViewCustomStateNoOlderFallbackAndRefusals(t *testing.T) {
	req := poolViewFixture(t)
	replacePoolRows(t, &req.Count, func(rows []metricRow) []metricRow {
		old := rows[0]
		latest := old
		latest.EndNS = ns(101100)
		latest.Flags = 1
		return append(rows, latest)
	})
	view, e := ComposePoolView(req)
	if e != nil || view.Utilization.Reason != "used-unavailable" || view.States[0].Latest.AtNS != ns(101100) {
		t.Fatal("older fallback", view, e)
	}
	req = poolViewFixture(t)
	replacePoolRows(t, &req.Count, func(rows []metricRow) []metricRow {
		rows[0].Attributes = map[string]string{"db.client.connection.pool.name": "primary", "db.client.connection.state": "custom"}
		return rows
	})
	req.Stream = req.Count.Stocks.Series[0].Stream
	view, e = ComposePoolView(req)
	if e != nil || view.Utilization.Reason != "used-not-reported" || view.States[0].Name != "custom" {
		t.Fatal(view, e)
	}
	req = poolViewFixture(t)
	req.Capacity.Question.Definition.Source.ExactStream = nil
	if _, e = ComposePoolView(req); e == nil {
		t.Fatal("unpinned result accepted")
	}
	req = poolViewFixture(t)
	req.Count.Stocks.Series[0].Samples[0].Value = nil
	if _, e = ComposePoolView(req); e == nil {
		t.Fatal("forged derived result accepted")
	}
	req = poolViewFixture(t)
	replacePoolRows(t, req.Capacity, func(rows []metricRow) []metricRow { rows[0].Value = "0"; return rows })
	replacePoolRows(t, &req.Count, func(rows []metricRow) []metricRow { rows[0].Value = "0"; return rows })
	view, e = ComposePoolView(req)
	if e != nil || view.Utilization.Reason != "zero-capacity" {
		t.Fatal("zero/zero inventedratio", view, e)
	}
}

func TestComposePoolViewRejectsValidWrongWindowAndIdentity(t *testing.T) {
	req := poolViewFixture(t)
	req.Capacity.Question.ToNS = "102001000123"
	req.Capacity.Question.ToMS = 102002
	replacePoolRows(t, req.Capacity, func(rows []metricRow) []metricRow { return rows })
	if _, e := ComposePoolView(req); e == nil {
		t.Fatal("different valid evaluation window accepted")
	}
	req = poolViewFixture(t)
	req.Capacity.Question.Definition.Source.ExactStream.Resource["service.instance.id"] = "other"
	replacePoolRows(t, req.Capacity, func(rows []metricRow) []metricRow { rows[0].Resource["service.instance.id"] = "other"; return rows })
	if _, e := ComposePoolView(req); e == nil {
		t.Fatal("different valid instance accepted")
	}
	req = poolViewFixture(t)
	replacePoolRows(t, &req.Count, func(rows []metricRow) []metricRow { rows[0].Value = "0"; return rows })
	view, e := ComposePoolView(req)
	if e != nil || view.Utilization.Numerator != "0" || view.Utilization.Denominator != "1" {
		t.Fatal("known zero not exact", view, e)
	}
}

func TestComposePoolViewZeroEpochIsUnreported(t *testing.T) {
	for _, both := range []bool{false, true} {
		req := poolViewFixture(t)
		replacePoolRows(t, req.Capacity, func(rows []metricRow) []metricRow { rows[0].StartNS = "0"; return rows })
		if both {
			replacePoolRows(t, &req.Count, func(rows []metricRow) []metricRow { rows[0].StartNS = "0"; return rows })
		}
		view, e := ComposePoolView(req)
		if e != nil || view.Utilization.State != "unavailable" || view.Utilization.Reason != "epoch-not-reported" || view.Utilization.Numerator != "" || view.Capacity.Latest.Value == nil || view.States[0].Latest.Value == nil {
			t.Fatal("unknown epoch joined or values hidden", view, e)
		}
	}
}
