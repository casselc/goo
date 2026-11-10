package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
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

type quantRaw397 struct {
	customizeRaw394
	DurationNS string `json:"duration_ns"`
}
type quantCase397 struct {
	customizeCase394
	Kind string
}
type quantLedger397 struct {
	Version string
	Fixture struct {
		Token   string
		Records []struct {
			Signal, ID, TraceID, Service, Environment, Namespace, Team, Phase, DurationNS string
			TimeMS                                                                        int64
			Numeric                                                                       *string
		}
	}
	Raw      struct{ Logs, Spans []quantRaw397 }
	Cases    []quantCase397
	Facts    []patternFact391
	Controls []struct {
		Signal, Name    string
		Source          customizeRead394
		Request         exploration.Query
		Status, Present int
		Disabled        *bool
	}
	Errors     []string
	ProxyCalls []literalWire393
}

func rawBase397(rows []quantRaw397) []customizeRaw394 {
	out := make([]customizeRaw394, len(rows))
	for i, row := range rows {
		out[i] = row.customizeRaw394
	}
	return out
}
func quantNames397() []string {
	names := []string{}
	for _, kind := range []string{"numeric-logs", "numeric-spans", "duration"} {
		signal := "spans"
		if kind == "numeric-logs" {
			signal = "logs"
		}
		for _, suffix := range []string{"source-run", "parent-find", "hidden-wrapper", "inert-wrapper", "customize-stage", "explicit-child-run", "local-presentation", "save-configuration", "held-back", "replacement-read", "pending-wrapper", "superseded-wrapper", "saved-reopen-staged", "destination-inspect-apply", "explicit-destination-run", "offset-unavailable", "empty-unavailable", "failed-unavailable"} {
			prefix := signal
			if suffix == "saved-reopen-staged" {
				prefix = kind
			}
			names = append(names, prefix+"-"+suffix)
		}
	}
	return append(names, "complete")
}
func quantCost397(name string) (reads, stores, configs int) {
	reads, stores, configs = customizeCost394(name)
	if strings.HasSuffix(name, "-saved-reopen-staged") {
		stores = 1
	}
	return
}
func quantValue397(display json.RawMessage) string {
	var p struct{ Value string }
	json.Unmarshal(display, &p)
	return p.Value
}
func quantRead397(read customizeRead394, rows []quantRaw397) error {
	if e := admitCustomizeRead394(read, rawBase397(rows)); e != nil {
		return e
	}
	members, _, e := customizeMembers394(read.Request, rawBase397(rows))
	if e != nil {
		return e
	}
	chosen := []quantRaw397{}
	for _, row := range rows {
		if _, ok := members[row.SpanID]; ok {
			chosen = append(chosen, row)
		}
	}
	if e = quantSummary397(read.Request, chosen, read.Response.Numeric, read.Response.Duration, nil); e != nil {
		return e
	}
	for _, wire := range read.Response.Groups {
		var g struct {
			G0       *string
			Numeric  *exploration.NumericSummary
			Duration *exploration.DurationSummary
		}
		if e = json.Unmarshal(wire, &g); e != nil {
			return e
		}
		group := []quantRaw397{}
		for _, row := range chosen {
			attrs := row.LogAttributes
			if read.Request.Signal == "spans" {
				attrs = row.SpanAttributes
			}
			value, present := attrs[read.Request.GroupBy[0].Key]
			if g.G0 == nil && !present || g.G0 != nil && present && *g.G0 == value {
				group = append(group, row)
			}
		}
		if e = quantSummary397(read.Request, group, g.Numeric, g.Duration, nil); e != nil {
			return e
		}
	}
	if d := read.Response.ViewDataset; d != nil {
		for _, row := range d.Rows {
			group := []quantRaw397{}
			for _, raw := range chosen {
				attrs := raw.LogAttributes
				if read.Request.Signal == "spans" {
					attrs = raw.SpanAttributes
				}
				value, present := attrs[read.Request.GroupBy[0].Key]
				category := row.Values["category"]
				if category == nil && !present || category != nil && present && *category == value {
					group = append(group, raw)
				}
			}
			if e = quantSummary397(read.Request, group, nil, nil, &row); e != nil {
				return e
			}
		}
	}
	return nil
}

