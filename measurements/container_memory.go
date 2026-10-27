package measurements

import (
	"encoding/json"
	"fmt"
	"maps"
	"math/big"
	"regexp"
	"strings"

	"github.com/casselc/goo/metrics"
)

const ContainerMemoryVersion = "goo/container-memory-view/v1"
const ContainerMemoryKubeletProfile = "container-memory-kubeletstats-v0.137.0"
const ContainerMemoryClusterProfile = "container-memory-k8scluster-v0.137.0"
const K8sClusterScope = "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver"

func ContainerMemoryRole(i Instrument) string {
	if i.Kind != "gauge" || i.Temporality != 0 || i.Monotonic || i.Unit != "By" || i.Version != "0.137.0" || i.SchemaURL != "" {
		return ""
	}
	if i.Scope == KubeletStatsScope && i.Metric == "container.memory.working_set" {
		return "working-set"
	}
	if i.Scope == K8sClusterScope {
		switch i.Metric {
		case "k8s.container.memory_request":
			return "request"
		case "k8s.container.memory_limit":
			return "limit"
		}
	}
	return ""
}
func containerMemoryProfile(role string) string {
	if role == "working-set" {
		return ContainerMemoryKubeletProfile
	}
	return ContainerMemoryClusterProfile
}
func containerMemoryTitle(role string) string {
	switch role {
	case "working-set":
		return "Container memory working set"
	case "request":
		return "Reported memory request"
	default:
		return "Reported memory limit"
	}
}

