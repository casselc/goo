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

type summaryRaw400 struct {
	quantRaw397
	StatusCode string
}
type summaryFrame400 struct {
	Name     string
	Snapshot scopedview.Dataset
	Display  struct {
		Fields  []string
		Title   string
		Compact bool
	}
	FieldOrder []string
}
type summaryWrapper400 struct {
	Before, After int
	Invoked       bool
}
type summaryCase400 struct {
	Kind, Signal, Title                           string
	SourceQuestion, DestinationQuestion           exploration.Query
	Source, ChildRun, Replacement, DestinationRun customizeRead394
	Hidden, Inert, Pending, Superseded            summaryWrapper400
	Fields                                        []string
	ParentBefore, ParentAfter                     summaryFrame400
	SummarySnapshots                              []summaryFrame400
	Staged                                        struct {
		Event struct {
			Question exploration.Query
			View     workbench.ExplorationView
			Stage    bool
		}
		SummaryCount int
	}
	Saved, DestinationSetup customizeSaved394
	Reuse                   struct {
		Question     exploration.Query
		Composition  workbench.ViewComposition
		SummaryCount int
	}
}
type summaryLedger400 struct {
	Version string
	Fixture struct {
		Token   string
		Records []struct {
			Signal, ID, TraceID, Service, Environment, Namespace, Team, Phase, DurationNS string
			TimeMS                                                                        int64
			Numeric                                                                       *string
			Status                                                                        int
		}
	}
	Raw      struct{ Logs, Spans []summaryRaw400 }
	Cases    []summaryCase400
	Controls []struct {
		Name     string
		Question exploration.Query
		Read     customizeRead394
		Snapshot scopedview.Dataset
		Display  json.RawMessage
	}
	Facts      []patternFact391
	Errors     []string
	ProxyCalls []literalWire393
}

