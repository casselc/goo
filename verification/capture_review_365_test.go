package verification

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

func reviewCapture365(t *testing.T) (exploration.CaptureStore, exploration.SavedCapture, []byte) {
	t.Helper()
	raw, err := os.ReadFile("../testdata/storage/exploration-capture.json")
	if err != nil {
		t.Fatal(err)
	}
	var c exploration.SavedCapture
	if err = json.Unmarshal(raw, &c); err != nil {
		t.Fatal(err)
	}
	row := bytes.Clone(c.Explore.Rows[0])
	c.Explore.Rows = make([]json.RawMessage, 101)
	for i := range c.Explore.Rows {
		c.Explore.Rows[i] = bytes.Clone(row)
	}
	c.Explore.Count = "900"
	c.Explore.HasMore = true
	c.Explore.Groups = nil
	c.Explore.TraceExamples = nil
	c.ID = ""
	body, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(body)
	c.ID = hex.EncodeToString(sum[:])
	body, err = json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	store := exploration.CaptureStore{Dir: t.TempDir()}
	if err = os.WriteFile(filepath.Join(store.Dir, c.ID+".json"), body, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = store.Load(c.ID); err != nil {
		t.Fatal(err)
	}
	return store, c, row
}
func TestHegelCaptureReview365(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		capture, c, raw := reviewCapture365(t)
		store := workbench.Store{Dir: t.TempDir()}
		resolve := capture.ResolveRecord
		input := workbench.CaptureReviewTaskInput{Version: 1, Name: "classification", Title: "Observed examples", Labels: []workbench.CaptureReviewLabel{{Key: "positive", Title: "Positive"}, {Key: "negative", Title: "Explicit negative"}}}
		task, err := store.CreateCaptureReviewTask(input)
		if err != nil {
			tc.Fatal(err)
		}
		taskID := task.ID
		input.Labels[0].Key = "mutated"
		task.Labels[0].Title = "mutated"
		loaded, err := store.LoadCaptureReviewTask(taskID)
		if err != nil || loaded.Labels[0].Key != "positive" || loaded.Labels[0].Title != "Positive" {
			tc.Fatal("task input/output alias", err)
		}
		target := exploration.CaptureRecordRef{CaptureID: c.ID, Index: 100}
		var previous string
		var pinned *workbench.CaptureReviewManifest
		var frozen []byte
		for step := range 10 {
			kind := []string{"label", "uncertain", "abstain"}[hegel.Draw(tc, hegel.Integers(0, 2))]
			review := []string{"unreviewed", "reviewed", "disputed"}[hegel.Draw(tc, hegel.Integers(0, 2))]
			state := "active"
			if step > 0 && hegel.Draw(tc, hegel.Integers(0, 3)) == 0 {
				state = "withdrawn"
			}
			label := ""
			if kind == "label" {
				label = []string{"positive", "negative"}[hegel.Draw(tc, hegel.Integers(0, 1))]
			}
			if step == 0 {
				kind = "label"
				review = "reviewed"
				state = "active"
				label = "positive"
			}
			parents := []string{}
			if previous != "" {
				parents = append(parents, previous)
			}
			in := workbench.CaptureReviewDecisionInput{TaskID: taskID, Target: target, Parents: parents, Judgment: workbench.CaptureReviewJudgment{Kind: kind, Value: label}, Review: review, State: state, Note: "Explicit observation"}
			intentHash := sha256.Sum256([]byte(fmt.Sprintf("%s/%d", taskID, step)))
			requestID := hex.EncodeToString(intentHash[:])
			decision, err := store.CreateCaptureReviewDecisionOnce(requestID, in, resolve)
			if err != nil {
				tc.Fatal(err)
			}
			retried, retryErr := store.CreateCaptureReviewDecisionOnce(requestID, in, resolve)
			if retryErr != nil || retried.ID != decision.ID || retried.RecordedAt != decision.RecordedAt {
				tc.Fatal("same intent changed immutable revision/clock", retryErr)
			}
			changed := in
			changed.Note += " changed"
			if _, err := store.CreateCaptureReviewDecisionOnce(requestID, changed, resolve); err == nil {
				tc.Fatal("same request ID changed authored intent")
			}
			previous = decision.ID
			sum := sha256.Sum256(raw)
			if decision.Target != target || decision.Evidence.RecordDigest != hex.EncodeToString(sum[:]) || decision.Evidence.Reference != target || decision.Origin != "local-user-unverified" || decision.RecordedAt == "" {
				tc.Fatal("server did not bind exact occurrence/custody")
			}
			manifest, err := store.CreateCaptureReviewManifest(workbench.CaptureReviewManifestInput{TaskID: taskID, Selected: []string{decision.ID}, KnownHeads: []string{decision.ID}, Historical: true}, resolve)
			// Independent eligibility rule: deliberate abstention/uncertainty, missing
			// review, dispute and withdrawal never imply a negative training class.
			eligible := kind == "label" && review == "reviewed" && state == "active"
			if (err == nil) != eligible {
				tc.Fatal("direct judgment eligibility disagreement", kind, review, state, err)
			}
			if eligible {
				export, err := store.ExportCaptureReviewManifest(manifest.ID, resolve)
				if err != nil {
					tc.Fatal(err)
				}
				var line struct {
					TaskID, DecisionID, RecordDigest, Label string
					Target                                  exploration.CaptureRecordRef
					Record                                  json.RawMessage
				}
				if json.Unmarshal(bytes.TrimSpace(export), &line) != nil || line.TaskID != taskID || line.DecisionID != decision.ID || line.Target != target || line.Label != label || line.RecordDigest != hex.EncodeToString(sum[:]) || !bytes.Equal(line.Record, raw) || !bytes.Contains(export, []byte("18446744073709551615")) {
					tc.Fatal("export lost pinned label/raw number tokens")
				}
				if pinned == nil {
					pinned = manifest
					frozen = bytes.Clone(export)
				}
			}
			decision.Judgment.Value = "caller"
			decision.Evidence.Receipt.Title = "caller"
			again, err := store.LoadCaptureReviewDecision(previous)
			if err != nil || again.Judgment.Value != label || again.Evidence.Receipt.Title != c.Title {
				tc.Fatal("returned decision alias", err)
			}
			if pinned != nil {
				now, err := store.ExportCaptureReviewManifest(pinned.ID, resolve)
				if err != nil || !bytes.Equal(now, frozen) {
					tc.Fatal("later revision rewrote historical pinned export", err)
				}
			}
		}
		// Two supplied forks require explicit ancestry resolution; equal row bytes
		// at positions zero/100 remain separate task targets.
		create := func(parents []string, target exploration.CaptureRecordRef, task string) *workbench.CaptureReviewDecision {
			d, e := store.CreateCaptureReviewDecision(workbench.CaptureReviewDecisionInput{TaskID: task, Target: target, Parents: parents, Judgment: workbench.CaptureReviewJudgment{Kind: "label", Value: "positive"}, Review: "reviewed", State: "active"}, resolve)
			if e != nil {
				tc.Fatal(e)
			}
			return d
		}
		a := create([]string{previous}, target, taskID)
		b := create([]string{previous}, target, taskID)
		if _, err := store.CreateCaptureReviewManifest(workbench.CaptureReviewManifestInput{TaskID: taskID, Selected: []string{a.ID}, KnownHeads: []string{a.ID, b.ID}, Historical: true}, resolve); err == nil {
			tc.Fatal("known unresolved fork became consensus")
		}
		joined := create([]string{a.ID, b.ID}, target, taskID)
		if _, err := store.CreateCaptureReviewManifest(workbench.CaptureReviewManifestInput{TaskID: taskID, Selected: []string{joined.ID}, KnownHeads: []string{a.ID, b.ID}, Historical: true}, resolve); err != nil {
			tc.Fatal("explicit known-head resolution refused", err)
		}
		foreign := target
		foreign.Index = 0
		if _, err := store.CreateCaptureReviewDecision(workbench.CaptureReviewDecisionInput{TaskID: taskID, Target: foreign, Parents: []string{a.ID}, Judgment: workbench.CaptureReviewJudgment{Kind: "label", Value: "positive"}, Review: "reviewed", State: "active"}, resolve); err == nil {
			tc.Fatal("equal payload foreign occurrence joined")
		}
		other, err := store.CreateCaptureReviewTask(workbench.CaptureReviewTaskInput{Version: 1, Name: "foreign", Title: "Other vocabulary", Labels: []workbench.CaptureReviewLabel{{Key: "positive", Title: "Positive"}}})
		if err != nil {
			tc.Fatal(err)
		}
		if _, err := store.CreateCaptureReviewDecision(workbench.CaptureReviewDecisionInput{TaskID: other.ID, Target: target, Parents: []string{a.ID}, Judgment: workbench.CaptureReviewJudgment{Kind: "label", Value: "positive"}, Review: "reviewed", State: "active"}, resolve); err == nil {
			tc.Fatal("foreign task parent joined")
		}
		if _, err := store.CreateCaptureReviewManifest(workbench.CaptureReviewManifestInput{TaskID: taskID, Selected: nil, Historical: true}, resolve); err == nil {
			tc.Fatal("absence of decision became a label")
		}
		tooMany := workbench.CaptureReviewTaskInput{Version: 1, Name: "bounds", Title: "Bounds", Labels: make([]workbench.CaptureReviewLabel, 17)}
		if _, err := store.CreateCaptureReviewTask(tooMany); err == nil {
			tc.Fatal("label ceiling truncated")
		}
		if _, err := store.CreateCaptureReviewDecision(workbench.CaptureReviewDecisionInput{TaskID: taskID, Target: target, Judgment: workbench.CaptureReviewJudgment{Kind: "label", Value: "undeclared"}, Review: "reviewed", State: "active"}, resolve); err == nil {
			tc.Fatal("undeclared label admitted")
		}
		if _, err := store.CreateCaptureReviewDecision(workbench.CaptureReviewDecisionInput{TaskID: taskID, Target: target, Judgment: workbench.CaptureReviewJudgment{Kind: "label", Value: "positive"}, Review: "reviewed", State: "active", Note: strings.Repeat("x", 2049)}, resolve); err == nil {
			tc.Fatal("oversize note truncated")
		}
		disk, err := os.ReadFile(filepath.Join(capture.Dir, c.ID+".json"))
		if err != nil {
			tc.Fatal(err)
		}
		var intact exploration.SavedCapture
		if json.Unmarshal(disk, &intact) != nil || !reflect.DeepEqual(intact.Explore.Rows, c.Explore.Rows) {
			tc.Fatal("review mutated captured evidence")
		}
	}, options(40)...)
}
