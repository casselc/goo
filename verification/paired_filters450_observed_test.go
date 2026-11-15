package verification

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
)

type filters450Row struct {
	ID, AtNS, Service, SeverityNumber, DurationNS, Kind string
	Resource, Attributes                                map[string]string
}
type filters450Step struct {
	paired444Step
	Local, Physical     bool
	SQLBefore, SQLAfter *int
	Posts               []string
}
type filters450Report struct {
	Source []struct {
		SQL    string
		Status *int
	}
	Version, Status string
	Cleanup         bool
	Errors          []string
	Fixture         struct{ Token, Environment, Service string }
	Shared          exploration.Predicate
	Conditions      map[string]*exploration.Predicate
	Steps           []filters450Step
	Oracle          map[string][]filters450Row
	Memberships     []struct {
		Name, Signal, Mode string
		Selector           exploration.TailSelector
		Window             struct{ FromMS, ToMS int64 }
		IDs, OracleIDs     []string
	}
	Saved, Recalled workbench.ActivityView
	Recall          struct {
		SourcePosts, SQL int
		Paused           bool
		Rows             map[string]int
	}
	Bombadil struct {
		Snapshots int
		Completed struct {
			Version                                          string
			Phase, LogsReads, SpansReads, Forbidden          int
			SiblingHeld, LogApplied, LocalReadFree, Overflow bool
			LogsPhase, SpansPhase                            string
		}
	}
}

