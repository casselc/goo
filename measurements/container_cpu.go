package measurements

import (
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/casselc/goo/metrics"
)

const ContainerCPUVersion = "goo/container-cpu-view/v1"
const ContainerCPUKubeletProfile = "container-cpu-kubeletstats-v0.137.0"
const ContainerCPUClusterProfile = "container-cpu-k8scluster-v0.137.0"

func ContainerCPURole(i Instrument) string {
	if i.Kind != "gauge" || i.Temporality != 0 || i.Monotonic || i.Unit != "{cpu}" || i.Version != "0.137.0" || i.SchemaURL != "" {
		return ""
	}
	if i.Scope == KubeletStatsScope && i.Metric == "container.cpu.usage" {
		return "usage"
	}
	if i.Scope == K8sClusterScope {
		switch i.Metric {
		case "k8s.container.cpu_request":
			return "request"
		case "k8s.container.cpu_limit":
			return "limit"
		}
	}
	return ""
}
func containerCPUProfile(role string) string {
	if role == "usage" {
		return ContainerCPUKubeletProfile
	}
	return ContainerCPUClusterProfile
}
func containerCPUTitle(role string) string {
	switch role {
	case "usage":
		return "Reported container CPU usage"
	case "request":
		return "Reported CPU request"
	default:
		return "Reported CPU limit"
	}
}

type ContainerCPUViewRequest struct {
	Parent  PodMetricDiscoveryQuestion `json:"parent"`
	Usage   *Result                    `json:"usage,omitempty"`
	Request *Result                    `json:"request,omitempty"`
	Limit   *Result                    `json:"limit,omitempty"`
}
type ContainerCPUSignal struct {
	Role             string    `json:"role"`
	Title            string    `json:"title"`
	Question         *Question `json:"question,omitempty"`
	Value            string    `json:"value,omitempty"`
	State            string    `json:"state"`
	ObservationState string    `json:"observationState,omitempty"`
	AtNS             string    `json:"atNS,omitempty"`
	Reason           string    `json:"reason,omitempty"`
	ReadAt           string    `json:"readAt,omitempty"`
	Basis            string    `json:"basis,omitempty"`
}
type ContainerCPUPercentage struct {
	State  string `json:"state"`
	Reason string `json:"reason,omitempty"`
	Value  string `json:"value,omitempty"`
	Text   string `json:"text,omitempty"`
}
type ContainerCPUPercentages struct {
	Request ContainerCPUPercentage `json:"request"`
	Limit   ContainerCPUPercentage `json:"limit"`
}
type ContainerCPUView struct {
	Version     string                     `json:"version"`
	Parent      PodMetricDiscoveryQuestion `json:"parent"`
	Signals     []ContainerCPUSignal       `json:"signals"`
	Percentages ContainerCPUPercentages    `json:"percentages"`
	Warnings    []string                   `json:"warnings"`
}

// exactContainerCPU certifies the bounded reported decimal, not pre-export CPU
// precision. No derivative, reset interpretation or integer rounding is applied.
func exactContainerCPU(raw string) (*big.Rat, bool) {
	if len(raw) > 128 || !containerByteDecimal.MatchString(raw) {
		return nil, false
	}
	if at := strings.IndexAny(raw, "eE"); at >= 0 {
		exponent, ok := new(big.Int).SetString(raw[at+1:], 10)
		if !ok || !exponent.IsInt64() || exponent.Int64() < -400 || exponent.Int64() > 400 {
			return nil, false
		}
	}
	value, ok := new(big.Rat).SetString(raw)
	if !ok || value.Sign() < 0 {
		return nil, false
	}
	finite, err := strconv.ParseFloat(raw, 64)
	if err != nil || math.IsInf(finite, 0) || math.IsNaN(finite) || finite == 0 && value.Sign() != 0 {
		return nil, false
	}
	if len(containerCPUDecimal(value)) > 512 {
		return nil, false
	}
	return value, true
}
func containerCPUDecimal(value *big.Rat) string {
	denominator := new(big.Int).Set(value.Denom())
	two, five := 0, 0
	for new(big.Int).Mod(denominator, big.NewInt(2)).Sign() == 0 {
		denominator.Quo(denominator, big.NewInt(2))
		two++
	}
	for new(big.Int).Mod(denominator, big.NewInt(5)).Sign() == 0 {
		denominator.Quo(denominator, big.NewInt(5))
		five++
	}
	places := max(two, five)
	out := value.FloatString(places)
	if strings.Contains(out, ".") {
		out = strings.TrimRight(strings.TrimRight(out, "0"), ".")
	}
	return out
}

