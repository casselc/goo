package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/casselc/goo/annotations"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/capturehttp"
	"github.com/casselc/goo/prebuilt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func serveHTTP(args []string) error {
	server, err := configuredHTTPServer(args)
	if err != nil {
		return err
	}
	return server.ListenAndServe()
}

// configuredHTTPServer validates configuration and registers routes without
// listening or probing the telemetry source.
func configuredHTTPServer(args []string) (*http.Server, error) {
	fs := flag.NewFlagSet("serve-http", flag.ContinueOnError)
	profile := fs.String("profile", "lab", "startup profile: lab or generic")
	packagePaths := []string{}
	fs.Func("operation-package", "repeatable operation package YAML path", func(v string) error { packagePaths = append(packagePaths, v); return nil })
	binding := fs.String("binding", "", "package binding YAML")
	inventory := fs.String("inventory", "", "declared versioned inventory JSON")
	listen := fs.String("listen", "127.0.0.1:8089", "listen address")
	captureDir := fs.String("investigations-dir", "artifacts/investigations", "immutable capture directory")
	scope := fs.String("scope-name", "", "explicit instrumentation library")
	version := fs.String("scope-version", "", "explicit instrumentation version")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if *profile != "lab" && *profile != "generic" {
		return nil, fmt.Errorf("profile must be lab or generic")
	}
	configuredPrepared := *binding != "" && *inventory != "" && *scope != "" && *version != ""
	anyPrepared := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "binding" || f.Name == "inventory" || f.Name == "scope-name" || f.Name == "scope-version" {
			anyPrepared = true
		}
	})
	if fs.NArg() != 0 || (*profile == "lab" || anyPrepared) && !configuredPrepared {
		return nil, fmt.Errorf("require binding, inventory, scope-name and scope-version")
	}
	var expanded *prebuilt.Expansion
	if configuredPrepared {
		bf, err := os.Open(*binding)
		if err != nil {
			return nil, err
		}
		defer bf.Close()
		b, err := prebuilt.ReadBinding(bf)
		if err != nil {
			return nil, err
		}
		inf, err := os.Open(*inventory)
		if err != nil {
			return nil, err
		}
		defer inf.Close()
		in, err := prebuilt.ReadInventory(inf)
		if err != nil {
			return nil, err
		}
		expanded, err = prebuilt.Expand(in, b)
		if err != nil {
			return nil, err
		}
	}
	if len(packagePaths) == 0 && *profile == "lab" {
		packagePaths = []string{"config/operations/checkout.yaml"}
	}
	plans := map[string]*annotations.OperationPlan{}
	ordered := []*annotations.OperationPlan{}
	for _, path := range packagePaths {
		plan, e := annotations.LoadOperationPackage(path)
		if e != nil {
			return nil, fmt.Errorf("operation package %s: %w", path, e)
		}
		name := plan.Definition.Name
		if plans[name] != nil {
			return nil, fmt.Errorf("duplicate package name %s", name)
		}
		plans[name] = plan
		ordered = append(ordered, plan)
	}
	source := prebuilt.HTTPSource{ScopeName: *scope, ScopeVersion: *version, SchemaURL: "https://opentelemetry.io/schemas/1.37.0"}
	client := &http.Client{Timeout: 15 * time.Second}
	connection := clickHouseConnectionFromEnvironment(client)
	mux := http.NewServeMux()
	registerMeasurements(mux, connection.measurementSource(), *captureDir)
	operationMux := http.NewServeMux()
	registerOperationReads(operationMux, plans, ordered, expanded, source, connection)
	for _, path := range []string{"/http-package", "/operation-packages", "/operations", "/counter-rates", "/histograms"} {
		mux.Handle("GET "+path, runtimeOperationHandler(operationMux, path, configuredPrepared, len(ordered) > 0))
	}
	registerRuntimeCapabilities(mux, *profile, configuredPrepared, len(ordered) > 0)
	registerInvestigations(mux, plans, connection, *captureDir)
	registerDashboards(mux, ordered)
	recordSource := connection.explorationSource()
	// Ephemeral capabilities belong to this server, never another configured mux.
	resultSnapshots := &exploration.ResultSnapshots{}
	tailMembership := exploration.NewTailMembershipSnapshots(time.Now)
	registerExploration(mux, recordSource, resultSnapshots, tailMembership)
	capturehttp.Register(mux, recordSource, *captureDir, resultSnapshots)
	registerWorkbench(mux, *captureDir, retentionConnectionFromEnvironment(), resultSnapshots, tailMembership)
	registerTelemetryPackages(mux, filepath.Join(*captureDir, "telemetry-packages"))
	server := &http.Server{Addr: *listen, Handler: newQueryBudget(mux, 4), ReadHeaderTimeout: 5 * time.Second, WriteTimeout: 90 * time.Second}
	return server, nil
}

func registerRuntimeCapabilities(mux *http.ServeMux, profile string, prepared, operations bool) {
	mux.HandleFunc("GET /capabilities", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		if len(r.URL.Query()) != 0 {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{"error": "no query parameters supported"})
			return
		}
		state := func(configured bool) map[string]string {
			if configured {
				return map[string]string{"state": "configured"}
			}
			return map[string]string{"state": "unconfigured"}
		}
		json.NewEncoder(w).Encode(map[string]any{"version": "goo/runtime-capabilities/v1", "profile": profile, "preparedHTTP": state(prepared), "declaredOperations": state(operations)})
	})
}
func runtimeOperationHandler(next http.Handler, path string, prepared, operations bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capability := ""
		if path == "/http-package" && !prepared {
			capability = "prepared-http"
		}
		if path != "/http-package" && !operations {
			capability = "declared-operations"
		}
		if capability != "" {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "no-store")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"error": "capability is not configured for this runtime", "code": "capability_unconfigured", "capability": capability})
			return
		}
		next.ServeHTTP(w, r)
	})
}
