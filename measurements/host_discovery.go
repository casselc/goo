package measurements

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/metrics"
	"sort"
	"strconv"
)

const HostDiscoveryVersion = "goo/host-discovery/v1"
const HostListLimit = 100
const HostDiscoveryByteLimit = 8 << 20
const HostInstrumentLimit = 16

type HostDiscoveryQuestion struct {
	Environment string `json:"environment"`
	FromMS      int64  `json:"fromMs"`
	ToMS        int64  `json:"toMs"`
}
type HostEntry struct {
	Key                string            `json:"key"`
	Resource           map[string]string `json:"resource"`
	ResourceSchemaURL  string            `json:"resourceSchemaURL"`
	Instruments        []Instrument      `json:"instruments"`
	Records            int               `json:"records"`
	InstrumentsOmitted int               `json:"instrumentsOmitted"`
}
type HostDiscovery struct {
	Version              string                `json:"version"`
	Question             HostDiscoveryQuestion `json:"question"`
	Hosts                []HostEntry           `json:"hosts"`
	Records              int                   `json:"records"`
	MissingHostIDRecords int                   `json:"missingHostIDRecords"`
	HostsOmitted         int                   `json:"hostsOmitted"`
	Warning              string                `json:"warning"`
}

func (q HostDiscoveryQuestion) Validate() error {
	if q.Environment == "" || !text(q.Environment, 128) || q.FromMS < 0 || q.ToMS <= q.FromMS || q.ToMS-q.FromMS > 7200000 || q.ToMS >= 4102444800000 {
		return fmt.Errorf("host discovery requires an explicit environment and fixed half-open window <=120 minutes before 2100")
	}
	return nil
}
func hostResourceKey(resource map[string]string, schema string) string {
	return Hash(struct {
		Resource          map[string]string
		ResourceSchemaURL string
	}{resource, schema})
}
func (s Source) DiscoverHosts(ctx context.Context, q HostDiscoveryQuestion) (*HostDiscovery, error) {
	if e := q.Validate(); e != nil {
		return nil, e
	}
	if e := ctx.Err(); e != nil {
		return nil, e
	}
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
		parts = append(parts, fmt.Sprintf("SELECT '%s' AS kind, ScopeName AS scope, ScopeVersion AS version, ScopeSchemaUrl AS schemaURL, MetricName AS metric, MetricUnit AS unit, %s AS temporality, %s AS monotonic, ResourceAttributes AS resource, Attributes AS attributes, ResourceSchemaUrl AS resourceSchemaURL FROM otel_metrics_%s WHERE TimeUnix>=fromUnixTimestamp64Milli({from:Int64}) AND TimeUnix<fromUnixTimestamp64Milli({to:Int64}) AND ResourceAttributes['deployment.environment.name']={environment:String} AND mapContains(ResourceAttributes,'host.id') AND ResourceAttributes['host.id']!=''", kind, tempo, mono, table))
	}
	rows, e := s.read(ctx, Plan{SQL: "SELECT * FROM (" + parts[0] + " UNION ALL " + parts[1] + " UNION ALL " + parts[2] + ") LIMIT 10001 FORMAT JSONEachRow", Parameters: map[string]string{"from": strconv.FormatInt(q.FromMS, 10), "to": strconv.FormatInt(q.ToMS, 10), "environment": q.Environment}})
	if e != nil {
		return nil, e
	}
	if e = ctx.Err(); e != nil {
		return nil, e
	}
	return discoverHostRows(q, rows)
}
func validateHostLabels(values map[string]string) error {
	if len(values) > 64 {
		return fmt.Errorf("host discovery identity exceeds 64 labels")
	}
	for k, v := range values {
		if k == "" || !text(k, 256) || !text(v, 2048) {
			return fmt.Errorf("invalid host discovery identity label")
		}
	}
	return nil
}
func discoverHostRows(q HostDiscoveryQuestion, rows []json.RawMessage) (*HostDiscovery, error) {
	if e := q.Validate(); e != nil {
		return nil, e
	}
	if len(rows) > RowLimit {
		return nil, fmt.Errorf("host discovery exceeds 10000 records; narrow the window")
	}
	out := &HostDiscovery{Version: HostDiscoveryVersion, Question: q, Hosts: []HostEntry{}, Records: len(rows), Warning: "Observed metadata only, not host health or a complete inventory. Raw resource maps and schemas remain separate identities; equal reported host.id values do not prove the same host. Source inventory queries only host-bearing rows; unidentified rows are not counted, so missingHostIDRecords is only a decoder observation and not a source absence count. Host name is display only; no service identity or relationship is manufactured."}
	type group struct {
		entry        HostEntry
		instruments  map[string]*Instrument
		observations map[string][]DiscoveryObservation
		dimensions   map[string]map[string]bool
	}
	hosts := map[string]*group{}
	for _, row := range rows {
		var raw struct {
			Instrument
			DiscoveryObservation
		}
		if json.Unmarshal(row, &raw) != nil {
			return nil, fmt.Errorf("invalid host discovery row")
		}
		if e := validateHostLabels(raw.Resource); e != nil {
			return nil, e
		}
		if e := validateHostLabels(raw.Attributes); e != nil {
			return nil, e
		}
		if !text(raw.ResourceSchemaURL, 1024) {
			return nil, fmt.Errorf("invalid host resource schema")
		}
		if raw.Resource["deployment.environment.name"] != q.Environment {
			return nil, fmt.Errorf("host discovery row escaped environment")
		}
		encodedResource, _ := json.Marshal(struct {
			Resource map[string]string
			Schema   string
		}{raw.Resource, raw.ResourceSchemaURL})
		if len(encodedResource) > 16384 {
			return nil, fmt.Errorf("host resource identity exceeds 16 KiB")
		}
		id := raw.Resource["host.id"]
		if id == "" {
			out.MissingHostIDRecords++
			continue
		}
		if !text(id, 512) {
			return nil, fmt.Errorf("invalid host.id")
		}
		i := raw.Instrument
		if i.Kind != "gauge" && i.Kind != "counter" && i.Kind != "histogram" {
			return nil, fmt.Errorf("invalid host instrument family")
		}
		if !text(i.Scope, 256) || !text(i.Version, 128) || !text(i.SchemaURL, 1024) || !text(i.Metric, 256) || !text(i.Unit, 64) {
			return nil, fmt.Errorf("invalid host instrument metadata")
		}
		if i.Kind == "counter" && !i.Monotonic && i.Temporality == 2 {
			i.Kind = "stock"
		}
		i.CapabilityPlan = nil
		i.Definition = nil
		i.Dimensions = nil
		i.Records = 0
		i.Supported = false
		hostKey := hostResourceKey(raw.Resource, raw.ResourceSchemaURL)
		g := hosts[hostKey]
		if g == nil {
			g = &group{entry: HostEntry{Key: hostKey, Resource: cloneLabels(raw.Resource), ResourceSchemaURL: raw.ResourceSchemaURL, Instruments: []Instrument{}}, instruments: map[string]*Instrument{}, observations: map[string][]DiscoveryObservation{}, dimensions: map[string]map[string]bool{}}
			hosts[hostKey] = g
		}
		g.entry.Records++
		key := Hash(i)
		if g.instruments[key] == nil {
			copy := i
			g.instruments[key] = &copy
			g.dimensions[key] = map[string]bool{}
		}
		g.instruments[key].Records++
		g.observations[key] = append(g.observations[key], DiscoveryObservation{Resource: cloneLabels(raw.Resource), ResourceSchemaURL: raw.ResourceSchemaURL, Attributes: cloneLabels(raw.Attributes)})
		for k := range raw.Attributes {
			g.dimensions[key][k] = true
		}
	}
	hostKeys := []string{}
	for k := range hosts {
		hostKeys = append(hostKeys, k)
	}
	sort.Strings(hostKeys)
	for index, k := range hostKeys {
		if index >= HostListLimit {
			out.HostsOmitted++
			continue
		}
		g := hosts[k]
		keys := []string{}
		for key := range g.instruments {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for index, key := range keys {
			if index >= HostInstrumentLimit {
				g.entry.InstrumentsOmitted++
				continue
			}
			i := g.instruments[key]
			for d := range g.dimensions[key] {
				i.Dimensions = append(i.Dimensions, d)
			}
			sort.Strings(i.Dimensions)
			i.Supported = i.Scope != "" && i.Metric != "" && i.Unit != "" && (i.Kind == "gauge" || i.Kind == "stock" && i.Temporality == 2 && !i.Monotonic || (HostCPUTimeInstrument(*i) || HostNetworkInstrument(*i)))
			coverage := observedCoverage(*i, g.observations[key])
			if i.Supported && len(coverage.Streams) > 0 {
				stream := coverage.Streams[0].Stream
				d := Definition{APIVersion: Version, Name: "host-metric-" + key[:12], Title: i.Metric, Kind: i.Kind, Environment: q.Environment, LookbackMinutes: 2, MaxGapMS: 60000, Grain: "per-stream", Source: Binding{Resource: &ResourceBinding{Kind: "host", ID: g.entry.Resource["host.id"]}, Services: []string{}, Scope: i.Scope, Version: i.Version, SchemaURL: i.SchemaURL, Metric: i.Metric, Unit: i.Unit, ExactStream: &stream}}
				if HostCPUTimeInstrument(*i) || HostNetworkInstrument(*i) {
					d.CounterDomain = metrics.DecimalCounterDomain
				}
				if d.Validate() == nil {
					i.Definition = &d
				} else {
					i.Supported = false
				}
			}
			plan := InstrumentCapabilityPlan{Version: CapabilityPlanVersion, Profile: "host-metric", Title: "Reported host metric", Eligible: i.Supported && i.Definition != nil, Verification: "not-run", Reasons: []string{}, Hints: []string{"Host resource metadata only; no service identity, health, utilization or complete inventory is inferred. Exact stream identity and values require evaluation."}, Coverage: coverage}
			i.CapabilityPlan = &plan
			g.entry.Instruments = append(g.entry.Instruments, *i)
		}
		out.Hosts = append(out.Hosts, g.entry)
	}
	encoded, _ := json.Marshal(out)
	if len(encoded) > HostDiscoveryByteLimit {
		return nil, fmt.Errorf("host discovery exceeds 8 MiB; narrow the window")
	}
	return out, nil
}
