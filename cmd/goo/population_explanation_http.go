package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/httpjson"
	"net/http"
	"time"
)

func registerPopulationExplanations(mux *http.ServeMux, source exploration.Source) {
	fail := func(w http.ResponseWriter, code int, e error) {
		httpjson.Write(w, code, map[string]string{"error": e.Error()})
	}
	mux.HandleFunc("POST /population/explain/plan", func(w http.ResponseWriter, r *http.Request) {
		var request exploration.ExplanationRequest
		if e := httpjson.Decode(w, r, &request, 65536, "expected one JSON object"); e != nil {
			fail(w, 400, e)
			return
		}
		plan, e := exploration.PlanExplanation(request)
		if e != nil {
			fail(w, 400, e)
			return
		}
		httpjson.Write(w, 200, plan)
	})
	mux.HandleFunc("POST /population/explain", func(w http.ResponseWriter, r *http.Request) {
		var request exploration.ExplanationRequest
		if e := httpjson.Decode(w, r, &request, 65536, "expected one JSON object"); e != nil {
			fail(w, 400, e)
			return
		}
		if e := request.Validate(); e != nil {
			fail(w, 400, e)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 55*time.Second)
		defer cancel()
		result, e := source.Explain(ctx, request)
		if e != nil {
			fail(w, 502, e)
			return
		}
		if token, e := resultSnapshots.FreezeExplanation(result); e == nil {
			result.ResultToken = token
		} else {
			w.Header().Set("X-Goo-Save-Status", "unavailable")
		}
		httpjson.Write(w, 200, result)
	})
	mux.HandleFunc("POST /population/explain/filter", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Request exploration.ExplanationRequest `json:"request"`
			Field   exploration.Field              `json:"field"`
			Value   json.RawMessage                `json:"value"`
			Side    string                         `json:"side"`
			Mode    string                         `json:"mode"`
		}
		if e := httpjson.Decode(w, r, &in, 65536, "expected one JSON object"); e != nil {
			fail(w, 400, e)
			return
		}
		if len(in.Value) == 0 {
			fail(w, 400, fmt.Errorf("value is required; null selects missing keys"))
			return
		}
		var value *string
		if e := json.Unmarshal(in.Value, &value); e != nil {
			fail(w, 400, fmt.Errorf("value must be a string or null"))
			return
		}
		question, e := exploration.PlanExplanationFilter(in.Request, in.Field, value, in.Side, in.Mode)
		if e != nil {
			fail(w, 400, e)
			return
		}
		httpjson.Write(w, 200, struct {
			Question exploration.Query `json:"question"`
		}{question})
	})
}
