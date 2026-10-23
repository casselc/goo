package measurements

import (
	"github.com/casselc/goo/metrics"
	"sort"
)

const CapabilityPlanVersion = "goo/instrument-capability-plan/v1alpha1"
const HTTPRuleSet = "OpenTelemetry semantic conventions v1.30.0"
const HTTPSchema = "https://opentelemetry.io/schemas/1.30.0"
const DiscoveryStreamLimit = 25

type DimensionPresence struct {
	Key   string `json:"key"`
	State string `json:"state"`
}
type DimensionCoverage struct {
	Key     string `json:"key"`
	Present int    `json:"present"`
	Empty   int    `json:"empty"`
	Missing int    `json:"missing"`
}
type ObservedStream struct {
	ID         string              `json:"id"`
	Stream     metrics.Stream      `json:"stream"`
	Records    int                 `json:"records"`
	Instance   string              `json:"instance"`
	Dimensions []DimensionPresence `json:"dimensions"`
}
type InstrumentCoverage struct {
	State          string              `json:"state"`
	TotalStreams   int                 `json:"totalStreams"`
	OmittedStreams int                 `json:"omittedStreams"`
	Dimensions     []DimensionCoverage `json:"dimensions"`
	Streams        []ObservedStream    `json:"streams"`
}
type InstrumentCapabilityPlan struct {
	Version      string             `json:"version"`
	Profile      string             `json:"profile"`
	Title        string             `json:"title"`
	RuleSet      string             `json:"ruleSet,omitempty"`
	ShapeMatches bool               `json:"shapeMatches"`
	Eligible     bool               `json:"eligible"`
	Verification string             `json:"verification"`
	Reasons      []string           `json:"reasons"`
	Hints        []string           `json:"hints"`
	Coverage     InstrumentCoverage `json:"coverage"`
}

// DiscoveryObservation retains identity labels, not metric values or interval validity.
type DiscoveryObservation struct {
	Resource          map[string]string `json:"resource"`
	Attributes        map[string]string `json:"attributes"`
	ResourceSchemaURL string            `json:"resourceSchemaURL"`
}

