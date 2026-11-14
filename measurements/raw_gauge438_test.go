package measurements

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/internal/sourcefailure"
)

func rawGaugeQuestion438() RawGaugeQuestion {
	return RawGaugeQuestion{Question: DiscoveryQuestion{Environment: "prod", Service: "api", FromMS: 1000, ToMS: 2000}, Instrument: RawGaugeInstrument{Kind: "gauge", Scope: "", Version: "", SchemaURL: "", Metric: "temperature", Unit: "Cel"}}
}
func rawGaugeRow438() json.RawMessage {
	return json.RawMessage(`{"resource":{"service.name":"api","deployment.environment.name":"prod"},"attributes":{"room":"a"},"resource_schema_url":"","scope":"","version":"","schema_url":"","name":"temperature","unit":"Cel","start_ns":"0","end_ns":"1500000000","value":"12.5","flags":0}`)
}
func mutateRawGauge438(t *testing.T, edit func(map[string]json.RawMessage)) json.RawMessage {
	t.Helper()
	var fields map[string]json.RawMessage
	json.Unmarshal(rawGaugeRow438(), &fields)
	edit(fields)
	raw, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
func TestRawGauge438ConservesOccurrences(t *testing.T) {
	q := rawGaugeQuestion438()
	raw := rawGaugeRow438()
	r, err := rawGaugeRows(q, []json.RawMessage{raw, raw}, time.Unix(10, 0))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Rows) != 2 || r.Rows[0].OccurrenceKey == r.Rows[1].OccurrenceKey || r.ReadAt != "1970-01-01T00:00:10Z" || r.Rows[0].ValueState != "finite" || *r.Rows[0].Value != 12.5 || !bytes.Equal(r.Rows[0].RawRow, raw) || r.Rows[0].Stream.Resource["service.instance.id"] != "" {
		t.Fatalf("raw facts changed: %+v", r)
	}
	raw[0] = 'x'
	if r.Rows[0].RawRow[0] != '{' {
		t.Fatal("raw input aliased")
	}
	q.Question.Service = "changed"
	if r.Question.Service != "api" {
		t.Fatal("question aliased")
	}
	empty, err := rawGaugeRows(rawGaugeQuestion438(), nil, time.Unix(10, 0))
	if err != nil || empty.Rows == nil || len(empty.Rows) != 0 {
		t.Fatal(empty, err)
	}
}
func TestRawGauge438ScalarDiagnostics(t *testing.T) {
	cases := []struct{ name, field, raw, state, code string }{
		{"null", "value", "null", "absent", "absent"}, {"missing", "value", "", "absent", "absent"}, {"object", "value", "{}", "invalid", "invalid"}, {"boolean", "value", "false", "invalid", "invalid"}, {"NaN", "value", `"NaN"`, "nonfinite", "nonfinite"}, {"infinity", "value", `"+Inf"`, "nonfinite", "nonfinite"}, {"overflow", "value", `"1e999"`, "nonfinite", "nonfinite"}, {"zero", "value", `"0"`, "finite", ""}, {"numeric", "value", "1.25", "finite", ""}, {"no-recorded", "flags", "1", "absent", "no-recorded-value"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			raw := mutateRawGauge438(t, func(fields map[string]json.RawMessage) {
				if c.raw == "" {
					delete(fields, c.field)
				} else {
					fields[c.field] = json.RawMessage(c.raw)
				}
			})
			r, err := rawGaugeRows(rawGaugeQuestion438(), []json.RawMessage{raw}, time.Time{})
			if err != nil {
				t.Fatal(err)
			}
			row := r.Rows[0]
			if row.ValueState != c.state || c.state != "finite" && row.Value != nil {
				t.Fatalf("state/value: %+v", row)
			}
			if c.code != "" && !slices.Contains(row.Diagnostics, RawObservationDiagnostic{Field: "value", Code: c.code}) {
				t.Fatal(row.Diagnostics)
			}
		})
	}
	for _, flags := range []string{"null", `"0"`, "-1", "4294967296", "2"} {
		raw := mutateRawGauge438(t, func(fields map[string]json.RawMessage) { fields["flags"] = json.RawMessage(flags) })
		r, err := rawGaugeRows(rawGaugeQuestion438(), []json.RawMessage{raw}, time.Time{})
		if err != nil || len(r.Rows[0].Diagnostics) == 0 {
			t.Fatal(r, err)
		}
	}
	for _, start := range []string{"null", `"oops"`, `"1600000000"`, "12"} {
		raw := mutateRawGauge438(t, func(fields map[string]json.RawMessage) { fields["start_ns"] = json.RawMessage(start) })
		r, err := rawGaugeRows(rawGaugeQuestion438(), []json.RawMessage{raw}, time.Time{})
		if err != nil || r.Rows[0].StartState == "valid" {
			t.Fatal(r, err)
		}
	}
}
func TestRawGauge438AtomicScopeRefusal(t *testing.T) {
	for _, c := range []struct{ field, raw string }{{"end_ns", "null"}, {"end_ns", `"0999999999"`}, {"end_ns", `"2000000000"`}, {"end_ns", `"-1"`}, {"end_ns", "123"}, {"unit", `"K"`}, {"scope", `"other"`}, {"resource", `{"service.name":"api"}`}, {"attributes", `{"x":null}`}, {"attributes", "null"}, {"version", "null"}} {
		t.Run(c.field+c.raw, func(t *testing.T) {
			raw := mutateRawGauge438(t, func(fields map[string]json.RawMessage) { fields[c.field] = json.RawMessage(c.raw) })
			r, err := rawGaugeRows(rawGaugeQuestion438(), []json.RawMessage{rawGaugeRow438(), raw}, time.Time{})
			if err == nil || r != nil {
				t.Fatal("published partial", r, err)
			}
		})
	}
	r, err := rawGaugeRows(rawGaugeQuestion438(), make([]json.RawMessage, RowLimit+1), time.Time{})
	if r != nil || sourcefailure.Describe(err).Refusal.Kind != sourcefailure.Budget {
		t.Fatal(r, err)
	}
	q := rawGaugeQuestion438()
	selector := identity.EnvironmentSelector{Version: "goo/environment-selector/v1", Policy: "effective", Scope: identity.EnvironmentScope{Mode: "all"}}
	q.Question.Environment = ""
	q.Question.EnvironmentSelector = &selector
	if q.Validate() == nil {
		t.Fatal("All native scope admitted")
	}
}

