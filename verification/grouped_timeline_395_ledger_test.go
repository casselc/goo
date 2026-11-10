package verification

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/scopedview"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type groupedSnapshot395 struct {
	Name     string
	Snapshot scopedview.Dataset
	Display  json.RawMessage
	Find     string
}
type groupedCase395 struct {
	Signal                              string
	SourceQuestion, DestinationQuestion exploration.Query
	Source, ChildRun, DestinationRun    customizeRead394
	ParentBefore, ParentAfter           groupedSnapshot395
	Staged                              struct {
		Event struct {
			Question exploration.Query
			View     workbench.ExplorationView
			Stage    bool
		}
		ChartCount int
	}
	LocalSnapshots          []groupedSnapshot395
	Saved, DestinationSetup customizeSaved394
	Reuse                   struct {
		Question    exploration.Query
		Inspection  workbench.ExplorationView
		Composition workbench.ViewComposition
		ChartCount  int
	}
	DestinationSnapshot                 groupedSnapshot395
	StoredBytesBefore, StoredBytesAfter string
	KnownZero                           struct {
		Row          scopedview.Row
		StageEnabled bool
	}
	ZeroContinuation struct {
		Question                  exploration.Query
		ParentBefore, ParentAfter groupedSnapshot395
		Phase                     string
		StagedChartCount          int
	}
	Retired customizeWrapper394
}
type groupedLedger395 struct {
	Version string
	Fixture struct {
		Token, Environment, Namespace string
		Base                          int64
		Records                       []struct {
			Signal, Team, Environment, Namespace, Service, ID, TraceID string
			Group                                                      *string
			TimeMS                                                     int64
		}
	}
	Raw        struct{ Logs, Spans []customizeRaw394 }
	Cases      []groupedCase395
	Facts      []patternFact391
	ProxyCalls []literalWire393
	Errors     []string
}

