package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"github.com/casselc/goo/instrumentation"
	"log"
	"net/http"
	"regexp"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	exporter, err := otlptracehttp.New(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	appResource := resource.NewSchemaless(
		attribute.String("service.name", "lab-go-checkout"), attribute.String("deployment.environment.name", "local"), attribute.String("service.instance.id", "woven-demo"))
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter), sdktrace.WithResource(appResource))
	metricExporter, e := otlpmetrichttp.New(context.Background())
	if e != nil {
		log.Fatal(e)
	}
	meter := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(2*time.Second))), sdkmetric.WithResource(appResource))
	otel.SetMeterProvider(meter)
	otel.SetTracerProvider(tp)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Woven Go checkout</title><style>body{font:18px system-ui;max-width:800px;margin:50px auto;background:#0b1420;color:#d9e6f5}button{padding:15px;margin:8px;background:#204b45;color:white;border:1px solid #67ddbb}a{color:#67ddbb}pre{white-space:pre-wrap}</style><h1>Woven Go checkout</h1><p>Run real Go invocation paths. Operation spans are exported through the collector into ClickHouse. These calls demonstrate instrumentation, not durable business transactions.</p><label>Subject ID (empty creates a new one) <input id="subject"></label><label>Tenant <select id="tenant"><option>demo</option><option>other</option></select></label><button data-scenario="ok">Successful checkout</button><button data-scenario="error">Fail checkout</button><button data-scenario="panic">Panic checkout</button><pre id="result" role="status"></pre><a href="http://localhost:8088/#operations">Investigate operations</a><script>for(const b of document.querySelectorAll('button'))b.onclick=async()=>{b.disabled=true;try{const r=await fetch('/invoke?scenario='+b.dataset.scenario+'&subject='+encodeURIComponent(document.querySelector('#subject').value)+'&tenant='+encodeURIComponent(document.querySelector('#tenant').value),{method:'POST'});document.querySelector('#result').textContent='HTTP '+r.status+' '+await r.text()}finally{b.disabled=false}};</script></html>`))
	})
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	mux.HandleFunc("POST /invoke", func(w http.ResponseWriter, r *http.Request) {
		scenario := r.URL.Query().Get("scenario")
		if scenario != "ok" && scenario != "error" && scenario != "panic" {
			http.Error(w, "scenario must be ok, error or panic", 400)
			return
		}
		id := r.URL.Query().Get("subject")
		if id == "" {
			id = rand.Text()
		}
		tenant := r.URL.Query().Get("tenant")
		if tenant == "" {
			tenant = "demo"
		}
		if !regexp.MustCompile(`^[A-Za-z0-9_.-]{1,128}$`).MatchString(id) || (tenant != "demo" && tenant != "other") {
			http.Error(w, "bounded subject and demo/other tenant required", 400)
			return
		}
		ctx := instrumentation.WithSubjectScope(context.WithValue(r.Context(), subjectIDKey{}, id), "reservation", tenant)
		w.Header().Set("Content-Type", "application/json")
		defer func() {
			if value := recover(); value != nil {
				w.WriteHeader(500)
				json.NewEncoder(w).Encode(map[string]string{"scenario": scenario, "outcome": "panic"})
			}
		}()
		if scenario == "panic" {
			Panic(ctx)
			return
		}
		_, err := Reserve(ctx, scenario == "error")
		if err != nil {
			w.WriteHeader(503)
		}
		json.NewEncoder(w).Encode(map[string]string{"scenario": scenario, "id": id, "tenant": tenant})
	})
	log.Fatal((&http.Server{Addr: ":8090", Handler: mux, ReadHeaderTimeout: 5 * time.Second}).ListenAndServe())
}
