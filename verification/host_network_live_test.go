package verification

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/qtrace"
	"github.com/casselc/observability/quintgo/validate"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func hnRaw(f psMap, r *measurements.Result, direction string) bool {
	return hostRateRaw(f, r, "system.network.io", "By", map[string]string{"network.interface.name": "eth0", "network.io.direction": direction})
}
func hnOK(d psMap) bool {
	f := psObj(d["facts"])
	results := psArr(d["results"])
	if len(results) != 2 {
		return false
	}
	seen := map[string]*measurements.Result{}
	for _, v := range results {
		var r measurements.Result
		if !rsDecode(v, &r) || r.Question.Definition.Source.ExactStream == nil {
			return false
		}
		direction := r.Question.Definition.Source.ExactStream.Attributes["network.io.direction"]
		if seen[direction] != nil || !hnRaw(f, &r, direction) {
			return false
		}
		seen[direction] = &r
	}
	if seen["receive"] == nil || seen["transmit"] == nil {
		return false
	}
	var individual measurements.Result
	var saved exploration.CapturePage
	if !rsDecode(d["individual"], &individual) || !hnRaw(f, &individual, "receive") || !rsDecode(d["saved"], &saved) || saved.Measurement == nil || !hnRaw(f, saved.Measurement, "receive") || !peLiveCaptureHash(saved) {
		return false
	}
	if !hostRateContexts(f, psArr(d["contexts"])) || len(psArr(d["blocked"])) != 0 {
		return false
	}
	plots := psArr(d["plots"])
	if len(plots) != 2 {
		return false
	}
	for i, direction := range []string{"receive", "transmit"} {
		plot := psObj(plots[i])
		if !aeJSONEqual(plot["domain"], psMap{"min": "0", "max": "1024"}) || plot["unit"] != "By/s" {
			return false
		}
		points := psArr(plot["points"])
		intervals := seen[direction].Counters.Series[0].Intervals
		if len(points) != len(intervals) {
			return false
		}
		for j, p := range points {
			p := psObj(p)
			interval := intervals[j]
			if p["fromNS"] != interval.FromNS || p["toNS"] != interval.ToNS {
				return false
			}
			if interval.State == "ok" {
				value, e := strconv.ParseFloat(p["value"].(string), 64)
				if e != nil || interval.PerSecond == nil || value != *interval.PerSecond || p["state"] != "observed" {
					return false
				}
			} else if p["state"] != "unknown" || p["value"] != nil {
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
		if w.Question.Definition.Source.Metric == "system.network.io" {
			return false
		}
	}
	return true
}
func TestHostNetworkObserved(t *testing.T) {
	d, raw, path := rsRead(t, "GOO_HOST_NETWORK_LEDGER")
	if !hnOK(d) {
		t.Fatal("network raw/shared-scale/context/effect oracle")
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
		t.Fatal("full immutable interface capture")
	}
	changes := []func(psMap){func(x psMap) { psObj(psArr(x["events"])[0])["source"] = float64(4) }, func(x psMap) {
		x["requests"] = append(psArr(x["requests"]), psMap{"method": "POST", "effect": "source", "endpoint": "/api/measurements"})
	}, func(x psMap) { psObj(psArr(x["plots"])[0])["domain"] = psMap{"min": "0", "max": "512"} }, func(x psMap) { psObj(x["individual"])["basis"] = "bad" }, func(x psMap) {
		p := psObj(psArr(psObj(psArr(psObj(psObj(x["individual"])["counters"])["series"])[0])["intervals"])[0])
		p["exactPerSecond"] = "999"
	}, func(x psMap) {
		row := psObj(psArr(psObj(psObj(psArr(x["contexts"])[0])["result"])["rows"])[0])
		row["timestamp_ns"] = "1"
	}, func(x psMap) { x["results"] = psArr(x["results"])[:1] }}
	for i, change := range changes {
		var x psMap
		_ = json.Unmarshal(raw, &x)
		change(x)
		if hnOK(x) {
			t.Fatal("accepted network control", i)
		}
	}
	t.Log("seven changed controls rejected; two independent streams, shared scale and full capture verified")
}
func hnTrace(d psMap) *qobs.MemorySink {
	s := &qobs.MemorySink{}
	actions := []string{"defaults", "network", "local", "native", "logs", "spans", "capture"}
	for i, v := range psArr(d["events"]) {
		e := psObj(v)
		s.Record(context.Background(), &qtrace.Step{Seq: uint64(i + 1), Recorder: "host-network", Outcome: "ok", Action: actions[i], Obs: map[string]any{"source": e["source"], "plans": e["plans"]}})
	}
	return s
}
func TestQuintHostNetworkObserved(t *testing.T) {
	d, _, path := rsRead(t, "GOO_HOST_NETWORK_LEDGER")
	if !hnOK(d) {
		t.Fatal("raw precondition")
	}
	checker, e := validate.NewChecker("model/host_network.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
	if !check("network-observed", hnTrace(d)) {
		t.Fatal("actual network effect trace")
	}
	bad := hnTrace(d)
	bad.Steps()[2].Obs["source"] = float64(6)
	if check("network-bad-reopen-read", bad) {
		t.Fatal("reopen read accepted")
	}
	t.Log("one actual seven-step trace; one bad local-read trace rejected")
}