type ContainerMemoryViewRequest struct {
	Parent     PodMetricDiscoveryQuestion `json:"parent"`
	WorkingSet *Result                    `json:"workingSet,omitempty"`
	Request    *Result                    `json:"request,omitempty"`
	Limit      *Result                    `json:"limit,omitempty"`
}
type ContainerMemorySignal struct {
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
type ContainerMemoryPercentage struct {
	State  string `json:"state"`
	Reason string `json:"reason,omitempty"`
	Value  string `json:"value,omitempty"`
	Text   string `json:"text,omitempty"`
}
type ContainerMemoryPercentages struct {
	Request ContainerMemoryPercentage `json:"request"`
	Limit   ContainerMemoryPercentage `json:"limit"`
}
type ContainerMemoryView struct {
	Version     string                     `json:"version"`
	Parent      PodMetricDiscoveryQuestion `json:"parent"`
	Signals     []ContainerMemorySignal    `json:"signals"`
	Percentages ContainerMemoryPercentages `json:"percentages"`
	Warnings    []string                   `json:"warnings"`
}

func sameContainerContext(a, b map[string]string) bool {
	for _, key := range []string{"container.id", "k8s.pod.uid", "k8s.container.name"} {
		if a[key] == "" || !text(a[key], 512) || !text(b[key], 512) || a[key] != b[key] {
			return false
		}
	}
	for _, key := range []string{"k8s.cluster.uid", "k8s.cluster.name", "k8s.namespace.name"} {
		av, ap := a[key]
		bv, bp := b[key]
		if ap != bp || av != bv {
			return false
		}
	}
	return true
}

// exactContainerBytes certifies the original decimal, rather than a rounded
// float64 approximation. Bounded syntax prevents unbounded exponent arithmetic.
var containerByteDecimal = regexp.MustCompile(`^[+-]?([0-9]+(\.[0-9]*)?|\.[0-9]+)([eE][+-]?[0-9]+)?$`)

func exactContainerBytes(raw string) (*big.Int, bool) {
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
	if !ok || !value.IsInt() || value.Sign() < 0 || value.Num().Cmp(new(big.Int).Lsh(big.NewInt(1), 53)) >= 0 {
		return nil, false
	}
	return new(big.Int).Set(value.Num()), true
}

// ComposeContainerMemoryView validates independent held evidence. Matching
// resources and clocks certify only the descriptive ratio, never an atomic
// snapshot, capacity availability, application health or source authenticity.
func ComposeContainerMemoryView(input ContainerMemoryViewRequest) (*ContainerMemoryView, error) {
	if e := input.Parent.Validate(); e != nil {
		return nil, e
	}
	if !sameContainerContext(input.Parent.Resource, input.Parent.Resource) {
		return nil, fmt.Errorf("container memory parent requires actual container ID, pod UID and container name")
	}
	// Bound raw evidence and plan payloads before cloning, and replay before
	// marshaling derived result families supplied by callers.
	budget := 0
	for _, r := range []*Result{input.WorkingSet, input.Request, input.Limit} {
		if r == nil {
			continue
		}
		if len(r.Rows) > RowLimit || len(r.Plan.Parameters) > 128 {
			return nil, fmt.Errorf("container memory input exceeds bounded evidence limits")
		}
		budget += len(r.Plan.SQL) + len(r.ReadAt) + len(r.Basis)
		for k, v := range r.Plan.Parameters {
			budget += len(k) + len(v)
			if budget > 8<<20 {
				return nil, fmt.Errorf("container memory input exceeds 8 MiB")
			}
		}
		for _, raw := range r.Rows {
			budget += len(raw)
			if budget > 8<<20 {
				return nil, fmt.Errorf("container memory input exceeds 8 MiB")
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
		return nil, fmt.Errorf("container memory input exceeds 8 MiB")
	}
	var in ContainerMemoryViewRequest
	if e = json.Unmarshal(encoded, &in); e != nil {
		return nil, e
	}
	view := &ContainerMemoryView{Version: ContainerMemoryVersion, Parent: in.Parent, Signals: []ContainerMemorySignal{}, Warnings: []string{"Each signal retains independent captured evidence, source read time and observation clock. No paired custody or transactional snapshot is created.", "Matching container ID, pod UID, container name and cluster/namespace qualification permits independent display only. Different full resources, resource schemas, point attributes, clocks or freshness policies withhold percentages.", "Missing inputs are unavailable, not zero or unlimited. Explicit zero is reported evidence; a zero denominator has no percentage. No CPU, container aggregation, ownership or health is inferred."}}
	type certified struct {
		result *Result
		sample *metrics.GaugeSample
		stream metrics.Stream
		bytes  *big.Int
	}
	facts := map[string]certified{}
	for _, entry := range []struct {
		role   string
		result *Result
	}{{"working-set", in.WorkingSet}, {"request", in.Request}, {"limit", in.Limit}} {
		signal := ContainerMemorySignal{Role: entry.role, Title: containerMemoryTitle(entry.role), State: "unavailable", Reason: "not-supplied"}
		r := entry.result
		if r == nil {
			view.Signals = append(view.Signals, signal)
			continue
		}
		d := r.Question.Definition
		s := d.Source.ExactStream
		if d.Kind != "gauge" || d.Source.Resource == nil || d.Source.Resource.Kind != "kubernetes-container" || s == nil || ContainerMemoryRole(Instrument{Kind: d.Kind, Metric: d.Source.Metric, Unit: d.Source.Unit, Scope: d.Source.Scope, Version: d.Source.Version, SchemaURL: d.Source.SchemaURL}) != entry.role || r.Gauges == nil {
			return nil, fmt.Errorf("container memory signal requires its exact pinned gauge role and container authority")
		}
		if d.Environment != in.Parent.Environment || r.Question.FromMS != in.Parent.FromMS || r.Question.ToMS != in.Parent.ToMS {
			return nil, fmt.Errorf("container memory signals require the original parent window and environment")
		}
		from, to := r.Question.BoundsNS()
		if from != ns(in.Parent.FromMS) || to != ns(in.Parent.ToMS) {
			return nil, fmt.Errorf("container memory signal nanosecond bounds differ from parent window")
		}
		if !sameContainerContext(in.Parent.Resource, s.Resource) {
			return nil, fmt.Errorf("container memory signal does not match parent identity and qualification")
		}
		if len(r.Gauges.Series) > 1 {
			return nil, fmt.Errorf("container memory signal has ambiguous series")
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
				return nil, fmt.Errorf("container memory includes an unexpected stream")
			}
			if len(series.Samples) > 0 {
				sample := series.Samples[len(series.Samples)-1]
				fact.sample = &sample
				signal.State = sample.State
				signal.ObservationState = sample.State
				signal.Reason = sample.Reason
				signal.AtNS = sample.AtNS
				if sample.Value != nil {
					// Every raw record at the selected time must certify the same integer;
					// duplicate conflicts and missing flags remain governed by replay above.
					var value *big.Int
					valid := true
					found := false
					for _, raw := range r.Rows {
						var row metricRow
						if json.Unmarshal(raw, &row) != nil {
							return nil, fmt.Errorf("invalid container memory raw row")
						}
						if row.EndNS != sample.AtNS {
							continue
						}
						found = true
						if row.Flags&1 != 0 {
							valid = false
							break
						}
						n, ok := exactContainerBytes(row.Value)
						if !ok || value != nil && value.Cmp(n) != 0 {
							valid = false
							break
						}
						value = n
					}
					if valid && found && value != nil {
						fact.bytes = value
						signal.Value = value.String()
					} else {
						signal.State = "invalid"
						signal.Reason = "invalid-byte-domain"
					}
				}
			}
		}
		facts[entry.role] = fact
		view.Signals = append(view.Signals, signal)
	}
	percentage := func(role string) ContainerMemoryPercentage {
		refuse := func(reason string) ContainerMemoryPercentage {
			return ContainerMemoryPercentage{State: "unavailable", Reason: reason}
		}
		numerator, nok := facts["working-set"]
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
		if numerator.bytes == nil || denominator.bytes == nil {
			return refuse("invalid-byte-domain")
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
		if denominator.bytes.Sign() == 0 {
			return refuse("zero-denominator")
		}
		ratio := new(big.Rat).SetFrac(new(big.Int).Mul(numerator.bytes, big.NewInt(100)), denominator.bytes)
		return ContainerMemoryPercentage{State: "observed", Value: ratio.RatString(), Text: ratio.FloatString(1) + "%"}
	}
	view.Percentages = ContainerMemoryPercentages{Request: percentage("request"), Limit: percentage("limit")}
	return view, nil
}
