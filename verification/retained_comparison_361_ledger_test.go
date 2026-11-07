package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/url"
	"os"
	"slices"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type retainedLocation361 struct {
	ID, Cursor string
	Offset     int
}
type retainedMetadata361 struct {
	Receipt                                                        exploration.CaptureReceipt
	Question                                                       exploration.Query
	SourceReadAt, Basis, ObservationState, MatchingCapturedRecords string
	Page                                                           struct {
		Offset, Limit, Returned int
		HasMore                 *bool
	}
	RetainedRecords int
	RecordsLimited  bool
}
type retainedSide361 struct {
	Metadata retainedMetadata361
	Location retainedLocation361
}
type retainedInput361 struct {
	Baseline, Selection struct {
		Rows      []json.RawMessage
		Truncated bool
	}
}
type retainedState361 struct {
	Calls, Source, Store, SQL int
	CandidateLocation         *retainedLocation361
	DomID                     *int
	Input                     *retainedInput361
	Provenance                struct{ Baseline, Selection *retainedMetadata361 }
	InspectorRow              json.RawMessage
	BaselineStatus            string
}
type retainedFact361 struct {
	Name          string
	Before, After retainedState361
}
type retainedCapture361 struct {
	Key                                 string
	Question                            exploration.Query
	Capture, Page100                    exploration.CapturePage
	StoredBytesBefore, StoredBytesAfter string
	Oracle                              struct {
		Count string
		Rows  []json.RawMessage
	}
}
type retainedComparison361 struct {
	Input               retainedInput361
	Result              exploration.TracePageComparison
	Baseline, Selection retainedSide361
	DomID               int
}
type retainedLedger361 struct {
	Version string
	Fixture struct {
		Token, Environment, Namespace string
		Rows                          []struct {
			Side                                                                         string
			Ordinal, Kind                                                                int
			TraceID, SpanID, ParentSpanID, Service, Name, Namespace, StartNS, DurationNS string
			Env                                                                          map[string]string
		}
	}
	Captures    []retainedCapture361
	Comparisons []retainedComparison361
	Facts       []retainedFact361
	Calls       []struct {
		Path, URL, Method string
		Body              json.RawMessage
	}
	Errors    []string
	OldAction struct {
		Pending, ABA struct{ Before, After retainedState361 }
	}
	Inspections []struct {
		Location               retainedLocation361
		Index                  int
		Row, InspectorObserved json.RawMessage
		CapturedOccurrence     string
	}
}

var retainedNames361 = []string{"open-reference", "pin-reference", "open-selection", "compare", "open-baseline-side", "inspect-return-baseline", "open-selection-side", "inspect-return-selection", "candidate-page", "replace-baseline", "clear", "open-reference-aba", "pin-reference-aba", "open-selection-aba", "compare-aba", "old-action-aba", "open-baseline-aba", "inspect-return-aba"}
var retainedActions361 = []string{"open_reference", "pin_reference", "open_selection", "compare_pages", "open_baseline_side", "inspect_return_baseline", "open_selection_side", "inspect_return_selection", "candidate_page", "replace_baseline", "clear_reference", "open_reference_aba", "pin_reference_aba", "open_selection_aba", "compare_aba", "old_action_aba", "open_baseline_aba", "inspect_return_aba"}
var retainedDelta361 = []int{2, 0, 1, 0, 1, 0, 1, 0, 1, 0, 0, 2, 0, 1, 0, 0, 1, 0}
var retainedStores361 = []int{2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 6, 8, 8, 9, 9, 9, 10, 10}
var retainedPaired361 = []bool{false, false, false, true, true, true, true, true, true, false, false, false, false, false, true, true, true, true}

