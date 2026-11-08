package verification

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type traceSearchState382 struct{ Source, SQL, Calls, Documents, Captures int }
type traceSearchFact382 struct {
	Name          string
	Before, After traceSearchState382
}
type traceSearchRawRow382 struct {
	TraceID                  string `json:"traceId"`
	SpanID                   string `json:"spanId"`
	ParentSpanID             string `json:"parentSpanId"`
	Service, Operation, Kind string
	Resource, Attributes     map[string]string
	TimestampNS, DurationNS  string
}
type traceSearchLedger382 struct {
	Version                             string
	Saved                               workbench.Document
	SavedRequest                        workbench.DocumentInput
	SavedStep                           workbench.Step
	StagedStep                          workbench.Step
	StagedRequest                       exploration.TraceSearchRequest
	SavedCardConditions                 []exploration.Predicate
	Reopened                            workbench.Document
	StagedHref                          string
	SharedHref                          string
	StoredBytesBefore, StoredBytesAfter string
	Fixture                             struct {
		ParentQuestion exploration.Query
		First, Other   traceSearchRawRow382
		HighThreshold  string
	}
	RawRows  []traceSearchRawRow382
	Searches []struct {
		Request exploration.TraceSearchRequest
		Result  exploration.TraceSearchResult
		Status  int
	}
	Facts  []traceSearchFact382
	Stale  struct{ Before, After json.RawMessage }
	Errors []string
}

func documentHash382(document workbench.Document) bool {
	id := document.ID
	document.ID = ""
	wire, err := json.Marshal(document)
	if err != nil {
		return false
	}
	sum := sha256.Sum256(wire)
	return id == hex.EncodeToString(sum[:])
}
func traceSearchRecords382(rows []traceSearchRawRow382) []tsRecord {
	out := make([]tsRecord, 0, len(rows))
	for _, row := range rows {
		values := map[exploration.Field]string{}
		for key, value := range row.Resource {
			values[exploration.Field{Kind: "resource", Key: key}] = value
		}
		for key, value := range row.Attributes {
			values[exploration.Field{Kind: "attribute", Key: key}] = value
		}
		for key, value := range map[string]string{"trace_id": row.TraceID, "span_id": row.SpanID, "parent_span_id": row.ParentSpanID, "service": row.Service, "name": row.Operation, "kind": row.Kind, "timestamp_ns": row.TimestampNS, "duration_ns": row.DurationNS} {
			values[exploration.Field{Kind: "column", Key: key}] = value
		}
		out = append(out, tsRecord{Values: values})
	}
	return out
}

