package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"gopkg.in/yaml.v3"
)

func TestTypedNativeGaugeHTTPAndPackage319(t *testing.T) {
	reads := 0
	var resource map[string]string
	client := &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		reads++
		sql, _ := io.ReadAll(r.Body)
		var row any
		if strings.Contains(string(sql), " AS kind, ScopeName") {
			row = map[string]any{"kind": "gauge", "scope": "sdk", "version": "", "schemaURL": "", "metric": "load", "unit": "items", "temporality": 0, "monotonic": false, "dimensions": []string{}, "resource": resource, "attributes": map[string]string{}}
		} else {
			row = map[string]any{"resource": resource, "scope": "sdk", "version": "", "schema_url": "", "resource_schema_url": "", "name": "load", "unit": "items", "attributes": map[string]string{}, "start_ns": "0", "end_ns": "150000000000", "value": "456", "flags": 0}
		}
		raw, _ := json.Marshal(row)
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(string(raw) + "\n"))}, nil
	})}
	dir := t.TempDir()
	mux := http.NewServeMux()
	registerMeasurements(mux, measurements.Source{Endpoint: "http://source", Client: client}, dir)
	post := func(path string, value any) *httptest.ResponseRecorder {
		t.Helper()
		raw, _ := json.Marshal(value)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", path, strings.NewReader(string(raw))))
		return w
	}
	selectors := []identity.EnvironmentSelector{
		{Version: identity.EnvironmentSelectorVersion, Policy: "current-only", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "empty"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "unreported"}},
	}
	var prototype measurements.Definition
	for _, selector := range selectors {
		t.Run(selector.Policy+selector.Scope.Mode, func(t *testing.T) {
			resource = map[string]string{"service.name": "api", "service.instance.id": "same", "service.namespace": "team"}
			switch {
			case selector.Policy == "current-only":
				resource["deployment.environment.name"] = "prod"
			case selector.Scope.Mode == "named":
				resource["deployment.environment"] = "prod"
			case selector.Scope.Mode == "empty":
				resource["deployment.environment.name"] = ""
				resource["deployment.environment"] = "prod"
			}
			before := reads
			discovery := post("/metric-discovery", measurements.DiscoveryQuestion{EnvironmentSelector: &selector, Namespace: &identity.Namespace{Mode: "value", Value: "team"}, Service: "api", FromMS: 100000, ToMS: 160000})
			var inventory measurements.DiscoveryResult
			if discovery.Code != 200 || json.Unmarshal(discovery.Body.Bytes(), &inventory) != nil || len(inventory.Instruments) != 1 || inventory.Instruments[0].Definition == nil {
				t.Fatal("typed discovery failed", discovery.Code, discovery.Body.String())
			}
			prototype = *inventory.Instruments[0].Definition
			if !reflect.DeepEqual(prototype.EnvironmentSelector, &selector) || prototype.Environment != "" {
				t.Fatal("discovery discarded selector")
			}
			question := measurements.Question{Definition: prototype, FromMS: 100000, ToMS: 160000}
			response := post("/measurements", question)
			var result measurements.Result
			if response.Code != 200 || json.Unmarshal(response.Body.Bytes(), &result) != nil || !reflect.DeepEqual(result.Question, question) || result.Gauges == nil || len(result.Gauges.Series) != 1 || *result.Gauges.Series[0].Samples[0].Value != 456 || reads-before != 2 {
				t.Fatal("typed native values/receipt/budget mismatch", response.Code, response.Body.String())
			}
		})
	}
	text, _ := yaml.Marshal(prototype)
	before := reads
	preview := post("/measurement-packages/plan", map[string]any{"yaml": string(text), "fromMs": 100000, "toMs": 160000})
	if preview.Code != 200 || reads != before {
		t.Fatal("typed package preview failed/read source", preview.Code, preview.Body.String())
	}
	published := post("/measurement-packages", map[string]string{"yaml": string(text)})
	var saved struct {
		Definition measurements.Definition `json:"definition"`
		Revision   string                  `json:"revision"`
	}
	if published.Code != 201 || json.Unmarshal(published.Body.Bytes(), &saved) != nil || !reflect.DeepEqual(saved.Definition.EnvironmentSelector, prototype.EnvironmentSelector) || reads != before {
		t.Fatal("typed package publication lost policy", published.Code, published.Body.String())
	}
	reopened := httptest.NewRecorder()
	mux.ServeHTTP(reopened, httptest.NewRequest("GET", "/measurement-packages/revisions/"+saved.Revision, nil))
	if reopened.Code != 200 || json.Unmarshal(reopened.Body.Bytes(), &saved) != nil || !reflect.DeepEqual(saved.Definition.EnvironmentSelector, prototype.EnvironmentSelector) || reads != before {
		t.Fatal("typed revision lookup lost policy", reopened.Code, reopened.Body.String())
	}
	entries, _ := os.ReadDir(filepath.Join(dir, "measurement-packages"))
	for _, kind := range []string{"counter", "stock", "histogram", "state", "lifecycle"} {
		bad := prototype
		bad.Kind = kind
		response := post("/measurements", measurements.Question{Definition: bad, FromMS: 100000, ToMS: 160000})
		if response.Code != 400 || reads != before {
			t.Fatal("unsupported typed family reached source", kind, response.Code)
		}
		text, _ := yaml.Marshal(bad)
		response = post("/measurement-packages", map[string]string{"yaml": string(text)})
		after, _ := os.ReadDir(filepath.Join(dir, "measurement-packages"))
		if response.Code != 400 || len(after) != len(entries) || reads != before {
			t.Fatal("unsupported typed package produced effects", kind, response.Code)
		}
	}
	resourceBound := prototype
	resourceBound.Source.Resource = &measurements.ResourceBinding{Kind: "host", ID: "host"}
	refusal := post("/measurements", measurements.Question{Definition: resourceBound, FromMS: 100000, ToMS: 160000})
	if refusal.Code != 400 || reads != before {
		t.Fatal("typed resource binding reached source")
	}
	raw, _ := json.Marshal(measurements.Question{Definition: prototype, FromMS: 100000, ToMS: 160000})
	for _, body := range []string{strings.Replace(string(raw), `"environmentSelector":`, `"environment":"prod","environmentSelector":`, 1), strings.Replace(string(raw), `"environmentSelector":`, `"Environment":null,"environmentSelector":`, 1)} {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/measurements", strings.NewReader(body)))
		if w.Code != 400 || reads != before {
			t.Fatal("invalid native union reached source", w.Code)
		}
	}
}
