// Package api contains the HTTP contract and its read-only browser.
package api

import "embed"

//go:embed index.html openapi.json
var Assets embed.FS
