package annotations

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/internal/sourcehttp"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// OperationSource is an explicit physical binding, not an automatically inferred schema.
type OperationSource struct {
	Namespace        *identity.Namespace `json:"namespace,omitempty"`
	SubjectNamespace string              `json:"subjectNamespace,omitempty"`
	Service          string              `json:"service"`
	Environment      string              `json:"environment"`
	Instance         string              `json:"instance"`
	DeclaredSchema   string              `json:"declaredSchema"`
}
type OperationRow struct {
	EventTimes      []string            `json:"event_times"`
	EventNames      []string            `json:"event_names"`
	EventAttributes []map[string]string `json:"event_attributes"`
	TimestampNS     string              `json:"timestamp_ns"`
	DurationNS      string              `json:"duration_ns"`
	TraceID         string              `json:"trace_id"`
	SpanID          string              `json:"span_id"`
	ParentSpanID    string              `json:"parent_span_id"`
	Kind            string              `json:"kind"`
	Scope           string              `json:"scope"`
	Version         string              `json:"version"`
	Resource        map[string]string   `json:"resource"`
	Attributes      map[string]string   `json:"attributes"`
}

// OperationSpan is logical evidence shared by overview, distribution and inspection.
// Exact timestamps/durations remain decimal strings across browser transport.
type OperationSpan struct {
	Events       []SubjectEvent    `json:"events"`
	TraceID      string            `json:"traceID"`
	SpanID       string            `json:"spanID"`
	ParentSpanID string            `json:"parentSpanID"`
	StartNS      string            `json:"startNS"`
	DurationNS   string            `json:"durationNS"`
	Operation    string            `json:"operation"`
	Role         string            `json:"role"`
	Outcome      string            `json:"outcome"`
	Attributes   map[string]string `json:"attributes"`
}
type LiveOperations struct {
	Package *OperationPlan  `json:"package,omitempty"`
	Spans   []OperationSpan `json:"spans"`

	Conformance      Conformance       `json:"conformance"`
	Source           OperationSource   `json:"source"`
	FromMS           int64             `json:"fromMs"`
	ToMS             int64             `json:"toMs"`
	Query            string            `json:"query"`
	Parameters       map[string]string `json:"parameters"`
	DuplicateRecords int               `json:"duplicateRecords"`
}

