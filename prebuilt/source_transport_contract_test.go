package prebuilt

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

type prebuiltWireTransport func(*http.Request) (*http.Response, error)

func (f prebuiltWireTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type prebuiltWireBody struct {
	io.Reader
	reads, closes int
}

func (b *prebuiltWireBody) Read(p []byte) (int, error) { b.reads++; return b.Reader.Read(p) }
func (b *prebuiltWireBody) Close() error               { b.closes++; return nil }

type prebuiltWireFailure struct{ err error }

func (f prebuiltWireFailure) Read([]byte) (int, error) { return 0, f.err }

func TestPrebuiltTransportContract(t *testing.T) {
	e, row := sourceFixture(t)
	source := HTTPSource{ScopeName: "literal\\n\t'雪+%", ScopeVersion: "1", SchemaURL: "https://opentelemetry.io/schemas/1.37.0"}
	sql, params, err := e.HTTPSourceQuery(source, 100, 200)
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
		{name: "overflow bytes", status: 200, payload: strings.Repeat(" ", (8<<20)+1), want: "source payload exceeds 8 MiB; no partial evaluation"},
		{name: "row boundary", status: 200, payload: strings.Repeat(string(raw)+"\n", 1000)},
		{name: "row overflow", status: 200, payload: strings.Repeat(string(raw)+"\n", 1001), want: "source bound exceeded (1000 raw spans); narrow the window; partial population not evaluated"},
		{name: "malformed", status: 200, payload: "{", want: "unexpected EOF"},
		{name: "created", status: 201, broken: true, want: "ClickHouse source returned HTTP 201"},
		{name: "unavailable", status: 503, broken: true, want: "ClickHouse source returned HTTP 503"},
		{name: "read failure", status: 200, broken: true, want: "read failure"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var reader io.Reader = strings.NewReader(tc.payload)
			if tc.broken {
				reader = prebuiltWireFailure{failure}
			}
			body := &prebuiltWireBody{Reader: reader}
			client := &http.Client{Transport: prebuiltWireTransport(func(r *http.Request) (*http.Response, error) {
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
			result, err := FetchHTTP(context.Background(), client, "http://source/?database=otel&readonly=0", "u", "p", e, source, 100, 200)
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
			if tc.name == "duplicates" && result.DuplicateRecords != 1 {
				t.Fatal("dedup changed")
			}
		})
	}
}

func TestPrebuiltTransportCancellationAndErrors(t *testing.T) {
	e, _ := sourceFixture(t)
	source := HTTPSource{ScopeName: "test", ScopeVersion: "1", SchemaURL: "https://opentelemetry.io/schemas/1.37.0"}
	failure := errors.New("transport failure")
	for _, cancel := range []bool{false, true} {
		ctx, stop := context.WithCancel(context.Background())
		if cancel {
			stop()
		}
		client := &http.Client{Transport: prebuiltWireTransport(func(r *http.Request) (*http.Response, error) {
			if cancel {
				return nil, r.Context().Err()
			}
			return nil, failure
		})}
		result, err := FetchHTTP(ctx, client, "http://source", "u", "p", e, source, 100, 200)
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

func TestClickHousePrebuiltTransportIntegration(t *testing.T) {
	endpoint := os.Getenv("GOO_CLICKHOUSE_TEST_URL")
	if endpoint == "" {
		t.Skip("GOO_CLICKHOUSE_TEST_URL is not configured")
	}
	e, _ := sourceFixture(t)
	// Deliberately absent scope: exercises the actual table/query/evaluator without
	// inserting telemetry or attributing arbitrary live spans to a fixture binding.
	source := HTTPSource{ScopeName: "goo.transport.contract.absent", ScopeVersion: "1", SchemaURL: "https://opentelemetry.io/schemas/1.37.0"}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	result, err := FetchHTTP(ctx, &http.Client{Timeout: 15 * time.Second}, endpoint, os.Getenv("GOO_CLICKHOUSE_TEST_USER"), os.Getenv("GOO_CLICKHOUSE_TEST_PASSWORD"), e, source, 100, 200)
	if err != nil || result == nil {
		t.Fatal(err)
	}
	if len(result.Prepared.Facts) != 0 || result.DuplicateRecords != 0 || len(result.Warnings) == 0 {
		t.Fatal("unexpected fixture population")
	}
}
