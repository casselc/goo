package activitylogs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand/v2"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

type gateState struct {
	Scope           *string
	Generation      int
	Loading, Frozen bool
	Result          json.RawMessage
}
type gateCommand struct {
	Name  string  `json:"name"`
	Scope *string `json:"scope"`
	Token int     `json:"token"`
}
type projection struct {
	Gate     []gateState
	Query    exploration.Query
	Timeline struct {
		Ceiling string
		Points  []struct{ Key, State, Value string }
	}
	Severity []struct {
		Value, Count string
		Share        struct{ Numerator, Denominator string }
	}
	Rate      struct{ Numerator, Denominator, Exact, Text string }
	Selection exploration.Query
}

func integer(v string) *big.Int {
	n, ok := new(big.Int).SetString(v, 10)
	if !ok {
		panic("noninteger fixture")
	}
	return n
}
func project(t interface {
	Helper()
	Fatalf(string, ...any)
}, input any) []projection {
	t.Helper()
	wire, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	cmd := exec.Command("node", "bridge.mjs")
	cmd.Stdin = bytes.NewReader(wire)
	raw, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("actual JS log model: %v %s", err, raw)
	}
	var out []projection
	if err = json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("model output: %v %s", err, raw)
	}
	return out
}
func fixture(seed int) (map[string]any, []string, map[string]string, int64, int, int) {
	from := int64(1000000 + seed)
	bins := make([]string, 6)
	total := new(big.Int)
	parts := []*big.Int{new(big.Int), new(big.Int), new(big.Int)}
	series := []any{}
	for i := range 6 {
		n := new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 53), big.NewInt(int64(seed+i)))
		if (seed+i)%4 == 0 || seed%7 == 0 {
			n.SetInt64(0)
		}
		bins[i] = n.String()
		total.Add(total, n)
		parts[i%3].Add(parts[i%3], n)
		state := "observed"
		if n.Sign() == 0 {
			state = "no_data"
		}
		series = append(series, map[string]any{"fromMs": from + int64(i)*10000, "toMs": from + int64(i+1)*10000, "state": state, "count": n.String()})
	}
	labels := []string{"", "ERROR", "error"}
	groups := []any{}
	counts := map[string]string{}
	for i, label := range labels {
		if parts[i].Sign() > 0 {
			groups = append(groups, map[string]any{"g0": label, "count": parts[i].String()})
			counts[label] = parts[i].String()
		}
	}
	field := map[string]string{"kind": "attribute", "key": "route.'雪"}
	where := map[string]any{"any": []any{map[string]any{"field": field, "op": "eq", "value": "alpha"}, map[string]any{"field": field, "op": "missing"}}}
	selector := map[string]any{"signal": "logs", "environment": "prod-'雪", "where": where}
	if seed%2 == 0 {
		delete(selector, "environment")
		selector["environmentSelector"] = map[string]any{"version": "goo/environment-selector/v1", "policy": "effective", "scope": map[string]any{"mode": "unreported"}}
	}
	q := map[string]any{"signal": "logs", "fromMs": from, "toMs": from + 60000, "limit": 1, "offset": 0, "groupBy": []any{map[string]string{"kind": "column", "key": "severity"}}, "bucketMs": 10000, "where": where}
	for _, k := range []string{"environment", "environmentSelector"} {
		if v, ok := selector[k]; ok {
			q[k] = v
		}
	}
	read := map[string]any{"readAt": "2026-11-12T12:00:00Z", "basis": strings.Repeat("a", 64)}
	state := "observed"
	if total.Sign() == 0 {
		state = "no_data"
	}
	dto := map[string]any{"version": "goo/record-aggregates/v1", "question": q, "capturedAt": "2026-11-12T12:00:00Z", "basis": strings.Repeat("b", 64), "count": total.String(), "state": state, "groups": groups, "timeSeries": series, "warnings": []string{}, "reads": map[string]any{"count": read, "groups": read, "timeSeries": read}}
	a, b := seed%6, (seed/6)%6
	return map[string]any{"context": map[string]any{"selector": selector, "readWindow": map[string]int64{"fromMs": from, "toMs": from + 60000}}, "dto": dto, "a": a, "b": b}, bins, counts, from, a, b
}

