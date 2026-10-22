package annotations

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func histogramTransportFixture(t *testing.T) (*OperationPlan, HistogramRow) {
	t.Helper()
	p, err := LoadOperationPackage("../config/operations/checkout.yaml")
	if err != nil {
		t.Fatal(err)
	}
	definition := p.Definition
	second := definition.Histograms[0]
	second.Name = "duration-copy"
	definition.Histograms = append(definition.Histograms, second)
	p, err = definition.Bind(p.Manifest)
	if err != nil {
		t.Fatal(err)
	}
	contract := p.Definition.Histograms[0]
	base := CounterRow{Resource: map[string]string{"service.name": p.Source.Service, "service.instance.id": p.Source.Instance, "deployment.environment.name": p.Source.Environment}, Scope: p.Manifest.Scope, Version: p.Manifest.ScopeVersion, SchemaURL: p.Manifest.SchemaURL, Name: contract.Metric, Unit: contract.Unit, Attributes: map[string]string{"goo.operation.name": "checkout.reserve", "goo.operation.role": "server", "goo.operation.outcome": "ok"}, StartNS: "1000000", EndNS: "100000000", Temporality: 2}
	buckets := make([]string, len(contract.Bounds)+1)
	for i := range buckets {
		buckets[i] = "0"
	}
	buckets[1] = "9007199254740993"
	return p, HistogramRow{CounterRow: base, Bounds: contract.Bounds, BucketCounts: buckets, Count: "9007199254740993"}
}

func TestHistogramTransportContract(t *testing.T) {
	p, row := histogramTransportFixture(t)
	// Change the declaration and rebind: request parameters must remain unescaped
	// without invalidating the plan's identity checks.
	definition := p.Definition
	definition.Resource.Service = "literal\\n\t'雪+%"
	var err error
	p, err = definition.Bind(p.Manifest)
	if err != nil {
		t.Fatal(err)
	}
	row.Resource["service.name"] = p.Source.Service
	raw, _ := json.Marshal(row)
	later := row
	later.EndNS = "200000000"
	later.Count = "9007199254740994"
	later.BucketCounts = append([]string(nil), row.BucketCounts...)
	later.BucketCounts[1] = later.Count
	last, _ := json.Marshal(later)
	failure := errors.New("histogram read failure")
	cases := []struct {
		name, payload, want string
		status              int
		broken              bool
	}{
		{name: "empty", status: 200},
		{name: "valid delta", status: 200, payload: string(raw) + "\n" + string(last)},
		{name: "exact bytes", status: 200, payload: string(raw) + strings.Repeat(" ", (8<<20)-len(raw))},
		{name: "overflow bytes", status: 200, payload: strings.Repeat(" ", (8<<20)+1), want: "histogram source exceeds 8 MiB"},
		{name: "row boundary", status: 200, payload: strings.Repeat(string(raw)+"\n", 1000)},
		{name: "row overflow", status: 200, payload: strings.Repeat(string(raw)+"\n", 1001), want: "histogram source exceeds 1000 points"},
		{name: "malformed", status: 200, payload: "{", want: "unexpected EOF"},
		{name: "created", status: 201, payload: "ignored", want: "histogram source HTTP 201"},
		{name: "unavailable", status: 503, payload: "ignored", want: "histogram source HTTP 503"},
		{name: "status beats overflow", status: 503, payload: strings.Repeat(" ", (8<<20)+1), want: "histogram source HTTP 503"},
		{name: "read failure", status: 200, broken: true, want: "histogram read failure"},
		{name: "read beats status", status: 503, broken: true, want: "histogram read failure"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var bodies []*operationWireBody
			client := &http.Client{Transport: operationWireTransport(func(r *http.Request) (*http.Response, error) {
				for _, b := range bodies {
					if b.closes != 1 {
						t.Fatal("previous response not closed before next contract")
					}
				}
				index := len(bodies)
				contract := p.Definition.Histograms[index]
				got, _ := io.ReadAll(r.Body)
				u, password, ok := r.BasicAuth()
				if r.Method != "POST" || string(got) != histogramWireSQL || !ok || u != "u" || password != "p" {
					t.Fatal("request contract", string(got))
				}
				q := r.URL.Query()
				expected := map[string]string{"database": "otel", "readonly": "1", "max_execution_time": "10", "max_result_rows": "1001", "result_overflow_mode": "throw", "param_from": "100", "param_to": "200", "param_service": p.Source.Service, "param_environment": p.Source.Environment, "param_instance": p.Source.Instance, "param_scope": p.Manifest.Scope, "param_version": p.Manifest.ScopeVersion, "param_metric": contract.Metric}
				gap, _ := time.ParseDuration(contract.MaxGap)
				expected["param_baseline"] = strconv.FormatInt(100-int64(gap/time.Millisecond), 10)
				for k, v := range expected {
					if q.Get(k) != v {
						t.Fatal(k, q.Get(k), v)
					}
				}
				if len(q) != len(expected) {
					t.Fatal("unexpected settings", q)
				}
				var reader io.Reader = strings.NewReader(tc.payload)
				if tc.broken {
					reader = operationWireFailure{failure}
				}
				b := &operationWireBody{Reader: reader}
				bodies = append(bodies, b)
				return &http.Response{StatusCode: tc.status, Body: b, Header: make(http.Header)}, nil
			})}
			result, err := p.FetchHistograms(context.Background(), client, "http://source/?database=otel&readonly=0", "u", "p", 100, 200)
			for _, b := range bodies {
				if b.closes != 1 || b.reads == 0 {
					t.Fatal("read/close policy", b.reads, b.closes)
				}
			}
			if tc.want != "" {
				if err == nil || err.Error() != tc.want || result != nil || len(bodies) != 1 {
					t.Fatal(result, err, len(bodies))
				}
				if tc.broken && !errors.Is(err, failure) {
					t.Fatal("error identity")
				}
				return
			}
			if err != nil || result == nil || len(result.Histograms) != len(p.Definition.Histograms) || len(bodies) != len(p.Definition.Histograms) {
				t.Fatal(result, err)
			}
			if tc.name == "valid delta" {
				r := result.Histograms[0].Result
				if len(r.Series) != 1 || r.Series[0].Intervals[0].Distribution.Count != "1" || r.Basis == "" {
					t.Fatal("histogram semantics changed", r)
				}
			}
		})
	}
}

