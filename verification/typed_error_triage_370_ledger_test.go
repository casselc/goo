package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type typedTriageCounters370 struct{ Source, Streams, Store, API, SQL int }
type typedTriageFact370 struct {
	Name          string
	Before, After typedTriageCounters370
}
type typedTriageRecord370 struct {
	ID, TraceID, Service, Namespace string
	TimeMS                          int64 `json:"timeMs"`
	Env, Attributes                 map[string]string
	Type                            *string
	Error                           bool
}
type typedTriageFixture370 struct {
	Name, Namespace, Signal, Policy, Mode, TargetService, TargetType string
	Selector                                                         exploration.Query
	Records                                                          []typedTriageRecord370
}
type typedTriageOperation370 struct {
	Group             map[string]*string
	Expected, Request exploration.Query
	Response          exploration.Result
	Status            int
	OracleIDs         []string
	Type              *string
}
type typedTriageCase370 struct {
	Fixture    typedTriageFixture370
	ReadWindow struct{ FromMS, ToMS int64 }
	Parent     struct {
		Request   exploration.Query
		Response  exploration.Result
		Status    int
		OracleIDs []string
	}
	RawRows                   []struct{ ID, TimeMS string }
	ParentBefore, ParentAfter json.RawMessage
	Inspect, Compare          typedTriageOperation370
	Inspections               []typedTriageOperation370
	Pending                   *struct {
		Request                           exploration.Query
		Response                          exploration.Result
		Status, AbortsBefore, AbortsAfter int
		BeforeRelease, AfterRelease       json.RawMessage
	}
	OldActions    struct{ Before, After json.RawMessage }
	HiddenActions *struct {
		Before, After                     json.RawMessage
		SourceConnected, DefaultPrevented *bool
	}
	NarrowAction *struct {
		Gesture                string
		Before, After          typedTriageCounters370
		Question               exploration.Query
		Focused, CenterVisible bool
		Rect                   struct{ Top, Left, Right, Bottom float64 }
		Viewport               struct{ Width, Height float64 }
	}
	Facts []typedTriageFact370
}
type typedTriageReport370 struct {
	Version string
	Cases   []typedTriageCase370
	Errors  []string
}

// This oracle uses source facts, not the UI query builders or the production
// environment matcher. In particular a reported empty current value masks legacy.
func typedTriageEnvironment370(q exploration.Query, resource map[string]string) bool {
	s := q.EnvironmentSelector
	if s == nil || s.Validate() != nil {
		return false
	}
	current, c := resource["deployment.environment.name"]
	legacy, l := resource["deployment.environment"]
	if s.Policy == "current-only" {
		return c && current == s.Scope.Value
	}
	if !c && !l {
		return s.Scope.Mode == "unreported"
	}
	value := legacy
	if c {
		value = current
	}
	return s.Scope.Mode == "empty" && value == "" || s.Scope.Mode == "named" && value == s.Scope.Value
}

