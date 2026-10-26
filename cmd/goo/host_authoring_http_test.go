package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/casselc/goo/measurements"
	"gopkg.in/yaml.v3"
)

func TestHostYAMLPublishReopenAndDuplicateCanonicalServices(t *testing.T) {
	t.Chdir("../..")
	// Use an actual host-bound definition; only its YAML services spelling varies.
	var input measurements.HostDashboardRequest
	if err := json.Unmarshal([]byte(hostHTTPPlanBody(t, nil)), &input); err != nil {
		t.Fatal(err)
	}
	var definition measurements.Definition
	for _, instrument := range input.Host.Instruments {
		if instrument.Definition != nil {
			definition = *instrument.Definition
			break
		}
	}
	if definition.Source.Resource == nil {
		t.Fatal("missing host authoring fixture")
	}
	verifyHostYAMLPublicationVariants(t, definition, false)
}

func verifyHostYAMLPublicationVariants(t *testing.T, definition measurements.Definition, dimensionless bool) {
	t.Helper()
	canonical, err := yaml.Marshal(definition)
	if err != nil {
		t.Fatal(err)
	}
	source := string(canonical)
	line := ""
	for _, candidate := range strings.Split(source, "\n") {
		if strings.TrimSpace(candidate) == "services: []" {
			line = candidate
			break
		}
	}
	if line == "" {
		t.Fatal("expected canonical empty host services", source)
	}
	variants := []string{strings.Replace(source, line+"\n", "", 1), strings.Replace(source, "services: []", "services: null", 1), source}
	if dimensionless {
		attributeLine := ""
		for _, candidate := range strings.Split(source, "\n") {
			if strings.TrimSpace(candidate) == "attributes: {}" {
				attributeLine = candidate
				break
			}
		}
		if attributeLine == "" {
			t.Fatal("expected canonical empty exact attributes", source)
		}
		servicesVariants := variants
		variants = nil
		for _, wire := range servicesVariants {
			variants = append(variants, strings.Replace(wire, attributeLine+"\n", "", 1), strings.Replace(wire, "attributes: {}", "attributes: null", 1), wire)
		}
	}

	dir := t.TempDir()
	makeMux := func() *http.ServeMux {
		mux := http.NewServeMux()
		registerMeasurements(mux, measurements.Source{Client: http.DefaultClient}, dir)
		return mux
	}
	mux := makeMux()
	revision := ""
	winnerResponse := ""
	for index, wire := range variants {
		body, _ := json.Marshal(map[string]string{"yaml": wire})
		for pass := 0; pass < 2; pass++ {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, httptest.NewRequest("POST", "/measurement-packages", strings.NewReader(string(body))))
			if w.Code != 201 {
				t.Fatal("host publication/duplicate failed", index, pass, w.Code, w.Body.String())
			}
			var item struct {
				Revision   string                  `json:"revision"`
				Definition measurements.Definition `json:"definition"`
				YAML       string                  `json:"yaml"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &item); err != nil {
				t.Fatal(err)
			}
			if dimensionless && (item.Definition.Source.ExactStream == nil || item.Definition.Source.ExactStream.Attributes == nil || len(item.Definition.Source.ExactStream.Attributes) != 0) {
				t.Fatal("authoring did not canonicalize empty exact attributes")
			}
			if item.Definition.Source.Services == nil || len(item.Definition.Source.Services) != 0 {
				t.Fatal("authoring did not canonicalize empty services")
			}
			if revision == "" {
				revision = item.Revision
				winnerResponse = w.Body.String()
			} else if revision != item.Revision || winnerResponse != w.Body.String() {
				t.Fatal("services spelling changed published identity/bytes")
			}
			stored, err := os.ReadFile(filepath.Join(dir, "measurement-packages", revision+".yaml"))
			if err != nil {
				t.Fatal(err)
			}
			loaded, err := measurements.Read(string(stored))
			if err != nil || measurements.Hash(loaded) != revision {
				t.Fatal("published host winner cannot reopen", err)
			}
			if string(stored) != item.YAML {
				t.Fatal("stored canonical YAML differs from publication")
			}
		}
		// Fresh registration verifies the published winner through the real list route.
		mux = makeMux()
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("GET", "/measurement-packages", nil))
		if w.Code != 200 {
			t.Fatal("fresh host catalog reopen failed", w.Code, w.Body.String())
		}
		var items []struct {
			Revision   string                  `json:"revision"`
			Definition measurements.Definition `json:"definition"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
			t.Fatal(err)
		}
		found := false
		for _, item := range items {
			if item.Revision == revision {
				found = true
				if measurements.Hash(item.Definition) != revision {
					t.Fatal("listed host definition revision mismatch")
				}
			}
		}
		if !found {
			t.Fatal("published host absent from fresh catalog")
		}
	}
	entries, err := os.ReadDir(filepath.Join(dir, "measurement-packages"))
	if err != nil || len(entries) != 1 {
		t.Fatal("services spellings published multiple revisions", err, len(entries))
	}
}

func TestHostYAMLDimensionlessAttributesAndServicesPublicationMatrix(t *testing.T) {
	t.Chdir("../..")
	var input measurements.HostDashboardRequest
	if err := json.Unmarshal([]byte(hostHTTPPlanBody(t, nil)), &input); err != nil {
		t.Fatal(err)
	}
	var definition measurements.Definition
	for _, instrument := range input.Host.Instruments {
		if instrument.Definition != nil && instrument.Kind == "gauge" {
			definition = *instrument.Definition
			break
		}
	}
	if definition.Source.ExactStream == nil {
		t.Fatal("no exact host gauge fixture")
	}
	definition.Name = "dimensionless-host-gauge"
	definition.Title = "Dimensionless host gauge"
	definition.Source.Metric = "host.dimensionless"
	definition.Source.ExactStream.Name = definition.Source.Metric
	definition.Source.ExactStream.Attributes = map[string]string{}
	verifyHostYAMLPublicationVariants(t, definition, true)
}
