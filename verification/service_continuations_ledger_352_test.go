package verification

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type serviceContinuationCall352 struct {
	Path, Search, Method string
	Body                 json.RawMessage
}
type serviceContinuationFact352 struct {
	Name                                   string
	Before, After, CallsBefore, CallsAfter int
}
type serviceContinuationCase352 struct {
	Key             string
	Scope           exploration.InventoryQuestion
	Evidence        exploration.ServiceDiscovery
	WiderQuestion   exploration.Query
	Wider           exploration.Result
	Oracle          []struct{ ID string }
	View            workbench.ActivityView
	StartedSelector json.RawMessage
	Records         []json.RawMessage
	LibraryList     workbench.ActivityViewList
	LibraryURL      string
	Facts           []serviceContinuationFact352
}
type serviceContinuationLedger352 struct {
	Version string
	Fixture struct {
		Token, Environment, Parent, Child, Worker, ParentNS, ChildNS, WorkerNS string
		FromMS, ToMS                                                           int64
	}
	Cases  []serviceContinuationCase352
	Calls  []serviceContinuationCall352
	Errors []string
}
type serviceContinuationRow352 struct {
	ID         string            `json:"SpanId"`
	Service    string            `json:"ServiceName"`
	Resource   map[string]string `json:"ResourceAttributes"`
	Attributes map[string]string `json:"Attributes"`
	TimeNS     string            `json:"timestamp_ns"`
}

func serviceContinuationSource352(path string) bool {
	return slices.Contains([]string{"/api/explore", "/api/explore/page", "/api/discovery", "/api/inventory", "/api/metric-discovery", "/api/record-aggregates", "/api/measurements", "/api/tail"}, path)
}
func serviceContinuationCount352(calls []serviceContinuationCall352) int {
	n := 0
	for _, call := range calls {
		if serviceContinuationSource352(call.Path) {
			n++
		}
	}
	return n
}
func serviceContinuationHash352(value string) bool {
	raw, err := hex.DecodeString(value)
	return err == nil && len(raw) == 32
}

