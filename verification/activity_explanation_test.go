package verification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Activity authoring is client-owned. These generated requests exercise the
// existing Go wire/planner/SQL boundary, without recreating the UI helper.
func TestHegelActivityExplanationScopeRoundTrip(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		mode := hegel.Draw(tc, hegel.Integers(0, 2))
		signal := "spans"
		if seed%2 == 0 {
			signal = "logs"
		}
		typ := "int64"
		threshold := "9007199254740993"
		if seed%3 == 0 {
			typ = "uint64"
			threshold = "18446744073709551615"
		}
		b := predicateBinding(typ, "attribute")
		literal := fmt.Sprintf("core-'\\雪-%d", seed)
		where := &exploration.Predicate{All: []exploration.Predicate{
			{Any: []exploration.Predicate{{Field: &exploration.Field{Kind: "attribute", Key: "tenant"}, Op: "missing"}, {Field: &exploration.Field{Kind: "attribute", Key: "tenant"}, Op: "eq", Value: ""}}},
			{Numeric: &b, Op: "lte", Value: threshold},
			{Field: &exploration.Field{Kind: "resource", Key: "team"}, Op: "eq", Value: literal},
		}}
		end := int64(1800000000000 + seed*1000)
		q := exploration.Query{Signal: signal, Environment: "activity-property", FromMS: end - 60000, ToMS: end, Limit: 1, Offset: seed % 100, Where: where}
		baseline := comparisonClone(q)
		selection := comparisonClone(q)
		switch mode {
		case 0:
			baseline.FromMS = q.FromMS - 60000
			baseline.ToMS = q.FromMS
		case 1:
			selection.FromMS = q.FromMS + int64(seed%50)*1000
			selection.ToMS = selection.FromMS + 1000
		case 2:
			baseline.FromMS = q.FromMS - 120000
			baseline.ToMS = q.FromMS - 30000
			baseline.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "resource", Key: "service.version"}, Op: "eq", Value: fmt.Sprintf("old-%d", seed)}
		}
		request := exploration.ExplanationRequest{Version: "goo/explanation/v1", Baseline: baseline, Selection: selection, Fields: []exploration.Field{{Kind: "column", Key: "service"}, {Kind: "attribute", Key: "tenant"}}}
		raw, e := json.Marshal(request)
		if e != nil {
			tc.Fatal(e)
		}
		var decoded exploration.ExplanationRequest
		if e = json.Unmarshal(raw, &decoded); e != nil || !reflect.DeepEqual(request, decoded) {
			tc.Fatal("wire changed original scope", e)
		}
		plan, e := exploration.PlanExplanation(decoded)
		if e != nil {
			tc.Fatal(e)
		}
		if !reflect.DeepEqual(plan.Request, request) {
			tc.Fatal("plan mutated supplied scope")
		}
		for i, side := range []exploration.Query{plan.Baseline, plan.Selection} {
			original := []exploration.Query{baseline, selection}[i]
			if !reflect.DeepEqual(side, peLiveCountQuestion(original)) {
				tc.Fatal("count projection lost typed predicate/window")
			}
			compiled, e := exploration.Compile(side)
			if e != nil {
				tc.Fatal(e)
			}
			if compiled.Parameters["from"] != strconv.FormatInt(original.FromMS, 10) || compiled.Parameters["to"] != strconv.FormatInt(original.ToMS, 10) || compiled.Parameters["environment"] != q.Environment {
				tc.Fatal("compiled scope changed")
			}
			if strings.Contains(compiled.Count, threshold) || strings.Contains(compiled.Count, literal) {
				tc.Fatal("literal interpolated into SQL")
			}
		}
		if mode == 0 && (plan.Baseline.ToMS != plan.Selection.FromMS || plan.Baseline.ToMS-plan.Baseline.FromMS != plan.Selection.ToMS-plan.Selection.FromMS) {
			tc.Fatal("previous window not adjacent/equal width")
		}
		if mode == 1 && (plan.Selection.FromMS < q.FromMS || plan.Selection.ToMS > q.ToMS) {
			tc.Fatal("brush escaped fixed window")
		}
		if now, _ := json.Marshal(request); string(now) != string(raw) {
			tc.Fatal("request mutated")
		}
	}, options(70)...)
}

// The Activity chart is deliberately a held-row guide. Its six bins are checked
// against the actual retained wire records, independently of telemetry totals.
type aeLiveApplied struct {
	Selector   exploration.TailSelector
	ReadWindow struct {
		FromMS int64 `json:"fromMs"`
		ToMS   int64 `json:"toMs"`
	}
	HeldRows []struct{ Record map[string]json.RawMessage }
	Chart    struct {
		Title, Unit, Ceiling, FromLabel, ToLabel string
		Interactive                              bool
		Points                                   []struct{ Key, Label, Value string }
	}
}

