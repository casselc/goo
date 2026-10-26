package main

import (
	"context"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/httpjson"
	"net/http"
	"time"
)

func registerPodDiscovery(mux *http.ServeMux, source exploration.Source) {
	mux.HandleFunc("POST /pod-discovery", func(w http.ResponseWriter, r *http.Request) {
		fail := func(code int, e error) { httpjson.Write(w, code, map[string]string{"error": e.Error()}) }
		var q exploration.PodDiscoveryQuestion
		if e := httpjson.Decode(w, r, &q, 16384, "expected one JSON object"); e != nil {
			fail(400, e)
			return
		}
		if e := q.Validate(); e != nil {
			fail(400, e)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		result, e := source.DiscoverPods(ctx, q)
		if e != nil {
			fail(502, e)
			return
		}
		httpjson.Write(w, 200, result)
	})
}
