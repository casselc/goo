package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
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

type meanRaw398 struct {
	customizeRaw394
	DurationNS string `json:"duration_ns"`
}
type meanLedger398 struct {
	Version string
	Fixture struct {
		Token   string
		Records []struct {
			Signal, Team, Environment, Namespace, Service, ID, TraceID, DurationNS string
			Group                                                                  *string
			TimeMS                                                                 int64
		}
	}
	Raw        struct{ Spans []meanRaw398 }
	Cases      []groupedCase395
	Facts      []patternFact391
	ProxyCalls []literalWire393
	Errors     []string
}

func meanRawBase398(raw []meanRaw398) []customizeRaw394 {
	out := make([]customizeRaw394, len(raw))
	for i, row := range raw {
		out[i] = row.customizeRaw394
	}
	return out
}

// The oracle reconstructs typed category/half-open interval membership directly
// from physical rows; neither group summaries nor projected means are its input.
func admitMeanDataset398(d *scopedview.Dataset, q exploration.Query, raw []meanRaw398, basis string) error {
	base := meanRawBase398(raw)
	if err := admitGroupedDataset395(d, q, base, basis); err != nil {
		return err
	}
	if len(d.Fields) != 6 {
		return fmt.Errorf("mean profile must retain six typed fields")
	}
	for _, field := range d.Fields {
		if field.Key == "duration.mean" || field.Key == "duration.sum" {
			kind, method := "sum", "exact"
			if field.Key == "duration.mean" {
				kind, method = "mean", "exact-rational"
			}
			if field.Type != "rational" || field.Unit != "ns" || !field.Nullable || field.Aggregation == nil || field.Aggregation.Family != "duration" || field.Aggregation.Kind != kind || field.Aggregation.Method != method {
				return fmt.Errorf("duration field semantics changed")
			}
		}
	}
	members, _, err := customizeMembers394(q, base)
	if err != nil {
		return err
	}
	sums := map[string]*big.Int{}
	counts := map[string]int{}
	for _, row := range raw {
		if _, ok := members[row.SpanID]; !ok {
			continue
		}
		at, err := strconv.ParseInt(row.TimestampNS, 10, 64)
		if err != nil {
			return err
		}
		from := q.FromMS + ((at/1000000-q.FromMS)/q.BucketMS)*q.BucketMS
		var category *string
		if value, ok := row.SpanAttributes[q.GroupBy[0].Key]; ok {
			category = &value
		}
		key := continuationJSON352(category) + "|" + strconv.FormatInt(from, 10) + "|" + strconv.FormatInt(min(from+q.BucketMS, q.ToMS), 10)
		n, ok := new(big.Int).SetString(row.DurationNS, 10)
		if !ok || n.Sign() < 0 {
			return fmt.Errorf("invalid physical duration")
		}
		if sums[key] == nil {
			sums[key] = new(big.Int)
		}
		sums[key].Add(sums[key], n)
		counts[key]++
	}
	for _, row := range d.Rows {
		v := row.Values
		key := continuationJSON352(v["category"]) + "|" + *v["fromMs"] + "|" + *v["toMs"]
		n := counts[key]
		support := &scopedview.Support{Total: strconv.Itoa(n), Valid: strconv.Itoa(n), Missing: "0", Empty: "0", Invalid: "0"}
		for _, field := range []string{"duration.sum", "duration.mean"} {
			cell, ok := row.Cells[field]
			if !ok || !reflect.DeepEqual(cell.Support, support) {
				return fmt.Errorf("mean support differs from physical contributing count")
			}
			if n == 0 {
				if v[field] != nil || cell.State != "no_data" {
					return fmt.Errorf("empty cell invented zero latency")
				}
				continue
			}
			expected := sums[key].String()
			if field == "duration.mean" {
				expected = new(big.Rat).SetFrac(sums[key], big.NewInt(int64(n))).RatString()
			}
			if v[field] == nil || *v[field] != expected || cell.State != "derived" {
				return fmt.Errorf("mean/sum differs from exact physical duration support")
			}
		}
	}
	return nil
}
func admitMeanRead398(read customizeRead394, raw []meanRaw398) error {
	q, r := read.Request, read.Response
	if read.Status != 200 || q.Signal != "spans" || q.Summary != "span-duration" || !q.GroupTimeSeries || !reflect.DeepEqual(q, r.Question) || r.State != "observed" || !peLiveHex(r.Basis) || !peLiveClock(r.CapturedAt) {
		return fmt.Errorf("mean read identity/status changed")
	}
	members, groups, err := customizeMembers394(q, meanRawBase398(raw))
	if err != nil {
		return err
	}
	if r.Count != strconv.Itoa(len(members)) || len(r.Groups) != len(groups) || len(r.Rows) != min(q.Limit, max(0, len(members)-q.Offset)) || r.HasMore != (len(members) > q.Offset+len(r.Rows)) {
		return fmt.Errorf("mean population/page changed")
	}
	if r.GroupedTimeDataset == nil || r.GroupedTimeDataset.CapturedAt != r.CapturedAt {
		return fmt.Errorf("mean dataset custody clock changed")
	}
	return admitMeanDataset398(r.GroupedTimeDataset, q, raw, r.Basis)
}
func meanNames398() []string {
	names := []string{}
	for _, suffix := range []string{"source-run", "source-local", "customize-stage", "explicit-child-run", "local-presentation", "save-configuration", "known-zero-preview", "held-back", "retired-owner", "reopen-staged", "inspect-apply-destination", "explicit-destination-run", "destination-local"} {
		names = append(names, "spans-"+suffix)
		if suffix == "known-zero-preview" {
			names = append(names, "spans-zero-stage", "spans-zero-return")
		}
	}
	return append(names, "complete")
}
func meanCost398(name string) (read, store, config int) {
	if strings.HasSuffix(name, "-reopen-staged") {
		return 0, 1, 0
	}
	return groupedCost395(name)
}
func admitMeanLedger398(r meanLedger398) error {
	if r.Version != "goo.groupedMean398/1" || len(r.Errors) != 0 || len(r.Cases) != 1 || len(r.Raw.Spans) != 19 || len(r.Fixture.Records) != 19 {
		return fmt.Errorf("incomplete mean journey")
	}
	for _, seed := range r.Fixture.Records {
		found := false
		for _, row := range r.Raw.Spans {
			if row.SpanID != seed.ID {
				continue
			}
			if found || row.TraceID != seed.TraceID || row.ServiceName != seed.Service || row.TimestampNS != strconv.FormatInt(seed.TimeMS*1000000, 10) || row.DurationNS != seed.DurationNS || row.ResourceAttributes["service.namespace"] != seed.Namespace || row.ResourceAttributes["deployment.environment.name"] != seed.Environment {
				return fmt.Errorf("physical mean fixture differs from original OTLP")
			}
			value, present := row.SpanAttributes["work.group"]
			if row.SpanAttributes["team"] != seed.Team || seed.Group == nil && present || seed.Group != nil && (!present || value != *seed.Group) {
				return fmt.Errorf("mean category presence changed")
			}
			found = true
		}
		if !found {
			return fmt.Errorf("original mean span absent")
		}
	}
	c := r.Cases[0]
	if c.Signal != "spans" || !reflect.DeepEqual(c.Source.Request, c.SourceQuestion) || !reflect.DeepEqual(c.ChildRun.Request, c.SourceQuestion) || !reflect.DeepEqual(c.DestinationRun.Request, c.DestinationQuestion) {
		return fmt.Errorf("whole mean question changed")
	}
	for _, read := range []customizeRead394{c.Source, c.ChildRun, c.DestinationRun} {
		if err := admitMeanRead398(read, r.Raw.Spans); err != nil {
			return err
		}
	}
	if !c.Staged.Event.Stage || c.Staged.ChartCount != 0 || !reflect.DeepEqual(c.Staged.Event.Question, c.SourceQuestion) || !reflect.DeepEqual(c.Staged.Event.View.Queries, []exploration.Query{c.SourceQuestion}) || c.Staged.Event.View.Validate() != nil {
		return fmt.Errorf("mean staging copied evidence or changed question")
	}
	if !sameTransport384(c.ParentBefore, c.ParentAfter) || !sameTransport384(c.ParentBefore.Snapshot, c.Source.Response.GroupedTimeDataset) {
		return fmt.Errorf("held mean data/display changed")
	}
	for _, pub := range []customizeSaved394{c.Saved, c.DestinationSetup} {
		if pub.Status != 201 || pub.Request.Validate() != nil || pub.Response.Validate() != nil {
			return fmt.Errorf("mean configuration publication failed")
		}
		copy := pub.Response
		id := copy.ID
		copy.ID = ""
		digest, err := workbench.ViewContentID(copy)
		if err != nil || digest != id || !sameTransport384(copy, pub.Request) {
			return fmt.Errorf("immutable mean definition differs")
		}
	}
	if !reflect.DeepEqual(c.Saved.Response.Queries, []exploration.Query{c.SourceQuestion}) || !reflect.DeepEqual(c.DestinationSetup.Response.Queries, []exploration.Query{c.DestinationQuestion}) || c.StoredBytesBefore == "" || c.StoredBytesBefore != c.StoredBytesAfter {
		return fmt.Errorf("saved mean question/bytes changed")
	}
	var stored workbench.ExplorationView
	if json.Unmarshal([]byte(c.StoredBytesBefore), &stored) != nil || !sameTransport384(stored, c.Saved.Response) {
		return fmt.Errorf("stored mean bytes differ from publication")
	}
	reused, err := workbench.ReuseViewComposition(*c.Saved.Response.Composition, c.SourceQuestion, c.DestinationQuestion)
	if err != nil || c.Reuse.ChartCount != 0 || !reflect.DeepEqual(c.Reuse.Question, c.DestinationQuestion) || !sameTransport384(c.Reuse.Composition, reused) || !sameTransport384(c.Reuse.Inspection, c.Saved.Response) {
		return fmt.Errorf("mean Inspect/Apply changed destination or imported data")
	}
	for _, snap := range c.LocalSnapshots {
		want := c.ChildRun.Response.GroupedTimeDataset
		if strings.HasPrefix(snap.Name, "source-") {
			want = c.Source.Response.GroupedTimeDataset
		}
		if strings.HasPrefix(snap.Name, "destination-") {
			want = c.DestinationRun.Response.GroupedTimeDataset
		}
		if !sameTransport384(snap.Snapshot, want) {
			return fmt.Errorf("local mean presentation changed supplied data")
		}
	}
	zero := c.KnownZero.Row
	if !c.KnownZero.StageEnabled || zero.Values["category"] != nil || zero.Values["toMs"] == nil || *zero.Values["toMs"] != strconv.FormatInt(c.SourceQuestion.ToMS, 10) || zero.Values["count"] == nil || *zero.Values["count"] != "0" || zero.Values["duration.mean"] != nil || zero.State != "no_data" {
		return fmt.Errorf("known-zero count invented duration")
	}
	found := false
	for _, row := range c.ChildRun.Response.GroupedTimeDataset.Rows {
		if row.Key == zero.Key && sameTransport384(row, zero) {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("zero action lost original supplied cell")
	}
	continuation := c.ZeroContinuation
	expected := c.SourceQuestion
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
	if !reflect.DeepEqual(continuation.Question, expected) || continuation.Phase != "staged" || continuation.StagedChartCount != 0 || !sameTransport384(continuation.ParentBefore, continuation.ParentAfter) || !sameTransport384(continuation.ParentBefore.Snapshot, c.ChildRun.Response.GroupedTimeDataset) {
		return fmt.Errorf("zero mean context changed exact child or held parent")
	}
	if !c.Retired.Invoked || c.Retired.After.Events != c.Retired.Before.Events {
		return fmt.Errorf("retired mean callback staged")
	}
	// Compile the actual fixed questions and compare the full physical SQL multiset.
	type planned struct {
		sql    string
		params map[string]string
		used   bool
	}
	plans := []planned{}
	for _, read := range []customizeRead394{c.Source, c.ChildRun, c.DestinationRun} {
		p, err := exploration.Compile(read.Request)
		if err != nil {
			return err
		}
		for _, sql := range []string{p.Count, p.Rows, p.Groups, p.Fields, p.TimeSeries, p.GroupedTimeSeries} {
			if sql != "" {
				plans = append(plans, planned{sql: sql, params: p.Parameters})
			}
		}
	}
	if len(plans) != len(r.ProxyCalls) {
		return fmt.Errorf("mean source SQL count differs: got%d want%d", len(r.ProxyCalls), len(plans))
	}
	for _, call := range r.ProxyCalls {
		found := false
		for i := range plans {
			p := &plans[i]
			if p.used || p.sql != call.SQL {
				continue
			}
			same := true
			for key, value := range p.params {
				if call.Parameters["param_"+key] != value {
					same = false
					break
				}
			}
			if !same {
				continue
			}
			if call.Status != 200 || call.Parameters["readonly"] != "1" || call.Parameters["max_result_rows"] != "1002" || call.Parameters["max_execution_time"] != "10" {
				return fmt.Errorf("mean source readonly/settings changed")
			}
			p.used = true
			found = true
			break
		}
		if !found {
			return fmt.Errorf("unexpected mean SQL or parameters")
		}
	}
	names := meanNames398()
	if len(names) != len(r.Facts) {
		return fmt.Errorf("mean facts got%d want%d", len(r.Facts), len(names))
	}
	before := patternBudget391{}
	for i, f := range r.Facts {
		read, store, config := meanCost398(f.Name)
		if f.Name != names[i] || f.Before.Source != before.Source || f.Before.Store != before.Store || f.Before.ConfigWrites != before.ConfigWrites || f.After.Source-f.Before.Source != read || f.After.Store-f.Before.Store != store || f.After.ConfigWrites-f.Before.ConfigWrites != config || read == 0 && f.After.SQL != f.Before.SQL {
			return fmt.Errorf("mean read/config budget changed at%s", f.Name)
		}
		before = f.After
	}
	if before.Source != 3 || before.Store != 3 || before.ConfigWrites != 1 {
		return fmt.Errorf("mean final budget changed")
	}
	return nil
}

func replayMean398(t *testing.T, r meanLedger398, mutate bool) bool {
	t.Helper()
	reads, stores, configs := []int{}, []int{}, []int{}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, f := range r.Facts {
		action := "step"
		if f.Name == "complete" {
			action = "finish"
		} else {
			a, b, c := meanCost398(f.Name)
			reads = append(reads, a)
			stores = append(stores, b)
			configs = append(configs, c)
		}
		n := f.After.Source
		if mutate && f.Name == "spans-customize-stage" {
			n++
		}
		rec.Record(context.Background(), action, nil, "obs.phase", i+1, "obs.reads", n, "obs.stores", f.After.Store, "obs.configs", f.After.ConfigWrites, "obs.finished", f.Name == "complete")
	}
	list := func(v []int) string { b, _ := json.Marshal(v); return string(b) }
	spec, err := filepath.Abs("model/grouped_mean398.qnt")
	if err != nil {
		t.Fatal(err)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: grouped_mean398\nconstants: {READS: '%s', STORES: '%s', CONFIGS: '%s', STEPS: '15', READ_LIMIT: '3', STORE_LIMIT: '3', CONFIG_LIMIT: '1'}\ninvariants: [safety]\nactions:\n", spec, list(reads), list(stores), list(configs))
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
	return result.OK()
}
func TestQuintGroupedMeanObserved398(t *testing.T) {
	path := os.Getenv("GOO_GROUPED_MEAN_398_REPORT")
	if path == "" {
		t.Skip("actual grouped mean journey required")
	}
	wire, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var r meanLedger398
	if err = json.Unmarshal(wire, &r); err != nil {
		t.Fatal(err)
	}
	if err = admitMeanLedger398(r); err != nil {
		t.Fatal(err)
	}
	actorPath := os.Getenv("GOO_GROUPED_MEAN_398_ACTOR")
	if actorPath == "" {
		actorPath = filepath.Join(filepath.Dir(path), "actor.json")
	}
	actorWire, e := os.ReadFile(actorPath)
	if e != nil {
		t.Fatal(e)
	}
	var actor groupedActor395
	if e = json.Unmarshal(actorWire, &actor); e != nil {
		t.Fatal(e)
	}
	if e = admitMeanActor398(r, actor); e != nil {
		t.Fatal(e)
	}
	mutants := []func(*meanLedger398){func(r *meanLedger398) { r.Raw.Spans[0].DurationNS = "999" }, func(r *meanLedger398) {
		r.Cases[0].Source.Response.GroupedTimeDataset.Rows[0].Values["duration.mean"] = ptr395("999")
	}, func(r *meanLedger398) { r.Cases[0].ParentAfter.Find = "changed" }, func(r *meanLedger398) { r.Cases[0].Staged.Event.Question.Limit++ }, func(r *meanLedger398) { r.Cases[0].KnownZero.Row.Values["duration.mean"] = ptr395("0") }, func(r *meanLedger398) { r.Cases[0].Retired.After.Events++ }, func(r *meanLedger398) { r.Facts[2].After.Source++ }}
	for i, mutate := range mutants {
		var bad meanLedger398
		json.Unmarshal(wire, &bad)
		mutate(&bad)
		if admitMeanLedger398(bad) == nil {
			t.Fatalf("mean semantic mutant%d escaped", i)
		}
	}
	if !replayMean398(t, r, false) || replayMean398(t, r, true) {
		t.Fatal("mean observed model/staging read negative control incorrect")
	}
	t.Logf("19 physical spans, exact sum/count/rational means, %d facts; seven semantic mutants and staging-read control rejected", len(r.Facts))
}

func admitMeanActor398(ledger meanLedger398, actor groupedActor395) error {
	if actor.Version != "goo.groupedMean398ActorReport/1" || actor.SourceBank != ledger.Fixture.Token || !reflect.DeepEqual(actor.Question, ledger.Cases[0].SourceQuestion) || actor.Actor.Snapshots <= 0 || actor.Actor.Complete.Version != "goo.groupedMean398Actor/1" || actor.Actor.Complete.Phase != "complete" || len(actor.Actor.Complete.Requests) != 2 {
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
		if err := admitMeanRead398(customizeRead394{Request: request.Request, Response: request.Response, Status: request.Status}, ledger.Raw.Spans); err != nil {
			return err
		}
	}
	parent := actor.Actor.Complete.Parent
	if parent.Find != "雪" || quantValue397(parent.Display) != "duration.mean" || !sameTransport384(parent.Snapshot, actor.Actor.Complete.Requests[0].Response.GroupedTimeDataset) {
		return fmt.Errorf("actor Back lost held data/Find/clock")
	}
	zero := actor.Actor.Complete.Zero
	found := false
	for _, row := range actor.Actor.Complete.Requests[1].Response.GroupedTimeDataset.Rows {
		if row.Key == zero.Key && sameTransport384(row, zero) && row.State == "no_data" && row.Values["duration.mean"] == nil && row.Values["count"] != nil && *row.Values["count"] == "0" {
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
