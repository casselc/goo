package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/scopedview"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type rateCounters406 struct{ Source, Store, Writes, Expand, SQL int }
type rateFrame406 struct {
	Snapshot scopedview.Dataset
	Display  struct {
		Fields []string
		Value  string
	}
	Fields []string
}
type rateCase406 struct {
	Kind                                                   string
	SourceQuestion, DestinationQuestion                    exploration.Query
	Source, Child, Destination, Reopened                   customizeRead394
	Selected, Count, DestinationSnapshot, ReopenedSnapshot rateFrame406
	Saved, DestinationSetup, Setup                         customizeSaved394
	Expansion                                              templateExpansion404
	ParentBefore, ParentAfter                              struct {
		Text, HTML string
		Same       bool
	}
}
type rateLedger406 struct {
	Version, Status string
	Raw             struct{ Logs, Spans []summaryRaw400 }
	Fixture         struct {
		Token   string
		Records []struct {
			Signal, ID, TraceID, Environment, Namespace, Team, Service, Phase string
			TimeMS                                                            int64
			LegacyOnly                                                        bool
		}
	}
	Cases    []rateCase406
	Controls []struct {
		Name     string
		Request  exploration.Query
		Response exploration.Result
		Status   int
	}
	Facts []struct {
		Name          string
		Before, After rateCounters406
	}
	Counters   rateCounters406
	ProxyCalls []literalWire393
	Errors     []string
}

var rateNames406 = []string{"summary-source-read-default-count", "summary-stage", "summary-child-read", "summary-select-rate-order", "summary-retired-field-handler", "summary-save-template", "summary-held-back", "summary-reopen-definition", "summary-inspect-destination", "summary-apply", "summary-retired-apply", "summary-destination-read", "timeline-stage-definition", "timeline-source-read-default-count", "timeline-select-rate-short-final-zero", "timeline-table-bars-local", "timeline-save-rate", "timeline-reopen-rate", "timeline-reopened-read", "timeline-stage-clipped-interval", "timeline-clipped-child-read", "timeline-retired-stage", "empty-summary", "known-zero-interval"}

