package measurements

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"sort"

	"github.com/casselc/goo/metrics"
)

const GoRuntimeSchema = "https://opentelemetry.io/schemas/1.37.0"

func GoRuntimeMetricTitle(metric string) string {
	switch metric {
	case "go.memory.used":
		return "Go runtime memory used"
	case "go.memory.gc.goal":
		return "Go GC heap goal"
	case "go.goroutine.count":
		return "Go goroutines"
	}
	return ""
}
func goRuntimeShape(i Instrument) bool {
	unit := "By"
	if i.Metric == "go.goroutine.count" {
		unit = "{goroutine}"
	}
	return GoRuntimeMetricTitle(i.Metric) != "" && i.Supported && i.Definition != nil && i.Kind == "stock" && i.Unit == unit && i.SchemaURL == GoRuntimeSchema && i.Temporality == 2 && !i.Monotonic && i.Definition.Source.Resource == nil
}

// GoRuntimePreviewOmissions bounds supplied preview metadata before arithmetic.
func GoRuntimePreviewOmissions(i Instrument) (int, error) {
	if GoRuntimeMetricTitle(i.Metric) == "" || i.CapabilityPlan == nil {
		return 0, nil
	}
	coverage := i.CapabilityPlan.Coverage
	omitted := coverage.OmittedStreams
	if omitted < 0 || omitted > RowLimit || len(coverage.Streams) > RowLimit {
		return 0, fmt.Errorf("runtime preview counts must be within 0..10000")
	}
	extra := len(coverage.Streams) - DiscoveryStreamLimit
	if extra < 0 {
		extra = 0
	}
	if extra > RowLimit-omitted {
		return 0, fmt.Errorf("runtime preview counts must be within 0..10000")
	}
	return omitted + extra, nil
}

// GoRuntimeCandidates pins positive preview streams; absence from a bounded
// preview is never interpreted as source absence or unsupported telemetry.
func GoRuntimeCandidates(i Instrument) []Instrument {
	out := []Instrument{}
	if _, e := GoRuntimePreviewOmissions(i); e != nil {
		return out
	}
	if !goRuntimeShape(i) || i.CapabilityPlan == nil {
		return out
	}
	d := i.Definition
	if d.Validate() != nil || d.Kind != i.Kind || d.Source.Metric != i.Metric || d.Source.Unit != i.Unit || d.Source.Scope != i.Scope || d.Source.Version != i.Version || d.Source.SchemaURL != i.SchemaURL {
		return out
	}
	streams := i.CapabilityPlan.Coverage.Streams
	if len(streams) > DiscoveryStreamLimit {
		streams = streams[:DiscoveryStreamLimit]
	}
	byRevision := map[string]Instrument{}
	for _, preview := range streams {
		stream := preview.Stream
		if stream.Attributes == nil {
			stream.Attributes = map[string]string{}
		}
		if d.Source.ExactStream != nil && !sameExactStream(*d.Source.ExactStream, stream) {
			continue
		}
		candidate := *d
		candidate.Source.ExactStream = &stream
		candidate.StockDomain = "nonnegative-integer"
		// Fresh planned definitions use a canonical generated authoring identity.
		// Display names cannot multiply the same stream+evaluation policy option.
		seed, err := definitionSeed(candidate)
		if err != nil {
			continue
		}
		candidate.Name = "go-runtime-" + seed[:24]
		candidate.Title = GoRuntimeMetricTitle(i.Metric)
		key, err := candidate.Revision()
		if err != nil {
			continue
		}
		if _, exists := byRevision[key]; exists {
			continue
		}
		// Detach identity maps and source selectors from supplied metadata.
		encoded, _ := json.Marshal(candidate)
		var detached Definition
		if json.Unmarshal(encoded, &detached) != nil {
			continue
		}
		copy := i
		copy.Definition = &detached
		// Each option represents one pinned stream. Do not replicate the entire
		// discovery preview across every expanded option or alias caller metadata.
		encodedInstrument, _ := json.Marshal(copy)
		var exact Instrument
		if json.Unmarshal(encodedInstrument, &exact) != nil {
			continue
		}
		exact.Records = preview.Records
		exact.CapabilityPlan.Coverage = observedCoverage(exact, []DiscoveryObservation{{Resource: stream.Resource, Attributes: stream.Attributes, ResourceSchemaURL: stream.ResourceSchemaURL}})
		if len(exact.CapabilityPlan.Coverage.Streams) > 0 {
			exact.CapabilityPlan.Coverage.Streams[0].Records = preview.Records
		}
		byRevision[key] = exact
	}
	for _, key := range slices.Sorted(maps.Keys(byRevision)) {
		out = append(out, byRevision[key])
	}
	return out
}
func GoRuntimeExactCandidate(i Instrument) bool {
	if !goRuntimeShape(i) || i.Definition.StockDomain != "nonnegative-integer" || i.Definition.Source.ExactStream == nil || i.CapabilityPlan == nil {
		return false
	}
	d := i.Definition
	if d.Validate() != nil || d.Kind != i.Kind || d.Source.Metric != i.Metric || d.Source.Unit != i.Unit || d.Source.Scope != i.Scope || d.Source.Version != i.Version || d.Source.SchemaURL != i.SchemaURL {
		return false
	}
	streams := i.CapabilityPlan.Coverage.Streams
	if len(streams) > DiscoveryStreamLimit {
		streams = streams[:DiscoveryStreamLimit]
	}
	for _, preview := range streams {
		if sameExactStream(*d.Source.ExactStream, preview.Stream) {
			return true
		}
	}
	return false
}
func goRuntimeProcessKey(stream metrics.Stream) string {
	encoded, _ := json.Marshal(struct {
		Instance                                     string
		Resource                                     map[string]string
		ResourceSchemaURL, Scope, Version, SchemaURL string
	}{stream.Resource["service.instance.id"], stream.Resource, stream.ResourceSchemaURL, stream.Scope, stream.Version, stream.SchemaURL})
	return string(encoded)
}
func goRuntimeMemoryOrder(stream metrics.Stream) string {
	value, present := stream.Attributes["go.memory.type"]
	if !present {
		return "4:"
	}
	switch value {
	case "other":
		return "0:"
	case "stack":
		return "1:"
	case "":
		return "3:"
	}
	return "2:" + value
}

