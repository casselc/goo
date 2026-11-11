package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/scopedview"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type templateCounters404 struct{ Source, Store, ConfigWrites, Expand, SQL int }
type templateFact404 struct {
	Name          string
	Before, After templateCounters404
}
type templateExpansion404 struct {
	Request  workbench.ViewTemplateExpansionRequest
	Response workbench.ViewTemplateExpansion
	Status   int
}
type templateFrame404 struct {
	Name     string
	Snapshot scopedview.Dataset
	Display  struct {
		Fields  []string
		Title   string
		Compact bool
	}
	FieldOrder []string
}
type templateCase404 struct {
	Kind, Title                                     string
	SourceQuestion, DestinationQuestion             exploration.Query
	Source, ChildRun, DestinationRun, ReopenRun     customizeRead394
	SourceSaved, DestinationSetup, DestinationSaved customizeSaved394
	Expansion, PendingExpansion                     templateExpansion404
	Fields                                          []string
	AppliedComposition                              workbench.ViewComposition
	SummarySnapshots                                []templateFrame404
	ParentBefore, ParentAfter                       json.RawMessage
	ParentEvidenceBefore, ParentEvidenceAfter       struct {
		Text, HTML string
		Same       bool
	}
}
type templateLedger404 struct {
	Version, Status string
	Fixture         json.RawMessage
	Raw             struct{ Logs, Spans []summaryRaw400 }
	Cases           []templateCase404
	Controls        []struct {
		Name     string
		Question exploration.Query
		Read     customizeRead394
		Snapshot scopedview.Dataset
	}
	Facts      []templateFact404
	Counters   templateCounters404
	ProxyCalls []literalWire393
	Errors     []string
}

