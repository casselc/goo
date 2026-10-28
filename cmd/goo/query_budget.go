package main

import (
	"net/http"

	"github.com/casselc/goo/internal/capturehttp"
	"github.com/casselc/goo/internal/explanationhttp"
	"github.com/casselc/goo/internal/nativehttp"
	"github.com/casselc/goo/internal/operationhttp"
	"github.com/casselc/goo/internal/queryhttp"
	"github.com/casselc/goo/internal/recordhttp"
)

// Composition names policy owners; each family declares its own effects, limits
// and deadlines. The classifier never consumes a body or performs source work.
func newQueryBudget(next http.Handler, limit int) http.Handler {
	policies := [...]func(string, string) (queryhttp.Policy, bool){recordhttp.PolicyFor, nativehttp.PolicyFor, explanationhttp.PolicyFor, capturehttp.PolicyFor, operationhttp.PolicyFor}
	return queryhttp.NewBudget(next, limit, func(r *http.Request) bool {
		for _, lookup := range policies {
			if policy, ok := lookup(r.Method, r.URL.Path); ok {
				return policy.Effect != queryhttp.Pure
			}
		}
		return false
	})
}
