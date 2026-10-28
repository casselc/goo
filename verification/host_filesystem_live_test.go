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

// This oracle compares actual stock results to separately recorded raw OTLP facts.
// It does not call the production stock evaluator to construct expected samples.
func hfRaw(f psMap, r *measurements.Result, metric string) bool {
	d := r.Question.Definition
	s := d.Source.ExactStream
	if s == nil || !measurements.HostFilesystemDefinition(d) || s.Name != metric || s.Unit != "By" || s.SchemaURL != measurements.HostMetricsSchema || d.Environment != f["environment"] || d.Source.Resource.ID != "cpu-host" || len(d.Source.Services) != 0 || d.Source.Namespace != nil || !chBasis(r) || r.Stocks == nil || r.Gauges != nil || r.Counters != nil || r.Histograms != nil {
		return false
	}
	expectedAttrs := map[string]string{"system.device": "/dev/sda", "system.filesystem.mountpoint": "/data", "system.filesystem.mode": "rw", "system.filesystem.type": "ext4"}
	if metric == "system.filesystem.usage" {
		expectedAttrs["system.filesystem.state"] = "used"
	}
	if !reflect.DeepEqual(s.Attributes, expectedAttrs) {
		return false
	}
	lo, hi := r.Question.BoundsNS()
	if r.Stocks.FromNS != lo || r.Stocks.ToNS != hi || len(r.Stocks.Series) != 1 || !reflect.DeepEqual(r.Stocks.Series[0].Stream, *s) {
		return false
	}
	expected := []psMap{}
	for _, v := range psArr(f["points"]) {
		p := psObj(v)
		if p["metric"] == metric && p["unit"] == s.Unit && p["scope"] == s.Scope && p["version"] == s.Version && p["schemaURL"] == s.SchemaURL && p["resourceSchemaURL"] == s.ResourceSchemaURL && aeJSONEqual(p["resource"], s.Resource) && aeJSONEqual(p["attributes"], s.Attributes) && p["endNS"].(string) >= lo && p["endNS"].(string) < hi {
			expected = append(expected, p)
		}
	}
	sort.Slice(expected, func(i, j int) bool { return expected[i]["endNS"].(string) < expected[j]["endNS"].(string) })
	if len(expected) == 0 || len(r.Rows) != len(expected) || r.Count != strconv.Itoa(len(expected)) || len(r.Stocks.Series[0].Samples) != len(expected) {
		return false
	}
	upper, _ := strconv.ParseInt(hi, 10, 64)
	for i, p := range expected {
		var row psMap
		if json.Unmarshal(r.Rows[i], &row) != nil {
			return false
		}
		for actual, want := range map[string]string{"resource_schema_url": "resourceSchemaURL", "schema_url": "schemaURL", "scope": "scope", "version": "version", "name": "metric", "unit": "unit", "start_ns": "startNS", "end_ns": "endNS"} {
			if row[actual] != p[want] {
				return false
			}
		}
		value, ok := new(big.Rat).SetString(fmt.Sprint(row["value"]))
		want, ok2 := new(big.Rat).SetString(p["value"].(string))
		if !ok || !ok2 || value.Cmp(want) != 0 || !value.IsInt() || value.Sign() < 0 || value.Num().BitLen() > 53 || !aeJSONEqual(row["resource"], p["resource"]) || !aeJSONEqual(row["attributes"], p["attributes"]) || row["flags"] != p["flags"] || row["temporality"] != float64(2) || row["monotonic"] != false {
			return false
		}
		sample := r.Stocks.Series[0].Samples[i]
		at, _ := strconv.ParseInt(p["endNS"].(string), 10, 64)
		state, reason := "observed", ""
		if upper-at > d.MaxGapMS*1000000 {
			state = "stale"
			reason = "older than declared freshness"
		}
		wantFloat, _ := want.Float64()
		if sample.StartNS != p["startNS"] || sample.AtNS != p["endNS"] || sample.State != state || sample.Reason != reason || sample.Value == nil || *sample.Value != wantFloat {
			return false
		}
	}
	latest := r.Stocks.Series[0].Samples[len(expected)-1].State
	outer := "observed"
	if latest != "observed" {
		outer = "unknown"
	}
	if r.Stocks.Series[0].State != latest || r.State != outer {
		return false
	}

	return true
}
func hfContexts(d psMap) bool {
	f := psObj(d["facts"])
	host := psObj(psObj(d["host"])["resource"])
	scope := psObj(d["question"])
	contexts := psArr(d["contexts"])
	if len(contexts) != 2 {
		return false
	}
	seen := map[string]bool{}
	for _, v := range contexts {
		v := psObj(v)
		signal := fmt.Sprint(v["signal"])
		if signal != "logs" && signal != "spans" || seen[signal] {
			return false
		}
		seen[signal] = true
		var q exploration.Query
		var r exploration.Result
		if !rsDecode(v["question"], &q) || !rsDecode(v["result"], &r) || q.Environment != f["environment"] || q.Signal != signal || q.FromMS != int64(scope["fromMs"].(float64)) || q.ToMS != int64(scope["toMs"].(float64)) || q.Where == nil || len(q.Where.All) != len(host) || !nhLiveSameQuery(q, r.Question) {
			return false
		}
		want := exploration.Query{Signal: signal, Environment: f["environment"].(string), FromMS: int64(scope["fromMs"].(float64)), ToMS: int64(scope["toMs"].(float64)), Limit: 100, Where: &exploration.Predicate{All: []exploration.Predicate{}}}
		names := []string{}
		for k := range host {
			names = append(names, k)
		}
		sort.Strings(names)
		for _, k := range names {
			want.Where.All = append(want.Where.All, exploration.Predicate{Field: &exploration.Field{Kind: "resource", Key: k}, Op: "eq", Value: host[k].(string)})
		}
		if !aeJSONEqual(q, want) {
			return false
		}

		keys := map[string]bool{}
		for _, p := range q.Where.All {
			if p.Field == nil || p.Field.Kind != "resource" || p.Op != "eq" || host[p.Field.Key] != p.Value || keys[p.Field.Key] {
				return false
			}
			keys[p.Field.Key] = true
		}
		expected := map[string]psMap{}
		for _, v := range psArr(f["records"]) {
			p := psObj(v)
			if p["signal"] != signal {
				continue
			}
			match := true
			for k, v := range host {
				if psObj(p["resource"])[k] != v {
					match = false
				}
			}
			at, _ := strconv.ParseInt(p["timeNS"].(string), 10, 64)
			if match && at >= q.FromMS*1000000 && at < q.ToMS*1000000 {
				expected[p["id"].(string)] = p
			}
		}
		if len(expected) != 5 || len(r.Rows) != len(expected) || r.Count != fmt.Sprint(len(expected)) {
			return false
		}
		for _, raw := range r.Rows {
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
			p := expected[fmt.Sprint(attrs["work.id"])]
			if p == nil || !aeJSONEqual(row["ResourceAttributes"], p["resource"]) || !aeJSONEqual(attrs, p["attributes"]) || row["timestamp_ns"] != p["timeNS"] || row["TraceId"] != p["traceId"] || row["SpanId"] != p["spanId"] {
				return false
			}
			delete(expected, p["id"].(string))
		}
		if len(expected) != 0 {
			return false
		}
	}
	return true
}
func hfOK(d psMap) bool {
	f := psObj(d["facts"])
	results := psArr(d["results"])
	if len(results) != 2 {
		return false
	}
	seen := map[string]*measurements.Result{}
	for _, v := range results {
		var r measurements.Result
		if !rsDecode(v, &r) {
			return false
		}
		metric := r.Question.Definition.Source.Metric
		if seen[metric] != nil || !hfRaw(f, &r, metric) {
			return false
		}
		seen[metric] = &r
	}
	for _, metric := range []string{"system.filesystem.usage", "system.filesystem.limit"} {
		if seen[metric] == nil {
			return false
		}
	}
	var individual measurements.Result
	var saved exploration.CapturePage
	if !rsDecode(d["individual"], &individual) || !hfRaw(f, &individual, "system.filesystem.usage") || !rsDecode(d["saved"], &saved) || saved.Measurement == nil || !hfRaw(f, saved.Measurement, "system.filesystem.usage") || !peLiveCaptureHash(saved) || !hfContexts(d) || len(psArr(d["blocked"])) != 0 {
		return false
	}
	plots := psArr(d["plots"])
	if len(plots) != 2 {
		return false
	}
	for i, metric := range []string{"system.filesystem.usage", "system.filesystem.limit"} {
		plot := psObj(plots[i])
		if !aeJSONEqual(plot["domain"], psMap{"min": "0", "max": "134217728"}) || plot["unit"] != "By" || plot["kind"] != "stock" {
			return false
		}
		points := psArr(plot["points"])
		samples := seen[metric].Stocks.Series[0].Samples
		if len(points) != len(samples) {
			return false
		}
		for j, v := range points {
			p := psObj(v)
			sample := samples[j]
			value, e := strconv.ParseFloat(fmt.Sprint(p["value"]), 64)
			if e != nil || sample.Value == nil || value != *sample.Value || p["state"] != sample.State || p["fromNS"] != sample.AtNS || p["toNS"] != sample.AtNS {
				return false
			}
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
	if source != 9 || native != 5 || plans != 1 || captures != 1 {
		return false
	}
	events := psArr(d["events"])
	counts := []int{3, 5, 5, 6, 7, 8, 9}
	if len(events) != len(counts) {
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
		if measurements.HostFilesystemDefinition(w.Question.Definition) {
			return false
		}
	}
	return true
}
func TestHostFilesystemObserved(t *testing.T) {
	d, raw, path := rsRead(t, "GOO_HOST_FILESYSTEM_LEDGER")
	if !hfOK(d) {
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
		func(x psMap) { psObj(psArr(x["plots"])[0])["domain"] = psMap{"min": "0", "max": "8192"} },
		func(x psMap) { psObj(x["individual"])["basis"] = "bad" },
		func(x psMap) {
			sample := psObj(psArr(psObj(psArr(psObj(psObj(x["individual"])["stocks"])["series"])[0])["samples"])[0])
			sample["value"] = float64(999)
		},
		func(x psMap) {
			row := psObj(psArr(psObj(psObj(psArr(x["contexts"])[0])["result"])["rows"])[0])
			row["timestamp_ns"] = "1"
		},
		func(x psMap) { x["results"] = psArr(x["results"])[:1] },
	}
	changes = append(changes,
		func(x psMap) {
			sample := psObj(psArr(psObj(psArr(psObj(psObj(x["individual"])["stocks"])["series"])[0])["samples"])[0])
			sample["reason"] = "fabricated"
		},
		func(x psMap) {
			individual := psObj(x["individual"])
			individual["state"] = "unknown"
			psObj(psArr(psObj(individual["stocks"])["series"])[0])["state"] = "unknown"
		},
		func(x psMap) {
			question := psObj(psObj(psArr(x["contexts"])[0])["question"])
			psObj(psObj(psArr(x["contexts"])[0])["result"])["question"] = question
			psObj(question["where"])["any"] = []any{psMap{"field": psMap{"kind": "resource", "key": "system.device"}, "op": "eq", "value": "other"}}
		},
		func(x psMap) {
			ctx := psObj(psArr(x["contexts"])[0])
			question := psObj(ctx["question"])
			question["summary"] = "count"
			psObj(ctx["result"])["question"] = question
		},
	)

	for i, change := range changes {
		var x psMap
		_ = json.Unmarshal(raw, &x)
		change(x)
		if hfOK(x) {
			t.Fatal("accepted filesystem control", i)
		}
	}
	t.Log("eleven changed controls rejected; two independent streams, shared scale and full capture verified")
}
func hfTrace(d psMap) *qobs.MemorySink {
	s := &qobs.MemorySink{}
	actions := []string{"defaults", "filesystem", "local", "logs", "spans", "native", "capture"}
	for i, v := range psArr(d["events"]) {
		e := psObj(v)
		s.Record(context.Background(), &qtrace.Step{Seq: uint64(i + 1), Recorder: "host-filesystem", Outcome: "ok", Action: actions[i], Obs: map[string]any{"source": e["source"], "plans": e["plans"]}})
	}
	return s
}
func TestQuintHostFilesystemObserved(t *testing.T) {
	d, _, path := rsRead(t, "GOO_HOST_FILESYSTEM_LEDGER")
	if !hfOK(d) {
		t.Fatal("raw precondition")
	}
	checker, e := validate.NewChecker("model/host_filesystem.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
	if !check("filesystem-observed", hfTrace(d)) {
		t.Fatal("actual filesystem effect trace")
	}
	bad := hfTrace(d)
	bad.Steps()[2].Obs["source"] = float64(6)
	if check("filesystem-bad-reopen-read", bad) {
		t.Fatal("reopen read accepted")
	}
	t.Log("one actual seven-step trace; one bad local-read trace rejected")
}
