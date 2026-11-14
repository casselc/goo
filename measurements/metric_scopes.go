package measurements

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/internal/sourcecap"
	"github.com/casselc/goo/internal/sourcefailure"
)

const MetricScopeLimit = 128

// MetricScopesQuestion discovers observed identities without requiring an
// environment or service to be known in advance. It is not an execution scope.
type MetricScopesQuestion struct {
	FromMS int64 `json:"fromMs"`
	ToMS   int64 `json:"toMs"`
}

func (q MetricScopesQuestion) Validate() error {
	if q.FromMS < 0 || q.ToMS <= q.FromMS || q.ToMS-q.FromMS > 120*60000 || q.ToMS > 4102444800000 {
		return fmt.Errorf("choose a half-open window of at most 120 minutes before 2100")
	}
	return nil
}

// MetricScopeChoice keeps effective environment, namespace presence and service
// separate. Selectable admits instrument discovery, not kernel evaluation.
type MetricScopeChoice struct {
	Service             string                        `json:"service"`
	Namespace           *identity.Namespace           `json:"namespace"`
	Environment         identity.EnvironmentScope     `json:"environment"`
	EnvironmentSelector *identity.EnvironmentSelector `json:"environmentSelector,omitempty"`
	Points              string                        `json:"points"`
	FirstMS             int64                         `json:"firstMs"`
	LastMS              int64                         `json:"lastMs"`
	Selectable          bool                          `json:"selectable"`
	Reason              string                        `json:"reason,omitempty"`
}
type MetricScopesResult struct {
	Version     string                `json:"version"`
	Question    MetricScopesQuestion  `json:"question"`
	CapturedAt  string                `json:"capturedAt"`
	Sources     []sourcecap.Family    `json:"sources"`
	Diagnostics []DiscoveryDiagnostic `json:"diagnostics"`
	Choices     []MetricScopeChoice   `json:"choices"`
	Unit        string                `json:"unit"`
	Warning     string                `json:"warning"`
}

// DiscoverMetricScopes aggregates complete bounded identity groups over usable
// exporter families. An absent/denied family stays explicit, not empty. Metadata
// and data are independent reads; a later source failure refuses the catalog.
func (s Source) DiscoverMetricScopes(ctx context.Context, q MetricScopesQuestion) (result *MetricScopesResult, resultErr error) {
	if err := q.Validate(); err != nil {
		return nil, err
	}
	defer func() { resultErr = sourcefailure.Protocol(resultErr) }()
	sources, err := sourcecap.Inspect(ctx, func(ctx context.Context, sql string, params map[string]string) ([]json.RawMessage, error) {
		return s.read(ctx, Plan{SQL: sql, Parameters: params})
	}, sourcecap.Gauge, sourcecap.Counter, sourcecap.Histogram, sourcecap.ExponentialHistogram)
	if err != nil {
		return nil, err
	}
	kinds := []string{}
	diagnostics := []DiscoveryDiagnostic{}
	for _, family := range sources {
		if family.Compatible() {
			kinds = append(kinds, family.Kind)
		} else {
			diagnostics = append(diagnostics, DiscoveryDiagnostic{Kind: family.Kind, Reason: family.State})
		}
	}
	if len(kinds) == 0 {
		return nil, sourcefailure.Wrap(sourcefailure.Unavailable, fmt.Errorf("no compatible metric scope sources"))
	}
	plan, err := compileMetricScopes(q, kinds)
	if err != nil {
		return nil, err
	}
	rows, err := s.read(ctx, plan)
	if err != nil {
		return nil, err
	}
	choices, err := metricScopeRows(q, rows)
	if err != nil {
		return nil, err
	}
	result = &MetricScopesResult{Version: "goo/metric-scopes/v1", Question: q, CapturedAt: time.Now().UTC().Format(time.RFC3339Nano), Sources: sources, Diagnostics: diagnostics, Choices: choices, Unit: "exported-points", Warning: "Exported points in the searched event-time window across available source families; not metric values, unique streams, represented histogram observations, ingestion completeness or service health. Current environment-key presence wins over legacy, including empty labels. Metadata and point groups are independent source reads."}
	encoded, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	if len(encoded) > 1<<20 {
		return nil, sourcefailure.Wrap(sourcefailure.Budget, fmt.Errorf("metric scope result exceeds 1 MiB"))
	}
	return result, nil
}