// Reuse the existing independent exception/severity oracle, feeding it the raw
// OTLP fixture facts. These are not the production query evaluator or UI builder.
func typedTriageRaw370(r typedTriageRecord370, signal string) etRecord {
	attrs, resource := maps.Clone(r.Attributes), maps.Clone(r.Env)
	if attrs == nil {
		attrs = map[string]string{}
	}
	if resource == nil {
		resource = map[string]string{}
	}
	resource["service.namespace"] = r.Namespace
	resource["service.name"] = r.Service
	status, severity, number := "Ok", "INFO", "9"
	events := []etEvent{}
	if r.Error {
		status, severity, number = "Error", "ERROR", "17"
		if signal == "spans" {
			event := map[string]string{}
			if r.Type != nil {
				event["exception.type"] = *r.Type
			}
			events = append(events, etEvent{Name: "exception", Attributes: event})
		}
	}
	if signal == "logs" && r.Type != nil {
		attrs["exception.type"] = *r.Type
	}
	return etRecord{Signal: signal, ID: r.ID, TraceID: r.TraceID, Service: r.Service, TimeMS: r.TimeMS, Attributes: attrs, Resource: resource, Status: status, Severity: severity, SeverityNumber: number, Events: events}
}
func typedTriageValue370(r typedTriageRecord370, signal string, field exploration.Field) (*string, error) {
	return etField(field, typedTriageRaw370(r, signal))
}
func typedTriageMembers370(q exploration.Query, f typedTriageFixture370) ([]typedTriageRecord370, error) {
	if q.Validate() != nil || q.EnvironmentSelector == nil || q.Signal != f.Signal {
		return nil, fmt.Errorf("invalid typed fixture question")
	}
	rows := []typedTriageRecord370{}
	for _, r := range f.Records {
		if !typedTriageEnvironment370(q, r.Env) || r.TimeMS < q.FromMS || r.TimeMS >= q.ToMS {
			continue
		}
		yes, e := etMatch(q.Where, typedTriageRaw370(r, q.Signal))
		if e != nil {
			return nil, e
		}
		if yes {
			rows = append(rows, r)
		}
	}
	return rows, nil
}
func typedTriageResult370(r exploration.Result, q exploration.Query, f typedTriageFixture370) bool {
	if !nhLiveSameQuery(r.Question, q) || r.Basis == "" || r.CapturedAt == "" || q.Offset != 0 || q.Limit != 100 {
		return false
	}
	rows, e := typedTriageMembers370(q, f)
	if e != nil || r.Count != strconv.Itoa(len(rows)) {
		return false
	}
	expectedIDs := []string{}
	expectedGroups := map[string]int{}
	for _, row := range rows {
		expectedIDs = append(expectedIDs, row.ID)
		values := []*string{}
		for _, field := range q.GroupBy {
			v, e := typedTriageValue370(row, q.Signal, field)
			if e != nil {
				return false
			}
			values = append(values, v)
		}
		key, e := json.Marshal(values)
		if e != nil {
			return false
		}
		expectedGroups[string(key)]++
	}
	slices.Sort(expectedIDs)
	got, e := reportedResultIDs353(r)
	if e != nil || !slices.Equal(got, expectedIDs) {
		return false
	}
	if len(q.GroupBy) > 0 {
		if len(r.Groups) != len(expectedGroups) {
			return false
		}
		seen := map[string]bool{}
		for _, raw := range r.Groups {
			var group map[string]json.RawMessage
			if json.Unmarshal(raw, &group) != nil {
				return false
			}
			values := []*string{}
			for i := range q.GroupBy {
				raw, ok := group["g"+strconv.Itoa(i)]
				if !ok {
					return false
				}
				var v *string
				if json.Unmarshal(raw, &v) != nil {
					return false
				}
				values = append(values, v)
			}
			key, _ := json.Marshal(values)
			var count string
			if json.Unmarshal(group["count"], &count) != nil || count != strconv.Itoa(expectedGroups[string(key)]) || seen[string(key)] {
				return false
			}
			seen[string(key)] = true
		}
	}
	return len(rows) == 0 && r.State == "no_data" || len(rows) > 0 && r.State == "observed"
}

