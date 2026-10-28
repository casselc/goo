package nativehttp_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/nativehttp"
	"github.com/casselc/goo/internal/queryhttp"
	"github.com/casselc/goo/measurements"
)

func TestNativeAdmissionLeavesSuppliedDataPlansAvailable(t *testing.T) {
	mux := http.NewServeMux()
	nativehttp.RegisterMeasurements(mux, measurements.Source{})
	nativehttp.RegisterPodDiscovery(mux, exploration.Source{})
	budget := queryhttp.NewBudget(mux, 0, func(r *http.Request) bool {
		policy, ok := nativehttp.PolicyFor(r.Method, r.URL.Path)
		return ok && policy.Effect != queryhttp.Pure
	})
	for _, path := range []string{"/metric-discovery", "/measurements", "/host-discovery", "/pod-discovery", "/pod-metric-discovery"} {
		w := httptest.NewRecorder()
		budget.ServeHTTP(w, httptest.NewRequest("POST", path, strings.NewReader(`{}`)))
		if w.Code != 429 {
			t.Fatalf("source bypassed full budget: %s %d", path, w.Code)
		}
	}
	for _, name := range []string{"host", "pod-metrics"} {
		data, err := os.ReadFile("../../testdata/http/" + name + ".json")
		if err != nil {
			t.Fatal(err)
		}
		var fixtures struct {
			Cases []struct {
				Path, Body string
				Status     int
			}
		}
		if err := json.Unmarshal(data, &fixtures); err != nil {
			t.Fatal(err)
		}
		checked := false
		for _, fixture := range fixtures.Cases {
			if !strings.HasSuffix(fixture.Path, "/plan") || fixture.Status != 200 {
				continue
			}
			r := httptest.NewRequest("POST", fixture.Path, strings.NewReader(fixture.Body))
			r.Header.Set("X-Goo-Query-ID", "irrelevant-to-pure-planner")
			w := httptest.NewRecorder()
			budget.ServeHTTP(w, r)
			if w.Code != 200 {
				t.Fatalf("pure planner inherited source admission: %s %d %s", fixture.Path, w.Code, w.Body.String())
			}
			checked = true
			break
		}
		if !checked {
			t.Fatal("missing accepted planner fixture", name)
		}
	}
	for _, path := range []string{"/measurements/other", "/measurement-packages", "/captures", "/tail"} {
		if _, ok := nativehttp.PolicyFor("POST", path); ok {
			t.Fatal("native family claimed unrelated effect", path)
		}
	}
	if _, ok := nativehttp.PolicyFor("GET", "/measurements"); ok {
		t.Fatal("POST policy leaked to GET")
	}
}

func TestNativeCustodyUsesConfiguredDirectoryOnlyForLifecycle(t *testing.T) {
	definition, err := os.ReadFile("../../config/measurements/reservation-lifecycle.yaml")
	if err != nil {
		t.Fatal(err)
	}
	d, err := measurements.Read(string(definition))
	if err != nil {
		t.Fatal(err)
	}
	row := map[string]any{"body": d.Lifecycle.Accepted, "service": d.Source.Services[0], "resource": map[string]string{"deployment.environment.name": d.Environment}, "attributes": map[string]string{
		d.Lifecycle.EventID: "fixture-event", d.Lifecycle.EventTime: "150000000000", d.Lifecycle.Subject: "fixture-subject", d.Lifecycle.Tenant: "fixture-tenant", d.Lifecycle.Release: "fixture-release",
	}}
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			return
		}
		if strings.Contains(string(body), "FROM otel_logs") {
			json.NewEncoder(w).Encode(row)
		}
	}))
	defer source.Close()
	root := t.TempDir()
	dir := filepath.Join(root, "configured-custody")
	mux := http.NewServeMux()
	nativehttp.RegisterMeasurements(mux, measurements.Source{Endpoint: source.URL, Client: source.Client(), CustodyDir: dir})
	discovery := fmt.Sprintf(`{"environment":%q,"service":%q,"fromMs":100000,"toMs":220000}`, d.Environment, d.Source.Services[0])
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/metric-discovery", strings.NewReader(discovery)))
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatal("discovery created lifecycle custody", err)
	}
	body, err := json.Marshal(measurements.Question{Definition: d, FromMS: 100000, ToMS: 220000})
	if err != nil {
		t.Fatal(err)
	}
	var original []byte
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/measurements", bytes.NewReader(body)))
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		files, err := os.ReadDir(dir)
		if err != nil || len(files) != 1 {
			t.Fatal("expected one first-observation file", files, err)
		}
		data, err := os.ReadFile(filepath.Join(dir, files[0].Name()))
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			original = data
		} else if !bytes.Equal(original, data) {
			t.Fatal("repeated read rewrote custody observation")
		}
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 1 || entries[0].Name() != "configured-custody" {
		t.Fatal("native registration wrote outside its configured custody owner", entries, err)
	}
}
