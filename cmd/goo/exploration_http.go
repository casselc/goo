package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"io"
	"net/http"
	"os"
	"time"
)

var resultSnapshots exploration.ResultSnapshots

func registerExploration(mux *http.ServeMux, client *http.Client) {
	registerRecipes(mux)
	registerTail(mux, client)
	mux.HandleFunc("POST /compare", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		fail := func(e error) { w.WriteHeader(400); json.NewEncoder(w).Encode(map[string]string{"error": e.Error()}) }
		var request exploration.ComparisonRequest
		d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		d.DisallowUnknownFields()
		if e := d.Decode(&request); e != nil {
			fail(e)
			return
		}
		if d.Decode(new(any)) != io.EOF {
			fail(fmt.Errorf("expected one JSON object"))
			return
		}
		result, e := exploration.CompareCohorts(request)
		if e != nil {
			fail(e)
			return
		}
		json.NewEncoder(w).Encode(result)
	})

	mux.HandleFunc("POST /inventory", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		fail := func(code int, e error) {
			w.WriteHeader(code)
			json.NewEncoder(w).Encode(map[string]string{"error": e.Error()})
		}
		var q exploration.InventoryQuestion
		d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16384))
		d.DisallowUnknownFields()
		if e := d.Decode(&q); e != nil {
			fail(400, e)
			return
		}
		if d.Decode(new(any)) != io.EOF {
			fail(400, fmt.Errorf("expected one JSON object"))
			return
		}
		if e := q.Validate(); e != nil {
			fail(400, e)
			return
		}
		endpoint := os.Getenv("CLICKHOUSE_URL")
		if endpoint == "" {
			endpoint = "http://localhost:8123"
		}
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		source := exploration.Source{Endpoint: endpoint, User: os.Getenv("CLICKHOUSE_USER"), Password: os.Getenv("CLICKHOUSE_PASSWORD"), Client: client}
		result, e := source.Inventory(ctx, q)
		if e != nil {
			fail(502, e)
			return
		}
		json.NewEncoder(w).Encode(result)
	})
	mux.HandleFunc("POST /discovery", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		fail := func(code int, e error) {
			w.WriteHeader(code)
			json.NewEncoder(w).Encode(map[string]string{"error": e.Error()})
		}
		var q exploration.InventoryQuestion
		d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16384))
		d.DisallowUnknownFields()
		if e := d.Decode(&q); e != nil {
			fail(400, e)
			return
		}
		if d.Decode(new(any)) != io.EOF {
			fail(400, fmt.Errorf("expected one JSON object"))
			return
		}
		if e := q.Validate(); e != nil {
			fail(400, e)
			return
		}
		endpoint := os.Getenv("CLICKHOUSE_URL")
		if endpoint == "" {
			endpoint = "http://localhost:8123"
		}
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		source := exploration.Source{Endpoint: endpoint, User: os.Getenv("CLICKHOUSE_USER"), Password: os.Getenv("CLICKHOUSE_PASSWORD"), Client: client}
		result, e := source.DiscoverServices(ctx, q)
		if e != nil {
			fail(502, e)
			return
		}
		json.NewEncoder(w).Encode(result)
	})

	mux.HandleFunc("POST /explore", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		fail := func(code int, e error) {
			w.WriteHeader(code)
			json.NewEncoder(w).Encode(map[string]string{"error": e.Error()})
		}
		var q exploration.Query
		d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16384))
		d.DisallowUnknownFields()
		if e := d.Decode(&q); e != nil {
			fail(400, e)
			return
		}
		if d.Decode(new(any)) != io.EOF {
			fail(400, fmt.Errorf("expected one JSON object"))
			return
		}
		if e := q.Validate(); e != nil {
			fail(400, e)
			return
		}
		endpoint := os.Getenv("CLICKHOUSE_URL")
		if endpoint == "" {
			endpoint = "http://localhost:8123"
		}
		ctx, cancel := context.WithTimeout(r.Context(), 55*time.Second)
		defer cancel()
		source := exploration.Source{Endpoint: endpoint, User: os.Getenv("CLICKHOUSE_USER"), Password: os.Getenv("CLICKHOUSE_PASSWORD"), Client: client}
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
		json.NewEncoder(w).Encode(result)
	})
}
