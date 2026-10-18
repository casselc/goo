package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestQueryBudgetCancelAndAdmission(t *testing.T) {
	entered := make(chan struct{}, 1)
	cancelled := make(chan struct{}, 1)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/captures/local" {
			w.WriteHeader(200)
			return
		}
		entered <- struct{}{}
		<-r.Context().Done()
		cancelled <- struct{}{}
		w.WriteHeader(499)
	})
	srv := httptest.NewServer(newQueryBudget(handler, 1))
	defer srv.Close()
	id := "12345678-1234-1234-1234-123456789abc"
	done := make(chan int, 1)
	go func() {
		req, _ := http.NewRequest("POST", srv.URL+"/explore", bytes.NewBufferString("{}"))
		req.Header.Set("X-Goo-Query-ID", id)
		res, e := srv.Client().Do(req)
		if e != nil {
			done <- 0
			return
		}
		res.Body.Close()
		done <- res.StatusCode
	}()
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("not admitted")
	}
	res, e := srv.Client().Post(srv.URL+"/inventory", "application/json", bytes.NewBufferString("{}"))
	if e != nil {
		t.Fatal(e)
	}
	if res.StatusCode != 429 {
		t.Fatal(res.StatusCode)
	}
	res.Body.Close()
	res, e = srv.Client().Get(srv.URL + "/captures/local")
	if e != nil || res.StatusCode != 200 {
		t.Fatal(res, e)
	}
	res.Body.Close()
	res, e = srv.Client().Post(srv.URL+"/queries/"+id+"/cancel", "application/json", bytes.NewBufferString("{}"))
	if e != nil || res.StatusCode != 202 {
		t.Fatal(res, e)
	}
	res.Body.Close()
	select {
	case <-cancelled:
	case <-time.After(2 * time.Second):
		t.Fatal("cancel did not propagate")
	}
	if <-done != 499 {
		t.Fatal("original did not stop")
	}
	// Tombstones prevent a delayed request being admitted after cancellation.
	req, _ := http.NewRequest("POST", srv.URL+"/explore", bytes.NewBufferString("{}"))
	req.Header.Set("X-Goo-Query-ID", id)
	res, e = srv.Client().Do(req)
	if e != nil || res.StatusCode != 409 {
		t.Fatal(res, e)
	}
	io.Copy(io.Discard, res.Body)
	res.Body.Close()
}

func TestCancellationBeforeAdmission(t *testing.T) {
	called := false
	budget := newQueryBudget(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }), 1)
	id := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	w := httptest.NewRecorder()
	budget.ServeHTTP(w, httptest.NewRequest("POST", "/queries/"+id+"/cancel", bytes.NewBufferString("{}")))
	if w.Code != 202 {
		t.Fatal(w.Code)
	}
	r := httptest.NewRequest("POST", "/explore", bytes.NewBufferString("{}"))
	r.Header.Set("X-Goo-Query-ID", id)
	w = httptest.NewRecorder()
	budget.ServeHTTP(w, r)
	if w.Code != 409 || called {
		t.Fatal("cancelled request was admitted", w.Code, called)
	}
	w = httptest.NewRecorder()
	budget.ServeHTTP(w, httptest.NewRequest("POST", "/queries/not-a-query-id/cancel", nil))
	if w.Code != 400 {
		t.Fatal(w.Code)
	}
}
