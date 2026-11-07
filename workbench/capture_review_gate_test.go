package workbench

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/casselc/goo/exploration"
)

// These private types demonstrate the readiness contract, not a production
// review store, HTTP API, task registry or annotation UI.
type reviewTaskGate365 struct {
	Version string   `json:"version"`
	Labels  []string `json:"labels"`
}
type reviewDecisionGate365 struct {
	Task         string                       `json:"task"`
	Target       exploration.CaptureRecordRef `json:"target"`
	RecordDigest string                       `json:"recordDigest"`
	Parents      []string                     `json:"parents"`
	Judgment     string                       `json:"judgment"`
	Label        string                       `json:"label"`
	Review       string                       `json:"review"`
	State        string                       `json:"state"`
	RecordedAt   string                       `json:"recordedAt"`
	Origin       string                       `json:"origin"`
}
type reviewManifestGate365 struct {
	Task         string                       `json:"task"`
	Decision     string                       `json:"decision"`
	KnownHeads   []string                     `json:"knownHeads"`
	Target       exploration.CaptureRecordRef `json:"target"`
	RecordDigest string                       `json:"recordDigest"`
	Selection    string                       `json:"selection"`
}

func eligibleReviewGate365(task reviewTaskGate365, d reviewDecisionGate365, knownHeads []string) bool {
	if d.Judgment != "label" || !slices.Contains(task.Labels, d.Label) || d.Review != "reviewed" || d.State != "active" {
		return false
	}
	// A resolution must explicitly cover all supplied competing heads. Nothing
	// here claims that unknown concurrent branches do not exist.
	if len(knownHeads) > 8 || len(d.Parents) > 8 {
		return false
	}
	if len(knownHeads) > 1 {
		for _, id := range knownHeads {
			if !slices.Contains(d.Parents, id) {
				return false
			}
		}
	}
	return true
}
func publishReviewGate365(s Store, kind string, body any) (string, error) {
	id, err := digest(body)
	if err != nil {
		return "", err
	}
	if err = s.publish(kind, id, body); err != nil {
		return "", err
	}
	var raw json.RawMessage
	if err = s.read(kind, id, &raw); err != nil {
		return "", err
	}
	actual, err := digest(raw)
	if err != nil {
		return "", err
	}
	if actual != id {
		return "", os.ErrInvalid
	}
	return id, nil
}

// Export joins pinned stored artifacts again; no draft object is authoritative.
func exportReviewGate365(s Store, captures exploration.CaptureStore, manifest reviewManifestGate365) ([]byte, error) {
	read := func(kind, id string, out any) error {
		if err := s.read(kind, id, out); err != nil {
			return err
		}
		actual, err := digest(out)
		if err != nil {
			return err
		}
		if actual != id {
			return errors.New("pinned artifact digest mismatch")
		}
		return nil
	}
	var task reviewTaskGate365
	var decision reviewDecisionGate365
	if err := read("gate-tasks", manifest.Task, &task); err != nil {
		return nil, err
	}
	if err := read("gate-decisions", manifest.Decision, &decision); err != nil {
		return nil, err
	}
	if manifest.Selection != "explicit-pinned-branch-not-current-consensus" || decision.Task != manifest.Task || decision.Target != manifest.Target || decision.RecordDigest != manifest.RecordDigest || !eligibleReviewGate365(task, decision, manifest.KnownHeads) {
		return nil, errors.New("pinned task, target or judgment mismatch")
	}
	record, err := captures.ResolveRecord(manifest.Target)
	if err != nil {
		return nil, err
	}
	if record.RecordDigest != manifest.RecordDigest {
		return nil, errors.New("pinned record digest mismatch")
	}
	return json.Marshal(struct {
		Record json.RawMessage `json:"record"`
		Label  string          `json:"label"`
	}{record.Record, decision.Label})
}

