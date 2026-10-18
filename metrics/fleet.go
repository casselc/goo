package metrics

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"strings"
)

// FleetRequest fixes membership for the entire question window. A nil counter
// means the source was unavailable; an empty counter result means unobserved.
type FleetInput struct {
	Instance string         `json:"instance"`
	Counter  *CounterResult `json:"counter"`
}
type FleetRequest struct {
	Comparison *ComparisonPolicy `json:"comparison,omitempty"`
	FromNS     string            `json:"fromNS"`
	ToNS       string            `json:"toNS"`
	Members    []string          `json:"members"`
	Policy     RatioPolicy       `json:"policy"`
	Inputs     []FleetInput      `json:"inputs"`
}
type FleetMember struct {
	Instance string `json:"instance"`
	State    string `json:"state"`
	Basis    string `json:"basis,omitempty"`
}
type FleetInterval struct {
	RatioInterval
	Expected  int           `json:"expected"`
	Supported int           `json:"supported"`
	Members   []FleetMember `json:"members"`
}
type FleetSeries struct {
	Stream    Stream          `json:"stream"`
	State     string          `json:"state"`
	Intervals []FleetInterval `json:"intervals"`
}
type FleetResult struct {
	Comparison *FleetComparison `json:"comparison,omitempty"`
	APIVersion string           `json:"apiVersion"`
	Basis      string           `json:"basis"`
	FromNS     string           `json:"fromNS"`
	ToNS       string           `json:"toNS"`
	Policy     RatioPolicy      `json:"policy"`
	Members    []FleetMember    `json:"members"`
	Series     []FleetSeries    `json:"series"`
}

func ValidateFleetMembers(members []string) error {
	if len(members) < 2 || len(members) > 4 {
		return fmt.Errorf("fleet requires 2..4 explicit members")
	}
	seen := map[string]bool{}
	for _, m := range members {
		if m == "" || strings.TrimSpace(m) != m || len(m) > 128 || seen[m] {
			return fmt.Errorf("invalid or duplicate fleet member")
		}
		seen[m] = true
	}
	return nil
}

