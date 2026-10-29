package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"time"

	"github.com/casselc/goo/internal/httpjson"
	"github.com/casselc/goo/internal/nativecatalog"
	"github.com/casselc/goo/internal/nativehttp"
	"github.com/casselc/goo/measurements"
	"gopkg.in/yaml.v3"
)

func registerMeasurements(mux *http.ServeMux, source measurements.Source, dir string) {
	registerContainerMemoryView(mux)
	registerContainerCPUView(mux)
	registerContainerResourceHistory(mux)
	packagesDir := filepath.Join(dir, "measurement-packages")
	catalogStore := nativecatalog.Store{BuiltinDir: "config/measurements", SavedDir: packagesDir}
	source.CustodyDir = filepath.Join(dir, "exploration", "custody")
	nativehttp.RegisterMeasurements(mux, source)
	send := httpjson.Write
	fail := func(w http.ResponseWriter, e error) { send(w, 400, map[string]string{"error": e.Error()}) }
	decode := func(w http.ResponseWriter, r *http.Request, v any) bool {
		if e := httpjson.Decode(w, r, v, 65536, "one JSON object required"); e != nil {
			fail(w, e)
			return false
		}
		return true
	}
	item := func(d measurements.Definition, revision string) (map[string]any, error) {
		text, err := yaml.Marshal(d)
		if err != nil {
			return nil, err
		}
		return map[string]any{"definition": d, "revision": revision, "yaml": string(text)}, nil
	}
	mux.HandleFunc("GET /investigation-workflows", func(w http.ResponseWriter, r *http.Request) {
		f, e := os.Open("config/workflows.json")
		if e != nil {
			fail(w, e)
			return
		}
		defer f.Close()
		workflows, e := measurements.ReadWorkflows(f)
		if e != nil {
			fail(w, e)
			return
		}
		for _, workflow := range workflows {
			for _, step := range workflow.Steps {
				directory := "config/measurements"
				if step.Family == "telemetry" {
					directory = "examples/telemetry"
				}
				if _, e := os.Stat(filepath.Join(directory, step.Package+".yaml")); e != nil {
					fail(w, fmt.Errorf("workflow package unavailable: %s", step.Package))
					return
				}
			}
		}
		send(w, 200, workflows)
	})
	mux.HandleFunc("GET /measurement-packages", func(w http.ResponseWriter, r *http.Request) {
		entries, e := catalogStore.List(r.Context())
		if e != nil {
			fail(w, e)
			return
		}
		out := []any{}
		for _, v := range entries {
			entry, e := item(v.Definition, v.Revision)
			if e != nil {
				fail(w, e)
				return
			}
			entry["builtin"] = v.Builtin
			out = append(out, entry)
		}
		send(w, 200, out)
	})
	mux.HandleFunc("GET /measurement-packages/revisions/{revision}", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			fail(w, fmt.Errorf("exact revision lookup has no query parameters"))
			return
		}
		value, e := catalogStore.Resolve(r.Context(), r.PathValue("revision"))
		if errors.Is(e, os.ErrNotExist) {
			send(w, 404, map[string]string{"error": "native package revision not found"})
			return
		}
		if e != nil {
			fail(w, e)
			return
		}
		entry, e := item(value.Definition, value.Revision)
		if e != nil {
			fail(w, e)
			return
		}
		send(w, 200, entry)
	})

	mux.HandleFunc("POST /measurement-packages/plan", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			YAML   string `json:"yaml"`
			FromMS int64  `json:"fromMs"`
			ToMS   int64  `json:"toMs"`
			FromNS string `json:"fromNS"`
			ToNS   string `json:"toNS"`
		}
		if !decode(w, r, &input) {
			return
		}
		d, e := measurements.Read(input.YAML)
		if e != nil {
			fail(w, e)
			return
		}
		from, to := input.FromMS, input.ToMS
		if from == 0 && to == 0 && input.FromNS == "" && input.ToNS == "" {
			to = time.Now().UnixMilli()
			from = to - int64(d.LookbackMinutes)*60000
		}
		q := measurements.Question{Definition: d, FromMS: from, ToMS: to, FromNS: input.FromNS, ToNS: input.ToNS}
		p, e := measurements.Compile(q)
		if e != nil {
			fail(w, e)
			return
		}
		out, e := item(d, p.Revision)
		if e != nil {
			fail(w, e)
			return
		}
		out["question"] = q
		out["plan"] = p
		send(w, 200, out)
	})
	mux.HandleFunc("POST /measurement-packages", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			YAML string `json:"yaml"`
		}
		if !decode(w, r, &input) {
			return
		}
		d, e := measurements.Read(input.YAML)
		if e != nil {
			fail(w, e)
			return
		}
		revision, e := d.Revision()
		if e != nil {
			fail(w, e)
			return
		}
		entry, e := item(d, revision)
		if e != nil {
			fail(w, e)
			return
		}
		if e := catalogStore.Publish(r.Context(), d); e != nil {
			fail(w, e)
			return
		}

		send(w, 201, entry)
	})
}
