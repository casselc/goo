package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	mux.HandleFunc("GET /retained-evidence", func(w http.ResponseWriter, r *http.Request) {
		if !available(w) {
			return
		}
		q := r.URL.Query()
		for key, values := range q {
			if len(values) != 1 || (key != "tag" && key != "service" && key != "environment" && key != "from" && key != "to" && key != "cursor") {
				fail(w, 400, errors.New("unsupported retained catalog query"))
				return
			}
		}
		filter := retention.CatalogQuery{Tag: q.Get("tag"), Service: q.Get("service"), Environment: q.Get("environment"), From: q.Get("from"), To: q.Get("to"), Cursor: q.Get("cursor")}
		if e := filter.Validate(); e != nil {
			fail(w, 400, e)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		v, e := store.Catalog(ctx, filter)
		if e != nil {
			fail(w, 502, e)
			return
		}
		reply(w, 200, v)
	})
	mux.HandleFunc("GET /retained-evidence/{id}/annotations/{revision}", func(w http.ResponseWriter, r *http.Request) {
		if !available(w) {
			return
		}
		for key, values := range r.URL.Query() {
			if key != "awaitToken" || len(values) != 1 {
				fail(w, 400, errors.New("unsupported pinned review query"))
				return
			}
		}
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		v, e := (retention.Store{URL: url, AwaitToken: r.URL.Query().Get("awaitToken")}).Annotation(ctx, r.PathValue("id"), r.PathValue("revision"))
		if e != nil {
			fail(w, 404, e)
			return
		}
		reply(w, 200, v)
	})
	mux.HandleFunc("POST /retained-evidence", func(w http.ResponseWriter, r *http.Request) {
		if !available(w) {
			return
		}
		var in struct {
			RequestID   string   `json:"requestID"`
			CaptureID   string   `json:"captureID"`
			ResultToken string   `json:"resultToken"`
			Indexes     []int    `json:"indexes"`
			Tags        []string `json:"tags"`
			Note        string   `json:"note"`
		}
		if e := decode(w, r, &in); e != nil {
			fail(w, 400, e)
			return
		}
		if !workbench.ID.MatchString(in.RequestID) {
			fail(w, 400, errors.New("request ID must be 64 lowercase hex characters"))
			return
		}
		if (in.CaptureID == "") == (in.ResultToken == "") {
			fail(w, 400, errors.New("choose exactly one saved capture or live result page"))
			return
		}
		var c *exploration.SavedCapture
		var e error
		if in.ResultToken != "" {
			c, e = resultSnapshots.Load(in.ResultToken)
		} else {
			c, e = captures.Load(in.CaptureID)
		}
		if e != nil {
			// A successful earlier save remains recoverable after page expiry.
			sourceID := in.CaptureID
			if in.ResultToken != "" {
				sourceID = in.ResultToken
			}
			ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
			defer cancel()
			recovered, err := store.RecoverSelection(ctx, in.RequestID, sourceID, in.ResultToken != "", in.Indexes, in.Tags, in.Note)
			if err == nil {
				reply(w, 201, recovered)
				return
			}
			if !errors.Is(err, retention.ErrRequestNotFound) {
				fail(w, 502, fmt.Errorf("request recovery unavailable; preserve the draft and retry unchanged: %w", err))
				return
			}
			fail(w, 410, e)
			return
		}
		m, e := retention.Select(c, in.Indexes)
		if in.ResultToken != "" && e == nil {
			m, e = retention.SelectResult(c, in.Indexes)
		}
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
		for key, values := range r.URL.Query() {
			if key != "awaitToken" || len(values) != 1 || len(values[0]) > 16384 {
				fail(w, 400, errors.New("unsupported retained evidence read basis"))
				return
			}
		}
		if !available(w) {
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		v, e := (retention.Store{URL: url, AwaitToken: r.URL.Query().Get("awaitToken")}).Load(ctx, r.PathValue("id"))
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
			if len(values) != 1 || (key != "target" && key != "history" && key != "awaitToken") {
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
		v, e := (retention.Store{URL: url, AwaitToken: q.Get("awaitToken")}).Annotations(ctx, r.PathValue("id"))
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
		var in struct {
			workbench.MeaningInput
			RequestID  string `json:"requestID"`
			AwaitToken string `json:"awaitToken"`
		}
		if e := decode(w, r, &in); e != nil {
			fail(w, 400, e)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		v, e := (retention.Store{URL: url, AwaitToken: in.AwaitToken}).CorrectOnce(ctx, r.PathValue("id"), in.MeaningInput, in.RequestID)
		if e != nil {
			fail(w, 400, e)
			return
		}
		reply(w, 201, v)
	})
}
