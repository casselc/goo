package metrics

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"strconv"
)

// RatioPolicy declares an exhaustive denominator outcome set and its error subset.
// Missing members never imply zero. Counts come from one evaluated counter basis.
type RatioPolicy struct {
	Outcomes []string `json:"outcomes"`
	Errors   []string `json:"errors"`
}
type RatioInterval struct {
	FromNS   string   `json:"fromNS"`
	ToNS     string   `json:"toNS"`
	State    string   `json:"state"`
	Reason   string   `json:"reason,omitempty"`
	Calls    string   `json:"calls,omitempty"`
	Errors   string   `json:"errors,omitempty"`
	Fraction string   `json:"fraction,omitempty"`
	Percent  *float64 `json:"percent,omitempty"`
}
type RatioSeries struct {
	Stream    Stream          `json:"stream"`
	State     string          `json:"state"`
	Intervals []RatioInterval `json:"intervals"`
}
type RatioResult struct {
	APIVersion   string        `json:"apiVersion"`
	Basis        string        `json:"basis"`
	CounterBasis string        `json:"counterBasis"`
	Policy       RatioPolicy   `json:"policy"`
	FromNS       string        `json:"fromNS"`
	ToNS         string        `json:"toNS"`
	Series       []RatioSeries `json:"series"`
}

func ValidateRatioPolicy(policy RatioPolicy) error {
	all, errors := map[string]bool{}, map[string]bool{}
	if len(policy.Outcomes) == 0 || len(policy.Errors) == 0 {
		return fmt.Errorf("ratio requires outcomes and an error subset")
	}
	for _, o := range policy.Outcomes {
		if all[o] || (o != "ok" && o != "error" && o != "panic") {
			return fmt.Errorf("invalid ratio outcome")
		}
		all[o] = true
	}
	for _, o := range policy.Errors {
		if errors[o] || !all[o] || o == "ok" {
			return fmt.Errorf("errors must be distinct denominator outcomes")
		}
		errors[o] = true
	}
	return nil
}

