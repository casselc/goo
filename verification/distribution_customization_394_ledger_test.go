package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/scopedview"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type customizeRaw394 struct {
	ServiceName, SpanID, TraceID, Body, SpanName, SpanKind string
	TimestampNS                                            string `json:"timestamp_ns"`
	ResourceAttributes, LogAttributes, SpanAttributes      map[string]string
}
type customizeRead394 struct {
	Request  exploration.Query
	Response exploration.Result
	Status   int
}
type customizeParent394 struct {
	Find     string
	Snapshot scopedview.Dataset
	Display  json.RawMessage
}
type customizeWrapper394 struct {
	Before struct {
		Events                           int
		Connected, Exists, Hidden, Inert bool
	}
	After   struct{ Events int }
	Invoked bool
}
type customizeSaved394 struct {
	Request, Response workbench.ExplorationView
	Status            int
}
type customizeCase394 struct {
	Signal                                        string
	SourceQuestion, DestinationQuestion           exploration.Query
	Source, ChildRun, Replacement, DestinationRun customizeRead394
	ParentBefore, ParentAfter                     customizeParent394
	Hidden, Inert, Pending, Superseded            customizeWrapper394
	Staged                                        struct {
		Event struct {
			Question exploration.Query
			View     workbench.ExplorationView
			Stage    bool
		}
		ChartCount int
	}
	Customized, DestinationSnapshot struct {
		Snapshot scopedview.Dataset
		Display  json.RawMessage
	}
	Saved, DestinationSetup customizeSaved394
}
type customizeLedger394 struct {
	Version string
	Fixture struct {
		Token, Environment, Namespace string
		Records                       []struct {
			Signal, Team, Environment, Namespace, Service, Phase, ID, TraceID string
			TimeMS                                                            int64
		}
	}
	Raw      struct{ Logs, Spans []customizeRaw394 }
	Cases    []customizeCase394
	Facts    []patternFact391
	Controls []struct {
		Signal, Name    string
		Source          customizeRead394
		Request         exploration.Query
		Status, Present int
		Disabled        *bool
		Response        json.RawMessage
	}
	Errors []string
	Actor  json.RawMessage
}

