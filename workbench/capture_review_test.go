package workbench

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/casselc/goo/exploration"
)

func reviewFixture(t *testing.T) (Store, CaptureReviewResolver, exploration.CaptureRecordRef) {
	t.Helper()
	raw, err := os.ReadFile("../testdata/storage/exploration-capture.json")
	if err != nil {
		t.Fatal(err)
	}
	var c exploration.SavedCapture
	if err = json.Unmarshal(raw, &c); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err = os.WriteFile(filepath.Join(dir, c.ID+".json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	return Store{Dir: t.TempDir()}, exploration.CaptureStore{Dir: dir}.ResolveRecord, exploration.CaptureRecordRef{CaptureID: c.ID, Index: 0}
}
func reviewTask(t *testing.T, s Store) *CaptureReviewTask {
	t.Helper()
	v, err := s.CreateCaptureReviewTask(CaptureReviewTaskInput{Version: 1, Name: "timeouts", Title: "Timeout review", Labels: []CaptureReviewLabel{{"timeout", "Timeout"}, {"not-timeout", "Not timeout"}}})
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func reviewInput(task *CaptureReviewTask, target exploration.CaptureRecordRef) CaptureReviewDecisionInput {
	return CaptureReviewDecisionInput{TaskID: task.ID, Target: target, Parents: []string{}, Judgment: CaptureReviewJudgment{Kind: "label", Value: "timeout"}, Review: "reviewed", State: "active", Note: "local judgment"}
}
func TestCaptureReviewTaskAndDecisionIntegrity(t *testing.T) {
	s, resolve, target := reviewFixture(t)
	task := reviewTask(t, s)
	task.Labels[0].Title = "caller changed"
	loaded, err := s.LoadCaptureReviewTask(task.ID)
	if err != nil || loaded.Labels[0].Title != "Timeout" {
		t.Fatal("task input/return aliased", err)
	}
	in := reviewInput(loaded, target)
	decision, err := s.CreateCaptureReviewDecision(in, resolve)
	if err != nil {
		t.Fatal(err)
	}
	if decision.Evidence.Reference != target || decision.Evidence.Receipt.ID != target.CaptureID || decision.Evidence.RecordDigest == "" || decision.Origin != captureReviewOrigin {
		t.Fatal(decision)
	}
	decision.Evidence.Receipt.Question.Explore.Environment = "changed"
	again, err := s.LoadCaptureReviewDecision(decision.ID)
	if err != nil || again.Evidence.Receipt.Question.Explore.Environment == "changed" {
		t.Fatal("decision evidence alias", err)
	}
	for _, change := range []func(*CaptureReviewDecisionInput){func(v *CaptureReviewDecisionInput) { v.TaskID = strings.Repeat("a", 64) }, func(v *CaptureReviewDecisionInput) { v.Target.Index = -1 }, func(v *CaptureReviewDecisionInput) { v.Judgment.Value = "implicit-negative" }, func(v *CaptureReviewDecisionInput) {
		v.Judgment = CaptureReviewJudgment{Kind: "abstain", Value: "timeout"}
	}, func(v *CaptureReviewDecisionInput) { v.Parents = []string{decision.ID, decision.ID} }, func(v *CaptureReviewDecisionInput) { v.Review = "trusted" }, func(v *CaptureReviewDecisionInput) { v.State = "withdrawn" }} {
		bad := in
		change(&bad)
		if _, err = s.CreateCaptureReviewDecision(bad, resolve); err == nil {
			t.Fatal("invalid decision accepted", bad)
		}
	}
	foreign := reviewTask(t, s)
	bad := reviewInput(foreign, target)
	bad.Parents = []string{again.ID}
	if _, err = s.CreateCaptureReviewDecision(bad, resolve); err == nil {
		t.Fatal("foreign task parent accepted")
	}
	if _, err = s.CreateCaptureReviewDecision(in, func(ref exploration.CaptureRecordRef) (*exploration.CapturedRecord, error) {
		v, e := resolve(ref)
		v.Record = []byte(`{"forged":true}`)
		return v, e
	}); err == nil {
		t.Fatal("forged resolver hash accepted")
	}
	data, err := os.ReadFile(filepath.Join(s.Dir, "capture-review-decisions", again.ID+".json"))
	if err != nil {
		t.Fatal(err)
	}
	data = bytes.Replace(data, []byte("local judgment"), []byte("other judgment"), 1)
	if err = os.WriteFile(filepath.Join(s.Dir, "capture-review-decisions", again.ID+".json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = s.LoadCaptureReviewDecision(again.ID); err == nil {
		t.Fatal("corrupt decision accepted")
	}
}
func TestCaptureReviewPinnedForkResolutionExport(t *testing.T) {
	s, resolve, target := reviewFixture(t)
	task := reviewTask(t, s)
	base := reviewInput(task, target)
	base.Judgment = CaptureReviewJudgment{Kind: "uncertain"}
	base.Review = "unreviewed"
	root, err := s.CreateCaptureReviewDecision(base, resolve)
	if err != nil {
		t.Fatal(err)
	}
	branches := make([]*CaptureReviewDecision, 2)
	errs := make([]error, 2)
	var wg sync.WaitGroup
	for i := range branches {
		wg.Go(func() {
			in := reviewInput(task, target)
			in.Parents = []string{root.ID}
			if i == 1 {
				in.Judgment.Value = "not-timeout"
			}
			branches[i], errs[i] = s.CreateCaptureReviewDecision(in, resolve)
		})
	}
	wg.Wait()
	for _, e := range errs {
		if e != nil {
			t.Fatal(e)
		}
	}
	heads := []string{branches[0].ID, branches[1].ID}
	if heads[0] == heads[1] {
		t.Fatal("fork overwritten")
	}
	input := CaptureReviewManifestInput{TaskID: task.ID, Selected: []string{heads[0]}, KnownHeads: heads, Historical: true}
	if _, err = s.CreateCaptureReviewManifest(input, resolve); err == nil {
		t.Fatal("unresolved known fork exported")
	}
	resolved := reviewInput(task, target)
	resolved.Parents = slices.Clone(heads)
	decision, err := s.CreateCaptureReviewDecision(resolved, resolve)
	if err != nil {
		t.Fatal(err)
	}
	input.Selected = []string{decision.ID}
	manifest, err := s.CreateCaptureReviewManifest(input, resolve)
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Revisions) != 4 || len(manifest.Items) != 1 || manifest.Mode != "pinned-review-branches" {
		t.Fatal(manifest)
	}
	before, err := s.ExportCaptureReviewManifest(manifest.ID, resolve)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(before, []byte("18446744073709551615")) || !bytes.Contains(before, []byte("1700000000000000123")) {
		t.Fatal("raw number tokens rounded")
	}
	withdrawn := resolved
	withdrawn.Parents = []string{decision.ID}
	withdrawn.State = "withdrawn"
	later, err := s.CreateCaptureReviewDecision(withdrawn, resolve)
	if err != nil {
		t.Fatal(err)
	}
	after, err := s.ExportCaptureReviewManifest(manifest.ID, resolve)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("later withdrawal changed pinned export", err)
	}
	for _, id := range []string{root.ID, later.ID} {
		bad := input
		bad.Selected = []string{id}
		bad.KnownHeads = nil
		if _, err = s.CreateCaptureReviewManifest(bad, resolve); err == nil {
			t.Fatal("nonlabel/unreviewed/withdrawn decision exported")
		}
	}
	for _, change := range []func(*CaptureReviewDecisionInput){func(v *CaptureReviewDecisionInput) { v.Judgment = CaptureReviewJudgment{Kind: "abstain"} }, func(v *CaptureReviewDecisionInput) { v.Review = "disputed" }, func(v *CaptureReviewDecisionInput) { v.Review = "unreviewed" }} {
		bad := reviewInput(task, target)
		change(&bad)
		v, e := s.CreateCaptureReviewDecision(bad, resolve)
		if e != nil {
			t.Fatal(e)
		}
		m := input
		m.Selected = []string{v.ID}
		m.KnownHeads = nil
		if _, e = s.CreateCaptureReviewManifest(m, resolve); e == nil {
			t.Fatal("ineligible label exported")
		}
	}
	duplicate := input
	duplicate.Selected = []string{heads[0], heads[1]}
	duplicate.KnownHeads = nil
	if _, err = s.CreateCaptureReviewManifest(duplicate, resolve); err == nil {
		t.Fatal("duplicate target accepted")
	}
	historical := input
	historical.Historical = false
	if _, err = s.CreateCaptureReviewManifest(historical, resolve); err == nil {
		t.Fatal("consensus/latest mode accepted")
	}
	manifest.Items[0].Label = "changed"
	reopened, err := s.LoadCaptureReviewManifest(manifest.ID)
	if err != nil || reopened.Items[0].Label != "timeout" {
		t.Fatal("returned manifest aliased store", err)
	}
	if _, err = s.ExportCaptureReviewManifest(manifest.ID, func(ref exploration.CaptureRecordRef) (*exploration.CapturedRecord, error) {
		r, e := resolve(ref)
		r.Receipt.Title = "changed evidence"
		return r, e
	}); err == nil {
		t.Fatal("changed evidence metadata exported")
	}
}
func TestCaptureReviewBoundsAndDiamond(t *testing.T) {
	s, resolve, target := reviewFixture(t)
	task := reviewTask(t, s)
	input := reviewInput(task, target)
	root, err := s.CreateCaptureReviewDecision(input, resolve)
	if err != nil {
		t.Fatal(err)
	}
	heads := []string{root.ID}
	for range 4 {
		next := []string{}
		for range 4 {
			v := input
			v.Parents = slices.Clone(heads)
			d, e := s.CreateCaptureReviewDecision(v, resolve)
			if e != nil {
				t.Fatal(e)
			}
			next = append(next, d.ID)
		}
		heads = next
	}
	if _, err = s.LoadCaptureReviewDecision(heads[0]); err != nil {
		t.Fatal("bounded shared DAG refused", err)
	}
	all := map[string]*CaptureReviewDecision{}
	for i := 0; i < 64; i++ {
		id := strings.Repeat("a", 63) + string(rune('A'+i))
		parents := []string{}
		for j := 0; j < i; j++ {
			parents = append(parents, strings.Repeat("a", 63)+string(rune('A'+j)))
			if len(parents) == 8 {
				break
			}
		}
		all[id] = &CaptureReviewDecision{Parents: parents}
	}
	if ancestorReview(strings.Repeat("a", 63)+string(rune('A'+63)), "missing", all) {
		t.Fatal("nonancestor found")
	}
	labels := make([]CaptureReviewLabel, 17)
	if _, err = s.CreateCaptureReviewTask(CaptureReviewTaskInput{Version: 1, Name: "too-many", Title: "Bad", Labels: labels}); err == nil {
		t.Fatal("label limit ignored")
	}
	bad := input
	bad.Parents = make([]string, 9)
	if _, err = s.CreateCaptureReviewDecision(bad, resolve); err == nil {
		t.Fatal("parent limit ignored")
	}
	if _, err = s.CreateCaptureReviewManifest(CaptureReviewManifestInput{TaskID: task.ID, Selected: make([]string, 33), Historical: true}, resolve); err == nil {
		t.Fatal("manifest size ignored")
	}
	huge := func(ref exploration.CaptureRecordRef) (*exploration.CapturedRecord, error) {
		r, e := resolve(ref)
		if e != nil {
			return nil, e
		}
		r.Record = json.RawMessage(`{"value":"` + strings.Repeat("x", CaptureReviewExportMaxBytes) + `"}`)
		sum := sha256.Sum256(r.Record)
		r.RecordDigest = hex.EncodeToString(sum[:])
		return r, nil
	}
	d, err := s.CreateCaptureReviewDecision(input, huge)
	if err != nil {
		t.Fatal(err)
	}
	m, err := s.CreateCaptureReviewManifest(CaptureReviewManifestInput{TaskID: task.ID, Selected: []string{d.ID}, Historical: true}, huge)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.ExportCaptureReviewManifest(m.ID, huge); err == nil {
		t.Fatal("export byte limit ignored")
	}
}
