package verification

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/qtrace"
	"github.com/casselc/observability/quintgo/validate"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"testing"
	"time"
)

func ctRaw(f psMap, r *measurements.Result) bool {
	return hostRateRaw(f, r, "system.cpu.time", "s", map[string]string{"cpu.logical_number": "0", "cpu.mode": "user"})
}

// Independent finite-fixture oracle shared by CPU-time and network-byte journeys.
func hostRateRaw(f psMap, r *measurements.Result, name, unit string, attributes map[string]string) bool {
	d := r.Question.Definition
	s := d.Source.ExactStream
	if s == nil || d.Kind != "counter" || d.CounterDomain != metrics.DecimalCounterDomain || d.Environment != f["environment"] || d.Source.Resource == nil || d.Source.Resource.Kind != "host" || d.Source.Resource.ID != "cpu-host" || s.Name != name || s.Unit != unit || s.SchemaURL != measurements.HostMetricsSchema || len(d.Source.Services) != 0 || d.Source.Namespace != nil || !reflect.DeepEqual(s.Attributes, attributes) || !chBasis(r) || r.Counters == nil || r.Gauges != nil || r.Stocks != nil || r.Histograms != nil {
		return false
	}
	expected := []psMap{}
	for _, v := range psArr(f["points"]) {
		p := psObj(v)
		if p["metric"] == s.Name && p["unit"] == s.Unit && p["scope"] == s.Scope && p["version"] == s.Version && p["schemaURL"] == s.SchemaURL && p["resourceSchemaURL"] == s.ResourceSchemaURL && aeJSONEqual(p["resource"], s.Resource) && aeJSONEqual(p["attributes"], s.Attributes) {
			expected = append(expected, p)
		}
	}
	sort.Slice(expected, func(i, j int) bool { return expected[i]["endNS"].(string) < expected[j]["endNS"].(string) })
	if len(expected) != 7 || len(r.Rows) != 7 || r.Count != "7" || len(r.Counters.Series) != 1 || r.Counters.ValueDomain != metrics.DecimalCounterDomain {
		return false
	}
	canonical := []metrics.CounterPoint{}
	for i, raw := range r.Rows {
		var row psMap
		if json.Unmarshal(raw, &row) != nil {
			return false
		}
		p := expected[i]
		for actual, want := range map[string]string{"resource_schema_url": "resourceSchemaURL", "schema_url": "schemaURL", "scope": "scope", "version": "version", "name": "metric", "unit": "unit", "start_ns": "startNS", "end_ns": "endNS"} {
			if row[actual] != p[want] {
				return false
			}
		}
		a, ok := new(big.Rat).SetString(fmt.Sprint(row["value"]))
		b, ok2 := new(big.Rat).SetString(p["value"].(string))
		if !ok || !ok2 || a.Cmp(b) != 0 || !aeJSONEqual(row["resource"], p["resource"]) || !aeJSONEqual(row["attributes"], p["attributes"]) || row["flags"] != float64(0) || row["temporality"] != float64(2) || row["monotonic"] != true {
			return false
		}
		canonical = append(canonical, metrics.CounterPoint{Stream: *s, Temporality: "cumulative", Monotonic: true, StartNS: p["startNS"].(string), EndNS: p["endNS"].(string), Value: p["value"].(string)})
	}
	lo, hi := r.Question.BoundsNS()
	encoded, _ := json.Marshal(metrics.CounterRequest{ValueDomain: metrics.DecimalCounterDomain, FromNS: lo, ToNS: hi, MaxGapNS: fmt.Sprint(d.MaxGapMS * 1000000), Points: canonical})
	hash := sha256.Sum256(encoded)
	if r.Counters.Basis != hex.EncodeToString(hash[:]) || r.Counters.FromNS != lo || r.Counters.ToNS != hi {
		return false
	}
	series := r.Counters.Series[0]
	if !reflect.DeepEqual(series.Stream, *s) || series.Temporality != "cumulative" || series.State != "unknown" {
		return false
	}
	// This finite fixture deliberately fixes six interval cases. Expected values
	// come from independent raw decimals, not the production counter evaluator.
	reasons := []string{"", "", "counter-decrease", "invalid-counter-epoch", "counter-epoch-change", ""}
	lower, _ := strconv.ParseInt(lo, 10, 64)
	upper, _ := strconv.ParseInt(hi, 10, 64)
	expectedIndex := []int{}
	for i := 0; i < len(expected)-1; i++ {
		left, _ := strconv.ParseInt(expected[i]["endNS"].(string), 10, 64)
		right, _ := strconv.ParseInt(expected[i+1]["endNS"].(string), 10, 64)
		if right > lower && left < upper {
			expectedIndex = append(expectedIndex, i)
		}
	}
	if len(series.Intervals) != len(expectedIndex) {
		return false
	}
	for j, p := range series.Intervals {
		i := expectedIndex[j]
		a, z := expected[i], expected[i+1]
		left, _ := strconv.ParseInt(a["endNS"].(string), 10, 64)
		right, _ := strconv.ParseInt(z["endNS"].(string), 10, 64)
		wantFrom, wantTo := left, right
		if left < lower {
			wantFrom = lower
			reasons[i] = "boundary-crossing"
		}
		if right > upper {
			wantTo = upper
			reasons[i] = "boundary-crossing"
		}
		if p.FromNS != strconv.FormatInt(wantFrom, 10) || p.ToNS != strconv.FormatInt(wantTo, 10) || p.Reason != reasons[i] {
			return false
		}
		if reasons[i] != "" {
			if p.State != "unknown" || p.Delta != "" || p.PerSecond != nil || p.ExactPerSecond != "" {
				return false
			}
			continue
		}
		av, _ := new(big.Rat).SetString(a["value"].(string))
		zv, _ := new(big.Rat).SetString(z["value"].(string))
		delta := new(big.Rat).Sub(zv, av)
		actual, ok := new(big.Rat).SetString(p.Delta)
		if !ok || actual.Cmp(delta) != 0 || p.State != "ok" || p.PerSecond == nil {
			return false
		}
		start, _ := strconv.ParseInt(a["endNS"].(string), 10, 64)
		end, _ := strconv.ParseInt(z["endNS"].(string), 10, 64)
		rate := new(big.Rat).Mul(delta, big.NewRat(1000000000, end-start))
		value, _ := rate.Float64()
		if p.ExactPerSecond != rate.RatString() || *p.PerSecond != value {
			return false
		}
	}
	return true
}
func hostRateContexts(f psMap, contexts []any) bool {
	if len(contexts) != 2 {
		return false
	}
	seen := map[string]bool{}
	lo := psObj(f["interval"])["fromNS"].(string)
	hi := psObj(f["interval"])["toNS"].(string)
	lower, _ := strconv.ParseInt(lo, 10, 64)
	upper, _ := strconv.ParseInt(hi, 10, 64)
	for _, v := range contexts {
		v := psObj(v)
		signal := fmt.Sprint(v["signal"])
		if signal != "logs" && signal != "spans" || seen[signal] {
			return false
		}
		seen[signal] = true
		var q exploration.Query
		var result exploration.Result
		if !rsDecode(v["question"], &q) || !rsDecode(v["result"], &result) {
			return false
		}
		want := exploration.Query{Signal: signal, Environment: f["environment"].(string), FromMS: lower / 1000000, ToMS: (upper + 999999) / 1000000, Limit: 100, Where: &exploration.Predicate{All: []exploration.Predicate{{Field: &exploration.Field{Kind: "resource", Key: "host.id"}, Op: "eq", Value: "cpu-host"}, {Field: &exploration.Field{Kind: "column", Key: "timestamp_ns"}, Op: "gte", Value: lo}, {Field: &exploration.Field{Kind: "column", Key: "timestamp_ns"}, Op: "lte", Value: fmt.Sprint(upper - 1)}}}}
		if !aeJSONEqual(q, want) || !nhLiveSameQuery(q, result.Question) || result.Count != "3" || len(result.Rows) != 3 {
			return false
		}
		expected := map[string]psMap{}
		for _, v := range psArr(f["records"]) {
			p := psObj(v)
			at, _ := strconv.ParseInt(p["timeNS"].(string), 10, 64)
			if p["signal"] == signal && at >= lower && at < upper && psObj(p["resource"])["host.id"] == "cpu-host" {
				expected[p["id"].(string)] = p
			}
		}
		if len(expected) != 3 {
			return false
		}
		for _, raw := range result.Rows {
			var row psMap
			if json.Unmarshal(raw, &row) != nil {
				return false
			}
			attrs := psObj(row["Attributes"])
			if attrs == nil {
				attrs = psObj(row["LogAttributes"])
			}
			if attrs == nil {
				attrs = psObj(row["SpanAttributes"])
			}
			id := fmt.Sprint(attrs["work.id"])
			p := expected[id]
			if p == nil || !aeJSONEqual(row["ResourceAttributes"], p["resource"]) || !aeJSONEqual(attrs, p["attributes"]) || row["timestamp_ns"] != p["timeNS"] || row["TraceId"] != p["traceId"] || row["SpanId"] != p["spanId"] {
				return false
			}
			delete(expected, id)
		}
		if len(expected) != 0 {
			return false
		}
	}
	return true
}
func ctOK(d psMap) bool {
	f := psObj(d["facts"])
	defaults := psArr(d["defaults"])
	if len(defaults) != 2 {
		return false
	}
	counterSeen := 0
	for _, v := range defaults {
		var r measurements.Result
		if !rsDecode(v, &r) {
			return false
		}
		if r.Counters != nil {
			counterSeen++
			if !ctRaw(f, &r) {
				return false
			}
		} else if !hrRaw(f, &r) {
			return false
		}
	}
	if counterSeen != 1 {
		return false
	}
	var individual measurements.Result
	if !rsDecode(d["individual"], &individual) || !ctRaw(f, &individual) {
		return false
	}
	var page exploration.CapturePage
	if !rsDecode(d["saved"], &page) || page.Measurement == nil || !ctRaw(f, page.Measurement) || !peLiveCaptureHash(page) {
		return false
	}
	if !hostRateContexts(f, psArr(d["context"])) {
		return false
	}
	sources, hostPlans, nativeReads, captures := 0, 0, 0, 0
	for _, v := range psArr(d["requests"]) {
		r := psObj(v)
		if r["method"] != "POST" {
			continue
		}
		if r["effect"] == "source" {
			sources++
		}
		if r["endpoint"] == "/api/host-dashboard/plan" {
			hostPlans++
		}
		if r["endpoint"] == "/api/measurements" {
			nativeReads++
		}
		if r["endpoint"] == "/api/captures" {
			captures++
		}
	}
	if sources != 7 || hostPlans != 1 || nativeReads != 3 || captures != 1 {
		return false
	}
	events := psArr(d["events"])
	if len(events) != 5 {
		return false
	}
	for i, e := range events {
		if psObj(e)["source"] != float64(i+3) || psObj(e)["plans"] != float64(1) {
			return false
		}
	}
	return len(psArr(d["blocked"])) == 0
}
func TestCPUTimeObserved(t *testing.T) {
	d, raw, path := rsRead(t, "GOO_CPU_TIME_LEDGER")
	if !ctOK(d) {
		t.Fatal("CPU-time raw/context/read/capture oracle")
	}
	a, z := psObj(d["storedBefore"]), psObj(d["storedAfter"])
	before, e := os.ReadFile(filepath.Join(filepath.Dir(path), a["file"].(string)))
	if e != nil {
		t.Fatal(e)
	}
	after, e := os.ReadFile(filepath.Join(filepath.Dir(path), z["file"].(string)))
	hash := sha256.Sum256(before)
	var stored exploration.SavedCapture
	var page exploration.CapturePage
	if e != nil || !bytes.Equal(before, after) || len(before) != int(a["bytes"].(float64)) || hex.EncodeToString(hash[:]) != a["sha256"] || a["sha256"] != z["sha256"] || json.Unmarshal(before, &stored) != nil || !rsDecode(d["saved"], &page) || !reflect.DeepEqual(stored.CaptureReceipt, page.Receipt) || !aeJSONEqual(stored.Measurement, page.Measurement) {
		t.Fatal("full immutable capture bytes")
	}
	changes := []func(psMap){func(x psMap) { psObj(psArr(x["events"])[0])["source"] = float64(4) }, func(x psMap) {
		x["requests"] = append(psArr(x["requests"]), psMap{"method": "POST", "endpoint": "/api/measurements", "effect": "source"})
	}, func(x psMap) { psObj(x["individual"])["basis"] = "bad" }, func(x psMap) {
		p := psObj(psArr(psObj(psArr(psObj(psObj(x["individual"])["counters"])["series"])[0])["intervals"])[0])
		p["delta"] = "2"
	}, func(x psMap) { psObj(psObj(psArr(x["context"])[0])["result"])["count"] = "2" }}
	changes = append(changes, func(x psMap) {
		row := psObj(psArr(psObj(psObj(psArr(x["context"])[0])["result"])["rows"])[0])
		row["timestamp_ns"] = "1"
	}, func(x psMap) {
		p := psObj(psArr(psObj(psArr(psObj(psObj(x["individual"])["counters"])["series"])[0])["intervals"])[2])
		p["perSecond"] = float64(1)
	})
	for i, change := range changes {
		var x psMap
		_ = json.Unmarshal(raw, &x)
		change(x)
		if ctOK(x) {
			t.Fatal("accepted changed control", i)
		}
	}
	t.Log("seven raw/effect/context controls rejected; full capture bytes preserved")
}