func rateCosts406(name string) (source, writes, expand, sql int) {
	switch name {
	case "summary-source-read-default-count", "summary-child-read", "summary-destination-read", "empty-summary":
		source = 1
		sql = 3
	case "timeline-source-read-default-count", "timeline-reopened-read", "timeline-clipped-child-read", "known-zero-interval":
		source = 1
		sql = 4
	case "summary-save-template", "timeline-save-rate":
		writes = 1
	case "summary-inspect-destination":
		expand = 1
	}
	return
}
func rateRead406(read customizeRead394, raw []summaryRaw400) error {
	if read.Status != 200 || !reflect.DeepEqual(read.Request, read.Response.Question) {
		return fmt.Errorf("rate source request changed")
	}
	q, r := read.Request, read.Response
	members, e := summaryMembers400(q, raw)
	if e != nil {
		return e
	}
	count := strconv.Itoa(len(members))
	wantState := "observed"
	if len(members) == 0 {
		wantState = "no_data"
	}
	if r.Count != count || r.State != wantState || r.HasMore != (len(members) > q.Limit) || len(r.Rows) != min(q.Limit, len(members)) || r.Basis == "" || r.CapturedAt == "" {
		return fmt.Errorf("rate independent physical population/page differs")
	}
	ids := map[string]bool{}
	for _, m := range members {
		ids[m.SpanID] = true
	}
	retained := map[string]bool{}
	for _, wire := range r.Rows {
		var row summaryRaw400
		if json.Unmarshal(wire, &row) != nil || !ids[row.SpanID] || retained[row.SpanID] {
			return fmt.Errorf("foreign retained rate record")
		}
		retained[row.SpanID] = true
	}
	for _, d := range []*scopedview.Dataset{r.SummaryDataset, r.TimeDataset} {
		if d == nil {
			if q.BucketMS == 0 {
				continue
			}
			return fmt.Errorf("missing rate interval dataset")
		}
		if d.Revision != r.Basis || d.Scope.Generation != r.Basis || d.CapturedAt != r.CapturedAt || !sameTransport384(d.Scope.Selector, q) || d.Quality.Extent != "complete" || d.Completeness == nil || d.Completeness.Revision != r.Basis || d.Completeness.ScopeID != d.Scope.ID {
			return fmt.Errorf("rate source identity/coverage/clock differs")
		}
		if err := exploration.ValidateRecordRateDataset(d); err != nil {
			return err
		}
		if d.Shape == "time-series" && len(d.Rows) != int((q.ToMS-q.FromMS+q.BucketMS-1)/q.BucketMS) {
			return fmt.Errorf("missing exact interval coverage")
		}
		for index, row := range d.Rows {
			elapsed := q.ToMS - q.FromMS
			n := len(members)
			if d.Shape == "time-series" {
				a, _ := strconv.ParseInt(*row.Values["fromMs"], 10, 64)
				b, _ := strconv.ParseInt(*row.Values["toMs"], 10, 64)
				if a != q.FromMS+int64(index)*q.BucketMS || b != min(a+q.BucketMS, q.ToMS) {
					return fmt.Errorf("interval geometry changed")
				}
				elapsed = b - a
				n = 0
				for _, m := range members {
					at, _ := strconv.ParseInt(m.TimestampNS, 10, 64)
					if at >= a*1000000 && at < b*1000000 {
						n++
					}
				}
			}
			c := strconv.Itoa(n)
			if row.Values["count"] == nil || *row.Values["count"] != c || row.Values["record.rate"] == nil || *row.Values["record.rate"] != rateExpected406(c, elapsed) {
				return fmt.Errorf("rate must equal independently filtered raw count / actual exposure")
			}
		}
	}
	if r.SummaryDataset == nil {
		return fmt.Errorf("missing summary rate")
	}
	return nil
}
func admitRate406(r rateLedger406) error {
	fail := func(s string) error { return fmt.Errorf("recordRate406: %s", s) }
	if r.Version != "goo.recordRate406/1" || r.Status != "passed" || len(r.Errors) != 0 || len(r.Cases) != 2 || len(r.Controls) != 2 || len(r.Raw.Logs) != 12 || len(r.Raw.Spans) != 12 || len(r.Fixture.Records) != 24 || len(r.Facts) != len(rateNames406) {
		return fail("unfinished finite source journey")
	}
	for _, signal := range []string{"logs", "spans"} {
		raw := r.Raw.Logs
		if signal == "spans" {
			raw = r.Raw.Spans
		}
		seen := map[string]bool{}
		for _, row := range raw {
			if seen[row.SpanID] {
				return fail("duplicate raw identity")
			}
			seen[row.SpanID] = true
			found := false
			for _, f := range r.Fixture.Records {
				if f.Signal != signal || f.ID != row.SpanID {
					continue
				}
				attrs := row.LogAttributes
				if signal == "spans" {
					attrs = row.SpanAttributes
				}
				key := "deployment.environment.name"
				if f.LegacyOnly {
					key = "deployment.environment"
				}
				if row.TimestampNS != strconv.FormatInt(f.TimeMS*1000000, 10) || row.TraceID != f.TraceID || row.ServiceName != f.Service || row.ResourceAttributes[key] != f.Environment || row.ResourceAttributes["service.namespace"] != f.Namespace || attrs["team"] != f.Team || attrs["work.class"] != f.Phase {
					return fail("raw fixture membership differs")
				}
				found = true
			}
			if !found {
				return fail("unplanned physical record")
			}
		}
	}
	reads := []customizeRead394{}
	for i, c := range r.Cases {
		if c.Kind != []string{"summary", "timeline"}[i] || c.SourceQuestion.Limit != 2 || c.SourceQuestion.Offset != 0 {
			return fail("wrong case/source family")
		}
		raw := r.Raw.Logs
		if i == 1 {
			raw = r.Raw.Spans
		}
		if !reflect.DeepEqual(c.Source.Request, c.SourceQuestion) {
			return fail("source carrier changed full question")
		}
		if i == 0 {
			if !reflect.DeepEqual(c.Child.Request, c.SourceQuestion) {
				return fail("summary child changed source question")
			}
		} else {
			expectedChild := c.SourceQuestion
			expectedChild.FromMS += 3000
			if !reflect.DeepEqual(c.Child.Request, expectedChild) || !reflect.DeepEqual(c.Reopened.Request, c.SourceQuestion) {
				return fail("timeline child/reopen changed full question")
			}
		}
		rr := []customizeRead394{c.Source, c.Child}
		if i == 0 {
			rr = append(rr, c.Destination)
			if c.DestinationQuestion.Limit != 3 || c.DestinationQuestion.FromMS != c.SourceQuestion.FromMS+500 {
				return fail("destination scope changed")
			}
			if !reflect.DeepEqual(c.Destination.Request, c.DestinationQuestion) || !reflect.DeepEqual(c.Selected.Display.Fields, []string{"record.rate", "count"}) || !reflect.DeepEqual(c.DestinationSnapshot.Display.Fields, []string{"record.rate", "count"}) || !sameTransport384(c.Selected.Snapshot, *c.Child.Response.SummaryDataset) || !sameTransport384(c.DestinationSnapshot.Snapshot, *c.Destination.Response.SummaryDataset) {
				return fail("summary presentation/data moved")
			}
			if !c.ParentAfter.Same || c.ParentBefore.Text != c.ParentAfter.Text || c.ParentBefore.HTML != c.ParentAfter.HTML {
				return fail("held parent replaced")
			}
			if err := templateStored404(c.DestinationSetup); err != nil {
				return err
			}
			if c.Expansion.Status != 200 || c.Expansion.Request.Source == nil || !reflect.DeepEqual(*c.Expansion.Request.Source, c.Saved.Response) || c.Expansion.Request.Query == nil || !reflect.DeepEqual(*c.Expansion.Request.Query, c.DestinationQuestion) {
				return fail("template destination carrier changed")
			}
			expected, err := workbench.ExpandViewTemplate(c.Expansion.Request)
			if err != nil || !reflect.DeepEqual(expected, c.Expansion.Response) {
				return fail("actual rate expansion changed destination/presentation")
			}
		} else {
			rr = append(rr, c.Reopened)
			if c.Selected.Display.Value != "record.rate" || c.Count.Display.Value != "count" || c.ReopenedSnapshot.Display.Value != "record.rate" || !sameTransport384(c.Selected.Snapshot, c.Count.Snapshot) || !sameTransport384(c.Selected.Snapshot, *c.Source.Response.TimeDataset) || !sameTransport384(c.ReopenedSnapshot.Snapshot, *c.Reopened.Response.TimeDataset) || c.Child.Request.FromMS != c.SourceQuestion.FromMS+3000 || c.Child.Request.ToMS != c.SourceQuestion.ToMS {
				return fail("timeline local display/child changed")
			}
			if err := templateStored404(c.Setup); err != nil {
				return err
			}
		}
		for _, read := range rr {
			if err := rateRead406(read, raw); err != nil {
				return err
			}
			reads = append(reads, read)
		}
		if err := templateStored404(c.Saved); err != nil {
			return err
		}
	}
	for i, c := range r.Controls {
		if c.Name != []string{"empty-summary", "known-zero-interval"}[i] || c.Response.Count != "0" {
			return fail("zero controls absent")
		}
		read := customizeRead394{Request: c.Request, Response: c.Response, Status: c.Status}
		raw := r.Raw.Logs
		if i == 1 {
			raw = r.Raw.Spans
		}
		if err := rateRead406(read, raw); err != nil {
			return err
		}
		reads = append(reads, read)
	}
	previous := rateCounters406{}
	for i, f := range r.Facts {
		a, b, c, d := rateCosts406(f.Name)
		if f.Name != rateNames406[i] || f.Before != previous || f.After.Source-f.Before.Source != a || f.After.Writes-f.Before.Writes != b || f.After.Expand-f.Before.Expand != c || f.After.SQL-f.Before.SQL != d {
			return fail("unexpected source/config/SQL transition")
		}
		previous = f.After
	}
	if r.Counters != previous || r.Counters.Source != 8 || r.Counters.Writes != 2 || r.Counters.Expand != 1 || r.Counters.SQL != 28 {
		return fail("wrong finite budgets")
	}
	return summaryPlans400(reads, r.ProxyCalls)
}
func replayRate406(t *testing.T, r rateLedger406, mutate bool) bool {
	t.Helper()
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	costs := []int{}
	sql := []int{}
	for i, f := range r.Facts {
		a, _, _, d := rateCosts406(f.Name)
		costs = append(costs, a)
		sql = append(sql, d)
		n := f.After.Source
		if mutate && f.Name == "summary-select-rate-order" {
			n++
		}
		rec.Record(context.Background(), "step", nil, "obs.phase", i+1, "obs.reads", n, "obs.sql", f.After.SQL)
	}
	spec, _ := filepath.Abs("model/record_rate406.qnt")
	a, _ := json.Marshal(costs)
	b, _ := json.Marshal(sql)
	binding := fmt.Sprintf("spec: %q\nmodule: record_rate406\nconstants: {READS: '%s', SQL: '%s'}\ninvariants: [safety]\nactions:\n  step: {expect: 'phase == {{.obs.phase}} and reads == {{.obs.reads}} and sql == {{.obs.sql}}'}\n", spec, a, b)
	dir := t.TempDir()
	file := filepath.Join(dir, "binding.yaml")
	if e := os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	checker, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: dir})
	if e != nil {
		t.Fatal(e)
	}
	result, e := checker.Check(t.Context(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return result.OK()
}
func TestQuintRecordRateObserved406(t *testing.T) {
	path := os.Getenv("GOO_RECORD_RATE_406_REPORT")
	if path == "" {
		t.Skip("actual record rate journey required")
	}
	wire, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var r rateLedger406
	if json.Unmarshal(wire, &r) != nil {
		t.Fatal("invalid rate report")
	}
	if e = admitRate406(r); e != nil {
		t.Fatal(e)
	}
	mutants := []func(*rateLedger406){func(x *rateLedger406) { x.Raw.Logs[0].TimestampNS = "0" }, func(x *rateLedger406) { x.Cases[0].DestinationQuestion.Limit++ }, func(x *rateLedger406) { x.Cases[0].Selected.Snapshot.Rows[0].Cells["record.rate"].Rate.ElapsedMS = "1" }, func(x *rateLedger406) {
		x.Cases[1].Selected.Snapshot.Rows[3].Cells["record.rate"].Rate.ElapsedMS = "1000"
	}, func(x *rateLedger406) { x.Cases[0].ParentAfter.Same = false }, func(x *rateLedger406) { x.Facts[3].After.Source++ }, func(x *rateLedger406) { x.ProxyCalls[0].Parameters["param_from"] = "0" }}
	for i, m := range mutants {
		var bad rateLedger406
		_ = json.Unmarshal(wire, &bad)
		m(&bad)
		if admitRate406(bad) == nil {
			t.Fatalf("semantic mutant %d escaped", i)
		}
	}
	if !replayRate406(t, r, false) || replayRate406(t, r, true) {
		t.Fatal("actual rate model or local-source negative control incorrect")
	}
	t.Logf("24 raw source records, 24 observed boundaries, exact rates and 28 unchanged source statements; %d semantic controls rejected", len(mutants))
}

func TestRecordRateActorObserved406(t *testing.T) {
	path := os.Getenv("GOO_RECORD_RATE_406_ACTOR_REPORT")
	if path == "" {
		t.Skip("actual Bombadil record rate report required")
	}
	mainPath := os.Getenv("GOO_RECORD_RATE_406_REPORT")
	if mainPath == "" {
		t.Fatal("actor requires its actual raw source bank")
	}
	wire, e := os.ReadFile(mainPath)
	if e != nil {
		t.Fatal(e)
	}
	var main rateLedger406
	if json.Unmarshal(wire, &main) != nil {
		t.Fatal("invalid main report")
	}
	if e = admitRate406(main); e != nil {
		t.Fatal(e)
	}
	wire, e = os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var actor struct {
		Version             string
		SourceBank          string
		Question            exploration.Query
		Fields              []string
		Setup               []customizeSaved394
		SQLBefore, SQLAfter int
		ProxyCalls          []literalWire393
		Actor               struct {
			Snapshots int
			Phases    []string
			Complete  struct {
				Version, Phase, Population string
				SummaryCount               int
				Summary                    rateFrame406
				Overflow                   bool
				Calls                      []struct {
					Path, Method      string
					Request, Response json.RawMessage
					Status            int
				}
			}
		}
	}
	if json.Unmarshal(wire, &actor) != nil {
		t.Fatal("invalid actor report")
	}
	c := actor.Actor.Complete
	if actor.Version != "goo.recordRate406ActorReport/1" || actor.SourceBank != main.Fixture.Token || c.Version != "goo.recordRate406Actor/1" || c.Phase != "complete" || c.Population != "applied" || c.SummaryCount != 1 || c.Overflow || actor.Actor.Snapshots < 1 || len(actor.Setup) != 2 || !reflect.DeepEqual(actor.Question, main.Cases[0].DestinationQuestion) || !reflect.DeepEqual(actor.Fields, []string{"record.rate", "count"}) || !reflect.DeepEqual(c.Summary.Display.Fields, actor.Fields) || actor.SQLBefore != 0 || actor.SQLAfter != 3 {
		t.Fatal("unfinished or scope-shifted actor")
	}
	for _, saved := range actor.Setup {
		if e = templateStored404(saved); e != nil {
			t.Fatal(e)
		}
	}
	reads := []customizeRead394{}
	expands := 0
	for _, call := range c.Calls {
		if call.Method != "POST" {
			continue
		}
		if call.Path == "/api/explore" {
			var read customizeRead394
			read.Status = call.Status
			if json.Unmarshal(call.Request, &read.Request) != nil || json.Unmarshal(call.Response, &read.Response) != nil || !reflect.DeepEqual(read.Request, actor.Question) {
				t.Fatal("actor source carrier changed")
			}
			if e = rateRead406(read, main.Raw.Logs); e != nil {
				t.Fatal(e)
			}
			if !sameTransport384(c.Summary.Snapshot, *read.Response.SummaryDataset) {
				t.Fatal("actor DOM source snapshot differs")
			}
			reads = append(reads, read)
		} else if call.Path == "/api/view-templates/expand" {
			expands++
			var request workbench.ViewTemplateExpansionRequest
			var response workbench.ViewTemplateExpansion
			if call.Status != 200 || json.Unmarshal(call.Request, &request) != nil || json.Unmarshal(call.Response, &response) != nil {
				t.Fatal("actor expansion malformed")
			}
			expected, e := workbench.ExpandViewTemplate(request)
			if e != nil || !reflect.DeepEqual(expected, response) || !reflect.DeepEqual(response.View.Queries, []exploration.Query{actor.Question}) {
				t.Fatal("actor expansion changed scope")
			}
		} else {
			t.Fatal("actor performed unexpected write/source action", call.Path)
		}
	}
	if len(reads) != 1 || expands != 1 {
		t.Fatal("actor source/expand budget")
	}
	if e = summaryPlans400(reads, actor.ProxyCalls); e != nil {
		t.Fatal(e)
	}
}
