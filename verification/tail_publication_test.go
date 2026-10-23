package verification

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/retention"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
)

func tailDigest(v any) string {
	raw, _ := json.Marshal(v)
	h := sha256.Sum256(raw)
	return hex.EncodeToString(h[:])
}
func TestTailManifestProperties(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		m := &tailMembershipMachine{now: time.UnixMilli(200000), connection: strings.Repeat("a", 64), selector: exploration.TailSelector{Environment: "test"}, keys: map[string][]string{}}
		m.store = exploration.NewTailMembershipSnapshots(func() time.Time { return m.now })
		m.RuleIssue(tc)
		cap := m.caps[0]
		refs := []exploration.TailMembershipRef{}
		for _, key := range m.keys[cap.Token] {
			refs = append(refs, exploration.TailMembershipRef{Token: cap.Token, Key: key})
		}
		draft, e := m.store.Resolve(m.connection, m.selector, refs)
		if e != nil {
			tc.Fatal(e)
		}
		manifest := retention.TailManifest{Version: retention.TailEvidenceVersion, Selection: *draft}
		manifest.ID = tailDigest(manifest)
		if e = manifest.Validate(); e != nil {
			tc.Fatal(e)
		}
		view := retention.Evidence{Origin: "tail-selection", Tail: &manifest}
		if view.Validate() != nil {
			tc.Fatal("tail origin rejected")
		}
		view.Origin = "saved-capture"
		if view.Validate() == nil {
			tc.Fatal("tail masqueraded as captured evidence")
		}
		raw, _ := json.Marshal(manifest)
		var restored retention.TailManifest
		if json.Unmarshal(raw, &restored) != nil || restored.Validate() != nil || !reflect.DeepEqual(manifest, restored) {
			tc.Fatal("durable round trip changed membership")
		}
		// Integrity validation cannot depend on the live proof's remaining lifetime.
		m.now = cap.ExpiresAt.Add(time.Hour)
		if restored.Validate() != nil {
			tc.Fatal("durable evidence expired")
		}
		restored.Selection.Members[0].FromMS++
		restored.Selection.Digest = ""
		restored.Selection.Digest = tailDigest(restored.Selection)
		restored.ID = ""
		restored.ID = tailDigest(restored)
		if restored.Validate() == nil {
			tc.Fatal("rehashing forged read window bypassed validation")
		}
	}, options(50)...)
}
func TestQuintObservedTailPublication(t *testing.T) {
	url := os.Getenv("GOO_XTDB_TEST_URL")
	if url == "" {
		t.Skip("optional XTDB instance required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	m := &tailMembershipMachine{now: time.UnixMilli(200000), connection: strings.Repeat("b", 64), selector: exploration.TailSelector{Environment: "test"}, keys: map[string][]string{}}
	m.store = exploration.NewTailMembershipSnapshots(func() time.Time { return m.now })
	tracker := exploration.TailTracker{}
	rows, e := tracker.Observe([]json.RawMessage{json.RawMessage(`{"timestamp_ns":"150000000000","n":18446744073709551615}`)})
	if e != nil {
		t.Fatal(e)
	}
	cap, e := m.store.Freeze(exploration.TailMembershipBatch{Connection: m.connection, Selector: m.selector, FromMS: 100000, ToMS: 160000, ObservedAt: m.now, Rows: rows})
	if e != nil {
		t.Fatal(e)
	}
	in := retention.TailIntent{Version: retention.TailEvidenceVersion, RequestID: tailDigest(time.Now().String()), Connection: m.connection, Selector: m.selector, Refs: []exploration.TailMembershipRef{{Token: cap.Token, Key: rows[0].Key}}, Note: "observed replay"}
	store := retention.Store{URL: url}
	first, e := store.PublishTail(ctx, m.store, in)
	if e != nil {
		t.Fatal(e)
	}
	m.now = cap.ExpiresAt
	recovered, e := store.PublishTail(ctx, nil, in)
	if e != nil {
		t.Fatal(e)
	}
	changed := in
	changed.Note = "another intent"
	_, conflict := store.PublishTail(ctx, nil, changed)
	absent := in
	absent.RequestID = tailDigest(in.RequestID + "absent")
	_, expired := store.PublishTail(ctx, m.store, absent)
	loaded, e := store.LoadTail(ctx, first.Manifest.ID)
	if e != nil {
		t.Fatal(e)
	}
	first.AwaitToken = ""
	recovered.AwaitToken = ""
	durable := reflect.DeepEqual(first, recovered) && reflect.DeepEqual(*loaded, first.Manifest) && loaded.Validate() == nil
	check := func(bad int) bool {
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		actions := []string{"publish", "recover", "conflict", "absent"}
		denied := []bool{false, false, errors.Is(conflict, retention.ErrTailIntentConflict), errors.Is(expired, exploration.ErrTailMembershipUnavailable)}
		for i, action := range actions {
			ok := durable
			refused := denied[i]
			if i == bad {
				if i < 2 {
					ok = false
				} else {
					refused = false
				}
			}
			rec.Record(ctx, action, nil, "obs.durable", ok, "obs.refused", refused)
		}
		checker, e := validate.NewChecker("model/tail_publication.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("valid publication trace rejected")
	}
	for bad := 0; bad < 4; bad++ {
		if check(bad) {
			t.Fatal("forged publication trace accepted", bad)
		}
	}
}
