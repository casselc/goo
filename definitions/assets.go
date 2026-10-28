// Package definitions contains the optional bundled convention proposal catalog.
package definitions

import _ "embed"

//go:embed prebuilts.json
var Prebuilts []byte
