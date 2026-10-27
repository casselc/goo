package main

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/measurements"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func containerHistoryHTTPRequest(t *testing.T) measurements.ContainerResourceHistoryRequest {
	memory := containerMemoryHTTPRequest(t)
	cpu := containerCPUHTTPRequest(t)
	return measurements.ContainerResourceHistoryRequest{Parent: memory.Parent, Memory: &measurements.ContainerMemoryHistoryResults{WorkingSet: memory.WorkingSet, Request: memory.Request, Limit: memory.Limit}, CPU: &measurements.ContainerCPUHistoryResults{Usage: cpu.Usage, Request: cpu.Request, Limit: cpu.Limit}}
}
func containerHistoryHTTPBody(t *testing.T, edit func(*measurements.ContainerResourceHistoryRequest)) string {
	input := containerHistoryHTTPRequest(t)
	if edit != nil {
		edit(&input)
	}
	encoded, _ := json.Marshal(input)
	return string(encoded)
}
func containerHistoryHTTPResponse(t *testing.T, c planningWireCase) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	registerContainerResourceHistory(mux)
	body := c.Body + c.Suffix
	if c.PadTo > len(body) {
		body += strings.Repeat(" ", c.PadTo-len(body))
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(c.Method, c.Path, strings.NewReader(body)))
	return w
}
func TestContainerResourceHistoryHTTPFrozenContracts(t *testing.T) {
	full := containerHistoryHTTPBody(t, nil)
	cases := []planningWireCase{
		{Name: "held-memory-and-cpu", Body: full, Status: 200},
		{Name: "missing-family-observations", Body: containerHistoryHTTPBody(t, func(r *measurements.ContainerResourceHistoryRequest) {
			r.CPU = nil
			r.Memory = &measurements.ContainerMemoryHistoryResults{}
		}), Status: 200},
		{Name: "cpu-only", Body: containerHistoryHTTPBody(t, func(r *measurements.ContainerResourceHistoryRequest) { r.Memory = nil }), Status: 200},
		{Name: "historical-invalid-byte-observation", Body: containerHistoryHTTPBody(t, func(r *measurements.ContainerResourceHistoryRequest) {
			var row map[string]any
			json.Unmarshal(r.Memory.WorkingSet.Rows[0], &row)
			row["end_ns"] = "110000000000"
			row["value"] = "0.5"
			old, _ := json.Marshal(row)
			result, e := measurements.Evaluate(r.Memory.WorkingSet.Question, []json.RawMessage{r.Memory.WorkingSet.Rows[0], old}, time.Unix(200, 0), "", nil)
			if e != nil {
				t.Fatal(e)
			}
			r.Memory.WorkingSet = result
		}), Status: 200},
		{Name: "no-data", Body: containerHistoryHTTPBody(t, func(r *measurements.ContainerResourceHistoryRequest) {
			result, e := measurements.Evaluate(r.Memory.WorkingSet.Question, nil, time.Unix(200, 0), "", nil)
			if e != nil {
				t.Fatal(e)
			}
			r.Memory.WorkingSet = result
		}), Status: 200},
		{Name: "no-family", Body: containerHistoryHTTPBody(t, func(r *measurements.ContainerResourceHistoryRequest) { r.Memory = nil; r.CPU = nil }), Status: 400},
		{Name: "wrong-role", Body: containerHistoryHTTPBody(t, func(r *measurements.ContainerResourceHistoryRequest) { r.CPU.Usage = r.Memory.WorkingSet }), Status: 400},
		{Name: "forged-observation", Body: containerHistoryHTTPBody(t, func(r *measurements.ContainerResourceHistoryRequest) {
			r.Memory.WorkingSet.Gauges.Series[0].Samples[0].Value = nil
		}), Status: 400},
		{Name: "wrong-context", Body: containerHistoryHTTPBody(t, func(r *measurements.ContainerResourceHistoryRequest) {
			parent := containerCPUHTTPParent()
			parent.Resource["container.id"] = "replacement"
			r.CPU.Usage = containerCPUHTTPResult(t, parent, "usage", "0.125")
		}), Status: 400},
		{Name: "unknown-field", Body: `{"unknown":true}`, Status: 400},
		{Name: "second-document", Body: full, Suffix: " {}", Status: 400},
		{Name: "oversize", Body: full, PadTo: (8 << 20) + 1, Status: 400},
		{Name: "method", Method: "GET", Status: 405},
	}
	for i := range cases {
		cases[i].Path = "/container-resources/history"
		if cases[i].Method == "" {
			cases[i].Method = "POST"
		}
	}
	const file = "../../testdata/http/container-resource-history.json"
	if os.Getenv("GOO_UPDATE_CONTAINER_HISTORY_HTTP_FIXTURES") == "1" {
		for i := range cases {
			w := containerHistoryHTTPResponse(t, cases[i])
			if w.Code != cases[i].Status {
				t.Fatal(cases[i].Name, w.Code, w.Body.String())
			}
			cases[i].Response = w.Body.String()
			cases[i].Headers = map[string]string{}
			for _, key := range []string{"Content-Type", "Cache-Control", "Allow", "X-Content-Type-Options"} {
				cases[i].Headers[key] = w.Header().Get(key)
			}
		}
		data, _ := json.MarshalIndent(planningWireFixtures{Version: 1, Cases: cases}, "", "  ")
		if e := os.WriteFile(file, append(data, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	}
	fixture := dashboardProfileFixture(t, file, len(cases))
	for _, c := range fixture.Cases {
		w := containerHistoryHTTPResponse(t, c)
		if w.Code != c.Status || w.Body.String() != c.Response {
			t.Fatal(c.Name, w.Code, w.Body.String())
		}
		for key, value := range c.Headers {
			if w.Header().Get(key) != value {
				t.Fatal(key)
			}
		}
	}
}
func TestContainerResourceHistoryPureServedMux(t *testing.T) {
	mux := http.NewServeMux()
	registerMeasurements(mux, measurements.Source{}, t.TempDir())
	body := containerHistoryHTTPBody(t, nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/container-resources/history", strings.NewReader(body)))
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/container-resources/history", strings.NewReader(body)).WithContext(ctx))
	if w.Code != 400 {
		t.Fatal("canceled pure projection", w.Code)
	}
}