// Independent ordinary-record membership uses raw source fields and the existing
// verification predicate interpreter; it does not consult rendered row labels.
func customizeMembers394(q exploration.Query, rows []customizeRaw394) (map[string]customizeRaw394, map[string]int, error) {
	selected := map[string]customizeRaw394{}
	groups := map[string]int{}
	if q.Validate() != nil || len(q.GroupBy) != 1 {
		return nil, nil, fmt.Errorf("invalid ordinary source question")
	}
	for _, row := range rows {
		at, e := strconv.ParseInt(row.TimestampNS, 10, 64)
		if e != nil {
			return nil, nil, e
		}
		if at < q.FromMS*1000000 || at >= q.ToMS*1000000 {
			continue
		}
		mode, name := 0, q.Environment
		if s := q.EnvironmentSelector; s != nil {
			name = s.Scope.Value
			if s.Policy == "effective" {
				mode = 2
			}
			if s.Scope.Mode == "empty" {
				mode = 3
			} else if s.Scope.Mode == "unreported" {
				mode = 4
			}
		}
		if !continuationEnvironment352(mode, name, row.ResourceAttributes) {
			continue
		}
		attrs := row.LogAttributes
		if q.Signal == "spans" {
			attrs = row.SpanAttributes
		}
		values := map[exploration.Field]string{{Kind: "column", Key: "service"}: row.ServiceName, {Kind: "column", Key: "span_id"}: row.SpanID, {Kind: "column", Key: "trace_id"}: row.TraceID, {Kind: "column", Key: "name"}: row.SpanName, {Kind: "column", Key: "kind"}: row.SpanKind, {Kind: "column", Key: "body"}: row.Body}
		for k, v := range attrs {
			values[exploration.Field{Kind: "attribute", Key: k}] = v
		}
		for k, v := range row.ResourceAttributes {
			values[exploration.Field{Kind: "resource", Key: k}] = v
		}
		if !hdPredicate(q.Where, values) {
			continue
		}
		if _, exists := selected[row.SpanID]; exists {
			return nil, nil, fmt.Errorf("duplicate raw record identity")
		}
		selected[row.SpanID] = row
		var category *string
		if value, ok := values[q.GroupBy[0]]; ok {
			copy := value
			category = &copy
		}
		groups[continuationJSON352(category)]++
	}
	return selected, groups, nil
}
func admitCustomizeRead394(read customizeRead394, rows []customizeRaw394) error {
	if read.Status != 200 || !reflect.DeepEqual(read.Request, read.Response.Question) || !peLiveHex(read.Response.Basis) || !peLiveClock(read.Response.CapturedAt) {
		return fmt.Errorf("source read identity/clock changed")
	}
	members, groups, e := customizeMembers394(read.Request, rows)
	if e != nil {
		return e
	}
	r := read.Response
	if r.Count != strconv.Itoa(len(members)) || len(r.Rows) != min(read.Request.Limit, max(0, len(members)-read.Request.Offset)) || r.HasMore != (len(members) > read.Request.Offset+len(r.Rows)) {
		return fmt.Errorf("source count/page membership changed")
	}
	state := "observed"
	if len(members) == 0 {
		state = "no_data"
	}
	if r.State != state {
		return fmt.Errorf("source population state changed")
	}
	seen := map[string]bool{}
	for _, wire := range r.Rows {
		var actual struct {
			SpanID, TraceID, TimestampNS, Body, SpanName, ServiceName string
			Attributes, ResourceAttributes                            map[string]string
		}
		var fields map[string]json.RawMessage
		if json.Unmarshal(wire, &fields) != nil {
			return fmt.Errorf("source row invalid")
		}
		if e = json.Unmarshal(wire, &actual); e != nil {
			return e
		}
		json.Unmarshal(fields["timestamp_ns"], &actual.TimestampNS)
		raw, ok := members[actual.SpanID]
		attrs := raw.LogAttributes
		if read.Request.Signal == "spans" {
			attrs = raw.SpanAttributes
		}
		if !ok || seen[actual.SpanID] || actual.TraceID != raw.TraceID || actual.TimestampNS != raw.TimestampNS || actual.ServiceName != raw.ServiceName || !reflect.DeepEqual(actual.Attributes, attrs) || !reflect.DeepEqual(actual.ResourceAttributes, raw.ResourceAttributes) {
			return fmt.Errorf("returned record differs from physical source")
		}
		seen[actual.SpanID] = true
	}
	observed := map[string]int{}
	for _, wire := range r.Groups {
		var group struct {
			G0    *string
			Count string
		}
		if json.Unmarshal(wire, &group) != nil {
			return fmt.Errorf("group shape invalid")
		}
		n, e := strconv.Atoi(group.Count)
		if e != nil {
			return e
		}
		observed[continuationJSON352(group.G0)] = n
	}
	if !reflect.DeepEqual(observed, groups) {
		return fmt.Errorf("full group partition differs from physical membership")
	}
	if r.ViewDataset != nil {
		d := r.ViewDataset
		var q exploration.Query
		if json.Unmarshal(d.Scope.Selector, &q) != nil || !reflect.DeepEqual(q, read.Request) || d.Revision != r.Basis || d.CapturedAt != r.CapturedAt || d.Partition == nil || d.Partition.Total != r.Count || d.Completeness == nil {
			return fmt.Errorf("source dataset lost exact population/basis/clock")
		}
		full := map[string]int{}
		for _, row := range d.Rows {
			if row.Values["count"] == nil {
				return fmt.Errorf("supplied group count missing")
			}
			count, e := strconv.Atoi(*row.Values["count"])
			if e != nil {
				return e
			}
			full[continuationJSON352(row.Values["category"])] = count
		}
		if !reflect.DeepEqual(full, groups) {
			return fmt.Errorf("supplied dataset lost full group counts")
		}
	}
	return nil
}
func admitCustomization394(r customizeLedger394) error {
	fail := func(s string) error { return fmt.Errorf("customize394: %s", s) }
	if r.Version != "goo.distributionCustomize394/1" || len(r.Errors) != 0 || len(r.Cases) != 2 || len(r.Raw.Logs) != 18 || len(r.Raw.Spans) != 18 || len(r.Fixture.Records) != 36 {
		return fail("finite actual ordinary bank absent")
	}
	for _, signal := range []string{"logs", "spans"} {
		rows := r.Raw.Logs
		if signal == "spans" {
			rows = r.Raw.Spans
		}
		for _, raw := range rows {
			found := false
			for _, seed := range r.Fixture.Records {
				if seed.Signal != signal || seed.ID != raw.SpanID {
					continue
				}
				found = true
				attrs := raw.LogAttributes
				if signal == "spans" {
					attrs = raw.SpanAttributes
				}
				if raw.TraceID != seed.TraceID || raw.TimestampNS != strconv.FormatInt(seed.TimeMS*1000000, 10) || raw.ServiceName != seed.Service || raw.ResourceAttributes["deployment.environment.name"] != seed.Environment || raw.ResourceAttributes["service.namespace"] != seed.Namespace || attrs["team"] != seed.Team || attrs["work.class"] != seed.Phase {
					return fail("raw physical identity differs from OTLP seed")
				}
			}
			if !found {
				return fail("foreign physical row")
			}
		}
	}
	for index, c := range r.Cases {
		if c.Signal != []string{"logs", "spans"}[index] || c.SourceQuestion.Signal != c.Signal || c.DestinationQuestion.Signal != c.Signal || reflect.DeepEqual(c.SourceQuestion, c.DestinationQuestion) {
			return fail("signal/source/destination case changed")
		}
		rows := r.Raw.Logs
		if c.Signal == "spans" {
			rows = r.Raw.Spans
		}
		for _, read := range []customizeRead394{c.Source, c.ChildRun, c.Replacement, c.DestinationRun} {
			if e := admitCustomizeRead394(read, rows); e != nil {
				return e
			}
		}
		if !reflect.DeepEqual(c.Source.Request, c.SourceQuestion) || !reflect.DeepEqual(c.ChildRun.Request, c.SourceQuestion) || !reflect.DeepEqual(c.Replacement.Request, c.SourceQuestion) || !reflect.DeepEqual(c.DestinationRun.Request, c.DestinationQuestion) {
			return fail("read changed complete fixed question")
		}
		if c.Staged.ChartCount != 0 || !c.Staged.Event.Stage || !reflect.DeepEqual(c.Staged.Event.Question, c.SourceQuestion) || !reflect.DeepEqual(c.Staged.Event.View.Queries, []exploration.Query{c.SourceQuestion}) || c.Staged.Event.View.Validate() != nil {
			return fail("Customize borrowed evidence or changed question")
		}
		if c.ParentBefore.Find != "read" || !sameTransport384(c.ParentBefore, c.ParentAfter) || !sameTransport384(c.ParentBefore.Snapshot, c.Source.Response.ViewDataset) || !sameTransport384(c.Customized.Snapshot, c.ChildRun.Response.ViewDataset) || !sameTransport384(c.DestinationSnapshot.Snapshot, c.DestinationRun.Response.ViewDataset) {
			return fail("local/held presentation changed source data or clocks")
		}
		for _, guard := range []customizeWrapper394{c.Hidden, c.Inert, c.Pending, c.Superseded} {
			if !guard.Invoked || guard.Before.Events != guard.After.Events {
				return fail("retired/hidden/inert producer acquired continuation authority")
			}
		}
		if !c.Hidden.Before.Connected || !c.Hidden.Before.Hidden || !c.Inert.Before.Connected || !c.Inert.Before.Inert {
			return fail("wrapper refusal was not exercised on actual attached producer")
		}
		for _, saved := range []customizeSaved394{c.Saved, c.DestinationSetup} {
			if saved.Status != 201 || saved.Response.Validate() != nil || !reflect.DeepEqual(saved.Request.Queries, saved.Response.Queries) {
				return fail("published configuration altered question")
			}
			copy := saved.Response
			id := copy.ID
			copy.ID = ""
			hash, e := workbench.ViewContentID(copy)
			if e != nil || hash != id {
				return fail("configuration content identity changed")
			}
		}
		if !reflect.DeepEqual(c.Saved.Response.Queries, []exploration.Query{c.SourceQuestion}) || !reflect.DeepEqual(c.DestinationSetup.Response.Queries, []exploration.Query{c.DestinationQuestion}) {
			return fail("source configuration replaced destination scope")
		}
	}
	if len(r.Controls) != 6 {
		return fail("failed/empty/offset acceptance controls missing")
	}
	for _, control := range r.Controls {
		if control.Name == "failed" {
			if control.Status < 400 || control.Present != 0 {
				return fail("failed source acquired continuation")
			}
			continue
		}
		rows := r.Raw.Logs
		if control.Signal == "spans" {
			rows = r.Raw.Spans
		}
		if e := admitCustomizeRead394(control.Source, rows); e != nil {
			return e
		}
		if control.Present > 0 && (control.Disabled == nil || !*control.Disabled) {
			return fail("unsupported source acquired enabled Customize")
		}
		if control.Name == "offset" && control.Source.Request.Offset != 1 || control.Name == "empty" && control.Source.Response.Count != "0" {
			return fail("unsupported control did not reach intended source condition")
		}
	}
	if len(r.Facts) != 35 {
		return fail("unfinished measured workflow")
	}
	for index, f := range r.Facts {
		name := customizeNames394()[index]
		if f.Name != name || index > 0 && f.Before != r.Facts[index-1].After {
			return fail("action/counter continuity changed")
		}
		source, store, config := customizeCost394(name)
		if f.After.Source-f.Before.Source != source || f.After.Store-f.Before.Store != store || f.After.ConfigWrites-f.Before.ConfigWrites != config || source == 0 && f.After.SQL != f.Before.SQL {
			return fail("local/staging/Back/configuration attempted unsupported read")
		}
	}
	return nil
}
func customizeNames394() []string {
	names := []string{}
	for _, signal := range []string{"logs", "spans"} {
		for _, suffix := range []string{"source-run", "parent-find", "hidden-wrapper", "inert-wrapper", "customize-stage", "explicit-child-run", "local-presentation", "save-configuration", "held-back", "replacement-read", "pending-wrapper", "superseded-wrapper", "destination-inspect-apply", "explicit-destination-run", "offset-unavailable", "empty-unavailable", "failed-unavailable"} {
			names = append(names, signal+"-"+suffix)
		}
	}
	return append(names, "complete")
}
func customizeCost394(name string) (source, store, config int) {
	if strings.HasSuffix(name, "-source-run") || strings.HasSuffix(name, "-explicit-child-run") || strings.HasSuffix(name, "-replacement-read") || strings.HasSuffix(name, "-explicit-destination-run") || strings.HasSuffix(name, "-unavailable") {
		source = 1
	}
	if strings.HasSuffix(name, "-destination-inspect-apply") {
		store = 2
	}
	if strings.HasSuffix(name, "-save-configuration") {
		config = 1
	}
	return
}

