package measurements

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/metrics"
	"sort"
	"strconv"
)

const HostDashboardVersion = "goo/host-dashboard/v1"
const HostMetricsSchema = "https://opentelemetry.io/schemas/1.37.0"
const HostCandidateLimit = 16

// HostCPUTimeInstrument is the pinned monotonic cumulative seconds binding.
func HostCPUTimeInstrument(i Instrument) bool {
	return i.SchemaURL == HostMetricsSchema && i.Kind == "counter" && i.Metric == "system.cpu.time" && i.Unit == "s" && i.Temporality == 2 && i.Monotonic
}

type HostDashboardRequest struct {
	Question HostDiscoveryQuestion `json:"question"`
	Host     HostEntry             `json:"host"`
}
type HostWidget struct {
	Key      string   `json:"key"`
	Title    string   `json:"title"`
	State    string   `json:"state"`
	Grain    string   `json:"grain"`
	Question Question `json:"question"`
}
type HostDashboardPlan struct {
	Version            string                `json:"version"`
	Question           HostDiscoveryQuestion `json:"question"`
	HostKey            string                `json:"hostKey"`
	State              string                `json:"state"`
	Widgets            []HostWidget          `json:"widgets"`
	Candidates         []HostWidget          `json:"candidates"`
	CandidatesOmitted  int                   `json:"candidatesOmitted"`
	InstrumentsOmitted int                   `json:"instrumentsOmitted"`
	StreamsOmitted     int                   `json:"streamsOmitted"`
	Warnings           []string              `json:"warnings"`
}

