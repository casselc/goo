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
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func hcPoints(plot psMap, r *measurements.Result) bool {
	d := r.Question.Definition
	from, to := r.Question.BoundsNS()
	if plot["kind"] != d.Kind || plot["fromNS"] != from || plot["toNS"] != to {
		return false
	}
	points := psArr(plot["points"])
	if r.Counters != nil {
		s := r.Counters.Series[0]
		if plot["unit"] != s.Stream.Unit+"/s" || len(points) != len(s.Intervals) {
			return false
		}
		for i, v := range points {
			p := psObj(v)
			interval := s.Intervals[i]
			if p["fromNS"] != interval.FromNS || p["toNS"] != interval.ToNS {
				return false
			}
			if interval.State == "ok" {
				value, e := strconv.ParseFloat(fmt.Sprint(p["value"]), 64)
				if e != nil || interval.PerSecond == nil || value != *interval.PerSecond || p["state"] != "observed" {
					return false
				}
			} else if p["state"] != "unknown" || p["value"] != nil {
				return false
			}
		}
	} else if r.Gauges != nil {
		s := r.Gauges.Series[0]
		if plot["unit"] != s.Stream.Unit || len(points) != len(s.Samples) {
			return false
		}
		for i, v := range points {
			p := psObj(v)
			sample := s.Samples[i]
			value, e := strconv.ParseFloat(fmt.Sprint(p["value"]), 64)
			if e != nil || sample.Value == nil || value != *sample.Value || p["state"] != sample.State || p["fromNS"] != sample.AtNS || p["toNS"] != sample.AtNS {
				return false
			}
		}
	} else if r.Stocks != nil {
		s := r.Stocks.Series[0]
		if plot["unit"] != s.Stream.Unit || len(points) != len(s.Samples) {
			return false
		}
		for i, v := range points {
			p := psObj(v)
			sample := s.Samples[i]
			value, e := strconv.ParseFloat(fmt.Sprint(p["value"]), 64)
			if e != nil || sample.Value == nil || value != *sample.Value || p["state"] != sample.State || p["fromNS"] != sample.AtNS || p["toNS"] != sample.AtNS {
				return false
			}
		}
	} else {
		return false
	}
	var domain any
	switch d.Source.Metric {
	case "system.network.io", "system.disk.io":
		domain = psMap{"min": "0", "max": "1024"}
	case "system.filesystem.usage", "system.filesystem.limit":
		domain = psMap{"min": "0", "max": "134217728"}
	case "system.filesystem.utilization":
		domain = psMap{"min": "0", "max": "1"}
	}
	return aeJSONEqual(plot["domain"], domain) && plot["total"] == float64(len(points)) && plot["omitted"] == float64(0)
}
func hcOK(d psMap) bool {
	f := psObj(d["facts"])
	var plan measurements.HostDashboardPlan
	if !rsDecode(d["plan"], &plan) || len(plan.Widgets) != 2 {
		return false
	}
	results := []measurements.Result{}
	families := map[string]int{}
	for _, v := range psArr(d["responses"]) {
		x := psObj(v)
		if x["endpoint"] != "/api/measurements" {
			continue
		}
		var r measurements.Result
		if !rsDecode(x["body"], &r) || !chBasis(&r) {
			return false
		}
		metric := r.Question.Definition.Source.Metric
		valid := false
		switch metric {
		case "system.cpu.time":
			valid = ctRaw(f, &r)
		case "system.memory.usage":
			valid = hrRaw(f, &r)
		case "system.network.io":
			valid = hnRaw(f, &r, r.Question.Definition.Source.ExactStream.Attributes["network.io.direction"])
		case "system.disk.io":
			valid = hdRaw(f, &r, r.Question.Definition.Source.ExactStream.Attributes["disk.io.direction"])
		case "system.filesystem.usage", "system.filesystem.limit":
			valid = hfRaw(f, &r, metric)
		case "system.filesystem.utilization":
			valid = huRaw(f, &r)
		}
		if !valid {
			return false
		}
		families[metric]++
		results = append(results, r)
	}
	wantFamilies := map[string]int{"system.cpu.time": 1, "system.memory.usage": 1, "system.network.io": 4, "system.disk.io": 2, "system.filesystem.usage": 1, "system.filesystem.limit": 1, "system.filesystem.utilization": 2}
	if !reflect.DeepEqual(families, wantFamilies) {
		return false
	}
	proofs := psArr(d["plotProofs"])
	if len(proofs) != 9 {
		return false
	}
	keys := map[string]bool{}
	for _, v := range proofs {
		p := psObj(v)
		key := fmt.Sprint(p["key"])
		if keys[key] {
			return false
		}
		keys[key] = true
		found := false
		for _, w := range plan.Candidates {
			if w.Key != key {
				continue
			}
			for i := range results {
				r := &results[i]
				if aeJSONEqual(w.Question, r.Question) && hcPoints(psObj(p["data"]), r) {
					found = true
					break
				}
			}
		}
		if !found {
			return false
		}
	}
	var individual measurements.Result
	var saved exploration.CapturePage
	if !rsDecode(d["individual"], &individual) || !huRaw(f, &individual) || !rsDecode(d["saved"], &saved) || saved.Measurement == nil || !huRaw(f, saved.Measurement) || !peLiveCaptureHash(saved) || len(psArr(d["blocked"])) != 0 {
		return false
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
	if source != 14 || native != 12 || plans != 1 || captures != 1 {
		return false
	}
	counts := []int{3, 5, 7, 9, 10, 10, 12, 13, 14}
	names := []string{"defaults", "network", "disk", "bytes", "utilization", "local", "refresh", "native", "capture"}
	events := psArr(d["events"])
	if len(events) != len(counts) {
		return false
	}
	for i, v := range events {
		e := psObj(v)
		if e["source"] != float64(counts[i]) || e["plans"] != float64(1) || e["action"] != names[i] {
			return false
		}
	}
	return true
}
func TestHostControlsObserved(t *testing.T) {
	d, raw, path := rsRead(t, "GOO_HOST_CONTROLS_LEDGER")
	if !hcOK(d) {
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
		func(x psMap) { psObj(psArr(x["events"])[5])["source"] = float64(11) },
		func(x psMap) {
			x["requests"] = append(psArr(x["requests"]), psMap{"method": "POST", "effect": "source", "endpoint": "/api/measurements"})
		},
		func(x psMap) {
			p := psObj(psArr(x["plotProofs"])[0])
			psObj(psArr(psObj(p["data"])["points"])[0])["value"] = "999"
		},
		func(x psMap) {
			for _, v := range psArr(x["plotProofs"]) {
				p := psObj(psObj(v)["data"])
				if p["unit"] == "By/s" {
					p["domain"] = psMap{"min": "0", "max": "1"}
					break
				}
			}
		},
		func(x psMap) { psObj(x["individual"])["basis"] = "bad" },
		func(x psMap) { psObj(psArr(x["plotProofs"])[0])["key"] = "bad" },
	}
	for i, change := range changes {
		var x psMap
		_ = json.Unmarshal(raw, &x)
		change(x)
		if hcOK(x) {
			t.Fatal("accepted filesystem control", i)
		}
	}
	t.Log("six changed controls rejected; nine actual plots retain raw values/domains and full capture")
}
func hcTrace(d psMap) *qobs.MemorySink {
	s := &qobs.MemorySink{}
	actions := []string{"defaults", "network", "disk", "bytes", "utilization", "local", "refresh", "native", "capture"}
	for i, v := range psArr(d["events"]) {
		e := psObj(v)
		s.Record(context.Background(), &qtrace.Step{Seq: uint64(i + 1), Recorder: "filesystem-utilization", Outcome: "ok", Action: actions[i], Obs: map[string]any{"source": e["source"], "plans": e["plans"]}})
	}
	return s
}
func TestQuintHostControlsObserved(t *testing.T) {
	d, _, path := rsRead(t, "GOO_HOST_CONTROLS_LEDGER")
	if !hcOK(d) {
		t.Fatal("raw precondition")
	}
	checker, e := validate.NewChecker("model/host_controls.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
	if !check("filesystem-observed", hcTrace(d)) {
		t.Fatal("actual filesystem effect trace")
	}
	bad := hcTrace(d)
	bad.Steps()[5].Obs["source"] = float64(11)
	if check("filesystem-bad-reopen-read", bad) {
		t.Fatal("reopen read accepted")
	}
	t.Log("one actual nine-step trace; one bad local-read trace rejected")
}
