// Package ui contains the browser application's build-time display assets.
package ui

import "embed"

// Assets is embedded so the installed Goo binary serves its matching UI without
// a source checkout, writable asset directory, or a separate application server.
//
//go:embed *.html *.css *.mjs
var Assets embed.FS