func presence(values map[string]string, key string) string {
	v, ok := values[key]
	if !ok {
		return "missing"
	}
	if v == "" {
		return "empty"
	}
	return "present"
}
func cloneLabels(m map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range m {
		out[k] = v
	}
	return out
}
func observedCoverage(i Instrument, observations []DiscoveryObservation) InstrumentCoverage {
	out := InstrumentCoverage{State: "unavailable", Dimensions: []DimensionCoverage{}, Streams: []ObservedStream{}}
	if len(observations) == 0 {
		return out
	}
	keys := map[string]bool{}
	if i.Metric == "http.server.request.duration" {
		keys["http.request.method"] = true
		keys["http.response.status_code"] = true
	}
	streams := map[string]*ObservedStream{}
	for _, o := range observations {
		if o.Resource == nil || o.Attributes == nil {
			return out
		}
		// The pinned HTTP profile watches two keys; arbitrary attribute cardinality
		// must not multiply this bounded coverage matrix.
		s := metrics.Stream{Resource: cloneLabels(o.Resource), Attributes: cloneLabels(o.Attributes), ResourceSchemaURL: o.ResourceSchemaURL, Scope: i.Scope, Version: i.Version, SchemaURL: i.SchemaURL, Name: i.Metric, Unit: i.Unit}
		id := Hash(s)
		if streams[id] == nil {
			streams[id] = &ObservedStream{ID: id, Stream: s, Instance: presence(s.Resource, "service.instance.id"), Dimensions: []DimensionPresence{}}
		}
		streams[id].Records++
	}
	names := []string{}
	for k := range keys {
		names = append(names, k)
	}
	sort.Strings(names)
	ids := []string{}
	for id := range streams {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out.State = "observed"
	out.TotalStreams = len(ids)
	for _, key := range names {
		d := DimensionCoverage{Key: key}
		for _, id := range ids {
			switch presence(streams[id].Stream.Attributes, key) {
			case "present":
				d.Present++
			case "empty":
				d.Empty++
			case "missing":
				d.Missing++
			}
		}
		out.Dimensions = append(out.Dimensions, d)
	}
	for index, id := range ids {
		if index >= DiscoveryStreamLimit {
			out.OmittedStreams++
			continue
		}
		s := streams[id]
		for _, key := range names {
			s.Dimensions = append(s.Dimensions, DimensionPresence{Key: key, State: presence(s.Stream.Attributes, key)})
		}
		out.Streams = append(out.Streams, *s)
	}
	return out
}

// PlanInstrument is pure metadata planning. It does not verify telemetry semantics.
func PlanInstrument(i Instrument, observations []DiscoveryObservation) InstrumentCapabilityPlan {
	p := InstrumentCapabilityPlan{Version: CapabilityPlanVersion, Profile: "generic", Title: "Generic metric", Eligible: i.Supported && i.Definition != nil, Verification: "not-run", Reasons: []string{}, Hints: []string{}, Coverage: observedCoverage(i, observations)}
	if i.Scope == "" {
		p.Reasons = append(p.Reasons, "Instrumentation scope is omitted.")
	}
	if i.Unit == "" {
		p.Reasons = append(p.Reasons, "Unit is omitted.")
	}
	if (i.Kind == "counter" || i.Kind == "histogram") && i.Temporality != 1 && i.Temporality != 2 {
		p.Reasons = append(p.Reasons, "Aggregation temporality is unsupported.")
	}
	if i.Kind == "counter" && !i.Monotonic {
		p.Reasons = append(p.Reasons, "Non-monotonic sums are not supported by the counter kernel.")
	}
	if i.Kind == "histogram" && i.Unit != "s" && i.Unit != "ms" {
		p.Reasons = append(p.Reasons, "The current histogram kernel requires seconds or milliseconds.")
	}
	if !p.Eligible && len(p.Reasons) == 0 {
		p.Reasons = append(p.Reasons, "No valid package binding is available from the current kernels.")
	}
	if i.Metric != "http.server.request.duration" {
		p.Hints = append(p.Hints, "Generic exact metric binding; no application convention is inferred.")
		return p
	}
	p.Profile = "http-server-duration"
	p.Title = "HTTP server duration candidate"
	p.RuleSet = HTTPRuleSet
	p.ShapeMatches = i.Kind == "histogram" && i.Unit == "s"
	if i.Kind != "histogram" {
		p.Hints = append(p.Hints, "The pinned HTTP duration convention requires a histogram. Use the generic binding.")
	}
	if i.Unit != "s" {
		p.Hints = append(p.Hints, "The pinned HTTP duration convention requires unit s. No automatic conversion is applied.")
	}
	switch i.SchemaURL {
	case HTTPSchema:
		p.Hints = append(p.Hints, "Schema URL matches the pinned rule set; this does not verify conformance.")
	case "":
		p.Hints = append(p.Hints, "Schema URL omitted; convention conformance is unverified.")
	default:
		p.Hints = append(p.Hints, "Other schema URL; no schema migration or conformance is inferred.")
	}
	method, status := false, false
	for _, key := range i.Dimensions {
		method = method || key == "http.request.method"
		status = status || key == "http.response.status_code"
	}
	if method {
		p.Hints = append(p.Hints, "HTTP method key observed; inspect per-stream presence below.")
	} else {
		p.Hints = append(p.Hints, "HTTP method key not observed in this window.")
	}
	if status {
		p.Hints = append(p.Hints, "HTTP status key observed; no error policy or denominator has been verified.")
	} else {
		p.Hints = append(p.Hints, "HTTP status key not observed; no HTTP error-rate view is offered.")
	}
	if i.Kind == "histogram" && p.Eligible && (i.Unit == "s" || i.Unit == "ms") {
		p.Hints = append(p.Hints, "Run can evaluate retained histogram observations and bucket ranges per exact stream. Complete interval coverage is still required; no fleet percentile or request rate is inferred.")
	} else {
		p.Hints = append(p.Hints, "HTTP duration views are not established by this metadata. Retain the generic binding if it is eligible.")
	}
	return p
}
