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
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type link452Read struct {
	Request  exploration.Query
	Status   int
	Response struct {
		Question exploration.Query
		Rows     []map[string]any
	}
}
type link452Context struct {
	Name         string
	Target       struct{ TraceID, SpanID string }
	Question     exploration.Query
	ReportedLink struct {
		Question            exploration.Query
		RowIndex, LinkIndex int
		Source              struct{ TraceID, SpanID string }
	}
	Reads       []link452Read
	ExpectedIDs []string
}
type link452Step struct {
	Name, Action  string
	Local         bool
	Before, After struct{ Source, SQL *int }
	Posts         []string
}
type link452Report struct {
	Version, Status string
	Cleanup         bool
	Errors          []string
	Origin          exploration.Query
	Fixture         struct {
		Environment string
		IDs         map[string]string
	}
	Contexts []link452Context
	Steps    []link452Step
	Banks    []map[string]any
	Oracle   []struct {
		TraceID, SpanID, AtNS, Service string
		Resource                       map[string]string
		LinkedTraces, LinkedSpans      []string
	}
	Source []struct {
		SQL, Family, Fault string
		Index              int
		Status             *int
	}
	Cancellation struct{ SourceCompleted, Ignored bool }
	Late         struct {
		link452Read
		Ignored bool
	}
	Stale    struct{ RunRefused, ClosedRefused bool }
	Bombadil struct {
		Snapshots int
		Completed struct {
			Version                                         string
			Phase, Source, Forbidden, LookupRows, TraceRows int
			StageReadFree, HeldRestored, FocusRestored      bool
		}
	}
}

