package main

import (
	"context"
	"net/http"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/httpjson"
)

// Trace search owns its request and source read independently of record queries.
// It does not freeze a result token or publish evidence.
func registerTraceSearch(mux *http.ServeMux, source exploration.Source) {
	mux.HandleFunc("POST /trace-search", func(w http.ResponseWriter, r *http.Request) {
		fail := func(code int, e error) { httpjson.Write(w, code, map[string]string{"error": e.Error()}) }
		var request exploration.TraceSearchRequest
		if err := httpjson.Decode(w, r, &request, 65536, "expected one JSON object"); err != nil {
			fail(400, err)
			return
		}
		if _, err := exploration.CompileTraceSearch(request); err != nil {
			fail(400, err)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 55*time.Second)
		defer cancel()
		result, err := source.SearchTraces(ctx, request)
		if err != nil {
			fail(502, err)
			return
		}
		httpjson.Write(w, 200, result)
	})
}
