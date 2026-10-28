package main

import (
	"net/http"

	"github.com/casselc/goo/internal/queryhttp"
	"github.com/casselc/goo/internal/recordhttp"
)

// Application composition classifies the remaining native/publication families;
// the extracted record family owns its own body, deadline and effect policies.
func newQueryBudget(next http.Handler, limit int) http.Handler {
	return queryhttp.NewBudget(next, limit, func(r *http.Request) bool {
		if policy, ok := recordhttp.PolicyFor(r.Method, r.URL.Path); ok {
			return policy.Effect != recordhttp.Pure
		}
		if r.Method != "POST" {
			return false
		}
		switch r.URL.Path {
		case "/metric-discovery", "/host-discovery", "/pod-discovery", "/pod-metric-discovery", "/measurements", "/population/explain", "/captures", "/captures/live":
			return true
		default:
			return false
		}
	})
}
