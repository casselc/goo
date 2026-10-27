package main

import (
	"context"
	"github.com/casselc/goo/internal/httpjson"
	"github.com/casselc/goo/measurements"
	"net/http"
	"time"
)

func registerContainerMemoryView(mux *http.ServeMux) {
	mux.HandleFunc("POST /container-memory/view", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		fail := func(err error) { httpjson.Write(w, 400, map[string]string{"error": err.Error()}) }
		var request measurements.ContainerMemoryViewRequest
		if err := httpjson.Decode(w, r, &request, 8<<20, "expected one JSON object"); err != nil {
			fail(err)
			return
		}
		if err := ctx.Err(); err != nil {
			fail(err)
			return
		}
		result, err := measurements.ComposeContainerMemoryView(request)
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
