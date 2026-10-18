// Package metrics implements bounded native metric semantics independently of storage.
package metrics

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"strconv"
)

type Stream struct {
	ResourceSchemaURL string            `json:"resourceSchemaURL,omitempty"`
	Resource          map[string]string `json:"resource"`
	Scope             string            `json:"scope"`
	Version           string            `json:"version"`
	SchemaURL         string            `json:"schemaURL"`
	Name              string            `json:"name"`
	Unit              string            `json:"unit"`
	Attributes        map[string]string `json:"attributes"`
}
type CounterPoint struct {
	Stream          Stream `json:"stream"`
	Temporality     string `json:"temporality"`
	Monotonic       bool   `json:"monotonic"`
	StartNS         string `json:"startNS"`
	EndNS           string `json:"endNS"`
	Value           string `json:"value"`
	NoRecordedValue bool   `json:"noRecordedValue,omitempty"`
}
type CounterRequest struct {
	FromNS   string         `json:"fromNS"`
	ToNS     string         `json:"toNS"`
	MaxGapNS string         `json:"maxGapNS"`
	Points   []CounterPoint `json:"points"`
}
type CounterInterval struct {
	FromNS    string   `json:"fromNS"`
	ToNS      string   `json:"toNS"`
	State     string   `json:"state"`
	Reason    string   `json:"reason,omitempty"`
	Delta     string   `json:"delta,omitempty"`
	PerSecond *float64 `json:"perSecond,omitempty"`
}
type CounterSeries struct {
	Stream      Stream            `json:"stream"`
	Temporality string            `json:"temporality"`
	State       string            `json:"state"`
	Intervals   []CounterInterval `json:"intervals"`
}
type CounterResult struct {
	APIVersion string          `json:"apiVersion"`
	Basis      string          `json:"basis"`
	FromNS     string          `json:"fromNS"`
	ToNS       string          `json:"toNS"`
	Series     []CounterSeries `json:"series"`
	Warnings   []string        `json:"warnings"`
}
type parsedPoint struct {
	raw        CounterPoint
	start, end int64
	value      uint64
}

func nano(s string) (int64, error) {
	v, e := strconv.ParseInt(s, 10, 64)
	if e != nil || v < 0 || strconv.FormatInt(v, 10) != s {
		return 0, fmt.Errorf("require canonical nonnegative int64 nanoseconds")
	}
	return v, nil
}
func cloneStream(s Stream) Stream {
	r := map[string]string{}
	for k, v := range s.Resource {
		r[k] = v
	}
	a := map[string]string{}
	for k, v := range s.Attributes {
		a[k] = v
	}
	s.Resource = r
	s.Attributes = a
	return s
}
func rate(delta uint64, ns int64) float64 {
	n := new(big.Int).SetUint64(delta)
	n.Mul(n, big.NewInt(1000000000))
	r := new(big.Rat).SetFrac(n, big.NewInt(ns))
	f, _ := r.Float64()
	return f
}

