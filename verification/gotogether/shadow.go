// Package shadow judges the browser's request log without changing its clocks.
package shadow

import (
	"fmt"
	"github.com/chucklehead-dev/gotogether/qdriver"
	"github.com/chucklehead-dev/gotogether/qtest"
	"github.com/chucklehead-dev/gotogether/qtrace"
)

// Scenario names the visible journey that must occur, independently of transport.
func Scenario() qtest.Expectation {
	return qtest.NamedScenario("439 acquisition, local inspection and retained return",
		qtest.Takes("read").With("read", "1"),
		qtest.Takes("show").With("phase", `"explicit-read"`),
		qtest.Takes("show").With("phase", `"local-inspect"`),
		qtest.Takes("show").With("phase", `"retained-return"`),
		qtest.Takes("show").With("phase", `"cancel-pending-reread"`),
		qtest.Takes("show").With("phase", `"late-success-ignored"`),
		qtest.Takes("show").With("phase", `"final"`))
}

// Steps preserves driver sequence, invocation/completion identities and times.
// Only an explicit driver drain marker becomes a settle. EOF is never a settle.
func Steps(log *qdriver.Log) ([]qtrace.Step, error) {
	steps := log.Steps()
	for i := range steps {
		s := &steps[i]
		if s.Action == "drained" {
			if (i != len(steps)-1 && !(i == len(steps)-2 && steps[i+1].Lost != "")) || s.Invocation {
				return nil, fmt.Errorf("drained must be the final driver step")
			}
			s.Settled = true
		}
	}
	return steps, nil
}