func TestQuintDistributionCustomizationObserved394(t *testing.T) {
	path := os.Getenv("GOO_DISTRIBUTION_CUSTOMIZE_394_REPORT")
	if path == "" {
		t.Skip("actual direct distribution journey required")
	}
	wire, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var r customizeLedger394
	if e = json.Unmarshal(wire, &r); e != nil {
		t.Fatal(e)
	}
	if e = admitCustomization394(r); e != nil {
		t.Fatal(e)
	}
	actorPath := os.Getenv("GOO_DISTRIBUTION_CUSTOMIZE_394_ACTOR")
	if actorPath == "" {
		actorPath = filepath.Join(filepath.Dir(path), "actor.json")
	}
	actorWire, e := os.ReadFile(actorPath)
	if e != nil {
		t.Fatal(e)
	}
	var actor customizeActor394
	if e = json.Unmarshal(actorWire, &actor); e != nil {
		t.Fatal(e)
	}
	if e = admitCustomizationActor394(r, actor); e != nil {
		t.Fatal(e)
	}

	if !replayCustomization394(t, r, false) {
		t.Fatal("actual measured workflow violates model")
	}
	mutations := []func(*customizeLedger394){
		func(r *customizeLedger394) { r.Cases[0].Staged.Event.Question.Limit++ },
		func(r *customizeLedger394) { r.Cases[0].ParentAfter.Snapshot.CapturedAt = "2026-01-01T00:00:00Z" },
		func(r *customizeLedger394) { r.Cases[0].Hidden.After.Events++ },
		func(r *customizeLedger394) { r.Cases[0].DestinationQuestion.Where = nil },
		func(r *customizeLedger394) { r.Cases[0].Customized.Snapshot.Partition.Total = "1" },
		func(r *customizeLedger394) { r.Facts[4].After.Source++ },
	}
	for index, mutate := range mutations {
		var bad customizeLedger394
		json.Unmarshal(wire, &bad)
		mutate(&bad)
		if admitCustomization394(bad) == nil {
			t.Fatalf("semantic mutant%d escaped", index)
		}
	}
	if replayCustomization394(t, r, true) {
		t.Fatal("Customize source-read model control escaped")
	}
	t.Logf("admitted two ordinary signal journeys,36physical rows,35facts; six semantic mutants and staging-read control rejected")
}
func replayCustomization394(t *testing.T, r customizeLedger394, mutate bool) bool {
	t.Helper()
	names := customizeNames394()
	if len(r.Facts) != len(names) {
		return false
	}
	reads, stores, configs := make([]int, len(names)), make([]int, len(names)), make([]int, len(names))
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for index, f := range r.Facts {
		if f.Name != names[index] {
			return false
		}
		reads[index], stores[index], configs[index] = customizeCost394(f.Name)
		source := f.After.Source
		if mutate && f.Name == "logs-customize-stage" {
			source++
		}
		action := "step"
		if f.Name == "complete" {
			action = "finish"
		}
		rec.Record(context.Background(), action, nil, "obs.phase", index+1, "obs.reads", source, "obs.stores", f.After.Store, "obs.configs", f.After.ConfigWrites, "obs.finished", f.Name == "complete")
	}
	list := func(v []int) string { wire, _ := json.Marshal(v); return string(wire) }
	spec, e := filepath.Abs("model/distribution_customize394.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: distribution_customize394\nconstants: {READS: '%s', STORES: '%s', CONFIGS: '%s'}\ninvariants: [safety]\nactions:\n", spec, list(reads), list(stores), list(configs))
	for _, action := range []string{"step", "finish"} {
		binding += "  " + action + ": {expect: 'phase == {{.obs.phase}} and reads == {{.obs.reads}} and stores == {{.obs.stores}} and configs == {{.obs.configs}} and finished == {{.obs.finished}}'}\n"
	}
	dir := t.TempDir()
	file := filepath.Join(dir, "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
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
	label := "actual"
	if mutate {
		label = "stage-read-mutant"
	}
	steps, _ := json.Marshal(sink.Steps())
	if e = artifact("distribution-customize394-"+label+"-steps.json", steps); e != nil {
		t.Fatal(e)
	}
	if e = artifact("distribution-customize394-"+label+"-model.txt", []byte(result.Summary())); e != nil {
		t.Fatal(e)
	}
	return result.OK()
}

type customizeActor394 struct {
	Version, SourceBank string
	Question            exploration.Query
	SourceFrame         customizeRead394
	SQLBefore, SQLAfter int
	ProxyCalls          []literalWire393
	Actor               struct {
		Snapshots int
		Phases    []string
		Complete  struct {
			Version, Phase string
			Parent         customizeParent394
			Requests       []struct {
				Path     string
				Request  exploration.Query
				Response exploration.Result
				Status   int
			}
		}
	}
}

func admitCustomizationActor394(r customizeLedger394, actor customizeActor394) error {
	fail := func(s string) error { return fmt.Errorf("customize394 actor: %s", s) }
	if actor.Version != "goo.distributionCustomize394ActorReport/1" || actor.SourceBank != r.Fixture.Token || !reflect.DeepEqual(actor.Question, r.Cases[0].SourceQuestion) || actor.Actor.Snapshots <= 0 || actor.Actor.Complete.Phase != "complete" || actor.Actor.Complete.Version != "goo.distributionCustomize394Actor/1" || len(actor.Actor.Complete.Requests) != 2 || !slices.Contains(actor.Actor.Phases, "staged") {
		return fail("actual actor completion missing")
	}
	if e := admitCustomizeRead394(actor.SourceFrame, r.Raw.Logs); e != nil {
		return e
	}
	for _, request := range actor.Actor.Complete.Requests {
		if request.Path != "/api/explore" || !reflect.DeepEqual(request.Request, actor.Question) {
			return fail("actor changed fixed source question")
		}
		if e := admitCustomizeRead394(customizeRead394{request.Request, request.Response, request.Status}, r.Raw.Logs); e != nil {
			return e
		}
	}
	if actor.Actor.Complete.Parent.Find != "read" || !sameTransport384(actor.Actor.Complete.Parent.Snapshot, actor.Actor.Complete.Requests[0].Response.ViewDataset) {
		return fail("held Back replaced actor source identity/data/clock")
	}
	if actor.SQLAfter-actor.SQLBefore != 8 || actor.SQLBefore < 0 || actor.SQLAfter > len(actor.ProxyCalls) {
		return fail("actor source/staging/Back budget changed")
	}
	plan, e := exploration.Compile(actor.Question)
	if e != nil {
		return e
	}
	expected := map[string]int{plan.Count: 2, plan.Rows: 2, plan.Groups: 2, plan.Fields: 2}
	for _, call := range actor.ProxyCalls[actor.SQLBefore:actor.SQLAfter] {
		if call.Status != 200 || call.Parameters["readonly"] != "1" || expected[call.SQL] <= 0 {
			return fail("actual physical actor source departed from plan")
		}
		expected[call.SQL]--
		for key, value := range plan.Parameters {
			if call.Parameters["param_"+key] != value {
				return fail("actor typed source parameter changed")
			}
		}
	}
	for _, left := range expected {
		if left != 0 {
			return fail("actor source reads incomplete")
		}
	}
	return nil
}
