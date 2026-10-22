package main

import (
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/httpjson"
	"net/http"
)

func registerRecipes(mux *http.ServeMux) {
	mux.HandleFunc("GET /recipes", func(w http.ResponseWriter, r *http.Request) { httpjson.Write(w, 200, exploration.Recipes()) })
	mux.HandleFunc("POST /recipes", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			ID       string            `json:"id"`
			Question exploration.Query `json:"question"`
		}
		e := httpjson.Decode(w, r, &input, 16384, "require one JSON object")
		if e != nil {
			httpjson.Write(w, 400, map[string]string{"error": e.Error()})
			return
		}
		p, e := exploration.PlanRecipe(input.ID, input.Question)
		if e != nil {
			httpjson.Write(w, 400, map[string]string{"error": e.Error()})
			return
		}
		httpjson.Write(w, 200, p)
	})
}