func retainedCaptureCheck361(l retainedLedger361, c retainedCapture361) error {
	if c.Capture.Explore == nil || c.Page100.Explore == nil || c.StoredBytesBefore == "" || c.StoredBytesBefore != c.StoredBytesAfter || c.Question.Validate() != nil {
		return fmt.Errorf("complete immutable stored capture")
	}
	var stored exploration.SavedCapture
	if json.Unmarshal([]byte(c.StoredBytesBefore), &stored) != nil || stored.Explore == nil || stored.ID != c.Capture.Receipt.ID || !peLiveCaptureHash(exploration.CapturePage{Receipt: stored.CaptureReceipt, Explore: stored.Explore}) {
		return fmt.Errorf("real full capture hash")
	}
	if savedJSON360(stored.CaptureReceipt) != savedJSON360(c.Capture.Receipt) {
		return fmt.Errorf("capture DTO receipt differs from immutable stored body")
	}
	mode := 0
	if c.Key == "selection" {
		mode = 2
	}
	indices := []int{}
	for i, r := range l.Fixture.Rows {
		start, ok := new(big.Int).SetString(r.StartNS, 10)
		if !ok {
			return fmt.Errorf("fixture exact start")
		}
		inside := start.Cmp(new(big.Int).Mul(big.NewInt(c.Question.FromMS), big.NewInt(1000000))) >= 0 && start.Cmp(new(big.Int).Mul(big.NewInt(c.Question.ToMS), big.NewInt(1000000))) < 0
		want := r.Side == c.Key && r.Namespace == l.Fixture.Namespace && inside && continuationEnvironment352(mode, l.Fixture.Environment, r.Env)
		values := map[exploration.Field]string{{Kind: "attribute", Key: "pair.run"}: l.Fixture.Token, {Kind: "attribute", Key: "pair.side"}: r.Side, {Kind: "resource", Key: "service.namespace"}: r.Namespace}
		if (hdPredicate(c.Question.Where, values) && inside && continuationEnvironment352(mode, l.Fixture.Environment, r.Env)) != want {
			return fmt.Errorf("literal captured role/window/namespace predicate")
		}
		if want {
			indices = append(indices, i)
		}
	}
	slices.SortFunc(indices, func(i, j int) int {
		a, _ := new(big.Int).SetString(l.Fixture.Rows[i].StartNS, 10)
		b, _ := new(big.Int).SetString(l.Fixture.Rows[j].StartNS, 10)
		return a.Cmp(b)
	})
	wantCount := 120
	if c.Key == "selection" {
		wantCount = 135
	}
	if len(indices) != wantCount || c.Oracle.Count != strconv.Itoa(wantCount) || len(stored.Explore.Rows) != wantCount {
		return fmt.Errorf("independent raw role retention/count")
	}
	for _, p := range []exploration.CapturePage{c.Capture, c.Page100} {
		if p.Explore == nil || savedJSON360(p.Receipt) != savedJSON360(c.Capture.Receipt) || p.Explore.Basis != c.Capture.Explore.Basis || p.Explore.CapturedAt != c.Capture.Explore.CapturedAt || p.RetainedRecords != wantCount || p.RecordsLimited || p.Explore.Count != strconv.Itoa(wantCount) {
			return fmt.Errorf("independent page custody/count/clock")
		}
		expected := c.Question
		expected.Offset = p.Explore.Question.Offset
		if savedJSON360(p.Explore.Question) != savedJSON360(expected) || expected.Limit != 100 || expected.Offset != 0 && expected.Offset != 100 {
			return fmt.Errorf("whole original page selector")
		}
		start, end := expected.Offset, min(expected.Offset+100, len(indices))
		if len(p.Explore.Rows) != end-start {
			return fmt.Errorf("actual current stored page delivery")
		}
		if savedJSON360(p.Explore.Rows) != savedJSON360(stored.Explore.Rows[start:end]) {
			return fmt.Errorf("whole rich page differs from immutable stored occurrences")
		}
		for local, index := range indices[start:end] {
			r := l.Fixture.Rows[index]
			var row map[string]any
			if json.Unmarshal(p.Explore.Rows[local], &row) != nil {
				return fmt.Errorf("raw captured row")
			}
			for key, value := range map[string]string{"TraceId": r.TraceID, "SpanId": r.SpanID, "ParentSpanId": r.ParentSpanID, "ServiceName": r.Service, "SpanName": r.Name, "timestamp_ns": r.StartNS, "duration_ns": r.DurationNS} {
				if row[key] != value {
					return fmt.Errorf("original role/page occurrence%d/%s", start+local, key)
				}
			}
			attrs, _ := row["Attributes"].(map[string]any)
			resource, _ := row["ResourceAttributes"].(map[string]any)
			if attrs["pair.run"] != l.Fixture.Token || attrs["pair.side"] != c.Key || attrs["sample.ordinal"] != strconv.Itoa(r.Ordinal) || resource["service.namespace"] != r.Namespace {
				return fmt.Errorf("original qualified occurrence attributes")
			}
		}
	}
	return nil
}
func retainedSideCheck361(l retainedLedger361, s retainedSide361, c retainedCapture361, p exploration.CapturePage, initialFact int) error {
	m := s.Metadata
	q := p.Explore.Question
	if savedJSON360(m.Receipt) != savedJSON360(p.Receipt) || savedJSON360(m.Question) != savedJSON360(q) || m.SourceReadAt != p.Explore.CapturedAt || m.Basis != p.Explore.Basis || m.MatchingCapturedRecords != p.Explore.Count || m.RetainedRecords != p.RetainedRecords || m.RecordsLimited != p.RecordsLimited || m.ObservationState != p.Explore.State || m.Page.Offset != q.Offset || m.Page.Limit != q.Limit || m.Page.Returned != len(p.Explore.Rows) || m.Page.HasMore == nil || *m.Page.HasMore != p.Explore.HasMore || s.Location.ID != p.Receipt.ID || s.Location.Offset != q.Offset {
		return fmt.Errorf("independent role metadata/location binding")
	}
	f := l.Facts[initialFact]
	found := false
	for _, call := range l.Calls[f.Before.Calls:f.After.Calls] {
		u, err := url.Parse(call.URL)
		if err == nil && call.Method == "GET" && call.Path == "/api/captures/"+c.Capture.Receipt.ID && u.Query().Get("cursor") == s.Location.Cursor {
			found = true
		}
	}
	if !found || q.Offset > 0 && s.Location.Cursor == "" {
		return fmt.Errorf("opaque stored cursor not from actual verified page request")
	}
	return nil
}
func retainedAdmission361(l retainedLedger361) error {
	if l.Version != "goo.retainedComparison361/1" || len(l.Errors) != 0 || len(l.Captures) != 2 || len(l.Comparisons) != 2 || len(l.Facts) != 18 {
		return fmt.Errorf("complete observed retained comparison campaign")
	}
	for _, c := range l.Captures {
		if err := retainedCaptureCheck361(l, c); err != nil {
			return err
		}
	}
	ref, sel := l.Captures[0], l.Captures[1]
	if ref.Key != "reference" || sel.Key != "selection" {
		return fmt.Errorf("explicit comparison roles")
	}
	a, b := ref.Page100, sel.Capture
	for _, comparison := range l.Comparisons {
		if err := retainedSideCheck361(l, comparison.Baseline, ref, a, 0); err != nil {
			return err
		}
		if err := retainedSideCheck361(l, comparison.Selection, sel, b, 2); err != nil {
			return err
		}
		in := comparison.Input
		if savedJSON360(in.Baseline.Rows) != savedJSON360(a.Explore.Rows) || savedJSON360(in.Selection.Rows) != savedJSON360(b.Explore.Rows) || !in.Baseline.Truncated || !in.Selection.Truncated {
			return fmt.Errorf("whole original page rows/roles/known partial flags")
		}
		actual, err := exploration.CompareTracePages(in.Baseline.Rows, in.Selection.Rows, in.Baseline.Truncated, in.Selection.Truncated)
		if err != nil || !tcOracleEqual(tcOracle(in.Baseline.Rows, in.Selection.Rows, in.Baseline.Truncated, in.Selection.Truncated), *actual) || savedJSON360(actual) != savedJSON360(comparison.Result) {
			return fmt.Errorf("independent raw timing/group/original-index comparison")
		}
		for _, g := range actual.Groups {
			if g.Service != nil && g.Operation != nil && *g.Service == "gateway" && *g.Operation == "checkout" {
				if g.InclusiveDifferenceNS == nil || *g.InclusiveDifferenceNS != "25000000" || g.OutsideDifferenceNS == nil || *g.OutsideDifferenceNS != "17000000" {
					return fmt.Errorf("manual independent checkout duration deltas")
				}
			}
			if g.Operation != nil && (*g.Operation == "old-path" || *g.Operation == "retry") && (g.Match != nil || g.InclusiveDifferenceNS != nil || g.OutsideDifferenceNS != nil) {
				return fmt.Errorf("one-sided page absence became zero or row identity")
			}
		}
	}
	first, second := l.Comparisons[0], l.Comparisons[1]
	if first.DomID <= 0 || second.DomID <= 0 || first.DomID == second.DomID || savedJSON360(first.Input) != savedJSON360(second.Input) {
		return fmt.Errorf("explicit recompare/ABA lifecycle")
	}
	baseSource, baseStore := l.Facts[0].Before.Source, l.Facts[0].Before.Store
	for i, f := range l.Facts {
		if f.Name != retainedNames361[i] || f.Before.Calls < 0 || f.After.Calls < f.Before.Calls || f.After.Calls > len(l.Calls) {
			return fmt.Errorf("actual fact order/bounds")
		}
		stores, source := 0, 0
		gets := []string{}
		for _, call := range l.Calls[f.Before.Calls:f.After.Calls] {
			if call.Method == "GET" && (call.Path == "/api/captures/"+ref.Capture.Receipt.ID || call.Path == "/api/captures/"+sel.Capture.Receipt.ID) {
				stores++
				gets = append(gets, call.URL)
			}
			if serviceTemplateSource356(serviceContinuationCall352{Path: call.Path, Method: call.Method}) {
				source++
			}
		}
		if stores != retainedDelta361[i] || source != 0 || f.After.Source != baseSource || f.After.SQL != f.Before.SQL || f.After.Store-f.Before.Store != stores || f.After.Store-baseStore != retainedStores361[i] {
			return fmt.Errorf("source-free comparison/store-only metadata boundary")
		}
		if retainedPaired361[i] {
			which := first
			if i >= 14 {
				which = second
			}
			if f.After.DomID == nil || *f.After.DomID != which.DomID || f.After.Input == nil || savedJSON360(f.After.Input) != savedJSON360(which.Input) || f.After.Provenance.Baseline == nil || f.After.Provenance.Selection == nil || savedJSON360(*f.After.Provenance.Baseline) != savedJSON360(which.Baseline.Metadata) || savedJSON360(*f.After.Provenance.Selection) != savedJSON360(which.Selection.Metadata) {
				return fmt.Errorf("candidate update relabelled/remounted frozen applied pair")
			}
		} else if f.After.DomID != nil || f.After.Input != nil {
			return fmt.Errorf("retired or implicit comparison still applied")
		}
		if i == 4 || i == 6 || i == 16 {
			side := first.Baseline
			if i == 6 {
				side = first.Selection
			}
			if len(gets) != 1 {
				return fmt.Errorf("one exact operand page open")
			}
			u, e := url.Parse(gets[0])
			if e != nil || u.Path != "/api/captures/"+side.Location.ID || u.Query().Get("cursor") != side.Location.Cursor {
				return fmt.Errorf("operand side open reconstructed/moved cursor")
			}
		}
	}
	if savedJSON360(l.OldAction.Pending.Before) != savedJSON360(l.OldAction.Pending.After) || savedJSON360(l.OldAction.ABA.Before) != savedJSON360(l.OldAction.ABA.After) {
		return fmt.Errorf("retired action renewed opaque location authority")
	}
	if len(l.Inspections) != 3 {
		return fmt.Errorf("actual local original occurrence inspections missing")
	}
	for _, s := range l.Inspections {
		var p exploration.CapturePage
		if s.Location.ID == ref.Capture.Receipt.ID {
			p = a
		} else if s.Location.ID == sel.Capture.Receipt.ID {
			p = b
		} else {
			return fmt.Errorf("unknown inspected capture")
		}
		if s.Location.Offset != p.Explore.Question.Offset || s.Index < 0 || s.Index >= len(p.Explore.Rows) || savedJSON360(s.Row) != savedJSON360(p.Explore.Rows[s.Index]) || savedJSON360(s.Row) != savedJSON360(s.InspectorObserved) || s.CapturedOccurrence != strconv.Itoa(s.Location.Offset+s.Index+1) {
			return fmt.Errorf("original captured occurrence inspection")
		}
	}
	return nil
}
func retainedModel361(t *testing.T, label string, stores, sources []int, paired []bool, coherent bool) bool {
	t.Helper()
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink, qobs.WithActor(label))
	for i, a := range retainedActions361 {
		rec.Record(context.Background(), a, nil, "obs.stores", stores[i], "obs.source", sources[i], "obs.paired", paired[i], "obs.coherent", coherent || i != 17)
	}
	checker, e := validate.NewChecker("model/retained_comparison_361.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	r, e := checker.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	artifact("retained-comparison361-"+label+".txt", []byte(r.Summary()))
	return r.OK()
}
func TestQuintRetainedPageComparisonProtocol361(t *testing.T) {
	zero := make([]int, 18)
	if !retainedModel361(t, "protocol-complete", retainedStores361, zero, retainedPaired361, true) {
		t.Fatal("complete retained protocol refused")
	}
	wrong := slices.Clone(zero)
	wrong[3] = 1
	if retainedModel361(t, "protocol-source-read", retainedStores361, wrong, retainedPaired361, true) {
		t.Fatal("comparison acquired telemetry")
	}
	pair := slices.Clone(retainedPaired361)
	pair[8] = false
	if retainedModel361(t, "protocol-candidate-dropped-pair", retainedStores361, zero, pair, true) {
		t.Fatal("candidate changed applied pair")
	}
}
func TestQuintRetainedPageComparisonObserved361(t *testing.T) {
	path := os.Getenv("GOO_RETAINED_PAGE_COMPARISON_361_REPORT")
	if path == "" {
		t.Skip("explicit genuine stored comparison visit required")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var l retainedLedger361
	if json.Unmarshal(raw, &l) != nil {
		t.Fatal("invalid retained report")
	}
	if e = retainedAdmission361(l); e != nil {
		t.Fatal(e)
	}
	stores, sources, pairs := make([]int, 18), make([]int, 18), make([]bool, 18)
	baseStore, baseSource := l.Facts[0].Before.Store, l.Facts[0].Before.Source
	for i, f := range l.Facts {
		stores[i] = f.After.Store - baseStore
		sources[i] = f.After.Source - baseSource
		pairs[i] = f.After.DomID != nil
	}
	if !retainedModel361(t, "actual", stores, sources, pairs, true) {
		t.Fatal("observed protocol refused")
	}
	mutants := []func(*retainedLedger361){func(l *retainedLedger361) { l.Comparisons[0].Baseline.Metadata.Question.FromMS++ }, func(l *retainedLedger361) { l.Comparisons[0].Baseline.Location.Cursor = "reconstructed" }, func(l *retainedLedger361) {
		l.Comparisons[0].Input.Baseline.Rows = l.Comparisons[0].Input.Selection.Rows
	}, func(l *retainedLedger361) { l.Comparisons[0].Baseline.Metadata.SourceReadAt = "2020-01-01T00:00:00Z" }, func(l *retainedLedger361) { l.Facts[3].After.Source++ }, func(l *retainedLedger361) { l.Captures[0].StoredBytesAfter += " " }}
	for i, mutate := range mutants {
		var fresh retainedLedger361
		json.Unmarshal(raw, &fresh)
		mutate(&fresh)
		if retainedAdmission361(fresh) == nil {
			t.Fatalf("negative%d admitted", i)
		}
		if retainedModel361(t, fmt.Sprintf("poison-%d", i), stores, sources, pairs, false) {
			t.Fatalf("negative%d model admitted", i)
		}
	}
}
