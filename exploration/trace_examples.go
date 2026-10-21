package exploration

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
)

// TraceExampleSet describes a bounded record page, never a sample or full trace.
type TraceExampleSet struct {
	Version    string         `json:"version"`
	State      string         `json:"state"`
	Considered int            `json:"considered"`
	Excluded   int            `json:"excluded"`
	Items      []TraceExample `json:"items"`
}
type TraceExample struct {
	TraceID     string `json:"traceId"`
	SpanID      string `json:"spanId"`
	Service     string `json:"service"`
	Name        string `json:"name"`
	DurationNS  string `json:"durationNS"`
	TimestampNS string `json:"timestampNS"`
	Status      string `json:"status"`
	Reason      string `json:"reason"`
}

func exampleID(s string, n int) bool {
	return len(s) == n && strings.Trim(s, "0") != "" && strings.IndexFunc(s, func(r rune) bool { return !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f') }) < 0
}

// SelectTraceExamples uses exact uint64 durations and deterministic identity ties.
// Distinct trace IDs are selected at most once. No extra source query is made.
func SelectTraceExamples(r Result) *TraceExampleSet {
	if r.Question.Signal != "spans" {
		return nil
	}
	out := &TraceExampleSet{Version: "goo/trace-page-examples/v1", State: "unavailable", Considered: len(r.Rows), Items: []TraceExample{}}
	if r.Question.Validate() != nil || len(r.Rows) > CaptureRowLimit || r.State != "observed" {
		if r.State == "no_data" {
			out.State = "no_data"
		}
		return out
	}
	candidates := []TraceExample{}
	for _, raw := range r.Rows {
		var row struct {
			TraceID   string `json:"TraceId"`
			SpanID    string `json:"SpanId"`
			Service   string `json:"ServiceName"`
			Name      string `json:"SpanName"`
			Duration  string `json:"duration_ns"`
			Timestamp string `json:"timestamp_ns"`
			Status    string `json:"StatusCode"`
		}
		if json.Unmarshal(raw, &row) != nil {
			out.Excluded++
			continue
		}
		duration, err := strconv.ParseUint(row.Duration, 10, 64)
		timestamp, te := strconv.ParseUint(row.Timestamp, 10, 64)
		if err != nil || te != nil || strconv.FormatUint(duration, 10) != row.Duration || strconv.FormatUint(timestamp, 10) != row.Timestamp || timestamp/1000000 < uint64(r.Question.FromMS) || timestamp/1000000 >= uint64(r.Question.ToMS) || !exampleID(row.TraceID, 32) || !exampleID(row.SpanID, 16) {
			out.Excluded++
			continue
		}
		candidates = append(candidates, TraceExample{TraceID: row.TraceID, SpanID: row.SpanID, Service: row.Service, Name: row.Name, DurationNS: row.Duration, TimestampNS: row.Timestamp, Status: row.Status})
	}
	sort.Slice(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		x, _ := strconv.ParseUint(a.DurationNS, 10, 64)
		y, _ := strconv.ParseUint(b.DurationNS, 10, 64)
		if x != y {
			return x > y
		}
		left, right := []string{a.TraceID, a.SpanID, a.TimestampNS, a.Service, a.Name, a.Status}, []string{b.TraceID, b.SpanID, b.TimestampNS, b.Service, b.Name, b.Status}
		for k := range left {
			if left[k] != right[k] {
				return left[k] < right[k]
			}
		}
		return false
	})
	seen := map[string]bool{}
	add := func(v TraceExample, reason string) {
		if seen[v.TraceID] {
			return
		}
		seen[v.TraceID] = true
		v.Reason = reason
		out.Items = append(out.Items, v)
	}
	for _, c := range candidates {
		if c.Status == "Error" {
			add(c, "Recorded Error span on this page")
			break
		}
	}
	if len(candidates) > 0 {
		add(candidates[0], "Longest linked span on this page")
		add(candidates[len(candidates)-1], "Shortest linked span on this page")
		out.State = "observed"
	} else {
		out.State = "no_linked_spans"
	}
	return out
}
