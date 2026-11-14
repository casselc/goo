package shadow

import (
	"github.com/chucklehead-dev/gotogether/qdriver"
	"github.com/chucklehead-dev/gotogether/qdriver/bombadil"
	"github.com/chucklehead-dev/gotogether/qtrace"
	"os"
	"testing"
)

// This actor's six generic phases do not claim the deterministic 439 scenario.
// Check adapter admission and its actual request/completion separation only.
func TestBombadilRecorded(t *testing.T) {
	path := os.Getenv("GOO_GOTOGETHER_BOMBADIL")
	if path == "" {
		t.Skip("requires recorded bounded live actor")
	}
	log, err := bombadil.LogFile(path, bombadil.Options{
		Header: qdriver.Header{Recorder: "raw-gauge-actor", Actor: "acquisition"}, Snapshot: "state",
		Clicks: map[string][]string{
			"goo.stageRawGaugeScope": {}, "goo.discoverRawGaugeInstrument": {},
			"goo.readRawGaugeExplicitly": {"read"}, "goo.findHeldGaugeObservation": {},
			"goo.inspectHeldGaugeOccurrence": {}, "goo.observeLocalGaugeDetail": {},
		},
		Show: &bombadil.Show{Key: "bank", Fields: []string{"bank", "clock", "question", "phase", "find", "detail"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	traces, issues := qtrace.Reconstruct(log.Steps())
	for _, issue := range issues {
		if issue.Severity == "error" {
			t.Fatal(issue)
		}
	}
	calls, shows := 0, 0
	for _, r := range log.Records {
		if r.Event == qdriver.Request {
			calls++
		}
		if r.Event == qdriver.Step {
			shows++
		}
	}
	if calls == 0 || shows == 0 {
		t.Fatalf("actor adapter saw calls=%d DOM shows=%d", calls, shows)
	}
	for _, tr := range traces {
		if len(tr.Pending) > 0 {
			t.Fatalf("actor left %d requests unanswered", len(tr.Pending))
		}
	}
	t.Logf("Bombadil protocol admitted %d actual read invocations and %d independent DOM shows; no finished-439 claim", calls, shows)
}
