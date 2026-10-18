package measurements

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/metrics"
	"io"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Source struct {
	Endpoint, User, Password, CustodyDir string
	Client                               *http.Client
}

func Compile(q Question) (Plan, error) {
	if e := q.Validate(); e != nil {
		return Plan{}, e
	}
	d := q.Definition
	p := Plan{Revision: Hash(d), Parameters: map[string]string{}}
	param := func(v, t string) string {
		k := fmt.Sprintf("p%d", len(p.Parameters))
		p.Parameters[k] = v
		return "{" + k + ":" + t + "}"
	}
	from := q.FromMS
	if d.Kind == "counter" || d.Kind == "histogram" {
		from -= d.MaxGapMS
		if from < 0 {
			from = 0
		}
	}
	if d.Kind == "lifecycle" {
		from -= d.Lifecycle.DeadlineMS + d.Lifecycle.AllowanceMS
	}
	timeColumn := "TimeUnix"
	table := "otel_metrics_" + map[string]string{"gauge": "gauge", "counter": "sum", "histogram": "histogram"}[d.Kind]
	fields := `mapUpdate(ResourceAttributes,map('service.name',toString(ServiceName))) AS resource, ScopeName AS scope, ScopeVersion AS version, ScopeSchemaUrl AS schema_url, ResourceSchemaUrl AS resource_schema_url, MetricName AS name, MetricUnit AS unit, Attributes AS attributes, toString(toUnixTimestamp64Nano(StartTimeUnix)) AS start_ns, toString(toUnixTimestamp64Nano(TimeUnix)) AS end_ns, Flags AS flags`
	switch d.Kind {
	case "gauge":
		fields += ", toString(Value) AS value"
	case "counter":
		fields += ", toString(Value) AS value, AggregationTemporality AS temporality, IsMonotonic AS monotonic"
	case "histogram":
		fields += ", AggregationTemporality AS temporality, toString(Count) AS count, arrayMap(c -> toString(c),BucketCounts) AS bucket_counts, ExplicitBounds AS bounds"
	case "lifecycle":
		table = "otel_logs"
		timeColumn = "Timestamp"
		fields = "toString(toUnixTimestamp64Nano(Timestamp)) AS timestamp_ns, Body AS body, TraceId AS trace_id, SpanId AS span_id, LogAttributes AS attributes, ResourceAttributes AS resource, toString(ServiceName) AS service"
	}
	upper := "<"
	if d.Kind == "counter" || d.Kind == "histogram" {
		upper = "<="
	}
	where := timeColumn + ">=fromUnixTimestamp64Milli(" + param(strconv.FormatInt(from, 10), "Int64") + ") AND " + timeColumn + upper + "fromUnixTimestamp64Milli(" + param(strconv.FormatInt(q.ToMS, 10), "Int64") + ") AND ResourceAttributes['deployment.environment.name']=" + param(d.Environment, "String")
	services := []string{}
	for _, s := range d.Source.Services {
		services = append(services, param(s, "String"))
	}
	where += " AND ServiceName IN (" + strings.Join(services, ",") + ")"
	if d.Kind == "lifecycle" {
		l := d.Lifecycle
		where += " AND Body IN (" + param(l.Accepted, "String") + "," + param(l.Terminal, "String") + ")"
	} else {
		where += " AND ScopeName=" + param(d.Source.Scope, "String") + " AND ScopeVersion=" + param(d.Source.Version, "String") + " AND ScopeSchemaUrl=" + param(d.Source.SchemaURL, "String") + " AND MetricName=" + param(d.Source.Metric, "String") + " AND MetricUnit=" + param(d.Source.Unit, "String")
	}
	keys := []string{}
	for k := range d.Source.Attributes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		key := param(k, "String")
		where += " AND mapContains(Attributes," + key + ") AND Attributes[" + key + "]=" + param(d.Source.Attributes[k], "String")
	}
	p.SQL = "SELECT " + fields + " FROM " + table + " WHERE " + where + " ORDER BY " + timeColumn + " LIMIT 10001 FORMAT JSONEachRow"
	return p, nil
}
func (s Source) Run(ctx context.Context, q Question, previous *Result) (*Result, error) {
	plan, e := Compile(q)
	if e != nil {
		return nil, e
	}
	u, e := url.Parse(s.Endpoint)
	if e != nil {
		return nil, e
	}
	v := u.Query()
	for k, x := range plan.Parameters {
		esc := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\t", "\\t", "\r", "\\r", "\x00", "\\0").Replace(x)
		v.Set("param_"+k, esc)
	}
	for k, x := range map[string]string{"readonly": "1", "max_execution_time": "10", "max_result_rows": "10001", "result_overflow_mode": "throw", "max_bytes_to_read": "268435456", "read_overflow_mode": "throw"} {
		v.Set(k, x)
	}
	u.RawQuery = v.Encode()
	req, e := http.NewRequestWithContext(ctx, "POST", u.String(), strings.NewReader(plan.SQL))
	if e != nil {
		return nil, e
	}
	req.SetBasicAuth(s.User, s.Password)
	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	resp, e := client.Do(req)
	if e != nil {
		return nil, e
	}
	defer resp.Body.Close()
	b, e := io.ReadAll(io.LimitReader(resp.Body, (8<<20)+1))
	if e != nil {
		return nil, e
	}
	if resp.StatusCode != 200 || len(b) > 8<<20 {
		return nil, fmt.Errorf("bounded measurement source HTTP %d / payload limit", resp.StatusCode)
	}
	rows := []json.RawMessage{}
	scanner := bufio.NewScanner(bytes.NewReader(b))
	scanner.Buffer(make([]byte, 4096), 8<<20)
	for scanner.Scan() {
		rows = append(rows, append(json.RawMessage(nil), scanner.Bytes()...))
		if len(rows) > RowLimit {
			return nil, fmt.Errorf("source exceeds 10000 rows; no partial measurement")
		}
	}
	if e = scanner.Err(); e != nil {
		return nil, e
	}
	return Evaluate(q, rows, time.Now().UTC(), s.CustodyDir, previous)
}

