package verification

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
)

type typedTraceRecord340 struct {
	TraceID              string `json:"traceId"`
	SpanID               string `json:"spanId"`
	ParentSpanID         string `json:"parentSpanId"`
	Service, Operation   string
	StartNS              string `json:"startNs"`
	DurationNS           string `json:"durationNs"`
	Status, Kind         int
	Resource, Attributes map[string]string
}
type typedTraceProfileDefinition340 struct {
	Name     string
	Selector identity.EnvironmentSelector
	Expected int
}
type typedTraceFixture340 struct {
	Version, Token, Environment, Service, Operation string
	Records                                         []typedTraceRecord340
	Profiles                                        []typedTraceProfileDefinition340
}
type typedTraceProfile340 struct {
	Name            string
	ParentQuestion  exploration.Query
	Request         exploration.TraceSearchRequest
	Result          exploration.TraceSearchResult
	Events          []tsLiveEvent
	TraceQueries    []exploration.Query
	TraceID         string `json:"traceId"`
	PolicyPreserved bool
}
type typedTraceLedger340 struct {
	Fixture  typedTraceFixture340
	Profiles []typedTraceProfile340
}

func typedTraceRecords340(f typedTraceFixture340) []tsRecord {
	kinds := []string{"Unspecified", "Internal", "Server", "Client", "Producer", "Consumer"}
	statuses := []string{"Unset", "Ok", "Error"}
	states := []string{"unset", "ok", "error"}
	out := make([]tsRecord, 0, len(f.Records))
	for _, r := range f.Records {
		values := map[exploration.Field]string{}
		put := func(kind, key, value string) { values[exploration.Field{Kind: kind, Key: key}] = value }
		for key, value := range r.Resource {
			put("resource", key, value)
		}
		for key, value := range r.Attributes {
			put("attribute", key, value)
		}
		for key, value := range map[string]string{
			"trace_id": r.TraceID, "span_id": r.SpanID, "parent_span_id": r.ParentSpanID,
			"service": r.Service, "name": r.Operation, "timestamp_ns": r.StartNS, "duration_ns": r.DurationNS,
			"scope": "native.typed.trace340", "scope_version": "", "exception_event": "no", "exception_type_state": "unreported",
		} {
			put("column", key, value)
		}
		if r.Kind >= 0 && r.Kind < len(kinds) {
			put("column", "kind", kinds[r.Kind])
		}
		if r.Status >= 0 && r.Status < len(statuses) {
			put("column", "status", statuses[r.Status])
			put("column", "status_class", states[r.Status])
		}
		out = append(out, tsRecord{Values: values})
	}
	return out
}
func typedTraceProfileOK340(f typedTraceFixture340, definition typedTraceProfileDefinition340, p typedTraceProfile340) bool {
	if p.Name != definition.Name || !p.PolicyPreserved || p.Request.Validate() != nil || p.Request.Environment != "" || !aeJSONEqual(p.Request.EnvironmentSelector, definition.Selector) || !aeJSONEqual(p.Result.Request, p.Request) || !peLiveClock(p.Result.ReadAt) || !peLiveHex(p.Result.Basis) {
		return false
	}
	parent := p.ParentQuestion
	if parent.Signal != "spans" && parent.Signal != "logs" || parent.Environment != "" || !aeJSONEqual(parent.EnvironmentSelector, definition.Selector) || parent.FromMS != p.Request.FromMS || parent.ToMS != p.Request.ToMS || len(p.Request.Conditions) != 2 {
		return false
	}
	// The contextual condition inherits the complete original predicate plus the
	// observed service/operation; the added condition stays trace-wide.
	clauses := []exploration.Predicate{}
	if parent.Where != nil {
		clauses = append(clauses, *parent.Where)
	}
	clauses = append(clauses, tsLeaf("column", "service", "eq", f.Service), tsLeaf("column", "name", "eq", f.Operation))
	expected := exploration.Predicate{All: clauses}
	want, wok := tsFlatConj(expected)
	got, gok := tsFlatConj(p.Request.Conditions[0])
	if !wok || !gok || !aeJSONEqual(want, got) {
		return false
	}
	condition := p.Request.Conditions[1]
	if !(aeJSONEqual(condition, tsLeaf("column", "status", "eq", "Error")) || aeJSONEqual(condition, tsLeaf("column", "status_class", "eq", "error"))) {
		return false
	}
	items, total := tsOracle(p.Request, typedTraceRecords340(f))
	if total != strconv.Itoa(definition.Expected) || p.Result.MatchedTraces != total || p.Result.State != "observed" || p.Result.HasMore != (definition.Expected > p.Request.Limit) || !tsItemsEqual(p.Result.Items, items) {
		return false
	}
	if !tcID(p.TraceID, 32) || len(p.TraceQueries) != 2 {
		return false
	}
	found := false
	for _, item := range p.Result.Items {
		found = found || item.TraceID == p.TraceID
	}
	if !found {
		return false
	}
	signals := map[string]bool{}
	for _, q := range p.TraceQueries {
		if (q.Signal != "spans" && q.Signal != "logs") || signals[q.Signal] {
			return false
		}
		signals[q.Signal] = true
		if q.Validate() != nil || q.Environment != p.Request.Environment || !aeJSONEqual(q.EnvironmentSelector, p.Request.EnvironmentSelector) || q.FromMS != p.Request.FromMS || q.ToMS != p.Request.ToMS || q.Offset != 0 || !aeJSONEqual(q.Where, exploration.Predicate{All: []exploration.Predicate{tsLeaf("column", "trace_id", "eq", p.TraceID)}}) {
			return false
		}
	}
	return typedTraceEventsOK340(p.Events)
}
func typedTraceEventsOK340(events []tsLiveEvent) bool {
	return tsLiveEventCheck(events, []string{"parent-run", "stage", "search", "witness-details", "open-cross", "back"}, []int{1, 1, 2, 2, 4, 4})
}
func TestQuintTypedTraceSearchModel340(t *testing.T) {
	events := []tsLiveEvent{{Action: "parent-run", Source: 1}, {Action: "stage", Source: 1}, {Action: "search", Source: 2}, {Action: "witness-details", Source: 2}, {Action: "open-cross", Source: 4}, {Action: "back", Source: 4}}
	if !typedTraceEventsOK340(events) || !tsLiveReplay(t, events, "model/trace_search.yaml") {
		t.Fatal("typed return branch refused")
	}
	if typedTraceEventsOK340(events[:5]) {
		t.Fatal("unfinished typed return was accepted")
	}
	events[5].Source++
	if tsLiveReplay(t, events, "model/trace_search.yaml") {
		t.Fatal("return acquired source data")
	}
}
func typedTraceLedgerOK340(l typedTraceLedger340) bool {
	f := l.Fixture
	if f.Version != "goo.typedTraceFixture340" || f.Token == "" || f.Service == "" || f.Operation == "" || len(f.Records) < 20 || len(f.Profiles) != 4 || len(l.Profiles) != 4 {
		return false
	}
	seen := map[string]bool{}
	for i, definition := range f.Profiles {
		expected := map[string]struct {
			policy, mode string
			count        int
		}{"effective": {"effective", "named", 15}, "current": {"current-only", "named", 13}, "empty": {"effective", "empty", 2}, "unreported": {"effective", "unreported", 1}}
		want, known := expected[definition.Name]
		name := ""
		if want.mode == "named" {
			name = f.Environment
		}
		if !known || definition.Expected != want.count || !aeJSONEqual(definition.Selector, tsSelector340(want.policy, want.mode, name)) || seen[definition.Name] || definition.Selector.Validate() != nil || !typedTraceProfileOK340(f, definition, l.Profiles[i]) {
			return false
		}
		seen[definition.Name] = true
	}
	return seen["effective"] && seen["current"] && seen["empty"] && seen["unreported"]
}
func TestQuintTypedTraceSearchObserved340(t *testing.T) {
	file := os.Getenv("GOO_TYPED_TRACE_SEARCH_LEDGER")
	if file == "" {
		t.Skip("native typed trace-search ledger not selected")
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var base typedTraceLedger340
	if json.Unmarshal(raw, &base) != nil || !typedTraceLedgerOK340(base) {
		t.Fatal("typed native scope, witness, continuation or effect ledger refused")
	}
	for _, profile := range base.Profiles {
		if !tsLiveReplay(t, profile.Events, "model/trace_search.yaml") {
			t.Fatalf("typed %s lifecycle refused", profile.Name)
		}
	}
	changes := []struct {
		name   string
		mutate func(*typedTraceLedger340)
	}{
		{"winning-policy-erased", func(l *typedTraceLedger340) { l.Profiles[0].Request.EnvironmentSelector.Policy = "current-only" }},
		{"witness-current-absence", func(l *typedTraceLedger340) {
			for i, v := range l.Profiles[0].Result.Items[0].Witnesses[0].FieldValues {
				if v.Field.Key == "deployment.environment" {
					l.Profiles[0].Result.Items[0].Witnesses[0].FieldValues[i].Value = nil
				}
			}
		}},
		{"continuation-policy-weakened", func(l *typedTraceLedger340) {
			l.Profiles[0].TraceQueries[0].EnvironmentSelector.Policy = "current-only"
		}},
		{"continuation-filter-inherited", func(l *typedTraceLedger340) {
			l.Profiles[0].TraceQueries[0].Where = &l.Profiles[0].Request.Conditions[0]
		}},
		{"context-predicate-dropped", func(l *typedTraceLedger340) {
			l.Profiles[0].Request.Conditions[0] = tsLeaf("column", "service", "eq", l.Fixture.Service)
		}},
		{"population-count-replaced-by-limit", func(l *typedTraceLedger340) {
			l.Profiles[0].Result.MatchedTraces = strconv.Itoa(l.Profiles[0].Request.Limit)
		}},
		{"unfinished-return", func(l *typedTraceLedger340) { l.Profiles[0].Events = l.Profiles[0].Events[:5] }},
	}
	for _, change := range changes {
		t.Run("refuses "+change.name, func(t *testing.T) {
			var bad typedTraceLedger340
			if json.Unmarshal(raw, &bad) != nil {
				t.Fatal("bad clone")
			}
			change.mutate(&bad)
			if typedTraceLedgerOK340(bad) {
				t.Fatal("changed observed semantics accepted")
			}
		})
	}
	bad := append([]tsLiveEvent{}, base.Profiles[0].Events...)
	bad[len(bad)-1].Source++
	if tsLiveReplay(t, bad, "model/trace_search.yaml") {
		t.Fatal("return issued an unexpected source read")
	}
}