// Only a logically mandatory literal can bound the independent source bank.
// Merely finding a trace-id leaf somewhere under Any would be insufficient.
func requiredTrace382(p exploration.Predicate) string {
	if p.Field != nil && p.Field.Kind == "column" && p.Field.Key == "trace_id" && p.Op == "eq" {
		return p.Value
	}
	if p.All != nil {
		for _, child := range p.All {
			if id := requiredTrace382(child); id != "" {
				return id
			}
		}
	}
	if len(p.Any) > 0 {
		id := requiredTrace382(p.Any[0])
		if id == "" {
			return ""
		}
		for _, child := range p.Any[1:] {
			if requiredTrace382(child) != id {
				return ""
			}
		}
		return id
	}
	return ""
}
func stepFromURL382(raw string) (*workbench.Step, error) {
	link, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	payload, ok := strings.CutPrefix(link.Fragment, "step=")
	if !ok {
		return nil, fmt.Errorf("step link unavailable")
	}
	var step workbench.Step
	if err = json.Unmarshal([]byte(payload), &step); err != nil {
		return nil, err
	}
	return &step, step.Validate()
}
func traceSearchAdmission382(report traceSearchLedger382) error {
	fail := func(message string) error { return fmt.Errorf("saved trace-search382: %s", message) }
	if report.Version != "goo.savedTraceSearch382/1" || len(report.Errors) != 0 || report.Saved.DocumentInput.Validate() != nil || !documentHash382(report.Saved) || len(report.Saved.Cards) != 1 {
		return fail("immutable document unavailable")
	}
	card := report.Saved.Cards[0]
	if card.Kind != "question" || card.Step == nil || card.Step.Kind != "trace-search" || len(card.Evidence) != 0 || savedJSON360(report.Saved.DocumentInput) != savedJSON360(report.SavedRequest) || !reflect.DeepEqual(*card.Step, report.SavedStep) {
		return fail("authored exact search intent changed")
	}
	request := card.Step.TraceSearch
	if request == nil || request.Validate() != nil || len(request.Conditions) != 3 || request.Limit != 7 {
		return fail("nontrivial search intent unavailable")
	}
	if report.Fixture.HighThreshold != "9007199254740993" {
		return fail("large numeric threshold unavailable")
	}
	// Restrict the independent source bank to a trace literally present in the
	// original parent predicate. This proves a complete membership oracle for this
	// deliberately scoped journey rather than claiming the whole warehouse bank.
	trace := requiredTrace382(request.Conditions[0])
	var large bool
	var visit func(exploration.Predicate)
	visit = func(p exploration.Predicate) {
		if p.Field != nil && p.Field.Kind == "column" && p.Field.Key == "duration_ns" && p.Op == "lte" && p.Value == report.Fixture.HighThreshold {
			large = true
		}
		for _, child := range p.All {
			visit(child)
		}
		for _, child := range p.Any {
			visit(child)
		}
	}
	visit(request.Conditions[0])
	for _, p := range request.Conditions {
		visit(p)
	}
	if trace == "" || report.Fixture.ParentQuestion.Where == nil || requiredTrace382(*report.Fixture.ParentQuestion.Where) != trace || !large || len(report.RawRows) < 2 || len(report.RawRows) > 100 {
		return fail("scoped source/precision proof missing")
	}
	services := map[string]bool{}
	for _, row := range report.RawRows {
		if row.TraceID != trace {
			return fail("independent bank contains an unrelated trace")
		}
		services[row.Service] = true
	}
	if len(services) < 2 {
		return fail("distributed source witnesses missing")
	}
	if request.FromMS != report.Fixture.ParentQuestion.FromMS || request.ToMS != report.Fixture.ParentQuestion.ToMS || !(exploration.Query{Environment: request.Environment, EnvironmentSelector: request.EnvironmentSelector}).SameEnvironment(report.Fixture.ParentQuestion) {
		return fail("original fixed environment/window changed")
	}
	if report.StoredBytesBefore == "" || report.StoredBytesBefore != report.StoredBytesAfter {
		return fail("stored revision changed")
	}
	var stored workbench.Document
	if json.Unmarshal([]byte(report.StoredBytesBefore), &stored) != nil || !reflect.DeepEqual(stored, report.Saved) {
		return fail("stored revision is not saved document")
	}
	if len(report.Stale.Before) == 0 || savedJSON360(report.Stale.Before) != savedJSON360(report.Stale.After) {
		return fail("stale action changed owner")
	}
	for _, href := range []string{report.StagedHref, report.SharedHref} {
		decoded, err := stepFromURL382(href)
		if err != nil || !reflect.DeepEqual(decoded, card.Step) {
			return fail("shared link does not preserve typed search intent")
		}
	}
	if !reflect.DeepEqual(report.StagedStep, *card.Step) || !reflect.DeepEqual(report.StagedRequest, *request) || !reflect.DeepEqual(report.SavedCardConditions, request.Conditions) || !reflect.DeepEqual(report.Reopened, report.Saved) {
		return fail("shared serialized step changed intent")
	}
	if len(report.Searches) != 2 {
		return fail("explicit search reads missing")
	}
	records := traceSearchRecords382(report.RawRows)
	expected, total := tsOracle(*request, records)
	if total != "1" || len(expected) != 1 {
		return fail("independent existential intersection changed")
	}
	expectedPlan, err := exploration.CompileTraceSearch(*request)
	if err != nil {
		return fail("saved search intent cannot be compiled")
	}
	for _, search := range report.Searches {
		if search.Status != 200 || !reflect.DeepEqual(search.Request, *request) || !reflect.DeepEqual(search.Result.Request, *request) || search.Result.MatchedTraces != total || search.Result.HasMore || search.Result.State != "observed" || search.Result.Basis == "" || search.Result.ReadAt == "" || !reflect.DeepEqual(search.Result.Plan, expectedPlan) || !tsItemsEqual(search.Result.Items, expected) {
			return fail("actual search differs from exact saved intent/source witnesses")
		}
	}
	names := []string{"parent-stage", "parent-run", "open-search", "add-conditions", "initial-search", "save-intent", "stale-save", "save-revision", "reopen-revision", "inspect-saved-question", "reopen-question", "reopened-search", "share-link"}
	reads := []int{0, 1, 0, 0, 1, 0, 0, 0, 0, 0, 0, 1, 0}
	if len(report.Facts) != len(names) {
		return fail("actual source ownership facts missing")
	}
	for i, fact := range report.Facts {
		if fact.Name != names[i] || fact.After.Source-fact.Before.Source != reads[i] || fact.After.SQL < fact.Before.SQL || reads[i] == 0 && fact.After.SQL != fact.Before.SQL {
			return fail("passive/explicit source budget changed")
		}
		if fact.Name == "inspect-saved-question" && fact.After.Calls != fact.Before.Calls {
			return fail("saved-condition disclosure made a network request")
		}
		expectedDocuments := 0
		if fact.Name == "save-revision" {
			expectedDocuments = 1
		}
		if fact.After.Documents-fact.Before.Documents != expectedDocuments || fact.After.Captures != fact.Before.Captures {
			return fail("unexpected intent/evidence publication")
		}
		if i > 0 && fact.Before != report.Facts[i-1].After {
			return fail("source counters discontinuous")
		}
	}
	return nil
}