// GoRuntimeDefaults chooses one memory component and one heap-goal observation
// for a shared process identity when possible. This is display context, never a
// summed heap value or atomic GC snapshot. Goroutines always require selection.
func GoRuntimeDefaults(instruments []Instrument) map[string]bool {
	type group struct{ memory, goal []string }
	groups := map[string]*group{}
	memoryOrder := map[string]string{}
	for _, i := range instruments {
		if !GoRuntimeExactCandidate(i) || i.Metric == "go.goroutine.count" {
			continue
		}
		process := goRuntimeProcessKey(*i.Definition.Source.ExactStream)
		if groups[process] == nil {
			groups[process] = &group{}
		}
		key, err := i.Definition.Revision()
		if err != nil {
			continue
		}
		if i.Metric == "go.memory.used" {
			groups[process].memory = append(groups[process].memory, key)
			memoryOrder[key] = goRuntimeMemoryOrder(*i.Definition.Source.ExactStream)
		} else {
			groups[process].goal = append(groups[process].goal, key)
		}
	}
	names := []string{}
	for key, g := range groups {
		sort.Slice(g.memory, func(i, j int) bool {
			a, b := g.memory[i], g.memory[j]
			if memoryOrder[a] != memoryOrder[b] {
				return memoryOrder[a] < memoryOrder[b]
			}
			return a < b
		})
		sort.Strings(g.goal)
		names = append(names, key)
	}
	sort.Strings(names)
	chosen := map[string]bool{}
	for _, key := range names {
		g := groups[key]
		if len(g.memory) > 0 && len(g.goal) > 0 {
			chosen[g.memory[0]] = true
			chosen[g.goal[0]] = true
			return chosen
		}
	}
	// With no complete pair, choose at most one of each from independent streams,
	// preserving their distinct identity and clocks rather than implying a join.
	for _, key := range names {
		g := groups[key]
		if len(g.memory) > 0 {
			chosen[g.memory[0]] = true
			break
		}
	}
	for _, key := range names {
		g := groups[key]
		if len(g.goal) > 0 {
			chosen[g.goal[0]] = true
			break
		}
	}
	return chosen
}