func TestCapturedReviewPublicationReadiness365(t *testing.T) {
	wire, err := os.ReadFile("../testdata/storage/exploration-capture.json")
	if err != nil {
		t.Fatal(err)
	}
	var capture exploration.SavedCapture
	if err = json.Unmarshal(wire, &capture); err != nil {
		t.Fatal(err)
	}
	captureDir := t.TempDir()
	if err = os.WriteFile(filepath.Join(captureDir, capture.ID+".json"), wire, 0600); err != nil {
		t.Fatal(err)
	}
	captures := exploration.CaptureStore{Dir: captureDir}
	record, err := captures.ResolveRecord(exploration.CaptureRecordRef{CaptureID: capture.ID, Index: 0})
	if err != nil {
		t.Fatal(err)
	}
	store := Store{Dir: t.TempDir()}
	task := reviewTaskGate365{Version: "gate-only/categorical-task/1", Labels: []string{"positive", "negative"}}
	taskID, err := publishReviewGate365(store, "gate-tasks", task)
	if err != nil {
		t.Fatal(err)
	}
	root := reviewDecisionGate365{Task: taskID, Target: record.Reference, RecordDigest: record.RecordDigest, Judgment: "uncertain", Review: "unreviewed", State: "active", RecordedAt: "2026-11-07T00:00:00Z", Origin: "local-user-unverified"}
	rootID, err := publishReviewGate365(store, "gate-decisions", root)
	if err != nil {
		t.Fatal(err)
	}
	branches := []reviewDecisionGate365{root, root}
	branches[0].Judgment = "label"
	branches[0].Label = "positive"
	branches[1].Judgment = "abstain"
	ids := make([]string, 2)
	errs := make([]error, 2)
	var writers sync.WaitGroup
	for i := range branches {
		branches[i].Parents = []string{rootID}
		writers.Go(func() { ids[i], errs[i] = publishReviewGate365(store, "gate-decisions", branches[i]) })
	}
	writers.Wait()
	for _, err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if ids[0] == ids[1] {
		t.Fatal("concurrent corrections collapsed")
	}
	for i, id := range ids {
		var loaded reviewDecisionGate365
		if err := store.read("gate-decisions", id, &loaded); err != nil || loaded.Judgment != branches[i].Judgment {
			t.Fatal("a branch was replaced", err)
		}
	}
	resolution := root
	resolution.Parents = slices.Clone(ids)
	slices.Sort(resolution.Parents)
	resolution.Judgment = "label"
	resolution.Label = "negative"
	resolution.Review = "reviewed"
	if !eligibleReviewGate365(task, resolution, ids) {
		t.Fatal("explicit known-branch resolution refused")
	}
	for _, change := range []func(*reviewDecisionGate365){func(d *reviewDecisionGate365) { d.Judgment = "uncertain" }, func(d *reviewDecisionGate365) { d.Judgment = "abstain" }, func(d *reviewDecisionGate365) { d.Review = "unreviewed" }, func(d *reviewDecisionGate365) { d.Review = "disputed" }, func(d *reviewDecisionGate365) { d.State = "withdrawn" }, func(d *reviewDecisionGate365) { d.Label = "other" }, func(d *reviewDecisionGate365) { d.Parents = nil }} {
		bad := resolution
		change(&bad)
		if eligibleReviewGate365(task, bad, ids) {
			t.Fatal("ineligible decision silently became label", bad)
		}
	}
	decisionID, err := publishReviewGate365(store, "gate-decisions", resolution)
	if err != nil {
		t.Fatal(err)
	}
	manifest := reviewManifestGate365{Task: taskID, Decision: decisionID, KnownHeads: ids, Target: record.Reference, RecordDigest: record.RecordDigest, Selection: "explicit-pinned-branch-not-current-consensus"}
	manifestID, err := publishReviewGate365(store, "gate-manifests", manifest)
	if err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(store.Dir, "gate-manifests", manifestID+".json"))
	if err != nil {
		t.Fatal(err)
	}
	withdrawal := resolution
	withdrawal.Parents = []string{decisionID}
	withdrawal.State = "withdrawn"
	withdrawal.Label = ""
	withdrawal.RecordedAt = "2026-11-07T00:00:01Z"
	if _, err := publishReviewGate365(store, "gate-decisions", withdrawal); err != nil {
		t.Fatal(err)
	}
	var reopened reviewManifestGate365
	if err = store.read("gate-manifests", manifestID, &reopened); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(filepath.Join(store.Dir, "gate-manifests", manifestID+".json"))
	if err != nil || !bytes.Equal(before, after) || reopened.Decision != decisionID {
		t.Fatal("later withdrawal rewrote pinned export", err)
	}
	line, err := exportReviewGate365(store, captures, reopened)
	if err != nil || !bytes.Contains(line, []byte("18446744073709551615")) || !bytes.Contains(line, []byte("1700000000000000123")) || !bytes.Contains(line, []byte(`"label":"negative"`)) {
		t.Fatal("pinned export lost original tokens or stored decision", err)
	}
	foreignTask := task
	foreignTask.Version = "gate-only/categorical-task/2"
	foreignTaskID, err := publishReviewGate365(store, "gate-tasks", foreignTask)
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*reviewManifestGate365){func(m *reviewManifestGate365) { m.Task = foreignTaskID }, func(m *reviewManifestGate365) { m.Target.Index++ }, func(m *reviewManifestGate365) { m.RecordDigest = strings.Repeat("a", 64) }} {
		bad := reopened
		change(&bad)
		if _, err := exportReviewGate365(store, captures, bad); err == nil {
			t.Fatal("foreign pinned task/target/digest joined")
		}
	}
	if _, err := publishReviewGate365(store, "gate-manifests", strings.Repeat("x", MaxBytes+1)); err == nil {
		t.Fatal("publication silently exceeded bound")
	}
	// Existing publication is create-or-reuse, so the owning artifact layer must
	// verify a winner rather than accepting the filesystem path as identity.
	sum := sha256.Sum256(before)
	if hex.EncodeToString(sum[:]) != manifestID {
		t.Fatal("manifest path is not its content digest")
	}
	if err := os.WriteFile(filepath.Join(store.Dir, "gate-manifests", manifestID+".json"), []byte(`{"forged":true}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := publishReviewGate365(store, "gate-manifests", manifest); err == nil {
		t.Fatal("existing wrong winner was trusted")
	}
}
