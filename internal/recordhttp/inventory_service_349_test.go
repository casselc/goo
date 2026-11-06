package recordhttp_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/recordhttp"
)

type inventoryTransport349 func(*http.Request) (*http.Response, error)

func (f inventoryTransport349) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestInventoryServiceAdmission349(t *testing.T) {
	attempts := 0
	source := exploration.Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: inventoryTransport349(func(*http.Request) (*http.Response, error) {
		attempts++
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(""))}, nil
	})}}
	mux := http.NewServeMux()
	recordhttp.Register(mux, source, &exploration.ResultSnapshots{})
	for _, test := range []struct {
		path, body    string
		status, reads int
	}{
		{"/inventory", `{"environment":"prod","fromMs":1000,"toMs":61000,"service":""}`, 200, 1},
		{"/inventory", `{"environment":"prod","fromMs":1000,"toMs":61000,"service":null}`, 400, 0},
		{"/discovery", `{"environment":"prod","fromMs":1000,"toMs":61000,"service":""}`, 400, 0},
		{"/discovery", `{"environment":"prod","fromMs":1000,"toMs":61000,"service":"api"}`, 400, 0},
	} {
		before := attempts
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", test.path, strings.NewReader(test.body)))
		if w.Code != test.status || attempts-before != test.reads {
			t.Fatal("service scope admission or source boundary changed", test.path, w.Code, w.Body.String(), attempts-before)
		}
		if w.Code == 200 && !strings.Contains(w.Body.String(), `"service":""`) {
			t.Fatal("empty selected service omitted", w.Body.String())
		}
	}
}

func TestInventoryWholeEnvironmentBudgetRefusal349(t *testing.T) {
	attempts := 0
	source := exploration.Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: inventoryTransport349(func(*http.Request) (*http.Response, error) {
		attempts++
		var body strings.Builder
		for i := range 1001 {
			fmt.Fprintf(&body, `{"service":"unrelated-%d","name":"operation","count":"1","errors":"0","buckets":["1","0","0","0","0","0","0","0"]}`+"\n", i)
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body.String()))}, nil
	})}}
	mux := http.NewServeMux()
	recordhttp.Register(mux, source, &exploration.ResultSnapshots{})
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/inventory", strings.NewReader(`{"environment":"prod","fromMs":1000,"toMs":61000}`)))
	var result struct {
		Error   string
		Refusal struct{ Kind, Detail string }
	}
	if w.Code != 502 || attempts != 1 || json.Unmarshal(w.Body.Bytes(), &result) != nil || result.Refusal.Kind != "budget" || result.Error != "Source read exceeded a limit." || !strings.Contains(result.Refusal.Detail, "grouping limit") {
		t.Fatal("global cap lost budget meaning", w.Code, w.Body.String(), attempts)
	}
	if strings.Contains(w.Body.String(), "source.invalid") || strings.Contains(w.Body.String(), "unrelated-") {
		t.Fatal("source details leaked", w.Body.String())
	}
}
