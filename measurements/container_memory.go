package measurements

import (
	"math/big"
	"regexp"
	"strings"
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
	in, e := cloneContainerResourceEvidence(input, input.Parent, []*Result{input.WorkingSet, input.Request, input.Limit}, "container memory")
	if e != nil {
		return nil, e
	}
	view := &ContainerMemoryView{Version: ContainerMemoryVersion, Parent: in.Parent, Signals: []ContainerMemorySignal{}, Warnings: []string{"Each signal retains independent captured evidence, source read time and observation clock. No paired custody or transactional snapshot is created.", "Matching container ID, pod UID, container name and cluster/namespace qualification permits independent display only. Different full resources, resource schemas, point attributes, clocks or freshness policies withhold percentages.", "Missing inputs are unavailable, not zero or unlimited. Explicit zero is reported evidence; a zero denominator has no percentage. No CPU, container aggregation, ownership or health is inferred."}}
	policy := containerResourcePolicy{label: "container memory", invalidReason: "invalid-byte-domain", role: ContainerMemoryRole, title: containerMemoryTitle, parse: func(raw string) (*big.Rat, bool) {
		n, ok := exactContainerBytes(raw)
		if !ok {
			return nil, false
		}
		return new(big.Rat).SetInt(n), true
	}, format: func(value *big.Rat) string { return value.Num().String() }, percentage: func(ratio *big.Rat) containerResourcePercentageValue {
		return containerResourcePercentageValue{State: "observed", Value: ratio.RatString(), Text: ratio.FloatString(1) + "%"}
	}}
	signals, facts, e := composeContainerResourceSignals(in.Parent, []containerResourceEntry{{"working-set", in.WorkingSet}, {"request", in.Request}, {"limit", in.Limit}}, policy)
	if e != nil {
		return nil, e
	}
	for _, signal := range signals {
		view.Signals = append(view.Signals, ContainerMemorySignal(signal))
	}
	view.Percentages = ContainerMemoryPercentages{Request: ContainerMemoryPercentage(containerResourcePercentage(facts, "working-set", "request", policy)), Limit: ContainerMemoryPercentage(containerResourcePercentage(facts, "working-set", "limit", policy))}
	return view, nil
}
