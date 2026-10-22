package measurements

import (
	"bufio"
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

type nativeWireTransport func(*http.Request) (*http.Response, error)

func (f nativeWireTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type nativeWireBody struct {
	io.Reader
	closed int
	read   int
}

func (b *nativeWireBody) Read(p []byte) (int, error) {
	n, e := b.Reader.Read(p)
	b.read += n
	return n, e
}
func (b *nativeWireBody) Close() error { b.closed++; return nil }

type nativeWireFailure struct{ err error }

func (r nativeWireFailure) Read([]byte) (int, error) { return 0, r.err }

func TestNativeSourceRequestWireContract(t *testing.T) {
	const sql = "SELECT {p0:String} FORMAT JSONEachRow"
	const value = "literal\\n\n\t\r\x00\b\f'雪+%"
	params := map[string]string{"p0": value}
	body := &nativeWireBody{Reader: strings.NewReader("{}\n")}
	calls := 0
	client := &http.Client{Transport: nativeWireTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		defer r.Body.Close()
		b, e := io.ReadAll(r.Body)
		if e != nil || string(b) != sql || r.Method != "POST" || r.URL.Path != "/query" {
			t.Fatal("query wire changed", e)
		}
		u, p, ok := r.BasicAuth()
		if !ok || u != "fixture-user" || p != "local-only" {
			t.Fatal("source auth changed")
		}
		q := r.URL.Query()
		want := map[string]string{"custom": "keep", "param_p0": "literal\\\\n\\n\\t\\r\\0\b\f'雪+%", "readonly": "1", "max_execution_time": "10", "max_result_rows": "10001", "result_overflow_mode": "throw", "max_bytes_to_read": "268435456", "read_overflow_mode": "throw"}
		if len(q) != len(want) {
			t.Fatal("unexpected query policy", q)
		}
		for k, v := range want {
			if q.Get(k) != v {
				t.Fatal("query parameter changed", k)
			}
		}
		return &http.Response{StatusCode: 200, Body: body, Header: make(http.Header)}, nil
	})}
	source := Source{Endpoint: "http://fixture.invalid/query?custom=keep&readonly=0&param_p0=old", User: "fixture-user", Password: "local-only", Client: client}
	rows, e := source.read(context.Background(), Plan{SQL: sql, Parameters: params})
	if e != nil || len(rows) != 1 || string(rows[0]) != "{}" || calls != 1 || body.closed != 1 || params["p0"] != value {
		t.Fatal("request roundtrip/ownership changed", e)
	}
}
func TestNativeSourceResponseWireContract(t *testing.T) {
	const limit = 8 << 20
	atLimit := strings.Repeat(strings.Repeat("x", 1023)+"\n", 8192)
	failure := errors.New("fixture body read failure")
	for _, c := range []struct {
		name    string
		status  int
		body    string
		rows    int
		message string
		cause   error
	}{{"empty", 200, "", 0, "", nil}, {"raw lines", 200, "not json\n\n{}\r\n", 3, "", nil}, {"exact row count", 200, strings.Repeat("{}\n", RowLimit), RowLimit, "", nil}, {"row overflow", 200, strings.Repeat("{}\n", RowLimit+1), 0, "source exceeds 10000 rows; no partial measurement", nil}, {"exact byte limit", 200, atLimit, 8192, "", nil}, {"byte overflow", 200, atLimit + "x", 0, "bounded measurement source HTTP 200 / payload limit", nil}, {"single line scanner limit", 200, strings.Repeat("x", limit), 0, bufio.ErrTooLong.Error(), nil}, {"other success status", 201, "{}\n", 0, "bounded measurement source HTTP 201 / payload limit", nil}, {"non-200", 503, "backend detail", 0, "bounded measurement source HTTP 503 / payload limit", nil}, {"read error", 200, "{}\n", 0, failure.Error(), failure}, {"read error precedes status", 503, "backend detail", 0, failure.Error(), failure}} {
		t.Run(c.name, func(t *testing.T) {
			var reader io.Reader = strings.NewReader(c.body)
			if c.cause != nil {
				reader = io.MultiReader(reader, nativeWireFailure{c.cause})
			}
			body := &nativeWireBody{Reader: reader}
			client := &http.Client{Transport: nativeWireTransport(func(r *http.Request) (*http.Response, error) {
				r.Body.Close()
				return &http.Response{StatusCode: c.status, Body: body, Header: make(http.Header)}, nil
			})}
			rows, e := (Source{Endpoint: "http://fixture.invalid", Client: client}).read(context.Background(), Plan{SQL: "SELECT 1"})
			if c.message == "" {
				if e != nil || rows == nil || len(rows) != c.rows {
					t.Fatal("success changed", len(rows), e)
				}
			} else if e == nil || e.Error() != c.message || rows != nil {
				t.Fatal("failure/partial policy changed", len(rows), e)
			}
			if body.closed != 1 || body.read > limit+1 {
				t.Fatal("body lifetime/bound changed", body.closed, body.read)
			}
			if c.cause != nil && !errors.Is(e, c.cause) {
				t.Fatal("read error identity changed", e)
			}
		})
	}
}
func TestNativeSourceTransportErrorsAndCancellation(t *testing.T) {
	failure := errors.New("fixture transport failed")
	s := Source{Endpoint: "http://fixture.invalid", Client: &http.Client{Transport: nativeWireTransport(func(r *http.Request) (*http.Response, error) {
		r.Body.Close()
		if r.Context().Err() != nil {
			return nil, r.Context().Err()
		}
		return nil, failure
	})}}
	rows, e := s.read(context.Background(), Plan{})
	if !errors.Is(e, failure) || rows != nil {
		t.Fatal("transport failure changed", e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	rows, e = s.read(ctx, Plan{})
	if !errors.Is(e, context.Canceled) || rows != nil {
		t.Fatal("cancellation lost", e)
	}
}

func TestClickHouseNativeSourceTransportIntegration(t *testing.T) {
	address := os.Getenv("GOO_CLICKHOUSE_TEST_URL")
	if address == "" {
		t.Skip("optional real ClickHouse instance required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	source := Source{Endpoint: address, User: os.Getenv("GOO_CLICKHOUSE_TEST_USER"), Password: os.Getenv("GOO_CLICKHOUSE_TEST_PASSWORD")}
	const value = "literal\\n\n\t\r\x00'雪+%"
	rows, e := source.read(ctx, Plan{SQL: "SELECT {p0:String} AS message, toString(number) AS n FROM numbers(3) ORDER BY number FORMAT JSONEachRow", Parameters: map[string]string{"p0": value}})
	if e != nil || len(rows) != 3 {
		t.Fatal("real bounded source", len(rows), e)
	}
	for i, row := range rows {
		var got struct {
			Message string
			N       string
		}
		if json.Unmarshal(row, &got) != nil || got.Message != value || got.N != strconv.Itoa(i) {
			t.Fatal("real parameter/row bytes changed", i)
		}
	}
	wire, e := os.ReadFile("../testdata/storage/native-measurement.yaml")
	if e != nil {
		t.Fatal(e)
	}
	definition, e := Read(string(wire))
	if e != nil {
		t.Fatal(e)
	}
	now := time.Now().UnixMilli()
	question := Question{Definition: definition, FromMS: now - 60000, ToMS: now}
	result, e := source.Run(ctx, question, nil)
	if e != nil || result == nil {
		t.Fatal("real native query/evaluation", e)
	}
	if e = ValidateResult(result); e != nil {
		t.Fatal("real native result identity", e)
	}
	if _, e = source.Discover(ctx, DiscoveryQuestion{Environment: definition.Environment, Service: definition.Source.Services[0], FromMS: now - 60000, ToMS: now}); e != nil {
		t.Fatal("real native discovery", e)
	}
}
