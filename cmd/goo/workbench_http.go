package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/retention"
	"github.com/casselc/goo/workbench"
)

func registerWorkbench(mux *http.ServeMux, dir string) {
	registerRetention(mux, dir)
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
	mux.HandleFunc("POST /annotations", func(w http.ResponseWriter, r *http.Request) {
		var in workbench.MeaningInput
		if e := decode(w, r, &in); e != nil {
			fail(w, 400, e)
			return
		}
		v, e := store.CreateMeaning(in, func(id string) error {
			_, e := (exploration.PackageStore{Dir: filepath.Join(dir, "telemetry-packages")}).Load(id)
			return e
		})
		if e != nil {
			fail(w, 400, e)
			return
		}
		reply(w, 201, v)
	})
	mux.HandleFunc("GET /annotations/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			fail(w, 400, fmt.Errorf("revision has no query parameters"))
			return
		}
		v, e := store.LoadMeaning(r.PathValue("id"))
		if e != nil {
			fail(w, 404, e)
			return
		}
		reply(w, 200, v)
	})
	mux.HandleFunc("GET /annotations", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		for k, v := range q {
			if len(v) != 1 || (k != "target" && k != "tag" && k != "cursor" && k != "history") {
				fail(w, 400, fmt.Errorf("unsupported annotation query"))
				return
			}
		}
		if q.Has("history") && q.Get("history") != "1" {
			fail(w, 400, fmt.Errorf("history must be 1"))
			return
		}
		var target *workbench.MeaningTarget
		if q.Has("target") {
			target = new(workbench.MeaningTarget)
			d := json.NewDecoder(strings.NewReader(q.Get("target")))
			d.DisallowUnknownFields()
			if len(q.Get("target")) > 4096 || d.Decode(target) != nil || d.Decode(new(any)) != io.EOF {
				fail(w, 400, fmt.Errorf("invalid typed target"))
				return
			}
		}
		v, e := store.Meanings(target, q.Get("tag"), q.Get("cursor"), q.Has("history"))
		if e != nil {
			fail(w, 400, e)
			return
		}
		reply(w, 200, v)
	})
	mux.HandleFunc("POST /population/plan", func(w http.ResponseWriter, r *http.Request) {
		var in exploration.PopulationRequest
		if e := decode(w, r, &in); e != nil {
			fail(w, 400, e)
			return
		}
		p, e := exploration.PlanPopulation(in)
		if e != nil {
			fail(w, 400, e)
			return
		}
		reply(w, 200, p)
	})
	mux.HandleFunc("POST /exploration-views", func(w http.ResponseWriter, r *http.Request) {
		var in workbench.ExplorationView
		if e := decode(w, r, &in); e != nil {
			fail(w, 400, e)
			return
		}
		v, e := store.CreateExplorationView(in)
		if e != nil {
			fail(w, 400, e)
			return
		}
		reply(w, 201, v)
	})
	mux.HandleFunc("GET /exploration-views/{id}", func(w http.ResponseWriter, r *http.Request) {
		v, e := store.LoadExplorationView(r.PathValue("id"))
		if e != nil {
			fail(w, 404, e)
			return
		}
		reply(w, 200, v)
	})
	mux.HandleFunc("GET /exploration-views", func(w http.ResponseWriter, r *http.Request) {
		for key, values := range r.URL.Query() {
			if key != "cursor" || len(values) != 1 {
				fail(w, 400, fmt.Errorf("only one view cursor supported"))
				return
			}
		}
		v, e := store.ExplorationViews(r.URL.Query().Get("cursor"))
		if e != nil {
			fail(w, 400, e)
			return
		}
		reply(w, 200, v)
	})

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
	mux.HandleFunc("POST /captures/comparisons", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Baseline  string                          `json:"baseline"`
			Selection string                          `json:"selection"`
			Expected  *exploration.EvidenceComparison `json:"expected"`
		}
		if e := decode(w, r, &in); e != nil {
			fail(w, 400, e)
			return
		}
		receipt, e := store.CreateComparison(in.Baseline, in.Selection, in.Expected, captures.Load)
		if e != nil {
			code := 400
			if errors.Is(e, workbench.ErrComparisonChanged) {
				code = 409
			}
			fail(w, code, e)
			return
		}
		reply(w, 201, receipt)
	})
	mux.HandleFunc("GET /captures/comparisons/{id}/receipt", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			fail(w, 400, fmt.Errorf("stored comparison has no query parameters"))
			return
		}
		receipt, e := store.LoadComparison(r.PathValue("id"))
		if e != nil {
			fail(w, 404, e)
			return
		}
		reply(w, 200, receipt)
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
		result, e := store.CreateDocument(in.Document, in.Parent, func(id string) error { _, e := captures.Load(id); return e }, func(card workbench.Card) error {
			if os.Getenv("GOO_XTDB_URL") == "" {
				return fmt.Errorf("optional retention is not configured")
			}
			ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
			defer cancel()
			_, e := (retention.Store{URL: os.Getenv("GOO_XTDB_URL"), AwaitToken: card.ReadBasis}).Annotation(ctx, card.Reference, card.Review)
			return e
		})
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
