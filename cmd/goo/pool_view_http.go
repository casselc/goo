package main

import (
	"context"
	"net/http"
	"time"

	"github.com/casselc/goo/internal/httpjson"
	"github.com/casselc/goo/measurements"
)

func registerPoolView(mux *http.ServeMux) {
	mux.HandleFunc("POST /pool-signals/view", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		fail := func(err error) { httpjson.Write(w, 400, map[string]string{"error": err.Error()}) }
		var request measurements.PoolViewRequest
		if err := httpjson.Decode(w, r, &request, 2<<20, "expected one JSON object"); err != nil {
			fail(err)
			return
		}
		if err := ctx.Err(); err != nil {
			fail(err)
			return
		}
		result, err := measurements.ComposePoolView(request)
		if err != nil {
			fail(err)
			return
		}
		if err := ctx.Err(); err != nil {
			fail(err)
			return
		}
		httpjson.Write(w, 200, result)
	})
}
