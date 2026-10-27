package measurements

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/metrics"
	"maps"
	"sort"
	"strconv"
	"unicode/utf8"
)

const PodMetricDiscoveryVersion = "goo/pod-metric-discovery/v1"
const PodDashboardVersion = "goo/pod-dashboard/v1"
const PodMetricInstrumentLimit = 16
const PodMetricCandidateLimit = 16
const KubeletStatsScope = "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver"
const KubeletStatsVersion = "0.137.0"
const PodMetricsSchema = "https://opentelemetry.io/schemas/1.37.0"

type PodMetricDiscoveryQuestion struct {
	Environment       string            `json:"environment"`
	FromMS            int64             `json:"fromMs"`
	ToMS              int64             `json:"toMs"`
	Resource          map[string]string `json:"resource"`
	ResourceSchemaURL *string           `json:"resourceSchemaURL"`
}
type PodMetricDiscovery struct {
	Version            string                     `json:"version"`
	Question           PodMetricDiscoveryQuestion `json:"question"`
	Key                string                     `json:"key"`
	Instruments        []Instrument               `json:"instruments"`
	Records            int                        `json:"records"`
	InstrumentsOmitted int                        `json:"instrumentsOmitted"`
	Warning            string                     `json:"warning"`
}

func (q PodMetricDiscoveryQuestion) Validate() error {
	if e := (HostDiscoveryQuestion{Environment: q.Environment, FromMS: q.FromMS, ToMS: q.ToMS}).Validate(); e != nil {
		return fmt.Errorf("pod metric discovery requires explicit environment and fixed half-open window <=120 minutes before 2100")
	}
	if q.ResourceSchemaURL == nil {
		return fmt.Errorf("pod metric discovery requires a known reported resource schema; trace schema is unavailable")
	}
	if e := validateHostLabels(q.Resource); e != nil {
		return e
	}
	if q.Resource["k8s.pod.uid"] == "" || !text(q.Resource["k8s.pod.uid"], 512) || q.Resource["deployment.environment.name"] != q.Environment || !text(*q.ResourceSchemaURL, 1024) {
		return fmt.Errorf("pod metric discovery requires an exact pod UID and resource identity in the question environment")
	}
	encoded, _ := json.Marshal(struct {
		Resource map[string]string
		Schema   string
	}{q.Resource, *q.ResourceSchemaURL})
	if len(encoded) > 16384 {
		return fmt.Errorf("pod resource identity exceeds 16 KiB")
	}
	return nil
}
func (s Source) DiscoverPodMetrics(ctx context.Context, q PodMetricDiscoveryQuestion) (*PodMetricDiscovery, error) {
	if e := q.Validate(); e != nil {
		return nil, e
	}
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	resource, _ := json.Marshal(q.Resource)
	params := map[string]string{"from": strconv.FormatInt(q.FromMS, 10), "to": strconv.FormatInt(q.ToMS, 10), "environment": q.Environment, "pod": q.Resource["k8s.pod.uid"], "resource": string(resource), "resourceSchema": *q.ResourceSchemaURL}
	parts := []string{}
	for _, kind := range []string{"gauge", "counter", "histogram"} {
		table := map[string]string{"gauge": "gauge", "counter": "sum", "histogram": "histogram"}[kind]
		tempo, mono := "0", "false"
		if kind != "gauge" {
			tempo = "AggregationTemporality"
		}
		if kind == "counter" {
			mono = "IsMonotonic"
		}
		parts = append(parts, "SELECT '"+kind+"' AS kind, ScopeName AS scope, ScopeVersion AS version, ScopeSchemaUrl AS schemaURL, MetricName AS metric, MetricUnit AS unit, "+tempo+" AS temporality, "+mono+" AS monotonic, ResourceAttributes AS resource, Attributes AS attributes, ResourceSchemaUrl AS resourceSchemaURL FROM otel_metrics_"+table+" WHERE TimeUnix>=fromUnixTimestamp64Milli({from:Int64}) AND TimeUnix<fromUnixTimestamp64Milli({to:Int64}) AND ResourceAttributes['deployment.environment.name']={environment:String} AND ResourceAttributes['k8s.pod.uid']={pod:String} AND ResourceSchemaUrl={resourceSchema:String} AND arraySort(arrayZip(mapKeys(ResourceAttributes),mapValues(ResourceAttributes)))=arraySort(JSONExtractKeysAndValues({resource:String},'String'))")
	}
	rows, e := s.read(ctx, Plan{SQL: "SELECT * FROM (" + parts[0] + " UNION ALL " + parts[1] + " UNION ALL " + parts[2] + ") LIMIT 10001 FORMAT JSONEachRow", Parameters: params})
	if e != nil {
		return nil, e
	}
	if e = ctx.Err(); e != nil {
		return nil, e
	}
	return discoverPodMetricRows(q, rows)
}
func discoverPodMetricRows(q PodMetricDiscoveryQuestion, rows []json.RawMessage) (*PodMetricDiscovery, error) {
	if e := q.Validate(); e != nil {
		return nil, e
	}
	if len(rows) > RowLimit {
		return nil, fmt.Errorf("pod metric discovery exceeds 10000 observations; no partial inventory")
	}
	encoded, _ := json.Marshal(q)
	var detached PodMetricDiscoveryQuestion
	if e := json.Unmarshal(encoded, &detached); e != nil {
		return nil, e
	}
	q = detached
	out := &PodMetricDiscovery{Version: PodMetricDiscoveryVersion, Question: q, Key: hostResourceKey(q.Resource, *q.ResourceSchemaURL), Records: len(rows), Instruments: []Instrument{}, Warning: "Reported metadata for one exact raw resource and known resource schema. No service identity, complete pod inventory, phase, readiness, health, capacity or container aggregation is inferred. Receiver-specific names and semconv profiles remain distinct."}
	groups := map[string]*Instrument{}
	observations := map[string][]DiscoveryObservation{}
	dimensions := map[string]map[string]bool{}
	for _, raw := range rows {
		var values map[string]json.RawMessage
		if !utf8.Valid(raw) || json.Unmarshal(raw, &values) != nil {
			return nil, fmt.Errorf("invalid pod metric metadata")
		}
		var row struct {
			Instrument
			DiscoveryObservation
		}
		if json.Unmarshal(raw, &row) != nil {
			return nil, fmt.Errorf("invalid pod metric row")
		}
		for _, field := range []string{"kind", "scope", "version", "schemaURL", "metric", "unit", "temporality", "monotonic", "resource", "attributes", "resourceSchemaURL"} {
			if _, ok := values[field]; !ok {
				return nil, fmt.Errorf("pod metric metadata missing %s", field)
			}
		}
		for _, field := range []string{"kind", "scope", "version", "schemaURL", "metric", "unit", "resourceSchemaURL"} {
			value := bytes.TrimSpace(values[field])
			if len(value) == 0 || value[0] != '"' {
				return nil, fmt.Errorf("pod metric metadata %s must be a reported string", field)
			}
		}
		for _, field := range []string{"temporality", "monotonic"} {
			if bytes.Equal(bytes.TrimSpace(values[field]), []byte("null")) {
				return nil, fmt.Errorf("pod metric metadata %s must be reported", field)
			}
		}
		for _, name := range []string{"resource", "attributes"} {
			var labels map[string]json.RawMessage
			if json.Unmarshal(values[name], &labels) != nil || labels == nil {
				return nil, fmt.Errorf("invalid pod metric labels")
			}
			for _, v := range labels {
				if len(v) == 0 || v[0] != '"' {
					return nil, fmt.Errorf("pod metric labels must be strings")
				}
			}
		}
		if e := validateHostLabels(row.Resource); e != nil {
			return nil, e
		}
		if e := validateHostLabels(row.Attributes); e != nil {
			return nil, e
		}
		if !maps.Equal(row.Resource, q.Resource) || row.ResourceSchemaURL != *q.ResourceSchemaURL {
			return nil, fmt.Errorf("pod metric source returned a different resource identity")
		}
		i := row.Instrument
		if i.Kind != "gauge" && i.Kind != "counter" && i.Kind != "histogram" {
			return nil, fmt.Errorf("invalid pod metric family")
		}
		if !text(i.Scope, 256) || !text(i.Version, 128) || !text(i.SchemaURL, 1024) || !text(i.Metric, 256) || !text(i.Unit, 64) {
			return nil, fmt.Errorf("invalid pod metric metadata labels")
		}
		if i.Kind == "counter" && i.Temporality == 2 && !i.Monotonic {
			i.Kind = "stock"
		}
		i.Supported = false
		i.Definition = nil
		i.CapabilityPlan = nil
		i.Dimensions = nil
		i.Records = 0
		key := Hash(i)
		if groups[key] == nil {
			copy := i
			groups[key] = &copy
			dimensions[key] = map[string]bool{}
		}
		groups[key].Records++
		observations[key] = append(observations[key], DiscoveryObservation{Resource: row.Resource, Attributes: row.Attributes, ResourceSchemaURL: row.ResourceSchemaURL})
		for name := range row.Attributes {
			dimensions[key][name] = true
		}
	}
	keys := []string{}
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for index, key := range keys {
		if index >= PodMetricInstrumentLimit {
			out.InstrumentsOmitted++
			continue
		}
		i := groups[key]
		for name := range dimensions[key] {
			i.Dimensions = append(i.Dimensions, name)
		}
		sort.Strings(i.Dimensions)
		coverage := observedCoverage(*i, observations[key])
		clusterState := podMetricProfile(*i) == PodClusterProfile
		ready := clusterState && i.Metric == "k8s.container.ready"
		i.Supported = clusterState || i.Kind == "gauge" && i.Scope != "" && i.Metric != "" && i.Unit != "" && i.Temporality == 0
		if i.Supported && len(coverage.Streams) > 0 {
			stream := coverage.Streams[0].Stream
			def := Definition{APIVersion: Version, Name: "pod-metric-" + key[:12], Title: i.Metric, Kind: "gauge", Environment: q.Environment, LookbackMinutes: 2, MaxGapMS: 60000, Grain: "per-stream", Source: Binding{Resource: &ResourceBinding{Kind: "kubernetes-pod", ID: q.Resource["k8s.pod.uid"]}, Services: []string{}, Scope: i.Scope, Version: i.Version, SchemaURL: i.SchemaURL, Metric: i.Metric, Unit: i.Unit, ExactStream: &stream}}
			if clusterState {
				def.Kind = "state"
				def.StateProfile = metrics.PodPhaseProfile
				def.Title = "Reported pod phase"
				if ready {
					def.StateProfile = metrics.ContainerReadyProfile
					def.Source.Resource = &ResourceBinding{Kind: "kubernetes-container", ID: q.Resource["container.id"]}
					def.Title = "Reported container readiness"
				}
			}
			if def.Validate() == nil {
				i.Definition = &def
			} else {
				i.Supported = false
			}
		}
		p := InstrumentCapabilityPlan{Version: CapabilityPlanVersion, Profile: "pod-metric", Title: "Reported pod resource metric", Eligible: i.Supported && i.Definition != nil, Verification: "not-run", Reasons: []string{}, Hints: []string{"Pod-bearing resource discovery supports scalar pod gauges and exact pinned k8scluster pod-phase and container-readiness state profiles. Container readiness requires reported container.id, pod UID and container name, and uses container authority. Other unitless metrics and stock/counter/histogram metrics remain unsupported. No aliases, service labels or health state are inferred."}, Coverage: coverage}
		i.CapabilityPlan = &p
		out.Instruments = append(out.Instruments, *i)
	}
	output, _ := json.Marshal(out)
	if len(output) > 8<<20 {
		return nil, fmt.Errorf("pod metric discovery exceeds 8 MiB; narrow the window")
	}
	return out, nil
}
