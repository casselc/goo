package workbench

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
)

func traceSearchStep382(policy, mode, value string) Step {
	selector := &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: policy, Scope: identity.EnvironmentScope{Mode: mode, Value: value}}
	request := &exploration.TraceSearchRequest{Version: exploration.TraceSearchVersion, EnvironmentSelector: selector, FromMS: 1000, ToMS: 61000, Limit: 37, Conditions: []exploration.Predicate{
		{All: []exploration.Predicate{{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "frontend-雪"}, {Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "eq", Value: ""}}},
		{All: []exploration.Predicate{{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "database"}, {Numeric: &exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work.units"}, Type: "uint64", Unit: "{unit}", Parse: "canonical-decimal"}, Op: "gte", Value: "9007199254740993"}}},
	}}
	return Step{Version: 1, Kind: "trace-search", View: "Traces", Scope: StepScope{EnvironmentSelector: selector, FromMS: request.FromMS, ToMS: request.ToMS}, TraceSearch: request}
}

func TestTraceSearchStep382ImmutableIntentRoundTrip(t *testing.T) {
	for _, env := range []struct{ policy, mode, value string }{{"current-only", "named", "prod"}, {"effective", "named", "prod"}, {"effective", "empty", ""}, {"effective", "unreported", ""}} {
		t.Run(env.policy+"/"+env.mode, func(t *testing.T) {
			step := traceSearchStep382(env.policy, env.mode, env.value)
			if err := step.Validate(); err != nil {
				t.Fatal(err)
			}
			store := Store{Dir: t.TempDir()}
			first, err := store.CreateDocument(DocumentInput{Title: "Cross-service question", Cards: []Card{{Kind: "question", Step: &step, Key: strings.Repeat("a", 64)}}}, "", nil)
			if err != nil {
				t.Fatal(err)
			}
			before, _ := json.Marshal(first)
			step.TraceSearch.Conditions[1].All[1].Value = "18446744073709551615"
			child, err := store.CreateDocument(DocumentInput{Title: "Changed threshold", Cards: []Card{{Kind: "question", Step: &step}}}, first.ID, nil)
			if err != nil || child.Parent != first.ID {
				t.Fatal(child, err)
			}
			old, err := store.LoadDocument(first.ID)
			after, _ := json.Marshal(old)
			if err != nil || string(before) != string(after) || old.Cards[0].Step.TraceSearch.Conditions[1].All[1].Value != "9007199254740993" || old.Cards[0].Step.TraceSearch.Limit != 37 {
				t.Fatal("immutable intent changed", err)
			}
			wanted := traceSearchStep382(env.policy, env.mode, env.value)
			if !reflect.DeepEqual(*old.Cards[0].Step, wanted) {
				t.Fatal("typed scope/order/limit changed")
			}
		})
	}
}

func TestTraceSearchStep382RefusesMixedForeignAndUnsupportedIntent(t *testing.T) {
	for name, mutate := range map[string]func(*Step){
		"mixed-record": func(s *Step) { s.Record = &exploration.Query{} }, "mixed-measurement": func(s *Step) { s.Measurement = &measurements.Definition{} },
		"missing-request": func(s *Step) { s.TraceSearch = nil }, "view": func(s *Step) { s.View = "Records" }, "version": func(s *Step) { s.Version = 2 },
		"scope-window": func(s *Step) { s.Scope.ToMS++ }, "scope-ns": func(s *Step) { s.Scope.FromNS = "1000000000" },
		"environment": func(s *Step) {
			s.Scope.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "current-only", Scope: identity.EnvironmentScope{Mode: "named", Value: "other"}}
		},
		"condition-count": func(s *Step) { s.TraceSearch.Conditions = nil }, "limit": func(s *Step) { s.TraceSearch.Limit = 101 },
		"condition-field": func(s *Step) { s.TraceSearch.Conditions[0].All[0].Field.Key = "body" },
		"bad-numeric":     func(s *Step) { s.TraceSearch.Conditions[1].All[1].Value = "9007199254740993.0" },
	} {
		t.Run(name, func(t *testing.T) {
			s := traceSearchStep382("current-only", "named", "prod")
			mutate(&s)
			if s.Validate() == nil {
				t.Fatal("invalid trace-search step admitted")
			}
		})
	}
	s := traceSearchStep382("current-only", "named", "prod")
	called := false
	_, err := (Store{Dir: t.TempDir()}).CreateDocument(DocumentInput{Title: "Foreign evidence", Cards: []Card{{Kind: "question", Key: strings.Repeat("a", 64), Step: &s, Evidence: []string{strings.Repeat("b", 64)}}}}, "", func(string) error { called = true; return nil })
	if err == nil || called {
		t.Fatal("search evidence reached capture resolver", err, called)
	}
}

func TestTraceSearchStep382StrictWire(t *testing.T) {
	base, _ := json.Marshal(traceSearchStep382("current-only", "named", "prod"))
	for _, bad := range []string{
		strings.Replace(string(base), `"traceSearch":`, `"record":null,"traceSearch":`, 1),
		strings.Replace(string(base), `"scope":{`, `"scope":{"fromNS":"",`, 1),
		strings.Replace(string(base), `"scope":{`, `"scope":{"toNS":null,`, 1),
		strings.Replace(string(base), `"scope":{`, `"scope":{"unexpected":true,`, 1),
		strings.Replace(string(base), `"traceSearch":{`, `"traceSearch":{"result":{},`, 1),
		strings.Replace(string(base), `"traceSearch":{`, `"traceSearch":{"fromMS":1,`, 1),
		strings.Replace(string(base), `"fromMs":1000`, `"fromMs":null`, 1),
		strings.Replace(string(base), `"limit":37`, `"limit":null`, 1),
		strings.ReplaceAll(string(base), `"fromMs":1000,`, ``),
		strings.Replace(string(base), `"value":"frontend-雪"`, `"value":null`, 1),
		strings.Replace(string(base), `"numeric":{`, `"numeric":{"scale":null,`, 1),
		strings.Replace(string(base), `"numeric":{`, `"numeric":{"unexpected":1,`, 1),
		strings.Replace(string(base), `"field":{`, `"Field":{`, 1),
		strings.Replace(string(base), `"numeric":{`, `"numeric":{"Scale":0,`, 1),
		strings.Replace(string(base), `"field":{`, `"field":{"limit":1,`, 1),
		strings.Replace(string(base), `"field":{"kind":"column","key":"service"}`, `"field":{"kind":"resource","key":"name"},"field":{"kind":"column"}`, 1),
		strings.Replace(string(base), `"view":"Traces"`, `"view":"Traces","unexpected":1`, 1),
	} {
		var step Step
		if err := json.Unmarshal([]byte(bad), &step); err == nil && step.Validate() == nil {
			t.Fatalf("invalid wire admitted: %s", bad)
		}
	}
	var step Step
	if err := json.Unmarshal(base, &step); err != nil || step.Validate() != nil {
		t.Fatal("valid wire rejected", err)
	}
}

func largeTraceSearchStep382(valueBytes int) Step {
	step := traceSearchStep382("current-only", "named", "prod")
	step.TraceSearch.Conditions = make([]exploration.Predicate, 4)
	for index := range step.TraceSearch.Conditions {
		leaves := make([]exploration.Predicate, 15)
		for i := range leaves {
			leaves[i] = exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: strings.Repeat("k", 128)}, Op: "eq", Value: strings.Repeat("x", valueBytes)}
		}
		step.TraceSearch.Conditions[index] = exploration.Predicate{All: leaves}
	}
	return step
}
func TestTraceSearchStep382CanonicalRequestBudget(t *testing.T) {
	valid := largeTraceSearchStep382(800)
	encoded, _ := json.Marshal(valid.TraceSearch)
	if len(encoded) <= 48000 || len(encoded) > TraceSearchIntentMaxBytes || valid.Validate() != nil {
		t.Fatal("expected a legal large bounded request", len(encoded), valid.Validate())
	}
	if _, err := (Store{Dir: t.TempDir()}).CreateDocument(DocumentInput{Title: "Large search intent", Cards: []Card{{Kind: "question", Step: &valid}}}, "", nil); err != nil {
		t.Fatal(err)
	}
	oversized := largeTraceSearchStep382(1024)
	encoded, _ = json.Marshal(oversized.TraceSearch)
	if oversized.TraceSearch.Validate() != nil || len(encoded) <= TraceSearchIntentMaxBytes || oversized.Validate() == nil {
		t.Fatal("request body budget must constrain otherwise supported AST", len(encoded))
	}
	escaped := largeTraceSearchStep382(160)
	for i := range escaped.TraceSearch.Conditions {
		for j := range escaped.TraceSearch.Conditions[i].All {
			escaped.TraceSearch.Conditions[i].All[j].Value = strings.Repeat("<", 160)
		}
	}
	encoded, _ = json.Marshal(escaped.TraceSearch)
	if len(encoded) <= TraceSearchIntentMaxBytes || escaped.Validate() == nil {
		t.Fatal("Go HTML-escaped request budget was not admitted explicitly", len(encoded))
	}
}
