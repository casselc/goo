package measurements

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/internal/sourcefailure"
)

func rawSumQuestion441(tempo int32, mono bool) RawSumQuestion {
	kind := "counter"
	if tempo == 2 && !mono {
		kind = "stock"
	}
	return RawSumQuestion{Question: DiscoveryQuestion{Environment: "prod", Service: "api", FromMS: 1000, ToMS: 2000}, Instrument: RawSumInstrument{Kind: kind, Scope: "", Version: "", SchemaURL: "", Metric: "jobs.total", Unit: "{job}", Temporality: tempo, Monotonic: mono}}
}
func rawSumRow441(tempo int32, mono bool) json.RawMessage {
	return json.RawMessage(`{"resource":{"service.name":"api","deployment.environment.name":"prod"},"attributes":{"queue":"a"},"resource_schema_url":"","scope":"","version":"","schema_url":"","name":"jobs.total","unit":"{job}","start_ns":"0","end_ns":"1500000000","value":"12.5","flags":0,"temporality":` + strconv.FormatInt(int64(tempo), 10) + `,"monotonic":` + strconv.FormatBool(mono) + `}`)
}
func TestRawSum441TupleAndOccurrenceConservation(t *testing.T) {
	for _, tempo := range []int32{-2147483648, -1, 0, 1, 2, 3, 2147483647} {
		for _, mono := range []bool{false, true} {
			q := rawSumQuestion441(tempo, mono)
			raw := rawSumRow441(tempo, mono)
			r, err := rawSumRows(q, []json.RawMessage{raw, raw}, time.Unix(10, 0))
			if err != nil {
				t.Fatal(err)
			}
			if r.Version != "goo/raw-sum-observations/v1" || len(r.Rows) != 2 || r.Rows[0].OccurrenceKey == r.Rows[1].OccurrenceKey || r.Rows[0].Temporality != tempo || r.Rows[0].Monotonic != mono || !bytes.Equal(r.Rows[0].RawRow, raw) || r.Rows[0].Stream.Resource["service.instance.id"] != "" {
				t.Fatal("reported facts lost", r)
			}
			if tempo != 1 && tempo != 2 {
				code := "unrecognized"
				if tempo == 0 {
					code = "unspecified"
				}
				if !slices.Contains(r.Rows[0].Diagnostics, RawObservationDiagnostic{Field: "temporality", Code: code}) {
					t.Fatal("unknown temporality hidden", r)
				}
			}
			p, err := CompileRawSumObservations(q)
			if err != nil || !strings.Contains(p.SQL, "FROM otel_metrics_sum") || strings.Contains(p.SQL, "otel_metrics_gauge") || !strings.Contains(p.SQL, "AggregationTemporality={p9:Int32}") || !strings.Contains(p.SQL, "IsMonotonic={p10:Bool}") || p.Parameters["p9"] != strconv.FormatInt(int64(tempo), 10) || p.Parameters["p10"] != strconv.FormatBool(mono) {
				t.Fatal(p, err)
			}
			q.Instrument.Kind = "gauge"
			if q.Validate() == nil {
				t.Fatal("wrong kind admitted")
			}
		}
	}
}
func TestRawSum441StructuralMetadataRefusesAtomically(t *testing.T) {
	for _, change := range []struct{ key, raw string }{{"temporality", ""}, {"temporality", "null"}, {"temporality", `"1"`}, {"temporality", "1.1"}, {"temporality", "2147483648"}, {"temporality", "2"}, {"monotonic", ""}, {"monotonic", "null"}, {"monotonic", `"true"`}, {"monotonic", "1"}, {"monotonic", "false"}, {"end_ns", `"2000000000"`}} {
		var fields map[string]json.RawMessage
		json.Unmarshal(rawSumRow441(1, true), &fields)
		if change.raw == "" {
			delete(fields, change.key)
		} else {
			fields[change.key] = json.RawMessage(change.raw)
		}
		bad, _ := json.Marshal(fields)
		r, err := rawSumRows(rawSumQuestion441(1, true), []json.RawMessage{rawSumRow441(1, true), bad}, time.Time{})
		if err == nil || r != nil {
			t.Fatal("partial or unsupported identity published", change, r, err)
		}
	}
}
func TestRawSum441ExplicitRequestMetadata(t *testing.T) {
	good, _ := json.Marshal(rawSumQuestion441(2, false))
	var q RawSumQuestion
	if err := json.Unmarshal(good, &q); err != nil || q.Validate() != nil {
		t.Fatal(q, err)
	}
	for _, key := range []string{"kind", "scope", "version", "schemaURL", "metric", "unit", "temporality", "monotonic"} {
		var fields map[string]json.RawMessage
		json.Unmarshal(good, &fields)
		var instrument map[string]json.RawMessage
		json.Unmarshal(fields["instrument"], &instrument)
		delete(instrument, key)
		fields["instrument"], _ = json.Marshal(instrument)
		bad, _ := json.Marshal(fields)
		if json.Unmarshal(bad, &q) == nil {
			t.Fatal("omitted exact tuple field admitted", key)
		}
	}
	q = rawSumQuestion441(2, false)
	q.Instrument.Kind = "counter"
	if q.Validate() == nil {
		t.Fatal("stock tuple mislabeled counter")
	}
	q = rawSumQuestion441(1, false)
	q.Instrument.Kind = "stock"
	if q.Validate() == nil {
		t.Fatal("delta tuple mislabeled stock")
	}
	q = rawSumQuestion441(1, true)
	q.Question.Environment = ""
	q.Question.EnvironmentSelector = &identity.EnvironmentSelector{Version: "goo/environment-selector/v1", Policy: "effective", Scope: identity.EnvironmentScope{Mode: "all"}}
	if q.Validate() == nil {
		t.Fatal("wide native scope admitted")
	}
}
func TestRawSum441ScalarDiagnostics(t *testing.T) {
	q := rawSumQuestion441(1, true)
	for _, c := range []struct{ value, flags, state string }{{`"0"`, "0", "finite"}, {`"NaN"`, "0", "nonfinite"}, {"null", "0", "absent"}, {"{}", "0", "invalid"}, {`"0"`, "1", "absent"}} {
		var fields map[string]json.RawMessage
		json.Unmarshal(rawSumRow441(1, true), &fields)
		fields["value"] = json.RawMessage(c.value)
		fields["flags"] = json.RawMessage(c.flags)
		raw, _ := json.Marshal(fields)
		r, err := rawSumRows(q, []json.RawMessage{raw}, time.Time{})
		if err != nil || r.Rows[0].ValueState != c.state || c.state != "finite" && r.Rows[0].Value != nil {
			t.Fatal(r, err)
		}
		if c.flags == "1" && !slices.Contains(r.Rows[0].Diagnostics, RawObservationDiagnostic{Field: "value", Code: "no-recorded-value"}) {
			t.Fatal("flagged zero appears recorded")
		}
	}
}
func TestRawSum441SourceBoundaries(t *testing.T) {
	for _, mode := range []string{"success", "rows", "input-bytes", "output-bytes", "projection-failure", "metadata-mismatch"} {
		t.Run(mode, func(t *testing.T) {
			reads := 0
			source := Source{Endpoint: "http://source", Client: &http.Client{Transport: rawGaugeTransport438(func(req *http.Request) (*http.Response, error) {
				reads++
				sql, _ := io.ReadAll(req.Body)
				if strings.Contains(string(sql), "otel_metrics_gauge") {
					t.Fatal("sum read accessed gauge source")
				}
				body, handled := sourceCapabilityResponse431(string(sql))
				status := 200
				if !handled {
					body = string(rawSumRow441(1, true)) + "\n"
					switch mode {
					case "rows":
						body = strings.Repeat(body, RowLimit+1)
					case "input-bytes":
						body = `{"padding":"` + strings.Repeat("x", 8<<20) + `"}`
					case "output-bytes":
						body = strings.Replace(body, `"queue":"a"`, `"queue":"`+strings.Repeat("x", (4<<20)+1024)+`"`, 1)
					case "projection-failure":
						status = 403
						body = "denied"
					case "metadata-mismatch":
						body = strings.Replace(body, `"temporality":1`, `"temporality":2`, 1)
					}
				}
				return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
			})}}
			q := rawSumQuestion441(1, true)
			r, err := source.ReadRawSumObservations(t.Context(), q)
			if mode == "success" {
				if err != nil || len(r.Rows) != 1 || reads != 3 {
					t.Fatal(r, err, reads)
				}
			} else {
				want := sourcefailure.Budget
				if mode == "projection-failure" {
					want = sourcefailure.Unavailable
				}
				if mode == "metadata-mismatch" {
					want = sourcefailure.Malformed
				}
				if r != nil || err == nil || sourcefailure.Describe(err).Refusal.Kind != want || reads != 3 {
					t.Fatal(r, err, reads)
				}
			}
			q.Instrument.Kind = "gauge"
			if _, err = source.ReadRawSumObservations(t.Context(), q); err == nil || reads != 3 {
				t.Fatal("invalid profile reached source")
			}
			q = rawSumQuestion441(1, true)
			ctx, cancel := context.WithCancel(t.Context())
			cancel()
			if _, err = source.ReadRawSumObservations(ctx, q); err == nil || reads != 3 {
				t.Fatal("canceled request reached source")
			}
		})
	}
}
