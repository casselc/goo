package main

import (
	"flag"
	"fmt"
	"github.com/casselc/goo/annotations"
	"github.com/casselc/goo/prebuilt"
	"net/http"
	"os"
	"path/filepath"
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
	connection := clickHouseConnectionFromEnvironment(client)
	mux := http.NewServeMux()
	registerMeasurements(mux, connection.measurementSource(), *captureDir)
	registerOperationReads(mux, plans, ordered, expanded, source, connection)
	registerInvestigations(mux, plans, connection, *captureDir)
	registerDashboards(mux, ordered)
	recordSource := connection.explorationSource()
	registerExploration(mux, recordSource)
	registerCaptures(mux, recordSource, *captureDir)
	registerWorkbench(mux, *captureDir)
	registerTelemetryPackages(mux, filepath.Join(*captureDir, "telemetry-packages"))
	server := &http.Server{Addr: *listen, Handler: newQueryBudget(mux, 4), ReadHeaderTimeout: 5 * time.Second, WriteTimeout: 90 * time.Second}
	return server.ListenAndServe()
}