// ComposeContainerCPUView validates independent held evidence. Matching
// resources and clocks certify only the descriptive ratio, never an atomic
// snapshot, capacity availability, application health or source authenticity.
func ComposeContainerCPUView(input ContainerCPUViewRequest) (*ContainerCPUView, error) {
	if e := input.Parent.Validate(); e != nil {
		return nil, e
	}
	if !sameContainerContext(input.Parent.Resource, input.Parent.Resource) {
		return nil, fmt.Errorf("container CPU parent requires actual container ID, pod UID and container name")
	}
	// Bound raw evidence and plan payloads before cloning, and replay before
	// marshaling derived result families supplied by callers.
	budget := 0
	for _, r := range []*Result{input.Usage, input.Request, input.Limit} {
		if r == nil {
			continue
		}
		if len(r.Rows) > RowLimit || len(r.Plan.Parameters) > 128 {
			return nil, fmt.Errorf("container CPU input exceeds bounded evidence limits")
		}
		budget += len(r.Plan.SQL) + len(r.ReadAt) + len(r.Basis)
		for k, v := range r.Plan.Parameters {
			budget += len(k) + len(v)
			if budget > 8<<20 {
				return nil, fmt.Errorf("container CPU input exceeds 8 MiB")
			}
		}
		for _, raw := range r.Rows {
			budget += len(raw)
			if budget > 8<<20 {
				return nil, fmt.Errorf("container CPU input exceeds 8 MiB")
			}
		}
		if e := ValidateResult(r); e != nil {
			return nil, e
		}
	}
	encoded, e := json.Marshal(input)
	if e != nil {
		return nil, e
	}
	if len(encoded) > 8<<20 {
		return nil, fmt.Errorf("container CPU input exceeds 8 MiB")
	}
	var in ContainerCPUViewRequest
	if e = json.Unmarshal(encoded, &in); e != nil {
		return nil, e
	}
	view := &ContainerCPUView{Version: ContainerCPUVersion, Parent: in.Parent, Signals: []ContainerCPUSignal{}, Warnings: []string{"Each signal retains independent captured evidence, source read time and observation clock. No paired custody or transactional snapshot is created.", "Matching container ID, pod UID, container name and cluster/namespace qualification permits independent display only. Different full resources, resource schemas, point attributes, clocks or freshness policies withhold percentages.", "Missing inputs are unavailable, not zero or unlimited. Explicit zero is reported evidence; a zero denominator has no percentage. No CPU-time derivative, container aggregation, ownership or health is inferred."}}
	type certified struct {
		result *Result
		sample *metrics.GaugeSample
		stream metrics.Stream
		cores  *big.Rat
	}
	facts := map[string]certified{}
	for _, entry := range []struct {
		role   string
		result *Result
	}{{"usage", in.Usage}, {"request", in.Request}, {"limit", in.Limit}} {
		signal := ContainerCPUSignal{Role: entry.role, Title: containerCPUTitle(entry.role), State: "unavailable", Reason: "not-supplied"}
		r := entry.result
		if r == nil {
			view.Signals = append(view.Signals, signal)
			continue
		}
		d := r.Question.Definition
		s := d.Source.ExactStream
		if d.Kind != "gauge" || d.Source.Resource == nil || d.Source.Resource.Kind != "kubernetes-container" || s == nil || ContainerCPURole(Instrument{Kind: d.Kind, Metric: d.Source.Metric, Unit: d.Source.Unit, Scope: d.Source.Scope, Version: d.Source.Version, SchemaURL: d.Source.SchemaURL}) != entry.role || r.Gauges == nil {
			return nil, fmt.Errorf("container CPU signal requires its exact pinned gauge role and container authority")
		}
		if d.Environment != in.Parent.Environment || r.Question.FromMS != in.Parent.FromMS || r.Question.ToMS != in.Parent.ToMS {
			return nil, fmt.Errorf("container CPU signals require the original parent window and environment")
		}
		from, to := r.Question.BoundsNS()
		if from != ns(in.Parent.FromMS) || to != ns(in.Parent.ToMS) {
			return nil, fmt.Errorf("container CPU signal nanosecond bounds differ from parent window")
		}
		if !sameContainerContext(in.Parent.Resource, s.Resource) {
			return nil, fmt.Errorf("container CPU signal does not match parent identity and qualification")
		}
		if len(r.Gauges.Series) > 1 {
			return nil, fmt.Errorf("container CPU signal has ambiguous series")
		}
		signal.Question = &r.Question
		signal.ReadAt = r.ReadAt
		signal.Basis = r.Basis
		signal.State = "no_data"
		signal.Reason = "no-observations"
		fact := certified{result: r, stream: *s}
		if len(r.Gauges.Series) == 1 {
			series := r.Gauges.Series[0]
			if !sameExactStream(series.Stream, *s) {
				return nil, fmt.Errorf("container CPU includes an unexpected stream")
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
							return nil, fmt.Errorf("invalid container CPU raw row")
						}
						if row.EndNS != sample.AtNS {
							continue
						}
						found = true
						if row.Flags&1 != 0 {
							valid = false
							break
						}
						n, ok := exactContainerCPU(row.Value)
						if !ok || value != nil && value.Cmp(n) != 0 {
							valid = false
							break
						}
						value = n
					}
					if valid && found && value != nil {
						fact.cores = value
						signal.Value = containerCPUDecimal(value)
					} else {
						signal.State = "invalid"
						signal.Reason = "invalid-core-domain"
					}
				}
			}
		}
		facts[entry.role] = fact
		view.Signals = append(view.Signals, signal)
	}
	percentage := func(role string) ContainerCPUPercentage {
		refuse := func(reason string) ContainerCPUPercentage {
			return ContainerCPUPercentage{State: "unavailable", Reason: reason}
		}
		numerator, nok := facts["usage"]
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
		if numerator.cores == nil || denominator.cores == nil {
			return refuse("invalid-core-domain")
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
		if denominator.cores.Sign() == 0 {
			return refuse("zero-denominator")
		}
		ratio := new(big.Rat).Mul(new(big.Rat).Quo(numerator.cores, denominator.cores), big.NewRat(100, 1))
		value, text := ratio.RatString(), ratio.FloatString(1)+"%"
		if ratio.Sign() > 0 && ratio.Cmp(big.NewRat(1, 10)) < 0 {
			text = "<0.1%"
		} else if len(text) > 24 {
			text = "~" + new(big.Float).SetPrec(128).SetRat(ratio).Text('g', 4) + "%"
		}
		if len(value) > 1024 || len(text) > 1024 {
			return refuse("bounded-percentage-domain")
		}
		return ContainerCPUPercentage{State: "observed", Value: value, Text: text}
	}
	view.Percentages = ContainerCPUPercentages{Request: percentage("request"), Limit: percentage("limit")}
	return view, nil
}