func ctTrace(d psMap) *qobs.MemorySink {
	s := &qobs.MemorySink{}
	for i, v := range psArr(d["events"]) {
		e := psObj(v)
		s.Record(context.Background(), &qtrace.Step{Seq: uint64(i + 1), Recorder: "cpu-time-browser", Outcome: "ok", Action: []string{"defaults", "native", "logs", "spans", "capture"}[i], Obs: map[string]any{"source": e["source"], "plans": e["plans"]}})
	}
	return s
}
func TestQuintCPUTimeObserved(t *testing.T) {
	d, _, path := rsRead(t, "GOO_CPU_TIME_LEDGER")
	if !ctOK(d) {
		t.Fatal("raw precondition")
	}
	checker, e := validate.NewChecker("model/cpu_time.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	check := func(name string, s *qobs.MemorySink) bool {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		report, e := checker.Check(ctx, s.Steps())
		if e != nil {
			t.Fatal(e)
		}
		b, _ := json.MarshalIndent(s.Steps(), "", "  ")
		_ = os.WriteFile(filepath.Join(filepath.Dir(path), name+".json"), b, 0644)
		_ = os.WriteFile(filepath.Join(filepath.Dir(path), name+".txt"), []byte(report.Summary()), 0644)
		return report.OK()
	}
	if !check("cpu-time-observed", ctTrace(d)) {
		t.Fatal("actual read protocol")
	}
	bad := ctTrace(d)
	bad.Steps()[1].Obs["source"] = float64(5)
	if check("cpu-time-bad-extra-read", bad) {
		t.Fatal("accepted extra read")
	}
	t.Log("one actual browser trace / five macrosteps; one bad read trace rejected")
}
