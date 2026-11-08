package measurements

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/internal/sourcefailure"
)

func TestNativeMeasurementRefusalClassification378(t *testing.T) {
	d := definition(t, "queue-lag")
	q := Question{Definition: d, FromMS: 100000, ToMS: 220000}
	valid := gaugeRow(d, "one", 210000, "2")
	invalid := valid
	invalid.Value = "secret-SQL-not-a-number"
	cases := []struct {
		name, body string
		status     int
		kind       sourcefailure.Kind
	}{
		{"source unavailable", "secret credentials SELECT statement", 503, sourcefailure.Unavailable},
		{"invalid JSON", "{secret", 200, sourcefailure.Malformed},
		{"nonobject JSON", "[]", 200, sourcefailure.Malformed},
		{"null JSON", "null", 200, sourcefailure.Malformed},
		{"invalid numeric protocol", string(raw(invalid)) + "\n", 200, sourcefailure.Malformed},
		{"global row budget", strings.Repeat("{}\n", RowLimit+1), 200, sourcefailure.Budget},
		{"payload budget", strings.Repeat(" ", 8<<20+1), 200, sourcefailure.Budget},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			s := Source{Endpoint: "http://private-source.invalid", Client: &http.Client{Transport: nativeWireTransport(func(r *http.Request) (*http.Response, error) {
				calls++
				r.Body.Close()
				return &http.Response{StatusCode: tc.status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(tc.body))}, nil
			})}}
			result, err := s.Run(t.Context(), q, nil)
			if err == nil || result != nil || calls != 1 || sourcefailure.Describe(err).Refusal.Kind != tc.kind {
				t.Fatal("native refusal kind/partial result", result, err, calls)
			}
			wire, _ := json.Marshal(sourcefailure.Describe(err))
			for _, unsafe := range []string{"secret", "credentials", "SELECT", "private-source"} {
				if strings.Contains(string(wire), unsafe) {
					t.Fatal("source detail leaked", string(wire))
				}
			}
			bad := q
			bad.ToMS = bad.FromMS
			if result, err = s.Run(t.Context(), bad, nil); err == nil || result != nil || calls != 1 {
				t.Fatal("invalid question reached source", result, err, calls)
			} else if _, typed := errors.AsType[*sourcefailure.Error](err); typed {
				t.Fatal("compile error reclassified as source failure", err)
			}
		})
	}
	// Pure supplied-data evaluation is not a source transport boundary.
	if _, err := Evaluate(q, []json.RawMessage{raw(invalid)}, time.Unix(230, 0)); err == nil {
		t.Fatal("invalid scalar accepted")
	} else if _, typed := errors.AsType[*sourcefailure.Error](err); typed {
		t.Fatal("pure kernel gained source failure policy")
	}
}
func TestNativeMeasurementFailureCauses378(t *testing.T) {
	d := definition(t, "queue-lag")
	q := Question{Definition: d, FromMS: 100000, ToMS: 220000}
	timeout := &net.DNSError{Err: "secret timeout", IsTimeout: true}
	for _, tc := range []struct {
		cause error
		kind  sourcefailure.Kind
	}{{context.Canceled, sourcefailure.Cancelled}, {context.DeadlineExceeded, sourcefailure.Timeout}, {timeout, sourcefailure.Timeout}, {errors.New("secret upstream error"), sourcefailure.Unavailable}} {
		source := Source{Endpoint: "http://private-source.invalid", Client: &http.Client{Transport: nativeWireTransport(func(r *http.Request) (*http.Response, error) { r.Body.Close(); return nil, tc.cause })}}
		result, err := source.Run(t.Context(), q, nil)
		if result != nil || !errors.Is(err, tc.cause) || sourcefailure.Describe(err).Refusal.Kind != tc.kind {
			t.Fatal("native cause/classification lost", result, err)
		}
		if tc.cause == timeout {
			if _, ok := errors.AsType[*net.DNSError](err); !ok {
				t.Fatal("network cause lost")
			}
		}
	}
	cause := errors.New("secret source body read failure")
	source := Source{Endpoint: "http://private-source.invalid", Client: &http.Client{Transport: nativeWireTransport(func(r *http.Request) (*http.Response, error) {
		r.Body.Close()
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(nativeWireFailure{cause})}, nil
	})}}
	if result, err := source.Run(t.Context(), q, nil); result != nil || !errors.Is(err, cause) || sourcefailure.Describe(err).Refusal.Kind != sourcefailure.Unavailable {
		t.Fatal("body failure cause lost", result, err)
	}
}
func TestNativeLifecycleCustodyFailureIsNotProtocol378(t *testing.T) {
	d := definition(t, "reservation-lifecycle")
	q := Question{Definition: d, FromMS: 100000, ToMS: 220000}
	row := raw(map[string]any{"body": d.Lifecycle.Accepted, "service": "lab-grpc", "resource": map[string]string{"deployment.environment.name": d.Environment}, "attributes": map[string]string{"goo.event.id": "e1", "goo.event.time_ns": "150000000000", "goo.subject.id": "subject", "goo.subject.tenant": "demo", "app.release": "r1"}})
	blocked := filepath.Join(t.TempDir(), "custody-is-a-file")
	if err := os.WriteFile(blocked, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	source := Source{Endpoint: "http://source.invalid", CustodyDir: blocked, Client: &http.Client{Transport: nativeWireTransport(func(r *http.Request) (*http.Response, error) {
		r.Body.Close()
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(string(row) + "\n"))}, nil
	})}}
	result, err := source.Run(t.Context(), q, nil)
	if result != nil || err == nil || sourcefailure.Describe(err).Refusal.Kind != sourcefailure.Unavailable {
		t.Fatal("custody failure became invalid source data", result, err)
	}
	if _, typed := errors.AsType[*sourcefailure.Error](err); typed {
		t.Fatal("lifecycle custody was wrapped by numeric protocol")
	}
	if _, path := errors.AsType[*os.PathError](err); !path {
		t.Fatal("filesystem cause lost", err)
	}
}
