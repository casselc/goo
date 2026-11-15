package verification

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type paired444Lane struct {
	Generation int      `json:"generation"`
	Phase      string   `json:"phase"`
	Keys       []string `json:"keys"`
	Connection string   `json:"connection"`
	Find       struct {
		Text string `json:"text"`
	} `json:"find"`
	ReadWindow *struct {
		FromMS int64 `json:"fromMs"`
		ToMS   int64 `json:"toMs"`
	} `json:"readWindow"`
	Selector json.RawMessage `json:"selector"`
}
type paired444Snapshot struct {
	Reads map[string]int           `json:"reads"`
	Lanes map[string]paired444Lane `json:"lanes"`
}
type paired444Step struct {
	Name   string            `json:"name"`
	Action string            `json:"action"`
	Lane   string            `json:"lane"`
	Before paired444Snapshot `json:"before"`
	After  paired444Snapshot `json:"after"`
}
type paired444Report struct {
	Version, Status string
	Cleanup         bool
	Errors          []string
	Steps           []paired444Step
}

func paired444Active(phase string) bool {
	return slices.Contains([]string{"loading", "unknown", "partial", "stale"}, phase)
}
func paired444MetadataHeld(a, b paired444Lane, find bool) bool {
	if a.Connection != b.Connection {
		return false
	}
	var x, y any
	if len(a.Selector) > 0 {
		if json.Unmarshal(a.Selector, &x) != nil {
			return false
		}
	}
	if len(b.Selector) > 0 {
		if json.Unmarshal(b.Selector, &y) != nil {
			return false
		}
	}
	return reflect.DeepEqual(x, y) && (!find || a.Find.Text == b.Find.Text)
}
func paired444Held(a, b paired444Lane, find bool) bool {
	return slices.Equal(a.Keys, b.Keys) && reflect.DeepEqual(a.ReadWindow, b.ReadWindow) && paired444MetadataHeld(a, b, find)
}
func paired444SettlingHeld(a, b paired444Lane) bool {
	return paired444Held(a, b, true) || paired444Active(a.Phase) && paired444MetadataHeld(a, b, true)
}

func paired444Projection(step paired444Step) (string, bool, error) {
	lane := step.Lane
	if lane != "" && lane != "logs" && lane != "spans" {
		return "", false, fmt.Errorf("unknown lane %q", lane)
	}
	for _, snapshot := range []paired444Snapshot{step.Before, step.After} {
		for _, name := range []string{"logs", "spans"} {
			l, ok := snapshot.Lanes[name]
			if !ok {
				return "", false, fmt.Errorf("missing %s lane", name)
			}
			if _, ok := snapshot.Reads[name]; !ok || snapshot.Reads[name] < 0 {
				return "", false, fmt.Errorf("missing or invalid %s read count", name)
			}
			if !slices.Contains([]string{"loading", "unknown", "partial", "stale", "paused", "stopped", "unavailable"}, l.Phase) {
				return "", false, fmt.Errorf("unknown %s phase %q", name, l.Phase)
			}
			if l.ReadWindow != nil && (l.ReadWindow.FromMS < 0 || l.ReadWindow.ToMS-l.ReadWindow.FromMS != 60000) {
				return "", false, fmt.Errorf("invalid original read window")
			}
			if l.Generation < 0 {
				return "", false, fmt.Errorf("invalid owner generation")
			}
			seen := make(map[string]bool, len(l.Keys))
			for _, key := range l.Keys {
				if key == "" || seen[key] {
					return "", false, fmt.Errorf("duplicate or empty occurrence key")
				}
				seen[key] = true
			}
		}
	}
	for _, name := range []string{"logs", "spans"} {
		if step.After.Lanes[name].Generation < step.Before.Lanes[name].Generation {
			return "", false, fmt.Errorf("owner generation regressed")
		}
	}
	peer := "logs"
	if lane == "logs" {
		peer = "spans"
	}
	same := func(name string, find bool) bool {
		return paired444Held(step.Before.Lanes[name], step.After.Lanes[name], find)
	}
	held := true
	// A running sibling may receive an ordinary batch during a local gesture.
	peerHeld := func() bool { return paired444SettlingHeld(step.Before.Lanes[peer], step.After.Lanes[peer]) }
	suffix := "Logs"
	if lane == "spans" {
		suffix = "Spans"
	}
	switch step.Action {
	case "stage":
		return "stage", true, nil
	case "start":
		return "start", true, nil
	case "resume", "refuse":
		if lane == "" {
			return "", false, fmt.Errorf("resume requires lane")
		}
		return "resume" + suffix, peerHeld(), nil
	case "inspect", "pause", "failure":
		if lane == "" {
			return "pauseBoth", same("logs", true) && same("spans", true), nil
		}
		return "pause" + suffix, same(lane, true) && peerHeld(), nil
	case "navigate":
		return "navigate", paired444SettlingHeld(step.Before.Lanes["logs"], step.After.Lanes["logs"]) && paired444SettlingHeld(step.Before.Lanes["spans"], step.After.Lanes["spans"]), nil
	case "return":
		return "back", same("logs", true) && same("spans", true), nil
	case "late":
		if lane == "" {
			return "observe", same("logs", true) && same("spans", true), nil
		}
		return "observe", same(lane, true) && peerHeld() && step.Before.Lanes[lane].Generation == step.After.Lanes[lane].Generation, nil
	case "find":
		if lane == "" {
			return "", false, fmt.Errorf("find requires lane")
		}
		return "observe", same(lane, false) && peerHeld(), nil
	default:
		return "", held, fmt.Errorf("unknown paired action %q", step.Action)
	}
}