func link452Admission(r link452Report) string {
	if r.Version != "goo.spanLinkTarget452/1" || r.Status != "passed" || !r.Cleanup || len(r.Errors) > 0 || r.Origin.Validate() != nil || r.Origin.Signal != "spans" || len(r.Oracle) != 8 {
		return "unfinished real-source run"
	}
	if len(r.Contexts) != 8 || len(r.Steps) < 20 || len(r.Source) < 10 {
		return "missing journey or source ledger"
	}
	contexts := map[string]link452Context{}
	for _, c := range r.Contexts {
		if _, exists := contexts[c.Name]; exists {
			return "duplicate context"
		}
		contexts[c.Name] = c
		expected := exploration.Query{Signal: "spans", Environment: r.Origin.Environment, EnvironmentSelector: r.Origin.EnvironmentSelector, FromMS: r.Origin.FromMS, ToMS: r.Origin.ToMS, Limit: 100, Where: &exploration.Predicate{All: []exploration.Predicate{{Field: &exploration.Field{Kind: "column", Key: "trace_id"}, Op: "eq", Value: c.Target.TraceID}, {Field: &exploration.Field{Kind: "column", Key: "span_id"}, Op: "eq", Value: c.Target.SpanID}}}}
		if !reflect.DeepEqual(c.Question, expected) || c.Question.Validate() != nil || c.ReportedLink.Question.Signal != "spans" || !c.ReportedLink.Question.SameEnvironment(r.Origin) || c.ReportedLink.Question.FromMS != r.Origin.FromMS || c.ReportedLink.Question.ToMS != r.Origin.ToMS || c.ReportedLink.Source.TraceID != r.Fixture.IDs["sourceTrace"] || c.ReportedLink.Source.SpanID != r.Fixture.IDs["sourceSpan"] {
			return "changed exact lookup scope"
		}
		targetTrace, targetSpan := r.Fixture.IDs["targetTrace"], r.Fixture.IDs["targetSpan"]
		switch c.Name {
		case "absent":
			targetTrace, targetSpan = r.Fixture.IDs["absentTrace"], r.Fixture.IDs["absentSpan"]
		case "excluded":
			targetTrace, targetSpan = r.Fixture.IDs["excludedTrace"], r.Fixture.IDs["excludedSpan"]
		case "ambiguous":
			targetTrace, targetSpan = r.Fixture.IDs["sourceTrace"], r.Fixture.IDs["childSpan"]
		}
		if c.Target.TraceID != targetTrace || c.Target.SpanID != targetSpan {
			return "wrong reported target"
		}
		witnessed := false
		for _, row := range r.Oracle {
			if row.TraceID == c.ReportedLink.Source.TraceID && row.SpanID == c.ReportedLink.Source.SpanID {
				i := c.ReportedLink.LinkIndex
				if i >= 0 && i < len(row.LinkedTraces) && len(row.LinkedTraces) == len(row.LinkedSpans) && row.LinkedTraces[i] == targetTrace && row.LinkedSpans[i] == targetSpan {
					witnessed = true
				}
			}
		}
		if !witnessed {
			return "reported target not witnessed in source link arrays"
		}
		ids := []string{}
		for _, row := range r.Oracle {
			at, e := strconv.ParseUint(row.AtNS, 10, 64)
			if e != nil {
				return "invalid oracle clock"
			}
			env, has := row.Resource["deployment.environment.name"]
			if !has {
				env = row.Resource["deployment.environment"]
			}
			if row.TraceID == targetTrace && row.SpanID == targetSpan && at >= uint64(r.Origin.FromMS)*1000000 && at < uint64(r.Origin.ToMS)*1000000 && env == r.Fixture.Environment {
				ids = append(ids, row.SpanID)
			}
		}
		slices.Sort(ids)
		successes, refusals := 0, 0
		for _, read := range c.Reads {
			if !reflect.DeepEqual(read.Request, expected) {
				return "request differs from staged target"
			}
			if read.Status == 502 {
				refusals++
				continue
			}
			if read.Status != 200 || !reflect.DeepEqual(read.Response.Question, expected) {
				return "source response scope drift"
			}
			successes++
			got := []string{}
			for _, row := range read.Response.Rows {
				if row["TraceId"] != targetTrace || row["SpanId"] != targetSpan {
					return "crossed source tuple"
				}
				got = append(got, row["SpanId"].(string))
			}
			slices.Sort(got)
			if !reflect.DeepEqual(got, ids) || !reflect.DeepEqual(c.ExpectedIDs, ids) {
				return "oracle differs from returned matches"
			}
		}
		switch c.Name {
		case "target", "absent", "excluded", "ambiguous":
			if successes != 1 || refusals != 0 {
				return "missing completed lookup"
			}
		case "refused":
			if successes != 1 || refusals != 1 {
				return "missing refusal recovery"
			}
		case "canceled", "late", "stale":
			if len(c.Reads) != 0 {
				return "retired response adopted"
			}
		default:
			return "unknown context"
		}
	}
	for _, name := range []string{"target", "absent", "excluded", "ambiguous", "refused", "canceled", "late", "stale"} {
		if _, ok := contexts[name]; !ok {
			return "missing target case"
		}
	}
	names := map[string]bool{}
	lastSource, lastSQL := 0, 0
	for _, s := range r.Steps {
		names[s.Name] = true
		if s.Before.Source == nil || s.After.Source == nil || s.Before.SQL == nil || s.After.SQL == nil {
			return "missing acquisition counter"
		}
		a, b, x, y := *s.Before.Source, *s.After.Source, *s.Before.SQL, *s.After.SQL
		if a < lastSource || x < lastSQL || b < a || y < x || y > len(r.Source) {
			return "source ledger discontinuity"
		}
		lastSource, lastSQL = b, y
		if s.Local {
			if a != b || x != y {
				return "local interaction acquired source"
			}
			for _, post := range s.Posts {
				if (s.Name != "back-canceled" && s.Name != "back-late") || !histogram447Cancel.MatchString(post) {
					return "local interaction posted acquisition"
				}
			}
		}
		if strings.HasPrefix(s.Name, "stage-") || strings.HasPrefix(s.Name, "back-") {
			if !s.Local {
				return "staging or Return missing source-free witness"
			}
		}
	}
	for _, name := range []string{"run-origin", "expand-held-trace", "stage-target", "back-target", "complete-canceled-source", "deliver-late-success", "retired-run-refused", "close-trace"} {
		if !names[name] {
			return "missing control"
		}
	}
	for i, call := range r.Source {
		if call.Index != i || call.SQL == "" || call.Status == nil || (*call.Status != 200 && !(call.Fault == "refuse" && *call.Status == 403)) {
			return "unfinished source call"
		}
	}
	for _, name := range []string{"target", "absent", "excluded", "ambiguous", "refused", "canceled", "late", "stale"} {
		var before, after map[string]any
		for _, bank := range r.Banks {
			if bank["name"] == "before-"+name {
				before = bank
			}
			if bank["name"] == "after-"+name {
				after = bank
			}
		}
		if before == nil || after == nil {
			return "missing held Return bank"
		}
		a, b := map[string]any{}, map[string]any{}
		for k, v := range before {
			if k != "name" {
				a[k] = v
			}
		}
		for k, v := range after {
			if k != "name" {
				b[k] = v
			}
		}
		if !reflect.DeepEqual(a, b) {
			return "held origin changed on Return"
		}
	}
	if !r.Cancellation.SourceCompleted || !r.Cancellation.Ignored || !r.Late.Ignored || r.Late.Status != 200 || !r.Stale.RunRefused || !r.Stale.ClosedRefused {
		return "missing retirement control"
	}
	if !reflect.DeepEqual(r.Late.Request, contexts["late"].Question) || !reflect.DeepEqual(r.Late.Response.Question, contexts["late"].Question) {
		return "late response not actual target read"
	}
	actor := r.Bombadil.Completed
	if r.Bombadil.Snapshots < 1 || actor.Version != "goo.spanLinkTarget452Actor/1" || actor.Phase != 8 || actor.Forbidden != 0 || actor.LookupRows != 1 || actor.TraceRows != 3 || !actor.StageReadFree || !actor.HeldRestored || !actor.FocusRestored {
		return "actor did not finish observed workflow"
	}
	return ""
}
func TestQuintSpanLinkTarget452Observed(t *testing.T) {
	path := os.Getenv("GOO_SPAN_LINK_TARGET452_REPORT")
	if path == "" {
		t.Skip("requires actual reported-span-link journey report")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var report link452Report
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if why := link452Admission(report); why != "" {
		t.Fatal("source admission:", why)
	}
	checker, err := validate.NewChecker("model/population_drilldown.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	check := func(r link452Report) bool {
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		depth, reads := 0, 0
		coherent := link452Admission(r) == ""
		for _, s := range r.Steps {
			action := "history"
			if s.Action == "stage" {
				action = "stage"
				depth++
			}
			if s.Action == "back" {
				action = "back"
				depth--
			}
			count := 0
			if s.Before.SQL != nil && s.After.SQL != nil {
				count = *s.After.SQL - *s.Before.SQL
			}
			if count == 0 {
				rec.Record(t.Context(), action, nil, "obs.depth", depth, "obs.reads", reads, "obs.coherent", coherent)
			}
			for range max(0, count) {
				reads++
				rec.Record(t.Context(), "read", nil, "obs.depth", depth, "obs.reads", reads, "obs.coherent", coherent)
			}
		}
		verdict, e := checker.Check(t.Context(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return verdict.OK()
	}
	if !check(report) {
		t.Fatal("actual stage/read/back trace rejected")
	}
	mutations := map[string]func(*link452Report){"unfinished": func(r *link452Report) { r.Cleanup = false }, "wide-window": func(r *link452Report) { r.Contexts[0].Question.FromMS-- }, "scope-leak": func(r *link452Report) { r.Contexts[0].Question.Where = nil }, "wrong-target": func(r *link452Report) { r.Contexts[0].Target.SpanID = r.Fixture.IDs["wrongSpan"] }, "missing-oracle": func(r *link452Report) { r.Oracle = nil }, "wrong-source-result": func(r *link452Report) { r.Contexts[0].Reads[0].Response.Rows = nil }, "held-bank-drift": func(r *link452Report) { r.Banks[1]["detail"] = "changed" }, "stale-action": func(r *link452Report) { r.Stale.RunRefused = false }, "late-adopted": func(r *link452Report) { r.Late.Ignored = false }, "actor-incomplete": func(r *link452Report) { r.Bombadil.Completed.HeldRestored = false }, "stage-read": func(r *link452Report) {
		for i := range r.Steps {
			if r.Steps[i].Name == "stage-target" {
				n := *r.Steps[i].After.SQL + 1
				r.Steps[i].After.SQL = &n
			}
		}
	}, "back-read": func(r *link452Report) {
		for i := range r.Steps {
			if r.Steps[i].Name == "back-target" {
				r.Steps[i].Posts = append(r.Steps[i].Posts, "/api/explore")
			}
		}
	}, "missing-counter": func(r *link452Report) { r.Steps[0].After.SQL = nil }, "source-ledger": func(r *link452Report) { r.Source[0].SQL = "" }}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			var r link452Report
			_ = json.Unmarshal(raw, &r)
			mutate(&r)
			if check(r) {
				t.Fatal("invalid observed run admitted")
			}
		})
	}
	t.Log("actual source-linked stage/read/back trace accepted; 14 negative controls rejected; model proves only declared depth/read/coherence transitions")
}
