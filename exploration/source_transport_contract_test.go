package exploration

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

type recordWireTransport func(*http.Request) (*http.Response, error)

func (f recordWireTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type recordWireBody struct {
	io.Reader
	reads, closes int
}

func (b *recordWireBody) Read(p []byte) (int, error) { b.reads++; return b.Reader.Read(p) }
func (b *recordWireBody) Close() error               { b.closes++; return nil }

type recordWireFailure struct{ err error }

func (f recordWireFailure) Read([]byte) (int, error) { return 0, f.err }

func TestRecordSourceRequestWireContract(t *testing.T) {
	value := "\\\x00\t\n\r\x1f\x7f雪+%'"
	params := map[string]string{"p0": value}
	client := &http.Client{Transport: recordWireTransport(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		user, password, ok := r.BasicAuth()
		if r.Method != "POST" || string(body) != "SELECT {p0:String}\nFORMAT JSONEachRow" || !ok || user != "u" || password != "p" {
			t.Fatal("request changed", r.Method, string(body), ok)
		}
		expected := map[string]string{"database": "otel", "readonly": "1", "cancel_http_readonly_queries_on_client_close": "1", "max_execution_time": "10", "max_result_rows": "1002", "result_overflow_mode": "throw", "max_memory_usage": "268435456", "max_bytes_to_read": "268435456", "read_overflow_mode": "throw", "param_p0": `\\\x00\x09\x0a\x0d\x1f\x7f雪+%'`}
		for key, want := range expected {
			if got := r.URL.Query().Get(key); got != want {
				t.Errorf("%s: %q != %q", key, got, want)
			}
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("{}\n")), Header: make(http.Header)}, nil
	})}
	_, err := (Source{Endpoint: "http://source/?database=otel&readonly=0", User: "u", Password: "p", Client: client}).read(context.Background(), "SELECT {p0:String}\nFORMAT JSONEachRow", params)
	if err != nil || params["p0"] != value {
		t.Fatal(err, params)
	}
}

func TestRecordSourceResponseWireContract(t *testing.T) {
	failure := errors.New("body failure")
	exactLine := "{\"x\":\"" + strings.Repeat("a", (1<<20)-9) + "\"}\n"
	if len(exactLine) != 1<<20 {
		t.Fatal(len(exactLine))
	}
	cases := []struct {
		name, payload, wantError string
		status, wantRows         int
		broken, scanner          bool
	}{
		{name: "empty", status: 200},
		{name: "trimmed objects", status: 200, payload: " \r\n {\"n\":18446744073709551615} \r\n\t{}\n", wantRows: 2},
		{name: "array", status: 200, payload: "{}\n[]", wantError: "invalid source row"},
		{name: "null", status: 200, payload: "null", wantError: "invalid source row"},
		{name: "scalar", status: 200, payload: "42", wantError: "invalid source row"},
		{name: "malformed", status: 200, payload: "{", wantError: "invalid source row"},
		{name: "row boundary", status: 200, payload: strings.Repeat("{}\n", 1001), wantRows: 1001},
		{name: "row overflow", status: 200, payload: strings.Repeat("{}\n", 1002), wantError: "source row budget exceeded"},
		{name: "byte boundary", status: 200, payload: strings.Repeat(exactLine, 8), wantRows: 8},
		{name: "byte overflow", status: 200, payload: strings.Repeat(" ", (8<<20)+1), wantError: "source exceeds 8 MiB"},
		{name: "scanner boundary", status: 200, payload: strings.Repeat(" ", 8<<20), scanner: true},
		{name: "created", status: 201, broken: true, wantError: "exploration source HTTP 201"},
		{name: "unavailable", status: 503, broken: true, wantError: "exploration source HTTP 503"},
		{name: "read failure", status: 200, broken: true, wantError: "body failure"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var reader io.Reader = strings.NewReader(tc.payload)
			if tc.broken {
				reader = recordWireFailure{failure}
			}
			body := &recordWireBody{Reader: reader}
			client := &http.Client{Transport: recordWireTransport(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tc.status, Body: body, Header: make(http.Header)}, nil
			})}
			rows, err := (Source{Endpoint: "http://source", Client: client}).read(context.Background(), "SELECT", nil)
			if body.closes != 1 {
				t.Fatal("body closure", body.closes)
			}
			if tc.status != 200 && body.reads != 0 {
				t.Fatal("error body consumed")
			}
			if tc.scanner {
				if !errors.Is(err, bufio.ErrTooLong) || rows == nil {
					t.Fatal("scanner contract", rows, err)
				}
				return
			}
			if tc.wantError != "" {
				if err == nil || err.Error() != tc.wantError || rows != nil {
					t.Fatal(rows, err)
				}
				if tc.status == 200 && tc.broken && !errors.Is(err, failure) {
					t.Fatal("lost error identity")
				}
				return
			}
			if err != nil || rows == nil || len(rows) != tc.wantRows {
				t.Fatal(len(rows), err)
			}
			if tc.name == "trimmed objects" && string(rows[0]) != `{"n":18446744073709551615}` {
				t.Fatal(string(rows[0]))
			}
		})
	}
}

func TestRecordSourceTransportErrorsAndCancellation(t *testing.T) {
	failure := errors.New("transport failure")
	for _, cancel := range []bool{false, true} {
		ctx, stop := context.WithCancel(context.Background())
		if cancel {
			stop()
		}
		client := &http.Client{Transport: recordWireTransport(func(r *http.Request) (*http.Response, error) {
			if cancel {
				return nil, r.Context().Err()
			}
			return nil, failure
		})}
		rows, err := (Source{Endpoint: "http://source", Client: client}).read(ctx, "SELECT", nil)
		stop()
		want := failure
		if cancel {
			want = context.Canceled
		}
		if rows != nil || !errors.Is(err, want) {
			t.Fatal(rows, err)
		}
	}
}

func TestClickHouseRecordSourceTransportIntegration(t *testing.T) {
	endpoint := os.Getenv("GOO_CLICKHOUSE_TEST_URL")
	if endpoint == "" {
		t.Skip("GOO_CLICKHOUSE_TEST_URL is not configured")
	}
	source := Source{Endpoint: endpoint, User: os.Getenv("GOO_CLICKHOUSE_TEST_USER"), Password: os.Getenv("GOO_CLICKHOUSE_TEST_PASSWORD")}
	value := "backslash\\ quote' 雪+%"
	for i := 0; i < 32; i++ {
		value += string(rune(i))
	}
	value += string(rune(127))
	rows, err := source.read(context.Background(), "SELECT {p0:String} AS message, toString(number) AS n FROM numbers(3) FORMAT JSONEachRow", map[string]string{"p0": value})
	if err != nil || len(rows) != 3 {
		t.Fatal(len(rows), err)
	}
	for i, row := range rows {
		var got struct{ Message, N string }
		if err := json.Unmarshal(row, &got); err != nil || got.Message != value || got.N != fmt.Sprint(i) {
			t.Fatal(string(row), err)
		}
	}
}