func compileMetricScopes(q MetricScopesQuestion, kinds []string) (Plan, error) {
	if err := q.Validate(); err != nil {
		return Plan{}, err
	}
	parts := []string{}
	seen := map[string]bool{}
	for _, kind := range kinds {
		table := map[string]string{sourcecap.Gauge: "otel_metrics_gauge", sourcecap.Counter: "otel_metrics_sum", sourcecap.Histogram: "otel_metrics_histogram", sourcecap.ExponentialHistogram: "otel_metrics_exp_histogram"}[kind]
		if table == "" || seen[kind] {
			return Plan{}, fmt.Errorf("choose distinct supported metric source families")
		}
		seen[kind] = true
		// Presence, not truthiness, determines effective labels. Grouping canonical
		// effective values merges current/legacy combinations that select one scope.
		part := "SELECT toUInt8(mapContains(ResourceAttributes,'" + identity.EnvironmentKey + "') OR mapContains(ResourceAttributes,'" + identity.LegacyEnvironmentKey + "')) AS env_present, if(mapContains(ResourceAttributes,'" + identity.EnvironmentKey + "'),ResourceAttributes['" + identity.EnvironmentKey + "'],ResourceAttributes['" + identity.LegacyEnvironmentKey + "']) AS env_value, toUInt8(mapContains(ResourceAttributes,'" + identity.NamespaceKey + "')) AS ns_present, ResourceAttributes['" + identity.NamespaceKey + "'] AS ns_value, ServiceName AS service, count() AS n, min(toUnixTimestamp64Milli(TimeUnix)) AS first_ms, max(toUnixTimestamp64Milli(TimeUnix)) AS last_ms FROM " + table + " WHERE TimeUnix>=fromUnixTimestamp64Milli({from:Int64}) AND TimeUnix<fromUnixTimestamp64Milli({to:Int64}) GROUP BY env_present,env_value,ns_present,ns_value,service"
		parts = append(parts, part)
	}
	if len(parts) == 0 {
		return Plan{}, fmt.Errorf("choose a usable metric source family")
	}
	sql := "SELECT toString(env_present) AS environment_present, env_value AS environment, toString(ns_present) AS namespace_present, ns_value AS namespace, service, toString(sum(toUInt128(n))) AS points, toString(min(first_ms)) AS first_ms, toString(max(last_ms)) AS last_ms FROM (" + strings.Join(parts, " UNION ALL ") + ") GROUP BY env_present,env_value,ns_present,ns_value,service ORDER BY env_present,env_value,ns_present,ns_value,service LIMIT 129 FORMAT JSONEachRow"
	return Plan{SQL: sql, Parameters: map[string]string{"from": strconv.FormatInt(q.FromMS, 10), "to": strconv.FormatInt(q.ToMS, 10)}}, nil
}
func metricScopeRows(q MetricScopesQuestion, rows []json.RawMessage) ([]MetricScopeChoice, error) {
	if len(rows) > MetricScopeLimit {
		return nil, sourcefailure.Wrap(sourcefailure.Budget, fmt.Errorf("metric scope/service groups exceed 128; narrow the window"))
	}
	choices := make([]MetricScopeChoice, 0, len(rows))
	seen := map[string]bool{}
	size := 0
	for _, raw := range rows {
		size += len(raw)
		if size > 1<<20 || len(raw) > 32<<10 {
			return nil, sourcefailure.Wrap(sourcefailure.Budget, fmt.Errorf("metric scope identities exceed response budget"))
		}
		var row struct {
			EnvironmentPresent *string `json:"environment_present"`
			Environment        *string `json:"environment"`
			NamespacePresent   *string `json:"namespace_present"`
			Namespace          *string `json:"namespace"`
			Service            *string `json:"service"`
			Points             *string `json:"points"`
			First              *string `json:"first_ms"`
			Last               *string `json:"last_ms"`
		}
		if json.Unmarshal(raw, &row) != nil || row.EnvironmentPresent == nil || row.Environment == nil || row.NamespacePresent == nil || row.Namespace == nil || row.Service == nil || row.Points == nil || row.First == nil || row.Last == nil {
			return nil, fmt.Errorf("metric scope source row was rejected")
		}
		for _, value := range []*string{row.Environment, row.Namespace, row.Service} {
			if len(*value) > 4096 || !utf8.ValidString(*value) {
				return nil, fmt.Errorf("metric scope identity exceeds source limits")
			}
		}
		if (*row.EnvironmentPresent != "0" && *row.EnvironmentPresent != "1") || (*row.NamespacePresent != "0" && *row.NamespacePresent != "1") || (*row.EnvironmentPresent == "0" && *row.Environment != "") || (*row.NamespacePresent == "0" && *row.Namespace != "") {
			return nil, fmt.Errorf("metric scope presence was rejected")
		}
		points, ok := new(big.Int).SetString(*row.Points, 10)
		if !ok || points.Sign() <= 0 || points.BitLen() > 128 || points.String() != *row.Points {
			return nil, fmt.Errorf("metric scope point count was rejected")
		}
		first, e := strconv.ParseInt(*row.First, 10, 64)
		if e != nil {
			return nil, fmt.Errorf("metric scope first time was rejected")
		}
		last, e := strconv.ParseInt(*row.Last, 10, 64)
		if e != nil || first < q.FromMS || last < first || last >= q.ToMS || strconv.FormatInt(first, 10) != *row.First || strconv.FormatInt(last, 10) != *row.Last {
			return nil, fmt.Errorf("metric scope event bounds were rejected")
		}
		environment := identity.EnvironmentScope{Mode: "unreported"}
		if *row.EnvironmentPresent == "1" {
			environment.Mode = "empty"
			if *row.Environment != "" {
				environment.Mode = "named"
				environment.Value = *row.Environment
			}
		}
		namespace := &identity.Namespace{Mode: "missing"}
		if *row.NamespacePresent == "1" {
			namespace.Mode = "value"
			namespace.Value = *row.Namespace
		}
		choice := MetricScopeChoice{Service: *row.Service, Namespace: namespace, Environment: environment, Points: *row.Points, FirstMS: first, LastMS: last, Selectable: true}
		if selection, e := identity.EffectiveEnvironment(environment); e == nil {
			wire, _ := selection.Selector()
			choice.EnvironmentSelector = &wire
		} else {
			choice.Selectable = false
			choice.Reason = "Environment label exceeds selector limits."
		}
		if namespace.Validate() != nil {
			choice.Selectable = false
			choice.Reason = "Namespace exceeds selector limits."
		}
		if *row.Service == "" {
			choice.Selectable = false
			choice.Reason = "Service name is empty or not reported."
		} else if len(*row.Service) > 128 || strings.ContainsRune(*row.Service, 0) {
			choice.Selectable = false
			choice.Reason = "Service name exceeds selector limits."
		}
		key, _ := json.Marshal([]any{environment, namespace, *row.Service})
		if seen[string(key)] {
			return nil, fmt.Errorf("duplicate metric scope identity")
		}
		seen[string(key)] = true
		choices = append(choices, choice)
	}
	return choices, nil
}
