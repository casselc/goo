package measurements

import (
	"encoding/json"
	"fmt"
	"maps"
	"sort"
)

const PodReceiverProfile = "collector-kubeletstats-v0.137.0"
const PodSemconvProfile = "otel-k8s-v1.37.0"

type PodDashboardRequest struct {
	Discovery PodMetricDiscovery `json:"discovery"`
}
type PodMetricWidget struct {
	Key      string   `json:"key"`
	Title    string   `json:"title"`
	State    string   `json:"state"`
	Profile  string   `json:"profile"`
	Grain    string   `json:"grain"`
	Question Question `json:"question"`
}
type PodDashboardPlan struct {
	Version            string                     `json:"version"`
	Question           PodMetricDiscoveryQuestion `json:"question"`
	ResourceKey        string                     `json:"resourceKey"`
	State              string                     `json:"state"`
	Widgets            []PodMetricWidget          `json:"widgets"`
	Candidates         []PodMetricWidget          `json:"candidates"`
	CandidatesOmitted  int                        `json:"candidatesOmitted"`
	InstrumentsOmitted int                        `json:"instrumentsOmitted"`
	StreamsOmitted     int                        `json:"streamsOmitted"`
	Warnings           []string                   `json:"warnings"`
}

func podMetricProfile(i Instrument) string {
	if i.Kind != "gauge" || i.Unit != "By" || i.Temporality != 0 || i.Scope == "" {
		return ""
	}
	if i.Scope == KubeletStatsScope && i.Version == KubeletStatsVersion && i.SchemaURL == "" && (i.Metric == "k8s.pod.memory.usage" || i.Metric == "k8s.pod.memory.working_set") {
		return PodReceiverProfile
	}
	if i.SchemaURL == PodMetricsSchema && i.Metric == "k8s.pod.memory.usage" {
		return PodSemconvProfile
	}
	return ""
}
func PlanPodDashboard(input PodDashboardRequest) (*PodDashboardPlan, error) {
	discovery := input.Discovery
	if e := discovery.Question.Validate(); e != nil {
		return nil, e
	}
	if discovery.Version != PodMetricDiscoveryVersion || discovery.Key != hostResourceKey(discovery.Question.Resource, *discovery.Question.ResourceSchemaURL) || len(discovery.Instruments) > PodMetricInstrumentLimit || discovery.InstrumentsOmitted < 0 || discovery.InstrumentsOmitted > RowLimit || discovery.Records < 0 || discovery.Records > RowLimit {
		return nil, fmt.Errorf("pod dashboard requires bounded matching discovery metadata")
	}
	encoded, e := json.Marshal(input)
	if e != nil {
		return nil, e
	}
	var in PodDashboardRequest
	if e = json.Unmarshal(encoded, &in); e != nil {
		return nil, e
	}
	discovery = in.Discovery
	q := discovery.Question
	result := &PodDashboardPlan{Version: PodDashboardVersion, Question: q, ResourceKey: discovery.Key, State: "unconfirmed", Widgets: []PodMetricWidget{}, Candidates: []PodMetricWidget{}, InstrumentsOmitted: discovery.InstrumentsOmitted, Warnings: []string{"Reported metadata for one exact pod resource variant and known schema, not a verified pod entity, health or complete Kubernetes inventory.", "Collector kubeletstats v0.137.0 memory usage and working set gauges remain separate from the semconv v1.37.0 usage-only profile. No aliases, container sums, CPU/limit ratios or ownership are inferred.", "At most two exact-stream default reads run; all values, freshness and read clocks require independent evaluation. Negative, fractional or unsafe byte values remain raw gauge evidence but cannot be displayed as valid byte counts.", "Ordinal pod phase and container readiness from other receiver profiles are not supported state or health panels in this slice. Missing or omitted metadata does not establish absence."}}
	type candidate struct {
		widget PodMetricWidget
		family string
	}
	all := []candidate{}
	seen := map[string]bool{}
	for _, i := range discovery.Instruments {
		if i.CapabilityPlan == nil {
			continue
		}
		coverage := i.CapabilityPlan.Coverage
		if coverage.OmittedStreams < 0 || coverage.OmittedStreams > RowLimit || len(coverage.Streams) > RowLimit {
			return nil, fmt.Errorf("pod preview counts must be within 0..10000")
		}
		preview := coverage.Streams
		extra := 0
		if len(preview) > DiscoveryStreamLimit {
			extra = len(preview) - DiscoveryStreamLimit
			preview = preview[:DiscoveryStreamLimit]
		}
		if extra > RowLimit-coverage.OmittedStreams {
			return nil, fmt.Errorf("pod preview counts must be within 0..10000")
		}
		result.StreamsOmitted += coverage.OmittedStreams + extra
		profile := podMetricProfile(i)
		if profile == "" || !i.Supported || i.Definition == nil {
			continue
		}
		def := i.Definition
		if def.Validate() != nil || def.Kind != i.Kind || def.Source.Metric != i.Metric || def.Source.Unit != i.Unit || def.Source.Scope != i.Scope || def.Source.Version != i.Version || def.Source.SchemaURL != i.SchemaURL || def.Source.Resource == nil || def.Source.Resource.Kind != "kubernetes-pod" || def.Source.Resource.ID != q.Resource["k8s.pod.uid"] || def.Environment != q.Environment {
			continue
		}
		for _, observed := range preview {
			stream := observed.Stream
			if stream.ResourceSchemaURL != *q.ResourceSchemaURL || !maps.Equal(stream.Resource, q.Resource) || stream.Scope != i.Scope || stream.Version != i.Version || stream.SchemaURL != i.SchemaURL || stream.Name != i.Metric || stream.Unit != i.Unit {
				continue
			}
			if e := validateHostLabels(stream.Attributes); e != nil {
				return nil, e
			}
			// Discovery's initial exact stream is a convenience preview, not authored
			// exclusive intent. Fresh questions pin each matching positive preview;
			// original Source.Attributes selectors still validate below.
			if def.Source.ExactStream == nil {
				continue
			}
			selected := *def
			if stream.Attributes == nil {
				stream.Attributes = map[string]string{}
			}
			selected.Source.ExactStream = &stream
			selected.Name = ""
			selected.Title = ""
			selected.Name = "pod-observation-" + Hash(selected)[:24]
			selected.Title = "Pod memory usage"
			if i.Metric == "k8s.pod.memory.working_set" {
				selected.Title = "Pod memory working set"
			}
			question := Question{Definition: selected, FromMS: q.FromMS, ToMS: q.ToMS}
			if question.Validate() != nil {
				continue
			}
			// Names do not multiply equivalent exact stream/policy choices.
			identity := selected
			identity.Name = ""
			identity.Title = ""
			dedupKey := Hash(identity)
			if seen[dedupKey] {
				continue
			}
			seen[dedupKey] = true
			family := "usage"
			title := "Pod memory usage"
			if i.Metric == "k8s.pod.memory.working_set" {
				family = "working-set"
				title = "Pod memory working set"
			}
			all = append(all, candidate{widget: PodMetricWidget{Key: Hash(question), Title: title, State: "candidate", Profile: profile, Grain: "pod-resource-stream", Question: question}, family: family})
		}
	}
	sort.Slice(all, func(a, b int) bool {
		left, right := all[a], all[b]
		if left.widget.Profile != right.widget.Profile {
			return left.widget.Profile == PodReceiverProfile
		}
		if left.family != right.family {
			return left.family == "usage"
		}
		return left.widget.Key < right.widget.Key
	})
	defaultProfile := ""
	if len(all) > 0 {
		defaultProfile = all[0].widget.Profile
	}
	families := map[string]bool{}
	defaultKeys := map[string]bool{}
	for _, c := range all {
		if c.widget.Profile == defaultProfile && !families[c.family] && len(defaultKeys) < 2 {
			defaultKeys[c.widget.Key] = true
			families[c.family] = true
		}
	}
	// Protect selected defaults before the presentation cap: many usage streams
	// cannot hide a positive working-set family from the default pair.
	sort.SliceStable(all, func(a, b int) bool { return defaultKeys[all[a].widget.Key] && !defaultKeys[all[b].widget.Key] })
	for index, c := range all {
		if index >= PodMetricCandidateLimit {
			result.CandidatesOmitted++
			continue
		}
		result.Candidates = append(result.Candidates, c.widget)
		if defaultKeys[c.widget.Key] {
			result.Widgets = append(result.Widgets, c.widget)
		}
	}
	if len(result.Candidates) > 0 {
		result.State = "candidates"
	}
	return result, nil
}
