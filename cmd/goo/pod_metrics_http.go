package main

import (
	"context"
	"github.com/casselc/goo/internal/httpjson"
	"github.com/casselc/goo/measurements"
	"net/http"
	"time"
)

// The source response is bounded at 8 MiB; permit its JSON request wrapper too.
const podDashboardBodyLimit = (8 << 20) + 1024

func registerPodMeasurements(mux *http.ServeMux, source measurements.Source) {
	mux.HandleFunc("POST /pod-metric-discovery", func(w http.ResponseWriter, r *http.Request) {
		fail := func(code int, e error) { httpjson.Write(w, code, map[string]string{"error": e.Error()}) }
		var q measurements.PodMetricDiscoveryQuestion
		if e := httpjson.Decode(w, r, &q, 65536, "expected one JSON object"); e != nil {
			fail(400, e)
			return
		}
		if e := q.Validate(); e != nil {
			fail(400, e)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		result, e := source.DiscoverPodMetrics(ctx, q)
		if e != nil {
			fail(502, e)
			return
		}
		httpjson.Write(w, 200, result)
	})
	mux.HandleFunc("POST /pod-dashboard/plan", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		fail := func(e error) { httpjson.Write(w, 400, map[string]string{"error": e.Error()}) }
		var in measurements.PodDashboardRequest
		if e := httpjson.Decode(w, r, &in, podDashboardBodyLimit, "expected one JSON object"); e != nil {
			fail(e)
			return
		}
		if e := ctx.Err(); e != nil {
			fail(e)
			return
		}
		result, e := measurements.PlanPodDashboard(in)
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
