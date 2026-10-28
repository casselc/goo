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

// Expected fractions/freshness are reconstructed from independent raw OTLP facts.
func huRaw(f psMap, r *measurements.Result) bool {
	d := r.Question.Definition
	s := d.Source.ExactStream
	attrs := map[string]string{"system.device": "/dev/sda", "system.filesystem.mountpoint": "/data", "system.filesystem.mode": "rw", "system.filesystem.type": "ext4", "system.filesystem.state": "used"}
	if !measurements.HostFilesystemUtilizationDefinition(d) || s == nil || d.Source.Resource.ID != "cpu-host" || d.Environment != f["environment"] || len(d.Source.Services) != 0 || d.Source.Namespace != nil || s.Name != d.Source.Metric || s.Unit != "1" || s.SchemaURL != measurements.HostMetricsSchema || !reflect.DeepEqual(s.Attributes, attrs) || !chBasis(r) || r.Gauges == nil || r.Stocks != nil || r.Counters != nil || r.Histograms != nil {
		return false
	}
	lo, hi := r.Question.BoundsNS()
	if r.Gauges.FromNS != lo || r.Gauges.ToNS != hi || len(r.Gauges.Series) != 1 || !reflect.DeepEqual(r.Gauges.Series[0].Stream, *s) {
		return false
	}
	expected := []psMap{}
	for _, v := range psArr(f["points"]) {
		p := psObj(v)
		if p["metric"] == s.Name && p["unit"] == s.Unit && p["scope"] == s.Scope && p["version"] == s.Version && p["schemaURL"] == s.SchemaURL && p["resourceSchemaURL"] == s.ResourceSchemaURL && aeJSONEqual(p["resource"], s.Resource) && aeJSONEqual(p["attributes"], s.Attributes) && p["endNS"].(string) >= lo && p["endNS"].(string) < hi {
			expected = append(expected, p)
		}
	}
	sort.Slice(expected, func(i, j int) bool { return expected[i]["endNS"].(string) < expected[j]["endNS"].(string) })
	if len(expected) != 3 || len(r.Rows) != 3 || r.Count != "3" || len(r.Gauges.Series[0].Samples) != 3 {
		return false
	}
	upper, _ := strconv.ParseInt(hi, 10, 64)
	for i, p := range expected {
		var row psMap
		if json.Unmarshal(r.Rows[i], &row) != nil {
			return false
		}
		for actual, want := range map[string]string{"resource_schema_url": "resourceSchemaURL", "schema_url": "schemaURL", "scope": "scope", "version": "version", "name": "metric", "unit": "unit", "end_ns": "endNS"} {
			if row[actual] != p[want] {
				return false
			}
		}
		a, ok := new(big.Rat).SetString(fmt.Sprint(row["value"]))
		b, ok2 := new(big.Rat).SetString(p["value"].(string))
		if !ok || !ok2 || a.Cmp(b) != 0 || a.Sign() < 0 || a.Cmp(big.NewRat(1, 1)) > 0 || row["flags"] != float64(0) || !aeJSONEqual(row["resource"], p["resource"]) || !aeJSONEqual(row["attributes"], p["attributes"]) {
			return false
		}
		sample := r.Gauges.Series[0].Samples[i]
		at, _ := strconv.ParseInt(p["endNS"].(string), 10, 64)
		state, reason := "observed", ""
		if upper-at > d.MaxGapMS*1000000 {
			state = "stale"
			reason = "older than declared freshness"
		}
		value, _ := b.Float64()
		if sample.AtNS != p["endNS"] || sample.State != state || sample.Reason != reason || sample.Value == nil || *sample.Value != value {
			return false
		}
	}
	last := r.Gauges.Series[0].Samples[2].State
	outer := "observed"
	if last != "observed" {
		outer = "unknown"
	}
	return r.Gauges.Series[0].State == last && r.State == outer
}
func huOK(d psMap) bool {
	f := psObj(d["facts"])
	results := psArr(d["results"])
	if len(results) != 1 {
		return false
	}
	var result, individual measurements.Result
	var saved exploration.CapturePage
	if !rsDecode(results[0], &result) || !huRaw(f, &result) || !rsDecode(d["individual"], &individual) || !huRaw(f, &individual) || !rsDecode(d["saved"], &saved) || saved.Measurement == nil || !huRaw(f, saved.Measurement) || !peLiveCaptureHash(saved) || len(psArr(d["blocked"])) != 0 {
		return false
	}
	plot := psObj(d["plot"])
	if !aeJSONEqual(plot["domain"], psMap{"min": "0", "max": "1"}) || plot["unit"] != "1" || plot["kind"] != "gauge" {
		return false
	}
	points := psArr(plot["points"])
	samples := result.Gauges.Series[0].Samples
	if len(points) != 3 {
		return false
	}
	for i, v := range points {
		p := psObj(v)
		value, e := strconv.ParseFloat(fmt.Sprint(p["value"]), 64)
		if e != nil || samples[i].Value == nil || value != *samples[i].Value || p["state"] != samples[i].State || p["fromNS"] != samples[i].AtNS || p["toNS"] != samples[i].AtNS {
			return false
		}
	}
	if *samples[2].Value != .625 {
		return false
	}
	bytes := psArr(d["byteResults"])
	if len(bytes) != 2 {
		return false
	}
	seen := map[string]bool{}
	for _, v := range bytes {
		var r measurements.Result
		if !rsDecode(v, &r) || !hfRaw(f, &r, r.Question.Definition.Source.Metric) || seen[r.Question.Definition.Source.Metric] {
			return false
		}
		seen[r.Question.Definition.Source.Metric] = true
	}
	if !seen["system.filesystem.usage"] || !seen["system.filesystem.limit"] {
		return false
	}
	plots := psArr(d["bytePlots"])
	if len(plots) != 2 {
		return false
	}
	for _, v := range plots {
		p := psObj(v)
		if p["unit"] != "By" || !aeJSONEqual(p["domain"], psMap{"min": "0", "max": "134217728"}) {
			return false
		}
	}
	source, native, plans, captures := 0, 0, 0, 0
	for _, v := range psArr(d["requests"]) {
		r := psObj(v)
		if r["method"] != "POST" {
			continue
		}
		if r["effect"] == "source" {
			source++
		}
		switch r["endpoint"] {
		case "/api/measurements":
			native++
		case "/api/host-dashboard/plan":
			plans++
		case "/api/captures":
			captures++
		}
	}
	if source != 8 || native != 6 || plans != 1 || captures != 1 {
		return false
	}
	events := psArr(d["events"])
	counts := []int{3, 5, 6, 6, 6, 7, 8}
	if len(events) != 7 {
		return false
	}
	for i, v := range events {
		e := psObj(v)
		if e["source"] != float64(counts[i]) || e["plans"] != float64(1) {
			return false
		}
	}
	var plan measurements.HostDashboardPlan
	if !rsDecode(d["plan"], &plan) || len(plan.Widgets) != 2 {
		return false
	}
	for _, w := range plan.Widgets {
		if measurements.HostFilesystemUtilizationDefinition(w.Question.Definition) {
			return false
		}
	}
	return true
}
func TestFilesystemUtilizationObserved(t *testing.T) {
	d, raw, path := rsRead(t, "GOO_FILESYSTEM_UTILIZATION_LEDGER")
	if !huOK(d) {
		t.Fatal("filesystem raw/shared-scale/context/effect oracle")
	}
	before, after := psObj(d["storedBefore"]), psObj(d["storedAfter"])
	a, e := os.ReadFile(filepath.Join(filepath.Dir(path), before["file"].(string)))
	if e != nil {
		t.Fatal(e)
	}
	z, e := os.ReadFile(filepath.Join(filepath.Dir(path), after["file"].(string)))
	hash := sha256.Sum256(a)
	var stored exploration.SavedCapture
	var page exploration.CapturePage
	if e != nil || !bytes.Equal(a, z) || len(a) != int(before["bytes"].(float64)) || hex.EncodeToString(hash[:]) != before["sha256"] || before["sha256"] != after["sha256"] || json.Unmarshal(a, &stored) != nil || !rsDecode(d["saved"], &page) || !reflect.DeepEqual(stored.CaptureReceipt, page.Receipt) || !aeJSONEqual(stored.Measurement, page.Measurement) {
		t.Fatal("full immutable filesystem capture")
	}
	changes := []func(psMap){
		func(x psMap) { psObj(psArr(x["events"])[0])["source"] = float64(4) },
		func(x psMap) {
			x["requests"] = append(psArr(x["requests"]), psMap{"method": "POST", "effect": "source", "endpoint": "/api/measurements"})
		},
		func(x psMap) { psObj(x["plot"])["domain"] = psMap{"min": "0", "max": "134217728"} },
		func(x psMap) { psObj(psArr(x["bytePlots"])[0])["domain"] = psMap{"min": "0", "max": "1"} },
		func(x psMap) { psObj(x["individual"])["basis"] = "bad" },
		func(x psMap) {
			s := psObj(psArr(psObj(psObj(x["individual"])["gauges"])["series"])[0])
			psObj(psArr(s["samples"])[2])["value"] = .375
		},
		func(x psMap) { x["results"] = []any{} },
		func(x psMap) { psObj(x["individual"])["state"] = "unknown" },
		func(x psMap) {
			s := psObj(psArr(psObj(psObj(x["individual"])["gauges"])["series"])[0])
			psObj(psArr(s["samples"])[0])["reason"] = "fabricated"
		},
	}
	for i, change := range changes {
		var x psMap
		_ = json.Unmarshal(raw, &x)
		change(x)
		if huOK(x) {
			t.Fatal("accepted filesystem control", i)
		}
	}
	t.Log("nine changed controls rejected; reported fraction, separate byte scales and full capture verified")
}
func huTrace(d psMap) *qobs.MemorySink {
	s := &qobs.MemorySink{}
	actions := []string{"defaults", "bytes", "utilization", "local", "parent", "native", "capture"}
	for i, v := range psArr(d["events"]) {
		e := psObj(v)
		s.Record(context.Background(), &qtrace.Step{Seq: uint64(i + 1), Recorder: "filesystem-utilization", Outcome: "ok", Action: actions[i], Obs: map[string]any{"source": e["source"], "plans": e["plans"]}})
	}
	return s
}
func TestQuintFilesystemUtilizationObserved(t *testing.T) {
	d, _, path := rsRead(t, "GOO_FILESYSTEM_UTILIZATION_LEDGER")
	if !huOK(d) {
		t.Fatal("raw precondition")
	}
	checker, e := validate.NewChecker("model/filesystem_utilization.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
	if !check("filesystem-observed", huTrace(d)) {
		t.Fatal("actual filesystem effect trace")
	}
	bad := huTrace(d)
	bad.Steps()[3].Obs["source"] = float64(7)
	if check("filesystem-bad-reopen-read", bad) {
		t.Fatal("reopen read accepted")
	}
	t.Log("one actual seven-step trace; one bad local-read trace rejected")
}
