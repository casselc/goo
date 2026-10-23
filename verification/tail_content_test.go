package verification

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/casselc/goo/retention"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

func TestQuintTailContentLedger(t *testing.T) {
	file := os.Getenv("GOO_TAIL_CONTENT_LEDGER")
	if file == "" {
		t.Skip("observed tail-content browser ledger required")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	check := func(bad int) bool {
		var l struct {
			Receipt  retention.TailReceipt
			Document workbench.Document
			Current  workbench.MeaningRevision
			Original retention.EvidenceContent
			Events   []struct {
				Action           string
				Reads, Published int
				Coherent         bool
			}
		}
		if e = json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch bad {
		case 0:
			l.Events[0].Reads++
		case 1:
			l.Events[1].Published++
		case 2:
			l.Document.Cards[0].Review = l.Current.ID
		case 3:
			l.Original.RetainedAt = "2000-01-01T00:00:00Z"
		case 4:
			l.Original.ManifestJSON = "{}"
		}
		var manifest retention.TailManifest
		e := json.Unmarshal([]byte(l.Original.ManifestJSON), &manifest)
		at, clockErr := time.Parse(time.RFC3339Nano, l.Original.RetainedAt)
		coherent := e == nil && manifest.Validate() == nil && l.Original.Evidence.Validate() == nil && l.Original.Origin == "tail-selection" && l.Original.Capture == nil && equalTailJSON(manifest, l.Receipt.Manifest) && clockErr == nil && at.Equal(l.Receipt.RetainedAt) && l.Current.Parent == l.Receipt.Annotation.ID && workbench.ValidateMeaningRevision(l.Current) == nil
		if len(l.Document.Cards) != 1 {
			coherent = false
		} else {
			card := l.Document.Cards[0]
			coherent = coherent && card.Reference == manifest.ID && card.Review == l.Receipt.Annotation.ID && card.ReadBasis != ""
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, event := range l.Events {
			rec.Record(context.Background(), event.Action, nil, "obs.reads", event.Reads, "obs.published", event.Published, "obs.coherent", coherent && event.Coherent)
		}
		checker, e := validate.NewChecker("model/annotations.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		result, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return result.OK()
	}
	if !check(-1) {
		t.Fatal("observed tail-content journey rejected")
	}
	for bad := 0; bad < 5; bad++ {
		if check(bad) {
			t.Fatal("forged tail-content journey accepted", bad)
		}
	}
}

func equalTailJSON(a, b retention.TailManifest) bool {
	x, e := json.Marshal(a)
	y, f := json.Marshal(b)
	return e == nil && f == nil && bytes.Equal(x, y)
}