type rawGaugeTransport438 func(*http.Request) (*http.Response, error)

func (f rawGaugeTransport438) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestRawGauge438Transport(t *testing.T) {
	reads := 0
	q := rawGaugeQuestion438()
	q.Instrument.Metric = "temperature' OR 1=1"
	row := bytes.ReplaceAll(rawGaugeRow438(), []byte(`"temperature"`), []byte(`"temperature' OR 1=1"`))
	source := Source{Endpoint: "http://source", Client: &http.Client{Transport: rawGaugeTransport438(func(req *http.Request) (*http.Response, error) {
		reads++
		sql, _ := io.ReadAll(req.Body)
		response, handled := sourceCapabilityResponse431(string(sql))
		if !handled {
			if !strings.Contains(string(sql), "FROM otel_metrics_gauge") || strings.Contains(string(sql), q.Instrument.Metric) || !strings.Contains(string(sql), "LIMIT 10001") || req.URL.Query().Get("readonly") != "1" {
				t.Fatal("unsafe source plan", string(sql))
			}
			response = string(row) + "\n"
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(response)), Header: make(http.Header)}, nil
	})}}
	result, err := source.ReadRawGaugeObservations(t.Context(), q)
	if err != nil || len(result.Rows) != 1 || reads != 3 {
		t.Fatal(result, err, reads)
	}
	q.Instrument.Kind = "counter"
	if _, err = source.ReadRawGaugeObservations(t.Context(), q); err == nil || reads != 3 {
		t.Fatal("unsupported kind reached source")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	q.Instrument.Kind = "gauge"
	if _, err = source.ReadRawGaugeObservations(ctx, q); err == nil || reads != 3 {
		t.Fatal("cancelled read reached source", err, reads)
	}
}

func TestRawGauge438FlagsSemanticInteger(t *testing.T) {
	for _, raw := range []string{"0.0", "1e0", "4294967295.000", "0e20"} {
		if _, ok := rawObservationFlags(json.RawMessage(raw)); !ok {
			t.Fatal("semantic uint32 refused", raw)
		}
	}
	for _, raw := range []string{"0.1", "-1", "4294967296", "4294967295.0000001", `"0"`, "true"} {
		if _, ok := rawObservationFlags(json.RawMessage(raw)); ok {
			t.Fatal("invalid flags admitted", raw)
		}
	}
}

func TestRawGauge438SourceBudgets(t *testing.T) {
	for _, kind := range []string{"rows", "input-bytes", "output-bytes", "projection-failure"} {
		t.Run(kind, func(t *testing.T) {
			reads := 0
			source := Source{Endpoint: "http://source", Client: &http.Client{Transport: rawGaugeTransport438(func(req *http.Request) (*http.Response, error) {
				reads++
				sql, _ := io.ReadAll(req.Body)
				body, handled := sourceCapabilityResponse431(string(sql))
				status := 200
				if !handled {
					switch kind {
					case "rows":
						body = strings.Repeat(string(rawGaugeRow438())+"\n", RowLimit+1)
					case "input-bytes":
						body = `{"padding":"` + strings.Repeat("x", 8<<20) + `"}` + "\n"
					case "output-bytes":
						raw := mutateRawGauge438(t, func(fields map[string]json.RawMessage) {
							attributes, _ := json.Marshal(map[string]string{"large": strings.Repeat("x", (4<<20)+1024)})
							fields["attributes"] = attributes
						})
						body = string(raw) + "\n"
					case "projection-failure":
						status = 403
						body = "denied"
					}
				}
				return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
			})}}
			result, err := source.ReadRawGaugeObservations(t.Context(), rawGaugeQuestion438())
			want := sourcefailure.Budget
			if kind == "projection-failure" {
				want = sourcefailure.Unavailable
			}
			if result != nil || err == nil || sourcefailure.Describe(err).Refusal.Kind != want || reads != 3 {
				t.Fatal(result, err, reads)
			}
		})
	}
}