func paired444Checker(t *testing.T) *validate.Checker {
	t.Helper()
	dir := t.TempDir()
	data, err := os.ReadFile("model/paired_activity444.qnt")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "paired_activity444.qnt"), data, 0600); err != nil {
		t.Fatal(err)
	}
	binding := "spec: paired_activity444.qnt\nmodule: paired_activity444\nconstants: {INITIAL: '0'}\ninvariants: [safety]\nactions:\n"
	for _, name := range []string{"stage", "start", "resumeLogs", "resumeSpans", "pauseLogs", "pauseSpans", "pauseBoth", "navigate", "back", "observe"} {
		binding += "  " + name + ":\n"
		switch name {
		case "start":
			binding += "    args: {l: '{{.obs.logsActive}}', s: '{{.obs.spansActive}}'}\n"
		case "resumeLogs":
			binding += "    args: {running: '{{.obs.logsActive}}'}\n"
		case "resumeSpans":
			binding += "    args: {running: '{{.obs.spansActive}}'}\n"
		}
		binding += "    expect: 'logsReads == {{.obs.logsReads}} and spansReads == {{.obs.spansReads}} and logsActive == {{.obs.logsActive}} and spansActive == {{.obs.spansActive}} and held == {{.obs.held}}'\n"
	}
	file := filepath.Join(dir, "binding.yaml")
	if err = os.WriteFile(file, []byte(binding), 0600); err != nil {
		t.Fatal(err)
	}
	checker, err := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	return checker
}
func paired444Replay(t *testing.T, checker *validate.Checker, report paired444Report) bool {
	t.Helper()
	if report.Version != "goo.paired444/1" || report.Status != "passed" || !report.Cleanup || len(report.Errors) > 0 || len(report.Steps) == 0 {
		return false
	}
	base := report.Steps[0].Before.Reads
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for _, step := range report.Steps {
		action, held, err := paired444Projection(step)
		if err != nil {
			t.Log(err)
			return false
		}
		rec.Record(t.Context(), action, nil, "obs.logsReads", step.After.Reads["logs"]-base["logs"], "obs.spansReads", step.After.Reads["spans"]-base["spans"], "obs.logsActive", paired444Active(step.After.Lanes["logs"].Phase), "obs.spansActive", paired444Active(step.After.Lanes["spans"].Phase), "obs.held", held)
	}
	verdict, err := checker.Check(t.Context(), sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	if !verdict.OK() {
		t.Logf("paired verdict: %s", verdict.Summary())
		for _, actor := range verdict.Actors {
			t.Logf("actor failure %+v output %s", actor.Failure, actor.Output)
		}
	}
	return verdict.OK()
}

func TestQuintPaired444Model(t *testing.T) {
	cmd := exec.Command(quintPath(t), "run", "model/paired_activity444.qnt", "--main=paired_activity444_instance", "--backend=typescript", "--max-samples=50", "--max-steps=30", "--invariant=safety", "--seed=42")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("model simulation: %v: %s", err, out)
	}
}
func TestQuintPaired444Observed(t *testing.T) {
	file := os.Getenv("GOO_PAIRED444_REPORT")
	if file == "" {
		t.Skip("requires actual paired Activity browser report")
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var report paired444Report
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	checker := paired444Checker(t)
	if !paired444Replay(t, checker, report) {
		t.Fatal("actual paired Activity trace rejected")
	}
	// Mutations preserve report shape and observation order; each violates one
	// independently observed authority/custody rule rather than inventing a click.
	for _, kind := range []string{"return-read", "return-keys", "inspect-peer-read", "late-keys", "late-generation"} {
		t.Run(kind, func(t *testing.T) {
			var bad paired444Report
			_ = json.Unmarshal(raw, &bad)
			changed := false
			for i := range bad.Steps {
				s := &bad.Steps[i]
				switch {
				case strings.HasPrefix(kind, "return-") && s.Action == "return":
					if kind == "return-read" {
						s.After.Reads["logs"]++
					} else {
						l := s.After.Lanes["logs"]
						l.Keys = append(l.Keys, "invented-obsolete-occurrence")
						s.After.Lanes["logs"] = l
					}
					changed = true
				case kind == "inspect-peer-read" && s.Action == "inspect":
					peer := "logs"
					if s.Lane == "logs" {
						peer = "spans"
					}
					s.After.Reads[peer]++
					changed = true
				case (kind == "late-keys" || kind == "late-generation") && s.Action == "late":
					lane := s.Lane
					if lane == "" {
						lane = "logs"
					}
					l := s.After.Lanes[lane]
					if kind == "late-keys" {
						l.Keys = append(l.Keys, "late-cancelled-occurrence")
					} else {
						l.Generation++
					}
					s.After.Lanes[lane] = l
					changed = true
				}
				if changed {
					break
				}
			}
			if !changed {
				t.Fatalf("actual journey did not exercise %s", kind)
			}
			if paired444Replay(t, checker, bad) {
				t.Fatalf("%s mutant admitted", kind)
			}
		})
	}
}

// Protocol fixtures exercise checker admission, not a real browser execution.
func TestQuintPaired444ProtocolControls(t *testing.T) {
	lane := func(phase string) paired444Lane {
		return paired444Lane{Phase: phase, Keys: []string{"fixture-occurrence"}, Connection: "fixture", Selector: json.RawMessage(`{"environment":"fixture"}`)}
	}
	state := func(lr, sr int, lp, sp string) paired444Snapshot {
		return paired444Snapshot{Reads: map[string]int{"logs": lr, "spans": sr}, Lanes: map[string]paired444Lane{"logs": lane(lp), "spans": lane(sp)}}
	}
	report := paired444Report{Version: "goo.paired444/1", Status: "passed", Cleanup: true}
	appendStep := func(action, selected string, before, after paired444Snapshot) {
		report.Steps = append(report.Steps, paired444Step{Name: action, Action: action, Lane: selected, Before: before, After: after})
	}
	appendStep("stage", "", state(0, 0, "paused", "paused"), state(0, 0, "paused", "paused"))
	appendStep("start", "", state(0, 0, "paused", "paused"), state(1, 1, "unknown", "unknown"))
	appendStep("inspect", "logs", state(1, 1, "unknown", "unknown"), state(1, 1, "paused", "unknown"))
	appendStep("pause", "spans", state(1, 1, "paused", "unknown"), state(1, 1, "paused", "paused"))
	appendStep("navigate", "", state(1, 1, "paused", "paused"), state(1, 1, "paused", "paused"))
	appendStep("return", "", state(1, 1, "paused", "paused"), state(1, 1, "paused", "paused"))
	appendStep("refuse", "spans", state(1, 1, "paused", "paused"), state(1, 2, "paused", "unavailable"))
	appendStep("late", "logs", state(1, 2, "paused", "unavailable"), state(1, 2, "paused", "unavailable"))
	checker := paired444Checker(t)
	if !paired444Replay(t, checker, report) {
		t.Fatal("faithful protocol fixture refused")
	}
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"return-read", "return-keys", "inspect-peer-read", "late-keys", "refusal-peer-keys", "resume-away", "late-generation"} {
		t.Run(kind, func(t *testing.T) {
			var bad paired444Report
			if err := json.Unmarshal(raw, &bad); err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "return-read":
				bad.Steps[5].After.Reads["logs"]++
			case "return-keys":
				l := bad.Steps[5].After.Lanes["logs"]
				l.Keys = []string{"replacement"}
				bad.Steps[5].After.Lanes["logs"] = l
			case "inspect-peer-read":
				bad.Steps[2].After.Reads["spans"]++
			case "late-generation":
				l := bad.Steps[7].After.Lanes["logs"]
				l.Generation++
				bad.Steps[7].After.Lanes["logs"] = l
			case "late-keys":
				l := bad.Steps[7].After.Lanes["logs"]
				l.Keys = []string{"obsolete"}
				bad.Steps[7].After.Lanes["logs"] = l
			case "refusal-peer-keys":
				l := bad.Steps[6].After.Lanes["logs"]
				l.Keys = []string{"sibling-replaced"}
				bad.Steps[6].After.Lanes["logs"] = l
			case "resume-away":
				bad.Steps = bad.Steps[:6]
				bad.Steps[5].Action = "resume"
				bad.Steps[5].Lane = "logs"
				bad.Steps[5].After.Reads["logs"]++
			}
			if paired444Replay(t, checker, bad) {
				t.Fatalf("%s mutant admitted", kind)
			}
		})
	}
}
