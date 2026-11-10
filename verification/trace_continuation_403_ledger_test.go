package verification

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type pageCounters403 struct{ Source, Stores, Captures, Documents, SQL int }
type pageFact403 struct {
	Name          string
	Before, After pageCounters403
}
type pageAttempt403 struct {
	Path    string
	Request json.RawMessage
	Result  exploration.TraceSearchResult
	Status  int
}
type pageProxy403 struct {
	SQL        string
	Parameters map[string]string
	Status     int
}
type pageLedger403 struct {
	Version, Status string
	Errors          []string
	Fixture         struct {
		Token    string
		Question exploration.TraceSearchRequest
		Rows     []traceSearchRawRow382
	}
	RawRows                             []traceSearchRawRow382
	Pages                               []pageAttempt403
	EmptyPage, Pending                  pageAttempt403
	CaptureRequest                      struct{ Question exploration.CaptureQuestion }
	CapturePage, ReopenedCapture        exploration.CapturePage
	SavedDocument                       workbench.Document
	StoredBytesBefore, StoredBytesAfter string
	HeldBefore, HeldAfter               json.RawMessage
	Children                            []struct {
		Request exploration.Query
		Result  json.RawMessage
		Status  int
	}
	Calls      []struct{ Path, Method string }
	Facts      []pageFact403
	Counters   pageCounters403
	ProxyCalls []pageProxy403
}