func serviceContinuationCaseCheck352(l serviceContinuationLedger352, c serviceContinuationCase352) (string, error) {
	phase := "browse"
	f := l.Fixture
	mode, count, logs := 0, "", 1
	namespace := &identity.Namespace{Mode: "missing"}
	switch c.Key {
	case "current":
		mode, count = 1, "3"
	case "effective":
		mode, count, logs = 2, "5", 2
		namespace = &identity.Namespace{Mode: "value", Value: f.ChildNS}
	case "empty":
		mode, count = 3, "3"
		namespace = &identity.Namespace{Mode: "value", Value: "empty352-" + f.Token}
	case "unreported":
		mode, count = 4, "2"
	default:
		return phase, fmt.Errorf("unknown policy case")
	}
	selector := identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: f.Environment}}
	if mode == 1 {
		selector.Policy = "current-only"
	} else if mode == 3 {
		selector.Scope = identity.EnvironmentScope{Mode: "empty"}
	} else if mode == 4 {
		selector.Scope = identity.EnvironmentScope{Mode: "unreported"}
	}
	if c.Scope.Environment != "" || c.Scope.EnvironmentSelector == nil || continuationJSON352(*c.Scope.EnvironmentSelector) != continuationJSON352(selector) || c.Scope.FromMS != f.FromMS || c.Scope.ToMS != f.ToMS || c.Scope.Namespace != nil || c.Scope.Service != nil || c.Scope.Validate() != nil {
		return phase, fmt.Errorf("discovery policy or fixed bounds changed")
	}
	if c.Evidence.Identity != "otel-service-namespace/v1" || continuationJSON352(c.Evidence.Question) != continuationJSON352(c.Scope) || !serviceContinuationHash352(c.Evidence.Basis) {
		return phase, fmt.Errorf("discovery lacks its qualified scope/evidence")
	}
	if _, err := time.Parse(time.RFC3339Nano, c.Evidence.CapturedAt); err != nil {
		return phase, fmt.Errorf("discovery lacks independent read clock")
	}
	resolved := false
	for _, edge := range c.Evidence.Dependencies {
		childNS := &identity.Namespace{Mode: "missing"}
		if edge.ChildNamespace != nil {
			childNS = &identity.Namespace{Mode: "value", Value: *edge.ChildNamespace}
		}
		if edge.Kind == "edge" && edge.Parent == f.Parent && edge.Child == f.Child && edge.ParentNamespace != nil && *edge.ParentNamespace == f.ParentNS && continuationJSON352(childNS) == continuationJSON352(namespace) && edge.Count == "1" {
			resolved = true
		}
	}
	if !resolved {
		return phase, fmt.Errorf("selected child has no corresponding reported relationship")
	}
	phase = "follow"
	where := &exploration.Predicate{All: []exploration.Predicate{{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: f.Child}, {Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "missing"}}}
	if namespace.Mode == "value" {
		where.All[1].Op, where.All[1].Value = "eq", namespace.Value
	}
	want := exploration.Query{Signal: "spans", EnvironmentSelector: &selector, FromMS: f.FromMS, ToMS: f.ToMS, Where: where, Limit: 100}
	if c.WiderQuestion.Validate() != nil || continuationJSON352(c.WiderQuestion) != continuationJSON352(want) || continuationJSON352(c.Wider.Question) != continuationJSON352(want) || c.Wider.Count != count || !serviceContinuationHash352(c.Wider.Basis) {
		return phase, fmt.Errorf("wider child lost policy/namespace/window or acquired an exact-edge profile")
	}
	if _, err := time.Parse(time.RFC3339Nano, c.Wider.CapturedAt); err != nil {
		return phase, fmt.Errorf("wider records lack read clock")
	}
	n, _ := strconv.Atoi(count)
	if len(c.Wider.Rows) != n || len(c.Oracle) != n || c.Wider.HasMore {
		return phase, fmt.Errorf("bounded fixture rows disagree with independent source count")
	}
	ids, oracle := []string{}, []string{}
	for _, raw := range c.Wider.Rows {
		var row serviceContinuationRow352
		if json.Unmarshal(raw, &row) != nil || row.ID == "" || row.Service != f.Child || !continuationEnvironment352(mode, f.Environment, row.Resource) || !continuationNamespace352(namespace, row.Resource) {
			return phase, fmt.Errorf("wider row is outside independently selected population")
		}
		timestamp, err := strconv.ParseInt(row.TimeNS, 10, 64)
		if err != nil || timestamp < f.FromMS*1000000 || timestamp >= f.ToMS*1000000 {
			return phase, fmt.Errorf("wider row is outside half-open discovery window")
		}
		ids = append(ids, row.ID)
	}
	for _, row := range c.Oracle {
		oracle = append(oracle, row.ID)
	}
	slices.Sort(ids)
	slices.Sort(oracle)
	if !slices.Equal(ids, oracle) || len(slices.Compact(slices.Clone(ids))) != n {
		return phase, fmt.Errorf("child records disagree with independent SQL IDs")
	}
	phase = "returnLocal"
	if err := serviceContinuationFactsCheck352(l, c); err != nil {
		return phase, err
	}
	phase = "preview"
	if c.View.Validate() != nil || !serviceContinuationHash352(c.View.ID) || c.View.Selector.Signal != "logs" || c.View.Selector.EnvironmentSelector == nil || continuationJSON352(*c.View.Selector.EnvironmentSelector) != continuationJSON352(selector) {
		return phase, fmt.Errorf("saved view policy changed")
	}
	savedWhere := &exploration.Predicate{All: []exploration.Predicate{{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: f.Worker}, {Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "eq", Value: f.WorkerNS}, {Field: &exploration.Field{Kind: "attribute", Key: "release"}, Op: "eq", Value: "saved-" + c.Key}}}
	if continuationJSON352(c.View.Selector.Where) != continuationJSON352(savedWhere) || len(c.LibraryList.Items) != 1 || continuationJSON352(c.LibraryList.Items[0]) != continuationJSON352(c.View) {
		return phase, fmt.Errorf("environment-scoped library injected the selected service or changed saved configuration")
	}
	u, err := url.Parse(c.LibraryURL)
	if err != nil || u.Path != "/api/activity-views" {
		return phase, fmt.Errorf("missing actual library request")
	}
	var listedSelector identity.EnvironmentSelector
	if json.Unmarshal([]byte(u.Query().Get("environmentSelector")), &listedSelector) != nil || continuationJSON352(listedSelector) != continuationJSON352(selector) || u.Query().Has("environment") {
		return phase, fmt.Errorf("library list silently projected the environment policy")
	}
	var rawSelector map[string]json.RawMessage
	if json.Unmarshal(c.StartedSelector, &rawSelector) != nil {
		return phase, fmt.Errorf("missing actual explicit Start selector")
	}
	for key := range rawSelector {
		if !slices.Contains([]string{"signal", "environment", "environmentSelector", "where"}, key) {
			return phase, fmt.Errorf("rolling view inherited fixed discovery fields")
		}
	}
	var started exploration.TailSelector
	if json.Unmarshal(c.StartedSelector, &started) != nil {
		return phase, fmt.Errorf("invalid actual explicit Start selector")
	}
	// The existing log-tail transport omits its default signal. Keep the raw
	// request above for closed-field admission; compare its checked semantics.
	if started.Signal == "" {
		started.Signal = "logs"
	}
	if continuationJSON352(started) != continuationJSON352(c.View.Selector) || len(c.Records) != logs {
		return phase, fmt.Errorf("explicit Start replaced the saved selector or wrong live population")
	}
	for _, raw := range c.Records {
		var row serviceContinuationRow352
		if json.Unmarshal(raw, &row) != nil || row.Service != f.Worker || !continuationEnvironment352(mode, f.Environment, row.Resource) || row.Resource["service.namespace"] != f.WorkerNS || row.Attributes["release"] != "saved-"+c.Key {
			return phase, fmt.Errorf("Activity records did not follow saved view's independent selector")
		}
	}
	return "", nil
}

