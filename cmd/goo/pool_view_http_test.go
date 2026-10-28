package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
)

// Held facts are produced by the real native evaluator at a fixed receipt clock,
// so pure composition fixtures have stable, genuinely validated replay bases.
func poolViewHTTPResult(t *testing.T, q measurements.Question, stream metrics.Stream) *measurements.Result {
	t.Helper()
	rows := []json.RawMessage{}
	for _, text := range strings.Split(strings.TrimSpace(stockHTTPRows(2, false)), "\n") {
		var row map[string]any
		_ = json.Unmarshal([]byte(text), &row)
		row["resource"] = stream.Resource
		row["attributes"] = stream.Attributes
		row["resource_schema_url"] = stream.ResourceSchemaURL
		row["scope"] = stream.Scope
		row["version"] = stream.Version
		row["schema_url"] = stream.SchemaURL
		row["name"] = stream.Name
		row["unit"] = stream.Unit
		if row["value"] == "7.5" {
			row["value"] = "7"
		}
		if q.Definition.Kind == "histogram" {
			row["count"] = "2"
			row["bounds"] = []float64{0.1}
			row["bucket_counts"] = []string{"1", "1"}
			row["start_ns"] = "100000000000"
			row["end_ns"] = "102500000000"
		}
		raw, _ := json.Marshal(row)
		rows = append(rows, raw)
		if q.Definition.Kind == "histogram" {
			break
		}
	}
	result, err := measurements.Evaluate(q, rows, time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if err := measurements.ValidateResult(result); err != nil {
		t.Fatal(err)
	}
	return result
}
func poolViewHTTPRequest(t *testing.T) measurements.PoolViewRequest {
	t.Helper()
	input := poolSignalsHTTPRequest()
	input.Stream.Attributes["db.client.connection.state"] = "used"
	plan, err := measurements.PlanPoolSignals(input)
	if err != nil || len(plan.Capacity) != 1 || len(plan.Wait) != 1 {
		t.Fatal("pool view fixture plan", err)
	}
	capq, waitq := plan.Capacity[0].Question, plan.Wait[0].Question
	return measurements.PoolViewRequest{Count: *poolViewHTTPResult(t, input.Parent, input.Stream), Stream: input.Stream, Capacity: poolViewHTTPResult(t, capq, *capq.Definition.Source.ExactStream), Wait: poolViewHTTPResult(t, waitq, *waitq.Definition.Source.ExactStream)}
}
func poolViewBody(t *testing.T, change func(*measurements.PoolViewRequest)) string {
	t.Helper()
	request := poolViewHTTPRequest(t)
	if change != nil {
		change(&request)
	}
	raw, _ := json.Marshal(request)
	return string(raw)
}
func poolViewWireResponse(t *testing.T, c planningWireCase) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	registerPoolView(mux)
	body := c.Body + c.Suffix
	if c.PadTo > len(body) {
		body += strings.Repeat(" ", c.PadTo-len(body))
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(c.Method, c.Path, strings.NewReader(body)))
	return w
}
func TestPoolViewHTTPFrozenContracts(t *testing.T) {
	full := poolViewBody(t, nil)
	cases := []planningWireCase{
		{Name: "validated-count-capacity-wait", Body: full, Status: 200},
		{Name: "count-only-unavailable", Body: poolViewBody(t, func(r *measurements.PoolViewRequest) { r.Capacity = nil; r.Wait = nil }), Status: 200},
		{Name: "wrong-anchor-identity", Body: poolViewBody(t, func(r *measurements.PoolViewRequest) { r.Stream.Resource["service.instance.id"] = "other" }), Status: 400},
		{Name: "different-valid-capacity-window", Body: poolViewBody(t, func(r *measurements.PoolViewRequest) {
			q := r.Capacity.Question
			q.ToMS++
			r.Capacity = poolViewHTTPResult(t, q, *q.Definition.Source.ExactStream)
		}), Status: 400},
		{Name: "different-valid-capacity-identity", Body: poolViewBody(t, func(r *measurements.PoolViewRequest) {
			q := r.Capacity.Question
			s := *q.Definition.Source.ExactStream
			s.Resource["service.instance.id"] = "other"
			q.Definition.Source.ExactStream = &s
			r.Capacity = poolViewHTTPResult(t, q, s)
		}), Status: 400},
		{Name: "forged-result-basis", Body: poolViewBody(t, func(r *measurements.PoolViewRequest) { r.Count.Basis = strings.Repeat("0", 64) }), Status: 400},
		{Name: "forged-evaluated-value", Body: poolViewBody(t, func(r *measurements.PoolViewRequest) {
			value := 123.0
			r.Count.Stocks.Series[0].Samples[0].Value = &value
		}), Status: 400},
		{Name: "unknown-field", Body: strings.TrimSuffix(full, "}") + `,"sql":"SELECT 1"}`, Status: 400},
		{Name: "wrong-count-type", Body: `{"count":true}`, Status: 400},
		{Name: "trailing-object", Body: full, Suffix: ` {}`, Status: 400},
		{Name: "exact-2MiB", Body: full, PadTo: 2 << 20, Status: 200},
		{Name: "oversize", Body: full, PadTo: (2 << 20) + 1, Status: 400},
		{Name: "method", Method: "GET", Status: 405},
	}
	for i := range cases {
		cases[i].Path = "/pool-signals/view"
		if cases[i].Method == "" {
			cases[i].Method = "POST"
		}
	}
	const filename = "../../testdata/http/pool-view.json"
	if os.Getenv("GOO_UPDATE_POOL_VIEW_HTTP_FIXTURES") == "1" {
		for i := range cases {
			w := poolViewWireResponse(t, cases[i])
			if w.Code != cases[i].Status {
				t.Fatal(cases[i].Name, w.Code, w.Body.String())
			}
			cases[i].Response = w.Body.String()
			cases[i].Headers = map[string]string{}
			for _, k := range []string{"Content-Type", "Cache-Control", "Allow", "X-Content-Type-Options"} {
				cases[i].Headers[k] = w.Header().Get(k)
			}
		}
		out, _ := json.MarshalIndent(planningWireFixtures{Version: 1, Cases: cases}, "", "  ")
		if err := os.WriteFile(filename, append(out, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
	fixture := dashboardProfileFixture(t, filename, len(cases))
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			w := poolViewWireResponse(t, c)
			if w.Code != c.Status || w.Body.String() != c.Response {
				t.Fatal(w.Code, w.Body.String())
			}
			for k, v := range c.Headers {
				if w.Header().Get(k) != v {
					t.Fatal("header changed", k)
				}
			}
		})
	}
}
func TestPoolViewHTTPPureAdmissionAndCancellation(t *testing.T) {
	mux := http.NewServeMux()
	registerPoolView(mux)
	budget := newQueryBudget(mux, 0)
	body := poolViewBody(t, nil)
	r := httptest.NewRequest("POST", "/pool-signals/view", strings.NewReader(body))
	r.Header.Set("X-Goo-Query-ID", "not-a-source-query-id")
	w := httptest.NewRecorder()
	budget.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatal("composition occupied source owner", w.Code, w.Body.String())
	}
	var result measurements.PoolView
	if json.Unmarshal(w.Body.Bytes(), &result) != nil || result.Version != "goo/pool-view/v1" || result.Utilization.State != "observed" || result.Utilization.Numerator != "1" || result.Utilization.Denominator != "1" {
		t.Fatal("wrong certified rational view", w.Body.String())
	}
	for _, expired := range []bool{false, true} {
		var ctx context.Context
		var cancel context.CancelFunc
		if expired {
			ctx, cancel = context.WithDeadline(context.Background(), time.Unix(0, 0))
		} else {
			ctx, cancel = context.WithCancel(context.Background())
			cancel()
		}
		request := httptest.NewRequest("POST", "/pool-signals/view", strings.NewReader(body)).WithContext(ctx)
		w = httptest.NewRecorder()
		mux.ServeHTTP(w, request)
		cancel()
		if w.Code != 400 {
			t.Fatal("cancelled/expired composition returned view", w.Code)
		}
	}
}

func TestPoolViewHTTPAvailableInGenericRuntimeWithoutPublication(t *testing.T) {
	dir := t.TempDir()
	server, err := configuredHTTPServer([]string{"-profile", "generic", "-investigations-dir", dir})
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	server.Handler.ServeHTTP(w, httptest.NewRequest("POST", "/pool-signals/view", strings.NewReader(poolViewBody(t, nil))))
	if w.Code != 200 {
		t.Fatal("generic runtime omitted pure composer", w.Code, w.Body.String())
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 0 {
		t.Fatal("pure composition created custody/publication files", err, entries)
	}
}
