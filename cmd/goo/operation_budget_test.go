package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/casselc/goo/annotations"
	"github.com/casselc/goo/internal/operationhttp"
	"github.com/casselc/goo/prebuilt"
)

func TestLegacyOperationAdmission(t *testing.T) {
	mux := http.NewServeMux()
	registerOperationReads(mux, nil, nil, nil, prebuilt.HTTPSource{}, clickHouseConnection{})
	handler := newQueryBudget(mux, 0)
	for _, path := range []string{"/http-package", "/operations", "/counter-rates", "/histograms"} {
		for _, method := range []string{"GET", "HEAD"} {
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, httptest.NewRequest(method, path, nil))
			if w.Code != 429 || w.Header().Get("Retry-After") != "1" {
				t.Fatalf("%s %s: %d", method, path, w.Code)
			}
		}
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest("GET", "/operation-packages", nil))
	if w.Code != 200 {
		t.Fatalf("declaration blocked: %d", w.Code)
	}
	for _, c := range []struct{ method, path string }{{"POST", "/operations"}, {"GET", "/operations/other"}} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(c.method, c.path, nil))
		if w.Code == 429 {
			t.Fatal("unrelated request admitted")
		}
	}
}

func TestOperationSourceCancellationAndDeadline(t *testing.T) {
	plan, err := annotations.LoadOperationPackage("../../config/operations/checkout.yaml")
	if err != nil {
		t.Fatal(err)
	}
	started := make(chan time.Duration, 1)
	client := &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		deadline, ok := r.Context().Deadline()
		if !ok {
			t.Error("missing source deadline")
		}
		started <- time.Until(deadline)
		<-r.Context().Done()
		return nil, r.Context().Err()
	})}
	mux := http.NewServeMux()
	registerOperationReads(mux, map[string]*annotations.OperationPlan{"checkout": plan}, []*annotations.OperationPlan{plan}, nil, prebuilt.HTTPSource{}, clickHouseConnection{Client: client, Endpoint: "http://fixture.invalid"})
	handler := newQueryBudget(mux, 1)
	const id = "12345678-1234-1234-1234-123456789abc"
	req := httptest.NewRequest("GET", "/operations", nil)
	req.Header.Set("X-Goo-Query-ID", id)
	result := make(chan *httptest.ResponseRecorder, 1)
	go func() { w := httptest.NewRecorder(); handler.ServeHTTP(w, req); result <- w }()
	select {
	case remaining := <-started:
		policy, _ := operationhttp.PolicyFor("GET", "/operations")
		if remaining <= 0 || remaining > policy.Timeout {
			t.Fatalf("deadline: %v", remaining)
		}
	case <-time.After(time.Second):
		t.Fatal("source did not start")
	}

	// Legacy reads share capacity with the modern record family; declarations stay available.
	for _, c := range []struct {
		method, path string
		status       int
	}{{"POST", "/explore", 429}, {"GET", "/operation-packages", 200}} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(c.method, c.path, nil))
		if w.Code != c.status {
			t.Fatalf("shared capacity %s: %d", c.path, w.Code)
		}
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest("POST", "/queries/"+id+"/cancel", nil))
	if w.Code != 202 {
		t.Fatal(w.Code)
	}
	select {
	case w := <-result:
		if w.Code != 502 {
			t.Fatalf("cancel status: %d", w.Code)
		}
	case <-time.After(time.Second):
		t.Fatal("source did not cancel")
	}
	// Completed/cancelled IDs cannot accidentally start another source read.
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req.Clone(context.Background()))
	if w.Code != 409 {
		t.Fatal(w.Code)
	}
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest("GET", "/operations?minutes=0", nil))
	if w.Code != 400 {
		t.Fatalf("capacity not released: %d", w.Code)
	}
}
