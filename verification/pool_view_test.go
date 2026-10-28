package verification

import (
	"encoding/json"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"hegel.dev/go/hegel"
)

// pvStock uses the real adapter as input construction, not as the expected
// ratio oracle. Each expectation below comes directly from raw supplied facts.
func pvStock(q measurements.Question, s metrics.Stream, value int, start, end int64, flags int) (*measurements.Result, error) {
	row := map[string]any{"resource": s.Resource, "resource_schema_url": s.ResourceSchemaURL, "scope": s.Scope, "version": s.Version, "schema_url": s.SchemaURL, "name": s.Name, "unit": s.Unit, "attributes": s.Attributes, "start_ns": strconv.FormatInt(start, 10), "end_ns": strconv.FormatInt(end, 10), "value": strconv.Itoa(value), "flags": flags, "temporality": 2, "monotonic": false}
	b, _ := json.Marshal(row)
	return measurements.Evaluate(q, []json.RawMessage{b}, time.Unix(200, 0))
}
func TestHegelPoolViewReportedAlignment(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		used := hegel.Draw(tc, hegel.Integers(0, 100))
		max := hegel.Draw(tc, hegel.Integers(0, 100))
		base := psFixture("generated")
		plan, e := measurements.PlanPoolSignals(base)
		if e != nil {
			tc.Errorf("setup %v", e)
			return
		}
		cq := plan.Capacity[0].Question
		for _, mode := range []string{"aligned", "different-time", "different-epoch", "unknown-epoch", "both-unknown-epoch", "missing-used"} {
			start, capStart, end, capEnd := int64(90000000000), int64(90000000000), int64(101000000000), int64(101000000000)
			flags := 0
			switch mode {
			case "different-time":
				capEnd++
			case "different-epoch":
				capStart++
			case "unknown-epoch":
				start = 0
			case "both-unknown-epoch":
				start = 0
				capStart = 0
			case "missing-used":
				flags = 1
			}
			count, err := pvStock(base.Parent, base.Stream, used, start, end, flags)
			if err != nil {
				tc.Errorf("count setup %v", err)
				return
			}
			capacity, err := pvStock(cq, *cq.Definition.Source.ExactStream, max, capStart, capEnd, 0)
			if err != nil {
				tc.Errorf("capacity setup %v", err)
				return
			}
			req := measurements.PoolViewRequest{Count: *count, Stream: base.Stream, Capacity: capacity}
			before, _ := json.Marshal(req)
			v, err := measurements.ComposePoolView(req)
			if err != nil {
				tc.Errorf("compose %s %v", mode, err)
				return
			}
			state, reason := "observed", "matching-reported-observation-and-epoch"
			switch mode {
			case "missing-used":
				state, reason = "unavailable", "used-unavailable"
			case "different-time":
				state, reason = "unavailable", "observation-time-mismatch"
			case "different-epoch":
				state, reason = "unavailable", "epoch-mismatch"
			case "unknown-epoch", "both-unknown-epoch":
				state, reason = "unavailable", "epoch-not-reported"
			default:
				if max == 0 {
					state, reason = "unavailable", "zero-capacity"
				} else if used > max {
					state, reason = "inconsistent", "used-exceeds-capacity"
				}
			}
			if v.Utilization.State != state || v.Utilization.Reason != reason {
				tc.Errorf("raw oracle used%d/max%d mode%s got%+v", used, max, mode, v.Utilization)
				return
			}
			if state == "observed" {
				ratio := new(big.Rat).SetFrac(big.NewInt(int64(used)), big.NewInt(int64(max)))
				if v.Utilization.Numerator != ratio.Num().String() || v.Utilization.Denominator != ratio.Denom().String() {
					tc.Errorf("exact rational mismatch")
					return
				}
			} else if v.Utilization.Numerator != "" || v.Utilization.Denominator != "" {
				tc.Errorf("unavailable ratio leaked")
				return
			}
			if v.CountEvidence.Basis != count.Basis || v.CountEvidence.ReadAt != count.ReadAt || v.Capacity.Evidence.Basis != capacity.Basis || v.Capacity.Evidence.ReadAt != capacity.ReadAt {
				tc.Errorf("independent provenance altered")
				return
			}
			v.Stream.Resource["service.instance.id"] = "mutated"
			v.States[0].Stream.Attributes["db.client.connection.state"] = "mutated"
			after, _ := json.Marshal(req)
			if string(before) != string(after) {
				tc.Errorf("input alias")
				return
			}
		}
	}, options(12)...)
}
func TestPoolViewIndependentIdentityRefusal(t *testing.T) {
	base := psFixture("identity")
	plan, e := measurements.PlanPoolSignals(base)
	if e != nil {
		t.Fatal(e)
	}
	count, e := pvStock(base.Parent, base.Stream, 3, 90000000000, 101000000000, 0)
	if e != nil {
		t.Fatal(e)
	}
	for _, edit := range []func(*metrics.Stream){func(s *metrics.Stream) { s.Resource["service.instance.id"] = "other" }, func(s *metrics.Stream) { s.ResourceSchemaURL = "other" }, func(s *metrics.Stream) { s.Attributes["extra"] = "x" }} {
		q := plan.Capacity[0].Question
		raw, _ := json.Marshal(q)
		_ = json.Unmarshal(raw, &q)
		edit(q.Definition.Source.ExactStream)
		capacity, e := pvStock(q, *q.Definition.Source.ExactStream, 10, 90000000000, 101000000000, 0)
		if e != nil {
			t.Fatal(e)
		}
		if _, e := measurements.ComposePoolView(measurements.PoolViewRequest{Count: *count, Stream: base.Stream, Capacity: capacity}); e == nil {
			t.Fatal("different identity accepted")
		}
	}
}

func TestPoolViewExactBoundaryTable(t *testing.T) {
	base := psFixture("boundary")
	plan, e := measurements.PlanPoolSignals(base)
	if e != nil {
		t.Fatal(e)
	}
	q := plan.Capacity[0].Question
	for _, c := range []struct {
		used, max     int
		state, reason string
	}{{0, 0, "unavailable", "zero-capacity"}, {0, 1, "observed", "matching-reported-observation-and-epoch"}, {1, 1, "observed", "matching-reported-observation-and-epoch"}, {2, 1, "inconsistent", "used-exceeds-capacity"}, {9007199254740990, 9007199254740991, "observed", "matching-reported-observation-and-epoch"}} {
		count, e := pvStock(base.Parent, base.Stream, c.used, 90000000000, 101000000000, 0)
		if e != nil {
			t.Fatal(e)
		}
		cap, e := pvStock(q, *q.Definition.Source.ExactStream, c.max, 90000000000, 101000000000, 0)
		if e != nil {
			t.Fatal(e)
		}
		v, e := measurements.ComposePoolView(measurements.PoolViewRequest{Count: *count, Stream: base.Stream, Capacity: cap})
		if e != nil || v.Utilization.State != c.state || v.Utilization.Reason != c.reason {
			t.Fatal(c, v, e)
		}
		if c.state == "observed" {
			r := new(big.Rat).SetFrac(big.NewInt(int64(c.used)), big.NewInt(int64(c.max)))
			if v.Utilization.Numerator != r.Num().String() || v.Utilization.Denominator != r.Denom().String() {
				t.Fatal("boundary rational")
			}
		}
	}
}
