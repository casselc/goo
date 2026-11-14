package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/casselc/goo/exploration"
)

type tailTransport433 func(*http.Request) (*http.Response, error)

func (f tailTransport433) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestTailStructuredSourceRefusal433(t *testing.T) {
	for _, tc := range []struct {
		status, code int
		body, reason string
	}{
		{404, 60, "password SELECT privatehost", "source_missing"},
		{403, 0, "password SELECT privatehost", "source_access"},
		{500, 47, "password SELECT privatehost", "source_schema"},
		{503, 0, "password SELECT privatehost", "source_unavailable"},
		{200, 0, "not-json password SELECT privatehost", "source_malformed"},
	} {
		client := &http.Client{Transport: tailTransport433(func(r *http.Request) (*http.Response, error) {
			header := http.Header{}
			if tc.code != 0 {
				header.Set("X-ClickHouse-Exception-Code", fmt.Sprint(tc.code))
			}
			return &http.Response{StatusCode: tc.status, Header: header, Body: io.NopCloser(strings.NewReader(tc.body))}, nil
		})}
		mux := http.NewServeMux()
		registerTail(mux, exploration.Source{Endpoint: "http://source.invalid", Client: client}, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/tail", strings.NewReader(`{"environment":"prod"}`)))
		text := w.Body.String()
		if !strings.Contains(text, "id: 1\nevent: reset") || !strings.Contains(text, "id: 2\nevent: gap") || !strings.Contains(text, `"reason":"`+tc.reason+`"`) || !strings.Contains(text, `"retry":"explicit"`) {
			t.Fatal(text)
		}
		for _, secret := range []string{"password", "SELECT", "privatehost"} {
			if strings.Contains(text, secret) {
				t.Fatal(text)
			}
		}
	}
}
func TestTailSafeAdmissionAndCancelled433(t *testing.T) {
	mux := http.NewServeMux()
	registerTail(mux, exploration.Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: tailTransport433(func(r *http.Request) (*http.Response, error) { return nil, context.Canceled })}}, nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/tail", strings.NewReader(`{"environment":"password SELECT privatehost","sql":"secret"}`)))
	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if w.Code != 400 || body["reason"] != "invalid_selector" || body["retry"] != "explicit" || strings.Contains(w.Body.String(), "secret") {
		t.Fatal(w.Body.String())
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/tail", strings.NewReader(`{"environment":"prod"}`)).WithContext(ctx))
	if strings.Contains(w.Body.String(), "event: gap") {
		t.Fatal(w.Body.String())
	}
}

func TestTailRejectsQueryDirectAndAlias433(t *testing.T) {
	reads := 0
	mux := http.NewServeMux()
	registerTail(mux, exploration.Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: tailTransport433(func(r *http.Request) (*http.Response, error) { reads++; return nil, context.Canceled })}}, nil)
	outer := http.NewServeMux()
	outer.Handle("/api/", http.StripPrefix("/api", mux))
	outer.Handle("/", mux)
	for _, path := range []string{"/tail?sql=password", "/api/tail?sql=password"} {
		w := httptest.NewRecorder()
		outer.ServeHTTP(w, httptest.NewRequest("POST", path, strings.NewReader(`{"environment":"prod"}`)))
		if w.Code != 400 || w.Body.String() != `{"error":"Invalid tail selector.","reason":"invalid_selector","retry":"explicit"}`+"\n" {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	if reads != 0 {
		t.Fatal("query rejection read the source", reads)
	}
}

func TestTailHandlerRolloverDeadline433(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		parent, cancel := context.WithCancel(context.Background())
		defer cancel()
		reads := 0
		client := &http.Client{Transport: tailTransport433(func(r *http.Request) (*http.Response, error) {
			reads++
			<-r.Context().Done()
			return nil, r.Context().Err()
		})}
		mux := http.NewServeMux()
		registerTail(mux, exploration.Source{Endpoint: "http://source.invalid", Client: client}, nil)
		w := httptest.NewRecorder()
		started := time.Now()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/tail", strings.NewReader(`{"environment":"prod"}`)).WithContext(parent))
		if elapsed := time.Since(started); elapsed != 60*time.Second {
			t.Fatal("unexpected connection duration", elapsed)
		}
		text := w.Body.String()
		if reads != 1 || parent.Err() != nil || !strings.Contains(text, "id: 1\nevent: reset") || !strings.Contains(text, "id: 2\nevent: gap\ndata: {\"reason\":\"connection_ended\",\"retry\":\"automatic\"}") || strings.Contains(text, "source_timeout") {
			t.Fatal(reads, parent.Err(), text)
		}
	})
}
