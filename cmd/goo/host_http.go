package main

import (
	"context"
	"github.com/casselc/goo/internal/httpjson"
	"github.com/casselc/goo/measurements"
	"net/http"
	"time"
)

func registerHostMeasurements(mux *http.ServeMux, source measurements.Source) {
	mux.HandleFunc("POST /host-discovery", func(w http.ResponseWriter, r *http.Request) {
		fail := func(code int, e error) { httpjson.Write(w, code, map[string]string{"error": e.Error()}) }
		var q measurements.HostDiscoveryQuestion
		if e := httpjson.Decode(w, r, &q, 16384, "expected one JSON object"); e != nil {
			fail(400, e)
			return
		}
		if e := q.Validate(); e != nil {
			fail(400, e)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		result, e := source.DiscoverHosts(ctx, q)
		if e != nil {
			fail(502, e)
			return
		}
		httpjson.Write(w, 200, result)
	})
	mux.HandleFunc("POST /host-dashboard/plan", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		fail := func(e error) { httpjson.Write(w, 400, map[string]string{"error": e.Error()}) }
		var in measurements.HostDashboardRequest
		if e := httpjson.Decode(w, r, &in, 65536, "expected one JSON object"); e != nil {
			fail(e)
			return
		}
		if e := ctx.Err(); e != nil {
			fail(e)
			return
		}
		result, e := measurements.PlanHostDashboard(in)
		if e != nil {
			fail(e)
			return
		}
		if e := ctx.Err(); e != nil {
			fail(e)
			return
		}
		httpjson.Write(w, 200, result)
	})
}