// The independently executed raw query has only environment, namespace and
// time bounds. Admit its exact ID/time multiset before considering child scopes.
func typedTriageRawAdmission370(c typedTriageCase370) bool {
	q, f := c.Parent.Request, c.Fixture
	want := []string{}
	for _, r := range f.Records {
		if typedTriageEnvironment370(q, r.Env) && r.Namespace == f.Namespace && r.TimeMS >= q.FromMS && r.TimeMS < q.ToMS {
			want = append(want, r.ID+":"+strconv.FormatInt(r.TimeMS, 10))
		}
	}
	got := []string{}
	for _, r := range c.RawRows {
		at, e := strconv.ParseInt(r.TimeMS, 10, 64)
		if e != nil || strconv.FormatInt(at, 10) != r.TimeMS || r.ID == "" {
			return false
		}
		got = append(got, r.ID+":"+r.TimeMS)
	}
	slices.Sort(want)
	slices.Sort(got)
	if len(want) == 0 || !slices.Equal(want, got) {
		return false
	}
	members, e := typedTriageMembers370(q, f)
	if e != nil {
		return false
	}
	ids := []string{}
	for _, r := range members {
		ids = append(ids, r.ID)
	}
	slices.Sort(ids)
	oracle := slices.Clone(c.Parent.OracleIDs)
	slices.Sort(oracle)
	return slices.Equal(ids, oracle)
}
func typedTriageChild370(parent exploration.Query, group map[string]*string, inspect bool) (exploration.Query, bool) {
	if len(parent.GroupBy) != 2 || parent.GroupBy[0] != (exploration.Field{Kind: "column", Key: "service"}) {
		return exploration.Query{}, false
	}
	service, ok := group["g0"]
	if !ok {
		return exploration.Query{}, false
	}
	typ, ok := group["g1"]
	if !ok {
		return exploration.Query{}, false
	}
	raw, e := json.Marshal(parent)
	if e != nil {
		return exploration.Query{}, false
	}
	var q exploration.Query
	if json.Unmarshal(raw, &q) != nil {
		return q, false
	}
	all := []exploration.Predicate{}
	if q.Where != nil {
		if len(q.Where.All) > 0 {
			all = append(all, q.Where.All...)
		} else {
			all = append(all, *q.Where)
		}
	}
	clause := func(field exploration.Field, v *string) exploration.Predicate {
		p := exploration.Predicate{Field: &field, Op: "missing"}
		if v != nil {
			p.Op = "eq"
			p.Value = *v
		}
		return p
	}
	all = append(all, clause(parent.GroupBy[0], service))
	if inspect {
		all = append(all, clause(parent.GroupBy[1], typ))
	}
	q.Where = &exploration.Predicate{All: all}
	q.GroupBy = []exploration.Field{parent.GroupBy[1]}
	q.Offset = 0
	q.GroupTimeSeries = false
	return q, q.Validate() == nil
}
func typedTriageOperationAdmission370(op typedTriageOperation370, parent exploration.Result, f typedTriageFixture370, inspect bool) bool {
	expected, ok := typedTriageChild370(parent.Question, op.Group, inspect)
	if !ok || op.Status != 200 || !nhLiveSameQuery(expected, op.Expected) || !nhLiveSameQuery(expected, op.Request) || !typedTriageResult370(op.Response, expected, f) {
		return false
	}
	actualGroup := false
	for _, raw := range parent.Groups {
		var g map[string]json.RawMessage
		if json.Unmarshal(raw, &g) != nil {
			return false
		}
		var a, b *string
		if json.Unmarshal(g["g0"], &a) != nil || json.Unmarshal(g["g1"], &b) != nil {
			return false
		}
		if savedJSON360([]any{a, b}) == savedJSON360([]any{op.Group["g0"], op.Group["g1"]}) {
			actualGroup = true
		}
	}
	if !actualGroup {
		return false
	}
	rows, e := typedTriageMembers370(expected, f)
	if e != nil {
		return false
	}
	ids := []string{}
	for _, r := range rows {
		ids = append(ids, r.ID)
	}
	slices.Sort(ids)
	oracle := slices.Clone(op.OracleIDs)
	slices.Sort(oracle)
	return len(ids) > 0 && slices.Equal(ids, oracle)
}
func typedTriageCaseAdmission370(c typedTriageCase370, requirePending bool) bool {
	f := c.Fixture
	q := c.Parent.Request
	s := q.EnvironmentSelector
	if f.Name == "" || f.Namespace == "" || s == nil || f.Selector.EnvironmentSelector == nil || savedJSON360(s) != savedJSON360(f.Selector.EnvironmentSelector) || s.Policy != f.Policy || s.Scope.Mode != f.Mode || q.Signal != f.Signal || f.Selector.Signal != f.Signal || q.FromMS != c.ReadWindow.FromMS || q.ToMS != c.ReadWindow.ToMS || c.Parent.Status != 200 || !typedTriageResult370(c.Parent.Response, q, f) || !typedTriageRawAdmission370(c) {
		return false
	}
	if q.Where == nil || len(q.Where.All) < 2 || savedJSON360(q.Where.All[0]) != savedJSON360(f.Selector.Where) {
		return false
	}
	if len(c.ParentBefore) == 0 || savedJSON360(c.ParentBefore) != savedJSON360(c.ParentAfter) || len(c.OldActions.Before) == 0 || savedJSON360(c.OldActions.Before) != savedJSON360(c.OldActions.After) {
		return false
	}
	if len(c.Inspections) != 3 || !typedTriageOperationAdmission370(c.Compare, c.Parent.Response, f, false) {
		return false
	}
	types := map[string]bool{}
	for _, op := range c.Inspections {
		if !typedTriageOperationAdmission370(op, c.Parent.Response, f, true) {
			return false
		}
		k := savedJSON360(op.Group["g1"])
		if types[k] {
			return false
		}
		types[k] = true
	}
	if !types["null"] || !types[`""`] || !types[savedJSON360(f.TargetType)] || !typedTriageOperationAdmission370(c.Inspect, c.Parent.Response, f, true) {
		return false
	}
	if requirePending && (c.Pending == nil || c.HiddenActions == nil || c.NarrowAction == nil) {
		return false
	}
	if c.HiddenActions != nil {
		h := c.HiddenActions
		if h.SourceConnected == nil || *h.SourceConnected || h.DefaultPrevented == nil || !*h.DefaultPrevented || len(h.Before) == 0 || savedJSON360(h.Before) != savedJSON360(h.After) {
			return false
		}
	}
	if c.NarrowAction != nil {
		n := c.NarrowAction
		if n.Gesture != "Focus current Inspect action after viewport resize" || !n.Focused || !n.CenterVisible || n.Before != n.After || !nhLiveSameQuery(n.Question, c.Inspect.Request) || n.Viewport.Width != 390 || n.Viewport.Height != 844 || n.Rect.Top < 0 || n.Rect.Left < 0 || n.Rect.Right > n.Viewport.Width || n.Rect.Bottom > n.Viewport.Height || n.Rect.Right <= n.Rect.Left || n.Rect.Bottom <= n.Rect.Top {
			return false
		}
	}
	if c.Pending != nil {
		p := c.Pending
		if p.Status != 200 || p.AbortsAfter <= p.AbortsBefore || !nhLiveSameQuery(p.Request, c.Inspect.Request) || !typedTriageResult370(p.Response, p.Request, f) || len(p.BeforeRelease) == 0 || savedJSON360(p.BeforeRelease) != savedJSON360(p.AfterRelease) {
			return false
		}
	}
	counts := map[string]int{}
	for i, fact := range c.Facts {
		counts[fact.Name]++
		delta := 0
		switch fact.Name {
		case "parentrun", "inspectrun", "comparerun", "pendingrun":
			delta = 1
		case "parentstage", "inspectstage", "inspectback", "comparestage", "compareback", "oldactions", "hiddenactions", "pendingstage", "pendingback", "laterelease":
		default:
			return false
		}
		if fact.After.Source-fact.Before.Source != delta || fact.After.Store != fact.Before.Store || fact.After.Streams != fact.Before.Streams || fact.After.SQL < fact.Before.SQL || delta == 0 && fact.After.SQL != fact.Before.SQL || fact.After.API < fact.Before.API {
			return false
		}
		if i > 0 {
			prior := c.Facts[i-1].After
			if fact.Before.Source != prior.Source || fact.Before.Store != prior.Store || fact.Before.Streams != prior.Streams || fact.Before.API < prior.API || fact.Before.SQL < prior.SQL {
				return false
			}
		}
	}
	for name, n := range map[string]int{"parentstage": 1, "parentrun": 1, "inspectstage": 3, "inspectrun": 3, "inspectback": 3, "comparestage": 1, "comparerun": 1, "compareback": 1, "oldactions": 1} {
		if counts[name] != n {
			return false
		}
	}
	hiddenN := 0
	if c.HiddenActions != nil {
		hiddenN = 1
	}
	if counts["hiddenactions"] != hiddenN {
		return false
	}
	pendingN := 0
	if c.Pending != nil {
		pendingN = 1
	}
	for _, name := range []string{"pendingstage", "pendingrun", "pendingback", "laterelease"} {
		if counts[name] != pendingN {
			return false
		}
	}
	return true
}

