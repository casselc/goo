package measurements

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/metrics"
	"maps"
	"math/big"
)

// These private representations preserve the concrete public wire contracts.
// Domain policy stays with CPU or memory; evidence authority is shared.
type containerResourceSignal ContainerMemorySignal
type containerResourcePercentageValue ContainerMemoryPercentage
type containerResourceEntry struct {
	role   string
	result *Result
}
type containerResourceFact struct {
	result *Result
	sample *metrics.GaugeSample
	stream metrics.Stream
	value  *big.Rat
}
type containerResourcePolicy struct {
	label, invalidReason string
	role                 func(Instrument) string
	title                func(string) string
	parse                func(string) (*big.Rat, bool)
	format               func(*big.Rat) string
	percentage           func(*big.Rat) containerResourcePercentageValue
}

func cloneContainerResourceEvidence[T any](input T, parent PodMetricDiscoveryQuestion, results []*Result, label string) (T, error) {
	var zero T
	if e := parent.Validate(); e != nil {
		return zero, e
	}
	if !sameContainerContext(parent.Resource, parent.Resource) {
		return zero, fmt.Errorf("%s parent requires actual container ID, pod UID and container name", label)
	}
	// Bound raw evidence and plan payloads before cloning, and replay before
	// marshaling derived result families supplied by callers.
	budget := 0
	for _, r := range results {
		if r == nil {
			continue
		}
		if len(r.Rows) > RowLimit || len(r.Plan.Parameters) > 128 {
			return zero, fmt.Errorf("%s input exceeds bounded evidence limits", label)
		}
		budget += len(r.Plan.SQL) + len(r.ReadAt) + len(r.Basis)
		for k, v := range r.Plan.Parameters {
			budget += len(k) + len(v)
			if budget > 8<<20 {
				return zero, fmt.Errorf("%s input exceeds 8 MiB", label)
			}
		}
		for _, raw := range r.Rows {
			budget += len(raw)
			if budget > 8<<20 {
				return zero, fmt.Errorf("%s input exceeds 8 MiB", label)
			}
		}
		if e := ValidateResult(r); e != nil {
			return zero, e
		}
	}
	encoded, e := json.Marshal(input)
	if e != nil {
		return zero, e
	}
	if len(encoded) > 8<<20 {
		return zero, fmt.Errorf("%s input exceeds 8 MiB", label)
	}
	var detached T
	if e = json.Unmarshal(encoded, &detached); e != nil {
		return zero, e
	}
	return detached, nil
}
func composeContainerResourceSignals(parent PodMetricDiscoveryQuestion, entries []containerResourceEntry, policy containerResourcePolicy) ([]containerResourceSignal, map[string]containerResourceFact, error) {
	out := []containerResourceSignal{}
	facts := map[string]containerResourceFact{}
	for _, entry := range entries {
		signal := containerResourceSignal{Role: entry.role, Title: policy.title(entry.role), State: "unavailable", Reason: "not-supplied"}
		r := entry.result
		if r == nil {
			out = append(out, signal)
			continue
		}
		d := r.Question.Definition
		s := d.Source.ExactStream
		if d.Kind != "gauge" || d.Source.Resource == nil || d.Source.Resource.Kind != "kubernetes-container" || s == nil || policy.role(Instrument{Kind: d.Kind, Metric: d.Source.Metric, Unit: d.Source.Unit, Scope: d.Source.Scope, Version: d.Source.Version, SchemaURL: d.Source.SchemaURL}) != entry.role || r.Gauges == nil {
			return nil, nil, fmt.Errorf("%s signal requires its exact pinned gauge role and container authority", policy.label)
		}
		if d.Environment != parent.Environment || r.Question.FromMS != parent.FromMS || r.Question.ToMS != parent.ToMS {
			return nil, nil, fmt.Errorf("%s signals require the original parent window and environment", policy.label)
		}
		from, to := r.Question.BoundsNS()
		if from != ns(parent.FromMS) || to != ns(parent.ToMS) {
			return nil, nil, fmt.Errorf("%s signal nanosecond bounds differ from parent window", policy.label)
		}
		if !sameContainerContext(parent.Resource, s.Resource) {
			return nil, nil, fmt.Errorf("%s signal does not match parent identity and qualification", policy.label)
		}
		if len(r.Gauges.Series) > 1 {
			return nil, nil, fmt.Errorf("%s signal has ambiguous series", policy.label)
		}
		signal.Question = &r.Question
		signal.ReadAt = r.ReadAt
		signal.Basis = r.Basis
		signal.State = "no_data"
		signal.Reason = "no-observations"
		fact := containerResourceFact{result: r, stream: *s}
		if len(r.Gauges.Series) == 1 {
			series := r.Gauges.Series[0]
			if !sameExactStream(series.Stream, *s) {
				return nil, nil, fmt.Errorf("%s includes an unexpected stream", policy.label)
			}
			if len(series.Samples) > 0 {
				sample := series.Samples[len(series.Samples)-1]
				fact.sample = &sample
				signal.State = sample.State
				signal.ObservationState = sample.State
				signal.Reason = sample.Reason
				signal.AtNS = sample.AtNS
				if sample.Value != nil {
					// Every raw record at the selected time must certify the same reported decimal;
					// duplicate conflicts and missing flags remain governed by replay above.
					var value *big.Rat
					valid := true
					found := false
					for _, raw := range r.Rows {
						var row metricRow
						if json.Unmarshal(raw, &row) != nil {
							return nil, nil, fmt.Errorf("invalid %s raw row", policy.label)
						}
						if row.EndNS != sample.AtNS {
							continue
						}
						found = true
						if row.Flags&1 != 0 {
							valid = false
							break
						}
						n, ok := policy.parse(row.Value)
						if !ok || value != nil && value.Cmp(n) != 0 {
							valid = false
							break
						}
						value = n
					}
					if valid && found && value != nil {
						fact.value = value
						signal.Value = policy.format(value)
					} else {
						signal.State = "invalid"
						signal.Reason = policy.invalidReason
					}
				}
			}
		}
		facts[entry.role] = fact
		out = append(out, signal)
	}
	return out, facts, nil
}
func containerResourcePercentage(facts map[string]containerResourceFact, numeratorRole, role string, policy containerResourcePolicy) containerResourcePercentageValue {
	refuse := func(reason string) containerResourcePercentageValue {
		return containerResourcePercentageValue{State: "unavailable", Reason: reason}
	}
	numerator, nok := facts[numeratorRole]
	denominator, dok := facts[role]
	if !nok || !dok {
		return refuse("input-not-supplied")
	}
	if numerator.sample == nil || denominator.sample == nil {
		return refuse("no-observations")
	}
	if numerator.sample.State != "observed" || denominator.sample.State != "observed" {
		return refuse("observation-not-current")
	}
	if numerator.value == nil || denominator.value == nil {
		return refuse(policy.invalidReason)
	}
	if !maps.Equal(numerator.stream.Resource, denominator.stream.Resource) || numerator.stream.ResourceSchemaURL != denominator.stream.ResourceSchemaURL || !maps.Equal(numerator.stream.Attributes, denominator.stream.Attributes) {
		return refuse("resource-or-attribute-mismatch")
	}
	if numerator.sample.AtNS != denominator.sample.AtNS {
		return refuse("observation-time-mismatch")
	}
	if numerator.result.Question.Definition.MaxGapMS != denominator.result.Question.Definition.MaxGapMS {
		return refuse("freshness-policy-mismatch")
	}
	if denominator.value.Sign() == 0 {
		return refuse("zero-denominator")
	}
	ratio := new(big.Rat).Mul(new(big.Rat).Quo(numerator.value, denominator.value), big.NewRat(100, 1))
	return policy.percentage(ratio)
}
