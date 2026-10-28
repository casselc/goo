// Package nativehttp owns native measurement/discovery and resource-planner HTTP
// boundaries. Domain execution and optional lifecycle custody stay in measurements.
package nativehttp

import (
	"time"

	"github.com/casselc/goo/internal/queryhttp"
)

// PodDashboardBodyLimit permits the bounded 8 MiB discovery response and wrapper.
const PodDashboardBodyLimit = (8 << 20) + 1024

var policies = map[string]queryhttp.Policy{
	"/metric-discovery":     {Effect: queryhttp.Read, BodyLimit: 65536, Timeout: 15 * time.Second, TrailingMessage: "one JSON object required"},
	"/measurements":         {Effect: queryhttp.ReadWithOptionalCustody, BodyLimit: 65536, Timeout: 15 * time.Second, TrailingMessage: "one JSON object required"},
	"/host-discovery":       {Effect: queryhttp.Read, BodyLimit: 16384, Timeout: 15 * time.Second, TrailingMessage: "expected one JSON object"},
	"/pod-discovery":        {Effect: queryhttp.Read, BodyLimit: 16384, Timeout: 30 * time.Second, TrailingMessage: "expected one JSON object"},
	"/pod-metric-discovery": {Effect: queryhttp.Read, BodyLimit: 65536, Timeout: 30 * time.Second, TrailingMessage: "expected one JSON object"},
	"/host-dashboard/plan":  {Effect: queryhttp.Pure, BodyLimit: 65536, Timeout: 2 * time.Second, TrailingMessage: "expected one JSON object"},
	"/pod-dashboard/plan":   {Effect: queryhttp.Pure, BodyLimit: PodDashboardBodyLimit, Timeout: 2 * time.Second, TrailingMessage: "expected one JSON object"},
}

// PolicyFor returns a copy for an exact native method/path; catalog publication,
// container planners, captures and SSE retain separate owners.
func PolicyFor(method, path string) (queryhttp.Policy, bool) {
	if method != "POST" {
		return queryhttp.Policy{}, false
	}
	policy, ok := policies[path]
	return policy, ok
}
