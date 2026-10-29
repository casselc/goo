package main

import (
	"errors"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/httpjson"
	"github.com/casselc/goo/workbench"
	"net/http"
	"path/filepath"
)

func registerActivityViews(mux *http.ServeMux, dir string) {
	store := workbench.Store{Dir: filepath.Join(dir, "workbench")}
	captures := exploration.CaptureStore{Dir: filepath.Join(dir, "exploration")}
	fail := func(w http.ResponseWriter, e error) { httpjson.Write(w, 400, map[string]string{"error": e.Error()}) }
	mux.HandleFunc("POST /activity-views", func(w http.ResponseWriter, r *http.Request) {
		var in workbench.ActivityView
		if r.URL.RawQuery != "" {
			fail(w, errors.New("Activity publication has no query parameters"))
			return
		}
		if e := httpjson.Decode(w, r, &in, 16384, "expected one Activity view"); e != nil {
			fail(w, e)
			return
		}
		if e := in.Validate(); e != nil {
			fail(w, e)
			return
		}
		if in.BaselineCapture != "" {
			c, e := captures.Load(in.BaselineCapture)
			if e != nil {
				fail(w, fmt.Errorf("baseline capture unavailable: %w", e))
				return
			}
			q := c.Question.Explore
			if c.Question.Kind != "explore" || q == nil || q.EnvironmentSelector != nil || !q.SameEnvironment(in.Selector.Query(60000)) {
				fail(w, errors.New("baseline capture must use this environment"))
				return
			}
			if e = exploration.ValidateDurationDistribution(c.Explore); e != nil {
				fail(w, fmt.Errorf("invalid baseline distribution: %w", e))
				return
			}
		}
		v, e := store.CreateActivityView(in)
		if e != nil {
			fail(w, e)
			return
		}
		httpjson.Write(w, 201, v)
	})
	mux.HandleFunc("GET /activity-views/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			fail(w, errors.New("Activity view has no query parameters"))
			return
		}
		v, e := store.LoadActivityView(r.PathValue("id"))
		if e != nil {
			httpjson.Write(w, 404, map[string]string{"error": e.Error()})
			return
		}
		httpjson.Write(w, 200, v)
	})
	mux.HandleFunc("GET /activity-views", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		for k, v := range q {
			if len(v) != 1 || (k != "cursor" && k != "environment") {
				fail(w, errors.New("unsupported Activity view query"))
				return
			}
		}
		v, e := store.ActivityViews(q.Get("cursor"), q.Get("environment"))
		if e != nil {
			fail(w, e)
			return
		}
		httpjson.Write(w, 200, v)
	})
}
