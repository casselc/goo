package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"time"

	apiassets "github.com/casselc/goo/api"
	"github.com/casselc/goo/definitions"
	telemetryexamples "github.com/casselc/goo/examples/telemetry"
	"github.com/casselc/goo/internal/httpjson"
	uiassets "github.com/casselc/goo/ui"
)

// registerUIAPI registers pure metadata on the same API mux as source-backed
// routes. The selected server profile is the sole UI/capability configuration.
func registerUIAPI(mux *http.ServeMux, profile string) error {
	entries, err := fs.ReadDir(telemetryexamples.Assets, ".")
	if err != nil {
		return fmt.Errorf("embedded example catalog: %w", err)
	}
	type example struct {
		Name string `json:"name"`
		YAML string `json:"yaml"`
	}
	catalog := make([]example, 0, len(entries))
	for _, entry := range entries {
		body, err := telemetryexamples.Assets.ReadFile(entry.Name())
		if err != nil {
			return err
		}
		catalog = append(catalog, example{Name: strings.TrimSuffix(entry.Name(), ".yaml"), YAML: string(body)})
	}
	metadata := func(w http.ResponseWriter, r *http.Request) bool {
		if r.URL.RawQuery != "" {
			httpjson.Write(w, 400, map[string]string{"error": "no query parameters supported"})
			return false
		}
		return true
	}
	enabled := func(w http.ResponseWriter, r *http.Request) bool {
		if !metadata(w, r) {
			return false
		}
		if profile != "lab" {
			httpjson.Write(w, 503, map[string]string{"error": "Example catalog is not configured in this runtime profile", "code": "capability_unconfigured", "capability": "example-catalog"})
			return false
		}
		return true
	}
	mux.HandleFunc("GET /ui-config", func(w http.ResponseWriter, r *http.Request) {
		if !metadata(w, r) {
			return
		}
		state := "unconfigured"
		if profile == "lab" {
			state = "configured"
		}
		httpjson.Write(w, 200, map[string]any{"version": "goo/ui-config/v1", "profile": profile, "examples": map[string]string{"state": state}})
	})
	mux.HandleFunc("GET /example-questions", func(w http.ResponseWriter, r *http.Request) {
		if enabled(w, r) {
			httpjson.Write(w, 200, catalog)
		}
	})
	mux.HandleFunc("GET /prebuilts", func(w http.ResponseWriter, r *http.Request) {
		if enabled(w, r) {
			httpjson.Write(w, 200, json.RawMessage(definitions.Prebuilts))
		}
	})
	// Liveness only. ClickHouse availability is established by source queries,
	// never a hidden read during startup or a static-resource health probe.
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if metadata(w, r) {
			httpjson.Write(w, 200, map[string]bool{"ok": true})
		}
	})
	return nil
}

// newUIHandler maps browser requests into the existing API handler in process.
// Both prefixes share admission, cancellation, snapshots and source ownership.
// There is no network hop, JSON re-encoding, response buffering or SPA fallback.
func newUIHandler(api http.Handler) (http.Handler, error) {
	mux := http.NewServeMux()
	mux.Handle("/api/", http.StripPrefix("/api", api))
	mux.Handle("/api", http.NotFoundHandler())
	serve := func(pattern string, assets fs.FS, name, contentType string) error {
		body, err := fs.ReadFile(assets, name)
		if err != nil {
			return fmt.Errorf("embedded asset %s: %w", name, err)
		}
		mux.HandleFunc("GET "+pattern, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", contentType)
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Cache-Control", "no-store")
			http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(body))
		})
		return nil
	}
	for route, name := range map[string]string{"/{$}": "index.html", "/index.html": "index.html", "/ui-components": "components.html"} {
		if err := serve(route, uiassets.Assets, name, "text/html; charset=utf-8"); err != nil {
			return nil, err
		}
	}
	entries, err := fs.ReadDir(uiassets.Assets, ".")
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		name := entry.Name()
		contentType := ""
		if strings.HasSuffix(name, ".mjs") {
			contentType = "application/javascript; charset=utf-8"
		} else if strings.HasSuffix(name, ".css") {
			contentType = "text/css; charset=utf-8"
		}
		if contentType != "" {
			if err := serve("/"+name, uiassets.Assets, name, contentType); err != nil {
				return nil, err
			}
		}
	}
	if err := serve("/api-docs", apiassets.Assets, "index.html", "text/html; charset=utf-8"); err != nil {
		return nil, err
	}
	if err := serve("/openapi.json", apiassets.Assets, "openapi.json", "application/json"); err != nil {
		return nil, err
	}
	mux.Handle("/", api)
	return mux, nil
}