// EvaluateErrorRatios partitions at every observed endpoint. Only complete,
// exactly aligned intervals are divided; it never prorates counts or fills gaps.
func EvaluateErrorRatios(input CounterResult, policy RatioPolicy) (*RatioResult, error) {
	if e := ValidateRatioPolicy(policy); e != nil {
		return nil, e
	}
	from, e := nano(input.FromNS)
	if e != nil {
		return nil, e
	}
	to, e := nano(input.ToNS)
	if e != nil || to <= from {
		return nil, fmt.Errorf("invalid ratio window")
	}
	policy = RatioPolicy{append([]string{}, policy.Outcomes...), append([]string{}, policy.Errors...)}
	sort.Strings(policy.Outcomes)
	sort.Strings(policy.Errors)
	selected, failures := map[string]bool{}, map[string]bool{}
	for _, o := range policy.Outcomes {
		selected[o] = true
	}
	for _, o := range policy.Errors {
		failures[o] = true
	}
	type group struct {
		stream      Stream
		members     map[string]CounterSeries
		edges       map[int64]bool
		temporality string
	}
	groups := map[string]*group{}
	points := 0
	for _, s := range input.Series {
		if s.Stream.Unit != "{invocation}" || s.Stream.Resource["service.name"] == "" || s.Stream.Resource["service.instance.id"] == "" || s.Stream.Scope == "" || s.Stream.Name == "" {
			return nil, fmt.Errorf("ratio requires a complete invocation stream identity")
		}
		o := s.Stream.Attributes["goo.operation.outcome"]
		if !selected[o] {
			return nil, fmt.Errorf("counter outcome outside ratio denominator")
		}
		if s.Temporality != "cumulative" && s.Temporality != "delta" {
			return nil, fmt.Errorf("unsupported ratio temporality")
		}
		identity := cloneStream(s.Stream)
		delete(identity.Attributes, "goo.operation.outcome")
		b, _ := json.Marshal(identity)
		key := string(b)
		g := groups[key]
		if g == nil {
			g = &group{identity, map[string]CounterSeries{}, map[int64]bool{from: true, to: true}, s.Temporality}
			groups[key] = g
		}
		if _, ok := g.members[o]; ok {
			return nil, fmt.Errorf("duplicate ratio member")
		}
		if g.temporality != s.Temporality {
			return nil, fmt.Errorf("mixed ratio temporalities")
		}
		previous := from
		for _, i := range s.Intervals {
			points++
			if points > 10000 {
				return nil, fmt.Errorf("ratio exceeds 10000 input intervals")
			}
			a, e1 := nano(i.FromNS)
			b, e2 := nano(i.ToNS)
			if e1 != nil || e2 != nil || a < previous || b <= a || b > to {
				return nil, fmt.Errorf("invalid or overlapping ratio interval")
			}
			previous = b
			g.edges[a] = true
			g.edges[b] = true
			if i.State == "ok" {
				v, e := strconv.ParseUint(i.Delta, 10, 64)
				if e != nil || strconv.FormatUint(v, 10) != i.Delta {
					return nil, fmt.Errorf("invalid exact ratio count")
				}
			} else if i.State != "unknown" || i.Delta != "" {
				return nil, fmt.Errorf("invalid ratio interval state")
			}
		}
		g.members[o] = s
	}
	result := &RatioResult{APIVersion: "goo/error-ratio-result/v1alpha1", CounterBasis: input.Basis, Policy: policy, FromNS: input.FromNS, ToNS: input.ToNS, Series: []RatioSeries{}}
	keys := []string{}
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		g := groups[key]
		edges := []int64{}
		for t := range g.edges {
			edges = append(edges, t)
		}
		sort.Slice(edges, func(i, j int) bool { return edges[i] < edges[j] })
		series := RatioSeries{Stream: g.stream, State: "ok", Intervals: []RatioInterval{}}
		hasTraffic := false
		lookup := map[string]map[string]CounterInterval{}
		for o, s := range g.members {
			lookup[o] = map[string]CounterInterval{}
			for _, i := range s.Intervals {
				lookup[o][i.FromNS+"/"+i.ToNS] = i
			}
		}
		for n := 1; n < len(edges); n++ {
			a, b := fmt.Sprint(edges[n-1]), fmt.Sprint(edges[n])
			interval := RatioInterval{FromNS: a, ToNS: b, State: "ok"}
			calls, errors := new(big.Int), new(big.Int)
			for _, o := range policy.Outcomes {
				_, exists := g.members[o]
				if !exists {
					interval.State = "unknown"
					interval.Reason = "missing-outcome"
					break
				}
				i, ok := lookup[o][a+"/"+b]
				if !ok {
					interval.State = "unknown"
					interval.Reason = "unaligned-or-missing-interval"
					break
				}
				if i.State != "ok" {
					interval.State = "unknown"
					interval.Reason = "unsupported-counter-interval"
					break
				}
				v, _ := new(big.Int).SetString(i.Delta, 10)
				calls.Add(calls, v)
				if failures[o] {
					errors.Add(errors, v)
				}
			}
			if interval.State == "ok" {
				interval.Calls = calls.String()
				interval.Errors = errors.String()
				if calls.Sign() == 0 {
					interval.State = "no_data"
					interval.Reason = "zero-traffic"
				} else {
					interval.Fraction = errors.String() + "/" + calls.String()
					v, _ := new(big.Rat).Mul(new(big.Rat).SetFrac(errors, calls), big.NewRat(100, 1)).Float64()
					interval.Percent = &v
					hasTraffic = true
				}
			}
			if interval.State == "unknown" {
				series.State = "unknown"
			} else if interval.State == "no_data" && series.State == "ok" {
				series.State = "no_data"
			}
			series.Intervals = append(series.Intervals, interval)
		}
		if series.State != "unknown" && hasTraffic {
			series.State = "ok"
		}
		result.Series = append(result.Series, series)
	}
	// Hash the full canonical input/result, not a caller-supplied basis alone.
	canonical := input
	canonical.Series = append([]CounterSeries{}, input.Series...)
	sort.Slice(canonical.Series, func(i, j int) bool {
		a, _ := json.Marshal(canonical.Series[i])
		b, _ := json.Marshal(canonical.Series[j])
		return string(a) < string(b)
	})
	payload, e := json.Marshal(struct {
		Input  CounterResult
		Result *RatioResult
	}{canonical, result})
	if e != nil {
		return nil, e
	}
	hash := sha256.Sum256(payload)
	result.Basis = fmt.Sprintf("sha256:%x", hash)
	return result, nil
}