func aeLiveHeldFacts(a aeLiveApplied, environment string, records []peLiveRecord) bool {
	facts := map[string]peLiveRecord{}
	for _, r := range records {
		facts[r.ID] = r
	}
	seen := map[string]bool{}
	for _, held := range a.HeldRows {
		raw, _ := json.Marshal(held.Record)
		var r struct {
			ID         string `json:"SpanId"`
			Body       string
			Time       string `json:"timestamp_ns"`
			Service    string `json:"ServiceName"`
			Attributes map[string]string
			Resource   map[string]string `json:"ResourceAttributes"`
		}
		if json.Unmarshal(raw, &r) != nil {
			return false
		}
		id := r.ID
		if a.Selector.Signal == "logs" {
			id = strings.TrimPrefix(r.Body, "work observed ")
		}
		fact, ok := facts[id]
		if !ok || seen[id] || fact.Signal != a.Selector.Signal || r.Time != strconv.FormatInt(fact.TimeMS*1000000, 10) || r.Service != fact.Service || !reflect.DeepEqual(r.Attributes, fact.Attributes) || r.Resource["deployment.environment.name"] != environment {
			return false
		}
		yes, e := peLiveMatch(a.Selector.Where, fact)
		if e != nil || !yes {
			return false
		}
		for key, value := range fact.Resource {
			if r.Resource[key] != value {
				return false
			}
		}
		seen[id] = true
	}
	return len(a.HeldRows) > 0
}
func aeLiveWindow(a aeLiveApplied) exploration.Query {
	return exploration.Query{Signal: a.Selector.Signal, Environment: a.Selector.Environment, Where: a.Selector.Where, FromMS: a.ReadWindow.FromMS, ToMS: a.ReadWindow.ToMS, Limit: 100}
}
func aeLiveChart(a aeLiveApplied) bool {
	if a.ReadWindow.FromMS < 60000 || a.ReadWindow.ToMS-a.ReadWindow.FromMS != 60000 || len(a.Chart.Points) != 6 || !a.Chart.Interactive || a.Chart.Unit != "held occurrences" || !strings.Contains(a.Chart.Title, "sampled occurrences, not full population counts") {
		return false
	}
	iso := func(ms int64) string { return time.UnixMilli(ms).UTC().Format("2006-01-02T15:04:05.000Z") }
	if a.Chart.FromLabel != iso(a.ReadWindow.FromMS) || a.Chart.ToLabel != iso(a.ReadWindow.ToMS) {
		return false
	}
	counts := [6]int{}
	maximum := 1
	for _, row := range a.HeldRows {
		var value string
		if json.Unmarshal(row.Record["timestamp_ns"], &value) != nil {
			continue
		}
		if value == "" || strings.IndexFunc(value, func(r rune) bool { return r < '0' || r > '9' }) >= 0 {
			continue
		}
		// Millisecond fixture timestamps remain exact: parse nanoseconds as uint64,
		// divide only after checking the complete half-open source window.
		ns, e := strconv.ParseUint(value, 10, 64)
		if e != nil {
			return false
		}
		start := uint64(a.ReadWindow.FromMS) * 1000000
		end := uint64(a.ReadWindow.ToMS) * 1000000
		if ns >= start && ns < end {
			counts[(ns-start)/10000000000]++
		}
	}
	for i, p := range a.Chart.Points {
		if p.Key != strconv.Itoa(i) || p.Value != strconv.Itoa(counts[i]) || p.Label != iso(a.ReadWindow.FromMS+int64(i)*10000)+" · "+strconv.Itoa(counts[i])+" held occurrences" {
			return false
		}
		if counts[i] > maximum {
			maximum = counts[i]
		}
	}
	return a.Chart.Ceiling == strconv.Itoa(maximum)
}

// Shared DTOs carry observed evidence only; peLiveResult independently rebuilds
// counts, exact rational shares and ranked references from emitted OTLP facts.
type aeLiveEvent struct {
	Action                                                 string
	SourceOps, Promotions, PurePlans, Streams, CaptureGets int
}
type aeLiveLedger struct {
	ScopeStored exploration.SavedCapture
	Events      []aeLiveEvent
	peLiveLedger
	Signal   string
	View     json.RawMessage
	Applied  aeLiveApplied
	Previous peLiveRun
	Brush    struct {
		FromBin, ToBin int
		Request        exploration.ExplanationRequest
		Result         exploration.ExplanationResult
	}
	ScopeCapture, ScopeCaptureAfter exploration.CapturePage
	LoadedScope                     peLiveRun
	Filter                          peLiveFilter
}

