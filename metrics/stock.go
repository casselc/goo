package metrics

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
)

// StockPoint is a cumulative nonmonotonic sum, not a gauge or a counter rate.
// StartNS identifies the accumulation epoch; decreases are valid observations.
type StockPoint struct {
	Stream          Stream  `json:"stream"`
	Temporality     string  `json:"temporality"`
	Monotonic       bool    `json:"monotonic"`
	StartNS         string  `json:"startNS"`
	EndNS           string  `json:"endNS"`
	Value           float64 `json:"value"`
	NoRecordedValue bool    `json:"noRecordedValue,omitempty"`
}
type StockSample struct {
	StartNS string   `json:"startNS"`
	AtNS    string   `json:"atNS"`
	Value   *float64 `json:"value,omitempty"`
	State   string   `json:"state"`
	Reason  string   `json:"reason,omitempty"`
}
type StockSeries struct {
	Stream  Stream        `json:"stream"`
	State   string        `json:"state"`
	Samples []StockSample `json:"samples"`
}
type StockResult struct {
	FromNS string        `json:"fromNS"`
	ToNS   string        `json:"toNS"`
	Series []StockSeries `json:"series"`
}

// EvaluateStocks observes cumulative stocks independently at each timestamp.
// It does not integrate delta sums, interpolate gaps, sum instances, infer
// capacity, or calculate rates. Service identity remains deliberately required.
func EvaluateStocks(fromText, toText, maxAgeText string, points []StockPoint) (*StockResult, error) {
	return evaluateStocks(fromText, toText, maxAgeText, points, serviceObservationIdentity)
}

// EvaluateHostStocks is an explicit host-resource path; service identity remains
// mandatory for the original EvaluateStocks entry point.
func EvaluateHostStocks(fromText, toText, maxAgeText, hostID string, points []StockPoint) (*StockResult, error) {
	if err := validateHostObservationID(hostID); err != nil {
		return nil, err
	}
	return evaluateStocks(fromText, toText, maxAgeText, points, func(s Stream) bool { return s.Resource["host.id"] == hostID })
}
func evaluateStocks(fromText, toText, maxAgeText string, points []StockPoint, identity func(Stream) bool) (*StockResult, error) {
	from, err := nano(fromText)
	if err != nil {
		return nil, err
	}
	to, err := nano(toText)
	if err != nil || to <= from {
		return nil, fmt.Errorf("require nonempty stock window")
	}
	age, err := nano(maxAgeText)
	if err != nil || age <= 0 || len(points) > 10000 {
		return nil, fmt.Errorf("require positive freshness and at most 10000 points")
	}
	groups := map[string][]StockPoint{}
	streams := map[string]Stream{}
	for _, p := range points {
		start, se := nano(p.StartNS)
		at, ae := nano(p.EndNS)
		if se != nil || ae != nil || start > at || at < from || at >= to || p.Temporality != "cumulative" || p.Monotonic || math.IsNaN(p.Value) || math.IsInf(p.Value, 0) || !identity(p.Stream) || p.Stream.Scope == "" || p.Stream.Name == "" || p.Stream.Unit == "" {
			return nil, fmt.Errorf("invalid cumulative stock observation")
		}
		p.Stream = cloneStream(p.Stream)
		if p.Value == 0 {
			p.Value = 0 // Canonical positive zero makes replay bytes order-independent.
		}
		encoded, _ := json.Marshal(p.Stream)
		key := string(encoded)
		groups[key] = append(groups[key], p)
		streams[key] = cloneStream(p.Stream)
	}
	out := &StockResult{FromNS: fromText, ToNS: toText, Series: []StockSeries{}}
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		ps := groups[key]
		sort.Slice(ps, func(i, j int) bool { a, _ := nano(ps[i].EndNS); b, _ := nano(ps[j].EndNS); return a < b })
		s := StockSeries{Stream: streams[key], State: "unknown", Samples: []StockSample{}}
		for i := 0; i < len(ps); {
			p := ps[i]
			end := i + 1
			conflict := false
			for end < len(ps) && ps[end].EndNS == p.EndNS {
				q := ps[end]
				conflict = conflict || q.StartNS != p.StartNS || q.Value != p.Value || q.NoRecordedValue != p.NoRecordedValue
				end++
			}
			sample := StockSample{StartNS: p.StartNS, AtNS: p.EndNS, State: "observed"}
			at, _ := nano(p.EndNS)
			switch {
			case conflict:
				// No arbitrary winning epoch is exposed when simultaneous reports disagree.
				sample.StartNS = ""
				sample.State = "unknown"
				sample.Reason = "conflicting observations at one timestamp"
			case p.NoRecordedValue:
				sample.State = "unknown"
				sample.Reason = "no recorded value"
			default:
				v := p.Value
				sample.Value = &v
				if to-at > age {
					sample.State = "stale"
					sample.Reason = "older than declared freshness"
				}
			}
			s.Samples = append(s.Samples, sample)
			s.State = sample.State
			i = end
		}
		out.Series = append(out.Series, s)
	}
	return out, nil
}
