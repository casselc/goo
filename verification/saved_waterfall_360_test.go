package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"hegel.dev/go/hegel"
)

type savedWaterfallSchedule360 struct {
	mask    int
	actions []map[string]any
}

func (m *savedWaterfallSchedule360) RuleKind(tc hegel.TestCase) {
	if m.mask&1 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "kind", "value": []string{"record-table", "record-span-waterfall"}[hegel.Draw(tc, hegel.Integers(0, 1))]})
	}
}
func (m *savedWaterfallSchedule360) RuleInspect(tc hegel.TestCase) {
	if m.mask&2 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "inspect", "index": hegel.Draw(tc, hegel.Integers(0, 15))})
	}
}
func (m *savedWaterfallSchedule360) RuleSelection(tc hegel.TestCase) {
	if m.mask&4 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "selection", "selected": hegel.Draw(tc, hegel.Integers(0, 2)), "busy": hegel.Draw(tc, hegel.Booleans())})
	}
}
func (m *savedWaterfallSchedule360) RulePage(tc hegel.TestCase) {
	if m.mask&8 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "page", "offset": 100 * hegel.Draw(tc, hegel.Integers(0, 3))})
	}
}
func (m *savedWaterfallSchedule360) RuleRetire(tc hegel.TestCase) {
	if m.mask&16 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "retire"})
	}
}
func (m *savedWaterfallSchedule360) RuleResume(tc hegel.TestCase) {
	if m.mask&32 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "resume"})
	}
}
func (m *savedWaterfallSchedule360) RuleScope(tc hegel.TestCase) {
	if m.mask&64 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "scope", "kind": hegel.Draw(tc, hegel.Integers(0, 2))})
	}
}
func (m *savedWaterfallSchedule360) RuleOld(tc hegel.TestCase) {
	if m.mask&128 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "old"})
	}
}
func (m *savedWaterfallSchedule360) RuleHidden(tc hegel.TestCase) {
	if m.mask&256 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "hidden"})
	}
}
func (m *savedWaterfallSchedule360) RuleCorrupt(tc hegel.TestCase) {
	if m.mask&512 != 0 {
		m.actions = append(m.actions, map[string]any{"type": "corrupt"})
	}
}
func waterfallNS360(value any) *big.Int {
	s, ok := value.(string)
	if !ok || s == "" || len(s) > 20 || len(s) > 1 && s[0] == '0' {
		return nil
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return nil
		}
	}
	n, ok := new(big.Int).SetString(s, 10)
	if !ok || n.BitLen() > 64 {
		return nil
	}
	return n
}
func waterfallID360(value any, size int) bool {
	s, ok := value.(string)
	if !ok || len(s) != size || strings.Trim(s, "0") == "" {
		return false
	}
	for _, c := range s {
		if !strings.ContainsRune("0123456789abcdef", c) {
			return false
		}
	}
	return true
}
func waterfallOracle360(rows []map[string]any) []map[string]any {
	origins, ends := map[string]*big.Int{}, map[string]*big.Int{}
	keys := make([]string, len(rows))
	for i, r := range rows {
		key := fmt.Sprintf("observation:%d", i)
		if waterfallID360(r["TraceId"], 32) {
			key = "trace:" + r["TraceId"].(string)
		}
		keys[i] = key
		start, duration := waterfallNS360(r["timestamp_ns"]), waterfallNS360(r["duration_ns"])
		if start == nil || duration == nil {
			continue
		}
		end := new(big.Int).Add(start, duration)
		if origins[key] == nil || start.Cmp(origins[key]) < 0 {
			origins[key] = new(big.Int).Set(start)
		}
		if ends[key] == nil || end.Cmp(ends[key]) > 0 {
			ends[key] = end
		}
	}
	scale := new(big.Int)
	known := false
	for key, origin := range origins {
		extent := new(big.Int).Sub(ends[key], origin)
		known = true
		if extent.Cmp(scale) > 0 {
			scale.Set(extent)
		}
	}
	rangeNS := new(big.Int).Set(scale)
	if rangeNS.Sign() == 0 {
		rangeNS.SetInt64(1)
	}
	positions := map[string][]int{}
	for i, r := range rows {
		if waterfallID360(r["TraceId"], 32) && waterfallID360(r["SpanId"], 16) {
			key := r["TraceId"].(string) + ":" + r["SpanId"].(string)
			positions[key] = append(positions[key], i)
		}
	}
	parents := make([]int, len(rows))
	states := make([]string, len(rows))
	for i, r := range rows {
		parents[i] = -1
		states[i] = "parent-unreported"
		if !waterfallID360(r["TraceId"], 32) {
			states[i] = "trace-unreported"
			continue
		}
		if !waterfallID360(r["SpanId"], 16) {
			states[i] = "span-unreported"
			continue
		}
		p, present := r["ParentSpanId"]
		if !present || p == nil {
			continue
		}
		if p == "" || p == strings.Repeat("0", 16) {
			states[i] = "no-reported-parent"
			continue
		}
		if !waterfallID360(p, 16) {
			states[i] = "invalid-parent"
			continue
		}
		owners := positions[r["TraceId"].(string)+":"+p.(string)]
		if len(owners) == 0 {
			states[i] = "parent-not-held"
		} else if len(owners) > 1 {
			states[i] = "ambiguous-parent"
		} else {
			states[i] = "held-parent"
			parents[i] = owners[0]
		}
	}
	out := make([]map[string]any, len(rows))
	for i, r := range rows {
		start, duration := waterfallNS360(r["timestamp_ns"]), waterfallNS360(r["duration_ns"])
		item := map[string]any{"index": i, "groupKey": keys[i], "timingValid": start != nil && duration != nil, "parentState": states[i], "parentIndex": nil, "left": nil, "width": nil, "scaleExtent": nil, "cycle": false}
		if known {
			item["scaleExtent"] = scale.String()
		}
		if parents[i] >= 0 {
			item["parentIndex"] = parents[i]
		}
		visited := map[int]bool{i: true}
		p := parents[i]
		for p >= 0 {
			if visited[p] {
				item["cycle"] = true
				break
			}
			visited[p] = true
			p = parents[p]
		}
		if start != nil && duration != nil {
			percent := func(n *big.Int) float64 {
				v := new(big.Int).Mul(n, big.NewInt(100000000))
				v.Quo(v, rangeNS)
				return float64(v.Int64()) / 1000000
			}
			item["left"] = percent(new(big.Int).Sub(start, origins[keys[i]]))
			item["width"] = percent(duration)
		}
		out[i] = item
	}
	return out
}
func waterfallRows360(seed int) []map[string]any {
	origin := big.NewInt(1700000000000000000 + int64(seed)*1000000)
	id := func(n int) string { return fmt.Sprintf("%016x", n) }
	trace := func(n int) string { return fmt.Sprintf("%032x", n) }
	rows := []map[string]any{}
	add := func(t string, s, p any, delta, duration any) {
		var timestamp any = delta
		if n, ok := delta.(int64); ok {
			timestamp = new(big.Int).Add(origin, big.NewInt(n)).String()
		}
		var length any = duration
		if n, ok := duration.(int64); ok {
			length = strconv.FormatInt(n, 10)
		}
		rows = append(rows, map[string]any{"TraceId": t, "SpanId": s, "ParentSpanId": p, "timestamp_ns": timestamp, "duration_ns": length, "ServiceName": fmt.Sprintf("service-%d", len(rows)%3), "SpanName": "observation " + strconv.Itoa(len(rows)), "StatusCode": "Ok", "Attributes": map[string]any{"original.position": strconv.Itoa(len(rows)), "empty": "", "null": nil}, "ResourceAttributes": map[string]string{"service.namespace": "team"}})
	}
	add(trace(1), id(2), id(1), int64(10), int64(50))
	add(trace(1), id(1), "", int64(0), int64(100))
	add(trace(2), id(1), "", int64(30000000000), int64(200))
	add(trace(1), id(4), "", int64(2), int64(0))
	add(trace(1), id(4), "", int64(3), int64(1))
	add(trace(1), id(5), id(4), int64(4), int64(1))
	add(trace(3), id(6), id(7), int64(0), int64(30))
	add(trace(3), id(7), id(6), int64(1), int64(1))
	add(trace(4), id(8), id(9), int64(0), int64(1))
	add("bad", id(9), "", int64(0), int64(1))
	add("bad", id(10), id(9), int64(1), int64(1))
	add(trace(5), id(11), nil, "bad", int64(20))
	add(trace(5), id(12), "", int64(0), nil)
	add(trace(6), id(13), "", int64(0), int64(0))
	add(strings.Repeat("0", 32), id(14), "", int64(0), "18446744073709551615")
	add(trace(1), "bad", "invalid", int64(1), int64(1))
	return rows
}
func TestHegelSavedWaterfall360(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 10000))
		mode := hegel.Draw(tc, hegel.Integers(0, 4))
		m := &savedWaterfallSchedule360{mask: hegel.Draw(tc, hegel.Integers(1, 1023))}
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(24))
		q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1700000000000, ToMS: 1700000060000, Limit: 100, Offset: 0, Where: &exploration.Predicate{Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "eq", Value: "team"}}
		if mode > 0 {
			s := identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}}
			if mode == 1 {
				s.Policy = "current-only"
			}
			if mode == 3 {
				s.Scope = identity.EnvironmentScope{Mode: "empty"}
			}
			if mode == 4 {
				s.Scope = identity.EnvironmentScope{Mode: "unreported"}
			}
			q.Environment = ""
			q.EnvironmentSelector = &s
		}
		if err := q.Validate(); err != nil {
			tc.Fatal(err)
		}
		rows := waterfallRows360(seed)
		raw, _ := json.Marshal(map[string]any{"question": q, "rows": rows, "oracle": waterfallOracle360(rows), "actions": m.actions})
		cmd := exec.CommandContext(t.Context(), "node", "--experimental-test-module-mocks", "saved_waterfall_360_fixture.mjs")
		cmd.Stdin = bytes.NewReader(raw)
		out, err := cmd.Output()
		if err != nil {
			if f, ok := err.(*exec.ExitError); ok {
				tc.Fatal("actual savedowner/page waterfall", err, string(f.Stderr))
			} else {
				tc.Fatal(err)
			}
			return
		}
		var verdict struct {
			Complete                       bool
			Checks, Telemetry, Inspections int
		}
		if json.Unmarshal(out, &verdict) != nil || !verdict.Complete || verdict.Checks < 15 || verdict.Telemetry != 0 || verdict.Inspections < 2 {
			tc.Fatal("incomplete actual saved-page schedule", string(out))
		}
	}, options(40)...)
}
