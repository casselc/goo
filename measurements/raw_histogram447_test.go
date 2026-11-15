package measurements

import (
	"bytes"
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

func rawHistogramQuestion447() RawHistogramQuestion {
	return RawHistogramQuestion{Question: DiscoveryQuestion{Environment: "prod", Service: "api", FromMS: 1000, ToMS: 2000}, Instrument: RawHistogramInstrument{Kind: "histogram", Scope: "", Version: "", SchemaURL: "", Metric: "request.duration", Unit: "s", Temporality: 1}}
}
func rawHistogramRow447() json.RawMessage {
	return json.RawMessage(`{"resource":{"service.name":"api","deployment.environment.name":"prod"},"attributes":{"route":"/api"},"resource_schema_url":"","scope":"","version":"","schema_url":"","name":"request.duration","unit":"s","start_ns":"0","end_ns":"1500000000","flags":0,"temporality":1,"count":"9007199254740993","bounds":["1"],"bucket_counts":["9007199254740992","1"],"sum":"0","min":"0","max":"0"}`)
}
func changeHistogram447(t *testing.T, changes map[string]string) json.RawMessage {
	t.Helper()
	var f map[string]json.RawMessage
	if err := json.Unmarshal(rawHistogramRow447(), &f); err != nil {
		t.Fatal(err)
	}
	for k, v := range changes {
		if v == "" {
			delete(f, k)
		} else {
			f[k] = json.RawMessage(v)
		}
	}
	out, e := json.Marshal(f)
	if e != nil {
		t.Fatal(e)
	}
	return out
}
func TestRawHistogram447ExactFactsAndDuplicates(t *testing.T) {
	raw := rawHistogramRow447()
	r, err := rawHistogramRows(rawHistogramQuestion447(), []json.RawMessage{raw, raw}, time.Unix(10, 0))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Rows) != 2 || r.Rows[0].OccurrenceKey == r.Rows[1].OccurrenceKey || r.Rows[0].Count == nil || *r.Rows[0].Count != "9007199254740993" || r.Rows[0].BucketCounts[0] != "9007199254740992" || r.Rows[0].BucketsState != "valid" || r.Rows[0].RecordingState != "recorded" || !bytes.Equal(r.Rows[0].RawRow, raw) {
		t.Fatal("reported facts changed", r)
	}
	if r.Rows[0].Sum.PresenceState != "unavailable" || r.Rows[0].Sum.Value == nil || *r.Rows[0].Sum.Value != 0 {
		t.Fatal("source zero became producer presence", r)
	}
	raw[0] = ' '
	if r.Rows[0].RawRow[0] != '{' {
		t.Fatal("raw payload aliased caller")
	}
	p, e := CompileRawHistogramObservations(rawHistogramQuestion447())
	if e != nil || !strings.Contains(p.SQL, "FROM otel_metrics_histogram") || strings.Contains(p.SQL, "otel_metrics_exp_histogram") || !strings.Contains(p.SQL, "AggregationTemporality={p9:Int32}") || p.Parameters["p9"] != "1" || !strings.Contains(p.SQL, "arrayMap(b -> toString(b),ExplicitBounds)") {
		t.Fatal(p, e)
	}
}
func TestRawHistogram447PayloadDiagnosticsNotWholeReadRefusal(t *testing.T) {
	cases := []struct {
		name    string
		changes map[string]string
		state   string
	}{
		{"count-only", map[string]string{"bounds": "[]", "bucket_counts": "[]"}, "absent"},
		{"all-values", map[string]string{"bounds": "[]", "bucket_counts": `["9007199254740993"]`}, "valid"},
		{"shape", map[string]string{"bucket_counts": `["9007199254740993"]`}, "invalid"},
		{"overflow-sum", map[string]string{"count": `"0"`, "bucket_counts": `["18446744073709551615","1"]`}, "invalid"},
		{"count-invalid", map[string]string{"count": `"18446744073709551616"`}, "invalid"},
		{"bounds-nan", map[string]string{"bounds": `["nan"]`}, "invalid"},
		{"bounds-infinite", map[string]string{"bounds": `["+Inf"]`}, "invalid"},
		{"bounds-unordered", map[string]string{"bounds": `["2","1"]`, "bucket_counts": `["1","1","1"]`}, "invalid"},
		{"missing-count", map[string]string{"count": ""}, "absent"},
		{"null-bounds", map[string]string{"bounds": "null"}, "absent"},
		{"mixed-invalid-absent", map[string]string{"count": `"bad"`, "bounds": "null"}, "invalid"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := changeHistogram447(t, tc.changes)
			r, e := rawHistogramRows(rawHistogramQuestion447(), []json.RawMessage{raw}, time.Time{})
			if e != nil || len(r.Rows) != 1 || r.Rows[0].BucketsState != tc.state || !bytes.Equal(r.Rows[0].RawRow, raw) {
				t.Fatal(r, e)
			}
		})
	}
}
func TestRawHistogram447RecordingAndOptionalStatistics(t *testing.T) {
	for _, flags := range []string{"0", "1", "3", "8", "null", "1.5"} {
		t.Run(flags, func(t *testing.T) {
			r, e := rawHistogramRows(rawHistogramQuestion447(), []json.RawMessage{changeHistogram447(t, map[string]string{"flags": flags, "sum": `"NaN"`, "min": "null", "max": `"0"`})}, time.Time{})
			if e != nil {
				t.Fatal(e)
			}
			row := r.Rows[0]
			state := "unknown"
			if flags == "0" {
				state = "recorded"
			}
			if flags == "1" || flags == "3" {
				state = "no-recorded-value"
				if row.Count != nil || row.Bounds != nil || row.BucketCounts != nil || row.CountState != "absent" || row.BucketsState != "absent" || row.Max.Value != nil || row.Max.ValueState != "absent" {
					t.Fatal("flagged value promoted", row)
				}
			}
			if row.RecordingState != state || row.Sum.PresenceState != "unavailable" || row.Min.PresenceState != "unavailable" || row.Max.PresenceState != "unavailable" {
				t.Fatal(row)
			}
			if !slices.Contains(row.Diagnostics, RawObservationDiagnostic{Field: "sum", Code: "nonfinite"}) {
				t.Fatal("nonfinite source fact hidden", row)
			}
		})
	}
}
func TestRawHistogram447AtomicScopeAndBudgets(t *testing.T) {
	for _, change := range []map[string]string{{"temporality": "2"}, {"temporality": "null"}, {"end_ns": `"2000000000"`}, {"unit": `"ms"`}, {"resource": `{"service.name":"api","deployment.environment.name":"other"}`}} {
		r, e := rawHistogramRows(rawHistogramQuestion447(), []json.RawMessage{rawHistogramRow447(), changeHistogram447(t, change)}, time.Time{})
		if r != nil || e == nil {
			t.Fatal("partial scope admitted", r, e)
		}
	}
	q := rawHistogramQuestion447()
	q.Question.Environment = ""
	q.Question.EnvironmentSelector = &identity.EnvironmentSelector{Version: "goo/environment-selector/v1", Policy: "effective", Scope: identity.EnvironmentScope{Mode: "all"}}
	if q.Validate() == nil {
		t.Fatal("all environment admitted")
	}
	q = rawHistogramQuestion447()
	q.Instrument.Kind = "exponential-histogram"
	if q.Validate() == nil {
		t.Fatal("different histogram profile admitted")
	}
	many := make([]json.RawMessage, 10001)
	for i := range many {
		many[i] = rawHistogramRow447()
	}
	if r, e := rawHistogramRows(rawHistogramQuestion447(), many, time.Time{}); r != nil || sourcefailure.Describe(e).Refusal.Kind != sourcefailure.Budget {
		t.Fatal("row budget", e)
	}
	vector := "[" + strings.Repeat(`"1",`, 10000) + `"1"]`
	if r, e := rawHistogramRows(rawHistogramQuestion447(), []json.RawMessage{changeHistogram447(t, map[string]string{"bounds": vector})}, time.Time{}); r != nil || sourcefailure.Describe(e).Refusal.Kind != sourcefailure.Budget {
		t.Fatal("vector budget", e)
	}
}
func TestRawHistogram447OptionalSourceColumns(t *testing.T) {
	q := rawHistogramQuestion447()
	calls := 0
	source := Source{Endpoint: "http://source", Client: &http.Client{Transport: rawGaugeTransport438(func(r *http.Request) (*http.Response, error) {
		calls++
		body, _ := io.ReadAll(r.Body)
		sql := string(body)
		payload := ""
		switch {
		case strings.HasPrefix(sql, "DESCRIBE TABLE"):
			payload, _ = sourceCapabilityResponse431(sql)
		case strings.Contains(sql, "LIMIT 0"):
		default:
			if !strings.Contains(sql, "NULL AS sum, NULL AS min, NULL AS max") {
				t.Errorf("missing optional source columns referenced: %s", sql)
			}
			payload = string(changeHistogram447(t, map[string]string{"sum": "null", "min": "null", "max": "null"})) + "\n"
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(payload))}, nil
	})}}
	out, e := source.ReadRawHistogramObservations(t.Context(), q)
	if e != nil || calls != 3 || len(out.Rows) != 1 || out.Rows[0].Sum.ValueState != "absent" {
		t.Fatal(out, e, calls)
	}
}
