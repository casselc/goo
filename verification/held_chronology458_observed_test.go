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

type chronology458Window struct {
	FromMS int64 `json:"fromMs"`
	ToMS   int64 `json:"toMs"`
}
type chronology458Item struct {
	Key           string              `json:"key"`
	Record        json.RawMessage     `json:"record"`
	ReadWindow    chronology458Window `json:"readWindow"`
	ObservedAt    string              `json:"observedAt"`
	Limited, Late bool
}
type chronology458OriginWindow struct {
	Key    string `json:"key"`
	FromMS int64  `json:"fromMs"`
	ToMS   int64  `json:"toMs"`
}
type chronology458Lane struct {
	Windows       []chronology458OriginWindow `json:"windows,omitempty"`
	Selector      json.RawMessage             `json:"selector"`
	Phase         string                      `json:"phase"`
	Rows          []chronology458Item         `json:"rows"`
	ReadWindow    *chronology458Window        `json:"readWindow"`
	Connection    string                      `json:"connection"`
	CompletedRead json.RawMessage             `json:"completedRead"`
	Evicted       int                         `json:"evicted"`
	Gap           bool                        `json:"gap"`
}
type chronology458Held struct {
	Version string                       `json:"version"`
	Lanes   map[string]chronology458Lane `json:"lanes"`
}
type chronology458Selected struct {
	Signal, Key string
	ReadWindow  *chronology458Window `json:"readWindow"`
	Window      *chronology458Window `json:"window"`
	Selector    json.RawMessage      `json:"selector"`
}
type chronology458Snapshot struct {
	Banks          map[string]chronology458Lane `json:"banks"`
	TailReads      map[string]int               `json:"tailReads"`
	SQL            int                          `json:"sql"`
	View           string                       `json:"view"`
	Paused         map[string]bool              `json:"paused"`
	SourceKeys     map[string][]string          `json:"sourceKeys"`
	ChronologyKeys json.RawMessage              `json:"chronologyKeys"`
	Held           *chronology458Held           `json:"held"`
	Selected       *chronology458Selected       `json:"selected"`
}
type chronology458Step struct {
	Name, Action, Signal, Key string
	Before, After             chronology458Snapshot
	SQLBefore, SQLAfter       *int
}
type chronology458Report struct {
	Version, Status          string
	Cleanup, DatabaseCleanup bool
	Errors                   []string
	Steps                    []chronology458Step
	FixedQuestion            json.RawMessage `json:"fixedQuestion"`
	Source                   []struct {
		SQL    string
		Status *int
	}
}