type metricRow struct {
	ResourceSchemaURL string            `json:"resource_schema_url"`
	Resource          map[string]string `json:"resource"`
	Scope             string            `json:"scope"`
	Version           string            `json:"version"`
	SchemaURL         string            `json:"schema_url"`
	Name              string            `json:"name"`
	Unit              string            `json:"unit"`
	Attributes        map[string]string `json:"attributes"`
	StartNS           string            `json:"start_ns"`
	EndNS             string            `json:"end_ns"`
	Value             string            `json:"value"`
	Flags             uint32            `json:"flags"`
	Temporality       int               `json:"temporality"`
	Monotonic         bool              `json:"monotonic"`
	Count             string            `json:"count"`
	Bounds            []float64         `json:"bounds"`
	BucketCounts      []string          `json:"bucket_counts"`
}

func ns(ms int64) string { return strconv.FormatInt(ms*1000000, 10) }
func Evaluate(q Question, rows []json.RawMessage, at time.Time, custodyDir string, previous *Result) (*Result, error) {
	encodedQuestion, _ := json.Marshal(q)
	var cloned Question
	json.Unmarshal(encodedQuestion, &cloned)
	q = cloned
	plan, e := Compile(q)
	if e != nil {
		return nil, e
	}
	if len(rows) > RowLimit {
		return nil, fmt.Errorf("no partial source evaluation")
	}
	copied, e := canonicalRows(rows)
	if e != nil {
		return nil, e
	}
	r := &Result{Question: q, Plan: plan, ReadAt: at.UTC().Format(time.RFC3339Nano), Rows: copied, Count: strconv.Itoa(len(rows)), State: "no_data", Warnings: []string{"Captured observations do not establish ingestion completeness or application health. Source limits fail closed; no measurement total comes from a row page.", "Streams retain complete resource, scope, metric and attribute identity. No automatic fleet merging, interpolation or averaging of percentiles.", "Consumer-group observations may be repeated by multiple workers. Never sum repeated broker observations across instances."}}
	r.Basis = Hash(struct {
		Q    Question
		At   string
		Rows []json.RawMessage
	}{q, r.ReadAt, copied})
	if q.Definition.Kind == "lifecycle" {
		return evaluateLifecycle(r, at, custodyDir, previous)
	}
	gs := []metrics.GaugePoint{}
	cs := []metrics.CounterPoint{}
	hs := []metrics.HistogramPoint{}
	d := q.Definition
	for _, raw := range copied {
		var row metricRow
		if e = json.Unmarshal(raw, &row); e != nil {
			return nil, e
		}
		serviceOK := false
		for _, name := range d.Source.Services {
			serviceOK = serviceOK || row.Resource["service.name"] == name
		}
		if !serviceOK || row.Resource["deployment.environment.name"] != d.Environment || row.Resource["service.instance.id"] == "" || row.Scope != d.Source.Scope || row.Version != d.Source.Version || row.SchemaURL != d.Source.SchemaURL || row.Name != d.Source.Metric || row.Unit != d.Source.Unit || row.Flags&^uint32(1) != 0 {
			return nil, fmt.Errorf("metric source identity/unit/flags mismatch")
		}
		for k, v := range d.Source.Attributes {
			actual, ok := row.Attributes[k]
			if !ok || actual != v {
				return nil, fmt.Errorf("metric selector mismatch")
			}
		}
		stream := metrics.Stream{ResourceSchemaURL: row.ResourceSchemaURL, Resource: row.Resource, Scope: row.Scope, Version: row.Version, SchemaURL: row.SchemaURL, Name: row.Name, Unit: row.Unit, Attributes: row.Attributes}
		missing := row.Flags&1 != 0
		if d.Kind == "gauge" || d.Kind == "counter" {
			value := float64(0)
			if !missing {
				value, e = strconv.ParseFloat(row.Value, 64)
				if e != nil || math.IsNaN(value) || math.IsInf(value, 0) {
					return nil, fmt.Errorf("invalid numeric observation")
				}
			}
			if d.Kind == "gauge" {
				gs = append(gs, metrics.GaugePoint{Stream: stream, EndNS: row.EndNS, Value: value, NoRecordedValue: missing})
				continue
			}
			if value < 0 || value >= 9007199254740992 || math.Trunc(value) != value || !row.Monotonic {
				return nil, fmt.Errorf("counter precision/monotonicity unsupported")
			}
			temp := map[int]string{1: "delta", 2: "cumulative"}[row.Temporality]
			cs = append(cs, metrics.CounterPoint{Stream: stream, Temporality: temp, Monotonic: true, StartNS: row.StartNS, EndNS: row.EndNS, Value: strconv.FormatUint(uint64(value), 10), NoRecordedValue: missing})
		} else {
			hs = append(hs, metrics.HistogramPoint{Stream: stream, Temporality: map[int]string{1: "delta", 2: "cumulative"}[row.Temporality], StartNS: row.StartNS, EndNS: row.EndNS, Bounds: row.Bounds, BucketCounts: row.BucketCounts, Count: row.Count, NoRecordedValue: missing})
		}
	}
	switch d.Kind {
	case "gauge":
		r.Gauges, e = metrics.EvaluateGauges(ns(q.FromMS), ns(q.ToMS), ns(d.MaxGapMS), gs)
	case "counter":
		r.Counters, e = metrics.EvaluateCounters(metrics.CounterRequest{FromNS: ns(q.FromMS), ToNS: ns(q.ToMS), MaxGapNS: ns(d.MaxGapMS), Points: cs})
	case "histogram":
		r.Histograms, e = metrics.EvaluateHistograms(metrics.HistogramRequest{FromNS: ns(q.FromMS), ToNS: ns(q.ToMS), MaxGapNS: ns(d.MaxGapMS), Points: hs})
	}
	if e != nil {
		return nil, e
	}
	if len(rows) > 0 {
		r.State = "observed"
	}
	if r.Gauges != nil {
		for _, s := range r.Gauges.Series {
			if s.State != "observed" {
				r.State = "unknown"
			}
		}
	}
	if r.Counters != nil {
		for _, s := range r.Counters.Series {
			if s.State != "ok" {
				r.State = "unknown"
			}
		}
	}
	if r.Histograms != nil {
		for _, s := range r.Histograms.Series {
			if s.State != "ok" {
				r.State = "unknown"
			}
		}
	}

	return r, nil
}
