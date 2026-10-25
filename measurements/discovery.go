package measurements

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/identity"
	"sort"
	"strconv"
)

// DiscoveryQuestion is a bounded inventory of observations, never a completeness claim.
type DiscoveryQuestion struct {
	Namespace   *identity.Namespace `json:"namespace,omitempty"`
	Environment string              `json:"environment"`
	Service     string              `json:"service"`
	FromMS      int64               `json:"fromMs"`
	ToMS        int64               `json:"toMs"`
}
type Instrument struct {
	CapabilityPlan *InstrumentCapabilityPlan `json:"capabilityPlan,omitempty"`
	Kind           string                    `json:"kind"`
	Scope          string                    `json:"scope"`
	Version        string                    `json:"version"`
	SchemaURL      string                    `json:"schemaURL"`
	Metric         string                    `json:"metric"`
	Unit           string                    `json:"unit"`
	Temporality    int                       `json:"temporality"`
	Monotonic      bool                      `json:"monotonic"`
	Dimensions     []string                  `json:"dimensions"`
	Records        int                       `json:"records"`
	Supported      bool                      `json:"supported"`
	Definition     *Definition               `json:"definition,omitempty"`
}
type DiscoveryResult struct {
	Question    DiscoveryQuestion `json:"question"`
	Instruments []Instrument      `json:"instruments"`
	Records     int               `json:"records"`
	Warning     string            `json:"warning"`
}

func (q DiscoveryQuestion) Validate() error {
	d := Definition{APIVersion: Version, Name: "discovered", Title: "Discovered metric", Kind: "gauge", Environment: q.Environment, LookbackMinutes: 2, MaxGapMS: 60000, Grain: "per-stream", Source: Binding{Namespace: q.Namespace.Clone(), Services: []string{q.Service}, Scope: "discovery", Metric: "discovery", Unit: "1"}}
	return (Question{Definition: d, FromMS: q.FromMS, ToMS: q.ToMS}).Validate()
}
func (s Source) Discover(ctx context.Context, q DiscoveryQuestion) (*DiscoveryResult, error) {
	if e := q.Validate(); e != nil {
		return nil, e
	}
	parts := []string{}
	namespaceSQL := ""
	params := map[string]string{"from": strconv.FormatInt(q.FromMS, 10), "to": strconv.FormatInt(q.ToMS, 10), "service": q.Service, "environment": q.Environment}
	if q.Namespace != nil {
		namespaceSQL = " AND NOT mapContains(ResourceAttributes,'service.namespace')"
		if q.Namespace.Mode == "value" {
			namespaceSQL = " AND mapContains(ResourceAttributes,'service.namespace') AND ResourceAttributes['service.namespace']={namespace:String}"
			params["namespace"] = q.Namespace.Value
		}
	}
	for _, kind := range []string{"gauge", "counter", "histogram"} {
		table := map[string]string{"gauge": "gauge", "counter": "sum", "histogram": "histogram"}[kind]
		tempo, mono := "0", "false"
		if kind != "gauge" {
			tempo = "AggregationTemporality"
		}
		if kind == "counter" {
			mono = "IsMonotonic"
		}
		parts = append(parts, fmt.Sprintf("SELECT '%s' AS kind, ScopeName AS scope, ScopeVersion AS version, ScopeSchemaUrl AS schemaURL, MetricName AS metric, MetricUnit AS unit, %s AS temporality, %s AS monotonic, mapKeys(Attributes) AS dimensions, mapUpdate(ResourceAttributes,map('service.name',toString(ServiceName))) AS resource, Attributes AS attributes, ResourceSchemaUrl AS resourceSchemaURL FROM otel_metrics_%s WHERE TimeUnix>=fromUnixTimestamp64Milli({from:Int64}) AND TimeUnix<fromUnixTimestamp64Milli({to:Int64}) AND ServiceName={service:String} AND ResourceAttributes['deployment.environment.name']={environment:String}", kind, tempo, mono, table)+namespaceSQL)
	}
	// A single global limit bounds all families. Exceeding it fails without publishing a partial catalog.
	sql := "SELECT * FROM (" + parts[0] + " UNION ALL " + parts[1] + " UNION ALL " + parts[2] + ") LIMIT 10001 FORMAT JSONEachRow"
	rows, e := s.read(ctx, Plan{SQL: sql, Parameters: params})
	if e != nil {
		return nil, e
	}
	return discoverRows(q, rows)
}
func discoverRows(q DiscoveryQuestion, rows []json.RawMessage) (*DiscoveryResult, error) {
	out := &DiscoveryResult{Question: q, Instruments: []Instrument{}, Records: len(rows), Warning: "Observed metadata only; not a complete instrument or dimension inventory. No fleet aggregation is implied. Empty version/schema values are exact matches. Counter and histogram support is restricted by the measurement kernels."}
	groups := map[string]*Instrument{}
	dims := map[string]map[string]bool{}
	observations := map[string][]DiscoveryObservation{}
	for _, row := range rows {
		var raw struct {
			Instrument
			DiscoveryObservation
		}
		if e := json.Unmarshal(row, &raw); e != nil {
			return nil, e
		}
		i := raw.Instrument
		if i.Kind == "counter" && !i.Monotonic && i.Temporality == 2 {
			i.Kind = "stock"
		}
		i.CapabilityPlan = nil
		keys := i.Dimensions
		i.Dimensions = nil
		i.Records = 0
		key := Hash(i)
		if groups[key] == nil {
			copy := i
			groups[key] = &copy
			dims[key] = map[string]bool{}
		}
		groups[key].Records++
		observations[key] = append(observations[key], raw.DiscoveryObservation)
		for _, k := range keys {
			dims[key][k] = true
		}
	}
	keys := []string{}
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		i := groups[k]
		for d := range dims[k] {
			i.Dimensions = append(i.Dimensions, d)
		}
		sort.Strings(i.Dimensions)
		i.Supported = i.Scope != "" && i.Unit != "" && (i.Kind == "gauge" || i.Kind == "stock" && i.Temporality == 2 && !i.Monotonic || (i.Temporality == 1 || i.Temporality == 2) && (i.Kind == "histogram" && (i.Unit == "s" || i.Unit == "ms") || i.Kind == "counter" && i.Monotonic))
		if i.Supported {
			d := Definition{APIVersion: Version, Name: "metric-" + k[:12], Title: i.Metric, Kind: i.Kind, Environment: q.Environment, LookbackMinutes: 2, MaxGapMS: 60000, Grain: "per-stream", Source: Binding{Namespace: q.Namespace.Clone(), Services: []string{q.Service}, Scope: i.Scope, Version: i.Version, SchemaURL: i.SchemaURL, Metric: i.Metric, Unit: i.Unit}}
			if d.Validate() == nil {
				i.Definition = &d
			} else {
				i.Supported = false
			}
		}
		plan := PlanInstrument(*i, observations[k])
		i.CapabilityPlan = &plan
		out.Instruments = append(out.Instruments, *i)
	}
	return out, nil
}