func chronology458JSONEqual(a, b json.RawMessage) bool {
	var x, y any
	if json.Unmarshal(a, &x) != nil || json.Unmarshal(b, &y) != nil {
		return false
	}
	return reflect.DeepEqual(x, y)
}
func chronology458Keys(held *chronology458Held) map[string][]string {
	out := map[string][]string{}
	if held == nil {
		return out
	}
	for _, signal := range []string{"logs", "spans"} {
		for _, row := range held.Lanes[signal].Rows {
			out[signal] = append(out[signal], row.Key)
		}
		if out[signal] == nil {
			out[signal] = []string{}
		}
	}
	return out
}
func chronology458SameHeld(a, b *chronology458Held) bool {
	if a == nil || b == nil {
		return a == b
	}
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	return chronology458JSONEqual(x, y)
}
func chronology458Has(held *chronology458Held, signal, key string) (*chronology458Item, bool) {
	if held == nil || !slices.Contains([]string{"logs", "spans"}, signal) {
		return nil, false
	}
	for _, row := range held.Lanes[signal].Rows {
		if row.Key == key {
			return &row, true
		}
	}
	return nil, false
}
func chronology458ValidHeld(held *chronology458Held) bool {
	if held == nil || held.Version != "goo.held-activity/1" || len(held.Lanes) != 2 {
		return false
	}
	for _, signal := range []string{"logs", "spans"} {
		lane, ok := held.Lanes[signal]
		if !ok || !slices.Contains([]string{"paused", "unavailable", "stopped"}, lane.Phase) || lane.Evicted < 0 || len(lane.Rows) > 100 {
			return false
		}
		if string(lane.Selector) == "null" {
			if len(lane.Rows) != 0 || lane.ReadWindow != nil || string(lane.CompletedRead) != "null" || lane.Connection != "" || lane.Evicted != 0 || lane.Gap || !slices.Contains([]string{"paused", "stopped"}, lane.Phase) {
				return false
			}
			continue
		}
		var selector map[string]any
		if json.Unmarshal(lane.Selector, &selector) != nil || (selector["signal"] != signal && !(signal == "logs" && selector["signal"] == nil)) {
			return false
		}
		seen := map[string]bool{}
		for _, row := range lane.Rows {
			if row.Key == "" || seen[row.Key] || row.ReadWindow.FromMS < 0 || row.ReadWindow.ToMS-row.ReadWindow.FromMS != 60000 || row.ObservedAt == "" {
				return false
			}
			seen[row.Key] = true
		}
	}
	return true
}
func chronology458VisibleMembers(raw json.RawMessage, held *chronology458Held) bool {
	if len(raw) == 0 || string(raw) == "null" {
		return true
	}
	var values []json.RawMessage
	if json.Unmarshal(raw, &values) != nil {
		return false
	}
	seen := map[string]bool{}
	for _, value := range values {
		var key string
		var pair struct{ Signal, Key string }
		if json.Unmarshal(value, &key) == nil {
			parts := strings.SplitN(key, ":", 2)
			if len(parts) != 2 {
				return false
			}
			pair.Signal, pair.Key = parts[0], parts[1]
		} else if json.Unmarshal(value, &pair) != nil {
			return false
		}
		compound := pair.Signal + ":" + pair.Key
		if seen[compound] {
			return false
		}
		seen[compound] = true
		if _, ok := chronology458Has(held, pair.Signal, pair.Key); !ok {
			return false
		}
	}
	return true
}
func chronology458Projection(step chronology458Step) (string, bool, error) {
	action := step.Action
	if action == "return" {
		action = "back"
	}
	if !slices.Contains([]string{"start", "resumeLogs", "resumeSpans", "hold", "find", "page", "inspect", "analyze", "back", "lanes", "retire", "observe"}, action) {
		return "", false, fmt.Errorf("unknown action %q", action)
	}
	valid := step.After.SQL >= step.Before.SQL
	for _, snapshot := range []chronology458Snapshot{step.Before, step.After} {
		if !slices.Contains([]string{"lanes", "chronology"}, snapshot.View) || snapshot.SQL < 0 || len(snapshot.TailReads) != 2 || len(snapshot.Paused) != 2 || len(snapshot.SourceKeys) != 2 {
			return "", false, fmt.Errorf("missing observed lane/read metadata")
		}
		for _, signal := range []string{"logs", "spans"} {
			if snapshot.TailReads[signal] < 0 {
				return "", false, fmt.Errorf("negative reads")
			}
		}
		if snapshot.View == "chronology" {
			valid = valid && snapshot.Paused["logs"] && snapshot.Paused["spans"] && chronology458ValidHeld(snapshot.Held) && chronology458VisibleMembers(snapshot.ChronologyKeys, snapshot.Held)
		}
	}
	local := !slices.Contains([]string{"start", "resumeLogs", "resumeSpans"}, action)
	liveObserve := action == "observe" && step.Before.View == "lanes" && (!step.Before.Paused["logs"] || !step.Before.Paused["spans"])
	if local {
		valid = valid && reflect.DeepEqual(step.Before.TailReads, step.After.TailReads) && (liveObserve || step.Before.SQL == step.After.SQL)
		if step.SQLBefore != nil && step.SQLAfter != nil {
			valid = valid && (liveObserve || *step.SQLBefore == *step.SQLAfter)
		}
	}
	if action == "hold" {
		valid = valid && step.After.View == "chronology"
		keys := chronology458Keys(step.After.Held)
		for _, signal := range []string{"logs", "spans"} {
			valid = valid && slices.Equal(step.Before.SourceKeys[signal], keys[signal]) && slices.Equal(step.Before.SourceKeys[signal], step.After.SourceKeys[signal])
			if step.After.Held != nil {
				source, ok := step.Before.Banks[signal]
				capture := step.After.Held.Lanes[signal]
				a, scopeOK := chronology458Scope(source.Selector, signal)
				b, captureOK := chronology458Scope(capture.Selector, signal)
				valid = valid && ok && ((scopeOK && captureOK && chronology458JSONEqual(a, b)) || (string(source.Selector) == "null" && string(capture.Selector) == "null")) && reflect.DeepEqual(source.ReadWindow, capture.ReadWindow) && chronology458JSONEqual(source.CompletedRead, capture.CompletedRead) && source.Connection == capture.Connection && source.Evicted == capture.Evicted && source.Gap == capture.Gap
				windows := map[string]chronology458Window{}
				for _, window := range source.Windows {
					if _, duplicate := windows[window.Key]; duplicate {
						valid = false
					}
					windows[window.Key] = chronology458Window{window.FromMS, window.ToMS}
				}
				valid = valid && len(windows) == len(capture.Rows)
				for _, item := range capture.Rows {
					window, ok := windows[item.Key]
					valid = valid && ok && window == item.ReadWindow
				}
			}
		}
	}
	if step.Before.View == "chronology" && step.After.View == "chronology" && action != "hold" {
		valid = valid && chronology458SameHeld(step.Before.Held, step.After.Held) && reflect.DeepEqual(step.Before.SourceKeys, step.After.SourceKeys)
	}
	if slices.Contains([]string{"start", "resumeLogs", "resumeSpans", "retire", "lanes"}, action) {
		valid = valid && step.After.View == "lanes"
	}
	if slices.Contains([]string{"inspect", "analyze"}, action) {
		sel := step.After.Selected
		if sel == nil {
			return action, false, nil
		}
		row, ok := chronology458Has(step.After.Held, sel.Signal, sel.Key)
		valid = valid && ok
		if ok {
			window := sel.ReadWindow
			if window == nil {
				window = sel.Window
			}
			valid = valid && window != nil && *window == row.ReadWindow && func() bool {
				a, ok := chronology458Scope(sel.Selector, sel.Signal)
				b, valid := chronology458Scope(step.After.Held.Lanes[sel.Signal].Selector, sel.Signal)
				return ok && valid && chronology458JSONEqual(a, b)
			}()
		}
	}
	return action, valid, nil
}
func chronology458Scope(raw json.RawMessage, signal string) (json.RawMessage, bool) {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil {
		return nil, false
	}
	var reported string
	if value, ok := fields["signal"]; ok {
		if json.Unmarshal(value, &reported) != nil || reported != signal {
			return nil, false
		}
	} else if signal != "logs" {
		return nil, false
	}
	delete(fields, "signal")
	for _, key := range []string{"fromMs", "toMs", "limit", "offset"} {
		delete(fields, key)
	}
	out, err := json.Marshal(fields)
	return out, err == nil
}
func chronology458FixedQuestion(raw json.RawMessage, step chronology458Step) bool {
	selected := step.After.Selected
	if selected == nil {
		return false
	}
	row, ok := chronology458Has(step.After.Held, selected.Signal, selected.Key)
	if !ok {
		return false
	}
	var q struct {
		Signal string
		FromMS int64 `json:"fromMs"`
		ToMS   int64 `json:"toMs"`
	}
	if json.Unmarshal(raw, &q) != nil || q.Signal != selected.Signal || q.FromMS != row.ReadWindow.FromMS || q.ToMS != row.ReadWindow.ToMS {
		return false
	}
	actual, ok := chronology458Scope(raw, selected.Signal)
	if !ok {
		return false
	}
	expected, ok := chronology458Scope(step.After.Held.Lanes[selected.Signal].Selector, selected.Signal)
	return ok && chronology458JSONEqual(actual, expected)
}

