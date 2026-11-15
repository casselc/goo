package verification

import (
	"encoding/json"
	"fmt"
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

type incoming453Oracle struct {
	TraceID, SpanID, AtNS     string
	Resource                  map[string]string
	LinkedTraces, LinkedSpans []string
	LinkedAttributes          []map[string]string
}
type incoming453Read struct {
	Request  exploration.Query
	Status   int
	Response struct {
		Question exploration.Query
		Count    string
		HasMore  bool
		Rows     []map[string]any
	}
	ExpectedKeys []string
}
type incoming453Context struct {
	Name         string
	Target       exploration.SpanLinkTarget
	Question     exploration.Query
	SelectedSpan struct {
		Question exploration.Query
		RowIndex int
		Source   exploration.SpanLinkTarget
	}
	Pages        []incoming453Read
	ExpectedKeys []string
}
type incoming453Report struct {
	Version, Status          string
	Cleanup, DatabaseCleanup bool
	Errors                   []string
	Origin                   exploration.Query
	Fixture                  struct {
		Environment string
		IDs         map[string]string
	}
	Oracle   []incoming453Oracle
	Contexts []incoming453Context
	Steps    []link452Step
	Banks    []map[string]any
	Source   []struct {
		Index              int
		SQL, Family, Fault string
		Status             *int
	}
	DuplicateControls struct{ DuplicateLinksOneRecord, DuplicateObservationsRetained, CrossedTupleExcluded bool }
	MissingSelector   struct {
		Refused bool
		Request exploration.Query
		Status  int
	}
	Cancellation struct{ SourceCompleted, Ignored bool }
	Late         struct {
		incoming453Read
		Ignored bool
	}
	Stale    struct{ RunRefused, ClosedRefused bool }
	Bombadil struct {
		Snapshots int
		Completed map[string]any
	}
}

func incoming453Members(r incoming453Report, target exploration.SpanLinkTarget) ([]string, string) {
	rows := []incoming453Oracle{}
	for _, row := range r.Oracle {
		at, err := strconv.ParseUint(row.AtNS, 10, 64)
		if err != nil {
			return nil, "invalid oracle timestamp"
		}
		env, has := row.Resource["deployment.environment.name"]
		if !has {
			env = row.Resource["deployment.environment"]
		}
		if env != r.Fixture.Environment || at < uint64(r.Origin.FromMS)*1000000 || at >= uint64(r.Origin.ToMS)*1000000 {
			continue
		}
		// Independent same-position oracle, not the backend helper being validated.
		if len(row.LinkedTraces) != len(row.LinkedSpans) || len(row.LinkedTraces) != len(row.LinkedAttributes) {
			continue
		}
		match := false
		for i, trace := range row.LinkedTraces {
			if trace == target.TraceID && row.LinkedSpans[i] == target.SpanID && row.LinkedAttributes[i] != nil {
				match = true
			}
		}
		if match {
			rows = append(rows, row)
		}
	}
	slices.SortFunc(rows, func(a, b incoming453Oracle) int {
		aa, _ := strconv.ParseUint(a.AtNS, 10, 64)
		bb, _ := strconv.ParseUint(b.AtNS, 10, 64)
		if aa < bb {
			return -1
		}
		if aa > bb {
			return 1
		}
		return 0
	})
	keys := []string{}
	for _, row := range rows {
		keys = append(keys, row.TraceID+"/"+row.SpanID+"@"+row.AtNS)
	}
	return keys, ""
}
func incoming453ReadAdmission(read incoming453Read, expected exploration.Query, keys []string) string {
	if !reflect.DeepEqual(read.Request, expected) || read.Status != 200 || !reflect.DeepEqual(read.Response.Question, expected) || read.Response.Count != strconv.Itoa(len(keys)) {
		return "read scope/count differs"
	}
	end := min(len(keys), expected.Offset+100)
	start := min(len(keys), expected.Offset)
	want := keys[start:end]
	got := []string{}
	for _, row := range read.Response.Rows {
		trace, tok := row["TraceId"].(string)
		span, sok := row["SpanId"].(string)
		at, aok := row["timestamp_ns"].(string)
		if !tok || !sok || !aok {
			return "malformed returned row"
		}
		traces, tok := row["linked_traces"].([]any)
		spans, sok := row["linked_spans"].([]any)
		attrs, aok := row["linked_attributes"].([]any)
		if !tok || !sok || !aok || len(traces) != len(spans) || len(traces) != len(attrs) || expected.LinksTo == nil {
			return "returned reference arrays malformed"
		}
		matched := false
		for i, value := range traces {
			if value == expected.LinksTo.TraceID && spans[i] == expected.LinksTo.SpanID {
				attributes, ok := attrs[i].(map[string]any)
				valid := ok
				for _, value := range attributes {
					if _, ok := value.(string); !ok {
						valid = false
					}
				}
				if valid {
					matched = true
				}
			}
		}
		if !matched {
			return "returned reference tuple not witnessed"
		}
		got = append(got, trace+"/"+span+"@"+at)
	}
	if !reflect.DeepEqual(got, want) || read.Response.HasMore != (len(keys) > expected.Offset+100) {
		return "physical page differs from oracle"
	}
	return ""
}
func incoming453Admission(r incoming453Report) string {
	if r.Version != "goo.spanLinkIncoming453/1" || r.Status != "passed" || !r.Cleanup || !r.DatabaseCleanup || len(r.Errors) > 0 || r.Origin.Validate() != nil || r.Origin.Signal != "spans" || len(r.Oracle) != 110 || len(r.Contexts) != 7 {
		return "unfinished physical run"
	}
	target := exploration.SpanLinkTarget{TraceID: r.Fixture.IDs["targetTrace"], SpanID: r.Fixture.IDs["targetSpan"]}
	members, why := incoming453Members(r, target)
	if why != "" || len(members) != 105 {
		return "physical incoming count differs"
	}
	identities := map[string]int{}
	multipleLinks := false
	for _, row := range r.Oracle {
		key := row.TraceID + "/" + row.SpanID
		identities[key]++
		if len(row.LinkedTraces) > 1 && len(row.LinkedTraces) == len(row.LinkedSpans) && row.LinkedTraces[0] == target.TraceID && row.LinkedTraces[1] == target.TraceID && row.LinkedSpans[0] == target.SpanID && row.LinkedSpans[1] == target.SpanID {
			multipleLinks = true
		}
	}
	duplicateRows := false
	for _, n := range identities {
		if n > 1 {
			duplicateRows = true
		}
	}
	if !multipleLinks || !duplicateRows || !r.DuplicateControls.DuplicateLinksOneRecord || !r.DuplicateControls.DuplicateObservationsRetained || !r.DuplicateControls.CrossedTupleExcluded {
		return "missing physical duplicate controls"
	}
	contexts := map[string]incoming453Context{}
	for _, c := range r.Contexts {
		if _, exists := contexts[c.Name]; exists {
			return "duplicate context"
		}
		contexts[c.Name] = c
		wantTarget := target
		if c.Name == "empty" {
			wantTarget.SpanID = r.Fixture.IDs["emptySpan"]
		}
		if c.Target != wantTarget || c.SelectedSpan.Source != wantTarget {
			return "selected target differs"
		}
		expected := exploration.Query{Signal: "spans", Environment: r.Origin.Environment, EnvironmentSelector: r.Origin.EnvironmentSelector, FromMS: r.Origin.FromMS, ToMS: r.Origin.ToMS, Limit: 100, LinksTo: &wantTarget}
		if !reflect.DeepEqual(c.Question, expected) || c.Question.Validate() != nil || !c.SelectedSpan.Question.SameEnvironment(r.Origin) || c.SelectedSpan.Question.FromMS != r.Origin.FromMS || c.SelectedSpan.Question.ToMS != r.Origin.ToMS {
			return "incoming scope changed"
		}
		keys, err := incoming453Members(r, wantTarget)
		if err != "" {
			return err
		}
		successes, failures := 0, 0
		offsets := []int{}
		for _, page := range c.Pages {
			q := expected
			q.Offset = page.Request.Offset
			if !reflect.DeepEqual(page.Request, q) {
				return "page query drift"
			}
			if page.Status >= 400 {
				if len(page.Response.Rows) > 0 {
					return "refusal fabricated rows"
				}
				failures++
				continue
			}
			successes++
			offsets = append(offsets, q.Offset)
			if why := incoming453ReadAdmission(page, q, keys); why != "" {
				return why
			}
		}
		switch c.Name {
		case "references":
			if !reflect.DeepEqual(offsets, []int{0, 100, 0}) || failures != 0 || !reflect.DeepEqual(c.ExpectedKeys, keys) {
				return "missing100/5/Previous pages"
			}
		case "empty":
			if successes != 1 || failures != 0 || len(keys) != 0 {
				return "missing empty result"
			}
		case "refused", "missing-selector":
			if successes != 1 || failures != 1 {
				return "missing retry after refusal"
			}
		case "canceled", "late", "stale":
			if len(c.Pages) != 0 {
				return "retired owner adopted result"
			}
		default:
			return "unknown context"
		}
	}
	for _, name := range []string{"references", "empty", "refused", "missing-selector", "canceled", "late", "stale"} {
		if _, ok := contexts[name]; !ok {
			return "missing scenario"
		}
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
			return "held bank changed"
		}
	}
	if !r.MissingSelector.Refused || r.MissingSelector.Status < 400 || !reflect.DeepEqual(r.MissingSelector.Request, contexts["missing-selector"].Question) || !r.Cancellation.SourceCompleted || !r.Cancellation.Ignored || !r.Late.Ignored || !r.Stale.RunRefused || !r.Stale.ClosedRefused {
		return "missing refusal/retirement control"
	}
	if why := incoming453ReadAdmission(r.Late.incoming453Read, contexts["late"].Question, members); why != "" {
		return "late actual read: " + why
	}
	if len(r.Steps) < 25 || len(r.Source) < 10 {
		return "missing source ledger"
	}
	lastSource, lastSQL := 0, 0
	schemaFailureIndexes := map[int]bool{}
	for _, s := range r.Steps {
		if s.Before.Source == nil || s.After.Source == nil || s.Before.SQL == nil || s.After.SQL == nil {
			return "missing read counter"
		}
		a, b, x, y := *s.Before.Source, *s.After.Source, *s.Before.SQL, *s.After.SQL
		if a < lastSource || x < lastSQL || b < a || y < x || y > len(r.Source) {
			return "counter discontinuity"
		}
		lastSource, lastSQL = b, y
		if s.Name == "run-missing-selector-0" {
			for i := x; i < y; i++ {
				schemaFailureIndexes[i] = true
			}
		}
		if s.Local {
			if a != b || x != y {
				return "local source acquisition"
			}
			for _, post := range s.Posts {
				if (s.Name != "back-canceled" && s.Name != "back-late") || !histogram447Cancel.MatchString(post) {
					return "local post acquisition"
				}
			}
		}
	}
	for i, call := range r.Source {
		if call.Index != i || call.SQL == "" || call.Status == nil {
			return "source ledger unfinished"
		}
		allowed := *call.Status == 200 || call.Fault == "refuse" && *call.Status == 403 || schemaFailureIndexes[i] && *call.Status >= 400 && *call.Status < 600 && strings.Contains(call.SQL, "Links.SpanId")
		if !allowed {
			return "undeclared source failure"
		}
	}
	actor := r.Bombadil.Completed
	if r.Bombadil.Snapshots < 1 || actor["version"] != "goo.spanLinkIncoming453Actor/1" || actor["phase"] != float64(9) || actor["forbidden"] != float64(0) || actor["source"] != float64(6) || actor["count"] != "105" || actor["lookupRows"] != float64(5) || actor["firstPageRows"] != float64(100) || actor["traceRows"] != float64(2) || actor["stageReadFree"] != true || actor["heldRestored"] != true || actor["focusRestored"] != true {
		return "incomplete actor"
	}
	return ""
}
func TestQuintSpanLinkIncoming453Observed(t *testing.T) {
	path := os.Getenv("GOO_SPAN_LINK_INCOMING453_REPORT")
	if path == "" {
		t.Skip("requires actual incoming-span-link report")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var report incoming453Report
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if why := incoming453Admission(report); why != "" {
		t.Fatal("physical source admission:", why)
	}
	checker, err := validate.NewChecker("model/population_drilldown.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	check := func(r incoming453Report) bool {
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		depth, reads := 0, 0
		coherent := incoming453Admission(r) == ""
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
	mutations := map[string]func(*incoming453Report){"unfinished": func(r *incoming453Report) { r.DatabaseCleanup = false }, "wide-window": func(r *incoming453Report) { r.Contexts[0].Question.FromMS-- }, "target-leak": func(r *incoming453Report) { r.Contexts[0].Question.LinksTo = nil }, "oracle-membership": func(r *incoming453Report) {
		target := exploration.SpanLinkTarget{TraceID: r.Fixture.IDs["targetTrace"], SpanID: r.Fixture.IDs["targetSpan"]}
		members, _ := incoming453Members(*r, target)
		for i, row := range r.Oracle {
			if row.TraceID+"/"+row.SpanID+"@"+row.AtNS == members[0] {
				r.Oracle[i].LinkedAttributes = nil
				return
			}
		}
	}, "wrong-count": func(r *incoming453Report) { r.Contexts[0].Pages[0].Response.Count = "106" }, "wrong-page": func(r *incoming453Report) { r.Contexts[0].Pages[1].Response.Rows = nil }, "missing-Previous": func(r *incoming453Report) { r.Contexts[0].Pages = r.Contexts[0].Pages[:2] }, "held-bank": func(r *incoming453Report) { r.Banks[1]["detail"] = "changed" }, "schema-failure-hidden": func(r *incoming453Report) { r.MissingSelector.Refused = false }, "late-adopted": func(r *incoming453Report) { r.Late.Ignored = false }, "stale-authority": func(r *incoming453Report) { r.Stale.RunRefused = false }, "actor-incomplete": func(r *incoming453Report) { r.Bombadil.Completed["heldRestored"] = false }, "stage-acquires": func(r *incoming453Report) {
		for i := range r.Steps {
			if r.Steps[i].Action == "stage" {
				n := *r.Steps[i].After.SQL + 1
				r.Steps[i].After.SQL = &n
				break
			}
		}
	}, "source-ledger": func(r *incoming453Report) { r.Source[0].SQL = "" }}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			var r incoming453Report
			_ = json.Unmarshal(raw, &r)
			mutate(&r)
			if check(r) {
				t.Fatal("invalid incoming report admitted")
			}
		})
	}
	t.Log(fmt.Sprintf("actual incoming source/query/pages accepted; %d negative controls rejected; model covers only observed stage/read/back/read counts", len(mutations)))
}
