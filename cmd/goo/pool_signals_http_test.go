package main

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func poolSignalsHTTPRequest() measurements.PoolSignalsRequest {
	parent := stockHTTPQuestion()
	parent.Definition.StockDomain = "nonnegative-integer"
	parent.Definition.Source.SchemaURL = measurements.PoolConnectionsSchema
	stream := metrics.Stream{Resource: map[string]string{"service.name": "api", "service.instance.id": "sdk-1", "deployment.environment.name": "prod", "service.namespace": ""}, Scope: parent.Definition.Source.Scope, SchemaURL: measurements.PoolConnectionsSchema, Name: parent.Definition.Source.Metric, Unit: parent.Definition.Source.Unit, Attributes: map[string]string{"db.client.connection.pool.name": "main", "db.client.connection.state": "idle"}}
	req := measurements.PoolSignalsRequest{Parent: parent, Stream: stream, Instruments: []measurements.Instrument{}}
	for _, kind := range []string{"capacity", "wait"} {
		d := parent.Definition
		d.Name = kind
		d.Title = kind
		d.StockDomain = ""
		d.Source.ExactStream = nil
		d.Source.Attributes = map[string]string{"db.client.connection.pool.name": "main"}
		d.Source.Metric = "db.client.connection.max"
		if kind == "wait" {
			d.Kind = "histogram"
			d.Source.Metric = "db.client.connection.wait_time"
			d.Source.Unit = "s"
		}
		candidate := stream
		candidate.Name = d.Source.Metric
		candidate.Unit = d.Source.Unit
		candidate.Attributes = map[string]string{"db.client.connection.pool.name": "main"}
		req.Instruments = append(req.Instruments, measurements.Instrument{Kind: d.Kind, Scope: d.Source.Scope, Version: d.Source.Version, SchemaURL: d.Source.SchemaURL, Metric: d.Source.Metric, Unit: d.Source.Unit, Temporality: 2, Supported: true, Definition: &d, CapabilityPlan: &measurements.InstrumentCapabilityPlan{Coverage: measurements.InstrumentCoverage{Streams: []measurements.ObservedStream{{Stream: candidate}}}}})
	}
	return req
}
func poolSignalsBody(change func(*measurements.PoolSignalsRequest)) string {
	req := poolSignalsHTTPRequest()
	if change != nil {
		change(&req)
	}
	raw, _ := json.Marshal(req)
	return string(raw)
}
func poolSignalsWireResponse(t *testing.T, c planningWireCase) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	registerPoolSignals(mux)
	body := c.Body + c.Suffix
	if c.PadTo > len(body) {
		body += strings.Repeat(" ", c.PadTo-len(body))
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(c.Method, c.Path, strings.NewReader(body)))
	return w
}
func TestPoolSignalsHTTPFrozenContracts(t *testing.T) {
	valid := poolSignalsBody(nil)
	cases := []planningWireCase{
		{Name: "matching-capacity-wait", Body: valid, Status: 200},
		{Name: "no-candidates-unknown", Body: poolSignalsBody(func(r *measurements.PoolSignalsRequest) { r.Instruments = nil }), Status: 200},
		{Name: "different-resource-no-broadening", Body: poolSignalsBody(func(r *measurements.PoolSignalsRequest) {
			for i := range r.Instruments {
				r.Instruments[i].CapabilityPlan.Coverage.Streams[0].Stream.Resource = map[string]string{"service.name": "other"}
			}
		}), Status: 200},
		{Name: "missing-pool", Body: poolSignalsBody(func(r *measurements.PoolSignalsRequest) {
			delete(r.Stream.Attributes, "db.client.connection.pool.name")
		}), Status: 400},
		{Name: "parent-not-domain-qualified", Body: poolSignalsBody(func(r *measurements.PoolSignalsRequest) { r.Parent.Definition.StockDomain = "" }), Status: 400},
		{Name: "pinned-parent-drift", Body: poolSignalsBody(func(r *measurements.PoolSignalsRequest) {
			old := r.Stream
			old.Name = "other"
			r.Parent.Definition.Source.ExactStream = &old
		}), Status: 400},
		{Name: "instrument-budget", Body: poolSignalsBody(func(r *measurements.PoolSignalsRequest) {
			for len(r.Instruments) < 17 {
				r.Instruments = append(r.Instruments, r.Instruments[0])
			}
		}), Status: 400},
		{Name: "negative-preview-omissions", Body: poolSignalsBody(func(r *measurements.PoolSignalsRequest) { r.Instruments[0].CapabilityPlan.Coverage.OmittedStreams = -1 }), Status: 400},
		{Name: "overflow-preview-omissions", Body: poolSignalsBody(func(r *measurements.PoolSignalsRequest) {
			for i := range r.Instruments {
				r.Instruments[i].CapabilityPlan.Coverage.OmittedStreams = int(^uint(0) >> 1)
			}
		}), Status: 400},
		{Name: "unknown-field", Body: strings.TrimSuffix(valid, "}") + `,"sql":"SELECT 1"}`, Status: 400},
		{Name: "wrong-instruments-type", Body: `{"instruments":true}`, Status: 400},
		{Name: "trailing-object", Body: valid, Suffix: ` {}`, Status: 400},
		{Name: "exact-64k", Body: valid, PadTo: 65536, Status: 200},
		{Name: "oversize", Body: valid, PadTo: 65537, Status: 400},
		{Name: "method", Method: "GET", Status: 405},
	}
	for i := range cases {
		cases[i].Path = "/pool-signals/plan"
		if cases[i].Method == "" {
			cases[i].Method = "POST"
		}
	}
	const filename = "../../testdata/http/pool-signals.json"
	if os.Getenv("GOO_UPDATE_POOL_SIGNALS_HTTP_FIXTURES") == "1" {
		for i := range cases {
			w := poolSignalsWireResponse(t, cases[i])
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
			w := poolSignalsWireResponse(t, c)
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
func TestPoolSignalsHTTPIsPureAndHonorsCancelledContext(t *testing.T) {
	mux := http.NewServeMux()
	registerPoolSignals(mux)
	// Pure planning ignores source-query capabilities and occupies no source slot.
	budget := newQueryBudget(mux, 0)
	request := httptest.NewRequest("POST", "/pool-signals/plan", strings.NewReader(poolSignalsBody(nil)))
	request.Header.Set("X-Goo-Query-ID", "not-a-query-capability")
	w := httptest.NewRecorder()
	budget.ServeHTTP(w, request)
	if w.Code != 200 {
		t.Fatal("pure planning incorrectly admitted as source query", w.Code, w.Body.String())
	}
	var result measurements.PoolSignalsPlan
	if json.Unmarshal(w.Body.Bytes(), &result) != nil || len(result.Capacity) != 1 || len(result.Wait) != 1 {
		t.Fatal(w.Body.String())
	}
	if result.Capacity[0].Question.Definition.StockDomain != "nonnegative-integer" || result.Wait[0].Question.Definition.Kind != "histogram" {
		t.Fatal("wrong executable signal question", result)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	request = httptest.NewRequest("POST", "/pool-signals/plan", strings.NewReader(poolSignalsBody(nil))).WithContext(ctx)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, request)
	if w.Code != 400 {
		t.Fatal("cancelled planner returned candidates", w.Code)
	}
	expired, stop := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer stop()
	request = httptest.NewRequest("POST", "/pool-signals/plan", strings.NewReader(poolSignalsBody(nil))).WithContext(expired)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, request)
	if w.Code != 400 || !strings.Contains(w.Body.String(), "deadline exceeded") {
		t.Fatal("expired pure planner returned candidates", w.Code, w.Body.String())
	}

}

func TestPoolSignalsHTTPExecuteCaptureOfflineReopen(t *testing.T) {
	resultSnapshots := &exploration.ResultSnapshots{}
	dir := t.TempDir()
	req := poolSignalsHTTPRequest()
	reads := 0
	offline := false
	streams := map[string]metrics.Stream{req.Stream.Name: req.Stream}
	for _, i := range req.Instruments {
		s := i.CapabilityPlan.Coverage.Streams[0].Stream
		streams[s.Name] = s
	}
	client := &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		if offline {
			t.Fatal("offline pool evidence reread source")
		}
		reads++
		name := ""
		for _, values := range r.URL.Query() {
			for _, value := range values {
				if _, ok := streams[value]; ok {
					name = value
				}
			}
		}
		stream, ok := streams[name]
		if !ok {
			t.Fatal("unexpected source instrument", r.URL.RawQuery)
		}
		raw := strings.Split(strings.TrimSpace(stockHTTPRows(2, false)), "\n")
		rows := []string{}
		for _, text := range raw {
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
			if strings.HasSuffix(name, "wait_time") {
				row["count"] = "2"
				row["bounds"] = []float64{0.1}
				row["bucket_counts"] = []string{"1", "1"}
				row["start_ns"] = "100000000000"
				row["end_ns"] = "102500000000"
			}
			b, _ := json.Marshal(row)
			rows = append(rows, string(b))
			if strings.HasSuffix(name, "wait_time") {
				break
			}
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(strings.Join(rows, "\n") + "\n")), Header: make(http.Header)}, nil
	})}
	makeMux := func() *http.ServeMux {
		m := http.NewServeMux()
		registerPoolSignals(m)
		registerMeasurements(m, measurements.Source{Endpoint: "http://source", Client: client}, dir)
		registerCaptures(m, exploration.Source{Endpoint: "http://source", Client: client}, dir, resultSnapshots)
		return m
	}
	mux := makeMux()
	observations := []apiObservation{}
	call := func(method, path, pattern string, input any, want int) []byte {
		var b []byte
		if input != nil {
			b, _ = json.Marshal(input)
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(string(b))))
		if w.Code != want {
			t.Fatal(method, path, w.Code, w.Body.String())
		}
		observations = append(observations, apiObservation{method, pattern, w.Code, b, append(json.RawMessage(nil), w.Body.Bytes()...)})
		return w.Body.Bytes()
	}
	var plan measurements.PoolSignalsPlan
	raw := call("POST", "/pool-signals/plan", "/pool-signals/plan", req, 200)
	if json.Unmarshal(raw, &plan) != nil || reads != 0 || len(plan.Capacity) != 1 || len(plan.Wait) != 1 {
		t.Fatal("pure candidate plan", string(raw), reads)
	}
	for _, q := range []measurements.Question{plan.Capacity[0].Question, plan.Wait[0].Question} {
		raw := call("POST", "/measurements", "/measurements", q, 200)
		var result measurements.Result
		if json.Unmarshal(raw, &result) != nil || measurements.ValidateResult(&result) != nil {
			t.Fatal("invalid executed related signal", string(raw))
		}
		if q.Definition.Kind == "stock" && (result.Stocks == nil || len(result.Stocks.Series) != 1) {
			t.Fatal("capacity missing stock", string(raw))
		}
		if q.Definition.Kind == "histogram" && (result.Histograms == nil || len(result.Histograms.Series) != 1) {
			t.Fatal("wait missing histogram", string(raw))
		}
	}
	pages := []exploration.CapturePage{}
	retained := [][]byte{}
	for _, q := range []measurements.Question{req.Parent, plan.Capacity[0].Question, plan.Wait[0].Question} {
		raw := call("POST", "/captures", "/captures", map[string]any{"title": "Pool signal evidence", "question": exploration.CaptureQuestion{Kind: "measurement", Measurement: &q}}, 201)
		var page exploration.CapturePage
		if json.Unmarshal(raw, &page) != nil || page.Measurement == nil {
			t.Fatal(string(raw))
		}
		pages = append(pages, page)
		native, _ := json.Marshal(page.Measurement)
		retained = append(retained, native)
	}
	before := reads
	offline = true
	mux = makeMux()
	for i, page := range pages {
		raw := call("GET", "/captures/"+page.Receipt.ID, "/captures/{id}", nil, 200)
		var reopened exploration.CapturePage
		_ = json.Unmarshal(raw, &reopened)
		native, _ := json.Marshal(reopened.Measurement)
		if string(native) != string(retained[i]) || reads != before {
			t.Fatal("original count/related signal changed on offline reopen")
		}
	}
	if path := os.Getenv("GOO_POOL_SIGNALS_API_OBSERVATIONS"); path != "" {
		raw, _ := json.MarshalIndent(observations, "", "  ")
		if err := os.WriteFile(path, append(raw, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
}