// EvaluateFleet sums member populations only when every declared member has a
// valid, exactly aligned outcome distribution. It never averages percentages.
func EvaluateFleet(request FleetRequest) (*FleetResult, error) {
	if e := ValidateFleetMembers(request.Members); e != nil {
		return nil, e
	}
	if e := ValidateRatioPolicy(request.Policy); e != nil {
		return nil, e
	}
	from, e := nano(request.FromNS)
	if e != nil {
		return nil, e
	}
	to, e := nano(request.ToNS)
	if e != nil || to <= from {
		return nil, fmt.Errorf("invalid fleet window")
	}
	members := append([]string{}, request.Members...)
	sort.Strings(members)
	allowed := map[string]bool{}
	for _, m := range members {
		allowed[m] = true
	}
	inputs := map[string]*CounterResult{}
	for _, input := range request.Inputs {
		if !allowed[input.Instance] {
			return nil, fmt.Errorf("undeclared fleet input")
		}
		if _, ok := inputs[input.Instance]; ok {
			return nil, fmt.Errorf("duplicate fleet input")
		}
		inputs[input.Instance] = input.Counter
	}
	policy := RatioPolicy{Outcomes: append([]string{}, request.Policy.Outcomes...), Errors: append([]string{}, request.Policy.Errors...)}
	sort.Strings(policy.Outcomes)
	sort.Strings(policy.Errors)
	result := &FleetResult{APIVersion: "goo/fleet-result/v1alpha1", FromNS: request.FromNS, ToNS: request.ToNS, Policy: policy, Members: []FleetMember{}, Series: []FleetSeries{}}
	type group struct {
		stream Stream
		series map[string]RatioSeries
		edges  map[int64]bool
	}
	groups := map[string]*group{}
	ratios := map[string]*RatioResult{}
	availability := map[string]string{}
	points := 0
	for _, member := range members {
		input := inputs[member]
		status := FleetMember{Instance: member, State: "unavailable"}
		if input != nil {
			if input.FromNS != request.FromNS || input.ToNS != request.ToNS {
				return nil, fmt.Errorf("fleet member window mismatch")
			}
			for _, series := range input.Series {
				if series.Stream.Resource["service.instance.id"] != member {
					return nil, fmt.Errorf("fleet input instance mismatch")
				}
				points += len(series.Intervals)
				if points > 10000 {
					return nil, fmt.Errorf("fleet exceeds 10000 input intervals")
				}
			}
			ratio, e := EvaluateErrorRatios(*input, policy)
			if e != nil {
				return nil, e
			}
			ratios[member] = ratio
			status.Basis = ratio.Basis
			status.State = "unobserved"
			if len(ratio.Series) > 0 {
				status.State = "observed"
			}
			for _, series := range ratio.Series {
				identity := cloneStream(series.Stream)
				delete(identity.Resource, "service.instance.id")
				keyJSON, _ := json.Marshal(identity)
				key := string(keyJSON)
				g := groups[key]
				if g == nil {
					g = &group{identity, map[string]RatioSeries{}, map[int64]bool{from: true, to: true}}
					groups[key] = g
				}
				g.series[member] = series
				for _, i := range series.Intervals {
					a, _ := nano(i.FromNS)
					b, _ := nano(i.ToNS)
					g.edges[a] = true
					g.edges[b] = true
				}
			}
		}
		availability[member] = status.State
		result.Members = append(result.Members, status)
	}
	keys := []string{}
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		g := groups[key]
		edges := []int64{}
		for edge := range g.edges {
			edges = append(edges, edge)
		}
		sort.Slice(edges, func(i, j int) bool { return edges[i] < edges[j] })
		lookup := map[string]map[string]RatioInterval{}
		for member, s := range g.series {
			lookup[member] = map[string]RatioInterval{}
			for _, i := range s.Intervals {
				lookup[member][i.FromNS+"/"+i.ToNS] = i
			}
		}
		series := FleetSeries{Stream: g.stream, State: "no_data", Intervals: []FleetInterval{}}
		hasTraffic := false
		for n := 1; n < len(edges); n++ {
			a, b := fmt.Sprint(edges[n-1]), fmt.Sprint(edges[n])
			interval := FleetInterval{RatioInterval: RatioInterval{FromNS: a, ToNS: b, State: "unknown", Reason: "incomplete-member-coverage"}, Expected: len(members), Members: []FleetMember{}}
			calls, errors := new(big.Int), new(big.Int)
			for _, member := range members {
				status := FleetMember{Instance: member, State: availability[member]}
				if status.State == "observed" {
					if _, ok := g.series[member]; !ok {
						status.State = "missing-compatible-stream"
					} else if i, ok := lookup[member][a+"/"+b]; !ok {
						status.State = "unaligned-interval"
					} else if i.State != "ok" && i.State != "no_data" {
						status.State = "unknown-counter-interval"
					} else {
						status.State = i.State
						interval.Supported++
						c, _ := new(big.Int).SetString(i.Calls, 10)
						e, _ := new(big.Int).SetString(i.Errors, 10)
						calls.Add(calls, c)
						errors.Add(errors, e)
					}
				}
				interval.Members = append(interval.Members, status)
			}
			if interval.Supported == interval.Expected {
				interval.Calls = calls.String()
				interval.Errors = errors.String()
				interval.Reason = ""
				interval.State = "ok"
				if calls.Sign() == 0 {
					interval.State = "no_data"
					interval.Reason = "zero-traffic"
				} else {
					interval.Fraction = errors.String() + "/" + calls.String()
					v, _ := new(big.Rat).Mul(new(big.Rat).SetFrac(errors, calls), big.NewRat(100, 1)).Float64()
					interval.Percent = &v
					hasTraffic = true
				}
			} else {
				series.State = "unknown"
			}
			series.Intervals = append(series.Intervals, interval)
		}
		if series.State != "unknown" && hasTraffic {
			series.State = "ok"
		}
		result.Series = append(result.Series, series)
	}
	if request.Comparison != nil {
		comparison, e := compareFleet(request, ratios)
		if e != nil {
			return nil, e
		}
		result.Comparison = comparison
	}
	payload, e := json.Marshal(result)
	if e != nil {
		return nil, e
	}
	hash := sha256.Sum256(payload)
	result.Basis = fmt.Sprintf("sha256:%x", hash)
	return result, nil
}
