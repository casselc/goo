package verification

import (
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type raw449Context struct {
	Name, Signal                        string
	FromNS, ToNS, AtNS, RadiusNS, Scope string
	Origin                              struct {
		Version, ReadAt, OccurrenceKey, EndNS string
		Question                              measurements.DiscoveryQuestion
		Stream                                struct{ Resource map[string]string }
	}
	Question, Request exploration.Query
	Response          exploration.Result
	Status            int
	ExpectedIDs       []string
	SourceRows        []json.RawMessage
}
type raw449Bank struct {
	Name, ReadAt, Find string
	Keys               []string
	Detail             json.RawMessage
	Focus              json.RawMessage
	Ports              json.RawMessage
	PageY              json.RawMessage
}
type raw449Report struct {
	Version, Status          string
	Cleanup, DatabaseCleanup bool
	Errors                   []string
	Contexts                 []raw449Context
	Steps                    []histogram447Step
	Banks                    []raw449Bank
	SourceScopeRetirement    struct{ Refused bool }
	Bombadil                 struct {
		Snapshots int
		Completed struct {
			Version                                                             string
			Phase, RawReads, RecordReads, CompletedRecords, Forbidden           int
			HeldRestored, FocusRestored, StageReadFree, DetailVisible, Overflow bool
		}
	}
}

func raw449Admission(r raw449Report) string {
	if r.Version != "goo.rawObservationContext449/1" || r.Status != "passed" || !r.Cleanup || !r.DatabaseCleanup || len(r.Errors) != 0 {
		return "unfinished run"
	}
	if len(r.Contexts) < 2 {
		return "missing nearby contexts"
	}
	signals := map[string]bool{}
	for _, c := range r.Contexts {
		signals[c.Signal] = true
		o := c.Origin
		at, err := strconv.ParseInt(o.EndNS, 10, 64)
		if err != nil || o.ReadAt == "" || o.OccurrenceKey == "" || o.Version == "" {
			return "missing exact held origin"
		}
		lo, hi := max(o.Question.FromMS*1000000, at-60000000000), min(o.Question.ToMS*1000000, at+60000000000)
		if lo >= hi || at < o.Question.FromMS*1000000 || at >= o.Question.ToMS*1000000 {
			return "invalid origin bounds"
		}
		leaf := func(kind, key, op, value string) exploration.Predicate {
			return exploration.Predicate{Field: &exploration.Field{Kind: kind, Key: key}, Op: op, Value: value}
		}
		predicates := []exploration.Predicate{leaf("column", "service", "eq", o.Question.Service)}
		if value, ok := o.Stream.Resource["service.namespace"]; ok {
			predicates = append(predicates, leaf("resource", "service.namespace", "eq", value))
		} else {
			predicates = append(predicates, leaf("resource", "service.namespace", "missing", ""))
		}
		if value, ok := o.Stream.Resource["service.instance.id"]; ok {
			predicates = append(predicates, leaf("resource", "service.instance.id", "eq", value))
		}
		predicates = append(predicates, leaf("column", "timestamp_ns", "gte", strconv.FormatInt(lo, 10)), leaf("column", "timestamp_ns", "lte", strconv.FormatInt(hi-1, 10)))
		want := exploration.Query{Signal: c.Signal, Environment: o.Question.Environment, EnvironmentSelector: o.Question.EnvironmentSelector, FromMS: lo / 1000000, ToMS: (hi + 999999) / 1000000, Where: &exploration.Predicate{All: predicates}, Limit: 100}
		if c.Signal != "logs" && c.Signal != "spans" || c.Question.Validate() != nil || !reflect.DeepEqual(c.Question, want) || !reflect.DeepEqual(c.Request, want) || !reflect.DeepEqual(c.Response.Question, want) || c.Status != 200 {
			return "context differs from independent held scope"
		}
		if c.FromNS != strconv.FormatInt(lo, 10) || c.ToNS != strconv.FormatInt(hi, 10) || c.AtNS != o.EndNS || c.RadiusNS != "60000000000" {
			return "context geometry differs"
		}
		scope := "service"
		if _, ok := o.Stream.Resource["service.instance.id"]; ok {
			scope = "reported-instance"
		}
		if c.Scope != scope {
			return "context scope differs"
		}
		if len(c.SourceRows) == 0 || len(c.ExpectedIDs) == 0 || len(c.Response.Rows) == 0 {
			return "missing independent record witness"
		}
		var expected []string
		for _, raw := range c.SourceRows {
			var row struct {
				ID, Signal, AtNS, Service string
				Resource                  map[string]string
			}
			if json.Unmarshal(raw, &row) != nil || row.ID == "" {
				return "malformed independent source row"
			}
			stamp, e := strconv.ParseInt(row.AtNS, 10, 64)
			if e != nil {
				return "malformed source timestamp"
			}
			matched := row.Signal == c.Signal && row.Service == o.Question.Service && stamp >= lo && stamp < hi
			for _, key := range []string{"service.namespace", "service.instance.id"} {
				value, present := o.Stream.Resource[key]
				actual, exists := row.Resource[key]
				if key == "service.namespace" {
					matched = matched && present == exists && value == actual
				} else if present {
					matched = matched && exists && actual == value
				}
			}
			selector := o.Question.EnvironmentSelector
			if selector == nil {
				matched = matched && row.Resource["deployment.environment.name"] == o.Question.Environment
			} else if selector.Policy == "current-only" {
				matched = matched && row.Resource["deployment.environment.name"] == selector.Scope.Value
			} else {
				value, present := row.Resource["deployment.environment.name"]
				if !present {
					value, present = row.Resource["deployment.environment"]
				}
				mode := "unreported"
				if present {
					mode = "named"
					if value == "" {
						mode = "empty"
					}
				}
				matched = matched && (selector.Scope.Mode == "all" || selector.Scope.Mode == mode && selector.Scope.Value == value)
			}
			if matched {
				expected = append(expected, row.ID)
			}
		}
		actual := []string{}
		for _, raw := range c.Response.Rows {
			var row struct {
				SpanID string `json:"SpanId"`
			}
			if json.Unmarshal(raw, &row) != nil || row.SpanID == "" {
				return "malformed acquired record"
			}
			actual = append(actual, row.SpanID)
		}
		slices.Sort(expected)
		slices.Sort(actual)
		reported := slices.Clone(c.ExpectedIDs)
		slices.Sort(reported)
		if !slices.Equal(expected, actual) || !slices.Equal(expected, reported) {
			return "record membership differs from independent source"
		}
	}
	if !signals["logs"] || !signals["spans"] {
		return "both record signals required"
	}
	required := map[string]bool{"stage-logs": false, "run-logs": false, "return-logs": false, "stage-spans": false, "run-spans": false, "return-spans": false, "source-scope-retirement": false}
	for _, s := range r.Steps {
		if _, ok := required[s.Name]; !ok {
			continue
		}
		required[s.Name] = true
		if s.SourceBefore == nil || s.SourceAfter == nil || s.SQLBefore == nil || s.SQLAfter == nil {
			return "missing authority counter"
		}
		reads := 0
		for _, path := range s.Posts {
			if path == "/api/explore" {
				reads++
			} else if path != "/api/cel/context" && path != "/api/cel/plan" {
				return "unexpected context authority"
			}
		}
		run := strings.HasPrefix(s.Name, "run-")
		if run {
			if s.Local || reads != 1 || *s.SourceAfter-*s.SourceBefore != 1 || (*s.SQLAfter-*s.SQLBefore < 1 || *s.SQLAfter-*s.SQLBefore > 16) {
				return "run must acquire once"
			}
		} else if !s.Local || reads != 0 || *s.SourceAfter != *s.SourceBefore || *s.SQLAfter != *s.SQLBefore {
			return "stage or return acquired data"
		}
	}
	for _, seen := range required {
		if !seen {
			return "missing context transition"
		}
	}
	for _, signal := range []string{"logs", "spans"} {
		var before, after *raw449Bank
		for i := range r.Banks {
			if r.Banks[i].Name == "before-"+signal {
				before = &r.Banks[i]
			}
			if r.Banks[i].Name == "after-"+signal {
				after = &r.Banks[i]
			}
		}
		if before == nil || after == nil || before.ReadAt == "" || len(before.Keys) == 0 || len(before.Detail) == 0 {
			return "missing held bank"
		}
		a, b := *before, *after
		a.Name = ""
		b.Name = ""
		if !reflect.DeepEqual(a, b) {
			return "return changed held bank"
		}
	}
	if !r.SourceScopeRetirement.Refused {
		return "retired context control gained authority"
	}
	a := r.Bombadil.Completed
	if r.Bombadil.Snapshots == 0 || a.Version != "goo.rawObservationContext449Actor/1" || a.Phase != 12 || a.RawReads != 1 || a.RecordReads != 2 || a.CompletedRecords != 2 || a.Forbidden != 0 || !a.HeldRestored || !a.FocusRestored || !a.StageReadFree || !a.DetailVisible || a.Overflow {
		return "actor did not complete nearby records journey"
	}
	return ""
}
func TestQuintRawObservationContext449Observed(t *testing.T) {
	path := os.Getenv("GOO_RAW_CONTEXT449_REPORT")
	if path == "" {
		t.Skip("requires completed real raw-point nearby-records journey")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var r raw449Report
	if err = json.Unmarshal(raw, &r); err != nil {
		t.Fatal(err)
	}
	if reason := raw449Admission(r); reason != "" {
		t.Fatal("actual context admission:", reason)
	}
	checker, err := validate.NewChecker("model/population_drilldown.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	check := func(report raw449Report) bool {
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		reads, depth := 0, 0
		coherent := raw449Admission(report) == ""
		for _, step := range report.Steps {
			action := "history"
			if strings.HasPrefix(step.Name, "stage-") {
				action = "stage"
				depth++
			}
			if strings.HasPrefix(step.Name, "return-") {
				action = "back"
				depth--
			}
			count := 0
			if step.SQLBefore != nil && step.SQLAfter != nil {
				count = *step.SQLAfter - *step.SQLBefore
			}
			if count == 0 {
				rec.Record(t.Context(), action, nil, "obs.reads", reads, "obs.depth", depth, "obs.coherent", coherent)
			}
			for range max(0, count) {
				reads++
				rec.Record(t.Context(), "read", nil, "obs.reads", reads, "obs.depth", depth, "obs.coherent", coherent)
			}
		}
		verdict, e := checker.Check(t.Context(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return verdict.OK()
	}
	if !check(r) {
		t.Fatal("actual context trace rejected")
	}
	mutations := map[string]func(*raw449Report){
		"source-row-drift": func(r *raw449Report) {
			r.Contexts[0].SourceRows = []json.RawMessage{json.RawMessage(`{"id":"different","signal":"logs","atNS":"1","service":"wrong","resource":{}}`)}
		},
		"record-id-drift":  func(r *raw449Report) { r.Contexts[0].ExpectedIDs = []string{"different"} },
		"context-geometry": func(r *raw449Report) { r.Contexts[0].ToNS = "0" },
		"context-scope":    func(r *raw449Report) { r.Contexts[0].Scope = "unbounded" },

		"unfinished":        func(r *raw449Report) { r.Cleanup = false },
		"missing-signal":    func(r *raw449Report) { r.Contexts = r.Contexts[:1] },
		"wide-window":       func(r *raw449Report) { r.Contexts[0].Question.FromMS-- },
		"scope-leak":        func(r *raw449Report) { r.Contexts[0].Question.Where = nil },
		"origin-time":       func(r *raw449Report) { r.Contexts[0].Origin.EndNS = "0" },
		"no-record-witness": func(r *raw449Report) { r.Contexts[0].SourceRows = nil },
		"missing-counter": func(r *raw449Report) {
			for i := range r.Steps {
				if r.Steps[i].Name == "stage-logs" {
					r.Steps[i].SourceAfter = nil
				}
			}
		},
		"return-read": func(r *raw449Report) {
			for i := range r.Steps {
				if r.Steps[i].Name == "return-logs" {
					r.Steps[i].Posts = append(r.Steps[i].Posts, "/api/explore")
				}
			}
		},
		"bank-clock": func(r *raw449Report) {
			for i := range r.Banks {
				if r.Banks[i].Name == "after-logs" {
					r.Banks[i].ReadAt = "changed"
				}
			}
		},
		"stale-control": func(r *raw449Report) { r.SourceScopeRetirement.Refused = false },
		"pending-click": func(r *raw449Report) { r.Bombadil.Completed.CompletedRecords = 1 },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			var mutant raw449Report
			_ = json.Unmarshal(raw, &mutant)
			mutate(&mutant)
			if check(mutant) {
				t.Fatal("invalid context trace admitted")
			}
		})
	}
	t.Log("actual nearby-record trace admitted; 15 negative controls rejected")
}
