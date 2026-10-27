package measurements

import (
	"github.com/casselc/goo/metrics"
	"math/big"
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
func exactContainerCPU(raw string) (*big.Rat, bool) { return metrics.ExportedNonnegativeDecimal(raw) }
func containerCPUDecimal(value *big.Rat) string     { return metrics.CounterDecimal(value) }

func ComposeContainerCPUView(input ContainerCPUViewRequest) (*ContainerCPUView, error) {
	in, e := cloneContainerResourceEvidence(input, input.Parent, []*Result{input.Usage, input.Request, input.Limit}, "container CPU")
	if e != nil {
		return nil, e
	}
	view := &ContainerCPUView{Version: ContainerCPUVersion, Parent: in.Parent, Signals: []ContainerCPUSignal{}, Warnings: []string{"Each signal retains independent captured evidence, source read time and observation clock. No paired custody or transactional snapshot is created.", "Matching container ID, pod UID, container name and cluster/namespace qualification permits independent display only. Different full resources, resource schemas, point attributes, clocks or freshness policies withhold percentages.", "Missing inputs are unavailable, not zero or unlimited. Explicit zero is reported evidence; a zero denominator has no percentage. No CPU-time derivative, container aggregation, ownership or health is inferred."}}
	policy := containerCPUPolicy()
	signals, facts, e := composeContainerResourceSignals(in.Parent, []containerResourceEntry{{"usage", in.Usage}, {"request", in.Request}, {"limit", in.Limit}}, policy)
	if e != nil {
		return nil, e
	}
	for _, signal := range signals {
		view.Signals = append(view.Signals, ContainerCPUSignal(signal))
	}
	view.Percentages = ContainerCPUPercentages{Request: ContainerCPUPercentage(containerResourcePercentage(facts, "usage", "request", policy)), Limit: ContainerCPUPercentage(containerResourcePercentage(facts, "usage", "limit", policy))}
	return view, nil
}

func containerCPUPolicy() containerResourcePolicy {
	return containerResourcePolicy{label: "container CPU", invalidReason: "invalid-core-domain", role: ContainerCPURole, title: containerCPUTitle, parse: exactContainerCPU, format: containerCPUDecimal, percentage: func(ratio *big.Rat) containerResourcePercentageValue {
		value, text := ratio.RatString(), ratio.FloatString(1)+"%"
		if ratio.Sign() > 0 && ratio.Cmp(big.NewRat(1, 10)) < 0 {
			text = "<0.1%"
		} else if len(text) > 24 {
			text = "~" + new(big.Float).SetPrec(128).SetRat(ratio).Text('g', 4) + "%"
		}
		if len(value) > 1024 || len(text) > 1024 {
			return containerResourcePercentageValue{State: "unavailable", Reason: "bounded-percentage-domain"}
		}
		return containerResourcePercentageValue{State: "observed", Value: value, Text: text}
	}}
}
