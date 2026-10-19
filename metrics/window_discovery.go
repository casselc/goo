package metrics

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strconv"
)

// SupportedWindow is an explicit sub-scope, never a replacement for the input scope.
type SupportedWindow struct {
	FromNS    string         `json:"fromNS"`
	ToNS      string         `json:"toNS"`
	Signature string         `json:"signature"`
	Members   []WindowMember `json:"members"`
}
type WindowMember struct {
	Stream  Stream         `json:"stream"`
	Summary *WindowSummary `json:"summary"`
}
type WindowExclusion struct {
	Stream Stream `json:"stream"`
	FromNS string `json:"fromNS"`
	ToNS   string `json:"toNS"`
	Reason string `json:"reason"`
}
type WindowDiscovery struct {
	FromNS         string            `json:"fromNS"`
	ToNS           string            `json:"toNS"`
	Streams        int               `json:"streams"`
	Windows        []SupportedWindow `json:"windows"`
	Exclusions     []WindowExclusion `json:"exclusions"`
	ExclusionCount int               `json:"exclusionCount"`
	WindowCount    int               `json:"windowCount"`
	Warnings       []string          `json:"warnings"`
}
type supportInterval struct {
	from, to, state, reason string
	counter                 *CounterInterval
	histogram               *HistogramInterval
}
type supportSeries struct {
	stream    Stream
	temp      string
	intervals []supportInterval
}

