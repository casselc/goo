package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/casselc/goo/exploration"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"time"
)

func registerCaptures(mux *http.ServeMux, client *http.Client, dir string) {
	store := exploration.CaptureStore{Dir: filepath.Join(dir, "exploration")}
	reply := func(w http.ResponseWriter, code int, value any) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(value)
	}
	fail := func(w http.ResponseWriter, code int, e error) { reply(w, code, map[string]string{"error": e.Error()}) }
	cursor := func(r *http.Request) (string, error) {
		v := r.URL.Query()
		for key, values := range v {
			if key != "cursor" || len(values) != 1 {
				return "", fmt.Errorf("only one cursor parameter is supported")
			}
		}
		return v.Get("cursor"), nil
	}
	mux.HandleFunc("GET /captures", func(w http.ResponseWriter, r *http.Request) {
		after, e := cursor(r)
		if e != nil {
			fail(w, 400, e)
			return
		}
		result, e := store.List(after)
		if e != nil {
			fail(w, 500, e)
			return
		}
		reply(w, 200, result)
	})
	mux.HandleFunc("GET /captures/{id}/receipt", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			fail(w, 400, fmt.Errorf("receipt metadata has no query parameters"))
			return
		}
		c, e := store.Load(r.PathValue("id"))
		if e != nil {
			fail(w, 404, e)
			return
		}
		revision := ""
		if c.Measurement != nil {
			revision = c.Measurement.Plan.Revision
		}
		reply(w, 200, map[string]any{"receipt": c.CaptureReceipt, "definitionRevision": revision})
	})
	mux.HandleFunc("GET /captures/{id}", func(w http.ResponseWriter, r *http.Request) {
		cur, e := cursor(r)
		if e != nil {
			fail(w, 400, e)
			return
		}
		c, e := store.Load(r.PathValue("id"))
		if e != nil {
			fail(w, 404, e)
			return
		}
		p, e := c.Page(cur)
		if e != nil {
			fail(w, 400, e)
			return
		}
		reply(w, 200, p)
	})
	mux.HandleFunc("POST /captures", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Title    string                      `json:"title"`
			Question exploration.CaptureQuestion `json:"question"`
			Parent   string                      `json:"parent"`
		}
		d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16384))
		d.DisallowUnknownFields()
		if e := d.Decode(&in); e != nil {
			fail(w, 400, e)
			return
		}
		if d.Decode(new(any)) != io.EOF {
			fail(w, 400, fmt.Errorf("expected one JSON object"))
			return
		}
		if in.Parent != "" {
			p, e := store.Load(in.Parent)
			if e != nil {
				fail(w, 404, e)
				return
			}
			if in.Question.Kind != "" {
				if e := in.Question.Validate(); e != nil {
					fail(w, 400, e)
					return
				}
			}
			if in.Question.Kind != "" && !reflect.DeepEqual(in.Question.Normalized(), p.Question) {
				fail(w, 409, fmt.Errorf("refresh cannot change the captured question"))
				return
			}
			if in.Question.Kind == "" && (in.Question.Explore != nil || in.Question.Inventory != nil || in.Question.Comparison != nil || in.Question.Measurement != nil) {
				fail(w, 400, fmt.Errorf("partial refresh question is invalid"))
				return
			}
			in.Question = p.Question
			if in.Title == "" {
				in.Title = p.Title
			}
		}
		if e := exploration.ValidateCaptureTitle(in.Title); e != nil {
			fail(w, 400, e)
			return
		}
		if e := in.Question.Validate(); e != nil {
			fail(w, 400, e)
			return
		}
		endpoint := os.Getenv("CLICKHOUSE_URL")
		if endpoint == "" {
			endpoint = "http://localhost:8123"
		}
		source := exploration.Source{Endpoint: endpoint, User: os.Getenv("CLICKHOUSE_USER"), Password: os.Getenv("CLICKHOUSE_PASSWORD"), Client: client}
		ctx, cancel := context.WithTimeout(r.Context(), 55*time.Second)
		defer cancel()
		c, e := store.Create(ctx, source, in.Title, in.Question, in.Parent)
		if e != nil {
			if errors.Is(e, exploration.ErrCapturePlanChanged) {
				fail(w, 409, e)
			} else {
				fail(w, 502, e)
			}
			return
		}
		p, e := c.Page("")
		if e != nil {
			fail(w, 500, e)
			return
		}
		reply(w, 201, p)
	})
}
