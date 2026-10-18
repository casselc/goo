package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
)

func registerWorkbench(mux *http.ServeMux, dir string) {
	store := workbench.Store{Dir: filepath.Join(dir, "workbench")}
	captures := exploration.CaptureStore{Dir: filepath.Join(dir, "exploration")}
	reply := func(w http.ResponseWriter, code int, v any) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(v)
	}
	fail := func(w http.ResponseWriter, code int, e error) { reply(w, code, map[string]string{"error": e.Error()}) }
	decode := func(w http.ResponseWriter, r *http.Request, v any) error {
		d := json.NewDecoder(http.MaxBytesReader(w, r.Body, workbench.MaxBytes))
		d.DisallowUnknownFields()
		if e := d.Decode(v); e != nil {
			return e
		}
		if d.Decode(new(any)) != io.EOF {
			return fmt.Errorf("expected one JSON object")
		}
		return nil
	}
	mux.HandleFunc("POST /deployments", func(w http.ResponseWriter, r *http.Request) {
		var in workbench.DeploymentInput
		if e := decode(w, r, &in); e != nil {
			fail(w, 400, e)
			return
		}
		if e := in.Validate(); e != nil {
			fail(w, 400, e)
			return
		}
		m, e := store.CreateDeployment(in)
		if e != nil {
			fail(w, 500, e)
			return
		}
		reply(w, 201, m)
	})
	mux.HandleFunc("GET /deployments/{id}", func(w http.ResponseWriter, r *http.Request) {
		m, e := store.LoadDeployment(r.PathValue("id"))
		if e != nil {
			fail(w, 404, e)
			return
		}
		reply(w, 200, m)
	})
	mux.HandleFunc("GET /deployments", func(w http.ResponseWriter, r *http.Request) {
		v := r.URL.Query()
		for k, values := range v {
			if (k != "environment" && k != "fromMs" && k != "toMs") || len(values) != 1 {
				fail(w, 400, fmt.Errorf("require one environment/fromMs/toMs"))
				return
			}
		}
		from, e1 := strconv.ParseInt(v.Get("fromMs"), 10, 64)
		to, e2 := strconv.ParseInt(v.Get("toMs"), 10, 64)
		q := workbench.DeploymentScope{Environment: v.Get("environment"), FromMS: from, ToMS: to}
		if e1 != nil || e2 != nil || strconv.FormatInt(from, 10) != v.Get("fromMs") || strconv.FormatInt(to, 10) != v.Get("toMs") || q.Validate() != nil {
			fail(w, 400, fmt.Errorf("require explicit valid environment and half-open fixed window"))
			return
		}
		result, e := store.Deployments(q)
		if e != nil {
			fail(w, 500, e)
			return
		}
		reply(w, 200, result)
	})
	mux.HandleFunc("POST /documents", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Parent   string                  `json:"parent"`
			Document workbench.DocumentInput `json:"document"`
		}
		if e := decode(w, r, &in); e != nil {
			fail(w, 400, e)
			return
		}
		result, e := store.CreateDocument(in.Document, in.Parent, func(id string) error { _, e := captures.Load(id); return e })
		if e != nil {
			fail(w, 400, e)
			return
		}
		reply(w, 201, result)
	})
	mux.HandleFunc("GET /documents/{id}", func(w http.ResponseWriter, r *http.Request) {
		result, e := store.LoadDocument(r.PathValue("id"))
		if e != nil {
			fail(w, 404, e)
			return
		}
		reply(w, 200, result)
	})
	mux.HandleFunc("GET /documents", func(w http.ResponseWriter, r *http.Request) {
		v := r.URL.Query()
		for k, values := range v {
			if k != "cursor" || len(values) != 1 {
				fail(w, 400, fmt.Errorf("only one document cursor supported"))
				return
			}
		}
		result, e := store.Documents(v.Get("cursor"))
		if e != nil {
			fail(w, 400, e)
			return
		}
		reply(w, 200, result)
	})
}
