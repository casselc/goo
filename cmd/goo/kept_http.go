package main

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/httpjson"
)

func registerKeptObservations(mux *http.ServeMux, dir string, bank *exploration.TailMembershipSnapshots, results *exploration.ResultSnapshots) {
	store := exploration.KeptObservationStore{Dir: filepath.Join(dir, "kept-observations")}
	fail := func(w http.ResponseWriter, code int, err error) {
		httpjson.Write(w, code, map[string]string{"error": err.Error()})
	}
	mux.HandleFunc("POST /kept-observations/tail", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			fail(w, 400, errors.New("keep has no query parameters"))
			return
		}
		var in exploration.KeepTailIntent
		if err := httpjson.Decode(w, r, &in, 16384, "expected one reference-only keep request"); err != nil {
			fail(w, 400, err)
			return
		}
		if err := in.Validate(); err != nil {
			fail(w, 400, err)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		v, err := store.Keep(ctx, bank, in)
		if err != nil {
			code := 503
			if errors.Is(err, exploration.ErrKeepConflict) || errors.Is(err, exploration.ErrKeepPrecondition) || errors.Is(err, exploration.ErrKeepCapacity) {
				code = 409
			}
			if errors.Is(err, exploration.ErrTailMembershipUnavailable) {
				code = 410
			}
			fail(w, code, err)
			return
		}
		httpjson.Write(w, 201, v)
	})
	mux.HandleFunc("POST /kept-observations/records", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			fail(w, 400, errors.New("keep has no query parameters"))
			return
		}
		var in exploration.KeepRecordsIntent
		if err := httpjson.Decode(w, r, &in, 16384, "expected one reference-only keep request"); err != nil {
			fail(w, 400, err)
			return
		}
		if err := in.Validate(); err != nil {
			fail(w, 400, err)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		v, err := store.KeepRecords(ctx, results, exploration.CaptureStore{Dir: filepath.Join(dir, "exploration")}, in)
		if err != nil {
			code := 503
			if errors.Is(err, exploration.ErrKeepConflict) || errors.Is(err, exploration.ErrKeepPrecondition) || errors.Is(err, exploration.ErrKeepCapacity) {
				code = 409
			}
			if errors.Is(err, exploration.ErrKeptRecordsUnavailable) {
				code = 410
			}
			fail(w, code, err)
			return
		}
		httpjson.Write(w, 201, v)
	})
	mux.HandleFunc("GET /kept-observations/search", func(w http.ResponseWriter, r *http.Request) {
		values, parseError := url.ParseQuery(r.URL.RawQuery)
		if parseError != nil {
			fail(w, 400, errors.New("invalid kept search query encoding"))
			return
		}
		for key, list := range values {
			if len(list) != 1 || (key != "origin" && key != "signal" && key != "cursor") {
				fail(w, 400, errors.New("unsupported kept search query"))
				return
			}
		}
		q := exploration.KeptSearchQuery{Origin: values.Get("origin"), Signal: values.Get("signal")}
		if err := q.Validate(); err != nil {
			fail(w, 400, err)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		page, err := store.Search(ctx, q, values.Get("cursor"))
		if err != nil {
			code := 503
			if errors.Is(err, exploration.ErrKeptCursor) {
				code = 409
			}
			httpjson.Write(w, code, map[string]any{"error": err.Error(), "repairAvailable": errors.Is(err, exploration.ErrKeptCatalog)})
			return
		}
		httpjson.Write(w, 200, page)
	})
	mux.HandleFunc("POST /kept-observations/repair", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			fail(w, 400, errors.New("repair has no query parameters"))
			return
		}
		var in exploration.KeptRepairRequest
		if err := httpjson.Decode(w, r, &in, 16384, "expected one bounded repair request"); err != nil {
			fail(w, 400, err)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		page, err := store.RepairCatalog(ctx, in.Cursor)
		if err != nil {
			code := 503
			if errors.Is(err, exploration.ErrKeptCursor) {
				code = 409
			}
			fail(w, code, err)
			return
		}
		httpjson.Write(w, 200, page)
	})
	mux.HandleFunc("GET /kept-observations/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			fail(w, 400, errors.New("kept observations have no query parameters"))
			return
		}
		v, err := store.Open(r.PathValue("id"))
		if err != nil {
			code := 400
			if errors.Is(err, os.ErrNotExist) {
				code = 404
			}
			fail(w, code, err)
			return
		}
		httpjson.Write(w, 200, v)
	})
}
