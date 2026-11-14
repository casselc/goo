package exploration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

type capabilityTransport431 struct {
	mode  string
	reads []string
}

func (c *capabilityTransport431) RoundTrip(r *http.Request) (*http.Response, error) {
	body, _ := io.ReadAll(r.Body)
	sql := string(body)
	c.reads = append(c.reads, sql)
	status, out, header := 200, "", http.Header{}
	if schema, ok := sourceCapabilityResponse431(sql); ok {
		out = schema
	}
	switch {
	case strings.HasPrefix(sql, "DESCRIBE") && c.mode == "denied":
		status = 403
		header.Set("X-ClickHouse-Exception-Code", "497")
		out = "secret endpoint credentials SELECT"
	case strings.HasPrefix(sql, "DESCRIBE") && c.mode == "absent":
		status = 404
		header.Set("X-ClickHouse-Exception-Code", "60")
	case strings.HasPrefix(sql, "SELECT 1 AS observed"):
		switch c.mode {
		case "empty":
		case "malformed":
			out = `{"observed":2}`
		case "race":
			status = 500
		default:
			out = `{"observed":1}`
		}
	}
	return &http.Response{StatusCode: status, Header: header, Body: io.NopCloser(strings.NewReader(out))}, nil
}
func TestSourceCapabilitiesPresenceAndReadiness431(t *testing.T) {
	for _, mode := range []string{"observed", "empty", "denied", "absent", "malformed", "race"} {
		t.Run(mode, func(t *testing.T) {
			transport := &capabilityTransport431{mode: mode}
			s := Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: transport}}
			result, err := s.DiscoverSourceCapabilities(context.Background(), SourceCapabilitiesQuestion{FromMS: 1000, ToMS: 2000})
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Families) != 6 {
				t.Fatal("missing profile states")
			}
			for _, f := range result.Families {
				wantState, wantObservation := "compatible", mode
				switch mode {
				case "denied":
					wantState, wantObservation = "unreadable", "unknown"
				case "absent":
					wantState, wantObservation = "absent", "unknown"
				case "malformed", "race":
					wantObservation = "unreadable"
				}
				if f.State != wantState || f.Observation.State != wantObservation {
					t.Fatalf("%s: %+v", mode, f)
				}
				if strings.Contains(f.Detail, "secret") {
					t.Fatal("unsafe detail")
				}
			}
			if mode == "denied" || mode == "absent" {
				if len(transport.reads) != 6 {
					t.Fatal("inspected denied observations")
				}
			} else if len(transport.reads) != 18 {
				t.Fatal("unbounded/repeated probes", len(transport.reads))
			}
			encoded, _ := json.Marshal(result)
			if strings.Contains(string(encoded), `"count"`) {
				t.Fatal("presence became a total")
			}
		})
	}
}
func TestSourceCapabilitiesValidationBeforeRead431(t *testing.T) {
	transport := &capabilityTransport431{}
	s := Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: transport}}
	for _, q := range []SourceCapabilitiesQuestion{{}, {FromMS: -1, ToMS: 1}, {FromMS: 1, ToMS: 1}, {FromMS: 1, ToMS: 120*60000 + 2}} {
		if _, err := s.DiscoverSourceCapabilities(context.Background(), q); err == nil {
			t.Fatal(fmt.Sprint("accepted", q))
		}
	}
	if len(transport.reads) != 0 {
		t.Fatal("invalid request read source")
	}
}
