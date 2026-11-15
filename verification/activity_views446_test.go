package verification

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

func activity446Clone(t *testing.T, value workbench.ActivityView) workbench.ActivityView {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var clone workbench.ActivityView
	if err := json.Unmarshal(raw, &clone); err != nil {
		t.Fatal(err)
	}
	return clone
}

// Store/hash properties use real definitions, not abstract presentation booleans.
func TestHegelPairedActivityViews446Identity(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		n := hegel.Draw(tc, hegel.Integers(0, 100000))
		kind := hegel.Draw(tc, hegel.Integers(0, 4))
		value := []string{"", "prod", "all", "unknown", "team/\"quoted\"", "β"}[n%6]
		selector := workbench.ActivitySelector{Signal: "both", Environment: fmt.Sprintf("prod-%d", n), Where: &exploration.Predicate{Any: []exploration.Predicate{
			{Field: &exploration.Field{Kind: "resource", Key: "team"}, Op: "missing"},
			{Field: &exploration.Field{Kind: "attribute", Key: "status"}, Op: "eq", Value: value},
		}}}
		if kind > 0 {
			selector.Environment = ""
			selector.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: fmt.Sprintf("prod-%d", n)}}
			if kind == 1 {
				selector.EnvironmentSelector.Policy = "current-only"
			}
			if kind == 3 {
				selector.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "empty"}
			}
			if kind == 4 {
				selector.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "unreported"}
			}
		}
		v := workbench.ActivityView{Version: 1, Title: fmt.Sprintf("Paired %d", n), Selector: selector, Lanes: &workbench.ActivityLanes{
			Logs:  workbench.ActivityLayout{Columns: []string{"summary", "utc"}, FieldColumns: []exploration.Field{{Kind: "resource", Key: "service.namespace"}}},
			Spans: workbench.ActivityLayout{Columns: []string{"service", "summary", "utc"}, FieldColumns: []exploration.Field{{Kind: "attribute", Key: "status"}}},
		}}
		source := activity446Clone(t, v)
		store := workbench.Store{Dir: t.TempDir()}
		saved, err := store.CreateActivityView(v)
		if err != nil {
			t.Fatal(err)
		}
		repeated, err := store.CreateActivityView(v)
		if err != nil || repeated.ID != saved.ID {
			t.Fatal("unstable paired definition", err)
		}
		got := activity446Clone(t, *saved)
		got.ID = ""
		if !reflect.DeepEqual(got, source) {
			t.Fatal("paired definition changed during publication")
		}
		queries, err := saved.Selector.Queries()
		if err != nil || len(queries) != 2 || queries[0].Signal != "logs" || queries[1].Signal != "spans" {
			t.Fatal("pair did not project to two physical queries", err)
		}
		for _, q := range queries {
			if !reflect.DeepEqual(q.Where, source.Selector.Where) {
				t.Fatal("shared filter lost a branch")
			}
		}
		queries[0].Where.Any[0].Field.Key = "mutated-query"
		if !reflect.DeepEqual(queries[1].Where, source.Selector.Where) || !reflect.DeepEqual(saved.Selector.Where, source.Selector.Where) {
			t.Fatal("a returned query mutated the saved selector or sibling predicate")
		}
		if exploration.TailSelector(saved.Selector).Query(60000).Validate() == nil {
			t.Fatal("mixed selector became an executable physical query")
		}
		selection, err := queries[0].EnvironmentSelection()
		if err != nil {
			t.Fatal(err)
		}
		page, err := store.ActivityViewsForEnvironment("", &selection)
		if err != nil || len(page.Items) != 1 || page.Items[0].ID != saved.ID {
			t.Fatal("catalog lost exact paired scope", err)
		}
		// Caller and returned values cannot mutate a published definition.
		v.Selector.Where.Any[0].Field.Key = "caller-mutated"
		v.Lanes.Logs.Columns[0] = "service"
		saved.Lanes.Spans.FieldColumns[0].Key = "returned-mutated"
		stable, err := store.LoadActivityView(repeated.ID)
		if err != nil {
			t.Fatal(err)
		}
		stable.ID = ""
		if !reflect.DeepEqual(*stable, source) {
			t.Fatal("published definition retained mutable caller/return storage")
		}
		for _, mutate := range []func(*workbench.ActivityView){
			func(x *workbench.ActivityView) {
				mode := "unreported"
				if x.Selector.EnvironmentSelector != nil && x.Selector.EnvironmentSelector.Scope.Mode == "unreported" {
					mode = "empty"
				}
				x.Selector.Environment = ""
				x.Selector.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: mode}}
			},
			func(x *workbench.ActivityView) { x.Lanes.Logs.Columns = []string{"utc", "summary"} },
			func(x *workbench.ActivityView) { x.Lanes.Spans.Columns = []string{"summary", "service", "utc"} },
			func(x *workbench.ActivityView) { x.Lanes.Logs.FieldColumns[0].Key = "team" },
			func(x *workbench.ActivityView) { x.Selector.Where.Any[1].Value = value + "-other" },
		} {
			changed := activity446Clone(t, source)
			mutate(&changed)
			altered, err := store.CreateActivityView(changed)
			if err != nil || altered.ID == repeated.ID {
				t.Fatal("presentation or complete predicate did not bind content identity", err)
			}
		}
	}, options(60)...)
}

