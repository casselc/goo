package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/httpjson"
)

func registerKeptObservations(mux *http.ServeMux, dir string, bank *exploration.TailMembershipSnapshots) {
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
			if errors.Is(err, exploration.ErrKeepConflict) {
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
	mux.HandleFunc("GET /kept-observations/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			fail(w, 400, errors.New("kept observations have no query parameters"))
			return
		}
		v, err := store.Load(r.PathValue("id"))
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
