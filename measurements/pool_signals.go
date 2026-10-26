package measurements

import (
	"encoding/json"
	"fmt"
	"maps"
	"sort"

	"github.com/casselc/goo/metrics"
)

const PoolSignalsVersion = "goo/pool-signals/v1"
const PoolSignalsInstrumentLimit = 16
const PoolSignalsOptionLimit = 2

type PoolSignalsRequest struct {
	Parent      Question       `json:"parent"`
	Stream      metrics.Stream `json:"stream"`
	Instruments []Instrument   `json:"instruments"`
}
type PoolSignalOption struct {
	Key      string   `json:"key"`
	Title    string   `json:"title"`
	Question Question `json:"question"`
}
type PoolSignalsPlan struct {
	Version         string             `json:"version"`
	Parent          Question           `json:"parent"`
	Stream          metrics.Stream     `json:"stream"`
	Capacity        []PoolSignalOption `json:"capacity"`
	Wait            []PoolSignalOption `json:"wait"`
	State           string             `json:"state"`
	CapacityOmitted int                `json:"capacityOmitted"`
	WaitOmitted     int                `json:"waitOmitted"`
	StreamsOmitted  int                `json:"streamsOmitted"`
	Warnings        []string           `json:"warnings"`
}

// PlanPoolSignals stages independent exact-stream measurements. Preview absence
// is not source absence, and matching labels do not certify capacity alignment.
func PlanPoolSignals(in PoolSignalsRequest) (*PoolSignalsPlan, error) {
	if len(in.Instruments) > PoolSignalsInstrumentLimit {
		return nil, fmt.Errorf("pool signals accepts at most 16 candidate instruments")
	}
	if e := in.Parent.Validate(); e != nil {
		return nil, e
	}
	d := in.Parent.Definition
	if d.Kind != "stock" || d.StockDomain != "nonnegative-integer" || d.Source.Metric != "db.client.connection.count" || d.Source.Unit != "{connection}" || d.Source.SchemaURL != PoolConnectionsSchema {
		return nil, fmt.Errorf("pool signals requires a certified pinned connection count definition")
	}
	if in.Stream.Attributes["db.client.connection.pool.name"] == "" || in.Stream.Attributes["db.client.connection.state"] == "" || in.Stream.Resource["service.instance.id"] == "" {
		return nil, fmt.Errorf("pool signals requires a reported instance, pool name and state")
	}
	if d.Source.ExactStream != nil && !sameExactStream(*d.Source.ExactStream, in.Stream) {
		return nil, fmt.Errorf("pool signals stream differs from pinned parent stream")
	}
	bound := d
	bound.Source.ExactStream = &in.Stream
	if e := bound.Validate(); e != nil {
		return nil, e
	}
	// Detach all retained inputs before deriving candidate definitions.
	encoded, e := json.Marshal(in)
	if e != nil {
		return nil, e
	}
	var req PoolSignalsRequest
	if e = json.Unmarshal(encoded, &req); e != nil {
		return nil, e
	}
	result := &PoolSignalsPlan{Version: PoolSignalsVersion, Parent: req.Parent, Stream: req.Stream, Capacity: []PoolSignalOption{}, Wait: []PoolSignalOption{}, State: "unconfirmed", Warnings: []string{"Candidates are supplied discovery metadata, not fresh source evidence. Missing candidates or omitted preview streams do not prove absence.", "Capacity and wait are independent measurements with separate observation clocks and populations. No utilization ratio, queue health, capacity alignment or causal relation is inferred.", "Only exact shared resource and instrumentation identity and pool labels match. Candidate attributes must equal the count stream attributes with only connection state removed; no alias or dimension broadening is inferred."}}
	expectedAttrs := maps.Clone(req.Stream.Attributes)
	delete(expectedAttrs, "db.client.connection.state")
	type candidate struct {
		identity, encoded, kind string
		option                  PoolSignalOption
	}
	candidates := []candidate{}
	for _, instrument := range req.Instruments {
		if instrument.CapabilityPlan == nil {
			continue
		}
		coverage := instrument.CapabilityPlan.Coverage
		preview := coverage.Streams
		omitted := coverage.OmittedStreams
		// Discovery is bounded by RowLimit. Refuse impossible supplied counters
		// before summing them; request byte limits do not prevent int overflow.
		if omitted < 0 || omitted > RowLimit || len(preview) > RowLimit {
			return nil, fmt.Errorf("pool signal preview counts must be within 0..10000")
		}
		if len(preview) > DiscoveryStreamLimit {
			extra := len(preview) - DiscoveryStreamLimit
			if extra > RowLimit-omitted {
				return nil, fmt.Errorf("pool signal preview counts must be within 0..10000")
			}
			omitted += extra
			preview = preview[:DiscoveryStreamLimit]
		}
		result.StreamsOmitted += omitted
		if !instrument.Supported || instrument.Definition == nil || instrument.SchemaURL != PoolConnectionsSchema {
			continue
		}
		kind := ""
		if instrument.Kind == "stock" && instrument.Metric == "db.client.connection.max" && instrument.Unit == "{connection}" && instrument.Temporality == 2 && !instrument.Monotonic {
			kind = "capacity"
		}
		if instrument.Kind == "histogram" && instrument.Metric == "db.client.connection.wait_time" && instrument.Unit == "s" && (instrument.Temporality == 1 || instrument.Temporality == 2) {
			kind = "wait"
		}
		if kind == "" {
			continue
		}
		def := *instrument.Definition
		if def.Validate() != nil || def.Kind != instrument.Kind || def.Source.Metric != instrument.Metric || def.Source.Unit != instrument.Unit || def.Source.Scope != instrument.Scope || def.Source.Version != instrument.Version || def.Source.SchemaURL != instrument.SchemaURL {
			continue
		}
		for _, observed := range preview {
			stream := observed.Stream
			if !maps.Equal(stream.Resource, req.Stream.Resource) || stream.ResourceSchemaURL != req.Stream.ResourceSchemaURL || stream.Scope != req.Stream.Scope || stream.Version != req.Stream.Version || stream.SchemaURL != req.Stream.SchemaURL || !maps.Equal(stream.Attributes, expectedAttrs) {
				continue
			}
			if def.Source.ExactStream != nil && !sameExactStream(*def.Source.ExactStream, stream) {
				continue
			}
			selected := def
			selected.Source.ExactStream = &stream
			if kind == "capacity" {
				selected.StockDomain = "nonnegative-integer"
			}
			q := req.Parent
			q.Definition = selected
			if q.Validate() != nil {
				continue
			}
			// Identity dedup deliberately excludes user-facing package names/titles.
			identity := Hash(struct {
				Kind   string
				Stream metrics.Stream
			}{kind, stream})
			bytes, _ := json.Marshal(q)
			title := "Pool capacity · independent cumulative stock"
			if kind == "wait" {
				title = "Pool wait · independent duration histogram"
			}
			candidates = append(candidates, candidate{identity: identity, encoded: string(bytes), kind: kind, option: PoolSignalOption{Key: Hash(q), Title: title, Question: q}})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].kind != candidates[j].kind {
			return candidates[i].kind < candidates[j].kind
		}
		if candidates[i].identity != candidates[j].identity {
			return candidates[i].identity < candidates[j].identity
		}
		return candidates[i].encoded < candidates[j].encoded
	})
	seen := map[string]bool{}
	for _, c := range candidates {
		key := c.kind + ":" + c.identity
		if seen[key] {
			continue
		}
		seen[key] = true
		if c.kind == "capacity" {
			if len(result.Capacity) < PoolSignalsOptionLimit {
				result.Capacity = append(result.Capacity, c.option)
			} else {
				result.CapacityOmitted++
			}
		} else {
			if len(result.Wait) < PoolSignalsOptionLimit {
				result.Wait = append(result.Wait, c.option)
			} else {
				result.WaitOmitted++
			}
		}
	}
	if len(result.Capacity)+len(result.Wait) > 0 {
		result.State = "candidates"
	}
	return result, nil
}