func traceSearchReplay382(t *testing.T, label string, facts []traceSearchFact382) bool {
	t.Helper()
	if len(facts) == 0 {
		t.Fatal("actual facts required")
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	baseline := facts[0].Before.Source
	for _, fact := range facts {
		action := "history"
		switch fact.Name {
		case "parent-run", "initial-search", "reopened-search":
			action = "read"
		case "parent-stage", "open-search", "add-conditions", "save-intent", "stale-save", "save-revision", "reopen-revision", "inspect-saved-question", "reopen-question", "share-link":
		default:
			t.Fatal("unknown recorded trace-search action", fact.Name)
		}
		rec.Record(context.Background(), action, nil, "obs.reads", fact.After.Source-baseline)
	}
	steps, _ := json.Marshal(sink.Steps())
	if err := artifact("trace-search382-"+label+"-steps.json", steps); err != nil {
		t.Fatal(err)
	}
	checker, err := validate.NewChecker("model/saved_trace_search_382.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	report, err := checker.Check(t.Context(), sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	if err = artifact("trace-search382-"+label+"-model.txt", []byte(report.Summary())); err != nil {
		t.Fatal(err)
	}
	return report.OK()
}
func TestQuintSavedTraceSearchObserved382(t *testing.T) {
	path := os.Getenv("GOO_SAVED_TRACE_SEARCH_382_REPORT")
	if path == "" {
		t.Skip("actual saved trace-search report opt-in")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var report traceSearchLedger382
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if err = traceSearchAdmission382(report); err != nil {
		t.Fatal(err)
	}
	clone := func() traceSearchLedger382 {
		var value traceSearchLedger382
		if err := json.Unmarshal(raw, &value); err != nil {
			t.Fatal(err)
		}
		return value
	}
	mutants := []struct {
		name   string
		change func(*traceSearchLedger382)
	}{
		{"scope", func(r *traceSearchLedger382) { r.SavedStep.Scope.ToMS++ }},
		{"order", func(r *traceSearchLedger382) { slices.Reverse(r.Searches[1].Request.Conditions) }},
		{"limit", func(r *traceSearchLedger382) { r.Searches[1].Request.Limit++ }},
		{"disclosure-fetch", func(r *traceSearchLedger382) { r.Facts[9].After.Calls++ }},
		{"disclosed-condition", func(r *traceSearchLedger382) { slices.Reverse(r.SavedCardConditions) }},
		{"link", func(r *traceSearchLedger382) { r.SharedHref = "#step=%7B%7D" }},
		{"precision", func(r *traceSearchLedger382) { r.Fixture.HighThreshold = "9007199254740992" }},
		{"revision", func(r *traceSearchLedger382) {
			r.Saved.ID = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
		}},
		{"custody", func(r *traceSearchLedger382) { r.StoredBytesAfter += "x" }},
		{"stale", func(r *traceSearchLedger382) { r.Stale.After = json.RawMessage(`{"changed":true}`) }},
		{"passive-source", func(r *traceSearchLedger382) { r.Facts[5].After.Source++ }},
		{"passive-sql", func(r *traceSearchLedger382) { r.Facts[8].After.SQL++ }},
		{"plan", func(r *traceSearchLedger382) { r.Searches[1].Result.Plan.SQL += " LIMIT 1" }},
		{"witness", func(r *traceSearchLedger382) { r.Searches[1].Result.Items[0].Witnesses[0].Service = "foreign" }},
	}
	for _, mutant := range mutants {
		bad := clone()
		mutant.change(&bad)
		if traceSearchAdmission382(bad) == nil {
			t.Fatal("admitted actual-ledger mutant", mutant.name)
		}
	}
	if !traceSearchReplay382(t, "actual", report.Facts) {
		t.Fatal("actual source lifecycle failed model")
	}
	bad := slices.Clone(report.Facts)
	bad[5].After.Source++
	if traceSearchReplay382(t, "save-source", bad) {
		t.Fatal("model admitted source read while saving intent")
	}
	bad = slices.Clone(report.Facts)
	bad[11].After.Source--
	if traceSearchReplay382(t, "missing-explicit-search", bad) {
		t.Fatal("model admitted explicit search without recorded read")
	}
	t.Log("checked " + strconv.Itoa(len(report.Facts)) + " measured source ownership facts")
}
