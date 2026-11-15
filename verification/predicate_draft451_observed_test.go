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
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type draft451Snapshot struct {
	Reads, SQL                 int
	Phase, Environment, Signal string
	Keys                       []string
	State                      struct {
		Selector   exploration.TailSelector
		ReadWindow *struct{ FromMS, ToMS int64 }
	}
}
type draft451Report struct {
	ModeDiscard struct {
		Signal   string
		Disabled bool
		Where    exploration.Predicate
	}
	OpenedSuggestions struct {
		Nodes  int
		Tenant bool
	}
	Version, Label, Status string
	Cleanup                bool
	Errors                 []string
	Fixture                struct{ Token, Environment, Service string }
	Base, Compiled         exploration.Predicate
	Steps                  []struct {
		Name, Action  string
		Local         bool
		Before, After draft451Snapshot
		Posts         []string
	}
	Oracle      []filters450Row
	Memberships []struct {
		Name, Minimum string
		IDs, Expected []string
	}
	Measurement struct {
		Before, After                               draft451Snapshot
		Mutations, Nodes, AfterNodes, DocumentNodes int
	}
	Discard struct {
		AppliedEnvironment  string
		AppliedVisible      bool
		AppliedSelector     exploration.TailSelector
		Environment, Signal string
		Where               exploration.Predicate
	}
	Saved                 workbench.ActivityView
	AppliedRequest        exploration.TailSelector
	FinalTitle, FinalText string
	Source                []struct {
		SQL    string
		Status *int
	}
}