func (m Manifest) OperationQuery(s OperationSource, from, to int64) (string, map[string]string, error) {
	if err := s.Namespace.Validate(); err != nil {
		return "", nil, err
	}
	if from < 0 || to <= from || to-from > 120*60000 || to > 4102444800000 {
		return "", nil, fmt.Errorf("require a half-open window of at most 120 minutes before 2100")
	}
	if s.Service == "" || s.Environment == "" || s.Instance == "" || s.DeclaredSchema != m.SchemaURL {
		return "", nil, fmt.Errorf("explicit service/environment/instance/schema binding required")
	}
	p := map[string]string{"from": strconv.FormatInt(from, 10), "to": strconv.FormatInt(to, 10), "service": s.Service, "environment": s.Environment, "instance": s.Instance, "scope": m.Scope, "version": m.ScopeVersion}
	q := `SELECT toString(toUnixTimestamp64Nano(Timestamp)) AS timestamp_ns, toString(Duration) AS duration_ns, TraceId AS trace_id, SpanId AS span_id, ParentSpanId AS parent_span_id, SpanKind AS kind, ScopeName AS scope, ScopeVersion AS version, mapUpdate(ResourceAttributes, map('service.name', toString(ServiceName))) AS resource, SpanAttributes AS attributes, arrayMap(t -> toString(toUnixTimestamp64Nano(t)), Events.Timestamp) AS event_times, Events.Name AS event_names, Events.Attributes AS event_attributes FROM otel_traces WHERE Timestamp >= fromUnixTimestamp64Milli({from:Int64}) AND Timestamp < fromUnixTimestamp64Milli({to:Int64}) AND ServiceName = {service:String} AND ResourceAttributes['deployment.environment.name'] = {environment:String} AND ResourceAttributes['service.instance.id'] = {instance:String} AND ScopeName = {scope:String} AND ScopeVersion = {version:String} ORDER BY Timestamp, TraceId, SpanId LIMIT 1001 FORMAT JSONEachRow`
	q = bindOperationNamespace(q, s.Namespace, p)
	return q, p, nil
}
func (m Manifest) ObserveRows(s OperationSource, from, to int64, rows []OperationRow) (Conformance, int, error) {
	if _, _, err := m.OperationQuery(s, from, to); err != nil {
		return Conformance{}, 0, err
	}
	if len(rows) > 1000 {
		return Conformance{}, 0, fmt.Errorf("source exceeds 1000 raw spans; narrow window; no partial totals")
	}
	observations := []Observation{}
	seen := map[string]string{}
	duplicates := 0
	for _, r := range rows {
		ns, err := strconv.ParseInt(r.TimestampNS, 10, 64)
		if err != nil || ns < from*1000000 || ns >= to*1000000 {
			return Conformance{}, 0, fmt.Errorf("span outside exact window")
		}
		if r.ParentSpanID != "" && r.ParentSpanID != "0000000000000000" && !hexID(r.ParentSpanID, 16) {
			return Conformance{}, 0, fmt.Errorf("invalid parent span identity")
		}
		duration, err := strconv.ParseUint(r.DurationNS, 10, 64)
		if err != nil {
			return Conformance{}, 0, fmt.Errorf("invalid span duration")
		}
		if !s.Namespace.Matches(r.Resource) || r.Resource["service.name"] != s.Service || r.Resource["deployment.environment.name"] != s.Environment || r.Resource["service.instance.id"] != s.Instance {
			return Conformance{}, 0, fmt.Errorf("source crossed resource binding")
		}
		role := r.Attributes["goo.operation.role"]
		kinds := map[string]string{"server": "Server", "client": "Client", "internal": "Internal", "producer": "Producer", "consumer": "Consumer"}
		if kinds[role] == "" || r.Kind != kinds[role] {
			return Conformance{}, 0, fmt.Errorf("operation role/span kind mismatch")
		}
		if _, err := m.SubjectEvents(s, r); err != nil {
			return Conformance{}, 0, err
		}
		payload, _ := json.Marshal(r)
		id := r.TraceID + "/" + r.SpanID
		if prior, ok := seen[id]; ok {
			if prior != string(payload) {
				return Conformance{}, 0, fmt.Errorf("conflicting raw span identity")
			}
			duplicates++
			continue
		}
		seen[id] = string(payload)
		observations = append(observations, Observation{Resource: s.Service + "/" + s.Environment + "/" + s.Instance, Scope: r.Scope, ScopeVersion: r.Version, SchemaURL: s.DeclaredSchema, TraceID: r.TraceID, SpanID: r.SpanID, Name: r.Attributes["goo.operation.name"], Role: role, Outcome: r.Attributes["goo.operation.outcome"], DurationNS: duration})
	}
	payload, _ := json.Marshal(struct {
		Revision string
		Source   OperationSource
		From, To int64
		Rows     []OperationRow
	}{m.Revision, s, from, to, rows})
	report, err := m.Observe(s.Service+"/"+s.Environment+"/"+s.Instance, hash(payload), observations)
	report.Warnings = append(report.Warnings, "Schema compatibility is explicitly declared: this exporter drops scope schema URLs.", "Window is fixed by span start time, not custody time; later arrivals can change reruns.", "At most 1000 raw records; overflow fails instead of returning partial totals.")
	return report, duplicates, err
}
func FetchOperations(ctx context.Context, client *http.Client, endpoint, user, password string, m Manifest, s OperationSource, from, to int64) (*LiveOperations, error) {
	sql, params, err := m.OperationQuery(s, from, to)
	if err != nil {
		return nil, err
	}
	address, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	q := address.Query()
	for k, v := range params {
		q.Set("param_"+k, v)
	}
	q.Set("readonly", "1")
	q.Set("max_execution_time", "10")
	q.Set("max_result_rows", "1001")
	q.Set("result_overflow_mode", "throw")
	address.RawQuery = q.Encode()
	response, err := sourcehttp.Post(ctx, client, address.String(), user, password, sql)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("ClickHouse source returned HTTP %d", response.StatusCode)
	}
	data, oversize, err := sourcehttp.ReadBounded(response.Body, 8<<20)
	if err != nil {
		return nil, err
	}
	if oversize {
		return nil, fmt.Errorf("source exceeds 8 MiB; no partial totals")
	}
	rows := []OperationRow{}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	for {
		var row OperationRow
		err = decoder.Decode(&row)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
		if len(rows) > 1000 {
			return nil, fmt.Errorf("source exceeds 1000 raw spans; narrow window")
		}
	}
	report, duplicates, err := m.ObserveRows(s, from, to, rows)
	if err != nil {
		return nil, err
	}
	spans := []OperationSpan{}
	seen := map[string]bool{}
	for _, row := range rows {
		id := row.TraceID + "/" + row.SpanID
		if seen[id] {
			continue
		}
		seen[id] = true
		events, err := m.SubjectEvents(s, row)
		if err != nil {
			return nil, err
		}
		spans = append(spans, OperationSpan{Events: events, TraceID: row.TraceID, SpanID: row.SpanID, ParentSpanID: row.ParentSpanID, StartNS: row.TimestampNS, DurationNS: row.DurationNS, Operation: row.Attributes["goo.operation.name"], Role: row.Attributes["goo.operation.role"], Outcome: row.Attributes["goo.operation.outcome"], Attributes: row.Attributes})
	}
	return &LiveOperations{Spans: spans, Conformance: report, Source: s, FromMS: from, ToMS: to, Query: sql, Parameters: params, DuplicateRecords: duplicates}, nil
}

func bindOperationNamespace(sql string, n *identity.Namespace, params map[string]string) string {
	if n == nil {
		return sql
	}
	condition := " AND NOT mapContains(ResourceAttributes,'service.namespace')"
	if n.Mode == "value" {
		params["namespace"] = n.Value
		condition = " AND mapContains(ResourceAttributes,'service.namespace') AND ResourceAttributes['service.namespace']={namespace:String}"
	}
	return strings.Replace(sql, " ORDER BY", condition+" ORDER BY", 1)
}
