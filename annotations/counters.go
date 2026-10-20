package annotations

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/metrics"
	"io"
	"math"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type RateContract struct {
	Name        string   `yaml:"name" json:"name"`
	Metric      string   `yaml:"metric" json:"metric"`
	Unit        string   `yaml:"unit" json:"unit"`
	Temporality string   `yaml:"temporality" json:"temporality"`
	MaxGap      string   `yaml:"maxGap" json:"maxGap"`
	Outcomes    []string `yaml:"outcomes" json:"outcomes"`
}
type CounterRow struct {
	Resource    map[string]string `json:"resource"`
	Scope       string            `json:"scope"`
	Version     string            `json:"version"`
	SchemaURL   string            `json:"schema_url"`
	Name        string            `json:"name"`
	Unit        string            `json:"unit"`
	Attributes  map[string]string `json:"attributes"`
	StartNS     string            `json:"start_ns"`
	EndNS       string            `json:"end_ns"`
	Value       string            `json:"value"`
	Flags       uint32            `json:"flags"`
	Temporality int               `json:"temporality"`
	Monotonic   bool              `json:"monotonic"`
}
type LiveRate struct {
	Contract RateContract           `json:"contract"`
	Result   *metrics.CounterResult `json:"result"`
}
type LiveRates struct {
	Fleet           *LiveFleet  `json:"fleet,omitempty"`
	PackageRevision string      `json:"packageRevision"`
	FromMS          int64       `json:"fromMs"`
	ToMS            int64       `json:"toMs"`
	Ratios          []LiveRatio `json:"ratios"`
	Rates           []LiveRate  `json:"rates"`
	Warnings        []string    `json:"warnings"`
}