// Replay follows the actual recorded boundary names and observed cumulative read
// counters. Navigation depth is the logical depth of those observed actions, not
// a claimed intermediate DOM snapshot. Scope/membership/held-state admission is
// independent of the existing model's read/depth safety check.
func typedTriageReplay370(t *testing.T, label string, facts []typedTriageFact370) bool {
	t.Helper()
	if len(facts) == 0 {
		t.Fatal("empty typed triage observations")
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	baseline := facts[0].Before.Source
	depth := 0
	for _, fact := range facts {
		action := "history"
		switch fact.Name {
		case "parentstage", "inspectstage", "comparestage", "pendingstage":
			action = "stage"
			depth++
		case "parentrun", "inspectrun", "comparerun", "pendingrun":
			action = "read"
		case "inspectback", "compareback", "pendingback":
			action = "back"
			depth--
		case "oldactions", "hiddenactions", "laterelease":
		default:
			t.Fatal("unknown typed triage observation", fact.Name)
		}
		rec.Record(context.Background(), action, nil, "obs.reads", fact.After.Source-baseline, "obs.depth", depth, "obs.coherent", true)
	}
	encoded, e := json.Marshal(sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	if e = artifact("typed-triage370-"+label+"-steps.json", encoded); e != nil {
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
	if e = artifact("typed-triage370-"+label+"-model.txt", []byte(report.Summary())); e != nil {
		t.Fatal(e)
	}
	return report.OK()
}
func TestQuintTypedErrorTriageObserved370(t *testing.T) {
	path := os.Getenv("GOO_TYPED_ERROR_TRIAGE_370_REPORT")
	if path == "" {
		t.Skip("actual typed error-triage report opt-in")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var r typedTriageReport370
	if json.Unmarshal(raw, &r) != nil || r.Version != "goo.typedErrorTriage370/1" || len(r.Errors) > 0 || len(r.Cases) != 4 {
		t.Fatal("missing actual typed triage report")
	}
	seen := map[string]bool{}
	for i, c := range r.Cases {
		if !typedTriageCaseAdmission370(c, i == 0) {
			t.Fatalf("typed triage case %d scope/member/owner/source admission", i)
		}
		scope := c.Fixture.Policy + ":" + c.Fixture.Mode
		if seen[scope] {
			t.Fatal("duplicate typed policy coverage")
		}
		seen[scope] = true
		if !typedTriageReplay370(t, strconv.Itoa(i), c.Facts) {
			t.Fatal("actual typed triage read/return lifecycle", i)
		}
	}
	for _, scope := range []string{"current-only:named", "effective:named", "effective:empty", "effective:unreported"} {
		if !seen[scope] {
			t.Fatal("missing typed policy", scope)
		}
	}
	first := r.Cases[0]
	bad := first
	bad.ParentAfter = json.RawMessage(`{"changed":true}`)
	if typedTriageCaseAdmission370(bad, true) {
		t.Fatal("replaced held parent admitted")
	}
	bad = first
	raw, _ = json.Marshal(first.Inspect)
	if json.Unmarshal(raw, &bad.Inspect) != nil {
		t.Fatal("copy mutation")
	}
	bad.Inspect.Request.ToMS++
	if typedTriageCaseAdmission370(bad, true) {
		t.Fatal("changed typed child scope admitted")
	}
	bad = first
	bad.RawRows = slices.Clone(first.RawRows)
	bad.RawRows[0].TimeMS = "0"
	if typedTriageCaseAdmission370(bad, true) {
		t.Fatal("changed independent raw source occurrence admitted")
	}
	bad = first
	hidden := *first.HiddenActions
	refused := false
	hidden.DefaultPrevented = &refused
	bad.HiddenActions = &hidden
	if typedTriageCaseAdmission370(bad, true) {
		t.Fatal("unrefused hidden source action admitted")
	}
	counter := slices.Clone(first.Facts)
	counter[0].After.Source++
	if typedTriageReplay370(t, "bad-stage-read", counter) {
		t.Fatal("model accepted Stage source read")
	}
	order := slices.Clone(first.Facts)
	back := slices.IndexFunc(order, func(f typedTriageFact370) bool { return f.Name == "inspectback" })
	if back < 0 {
		t.Fatal("missing actual Back observation")
	}
	order[0], order[back] = order[back], order[0]
	if typedTriageReplay370(t, "bad-return-order", order) {
		t.Fatal("model accepted Return before Stage")
	}
}
