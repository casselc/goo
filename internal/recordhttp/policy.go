// Package recordhttp exposes the record exploration HTTP family. It owns bounded
// decoding, per-route source deadlines and effects; semantic execution stays in
// exploration. Snapshot storage is supplied by the application.
package recordhttp

import (
	"time"

	"github.com/casselc/goo/internal/queryhttp"
)

var policies = map[string]queryhttp.Policy{
	"/environment-scopes":         {Effect: queryhttp.Read, BodyLimit: 4096, Timeout: 15 * time.Second, TrailingMessage: "expected one JSON object"},
	"/environment-scopes/records": {Effect: queryhttp.Read, BodyLimit: 4096, Timeout: 15 * time.Second, TrailingMessage: "expected one JSON object"},
	"/compare":                    {Effect: queryhttp.Pure, BodyLimit: 1 << 20, Timeout: 0, TrailingMessage: "expected one JSON object"},
	"/inventory":                  {Effect: queryhttp.Read, BodyLimit: 16384, Timeout: 15 * time.Second, TrailingMessage: "expected one JSON object"},
	"/discovery":                  {Effect: queryhttp.Read, BodyLimit: 16384, Timeout: 30 * time.Second, TrailingMessage: "expected one JSON object"},
	"/record-aggregates":          {Effect: queryhttp.Read, BodyLimit: 16384, Timeout: 55 * time.Second, TrailingMessage: "expected one JSON object"},
	"/explore/counted-values":     {Effect: queryhttp.Read, BodyLimit: 16384, Timeout: 55 * time.Second, TrailingMessage: "expected one JSON object"},
	"/explore/page":               {Effect: queryhttp.ReadAndSnapshot, BodyLimit: 16384, Timeout: 55 * time.Second, TrailingMessage: "expected one JSON object"},
	"/explore":                    {Effect: queryhttp.ReadAndSnapshot, BodyLimit: 16384, Timeout: 55 * time.Second, TrailingMessage: "expected one JSON object"},
	"/trace-search":               {Effect: queryhttp.Read, BodyLimit: 65536, Timeout: 55 * time.Second, TrailingMessage: "expected one JSON object"},
}

// PolicyFor returns a value copy for an exact method/path. Admission and handlers
// use these same declarations; similarly named subpaths do not inherit policy.
func PolicyFor(method, path string) (queryhttp.Policy, bool) {
	if method != "POST" {
		return queryhttp.Policy{}, false
	}
	policy, ok := policies[path]
	return policy, ok
}
