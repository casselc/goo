package main

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"io"
	"net/http"
)

func registerRecipes(mux *http.ServeMux) {
	send := func(w http.ResponseWriter, code int, v any) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(v)
	}
	mux.HandleFunc("GET /recipes", func(w http.ResponseWriter, r *http.Request) { send(w, 200, exploration.Recipes()) })
	mux.HandleFunc("POST /recipes", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			ID       string            `json:"id"`
			Question exploration.Query `json:"question"`
		}
		d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16384))
		d.DisallowUnknownFields()
		e := d.Decode(&input)
		if e == nil && d.Decode(new(any)) != io.EOF {
			e = fmt.Errorf("require one JSON object")
		}
		if e != nil {
			send(w, 400, map[string]string{"error": e.Error()})
			return
		}
		p, e := exploration.PlanRecipe(input.ID, input.Question)
		if e != nil {
			send(w, 400, map[string]string{"error": e.Error()})
			return
		}
		send(w, 200, p)
	})
}
