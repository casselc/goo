package annotations

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

type operationWireTransport func(*http.Request) (*http.Response, error)

func (f operationWireTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type operationWireBody struct {
	io.Reader
	reads, closes int
}

func (b *operationWireBody) Read(p []byte) (int, error) { b.reads++; return b.Reader.Read(p) }
func (b *operationWireBody) Close() error               { b.closes++; return nil }

type operationWireFailure struct{ err error }

func (f operationWireFailure) Read([]byte) (int, error) { return 0, f.err }

func TestOperationTransportContract(t *testing.T) {
	m, source, row := operationTransportFixture(t)
	source.Service = "literal\\n\t'雪+%"
	row.Resource["service.name"] = source.Service
	sql, params, err := m.OperationQuery(source, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(row)
	failure := errors.New("read failure")
	cases := []struct {
		name, payload, want string
		status              int
		broken              bool
	}{
		{name: "empty", status: 200},
		{name: "valid", status: 200, payload: string(raw)},
		{name: "duplicates", status: 200, payload: string(raw) + "\n" + string(raw)},
		{name: "exact bytes", status: 200, payload: string(raw) + strings.Repeat(" ", (8<<20)-len(raw))},
		{name: "overflow bytes", status: 200, payload: strings.Repeat(" ", (8<<20)+1), want: "source exceeds 8 MiB; no partial totals"},
		{name: "row boundary", status: 200, payload: strings.Repeat(string(raw)+"\n", 1000)},
		{name: "row overflow", status: 200, payload: strings.Repeat(string(raw)+"\n", 1001), want: "source exceeds 1000 raw spans; narrow window"},
		{name: "malformed", status: 200, payload: "{", want: "unexpected EOF"},
		{name: "created", status: 201, broken: true, want: "ClickHouse source returned HTTP 201"},
		{name: "unavailable", status: 503, broken: true, want: "ClickHouse source returned HTTP 503"},
		{name: "read failure", status: 200, broken: true, want: "read failure"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var reader io.Reader = strings.NewReader(tc.payload)
			if tc.broken {
				reader = operationWireFailure{failure}
			}
			body := &operationWireBody{Reader: reader}
			client := &http.Client{Transport: operationWireTransport(func(r *http.Request) (*http.Response, error) {
				got, _ := io.ReadAll(r.Body)
				u, p, ok := r.BasicAuth()
				if r.Method != "POST" || string(got) != sql || !ok || u != "u" || p != "p" {
					t.Fatal("request contract")
				}
				q := r.URL.Query()
				for k, v := range map[string]string{"database": "otel", "readonly": "1", "max_execution_time": "10", "max_result_rows": "1001", "result_overflow_mode": "throw"} {
					if q.Get(k) != v {
						t.Fatal(k, q.Get(k))
					}
				}
				for k, v := range params {
					if q.Get("param_"+k) != v {
						t.Fatal("raw parameter changed", k)
					}
				}
				if len(q) != len(params)+5 {
					t.Fatal("unexpected query settings", q)
				}
				return &http.Response{StatusCode: tc.status, Body: body, Header: make(http.Header)}, nil
			})}
			result, err := FetchOperations(context.Background(), client, "http://source/?database=otel&readonly=0", "u", "p", *m, source, 1, 2)
			if body.closes != 1 {
				t.Fatal("body closure", body.closes)
			}
			if tc.status != 200 && body.reads != 0 {
				t.Fatal("error body consumed")
			}
			if tc.want != "" {
				if err == nil || err.Error() != tc.want || result != nil {
					t.Fatal(result, err)
				}
				if tc.status == 200 && tc.broken && !errors.Is(err, failure) {
					t.Fatal("error identity")
				}
				return
			}
			if err != nil || result == nil {
				t.Fatal(err)
			}
			if result.Query != sql || !reflect.DeepEqual(result.Parameters, params) {
				t.Fatal("receipt query changed")
			}
			if tc.name == "valid" && (len(result.Spans) != 1 || result.Spans[0].DurationNS != row.DurationNS) {
				t.Fatal("lost exact duration")
			}
			if tc.name == "duplicates" && result.DuplicateRecords != 1 {
				t.Fatal("dedup changed")
			}
		})
	}
}

func TestOperationTransportCancellationAndErrors(t *testing.T) {
	m, source, _ := operationTransportFixture(t)
	failure := errors.New("transport failure")
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
		result, err := FetchOperations(ctx, client, "http://source", "u", "p", *m, source, 1, 2)
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

func TestClickHouseOperationTransportIntegration(t *testing.T) {
	endpoint := os.Getenv("GOO_CLICKHOUSE_TEST_URL")
	if endpoint == "" {
		t.Skip("GOO_CLICKHOUSE_TEST_URL is not configured")
	}
	m, source, _ := operationTransportFixture(t)
	// Deliberately absent service: exercises the actual table/query/evaluator without
	// inserting telemetry or attributing arbitrary live spans to a fixture binding.
	source.Service = "goo.transport.contract.absent"
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	result, err := FetchOperations(ctx, &http.Client{Timeout: 15 * time.Second}, endpoint, os.Getenv("GOO_CLICKHOUSE_TEST_USER"), os.Getenv("GOO_CLICKHOUSE_TEST_PASSWORD"), *m, source, 1, 2)
	if err != nil || result == nil {
		t.Fatal(err)
	}
	if len(result.Spans) != 0 || result.DuplicateRecords != 0 || len(result.Conformance.Warnings) == 0 {
		t.Fatal("unexpected fixture population")
	}
}

func operationTransportFixture(t *testing.T) (*Manifest, OperationSource, OperationRow) {
	t.Helper()
	m, err := ScanDir("../examples/operation-app")
	if err != nil {
		t.Fatal(err)
	}
	source := OperationSource{Service: "lab-go-checkout", Environment: "local", Instance: "woven-demo", DeclaredSchema: m.SchemaURL}
	row := OperationRow{TimestampNS: "1000000", DurationNS: "9007199254740993", TraceID: strings.Repeat("1", 32), SpanID: strings.Repeat("2", 16), Kind: "Server", Scope: m.Scope, Version: m.ScopeVersion, Resource: map[string]string{"service.name": source.Service, "deployment.environment.name": source.Environment, "service.instance.id": source.Instance}, Attributes: map[string]string{"goo.operation.name": "checkout.reserve", "goo.operation.role": "server", "goo.operation.outcome": "error"}}
	return m, source, row
}
