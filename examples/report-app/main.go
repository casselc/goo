package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/casselc/goo/instrumentation"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Report struct {
	ID      string `json:"id"`
	Outcome string `json:"outcome"`
}

//goo:operation name:report.generate role:server
func Generate(ctx context.Context) (result *Report, err error) {
	result = &Report{ID: rand.Text(), Outcome: "ready"}
	emitLog(ctx, "report.accepted", result.ID, "")
	err = Ready(ctx, result)
	return result, err
}

//goo:event name:report.ready subject:report.ID outcome:report.Outcome when:ok
func Ready(ctx context.Context, report *Report) (err error) {
	emitLog(ctx, "report.ready", report.ID, report.Outcome)
	return nil
}
func emitLog(ctx context.Context, name, id, outcome string) {
	if e := instrumentation.EmitExampleLog(ctx, "lab-go-reports", "reports-demo", name, id, outcome); e != nil {
		log.Printf("example log export failed: %v", e)
	}
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	exporter, err := otlptracehttp.New(context.Background())
	if err != nil {
		return err
	}
	provider := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter), sdktrace.WithResource(resource.NewSchemaless(attribute.String("service.name", "lab-go-reports"), attribute.String("deployment.environment.name", "local"), attribute.String("service.instance.id", "reports-demo"))))
	defer shutdownProvider(provider.Shutdown)
	otel.SetTracerProvider(provider)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<!doctype html><html lang="en"><meta charset="utf-8"><title>Report generator</title><style>body{font:18px system-ui;max-width:800px;margin:50px auto;background:#0b1420;color:#d9e6f5}button{padding:15px;background:#204b45;color:white}a{color:#67ddbb}</style><h1>Report generator</h1><p>A second Go application uses the shared operation package views. A reported ready event is an observation, not a durable job guarantee.</p><button id="generate">Generate report</button><pre id="result" role="status"></pre><a href="http://localhost:8088/#operations">Investigate reports</a><script>document.querySelector('#generate').onclick=async()=>{const b=document.querySelector('#generate');b.disabled=true;try{const r=await fetch('/generate',{method:'POST'});document.querySelector('#result').textContent=await r.text()}finally{b.disabled=false}};</script></html>`))
	})
	mux.HandleFunc("POST /generate", func(w http.ResponseWriter, r *http.Request) {
		report, err := Generate(instrumentation.WithSubjectScope(r.Context(), "report", "demo"))
		if err != nil {
			http.Error(w, "generation failed", 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(report)
	})
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return serve(ctx, &http.Server{Addr: ":8091", Handler: mux, ReadHeaderTimeout: 5 * time.Second})
}
