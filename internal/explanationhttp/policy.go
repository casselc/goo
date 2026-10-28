package explanationhttp

import (
	"time"

	"github.com/casselc/goo/internal/queryhttp"
)

var policies = map[string]queryhttp.Policy{
	"/population/explain/plan":   {Effect: queryhttp.Pure, BodyLimit: 65536, TrailingMessage: "expected one JSON object"},
	"/population/explain":        {Effect: queryhttp.ReadAndSnapshot, BodyLimit: 65536, Timeout: 55 * time.Second, TrailingMessage: "expected one JSON object"},
	"/population/explain/filter": {Effect: queryhttp.Pure, BodyLimit: 65536, TrailingMessage: "expected one JSON object"},
}

// PolicyFor distinguishes pure supplied-data planning from observed source reads.
func PolicyFor(method, path string) (queryhttp.Policy, bool) {
	if method != "POST" {
		return queryhttp.Policy{}, false
	}
	policy, ok := policies[path]
	return policy, ok
}
