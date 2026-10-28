package main

import (
	"bytes"
	"context"
	"github.com/casselc/goo/exploration"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTailHTTPValidationFramesAndDisconnect(t *testing.T) {
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "{\"timestamp_ns\":\"100\",\"Body\":\"hello\"}\n")
	}))
	defer source.Close()
	mux := http.NewServeMux()
	registerTail(mux, exploration.Source{Endpoint: source.URL, Client: source.Client()}, nil)
	for _, s := range []string{`{"environment":"prod","sql":"select 1"}`, `{"environment":""}`, `{"environment":"prod","signal":"metrics"}`, `{"environment":"prod"} {}`} {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/tail", strings.NewReader(s)))
		if w.Code != 400 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	server := httptest.NewServer(mux)
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "POST", server.URL+"/tail", bytes.NewBufferString(`{"environment":"prod"}`))
	resp, e := server.Client().Do(req)
	if e != nil {
		t.Fatal(e)
	}
	defer resp.Body.Close()
	b := make([]byte, 4096)
	n, e := resp.Body.Read(b)
	if e != nil || !strings.Contains(string(b[:n]), "event: reset") || resp.Header.Get("Content-Type") != "text/event-stream" {
		t.Fatal(string(b[:n]), e)
	}
	cancel()
}

func TestTailAdmissionAndReleasedCancellation(t *testing.T) {
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { io.Copy(io.Discard, r.Body); <-r.Context().Done() }))
	defer source.Close()
	// Keep the source read pending after the initial reset so all four slots are held.

	mux := http.NewServeMux()
	registerTail(mux, exploration.Source{Endpoint: source.URL, Client: source.Client()}, nil)
	server := httptest.NewServer(mux)
	defer server.Close()
	var cancels []context.CancelFunc
	var responses []*http.Response
	defer func() {
		for _, cancel := range cancels {
			cancel()
		}
		for _, r := range responses {
			r.Body.Close()
		}
	}()
	for i := 0; i < 4; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		cancels = append(cancels, cancel)
		req, _ := http.NewRequestWithContext(ctx, "POST", server.URL+"/tail", strings.NewReader(`{"environment":"prod"}`))
		r, e := server.Client().Do(req)
		if e != nil || r.StatusCode != 200 {
			t.Fatal("admission", e)
		}
		responses = append(responses, r)
	}
	r, e := server.Client().Post(server.URL+"/tail", "application/json", strings.NewReader(`{"environment":"prod"}`))
	if e != nil {
		t.Fatal(e)
	}
	if r.StatusCode != 429 {
		t.Fatal("fifth stream admitted", r.StatusCode)
	}
	r.Body.Close()
	cancels[0]()
	responses[0].Body.Close()
	deadline := time.Now().Add(2 * time.Second)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		req, _ := http.NewRequestWithContext(ctx, "POST", server.URL+"/tail", strings.NewReader(`{"environment":"prod"}`))
		next, e := server.Client().Do(req)
		if e != nil {
			cancel()
			t.Fatal(e)
		}
		code := next.StatusCode
		cancel()
		next.Body.Close()
		if code == 200 {
			break
		}
		if code != 429 || time.Now().After(deadline) {
			t.Fatal("slot not released", code)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
