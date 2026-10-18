package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/casselc/goo/annotations"
	"github.com/casselc/goo/prebuilt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func serveHTTP(args []string) error {
	fs := flag.NewFlagSet("serve-http", flag.ContinueOnError)
	packagePaths := []string{}
	fs.Func("operation-package", "repeatable operation package YAML path", func(v string) error { packagePaths = append(packagePaths, v); return nil })
	binding := fs.String("binding", "", "package binding YAML")
	inventory := fs.String("inventory", "", "declared versioned inventory JSON")
	listen := fs.String("listen", "127.0.0.1:8089", "listen address")
	captureDir := fs.String("investigations-dir", "artifacts/investigations", "immutable capture directory")
	scope := fs.String("scope-name", "", "explicit instrumentation library")
	version := fs.String("scope-version", "", "explicit instrumentation version")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *binding == "" || *inventory == "" || *scope == "" || *version == "" || fs.NArg() != 0 {
		return fmt.Errorf("require binding, inventory, scope-name and scope-version")
	}
	bf, err := os.Open(*binding)
	if err != nil {
		return err
	}
	defer bf.Close()
	b, err := prebuilt.ReadBinding(bf)
	if err != nil {
		return err
	}
	inf, err := os.Open(*inventory)
	if err != nil {
		return err
	}
	defer inf.Close()
	in, err := prebuilt.ReadInventory(inf)
	if err != nil {
		return err
	}
	expanded, err := prebuilt.Expand(in, b)
	if err != nil {
		return err
	}
	if len(packagePaths) == 0 {
		packagePaths = []string{"config/operations/checkout.yaml"}
	}
	plans := map[string]*annotations.OperationPlan{}
	ordered := []*annotations.OperationPlan{}
	for _, path := range packagePaths {
		plan, e := annotations.LoadOperationPackage(path)
		if e != nil {
			return fmt.Errorf("operation package %s: %w", path, e)
		}
		name := plan.Definition.Name
		if plans[name] != nil {
			return fmt.Errorf("duplicate package name %s", name)
		}
		plans[name] = plan
		ordered = append(ordered, plan)
	}
	source := prebuilt.HTTPSource{ScopeName: *scope, ScopeVersion: *version, SchemaURL: "https://opentelemetry.io/schemas/1.37.0"}
	client := &http.Client{Timeout: 15 * time.Second}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /http-package", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		minutes := 5
		if value := r.URL.Query().Get("minutes"); value != "" {
			parsed, e := strconv.Atoi(value)
			if e != nil || parsed < 1 || parsed > 120 {
				http.Error(w, `{"error":"minutes must be 1..120"}`, 400)
				return
			}
			minutes = parsed
		}
		for key, values := range r.URL.Query() {
			if key != "minutes" || len(values) != 1 {
				http.Error(w, `{"error":"unsupported query parameter"}`, 400)
				return
			}
		}
		to := time.Now().UnixMilli()
		endpoint := os.Getenv("CLICKHOUSE_URL")
		if endpoint == "" {
			endpoint = "http://localhost:8123"
		}
		result, e := prebuilt.FetchHTTP(r.Context(), client, endpoint, os.Getenv("CLICKHOUSE_USER"), os.Getenv("CLICKHOUSE_PASSWORD"), expanded, source, to-int64(minutes)*60000, to)
		if e != nil {
			w.WriteHeader(502)
			json.NewEncoder(w).Encode(map[string]string{"error": e.Error()})
			return
		}
		json.NewEncoder(w).Encode(result)
	})
	mux.HandleFunc("GET /operation-packages", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		if len(r.URL.Query()) != 0 {
			http.Error(w, `{"error":"no query parameters supported"}`, 400)
			return
		}
		json.NewEncoder(w).Encode(ordered)
	})
	mux.HandleFunc("GET /operations", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		minutes := 5
		for key, values := range r.URL.Query() {
			if (key != "minutes" && key != "package") || len(values) != 1 {
				http.Error(w, `{"error":"unsupported query parameter"}`, 400)
				return
			}
		}
		if v := r.URL.Query().Get("minutes"); v != "" {
			n, e := strconv.Atoi(v)
			if e != nil || n < 1 || n > 120 {
				http.Error(w, `{"error":"minutes must be 1..120"}`, 400)
				return
			}
			minutes = n
		}
		name := r.URL.Query().Get("package")
		if name == "" {
			name = ordered[0].Definition.Name
		}
		plan := plans[name]
		if plan == nil {
			http.Error(w, `{"error":"unknown operation package"}`, 400)
			return
		}
		to := time.Now().UnixMilli()
		endpoint := os.Getenv("CLICKHOUSE_URL")
		if endpoint == "" {
			endpoint = "http://localhost:8123"
		}
		result, e := plan.Fetch(r.Context(), client, endpoint, os.Getenv("CLICKHOUSE_USER"), os.Getenv("CLICKHOUSE_PASSWORD"), to-int64(minutes)*60000, to)
		if e != nil {
			w.WriteHeader(502)
			json.NewEncoder(w).Encode(map[string]string{"error": e.Error()})
			return
		}
		json.NewEncoder(w).Encode(result)
	})
	mux.HandleFunc("GET /counter-rates", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		for key, values := range r.URL.Query() {
			if (key != "package" && key != "fromMs" && key != "toMs") || len(values) != 1 {
				http.Error(w, `{"error":"unsupported rate query parameter"}`, 400)
				return
			}
		}
		plan := plans[r.URL.Query().Get("package")]
		from, e1 := strconv.ParseInt(r.URL.Query().Get("fromMs"), 10, 64)
		to, e2 := strconv.ParseInt(r.URL.Query().Get("toMs"), 10, 64)
		if plan == nil || e1 != nil || e2 != nil || from < 60000 || to <= from || to-from > 120*60000 {
			http.Error(w, `{"error":"require package and bounded fromMs/toMs"}`, 400)
			return
		}
		endpoint := os.Getenv("CLICKHOUSE_URL")
		if endpoint == "" {
			endpoint = "http://localhost:8123"
		}
		result, e := plan.FetchRates(r.Context(), client, endpoint, os.Getenv("CLICKHOUSE_USER"), os.Getenv("CLICKHOUSE_PASSWORD"), from, to)
		if e != nil {
			w.WriteHeader(502)
			json.NewEncoder(w).Encode(map[string]string{"error": e.Error()})
			return
		}
		json.NewEncoder(w).Encode(result)
	})
	mux.HandleFunc("GET /histograms", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		for key, values := range r.URL.Query() {
			if (key != "package" && key != "fromMs" && key != "toMs") || len(values) != 1 {
				http.Error(w, `{"error":"unsupported histogram query parameter"}`, 400)
				return
			}
		}
		plan := plans[r.URL.Query().Get("package")]
		from, e1 := strconv.ParseInt(r.URL.Query().Get("fromMs"), 10, 64)
		to, e2 := strconv.ParseInt(r.URL.Query().Get("toMs"), 10, 64)
		if plan == nil || e1 != nil || e2 != nil || from < 60000 || to <= from || to-from > 120*60000 {
			http.Error(w, `{"error":"require package and bounded fromMs/toMs"}`, 400)
			return
		}
		endpoint := os.Getenv("CLICKHOUSE_URL")
		if endpoint == "" {
			endpoint = "http://localhost:8123"
		}
		result, e := plan.FetchHistograms(r.Context(), client, endpoint, os.Getenv("CLICKHOUSE_USER"), os.Getenv("CLICKHOUSE_PASSWORD"), from, to)
		if e != nil {
			w.WriteHeader(502)
			json.NewEncoder(w).Encode(map[string]string{"error": e.Error()})
			return
		}
		json.NewEncoder(w).Encode(result)
	})
	registerInvestigations(mux, plans, client, *captureDir)
	registerExploration(mux, client)
	registerCaptures(mux, client, *captureDir)
	registerWorkbench(mux, *captureDir)
	registerTelemetryPackages(mux, filepath.Join(*captureDir, "telemetry-packages"))
	server := &http.Server{Addr: *listen, Handler: newQueryBudget(mux, 4), ReadHeaderTimeout: 5 * time.Second, WriteTimeout: 90 * time.Second}
	return server.ListenAndServe()
}
