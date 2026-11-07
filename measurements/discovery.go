package measurements

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/identity"
	"sort"
	"strconv"
	"strings"
)

// DiscoveryQuestion is a bounded inventory of observations, never a completeness claim.
type DiscoveryQuestion struct {
	Namespace           *identity.Namespace           `json:"namespace,omitempty"`
	Environment         string                        `json:"environment,omitempty"`
	EnvironmentSelector *identity.EnvironmentSelector `json:"environmentSelector,omitempty"`
	Service             string                        `json:"service"`
	FromMS              int64                         `json:"fromMs"`
	ToMS                int64                         `json:"toMs"`
}
type Instrument struct {
	UnsupportedReason string                    `json:"unsupportedReason,omitempty"`
	CapabilityPlan    *InstrumentCapabilityPlan `json:"capabilityPlan,omitempty"`
	Kind              string                    `json:"kind"`
	Scope             string                    `json:"scope"`
	Version           string                    `json:"version"`
	SchemaURL         string                    `json:"schemaURL"`
	Metric            string                    `json:"metric"`
	Unit              string                    `json:"unit"`
	Temporality       int                       `json:"temporality"`
	Monotonic         bool                      `json:"monotonic"`
	Dimensions        []string                  `json:"dimensions"`
	Records           int                       `json:"records"`
	Supported         bool                      `json:"supported"`
	Definition        *Definition               `json:"definition,omitempty"`
}
type DiscoveryDiagnostic struct {
	Kind   string `json:"kind"`
	Reason string `json:"reason"`
}

const exponentialHistogramKind = "exponential-histogram"
const exponentialHistogramUnsupportedReason = "Exponential histogram evaluation is unavailable: zero-threshold and optional-statistic fidelity and exponential bucket geometry are unsupported by the current source adapter and kernel."

// Diagnostics describe checked source availability, never instrument absence.
type DiscoveryResult struct {
	Question    DiscoveryQuestion     `json:"question"`
	Diagnostics []DiscoveryDiagnostic `json:"diagnostics"`
	Instruments []Instrument          `json:"instruments"`
	Records     int                   `json:"records"`
	Warning     string                `json:"warning"`
}

func (q DiscoveryQuestion) Validate() error {
	d := Definition{APIVersion: Version, Name: "discovered", Title: "Discovered metric", Kind: "gauge", Environment: q.Environment, EnvironmentSelector: copyEnvironmentSelector(q.EnvironmentSelector), LookbackMinutes: 2, MaxGapMS: 60000, Grain: "per-stream", Source: Binding{Namespace: q.Namespace.Clone(), Services: []string{q.Service}, Scope: "discovery", Metric: "discovery", Unit: "1"}}
	return (Question{Definition: d, FromMS: q.FromMS, ToMS: q.ToMS}).Validate()
}
func (s Source) Discover(ctx context.Context, q DiscoveryQuestion) (*DiscoveryResult, error) {
	q = q.clone()
	if e := q.Validate(); e != nil {
		return nil, e
	}
	optionalPresent, err := s.exponentialHistogramTablePresent(ctx)
	if err != nil {
		return nil, err
	}
	diagnostics := []DiscoveryDiagnostic{}
	kinds := []string{"gauge", "counter", "histogram"}
	if optionalPresent {
		kinds = append(kinds, exponentialHistogramKind)
	} else {
		diagnostics = append(diagnostics, DiscoveryDiagnostic{Kind: exponentialHistogramKind, Reason: "optional-table-missing"})
	}
	parts := []string{}
	namespaceSQL := ""
	params := map[string]string{"from": strconv.FormatInt(q.FromMS, 10), "to": strconv.FormatInt(q.ToMS, 10), "service": q.Service}
	environmentSQL := "ResourceAttributes['deployment.environment.name']={environment:String}"
	if q.EnvironmentSelector == nil {
		params["environment"] = q.Environment
	} else {
		selection, _ := q.EnvironmentSelection()
		var e error
		environmentSQL, e = metricEnvironmentPredicate(selection, func(value, _ string) string { params["environment"] = value; return "{environment:String}" })
		if e != nil {
			return nil, e
		}
	}
	if q.Namespace != nil {
		namespaceSQL = " AND NOT mapContains(ResourceAttributes,'service.namespace')"
		if q.Namespace.Mode == "value" {
			namespaceSQL = " AND mapContains(ResourceAttributes,'service.namespace') AND ResourceAttributes['service.namespace']={namespace:String}"
			params["namespace"] = q.Namespace.Value
		}
	}
	for _, kind := range kinds {
		table := map[string]string{"gauge": "gauge", "counter": "sum", "histogram": "histogram", exponentialHistogramKind: "exp_histogram"}[kind]
		tempo, mono := "0", "false"
		if kind != "gauge" {
			tempo = "AggregationTemporality"
		}
		if kind == "counter" {
			mono = "IsMonotonic"
		}
		parts = append(parts, fmt.Sprintf("SELECT '%s' AS kind, ScopeName AS scope, ScopeVersion AS version, ScopeSchemaUrl AS schemaURL, MetricName AS metric, MetricUnit AS unit, %s AS temporality, %s AS monotonic, mapKeys(Attributes) AS dimensions, mapUpdate(ResourceAttributes,map('service.name',toString(ServiceName))) AS resource, Attributes AS attributes, ResourceSchemaUrl AS resourceSchemaURL FROM otel_metrics_%s WHERE TimeUnix>=fromUnixTimestamp64Milli({from:Int64}) AND TimeUnix<fromUnixTimestamp64Milli({to:Int64}) AND ServiceName={service:String} AND %s", kind, tempo, mono, table, environmentSQL)+namespaceSQL)
	}
	// A single global limit bounds all families. Exceeding it fails without publishing a partial catalog.
	sql := "SELECT * FROM (" + strings.Join(parts, " UNION ALL ") + ") LIMIT 10001 FORMAT JSONEachRow"
	rows, e := s.read(ctx, Plan{SQL: sql, Parameters: params})
	if e != nil {
		return nil, e
	}
	result, err := discoverRows(q, rows)
	if err != nil {
		return nil, err
	}
	result.Diagnostics = diagnostics
	return result, nil
}

