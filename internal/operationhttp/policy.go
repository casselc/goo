// Package operationhttp declares the legacy operation-read HTTP boundaries.
package operationhttp

import (
	"github.com/casselc/goo/internal/queryhttp"
	"time"
)

// PolicyFor classifies exact GET and implicit HEAD routes. Declaration reads do
// not use source capacity. JSON source bodies remain absent on these GET routes.
func PolicyFor(method, path string) (queryhttp.Policy, bool) {
	if method != "GET" && method != "HEAD" {
		return queryhttp.Policy{}, false
	}
	switch path {
	case "/operation-packages":
		return queryhttp.Policy{Effect: queryhttp.Pure}, true
	case "/http-package", "/operations", "/counter-rates", "/histograms":
		return queryhttp.Policy{Effect: queryhttp.Read, Timeout: 55 * time.Second}, true
	default:
		return queryhttp.Policy{}, false
	}
}
