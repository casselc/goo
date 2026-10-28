package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/casselc/goo/internal/httpjson"
	"github.com/casselc/goo/internal/immutablefile"
	"github.com/casselc/goo/internal/nativehttp"
	"github.com/casselc/goo/measurements"
	"gopkg.in/yaml.v3"
)

func registerMeasurements(mux *http.ServeMux, source measurements.Source, dir string) {
	registerContainerMemoryView(mux)
	registerContainerCPUView(mux)
	registerContainerResourceHistory(mux)
	packagesDir := filepath.Join(dir, "measurement-packages")
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
		revision := measurements.Hash(d)
		created, e := immutablefile.PublishFileOnly(packagesDir, revision+".yaml", ".package-", b)
		if e != nil {
			fail(w, e)
			return
		}
		if !created {
			if e := verifyMeasurementPackageWinner(filepath.Join(packagesDir, revision+".yaml"), revision); e != nil {
				fail(w, e)
				return
			}
		}
		send(w, 201, item(d))
	})
}

// A duplicate publication is successful only when its immutable winner remains
// a valid package with the requested identity. Never repair or replace it here.
func verifyMeasurementPackageWinner(path, revision string) error {
	const limit = 32768
	invalid := func() error {
		return fmt.Errorf("existing measurement package is unavailable, invalid, or has a different revision")
	}
	info, e := os.Lstat(path)
	if e != nil || !info.Mode().IsRegular() || info.Size() > limit {
		return invalid()
	}
	f, e := os.Open(path)
	if e != nil {
		return invalid()
	}
	defer f.Close()
	info, e = f.Stat()
	if e != nil || !info.Mode().IsRegular() || info.Size() > limit {
		return invalid()
	}
	b, e := io.ReadAll(io.LimitReader(f, limit+1))
	if e != nil || len(b) > limit {
		return invalid()
	}
	d, e := measurements.Read(string(b))
	if e != nil || measurements.Hash(d) != revision {
		return invalid()
	}
	return nil
}