func (p *OperationPlan) EvaluateRateRows(contract RateContract, from, to int64, rows []CounterRow) (*metrics.CounterResult, error) {
	if len(rows) > 3000 {
		return nil, fmt.Errorf("counter source exceeds 3000 points; no partial rates")
	}
	gap, err := time.ParseDuration(contract.MaxGap)
	if err != nil {
		return nil, err
	}
	points := []metrics.CounterPoint{}
	outcomes := map[string]bool{}
	for _, v := range contract.Outcomes {
		outcomes[v] = true
	}
	ops := map[string]bool{}
	for _, op := range p.Definition.Operations {
		ops[op.Name+"/"+op.Role] = true
	}
	for _, row := range rows {
		if !p.Source.Namespace.Matches(row.Resource) || row.Resource["service.name"] != p.Source.Service || row.Resource["deployment.environment.name"] != p.Source.Environment || row.Resource["service.instance.id"] != p.Source.Instance || row.Scope != p.Manifest.Scope || row.Version != p.Manifest.ScopeVersion || row.SchemaURL != p.Manifest.SchemaURL || row.Name != contract.Metric || row.Unit != contract.Unit {
			return nil, fmt.Errorf("counter resource/instrumentation/unit binding mismatch")
		}
		if len(row.Attributes) != 3 || !ops[row.Attributes["goo.operation.name"]+"/"+row.Attributes["goo.operation.role"]] {
			return nil, fmt.Errorf("counter operation/dimension contract mismatch")
		}
		outcome := row.Attributes["goo.operation.outcome"]
		if outcome != "ok" && outcome != "error" && outcome != "panic" {
			return nil, fmt.Errorf("unknown counter outcome")
		}
		end, e := strconv.ParseInt(row.EndNS, 10, 64)
		if e != nil || end < from*1000000-int64(gap) || end > to*1000000 {
			return nil, fmt.Errorf("counter point outside bounded source window")
		}
		if row.StartNS == "0" {
			return nil, fmt.Errorf("missing OTLP counter start time")
		}
		temporality := map[int]string{1: "delta", 2: "cumulative"}[row.Temporality]
		if temporality != contract.Temporality || !row.Monotonic || row.Flags&^uint32(1) != 0 {
			return nil, fmt.Errorf("unsupported counter temporality/flags/monotonicity")
		}
		value := ""
		if row.Flags&1 == 0 {
			f, e := strconv.ParseFloat(row.Value, 64)
			if e != nil || math.IsInf(f, 0) || math.IsNaN(f) || f < 0 || f >= 9007199254740992 || math.Trunc(f) != f {
				return nil, fmt.Errorf("exported counter must be an exact nonnegative integer below 2^53")
			}
			value = strconv.FormatUint(uint64(f), 10)
		}
		if !outcomes[outcome] {
			continue
		}
		points = append(points, metrics.CounterPoint{Stream: metrics.Stream{Resource: row.Resource, Scope: row.Scope, Version: row.Version, SchemaURL: row.SchemaURL, Name: row.Name, Unit: row.Unit, Attributes: row.Attributes}, Temporality: temporality, Monotonic: true, StartNS: row.StartNS, EndNS: row.EndNS, Value: value, NoRecordedValue: row.Flags&1 != 0})
	}
	result, err := metrics.EvaluateCounters(metrics.CounterRequest{FromNS: strconv.FormatInt(from*1000000, 10), ToNS: strconv.FormatInt(to*1000000, 10), MaxGapNS: strconv.FormatInt(int64(gap), 10), Points: points})
	if err != nil {
		return nil, err
	}
	payload, _ := json.Marshal([]string{p.Revision, contract.Name, result.Basis})
	result.Basis = hash(payload)
	return result, nil
}
func (p *OperationPlan) fetchInstanceRates(ctx context.Context, client *http.Client, endpoint, user, password string, from, to int64) (*LiveRates, error) {
	fresh, err := p.Definition.Bind(p.Manifest)
	if err != nil || fresh.Revision != p.Revision || !reflect.DeepEqual(fresh.Source, p.Source) {
		return nil, fmt.Errorf("invalid operation plan")
	}
	_, params, err := p.Manifest.OperationQuery(p.Source, from, to)
	if err != nil {
		return nil, err
	}
	result := &LiveRates{PackageRevision: p.Revision, FromMS: from, ToMS: to, Rates: []LiveRate{}, Warnings: []string{"Observed counter intervals only; sampling and ingestion completeness remain unknown.", "Metric and span queries have separate evidence bases; time-matched spans are contextual, not exact counter contributors.", "Float64 exporter values at or above 2^53 are rejected; no precision is invented."}}
	for _, contract := range p.Definition.Rates {
		gap, _ := time.ParseDuration(contract.MaxGap)
		params["baseline"] = strconv.FormatInt(from-int64(gap/time.Millisecond), 10)
		params["metric"] = contract.Metric
		sql := `SELECT mapUpdate(ResourceAttributes,map('service.name',toString(ServiceName))) AS resource, ScopeName AS scope, ScopeVersion AS version, ScopeSchemaUrl AS schema_url, MetricName AS name, MetricUnit AS unit, Attributes AS attributes, toString(toUnixTimestamp64Nano(StartTimeUnix)) AS start_ns, toString(toUnixTimestamp64Nano(TimeUnix)) AS end_ns, toString(Value) AS value, Flags AS flags, AggregationTemporality AS temporality, IsMonotonic AS monotonic FROM otel_metrics_sum WHERE TimeUnix >= fromUnixTimestamp64Milli({baseline:Int64}) AND TimeUnix <= fromUnixTimestamp64Milli({to:Int64}) AND ServiceName={service:String} AND ResourceAttributes['deployment.environment.name']={environment:String} AND ResourceAttributes['service.instance.id']={instance:String} AND ScopeName={scope:String} AND ScopeVersion={version:String} AND MetricName={metric:String} ORDER BY TimeUnix LIMIT 3001 FORMAT JSONEachRow`
		sql = bindOperationNamespace(sql, p.Source.Namespace, params)
		address, e := url.Parse(endpoint)
		if e != nil {
			return nil, e
		}
		q := address.Query()
		for k, v := range params {
			q.Set("param_"+k, v)
		}
		q.Set("readonly", "1")
		q.Set("max_execution_time", "10")
		q.Set("max_result_rows", "3001")
		q.Set("result_overflow_mode", "throw")
		address.RawQuery = q.Encode()
		req, e := http.NewRequestWithContext(ctx, "POST", address.String(), strings.NewReader(sql))
		if e != nil {
			return nil, e
		}
		req.SetBasicAuth(user, password)
		response, e := client.Do(req)
		if e != nil {
			return nil, e
		}
		data, e := io.ReadAll(io.LimitReader(response.Body, (8<<20)+1))
		response.Body.Close()
		if e != nil {
			return nil, e
		}
		if response.StatusCode != 200 {
			return nil, fmt.Errorf("counter source HTTP %d", response.StatusCode)
		}
		if len(data) > 8<<20 {
			return nil, fmt.Errorf("counter source payload exceeds 8 MiB")
		}
		rows := []CounterRow{}
		dec := json.NewDecoder(strings.NewReader(string(data)))
		for {
			var row CounterRow
			e = dec.Decode(&row)
			if e == io.EOF {
				break
			}
			if e != nil {
				return nil, e
			}
			rows = append(rows, row)
			if len(rows) > 3000 {
				return nil, fmt.Errorf("counter source exceeds 3000 points")
			}
		}
		evaluated, e := p.EvaluateRateRows(contract, from, to, rows)
		if e != nil {
			return nil, e
		}
		result.Rates = append(result.Rates, LiveRate{contract, evaluated})
	}
	result.Ratios, err = p.evaluateRatios(result.Rates)
	if err != nil {
		return nil, err
	}
	return result, nil
}
