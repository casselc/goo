package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/casselc/goo/annotations"
	"github.com/casselc/goo/prebuilt"
)

func registerOperationReads(mux *http.ServeMux, plans map[string]*annotations.OperationPlan, ordered []*annotations.OperationPlan, expanded *prebuilt.Expansion, source prebuilt.HTTPSource, connection clickHouseConnection) {
	mux.HandleFunc("GET /http-package", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		minutes := 5
		if value := r.URL.Query().Get("minutes"); value != "" {
			parsed, e := strconv.Atoi(value)
			if e != nil || parsed < 1 || parsed > 120 {
				http.Error(w, `{"error":"minutes must be 1..120"}`, 400)
				return
			}
			minutes = parsed
		}
		for key, values := range r.URL.Query() {
			if key != "minutes" || len(values) != 1 {
				http.Error(w, `{"error":"unsupported query parameter"}`, 400)
				return
			}
		}
		to := time.Now().UnixMilli()
		result, e := prebuilt.FetchHTTP(r.Context(), connection.Client, connection.Endpoint, connection.User, connection.Password, expanded, source, to-int64(minutes)*60000, to)
		if e != nil {
			w.WriteHeader(502)
			json.NewEncoder(w).Encode(map[string]string{"error": e.Error()})
			return
		}
		json.NewEncoder(w).Encode(result)
	})
	mux.HandleFunc("GET /operation-packages", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		if len(r.URL.Query()) != 0 {
			http.Error(w, `{"error":"no query parameters supported"}`, 400)
			return
		}
		json.NewEncoder(w).Encode(ordered)
	})
	mux.HandleFunc("GET /operations", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		minutes := 5
		for key, values := range r.URL.Query() {
			if (key != "minutes" && key != "package") || len(values) != 1 {
				http.Error(w, `{"error":"unsupported query parameter"}`, 400)
				return
			}
		}
		if v := r.URL.Query().Get("minutes"); v != "" {
			n, e := strconv.Atoi(v)
			if e != nil || n < 1 || n > 120 {
				http.Error(w, `{"error":"minutes must be 1..120"}`, 400)
				return
			}
			minutes = n
		}
		name := r.URL.Query().Get("package")
		if name == "" {
			name = ordered[0].Definition.Name
		}
		plan := plans[name]
		if plan == nil {
			http.Error(w, `{"error":"unknown operation package"}`, 400)
			return
		}
		to := time.Now().UnixMilli()
		result, e := plan.Fetch(r.Context(), connection.Client, connection.Endpoint, connection.User, connection.Password, to-int64(minutes)*60000, to)
		if e != nil {
			w.WriteHeader(502)
			json.NewEncoder(w).Encode(map[string]string{"error": e.Error()})
			return
		}
		json.NewEncoder(w).Encode(result)
	})
	mux.HandleFunc("GET /counter-rates", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		for key, values := range r.URL.Query() {
			if (key != "package" && key != "fromMs" && key != "toMs") || len(values) != 1 {
				http.Error(w, `{"error":"unsupported rate query parameter"}`, 400)
				return
			}
		}
		plan := plans[r.URL.Query().Get("package")]
		from, e1 := strconv.ParseInt(r.URL.Query().Get("fromMs"), 10, 64)
		to, e2 := strconv.ParseInt(r.URL.Query().Get("toMs"), 10, 64)
		if plan == nil || e1 != nil || e2 != nil || from < 60000 || to <= from || to-from > 120*60000 {
			http.Error(w, `{"error":"require package and bounded fromMs/toMs"}`, 400)
			return
		}
		result, e := plan.FetchRates(r.Context(), connection.Client, connection.Endpoint, connection.User, connection.Password, from, to)
		if e != nil {
			w.WriteHeader(502)
			json.NewEncoder(w).Encode(map[string]string{"error": e.Error()})
			return
		}
		json.NewEncoder(w).Encode(result)
	})
	mux.HandleFunc("GET /histograms", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		for key, values := range r.URL.Query() {
			if (key != "package" && key != "fromMs" && key != "toMs") || len(values) != 1 {
				http.Error(w, `{"error":"unsupported histogram query parameter"}`, 400)
				return
			}
		}
		plan := plans[r.URL.Query().Get("package")]
		from, e1 := strconv.ParseInt(r.URL.Query().Get("fromMs"), 10, 64)
		to, e2 := strconv.ParseInt(r.URL.Query().Get("toMs"), 10, 64)
		if plan == nil || e1 != nil || e2 != nil || from < 60000 || to <= from || to-from > 120*60000 {
			http.Error(w, `{"error":"require package and bounded fromMs/toMs"}`, 400)
			return
		}
		result, e := plan.FetchHistograms(r.Context(), connection.Client, connection.Endpoint, connection.User, connection.Password, from, to)
		if e != nil {
			w.WriteHeader(502)
			json.NewEncoder(w).Encode(map[string]string{"error": e.Error()})
			return
		}
		json.NewEncoder(w).Encode(result)
	})
}
