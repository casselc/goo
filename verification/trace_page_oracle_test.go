package verification

import (
	"encoding/json"
	"math/big"
	"sort"
	"strings"

	"github.com/casselc/goo/exploration"
)

// This oracle partitions endpoint cells and uses independent reachability for
// cycle membership. It deliberately does not call AnalyzeTracePage or merge
// intervals. A child's residual is never an input to its parent's arithmetic.
func tpbOracle(rows []json.RawMessage, truncated bool) exploration.TracePageBreakdown {
	type fact struct {
		trace, id, parent         string
		start, end, duration      *big.Int
		identity, parentOK, valid bool
		service, operation        *string
		reasons                   map[string]bool
	}
	facts := make([]fact, len(rows))
	byID := map[string][]int{}
	integer := func(v any) *big.Int {
		s, ok := v.(string)
		if !ok || s == "" || (len(s) > 1 && s[0] == '0') {
			return nil
		}
		for _, c := range s {
			if c < '0' || c > '9' {
				return nil
			}
		}
		n, ok := new(big.Int).SetString(s, 10)
		if !ok || n.BitLen() > 64 {
			return nil
		}
		return n
	}
	id := func(s string, n int) bool {
		if len(s) != n || strings.Trim(s, "0") == "" {
			return false
		}
		for _, c := range s {
			if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
				return false
			}
		}
		return true
	}
	for i, raw := range rows {
		f := fact{valid: true, reasons: map[string]bool{}}
		var obj map[string]any
		malformed := json.Unmarshal(raw, &obj) != nil || obj == nil
		text := func(k string) (string, bool) {
			v, ok := obj[k]
			if !ok {
				return "", false
			}
			s, ok := v.(string)
			return s, ok
		}
		f.trace, _ = text("TraceId")
		f.id, _ = text("SpanId")
		f.parent, _ = text("ParentSpanId")
		_, parentExists := obj["ParentSpanId"]
		_, parentText := text("ParentSpanId")
		f.parentOK = !parentExists || parentText && (f.parent == "" || f.parent == "0000000000000000" || id(f.parent, 16))
		f.identity = id(f.trace, 32) && id(f.id, 16)
		f.start = integer(obj["timestamp_ns"])
		f.duration = integer(obj["duration_ns"])
		if f.start != nil && f.duration != nil {
			f.end = new(big.Int).Add(f.start, f.duration)
			if f.end.BitLen() > 64 {
				f.end = nil
			}
		}
		if f.end == nil {
			f.reasons["invalid-interval"] = true
			f.valid = false
		}
		if !f.identity {
			f.reasons["invalid-identity"] = true
			f.valid = false
		} else {
			byID[f.trace+":"+f.id] = append(byID[f.trace+":"+f.id], i)
		}
		if !f.parentOK {
			f.reasons["invalid-parent-identity"] = true
			f.valid = false
		}
		for k, dest := range map[string]**string{"ServiceName": &f.service, "SpanName": &f.operation} {
			if v, exists := obj[k]; exists && v != nil {
				s, ok := v.(string)
				if !ok {
					malformed = true
				} else {
					*dest = &s
				}
			}
		}
		if malformed {
			f.reasons["invalid-row"] = true
			f.valid = false
			f.end = nil
		}
		facts[i] = f
	}
	parent := make([]int, len(rows))
	for i := range parent {
		parent[i] = -1
		f := &facts[i]
		if f.identity && len(byID[f.trace+":"+f.id]) > 1 {
			f.valid = false
			f.reasons["duplicate-identity"] = true
		}
		if f.parentOK && f.parent != "" && f.parent != "0000000000000000" && id(f.trace, 32) {
			matches := byID[f.trace+":"+f.parent]
			if len(matches) == 0 {
				f.reasons["parent-outside-page"] = true
			} else if len(matches) > 1 {
				f.valid = false
				f.reasons["ambiguous-parent"] = true
			} else {
				parent[i] = matches[0]
			}
		}
	}
	cycles := make([]bool, len(rows))
	for origin := range rows {
		at := parent[origin]
		seen := map[int]bool{}
		for at >= 0 && !seen[at] {
			if at == origin {
				cycles[origin] = true
				break
			}
			seen[at] = true
			at = parent[at]
		}
		if cycles[origin] {
			facts[origin].valid = false
			facts[origin].reasons["parent-cycle"] = true
		}
	}
	out := exploration.TracePageBreakdown{Version: "goo/trace-page-breakdown/v1", State: "observed", Truncated: truncated, Rows: []exploration.TraceSpanBreakdown{}, Groups: []exploration.TracePageBreakdownGroup{}}
	if len(rows) == 0 {
		out.State = "no_data"
	}
	for i, f := range facts {
		r := exploration.TraceSpanBreakdown{RowIndex: i, Service: f.service, Operation: f.operation, State: "available", Reasons: []string{}}
		if f.duration != nil {
			s := f.duration.String()
			r.InclusiveNS = &s
		}
		covered := [][2]uint64{}
		valid := f.valid
		for j, c := range facts {
			// Every reported, well-formed parent ID attaches the observation to all
			// candidate parents; ambiguity must withhold rather than choose one.
			if !c.parentOK || c.parent == "" || c.parent == "0000000000000000" || c.trace != f.trace || c.parent != f.id || !id(c.trace, 32) || !f.identity {
				continue
			}
			if !c.identity || len(byID[c.trace+":"+c.id]) != 1 || c.end == nil || cycles[j] || parent[j] != i {
				valid = false
				f.reasons["unusable-child-edge"] = true
				continue
			}
			if f.end != nil && (c.start.Cmp(f.start) < 0 || c.end.Cmp(f.end) > 0) {
				valid = false
				f.reasons["child-outside-parent"] = true
				continue
			}
			covered = append(covered, [2]uint64{c.start.Uint64(), c.end.Uint64()})
		}
		if valid {
			u := new(big.Int).SetUint64(tpbPartition(covered))
			a, b := u.String(), new(big.Int).Sub(f.duration, u).String()
			r.ChildUnionNS = &a
			r.OutsideChildrenNS = &b
		} else {
			r.State = "unavailable"
			out.State = "partial"
		}
		for reason := range f.reasons {
			r.Reasons = append(r.Reasons, reason)
		}
		sort.Strings(r.Reasons)
		out.Rows = append(out.Rows, r)
	}
	groups := map[string]*exploration.TracePageBreakdownGroup{}
	for _, r := range out.Rows {
		keyBytes, _ := json.Marshal([2]*string{r.Service, r.Operation})
		key := string(keyBytes)
		g := groups[key]
		if g == nil {
			g = &exploration.TracePageBreakdownGroup{Service: r.Service, Operation: r.Operation, RowIndexes: []int{}, State: "observed"}
			groups[key] = g
		}
		g.RowIndexes = append(g.RowIndexes, r.RowIndex)
		add := func(sum **string, v *string) {
			if v == nil {
				return
			}
			n := new(big.Int)
			if *sum != nil {
				n.SetString(**sum, 10)
			}
			x, _ := new(big.Int).SetString(*v, 10)
			s := n.Add(n, x).String()
			*sum = &s
		}
		if r.InclusiveNS == nil {
			g.InclusiveUnknown++
		} else {
			g.InclusiveKnown++
			add(&g.InclusiveNS, r.InclusiveNS)
		}
		if r.OutsideChildrenNS == nil {
			g.OutsideUnknown++
		} else {
			g.OutsideKnown++
			add(&g.OutsideChildrenNS, r.OutsideChildrenNS)
		}
		if r.InclusiveNS == nil || r.OutsideChildrenNS == nil {
			g.State = "partial"
		}
	}
	for _, g := range groups {
		out.Groups = append(out.Groups, *g)
	}
	compare := func(a, b *string) int {
		if a == nil {
			if b == nil {
				return 0
			}
			return -1
		}
		if b == nil {
			return 1
		}
		return strings.Compare(*a, *b)
	}
	sort.Slice(out.Groups, func(i, j int) bool {
		a, b := out.Groups[i], out.Groups[j]
		if c := compare(a.Service, b.Service); c != 0 {
			return c < 0
		}
		return compare(a.Operation, b.Operation) < 0
	})
	return out
}
func tpbOracleEqual(rows []json.RawMessage, truncated bool, got exploration.TracePageBreakdown) bool {
	want := tpbOracle(rows, truncated)
	// Reason order is presentation-only; Go/JS whole DTO parity separately checks
	// the stable ordering. This semantic oracle requires the exact reason sets.
	got.Rows = append([]exploration.TraceSpanBreakdown{}, got.Rows...)
	for i := range got.Rows {
		got.Rows[i].Reasons = append([]string{}, got.Rows[i].Reasons...)
		sort.Strings(got.Rows[i].Reasons)
	}
	return aeJSONEqual(want, got)
}
