package scopehistory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"hegel.dev/go/hegel"
)

const day int64 = 86400000

type command struct {
	Name    string         `json:"name"`
	Intent  map[string]any `json:"intent,omitempty"`
	MS      int64          `json:"ms"`
	Enabled bool           `json:"enabled,omitempty"`
}
type item struct {
	Key    string `json:"key"`
	UsedAt int64  `json:"usedAt"`
	Pinned bool   `json:"pinned"`
	Offset *int   `json:"offset"`
}
type state struct {
	Durable  bool   `json:"durable"`
	Items    []item `json:"items"`
	Rejected bool   `json:"rejected"`
}
type oracle struct {
	now         int64
	durable     bool
	items, disk []item
	stored      bool
}

func (o *oracle) bound() {
	kept := make([]item, 0, len(o.items))
	for _, v := range o.items {
		if v.Pinned || o.now-v.UsedAt <= 30*day {
			kept = append(kept, v)
		}
	}
	slices.SortStableFunc(kept, func(a, b item) int {
		if a.UsedAt > b.UsedAt {
			return -1
		}
		if a.UsedAt < b.UsedAt {
			return 1
		}
		return 0
	})
	pinned, recent := []item{}, []item{}
	for _, v := range kept {
		if v.Pinned {
			pinned = append(pinned, v)
		} else if len(recent) < 20 {
			recent = append(recent, v)
		}
	}
	o.items = append(pinned, recent...)
}
func intent(n, offset int) map[string]any {
	environment := fmt.Sprintf("env-%d", n)
	if n%2 == 0 {
		return map[string]any{"kind": "activity", "selector": map[string]any{"signal": "logs", "environment": environment}, "columns": []string{"utc", "summary"}, "fieldColumns": []any{}}
	}
	return map[string]any{"kind": "question", "question": map[string]any{"signal": "spans", "environment": environment, "fromMs": 1000000, "toMs": 1060000, "limit": 100, "offset": offset}, "view": "Records"}
}
func key(v map[string]any) string {
	kind := v["kind"].(string)
	scope := v["question"]
	if kind == "activity" {
		scope = v["selector"]
	}
	return kind + ":" + scope.(map[string]any)["environment"].(string)
}
func (o *oracle) step(c command) state {
	rejected := false
	find := func(k string) int { return slices.IndexFunc(o.items, func(v item) bool { return v.Key == k }) }
	save := false
	switch c.Name {
	case "advance":
		o.now += c.MS
	case "reload":
		o.items = []item{}
		o.durable = o.stored
		if o.stored {
			o.items = slices.Clone(o.disk)
		}
	case "persistence":
		o.durable = c.Enabled
		if c.Enabled {
			save = true
		} else {
			o.disk = nil
			o.stored = false
		}
	case "reset":
		o.items = []item{}
		o.disk = nil
		o.stored = false
		o.durable = false
	case "clearRecent":
		o.items = slices.DeleteFunc(o.items, func(v item) bool { return !v.Pinned })
		save = true
	case "remember", "pin":
		k := key(c.Intent)
		i := find(k)
		wasPinned := i >= 0 && o.items[i].Pinned
		if c.Name == "pin" && !wasPinned && len(slices.DeleteFunc(slices.Clone(o.items), func(v item) bool { return !v.Pinned })) >= 20 {
			rejected = true
			break
		}
		if i >= 0 {
			o.items = slices.Delete(o.items, i, i+1)
		}
		var offset *int
		if c.Intent["kind"] == "question" {
			zero := 0
			offset = &zero
		}
		o.items = append([]item{{Key: k, UsedAt: o.now, Pinned: wasPinned || c.Name == "pin", Offset: offset}}, o.items...)
		save = true
	case "forget":
		k := key(c.Intent)
		o.items = slices.DeleteFunc(o.items, func(v item) bool { return v.Key == k })
		save = true
	case "unpin":
		if i := find(key(c.Intent)); i >= 0 {
			o.items[i].Pinned = false
		}
		save = true
	default:
		panic("unknown command")
	}
	o.bound()
	if save && o.durable {
		o.disk = slices.Clone(o.items)
		o.stored = true
	}
	return state{Durable: o.durable, Items: slices.Clone(o.items), Rejected: rejected}
}
func run(t interface {
	Helper()
	Fatalf(string, ...any)
}, input any) []state {
	t.Helper()
	wire, e := json.Marshal(input)
	if e != nil {
		t.Fatalf("encode: %v", e)
	}
	cmd := exec.Command("node", "bridge.mjs")
	cmd.Stdin = bytes.NewReader(wire)
	raw, e := cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("actual Journal: %v %s", e, raw)
	}
	var states []state
	if e = json.Unmarshal(raw, &states); e != nil {
		t.Fatalf("decode: %v %s", e, raw)
	}
	return states
}
func TestHistoryStateAdmissionRetentionAndExplicitPersistence(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		random := rand.New(rand.NewPCG(uint64(seed), 414))
		now := int64(1800000000000)
		commands := []command{{Name: "persistence", Enabled: true}}
		// Force capacity, duplicate-normalization and pinned TTL boundaries before
		// generated interleavings; random streams alone rarely fill all pin slots.
		for n := range 21 {
			commands = append(commands, command{Name: "pin", Intent: intent(n, n%3*100)})
		}
		commands = append(commands, command{Name: "advance", MS: 30*day + 1}, command{Name: "reload"}, command{Name: "clearRecent"})
		for range 100 {
			name := []string{"remember", "remember", "pin", "unpin", "forget", "advance", "reload", "persistence", "clearRecent", "reset"}[random.IntN(10)]
			commands = append(commands, command{Name: name, Intent: intent(random.IntN(30), random.IntN(3)*100), MS: []int64{0, 1, day, 30 * day, 30*day + 1}[random.IntN(5)], Enabled: random.IntN(2) == 0})
		}
		o := oracle{now: now, items: []item{}}
		expected := make([]state, 0, len(commands))
		for _, c := range commands {
			expected = append(expected, o.step(c))
		}
		poison := []any{}
		for _, field := range []string{"rows", "resultToken", "basis", "membership", "readAt"} {
			v := intent(1, 0)
			v["question"].(map[string]any)[field] = "unsafe"
			poison = append(poison, v)
		}
		for _, nested := range []string{"predicate", "field", "environment", "layout"} {
			v := intent(1, 0)
			q := v["question"].(map[string]any)
			switch nested {
			case "predicate", "field":
				f := map[string]any{"kind": "attribute", "key": "route"}
				p := map[string]any{"field": f, "op": "eq", "value": "checkout"}
				if nested == "field" {
					f["rows"] = []any{"secret"}
				} else {
					p["resultToken"] = "secret"
				}
				q["where"] = p
			case "environment":
				delete(q, "environment")
				q["environmentSelector"] = map[string]any{"version": "goo/environment-selector/v1", "policy": "effective", "scope": map[string]any{"mode": "unreported", "rows": []any{"secret"}}}
			case "layout":
				v = intent(0, 0)
				v["fieldColumns"] = []any{map[string]any{"kind": "attribute", "key": "route", "body": "secret"}}
			}
			poison = append(poison, v)
		}
		got := run(tc, map[string]any{"now": now, "commands": commands, "poison": poison})
		if len(got) != len(expected) {
			tc.Fatalf("states: got %d want %d", len(got), len(expected))
		}
		for i := range expected {
			if !reflect.DeepEqual(got[i], expected[i]) {
				tc.Fatalf("seed %d command %d %s: got %+v want %+v", seed, i, commands[i].Name, got[i], expected[i])
			}
		}
		// Rejecting controls prove comparison does not tolerate lost pins,
		// invented persistence, or continuation offsets in stored intent.
		bad := expected[20]
		bad.Items = slices.Clone(bad.Items)
		bad.Items = bad.Items[1:]
		if reflect.DeepEqual(bad, got[20]) {
			tc.Fatalf("lost-pin control accepted")
		}
		bad = expected[0]
		bad.Durable = !bad.Durable
		if reflect.DeepEqual(bad, got[0]) {
			tc.Fatalf("persistence control accepted")
		}
	}, hegel.WithTestCases(64), hegel.WithSeed(414), hegel.WithDatabase(filepath.Join(t.TempDir(), "hegel")))
}
