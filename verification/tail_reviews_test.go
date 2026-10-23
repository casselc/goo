package verification

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/retention"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

func TestQuintObservedTailReviews(t *testing.T) {
	url := os.Getenv("GOO_XTDB_TEST_URL")
	if url == "" {
		t.Skip("optional XTDB instance required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	now := time.UnixMilli(200000)
	bank := exploration.NewTailMembershipSnapshots(func() time.Time { return now })
	tracker := exploration.TailTracker{}
	rows, e := tracker.Observe([]json.RawMessage{json.RawMessage(`{"timestamp_ns":"150000000000","n":18446744073709551615}`)})
	if e != nil {
		t.Fatal(e)
	}
	in := retention.TailIntent{Version: retention.TailEvidenceVersion, Connection: tailDigest(t.Name()), RequestID: tailDigest(time.Now().String()), Selector: exploration.TailSelector{Environment: "test"}}
	tag := "review:model-" + in.RequestID[:12]
	in.Tags = []string{tag}
	cap, e := bank.Freeze(exploration.TailMembershipBatch{Connection: in.Connection, Selector: in.Selector, FromMS: 100000, ToMS: 160000, ObservedAt: now, Rows: rows})
	if e != nil {
		t.Fatal(e)
	}
	in.Refs = []exploration.TailMembershipRef{{Token: cap.Token, Key: rows[0].Key}}
	store := retention.Store{URL: url}
	original, e := store.PublishTail(ctx, bank, in)
	if e != nil {
		t.Fatal(e)
	}
	type observation struct {
		action             string
		heads, active      int
		visible, immutable bool
	}
	observations := []observation{}
	observe := func(action string) {
		history, e := store.Annotations(ctx, original.Manifest.ID)
		if e != nil {
			t.Fatal(e)
		}
		pinned, e := store.Annotation(ctx, original.Manifest.ID, original.Annotation.ID)
		if e != nil {
			t.Fatal(e)
		}
		view, e := store.ReadEvidence(ctx, original.Manifest.ID)
		if e != nil {
			t.Fatal(e)
		}
		catalog, e := store.CatalogEvidence(ctx, retention.EvidenceCatalogQuery{CatalogQuery: retention.CatalogQuery{Tag: tag}, Origin: "tail-selection"})
		if e != nil {
			t.Fatal(e)
		}
		o := observation{action: action, visible: len(catalog.Items) == 1 && catalog.Items[0].ID == original.Manifest.ID, immutable: reflect.DeepEqual(pinned, original.Annotation) && view.Capture == nil && reflect.DeepEqual(*view.Tail, original.Manifest)}
		for _, item := range history.Items {
			if item.Head {
				o.heads++
				if item.Revision.State == "active" {
					o.active++
				}
			}
		}
		observations = append(observations, o)
	}
	store.AwaitToken = original.AwaitToken
	observe("publish")
	corrected, e := store.CorrectOnce(ctx, original.Manifest.ID, workbench.MeaningInput{Target: original.Annotation.Target, Parent: original.Annotation.ID, Tags: []string{tag, "review:changed"}, Note: "Interpretation changed", State: "active"}, tailDigest(in.RequestID+"correct"))
	if e != nil {
		t.Fatal(e)
	}
	store.AwaitToken = corrected.AwaitToken
	observe("correct")
	withdrawn, e := store.CorrectOnce(ctx, original.Manifest.ID, workbench.MeaningInput{Target: original.Annotation.Target, Parent: corrected.ID, State: "withdrawn"}, tailDigest(in.RequestID+"withdraw"))
	if e != nil {
		t.Fatal(e)
	}
	store.AwaitToken = withdrawn.AwaitToken
	observe("withdraw")
	now = cap.ExpiresAt
	recovered, e := store.PublishTail(ctx, nil, in)
	if e != nil || recovered.Annotation.ID != original.Annotation.ID || !recovered.RetainedAt.Equal(original.RetainedAt) {
		t.Fatal("retry restored wrong origin receipt", e)
	}
	store.AwaitToken = recovered.AwaitToken
	observe("retry_original")
	check := func(bad int) bool {
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for i, o := range observations {
			if i == bad {
				switch i {
				case 0:
					o.immutable = false
				case 1:
					o.heads++
				case 2:
					o.active++
				case 3:
					o.visible = true
				}
			}
			rec.Record(ctx, o.action, nil, "obs.heads", o.heads, "obs.active", o.active, "obs.visible", o.visible, "obs.immutable", o.immutable)
		}
		checker, e := validate.NewChecker("model/tail_reviews.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		result, e := checker.Check(ctx, sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return result.OK()
	}
	if !check(-1) {
		t.Fatal("valid tail review replay rejected")
	}
	for bad := 0; bad < 4; bad++ {
		if check(bad) {
			t.Fatal("forged review replay accepted", bad)
		}
	}
}
