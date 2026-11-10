package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/scopedview"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type heldVisible396 struct {
	Snapshot scopedview.Dataset
	Display  struct {
		Kind, Query, Order string
		RequiresComplete   bool
	}
	Find, Order, Kind string
	Valid             struct{ Valid bool }
}
type heldVariant396 struct {
	Visible heldVisible396
	Saved   customizeSaved394
}
type heldLedger396 struct {
	Version, Mode string
	Question      exploration.Query
	Raw           struct{ Logs, Spans []customizeRaw394 }
	Reads         []struct {
		Name     string
		Request  exploration.Query
		Response exploration.Result
		Status   int
	}
	ProxyCalls                            []literalWire393
	Facts                                 []patternFact391
	Variants                              struct{ Nested, Trail heldVariant396 }
	NestedParentBefore, NestedParentAfter heldVisible396
	TrailParentBefore, RecoveredVisible   heldVisible396
	Reopened                              struct {
		Status, ChartCount int
		Response           workbench.ExplorationView
		Composition        workbench.ViewComposition
	}
	Invalid struct {
		Draft string
		Bytes int
		heldVisible396
		SaveDisabled bool
		Saved        *customizeSaved394
	}
	RecoveredSaved    customizeSaved394
	FailedReplacement struct {
		Prior                 heldVisible396
		Zero                  scopedview.Row
		Pending, AfterFailure struct {
			Before, After struct {
				Counters patternBudget391
				Draft    json.RawMessage
			}
			Callbacks struct {
				Invoked                                      bool
				FindConnected, StageConnected, SaveConnected bool
			}
		}
		Failure struct {
			Request  exploration.Query
			Response struct{ Refusal struct{ Kind string } }
			Status   int
		}
		UnavailableText string
	}
	Outcome struct{ NestedMatches, TrailMatches, InvalidSaveBlocked, InvalidPublished, RecoveredMatches bool }
	Errors  []string
}