func serviceContinuationFactsCheck352(l serviceContinuationLedger352, c serviceContinuationCase352) error {
	names := []string{"reopen-capture", "distributed-display", "wider-stage", "wider-run", "return-distributed", "library-list", "open-view", "return-service", "return-activity", "start-view"}
	if len(c.Facts) != len(names) {
		return fmt.Errorf("missing observed operation boundaries")
	}
	previous := -1
	for i, fact := range c.Facts {
		if fact.Name != names[i] || fact.CallsBefore < 0 || fact.CallsAfter < fact.CallsBefore || fact.CallsAfter > len(l.Calls) || fact.CallsBefore < previous || fact.Before != serviceContinuationCount352(l.Calls[:fact.CallsBefore]) || fact.After != serviceContinuationCount352(l.Calls[:fact.CallsAfter]) {
			return fmt.Errorf("operation fact lacks ordered raw request accounting: %s", fact.Name)
		}
		previous = fact.CallsAfter
		delta := 0
		if fact.Name == "wider-run" || fact.Name == "start-view" {
			delta = 1
		}
		if fact.After-fact.Before != delta {
			return fmt.Errorf("hidden or missing telemetry acquisition: %s", fact.Name)
		}
		for _, call := range l.Calls[fact.CallsBefore:fact.CallsAfter] {
			if serviceContinuationSource352(call.Path) && (fact.Name == "wider-run" && call.Path != "/api/explore" || fact.Name == "start-view" && call.Path != "/api/tail") {
				return fmt.Errorf("operation acquired unrelated telemetry")
			}
		}
	}
	return nil
}

func serviceContinuationLedgerCheck352(l serviceContinuationLedger352) error {
	if l.Version != "goo.serviceContinuations352/1" || len(l.Errors) != 0 || len(l.Cases) != 4 || l.Fixture.FromMS <= 0 || l.Fixture.ToMS <= l.Fixture.FromMS || l.Fixture.Child == "" || l.Fixture.Worker == "" {
		return fmt.Errorf("incomplete observed campaign")
	}
	seen := map[string]bool{}
	for _, c := range l.Cases {
		if seen[c.Key] {
			return fmt.Errorf("duplicate policy case")
		}
		seen[c.Key] = true
		if _, err := serviceContinuationCaseCheck352(l, c); err != nil {
			return fmt.Errorf("%s: %w", c.Key, err)
		}
	}
	return nil
}

