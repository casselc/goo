package verification

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/qtrace"
	"github.com/casselc/observability/quintgo/validate"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Bind every supplied native result and history to the existing independent
// physical-row/OTLP oracle; counters alone cannot establish automatic defaults.
func rdOK(d psMap) bool {
	f := psObj(d["facts"])
	origin, ok := rsOrigin(psObj(d["initial"])["request"])
	if !ok {
		return false
	}
	responses := psArr(d["responses"])
	requests := psArr(d["requests"])
	source, plans, composes, histories, native := 0, 0, 0, 0, 0
	for _, v := range requests {
		r := psObj(v)
		if r["method"] != "POST" {
			continue
		}
		switch r["endpoint"] {
		case "/api/pod-discovery", "/api/pod-metric-discovery", "/api/measurements":
			source++
		case "/api/pod-dashboard/plan":
			plans++
		case "/api/container-memory/view", "/api/container-cpu/view":
			composes++
		case "/api/container-resources/history":
			histories++
		default:
			if r["effect"] == "source" {
				return false
			}
		}
		if r["endpoint"] == "/api/measurements" {
			native++
			q := psObj(psObj(r["payload"])["definition"])
			s := psObj(q["source"])
			metric := s["metric"]
			if metric != "container.memory.working_set" && metric != "container.cpu.usage" && metric != "k8s.container.memory_request" && metric != "k8s.container.memory_limit" {
				return false
			}
		}
	}
	if source != 8 || plans != 2 || composes != 4 || histories != 4 || native != 5 {
		return false
	}
	// Exactly one primary-container usage read per family; the ambiguous container
	// admits CPU only. Capacity must reuse the byte-identical primary result.
	seen := map[string]int{}
	for _, v := range responses {
		r := psObj(v)
		if r["endpoint"] != "/api/measurements" {
			continue
		}
		var result measurements.Result
		if !rsDecode(r["body"], &result) || !chBasis(&result) || !rsRaw(f, &result) {
			return false
		}
		s := result.Question.Definition.Source.ExactStream
		if s == nil {
			return false
		}
		key := s.Resource["container.id"] + "/" + s.Name
		seen[key]++
	}
	if len(seen) != 5 {
		return false
	}
	for _, n := range seen {
		if n != 1 {
			return false
		}
	}
	checked := 0
	for _, v := range responses {
		r := psObj(v)
		endpoint := r["endpoint"]
		family := "memory"
		if endpoint == "/api/container-cpu/view" {
			family = "cpu"
		} else if endpoint != "/api/container-memory/view" {
			continue
		}
		input := psObj(r["request"])
		var history psMap
		for _, h := range responses {
			h := psObj(h)
			if h["endpoint"] != "/api/container-resources/history" {
				continue
			}
			req := psObj(h["request"])
			fields := psMap{}
			for key, val := range input {
				if key != "parent" {
					fields[key] = val
				}
			}
			if aeJSONEqual(req[family], fields) && aeJSONEqual(req["parent"], input["parent"]) {
				history = psMap{"request": h["request"], "response": h["body"]}
				break
			}
		}
		if history == nil || !chPair(f, psMap{"request": r["request"], "response": r["body"], "history": history}, family, origin, responses) {
			return false
		}
		checked++
	}
	if checked != 4 {
		return false
	}
	want := [][4]int{{4, 1, 2, 2}, {6, 1, 3, 3}, {6, 1, 3, 3}, {6, 1, 3, 3}, {8, 2, 4, 4}}
	events := psArr(d["events"])
	if len(events) != len(want) {
		return false
	}
	for i, v := range events {
		e := psObj(v)
		for j, k := range []string{"source", "plans", "compositions", "histories"} {
			if e[k] != float64(want[i][j]) {
				return false
			}
		}
	}
	return true
}
func TestResourceDefaultsObserved(t *testing.T) {
	d, raw, _ := rsRead(t, "GOO_RESOURCE_DEFAULTS_LEDGER")
	if !rdOK(d) {
		t.Fatal("automatic resource raw/history/read-budget oracle")
	}
	changes := []func(psMap){
		func(x psMap) { psObj(psArr(x["events"])[0])["source"] = float64(5) },
		func(x psMap) {
			x["requests"] = append(psArr(x["requests"]), psMap{"method": "POST", "endpoint": "/api/measurements"})
		},
		func(x psMap) {
			for _, v := range psArr(x["responses"]) {
				r := psObj(v)
				if r["endpoint"] == "/api/measurements" {
					psObj(r["body"])["basis"] = "bad"
					return
				}
			}
		},
		func(x psMap) {
			for _, v := range psArr(x["responses"]) {
				r := psObj(v)
				if r["endpoint"] == "/api/container-resources/history" {
					series := psObj(psArr(psObj(r["body"])["series"])[0])
					p := psObj(psArr(series["points"])[0])
					p["value"] = "999"
					return
				}
			}
		},
	}
	for i, change := range changes {
		var x psMap
		if json.Unmarshal(raw, &x) != nil {
			t.Fatal("decode")
		}
		change(x)
		if rdOK(x) {
			t.Fatal("accepted changed control", i)
		}
	}
	t.Log("rejected four changed raw/effect controls")
}
func rdTrace(d psMap) *qobs.MemorySink {
	s := &qobs.MemorySink{}
	actions := []string{"usage", "capacity", "local", "local", "ambiguous"}
	for i, v := range psArr(d["events"]) {
		e := psObj(v)
		s.Record(context.Background(), &qtrace.Step{Seq: uint64(i + 1), Recorder: "resource-defaults", Outcome: "ok", Action: actions[i], Obs: map[string]any{"source": e["source"], "plans": e["plans"], "compositions": e["compositions"], "histories": e["histories"]}})
	}
	return s
}
func TestQuintResourceDefaultsObserved(t *testing.T) {
	d, _, path := rsRead(t, "GOO_RESOURCE_DEFAULTS_LEDGER")
	if !rdOK(d) {
		t.Fatal("raw precondition")
	}
	checker, e := validate.NewChecker("model/resource_defaults.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
	if !check("defaults-observed", rdTrace(d)) {
		t.Fatal("actual five-step trace")
	}
	bad := rdTrace(d)
	bad.Steps()[2].Obs["source"] = float64(7)
	if check("defaults-bad-local-read", bad) {
		t.Fatal("accepted local read")
	}
	bad = rdTrace(d)
	bad.Steps()[1].Obs["source"] = float64(7)
	if check("defaults-bad-primary-reread", bad) {
		t.Fatal("accepted primary reread")
	}
	t.Log("one actual finite trace / five macrosteps; two bad traces rejected")
}
