// Package recordhttp exposes the record exploration HTTP family. It owns bounded
// decoding, per-route source deadlines and effects; semantic execution stays in
// exploration. Snapshot storage is supplied by the application.
package recordhttp

import "time"

// Effect distinguishes planning, independent reads and ephemeral result holding.
// None of these routes publishes a durable capture or runs a live stream.
type Effect string

const (
	Pure            Effect = "pure"
	Read            Effect = "source-read"
	ReadAndSnapshot Effect = "source-read-and-snapshot"
)

type Policy struct {
	Effect        Effect
	BodyLimit     int64
	SourceTimeout time.Duration
}

var policies = map[string]Policy{
	"/compare":                {Pure, 1 << 20, 0},
	"/inventory":              {Read, 16384, 15 * time.Second},
	"/discovery":              {Read, 16384, 30 * time.Second},
	"/record-aggregates":      {Read, 16384, 55 * time.Second},
	"/explore/counted-values": {Read, 16384, 55 * time.Second},
	"/explore":                {ReadAndSnapshot, 16384, 55 * time.Second},
	"/trace-search":           {Read, 65536, 55 * time.Second},
}

// PolicyFor returns a value copy for an exact method/path. Admission and handlers
// use these same declarations; similarly named subpaths do not inherit policy.
func PolicyFor(method, path string) (Policy, bool) {
	if method != "POST" {
		return Policy{}, false
	}
	policy, ok := policies[path]
	return policy, ok
}