func pageBudget403(name string) (source, store, capture, document int, ok bool) {
	ok = true
	switch name {
	case "first-read", "next-read", "root-reread", "second-page-reread", "saved-root-read", "saved-next-read", "third-page-read", "last-page-read", "first-page-explicit", "pending-page-read", "empty-page-direct-read":
		source = 1
	case "fresh-child":
		source = 2
	case "capture-positioned-page":
		capture = 1
	case "publish-root-intent":
		document = 1
	case "reopen-positioned-offline":
		store = 1
	case "stage-intent", "hidden-inert-next", "retired-next", "held-back", "stage-page-capture", "restage-root", "save-root-intent", "reopen-root-intent", "draft-retires-pending-page":
	default:
		ok = false
	}
	return
}
func admitPageAttempt403(a pageAttempt403, root exploration.TraceSearchRequest, rows []tsRecord) (string, error) {
	if a.Status != 200 {
		return "", fmt.Errorf("page status %d", a.Status)
	}
	var request exploration.TraceSearchRequest
	var position *exploration.TraceSearchPosition
	if strings.HasSuffix(a.Path, "/page") || a.Path == "" {
		var input exploration.TraceSearchPageRequest
		if json.Unmarshal(a.Request, &input) != nil || input.Validate() != nil {
			return "", fmt.Errorf("invalid positioned request")
		}
		request = input.Request
		position = &input.Position
	} else if json.Unmarshal(a.Request, &request) != nil || request.Validate() != nil {
		return "", fmt.Errorf("invalid root request")
	}
	if !reflect.DeepEqual(request, root) || !reflect.DeepEqual(a.Result.Request, request) || !reflect.DeepEqual(a.Result.Position, position) {
		return "", fmt.Errorf("page request/position scope changed")
	}
	if err := exploration.ValidateTraceSearchResult(&a.Result); err != nil {
		return "", err
	}
	full := root
	full.Limit = 100
	all, total := tsOracle(full, rows)
	eligible := []tsItem{}
	for _, item := range all {
		if position == nil || item.TraceID > position.LastTraceID {
			eligible = append(eligible, item)
		}
	}
	remaining := len(eligible)
	if len(eligible) > root.Limit {
		eligible = eligible[:root.Limit]
	}
	if a.Result.MatchedTraces != total || a.Result.RemainingTraces != strconv.Itoa(remaining) || a.Result.HasMore != (remaining > root.Limit) || !tsItemsEqual(a.Result.Items, eligible) {
		return "", fmt.Errorf("independent full-population/page/witness oracle differs")
	}
	plan, err := exploration.CompileTraceSearch(root)
	if position != nil {
		plan, err = exploration.CompileTraceSearchPage(exploration.TraceSearchPageRequest{Request: root, Position: *position})
	}
	if err != nil || !reflect.DeepEqual(plan, a.Result.Plan) {
		return "", fmt.Errorf("page compiled plan/binds changed")
	}
	return retainedStatement401(plan.SQL, plan.Parameters, false), nil
}
func admitPageLedger403(r pageLedger403) error {
	fail := func(s string) error { return fmt.Errorf("trace continuation403: %s", s) }
	if r.Version != "goo.traceContinuation403/1" || r.Status != "passed" || len(r.Errors) > 0 || len(r.Pages) < 8 || len(r.RawRows) != len(r.Fixture.Rows) || len(r.Facts) != 24 {
		return fail("incomplete actual journey")
	}
	root := r.Fixture.Question
	if root.Validate() != nil || root.Limit != 2 || len(root.Conditions) != 2 {
		return fail("fixture scope absent")
	}
	for _, p := range root.Conditions {
		if mandatoryBank403(p) != r.Fixture.Token {
			return fail("source oracle lacks mandatory fixture bank")
		}
	}
	for _, row := range r.RawRows {
		if row.Attributes["fixture.bank"] != r.Fixture.Token {
			return fail("foreign source oracle bank")
		}
	}
	rows := traceSearchRecords382(r.RawRows)
	expected := map[string]int{}
	add := func(a pageAttempt403) error {
		key, err := admitPageAttempt403(a, root, rows)
		if err != nil {
			return err
		}
		expected[key]++
		return nil
	}
	for _, p := range r.Pages {
		if err := add(p); err != nil {
			return err
		}
	}
	if err := add(r.EmptyPage); err != nil {
		return err
	}
	if r.EmptyPage.Result.RemainingTraces != "0" || r.EmptyPage.Result.MatchedTraces != "7" || r.EmptyPage.Result.State != "no_data" {
		return fail("empty page lost full count")
	}
	if err := add(r.Pending); err != nil {
		return err
	}
	if r.CapturePage.TraceSearch == nil || r.CapturePage.Receipt.Question.TraceSearchPosition == nil || !reflect.DeepEqual(r.CaptureRequest.Question, r.CapturePage.Receipt.Question) || !reflect.DeepEqual(r.CapturePage, r.ReopenedCapture) || r.StoredBytesBefore == "" || r.StoredBytesBefore != r.StoredBytesAfter {
		return fail("positioned custody/reopen changed")
	}
	if !same401(r.HeldBefore, r.HeldAfter) {
		return fail("held parent changed")
	}
	var stored exploration.SavedCapture
	if json.Unmarshal([]byte(r.StoredBytesBefore), &stored) != nil || stored.TraceSearch == nil || !reflect.DeepEqual(stored.TraceSearch, r.CapturePage.TraceSearch) || !reflect.DeepEqual(stored.CaptureReceipt, r.CapturePage.Receipt) {
		return fail("stored bytes do not own page")
	}
	storedID := stored.ID
	stored.ID = ""
	storedWire, _ := json.Marshal(stored)
	digest := sha256.Sum256(storedWire)
	if hex.EncodeToString(digest[:]) != storedID {
		return fail("positioned capture custody digest invalid")
	}
	input := exploration.TraceSearchPageRequest{Request: root, Position: *r.CapturePage.Receipt.Question.TraceSearchPosition}
	wire, _ := json.Marshal(input)
	if err := add(pageAttempt403{Request: wire, Result: *r.CapturePage.TraceSearch, Status: 200}); err != nil {
		return err
	}
	if !documentHash382(r.SavedDocument) || len(r.SavedDocument.Cards) != 1 || r.SavedDocument.Cards[0].Step == nil || r.SavedDocument.Cards[0].Step.Validate() != nil || !reflect.DeepEqual(r.SavedDocument.Cards[0].Step.TraceSearch, &root) {
		return fail("saved authored intent changed")
	}
	if len(r.Children) != 2 {
		return fail("fresh child records absent")
	}
	for _, child := range r.Children {
		if child.Status != 200 || child.Request.FromMS != root.FromMS || child.Request.ToMS != root.ToMS || !reflect.DeepEqual(child.Request.EnvironmentSelector, root.EnvironmentSelector) || child.Request.Environment != root.Environment || child.Request.Where == nil || requiredTrace382(*child.Request.Where) != r.Pages[1].Result.Items[0].TraceID {
			return fail("child fixed scope changed")
		}
		plan, err := exploration.Compile(child.Request)
		if err != nil {
			return err
		}
		for _, sql := range []string{plan.Count, plan.Rows, plan.Groups, plan.Fields, plan.TimeSeries, plan.GroupedTimeSeries} {
			if sql != "" {
				expected[retainedStatement401(sql, plan.Parameters, false)]++
			}
		}
	}
	seen := map[string]bool{}
	for _, f := range r.Facts {
		source, store, capture, document, ok := pageBudget403(f.Name)
		if !ok || seen[f.Name] || f.After.Source-f.Before.Source != source || f.After.Stores-f.Before.Stores != store || f.After.Captures-f.Before.Captures != capture || f.After.Documents-f.Before.Documents != document {
			return fail("owner read/publication budget differs")
		}
		if source == 0 && capture == 0 && f.After.SQL != f.Before.SQL {
			return fail("local action read source SQL")
		}
		seen[f.Name] = true
	}
	if r.Facts[len(r.Facts)-1].After != r.Counters {
		return fail("final owner counters differ")
	}
	actual := pageCounters403{SQL: len(r.ProxyCalls)}
	for _, call := range r.Calls {
		if call.Method == "POST" {
			switch call.Path {
			case "/api/trace-search", "/api/trace-search/page", "/api/explore":
				actual.Source++
			case "/api/captures":
				actual.Captures++
			case "/api/documents":
				actual.Documents++
			default:
				if !strings.HasPrefix(call.Path, "/api/queries/") {
					return fail("unexpected browser source/publication family")
				}
			}
		}
		if call.Method == "GET" {
			for _, prefix := range []string{"/api/captures/", "/api/documents/"} {
				if strings.HasPrefix(call.Path, prefix) && len(strings.TrimPrefix(call.Path, prefix)) == 64 {
					actual.Stores++
				}
			}
		}
	}
	if actual != r.Counters {
		return fail("actual requests differ from observed owner budgets")
	}
	physical := map[string]int{}
	for _, call := range r.ProxyCalls {
		if call.Status == 200 {
			physical[retainedStatement401(call.SQL, call.Parameters, true)]++
		}
	}
	wantSQL := 0
	for key, count := range expected {
		wantSQL += count
		if physical[key] != count {
			return fail("physical compiled SQL/binding multiset differs")
		}
	}
	if wantSQL != len(r.ProxyCalls) {
		return fail("unaccounted physical source statement")
	}
	return nil
}
func mandatoryBank403(p exploration.Predicate) string {
	if p.Field != nil && p.Field.Kind == "attribute" && p.Field.Key == "fixture.bank" && p.Op == "eq" {
		return p.Value
	}
	for _, c := range p.All {
		if v := mandatoryBank403(c); v != "" {
			return v
		}
	}
	return ""
}