type activity446Recall struct {
	TailPosts    *int           `json:"tailPosts"`
	SourceSQL    *int           `json:"sourceSQL"`
	StartVisible *bool          `json:"startVisible"`
	Rows         map[string]int `json:"rows"`
}
type activity446Report struct {
	Version, Status string
	Cleanup         bool
	Errors          []string
	Saved, Recalled workbench.ActivityView
	Recall          activity446Recall
	Steps           []paired444Step
}

func activity446ID(v workbench.ActivityView) string {
	v.ID = ""
	raw, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", sha256.Sum256(raw))
}

func activity446DefinitionHeld(report activity446Report) bool {
	if report.Version != "goo.activity-views446/1" || report.Status != "passed" || !report.Cleanup || len(report.Errors) != 0 {
		return false
	}
	if report.Saved.ID == "" || report.Saved.ID != activity446ID(report.Saved) || report.Saved.Validate() != nil || report.Recalled.Validate() != nil || report.Saved.Selector.Signal != "both" || !reflect.DeepEqual(report.Saved, report.Recalled) {
		return false
	}
	recall := report.Recall
	if recall.TailPosts == nil || *recall.TailPosts != 0 || recall.SourceSQL == nil || *recall.SourceSQL != 0 || recall.StartVisible == nil || !*recall.StartVisible {
		return false
	}
	for _, signal := range []string{"logs", "spans"} {
		rows, present := recall.Rows[signal]
		if !present || rows != 0 {
			return false
		}
	}
	seen := map[string]bool{}
	for _, step := range report.Steps {
		seen[step.Action] = true
		if step.Action == "start" {
			queries, err := report.Saved.Selector.Queries()
			if err != nil {
				return false
			}
			for _, expected := range queries {
				var physical exploration.TailSelector
				lane, present := step.After.Lanes[expected.Signal]
				if !present || json.Unmarshal(lane.Selector, &physical) != nil {
					return false
				}
				actual := physical.Query(60000)
				if actual.Validate() != nil || actual.Signal != expected.Signal || !actual.SameEnvironment(expected) || !reflect.DeepEqual(actual.Where, expected.Where) {
					return false
				}
			}
		}
	}
	for _, action := range []string{"start", "refuse", "inspect", "navigate", "return"} {
		if !seen[action] {
			return false
		}
	}
	return true
}

// The existing paired transport/retained-custody model remains the behavioral
// model; saved configuration gets an independent concrete admission check.
func TestQuintActivityViews446Observed(t *testing.T) {
	file := os.Getenv("GOO_ACTIVITY_VIEWS446_REPORT")
	if file == "" {
		t.Skip("requires actual saved paired Activity browser report")
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var report activity446Report
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	checker := paired444Checker(t)
	check := func(r activity446Report) bool {
		if !activity446DefinitionHeld(r) {
			return false
		}
		return paired444Replay(t, checker, paired444Report{Version: "goo.paired444/1", Status: r.Status, Cleanup: r.Cleanup, Errors: r.Errors, Steps: r.Steps})
	}
	if !check(report) {
		t.Fatal("actual saved paired Activity trace rejected")
	}
	for _, kind := range []string{"recall-read", "recall-sql", "recall-rows", "missing-recall-count", "recalled-layout", "recalled-selector", "forged-id", "start-selector", "return-read", "return-keys"} {
		t.Run(kind, func(t *testing.T) {
			var bad activity446Report
			if err := json.Unmarshal(raw, &bad); err != nil {
				t.Fatal(err)
			}
			changed := true
			switch kind {
			case "recall-read":
				*bad.Recall.TailPosts = 1
			case "recall-sql":
				*bad.Recall.SourceSQL = 1
			case "recall-rows":
				bad.Recall.Rows["spans"]++
			case "missing-recall-count":
				bad.Recall.TailPosts = nil
			case "recalled-layout":
				bad.Recalled.Lanes.Spans.Columns = []string{"utc", "summary"}
				if slices.Equal(bad.Saved.Lanes.Spans.Columns, bad.Recalled.Lanes.Spans.Columns) {
					bad.Recalled.Lanes.Spans.Columns = []string{"summary", "utc"}
				}
			case "forged-id":
				bad.Saved.ID = strings.Repeat("a", 64)
				bad.Recalled.ID = bad.Saved.ID
			case "start-selector":
				changed = false
				for i := range bad.Steps {
					if bad.Steps[i].Action == "start" {
						lane := bad.Steps[i].After.Lanes["spans"]
						lane.Selector = json.RawMessage(`{"signal":"logs","environment":"wrong"}`)
						bad.Steps[i].After.Lanes["spans"] = lane
						changed = true
						break
					}
				}
			case "recalled-selector":
				bad.Recalled.Selector.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "resource", Key: "invented"}, Op: "exists"}
			default:
				changed = false
				for i := range bad.Steps {
					s := &bad.Steps[i]
					if s.Action != "return" {
						continue
					}
					if strings.HasSuffix(kind, "read") {
						s.After.Reads["logs"]++
					} else {
						lane := s.After.Lanes["spans"]
						lane.Keys = append(lane.Keys, "invented-return-occurrence")
						s.After.Lanes["spans"] = lane
					}
					changed = true
					break
				}
			}
			if !changed {
				t.Fatal("control lacked an exercised target")
			}
			if check(bad) {
				t.Fatal("altered saved/retained definition accepted", kind)
			}
		})
	}
}
