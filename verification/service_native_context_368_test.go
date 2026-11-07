package verification

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"hegel.dev/go/hegel"
)

// Native service context shares held-export admission with captured origins,
// without borrowing a capture receipt. Cross-runtime plans preserve integer
// boundary and reported resource authority; arbitrary metric labels never become
// record selectors. Mounted lease schedules are covered by the UI owner tests.
func TestHegelServiceNativeContext368(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		offset := hegel.Draw(tc, hegel.Integers(0, 999999))
		width := hegel.Draw(tc, hegel.Integers(1, 2000001))
		nsMode := hegel.Draw(tc, hegel.Integers(0, 2))
		actions := make([]int, 20)
		for i := range actions {
			actions[i] = hegel.Draw(tc, hegel.Integers(0, 4))
		}
		var ns *identity.Namespace
		if nsMode > 0 {
			ns = &identity.Namespace{Mode: "value"}
			if nsMode == 2 {
				ns.Value = "team-π"
			}
		}
		p, s := nrcInputs(ns, "histogram")
		p.Definition.Environment = ""
		p.Definition.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: []string{"current-only", "effective"}[hegel.Draw(tc, hegel.Integers(0, 1))], Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}}
		p.Definition.Source.Unit = "s"
		s.Unit = "s"
		from := int64(1790000000000000000) + int64(offset)
		to := from + int64(width)
		a, b := strconv.FormatInt(from, 10), strconv.FormatInt(to, 10)
		r := measurements.Result{Question: p, Basis: "service-native-export", ReadAt: "2026-11-07T10:00:00Z", Histograms: &metrics.HistogramResult{Series: []metrics.HistogramSeries{{Stream: s, Intervals: []metrics.HistogramInterval{{FromNS: a, ToNS: b, State: "ok"}}}}}}
		cases := []any{}
		for _, signal := range []string{"logs", "spans"} {
			plan, err := exploration.PlanNativeRecordContext(p, s, signal, a, b)
			if err != nil {
				tc.Fatal(err)
			}
			for _, at := range []int64{from - 1, from, from + 1, to - 1, to, to + 1} {
				values := map[exploration.Field]string{{Kind: "column", Key: "service"}: "api", {Kind: "resource", Key: "service.instance.id"}: "instance-one", {Kind: "column", Key: "timestamp_ns"}: strconv.FormatInt(at, 10)}
				if ns != nil {
					values[exploration.Field{Kind: "resource", Key: "service.namespace"}] = ns.Value
				}
				if nrcPredicate(plan.Question.Where, values) != (at >= from && at < to) {
					tc.Fatal("independent exact integer membership")
				}
				values[exploration.Field{Kind: "resource", Key: "service.instance.id"}] = "foreign"
				if nrcPredicate(plan.Question.Where, values) {
					tc.Fatal("foreign instance admitted")
				}
			}
			cases = append(cases, map[string]any{"measurement": r, "detail": map[string]any{"question": plan.Question, "fromNS": a, "toNS": b, "grain": plan.Grain, "measurement": p, "stream": s, "basis": r.Basis, "readAt": r.ReadAt}})
		}
		encoded, err := json.Marshal(map[string]any{"cases": cases, "actions": actions})
		if err != nil {
			tc.Fatal(err)
		}
		cmd := exec.Command("node", "--experimental-test-module-mocks", "service_native_context_368_fixture.mjs")
		cmd.Stdin = bytes.NewReader(encoded)
		var diagnostics bytes.Buffer
		cmd.Stderr = &diagnostics
		out, err := cmd.Output()
		if err != nil {
			tc.Fatalf("service held export cross-runtime admission: %v\n%s\n%s", err, out, diagnostics.String())
		}
		var result struct {
			Complete bool
			Checks   int
		}
		if json.Unmarshal(out, &result) != nil || !result.Complete || result.Checks < 24 {
			tc.Fatalf("incomplete service admission property: %s", out)
		}
	}, options(40)...)
}
