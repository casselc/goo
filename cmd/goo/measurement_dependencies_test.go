package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/casselc/goo/measurements"
)

type nativeDependencyTransport func(*http.Request) (*http.Response, error)

func (f nativeDependencyTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestNativeMeasurementStartupDefaultsAndSnapshot(t *testing.T) {
	t.Setenv("CLICKHOUSE_URL", "")
	t.Setenv("CLICKHOUSE_USER", "fixture-user")
	t.Setenv("CLICKHOUSE_PASSWORD", "fixture-secret")
	client := &http.Client{}
	s := clickHouseConnectionFromEnvironment(client).measurementSource()
	if s.Endpoint != "http://localhost:8123" || s.User != "fixture-user" || s.Password != "fixture-secret" || s.Client != client {
		t.Fatal("startup defaults changed")
	}
	t.Setenv("CLICKHOUSE_URL", "http://example.invalid")
	t.Setenv("CLICKHOUSE_USER", "other")
	t.Setenv("CLICKHOUSE_PASSWORD", "other")
	if s.Endpoint != "http://localhost:8123" || s.User != "fixture-user" || s.Password != "fixture-secret" {
		t.Fatal("startup source mutated")
	}
}
func TestNativeMeasurementRegistrationsUseIndependentInjectedSources(t *testing.T) {
	wire, e := os.ReadFile("../../testdata/storage/native-measurement.yaml")
	if e != nil {
		t.Fatal(e)
	}
	d, e := measurements.Read(string(wire))
	if e != nil {
		t.Fatal(e)
	}
	var forbidden atomic.Int32
	trap := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forbidden.Add(1)
		http.Error(w, "unexpected ambient source", 500)
	}))
	defer trap.Close()
	t.Setenv("CLICKHOUSE_URL", trap.URL)
	t.Setenv("CLICKHOUSE_USER", "ambient")
	t.Setenv("CLICKHOUSE_PASSWORD", "ambient")
	for _, value := range []string{"2", "3"} {
		t.Run(value, func(t *testing.T) {
			var reads atomic.Int32
			backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				reads.Add(1)
				user, password, ok := r.BasicAuth()
				if !ok || user != "fixture-"+value || password != "local-only" || r.URL.Query().Get("readonly") != "1" {
					t.Error("injected authentication/readonly not preserved")
				}
				sql, _ := io.ReadAll(r.Body)
				if strings.Contains(string(sql), "UNION ALL") {
					io.WriteString(w, `{"kind":"gauge","scope":"ordinary.runtime","version":"","schemaURL":"","metric":"process.cpu.utilization","unit":"1","temporality":0,"monotonic":false,"dimensions":[]}`+"\n")
					return
				}
				json.NewEncoder(w).Encode(map[string]any{"resource": map[string]string{"service.name": "api", "service.instance.id": value, "deployment.environment.name": "prod"}, "scope": d.Source.Scope, "version": "", "schema_url": "", "name": d.Source.Metric, "unit": "1", "end_ns": "210000000000", "value": value})
			}))
			defer backend.Close()
			var clientCalls atomic.Int32
			client := backend.Client()
			client.Transport = nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
				clientCalls.Add(1)
				return http.DefaultTransport.RoundTrip(r)
			})
			root := t.TempDir()
			s := measurements.Source{Endpoint: backend.URL, User: "fixture-" + value, Password: "local-only", Client: client, CustodyDir: "ignored"}
			mux := http.NewServeMux()
			registerMeasurements(mux, s, root)
			s.Endpoint = trap.URL
			request := func(path string, v any) *httptest.ResponseRecorder {
				b, _ := json.Marshal(v)
				w := httptest.NewRecorder()
				mux.ServeHTTP(w, httptest.NewRequest("POST", path, strings.NewReader(string(b))))
				return w
			}
			// Planning stays source-free, even with live dependencies installed.
			r := request("/measurement-packages/plan", map[string]any{"yaml": string(wire), "fromMs": 100000, "toMs": 220000})
			if r.Code != 200 || reads.Load() != 0 {
				t.Fatal("planning executed source", r.Code)
			}
			r = request("/measurements", measurements.Question{Definition: d, FromMS: 100000, ToMS: 220000})
			var result measurements.Result
			if r.Code != 200 || json.Unmarshal(r.Body.Bytes(), &result) != nil || len(result.Rows) != 1 {
				t.Fatal("native execution", r.Code, r.Body.String())
			}
			var row struct{ Value string }
			json.Unmarshal(result.Rows[0], &row)
			if row.Value != value {
				t.Fatal("cross-registration source drift")
			}
			r = request("/metric-discovery", measurements.DiscoveryQuestion{Environment: "prod", Service: "api", FromMS: 100000, ToMS: 220000})
			if r.Code != 200 || reads.Load() != 2 {
				t.Fatal("native discovery dependency", r.Code, r.Body.String())
			}
			if clientCalls.Load() != 2 {
				t.Fatal("injected client bypassed")
			}
			invalid := measurements.Question{Definition: d, FromMS: 220000, ToMS: 100000}
			if r = request("/measurements", invalid); r.Code != 400 || reads.Load() != 2 {
				t.Fatal("invalid question reached source", r.Code)
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			body, _ := json.Marshal(measurements.Question{Definition: d, FromMS: 100000, ToMS: 220000})
			cancelled := httptest.NewRecorder()
			mux.ServeHTTP(cancelled, httptest.NewRequest("POST", "/measurements", strings.NewReader(string(body))).WithContext(ctx))
			// Cancellation may stop at body receive (400) or source execution (502).
			if (cancelled.Code != 400 && cancelled.Code != 502) || reads.Load() != 2 {
				t.Fatal("cancelled native request reached source", cancelled.Code)
			}
			if _, e := os.Stat(filepath.Join(root, "exploration", "custody")); !os.IsNotExist(e) {
				t.Fatal("gauge/discovery unexpectedly published lifecycle custody", e)
			}
		})
	}
	if forbidden.Load() != 0 {
		t.Fatal("handler consulted ambient source")
	}
}