func templateNames404() []string {
	out := []string{}
	for _, kind := range []string{"count-logs", "duration"} {
		for _, suffix := range []string{"source-read", "customize-stage", "child-read", "author-fields-template", "save-template", "held-back", "reopen-template"} {
			out = append(out, kind+"-"+suffix)
		}
		if kind == "count-logs" {
			out = append(out, kind+"-pending-inspect-retired")
		}
		for _, suffix := range []string{"destination-inspect", "hidden-inert-apply", "apply-presentation", "retired-apply", "destination-read", "save-destination", "reopen-destination", "reopened-read"} {
			out = append(out, kind+"-"+suffix)
		}
	}
	return append(out, "control-empty", "control-zero")
}
func templateCost404(name string) (source, store, config, expand int) {
	if strings.HasSuffix(name, "-source-read") || strings.HasSuffix(name, "-child-read") || strings.HasSuffix(name, "-destination-read") || strings.HasSuffix(name, "-reopened-read") || strings.HasPrefix(name, "control-") {
		source = 1
	}
	if strings.HasSuffix(name, "-save-template") || strings.HasSuffix(name, "-save-destination") {
		config = 1
	}
	if strings.HasSuffix(name, "-reopen-template") || strings.HasSuffix(name, "-reopen-destination") {
		store = 1
	}
	if strings.HasSuffix(name, "-pending-inspect-retired") {
		store = 2
		expand = 1
	}
	if strings.HasSuffix(name, "-destination-inspect") {
		store = 1
		expand = 1
		if strings.HasPrefix(name, "duration-") {
			store = 2
		}
	}
	return
}
func templateStored404(saved customizeSaved394) error {
	if saved.Status != 201 || saved.Request.ID != "" || saved.Request.Validate() != nil || saved.Response.Validate() != nil {
		return fmt.Errorf("invalid saved template carrier")
	}
	id, err := workbench.ViewContentID(saved.Request)
	if err != nil || id != saved.Response.ID {
		return fmt.Errorf("saved template content identity differs")
	}
	view := saved.Response
	view.ID = ""
	if !reflect.DeepEqual(view, saved.Request) {
		return fmt.Errorf("saved template response changed configuration")
	}
	return nil
}
func templateBinding404(view workbench.ExplorationView, q exploration.Query, fields []string) error {
	binding := workbench.ViewTemplateBinding{Version: "goo.template/1", Group: "none", Profile: "ordinary-summary"}
	if view.Validate() != nil || !reflect.DeepEqual(view.Template, &binding) || !reflect.DeepEqual(view.Queries, []exploration.Query{q}) || view.Composition == nil || len(view.Composition.Children) != 2 {
		return fmt.Errorf("record-summary template profile or complete question changed")
	}
	summary := summaryLeaf400(view.Composition)
	if summary == nil || !reflect.DeepEqual(summary.Display.Fields, fields) {
		return fmt.Errorf("template selected summary field order changed")
	}
	return nil
}
func admitTemplate404(r templateLedger404) error {
	fail := func(s string) error { return fmt.Errorf("summaryTemplate404: %s", s) }
	if r.Version != "goo.summaryTemplate404/1" || r.Status != "passed" || len(r.Cases) != 2 || len(r.Controls) != 2 || len(r.Errors) != 0 {
		return fail("unfinished finite journey")
	}
	physical := summaryLedger400{}
	if json.Unmarshal(r.Fixture, &physical.Fixture) != nil {
		return fail("invalid fixture bank")
	}
	physical.Raw.Logs = r.Raw.Logs
	physical.Raw.Spans = r.Raw.Spans
	if err := summaryPhysical400(physical); err != nil {
		return err
	}
	reads := []customizeRead394{}
	for i, c := range r.Cases {
		if c.Kind != []string{"count-logs", "duration"}[i] || c.SourceQuestion.Offset != 0 || c.DestinationQuestion.Offset != 0 || c.SourceQuestion.Limit != 3 || c.DestinationQuestion.Limit != 5 || len(c.SourceQuestion.GroupBy) != 0 || len(c.DestinationQuestion.GroupBy) != 0 || c.SourceQuestion.Numeric != nil || c.DestinationQuestion.Numeric != nil || reflect.DeepEqual(c.SourceQuestion, c.DestinationQuestion) {
			return fail("wrong finite root/scope family")
		}
		fields := []string{"count"}
		raw := r.Raw.Logs
		if i == 1 {
			fields = []string{"status.error_fraction", "count", "duration.mean"}
			raw = r.Raw.Spans
		}
		if !reflect.DeepEqual(c.Fields, fields) {
			return fail("unexpected selected facts")
		}
		for j, read := range []customizeRead394{c.Source, c.ChildRun, c.DestinationRun, c.ReopenRun} {
			q := c.SourceQuestion
			if j >= 2 {
				q = c.DestinationQuestion
			}
			if !reflect.DeepEqual(read.Request, q) {
				return fail("read changed complete source or destination question")
			}
			if err := summaryRead400(read, raw); err != nil {
				return err
			}
			reads = append(reads, read)
		}
		for _, saved := range []customizeSaved394{c.SourceSaved, c.DestinationSetup, c.DestinationSaved} {
			if err := templateStored404(saved); err != nil {
				return err
			}
		}
		if err := templateBinding404(c.SourceSaved.Response, c.SourceQuestion, fields); err != nil {
			return err
		}
		if err := templateBinding404(c.DestinationSaved.Response, c.DestinationQuestion, fields); err != nil {
			return err
		}
		if c.DestinationSetup.Response.Template != nil || !reflect.DeepEqual(c.DestinationSetup.Response.Queries, []exploration.Query{c.DestinationQuestion}) {
			return fail("destination setup imported donor intent")
		}
		if c.Expansion.Status != 200 || c.Expansion.Request.Source == nil || c.Expansion.Request.Query == nil || !reflect.DeepEqual(*c.Expansion.Request.Source, c.SourceSaved.Response) || !reflect.DeepEqual(*c.Expansion.Request.Query, c.DestinationQuestion) || c.Expansion.Request.BaselineCapture != nil {
			return fail("expansion request changed source/destination or invented baseline")
		}
		expected, err := workbench.ExpandViewTemplate(c.Expansion.Request)
		if err != nil || !reflect.DeepEqual(expected, c.Expansion.Response) {
			return fail("actual expansion differs from canonical source presentation")
		}
		if !reflect.DeepEqual(c.AppliedComposition, *expected.View.Composition) || !reflect.DeepEqual(c.DestinationSaved.Response.Composition, c.SourceSaved.Response.Composition) {
			return fail("Apply or republish changed summary/records presentation")
		}
		if i == 0 {
			if c.PendingExpansion.Status != 200 || !reflect.DeepEqual(c.PendingExpansion.Response.View.Queries, []exploration.Query{c.DestinationQuestion}) {
				return fail("held expansion control missing actual response")
			}
		}
		if !bytesEqualJSON404(c.ParentBefore, c.ParentAfter) || !c.ParentEvidenceAfter.Same || c.ParentEvidenceBefore.Text != c.ParentEvidenceAfter.Text || c.ParentEvidenceBefore.HTML != c.ParentEvidenceAfter.HTML {
			return fail("held Back replaced source evidence DOM or data")
		}
		if len(c.SummarySnapshots) != 4 {
			return fail("missing actual summary render stages")
		}
		for j, frame := range c.SummarySnapshots {
			if frame.Name != []string{"initial", "customized", "destination", "reopened"}[j] || !frame.Display.Compact || !reflect.DeepEqual(frame.Display.Fields, frame.FieldOrder) || len(frame.Display.Fields) == 0 {
				return fail("rendered summary order/compact contract changed")
			}
			read := c.ChildRun
			if j == 2 {
				read = c.DestinationRun
			}
			if j == 3 {
				read = c.ReopenRun
			}
			if err := summaryDataset400(read.Request, raw, &frame.Snapshot, read.Response.Basis, read.Response.CapturedAt); err != nil {
				return err
			}
			if j > 0 && !reflect.DeepEqual(frame.Display.Fields, fields) {
				return fail("rendered selected field order changed")
			}
		}
	}
	for i, c := range r.Controls {
		if c.Name != []string{"empty", "zero"}[i] || !reflect.DeepEqual(c.Question, c.Read.Request) {
			return fail("summary support control changed")
		}
		if err := summaryRead400(c.Read, r.Raw.Spans); err != nil {
			return err
		}
		if err := summaryDataset400(c.Question, r.Raw.Spans, &c.Snapshot, c.Read.Response.Basis, c.Read.Response.CapturedAt); err != nil {
			return err
		}
		reads = append(reads, c.Read)
		mean := c.Snapshot.Rows[0].Values["duration.mean"]
		if i == 0 && (c.Read.Response.Count != "0" || mean != nil || c.Snapshot.Rows[0].Cells["duration.mean"].State != "no_data") {
			return fail("empty duration claimed a numeric value")
		}
		if i == 1 && (mean == nil || *mean != "0" || c.Snapshot.Rows[0].Cells["duration.mean"].State != "derived") {
			return fail("real zero duration became unavailable")
		}
	}
	if err := summaryPlans400(reads, r.ProxyCalls); err != nil {
		return err
	}
	names := templateNames404()
	if len(r.Facts) != len(names) {
		return fail("missing local/read boundaries")
	}
	for i, f := range r.Facts {
		if f.Name != names[i] || i > 0 && f.Before != r.Facts[i-1].After {
			return fail("phase/counter continuity changed")
		}
		a, b, c, d := templateCost404(f.Name)
		if f.After.Source-f.Before.Source != a || f.After.Store-f.Before.Store != b || f.After.ConfigWrites-f.Before.ConfigWrites != c || f.After.Expand-f.Before.Expand != d || a == 0 && f.After.SQL != f.Before.SQL {
			return fail("source/store/publication/expand budget changed: " + f.Name)
		}
	}
	if r.Counters != r.Facts[len(r.Facts)-1].After || r.Counters.Source != 10 || r.Counters.ConfigWrites != 4 || r.Counters.Expand != 3 {
		return fail("final operation budget changed")
	}
	return nil
}
func bytesEqualJSON404(a, b json.RawMessage) bool {
	var x, y any
	return json.Unmarshal(a, &x) == nil && json.Unmarshal(b, &y) == nil && reflect.DeepEqual(x, y)
}
func replayTemplate404(t *testing.T, r templateLedger404, mutate bool) bool {
	t.Helper()
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	reads, stores, configs, expands := []int{}, []int{}, []int{}, []int{}
	for i, f := range r.Facts {
		a, b, c, d := templateCost404(f.Name)
		reads = append(reads, a)
		stores = append(stores, b)
		configs = append(configs, c)
		expands = append(expands, d)
		source := f.After.Source
		if mutate && f.Name == "count-logs-apply-presentation" {
			source++
		}
		rec.Record(context.Background(), "step", nil, "obs.phase", i+1, "obs.reads", source, "obs.stores", f.After.Store, "obs.configs", f.After.ConfigWrites, "obs.expands", f.After.Expand)
	}
	list := func(v []int) string { b, _ := json.Marshal(v); return string(b) }
	spec, err := filepath.Abs("model/summary_template404.qnt")
	if err != nil {
		t.Fatal(err)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: summary_template404\nconstants: {READS: '%s', STORES: '%s', CONFIGS: '%s', EXPANDS: '%s'}\ninvariants: [safety]\nactions:\n  step: {expect: 'phase == {{.obs.phase}} and reads == {{.obs.reads}} and stores == {{.obs.stores}} and configs == {{.obs.configs}} and expands == {{.obs.expands}}'}\n", spec, list(reads), list(stores), list(configs), list(expands))
	dir := t.TempDir()
	file := filepath.Join(dir, "binding.yaml")
	if err = os.WriteFile(file, []byte(binding), 0600); err != nil {
		t.Fatal(err)
	}
	checker, err := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	result, err := checker.Check(t.Context(), sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	return result.OK()
}
func TestQuintSummaryTemplateObserved404(t *testing.T) {
	path := os.Getenv("GOO_SUMMARY_TEMPLATE_404_REPORT")
	if path == "" {
		t.Skip("actual summary template journey required")
	}
	wire, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var r templateLedger404
	if json.Unmarshal(wire, &r) != nil {
		t.Fatal("invalid summary template report")
	}
	if err = admitTemplate404(r); err != nil {
		t.Fatal(err)
	}
	mutants := []func(*templateLedger404){func(x *templateLedger404) { x.Cases[0].DestinationQuestion.Limit++ }, func(x *templateLedger404) { x.Cases[1].Expansion.Response.View.Queries[0].FromMS++ }, func(x *templateLedger404) {
		summaryLeaf400(x.Cases[1].SourceSaved.Response.Composition).Display.Fields = []string{"count"}
	}, func(x *templateLedger404) { x.Cases[0].SourceSaved.Response.Template.Group = "destination" }, func(x *templateLedger404) { x.Cases[0].ParentEvidenceAfter.Same = false }, func(x *templateLedger404) { x.Cases[1].SummarySnapshots[2].Snapshot.CapturedAt = "foreign" }, func(x *templateLedger404) { x.Controls[1].Snapshot.Rows[0].Values["duration.mean"] = nil }, func(x *templateLedger404) { x.Raw.Spans[0].StatusCode = "Unknown" }, func(x *templateLedger404) { x.Facts[10].After.Source++ }, func(x *templateLedger404) { x.ProxyCalls[0].Parameters["param_from"] = "1" }}
	for i, mutate := range mutants {
		var bad templateLedger404
		json.Unmarshal(wire, &bad)
		mutate(&bad)
		if admitTemplate404(bad) == nil {
			t.Fatalf("semantic mutant %d escaped", i)
		}
	}
	if !replayTemplate404(t, r, false) || replayTemplate404(t, r, true) {
		t.Fatal("actual summary template model or local-read negative control incorrect")
	}
	t.Logf("36 physical records, two record-summary template families, %d boundaries; ten semantic mutants and local-read control rejected", len(r.Facts))
}

type templateActor404 struct {
	Version, SourceBank string
	Question            exploration.Query
	Fields              []string
	Setup               []struct {
		Kind              string
		Request, Response workbench.ExplorationView
		Status            int
	}
	SQLBefore, SQLAfter int
	ProxyCalls          []literalWire393
	Actor               struct {
		Snapshots int
		Phases    []string
		Complete  struct {
			Version, Phase, Population string
			SummaryCount               int
			Overflow                   bool
			Summary                    struct {
				Snapshot scopedview.Dataset
				Display  struct {
					Fields  []string
					Compact bool
				}
				Order []string
			}
			Calls []struct {
				Path, Method      string
				Status            int
				Request, Response json.RawMessage
			}
		}
	}
}

func admitTemplateActor404(main templateLedger404, a templateActor404) error {
	fail := func(s string) error { return fmt.Errorf("summaryTemplate404 actor: %s", s) }
	var seed struct{ Token string }
	json.Unmarshal(main.Fixture, &seed)
	c := main.Cases[1]
	if a.Version != "goo.summaryTemplate404ActorReport/1" || a.SourceBank != seed.Token || !reflect.DeepEqual(a.Question, c.DestinationQuestion) || !reflect.DeepEqual(a.Fields, c.Fields) || len(a.Setup) != 2 {
		return fail("source bank, fixed question or setup changed")
	}
	for i, s := range a.Setup {
		saved := customizeSaved394{Request: s.Request, Response: s.Response, Status: s.Status}
		if err := templateStored404(saved); err != nil {
			return err
		}
		want := c.SourceSaved.Request
		if i == 1 {
			want = c.DestinationSetup.Request
		}
		if !reflect.DeepEqual(s.Request, want) {
			return fail("actor setup differs from independent main question/configuration")
		}
	}
	complete := a.Actor.Complete
	if a.Actor.Snapshots < 6 || complete.Version != "goo.summaryTemplate404Actor/1" || complete.Phase != "complete" || complete.Population != "applied" || complete.SummaryCount != 1 || complete.Overflow || !reflect.DeepEqual(a.Actor.Phases, []string{"initial", "library", "preview", "applied", "read", "complete"}) {
		return fail("driver incomplete or incorrect viewport/staging")
	}
	var read customizeRead394
	var expansion templateExpansion404
	reads, expands := 0, 0
	for _, call := range complete.Calls {
		if call.Method != "POST" {
			continue
		}
		switch call.Path {
		case "/api/explore":
			reads++
			read.Status = call.Status
			if json.Unmarshal(call.Request, &read.Request) != nil || json.Unmarshal(call.Response, &read.Response) != nil {
				return fail("invalid source call")
			}
		case "/api/view-templates/expand":
			expands++
			expansion.Status = call.Status
			if json.Unmarshal(call.Request, &expansion.Request) != nil || json.Unmarshal(call.Response, &expansion.Response) != nil {
				return fail("invalid expansion call")
			}
		default:
			return fail("unexpected source/publication operation")
		}
	}
	if reads != 1 || expands != 1 || !reflect.DeepEqual(read.Request, a.Question) || expansion.Status != 200 || expansion.Request.Source == nil || !reflect.DeepEqual(*expansion.Request.Source, a.Setup[0].Response) || expansion.Request.Query == nil || !reflect.DeepEqual(*expansion.Request.Query, a.Question) {
		return fail("explicit operation budget or whole destination query changed")
	}
	expected, err := workbench.ExpandViewTemplate(expansion.Request)
	if err != nil || !reflect.DeepEqual(expected, expansion.Response) {
		return fail("actor expansion differs from source presentation")
	}
	if err = summaryRead400(read, main.Raw.Spans); err != nil {
		return err
	}
	if err = summaryDataset400(read.Request, main.Raw.Spans, &complete.Summary.Snapshot, read.Response.Basis, read.Response.CapturedAt); err != nil {
		return err
	}
	if !complete.Summary.Display.Compact || !reflect.DeepEqual(complete.Summary.Display.Fields, c.Fields) || !reflect.DeepEqual(complete.Summary.Order, c.Fields) {
		return fail("summary selected order/compact display changed")
	}
	if a.SQLAfter-a.SQLBefore != 3 || len(a.ProxyCalls) != 3 {
		return fail("actor physical statement budget changed")
	}
	return summaryPlans400([]customizeRead394{read}, a.ProxyCalls)
}
func TestBombadilSummaryTemplateObserved404(t *testing.T) {
	mainPath, actorPath := os.Getenv("GOO_SUMMARY_TEMPLATE_404_REPORT"), os.Getenv("GOO_SUMMARY_TEMPLATE_404_ACTOR")
	if mainPath == "" || actorPath == "" {
		t.Skip("actual summary template main and actor reports required")
	}
	wire, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}
	var main templateLedger404
	if json.Unmarshal(wire, &main) != nil {
		t.Fatal("invalid main report")
	}
	if err = admitTemplate404(main); err != nil {
		t.Fatal(err)
	}
	wire, err = os.ReadFile(actorPath)
	if err != nil {
		t.Fatal(err)
	}
	var actor templateActor404
	if json.Unmarshal(wire, &actor) != nil {
		t.Fatal("invalid actor report")
	}
	if err = admitTemplateActor404(main, actor); err != nil {
		t.Fatal(err)
	}
	t.Logf("%d actor snapshots; one explicit expansion, one source read and exact physical statements", actor.Actor.Snapshots)
}