func TestHistogramTransportCancellationAndErrors(t *testing.T) {
	p, _ := histogramTransportFixture(t)
	failure := errors.New("histogram transport failure")
	for _, cancel := range []bool{false, true} {
		ctx, stop := context.WithCancel(context.Background())
		if cancel {
			stop()
		}
		client := &http.Client{Transport: operationWireTransport(func(r *http.Request) (*http.Response, error) {
			if cancel {
				return nil, r.Context().Err()
			}
			return nil, failure
		})}
		result, err := p.FetchHistograms(ctx, client, "http://source", "u", "p", 100, 200)
		stop()
		want := failure
		if cancel {
			want = context.Canceled
		}
		if result != nil || !errors.Is(err, want) {
			t.Fatal(result, err)
		}
	}
}

func TestClickHouseHistogramTransportIntegration(t *testing.T) {
	endpoint := os.Getenv("GOO_CLICKHOUSE_TEST_URL")
	if endpoint == "" {
		t.Skip("GOO_CLICKHOUSE_TEST_URL is not configured")
	}
	p, _ := histogramTransportFixture(t)
	definition := p.Definition
	definition.Resource.Service = "goo.transport.contract.absent"
	p, err := definition.Bind(p.Manifest)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	result, err := p.FetchHistograms(ctx, &http.Client{Timeout: 15 * time.Second}, endpoint, os.Getenv("GOO_CLICKHOUSE_TEST_USER"), os.Getenv("GOO_CLICKHOUSE_TEST_PASSWORD"), 100, 200)
	if err != nil || result == nil {
		t.Fatal(err)
	}
	if len(result.Histograms) != len(p.Definition.Histograms) || len(result.Warnings) == 0 {
		t.Fatal("missing contracts")
	}
	for _, rate := range result.Histograms {
		if len(rate.Result.Series) != 0 || rate.Result.Basis == "" {
			t.Fatal("unexpected fixture population", rate.Result)
		}
	}
}

const histogramWireSQL = `SELECT mapUpdate(ResourceAttributes,map('service.name',toString(ServiceName))) AS resource, ScopeName AS scope, ScopeVersion AS version, ScopeSchemaUrl AS schema_url, MetricName AS name, MetricUnit AS unit, Attributes AS attributes, toString(toUnixTimestamp64Nano(StartTimeUnix)) AS start_ns, toString(toUnixTimestamp64Nano(TimeUnix)) AS end_ns, toString(Count) AS count, arrayMap(c -> toString(c),BucketCounts) AS bucket_counts, ExplicitBounds AS bounds, Flags AS flags, AggregationTemporality AS temporality FROM otel_metrics_histogram WHERE TimeUnix >= fromUnixTimestamp64Milli({baseline:Int64}) AND TimeUnix <= fromUnixTimestamp64Milli({to:Int64}) AND ServiceName={service:String} AND ResourceAttributes['deployment.environment.name']={environment:String} AND ResourceAttributes['service.instance.id']={instance:String} AND ScopeName={scope:String} AND ScopeVersion={version:String} AND MetricName={metric:String} ORDER BY TimeUnix LIMIT 1001 FORMAT JSONEachRow`

func TestHistogramTransportLaterContractFailure(t *testing.T) {
	p, _ := histogramTransportFixture(t)
	calls := 0
	var bodies []*operationWireBody
	client := &http.Client{Transport: operationWireTransport(func(*http.Request) (*http.Response, error) {
		calls++
		status := 200
		if calls == 2 {
			status = 503
		}
		b := &operationWireBody{Reader: strings.NewReader("")}
		bodies = append(bodies, b)
		return &http.Response{StatusCode: status, Body: b, Header: make(http.Header)}, nil
	})}
	result, err := p.FetchHistograms(context.Background(), client, "http://source", "u", "p", 100, 200)
	if result != nil || err == nil || err.Error() != "histogram source HTTP 503" || calls != 2 {
		t.Fatal("partial contract success escaped", result, err, calls)
	}
	for _, b := range bodies {
		if b.closes != 1 {
			t.Fatal("unclosed response")
		}
	}
}