func chronology458Checker(t *testing.T) *validate.Checker {
	t.Helper()
	dir := t.TempDir()
	data, err := os.ReadFile("model/held_chronology458.qnt")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "held_chronology458.qnt"), data, 0600); err != nil {
		t.Fatal(err)
	}
	binding := "spec: held_chronology458.qnt\nmodule: held_chronology458\nconstants: {INITIAL: '0'}\ninvariants: [safety]\nactions:\n"
	for _, name := range []string{"start", "resumeLogs", "resumeSpans", "hold", "find", "page", "inspect", "analyze", "back", "lanes", "retire", "observe"} {
		binding += "  " + name + ":\n    expect: 'logsReads == {{.obs.logsReads}} and spansReads == {{.obs.spansReads}} and held == {{.obs.held}} and intact == {{.obs.valid}}'\n"
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
func chronology458Replay(t *testing.T, checker *validate.Checker, r chronology458Report) bool {
	t.Helper()
	if r.Version != "goo.heldChronology458/1" || r.Status != "passed" || !r.Cleanup || !r.DatabaseCleanup || len(r.Errors) != 0 || len(r.Steps) == 0 {
		return false
	}
	base := r.Steps[0].Before.TailReads
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range r.Steps {
		action, valid, err := chronology458Projection(s)
		if err != nil {
			t.Log(err)
			return false
		}
		if s.Action == "analyze" {
			valid = valid && chronology458FixedQuestion(r.FixedQuestion, s)
		}
		if i > 0 {
			prior := r.Steps[i-1].After
			valid = valid && reflect.DeepEqual(prior.TailReads, s.Before.TailReads) && prior.View == s.Before.View
		}
		rec.Record(t.Context(), action, nil, "obs.logsReads", s.After.TailReads["logs"]-base["logs"], "obs.spansReads", s.After.TailReads["spans"]-base["spans"], "obs.held", s.After.View == "chronology", "obs.valid", valid)
	}
	verdict, err := checker.Check(t.Context(), sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	if !verdict.OK() {
		t.Logf("chronology verdict: %s", verdict.Summary())
	}
	return verdict.OK()
}
func TestQuintHeldChronology458Model(t *testing.T) {
	cmd := exec.Command(quintPath(t), "run", "model/held_chronology458.qnt", "--main=held_chronology458_instance", "--backend=typescript", "--max-samples=50", "--max-steps=30", "--invariant=safety", "--seed=42")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("model: %v %s", err, out)
	}
}
func TestQuintHeldChronology458Observed(t *testing.T) {
	file := os.Getenv("GOO_HELD_CHRONOLOGY458_REPORT")
	if file == "" {
		t.Skip("requires actual held chronology browser report")
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var report chronology458Report
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Source) == 0 {
		t.Fatal("missing physical source ledger")
	}
	for _, call := range report.Source {
		if call.SQL == "" || call.Status == nil || *call.Status != 200 {
			t.Fatal("incomplete physical source witness")
		}
	}
	checker := chronology458Checker(t)
	if !chronology458Replay(t, checker, report) {
		t.Fatal("actual held chronology trace rejected")
	}
	rejected := 0
	for _, kind := range []string{"hold-read", "hold-sql", "hold-key", "live-held", "mutated-bank", "foreign-row", "inspect-window", "inspect-scope", "stale-retirement"} {
		t.Run(kind, func(t *testing.T) {
			var bad chronology458Report
			_ = json.Unmarshal(raw, &bad)
			changed := false
			for i := range bad.Steps {
				s := &bad.Steps[i]
				switch {
				case s.Action == "hold" && kind == "hold-read":
					s.After.TailReads["logs"]++
					changed = true
				case s.Action == "hold" && kind == "hold-sql":
					s.After.SQL++
					changed = true
				case s.Action == "hold" && kind == "hold-key":
					s.After.Held.Lanes["logs"] = func() chronology458Lane {
						l := s.After.Held.Lanes["logs"]
						l.Rows = append(l.Rows, chronology458Item{Key: "invented"})
						return l
					}()
					changed = true
				case s.Action == "hold" && kind == "live-held":
					s.After.Paused["spans"] = false
					changed = true
				case s.Before.View == "chronology" && s.After.View == "chronology" && s.Action != "hold" && kind == "mutated-bank":
					l := s.After.Held.Lanes["logs"]
					l.Evicted++
					s.After.Held.Lanes["logs"] = l
					changed = true
				case s.After.View == "chronology" && kind == "foreign-row":
					s.After.ChronologyKeys = json.RawMessage(`["logs:invented"]`)
					changed = true
				case s.Action == "inspect" && kind == "inspect-window":
					if s.After.Selected.ReadWindow != nil {
						s.After.Selected.ReadWindow.ToMS++
					} else if s.After.Selected.Window != nil {
						s.After.Selected.Window.ToMS++
					}
					changed = true
				case s.Action == "inspect" && kind == "inspect-scope":
					s.After.Selected.Selector = json.RawMessage(`{"environment":"other"}`)
					changed = true
				case s.Before.View == "chronology" && slices.Contains([]string{"retire", "start", "resumeLogs", "resumeSpans"}, s.Action) && kind == "stale-retirement":
					s.After.View = "chronology"
					changed = true
				}
				if changed {
					break
				}
			}
			if !changed {
				t.Fatalf("journey did not exercise %s", kind)
			}
			if chronology458Replay(t, checker, bad) {
				t.Fatalf("%s admitted", kind)
			}
			rejected++
		})
	}
	t.Logf("actual held chronology trace admitted; %d negative controls rejected", rejected)
}

// Explicit synthetic protocol fixtures test the checker, not a browser run.
func TestQuintHeldChronology458ProtocolControls(t *testing.T) {
	key := strings.Repeat("a", 64) + ":1"
	held := &chronology458Held{Version: "goo.held-activity/1", Lanes: map[string]chronology458Lane{}}
	for _, signal := range []string{"logs", "spans"} {
		selector := json.RawMessage(`{"environment":"fixture"}`)
		if signal == "spans" {
			selector = json.RawMessage(`{"environment":"fixture","signal":"spans"}`)
		}
		held.Lanes[signal] = chronology458Lane{Windows: []chronology458OriginWindow{{key, 1000, 61000}}, CompletedRead: json.RawMessage(`null`), Phase: "paused", Selector: selector, Rows: []chronology458Item{{Key: key, Record: json.RawMessage(`{"timestamp_ns":"1000000001","Body":"fixture","SpanName":"fixture"}`), ReadWindow: chronology458Window{1000, 61000}, ObservedAt: "2026-11-16T12:00:00Z"}}}
	}
	state := func(reads int, merged, paused bool) chronology458Snapshot {
		mode := "lanes"
		var data *chronology458Held
		if merged {
			mode = "chronology"
			data = held
		}
		return chronology458Snapshot{Banks: held.Lanes, TailReads: map[string]int{"logs": reads, "spans": reads}, SQL: 10, View: mode, Paused: map[string]bool{"logs": paused, "spans": paused}, SourceKeys: map[string][]string{"logs": {key}, "spans": {key}}, ChronologyKeys: json.RawMessage(`["logs:` + key + `","spans:` + key + `"]`), Held: data}
	}
	selected := &chronology458Selected{Signal: "logs", Key: key, ReadWindow: &chronology458Window{1000, 61000}, Selector: json.RawMessage(`{"environment":"fixture"}`)}
	r := chronology458Report{Version: "goo.heldChronology458/1", Status: "passed", Cleanup: true, DatabaseCleanup: true, FixedQuestion: json.RawMessage(`{"signal":"logs","environment":"fixture","fromMs":1000,"toMs":61000,"limit":100,"offset":0}`)}
	add := func(action string, before, after chronology458Snapshot) {
		r.Steps = append(r.Steps, chronology458Step{Name: action, Action: action, Before: before, After: after})
	}
	add("start", state(0, false, true), state(1, false, false))
	add("hold", state(1, false, false), state(1, true, true))
	add("find", state(1, true, true), state(1, true, true))
	a, b := state(1, true, true), state(1, true, true)
	b.Selected = selected
	add("inspect", a, b)
	add("analyze", b, b)
	add("return", b, b)
	add("lanes", b, state(1, false, true))
	add("hold", state(1, false, true), state(1, true, true))
	add("retire", state(1, true, true), state(1, false, true))
	checker := chronology458Checker(t)
	if !chronology458Replay(t, checker, r) {
		t.Fatal("faithful synthetic protocol refused")
	}
	raw, _ := json.Marshal(r)
	for _, kind := range []string{"unpaused-hold", "invented-key", "source-sql", "wrong-window", "wrong-selector", "stale-mode", "forbidden-start"} {
		t.Run(kind, func(t *testing.T) {
			var bad chronology458Report
			_ = json.Unmarshal(raw, &bad)
			switch kind {
			case "unpaused-hold":
				bad.Steps[1].After.Paused["spans"] = false
			case "invented-key":
				bad.Steps[1].After.SourceKeys["logs"] = []string{"invented"}
			case "source-sql":
				bad.Steps[2].After.SQL++
			case "wrong-window":
				bad.Steps[3].After.Selected.ReadWindow.ToMS++
			case "wrong-selector":
				bad.FixedQuestion = json.RawMessage(`{"signal":"logs","environment":"wrong","fromMs":1000,"toMs":61000,"limit":100,"offset":0}`)
			case "stale-mode":
				bad.Steps[8].After.View = "chronology"
			case "forbidden-start":
				bad.Steps[4].Action = "start"
			}
			if chronology458Replay(t, checker, bad) {
				t.Fatalf("%s admitted", kind)
			}
		})
	}
}