// Independent canonical integer parsing and rational arithmetic over physical
// records. No production numeric evaluator or rendered aggregate is the oracle.
func quantSummary397(q exploration.Query, rows []quantRaw397, numeric *exploration.NumericSummary, duration *exploration.DurationSummary, cell *scopedview.Row) error {
	total, valid, missing, empty, invalid := len(rows), 0, 0, 0, 0
	sum := new(big.Int)
	var lo, hi *big.Int
	for _, row := range rows {
		var n *big.Int
		if q.Numeric != nil {
			attrs := row.LogAttributes
			if q.Signal == "spans" {
				attrs = row.SpanAttributes
			}
			value, present := attrs[q.Numeric.Field.Key]
			if !present {
				missing++
				continue
			}
			if value == "" {
				empty++
				continue
			}
			parsed, ok := new(big.Int).SetString(value, 10)
			if !ok || parsed.String() != value || !parsed.IsInt64() {
				invalid++
				continue
			}
			n = parsed
		} else {
			var ok bool
			n, ok = new(big.Int).SetString(row.DurationNS, 10)
			if !ok || n.Sign() < 0 {
				return fmt.Errorf("invalid physical duration")
			}
		}
		valid++
		sum.Add(sum, n)
		if lo == nil || n.Cmp(lo) < 0 {
			lo = new(big.Int).Set(n)
		}
		if hi == nil || n.Cmp(hi) > 0 {
			hi = new(big.Int).Set(n)
		}
	}
	if numeric != nil {
		if numeric.Count != strconv.Itoa(total) || numeric.Valid != strconv.Itoa(valid) || numeric.Missing != strconv.Itoa(missing) || numeric.Empty != strconv.Itoa(empty) || numeric.Invalid != strconv.Itoa(invalid) || numeric.Sum != sum.String() || numeric.Unit != q.Numeric.Unit {
			return fmt.Errorf("quantitative numeric support/sum/unit differs from physical records")
		}
		if valid > 0 && (numeric.Min != lo.String() || numeric.Max != hi.String()) {
			return fmt.Errorf("numeric extrema differ")
		}
	}
	if duration != nil {
		if duration.Count != strconv.Itoa(total) || duration.SumNS != sum.String() || valid > 0 && (duration.MinNS != lo.String() || duration.MaxNS != hi.String()) {
			return fmt.Errorf("duration support/sum/extrema differs")
		}
	}
	if cell != nil {
		key := "duration.mean"
		if q.Numeric != nil {
			key = "numeric.mean"
		}
		support := &scopedview.Support{Total: strconv.Itoa(total), Valid: strconv.Itoa(valid), Missing: strconv.Itoa(missing), Empty: strconv.Itoa(empty), Invalid: strconv.Itoa(invalid)}
		actual, ok := cell.Cells[key]
		if !ok || !reflect.DeepEqual(actual.Support, support) {
			return fmt.Errorf("cell quantitative parse support differs")
		}
		value := cell.Values[key]
		if valid == 0 {
			if value != nil {
				return fmt.Errorf("unavailable mean invented zero")
			}
		} else {
			want := new(big.Rat).SetFrac(sum, big.NewInt(int64(valid))).RatString()
			if value == nil || *value != want {
				return fmt.Errorf("exact rational mean differs: got %v want %s", value, want)
			}
		}
	}
	return nil
}
func admitQuant397(r quantLedger397) error {
	fail := func(s string) error { return fmt.Errorf("quant397: %s", s) }
	if r.Version != "goo.distributionCustomize397/1" || len(r.Errors) != 0 || len(r.Cases) != 3 || len(r.Raw.Logs) != 18 || len(r.Raw.Spans) != 18 || len(r.Fixture.Records) != 36 {
		return fail("finite physical bank absent")
	}
	for _, signal := range []string{"logs", "spans"} {
		rows := r.Raw.Logs
		if signal == "spans" {
			rows = r.Raw.Spans
		}
		seen := map[string]bool{}
		for _, row := range rows {
			found := false
			for _, seed := range r.Fixture.Records {
				if seed.Signal != signal || seed.ID != row.SpanID {
					continue
				}
				if seen[row.SpanID] {
					return fail("duplicate physical identity")
				}
				seen[row.SpanID] = true
				found = true
				attrs := row.LogAttributes
				if signal == "spans" {
					attrs = row.SpanAttributes
					if row.DurationNS != seed.DurationNS {
						return fail("physical duration differs from OTLP seed")
					}
				}
				value, present := attrs["work.value"]
				if row.TraceID != seed.TraceID || row.TimestampNS != strconv.FormatInt(seed.TimeMS*1000000, 10) || row.ServiceName != seed.Service || row.ResourceAttributes["deployment.environment.name"] != seed.Environment || row.ResourceAttributes["service.namespace"] != seed.Namespace || attrs["team"] != seed.Team || attrs["work.class"] != seed.Phase || present != (seed.Numeric != nil) || present && value != *seed.Numeric {
					return fail("physical records differ from actual OTLP seed")
				}
			}
			if !found {
				return fail("foreign physical record")
			}
		}
	}
	for i, c := range r.Cases {
		if c.Kind != []string{"numeric-logs", "numeric-spans", "duration"}[i] || reflect.DeepEqual(c.SourceQuestion, c.DestinationQuestion) {
			return fail("case/query distinction lost")
		}
		rows := r.Raw.Logs
		if c.Signal == "spans" {
			rows = r.Raw.Spans
		}
		measure := "numeric.mean"
		if c.Kind == "duration" {
			measure = "duration.mean"
		}
		for _, read := range []customizeRead394{c.Source, c.ChildRun, c.Replacement, c.DestinationRun} {
			if e := quantRead397(read, rows); e != nil {
				return e
			}
		}
		if !reflect.DeepEqual(c.Source.Request, c.SourceQuestion) || !reflect.DeepEqual(c.ChildRun.Request, c.SourceQuestion) || !reflect.DeepEqual(c.Replacement.Request, c.SourceQuestion) || !reflect.DeepEqual(c.DestinationRun.Request, c.DestinationQuestion) {
			return fail("full question changed")
		}
		if c.Staged.ChartCount != 0 || !c.Staged.Event.Stage || !reflect.DeepEqual(c.Staged.Event.Question, c.SourceQuestion) || !reflect.DeepEqual(c.Staged.Event.View.Queries, []exploration.Query{c.SourceQuestion}) || c.Staged.Event.View.Validate() != nil || c.Staged.Event.View.Displays[c.Signal].Display.Value != measure {
			return fail("staging changed measure/question or borrowed evidence")
		}
		if c.ParentBefore.Find != "read" || !sameTransport384(c.ParentBefore, c.ParentAfter) || !sameTransport384(c.ParentBefore.Snapshot, c.Source.Response.ViewDataset) || !sameTransport384(c.Customized.Snapshot, c.ChildRun.Response.ViewDataset) || !sameTransport384(c.DestinationSnapshot.Snapshot, c.DestinationRun.Response.ViewDataset) || quantValue397(c.Customized.Display) != measure || quantValue397(c.DestinationSnapshot.Display) != measure {
			return fail("local/held presentation or data changed")
		}
		for _, guard := range []customizeWrapper394{c.Hidden, c.Inert, c.Pending, c.Superseded} {
			if !guard.Invoked || guard.Before.Events != guard.After.Events {
				return fail("retired producer acquired authority")
			}
		}
		if !c.Hidden.Before.Connected || !c.Hidden.Before.Hidden || !c.Inert.Before.Connected || !c.Inert.Before.Inert {
			return fail("attached refusals not exercised")
		}
		for _, saved := range []customizeSaved394{c.Saved, c.DestinationSetup} {
			if saved.Status != 201 || saved.Response.Validate() != nil || !reflect.DeepEqual(saved.Request.Queries, saved.Response.Queries) {
				return fail("configuration question changed")
			}
			copy := saved.Response
			id := copy.ID
			copy.ID = ""
			hash, e := workbench.ViewContentID(copy)
			if e != nil || hash != id {
				return fail("content identity changed")
			}
		}
		if !reflect.DeepEqual(c.Saved.Response.Queries, []exploration.Query{c.SourceQuestion}) || !reflect.DeepEqual(c.DestinationSetup.Response.Queries, []exploration.Query{c.DestinationQuestion}) || c.Saved.Response.Displays[c.Signal].Display.Value != measure {
			return fail("saved selected measure or destination scope changed")
		}
	}
	if len(r.Controls) != 9 {
		return fail("unsupported acceptance controls missing")
	}
	for _, c := range r.Controls {
		if c.Name == "failed" {
			if c.Status < 400 || c.Present != 0 {
				return fail("failed source acquired continuation")
			}
			continue
		}
		rows := r.Raw.Logs
		if c.Signal == "spans" {
			rows = r.Raw.Spans
		}
		if e := quantRead397(c.Source, rows); e != nil {
			return e
		}
		if c.Present > 0 && (c.Disabled == nil || !*c.Disabled) {
			return fail("unsupported continuation enabled")
		}
		if c.Name == "offset" && c.Source.Request.Offset != 1 || c.Name == "empty" && c.Source.Response.Count != "0" {
			return fail("unsupported control condition absent")
		}
	}

	// Real source requests each own only their existing bounded physical plan.
	// Failed acquisition attests its attempted Count, not a fabricated result.
	type planned struct {
		sql          string
		params       map[string]string
		failed, used bool
	}
	expected := []planned{}
	add := func(q exploration.Query, failed bool) error {
		p, e := exploration.Compile(q)
		if e != nil {
			return e
		}
		sqls := []string{p.Count}
		if !failed {
			sqls = append(sqls, p.Rows, p.Groups, p.Fields)
			if p.TimeSeries != "" {
				sqls = append(sqls, p.TimeSeries)
			}
			if p.GroupedTimeSeries != "" {
				sqls = append(sqls, p.GroupedTimeSeries)
			}
		}
		for _, sql := range sqls {
			if sql == "" {
				return fail("incomplete physical plan")
			}
			expected = append(expected, planned{sql: sql, params: p.Parameters, failed: failed})
		}
		return nil
	}
	for _, c := range r.Cases {
		for _, read := range []customizeRead394{c.Source, c.ChildRun, c.Replacement, c.DestinationRun} {
			if e := add(read.Request, false); e != nil {
				return e
			}
		}
	}
	for _, c := range r.Controls {
		q := c.Source.Request
		if c.Name == "failed" {
			q = c.Request
		}
		if e := add(q, c.Name == "failed"); e != nil {
			return e
		}
	}
	if len(r.ProxyCalls) != len(expected) {
		return fail(fmt.Sprintf("physical SQL budget %d want %d", len(r.ProxyCalls), len(expected)))
	}
	for _, call := range r.ProxyCalls {
		found := false
		for i := range expected {
			want := &expected[i]
			if want.used || call.SQL != want.sql || want.failed != (call.Status >= 400) {
				continue
			}
			match := call.Parameters["readonly"] == "1"
			for key, value := range want.params {
				if call.Parameters["param_"+key] != value {
					match = false
				}
			}
			if !match {
				continue
			}
			want.used = true
			found = true
			break
		}
		if !found {
			return fail("physical source SQL/status/readonly/typed parameters changed")
		}
	}
	names := quantNames397()
	if len(r.Facts) != len(names) {
		return fail(fmt.Sprintf("unfinished facts %d want %d", len(r.Facts), len(names)))
	}
	for i, f := range r.Facts {
		if f.Name != names[i] || i > 0 && f.Before != r.Facts[i-1].After {
			return fail("counter continuity changed")
		}
		read, store, config := quantCost397(f.Name)
		if f.After.Source-f.Before.Source != read || f.After.Store-f.Before.Store != store || f.After.ConfigWrites-f.Before.ConfigWrites != config || read == 0 && f.After.SQL != f.Before.SQL {
			return fail("local action source/store budget changed")
		}
	}
	return nil
}
func admitQuantActor397(r quantLedger397, a customizeActor394) error {
	if a.Version != "goo.distributionCustomize397ActorReport/1" || a.SourceBank != r.Fixture.Token || !reflect.DeepEqual(a.Question, r.Cases[0].SourceQuestion) || a.Actor.Snapshots <= 0 || a.Actor.Complete.Version != "goo.distributionCustomize397Actor/1" || a.Actor.Complete.Phase != "complete" || len(a.Actor.Complete.Requests) != 2 || !slices.Contains(a.Actor.Phases, "staged") {
		return fmt.Errorf("397 actor actual completion absent")
	}
	if e := quantRead397(a.SourceFrame, r.Raw.Logs); e != nil {
		return e
	}
	for _, read := range a.Actor.Complete.Requests {
		if read.Path != "/api/explore" || !reflect.DeepEqual(read.Request, a.Question) {
			return fmt.Errorf("actor question changed")
		}
		if e := quantRead397(customizeRead394{read.Request, read.Response, read.Status}, r.Raw.Logs); e != nil {
			return e
		}
	}
	if a.Actor.Complete.Parent.Find != "read" || quantValue397(a.Actor.Complete.Parent.Display) != "numeric.mean" || !sameTransport384(a.Actor.Complete.Parent.Snapshot, a.Actor.Complete.Requests[0].Response.ViewDataset) {
		return fmt.Errorf("actor held parent changed")
	}
	if a.SQLBefore < 0 || a.SQLAfter > len(a.ProxyCalls) || a.SQLAfter-a.SQLBefore != 8 {
		return fmt.Errorf("actor physical source budget changed")
	}
	plan, e := exploration.Compile(a.Question)
	if e != nil {
		return e
	}
	expected := map[string]int{plan.Count: 2, plan.Rows: 2, plan.Groups: 2, plan.Fields: 2}
	for _, call := range a.ProxyCalls[a.SQLBefore:a.SQLAfter] {
		if call.Status != 200 || call.Parameters["readonly"] != "1" || expected[call.SQL] <= 0 {
			return fmt.Errorf("actor physical plan changed")
		}
		expected[call.SQL]--
		for key, value := range plan.Parameters {
			if call.Parameters["param_"+key] != value {
				return fmt.Errorf("actor typed parameter changed")
			}
		}
	}
	for _, left := range expected {
		if left != 0 {
			return fmt.Errorf("actor physical reads incomplete")
		}
	}
	return nil
}
func TestQuintQuantitativeCustomizationObserved397(t *testing.T) {
	path := os.Getenv("GOO_DISTRIBUTION_CUSTOMIZE_397_REPORT")
	if path == "" {
		t.Skip("actual quantitative journey required")
	}
	wire, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var r quantLedger397
	if e = json.Unmarshal(wire, &r); e != nil {
		t.Fatal(e)
	}
	if e = admitQuant397(r); e != nil {
		t.Fatal(e)
	}
	actorPath := os.Getenv("GOO_DISTRIBUTION_CUSTOMIZE_397_ACTOR")
	if actorPath == "" {
		actorPath = filepath.Join(filepath.Dir(path), "actor.json")
	}
	actorWire, e := os.ReadFile(actorPath)
	if e != nil {
		t.Fatal(e)
	}
	var a customizeActor394
	if e = json.Unmarshal(actorWire, &a); e != nil {
		t.Fatal(e)
	}
	if e = admitQuantActor397(r, a); e != nil {
		t.Fatal(e)
	}
	for i, mutate := range []func(*quantLedger397){func(r *quantLedger397) { r.Cases[0].Staged.Event.Question.Limit++ }, func(r *quantLedger397) { r.Cases[0].Saved.Response.Displays["logs"] = workbench.ViewDeclaration{} }, func(r *quantLedger397) { r.Cases[0].ParentAfter.Snapshot.CapturedAt = "changed" }, func(r *quantLedger397) { r.Cases[0].Hidden.After.Events++ }, func(r *quantLedger397) { r.Cases[0].Source.Response.Numeric.Invalid = "0" }, func(r *quantLedger397) { r.Raw.Spans[0].DurationNS = "123" }, func(r *quantLedger397) { r.Facts[4].After.Source++ }} {
		var bad quantLedger397
		json.Unmarshal(wire, &bad)
		mutate(&bad)
		if admitQuant397(bad) == nil {
			t.Fatalf("quantitative semantic mutant%d escaped", i)
		}
	}
	if !replayQuant397(t, r, false) || replayQuant397(t, r, true) {
		t.Fatal("actual model or staging-read negative control incorrect")
	}
	t.Logf("admitted three actual quantitative cases,36 physical records,%d facts; seven semantic mutants and staging-read control rejected", len(r.Facts))
}
func replayQuant397(t *testing.T, r quantLedger397, mutate bool) bool {
	t.Helper()
	reads, stores, configs := []int{}, []int{}, []int{}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, f := range r.Facts {
		read, store, config := quantCost397(f.Name)
		reads = append(reads, read)
		stores = append(stores, store)
		configs = append(configs, config)
		source := f.After.Source
		if mutate && i == 4 {
			source++
		}
		action := "step"
		if f.Name == "complete" {
			action = "finish"
		}
		rec.Record(context.Background(), action, nil, "obs.phase", i+1, "obs.reads", source, "obs.stores", f.After.Store, "obs.configs", f.After.ConfigWrites, "obs.finished", f.Name == "complete")
	}
	list := func(x []int) string { wire, _ := json.Marshal(x); return string(wire) }
	spec, e := filepath.Abs("model/quantitative_customize397.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: quantitative_customize397\nconstants: {READS: '%s', STORES: '%s', CONFIGS: '%s'}\ninvariants: [safety]\nactions:\n", spec, list(reads), list(stores), list(configs))
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
	return result.OK()
}