func aeLiveScopes(l aeLiveLedger) bool {
	q := aeLiveWindow(l.Applied)
	if l.Version != 1 || q.Signal != l.Signal || q.Environment != l.Environment || !aeLiveChart(l.Applied) || !aeLiveHeldFacts(l.Applied, l.Environment, l.Records) {
		return false
	}
	full := l.Previous.Request
	previous := comparisonClone(q)
	previous.FromMS = q.FromMS - 60000
	previous.ToMS = q.FromMS
	if !nhLiveSameQuery(full.Selection, q) || !nhLiveSameQuery(full.Baseline, previous) || !peLiveResult(full, l.Previous.Result, l.Environment, l.Records) {
		return false
	}
	brush := l.Brush
	if brush.FromBin < 0 || brush.ToBin > 6 || brush.FromBin >= brush.ToBin {
		return false
	}
	selected := comparisonClone(q)
	selected.FromMS = q.FromMS + int64(brush.FromBin)*10000
	selected.ToMS = q.FromMS + int64(brush.ToBin)*10000
	prior := comparisonClone(selected)
	prior.FromMS = selected.FromMS - (selected.ToMS - selected.FromMS)
	prior.ToMS = selected.FromMS
	if !nhLiveSameQuery(brush.Request.Selection, selected) || !nhLiveSameQuery(brush.Request.Baseline, prior) || !reflect.DeepEqual(brush.Request.Fields, full.Fields) || !peLiveResult(brush.Request, brush.Result, l.Environment, l.Records) {
		return false
	}
	scope := l.ScopeCapture
	if scope.Receipt.Question.Kind != "explore" || scope.Receipt.Question.Explore == nil || scope.Explore == nil || !aeLiveStoredScope(l.ScopeStored, scope, l.Environment, l.Records) || !reflect.DeepEqual(scope, l.ScopeCaptureAfter) || !nhLiveSameQuery(*scope.Receipt.Question.Explore, scope.Explore.Question) || !peLiveSource(*scope.Explore, l.Environment, l.Records) {
		return false
	}
	loaded := l.LoadedScope
	if !nhLiveSameQuery(loaded.Request.Baseline, *scope.Receipt.Question.Explore) || !nhLiveSameQuery(loaded.Request.Selection, selected) || !reflect.DeepEqual(loaded.Request.Fields, full.Fields) || !peLiveResult(loaded.Request, loaded.Result, l.Environment, l.Records) {
		return false
	}
	all := append(append([]peLiveRecord{}, l.Records...), l.LateRecords...)
	if !reflect.DeepEqual(l.Reused.Request, loaded.Request) || !peLiveResult(l.Reused.Request, l.Reused.Result, l.Environment, all) || !peLiveSource(l.ReadyProbe, l.Environment, all) {
		return false
	}
	promotions := l.peLiveLedger
	promotions.Observed = loaded
	return peLivePromotions(promotions) && reflect.DeepEqual(l.Saved, l.Reopened) && peLiveFilterScope(l.Filter, loaded.Request, l.Environment, all) && peLiveSource(l.Filter.Result, l.Environment, all) && nhLiveSameQuery(l.Filter.Result.Question, l.Filter.Response.Question)
}

