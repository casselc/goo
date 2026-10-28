package main

import (
	"net/http"

	"github.com/casselc/goo/internal/nativehttp"
	"github.com/casselc/goo/internal/queryhttp"
	"github.com/casselc/goo/internal/recordhttp"
)

// Record and native families own their policies. Application composition still
// classifies the remaining explanation/publication routes.
func newQueryBudget(next http.Handler, limit int) http.Handler {
	return queryhttp.NewBudget(next, limit, func(r *http.Request) bool {
		if policy, ok := recordhttp.PolicyFor(r.Method, r.URL.Path); ok {
			return policy.Effect != queryhttp.Pure
		}
		if policy, ok := nativehttp.PolicyFor(r.Method, r.URL.Path); ok {
			return policy.Effect != queryhttp.Pure
		}
		if r.Method != "POST" {
			return false
		}
		switch r.URL.Path {
		case "/population/explain", "/captures", "/captures/live":
			return true
		default:
			return false
		}
	})
}
