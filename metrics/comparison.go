package metrics

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
)

// ComparisonPolicy permits comparison of actual windows, never shifting counts
// or asserting a common population. Exact fleet totals use a separate path.
type ComparisonPolicy struct {
	Mode      string `json:"mode"`
	MaxSkewNS string `json:"maxSkewNS"`
}
type ComparisonMember struct {
	Instance string `json:"instance"`
	RatioInterval
	Causes []string `json:"causes,omitempty"`
}
type ComparisonInterval struct {
	FromNS        string             `json:"fromNS"`
	ToNS          string             `json:"toNS"`
	State         string             `json:"state"`
	StartSpreadNS string             `json:"startSpreadNS,omitempty"`
	EndSpreadNS   string             `json:"endSpreadNS,omitempty"`
	Members       []ComparisonMember `json:"members"`
}
type ComparisonSeries struct {
	Stream            Stream               `json:"stream"`
	ReferenceInstance string               `json:"referenceInstance"`
	Intervals         []ComparisonInterval `json:"intervals"`
}
type FleetComparison struct {
	Policy ComparisonPolicy   `json:"policy"`
	Basis  string             `json:"basis"`
	Series []ComparisonSeries `json:"series"`
}

func ValidateComparisonPolicy(p ComparisonPolicy) error {
	v, e := nano(p.MaxSkewNS)
	if p.Mode != "bounded-skew" || e != nil || v <= 0 || v > 1000000000 {
		return fmt.Errorf("comparison requires bounded-skew and maxSkewNS in (0,1s]")
	}
	return nil
}
func fleetIdentity(s Stream) (Stream, string) {
	s = cloneStream(s)
	delete(s.Resource, "service.instance.id")
	delete(s.Attributes, "goo.operation.outcome")
	b, _ := json.Marshal(s)
	return s, string(b)
}
func distance(a, b int64) int64 {
	if a > b {
		return a - b
	}
	return b - a
}
func compareFleet(request FleetRequest, ratios map[string]*RatioResult) (*FleetComparison, error) {
	policy := *request.Comparison
	if e := ValidateComparisonPolicy(policy); e != nil {
		return nil, e
	}
	skew, _ := nano(policy.MaxSkewNS)
	members := append([]string{}, request.Members...)
	sort.Strings(members)
	type group struct {
		stream  Stream
		members map[string]RatioSeries
	}
	groups := map[string]*group{}
	bases := map[string]string{}
	raw := map[string]*CounterResult{}
	for _, input := range request.Inputs {
		raw[input.Instance] = input.Counter
	}
	for _, member := range members {
		if r := ratios[member]; r != nil {
			bases[member] = r.Basis
			for _, s := range r.Series {
				identity, key := fleetIdentity(s.Stream)
				g := groups[key]
				if g == nil {
					g = &group{identity, map[string]RatioSeries{}}
					groups[key] = g
				}
				g.members[member] = s
			}
		}
	}
	result := &FleetComparison{Policy: policy, Series: []ComparisonSeries{}}
	keys := []string{}
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		g := groups[key]
		reference := ""
		for _, member := range members {
			if _, ok := g.members[member]; ok {
				reference = member
				break
			}
		}
		series := ComparisonSeries{Stream: g.stream, ReferenceInstance: reference, Intervals: []ComparisonInterval{}}
		used := map[string]map[int]bool{}
		for _, member := range members {
			used[member] = map[int]bool{}
		}
		for refIndex, ref := range g.members[reference].Intervals {
			a, _ := nano(ref.FromNS)
			b, _ := nano(ref.ToNS)
			minA, maxA, minB, maxB := a, a, b, b
			row := ComparisonInterval{FromNS: ref.FromNS, ToNS: ref.ToNS, State: "unknown", Members: []ComparisonMember{}}
			chosen := map[string]int{reference: refIndex}
			supported := ref.State == "ok" || ref.State == "no_data"
			// Greedy nearest whole intervals, with deterministic earlier-time tie breaking.
			// Reserve candidates only after the whole row is supported; never reuse them.
			if supported {
				for _, member := range members {
					if member == reference {
						continue
					}
					best := -1
					var bestDistance int64
					for index, candidate := range g.members[member].Intervals {
						if used[member][index] || (candidate.State != "ok" && candidate.State != "no_data") {
							continue
						}
						ca, _ := nano(candidate.FromNS)
						cb, _ := nano(candidate.ToNS)
						loA, hiA, loB, hiB := minA, maxA, minB, maxB
						if ca < loA {
							loA = ca
						}
						if ca > hiA {
							hiA = ca
						}
						if cb < loB {
							loB = cb
						}
						if cb > hiB {
							hiB = cb
						}
						if hiA-loA > skew || hiB-loB > skew {
							continue
						}
						score := distance(ca, a) + distance(cb, b)
						if best < 0 || score < bestDistance {
							best = index
							bestDistance = score
						}
					}
					if best < 0 {
						supported = false
						continue
					}
					chosen[member] = best
					i := g.members[member].Intervals[best]
					ca, _ := nano(i.FromNS)
					cb, _ := nano(i.ToNS)
					if ca < minA {
						minA = ca
					}
					if ca > maxA {
						maxA = ca
					}
					if cb < minB {
						minB = cb
					}
					if cb > maxB {
						maxB = cb
					}
				}
			}
			for _, member := range members {
				reason := "outside-skew-tolerance"
				if ref.State != "ok" && ref.State != "no_data" {
					reason = "unsupported-reference-window"
				}
				m := ComparisonMember{Instance: member, RatioInterval: RatioInterval{State: "unknown", Reason: reason}}
				if i, ok := chosen[member]; ok {
					m.RatioInterval = g.members[member].Intervals[i]
				} else if raw[member] == nil {
					m.Reason = "source-unavailable"
				} else if len(ratios[member].Series) == 0 {
					m.Reason = "unobserved"
				} else if _, ok := g.members[member]; !ok {
					m.Reason = "missing-compatible-stream"
				}
				if m.State == "unknown" {
					causes := map[string]bool{}
					if input := raw[member]; input != nil {
						for _, s := range input.Series {
							_, identity := fleetIdentity(s.Stream)
							if identity != key {
								continue
							}
							for _, i := range s.Intervals {
								ca, _ := nano(i.FromNS)
								cb, _ := nano(i.ToNS)
								if i.State == "unknown" && ca < b && cb > a {
									causes[i.Reason] = true
								}
							}
						}
					}
					for c := range causes {
						m.Causes = append(m.Causes, c)
					}
					sort.Strings(m.Causes)
				}
				row.Members = append(row.Members, m)
			}
			if supported {
				row.State = "comparable"
				row.StartSpreadNS = fmt.Sprint(maxA - minA)
				row.EndSpreadNS = fmt.Sprint(maxB - minB)
				for member, index := range chosen {
					used[member][index] = true
				}
			}
			series.Intervals = append(series.Intervals, row)
		}
		result.Series = append(result.Series, series)
	}
	payload, e := json.Marshal(struct {
		Bases  map[string]string
		Result *FleetComparison
	}{bases, result})
	if e != nil {
		return nil, e
	}
	h := sha256.Sum256(payload)
	result.Basis = fmt.Sprintf("sha256:%x", h)
	return result, nil
}
