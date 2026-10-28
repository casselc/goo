package main

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/annotations"
	"github.com/casselc/goo/internal/httpjson"
	"github.com/casselc/goo/investigations"
	"net/http"
)

func registerInvestigations(mux *http.ServeMux, plans map[string]*annotations.OperationPlan, connection clickHouseConnection, dir string) {
	store := investigations.Store{Dir: dir}
	reply := func(w http.ResponseWriter, code int, v any) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(v)
	}
	fail := func(w http.ResponseWriter, code int, e error) { reply(w, code, map[string]string{"error": e.Error()}) }
	mux.HandleFunc("GET /investigations/{id}", func(w http.ResponseWriter, r *http.Request) {
		c, e := store.Load(r.PathValue("id"))
		if e != nil {
			fail(w, 404, e)
			return
		}
		reply(w, 200, c)
	})
	mux.HandleFunc("POST /investigations", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Question  investigations.Question `json:"question"`
			Selection json.RawMessage         `json:"selection"`
			Parent    string                  `json:"parent"`
		}
		if e := httpjson.Decode(w, r, &in, 16384, "expected one JSON object"); e != nil {
			fail(w, 400, e)
			return
		}
		if in.Parent != "" {
			p, e := store.Load(in.Parent)
			if e != nil {
				fail(w, 404, e)
				return
			}
			if in.Question != (investigations.Question{}) && in.Question != p.Question {
				fail(w, 409, fmt.Errorf("refresh cannot change original question"))
				return
			}
			in.Question = p.Question
			if len(in.Selection) == 0 {
				in.Selection = p.Selection
			}
		}
		if e := in.Question.Validate(); e != nil {
			fail(w, 400, e)
			return
		}
		if e := investigations.ValidateSelection(in.Selection); e != nil {
			fail(w, 400, e)
			return
		}
		plan := plans[in.Question.Package]
		if plan == nil || plan.Revision != in.Question.Revision {
			fail(w, 409, fmt.Errorf("package revision unavailable; original capture remains reopenable"))
			return
		}
		q := in.Question
		operations, e := plan.Fetch(r.Context(), connection.Client, connection.Endpoint, connection.User, connection.Password, q.FromMS, q.ToMS)
		if e != nil {
			fail(w, 502, e)
			return
		}
		evidence := investigations.Evidence{Errors: map[string]string{}}
		evidence.Operations, _ = json.Marshal(operations)
		for _, view := range plan.Definition.Views {
			if (view == "counter-rates" || view == "error-ratios" || view == "fleet-coverage") && evidence.Rates == nil && evidence.Errors["rates"] == "" {
				v, e := plan.FetchRates(r.Context(), connection.Client, connection.Endpoint, connection.User, connection.Password, q.FromMS, q.ToMS)
				if e != nil {
					evidence.Errors["rates"] = e.Error()
				} else {
					evidence.Rates, _ = json.Marshal(v)
				}
			}
			if view == "histogram-heatmap" {
				v, e := plan.FetchHistograms(r.Context(), connection.Client, connection.Endpoint, connection.User, connection.Password, q.FromMS, q.ToMS)
				if e != nil {
					evidence.Errors["histograms"] = e.Error()
				} else {
					evidence.Histograms, _ = json.Marshal(v)
				}
			}
		}
		c, e := store.Save(q, in.Selection, evidence, in.Parent)
		if e != nil {
			fail(w, 400, e)
			return
		}
		reply(w, 201, c)
	})
}
