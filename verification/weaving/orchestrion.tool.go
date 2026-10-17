//go:build tools

package tools

import (
	_ "github.com/DataDog/orchestrion"
	_ "github.com/casselc/goo/instrumentation"
	_ "github.com/casselc/observability/quintgo/qobs"
)
