package verification

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"slices"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type reviewCounters365 struct{ Calls, Source, StoreReads, StoreWrites, SQL int }
type reviewFact365 struct {
	Name          string
	Before, After reviewCounters365
	Task          *workbench.CaptureReviewTask
	Decision      *workbench.CaptureReviewDecision
	Manifest      *workbench.CaptureReviewManifest
	Input         json.RawMessage
}

// Browser report objects have been parsed and pretty-serialized. Compare their
// JSON meaning without interpreting numbers as float64; exact bytes/digests come
// only from the stored capture string and downloaded export string.
func reviewTransportEqual365(a, b json.RawMessage) bool {
	canonical := func(raw json.RawMessage) ([]byte, error) {
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		var value any
		if err := decoder.Decode(&value); err != nil {
			return nil, err
		}
		return json.Marshal(value)
	}
	left, e := canonical(a)
	right, f := canonical(b)
	return e == nil && f == nil && bytes.Equal(left, right)
}
func reviewEligible365(d *workbench.CaptureReviewDecision) bool {
	return d != nil && d.Judgment.Kind == "label" && d.Review == "reviewed" && d.State == "active"
}
func reviewModel365(t *testing.T, label string, facts []reviewFact365) bool {
	t.Helper()
	if len(facts) == 0 {
		t.Fatal("missing actual review facts")
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	base := facts[0].Before.Source
	hasTask, eligible, pinned := false, false, false
	for _, f := range facts {
		action := "history"
		switch f.Name {
		case "task-created":
			action = "task_ready"
			hasTask = true
		case "label-unreviewed", "label-reviewed", "correct-unreviewed", "disputed", "withdrawn", "resolved-unreviewed", "resolved-reviewed":
			if f.Decision == nil {
				t.Fatal("authored fact lacks actual decision", f.Name)
			}
			d := f.Decision
			eligible = reviewEligible365(d)
			switch {
			case d.State == "withdrawn":
				action = "withdrawn"
			case d.Review == "disputed":
				action = "disputed"
			case d.Judgment.Kind == "uncertain":
				action = "uncertain"
			case d.Judgment.Kind == "abstain":
				action = "abstain"
			case eligible:
				action = "label_reviewed"
			default:
				action = "label_unreviewed"
			}
		case "manifest-pinned", "resolution-pinned":
			action = "pin"
			pinned = true
		case "export-pinned", "export-original":
			action = "export_pinned"
		case "open-target", "fork-loaded", "reopen-original":
		default:
			t.Fatal("unknown actual review fact", f.Name)
		}
		rec.Record(context.Background(), action, nil, "obs.source", f.After.Source-base, "obs.task", hasTask, "obs.eligible", eligible, "obs.pinned", pinned)
	}
	raw, err := json.Marshal(sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	if err = artifact("capture-review365-"+label+"-steps.json", raw); err != nil {
		t.Fatal(err)
	}
	checker, err := validate.NewChecker("model/capture_review_365.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	report, err := checker.Check(t.Context(), sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	if err = artifact("capture-review365-"+label+"-model.txt", []byte(report.Summary())); err != nil {
		t.Fatal(err)
	}
	return report.OK()
}
func TestQuintCaptureReviewProtocol365(t *testing.T) {
	d := func(review string) *workbench.CaptureReviewDecision {
		return &workbench.CaptureReviewDecision{Judgment: workbench.CaptureReviewJudgment{Kind: "label", Value: "positive"}, Review: review, State: "active"}
	}
	facts := []reviewFact365{{Name: "open-target"}, {Name: "task-created"}, {Name: "label-unreviewed", Decision: d("unreviewed")}, {Name: "label-reviewed", Decision: d("reviewed")}, {Name: "manifest-pinned"}, {Name: "export-pinned"}, {Name: "withdrawn", Decision: &workbench.CaptureReviewDecision{Judgment: workbench.CaptureReviewJudgment{Kind: "label", Value: "positive"}, Review: "reviewed", State: "withdrawn"}}, {Name: "export-original"}}
	if !reviewModel365(t, "protocol", facts) {
		t.Fatal("review protocol smoke")
	}
	bad := slices.Clone(facts)
	bad[2].After.Source = 1
	if reviewModel365(t, "protocol-source-mutant", bad) {
		t.Fatal("implicit telemetry accepted")
	}
	bad = slices.Clone(facts)
	bad[1], bad[4] = bad[4], bad[1]
	if reviewModel365(t, "protocol-order-mutant", bad) {
		t.Fatal("pin without reviewed task accepted")
	}
}
func TestQuintCaptureReviewObserved365(t *testing.T) {
	path := os.Getenv("GOO_CAPTURE_REVIEW_365_REPORT")
	if path == "" {
		t.Skip("actual review journey report opt-in")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var r struct {
		Version                             string
		Capture, Page100                    exploration.CapturePage
		StoredBytesBefore, StoredBytesAfter string
		Resolved                            struct{ Zero, Selected exploration.CapturedRecord }
		Task                                workbench.CaptureReviewTask
		Decisions                           map[string]workbench.CaptureReviewDecision
		Manifests                           struct {
			Original, Resolution workbench.CaptureReviewManifest
		}
		Exports      struct{ OriginalBefore, OriginalAfter, Resolution string }
		RetryReceipt struct {
			Request, RetryRequest json.RawMessage
			Committed, Response   workbench.CaptureReviewDecision
			Status                int
		}
		Facts []reviewFact365
	}
	if err = json.Unmarshal(raw, &r); err != nil {
		t.Fatal(err)
	}
	if r.Version != "goo.captureReview365/1" || r.StoredBytesBefore == "" || r.StoredBytesBefore != r.StoredBytesAfter {
		t.Fatal("immutable capture custody changed")
	}
	var stored exploration.SavedCapture
	if json.Unmarshal([]byte(r.StoredBytesBefore), &stored) != nil || stored.ID != r.Capture.Receipt.ID || savedJSON360(stored.CaptureReceipt) != savedJSON360(r.Capture.Receipt) {
		t.Fatal("stored capture receipt binding")
	}
	id := stored.ID
	stored.ID = ""
	body, err := json.Marshal(stored)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(body)
	if hex.EncodeToString(digest[:]) != id {
		t.Fatal("stored complete capture hash")
	}
	a, b := r.Resolved.Zero, r.Resolved.Selected
	if a.Reference.Index != 0 || b.Reference.Index != 100 || a.Reference.CaptureID != r.Capture.Receipt.ID || b.Reference.CaptureID != r.Capture.Receipt.ID || !reviewTransportEqual365(a.Record, b.Record) || a.RecordDigest != b.RecordDigest || a.Reference == b.Reference {
		t.Fatal("duplicate values lost absolute occurrence identity")
	}
	if stored.Explore == nil || len(stored.Explore.Rows) <= 100 {
		t.Fatal("missing stored absolute row")
	}
	exact := stored.Explore.Rows[100]
	if !bytes.Equal(stored.Explore.Rows[0], exact) || !reviewTransportEqual365(b.Record, exact) || !reviewTransportEqual365(a.Record, stored.Explore.Rows[0]) {
		t.Fatal("transported record differs from exact stored occurrence")
	}
	sum := sha256.Sum256(exact)
	if b.RecordDigest != hex.EncodeToString(sum[:]) || r.Page100.Explore == nil || r.Page100.Explore.Question.Offset != 100 || !reviewTransportEqual365(r.Page100.Explore.Rows[0], exact) {
		t.Fatal("server raw occurrence/page binding")
	}
	if r.Exports.OriginalBefore == "" || r.Exports.OriginalBefore != r.Exports.OriginalAfter {
		t.Fatal("later revisions changed pinned export")
	}
	checkExport := func(manifest workbench.CaptureReviewManifest, text string) {
		t.Helper()
		if manifest.TaskID != r.Task.ID || manifest.Mode != "pinned-review-branches" || !manifest.Historical || len(manifest.Items) != 1 || len(manifest.Selected) != 1 {
			t.Fatal("manifest task/mode/membership")
		}
		var line struct {
			TaskID, DecisionID, RecordDigest, Label string
			Target                                  exploration.CaptureRecordRef
			Record                                  json.RawMessage
		}
		if json.Unmarshal(bytes.TrimSpace([]byte(text)), &line) != nil {
			t.Fatal("invalid actual JSONL export")
		}
		var decision *workbench.CaptureReviewDecision
		for _, d := range r.Decisions {
			if d.ID == line.DecisionID {
				copy := d
				decision = &copy
			}
		}
		if manifest.Items[0].DecisionID != line.DecisionID || manifest.Items[0].Label != line.Label || manifest.Items[0].Evidence.Reference != b.Reference || manifest.Items[0].Evidence.RecordDigest != b.RecordDigest {
			t.Fatal("manifest item differs from exported pinned decision/evidence")
		}
		if !reviewEligible365(decision) || line.TaskID != r.Task.ID || line.DecisionID != manifest.Selected[0] || line.Label != decision.Judgment.Value || line.Target != b.Reference || line.RecordDigest != b.RecordDigest || !bytes.Equal(line.Record, exact) {
			t.Fatal("export task/decision/target/raw/eligibility crossjoin")
		}
	}
	if r.RetryReceipt.Status != 201 || savedJSON360(r.RetryReceipt.Request) != savedJSON360(r.RetryReceipt.RetryRequest) || savedJSON360(r.RetryReceipt.Committed) != savedJSON360(r.RetryReceipt.Response) || r.RetryReceipt.Response.ID != r.Decisions["unreviewed"].ID {
		t.Fatal("lost acknowledgement retry changed authored intent/clock/artifact")
	}
	var retriedInput struct{ RequestID string }
	if json.Unmarshal(r.RetryReceipt.Request, &retriedInput) != nil || !serviceContinuationHash352(retriedInput.RequestID) {
		t.Fatal("missing actual idempotent request identity")
	}
	checkExport(r.Manifests.Original, r.Exports.OriginalBefore)
	checkExport(r.Manifests.Resolution, r.Exports.Resolution)
	names := []string{"open-target", "task-created", "label-unreviewed", "label-reviewed", "manifest-pinned", "export-pinned", "correct-unreviewed", "disputed", "withdrawn", "fork-loaded", "resolved-unreviewed", "resolved-reviewed", "resolution-pinned", "reopen-original", "export-original"}
	if len(r.Facts) != len(names) {
		t.Fatal("incomplete real review lifecycle")
	}
	for i, f := range r.Facts {
		if f.Name != names[i] || f.After.Source != f.Before.Source || f.After.SQL != f.Before.SQL {
			t.Fatal("review metadata gesture read telemetry or missing fact", f.Name)
		}
		if f.Decision != nil && (f.Decision.TaskID != r.Task.ID || f.Decision.Target != b.Reference || f.Decision.Evidence.RecordDigest != b.RecordDigest) {
			t.Fatal("authored decision escaped exact task/occurrence")
		}
	}
	if !reviewModel365(t, "observed", r.Facts) {
		t.Fatal("actual review fact projection rejected")
	}
	bad := slices.Clone(r.Facts)
	bad[2].After.Source++
	if reviewModel365(t, "observed-source-mutant", bad) {
		t.Fatal("model accepted actual counter mutation")
	}
	bad = slices.Clone(r.Facts)
	bad[1], bad[4] = bad[4], bad[1]
	if reviewModel365(t, "observed-order-mutant", bad) {
		t.Fatal("model accepted pin before task")
	}
}
