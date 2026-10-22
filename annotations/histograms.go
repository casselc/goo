package annotations

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/internal/sourcehttp"
	"github.com/casselc/goo/metrics"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type HistogramContract struct {
	RateContract `yaml:",inline"`
	Bounds       []float64 `yaml:"bounds" json:"bounds"`
}
type HistogramRow struct {
	CounterRow
	Bounds       []float64 `json:"bounds"`
	BucketCounts []string  `json:"bucket_counts"`
	Count        string    `json:"count"`
}
type LiveHistogram struct {
	Contract HistogramContract        `json:"contract"`
	Result   *metrics.HistogramResult `json:"result"`
}
type LiveHistograms struct {
	PackageRevision string          `json:"packageRevision"`
	FromMS          int64           `json:"fromMs"`
	ToMS            int64           `json:"toMs"`
	Histograms      []LiveHistogram `json:"histograms"`
	Warnings        []string        `json:"warnings"`
}

func (p *OperationPlan) EvaluateHistogramRows(contract HistogramContract, from, to int64, rows []HistogramRow) (*metrics.HistogramResult, error) {
	if len(rows) > 1000 {
		return nil, fmt.Errorf("histogram source exceeds 1000 points; no partial distributions")
	}
	gap, e := time.ParseDuration(contract.MaxGap)
	if e != nil {
		return nil, e
	}
	outcomes := map[string]bool{}
	for _, o := range contract.Outcomes {
		outcomes[o] = true
	}
	ops := map[string]bool{}
	for _, op := range p.Definition.Operations {
		ops[op.Name+"/"+op.Role] = true
	}
	points := []metrics.HistogramPoint{}
	for _, row := range rows {
		if !p.Source.Namespace.Matches(row.Resource) || row.Resource["service.name"] != p.Source.Service || row.Resource["deployment.environment.name"] != p.Source.Environment || row.Resource["service.instance.id"] != p.Source.Instance || row.Scope != p.Manifest.Scope || row.Version != p.Manifest.ScopeVersion || row.SchemaURL != p.Manifest.SchemaURL || row.Name != contract.Metric || row.Unit != contract.Unit {
			return nil, fmt.Errorf("histogram resource/instrumentation/unit mismatch")
		}
		if len(row.Attributes) != 3 || !ops[row.Attributes["goo.operation.name"]+"/"+row.Attributes["goo.operation.role"]] {
			return nil, fmt.Errorf("histogram operation/dimension mismatch")
		}
		outcome := row.Attributes["goo.operation.outcome"]
		if outcome != "ok" && outcome != "error" && outcome != "panic" {
			return nil, fmt.Errorf("unknown histogram outcome")
		}
		end, e := strconv.ParseInt(row.EndNS, 10, 64)
		if e != nil || end < from*1000000-int64(gap) || end > to*1000000 {
			return nil, fmt.Errorf("histogram point outside source window")
		}
		if row.StartNS == "0" {
			return nil, fmt.Errorf("unknown histogram start time")
		}
		temporality := map[int]string{1: "delta", 2: "cumulative"}[row.Temporality]
		if temporality != contract.Temporality || row.Flags&^uint32(1) != 0 {
			return nil, fmt.Errorf("histogram temporality/flags mismatch")
		}
		if row.Flags&1 == 0 {
			if len(row.Bounds) != len(contract.Bounds) {
				return nil, fmt.Errorf("histogram bucket layout does not match package")
			}
			for i, b := range row.Bounds {
				if b != contract.Bounds[i] {
					return nil, fmt.Errorf("histogram bucket layout does not match package")
				}
			}
		}
		if !outcomes[outcome] {
			continue
		}
		points = append(points, metrics.HistogramPoint{Stream: metrics.Stream{Resource: row.Resource, Scope: row.Scope, Version: row.Version, SchemaURL: row.SchemaURL, Name: row.Name, Unit: row.Unit, Attributes: row.Attributes}, Temporality: temporality, StartNS: row.StartNS, EndNS: row.EndNS, Bounds: row.Bounds, BucketCounts: row.BucketCounts, Count: row.Count, NoRecordedValue: row.Flags&1 != 0})
	}
	result, e := metrics.EvaluateHistograms(metrics.HistogramRequest{FromNS: strconv.FormatInt(from*1000000, 10), ToNS: strconv.FormatInt(to*1000000, 10), MaxGapNS: strconv.FormatInt(int64(gap), 10), Points: points})
	if e != nil {
		return nil, e
	}
	payload, _ := json.Marshal([]string{p.Revision, contract.Name, result.Basis})
	result.Basis = hash(payload)
	return result, nil
}
func (p *OperationPlan) FetchHistograms(ctx context.Context, client *http.Client, endpoint, user, password string, from, to int64) (*LiveHistograms, error) {
	fresh, e := p.Definition.Bind(p.Manifest)
	if e != nil || fresh.Revision != p.Revision || !reflect.DeepEqual(fresh.Source, p.Source) {
		return nil, fmt.Errorf("invalid histogram operation plan")
	}
	_, params, e := p.Manifest.OperationQuery(p.Source, from, to)
	if e != nil {
		return nil, e
	}
	result := &LiveHistograms{PackageRevision: p.Revision, FromMS: from, ToMS: to, Histograms: []LiveHistogram{}, Warnings: []string{"Native metric bucket populations; no span-derived reconstruction or scalar percentile averaging.", "P95 is a bucket range; metric and span bases are separate. Span links are contextual, not exact bucket contributors.", "Per-instance streams only; missing coverage, resets and incompatible layouts never become zero."}}
	for _, contract := range p.Definition.Histograms {
		gap, _ := time.ParseDuration(contract.MaxGap)
		params["baseline"] = strconv.FormatInt(from-int64(gap/time.Millisecond), 10)
		params["metric"] = contract.Metric
		sql := `SELECT mapUpdate(ResourceAttributes,map('service.name',toString(ServiceName))) AS resource, ScopeName AS scope, ScopeVersion AS version, ScopeSchemaUrl AS schema_url, MetricName AS name, MetricUnit AS unit, Attributes AS attributes, toString(toUnixTimestamp64Nano(StartTimeUnix)) AS start_ns, toString(toUnixTimestamp64Nano(TimeUnix)) AS end_ns, toString(Count) AS count, arrayMap(c -> toString(c),BucketCounts) AS bucket_counts, ExplicitBounds AS bounds, Flags AS flags, AggregationTemporality AS temporality FROM otel_metrics_histogram WHERE TimeUnix >= fromUnixTimestamp64Milli({baseline:Int64}) AND TimeUnix <= fromUnixTimestamp64Milli({to:Int64}) AND ServiceName={service:String} AND ResourceAttributes['deployment.environment.name']={environment:String} AND ResourceAttributes['service.instance.id']={instance:String} AND ScopeName={scope:String} AND ScopeVersion={version:String} AND MetricName={metric:String} ORDER BY TimeUnix LIMIT 1001 FORMAT JSONEachRow`
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
		q.Set("max_result_rows", "1001")
		q.Set("result_overflow_mode", "throw")
		address.RawQuery = q.Encode()
		response, e := sourcehttp.Post(ctx, client, address.String(), user, password, sql)
		if e != nil {
			return nil, e
		}
		data, oversize, e := sourcehttp.ReadBounded(response.Body, 8<<20)
		response.Body.Close()
		if e != nil {
			return nil, e
		}
		if response.StatusCode != 200 {
			return nil, fmt.Errorf("histogram source HTTP %d", response.StatusCode)
		}
		if oversize {
			return nil, fmt.Errorf("histogram source exceeds 8 MiB")
		}
		rows := []HistogramRow{}
		dec := json.NewDecoder(strings.NewReader(string(data)))
		for {
			var row HistogramRow
			e = dec.Decode(&row)
			if e == io.EOF {
				break
			}
			if e != nil {
				return nil, e
			}
			rows = append(rows, row)
			if len(rows) > 1000 {
				return nil, fmt.Errorf("histogram source exceeds 1000 points")
			}
		}
		evaluated, e := p.EvaluateHistogramRows(contract, from, to, rows)
		if e != nil {
			return nil, e
		}
		result.Histograms = append(result.Histograms, LiveHistogram{contract, evaluated})
	}
	return result, nil
}