// Counts come from physical records, not the report's convenient sourceOracle,
// the projection's cells, a rendered label or an independently read grand total.
func groupedExpectedCells395(q exploration.Query, raw []customizeRaw394) (map[string]string, error) {
	members, groups, err := customizeMembers394(q, raw)
	if err != nil {
		return nil, err
	}
	cells := map[string]int{}
	for category := range groups {
		for from := q.FromMS; from < q.ToMS; from += q.BucketMS {
			key := category + "|" + strconv.FormatInt(from, 10) + "|" + strconv.FormatInt(min(from+q.BucketMS, q.ToMS), 10)
			cells[key] = 0
		}
	}
	for _, row := range members {
		attributes := row.LogAttributes
		if q.Signal == "spans" {
			attributes = row.SpanAttributes
		}
		var category *string
		if value, ok := attributes[q.GroupBy[0].Key]; ok {
			category = &value
		}
		at, err := strconv.ParseInt(row.TimestampNS, 10, 64)
		if err != nil {
			return nil, err
		}
		from := q.FromMS + ((at/1000000-q.FromMS)/q.BucketMS)*q.BucketMS
		key := continuationJSON352(category) + "|" + strconv.FormatInt(from, 10) + "|" + strconv.FormatInt(min(from+q.BucketMS, q.ToMS), 10)
		cells[key]++
	}
	expected := map[string]string{}
	for key, count := range cells {
		expected[key] = strconv.Itoa(count)
	}
	return expected, nil
}
func admitGroupedDataset395(d *scopedview.Dataset, q exploration.Query, raw []customizeRaw394, basis string) error {
	if d == nil || d.Shape != "grouped-time-series" || d.Revision != basis || d.Scope.Generation != basis || d.Quality.Extent != "complete" || d.Completeness == nil || d.Completeness.Revision != basis || d.Completeness.ScopeID != d.Scope.ID || !reflect.DeepEqual(d.Completeness.Aspects, []string{"grouped-intervals"}) || d.Partition != nil || !reflect.DeepEqual(d.Dimensions, []string{"category"}) {
		return fmt.Errorf("grouped dataset identity/coverage changed")
	}
	var selector exploration.Query
	if json.Unmarshal(d.Scope.Selector, &selector) != nil || !reflect.DeepEqual(selector, q) || d.Scope.Signal != q.Signal || d.Interval == nil || d.Interval.FromMS != strconv.FormatInt(q.FromMS, 10) || d.Interval.ToMS != strconv.FormatInt(q.ToMS, 10) || d.Interval.BucketMS != strconv.FormatInt(q.BucketMS, 10) {
		return fmt.Errorf("grouped dataset changed exact selector/interval")
	}
	expected, err := groupedExpectedCells395(q, raw)
	if err != nil {
		return err
	}
	if len(d.Rows) != len(expected) {
		return fmt.Errorf("grouped grid membership differs from physical records")
	}
	keys := map[string]bool{}
	for _, row := range d.Rows {
		v := row.Values
		if v["fromMs"] == nil || v["toMs"] == nil || v["count"] == nil {
			return fmt.Errorf("coherent cell missing value")
		}
		key := continuationJSON352(v["category"]) + "|" + *v["fromMs"] + "|" + *v["toMs"]
		count, ok := expected[key]
		if !ok || keys[key] || count != *v["count"] || row.State != map[bool]string{true: "no_data", false: "observed"}[count == "0"] {
			return fmt.Errorf("group interval differs from exact raw count")
		}
		keys[key] = true
		tuple, _ := json.Marshal([]any{v["category"], *v["fromMs"], *v["toMs"]})
		digest := sha256.Sum256(tuple)
		if row.Key != "group-interval:"+hex.EncodeToString(digest[:]) {
			return fmt.Errorf("group row key changed typed tuple")
		}
	}
	return nil
}
func admitGroupedRead395(read customizeRead394, raw []customizeRaw394) error {
	q, r := read.Request, read.Response
	if read.Status != 200 || !exploration.GroupedCountTimelineQuery(q) || !reflect.DeepEqual(q, r.Question) || r.State != "observed" || !peLiveHex(r.Basis) || !peLiveClock(r.CapturedAt) {
		return fmt.Errorf("actual grouped read scope/state changed")
	}
	members, groups, err := customizeMembers394(q, raw)
	if err != nil {
		return err
	}
	if r.Count != strconv.Itoa(len(members)) || len(r.Groups) != len(groups) || len(r.Rows) != min(q.Limit, max(0, len(members)-q.Offset)) || r.HasMore != (len(members) > q.Offset+len(r.Rows)) {
		return fmt.Errorf("actual group total/page membership changed")
	}
	seen := map[string]bool{}
	for _, wire := range r.Rows {
		var row struct {
			SpanID     string
			Attributes map[string]string
		}
		if json.Unmarshal(wire, &row) != nil || seen[row.SpanID] {
			return fmt.Errorf("invalid/duplicate observed page row")
		}
		original, ok := members[row.SpanID]
		attributes := original.LogAttributes
		if q.Signal == "spans" {
			attributes = original.SpanAttributes
		}
		if !ok || !reflect.DeepEqual(attributes, row.Attributes) {
			return fmt.Errorf("page literal fields disagree with physical member")
		}
		seen[row.SpanID] = true
	}
	if r.GroupedTimeDataset == nil || r.GroupedTimeDataset.CapturedAt != r.CapturedAt {
		return fmt.Errorf("grouped capture clock changed")
	}
	return admitGroupedDataset395(r.GroupedTimeDataset, q, raw, r.Basis)
}
func groupedNames395() []string {
	names := []string{}
	for _, signal := range []string{"logs", "spans"} {
		for _, suffix := range []string{"source-run", "source-local", "customize-stage", "explicit-child-run", "local-presentation", "save-configuration", "known-zero-preview", "held-back", "retired-owner", "inspect-apply-destination", "explicit-destination-run", "destination-local"} {
			names = append(names, signal+"-"+suffix)
			if signal == "logs" && suffix == "known-zero-preview" {
				names = append(names, "logs-zero-stage", "logs-zero-return")
			}
		}
	}
	return append(names, "complete")
}
func groupedCost395(name string) (read, store, config int) {
	switch {
	case strings.HasSuffix(name, "-source-run"), strings.HasSuffix(name, "-explicit-child-run"), strings.HasSuffix(name, "-explicit-destination-run"):
		read = 1
	case strings.HasSuffix(name, "-inspect-apply-destination"):
		store = 2
	case strings.HasSuffix(name, "-save-configuration"):
		config = 1
	}
	return
}
func admitGroupedLedger395(r groupedLedger395) error {
	if r.Version != "goo.groupedTimeline395/1" || len(r.Errors) != 0 || len(r.Cases) != 2 || len(r.Raw.Logs) != 19 || len(r.Raw.Spans) != 19 || len(r.Fixture.Records) != 38 || len(r.Facts) != 27 {
		return fmt.Errorf("incomplete actual grouped timeline ledger")
	}
	// Verify raw physical evidence corresponds to every original ordinary OTLP
	// record. Null is a missing key, not a literal null/empty value.
	for _, seed := range r.Fixture.Records {
		rows := r.Raw.Logs
		if seed.Signal == "spans" {
			rows = r.Raw.Spans
		}
		found := false
		for _, row := range rows {
			if row.SpanID != seed.ID {
				continue
			}
			if found || row.TraceID != seed.TraceID || row.ServiceName != seed.Service || row.TimestampNS != strconv.FormatInt(seed.TimeMS*1000000, 10) || row.ResourceAttributes["service.namespace"] != seed.Namespace || row.ResourceAttributes["deployment.environment.name"] != seed.Environment {
				return fmt.Errorf("physical record differs from original OTLP fixture")
			}
			attrs := row.LogAttributes
			if seed.Signal == "spans" {
				attrs = row.SpanAttributes
			}
			value, present := attrs["work.group"]
			if attrs["team"] != seed.Team || (seed.Group == nil && present) || (seed.Group != nil && (!present || value != *seed.Group)) {
				return fmt.Errorf("physical missing/empty/group identity changed")
			}
			found = true
		}
		if !found {
			return fmt.Errorf("original OTLP member absent from physical bank")
		}
	}
	for index, c := range r.Cases {
		signal := []string{"logs", "spans"}[index]
		raw := r.Raw.Logs
		if signal == "spans" {
			raw = r.Raw.Spans
		}
		if c.Signal != signal || !reflect.DeepEqual(c.Source.Request, c.SourceQuestion) || !reflect.DeepEqual(c.ChildRun.Request, c.SourceQuestion) || !reflect.DeepEqual(c.DestinationRun.Request, c.DestinationQuestion) {
			return fmt.Errorf("whole source/destination question changed")
		}
		for _, read := range []customizeRead394{c.Source, c.ChildRun, c.DestinationRun} {
			if err := admitGroupedRead395(read, raw); err != nil {
				return err
			}
		}
		if !c.Staged.Event.Stage || c.Staged.ChartCount != 0 || !reflect.DeepEqual(c.Staged.Event.Question, c.SourceQuestion) || !reflect.DeepEqual(c.Staged.Event.View.Queries, []exploration.Query{c.SourceQuestion}) || c.Staged.Event.View.Validate() != nil {
			return fmt.Errorf("customize imported evidence or changed query")
		}
		if !sameTransport384(c.ParentBefore, c.ParentAfter) || !sameTransport384(c.ParentBefore.Snapshot, c.Source.Response.GroupedTimeDataset) {
			return fmt.Errorf("Back replaced held grouped data/display/clock")
		}
		if c.Saved.Status != 201 || !reflect.DeepEqual(c.Saved.Request.Queries, []exploration.Query{c.SourceQuestion}) || c.Saved.Request.Validate() != nil || c.Saved.Response.Validate() != nil || !sameTransport384(c.Saved.Request.Composition, c.Saved.Response.Composition) || c.StoredBytesBefore == "" || c.StoredBytesBefore != c.StoredBytesAfter {
			return fmt.Errorf("immutable saved grouped configuration changed")
		}
		for _, publication := range []customizeSaved394{c.Saved, c.DestinationSetup} {
			if publication.Status != 201 {
				return fmt.Errorf("configuration publication failed")
			}
			copy := publication.Response
			id := copy.ID
			copy.ID = ""
			digest, err := workbench.ViewContentID(copy)
			if err != nil || digest != id || !sameTransport384(copy, publication.Request) {
				return fmt.Errorf("configuration publication changed content or identity")
			}
		}
		if !reflect.DeepEqual(c.DestinationSetup.Response.Queries, []exploration.Query{c.DestinationQuestion}) {
			return fmt.Errorf("destination setup changed its question")
		}
		var stored workbench.ExplorationView
		if json.Unmarshal([]byte(c.StoredBytesBefore), &stored) != nil || !sameTransport384(stored, c.Saved.Response) {
			return fmt.Errorf("stored bytes differ from actual publication")
		}
		reused, err := workbench.ReuseViewComposition(*c.Saved.Response.Composition, c.SourceQuestion, c.DestinationQuestion)
		if err != nil || c.Reuse.ChartCount != 0 || !reflect.DeepEqual(c.Reuse.Question, c.DestinationQuestion) || !sameTransport384(c.Reuse.Composition, reused) || !sameTransport384(c.Reuse.Inspection, c.Saved.Response) {
			return fmt.Errorf("Inspect/Apply changed destination or copied data")
		}
		for _, snap := range c.LocalSnapshots {
			expected := c.ChildRun.Response.GroupedTimeDataset
			if strings.HasPrefix(snap.Name, "source-") {
				expected = c.Source.Response.GroupedTimeDataset
			}
			if strings.HasPrefix(snap.Name, "destination-") {
				expected = c.DestinationRun.Response.GroupedTimeDataset
			}
			if !sameTransport384(snap.Snapshot, expected) {
				return fmt.Errorf("local Find/kind changed supplied grouped data")
			}
		}
		zero := c.KnownZero.Row
		if !c.KnownZero.StageEnabled || zero.Values["category"] != nil || zero.Values["count"] == nil || *zero.Values["count"] != "0" || zero.State != "no_data" || zero.Values["fromMs"] == nil || zero.Values["toMs"] == nil || *zero.Values["toMs"] != strconv.FormatInt(c.SourceQuestion.ToMS, 10) {
			return fmt.Errorf("known-zero final interval preview lost exact context")
		}
		zeroFound := false
		for _, row := range c.ChildRun.Response.GroupedTimeDataset.Rows {
			if row.Key == zero.Key && sameTransport384(row, zero) {
				zeroFound = true
				break
			}
		}
		if !zeroFound {
			return fmt.Errorf("known-zero action did not name original supplied cell")
		}
		if signal == "logs" {
			rawQuestion, _ := json.Marshal(c.SourceQuestion)
			var expected exploration.Query
			json.Unmarshal(rawQuestion, &expected)
			expected.FromMS, _ = strconv.ParseInt(*zero.Values["fromMs"], 10, 64)
			expected.ToMS, _ = strconv.ParseInt(*zero.Values["toMs"], 10, 64)
			predicates := []exploration.Predicate{}
			if expected.Where != nil {
				if expected.Where.All != nil {
					predicates = append(predicates, expected.Where.All...)
				} else {
					predicates = append(predicates, *expected.Where)
				}
			}
			field := expected.GroupBy[0]
			predicates = append(predicates, exploration.Predicate{Field: &field, Op: "missing"})
			expected.Where = &exploration.Predicate{All: predicates}
			continuation := c.ZeroContinuation
			if !reflect.DeepEqual(continuation.Question, expected) || continuation.Phase != "staged" || continuation.StagedChartCount != 0 || !sameTransport384(continuation.ParentBefore, continuation.ParentAfter) || !sameTransport384(continuation.ParentBefore.Snapshot, c.ChildRun.Response.GroupedTimeDataset) {
				return fmt.Errorf("current known-zero continuation changed typed scope/window or held parent")
			}
		}
		if !c.Retired.Invoked || c.Retired.After.Events != c.Retired.Before.Events {
			return fmt.Errorf("retired owner staged scope")
		}
	}
	// Every public grouped source call owns exactly the six planned physical
	// reads. Configuration, local displays and Back add no alternate SQL path.
	type plannedRead struct {
		sql        string
		parameters map[string]string
		used       bool
	}
	planned := []plannedRead{}
	for _, c := range r.Cases {
		for _, read := range []customizeRead394{c.Source, c.ChildRun, c.DestinationRun} {
			plan, err := exploration.Compile(read.Request)
			if err != nil {
				return err
			}
			for _, sql := range []string{plan.Count, plan.Rows, plan.Groups, plan.Fields, plan.TimeSeries, plan.GroupedTimeSeries} {
				if sql == "" {
					return fmt.Errorf("grouped physical plan incomplete")
				}
				planned = append(planned, plannedRead{sql: sql, parameters: plan.Parameters})
			}
		}
	}
	if len(r.ProxyCalls) != len(planned) {
		return fmt.Errorf("unexpected physical SQL reads: got %d want %d", len(r.ProxyCalls), len(planned))
	}
	for _, call := range r.ProxyCalls {
		matched := false
		for index := range planned {
			expected := &planned[index]
			if expected.used || call.SQL != expected.sql {
				continue
			}
			same := true
			for key, value := range expected.parameters {
				if call.Parameters["param_"+key] != value {
					same = false
					break
				}
			}
			if same {
				if call.Status != 200 || call.Parameters["readonly"] != "1" || call.Parameters["max_result_rows"] != "1002" || call.Parameters["max_execution_time"] != "10" {
					return fmt.Errorf("grouped physical source settings changed")
				}
				expected.used = true
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("unexpected physical grouped query or typed parameters")
		}
	}
	names := groupedNames395()
	before := patternBudget391{}
	for index, fact := range r.Facts {
		read, store, config := groupedCost395(fact.Name)
		if fact.Name != names[index] || fact.Before.Source != before.Source || fact.Before.Store != before.Store || fact.Before.ConfigWrites != before.ConfigWrites || fact.After.Source-fact.Before.Source != read || fact.After.Store-fact.Before.Store != store || fact.After.ConfigWrites-fact.Before.ConfigWrites != config || read == 0 && fact.After.SQL != fact.Before.SQL {
			return fmt.Errorf("measured no-read/configuration budget changed at %s", fact.Name)
		}
		before = fact.After
	}
	if before.Source != 6 || before.Store != 4 || before.ConfigWrites != 2 {
		return fmt.Errorf("workflow final budget changed")
	}
	return nil
}
func replayGrouped395(t *testing.T, r groupedLedger395, mutate bool) bool {
	t.Helper()
	reads, stores, configs := []int{}, []int{}, []int{}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for index, fact := range r.Facts {
		action := "step"
		if fact.Name == "complete" {
			action = "finish"
		} else {
			a, b, c := groupedCost395(fact.Name)
			reads = append(reads, a)
			stores = append(stores, b)
			configs = append(configs, c)
		}
		read := fact.After.Source
		if mutate && fact.Name == "logs-customize-stage" {
			read++
		}
		rec.Record(context.Background(), action, nil, "obs.phase", index+1, "obs.reads", read, "obs.stores", fact.After.Store, "obs.configs", fact.After.ConfigWrites, "obs.finished", fact.Name == "complete")
	}
	list := func(v []int) string { wire, _ := json.Marshal(v); return string(wire) }
	spec, err := filepath.Abs("model/grouped_timeline395.qnt")
	if err != nil {
		t.Fatal(err)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: grouped_timeline395\nconstants: {READS: '%s', STORES: '%s', CONFIGS: '%s', STEPS: '26', READ_LIMIT: '6', STORE_LIMIT: '4', CONFIG_LIMIT: '2'}\ninvariants: [safety]\nactions:\n", spec, list(reads), list(stores), list(configs))
	for _, action := range []string{"step", "finish"} {
		binding += "  " + action + ": {expect: 'phase == {{.obs.phase}} and reads == {{.obs.reads}} and stores == {{.obs.stores}} and configs == {{.obs.configs}} and finished == {{.obs.finished}}'}\n"
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "binding.yaml")
	if err = os.WriteFile(path, []byte(binding), 0600); err != nil {
		t.Fatal(err)
	}
	checker, err := validate.NewChecker(path, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	result, err := checker.Check(t.Context(), sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	label := "actual"
	if mutate {
		label = "stage-read-mutant"
	}
	steps, _ := json.Marshal(sink.Steps())
	if err = artifact("grouped-timeline395-"+label+"-steps.json", steps); err != nil {
		t.Fatal(err)
	}
	if err = artifact("grouped-timeline395-"+label+"-model.txt", []byte(result.Summary())); err != nil {
		t.Fatal(err)
	}
	return result.OK()
}
func TestQuintGroupedTimelineObserved395(t *testing.T) {
	path := os.Getenv("GOO_GROUPED_TIMELINE_395_REPORT")
	if path == "" {
		t.Skip("actual ordinary OTel grouped timeline journey required")
	}
	wire, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var ledger groupedLedger395
	if err = json.Unmarshal(wire, &ledger); err != nil {
		t.Fatal(err)
	}
	if err = admitGroupedLedger395(ledger); err != nil {
		t.Fatal(err)
	}
	actorPath := os.Getenv("GOO_GROUPED_TIMELINE_395_ACTOR")
	if actorPath == "" {
		actorPath = filepath.Join(filepath.Dir(path), "actor.json")
	}
	actorWire, err := os.ReadFile(actorPath)
	if err != nil {
		t.Fatal(err)
	}
	var actor groupedActor395
	if err = json.Unmarshal(actorWire, &actor); err != nil {
		t.Fatal(err)
	}
	if err = admitGroupedActor395(ledger, actor); err != nil {
		t.Fatal(err)
	}
	for index, mutate := range []func(*groupedActor395){func(a *groupedActor395) { a.Actor.Complete.Requests[0].Request.Limit++ }, func(a *groupedActor395) { a.Actor.Complete.Parent.Snapshot.CapturedAt = "changed" }, func(a *groupedActor395) { a.Actor.Complete.Zero.Values["count"] = ptr395("1") }, func(a *groupedActor395) { a.SQLAfter++ }} {
		var bad groupedActor395
		json.Unmarshal(actorWire, &bad)
		mutate(&bad)
		if admitGroupedActor395(ledger, bad) == nil {
			t.Fatalf("actor semantic mutant%d admitted", index)
		}
	}
	if !replayGrouped395(t, ledger, false) {
		t.Fatal("actual workflow violates model")
	}
	mutations := []func(*groupedLedger395){func(r *groupedLedger395) {
		r.Cases[0].Source.Response.GroupedTimeDataset.Rows[0].Values["count"] = ptr395("999")
	}, func(r *groupedLedger395) {
		for index := range r.Cases[0].Source.Response.GroupedTimeDataset.Rows {
			row := &r.Cases[0].Source.Response.GroupedTimeDataset.Rows[index]
			if row.Values["category"] == nil {
				row.Values["category"] = ptr395("")
				break
			}
		}
	}, func(r *groupedLedger395) { r.Cases[0].Staged.Event.Question.Limit++ }, func(r *groupedLedger395) { r.Cases[0].ParentAfter.Snapshot.CapturedAt = "changed" }, func(r *groupedLedger395) { r.Cases[0].Reuse.Question.Where = nil }, func(r *groupedLedger395) { r.Cases[0].KnownZero.Row.State = "unknown" }, func(r *groupedLedger395) { r.Cases[0].Retired.After.Events++ }, func(r *groupedLedger395) { r.Facts[2].After.Source++ }, func(r *groupedLedger395) { r.Cases[0].StoredBytesAfter += " " }, func(r *groupedLedger395) { r.Cases[0].ZeroContinuation.Question.ToMS++ }, func(r *groupedLedger395) { r.Cases[0].ZeroContinuation.ParentAfter.Find = "changed" }}
	for index, mutate := range mutations {
		var bad groupedLedger395
		json.Unmarshal(wire, &bad)
		mutate(&bad)
		if admitGroupedLedger395(bad) == nil {
			t.Fatalf("grouped semantic mutant%d admitted", index)
		}
	}
	if replayGrouped395(t, ledger, true) {
		t.Fatal("Quint accepted hidden stage source read")
	}
	t.Logf("38 physical records, %d source/destination matrices, %d measured facts, %d rejecting semantic controls", len(ledger.Cases)*3, len(ledger.Facts), len(mutations)+1)
}
func ptr395(s string) *string { return &s }

type groupedActor395 struct {
	Version, SourceBank string
	Question            exploration.Query
	SQLBefore, SQLAfter int
	ProxyCalls          []literalWire393
	Actor               struct {
		Snapshots int
		Phases    []string
		Complete  struct {
			Version, Phase string
			Parent         groupedSnapshot395
			Zero           scopedview.Row
			Requests       []struct {
				Path     string
				Request  exploration.Query
				Response exploration.Result
				Status   int
			}
		}
	}
}

func admitGroupedActor395(ledger groupedLedger395, actor groupedActor395) error {
	if actor.Version != "goo.groupedTimeline395ActorReport/1" || actor.SourceBank != ledger.Fixture.Token || !reflect.DeepEqual(actor.Question, ledger.Cases[0].SourceQuestion) || actor.Actor.Snapshots <= 0 || actor.Actor.Complete.Version != "goo.groupedTimeline395Actor/1" || actor.Actor.Complete.Phase != "complete" || len(actor.Actor.Complete.Requests) != 2 {
		return fmt.Errorf("actual grouped actor completion missing")
	}
	phases := map[string]bool{}
	for _, phase := range actor.Actor.Phases {
		phases[phase] = true
	}
	for _, phase := range []string{"source", "customize", "staged", "read", "heatmap", "zero", "back", "complete"} {
		if !phases[phase] {
			return fmt.Errorf("actual grouped actor phase missing: %s", phase)
		}
	}
	for _, request := range actor.Actor.Complete.Requests {
		if request.Path != "/api/explore" || !reflect.DeepEqual(request.Request, actor.Question) {
			return fmt.Errorf("actor read a different scope")
		}
		if err := admitGroupedRead395(customizeRead394{Request: request.Request, Response: request.Response, Status: request.Status}, ledger.Raw.Logs); err != nil {
			return err
		}
	}
	parent := actor.Actor.Complete.Parent
	if parent.Find != "雪" || !sameTransport384(parent.Snapshot, actor.Actor.Complete.Requests[0].Response.GroupedTimeDataset) {
		return fmt.Errorf("actor Back lost held data/Find/clock")
	}
	zero := actor.Actor.Complete.Zero
	found := false
	for _, row := range actor.Actor.Complete.Requests[1].Response.GroupedTimeDataset.Rows {
		if row.Key == zero.Key && sameTransport384(row, zero) && row.State == "no_data" && row.Values["count"] != nil && *row.Values["count"] == "0" {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("actor did not select a supplied known-zero cell")
	}
	if actor.SQLBefore < 0 || actor.SQLAfter > len(actor.ProxyCalls) || actor.SQLAfter-actor.SQLBefore != 12 {
		return fmt.Errorf("actor hidden source/SQL read")
	}
	plan, err := exploration.Compile(actor.Question)
	if err != nil {
		return err
	}
	expected := map[string]int{}
	for _, sql := range []string{plan.Count, plan.Rows, plan.Groups, plan.Fields, plan.TimeSeries, plan.GroupedTimeSeries} {
		expected[sql] = 2
	}
	for _, call := range actor.ProxyCalls[actor.SQLBefore:actor.SQLAfter] {
		if call.Status != 200 || call.Parameters["readonly"] != "1" || expected[call.SQL] <= 0 {
			return fmt.Errorf("actor physical SQL departed from source plan")
		}
		expected[call.SQL]--
		for key, value := range plan.Parameters {
			if call.Parameters["param_"+key] != value {
				return fmt.Errorf("actor typed query value changed")
			}
		}
	}
	for _, remaining := range expected {
		if remaining != 0 {
			return fmt.Errorf("actor source plan incomplete")
		}
	}
	return nil
}
