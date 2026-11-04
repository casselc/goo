package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"gopkg.in/yaml.v3"
)

func TestTypedCounterHistogramDiscoveryValuePackageHTTP320(t *testing.T) {
	selector := identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}}
	resource := map[string]string{"service.name": "api", "service.instance.id": "same", "service.namespace": "team", "deployment.environment": "prod"}
	reads := 0
	client := &http.Client{Transport: nativeDependencyTransport(func(request *http.Request) (*http.Response, error) {
		reads++
		sql, _ := io.ReadAll(request.Body)
		rows := []map[string]any{}
		if strings.Contains(string(sql), " AS kind, ScopeName") {
			for _, kind := range []string{"counter", "histogram"} {
				unit := "items"
				if kind == "histogram" {
					unit = "ms"
				}
				rows = append(rows, map[string]any{"kind": kind, "scope": "sdk", "version": "", "schemaURL": "", "metric": "typed-" + kind, "unit": unit, "temporality": 2, "monotonic": kind == "counter", "dimensions": []string{}, "resource": resource, "attributes": map[string]string{}})
			}
		} else {
			histogram := strings.Contains(string(sql), "bucket_counts")
			kind, unit := "counter", "items"
			if histogram {
				kind, unit = "histogram", "ms"
			}
			for index, end := range []string{"100000000000", "102000000000", "104000000000"} {
				row := map[string]any{"resource": resource, "scope": "sdk", "version": "", "schema_url": "", "resource_schema_url": "", "name": "typed-" + kind, "unit": unit, "attributes": map[string]string{}, "start_ns": "50000000000", "end_ns": end, "flags": 0, "temporality": 2, "monotonic": true}
				if histogram {
					row["count"] = []string{"2", "5", "1"}[index]
					row["bounds"] = []float64{10, 20}
					row["bucket_counts"] = [][]string{{"1", "1", "0"}, {"2", "2", "1"}, {"1", "0", "0"}}[index]
				} else {
					row["value"] = []string{"10", "18", "3"}[index]
				}
				rows = append(rows, row)
			}
		}
		body := ""
		for _, row := range rows {
			raw, _ := json.Marshal(row)
			body += string(raw) + "\n"
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
	})}
	mux := http.NewServeMux()
	registerMeasurements(mux, measurements.Source{Endpoint: "http://source", Client: client}, t.TempDir())
	registerDashboards(mux, nil)
	post := func(path string, value any) *httptest.ResponseRecorder {
		t.Helper()
		raw, _ := json.Marshal(value)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest("POST", path, strings.NewReader(string(raw))))
		return response
	}
	response := post("/metric-discovery", measurements.DiscoveryQuestion{EnvironmentSelector: &selector, Namespace: &identity.Namespace{Mode: "value", Value: "team"}, Service: "api", FromMS: 100000, ToMS: 104000})
	var discovery measurements.DiscoveryResult
	if response.Code != 200 || json.Unmarshal(response.Body.Bytes(), &discovery) != nil || len(discovery.Instruments) != 2 || reads != 1 {
		t.Fatal("typed family discovery failed", response.Code, response.Body.String())
	}
	for _, instrument := range discovery.Instruments {
		if !instrument.Supported || instrument.Definition == nil || instrument.UnsupportedReason != "" || !reflect.DeepEqual(instrument.Definition.EnvironmentSelector, &selector) {
			t.Fatal("supported typed prototype lost policy")
		}
		question := measurements.Question{Definition: *instrument.Definition, FromMS: 100000, ToMS: 104000}
		before := reads
		response = post("/measurements", question)
		var result measurements.Result
		if response.Code != 200 || json.Unmarshal(response.Body.Bytes(), &result) != nil || !reflect.DeepEqual(result.Question, question) || reads != before+1 {
			t.Fatal("typed interval source receipt failed", response.Code, response.Body.String())
		}
		if instrument.Kind == "counter" {
			if result.Counters == nil || len(result.Counters.Series) != 1 || result.Counters.Series[0].Intervals[0].Delta != "8" || result.Counters.Series[0].Intervals[1].Reason != "counter-decrease" {
				t.Fatal("counter rate/reset semantics changed")
			}
		} else if result.Histograms == nil || len(result.Histograms.Series) != 1 || result.Histograms.Series[0].Intervals[0].Distribution.Count != "3" || result.Histograms.Series[0].Intervals[1].Distribution != nil {
			t.Fatal("explicit histogram support/reset semantics changed")
		}
		before = reads
		text, _ := yaml.Marshal(question.Definition)
		response = post("/measurement-packages/plan", map[string]any{"yaml": string(text), "fromMs": 100000, "toMs": 104000})
		if response.Code != 200 || reads != before {
			t.Fatal("typed package planning read source")
		}
		response = post("/measurement-packages", map[string]string{"yaml": string(text)})
		var saved struct {
			Revision   string                  `json:"revision"`
			Definition measurements.Definition `json:"definition"`
		}
		if response.Code != 201 || json.Unmarshal(response.Body.Bytes(), &saved) != nil || reads != before || !reflect.DeepEqual(saved.Definition.EnvironmentSelector, &selector) {
			t.Fatal("typed interval package did not preserve selector", response.Code, response.Body.String())
		}
		reopened := httptest.NewRecorder()
		mux.ServeHTTP(reopened, httptest.NewRequest("GET", "/measurement-packages/revisions/"+saved.Revision, nil))
		if reopened.Code != 200 || json.Unmarshal(reopened.Body.Bytes(), &saved) != nil || saved.Definition.Kind != instrument.Kind || !reflect.DeepEqual(saved.Definition.EnvironmentSelector, &selector) || reads != before {
			t.Fatal("typed interval revision lookup changed scope")
		}
	}
	before := reads
	response = post("/dashboards/plan", exploration.DashboardRequest{Scope: exploration.InventoryQuestion{EnvironmentSelector: &selector, Namespace: &identity.Namespace{Mode: "value", Value: "team"}, FromMS: 100000, ToMS: 104000}, Service: "api", Signals: []string{"metrics"}, Instruments: discovery.Instruments})
	var dashboard exploration.DashboardPlan
	if response.Code != 200 || json.Unmarshal(response.Body.Bytes(), &dashboard) != nil || dashboard.NativeCapability == nil || dashboard.NativeCapability.State != "service-per-stream" || len(dashboard.NativeOptions) != 2 || reads != before {
		t.Fatal("typed native dashboard options missing", response.Code, response.Body.String())
	}
	for _, widget := range dashboard.Widgets {
		if widget.Measurement != nil && !reflect.DeepEqual(widget.Measurement.Definition.EnvironmentSelector, &selector) {
			t.Fatal("dashboard native question downgraded policy")
		}
	}
	bad := *discovery.Instruments[0].Definition
	bad.Kind = "histogram"
	bad.Source.Unit = "items"
	response = post("/measurements", measurements.Question{Definition: bad, FromMS: 100000, ToMS: 104000})
	if response.Code != 400 || reads != before {
		t.Fatal("invalid typed histogram unit reached SQL")
	}
}
