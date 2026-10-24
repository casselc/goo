package exploration

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"
)

const TracePageBreakdownVersion = "goo/trace-page-breakdown/v1"

type TracePageBreakdown struct {
	Version   string                    `json:"version"`
	State     string                    `json:"state"`
	Truncated bool                      `json:"truncated"`
	Rows      []TraceSpanBreakdown      `json:"rows"`
	Groups    []TracePageBreakdownGroup `json:"groups"`
}
type TraceSpanBreakdown struct {
	RowIndex          int      `json:"rowIndex"`
	Service           *string  `json:"service"`
	Operation         *string  `json:"operation"`
	InclusiveNS       *string  `json:"inclusiveNs,omitempty"`
	ChildUnionNS      *string  `json:"childUnionNs,omitempty"`
	OutsideChildrenNS *string  `json:"outsideChildrenNs,omitempty"`
	State             string   `json:"state"`
	Reasons           []string `json:"reasons"`
}
type TracePageBreakdownGroup struct {
	Service           *string `json:"service"`
	Operation         *string `json:"operation"`
	RowIndexes        []int   `json:"rowIndexes"`
	InclusiveNS       *string `json:"inclusiveNs"`
	InclusiveKnown    int     `json:"inclusiveKnown"`
	InclusiveUnknown  int     `json:"inclusiveUnknown"`
	OutsideChildrenNS *string `json:"outsideChildrenNs"`
	OutsideKnown      int     `json:"outsideKnown"`
	OutsideUnknown    int     `json:"outsideUnknown"`
	State             string  `json:"state"`
}

type tracePageNode struct {
	trace, span, parent           string
	start, duration, end          uint64
	timing, identity, parentValid bool
	parentIndex                   int
	cycle                         bool
}

func canonicalNS(s string) (uint64, bool) {
	n, e := strconv.ParseUint(s, 10, 64)
	return n, e == nil && strconv.FormatUint(n, 10) == s
}

