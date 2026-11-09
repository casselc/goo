package verification

import (
	"context"
	"encoding/json"
	"net/url"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type nativeLibraryCounters373 struct{ Source, Metadata, SQL, Captures int }
type nativeLibraryFact373 struct {
	Name          string
	Before, After nativeLibraryCounters373
}
type nativeLibraryPage373 struct {
	URL      string
	Response workbench.NativeViewList
}
type nativeLibraryReport373 struct {
	base                                      nativeReuseReport371
	Facts                                     []nativeLibraryFact373
	SourceArtifactBefore, SourceArtifactAfter string
	RawStoreCandidateIDs                      []string
	CorruptArtifact                           struct{ ID, Bytes string }
	Seeds                                     []struct {
		Request  workbench.NativeViewInput
		Response workbench.NativeView
	}
	CatalogOracle         []nativeLibraryPage373
	FirstPage, SecondPage nativeLibraryPage373
	WrongUnitID           string
	Get, WrongUnit        struct {
		Response            workbench.NativeView
		Status              int
		ApplyDisabled       bool
		StatusText, RowText string
	}
	WrongUnitReuse struct {
		Request struct {
			ID       string
			Question json.RawMessage
		}
		Status   int
		Response json.RawMessage
	}
	SourceOracle, DestinationOracle []struct {
		FromNS, ToNS, Count string
		Counts              []string
	}
	HeldReplies []struct {
		URL      string
		Status   int
		Response json.RawMessage
	}
	StaleList struct {
		ReplacementQuestion json.RawMessage
		CurrentHash         string
		RowsAfter           []string
		ApplyDisabled       bool
	}
	HiddenPage   struct{ BeforeIDs, AfterIDs []string }
	StaleInspect struct {
		CurrentID, InspectText string
		ApplyDisabled          bool
	}
}

// Decode the actual captured browser hash independently of the UI builder.
func nativeLibraryCurrentQuestion373(hash string) json.RawMessage {
	if !strings.HasPrefix(hash, "#step=") || len(hash) > 48000 {
		return nil
	}
	raw, e := url.PathUnescape(strings.TrimPrefix(hash, "#step="))
	if e != nil {
		return nil
	}
	var step struct {
		Version     int
		Kind        string
		Measurement measurements.Definition
		Scope       struct {
			FromMS, ToMS int64
			FromNS, ToNS string
		}
	}
	if json.Unmarshal([]byte(raw), &step) != nil || step.Version != 1 || step.Kind != "measurement" {
		return nil
	}
	q := measurements.Question{Definition: step.Measurement, FromMS: step.Scope.FromMS, ToMS: step.Scope.ToMS, FromNS: step.Scope.FromNS, ToNS: step.Scope.ToNS}
	encoded, e := json.Marshal(q)
	if e != nil {
		return nil
	}
	return encoded
}

// Only this observed run's bounded pages are checked here. This is not a
// reusable catalog validator, and Quint below claims acquisition budgets only.
func nativeLibraryAdmission373(r nativeLibraryReport373) bool {
	b := r.base
	same := func(a, b json.RawMessage) bool {
		return len(a) > 0 && len(b) > 0 && nativeReuseJSON371(a) == nativeReuseJSON371(b)
	}
	if b.Version != "goo.nativeViewLibrary373/1" || len(b.Errors) != 0 || r.SourceArtifactBefore == "" || r.SourceArtifactBefore != r.SourceArtifactAfter {
		return false
	}
	var source workbench.NativeView
	if json.Unmarshal([]byte(r.SourceArtifactBefore), &source) != nil || source.ID == "" || source.Validate() != nil || r.Get.Status != 200 || r.Get.Response.ID != source.ID || savedJSON360(r.Get.Response) != savedJSON360(source) {
		return false
	}
	var saved workbench.NativeView
	if json.Unmarshal(b.Save.Response, &saved) != nil || savedJSON360(saved) != savedJSON360(source) {
		return false
	}
	q := b.DestinationQuestion
	if !same(b.SourceRun.Request, b.SourceQuestion) || !same(b.DestinationRun.Request, q) || !same(b.Reuse.Request.Question, q) || !same(b.Reuse.Response.Question, q) || !same(q, b.Fixture.Target.Question) || b.Reuse.Request.ID != source.ID || b.Reuse.Response.ID != "" || b.Reuse.Status != 200 || savedJSON360(b.Reuse.Response.Declaration) != savedJSON360(source.Declaration) {
		return false
	}
	expected := map[string]workbench.NativeViewSummary{}
	summary := func(v workbench.NativeView) workbench.NativeViewSummary {
		return workbench.NativeViewSummary{ID: v.ID, Title: v.Title, Family: v.Question.Definition.Kind, Selection: v.Declaration.Selection.Kind, Kind: v.Declaration.Kind.Name, Unit: v.Question.Definition.Source.Unit}
	}
	expected[source.ID] = summary(source)
	for _, s := range r.Seeds {
		if s.Response.Validate() != nil || s.Request.Title != s.Response.Title || s.Request.Version != s.Response.Version || savedJSON360(s.Request.Question) != savedJSON360(s.Response.Question) || savedJSON360(s.Request.Declaration) != savedJSON360(s.Response.Declaration) {
			return false
		}
		if _, exists := expected[s.Response.ID]; exists {
			return false
		}
		expected[s.Response.ID] = summary(s.Response)
	}
	candidates := slices.Clone(r.RawStoreCandidateIDs)
	if len(candidates) != 28 || !slices.IsSorted(candidates) || r.CorruptArtifact.ID != strings.Repeat("0", 64) || r.CorruptArtifact.Bytes == "" || len(expected) != 27 {
		return false
	}
	all := make([]string, 0, len(expected)+1)
	for id := range expected {
		all = append(all, id)
	}
	all = append(all, r.CorruptArtifact.ID)
	slices.Sort(all)
	if !slices.Equal(candidates, all) {
		return false
	}
	if len(r.CatalogOracle) != 2 {
		return false
	}
	offset := 0
	after := ""
	for _, page := range r.CatalogOracle {
		u, e := url.Parse(page.URL)
		if e != nil || u.Path != "/api/native-views" || u.Query().Get("after") != after {
			return false
		}
		end := min(offset+20, len(candidates))
		items := []workbench.NativeViewSummary{}
		diagnostics := []workbench.NativeViewListDiagnostic{}
		for _, id := range candidates[offset:end] {
			if id == r.CorruptArtifact.ID {
				diagnostics = append(diagnostics, workbench.NativeViewListDiagnostic{ID: id, Reason: "unavailable-artifact"})
			} else {
				items = append(items, expected[id])
			}
		}
		cursor := ""
		if end < len(candidates) {
			cursor = candidates[end-1]
		}
		want := workbench.NativeViewList{Items: items, Partial: len(diagnostics) > 0, Diagnostics: diagnostics, NextCursor: cursor}
		if savedJSON360(page.Response) != savedJSON360(want) {
			return false
		}
		offset = end
		after = cursor
	}
	if savedJSON360(r.FirstPage.Response) != savedJSON360(r.CatalogOracle[0].Response) || savedJSON360(r.SecondPage.Response) != savedJSON360(r.CatalogOracle[1].Response) {
		return false
	}
	if r.WrongUnit.Response.Validate() != nil || expected[r.WrongUnitID] != summary(r.WrongUnit.Response) || r.WrongUnit.Status != 200 || r.WrongUnit.Response.ID != r.WrongUnitID || r.WrongUnit.Response.Question.Definition.Source.Unit != "ms" || !r.WrongUnit.ApplyDisabled || !strings.Contains(r.WrongUnit.StatusText, "same exact unit") || !strings.Contains(r.WrongUnit.RowText, "no conversion") || r.WrongUnitReuse.Status != 400 || r.WrongUnitReuse.Request.ID != r.WrongUnitID || !same(r.WrongUnitReuse.Request.Question, q) {
		return false
	}
	v := b.ActualView
	s := v.Snapshot
	target := b.Fixture.Target
	if !v.VerifiedBins || v.Display.Kind != "table" || !v.Display.RequiresComplete || !same(s.Scope.Selector.Question, q) || !same(s.Scope.Selector.Stream, target.Stream) || s.Scope.Selector.FromNS != target.FromNS || s.Scope.Selector.ToNS != target.ToNS || s.Histogram.Total != "7" || s.Histogram.Unit != "s" || s.Histogram.Temporality != "delta" || !slices.Contains(s.Completeness.Aspects, "bins") || len(s.Rows) != len(target.Counts) || len(r.SourceOracle) != 1 || r.SourceOracle[0].Count != "4" || len(r.DestinationOracle) != 2 || r.DestinationOracle[0].Count != "7" || !slices.Equal(r.DestinationOracle[0].Counts, target.Counts) {
		return false
	}
	for i, row := range s.Rows {
		if row.Values.Count != target.Counts[i] {
			return false
		}
	}
	if len(r.HeldReplies) != 3 {
		return false
	}
	for _, held := range r.HeldReplies {
		if held.Status != 200 || len(held.Response) == 0 {
			return false
		}
	}
	currentQuestion := nativeLibraryCurrentQuestion373(r.StaleList.CurrentHash)
	if !same(currentQuestion, r.StaleList.ReplacementQuestion) || same(currentQuestion, q) || len(r.StaleList.RowsAfter) != 0 || !r.StaleList.ApplyDisabled || len(r.HiddenPage.BeforeIDs) == 0 || !slices.Equal(r.HiddenPage.BeforeIDs, r.HiddenPage.AfterIDs) || r.StaleInspect.CurrentID != strings.Repeat("f", 64) || !r.StaleInspect.ApplyDisabled || r.StaleInspect.InspectText != "" {
		return false
	}
	names := []string{"save-configuration", "open-disclosure-is-passive", "browse-bounded-metadata", "loaded-page-find-and-retired-row", "more-bounded-metadata", "incompatible-metadata-visible", "discover-find-and-inspect", "apply-current-question-stages", "explicit-destination-run", "select-actual-destination-export", "direct-id-remains-available", "late-list-after-current-plan-replacement", "hidden-late-page-does-not-advance", "late-inspect-after-id-replacement"}
	if len(r.Facts) != len(names) {
		return false
	}
	for i, f := range r.Facts {
		if f.Name != names[i] || f.Before.Captures != f.After.Captures {
			return false
		}
		if i > 0 {
			p := r.Facts[i-1].After
			if f.Before.Source != p.Source || f.Before.SQL != p.SQL || f.Before.Metadata < p.Metadata {
				return false
			}
		}
		if f.Name != "explicit-destination-run" && f.After.SQL != f.Before.SQL {
			return false
		}
		expectedMetadata := []int{1, 0, 1, 0, 1, 1, 1, 1, 0, 0, 1, 1, 1, 1}
		if f.After.Metadata-f.Before.Metadata != expectedMetadata[i] {
			return false
		}
	}
	return true
}

func nativeLibraryReplay373(t *testing.T, label string, facts []nativeLibraryFact373) bool {
	t.Helper()
	if len(facts) == 0 {
		t.Fatal("missing actual native library facts")
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	baseline := facts[0].Before.Source
	for _, f := range facts {
		action := "history"
		if f.Name == "explicit-destination-run" {
			action = "read"
		}
		rec.Record(context.Background(), action, nil, "obs.reads", f.After.Source-baseline, "obs.depth", 0, "obs.coherent", true)
	}
	raw, e := json.Marshal(sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	if e = artifact("native-library373-"+label+"-steps.json", raw); e != nil {
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
	if e = artifact("native-library373-"+label+"-model.txt", []byte(report.Summary())); e != nil {
		t.Fatal(e)
	}
	return report.OK()
}
func TestQuintNativeViewLibraryObserved373(t *testing.T) {
	path := os.Getenv("GOO_NATIVE_VIEW_LIBRARY_373_REPORT")
	if path == "" {
		t.Skip("actual native library journey report opt-in")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var r nativeLibraryReport373
	if json.Unmarshal(raw, &r) != nil || json.Unmarshal(raw, &r.base) != nil || !nativeLibraryAdmission373(r) {
		t.Fatal("actual native library bounded discovery/provenance/ownership admission")
	}
	if !nativeLibraryReplay373(t, "actual", r.Facts) {
		t.Fatal("actual native library acquisition budget")
	}
	for _, index := range []int{2, 7} {
		facts := slices.Clone(r.Facts)
		facts[index].After.Source++
		if nativeLibraryReplay373(t, r.Facts[index].Name+"-read", facts) {
			t.Fatal("metadata action acquired telemetry", index)
		}
	}
	cursor := r
	cursor.CatalogOracle = slices.Clone(r.CatalogOracle)
	cursor.CatalogOracle[0].Response.NextCursor = strings.Repeat("f", 64)
	if nativeLibraryAdmission373(cursor) {
		t.Fatal("cursor mutant survived")
	}
	provenance := r
	provenance.base.DestinationRun.Request = r.base.SourceQuestion
	if nativeLibraryAdmission373(provenance) {
		t.Fatal("donor question mutant survived")
	}
	bins := r
	bins.base.ActualView.Snapshot.Rows = slices.Clone(r.base.ActualView.Snapshot.Rows)
	bins.base.ActualView.Snapshot.Rows[0].Values.Count = "1"
	if nativeLibraryAdmission373(bins) {
		t.Fatal("donor bin mutant survived")
	}
}
