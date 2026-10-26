package main

import (
	"context"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/httpjson"
	"net/http"
	"time"
)

var resultSnapshots = &exploration.ResultSnapshots{}

func registerExploration(mux *http.ServeMux, source exploration.Source, banks ...*exploration.TailMembershipSnapshots) {
	registerTraceSearch(mux, source)
	registerRecipes(mux)
	registerPoolSignals(mux)
	registerPoolView(mux)
	registerPopulationExplanations(mux, source)
	registerTail(mux, source, banks...)
	mux.HandleFunc("POST /compare", func(w http.ResponseWriter, r *http.Request) {
		fail := func(e error) { httpjson.Write(w, 400, map[string]string{"error": e.Error()}) }
		var request exploration.ComparisonRequest
		if e := httpjson.Decode(w, r, &request, 1<<20, "expected one JSON object"); e != nil {
			fail(e)
			return
		}
		result, e := exploration.CompareCohorts(request)
		if e != nil {
			fail(e)
			return
		}
		httpjson.Write(w, 200, result)
	})

	mux.HandleFunc("POST /inventory", func(w http.ResponseWriter, r *http.Request) {
		fail := func(code int, e error) {
			httpjson.Write(w, code, map[string]string{"error": e.Error()})
		}
		var q exploration.InventoryQuestion
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
		result, e := source.Inventory(ctx, q)
		if e != nil {
			fail(502, e)
			return
		}
		httpjson.Write(w, 200, result)
	})
	mux.HandleFunc("POST /discovery", func(w http.ResponseWriter, r *http.Request) {
		fail := func(code int, e error) {
			httpjson.Write(w, code, map[string]string{"error": e.Error()})
		}
		var q exploration.InventoryQuestion
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
		result, e := source.DiscoverServices(ctx, q)
		if e != nil {
			fail(502, e)
			return
		}
		httpjson.Write(w, 200, result)
	})

	mux.HandleFunc("POST /record-aggregates", func(w http.ResponseWriter, r *http.Request) {
		fail := func(code int, e error) { httpjson.Write(w, code, map[string]string{"error": e.Error()}) }
		var q exploration.Query
		if e := httpjson.Decode(w, r, &q, 16384, "expected one JSON object"); e != nil {
			fail(400, e)
			return
		}
		if e := exploration.ValidateRecordAggregatesQuery(q); e != nil {
			fail(400, e)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 55*time.Second)
		defer cancel()
		result, e := source.FetchRecordAggregates(ctx, q)
		if e != nil {
			fail(502, e)
			return
		}
		httpjson.Write(w, 200, result)
	})

	mux.HandleFunc("POST /explore/counted-values", func(w http.ResponseWriter, r *http.Request) {
		fail := func(code int, e error) { httpjson.Write(w, code, map[string]string{"error": e.Error()}) }
		var q exploration.Query
		if e := httpjson.Decode(w, r, &q, 16384, "expected one JSON object"); e != nil {
			fail(400, e)
			return
		}
		if e := exploration.ValidateCountedValuesQuery(q); e != nil {
			fail(400, e)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 55*time.Second)
		defer cancel()
		result, e := source.FetchCountedValues(ctx, q)
		if e != nil {
			fail(502, e)
			return
		}
		httpjson.Write(w, 200, result)
	})

	mux.HandleFunc("POST /explore", func(w http.ResponseWriter, r *http.Request) {
		fail := func(code int, e error) {
			httpjson.Write(w, code, map[string]string{"error": e.Error()})
		}
		var q exploration.Query
		if e := httpjson.Decode(w, r, &q, 16384, "expected one JSON object"); e != nil {
			fail(400, e)
			return
		}
		if e := q.Validate(); e != nil {
			fail(400, e)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 55*time.Second)
		defer cancel()
		result, e := source.Fetch(ctx, q)
		if e != nil {
			fail(502, e)
			return
		}
		if token, err := resultSnapshots.Freeze(result); err == nil {
			result.ResultToken = token
		} else {
			result.Warnings = append(result.Warnings, "Live selection retention unavailable: "+err.Error())
		}
		httpjson.Write(w, 200, result)
	})
}