// AnalyzeTracePage is a source-free projection of supplied page observations.
// Union uses observed direct children only. It does not establish complete
// traces, CPU/self time, synchronous execution, or a critical path. Links are
// references and deliberately never create child edges. Group sums overlap.
func AnalyzeTracePage(rows []json.RawMessage, truncated bool) (*TracePageBreakdown, error) {
	if len(rows) > 1000 {
		return nil, fmt.Errorf("trace breakdown exceeds 1000 supplied rows")
	}
	out := &TracePageBreakdown{Version: TracePageBreakdownVersion, State: "observed", Truncated: truncated, Rows: []TraceSpanBreakdown{}, Groups: []TracePageBreakdownGroup{}}
	if len(rows) == 0 {
		out.State = "no_data"
		return out, nil
	}
	nodes := make([]tracePageNode, len(rows))
	keys := map[string][]int{}
	add := func(i int, reason string) {
		for _, r := range out.Rows[i].Reasons {
			if r == reason {
				return
			}
		}
		out.Rows[i].Reasons = append(out.Rows[i].Reasons, reason)
	}
	for i, raw := range rows {
		r := TraceSpanBreakdown{RowIndex: i, State: "available", Reasons: []string{}}
		n := tracePageNode{parentIndex: -1}
		var object map[string]json.RawMessage
		malformed := json.Unmarshal(raw, &object) != nil || object == nil
		read := func(key string, optional bool) (string, bool) {
			value, exists := object[key]
			if !exists {
				return "", optional
			}
			var text string
			if string(value) == "null" || json.Unmarshal(value, &text) != nil {
				return "", false
			}
			return text, true
		}
		n.trace, _ = read("TraceId", false)
		n.span, _ = read("SpanId", false)
		var parentOK bool
		n.parent, parentOK = read("ParentSpanId", true)
		for key, destination := range map[string]**string{"ServiceName": &r.Service, "SpanName": &r.Operation} {
			if value, present := object[key]; present && string(value) != "null" {
				var text string
				if json.Unmarshal(value, &text) != nil {
					malformed = true
				} else {
					*destination = &text
				}
			}
		}
		startText, startOK := read("timestamp_ns", false)
		durationText, durationOK := read("duration_ns", false)
		start, sok := canonicalNS(startText)
		duration, dok := canonicalNS(durationText)
		if durationOK && dok {
			v := durationText
			r.InclusiveNS = &v
		}
		n.start, n.duration = start, duration
		n.timing = startOK && durationOK && sok && dok && duration <= math.MaxUint64-start
		if n.timing {
			n.end = start + duration
		} else {
			r.State = "unavailable"
			r.Reasons = append(r.Reasons, "invalid-interval")
		}
		n.identity = exampleID(n.trace, 32) && exampleID(n.span, 16)
		if !n.identity {
			r.State = "unavailable"
			r.Reasons = append(r.Reasons, "invalid-identity")
		} else {
			key := n.trace + ":" + n.span
			keys[key] = append(keys[key], i)
		}
		n.parentValid = parentOK && (n.parent == "" || n.parent == "0000000000000000" || exampleID(n.parent, 16))
		if !n.parentValid {
			r.State = "unavailable"
			r.Reasons = append(r.Reasons, "invalid-parent-identity")
		}
		if malformed {
			r.State = "unavailable"
			r.Reasons = append(r.Reasons, "invalid-row")
			n.timing = false
		}

		nodes[i] = n
		out.Rows = append(out.Rows, r)
	}
	// Resolve only unique reported parent identities, never arbitrary duplicates.
	for i := range nodes {
		n := &nodes[i]
		if n.identity && len(keys[n.trace+":"+n.span]) > 1 {
			out.Rows[i].State = "unavailable"
			add(i, "duplicate-identity")
		}
		if n.parent != "" && n.parent != "0000000000000000" && n.parentValid && exampleID(n.trace, 32) {
			parents := keys[n.trace+":"+n.parent]
			if len(parents) == 0 {
				add(i, "parent-outside-page")
			} else if len(parents) > 1 {
				out.Rows[i].State = "unavailable"
				add(i, "ambiguous-parent")
			} else {
				n.parentIndex = parents[0]
			}
		}
	}
	// Identify exact cycle members. Descendants are not cycle members merely
	// because an ancestor's own residual cannot be calculated.
	for i := range nodes {
		path := []int{}
		seen := map[int]int{}
		current := i
		for current >= 0 {
			if at, ok := seen[current]; ok {
				for _, member := range path[at:] {
					nodes[member].cycle = true
				}
				break
			}
			seen[current] = len(path)
			path = append(path, current)
			current = nodes[current].parentIndex
		}
	}
	for i, n := range nodes {
		if n.cycle {
			out.Rows[i].State = "unavailable"
			add(i, "parent-cycle")
		}
	}
	children := make([][]int, len(nodes))
	for i, n := range nodes {
		if n.parent != "" && n.parent != "0000000000000000" && n.parentValid && exampleID(n.trace, 32) {
			for _, parent := range keys[n.trace+":"+n.parent] {
				children[parent] = append(children[parent], i)
			}
		}
	}
	type interval struct{ start, end uint64 }
	for i, n := range nodes {
		available := out.Rows[i].State == "available"
		intervals := []interval{}
		for _, childIndex := range children[i] {
			child := nodes[childIndex]
			// The child's own unknown residual does not make its valid interval
			// unusable. Only interval/edge identity authority determines subtraction.
			if !child.identity || len(keys[child.trace+":"+child.span]) != 1 || !child.timing || child.cycle || child.parentIndex != i {
				available = false
				add(i, "unusable-child-edge")
				continue
			}
			if n.timing && (child.start < n.start || child.end > n.end) {
				available = false
				add(i, "child-outside-parent")
				continue
			}
			intervals = append(intervals, interval{child.start, child.end})
		}
		if !available {
			out.Rows[i].State = "unavailable"
			out.State = "partial"
			continue
		}
		sort.Slice(intervals, func(a, b int) bool {
			if intervals[a].start != intervals[b].start {
				return intervals[a].start < intervals[b].start
			}
			return intervals[a].end < intervals[b].end
		})
		var union uint64
		if len(intervals) > 0 {
			start, end := intervals[0].start, intervals[0].end
			for _, v := range intervals[1:] {
				if v.start <= end {
					if v.end > end {
						end = v.end
					}
				} else {
					union += end - start
					start, end = v.start, v.end
				}
			}
			union += end - start
		}
		u, o := strconv.FormatUint(union, 10), strconv.FormatUint(n.duration-union, 10)
		out.Rows[i].ChildUnionNS = &u
		out.Rows[i].OutsideChildrenNS = &o
	}
	// Every valid duration is inclusive evidence even when identity or timestamp
	// authority is insufficient for a child-union measure.
	groups := map[string]*TracePageBreakdownGroup{}
	for _, r := range out.Rows {
		encodedKey, _ := json.Marshal([2]*string{r.Service, r.Operation})
		key := string(encodedKey)
		g := groups[key]
		if g == nil {
			g = &TracePageBreakdownGroup{Service: r.Service, Operation: r.Operation, RowIndexes: []int{}, State: "observed"}
			groups[key] = g
		}
		g.RowIndexes = append(g.RowIndexes, r.RowIndex)
		if r.InclusiveNS != nil {
			if g.InclusiveNS == nil {
				v := "0"
				g.InclusiveNS = &v
			}
			v := addNS(*g.InclusiveNS, *r.InclusiveNS)
			g.InclusiveNS = &v
			g.InclusiveKnown++
		} else {
			g.InclusiveUnknown++
		}
		if r.OutsideChildrenNS != nil {
			if g.OutsideChildrenNS == nil {
				v := "0"
				g.OutsideChildrenNS = &v
			}
			v := addNS(*g.OutsideChildrenNS, *r.OutsideChildrenNS)
			g.OutsideChildrenNS = &v
			g.OutsideKnown++
		} else {
			g.OutsideUnknown++
		}
		if r.InclusiveNS == nil || r.OutsideChildrenNS == nil {
			g.State = "partial"
		}
	}
	groupKeys := make([]string, 0, len(groups))
	for key := range groups {
		groupKeys = append(groupKeys, key)
	}
	sort.Slice(groupKeys, func(i, j int) bool {
		a, b := groups[groupKeys[i]], groups[groupKeys[j]]
		if compareOptionalTraceLabel(a.Service, b.Service) != 0 {
			return compareOptionalTraceLabel(a.Service, b.Service) < 0
		}
		return compareOptionalTraceLabel(a.Operation, b.Operation) < 0
	})
	for _, key := range groupKeys {
		out.Groups = append(out.Groups, *groups[key])
	}
	return out, nil
}
func addNS(a, b string) string {
	x, _ := new(big.Int).SetString(a, 10)
	y, _ := new(big.Int).SetString(b, 10)
	return x.Add(x, y).String()
}

// Optional labels sort missing first, then by UTF-8 byte order.
func compareOptionalTraceLabel(a, b *string) int {
	if a == nil {
		if b == nil {
			return 0
		}
		return -1
	}
	if b == nil {
		return 1
	}
	if *a < *b {
		return -1
	}
	if *a > *b {
		return 1
	}
	return 0
}