func heldTimeline396(view workbench.ExplorationView) *workbench.TimelineViewDeclaration {
	if view.Composition != nil {
		for _, child := range view.Composition.Children {
			if child.Timeline != nil {
				return child.Timeline
			}
		}
	}
	return nil
}
func heldMatches396(visible heldVisible396, saved customizeSaved394) bool {
	d := heldTimeline396(saved.Response)
	if d == nil {
		return false
	}
	order := d.Display.Order
	if order == "" {
		order = "source"
	}
	visibleOrder := visible.Display.Order
	if visibleOrder == "" {
		visibleOrder = "source"
	}
	return visible.Valid.Valid && visible.Kind == visible.Display.Kind && visible.Find == visible.Display.Query && visible.Order == visibleOrder && visible.Kind == d.Kind.Name && visible.Find == d.Display.Query && visible.Order == order && visible.Display.RequiresComplete == d.Display.RequiresComplete
}
func admitHeld396(r heldLedger396, final bool) error {
	fail := func(s string) error { return fmt.Errorf("held396: %s", s) }
	if r.Version != "goo.heldDraft396/1" || len(r.Errors) != 0 || len(r.Reads) != 3 || r.Question.Signal != "logs" {
		return fail("actual finite read flow missing")
	}
	for _, read := range r.Reads {
		if !reflect.DeepEqual(read.Request, r.Question) {
			return fail("held transition changed complete source question")
		}
		if e := admitGroupedRead395(customizeRead394{Request: read.Request, Response: read.Response, Status: read.Status}, r.Raw.Logs); e != nil {
			return e
		}
	}
	plan, err := exploration.Compile(r.Question)
	if err != nil {
		return err
	}
	planned := map[string]int{}
	for _, sql := range []string{plan.Count, plan.Rows, plan.Groups, plan.Fields, plan.TimeSeries, plan.GroupedTimeSeries} {
		planned[sql] += 3
	}
	want, failures := 18, 0
	if final {
		want++
	}
	if len(r.ProxyCalls) != want {
		return fail("physical source budget changed")
	}
	for _, call := range r.ProxyCalls {
		if call.Parameters["readonly"] != "1" || call.Parameters["max_result_rows"] != "1002" || call.Parameters["max_execution_time"] != "10" {
			return fail("physical source settings changed")
		}
		for key, value := range plan.Parameters {
			if call.Parameters["param_"+key] != value {
				return fail("physical source typed parameters changed")
			}
		}
		if call.Status == 503 && final {
			if call.SQL != plan.Count {
				return fail("failed source did not stop at planned count")
			}
			failures++
			continue
		}
		if call.Status != 200 || planned[call.SQL] == 0 {
			return fail("unplanned physical source query")
		}
		planned[call.SQL]--
	}
	if final && failures != 1 {
		return fail("actual failed physical read missing")
	}
	for _, remaining := range planned {
		if remaining != 0 {
			return fail("source plan incomplete")
		}
	}
	data := r.Reads[1].Response.GroupedTimeDataset
	if !sameTransport384(r.NestedParentBefore.Snapshot, data) || !sameTransport384(r.NestedParentAfter.Snapshot, data) || !sameTransport384(r.TrailParentBefore.Snapshot, data) || !sameTransport384(r.Variants.Nested.Visible.Snapshot, data) || !sameTransport384(r.Variants.Trail.Visible.Snapshot, data) {
		return fail("nested/trail return changed held source data/basis/clock")
	}
	for _, saved := range []customizeSaved394{r.Variants.Nested.Saved, r.Variants.Trail.Saved, r.RecoveredSaved} {
		if saved.Status != 201 || saved.Response.Validate() != nil || !reflect.DeepEqual(saved.Request.Queries, []exploration.Query{r.Question}) || !reflect.DeepEqual(saved.Response.Queries, []exploration.Query{r.Question}) {
			return fail("saved configuration changed source question")
		}
		copy := saved.Response
		id := copy.ID
		copy.ID = ""
		hash, e := workbench.ViewContentID(copy)
		if e != nil || hash != id {
			return fail("saved content identity invalid")
		}
	}
	if !heldMatches396(r.Variants.Nested.Visible, r.Variants.Nested.Saved) || !heldMatches396(r.RecoveredVisible, r.RecoveredSaved) {
		return fail("known-current visible definition did not publish")
	}
	if r.Reopened.Status != 200 || r.Reopened.ChartCount != 0 || !reflect.DeepEqual(r.Reopened.Response, r.Variants.Nested.Saved.Response) || !reflect.DeepEqual(&r.Reopened.Composition, r.Reopened.Response.Composition) {
		return fail("configuration reopen borrowed evidence or altered saved intent")
	}
	if !sameTransport384(r.Invalid.Snapshot, r.Reads[2].Response.GroupedTimeDataset) || !sameTransport384(r.RecoveredVisible.Snapshot, r.Reads[2].Response.GroupedTimeDataset) || r.Invalid.Valid.Valid || r.Invalid.Bytes != 258 || len([]byte(r.Invalid.Draft)) != 258 || r.Invalid.Find != r.Invalid.Draft {
		return fail("invalid visible draft/control was not actual or changed held source")
	}
	if final {
		failure := r.FailedReplacement
		zeroFound := false
		for _, row := range failure.Prior.Snapshot.Rows {
			if row.Key == failure.Zero.Key && sameTransport384(row, failure.Zero) {
				zeroFound = true
			}
		}
		if !zeroFound {
			return fail("failed replacement zero control was not an actual held cell")
		}
		if !sameTransport384(failure.Prior.Snapshot, r.Reads[2].Response.GroupedTimeDataset) || failure.Zero.State != "no_data" || failure.Zero.Values["count"] == nil || *failure.Zero.Values["count"] != "0" || failure.Failure.Status != 502 || failure.Failure.Response.Refusal.Kind != "unavailable" || !reflect.DeepEqual(failure.Failure.Request, r.Question) {
			return fail("actual failed replacement/control absent")
		}
		for _, probe := range []struct {
			Before, After struct {
				Counters patternBudget391
				Draft    json.RawMessage
			}
			Callbacks struct {
				Invoked                                      bool
				FindConnected, StageConnected, SaveConnected bool
			}
		}{failure.Pending, failure.AfterFailure} {
			if !probe.Callbacks.Invoked || probe.Before.Counters != probe.After.Counters || !sameTransport384(json.RawMessage(probe.Before.Draft), json.RawMessage(probe.After.Draft)) {
				return fail("retired failed/pending callbacks changed draft or performed IO")
			}
		}
		if r.Mode != "final" || !heldMatches396(r.Variants.Trail.Visible, r.Variants.Trail.Saved) || !r.Invalid.SaveDisabled || r.Invalid.Saved != nil {
			return fail("held/current invalid draft publication bug remains")
		}
	} else if r.Mode != "baseline" || heldMatches396(r.Variants.Trail.Visible, r.Variants.Trail.Saved) || r.Invalid.SaveDisabled || r.Invalid.Saved == nil || r.Invalid.Saved.Status != 201 {
		return fail("baseline did not reproduce the two actual failures")
	}
	return nil
}
func heldNames396() []string {
	return []string{"source-run", "customize-stage", "child-run", "nested-stage", "nested-back", "nested-local-edit", "nested-valid-save", "trail-back-forward", "trail-local-edit", "trail-valid-save", "reopen-configuration", "explicit-reopened-run", "invalid-find-draft", "correct-find", "corrected-save", "failed-replacement-start", "failed-replacement-pending-old", "failed-replacement-settle", "failed-replacement-failed-old", "complete"}
}
func heldCost396(name string) (read, store, write int) {
	switch name {
	case "source-run", "child-run", "explicit-reopened-run", "failed-replacement-start":
		read = 1
	case "reopen-configuration":
		store = 1
	case "nested-valid-save", "trail-valid-save", "corrected-save":
		write = 1
	}
	return
}
func TestQuintHeldDraftObserved396(t *testing.T) {
	path := os.Getenv("GOO_HELD_DRAFT_396_REPORT")
	if path == "" {
		t.Skip("actual fixed held-draft journey required")
	}
	wire, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var r heldLedger396
	if e = json.Unmarshal(wire, &r); e != nil {
		t.Fatal(e)
	}
	baselinePath := os.Getenv("GOO_HELD_DRAFT_396_BASELINE")
	if baselinePath == "" {
		t.Fatal("actual pre-fix characterization required")
	}
	baseWire, e := os.ReadFile(baselinePath)
	if e != nil {
		t.Fatal(e)
	}
	var baseline heldLedger396
	if e = json.Unmarshal(baseWire, &baseline); e != nil {
		t.Fatal(e)
	}
	if e = admitHeld396(baseline, false); e != nil {
		t.Fatal(e)
	}
	if e = admitHeld396(r, true); e != nil {
		t.Fatal(e)
	}
	actorPath := os.Getenv("GOO_HELD_DRAFT_396_ACTOR")
	if actorPath == "" {
		t.Fatal("actual companion actor required")
	}
	actorWire, err := os.ReadFile(actorPath)
	if err != nil {
		t.Fatal(err)
	}
	if err = admitHeldActor396(actorWire, r); err != nil {
		t.Fatal(err)
	}
	if !replayHeld396(t, r, false) {
		t.Fatal("measured held/local/publication workflow violates model")
	}
	mutations := []func(*heldLedger396){func(r *heldLedger396) { heldTimeline396(r.Variants.Trail.Saved.Response).Display.Query = "stale" }, func(r *heldLedger396) { r.Invalid.SaveDisabled = false }, func(r *heldLedger396) { r.NestedParentAfter.Snapshot.CapturedAt = "2026-01-01T00:00:00Z" }, func(r *heldLedger396) { r.Question.Limit++ }, func(r *heldLedger396) { heldTimeline396(r.RecoveredSaved.Response).Kind.Name = "small-multiples" }}
	for index, mutate := range mutations {
		var bad heldLedger396
		json.Unmarshal(wire, &bad)
		mutate(&bad)
		if admitHeld396(bad, true) == nil {
			t.Fatalf("semantic mutant%d escaped", index)
		}
	}
	if replayHeld396(t, r, true) {
		t.Fatal("held-local read model control escaped")
	}
	t.Log("baseline two observed publication failures retained; final visible definitions and invalid publication guards admitted; five mutants and local-read control rejected")
}
func replayHeld396(t *testing.T, r heldLedger396, mutate bool) bool {
	t.Helper()
	names := heldNames396()
	if len(r.Facts) != len(names) {
		return false
	}
	reads, stores, writes := make([]int, len(names)), make([]int, len(names)), make([]int, len(names))
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for index, f := range r.Facts {
		if f.Name != names[index] || index > 0 && f.Before != r.Facts[index-1].After {
			return false
		}
		reads[index], stores[index], writes[index] = heldCost396(f.Name)
		if f.After.Source-f.Before.Source != reads[index] || f.After.Store-f.Before.Store != stores[index] || f.After.ConfigWrites-f.Before.ConfigWrites != writes[index] || reads[index] == 0 && f.After.SQL != f.Before.SQL {
			return false
		}
		source := f.After.Source
		if mutate && f.Name == "trail-local-edit" {
			source++
		}
		action := "step"
		if f.Name == "complete" {
			action = "finish"
		}
		rec.Record(context.Background(), action, nil, "obs.phase", index+1, "obs.reads", source, "obs.stores", f.After.Store, "obs.writes", f.After.ConfigWrites, "obs.finished", f.Name == "complete")
	}
	spec, e := filepath.Abs("model/held_draft396.qnt")
	if e != nil {
		t.Fatal(e)
	}
	list := func(v []int) string { data, _ := json.Marshal(v); return string(data) }
	binding := fmt.Sprintf("spec: %q\nmodule: held_draft396\nconstants: {READS: '%s', STORES: '%s', WRITES: '%s'}\ninvariants: [safety]\nactions:\n", spec, list(reads), list(stores), list(writes))
	for _, action := range []string{"step", "finish"} {
		binding += "  " + action + ": {expect: 'phase == {{.obs.phase}} and reads == {{.obs.reads}} and stores == {{.obs.stores}} and writes == {{.obs.writes}} and finished == {{.obs.finished}}'}\n"
	}
	dir := t.TempDir()
	file := filepath.Join(dir, "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	checker, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: dir})
	if e != nil {
		t.Fatal(e)
	}
	result, e := checker.Check(t.Context(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	label := "actual"
	if mutate {
		label = "held-read-mutant"
	}
	steps, _ := json.Marshal(sink.Steps())
	if e = artifact("held-draft396-"+label+"-steps.json", steps); e != nil {
		t.Fatal(e)
	}
	if e = artifact("held-draft396-"+label+"-model.txt", []byte(result.Summary())); e != nil {
		t.Fatal(e)
	}
	return result.OK()
}

// Admit the companion's actual HTTP payloads and physical source reads; actor
// state labels alone never establish publication or source success.
func admitHeldActor396(wire []byte, main heldLedger396) error {
	var a struct {
		Version             string
		Question            exploration.Query
		SQLBefore, SQLAfter int
		ProxyCalls          []literalWire393
		Actor               struct {
			Snapshots int
			Phases    []string
			Complete  struct {
				Phase    string
				Requests []struct {
					Path              string
					Request, Response json.RawMessage
					Status            int
				}
				Expected map[string]struct {
					Kind, Query, Order string
					RequiresComplete   bool
				}
				Chart struct {
					Snapshot scopedview.Dataset
					Display  struct {
						Kind, Query, Order string
						RequiresComplete   bool
					}
				}
				Zero scopedview.Row
			}
		}
	}
	if err := json.Unmarshal(wire, &a); err != nil {
		return err
	}
	if a.Version != "goo.heldDraft396ActorReport/1" || !reflect.DeepEqual(a.Question, main.Question) || a.Actor.Snapshots < 13 || len(a.Actor.Phases) != 13 || a.Actor.Complete.Phase != "complete" || len(a.Actor.Complete.Requests) != 5 {
		return fmt.Errorf("held actor actual completion missing")
	}
	reads, writes := 0, 0
	var held *scopedview.Dataset
	for _, request := range a.Actor.Complete.Requests {
		switch request.Path {
		case "/api/explore":
			var q exploration.Query
			var result exploration.Result
			if json.Unmarshal(request.Request, &q) != nil || json.Unmarshal(request.Response, &result) != nil || !reflect.DeepEqual(q, main.Question) {
				return fmt.Errorf("actor source question changed")
			}
			if err := admitGroupedRead395(customizeRead394{Request: q, Response: result, Status: request.Status}, main.Raw.Logs); err != nil {
				return err
			}
			reads++
			held = result.GroupedTimeDataset
		case "/api/exploration-views":
			var view, stored workbench.ExplorationView
			if json.Unmarshal(request.Request, &view) != nil || json.Unmarshal(request.Response, &stored) != nil || request.Status != 201 || stored.Validate() != nil || !reflect.DeepEqual(view.Queries, []exploration.Query{main.Question}) || !reflect.DeepEqual(stored.Queries, view.Queries) {
				return fmt.Errorf("actor configuration publication invalid")
			}
			names := []string{"nested", "trail", "corrected"}
			if writes >= len(names) {
				return fmt.Errorf("extra actor publication")
			}
			expected, ok := a.Actor.Complete.Expected[names[writes]]
			wantKind, wantFind, wantOrder, wantComplete := "heatmap", "雪", "source", false
			if writes == 0 {
				wantKind, wantFind, wantOrder, wantComplete = "table", "api", "label", true
			}
			if expected.Kind != wantKind || expected.Query != wantFind || expected.Order != wantOrder || expected.RequiresComplete != wantComplete {
				return fmt.Errorf("actor visible gesture intent changed")
			}
			d := heldTimeline396(stored)
			if !ok || d == nil || d.Kind.Name != expected.Kind || d.Display.Query != expected.Query || d.Display.Order != expected.Order || d.Display.RequiresComplete != expected.RequiresComplete {
				return fmt.Errorf("actor saved definition differs from visible intent")
			}
			writes++
		default:
			return fmt.Errorf("unexpected actor request")
		}
	}
	if reads != 2 || writes != 3 || !sameTransport384(held, a.Actor.Complete.Chart.Snapshot) {
		return fmt.Errorf("actor held data/clock changed")
	}
	plan, err := exploration.Compile(main.Question)
	if err != nil {
		return err
	}
	expected := map[string]int{}
	for _, sql := range []string{plan.Count, plan.Rows, plan.Groups, plan.Fields, plan.TimeSeries, plan.GroupedTimeSeries} {
		expected[sql] += 2
	}
	if a.SQLBefore < 0 || a.SQLAfter != len(a.ProxyCalls) || a.SQLAfter-a.SQLBefore != 12 {
		return fmt.Errorf("actor physical source budget changed")
	}
	for _, call := range a.ProxyCalls[a.SQLBefore:a.SQLAfter] {
		if expected[call.SQL] == 0 || call.Status != 200 || call.Parameters["readonly"] != "1" || call.Parameters["max_result_rows"] != "1002" || call.Parameters["max_execution_time"] != "10" {
			return fmt.Errorf("actor unplanned physical query")
		}
		expected[call.SQL]--
		for key, value := range plan.Parameters {
			if call.Parameters["param_"+key] != value {
				return fmt.Errorf("actor typed source parameters changed")
			}
		}
	}
	return nil
}