func draft451Admission(r draft451Report) string {
	if r.Version != "goo.predicateDraft451/1" || r.Label != "final" || r.Status != "passed" || !r.Cleanup || len(r.Errors) > 0 || r.Fixture.Token == "" {
		return "unfinished real single-draft journey"
	}
	leaf := func(key, value string) exploration.Predicate {
		return exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: key}, Op: "gte", Value: value}
	}
	base := exploration.Predicate{All: []exploration.Predicate{{Field: &exploration.Field{Kind: "resource", Key: "fixture.bank"}, Op: "eq", Value: r.Fixture.Token}, {Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: r.Fixture.Service}, leaf("duration_ns", "100000000")}}
	compiled := exploration.Predicate{All: []exploration.Predicate{base, leaf("duration_ns", "150000000")}}
	if !reflect.DeepEqual(base, r.Base) || !reflect.DeepEqual(compiled, r.Compiled) || !reflect.DeepEqual(r.Saved.Selector.Where, &compiled) || r.Saved.Validate() != nil || r.Saved.Selector.Signal != "spans" || r.Saved.Selector.Environment != r.Fixture.Environment || !reflect.DeepEqual(r.AppliedRequest.Where, &compiled) || r.AppliedRequest.Signal != "spans" || r.AppliedRequest.Environment != r.Fixture.Environment {
		return "checked or saved AST changed"
	}
	if !r.Discard.AppliedVisible || !strings.Contains(r.Discard.AppliedEnvironment, r.Fixture.Environment) || strings.Contains(r.Discard.AppliedEnvironment, r.Fixture.Environment+"-draft") || !reflect.DeepEqual(r.Discard.AppliedSelector, exploration.TailSelector{Signal: "spans", Environment: r.Fixture.Environment, Where: &base}) {
		return "discard concealed or changed applied environment"
	}
	if r.Discard.Environment != r.Fixture.Environment+"-draft" || r.Discard.Signal != "spans" || !reflect.DeepEqual(r.Discard.Where, base) {
		return "discard changed scope intent or applied AST"
	}
	if r.ModeDiscard.Signal != "logs" || !r.ModeDiscard.Disabled || !reflect.DeepEqual(r.ModeDiscard.Where, compiled) {
		return "discard concealed incompatible signal intent"
	}
	var title exploration.Predicate
	if json.Unmarshal([]byte(strings.TrimPrefix(r.FinalTitle, "Exact predicate · ")), &title) != nil || !reflect.DeepEqual(title, compiled) || !strings.Contains(r.FinalText, "150 ms") {
		return "human summary lost exact AST"
	}
	if !r.OpenedSuggestions.Tenant || r.OpenedSuggestions.Nodes <= r.Measurement.AfterNodes {
		return "opening editor did not hydrate held suggestions"
	}
	m := r.Measurement
	if m.Mutations != 0 || m.After.Reads != m.Before.Reads || m.After.SQL-m.Before.SQL < 2 || m.Before.State.ReadWindow == nil || m.After.State.ReadWindow == nil || m.After.State.ReadWindow.ToMS <= m.Before.State.ReadWindow.ToMS || len(m.After.Keys) != 100 {
		return "closed-editor optimization stopped source or retained work"
	}
	required := map[string]bool{"invalid-visual-draft": false, "discard-visual-draft": false, "pending-cel-blocks-save": false, "check-cel-condition": false, "apply-checked-cel": false, "save-checked-ast": false, "discard-pending-cel": false, "discard-keeps-incompatible-signal-intent": false}
	if len(r.Source) == 0 {
		return "missing physical source witness"
	}
	for _, c := range r.Source {
		if c.SQL == "" || c.Status == nil || *c.Status != 200 {
			return "incomplete source witness"
		}
	}
	for _, s := range r.Steps {
		if _, ok := required[s.Name]; ok {
			required[s.Name] = true
		}
		if s.Before.SQL < 0 || s.After.SQL < s.Before.SQL || s.After.SQL > len(r.Source) {
			return "physical source counter differs"
		}
		if s.Local && (s.Before.Reads != s.After.Reads || s.Before.SQL != s.After.SQL || !slices.Equal(s.Before.Keys, s.After.Keys) || !reflect.DeepEqual(s.Before.State.Selector, s.After.State.Selector)) {
			return "local draft changed held source bank"
		}
		if s.Action == "connect" && s.After.Reads-s.Before.Reads != 1 {
			return "explicit Apply did not acquire exactly once"
		}
		if s.Action == "pause" && s.After.Phase != "paused" {
			return "pause not observed"
		}
		for _, p := range s.Posts {
			if p != "/api/tail" && p != "/api/cel/context" && p != "/api/cel/plan" && p != "/api/activity-views" && !histogram447Cancel.MatchString(p) {
				return "unexpected authority"
			}
		}
	}
	for _, seen := range required {
		if !seen {
			return "missing single draft transition"
		}
	}
	if len(r.Oracle) != 102 || len(r.Memberships) != 2 {
		return "missing independent telemetry oracle"
	}
	for _, membership := range r.Memberships {
		minimum, e := strconv.ParseUint(membership.Minimum, 10, 64)
		if e != nil {
			return "invalid membership bound"
		}
		expected := []string{}
		for _, row := range r.Oracle {
			duration, e := strconv.ParseUint(row.DurationNS, 10, 64)
			if e != nil {
				return "invalid physical duration"
			}
			if row.Service == r.Fixture.Service && row.Resource["fixture.bank"] == r.Fixture.Token && row.Resource["deployment.environment.name"] == r.Fixture.Environment && duration >= minimum {
				expected = append(expected, row.ID)
			}
		}
		actual, oracle := slices.Clone(membership.IDs), slices.Clone(membership.Expected)
		slices.Sort(expected)
		slices.Sort(actual)
		slices.Sort(oracle)
		if !slices.Equal(expected, actual) || !slices.Equal(expected, oracle) {
			return "physical record membership differs"
		}
	}
	return ""
}
func TestQuintPredicateDraft451Observed(t *testing.T) {
	file := os.Getenv("GOO_PREDICATE_DRAFT451_REPORT")
	if file == "" {
		t.Skip("requires actual single predicate-draft browser report")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var r draft451Report
	if e = json.Unmarshal(raw, &r); e != nil {
		t.Fatal(e)
	}
	if reason := draft451Admission(r); reason != "" {
		t.Fatal(reason)
	}
	checker, e := validate.NewChecker("model/stream_entry.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	replay := func(r draft451Report) bool {
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		coherent := draft451Admission(r) == ""
		for _, step := range r.Steps {
			rec.Record(t.Context(), step.Action, nil, "obs.reads", step.After.Reads, "obs.gap", false, "obs.coherent", coherent)
		}
		v, e := checker.Check(t.Context(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return v.OK()
	}
	if !replay(r) {
		t.Fatal("existing stream-entry model rejected actual draft journey")
	}
	mutants := map[string]func(*draft451Report){"stale-applied-environment": func(r *draft451Report) { r.Discard.AppliedVisible = false },
		"discard-mode":  func(r *draft451Report) { r.ModeDiscard.Disabled = false },
		"discard-scope": func(r *draft451Report) { r.Discard.Environment = "wrong" }, "discard-ast": func(r *draft451Report) { r.Discard.Where = exploration.Predicate{} }, "saved-ast": func(r *draft451Report) { r.Saved.Selector.Where = nil }, "hidden-work": func(r *draft451Report) { r.Measurement.Mutations = 1 }, "source-stopped": func(r *draft451Report) { r.Measurement.After.SQL = r.Measurement.Before.SQL }, "stale-window": func(r *draft451Report) { r.Measurement.After.State.ReadWindow = r.Measurement.Before.State.ReadWindow }, "raw-title": func(r *draft451Report) { r.FinalTitle = "150 ms" }, "record-id": func(r *draft451Report) { r.Memberships[0].IDs = []string{"wrong"} }, "source-ledger": func(r *draft451Report) { r.Source = nil }, "draft-read": func(r *draft451Report) {
			for i := range r.Steps {
				if r.Steps[i].Name == "discard-pending-cel" {
					r.Steps[i].After.Reads++
				}
			}
		}}
	for name, mutate := range mutants {
		t.Run(name, func(t *testing.T) {
			var bad draft451Report
			_ = json.Unmarshal(raw, &bad)
			mutate(&bad)
			if replay(bad) {
				t.Fatal("invalid single draft trace admitted")
			}
		})
	}
	t.Log("actual single draft journey accepted by existing stream-entry model; 12 negative controls rejected")
}