// Independent admission above checks the actual report and raw request budgets.
// The existing four-step model projects browse → follow → retained return →
// saved-view preview for this observed campaign; it does not model SQL/transport
// correctness or turn a conforming prefix into proof of an unobserved completion.
func TestQuintServiceContinuationsObserved352(t *testing.T) {
	name := os.Getenv("GOO_SERVICE_CONTINUATIONS_352_REPORT")
	if name == "" {
		t.Skip("native service-continuation report opt-in")
	}
	raw, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	var ledger serviceContinuationLedger352
	if err = json.Unmarshal(raw, &ledger); err != nil {
		t.Fatal(err)
	}
	check := func(label string, input serviceContinuationLedger352, want bool) {
		t.Helper()
		admission := serviceContinuationLedgerCheck352(input)
		if (admission == nil) != want {
			t.Fatalf("%s independent admission=%v", label, admission)
		}
		judged := 0
		for _, observed := range input.Cases {
			failedPhase, caseError := serviceContinuationCaseCheck352(input, observed)
			if !want && caseError == nil {
				continue
			}
			sink := &qobs.MemorySink{}
			rec := qobs.NewRecorder(sink, qobs.WithActor("services352-"+observed.Key), qobs.WithProcess("actual-browser"))
			for _, action := range []string{"browse", "follow", "returnLocal", "preview"} {
				coherent := action != failedPhase
				rec.Record(context.Background(), action, nil, "obs.coherent", coherent)
				if !coherent {
					break // Never report later, unjudged observations after a refusal.
				}
			}
			checker, err := validate.NewChecker("model/service_browser.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
			if err != nil {
				t.Fatal(err)
			}
			report, err := checker.Check(context.Background(), sink.Steps())
			if err != nil {
				t.Fatal(err)
			}
			artifact("service-continuations352-"+label+"-"+observed.Key+".txt", []byte(report.Summary()))
			if report.OK() != want {
				t.Fatalf("%s/%s admission=%v: %s", label, observed.Key, caseError, report.Summary())
			}
			judged++
		}
		if judged == 0 || want && judged != 4 {
			t.Fatalf("%s lacks actual policy projections", label)
		}
	}
	check("actual", ledger, true)
	mutations := []func(*serviceContinuationLedger352){
		func(l *serviceContinuationLedger352) { l.Cases[0].Scope.EnvironmentSelector.Policy = "effective" },
		func(l *serviceContinuationLedger352) { l.Cases[0].WiderQuestion.ToMS++ },
		func(l *serviceContinuationLedger352) { l.Cases[0].Wider.Question.Where.All[0].Value = l.Fixture.Parent },
		func(l *serviceContinuationLedger352) { l.Cases[0].Wider.Question.Summary = "span-duration" },
		func(l *serviceContinuationLedger352) { l.Cases[0].Wider.Count = "1" },
		func(l *serviceContinuationLedger352) { l.Cases[0].Oracle[0].ID = "foreign-id" },
		func(l *serviceContinuationLedger352) {
			l.Cases[0].LibraryList.Items[0].Selector.EnvironmentSelector.Policy = "effective"
		},
		func(l *serviceContinuationLedger352) {
			var selector map[string]any
			json.Unmarshal(l.Cases[0].StartedSelector, &selector)
			selector["fromMs"] = l.Fixture.FromMS
			l.Cases[0].StartedSelector, _ = json.Marshal(selector)
		},
		func(l *serviceContinuationLedger352) { l.Cases[0].Facts[2].After++ },
		func(l *serviceContinuationLedger352) { l.Cases[0].Facts = l.Cases[0].Facts[:9] },
	}
	for i, mutate := range mutations {
		var changed serviceContinuationLedger352
		if err := json.Unmarshal(raw, &changed); err != nil {
			t.Fatal(err)
		}
		mutate(&changed)
		check(fmt.Sprintf("negative-%d", i), changed, false)
	}
}
