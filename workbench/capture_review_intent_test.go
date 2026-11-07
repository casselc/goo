package workbench

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
)

func TestCaptureReviewDurableIntentRetry(t *testing.T) {
	s, resolve, target := reviewFixture(t)
	input := CaptureReviewTaskInput{Version: 1, Name: "review", Title: "Review", Labels: []CaptureReviewLabel{{"yes", "Yes"}, {"no", "No"}}}
	const taskRequest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	var results [2]*CaptureReviewTask
	var errs [2]error
	var wg sync.WaitGroup
	for i := range results {
		wg.Go(func() { results[i], errs[i] = s.CreateCaptureReviewTaskOnce(taskRequest, input) })
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if results[0].ID != results[1].ID || results[0].RecordedAt != results[1].RecordedAt {
		t.Fatal("concurrent intent did not recover same task")
	}
	changed := input
	changed.Title = "Changed"
	if _, err := s.CreateCaptureReviewTaskOnce(taskRequest, changed); !errors.Is(err, ErrCaptureReviewIntentConflict) {
		t.Fatal("changed intent reused", err)
	}
	decisionInput := reviewInput(results[0], target)
	decisionInput.Judgment.Value = "yes"
	first, err := s.CreateCaptureReviewDecisionOnce(strings.Repeat("b", 64), decisionInput, resolve)
	if err != nil {
		t.Fatal(err)
	}
	retry, err := s.CreateCaptureReviewDecisionOnce(strings.Repeat("b", 64), decisionInput, resolve)
	if err != nil || retry.ID != first.ID || retry.RecordedAt != first.RecordedAt {
		t.Fatal("lost acknowledgment created another branch", err)
	}
	if _, err = s.CreateCaptureReviewDecisionOnce(taskRequest, decisionInput, resolve); !errors.Is(err, ErrCaptureReviewIntentConflict) {
		t.Fatal("kind collision accepted", err)
	}
	m := CaptureReviewManifestInput{TaskID: results[0].ID, Selected: []string{first.ID}, KnownHeads: []string{}, Historical: true}
	manifest, err := s.CreateCaptureReviewManifestOnce(strings.Repeat("c", 64), m, resolve)
	if err != nil {
		t.Fatal(err)
	}
	again, err := s.CreateCaptureReviewManifestOnce(strings.Repeat("c", 64), m, resolve)
	if err != nil || manifest.ID != again.ID {
		t.Fatal("manifest retry drift", err)
	}
	for _, kind := range []string{"capture-review-tasks", "capture-review-decisions", "capture-review-manifests"} {
		ids, e := s.ids(kind)
		if e != nil || len(ids) != 1 {
			t.Fatal("retry orphaned revisions", kind, ids, e)
		}
	}
	m.Selected = slices.Clone(m.Selected)
	m.KnownHeads = []string{first.ID}
	if _, err = s.CreateCaptureReviewManifestOnce(strings.Repeat("c", 64), m, resolve); !errors.Is(err, ErrCaptureReviewIntentConflict) {
		t.Fatal("manifest changed intent accepted", err)
	}
}

func TestCaptureReviewIntentRejectsClockCorruption(t *testing.T) {
	s, _, _ := reviewFixture(t)
	input := CaptureReviewTaskInput{Version: 1, Name: "review", Title: "Review", Labels: []CaptureReviewLabel{{"yes", "Yes"}}}
	request := strings.Repeat("d", 64)
	if _, err := s.CreateCaptureReviewTaskOnce(request, input); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(s.Dir, "capture-review-intents", request+".json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var intent captureReviewIntent
	if err = json.Unmarshal(raw, &intent); err != nil {
		t.Fatal(err)
	}
	intent.RecordedAt = "2001-01-01T00:00:00Z"
	raw, err = json.Marshal(intent)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreateCaptureReviewTaskOnce(request, input); err == nil {
		t.Fatal("valid JSON clock corruption created another revision")
	}
	ids, err := s.ids("capture-review-tasks")
	if err != nil || len(ids) != 1 {
		t.Fatal("corrupt retry published revision", ids, err)
	}
}
