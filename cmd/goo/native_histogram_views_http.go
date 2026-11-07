package main

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/casselc/goo/internal/httpjson"
	"github.com/casselc/goo/workbench"
)

// Native view routes publish configuration only, never evaluate telemetry.
func registerNativeHistogramViews(mux *http.ServeMux, dir string) {
	store := workbench.Store{Dir: filepath.Join(dir, "workbench")}
	fail := func(w http.ResponseWriter, err error) {
		code := http.StatusBadRequest
		if errors.Is(err, os.ErrNotExist) {
			code = http.StatusNotFound
		} else if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			code = http.StatusRequestEntityTooLarge
		} else if _, ok := errors.AsType[*os.PathError](err); ok {
			code = http.StatusInternalServerError
		}
		httpjson.Write(w, code, map[string]string{"error": err.Error()})
	}
	query := func(w http.ResponseWriter, r *http.Request) bool {
		if r.URL.RawQuery != "" {
			fail(w, errors.New("native histogram view routes have no query parameters"))
			return false
		}
		return true
	}
	mux.HandleFunc("GET /native-histogram-views", func(w http.ResponseWriter, r *http.Request) {
		values, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			fail(w, errors.New("invalid native histogram view cursor query"))
			return
		}
		after := ""
		for key, items := range values {
			if key != "after" || len(items) != 1 || !workbench.ID.MatchString(items[0]) {
				fail(w, errors.New("choose one nonempty canonical after cursor, or omit it"))
				return
			}
			after = items[0]
		}
		page, err := store.NativeHistogramViews(after)
		if err != nil {
			httpjson.Write(w, http.StatusInternalServerError, map[string]string{"error": "Native histogram view catalog unavailable; direct IDs remain available."})
			return
		}
		httpjson.Write(w, http.StatusOK, page)
	})
	mux.HandleFunc("POST /native-histogram-views", func(w http.ResponseWriter, r *http.Request) {
		if !query(w, r) {
			return
		}
		var in workbench.NativeHistogramViewInput
		if err := httpjson.Decode(w, r, &in, workbench.MaxBytes, "expected one native histogram view object"); err != nil {
			fail(w, err)
			return
		}
		v, err := store.CreateNativeHistogramView(in)
		if err != nil {
			fail(w, err)
			return
		}
		httpjson.Write(w, http.StatusCreated, v)
	})
	mux.HandleFunc("GET /native-histogram-views/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !query(w, r) {
			return
		}
		v, err := store.LoadNativeHistogramView(r.PathValue("id"))
		if err != nil {
			fail(w, err)
			return
		}
		httpjson.Write(w, http.StatusOK, v)
	})
	mux.HandleFunc("POST /native-histogram-views/reuse", func(w http.ResponseWriter, r *http.Request) {
		if !query(w, r) {
			return
		}
		var in workbench.NativeHistogramViewReuseRequest
		if err := httpjson.Decode(w, r, &in, workbench.MaxBytes, "expected source view ID and destination measurement question"); err != nil {
			fail(w, err)
			return
		}
		source, err := store.LoadNativeHistogramView(in.ID)
		if err != nil {
			fail(w, err)
			return
		}
		v, err := workbench.ReuseNativeHistogramView(*source, in.Question)
		if err != nil {
			fail(w, err)
			return
		}
		httpjson.Write(w, http.StatusOK, v)
	})
}
