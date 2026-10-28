package capturehttp

import (
	"time"

	"github.com/casselc/goo/internal/queryhttp"
)

var policies = map[string]queryhttp.Policy{
	"/captures":      {Effect: queryhttp.ReadOrPromote, BodyLimit: 16384, Timeout: 55 * time.Second, TrailingMessage: "expected one JSON object"},
	"/captures/live": {Effect: queryhttp.ReadOrFreeze, BodyLimit: 16384, Timeout: 55 * time.Second, TrailingMessage: "provide exactly one baseline or freeze capability"},
}

// PolicyFor describes the existing POST admission ceiling. Held promotion/freezing
// remains admitted with source reads; only source branches start this timeout.
// Bodyless saved-capture GET routes retain their own query/file validation.
func PolicyFor(method, path string) (queryhttp.Policy, bool) {
	if method != "POST" {
		return queryhttp.Policy{}, false
	}
	policy, ok := policies[path]
	return policy, ok
}
