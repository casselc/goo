package main

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

const podDiscoveryHTTPBody = `{"environment":"prod","fromMs":1000,"toMs":61000}`

func podHTTPRows() string {
	resource := map[string]string{"deployment.environment.name": "prod", "k8s.pod.uid": "pod-a", "k8s.pod.name": "reported-pod", "k8s.namespace.name": "checkout"}
	rows := []string{}
	add := func(family string, schema any, resource map[string]string, count string) {
		raw, _ := json.Marshal(map[string]any{"family": family, "resource": resource, "resource_schema_url": schema, "count": count, "first_ns": "1000000001", "last_ns": "2000000001"})
		rows = append(rows, string(raw))
	}
	for _, family := range []string{"logs", "spans", "gauge", "sum", "histogram"} {
		var schema any = ""
		if family == "spans" {
			schema = nil
		}
		count := "2"
		if family == "logs" {
			count = "18446744073709551615"
		}
		add(family, schema, resource, count)
	}
	add("logs", "", map[string]string{"deployment.environment.name": "prod", "k8s.pod.name": "missing-uid"}, "1")
	add("spans", nil, map[string]string{"deployment.environment.name": "prod", "k8s.pod.name": "empty-uid", "k8s.pod.uid": ""}, "1")
	return strings.Join(rows, "\n") + "\n"
}
func podHTTPWireResponse(t *testing.T, c planningWireCase) *httptest.ResponseRecorder {
	resultSnapshots := &exploration.ResultSnapshots{}
	t.Helper()
	reads := 0
	source := exploration.Source{Endpoint: "http://source", Client: &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		reads++
		deadline, ok := r.Context().Deadline()
		if !ok || time.Until(deadline) > 30*time.Second || time.Until(deadline) < 25*time.Second {
			t.Fatal("pod source context missing")
		}
		if r.URL.Query().Get("readonly") != "1" {
			t.Fatal("pod inventory not readonly")
		}
		sql, _ := io.ReadAll(r.Body)
		if path := os.Getenv("GOO_POD_DISCOVERY_SQL"); path != "" {
			if e := os.WriteFile(path, sql, 0644); e != nil {
				t.Fatal(e)
			}
		}
		payload := podHTTPRows()
		status := 200
		if c.Name == "source-unavailable" {
			status = 503
			payload = "fixture unavailable"
		}
		if c.Name == "source-group-sentinel" {
			one := strings.Split(strings.TrimSpace(payload), "\n")[0]
			payload = strings.Repeat(one+"\n", 1001)
		}
		return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(payload)), Header: make(http.Header)}, nil
	})}}
	mux := http.NewServeMux()
	registerExploration(mux, source, resultSnapshots, nil)
	body := c.Body + c.Suffix
	if c.PadTo > len(body) {
		body += strings.Repeat(" ", c.PadTo-len(body))
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(c.Method, c.Path, strings.NewReader(body)))
	want := 0
	if c.Status == 200 || c.Status == 502 {
		want = 1
	}
	if reads != want {
		t.Fatal("pod source effects", reads, want)
	}
	if c.Status == 200 {
		var result exploration.PodDiscovery
		if e := json.Unmarshal(w.Body.Bytes(), &result); e != nil {
			t.Fatal(e)
		}
		if result.Version != "goo/pod-discovery/v1" || result.SourceGroups != 7 || len(result.Observations) != 4 {
			t.Fatal("pod resource variants lost", w.Body.String())
		}
		if _, e := time.Parse(time.RFC3339Nano, result.ReadAt); e != nil {
			t.Fatal(e)
		}
		unknown, knownEmpty, missing, empty := false, false, false, false
		for _, entry := range result.Observations {
			if entry.ResourceSchemaURL == nil {
				unknown = true
			} else if *entry.ResourceSchemaURL == "" {
				knownEmpty = true
			}
			if entry.PodUIDState == "missing" {
				missing = true
				if entry.Actionable {
					t.Fatal("missing UID actionable")
				}
			}
			if entry.PodUIDState == "empty" {
				empty = true
				if entry.Actionable {
					t.Fatal("empty UID actionable")
				}
			}
		}
		if !unknown || !knownEmpty || !missing || !empty {
			t.Fatal("nullable schema or UID states collapsed")
		}
		// Only the real clock is normalized; content basis is independent of readAt.
		result.ReadAt = "2026-10-26T00:00:00Z"
		raw, _ := json.Marshal(result)
		w.Body.Reset()
		w.Body.Write(append(raw, '\n'))
	}
	return w
}
func TestPodDiscoveryHTTPWireContracts(t *testing.T) {
	cases := []planningWireCase{{Name: "five-families-resource-variants", Body: podDiscoveryHTTPBody, Status: 200}, {Name: "missing-scope", Body: `{}`, Status: 400}, {Name: "unknown-field", Body: strings.TrimSuffix(podDiscoveryHTTPBody, "}") + `,"sql":"SELECT 1"}`, Status: 400}, {Name: "trailing", Body: podDiscoveryHTTPBody, Suffix: ` {}`, Status: 400}, {Name: "exact-16KiB", Body: podDiscoveryHTTPBody, PadTo: 16384, Status: 200}, {Name: "oversize", Body: podDiscoveryHTTPBody, PadTo: 16385, Status: 400}, {Name: "source-unavailable", Body: podDiscoveryHTTPBody, Status: 502}, {Name: "source-group-sentinel", Body: podDiscoveryHTTPBody, Status: 502}, {Name: "method", Method: "GET", Status: 405}}
	for i := range cases {
		cases[i].Path = "/pod-discovery"
		if cases[i].Method == "" {
			cases[i].Method = "POST"
		}
	}
	const filename = "../../testdata/http/pod-discovery.json"
	if os.Getenv("GOO_UPDATE_POD_DISCOVERY_HTTP_FIXTURES") == "1" {
		for i := range cases {
			w := podHTTPWireResponse(t, cases[i])
			if w.Code != cases[i].Status {
				t.Fatal(cases[i].Name, w.Code, w.Body.String())
			}
			cases[i].Response = w.Body.String()
			cases[i].Headers = map[string]string{}
			for _, k := range []string{"Content-Type", "Cache-Control", "Allow", "X-Content-Type-Options"} {
				cases[i].Headers[k] = w.Header().Get(k)
			}
		}
		raw, _ := json.MarshalIndent(planningWireFixtures{Version: 1, Cases: cases}, "", "  ")
		if e := os.WriteFile(filename, append(raw, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	}
	fixture := dashboardProfileFixture(t, filename, len(cases))
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			w := podHTTPWireResponse(t, c)
			if w.Code != c.Status || w.Body.String() != c.Response {
				t.Fatal(w.Code, w.Body.String())
			}
			for k, v := range c.Headers {
				if w.Header().Get(k) != v {
					t.Fatal("pod header changed", k)
				}
			}
		})
	}
}
func TestPodDiscoverySharedAdmissionAndInflightCancellation(t *testing.T) {
	resultSnapshots := &exploration.ResultSnapshots{}
	called := false
	budget := newQueryBudget(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }), 1)
	id := "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
	cancelled := httptest.NewRecorder()
	budget.ServeHTTP(cancelled, httptest.NewRequest("POST", "/queries/"+id+"/cancel", nil))
	req := httptest.NewRequest("POST", "/pod-discovery", strings.NewReader(podDiscoveryHTTPBody))
	req.Header.Set("X-Goo-Query-ID", id)
	w := httptest.NewRecorder()
	budget.ServeHTTP(w, req)
	if cancelled.Code != 202 || w.Code != 409 || called {
		t.Fatal("pre-cancelled pod query admitted")
	}
	started := make(chan struct{}, 1)
	source := exploration.Source{Endpoint: "http://source", Client: &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		started <- struct{}{}
		<-r.Context().Done()
		return nil, r.Context().Err()
	})}}
	mux := http.NewServeMux()
	registerExploration(mux, source, resultSnapshots, nil)
	budget = newQueryBudget(mux, 1)
	id = "ffffffff-ffff-ffff-ffff-ffffffffffff"
	req = httptest.NewRequest("POST", "/pod-discovery", strings.NewReader(podDiscoveryHTTPBody))
	req.Header.Set("X-Goo-Query-ID", id)
	done := make(chan *httptest.ResponseRecorder, 1)
	go func() { w := httptest.NewRecorder(); budget.ServeHTTP(w, req); done <- w }()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("pod source not started")
	}
	busy := httptest.NewRecorder()
	budget.ServeHTTP(busy, httptest.NewRequest("POST", "/explore", strings.NewReader(`{}`)))
	if busy.Code != 429 {
		t.Fatal("pod discovery has independent source admission")
	}
	cancelled = httptest.NewRecorder()
	budget.ServeHTTP(cancelled, httptest.NewRequest("POST", "/queries/"+id+"/cancel", nil))
	select {
	case w := <-done:
		if cancelled.Code != 202 || w.Code != 502 {
			t.Fatal("pod cancel became data", w.Code)
		}
	case <-time.After(time.Second):
		t.Fatal("pod source cancellation not propagated")
	}
}
func TestPodDiscoveryGenericRuntimeRouteExistsWithoutCredentials(t *testing.T) {
	dir := t.TempDir()
	server, e := configuredHTTPServer([]string{"-profile", "generic", "-investigations-dir", dir})
	if e != nil {
		t.Fatal(e)
	}
	w := httptest.NewRecorder()
	request := httptest.NewRequest("POST", "/pod-discovery", strings.NewReader(`{}`)).WithContext(context.Background())
	server.Handler.ServeHTTP(w, request)
	if w.Code != 400 {
		t.Fatal("generic pod route missing", w.Code)
	}
	entries, e := os.ReadDir(dir)
	if e != nil || len(entries) != 0 {
		t.Fatal("invalid pod request created custody/publication")
	}
}