// Membership is reconstructed from physical columns and exact typed predicates;
// it deliberately has no grouping prerequisite and consults no summary values.
func summaryMembers400(q exploration.Query, raw []summaryRaw400) ([]summaryRaw400, error) {
	if q.Validate() != nil {
		return nil, fmt.Errorf("invalid summary source question")
	}
	out := []summaryRaw400{}
	seen := map[string]bool{}
	for _, row := range raw {
		at, e := strconv.ParseInt(row.TimestampNS, 10, 64)
		if e != nil {
			return nil, e
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
		if seen[row.SpanID] {
			return nil, fmt.Errorf("duplicate physical identity")
		}
		seen[row.SpanID] = true
		out = append(out, row)
	}
	return out, nil
}
func summaryDataset400(q exploration.Query, rows []summaryRaw400, d *scopedview.Dataset, basis, captured string) error {
	fail := func(s string) error { return fmt.Errorf("summary400 dataset: %s", s) }
	if d == nil || d.Shape != "summary" || d.Revision != basis || d.Scope.Generation != basis || d.ID != d.Scope.ID+":summary" || d.Scope.Signal != q.Signal || d.CapturedAt != captured || !sameTransport384(d.Scope.Selector, q) || len(d.Rows) != 1 || d.Rows[0].Key != "population" || d.Partition != nil || d.Interval != nil || len(d.Dimensions) != 0 || d.Quality.Extent != "complete" || d.Quality.Basis != "captured-query" || d.Quality.Sampling != "unknown" || d.Quality.Value != "derived" || d.Completeness == nil || d.Completeness.Revision != basis || d.Completeness.ScopeID != d.Scope.ID || !reflect.DeepEqual(d.Completeness.Aspects, []string{"summary"}) {
		return fail("identity, clock or supplied summary guarantee changed")
	}
	members, e := summaryMembers400(q, rows)
	if e != nil {
		return e
	}
	count := strconv.Itoa(len(members))
	row := d.Rows[0]
	if row.Values["count"] == nil || *row.Values["count"] != count {
		return fail("whole-population count differs")
	}
	expected := map[string]*string{}
	cells := map[string]scopedview.Cell{}
	put := func(key string, v *string, state string, support *scopedview.Support) {
		expected[key] = v
		cells[key] = scopedview.Cell{State: state, Support: support}
	}
	text := func(v string) *string { return &v }
	if q.Numeric != nil {
		if q.Numeric.Type != "int64" || q.Numeric.Scale != 0 {
			return fail("finite oracle requires recorded canonical int64 binding")
		}
		values := []*big.Int{}
		missing, empty, invalid := 0, 0, 0
		sum := new(big.Int)
		for _, r := range members {
			attrs := r.LogAttributes
			if q.Signal == "spans" {
				attrs = r.SpanAttributes
			}
			v, ok := attrs[q.Numeric.Field.Key]
			if !ok {
				missing++
				continue
			}
			if v == "" {
				empty++
				continue
			}
			n, ok := new(big.Int).SetString(v, 10)
			if !ok || n.String() != v || !n.IsInt64() {
				invalid++
				continue
			}
			values = append(values, n)
			sum.Add(sum, n)
		}
		support := &scopedview.Support{Total: count, Valid: strconv.Itoa(len(values)), Missing: strconv.Itoa(missing), Empty: strconv.Itoa(empty), Invalid: strconv.Itoa(invalid)}
		state := "derived"
		if len(values) == 0 {
			state = "no_valid_values"
			if len(members) == 0 {
				state = "no_data"
			}
		}
		for _, kind := range []string{"mean", "sum", "min", "max"} {
			put("numeric."+kind, nil, state, support)
		}
		if len(values) > 0 {
			slices.SortFunc(values, func(a, b *big.Int) int { return a.Cmp(b) })
			expected["numeric.mean"] = text(new(big.Rat).SetFrac(sum, big.NewInt(int64(len(values)))).RatString())
			expected["numeric.sum"] = text(sum.String())
			expected["numeric.min"] = text(values[0].String())
			expected["numeric.max"] = text(values[len(values)-1].String())
		}
	}
	if q.Summary == "span-duration" {
		values := []*big.Int{}
		sum := new(big.Int)
		for _, r := range members {
			n, ok := new(big.Int).SetString(r.DurationNS, 10)
			if !ok || n.Sign() < 0 || n.String() != r.DurationNS {
				return fail("malformed physical duration")
			}
			values = append(values, n)
			sum.Add(sum, n)
		}
		support := &scopedview.Support{Total: count, Valid: count, Missing: "0", Empty: "0", Invalid: "0"}
		state := "derived"
		if len(values) == 0 {
			state = "no_data"
		}
		for _, kind := range []string{"mean", "min", "max", "p50", "p95", "p99"} {
			put("duration."+kind, nil, state, support)
		}
		if len(values) > 0 {
			slices.SortFunc(values, func(a, b *big.Int) int { return a.Cmp(b) })
			expected["duration.mean"] = text(new(big.Rat).SetFrac(sum, big.NewInt(int64(len(values)))).RatString())
			expected["duration.min"] = text(values[0].String())
			expected["duration.max"] = text(values[len(values)-1].String())
			for key, p := range map[string]int{"p50": 50, "p95": 95, "p99": 99} {
				rank := len(values) * p / 100
				if rank >= len(values) {
					rank = len(values) - 1
				}
				expected["duration."+key] = text(values[rank].String())
			}
		}
	}
	if q.SpanStatus {
		counts := map[string]int{"error": 0, "ok": 0, "unset": 0, "unknown": 0}
		for _, r := range members {
			switch r.StatusCode {
			case "Error":
				counts["error"]++
			case "Ok":
				counts["ok"]++
			case "Unset":
				counts["unset"]++
			default:
				counts["unknown"]++
			}
		}
		for key, v := range counts {
			put("status."+key, text(strconv.Itoa(v)), "derived", nil)
		}
		put("status.error_fraction", nil, "no_data", nil)
		if len(members) > 0 {
			put("status.error_fraction", text(new(big.Rat).SetFrac64(int64(counts["error"]), int64(len(members))).RatString()), "derived", nil)
		}
	}
	if len(d.Fields) != len(expected)+1 || len(row.Values) != len(expected)+1 || len(row.Cells) != len(cells) {
		return fail("supplied field/cell set changed")
	}
	for key, value := range expected {
		if !reflect.DeepEqual(row.Values[key], value) || !reflect.DeepEqual(row.Cells[key], cells[key]) {
			return fail("exact value/state/support changed for " + key)
		}
	}
	for _, field := range d.Fields {
		if field.Key == "count" {
			if field.Type != "number" || field.Unit != "records" || field.Aggregation != nil {
				return fail("count meaning changed")
			}
			continue
		}
		family, kind, _ := strings.Cut(field.Key, ".")
		unit := "ns"
		if family == "numeric" {
			unit = q.Numeric.Unit
		}
		if family == "status" {
			unit = "records"
			if kind == "error_fraction" {
				unit = "1"
			}
		}
		if field.Unit != unit || field.Aggregation == nil {
			return fail("measure unit/method missing")
		}
		method := "exact"
		if kind == "mean" || kind == "error_fraction" {
			method = "exact-rational"
		}
		if strings.HasPrefix(kind, "p") {
			method = "quantilesExactHigh"
		}
		if field.Aggregation.Method != method {
			return fail("measure method changed")
		}
	}
	return nil
}
func summaryRead400(read customizeRead394, raw []summaryRaw400) error {
	if read.Status != 200 || !reflect.DeepEqual(read.Request, read.Response.Question) || read.Response.State != "observed" && read.Response.State != "no_data" {
		return fmt.Errorf("summary read status/question/state changed")
	}
	members, e := summaryMembers400(read.Request, raw)
	if e != nil {
		return e
	}
	if read.Response.Count != strconv.Itoa(len(members)) || len(read.Response.Rows) != min(read.Request.Limit, len(members)) || len(read.Request.GroupBy) != 0 {
		return fmt.Errorf("summary no-group population/page changed")
	}
	ids := map[string]bool{}
	for _, r := range members {
		ids[r.SpanID] = true
	}
	for _, wire := range read.Response.Rows {
		var r struct {
			SpanID string `json:"SpanId"`
		}
		if json.Unmarshal(wire, &r) != nil || !ids[r.SpanID] {
			return fmt.Errorf("foreign source page member")
		}
	}
	return summaryDataset400(read.Request, raw, read.Response.SummaryDataset, read.Response.Basis, read.Response.CapturedAt)
}

func summaryNames400() []string {
	names := []string{}
	for _, kind := range []string{"numeric-logs", "numeric-spans", "duration"} {
		for _, suffix := range []string{"source-run", "hidden-wrapper", "inert-wrapper"} {
			names = append(names, kind+"-"+suffix)
		}
		if kind == "numeric-logs" {
			for _, suffix := range []string{"replacement-read", "pending-wrapper", "superseded-wrapper"} {
				names = append(names, kind+"-"+suffix)
			}
		}
		for _, suffix := range []string{"customize-stage", "explicit-child-run", "local-fields-order", "save-configuration", "held-back", "saved-reopen", "destination-inspect-apply", "explicit-destination-run"} {
			names = append(names, kind+"-"+suffix)
		}
	}
	return append(names, "control-empty", "control-no_valid", "control-zero", "complete")
}
func summaryCost400(name string) (read, store, config int) {
	if strings.HasSuffix(name, "-source-run") || strings.HasSuffix(name, "-replacement-read") || strings.HasSuffix(name, "-explicit-child-run") || strings.HasSuffix(name, "-explicit-destination-run") || strings.HasPrefix(name, "control-") {
		read = 1
	}
	if strings.HasSuffix(name, "-saved-reopen") {
		store = 1
	}
	if strings.HasSuffix(name, "-destination-inspect-apply") {
		store = 2
	}
	if strings.HasSuffix(name, "-save-configuration") {
		config = 1
	}
	return
}
func summaryLeaf400(c *workbench.ViewComposition) *workbench.SummaryViewDeclaration {
	if c != nil {
		for _, child := range c.Children {
			if child.Summary != nil {
				return child.Summary
			}
		}
	}
	return nil
}
func summaryPhysical400(r summaryLedger400) error {
	if len(r.Raw.Logs) != 18 || len(r.Raw.Spans) != 18 || len(r.Fixture.Records) != 36 {
		return fmt.Errorf("physical bank incomplete")
	}
	for _, signal := range []string{"logs", "spans"} {
		raw := r.Raw.Logs
		if signal == "spans" {
			raw = r.Raw.Spans
		}
		seen := map[string]bool{}
		for _, row := range raw {
			found := false
			for _, seed := range r.Fixture.Records {
				if seed.Signal != signal || seed.ID != row.SpanID {
					continue
				}
				if seen[row.SpanID] {
					return fmt.Errorf("duplicate physical source identity")
				}
				seen[row.SpanID] = true
				found = true
				attrs := row.LogAttributes
				if signal == "spans" {
					attrs = row.SpanAttributes
					if row.DurationNS != seed.DurationNS || row.StatusCode != []string{"Unset", "Ok", "Error"}[seed.Status] {
						return fmt.Errorf("physical duration/status differs from actual OTLP seed")
					}
				}
				value, present := attrs["work.value"]
				if row.TimestampNS != strconv.FormatInt(seed.TimeMS*1000000, 10) || row.TraceID != seed.TraceID || row.ServiceName != seed.Service || row.ResourceAttributes["deployment.environment.name"] != seed.Environment || row.ResourceAttributes["service.namespace"] != seed.Namespace || attrs["team"] != seed.Team || attrs["work.class"] != seed.Phase || present != (seed.Numeric != nil) || present && value != *seed.Numeric {
					return fmt.Errorf("physical attributes/timestamp differ from seed")
				}
			}
			if !found {
				return fmt.Errorf("foreign physical source member")
			}
		}
	}
	return nil
}
func summaryPlans400(reads []customizeRead394, calls []literalWire393) error {
	type planned struct {
		sql    string
		params map[string]string
		used   bool
	}
	want := []planned{}
	for _, read := range reads {
		p, e := exploration.Compile(read.Request)
		if e != nil {
			return e
		}
		for _, sql := range []string{p.Count, p.Rows, p.Groups, p.Fields, p.TimeSeries, p.GroupedTimeSeries} {
			if sql != "" {
				want = append(want, planned{sql: sql, params: p.Parameters})
			}
		}
	}
	if len(want) != len(calls) {
		return fmt.Errorf("physical plan budget %d want %d", len(calls), len(want))
	}
	for _, call := range calls {
		found := false
		for i := range want {
			p := &want[i]
			if p.used || call.SQL != p.sql || call.Status != 200 || call.Parameters["readonly"] != "1" {
				continue
			}
			same := true
			for key, value := range p.params {
				if call.Parameters["param_"+key] != value {
					same = false
				}
			}
			if same {
				p.used = true
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("physical SQL/status/readonly/typed parameters changed")
		}
	}
	return nil
}
func admitSummary400(r summaryLedger400) error {
	fail := func(s string) error { return fmt.Errorf("summary400: %s", s) }
	if r.Version != "goo.populationSummary400/1" || len(r.Errors) != 0 || len(r.Cases) != 3 || len(r.Controls) != 3 {
		return fail("unfinished finite journey")
	}
	if e := summaryPhysical400(r); e != nil {
		return e
	}
	reads := []customizeRead394{}
	for i, c := range r.Cases {
		if c.Kind != []string{"numeric-logs", "numeric-spans", "duration"}[i] || reflect.DeepEqual(c.SourceQuestion, c.DestinationQuestion) || c.SourceQuestion.Limit != 3 || len(c.SourceQuestion.GroupBy) != 0 || len(c.DestinationQuestion.GroupBy) != 0 {
			return fail("case/no-group scope distinction changed")
		}
		raw := r.Raw.Logs
		if c.Signal == "spans" {
			raw = r.Raw.Spans
		}
		accepted := c.Source
		for _, read := range []customizeRead394{c.Source, c.ChildRun, c.DestinationRun} {
			if e := summaryRead400(read, raw); e != nil {
				return e
			}
			reads = append(reads, read)
		}
		if c.Kind == "numeric-logs" {
			if e := summaryRead400(c.Replacement, raw); e != nil {
				return e
			}
			reads = append(reads, c.Replacement)
			accepted = c.Replacement
			if c.Pending.After != c.Pending.Before || c.Superseded.After != c.Superseded.Before {
				return fail("retired producer staged current source")
			}
		}
		if !reflect.DeepEqual(c.Source.Request, c.SourceQuestion) || !reflect.DeepEqual(c.ChildRun.Request, c.SourceQuestion) || !reflect.DeepEqual(c.DestinationRun.Request, c.DestinationQuestion) || c.Hidden.Before != c.Hidden.After || !c.Hidden.Invoked || c.Inert.Before != c.Inert.After || !c.Inert.Invoked {
			return fail("fixed query or hidden/inert ownership changed")
		}
		if !reflect.DeepEqual(c.ParentBefore, c.ParentAfter) || !sameTransport384(c.ParentBefore.Snapshot, accepted.Response.SummaryDataset) {
			return fail("held return changed summary evidence or presentation")
		}
		if !c.Staged.Event.Stage || c.Staged.SummaryCount != 0 || !reflect.DeepEqual(c.Staged.Event.Question, c.SourceQuestion) || len(c.Staged.Event.View.Queries) != 1 || !reflect.DeepEqual(c.Staged.Event.View.Queries[0], c.SourceQuestion) || c.Staged.Event.View.Validate() != nil {
			return fail("staging copied data or altered question")
		}
		if c.Saved.Status != 201 || c.Saved.Response.Validate() != nil || len(c.Saved.Response.Queries) != 1 || !reflect.DeepEqual(c.Saved.Response.Queries[0], c.SourceQuestion) {
			return fail("saved carrier invalid or altered scope")
		}
		saved := summaryLeaf400(c.Saved.Response.Composition)
		if saved == nil || !reflect.DeepEqual(saved.Display.Fields, c.Fields) {
			return fail("saved selected fields changed")
		}
		if !reflect.DeepEqual(c.Reuse.Question, c.DestinationQuestion) || c.Reuse.SummaryCount != 0 || !reflect.DeepEqual(c.Reuse.Composition, *c.Saved.Response.Composition) {
			return fail("reuse copied donor data or changed display")
		}
		if _, e := workbench.ReuseViewComposition(*c.Saved.Response.Composition, c.SourceQuestion, c.DestinationQuestion); e != nil {
			return fail("actual compatible reuse rejected")
		}
		if len(c.SummarySnapshots) != 3 {
			return fail("missing rendered summary stages")
		}
		for j, frame := range c.SummarySnapshots {
			read := c.ChildRun
			if frame.Name == "destination" {
				read = c.DestinationRun
			}
			if e := summaryDataset400(read.Request, raw, &frame.Snapshot, read.Response.Basis, read.Response.CapturedAt); e != nil {
				return e
			}
			if !frame.Display.Compact || len(frame.Display.Fields) == 0 || len(frame.Display.Fields) > 6 || !reflect.DeepEqual(frame.Display.Fields, frame.FieldOrder) {
				return fail("visible selected field order/compactness changed")
			}
			if j > 0 && !reflect.DeepEqual(frame.Display.Fields, c.Fields) {
				return fail("authored field order lost")
			}
		}
	}
	for i, c := range r.Controls {
		if c.Name != []string{"empty", "no_valid", "zero"}[i] || !reflect.DeepEqual(c.Question, c.Read.Request) {
			return fail("control question changed")
		}
		if e := summaryRead400(c.Read, r.Raw.Logs); e != nil {
			return e
		}
		if e := summaryDataset400(c.Question, r.Raw.Logs, &c.Snapshot, c.Read.Response.Basis, c.Read.Response.CapturedAt); e != nil {
			return e
		}
		reads = append(reads, c.Read)
		cell := c.Snapshot.Rows[0].Cells["numeric.mean"]
		switch c.Name {
		case "empty":
			if c.Read.Response.Count != "0" || cell.State != "no_data" {
				return fail("empty control is not empty")
			}
		case "no_valid":
			if c.Read.Response.Count == "0" || cell.State != "no_valid_values" {
				return fail("invalid numeric control collapsed")
			}
		case "zero":
			if cell.State != "derived" || c.Snapshot.Rows[0].Values["numeric.mean"] == nil || *c.Snapshot.Rows[0].Values["numeric.mean"] != "0" {
				return fail("real zero mean unavailable")
			}
		}
	}
	if e := summaryPlans400(reads, r.ProxyCalls); e != nil {
		return e
	}
	names := summaryNames400()
	if len(r.Facts) != len(names) {
		return fail(fmt.Sprintf("unfinished facts %d want %d", len(r.Facts), len(names)))
	}
	for i, f := range r.Facts {
		if f.Name != names[i] || i > 0 && f.Before != r.Facts[i-1].After {
			return fail("counter continuity or phase changed")
		}
		read, store, config := summaryCost400(f.Name)
		if f.After.Source-f.Before.Source != read || f.After.Store-f.Before.Store != store || f.After.ConfigWrites-f.Before.ConfigWrites != config || read == 0 && f.After.SQL != f.Before.SQL {
			return fail("local action source/store budget changed: " + f.Name)
		}
	}
	return nil
}
func admitSummaryActor400(r summaryLedger400, a customizeActor394, configured []string) error {
	if a.Version != "goo.populationSummary400ActorReport/1" || a.SourceBank != r.Fixture.Token || !reflect.DeepEqual(a.Question, r.Cases[0].SourceQuestion) || a.Actor.Snapshots < 1 || a.Actor.Complete.Version != "goo.populationSummary400Actor/1" || a.Actor.Complete.Phase != "complete" || len(a.Actor.Complete.Requests) != 2 || !slices.Contains(a.Actor.Phases, "staged") || !reflect.DeepEqual(configured, []string{"numeric.sum", "count", "numeric.mean"}) {
		return fmt.Errorf("independent actor completion/configuration missing")
	}
	reads := []customizeRead394{}
	for _, request := range a.Actor.Complete.Requests {
		read := customizeRead394{Request: request.Request, Response: request.Response, Status: request.Status}
		if request.Path != "/api/explore" || !reflect.DeepEqual(request.Request, a.Question) {
			return fmt.Errorf("actor source question changed")
		}
		if e := summaryRead400(read, r.Raw.Logs); e != nil {
			return e
		}
		reads = append(reads, read)
	}
	if !sameTransport384(a.Actor.Complete.Parent.Snapshot, reads[0].Response.SummaryDataset) {
		return fmt.Errorf("actor held parent evidence changed")
	}
	if a.SQLBefore != 0 || a.SQLAfter != len(a.ProxyCalls) {
		return fmt.Errorf("actor physical read boundary changed")
	}
	return summaryPlans400(reads, a.ProxyCalls)
}

func TestQuintSummaryViewObserved400(t *testing.T) {
	path := os.Getenv("GOO_SUMMARY_VIEW_400_REPORT")
	if path == "" {
		t.Skip("actual summary authoring journey report required")
	}
	wire, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var r summaryLedger400
	if e = json.Unmarshal(wire, &r); e != nil {
		t.Fatal(e)
	}
	if e = admitSummary400(r); e != nil {
		t.Fatal(e)
	}
	actorPath := os.Getenv("GOO_SUMMARY_VIEW_400_ACTOR")
	if actorPath == "" {
		actorPath = filepath.Join(filepath.Dir(path), "actor.json")
	}
	actorWire, e := os.ReadFile(actorPath)
	if e != nil {
		t.Fatal(e)
	}
	var actor customizeActor394
	var extra struct {
		Actor struct {
			Complete struct{ ConfiguredFields []string }
		}
	}
	if e = json.Unmarshal(actorWire, &actor); e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(actorWire, &extra); e != nil {
		t.Fatal(e)
	}
	if e = admitSummaryActor400(r, actor, extra.Actor.Complete.ConfiguredFields); e != nil {
		t.Fatal(e)
	}
	mutants := []func(*summaryLedger400){
		func(x *summaryLedger400) { x.Cases[0].Staged.Event.Question.Limit++ },
		func(x *summaryLedger400) {
			summaryLeaf400(x.Cases[0].Saved.Response.Composition).Display.Fields = []string{"count"}
		},
		func(x *summaryLedger400) { x.Cases[0].ParentAfter.Snapshot.CapturedAt = "changed" },
		func(x *summaryLedger400) { x.Cases[0].Hidden.After++ },
		func(x *summaryLedger400) {
			x.Cases[0].ChildRun.Response.SummaryDataset.Rows[0].Values["numeric.mean"] = nil
		},
		func(x *summaryLedger400) { x.Raw.Spans[0].StatusCode = "Unknown" },
		func(x *summaryLedger400) {
			x.Controls[1].Snapshot.Rows[0].Cells["numeric.mean"] = scopedview.Cell{State: "derived"}
		},
		func(x *summaryLedger400) { x.Facts[6].After.Source++ },
	}
	for i, mutate := range mutants {
		var bad summaryLedger400
		json.Unmarshal(wire, &bad)
		mutate(&bad)
		if admitSummary400(bad) == nil {
			t.Fatalf("summary semantic mutant%d escaped", i)
		}
	}
	if !replaySummary400(t, r, false) || replaySummary400(t, r, true) {
		t.Fatal("actual summary model or staging-read negative control incorrect")
	}
	t.Logf("36 physical records, three no-group summary cases and three support controls, %d facts; eight semantic mutants and staging-read control rejected", len(r.Facts))
}
func replaySummary400(t *testing.T, r summaryLedger400, mutate bool) bool {
	t.Helper()
	reads, stores, configs := []int{}, []int{}, []int{}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, f := range r.Facts {
		read, store, config := summaryCost400(f.Name)
		reads = append(reads, read)
		stores = append(stores, store)
		configs = append(configs, config)
		source := f.After.Source
		if mutate && f.Name == "numeric-logs-customize-stage" {
			source++
		}
		action := "step"
		if f.Name == "complete" {
			action = "finish"
		}
		rec.Record(context.Background(), action, nil, "obs.phase", i+1, "obs.reads", source, "obs.stores", f.After.Store, "obs.configs", f.After.ConfigWrites, "obs.finished", f.Name == "complete")
	}
	list := func(v []int) string { b, _ := json.Marshal(v); return string(b) }
	spec, e := filepath.Abs("model/summary_view400.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: summary_view400\nconstants: {READS: '%s', STORES: '%s', CONFIGS: '%s'}\ninvariants: [safety]\nactions:\n", spec, list(reads), list(stores), list(configs))
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
