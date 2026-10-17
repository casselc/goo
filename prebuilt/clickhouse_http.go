package prebuilt

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/casselc/goo/measurement"
)

// HTTPSource is an explicit library/version binding. The pinned exporter does
// not retain scope schema URLs, so schema compatibility remains a declaration.
type HTTPSource struct {
	ScopeName    string `json:"scopeName"`
	ScopeVersion string `json:"scopeVersion"`
	SchemaURL    string `json:"schemaURL"`
}
type HTTPSpanRow struct {
	TimestampNS string            `json:"timestamp_ns"`
	TraceID     string            `json:"trace_id"`
	SpanID      string            `json:"span_id"`
	Resource    map[string]string `json:"resource"`
	Attributes  map[string]string `json:"attributes"`
}
type LiveHTTPResult struct {
	Expansion        *Expansion           `json:"expansion"`
	Prepared         *HTTPPrepared        `json:"prepared"`
	Results          []measurement.Result `json:"results"`
	Source           HTTPSource           `json:"source"`
	FromMS           int64                `json:"fromMs"`
	ToMS             int64                `json:"toMs"`
	DuplicateRecords int                  `json:"duplicateRecords"`
	Query            string               `json:"query"`
	Parameters       map[string]string    `json:"parameters"`
	Warnings         []string             `json:"warnings"`
}

func (e *Expansion) HTTPSourceQuery(source HTTPSource, from, to int64) (string, map[string]string, error) {
	if from < 0 || to <= from || to-from > 120*60*1000 || to > 4102444800000 {
		return "", nil, fmt.Errorf("require a nonempty half-open window of at most 120 minutes before 2100")
	}
	if source.ScopeName == "" || source.ScopeVersion == "" || source.SchemaURL != "https://opentelemetry.io/schemas/1.37.0" {
		return "", nil, fmt.Errorf("explicit supported instrumentation scope/version/schema binding required")
	}
	params := map[string]string{"from": strconv.FormatInt(from, 10), "to": strconv.FormatInt(to, 10), "scope": source.ScopeName, "version": source.ScopeVersion}
	filters := []string{"Timestamp >= fromUnixTimestamp64Milli({from:Int64})", "Timestamp < fromUnixTimestamp64Milli({to:Int64})", "SpanKind = 'Server'", "ScopeName = {scope:String}", "ScopeVersion = {version:String}"}
	keys := []string{}
	for k := range e.Resource {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i, key := range keys {
		k, v := fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i)
		params[k] = key
		params[v] = e.Resource[key]
		if key == "service.name" {
			filters = append(filters, "ServiceName = {"+v+":String}")
		} else {
			filters = append(filters, "ResourceAttributes[{"+k+":String}] = {"+v+":String}")
		}
	}
	sql := "SELECT toString(toUnixTimestamp64Nano(Timestamp)) AS timestamp_ns, TraceId AS trace_id, SpanId AS span_id, mapUpdate(ResourceAttributes, map('service.name', toString(ServiceName))) AS resource, SpanAttributes AS attributes FROM otel_traces WHERE " + strings.Join(filters, " AND ") + " ORDER BY Timestamp, TraceId, SpanId LIMIT 1001 FORMAT JSONEachRow"
	return sql, params, nil
}
func (e *Expansion) EvaluateHTTPRows(rows []HTTPSpanRow, evaluation string) (*HTTPPrepared, []measurement.Result, int, error) {
	if len(rows) > 1000 {
		return nil, nil, 0, fmt.Errorf("source bound exceeded (1000 raw spans); narrow the window; partial population not evaluated")
	}
	requests := []HTTPRequest{}
	seen := map[string]string{}
	duplicates := 0
	for _, row := range rows {
		for _, identity := range []struct {
			value string
			size  int
		}{{row.TraceID, 32}, {row.SpanID, 16}} {
			if len(identity.value) != identity.size || strings.Trim(identity.value, "0") == "" {
				return nil, nil, 0, fmt.Errorf("invalid OTel trace/span identity")
			}
			if _, err := hex.DecodeString(identity.value); err != nil {
				return nil, nil, 0, fmt.Errorf("invalid OTel trace/span identity")
			}
		}
		if timestamp, err := strconv.ParseInt(row.TimestampNS, 10, 64); err != nil || timestamp < 0 {
			return nil, nil, 0, fmt.Errorf("invalid exact source timestamp")
		}

		// Equality includes the exact source timestamp and all source attributes.
		key := row.TraceID + ":" + row.SpanID
		payload := digest(row)
		if prior, ok := seen[key]; ok {
			if prior != payload {
				return nil, nil, 0, fmt.Errorf("conflicting source records for one span identity")
			}
			duplicates++
			continue
		}
		seen[key] = payload
		var status *int
		if raw, ok := row.Attributes["http.response.status_code"]; ok {
			if parsed, err := strconv.Atoi(raw); err == nil {
				status = &parsed
			}
		}
		requests = append(requests, HTTPRequest{Kind: "server", Resource: row.Resource, TraceID: row.TraceID, SpanID: row.SpanID, Route: row.Attributes["http.route"], Method: row.Attributes["http.request.method"], Status: status})
	}
	prepared, err := e.PrepareHTTP(requests, evaluation)
	if err != nil {
		return nil, nil, 0, err
	}
	results, err := e.Measurement.Evaluate(prepared.Facts, evaluation, e.Views[0].GroupBy)
	return prepared, results, duplicates, err
}
func FetchHTTP(ctx context.Context, client *http.Client, endpoint, user, password string, e *Expansion, source HTTPSource, from, to int64) (*LiveHTTPResult, error) {
	sql, params, err := e.HTTPSourceQuery(source, from, to)
	if err != nil {
		return nil, err
	}
	address, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	query := address.Query()
	for key, value := range params {
		query.Set("param_"+key, value)
	}
	query.Set("readonly", "1")
	query.Set("max_execution_time", "10")
	query.Set("max_result_rows", "1001")
	query.Set("result_overflow_mode", "throw")
	address.RawQuery = query.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, address.String(), strings.NewReader(sql))
	if err != nil {
		return nil, err
	}
	request.SetBasicAuth(user, password)
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("ClickHouse source returned HTTP %d", response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, (8<<20)+1))
	if err != nil {
		return nil, err
	}
	if len(data) > 8<<20 {
		return nil, fmt.Errorf("source payload exceeds 8 MiB; no partial evaluation")
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	rows := []HTTPSpanRow{}
	for {
		var row HTTPSpanRow
		err = decoder.Decode(&row)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	for _, row := range rows {
		timestamp, err := strconv.ParseInt(row.TimestampNS, 10, 64)
		if err != nil || timestamp < from*1000000 || timestamp >= to*1000000 {
			return nil, fmt.Errorf("source returned a span outside the bound window")
		}
	}
	evaluation := digest(struct {
		Revision string
		From, To int64
		Rows     []HTTPSpanRow
	}{e.Revision, from, to, rows})
	prepared, results, duplicates, err := e.EvaluateHTTPRows(rows, evaluation)
	if err != nil {
		return nil, err
	}
	return &LiveHTTPResult{Expansion: e, Prepared: prepared, Results: results, Source: source, FromMS: from, ToMS: to, DuplicateRecords: duplicates, Query: sql, Parameters: params, Warnings: []string{"Observed server spans only; sampling and ingestion completeness unverified.", "Window is fixed by span start time, not immutable custody time; later reruns can change.", "Schema version is declared by the library binding: this exporter omits scope schema URLs.", "Source bound is 1000 raw records; overflow fails instead of evaluating a partial population."}}, nil
}
