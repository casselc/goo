package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/measurements"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func registerMeasurements(mux *http.ServeMux, client *http.Client, dir string) {
	packagesDir := filepath.Join(dir, "measurement-packages")
	custodyDir := filepath.Join(dir, "exploration", "custody")
	send := func(w http.ResponseWriter, code int, v any) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(v)
	}
	fail := func(w http.ResponseWriter, e error) { send(w, 400, map[string]string{"error": e.Error()}) }
	decode := func(w http.ResponseWriter, r *http.Request, v any) bool {
		d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 65536))
		d.DisallowUnknownFields()
		if e := d.Decode(v); e != nil {
			fail(w, e)
			return false
		}
		if d.Decode(new(any)) != io.EOF {
			fail(w, fmt.Errorf("one JSON object required"))
			return false
		}
		return true
	}
	item := func(d measurements.Definition) map[string]any {
		text, _ := yaml.Marshal(d)
		return map[string]any{"definition": d, "revision": measurements.Hash(d), "yaml": string(text)}
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
		all := map[string]measurements.Definition{}
		builtin := map[string]bool{}
		count := 0
		for _, directory := range []string{"config/measurements", packagesDir} {
			catalog, e := os.Open(directory)
			if os.IsNotExist(e) {
				continue
			}
			if e != nil {
				fail(w, e)
				return
			}
			files, e := catalog.ReadDir(257)
			catalog.Close()
			if len(files) > 256 {
				fail(w, fmt.Errorf("package catalog scan limit exceeded"))
				return
			}
			if e != nil && e != io.EOF {
				fail(w, e)
				return
			}
			for _, f := range files {
				if f.IsDir() || filepath.Ext(f.Name()) != ".yaml" {
					continue
				}
				count++
				if count > 256 {
					fail(w, fmt.Errorf("package catalog limit exceeded"))
					return
				}
				info, e := f.Info()
				if e != nil || info.Size() > 32768 {
					fail(w, fmt.Errorf("package file exceeds limit or unavailable"))
					return
				}
				b, e := os.ReadFile(filepath.Join(directory, f.Name()))
				if e != nil {
					fail(w, e)
					return
				}
				d, e := measurements.Read(string(b))
				if e != nil {
					fail(w, e)
					return
				}
				revision := measurements.Hash(d)
				all[revision] = d
				if directory == "config/measurements" {
					builtin[revision] = true
				}
			}
		}
		ids := []string{}
		for id := range all {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		out := []any{}
		for _, id := range ids {
			entry := item(all[id])
			entry["builtin"] = builtin[id]
			out = append(out, entry)
		}
		send(w, 200, out)
	})
	mux.HandleFunc("POST /measurement-packages/plan", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			YAML   string `json:"yaml"`
			FromMS int64  `json:"fromMs"`
			ToMS   int64  `json:"toMs"`
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
		if from == 0 && to == 0 {
			to = time.Now().UnixMilli()
			from = to - int64(d.LookbackMinutes)*60000
		}
		q := measurements.Question{Definition: d, FromMS: from, ToMS: to}
		p, e := measurements.Compile(q)
		if e != nil {
			fail(w, e)
			return
		}
		out := item(d)
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
		b, _ := yaml.Marshal(d)
		if e = os.MkdirAll(packagesDir, 0700); e != nil {
			fail(w, e)
			return
		}
		f, e := os.CreateTemp(packagesDir, ".package-")
		if e != nil {
			fail(w, e)
			return
		}
		name := f.Name()
		defer os.Remove(name)
		if _, e = f.Write(b); e == nil {
			e = f.Sync()
		}
		closeErr := f.Close()
		if e == nil {
			e = closeErr
		}
		if e != nil {
			fail(w, e)
			return
		}
		if e = os.Link(name, filepath.Join(packagesDir, measurements.Hash(d)+".yaml")); e != nil && !os.IsExist(e) {
			fail(w, e)
			return
		}
		send(w, 201, item(d))
	})
	mux.HandleFunc("POST /measurements", func(w http.ResponseWriter, r *http.Request) {
		var q measurements.Question
		if !decode(w, r, &q) {
			return
		}
		if e := q.Validate(); e != nil {
			fail(w, e)
			return
		}
		endpoint := os.Getenv("CLICKHOUSE_URL")
		if endpoint == "" {
			endpoint = "http://localhost:8123"
		}
		source := measurements.Source{Endpoint: endpoint, User: os.Getenv("CLICKHOUSE_USER"), Password: os.Getenv("CLICKHOUSE_PASSWORD"), Client: client, CustodyDir: custodyDir}
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		result, e := source.Run(ctx, q, nil)
		if e != nil {
			send(w, 502, map[string]string{"error": e.Error()})
			return
		}
		send(w, 200, result)
	})
}
