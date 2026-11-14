// Package celhttp exposes bounded source-free record CEL planning.
package celhttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/casselc/goo/internal/httpjson"
	"github.com/casselc/goo/internal/queryhttp"
	"github.com/casselc/goo/querycel"
)

const BodyLimit int64 = 32768

// ErrorResponse contains no partially admitted question or target plan.
type ErrorResponse struct {
	Error      string               `json:"error"`
	Diagnostic *querycel.Diagnostic `json:"diagnostic,omitempty"`
}

// PolicyFor declares that checking a supplied draft has no source/store effects.
func PolicyFor(method, path string) (queryhttp.Policy, bool) {
	if method != "POST" || path != "/cel/plan" {
		return queryhttp.Policy{}, false
	}
	return queryhttp.Policy{Effect: queryhttp.Pure, BodyLimit: BodyLimit, Timeout: 2 * time.Second, TrailingMessage: "expected one JSON object"}, true
}

// Register accepts no source, store, clock or readiness dependencies.
func Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /cel/plan", func(w http.ResponseWriter, r *http.Request) {
		fail := func(err error) {
			response := ErrorResponse{Error: err.Error()}
			errors.As(err, &response.Diagnostic)
			httpjson.Write(w, http.StatusBadRequest, response)
		}
		if r.URL.RawQuery != "" {
			fail(fmt.Errorf("query parameters are unsupported"))
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		r = r.WithContext(ctx)
		var request querycel.PlanRequest
		if err := httpjson.Decode(w, r, &request, BodyLimit, "expected one JSON object"); err != nil {
			fail(err)
			return
		}
		plan, err := querycel.PlanRecord(request)
		if err != nil {
			fail(err)
			return
		}
		// Planning is synchronous and bounded by compiler budgets; this check
		// prevents a completed/cancelled caller from publishing its result.
		if err := ctx.Err(); err != nil {
			fail(err)
			return
		}
		httpjson.Write(w, http.StatusOK, plan)
	})
}
