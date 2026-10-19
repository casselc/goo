package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestWindowDiscoverySourceFreeAndStrictHTTP(t *testing.T) {
	b, e := os.ReadFile("../../config/measurements/checkout-calls.yaml")
	if e != nil {
		t.Fatal(e)
	}
	d, e := measurements.Read(string(b))
	if e != nil {
		t.Fatal(e)
	}
	reads := 0
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reads++
		fmt.Fprintf(w, `{"resource":{"service.name":%q,"service.instance.id":"one","deployment.environment.name":%q},"scope":%q,"version":%q,"schema_url":%q,"name":%q,"unit":%q,"start_ns":"100000000123","end_ns":"102000000123","value":"6","temporality":1,"monotonic":true}`, d.Source.Services[0], d.Environment, d.Source.Scope, d.Source.Version, d.Source.SchemaURL, d.Source.Metric, d.Source.Unit)
	}))
	defer source.Close()
	t.Setenv("CLICKHOUSE_URL", source.URL)
	mux := http.NewServeMux()
	registerCaptures(mux, source.Client(), t.TempDir())
	payload, _ := json.Marshal(map[string]any{"title": "SDK support", "question": exploration.CaptureQuestion{Kind: "measurement", Measurement: &measurements.Question{Definition: d, FromMS: 100000, ToMS: 103000}}})
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/captures", bytes.NewReader(payload)))
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	var c exploration.CapturePage
	json.Unmarshal(w.Body.Bytes(), &c)
	source.Close()
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/captures/"+c.Receipt.ID+"/windows", nil))
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var result exploration.CapturedWindowDiscovery
	json.Unmarshal(w.Body.Bytes(), &result)
	if reads != 1 || len(result.Support.Windows) != 1 || result.Support.Windows[0].FromNS != "100000000123" || result.Support.Windows[0].ToNS != "102000000123" || result.Receipt.ID != c.Receipt.ID {
		t.Fatal(reads, result)
	}
	for _, path := range []string{"/captures/" + c.Receipt.ID + "/windows?fromNS=1", "/captures/not-an-id/windows"} {
		w = httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code < 400 {
			t.Fatal("accepted invalid request", path)
		}
	}
}