func filters450Conditions() (exploration.Predicate, exploration.Predicate) {
	leaf := func(kind, key, op, value string) exploration.Predicate {
		return exploration.Predicate{Field: &exploration.Field{Kind: kind, Key: key}, Op: op, Value: value}
	}
	return exploration.Predicate{All: []exploration.Predicate{leaf("column", "severity_number", "gte", "17"), {Any: []exploration.Predicate{leaf("attribute", "tenant", "missing", ""), leaf("attribute", "tenant", "eq", "")}}}}, exploration.Predicate{All: []exploration.Predicate{leaf("column", "duration_ns", "gte", "100000000"), leaf("column", "kind", "eq", "Server")}}
}
func filters450Admission(r filters450Report) string {
	if r.Version != "goo.pairedFilters450/1" || r.Status != "passed" || !r.Cleanup || len(r.Errors) != 0 || r.Fixture.Token == "" || r.Fixture.Environment == "" || r.Fixture.Service == "" {
		return "unfinished run"
	}
	logCondition, spanCondition := filters450Conditions()
	shared := exploration.Predicate{All: []exploration.Predicate{{Field: &exploration.Field{Kind: "resource", Key: "fixture.bank"}, Op: "eq", Value: r.Fixture.Token}, {Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: r.Fixture.Service}}}
	if !reflect.DeepEqual(r.Shared, shared) || !reflect.DeepEqual(r.Conditions["logs"], &logCondition) || !reflect.DeepEqual(r.Conditions["spans"], &spanCondition) {
		return "fixture predicates changed"
	}
	if r.Saved.Validate() != nil || r.Saved.Selector.Signal != "both" || r.Saved.Selector.Environment != r.Fixture.Environment || !reflect.DeepEqual(r.Saved.Selector.Where, &shared) || r.Saved.Lanes == nil || !reflect.DeepEqual(r.Saved.Lanes.Logs.Where, &logCondition) || !reflect.DeepEqual(r.Saved.Lanes.Spans.Where, &spanCondition) || !reflect.DeepEqual(r.Saved, r.Recalled) {
		return "saved lane condition differs"
	}
	if len(r.Recall.Rows) != 2 || !r.Recall.Paused || r.Recall.SourcePosts != 0 || r.Recall.SQL != 0 || r.Recall.Rows["logs"] != 0 || r.Recall.Rows["spans"] != 0 {
		return "recall acquired data"
	}
	if len(r.Source) == 0 {
		return "missing physical source ledger"
	}
	for _, call := range r.Source {
		if call.SQL == "" || call.Status == nil || *call.Status != 200 {
			return "incomplete physical source witness"
		}
	}
	required := map[string]bool{"start-shared-pair": false, "active-log-draft": false, "draft-error-logs": false, "apply-error-logs": false, "draft-missing-or-empty": false, "apply-missing-or-empty": false, "draft-slow-server-spans": false, "apply-slow-server-spans": false, "resume-filtered-logs": false, "save-lane-conditions": false, "dirty-lane-draft": false, "recall-paused-pair": false, "start-recalled-filters": false}
	for _, s := range r.Steps {
		if _, ok := required[s.Name]; ok {
			required[s.Name] = true
		}
		if s.SQLBefore == nil || s.SQLAfter == nil || *s.SQLBefore < 0 || *s.SQLAfter < *s.SQLBefore {
			return "missing physical source counters"
		}
		if *s.SQLAfter > len(r.Source) {
			return "source counter exceeds independent ledger"
		}
		if s.Physical && *s.SQLBefore != *s.SQLAfter {
			return "local configuration read physical source"
		}
		if s.Local && !reflect.DeepEqual(s.Before.Reads, s.After.Reads) {
			return "draft or recall opened a source connection"
		}
		for _, p := range s.Posts {
			if p != "/api/tail" && p != "/api/cel/context" && p != "/api/cel/plan" && p != "/api/activity-views" && !histogram447Cancel.MatchString(p) {
				return "unrelated source or store API"
			}
		}
		if s.Action == "resume" {
			peer := "logs"
			if s.Lane == "logs" {
				peer = "spans"
			}
			if s.After.Reads[s.Lane]-s.Before.Reads[s.Lane] != 1 || s.After.Reads[peer] != s.Before.Reads[peer] || !reflect.DeepEqual(s.Before.Lanes[peer], s.After.Lanes[peer]) {
				return "applying one lane changed its sibling"
			}
		}
		if s.Action == "draft" {
			for _, signal := range []string{"logs", "spans"} {
				if !paired444MetadataHeld(s.Before.Lanes[signal], s.After.Lanes[signal], true) || s.Before.Lanes[signal].Generation != s.After.Lanes[signal].Generation {
					return "active draft changed existing source owner"
				}
			}
		}
		if s.Action == "late" {
			for _, signal := range []string{"logs", "spans"} {
				if !paired444Held(s.Before.Lanes[signal], s.After.Lanes[signal], true) || s.Before.Lanes[signal].Generation != s.After.Lanes[signal].Generation {
					return "draft changed applied source owner"
				}
			}
		}
	}
	for _, ok := range required {
		if !ok {
			return "missing lane transition"
		}
	}
	if len(r.Memberships) < 7 || len(r.Oracle["logs"]) < 7 || len(r.Oracle["spans"]) < 7 {
		return "missing independent membership witnesses"
	}
	for _, m := range r.Memberships {
		if m.Signal != "logs" && m.Signal != "spans" || m.Window.ToMS-m.Window.FromMS != 60000 {
			return "invalid source window"
		}
		var condition *exploration.Predicate
		switch m.Mode {
		case "all":
		case "severity":
			condition = &logCondition.All[0]
		case "presence":
			condition = &logCondition
		case "slow":
			condition = &spanCondition
		default:
			return "unknown membership mode"
		}
		where := &shared
		if condition != nil {
			where = &exploration.Predicate{All: []exploration.Predicate{shared, *condition}}
		}
		expectedSelector := exploration.TailSelector{Signal: m.Signal, Environment: r.Fixture.Environment, Where: where}
		actualSelector := m.Selector
		if actualSelector.Signal == "" {
			actualSelector.Signal = "logs"
		}
		if !reflect.DeepEqual(actualSelector, expectedSelector) || m.Selector.Query(m.Window.ToMS).Validate() != nil {
			return "effective lane query differs"
		}
		expected := []string{}
		for _, row := range r.Oracle[m.Signal] {
			at, e := strconv.ParseInt(row.AtNS, 10, 64)
			if e != nil || row.ID == "" {
				return "invalid source row"
			}
			matched := at >= m.Window.FromMS*1000000 && at < m.Window.ToMS*1000000 && row.Service == r.Fixture.Service && row.Resource["deployment.environment.name"] == r.Fixture.Environment && row.Resource["fixture.bank"] == r.Fixture.Token
			if m.Mode == "severity" || m.Mode == "presence" {
				severity, e := strconv.ParseUint(row.SeverityNumber, 10, 64)
				matched = matched && e == nil && severity >= 17
			}
			if m.Mode == "presence" {
				tenant, present := row.Attributes["tenant"]
				matched = matched && (!present || tenant == "")
			}
			if m.Mode == "slow" {
				duration, e := strconv.ParseUint(row.DurationNS, 10, 64)
				matched = matched && e == nil && duration >= 100000000 && row.Kind == "Server"
			}
			if matched {
				expected = append(expected, row.ID)
			}
		}
		actual, oracle := slices.Clone(m.IDs), slices.Clone(m.OracleIDs)
		slices.Sort(expected)
		slices.Sort(actual)
		slices.Sort(oracle)
		if len(expected) == 0 || !slices.Equal(expected, actual) || !slices.Equal(expected, oracle) {
			return fmt.Sprintf("membership differs: %s expected%v actual%v", m.Name, expected, actual)
		}
	}
	a := r.Bombadil.Completed
	if r.Bombadil.Snapshots == 0 || a.Version != "goo.pairedFilters450Actor/1" || a.Phase != 8 || a.LogsReads != 2 || a.SpansReads != 1 || a.Forbidden != 0 || !a.SiblingHeld || !a.LogApplied || !a.LocalReadFree || a.Overflow || a.LogsPhase != "paused" || a.SpansPhase != "paused" {
		return "actor did not finish independent filter application"
	}
	return ""
}
func TestQuintPairedFilters450Observed(t *testing.T) {
	file := os.Getenv("GOO_PAIRED_FILTERS450_REPORT")
	if file == "" {
		t.Skip("requires real paired lane filters OTLP/browser report")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var report filters450Report
	if e = json.Unmarshal(raw, &report); e != nil {
		t.Fatal(e)
	}
	if reason := filters450Admission(report); reason != "" {
		t.Fatal("actual lane-filter admission:", reason)
	}
	checker := paired444Checker(t)
	replay := func(r filters450Report) bool {
		if len(r.Steps) == 0 {
			return false
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		base := r.Steps[0].Before.Reads
		coherent := filters450Admission(r) == ""
		for _, step := range r.Steps {
			projected := step.paired444Step
			if projected.Action == "draft" {
				projected.Action = "stage"
			}
			action, held, err := paired444Projection(projected)
			if step.Action == "draft" {
				action = "observe"
			}
			if err != nil {
				return false
			}
			if coherent && !held {
				t.Logf("unheld observed step %s", step.Name)
			}
			rec.Record(t.Context(), action, nil, "obs.logsReads", step.After.Reads["logs"]-base["logs"], "obs.spansReads", step.After.Reads["spans"]-base["spans"], "obs.logsActive", paired444Active(step.After.Lanes["logs"].Phase), "obs.spansActive", paired444Active(step.After.Lanes["spans"].Phase), "obs.held", held && coherent)
		}
		verdict, err := checker.Check(t.Context(), sink.Steps())
		if err != nil {
			t.Fatal(err)
		}
		if coherent && !verdict.OK() {
			t.Logf("paired model: %s", verdict.Summary())
			for _, actor := range verdict.Actors {
				t.Logf("failure %+v output%s", actor.Failure, actor.Output)
			}
		}
		return verdict.OK()
	}
	if !replay(report) {
		t.Fatal("actual independently filtered pair rejected by existing ownership model")
	}
	mutations := map[string]func(*filters450Report){
		"saved-log-filter":  func(r *filters450Report) { r.Saved.Lanes.Logs.Where = nil },
		"saved-span-filter": func(r *filters450Report) { r.Saved.Lanes.Spans.Where = nil },
		"shared-scope":      func(r *filters450Report) { r.Shared.All[0].Value = "wrong" },
		"empty-is-missing":  func(r *filters450Report) { r.Conditions["logs"].All[1].Any[1].Op = "missing" },
		"recall-read":       func(r *filters450Report) { r.Recall.SourcePosts = 1 },
		"recall-unpaused":   func(r *filters450Report) { r.Recall.Paused = false },
		"source-row-drift": func(r *filters450Report) {
			for i := range r.Oracle["logs"] {
				r.Oracle["logs"][i].Service = "wrong"
			}
		},
		"membership-id":       func(r *filters450Report) { r.Memberships[0].IDs = []string{"wrong"} },
		"window":              func(r *filters450Report) { r.Memberships[0].Window.FromMS-- },
		"effective-predicate": func(r *filters450Report) { r.Memberships[0].Selector.Where = nil },
		"draft-read": func(r *filters450Report) {
			for i := range r.Steps {
				if r.Steps[i].Name == "draft-error-logs" {
					r.Steps[i].After.Reads["logs"]++
				}
			}
		},
		"apply-sibling": func(r *filters450Report) {
			for i := range r.Steps {
				if r.Steps[i].Name == "apply-error-logs" {
					r.Steps[i].After.Lanes["spans"] = r.Steps[i].After.Lanes["logs"]
				}
			}
		},
		"missing-source-meter":  func(r *filters450Report) { r.Steps[0].SQLAfter = nil },
		"missing-source-ledger": func(r *filters450Report) { r.Source = nil },
		"actor-pending":         func(r *filters450Report) { r.Bombadil.Completed.Phase = 5 },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			var bad filters450Report
			_ = json.Unmarshal(raw, &bad)
			mutate(&bad)
			if replay(bad) {
				t.Fatal("invalid lane-filter trace admitted")
			}
		})
	}
	t.Log("actual lane filters accepted by existing paired ownership model; 15 negative controls rejected")
}