// EvaluateCounters sorts late/out-of-order observations into a new result. It
// never interpolates boundaries, bridges gaps or subtracts across counter epochs.
func EvaluateCounters(request CounterRequest) (*CounterResult, error) {
	from, e := nano(request.FromNS)
	if e != nil {
		return nil, e
	}
	to, e := nano(request.ToNS)
	if e != nil || to <= from {
		return nil, fmt.Errorf("require a nonempty counter window")
	}
	gap, e := nano(request.MaxGapNS)
	if e != nil || gap == 0 {
		return nil, fmt.Errorf("require positive maxGapNS")
	}
	if len(request.Points) > 10000 {
		return nil, fmt.Errorf("counter point limit exceeded; no partial evaluation")
	}
	groups := map[string][]parsedPoint{}
	for _, p := range request.Points {
		if p.Stream.Resource["service.name"] == "" || p.Stream.Resource["service.instance.id"] == "" || p.Stream.Scope == "" || p.Stream.Name == "" || p.Stream.Unit == "" {
			return nil, fmt.Errorf("require explicit resource instance and instrumentation/name/unit identity")
		}
		if !p.Monotonic || (p.Temporality != "cumulative" && p.Temporality != "delta") {
			return nil, fmt.Errorf("only monotonic cumulative/delta counters supported")
		}
		start, e := nano(p.StartNS)
		if e != nil {
			return nil, e
		}
		end, e := nano(p.EndNS)
		if e != nil || end <= start {
			return nil, fmt.Errorf("require startNS < endNS")
		}
		value := uint64(0)
		if !p.NoRecordedValue {
			var e error
			value, e = strconv.ParseUint(p.Value, 10, 64)
			if e != nil || strconv.FormatUint(value, 10) != p.Value {
				return nil, fmt.Errorf("require canonical uint64 counter value")
			}
		}
		p.Stream = cloneStream(p.Stream)
		keyBytes, _ := json.Marshal(p.Stream)
		key := string(keyBytes)
		groups[key] = append(groups[key], parsedPoint{p, start, end, value})
	}
	result := &CounterResult{APIVersion: "goo/counter-result/v1alpha1", FromNS: request.FromNS, ToNS: request.ToNS, Series: []CounterSeries{}, Warnings: []string{"Rates describe observed intervals, not sampling or ingestion completeness.", "No boundary interpolation, fleet aggregation, error ratio or durable custody is inferred."}}
	keys := []string{}
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	canonical := []CounterPoint{}
	for _, key := range keys {
		points := groups[key]
		sort.Slice(points, func(i, j int) bool { return points[i].end < points[j].end })
		unique := []parsedPoint{}
		for _, p := range points {
			if len(unique) > 0 {
				prior := unique[len(unique)-1]
				if p.start < prior.start && p.raw.Temporality == "cumulative" {
					return nil, fmt.Errorf("counter epoch start regressed")
				}
				if prior.raw.Temporality != p.raw.Temporality {
					return nil, fmt.Errorf("mixed temporality in one stream")
				}
				if prior.end == p.end {
					a, _ := json.Marshal(prior.raw)
					b, _ := json.Marshal(p.raw)
					if string(a) != string(b) {
						return nil, fmt.Errorf("conflicting duplicate counter timestamp")
					}
					continue
				}
			}
			unique = append(unique, p)
			canonical = append(canonical, p.raw)
		}
		series := CounterSeries{Stream: unique[0].raw.Stream, Temporality: unique[0].raw.Temporality, State: "no_data", Intervals: []CounterInterval{}}
		tainted := false
		appendInterval := func(start, end int64, delta uint64, reason string) {
			if end <= from || start >= to {
				return
			}
			clippedStart, clippedEnd := start, end
			if clippedStart < from {
				clippedStart = from
			}
			if clippedEnd > to {
				clippedEnd = to
			}
			if start < from || end > to {
				reason = "boundary-crossing"
			}
			interval := CounterInterval{FromNS: strconv.FormatInt(clippedStart, 10), ToNS: strconv.FormatInt(clippedEnd, 10), State: "unknown", Reason: reason}
			if reason == "" {
				interval.State = "ok"
				interval.Delta = strconv.FormatUint(delta, 10)
				v := rate(delta, end-start)
				interval.PerSecond = &v
			}
			series.Intervals = append(series.Intervals, interval)
		}
		for i, p := range unique {
			if p.raw.Temporality == "delta" {
				if i > 0 && p.start < unique[i-1].end {
					return nil, fmt.Errorf("overlapping delta intervals")
				}
				reason := ""
				if p.end-p.start > gap {
					reason = "excessive-gap"
				}
				if p.raw.NoRecordedValue {
					reason = "no-recorded-value"
				}
				appendInterval(p.start, p.end, p.value, reason)
				continue
			}
			if i == 0 {
				continue
			}
			prior := unique[i-1]
			reason := ""
			delta := uint64(0)
			switch {
			case p.start != prior.start:
				reason = "counter-epoch-change"
				tainted = false
			case p.value < prior.value && !p.raw.NoRecordedValue && !prior.raw.NoRecordedValue:
				reason = "counter-decrease"
				tainted = true
			case tainted:
				reason = "invalid-counter-epoch"
			case p.raw.NoRecordedValue || prior.raw.NoRecordedValue:
				reason = "no-recorded-value"
			case p.end-prior.end > gap:
				reason = "excessive-gap"
			default:
				delta = p.value - prior.value
			}
			appendInterval(prior.end, p.end, delta, reason)
		}
		if len(series.Intervals) > 0 {
			series.State = "ok"
			cursor := from
			for _, interval := range series.Intervals {
				a, _ := nano(interval.FromNS)
				b, _ := nano(interval.ToNS)
				if a != cursor || interval.State != "ok" {
					series.State = "unknown"
				}
				cursor = b
			}
			if cursor != to {
				series.State = "unknown"
			}
		}
		result.Series = append(result.Series, series)
	}
	payload, _ := json.Marshal(CounterRequest{FromNS: request.FromNS, ToNS: request.ToNS, MaxGapNS: request.MaxGapNS, Points: canonical})
	sum := sha256.Sum256(payload)
	result.Basis = hex.EncodeToString(sum[:])
	return result, nil
}