// DiscoverCounterWindows uses validated kernel intervals only, including reset taint.
func DiscoverCounterWindows(r *CounterResult) WindowDiscovery {
	if r == nil {
		return discoverWindows("", "", nil, false)
	}
	ss := []supportSeries{}
	for _, s := range r.Series {
		v := supportSeries{stream: s.Stream, temp: s.Temporality}
		for j := range s.Intervals {
			i := &s.Intervals[j]
			v.intervals = append(v.intervals, supportInterval{i.FromNS, i.ToNS, i.State, i.Reason, i, nil})
		}
		ss = append(ss, v)
	}
	return discoverWindows(r.FromNS, r.ToNS, ss, false)
}
func DiscoverHistogramWindows(r *HistogramResult) WindowDiscovery {
	if r == nil {
		return discoverWindows("", "", nil, true)
	}
	ss := []supportSeries{}
	for _, s := range r.Series {
		v := supportSeries{stream: s.Stream, temp: s.Temporality}
		for j := range s.Intervals {
			i := &s.Intervals[j]
			state, reason := i.State, i.Reason
			if state == "ok" && i.Distribution != nil && i.Distribution.Count == "0" {
				state, reason = "no_data", "empty-distribution"
			}
			v.intervals = append(v.intervals, supportInterval{i.FromNS, i.ToNS, state, reason, nil, i})
		}
		ss = append(ss, v)
	}
	return discoverWindows(r.FromNS, r.ToNS, ss, true)
}
func discoverWindows(from, to string, ss []supportSeries, hist bool) WindowDiscovery {
	out := WindowDiscovery{FromNS: from, ToNS: to, Streams: len(ss), Windows: []SupportedWindow{}, Exclusions: []WindowExclusion{}, Warnings: []string{"Support describes retained observations, not ingestion completeness. Every retained stream must share exact boundaries. No interpolation, rebinning or member selection."}}
	exclude := func(s Stream, a, b, reason string) {
		out.ExclusionCount++
		if len(out.Exclusions) < 64 {
			out.Exclusions = append(out.Exclusions, WindowExclusion{s, a, b, reason})
		}
	}
	if len(ss) == 0 {
		out.Warnings = append(out.Warnings, "No retained native streams.")
		return out
	}
	sort.Slice(ss, func(i, j int) bool {
		a, _ := json.Marshal(ss[i].stream)
		b, _ := json.Marshal(ss[j].stream)
		return string(a) < string(b)
	})
	common := map[string]int{}
	for _, s := range ss {
		edges := map[string]bool{}
		cursor := from
		for _, i := range s.intervals {
			if i.from != cursor {
				exclude(s.stream, cursor, i.from, "missing-coverage")
			}
			if i.state == "ok" {
				edges[i.from] = true
				edges[i.to] = true
			} else {
				reason := i.reason
				if reason == "" {
					reason = "unknown-interval"
				}
				exclude(s.stream, i.from, i.to, reason)
			}
			cursor = i.to
		}
		if cursor != to {
			exclude(s.stream, cursor, to, "missing-coverage")
		}
		for edge := range edges {
			common[edge]++
		}
	}
	edges := []string{}
	for edge, n := range common {
		if n == len(ss) {
			edges = append(edges, edge)
		}
	}
	sort.Slice(edges, func(i, j int) bool {
		a, _ := strconv.ParseInt(edges[i], 10, 64)
		b, _ := strconv.ParseInt(edges[j], 10, 64)
		return a < b
	})
	cursors := make([]int, len(ss))
	var pending *SupportedWindow
	summarize := func(a, b string, parts [][]supportInterval) (SupportedWindow, bool) {
		w := SupportedWindow{FromNS: a, ToNS: b, Members: []WindowMember{}}
		keys := []any{}
		for n, s := range ss {
			for _, i := range parts[n] {
				if i.state != "ok" {
					return w, false
				}
			}
			var summary *WindowSummary
			if hist {
				series := HistogramSeries{Stream: s.stream, Temporality: s.temp, State: "ok"}
				for _, i := range parts[n] {
					series.Intervals = append(series.Intervals, *i.histogram)
				}
				summary = histogramWindow(a, b, series)
			} else {
				series := CounterSeries{Stream: s.stream, Temporality: s.temp, State: "ok"}
				for _, i := range parts[n] {
					series.Intervals = append(series.Intervals, *i.counter)
				}
				summary = counterWindow(a, b, series)
			}
			if summary.State != "observed" {
				return w, false
			}
			var bounds []float64
			if summary.Distribution != nil {
				bounds = summary.Distribution.Bounds
			}
			keys = append(keys, struct {
				Stream Stream
				Temp   string
				Bounds []float64
			}{s.stream, s.temp, bounds})
			w.Members = append(w.Members, WindowMember{s.stream, summary})
		}
		raw, _ := json.Marshal(keys)
		hash := sha256.Sum256(raw)
		w.Signature = hex.EncodeToString(hash[:])
		return w, true
	}
	var pendingParts [][]supportInterval
	flush := func() {
		if pending == nil {
			return
		}
		w, ok := summarize(pending.FromNS, pending.ToNS, pendingParts)
		if ok {
			out.WindowCount++
			if len(out.Windows) < 32 {
				out.Windows = append(out.Windows, w)
			}
		} else {
			exclude(ss[len(w.Members)].stream, pending.FromNS, pending.ToNS, "aggregate-unavailable")
		}
		pending = nil
	}
	for k := 0; k+1 < len(edges); k++ {
		a, b := edges[k], edges[k+1]
		parts := make([][]supportInterval, len(ss))
		for n, s := range ss {
			for cursors[n] < len(s.intervals) {
				i := s.intervals[cursors[n]]
				end, _ := strconv.ParseInt(i.to, 10, 64)
				start, _ := strconv.ParseInt(a, 10, 64)
				if end > start {
					break
				}
				cursors[n]++
			}
			j := cursors[n]
			for j < len(s.intervals) {
				i := s.intervals[j]
				end, _ := strconv.ParseInt(i.to, 10, 64)
				limit, _ := strconv.ParseInt(b, 10, 64)
				if end > limit {
					break
				}
				parts[n] = append(parts[n], i)
				j++
			}
		}
		w, ok := summarize(a, b, parts)
		if !ok {
			flush()
			continue
		}
		if pending != nil && pending.ToNS == a && pending.Signature == w.Signature {
			for n := range ss {
				pendingParts[n] = append(pendingParts[n], parts[n]...)
			}
			pending.ToNS = b
			continue
		}
		flush()
		pending = &w
		pendingParts = parts
	}
	flush()
	if len(out.Windows) == 0 {
		out.Warnings = append(out.Warnings, "No complete common-boundary window is supported by every retained stream.")
	}
	if out.WindowCount > len(out.Windows) {
		out.Warnings = append(out.Warnings, "Showing the first 32 supported windows; narrow the original question to inspect later support.")
	}
	if out.ExclusionCount > len(out.Exclusions) {
		out.Warnings = append(out.Warnings, "Showing the first 64 exclusions; counts include all exclusions.")
	}
	return out
}
