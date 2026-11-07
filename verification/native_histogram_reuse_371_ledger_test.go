package verification

import (
	"context"
	"encoding/json"
	"os"
	"slices"
	"testing"

	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type nativeReuseCounters371 struct{ Metrics, SQL, Metadata, Captures int }
type nativeReuseFact371 struct {
	Name          string
	Before, After nativeReuseCounters371
}
type nativeReuseReport371 struct {
	Version                                                                             string
	Errors                                                                              []string
	SourceQuestion, DestinationQuestion                                                 json.RawMessage
	SourceArtifactBytes, SourceArtifactAfter, StoredArtifactBefore, StoredArtifactAfter string
	SourceDisplayBefore, SourceDisplayAfter                                             json.RawMessage
	Save                                                                                struct{ Request, Response json.RawMessage }
	Reuse                                                                               struct {
		Request, Response struct {
			ID                    string
			Question, Declaration json.RawMessage
		}
		Status int
	}
	SourceRun, DestinationRun struct{ Request json.RawMessage }
	CanceledRun               struct {
		Request json.RawMessage
		Status  int
	}
	ActualView struct {
		VerifiedBins bool
		Display      struct {
			Kind             string
			RequiresComplete bool
		}
		Snapshot struct {
			Scope struct {
				Selector struct {
					Question     json.RawMessage
					FromNS, ToNS string
					Stream       json.RawMessage
				}
			}
			Histogram    struct{ Total, Unit, Temporality string }
			Rows         []struct{ Values struct{ Count string } }
			Completeness struct{ Aspects []string }
		}
	}
	Fixture struct {
		Target struct {
			Question, Stream    json.RawMessage
			FromNS, ToNS, Count string
			Counts              []string
		}
	}
	Facts []nativeReuseFact371
}

// Stream resource schema URL has the same empty default as native question
// identity; normalizing this wire omission does not discard resource or labels.
func nativeReuseJSON371(raw json.RawMessage) string {
	var v any
	if json.Unmarshal(raw, &v) != nil {
		return "invalid"
	}
	var visit func(any)
	visit = func(value any) {
		switch x := value.(type) {
		case map[string]any:
			if _, hasResource := x["resource"]; hasResource && x["name"] != nil && x["unit"] != nil {
				if _, ok := x["resourceSchemaURL"]; !ok {
					x["resourceSchemaURL"] = ""
				}
			}
			for _, child := range x {
				visit(child)
			}
		case []any:
			for _, child := range x {
				visit(child)
			}
		}
	}
	visit(v)
	return savedJSON360(v)
}

// Independent domain admission checks immutable configuration and destination
// provenance. The existing Quint model below checks acquisition budgets only.
func nativeReuseAdmission371(r nativeReuseReport371) bool {
	same := func(a, b json.RawMessage) bool {
		return len(a) > 0 && len(b) > 0 && nativeReuseJSON371(a) == nativeReuseJSON371(b)
	}
	if r.Version != "goo.nativeHistogramReuse371/1" || len(r.Errors) != 0 || len(r.Facts) == 0 || r.StoredArtifactBefore == "" || r.StoredArtifactBefore != r.StoredArtifactAfter || r.SourceArtifactBytes == "" || !same(json.RawMessage(r.SourceArtifactBytes), json.RawMessage(r.SourceArtifactAfter)) || !same(r.SourceDisplayBefore, r.SourceDisplayAfter) {
		return false
	}
	q := r.DestinationQuestion
	if !same(r.SourceRun.Request, r.SourceQuestion) || !same(r.DestinationRun.Request, q) || !same(q, r.Fixture.Target.Question) || !same(r.Reuse.Request.Question, q) || !same(r.Reuse.Response.Question, q) || r.Reuse.Response.ID != "" || r.Reuse.Status != 200 || r.CanceledRun.Status != 200 || !same(r.CanceledRun.Request, q) {
		return false
	}
	var saved struct {
		Question, Declaration json.RawMessage
		ID                    string
	}
	if json.Unmarshal(r.Save.Response, &saved) != nil || saved.ID == "" || !same(saved.Question, r.SourceQuestion) || !same(saved.Declaration, r.Reuse.Response.Declaration) {
		return false
	}
	v := r.ActualView
	s := v.Snapshot
	target := r.Fixture.Target
	if !v.VerifiedBins || v.Display.Kind != "table" || !v.Display.RequiresComplete || !same(s.Scope.Selector.Question, q) || !same(s.Scope.Selector.Stream, target.Stream) || s.Scope.Selector.FromNS != target.FromNS || s.Scope.Selector.ToNS != target.ToNS || s.Histogram.Total != target.Count || s.Histogram.Unit != "s" || s.Histogram.Temporality != "delta" || !slices.Contains(s.Completeness.Aspects, "bins") || len(s.Rows) != len(target.Counts) {
		return false
	}
	for i, row := range s.Rows {
		if row.Values.Count != target.Counts[i] {
			return false
		}
	}
	expected := []string{"save-configuration", "load-inspect", "apply-stages-current-question", "explicit-run", "explicit-destination-export-display", "unknown-is-unavailable", "late-load-does-not-adopt", "hidden-old-apply-is-inert", "cancel-late-success-does-not-publish"}
	if len(r.Facts) != len(expected) {
		return false
	}
	for i, f := range r.Facts {
		if f.Name != expected[i] || f.After.Captures != f.Before.Captures {
			return false
		}
		// Unmarked setup performs metadata only: acquisition counters must remain
		// contiguous, while metadata counters may advance between boundaries.
		if i > 0 {
			p := r.Facts[i-1].After
			if f.Before.Metrics != p.Metrics || f.Before.SQL != p.SQL || f.Before.Metadata < p.Metadata || f.Before.Captures < p.Captures {
				return false
			}
		}
		if f.Name != "explicit-run" && f.Name != "cancel-late-success-does-not-publish" && f.After.SQL != f.Before.SQL {
			return false
		}
	}
	return true
}

func nativeReuseReplay371(t *testing.T, label string, facts []nativeReuseFact371) bool {
	t.Helper()
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	if len(facts) == 0 {
		t.Fatal("missing actual native reuse facts")
	}
	baseline := facts[0].Before.Metrics
	for _, f := range facts {
		action := "history"
		if f.Name == "explicit-run" || f.Name == "cancel-late-success-does-not-publish" {
			action = "read"
		}
		rec.Record(context.Background(), action, nil, "obs.reads", f.After.Metrics-baseline, "obs.depth", 0, "obs.coherent", true)
	}
	steps, e := json.Marshal(sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	if e = artifact("native-reuse371-"+label+"-steps.json", steps); e != nil {
		t.Fatal(e)
	}
	checker, e := validate.NewChecker("model/population_drilldown.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	report, e := checker.Check(t.Context(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	if e = artifact("native-reuse371-"+label+"-model.txt", []byte(report.Summary())); e != nil {
		t.Fatal(e)
	}
	return report.OK()
}

func TestQuintNativeHistogramReuseObserved371(t *testing.T) {
	path := os.Getenv("GOO_NATIVE_HISTOGRAM_REUSE_371_REPORT")
	if path == "" {
		t.Skip("actual native histogram reuse report opt-in")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var r nativeReuseReport371
	if json.Unmarshal(raw, &r) != nil || !nativeReuseAdmission371(r) {
		t.Fatal("native reuse identity/provenance/order admission")
	}
	if !nativeReuseReplay371(t, "actual", r.Facts) {
		t.Fatal("actual native reuse acquisition budget")
	}
	read := slices.Clone(r.Facts)
	read[2].After.Metrics++
	if nativeReuseReplay371(t, "bad-apply-read", read) {
		t.Fatal("read-on-Apply mutant survived")
	}
	order := r
	order.Facts = slices.Clone(r.Facts)
	order.Facts[2], order.Facts[3] = order.Facts[3], order.Facts[2]
	if nativeReuseAdmission371(order) {
		t.Fatal("Run-before-Apply mutant survived")
	}
	wrong := r
	wrong.DestinationRun.Request = r.SourceQuestion
	if nativeReuseAdmission371(wrong) {
		t.Fatal("donor-question mutant survived")
	}
}