func replayPage403(t *testing.T, r pageLedger403, mutate bool) bool {
	t.Helper()
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	reads, stores, captures, documents := []int{}, []int{}, []int{}, []int{}
	for i, f := range r.Facts {
		a, b, c, d, _ := pageBudget403(f.Name)
		reads = append(reads, a)
		stores = append(stores, b)
		captures = append(captures, c)
		documents = append(documents, d)
		source := f.After.Source
		if mutate && f.Name == "hidden-inert-next" {
			source++
		}
		rec.Record(context.Background(), "step", nil, "obs.phase", i+1, "obs.reads", source, "obs.stores", f.After.Stores, "obs.captures", f.After.Captures, "obs.documents", f.After.Documents)
	}
	list := func(v []int) string { b, _ := json.Marshal(v); return string(b) }
	spec, err := filepath.Abs("model/trace_continuation403.qnt")
	if err != nil {
		t.Fatal(err)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: trace_continuation403\nconstants: {READS: '%s', STORES: '%s', CAPTURES: '%s', DOCUMENTS: '%s'}\ninvariants: [safety]\nactions:\n  step: {expect: 'phase == {{.obs.phase}} and reads == {{.obs.reads}} and stores == {{.obs.stores}} and captures == {{.obs.captures}} and documents == {{.obs.documents}}'}\n", spec, list(reads), list(stores), list(captures), list(documents))
	dir := t.TempDir()
	file := filepath.Join(dir, "binding.yaml")
	if err = os.WriteFile(file, []byte(binding), 0600); err != nil {
		t.Fatal(err)
	}
	checker, err := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	result, err := checker.Check(t.Context(), sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	return result.OK()
}
func TestQuintTraceContinuationObserved403(t *testing.T) {
	path := os.Getenv("GOO_TRACE_CONTINUATION_403_REPORT")
	if path == "" {
		t.Skip("actual trace continuation journey required")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var r pageLedger403
	if json.Unmarshal(raw, &r) != nil {
		t.Fatal("invalid observed page report")
	}
	if err = admitPageLedger403(r); err != nil {
		t.Fatal(err)
	}
	mutants := []func(*pageLedger403){func(x *pageLedger403) { x.Pages[1].Result.Position.RequestHash = strings.Repeat("a", 64) }, func(x *pageLedger403) { x.Pages[1].Result.RemainingTraces = "0" }, func(x *pageLedger403) { x.Pages[1].Result.Items[0].TraceID = strings.Repeat("f", 32) }, func(x *pageLedger403) { x.Pages[1].Result.Items[0].Witnesses[0].Service = "foreign" }, func(x *pageLedger403) { x.EmptyPage.Result.MatchedTraces = "0" }, func(x *pageLedger403) { x.HeldAfter = json.RawMessage(`{}`) }, func(x *pageLedger403) { x.CapturePage.Receipt.Question.TraceSearchPosition = nil }, func(x *pageLedger403) { x.SavedDocument.Cards[0].Step.TraceSearch.Limit++ }, func(x *pageLedger403) { x.Facts[2].After.Source++ }, func(x *pageLedger403) {
		for k := range x.ProxyCalls[0].Parameters {
			if strings.HasPrefix(k, "param_") {
				x.ProxyCalls[0].Parameters[k] = "foreign"
				break
			}
		}
	}}
	for i, mutate := range mutants {
		var bad pageLedger403
		json.Unmarshal(raw, &bad)
		mutate(&bad)
		if admitPageLedger403(bad) == nil {
			t.Fatalf("semantic mutant %d admitted", i)
		}
	}
	if !replayPage403(t, r, false) || replayPage403(t, r, true) {
		t.Fatal("observed model or local-read negative control incorrect")
	}
	t.Logf("%d pages / %d observed ownership boundaries / ten semantic mutants and one model control", len(r.Pages), len(r.Facts))
}

func TestBombadilTraceContinuationObserved403(t *testing.T) {
	mainPath, actorPath := os.Getenv("GOO_TRACE_CONTINUATION_403_REPORT"), os.Getenv("GOO_TRACE_CONTINUATION_403_ACTOR")
	if mainPath == "" || actorPath == "" {
		t.Skip("actual trace continuation main and actor required")
	}
	raw, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}
	var main pageLedger403
	if json.Unmarshal(raw, &main) != nil {
		t.Fatal("invalid main report")
	}
	if err = admitPageLedger403(main); err != nil {
		t.Fatal(err)
	}
	raw, err = os.ReadFile(actorPath)
	if err != nil {
		t.Fatal(err)
	}
	var actor struct {
		Version, SourceBank string
		Expected            struct {
			Question      exploration.TraceSearchRequest
			First, Second []string
		}
		Actor struct {
			Snapshots int
			Phases    []string
			Complete  struct {
				Version, Phase                         string
				Reads                                  []pageAttempt403
				IDs                                    []string
				Exact, BudgetSafe, Complete, NoPagePan bool
			}
		}
		ProxyCalls []pageProxy403
	}
	if json.Unmarshal(raw, &actor) != nil || actor.Version != "goo.traceContinuation403ActorReport/1" || actor.SourceBank != main.Fixture.Token || !reflect.DeepEqual(actor.Expected.Question, main.Fixture.Question) || actor.Actor.Snapshots < 3 || len(actor.Actor.Phases) < 6 {
		t.Fatal("incomplete independent actor")
	}
	complete := actor.Actor.Complete
	if complete.Version != "goo.traceContinuation403Actor/1" || complete.Phase != "complete" || !complete.Exact || !complete.BudgetSafe || !complete.Complete || !complete.NoPagePan || len(complete.Reads) != 3 || len(actor.ProxyCalls) != 3 {
		t.Fatal("actor completion/budget absent")
	}
	expected := map[string]int{}
	for i, read := range complete.Reads {
		path := "/api/trace-search"
		if i == 1 {
			path += "/page"
		}
		if read.Path != path {
			t.Fatal("actor source sequence changed")
		}
		key, err := admitPageAttempt403(read, main.Fixture.Question, traceSearchRecords382(main.RawRows))
		if err != nil {
			t.Fatal(err)
		}
		expected[key]++
	}
	if complete.Reads[0].Result.Position != nil || complete.Reads[2].Result.Position != nil || !reflect.DeepEqual(complete.Reads[1].Result.Position, complete.Reads[0].Result.NextPosition) {
		t.Fatal("actor page chain changed")
	}
	physical := map[string]int{}
	for _, call := range actor.ProxyCalls {
		if call.Status != 200 {
			t.Fatal("actor source failed")
		}
		physical[retainedStatement401(call.SQL, call.Parameters, true)]++
	}
	if !reflect.DeepEqual(physical, expected) {
		t.Fatal("actor physical source plans/binds differ")
	}
	ids := []string{}
	for _, item := range complete.Reads[2].Result.Items {
		ids = append(ids, item.TraceID)
	}
	if !reflect.DeepEqual(ids, complete.IDs) {
		t.Fatal("actor DOM page IDs differ from checked source result")
	}
	t.Logf("%d actor snapshots / three explicit first-next-first page reads and exact source statements", actor.Actor.Snapshots)
}
