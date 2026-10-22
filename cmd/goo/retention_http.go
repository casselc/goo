package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/retention"
	"github.com/casselc/goo/workbench"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func registerRetention(mux *http.ServeMux, dir string) {
	url := os.Getenv("GOO_XTDB_URL")
	store := retention.Store{URL: url}
	captures := exploration.CaptureStore{Dir: filepath.Join(dir, "exploration")}
	reply := func(w http.ResponseWriter, code int, value any) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(value)
	}
	fail := func(w http.ResponseWriter, code int, e error) { reply(w, code, map[string]string{"error": e.Error()}) }
	available := func(w http.ResponseWriter) bool {
		if url == "" {
			fail(w, 503, errors.New("optional XTDB retention is not configured"))
			return false
		}
		return true
	}
	decode := func(w http.ResponseWriter, r *http.Request, v any) error {
		d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16384))
		d.DisallowUnknownFields()
		if e := d.Decode(v); e != nil {
			return e
		}
		if d.Decode(new(any)) != io.EOF {
			return errors.New("expected one request")
		}
		return nil
	}
	mux.HandleFunc("GET /retained-evidence/status", func(w http.ResponseWriter, r *http.Request) { reply(w, 200, map[string]bool{"configured": url != ""}) })
	mux.HandleFunc("POST /retained-evidence", func(w http.ResponseWriter, r *http.Request) {
		if !available(w) {
			return
		}
		var in struct {
			RequestID string   `json:"requestID"`
			CaptureID string   `json:"captureID"`
			Indexes   []int    `json:"indexes"`
			Tags      []string `json:"tags"`
			Note      string   `json:"note"`
		}
		if e := decode(w, r, &in); e != nil {
			fail(w, 400, e)
			return
		}
		c, e := captures.Load(in.CaptureID)
		if e != nil {
			fail(w, 400, e)
			return
		}
		m, e := retention.Select(c, in.Indexes)
		if e != nil {
			fail(w, 400, e)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		v, e := store.Publish(ctx, m, in.Tags, in.Note, in.RequestID)
		if e != nil {
			fail(w, 502, e)
			return
		}
		reply(w, 201, v)
	})
	mux.HandleFunc("GET /retained-evidence/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !available(w) {
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		v, e := store.Load(ctx, r.PathValue("id"))
		if e != nil {
			fail(w, 404, e)
			return
		}
		raw, _ := json.Marshal(v)
		reply(w, 200, struct {
			*retention.Manifest
			ManifestJSON string `json:"manifestJSON"`
		}{v, string(raw)})
	})
	mux.HandleFunc("GET /retained-evidence/{id}/annotations", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		for key, values := range q {
			if len(values) != 1 || (key != "target" && key != "history") {
				fail(w, 400, errors.New("unsupported retained annotation query"))
				return
			}
		}
		if q.Has("history") && q.Get("history") != "1" {
			fail(w, 400, errors.New("history must be 1"))
			return
		}
		if q.Has("target") {
			var target workbench.MeaningTarget
			d := json.NewDecoder(strings.NewReader(q.Get("target")))
			d.DisallowUnknownFields()
			if len(q.Get("target")) > 4096 || d.Decode(&target) != nil || d.Decode(new(any)) != io.EOF || target.Validate() != nil || target.Kind != "evidence" || target.ID != r.PathValue("id") {
				fail(w, 400, errors.New("annotation target differs from retained evidence"))
				return
			}
		}

		if !available(w) {
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		v, e := store.Annotations(ctx, r.PathValue("id"))
		if e != nil {
			fail(w, 404, e)
			return
		}
		if !q.Has("history") {
			items := []workbench.MeaningItem{}
			for _, item := range v.Items {
				if item.Head && item.Revision.State == "active" {
					items = append(items, item)
				}
			}
			v.Items = items
		}

		reply(w, 200, v)
	})
	mux.HandleFunc("POST /retained-evidence/{id}/annotations", func(w http.ResponseWriter, r *http.Request) {
		if !available(w) {
			return
		}
		var in workbench.MeaningInput
		if e := decode(w, r, &in); e != nil {
			fail(w, 400, e)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		v, e := store.Correct(ctx, r.PathValue("id"), in)
		if e != nil {
			fail(w, 400, e)
			return
		}
		reply(w, 201, v)
	})
}