// The oracle tracks the lease as a small independent state machine. Generated
// command sequences include duplicate starts, stale completions, scope changes
// and pause/resume interleavings against the actual shared JS gate.
func gateSequence(seed int) ([]gateCommand, []gateState) {
	random := rand.New(rand.NewPCG(uint64(seed), 413))
	scopeA, scopeB, empty := "scope-a", "scope-b", ""
	scopes := []*string{nil, &empty, &scopeA, &scopeB}
	current := gateState{}
	commands := []gateCommand{}
	states := []gateState{}
	same := func(a, b *string) bool { return a == nil && b == nil || a != nil && b != nil && *a == *b }
	for i := range 60 {
		command := gateCommand{Name: []string{"bind", "start", "accept", "pause", "resume"}[random.IntN(5)], Scope: scopes[random.IntN(len(scopes))], Token: current.Generation + random.IntN(3) - 1}
		if i == 0 {
			command.Name = "bind"
			command.Scope = &scopeA
		}
		if i == 1 {
			command.Name = "start"
		}
		if i == 2 {
			command.Name = "accept"
			command.Scope = &scopeA
			command.Token = current.Generation
		}
		var result any
		switch command.Name {
		case "bind":
			changed := !same(command.Scope, current.Scope)
			result = changed
			if changed {
				current.Scope = command.Scope
				current.Generation++
				current.Loading = false
				current.Frozen = false
			}
		case "start":
			if current.Scope != nil && *current.Scope != "" && !current.Frozen && !current.Loading {
				current.Generation++
				current.Loading = true
				result = current.Generation
			}
		case "accept":
			accepted := command.Token == current.Generation && current.Loading && !current.Frozen && same(command.Scope, current.Scope)
			result = accepted
			if accepted {
				current.Loading = false
			}
		case "pause":
			current.Generation++
			current.Loading = false
			current.Frozen = true
		case "resume":
			current.Generation++
			current.Loading = false
			current.Frozen = false
		}
		current.Result, _ = json.Marshal(result)
		commands = append(commands, command)
		states = append(states, current)
	}
	return commands, states
}

func TestLogProjectionExactCountsAndFrozenWindow(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		input, bins, counts, from, a, b := fixture(seed)
		commands, expectedGate := gateSequence(seed)
		input["commands"] = commands
		got := project(tc, []any{input})
		if len(got) != 1 {
			tc.Fatal("one projection required")
		}
		p := got[0]
		if !reflect.DeepEqual(p.Gate, expectedGate) {
			tc.Fatal("actual shared ReadGate diverged from generated ownership state machine")
		}
		if err := exploration.ValidateRecordAggregatesQuery(p.Query); err != nil {
			tc.Fatal(err)
		}
		plan, err := exploration.Compile(p.Query)
		if err != nil {
			tc.Fatal(err)
		}
		if plan.Count == "" || plan.Groups == "" || plan.TimeSeries == "" || strings.Contains(plan.Groups, "prod-'雪") {
			tc.Fatal("three parameterized aggregate statements required")
		}
		if p.Query.Signal != "logs" || p.Query.BucketMS != 10000 || !reflect.DeepEqual(p.Query.GroupBy, []exploration.Field{{Kind: "column", Key: "severity"}}) {
			tc.Fatal("severity projection changed")
		}
		if len(p.Timeline.Points) != 6 {
			tc.Fatal("six full-window buckets required")
		}
		total, max := new(big.Int), big.NewInt(1)
		for i, point := range p.Timeline.Points {
			n := integer(bins[i])
			total.Add(total, n)
			if n.Cmp(max) > 0 {
				max.Set(n)
			}
			if point.Key != fmt.Sprint(i) || point.Value != bins[i] || point.State != map[bool]string{true: "no_data", false: "observed"}[n.Sign() == 0] {
				tc.Fatal("bucket count/state changed")
			}
		}
		if p.Timeline.Ceiling != max.String() {
			tc.Fatal("scale changed")
		}
		if len(p.Severity) != len(counts) {
			tc.Fatal("exact reported text identities collapsed")
		}
		for i, g := range p.Severity {
			if counts[g.Value] != g.Count || g.Share.Numerator != g.Count || g.Share.Denominator != total.String() {
				tc.Fatal("reported text population changed")
			}
			if i > 0 {
				prior := p.Severity[i-1]
				cmp := integer(prior.Count).Cmp(integer(g.Count))
				if cmp < 0 || cmp == 0 && prior.Value > g.Value {
					tc.Fatal("severity ranking lost exact integer order")
				}
			}
		}
		rate := new(big.Rat).SetFrac(total, big.NewInt(60))
		if p.Rate.Numerator != rate.Num().String() || p.Rate.Denominator != rate.Denom().String() {
			tc.Fatal("rounded or inferred rate")
		}
		rounded := new(big.Int).Add(new(big.Int).Mul(total, big.NewInt(1000)), big.NewInt(30))
		rounded.Quo(rounded, big.NewInt(60))
		whole, remainder := new(big.Int), new(big.Int)
		whole.QuoRem(rounded, big.NewInt(1000), remainder)
		if p.Rate.Exact != rate.RatString() || p.Rate.Text != fmt.Sprintf("%s.%03d logs/s", whole.String(), remainder.Int64()) {
			tc.Fatal("rate display rounded before exact arithmetic")
		}

		low, high := min(a, b), maxIndex(a, b)
		if p.Selection.FromMS != from+int64(low)*10000 || p.Selection.ToMS != from+int64(high+1)*10000 {
			tc.Fatal("selection lost frozen UTC bounds")
		}
		expected := p.Query
		expected.FromMS = p.Selection.FromMS
		expected.ToMS = p.Selection.ToMS
		expected.Limit = 100
		expected.GroupBy = nil
		expected.BucketMS = 0
		if !reflect.DeepEqual(expected, p.Selection) {
			tc.Fatal("selection changed selector, OR or environment policy")
		}
		if err = p.Selection.Validate(); err != nil {
			tc.Fatal(err)
		}
	}, hegel.WithTestCases(64), hegel.WithSeed(413), hegel.WithDatabase(filepath.Join(t.TempDir(), "hegel")))
}
func maxIndex(a, b int) int {
	if a > b {
		return a
	}
	return b
}
