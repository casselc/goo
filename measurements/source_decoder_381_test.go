package measurements

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/casselc/goo/internal/sourcefailure"
)

// Exercise the public source boundary through its real private reader, rather
// than sharing the extracted decoder's framing logic or its error sentinels.
func TestNativeSourceFraming381(t *testing.T) {
	const limit = 10000
	numeric := ` {"n":9007199254740993,"negativeZero":-0,"largeExponent":1e400,"n":2} `
	cases := []struct {
		name, body string
		rows       int
		kind       sourcefailure.Kind
	}{
		{"empty", "", 0, ""},
		{"object with CRLF", numeric + "\r\n", 1, ""},
		{"two JSON values", "{} {}\n", 0, sourcefailure.Malformed},
		{"garbage tail", "{}x\n", 0, sourcefailure.Malformed},
		{"array", "[]\n", 0, sourcefailure.Malformed},
		{"null", "null\n", 0, sourcefailure.Malformed},
		{"scalar", "1\n", 0, sourcefailure.Malformed},
		{"broken second row", "{}\n{", 0, sourcefailure.Malformed},
		{"max admitted rows", strings.Repeat("{}\n", limit), limit, ""},
		{"row overflow", strings.Repeat("{}\n", limit+1), 0, sourcefailure.Budget},
		{"blank after max rows", strings.Repeat("{}\n", limit) + "\n", 0, sourcefailure.Budget},
		{"malformed overflow", strings.Repeat("{}\n", limit) + "{", 0, sourcefailure.Budget},
		{"blank rows", "\n \t\r\n", 0, sourcefailure.Malformed},
		{"unicode outer whitespace", "\u00a0{}\u00a0", 0, sourcefailure.Malformed},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			body := &nativeWireBody{Reader: strings.NewReader(c.body)}
			source := Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: nativeWireTransport(func(r *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 200, Header: make(http.Header), Body: body}, nil
			})}}
			rows, err := source.read(context.Background(), Plan{SQL: "SELECT 1 FORMAT JSONEachRow"})
			if body.closed != 1 {
				t.Fatalf("body closed %d times", body.closed)
			}
			if c.kind != "" {
				if err == nil || rows != nil || sourcefailure.Describe(err).Refusal.Kind != c.kind {
					t.Fatalf("admitted partial/wrong failure: rows=%d err=%v", len(rows), err)
				}
				return
			}
			if err != nil || rows == nil || len(rows) != c.rows {
				t.Fatalf("rows=%d err=%v", len(rows), err)
			}
			if c.name == "object with CRLF" && string(rows[0]) != numeric+"\r" {
				t.Fatalf("raw numeric/whitespace representation changed: %q", rows[0])
			}
		})
	}
}

func TestNativeSourceByteLimitAndCauses381(t *testing.T) {
	const capBytes = 8 << 20
	prefix, suffix := `{"padding":"`, `"}`
	exact := prefix + strings.Repeat("a", capBytes-len(prefix)-len(suffix)) + suffix
	for _, extra := range []bool{false, true} {
		t.Run(map[bool]string{false: "exact", true: "excess"}[extra], func(t *testing.T) {
			payload := exact
			if extra {
				payload += " "
			}
			body := &nativeWireBody{Reader: strings.NewReader(payload)}
			source := Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: nativeWireTransport(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 200, Header: make(http.Header), Body: body}, nil
			})}}
			rows, err := source.read(context.Background(), Plan{SQL: "SELECT 1 FORMAT JSONEachRow"})
			if extra {
				if err == nil || rows != nil || sourcefailure.Describe(err).Refusal.Kind != sourcefailure.Budget {
					t.Fatalf("oversized body admitted: rows=%d err=%v", len(rows), err)
				}
			} else if err != nil || len(rows) != 1 || len(rows[0]) != capBytes {
				t.Fatalf("exact bound rejected: rows=%d err=%v", len(rows), err)
			}
		})
	}
	for _, cause := range []error{errors.New("reader failure"), context.Canceled, context.DeadlineExceeded} {
		body := &nativeWireBody{Reader: nativeFailure381{cause}}
		source := Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: nativeWireTransport(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Header: make(http.Header), Body: body}, nil
		})}}
		rows, err := source.read(context.Background(), Plan{SQL: "SELECT 1 FORMAT JSONEachRow"})
		if !errors.Is(err, cause) || rows != nil || body.closed != 1 {
			t.Fatalf("read cause/partial/close changed: rows=%d err=%v", len(rows), err)
		}
	}
	body := &nativeWireBody{Reader: nativeFailure381{context.Canceled}}
	source := Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: nativeWireTransport(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 503, Header: make(http.Header), Body: body}, nil
	})}}
	rows, err := source.read(context.Background(), Plan{SQL: "SELECT 1 FORMAT JSONEachRow"})
	if err == nil || rows != nil || body.read != 0 || body.closed != 1 || sourcefailure.Describe(err).Refusal.Kind != sourcefailure.Unavailable {
		t.Fatal("HTTP refusal read/reclassified its body", err)
	}
}

// Return both bytes and an error: buffered acquisition must preserve the error,
// never admit the syntactically valid partial object.
type nativeFailure381 struct{ err error }

func (r nativeFailure381) Read(p []byte) (int, error) { return copy(p, []byte("{}")), r.err }

var _ io.Reader = nativeFailure381{}