// Probe only the fixed optional exporter table. A source failure or malformed
// answer is not table absence; a later table race still fails the global read.
func (s Source) exponentialHistogramTablePresent(ctx context.Context) (bool, error) {
	rows, err := s.read(ctx, Plan{SQL: "EXISTS TABLE otel_metrics_exp_histogram FORMAT JSONEachRow"})
	if err != nil {
		return false, err
	}
	if len(rows) != 1 {
		return false, fmt.Errorf("invalid optional metric table availability response")
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(rows[0], &fields) != nil || len(fields) != 1 {
		return false, fmt.Errorf("invalid optional metric table availability response")
	}
	switch strings.TrimSpace(string(fields["result"])) {
	case "0":
		return false, nil
	case "1":
		return true, nil
	default:
		return false, fmt.Errorf("invalid optional metric table availability response")
	}
}
func discoverRows(q DiscoveryQuestion, rows []json.RawMessage) (*DiscoveryResult, error) {
	if len(rows) > RowLimit {
		return nil, fmt.Errorf("source exceeds 10000 rows; no partial discovery")
	}
	q = q.clone()
	out := &DiscoveryResult{Question: q, Diagnostics: []DiscoveryDiagnostic{}, Instruments: []Instrument{}, Records: len(rows), Warning: "Observed metadata only; not a complete instrument or dimension inventory. No fleet aggregation is implied. Empty version/schema values are exact matches. Records count exported data points, not represented histogram observations or health. Counter and explicit histogram support is restricted by the measurement kernels; exponential histograms remain observed metadata only."}
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
		if q.EnvironmentSelector != nil {
			selection, e := q.EnvironmentSelection()
			if e != nil || !selection.Matches(raw.DiscoveryObservation.Resource) || !q.Namespace.Matches(raw.DiscoveryObservation.Resource) || raw.DiscoveryObservation.Resource["service.name"] != q.Service {
				return nil, fmt.Errorf("discovered metric resource differs from typed environment/service scope")
			}
		}
		i := raw.Instrument
		if i.Kind == "counter" && !i.Monotonic && i.Temporality == 2 {
			i.Kind = "stock"
		}
		i.CapabilityPlan = nil
		i.Definition = nil
		i.UnsupportedReason = ""
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
		if i.Kind == exponentialHistogramKind {
			i.Supported = false
			i.UnsupportedReason = exponentialHistogramUnsupportedReason
		} else if q.EnvironmentSelector != nil && !typedServiceKind(i.Kind) {
			i.Supported = false
			i.UnsupportedReason = typedServiceNativeReason
		}
		if i.Supported {
			d := Definition{APIVersion: Version, Name: "metric-" + k[:12], Title: i.Metric, Kind: i.Kind, Environment: q.Environment, EnvironmentSelector: copyEnvironmentSelector(q.EnvironmentSelector), LookbackMinutes: 2, MaxGapMS: 60000, Grain: "per-stream", Source: Binding{Namespace: q.Namespace.Clone(), Services: []string{q.Service}, Scope: i.Scope, Version: i.Version, SchemaURL: i.SchemaURL, Metric: i.Metric, Unit: i.Unit}}
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