func PlanHostDashboard(input HostDashboardRequest) (*HostDashboardPlan, error) {
	if e := input.Question.Validate(); e != nil {
		return nil, e
	}
	h := input.Host
	if e := validateHostLabels(h.Resource); e != nil {
		return nil, e
	}
	if h.Resource["host.id"] == "" || !text(h.Resource["host.id"], 512) || !text(h.ResourceSchemaURL, 1024) || h.Resource["deployment.environment.name"] != input.Question.Environment || h.Key != hostResourceKey(h.Resource, h.ResourceSchemaURL) {
		return nil, fmt.Errorf("host dashboard requires an exact reported resource identity in the question environment")
	}
	if len(h.Instruments) > HostInstrumentLimit || h.InstrumentsOmitted < 0 || h.InstrumentsOmitted > RowLimit || h.Records < 0 || h.Records > RowLimit {
		return nil, fmt.Errorf("host dashboard metadata exceeds bounded discovery limits")
	}
	// Retain detached supplied facts; derived questions cannot mutate the inventory.
	b, e := json.Marshal(input)
	if e != nil {
		return nil, e
	}
	var in HostDashboardRequest
	if e = json.Unmarshal(b, &in); e != nil {
		return nil, e
	}
	h = in.Host
	out := &HostDashboardPlan{Version: HostDashboardVersion, Question: in.Question, HostKey: h.Key, State: "unconfirmed", Widgets: []HostWidget{}, Candidates: []HostWidget{}, Warnings: []string{"Reported host resource identity only; equal host.id across different full maps does not prove the same physical host. Host name is display only.", "Candidates come from supplied metadata, not fresh source evidence, health, capacity or whole-host utilization. Each selected question is an independent exact stream; a logical CPU is not a host aggregate.", "At most two default measurements run. Remaining candidates require explicit selection; absent or omitted preview metadata is not absence."}}
	type candidate struct {
		widget   HostWidget
		family   string
		priority int
		numeric  bool
		core     uint64
	}
	candidates := []candidate{}
	seen := map[string]bool{}
	out.InstrumentsOmitted = h.InstrumentsOmitted
	for _, i := range h.Instruments {
		if i.CapabilityPlan == nil {
			continue
		}
		coverage := i.CapabilityPlan.Coverage
		if coverage.OmittedStreams < 0 || coverage.OmittedStreams > RowLimit || len(coverage.Streams) > RowLimit {
			return nil, fmt.Errorf("host preview counts must be within 0..10000")
		}
		preview := coverage.Streams
		extra := 0
		if len(preview) > DiscoveryStreamLimit {
			extra = len(preview) - DiscoveryStreamLimit
			preview = preview[:DiscoveryStreamLimit]
		}
		if extra > RowLimit-coverage.OmittedStreams {
			return nil, fmt.Errorf("host preview counts must be within 0..10000")
		}
		out.StreamsOmitted += extra + coverage.OmittedStreams
		family := ""
		if HostNetworkInstrument(i) {
			family = "network"
		}
		if HostFilesystemInstrument(i) {
			family = "filesystem"
		}
		if HostDiskInstrument(i) {
			family = "disk"
		}
		if HostCPUTimeInstrument(i) {
			family = "cpu-time"
		}
		if i.SchemaURL == HostMetricsSchema && i.Kind == "gauge" && i.Metric == "system.cpu.utilization" && i.Unit == "1" {
			family = "cpu"
		}
		if i.SchemaURL == HostMetricsSchema && i.Kind == "stock" && i.Metric == "system.memory.usage" && i.Unit == "By" && i.Temporality == 2 && !i.Monotonic {
			family = "memory"
		}
		if i.SchemaURL == HostMetricsSchema && i.Kind == "gauge" && i.Metric == "system.memory.utilization" && i.Unit == "1" {
			family = "memory-fraction"
		}
		if family == "" {
			continue
		}
		for _, observed := range preview {
			s := observed.Stream
			if family == "network" {
				name := s.Attributes["network.interface.name"]
				direction := s.Attributes["network.io.direction"]
				if name == "" || !text(name, 128) || (direction != "receive" && direction != "transmit") {
					continue
				}
			}
			if family == "disk" {
				name, direction := s.Attributes["system.device"], s.Attributes["disk.io.direction"]
				if name == "" || !text(name, 128) || (direction != "read" && direction != "write") {
					continue
				}
			}
			if family == "filesystem" {
				device, mount := s.Attributes["system.device"], s.Attributes["system.filesystem.mountpoint"]
				if device == "" || mount == "" || !text(device, 512) || !text(mount, 512) {
					continue
				}
				state, hasState := s.Attributes["system.filesystem.state"]
				if i.Metric == "system.filesystem.usage" && state != "used" || i.Metric == "system.filesystem.limit" && hasState {
					continue
				}
			}
			if s.ResourceSchemaURL != h.ResourceSchemaURL || hostResourceKey(s.Resource, s.ResourceSchemaURL) != h.Key || s.Scope != i.Scope || s.Version != i.Version || s.SchemaURL != i.SchemaURL || s.Name != i.Metric || s.Unit != i.Unit {
				continue
			}
			if e := validateHostLabels(s.Attributes); e != nil {
				return nil, e
			}
			d := Definition{APIVersion: Version, Name: "host-signal", Title: i.Metric, Kind: i.Kind, Environment: in.Question.Environment, LookbackMinutes: 2, MaxGapMS: 60000, Grain: "per-stream", Source: Binding{Resource: &ResourceBinding{Kind: "host", ID: h.Resource["host.id"]}, Services: []string{}, Scope: i.Scope, Version: i.Version, SchemaURL: i.SchemaURL, Metric: i.Metric, Unit: i.Unit, ExactStream: &s}}
			if family == "cpu-time" || family == "network" || family == "disk" {
				d.CounterDomain = metrics.DecimalCounterDomain
			}
			if family == "memory" || family == "filesystem" {
				d.StockDomain = "nonnegative-integer"
			}
			q := Question{Definition: d, FromMS: in.Question.FromMS, ToMS: in.Question.ToMS}
			if q.Validate() != nil {
				continue
			}
			key := Hash(q)
			if seen[key] {
				continue
			}
			seen[key] = true
			title := "Memory usage"
			priority := 0
			numeric := false
			core := uint64(0)
			if family == "cpu" || family == "cpu-time" {
				title = "CPU utilization"
				if family == "cpu-time" {
					title = "CPU activity (CPU-seconds/s)"
				}
				mode, ok := s.Attributes["cpu.mode"]
				if ok {
					title += " · mode " + strconv.Quote(mode)
				} else {
					title += " · mode unreported"
				}
				if mode != "user" {
					priority = 1
				}
				label, ok := s.Attributes["cpu.logical_number"]
				if ok {
					title += " · logical CPU " + strconv.Quote(label)
					n, err := strconv.ParseUint(label, 10, 32)
					if err == nil && strconv.FormatUint(n, 10) == label {
						numeric = true
						core = n
					}
				} else {
					title += " · logical CPU unreported"
				}
			}
			if family != "cpu" && family != "cpu-time" && family != "network" && family != "disk" && family != "filesystem" {
				if family == "memory-fraction" {
					title = "Reported memory utilization fraction"
				}
				state, ok := s.Attributes["system.memory.state"]
				if ok {
					title += " · state " + strconv.Quote(state)
				} else {
					title += " · state unreported"
				}
				if state != "used" {
					priority = 1
				}
			}
			if family == "network" {
				title = "Network " + s.Attributes["network.io.direction"] + " · " + strconv.Quote(s.Attributes["network.interface.name"])
			}
			if family == "disk" {
				title = "Disk " + s.Attributes["disk.io.direction"] + " · " + strconv.Quote(s.Attributes["system.device"])
			}
			if family == "filesystem" {
				role := "Used bytes"
				if i.Metric == "system.filesystem.limit" {
					role = "Reported capacity"
				}
				title = role + " · " + strconv.Quote(s.Attributes["system.filesystem.mountpoint"]) + " · " + strconv.Quote(s.Attributes["system.device"])
			}
			candidates = append(candidates, candidate{widget: HostWidget{Key: key, Title: title, State: "candidate", Grain: "host-exact-stream", Question: q}, family: family, priority: priority, numeric: numeric, core: core})
		}
	}
	sort.Slice(candidates, func(a, b int) bool {
		x, y := candidates[a], candidates[b]
		rank := func(f string) int {
			switch f {
			case "memory":
				return 0
			case "cpu":
				return 1
			case "cpu-time":
				return 2
			default:
				return 3
			}
		}
		if rank(x.family) != rank(y.family) {
			return rank(x.family) < rank(y.family)
		}
		if x.priority != y.priority {
			return x.priority < y.priority
		}
		if (x.family == "cpu" || x.family == "cpu-time") && x.numeric != y.numeric {
			return x.numeric
		}
		if x.numeric && y.numeric && x.core != y.core {
			return x.core < y.core
		}
		if (x.family == "cpu" || x.family == "cpu-time") && !x.numeric && !y.numeric {
			a := x.widget.Question.Definition.Source.ExactStream.Attributes["cpu.logical_number"]
			b := y.widget.Question.Definition.Source.ExactStream.Attributes["cpu.logical_number"]
			if a != b {
				return a < b
			}
		}
		return x.widget.Key < y.widget.Key
	})
	// Defaults are selected before the display cap so each available family remains
	// selectable; retain defaults first, then the remaining deterministic options.
	defaults := []candidate{}
	haveCPU, haveMemory := false, false
	for _, c := range candidates {
		if c.family == "memory" && !haveMemory {
			defaults = append(defaults, c)
			haveMemory = true
		}
		if (c.family == "cpu" || c.family == "cpu-time") && !haveCPU {
			defaults = append(defaults, c)
			haveCPU = true
		}
	}
	if !haveMemory {
		for _, c := range candidates {
			if c.family == "memory-fraction" {
				defaults = append(defaults, c)
				break
			}
		}
	}
	ordered := []candidate{}
	chosen := map[string]bool{}
	for _, c := range defaults {
		out.Widgets = append(out.Widgets, c.widget)
		ordered = append(ordered, c)
		chosen[c.widget.Key] = true
	}
	for _, c := range candidates {
		if !chosen[c.widget.Key] {
			ordered = append(ordered, c)
		}
	}
	for index, c := range ordered {
		if index >= HostCandidateLimit {
			out.CandidatesOmitted++
			continue
		}
		out.Candidates = append(out.Candidates, c.widget)
	}
	if len(out.Candidates) > 0 {
		out.State = "candidates"
	}
	return out, nil
}
