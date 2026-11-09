package verification

import (
	"fmt"
	"math/big"
	"sort"
	"strconv"

	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
)

// Raw source membership and rational interval arithmetic are independent of the
// production counter evaluator. This bounded oracle handles the cumulative,
// monotonic source bank in the actual 389 journey, not arbitrary OTel counters.
type counterRaw389 struct {
	AtNS, StartNS, Value, ServiceName, MetricName, MetricUnit, ScopeName, ScopeVersion, ScopeSchemaUrl, ResourceSchemaUrl string
	ResourceAttributes, Attributes                                                                                        map[string]string
	Flags, Temporality                                                                                                    int
	Monotonic                                                                                                             bool
}

func counterReference389(q measurements.Question, rows []counterRaw389) ([]metrics.CounterInterval, error) {
	selected := []counterRaw389{}
	for _, r := range rows {
		g := gaugeRaw386{AtNS: r.AtNS, Value: r.Value, ServiceName: r.ServiceName, MetricName: r.MetricName, MetricUnit: r.MetricUnit, ScopeName: r.ScopeName, ScopeVersion: r.ScopeVersion, ScopeSchemaUrl: r.ScopeSchemaUrl, ResourceSchemaUrl: r.ResourceSchemaUrl, ResourceAttributes: r.ResourceAttributes, Attributes: r.Attributes}
		if gaugeSelect386(q, g) {
			epoch, e := strconv.ParseUint(r.StartNS, 10, 64)
			if e != nil || strconv.FormatUint(epoch, 10) != r.StartNS {
				return nil, fmt.Errorf("noncanonical raw epoch")
			}
			end, e := strconv.ParseUint(r.AtNS, 10, 64)
			if e != nil || strconv.FormatUint(end, 10) != r.AtNS || epoch >= end {
				return nil, fmt.Errorf("invalid raw timestamp")
			}
			if q.Definition.CounterDomain == "" && r.Flags&1 == 0 {
				value, e := strconv.ParseUint(r.Value, 10, 64)
				if e != nil || strconv.FormatUint(value, 10) != r.Value {
					return nil, fmt.Errorf("invalid raw integer counter")
				}
			}

			if r.Temporality != 2 || !r.Monotonic {
				return nil, fmt.Errorf("oracle requires reported monotonic cumulative data")
			}
			selected = append(selected, r)
		}
	}
	sort.Slice(selected, func(i, j int) bool {
		a, _ := new(big.Int).SetString(selected[i].AtNS, 10)
		b, _ := new(big.Int).SetString(selected[j].AtNS, 10)
		return a.Cmp(b) < 0
	})
	result := []metrics.CounterInterval{}
	poisoned := false
	for i := 1; i < len(selected); i++ {
		a, b := selected[i-1], selected[i]
		start, e := strconv.ParseInt(a.AtNS, 10, 64)
		if e != nil {
			return nil, e
		}
		end, e := strconv.ParseInt(b.AtNS, 10, 64)
		if e != nil || end <= start {
			return nil, fmt.Errorf("invalid raw counter order")
		}
		av, ok := new(big.Rat).SetString(a.Value)
		if !ok {
			return nil, fmt.Errorf("invalid raw value")
		}
		bv, ok := new(big.Rat).SetString(b.Value)
		if !ok {
			return nil, fmt.Errorf("invalid raw value")
		}
		epochChanged := a.StartNS != b.StartNS
		flags := a.Flags&1 != 0 || b.Flags&1 != 0
		decrease := !flags && bv.Cmp(av) < 0
		reason := ""
		if epochChanged {
			reason = "counter-epoch-change"
			poisoned = false
		} else if decrease {
			reason = "counter-decrease"
			poisoned = true
		} else if poisoned {
			reason = "invalid-counter-epoch"
		} else if flags {
			reason = "no-recorded-value"
		} else if end-start > q.Definition.MaxGapMS*1000000 {
			reason = "excessive-gap"
		}
		interval := metrics.CounterInterval{FromNS: a.AtNS, ToNS: b.AtNS, State: "unknown", Reason: reason}
		if reason == "" {
			delta := new(big.Rat).Sub(bv, av)
			rate := new(big.Rat).Mul(delta, big.NewRat(1000000000, end-start))
			v, _ := rate.Float64()
			interval.State = "ok"
			interval.PerSecond = &v
			if q.Definition.CounterDomain == "nonnegative-decimal" {
				interval.Delta = delta.FloatString(2)
				interval.ExactPerSecond = rate.RatString()
			} else {
				interval.Delta = delta.Num().String()
			}
		}
		result = append(result, interval)
	}
	return result, nil
}
