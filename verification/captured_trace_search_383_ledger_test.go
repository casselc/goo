package verification

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type searchCaptureCounters383 struct{ Calls, Source, SQL, Documents, Captures, StoreReads int }
type searchCaptureFact383 struct {
	Name          string
	Before, After searchCaptureCounters383
}
type searchCaptureRead383 struct {
	Request                   exploration.Query
	Response                  json.RawMessage
	Status                    int
	AbortsBefore, AbortsAfter int
}
type searchCaptureLedger383 struct {
	Version   string
	Errors    []string
	RawRows   []traceSearchRawRow382
	Displayed struct {
		Request exploration.TraceSearchRequest
		Result  exploration.TraceSearchResult
		Status  int
	}
	CapturePage, Reopened               exploration.CapturePage
	CapturedRequest                     struct{ Question exploration.CaptureQuestion }
	SourceRefusalProbe                  struct{ Status int }
	StoredBytesBefore, StoredBytesAfter string
	Planned                             exploration.Query
	Citation                            json.RawMessage
	HeldBefore, HeldAfter               json.RawMessage
	Child, Refused, Pending             searchCaptureRead383
	Facts                               []searchCaptureFact383
	ReturnFocus                         []bool
}

func searchCaptureAction383(name string) string {
	switch name {
	case "run-search", "run", "run-refused", "stage-pending-run":
		return "read"
	case "capture":
		return "capture"
	case "stage-search", "save-question", "reopen-offline", "inspect-witness", "stage", "return", "stage-refused", "return-refused", "return-pending":
		return "history"
	}
	return ""
}
func admitSearchCapture383(report searchCaptureLedger383) error {
	fail := func(s string) error { return fmt.Errorf("captured trace search383: %s", s) }
	if report.Version != "goo.capturedTraceSearch383/1" || len(report.Errors) > 0 {
		return fail("wrong version/errors")
	}
	page := report.CapturePage
	if page.TraceSearch == nil || page.Receipt.Question.Kind != "trace-search" || page.Receipt.Question.TraceSearch == nil || page.Receipt.Question.Validate() != nil {
		return fail("typed capture question/result absent")
	}
	q := *page.Receipt.Question.TraceSearch
	r := page.TraceSearch
	if !reflect.DeepEqual(r.Request, q) || !reflect.DeepEqual(report.CapturedRequest.Question, page.Receipt.Question) || !reflect.DeepEqual(report.Displayed.Request, q) || report.Displayed.Status != 200 {
		return fail("capture reread changed fixed question")
	}
	if !reflect.DeepEqual(page, report.Reopened) {
		return fail("offline reopen changed immutable evidence")
	}
	if report.StoredBytesBefore == "" || report.StoredBytesBefore != report.StoredBytesAfter {
		return fail("custody bytes changed")
	}
	var stored exploration.SavedCapture
	if json.Unmarshal([]byte(report.StoredBytesBefore), &stored) != nil || stored.ID != page.Receipt.ID || stored.TraceSearch == nil || !reflect.DeepEqual(stored.TraceSearch, r) || !reflect.DeepEqual(stored.CaptureReceipt, page.Receipt) {
		return fail("capture bytes do not own reopened evidence")
	}
	id := stored.ID
	stored.ID = ""
	wire, _ := json.Marshal(stored)
	sum := sha256.Sum256(wire)
	if hex.EncodeToString(sum[:]) != id {
		return fail("immutable capture digest invalid")
	}
	readAt, e := time.Parse(time.RFC3339Nano, r.ReadAt)
	capturedAt, f := time.Parse(time.RFC3339Nano, page.Receipt.CapturedAt)
	if e != nil || f != nil || readAt.After(capturedAt) || r.ReadAt == report.Displayed.Result.ReadAt || r.Basis == "" {
		return fail("independent read/custody clocks absent")
	}
	if page.RetainedRecords != len(r.Items) || page.RecordsLimited != r.HasMore || page.NextCursor != "" || page.PreviousCursor != "" {
		return fail("search witnesses represented as record pages")
	}
	plan, e := exploration.CompileTraceSearch(q)
	if e != nil || !reflect.DeepEqual(plan, r.Plan) {
		return fail("saved source plan changed")
	}
	// The independent source bank must be bounded by a mandatory literal trace ID;
	// list truncation cannot make an incomplete bank a full matching-count oracle.
	trace := ""
	for _, p := range q.Conditions {
		if value := requiredTrace382(p); value != "" {
			trace = value
			break
		}
	}
	if trace == "" || len(report.RawRows) < 2 {
		return fail("bounded independent raw-record bank absent")
	}
	services := map[string]bool{}
	for _, row := range report.RawRows {
		if row.TraceID != trace {
			return fail("unrelated trace in source bank")
		}
		services[row.Service] = true
	}
	if len(services) < 2 {
		return fail("distributed witness membership absent")
	}
	expected, total := tsOracle(q, traceSearchRecords382(report.RawRows))
	if r.State != "observed" || r.MatchedTraces != total || !tsItemsEqual(r.Items, expected) || r.HasMore {
		return fail("saved count/witnesses differ from independent raw records")
	}
	var citation struct{ Receipt exploration.CaptureReceipt }
	if json.Unmarshal(report.Citation, &citation) != nil || !reflect.DeepEqual(citation.Receipt, page.Receipt) {
		return fail("child citation changed immutable search receipt")
	}
	for _, child := range []searchCaptureRead383{report.Child, report.Pending} {
		var result exploration.Result
		if json.Unmarshal(child.Response, &result) != nil || result.Count != strconv.Itoa(len(report.RawRows)) || result.HasMore || result.Basis == "" {
			return fail("fresh trace child differs from independent source population")
		}
		want := []string{}
		got := []string{}
		for _, row := range report.RawRows {
			want = append(want, row.SpanID)
		}
		for _, raw := range result.Rows {
			var row struct {
				SpanID  string `json:"SpanId"`
				TraceID string `json:"TraceId"`
			}
			if json.Unmarshal(raw, &row) != nil || row.TraceID != trace {
				return fail("child returned foreign trace records")
			}
			got = append(got, row.SpanID)
		}
		slices.Sort(want)
		slices.Sort(got)
		if !slices.Equal(want, got) {
			return fail("child record membership changed")
		}
	}
	if report.Planned.Signal != "spans" || report.Planned.FromMS != q.FromMS || report.Planned.ToMS != q.ToMS || !(exploration.Query{Environment: q.Environment, EnvironmentSelector: q.EnvironmentSelector}).SameEnvironment(report.Planned) || report.Planned.Where == nil || requiredTrace382(*report.Planned.Where) != trace {
		return fail("child trace staged a foreign scope")
	}
	var traceOnly func(exploration.Predicate) bool
	traceOnly = func(p exploration.Predicate) bool {
		if len(p.All) == 1 && len(p.Any) == 0 && p.Field == nil && p.Numeric == nil {
			return traceOnly(p.All[0])
		}
		return len(p.All) == 0 && len(p.Any) == 0 && p.Numeric == nil && p.Field != nil && p.Field.Kind == "column" && p.Field.Key == "trace_id" && p.Op == "eq" && p.Value == trace
	}
	if report.Planned.Where == nil || !traceOnly(*report.Planned.Where) {
		return fail("trace child inherited span-condition restrictions")
	}
	if len(report.Citation) == 0 || !strings.Contains(string(report.Citation), id) || len(report.HeldBefore) == 0 || !reviewTransportEqual365(report.HeldBefore, report.HeldAfter) {
		return fail("held DOM/selection/citation/focus changed on return")
	}
	if len(report.ReturnFocus) != 3 || !report.ReturnFocus[0] || !report.ReturnFocus[1] || !report.ReturnFocus[2] || report.Pending.Status != 200 || report.Pending.AbortsAfter <= report.Pending.AbortsBefore || report.Child.Status != 200 || !reflect.DeepEqual(report.Child.Request, report.Planned) || report.Refused.Status < 400 || !reflect.DeepEqual(report.Refused.Request, report.Planned) || !reflect.DeepEqual(report.Pending.Request, report.Planned) {
		return fail("successful/refused/pending child scenarios absent")
	}
	names := []string{"stage-search", "run-search", "save-question", "capture", "reopen-offline", "inspect-witness", "stage", "run", "return", "stage-refused", "run-refused", "return-refused", "stage-pending-run", "return-pending"}
	if len(report.Facts) != len(names) {
		return fail("measured fact sequence absent")
	}
	for i, fact := range report.Facts {
		if fact.Name != names[i] || searchCaptureAction383(fact.Name) == "" {
			return fail("unexpected actual fact")
		}
		read := 0
		capture := 0
		if searchCaptureAction383(fact.Name) == "read" {
			read = 1
		}
		if fact.Name == "capture" {
			capture = 1
		}
		if fact.After.Source-fact.Before.Source != read || fact.After.Captures-fact.Before.Captures != capture || fact.After.Documents-fact.Before.Documents != func() int {
			if fact.Name == "save-question" || fact.Name == "capture" {
				return 1
			}
			return 0
		}() {
			return fail("implicit source read or custody write")
		}
		if read == 0 && capture == 0 && fact.After.SQL != fact.Before.SQL {
			return fail("passive action reread source SQL")
		}
		expectedStoreReads := 0
		if fact.Name == "reopen-offline" {
			expectedStoreReads = 1
		}
		if fact.After.StoreReads-fact.Before.StoreReads != expectedStoreReads {
			return fail("unexpected capture-store read; held parent and capture response must be reused")
		}
		if i > 0 {
			previous := report.Facts[i-1].After
			if fact.Name == "reopen-offline" {
				if report.SourceRefusalProbe.Status != 502 || fact.Before.SQL-previous.SQL != 1 {
					return fail("offline refusal probe absent")
				}
				previous.SQL = fact.Before.SQL
			}
			if fact.Before != previous {
				return fail("counter discontinuity")
			}
		}
	}
	return nil
}
func replaySearchCapture383(t *testing.T, label string, facts []searchCaptureFact383) bool {
	t.Helper()
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	base := facts[0].Before
	for _, f := range facts {
		action := searchCaptureAction383(f.Name)
		if action == "" {
			t.Fatal("unknown action", f.Name)
		}
		rec.Record(context.Background(), action, nil, "obs.reads", f.After.Source-base.Source, "obs.captures", f.After.Captures-base.Captures)
	}
	wire, _ := json.Marshal(sink.Steps())
	if err := artifact("captured-trace-search383-"+label+"-steps.json", wire); err != nil {
		t.Fatal(err)
	}
	checker, err := validate.NewChecker("model/captured_trace_search_383.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	result, err := checker.Check(t.Context(), sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	if err = artifact("captured-trace-search383-"+label+"-model.txt", []byte(result.Summary())); err != nil {
		t.Fatal(err)
	}
	return result.OK()
}
func TestQuintCapturedTraceSearchObserved383(t *testing.T) {
	path := os.Getenv("GOO_CAPTURED_TRACE_SEARCH_383_REPORT")
	if path == "" {
		t.Skip("actual captured trace-search report opt-in")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var report searchCaptureLedger383
	if json.Unmarshal(raw, &report) != nil {
		t.Fatal("report JSON")
	}
	if err = admitSearchCapture383(report); err != nil {
		t.Fatal(err)
	}
	if !replaySearchCapture383(t, "actual", report.Facts) {
		t.Fatal("actual measured facts violate model")
	}
	clone := func() searchCaptureLedger383 {
		var value searchCaptureLedger383
		json.Unmarshal(raw, &value)
		return value
	}
	mutations := []func(*searchCaptureLedger383){
		func(r *searchCaptureLedger383) { r.StoredBytesAfter += " " },
		func(r *searchCaptureLedger383) {
			for i := range r.RawRows {
				r.RawRows[i].DurationNS = "9007199254740994"
			}
		},
		func(r *searchCaptureLedger383) { r.CapturePage.TraceSearch.Request.Limit++ },
		func(r *searchCaptureLedger383) { r.CapturePage.TraceSearch.MatchedTraces = "2" },
		func(r *searchCaptureLedger383) {
			r.CapturePage.TraceSearch.HasMore = !r.CapturePage.TraceSearch.HasMore
		},
		func(r *searchCaptureLedger383) { r.CapturePage.TraceSearch.ReadAt = r.Displayed.Result.ReadAt },
		func(r *searchCaptureLedger383) { r.CapturePage.Receipt.CapturedAt = "not-a-clock" },
		func(r *searchCaptureLedger383) { r.HeldAfter = json.RawMessage(`{"changed":true}`) },
		func(r *searchCaptureLedger383) { r.Planned.ToMS++ },
		func(r *searchCaptureLedger383) { r.Citation = json.RawMessage(`{"unrelated":true}`) },
		func(r *searchCaptureLedger383) { r.Facts[8].After.Source++ },
		func(r *searchCaptureLedger383) { r.Facts[8].After.StoreReads++ },
		func(r *searchCaptureLedger383) {
			r.CapturePage.TraceSearch.Items[0].Witnesses[0].SpanID = "ffffffffffffffff"
		},
	}
	for i, mutate := range mutations {
		bad := clone()
		mutate(&bad)
		if admitSearchCapture383(bad) == nil {
			t.Fatalf("semantic mutant %d admitted", i)
		}
	}
	bad := clone()
	bad.Facts[len(bad.Facts)-1].After.Source++
	if replaySearchCapture383(t, "implicit-read-mutant", bad.Facts) {
		t.Fatal("model accepted implicit Back read")
	}
	bad = clone()
	bad.Facts[3].After.Captures--
	if replaySearchCapture383(t, "missing-capture-mutant", bad.Facts) {
		t.Fatal("model accepted absent capture write")
	}
	t.Logf("checked %d actual facts, %d semantic mutants and two model controls", len(report.Facts), len(mutations))
}