func aeLiveRequests(l aeLiveLedger) bool {
	counts := map[string]int{}
	sourceIndex := 0
	for _, r := range l.Requests {
		counts[r.Effect]++
		switch r.Effect {
		case "stream":
			var s exploration.TailSelector
			if r.URL != "/api/tail" || json.Unmarshal(r.Body, &s) != nil {
				return false
			}
			if s.Signal == "" {
				s.Signal = "logs"
			}
			if !reflect.DeepEqual(s, l.Applied.Selector) {
				return false
			}
		case "source":
			if sourceIndex == 3 {
				var q exploration.Query
				if r.URL != "/api/explore" || json.Unmarshal(r.Body, &q) != nil || !nhLiveSameQuery(q, l.Filter.Response.Question) {
					return false
				}
			} else {
				var req exploration.ExplanationRequest
				expected := []exploration.ExplanationRequest{l.Previous.Request, l.Brush.Request, l.LoadedScope.Request, l.Reused.Request}
				i := sourceIndex
				if i > 3 {
					i--
				}
				if i >= len(expected) || r.URL != "/api/population/explain" || json.Unmarshal(r.Body, &req) != nil || !reflect.DeepEqual(req, expected[i]) {
					return false
				}
			}
			sourceIndex++
		case "plan":
			if r.URL != "/api/population/explain/filter" {
				return false
			}
			body := peLiveFilter{}.Request
			if json.Unmarshal(r.Body, &body) != nil || !reflect.DeepEqual(body, l.Filter.Request) {
				return false
			}
		case "promotion":
			if r.URL != "/api/captures" {
				return false
			}
			var body map[string]string
			if json.Unmarshal(r.Body, &body) != nil || len(body) != 2 || body["title"] != l.Saved.Receipt.Title || body["resultToken"] != l.LoadedScope.Result.ResultToken {
				return false
			}
		default:
			return false
		}
	}
	return counts["stream"] == 1 && counts["source"] == 5 && counts["promotion"] == 2 && counts["plan"] == 1
}
func aeLiveReplay(t *testing.T, l aeLiveLedger, coherent bool) bool {
	t.Helper()
	if len(l.Events) != 19 {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	gets := 0
	for i, event := range l.Events {
		if event.Streams != 1 && i > 0 || event.CaptureGets < gets {
			coherent = false
		}
		if i == 0 && (event.Streams != 0 || event.CaptureGets != 0) {
			coherent = false
		}
		if (i < 7 && event.CaptureGets != 0) || (i >= 7 && i < 16 && event.CaptureGets != 1) || (i >= 16 && event.CaptureGets < 2) {
			coherent = false
		}
		gets = event.CaptureGets
		action := strings.ReplaceAll(event.Action, "-", "_")
		if action == "fields" {
			action = "edit_fields"
		}
		rec.Record(context.Background(), action, nil, "obs.coherent", coherent, "obs.sources", event.SourceOps, "obs.promotions", event.Promotions, "obs.plans", event.PurePlans)
	}
	checker, e := validate.NewChecker("model/activity_explanation.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	result, e := checker.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return result.OK()
}
func TestQuintActivityExplanationLiveLedger(t *testing.T) {
	file := os.Getenv("GOO_ACTIVITY_EXPLANATION_LEDGER")
	if file == "" {
		t.Skip("requires actual Activity explanation journey ledger")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	check := func(control int) bool {
		var l aeLiveLedger
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		if e := aeReadStoredScope(file, &l.ScopeStored); e != nil {
			t.Fatal(e)
		}
		before, _ := json.Marshal(l)
		switch control {
		case 0:
			l.Previous.Request.Selection.FromMS++
		case 1:
			l.Previous.Request.Baseline.ToMS++
		case 2:
			l.Brush.Request.Selection.ToMS++
		case 3:
			l.Brush.Request.Baseline.FromMS--
		case 4:
			l.Applied.ReadWindow.ToMS++
		case 5:
			l.Applied.Chart.Points[0].Value += "1"
		case 6:
			l.Applied.Chart.Title = "Full population counts"
		case 7:
			l.LoadedScope.Request.Baseline.Where = nil
		case 8:
			l.LoadedScope.Request.Baseline.FromMS++
		case 9:
			l.ScopeCaptureAfter.Receipt.Title += "changed"
		case 10:
			l.LoadedScope.Result.Selection.Count += "1"
			l.LoadedScope.Result.Basis = peLiveResultBasis(l.LoadedScope.Result)
		case 11:
			l.LoadedScope.Result.Fields[0].Rows[0].DifferencePP += "1"
			l.LoadedScope.Result.Basis = peLiveResultBasis(l.LoadedScope.Result)
		case 12:
			l.Saved.Explanation.Selection.Count += "1"
		case 13:
			l.Reopened.Receipt.ID = "changed"
		case 14:
			l.Reused.Request.Baseline.ToMS++
		case 15:
			l.ReadyProbe.Count += "1"
		case 16:
			l.Filter.Response.Question.Where = nil
		case 17:
			l.Events[4].SourceOps++
		case 18:
			l.Events[7].SourceOps++
		case 19:
			l.Events[10].SourceOps++
		case 20:
			l.Events[15].Streams++
		case 21:
			l.Events[0], l.Events[1] = l.Events[1], l.Events[0]
		case 22:
			l.Promotions[1].Response.Receipt.ID = "changed"
		case 23:
			l.Requests[0].URL = "/api/explore"
		case 24:
			l.Applied.HeldRows = nil
		}
		if control >= 0 {
			after, _ := json.Marshal(l)
			if string(after) == string(before) {
				t.Fatalf("control %d did not change ledger", control)
			}
		}
		scopes, requests := aeLiveScopes(l), aeLiveRequests(l)
		if control < 0 {
			t.Logf("scope evidence=%v request discipline=%v stored hash=%v held-row facts=%v", scopes, requests, peLiveCaptureHash(exploration.CapturePage{Receipt: l.ScopeStored.CaptureReceipt, Explore: l.ScopeStored.Explore}), aeLiveHeldFacts(l.Applied, l.Environment, l.Records))
		}
		return aeLiveReplay(t, l, scopes && requests)
	}
	if !check(-1) {
		t.Fatal("actual Activity explanation ledger rejected")
	}
	for i := 0; i < 25; i++ {
		if check(i) {
			t.Fatalf("control %d accepted", i)
		}
		t.Logf("control %d rejected", i)
	}
}

type aeLiveWindowBounds struct {
	FromMS int64 `json:"fromMs"`
	ToMS   int64 `json:"toMs"`
}
type aeOwnershipEvent struct {
	aeLiveEvent
	PanelHidden, LoadEnabled                   *bool
	FieldCount, StagedFilterCount, EditorCount *int
	ReadWindow                                 *aeLiveWindowBounds
}
type aeOwnershipLedger struct {
	ScopeStored         exploration.SavedCapture
	Version             int
	Environment, Signal string
	Records             []peLiveRecord
	Applied             aeLiveApplied
	ScopeCapture        exploration.CapturePage
	HeldRuns            []struct {
		Reason  string
		Request exploration.ExplanationRequest
		Result  exploration.ExplanationResult
	}
	HeldLoad   struct{ Response exploration.CapturePage }
	FreshLoad  exploration.CapturePage
	Fresh      peLiveRun
	HeldFilter peLiveFilter
	Events     []aeOwnershipEvent
	Requests   []struct {
		Effect, URL string
		Body        json.RawMessage
	}
}

func aeOwnershipEvidence(l aeOwnershipLedger) bool {
	if l.Version != 1 || l.Signal != "spans" || len(l.HeldRuns) != 3 || len(l.Events) != 17 || !aeLiveChart(l.Applied) || !aeLiveHeldFacts(l.Applied, l.Environment, l.Records) || !aeLiveStoredScope(l.ScopeStored, l.ScopeCapture, l.Environment, l.Records) || !aeJSONEqual(l.ScopeCapture, l.HeldLoad.Response) || !aeJSONEqual(l.ScopeCapture, l.FreshLoad) || l.ScopeCapture.Receipt.Question.Explore == nil {
		return false
	}
	window := aeLiveWindow(l.Applied)
	fields := []exploration.Field{{Kind: "column", Key: "service"}, {Kind: "attribute", Key: "tenant"}}
	for i, held := range l.HeldRuns {
		if held.Reason != []string{"resume", "navigation", "scope"}[i] || !reflect.DeepEqual(held.Request.Fields, fields) || !nhLiveSameQuery(held.Request.Selection, window) || !peLiveResult(held.Request, held.Result, l.Environment, l.Records) {
			return false
		}
		previous := comparisonClone(window)
		previous.FromMS = window.FromMS - 60000
		previous.ToMS = window.FromMS
		if !nhLiveSameQuery(held.Request.Baseline, previous) {
			return false
		}
		reopened := l.Events[4+i*3]
		if reopened.ReadWindow == nil {
			return false
		}
		window.FromMS = reopened.ReadWindow.FromMS
		window.ToMS = reopened.ReadWindow.ToMS
	}
	if window.ToMS-window.FromMS != 60000 || !nhLiveSameQuery(l.Fresh.Request.Selection, window) || !nhLiveSameQuery(l.Fresh.Request.Baseline, *l.ScopeCapture.Receipt.Question.Explore) || !reflect.DeepEqual(l.Fresh.Request.Fields, fields) || !peLiveResult(l.Fresh.Request, l.Fresh.Result, l.Environment, l.Records) || !peLiveFilterScope(l.HeldFilter, l.Fresh.Request, l.Environment, l.Records) {
		return false
	}
	sources, streams, plans := 0, 0, 0
	for _, r := range l.Requests {
		switch r.Effect {
		case "stream":
			streams++
			var selector exploration.TailSelector
			if r.URL != "/api/tail" || json.Unmarshal(r.Body, &selector) != nil {
				return false
			}
			if selector.Signal == "" {
				selector.Signal = "logs"
			}
			if !reflect.DeepEqual(selector, l.Applied.Selector) {
				return false
			}
		case "source":
			var req exploration.ExplanationRequest
			if r.URL != "/api/population/explain" || json.Unmarshal(r.Body, &req) != nil {
				return false
			}
			expected := l.Fresh.Request
			if sources < 3 {
				expected = l.HeldRuns[sources].Request
			}
			if sources >= 4 || !reflect.DeepEqual(req, expected) {
				return false
			}
			sources++
		case "plan":
			if r.URL != "/api/population/explain/filter" {
				return false
			}
			got := peLiveFilter{}.Request
			if json.Unmarshal(r.Body, &got) != nil || !reflect.DeepEqual(got, l.HeldFilter.Request) {
				return false
			}
			plans++
		default:
			return false
		}
	}
	return sources == 4 && streams == 3 && plans == 1
}
func aeOwnershipReplay(t *testing.T, l aeOwnershipLedger, coherent bool) bool {
	t.Helper()
	if len(l.Events) != 17 {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, event := range l.Events {
		streams := 1
		if i == 0 {
			streams = 0
		} else if i >= 3 {
			streams = 2
		}
		if i >= 10 {
			streams = 3
		}
		gets := 0
		if i >= 11 {
			gets = 1
		}
		if i >= 13 {
			gets = 2
		}
		if event.Streams != streams || event.CaptureGets != gets || event.Promotions != 0 {
			coherent = false
		}
		snapshot := strings.HasSuffix(event.Action, "-released") || strings.HasSuffix(event.Action, "-reopen") || event.Action == "load-return" || event.Action == "load-fresh"
		if snapshot {
			if event.FieldCount == nil || event.StagedFilterCount == nil || event.EditorCount == nil || event.PanelHidden == nil || event.LoadEnabled == nil || event.ReadWindow == nil {
				coherent = false
			} else {
				editors := 1
				if event.Action == "load-return" || event.Action == "filter-released" || event.Action == "resume-released" || event.Action == "scope-released" {
					editors = 0
				}
				hidden := event.Action == "resume-released" || event.Action == "scope-released" || event.Action == "filter-released"
				if *event.FieldCount != 0 || *event.StagedFilterCount != 0 || *event.EditorCount != editors || *event.PanelHidden != hidden || event.ReadWindow.ToMS-event.ReadWindow.FromMS != 60000 {
					coherent = false
				}
				if (event.Action == "load-return" || event.Action == "load-fresh") && !*event.LoadEnabled {
					coherent = false
				}
			}
		}
		rec.Record(context.Background(), strings.ReplaceAll(event.Action, "-", "_"), nil, "obs.coherent", coherent, "obs.sources", event.SourceOps, "obs.plans", event.PurePlans)
	}
	checker, e := validate.NewChecker("model/activity_explanation_ownership.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	result, e := checker.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return result.OK()
}
func TestQuintActivityExplanationOwnershipLedger(t *testing.T) {
	file := os.Getenv("GOO_ACTIVITY_EXPLANATION_OWNERSHIP_LEDGER")
	if file == "" {
		t.Skip("requires actual held Activity ownership ledger")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	check := func(control int) bool {
		var l aeOwnershipLedger
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		if e := aeReadStoredScope(file, &l.ScopeStored); e != nil {
			t.Fatal(e)
		}
		before, _ := json.Marshal(l)
		switch control {
		case 0:
			l.HeldRuns[0].Request.Selection.FromMS++
		case 1:
			l.HeldRuns[1].Request.Baseline.Where = nil
		case 2:
			l.HeldRuns[2].Result.Selection.Count += "1"
			l.HeldRuns[2].Result.Basis = peLiveResultBasis(l.HeldRuns[2].Result)
		case 3:
			l.HeldLoad.Response.Receipt.ID = "changed"
		case 4:
			l.FreshLoad.Receipt.Title += "changed"
		case 5:
			l.Fresh.Request.Baseline.ToMS++
		case 6:
			l.HeldFilter.Response.Question.Where = nil
		case 7:
			l.Events[3].SourceOps++
		case 8:
			l.Events[12].SourceOps++
		case 9:
			l.Events[16].SourceOps++
		case 10:
			v := 1
			l.Events[3].FieldCount = &v
		case 11:
			v := 1
			l.Events[16].StagedFilterCount = &v
		case 12:
			v := false
			l.Events[12].LoadEnabled = &v
		case 13:
			v := 1
			l.Events[12].EditorCount = &v
		case 14:
			l.Events[4].ReadWindow.FromMS++
		case 15:
			l.Events[0], l.Events[1] = l.Events[1], l.Events[0]
		}
		if control >= 0 {
			after, _ := json.Marshal(l)
			if string(after) == string(before) {
				t.Fatalf("ownership control %d did not change evidence", control)
			}
		}
		evidence := aeOwnershipEvidence(l)
		if control < 0 {
			t.Logf("ownership source evidence=%v full stored scope=%v held public page=%v fresh result=%v", evidence, aeLiveStoredScope(l.ScopeStored, l.ScopeCapture, l.Environment, l.Records), aeJSONEqual(l.ScopeCapture, l.HeldLoad.Response), peLiveResult(l.Fresh.Request, l.Fresh.Result, l.Environment, l.Records))
		}
		return aeOwnershipReplay(t, l, evidence)
	}
	if !check(-1) {
		t.Fatal("actual Activity ownership ledger rejected")
	}
	for i := 0; i < 16; i++ {
		if check(i) {
			t.Fatalf("ownership control %d accepted", i)
		}
		t.Logf("ownership control %d rejected", i)
	}
}

func aeReadStoredScope(ledger string, dst *exploration.SavedCapture) error {
	raw, e := os.ReadFile(filepath.Join(filepath.Dir(ledger), "scope-capture.json"))
	if e != nil {
		return fmt.Errorf("actual full stored source capture is required: %w", e)
	}
	return json.Unmarshal(raw, dst)
}
func aeLiveStoredScope(stored exploration.SavedCapture, page exploration.CapturePage, environment string, records []peLiveRecord) bool {
	if stored.Explore == nil || stored.Question.Kind != "explore" || stored.Question.Explore == nil || stored.Explanation != nil || stored.Comparison != nil || stored.Inventory != nil || stored.Measurement != nil || page.Explore == nil || !reflect.DeepEqual(stored.CaptureReceipt, page.Receipt) || !reflect.DeepEqual(*stored.Question.Explore, stored.Explore.Question) {
		return false
	}
	full := exploration.CapturePage{Receipt: stored.CaptureReceipt, Explore: stored.Explore}
	if !peLiveCaptureHash(full) {
		return false
	}
	rows, e := peLiveRows(stored.Explore.Question, environment, records)
	if e != nil || len(rows) == 0 || len(rows) > 1000 || len(stored.Explore.Rows) != len(rows) || stored.Explore.HasMore || page.RetainedRecords != len(rows) || page.RecordsLimited || page.Explore.Count != stored.Explore.Count || page.Explore.Basis != stored.Explore.Basis || page.Explore.CapturedAt != stored.Explore.CapturedAt {
		return false
	}
	n := len(rows)
	if n > 100 {
		n = 100
	}
	if !aeJSONEqual(page.Explore.Rows, stored.Explore.Rows[:n]) || page.Explore.HasMore != (len(rows) > n) {
		return false
	}
	// Capture storage retains beyond the displayed query page. Adjust only this
	// independent membership check's presentation limit, after hashing the exact
	// unmodified stored representation and checking its original question above.
	retained := *stored.Explore
	retained.Question.Limit = len(rows)
	return peLiveSource(retained, environment, records)
}

type aeSavedBaselineLedger struct {
	Version                                                    int
	Environment, Signal, OriginalLedger                        string
	OriginalRecords, SelectionLateRecords, BaselineLateRecords []peLiveRecord
	ScopeCaptureBefore, ScopeCaptureAfter                      exploration.CapturePage
	ScopeRawBefore, ScopeRawAfter                              string
	ScopeStoredBefore, ScopeStoredAfter                        exploration.SavedCapture
	ScopeStoredBeforePath, ScopeStoredAfterPath                string
	ReadyProbe                                                 exploration.Result
	Rerun                                                      peLiveRun
}

func aeSavedBaselineEvidence(l aeSavedBaselineLedger, original aeLiveLedger, before, after []byte) bool {
	if l.Version != 1 || l.Environment != original.Environment || l.Signal != original.Signal || len(l.BaselineLateRecords) != 1 || !reflect.DeepEqual(l.OriginalRecords, original.Records) || !reflect.DeepEqual(l.SelectionLateRecords, original.LateRecords) || string(before) != string(after) || l.ScopeRawBefore != l.ScopeRawAfter || !reflect.DeepEqual(l.ScopeStoredBefore, l.ScopeStoredAfter) || !reflect.DeepEqual(l.ScopeCaptureBefore, l.ScopeCaptureAfter) || !reflect.DeepEqual(l.ScopeCaptureBefore, original.ScopeCapture) || !reflect.DeepEqual(l.Rerun.Request, original.LoadedScope.Request) {
		return false
	}
	var a, z exploration.SavedCapture
	var public exploration.CapturePage
	if json.Unmarshal(before, &a) != nil || json.Unmarshal(after, &z) != nil || !aeJSONEqual(a, l.ScopeStoredBefore) || !aeJSONEqual(z, l.ScopeStoredAfter) || json.Unmarshal([]byte(l.ScopeRawBefore), &public) != nil || !aeJSONEqual(public, l.ScopeCaptureBefore) || !aeLiveStoredScope(a, l.ScopeCaptureBefore, l.Environment, l.OriginalRecords) {
		return false
	}
	records := append(append(append([]peLiveRecord{}, l.OriginalRecords...), l.SelectionLateRecords...), l.BaselineLateRecords...)
	old, e := peLiveRows(l.Rerun.Request.Baseline, l.Environment, l.OriginalRecords)
	if e != nil {
		return false
	}
	selected, e := peLiveRows(l.Rerun.Request.Selection, l.Environment, records)
	if e != nil {
		return false
	}
	if l.Rerun.Result.Baseline.Count != strconv.Itoa(len(old)+1) || l.Rerun.Result.Selection.Count != strconv.Itoa(len(selected)) || l.Rerun.Result.Selection.Count != original.Reused.Result.Selection.Count || !nhLiveSameQuery(l.ReadyProbe.Question, l.Rerun.Request.Baseline) || !peLiveSource(l.ReadyProbe, l.Environment, records) || !peLiveResult(l.Rerun.Request, l.Rerun.Result, l.Environment, records) {
		return false
	}
	read, e := time.Parse(time.RFC3339Nano, l.Rerun.Result.ReadAt)
	if e != nil {
		return false
	}
	probe, e := time.Parse(time.RFC3339Nano, l.ReadyProbe.CapturedAt)
	return e == nil && read.After(probe)
}

// This supplemental record is an actual API observation, not a browser action
// trace. Arithmetic, scope and byte custody are checked without inventing steps.
func TestActivityExplanationSavedBaselineLedger(t *testing.T) {
	file := os.Getenv("GOO_ACTIVITY_EXPLANATION_SAVED_BASELINE_LEDGER")
	if file == "" {
		t.Skip("requires actual saved-baseline late-arrival evidence")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var seed aeSavedBaselineLedger
	if e = json.Unmarshal(raw, &seed); e != nil {
		t.Fatal(e)
	}
	dir := filepath.Dir(file)
	before, e := os.ReadFile(filepath.Join(dir, seed.ScopeStoredBeforePath))
	if e != nil {
		t.Fatal(e)
	}
	after, e := os.ReadFile(filepath.Join(dir, seed.ScopeStoredAfterPath))
	if e != nil {
		t.Fatal(e)
	}
	originalRaw, e := os.ReadFile(filepath.Join(dir, seed.OriginalLedger))
	if e != nil {
		t.Fatal(e)
	}
	var original aeLiveLedger
	if e = json.Unmarshal(originalRaw, &original); e != nil {
		t.Fatal(e)
	}
	check := func(control int) bool {
		var l aeSavedBaselineLedger
		if e = json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		prior, _ := json.Marshal(l)
		switch control {
		case 0:
			l.ScopeStoredBefore.ID = "changed"
		case 1:
			l.ScopeStoredAfter.Title += "changed"
		case 2:
			l.Rerun.Request.Baseline.Where = nil
		case 3:
			l.Rerun.Result.Baseline.Count = original.LoadedScope.Result.Baseline.Count
			l.Rerun.Result.Basis = peLiveResultBasis(l.Rerun.Result)
		case 4:
			l.Rerun.Result.Selection.Count += "1"
			l.Rerun.Result.Basis = peLiveResultBasis(l.Rerun.Result)
		case 5:
			l.ScopeCaptureAfter.Receipt.Question.Explore.Where = nil
		case 6:
			l.ScopeRawAfter += "\n"
		case 7:
			l.BaselineLateRecords[0].Resource["team"] = "excluded"
		}
		if control >= 0 {
			now, _ := json.Marshal(l)
			if string(now) == string(prior) {
				t.Fatalf("saved baseline control %d did not change evidence", control)
			}
		}
		return aeSavedBaselineEvidence(l, original, before, after)
	}
	if !check(-1) {
		t.Fatal("actual saved-baseline reread/custody evidence rejected")
	}
	for i := 0; i < 8; i++ {
		if check(i) {
			t.Fatalf("saved baseline control %d accepted", i)
		}
		t.Logf("saved baseline control %d rejected", i)
	}
}

// RawMessage formatting differs after a browser JSON roundtrip. Compare record
// values without binary64 conversion; immutable custody still hashes actual full
// stored bytes/typed representation separately.
func aeJSONEqual(a, z any) bool {
	decode := func(value any) (any, error) {
		raw, e := json.Marshal(value)
		if e != nil {
			return nil, e
		}
		d := json.NewDecoder(bytes.NewReader(raw))
		d.UseNumber()
		var out any
		e = d.Decode(&out)
		return out, e
	}
	x, e := decode(a)
	if e != nil {
		return false
	}
	y, e := decode(z)
	return e == nil && reflect.DeepEqual(x, y)
}
