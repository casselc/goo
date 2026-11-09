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
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type patternBudget391 struct{ Source, Store, ConfigWrites, CaptureWrites, SQL int }
type patternFact391 struct {
	Name          string
	Before, After patternBudget391
}
type patternRead391 struct {
	Name     string
	Request  exploration.Query
	Response exploration.Result
	Status   int
}
type patternDefinition391 struct {
	Name              string
	Request, Response workbench.ExplorationView
	Status            int
}
type patternLedger391 struct {
	Version string
	Fixture struct {
		Environment                         string
		SourceQuestion, DestinationQuestion exploration.Query
		Records                             []struct {
			Environment, Service, Team, Body, ID, TraceID, Namespace string
			TimeMS                                                   int64
		}
	}
	Raw []struct {
		ServiceName, SeverityText, Body, SpanID, TraceID, AtNS string
		ResourceAttributes, LogAttributes                      map[string]string
	}
	RawSpans json.RawMessage
	Reads    []patternRead391
	Facts    []patternFact391
	Captures []struct {
		ID, StoredBytesBefore, StoredBytesAfter string
		Response                                exploration.CapturePage
	}
	Definitions []patternDefinition391
	Snapshots   []struct {
		Name          string
		ScopeViewData exploration.ViewDataset
		Declaration   json.RawMessage
		VisibleRows   []string
	}
	Errors     []string
	Setup      struct{ CaptureWrites, ConfigWrites int }
	ProxyCalls []struct {
		SQL        string
		Parameters map[string]string
		Status     int
	}
	Calls []struct {
		Path, Method   string
		Body, Response json.RawMessage
		Status         int
	}
	NegativeControls struct {
		SuppliedOnly    bool
		ForgedRefusal   string
		UnknownSnapshot exploration.ViewDataset
		Text            string
	}
	OwnerControl struct {
		Before struct {
			URL            string
			Connected      bool
			Exists         bool
			HiddenAncestor bool
			Rect           struct{ X, Y, Width, Height, Top, Right, Bottom, Left float64 }
		}
		After      struct{ URL string }
		Dispatched bool
	}
	Actor struct {
		Snapshots int
		Phases    []string
		Complete  map[string]json.RawMessage
	}
}

func admitPattern391(r patternLedger391) error {
	fail := func(s string) error { return fmt.Errorf("pattern391: %s", s) }
	if r.Version != "goo.logPatternDisplay391/1" || len(r.Errors) != 0 || len(r.Raw) != 27 || len(r.Fixture.Records) != 27 || r.Setup.CaptureWrites != 1 || r.Setup.ConfigWrites != 1 {
		return fail("actual finite source/setup absent")
	}
	physical := make([]lpLiveRecord, 0, len(r.Raw))
	seen := map[string]bool{}
	for _, raw := range r.Raw {
		var fixtureFound bool
		for _, f := range r.Fixture.Records {
			if f.ID != raw.SpanID {
				continue
			}
			fixtureFound = true
			ns, e := strconv.ParseInt(raw.AtNS, 10, 64)
			if e != nil || ns != f.TimeMS*1000000 || raw.Body != f.Body || raw.ServiceName != f.Service || raw.TraceID != f.TraceID || raw.ResourceAttributes["deployment.environment.name"] != f.Environment || raw.ResourceAttributes["service.namespace"] != f.Namespace || raw.LogAttributes["team"] != f.Team || seen[f.ID] {
				return fail("physical row differs from actual fixture")
			}
			seen[f.ID] = true
			physical = append(physical, lpLiveRecord{Environment: f.Environment, Signal: "logs", ID: f.ID, TraceID: f.TraceID, Service: f.Service, Namespace: f.Namespace, Body: raw.Body, Severity: raw.SeverityText, TimeMS: f.TimeMS, Attributes: raw.LogAttributes, Resource: map[string]string{"service.namespace": f.Namespace}})
		}
		if !fixtureFound {
			return fail("foreign physical source row")
		}
	}
	if reflect.DeepEqual(r.Fixture.SourceQuestion, r.Fixture.DestinationQuestion) || r.Fixture.SourceQuestion.FromMS == r.Fixture.DestinationQuestion.FromMS {
		return fail("reuse destination did not change exact scope")
	}
	readNames := map[string]bool{}
	for _, read := range r.Reads {
		if read.Name == "" || readNames[read.Name] {
			return fail("duplicate actual read identity")
		}
		readNames[read.Name] = true
		if read.Name == "source-analysis" && !reflect.DeepEqual(read.Request, r.Fixture.SourceQuestion) || read.Name == "destination-run" && !reflect.DeepEqual(read.Request, r.Fixture.DestinationQuestion) {
			return fail("named source/destination owner changed")
		}
		sourceOK := lpLiveSource(read.Response, r.Fixture.Environment, physical)
		if read.Request.Signal == "spans" {
			sourceOK = string(r.RawSpans) == "[]" && read.Request.Environment == r.Fixture.Environment && read.Response.Count == "0" && len(read.Response.Rows) == 0 && len(read.Response.Groups) == 0 && read.Response.State == "no_data" && peLiveClock(read.Response.CapturedAt) && peLiveHex(read.Response.Basis)
		}
		if read.Status != 200 || !reflect.DeepEqual(read.Request, read.Response.Question) || !sourceOK {
			return fail("actual source read/question mismatch: " + read.Name)
		}
		if read.Response.ViewDataset != nil && read.Request.Signal == "logs" {
			d := read.Response.ViewDataset
			if d.Partition == nil || d.Partition.Total != read.Response.Count || d.Completeness == nil || d.Fields[0].Format != "log-pattern-v1" || d.Revision != read.Response.Basis || d.CapturedAt != read.Response.CapturedAt {
				return fail("source pattern dataset lacks exact full population/clocks")
			}
			groups := map[string]string{}
			for _, raw := range read.Response.Groups {
				var g struct{ G0, Count string }
				if json.Unmarshal(raw, &g) != nil || !lpCanonical(g.G0) {
					return fail("noncanonical source pattern group")
				}
				groups[g.G0] = g.Count
			}
			if len(d.Rows) != len(groups) {
				return fail("source dataset changed full partition")
			}
			seen := map[string]bool{}
			for _, row := range d.Rows {
				id, count := row.Values["category"], row.Values["count"]
				if id == nil || count == nil || seen[*id] || groups[*id] != *count {
					return fail("decoded dataset changed canonical group/count")
				}
				seen[*id] = true
			}
			var selector exploration.Query
			if json.Unmarshal(d.Scope.Selector, &selector) != nil || !reflect.DeepEqual(selector, read.Request) || d.Scope.Generation != read.Response.Basis || d.Completeness.Revision != d.Revision || d.Completeness.ScopeID != d.Scope.ID {
				return fail("dataset selector/witness belongs to another source clock")
			}
		}
	}
	if !readNames["source-analysis"] || !readNames["destination-run"] {
		return fail("explicit source/destination reads missing")
	}
	if len(r.Reads) < 3 || len(r.Captures) != 1 || len(r.Definitions) != 2 || len(r.Snapshots) != 4 {
		return fail("finite actual journey incomplete")
	}
	capture := r.Captures[0]
	if capture.StoredBytesBefore == "" || capture.StoredBytesBefore != capture.StoredBytesAfter {
		return fail("frozen capture bytes changed")
	}
	var stored exploration.SavedCapture
	if e := json.Unmarshal([]byte(capture.StoredBytesBefore), &stored); e != nil {
		return e
	}
	id := stored.ID
	stored.ID = ""
	wire, _ := json.Marshal(stored)
	digest := sha256.Sum256(wire)
	if id != capture.ID || hex.EncodeToString(digest[:]) != id {
		return fail("capture canonical identity changed")
	}
	stored.ID = id
	page, e := stored.Page("")
	captureQuestion := r.Fixture.SourceQuestion
	captureQuestion.Limit = exploration.CapturePageSize
	captureQuestion.Offset = 0
	if e != nil || !sameTransport384(page, capture.Response) || stored.Explore == nil || !lpLiveSource(*stored.Explore, r.Fixture.Environment, physical) || !reflect.DeepEqual(stored.Question.Explore, &captureQuestion) {
		return fail("captured source/custody projection changed")
	}
	for _, definition := range r.Definitions {
		if definition.Status != 201 || definition.Response.Validate() != nil || len(definition.Response.Queries) < 1 || len(definition.Response.Queries) > 2 || !reflect.DeepEqual(definition.Request.Queries, definition.Response.Queries) {
			return fail("saved definition changed query")
		}
		detached := definition.Response
		id := detached.ID
		detached.ID = ""
		hash, err := workbench.ViewContentID(detached)
		if err != nil || hash != id {
			return fail("definition content identity changed")
		}
		if definition.Name == "source-save" {
			d, ok := definition.Response.Displays["logs"]
			var donor *exploration.Query
			for _, q := range definition.Response.Queries {
				if q.Signal == "logs" {
					copy := q
					donor = &copy
				}
			}
			if !ok || donor == nil || d.Validate(*donor) != nil || d.Display.Profile != "log-pattern-count-v1" {
				return fail("saved profile missing")
			}
		}
	}
	for _, snapshot := range r.Snapshots {
		d := snapshot.ScopeViewData
		var q exploration.Query
		if json.Unmarshal(d.Scope.Selector, &q) != nil {
			return fail("snapshot lost exact selector")
		}
		matched := false
		for _, read := range r.Reads {
			if reflect.DeepEqual(q, read.Request) && sameTransport384(d, read.Response.ViewDataset) {
				matched = true
				break
			}
		}
		if !matched {
			return fail("local view changed full data/basis: " + snapshot.Name)
		}
	}
	if !r.NegativeControls.SuppliedOnly || !strings.Contains(r.NegativeControls.ForgedRefusal, "cannot attest a complete population") || !strings.Contains(strings.ToLower(r.NegativeControls.Text), "shares unavailable") || r.NegativeControls.UnknownSnapshot.Partition != nil || r.NegativeControls.UnknownSnapshot.Completeness != nil || r.NegativeControls.UnknownSnapshot.Quality.Extent != "unknown" {
		return fail("supplied diagnostic acquired authority")
	}
	if len(r.Facts) != len(patternActions391) {
		return fail("measured journey incomplete")
	}
	for i, f := range r.Facts {
		if f.Name != patternActions391[i] || i > 0 && f.Before != r.Facts[i-1].After {
			return fail("action order or counter continuity changed")
		}
		source, configs := 0, 0
		switch f.Name {
		case "source-analysis-run", "explicit-destination-run":
			source = 1
		case "source-population-run":
			source = 2
		case "save-source-configuration":
			configs = 1
		}
		if f.After.Source-f.Before.Source != source || f.After.ConfigWrites-f.Before.ConfigWrites != configs || f.After.CaptureWrites != 0 || f.Before.CaptureWrites != 0 || source == 0 && f.After.SQL != f.Before.SQL {
			return fail("local/read/publication boundary changed")
		}
		if f.Name != "destination-stage-inspect" && f.After.Store != f.Before.Store {
			return fail("unexpected store read")
		}
	}
	if !r.OwnerControl.Dispatched || !r.OwnerControl.Before.Exists || !r.OwnerControl.Before.Connected || !r.OwnerControl.Before.HiddenAncestor || r.OwnerControl.Before.Rect.Width != 0 || r.OwnerControl.Before.Rect.Height != 0 || r.OwnerControl.Before.URL == "" || r.OwnerControl.Before.URL != r.OwnerControl.After.URL {
		return fail("retired pattern owner navigated")
	}
	var actor struct {
		Version, Phase string
		Reads          int
		Snapshot       exploration.ViewDataset
		Display        struct{ Kind, Query string }
		Overflow       bool
	}
	actorWire, _ := json.Marshal(r.Actor.Complete)
	if json.Unmarshal(actorWire, &actor) != nil || r.Actor.Snapshots <= 0 || actor.Version != "goo.logPatternDisplay391Actor/1" || actor.Phase != "complete" || actor.Display.Kind != "table" || actor.Display.Query != "queue" || actor.Overflow || actor.Snapshot.Partition == nil || actor.Snapshot.Partition.Total != "12" {
		return fail("actual actor completion/held denominator missing")
	}
	if err := admitActorSQL391(r); err != nil {
		return err
	}
	var actorQuestion exploration.Query
	if json.Unmarshal(actor.Snapshot.Scope.Selector, &actorQuestion) != nil {
		return fail("actor lost fixed source selector")
	}
	rows, err := lpLiveRows(actorQuestion, r.Fixture.Environment, physical)
	if err != nil || len(rows) != 12 {
		return fail("actor source scope differs")
	}
	counts := map[string]int{}
	for _, row := range rows {
		counts[lpOracle(row.Body).Identity]++
	}
	if len(actor.Snapshot.Rows) != len(counts) {
		return fail("actor lost canonical population")
	}
	for _, row := range actor.Snapshot.Rows {
		if row.Values["category"] == nil || row.Values["count"] == nil || *row.Values["count"] != strconv.Itoa(counts[*row.Values["category"]]) {
			return fail("actor changed pattern counts")
		}
	}
	return nil
}

func TestQuintLogPatternDisplayObserved391(t *testing.T) {
	path := os.Getenv("GOO_LOG_PATTERN_DISPLAY_391_REPORT")
	if path == "" {
		t.Skip("actual log-pattern display journey required")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var r patternLedger391
	if e = json.Unmarshal(raw, &r); e != nil {
		t.Fatal(e)
	}
	if e = admitPattern391(r); e != nil {
		t.Fatal(e)
	}
	if !replayPattern391(t, r, false) {
		t.Fatal("actual budget violates model")
	}
	mutations := []func(*patternLedger391){
		func(r *patternLedger391) { r.Raw[0].Body += " changed" },
		func(r *patternLedger391) { r.Captures[0].StoredBytesAfter += " " },
		func(r *patternLedger391) { r.Snapshots[0].ScopeViewData.Partition.Total = "1" },
		func(r *patternLedger391) { r.Fixture.DestinationQuestion.Environment = "donor" },
		func(r *patternLedger391) { r.OwnerControl.After.URL += "#foreign" },
		func(r *patternLedger391) {
			r.NegativeControls.UnknownSnapshot.Partition = r.Snapshots[0].ScopeViewData.Partition
		},
	}
	for i, mutate := range mutations {
		var bad patternLedger391
		json.Unmarshal(raw, &bad)
		mutate(&bad)
		if admitPattern391(bad) == nil {
			t.Fatalf("semantic mutant %d escaped", i)
		}
	}
	if replayPattern391(t, r, true) {
		t.Fatal("local read model control escaped")
	}
	t.Logf("admitted %d physical logs, %d reads, %d facts; six semantic mutants and local-read control rejected", len(r.Raw), len(r.Reads), len(r.Facts))
}

var patternActions391 = []string{"source-stage", "source-analysis-run", "source-population-run", "retired-pattern-owner", "local-find-order-kind", "save-source-configuration", "destination-stage-inspect", "apply-without-evidence", "explicit-destination-run", "destination-local-reset", "supplied-negative-authority", "complete"}

func replayPattern391(t *testing.T, r patternLedger391, mutate bool) bool {
	t.Helper()
	if len(r.Facts) != len(patternActions391) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	inspectStores := 0
	for i, f := range r.Facts {
		if f.Name != patternActions391[i] {
			return false
		}
		if i > 0 && f.Before != r.Facts[i-1].After {
			return false
		}
		action := "local"
		switch f.Name {
		case "source-analysis-run", "explicit-destination-run":
			action = "read"
		case "source-population-run":
			action = "readPair"
		case "save-source-configuration":
			action = "save"
		case "destination-stage-inspect":
			action = "inspect"
			inspectStores = f.After.Store - f.Before.Store
		case "complete":
			action = "finish"
		}
		reads := f.After.Source
		if mutate && f.Name == "local-find-order-kind" {
			reads++
		}
		rec.Record(context.Background(), action, nil, "obs.phase", i+1, "obs.reads", reads, "obs.stores", f.After.Store, "obs.configs", f.After.ConfigWrites, "obs.captures", f.After.CaptureWrites, "obs.finished", f.Name == "complete")
	}
	if inspectStores < 1 || inspectStores > 3 {
		return false
	}
	spec, e := filepath.Abs("model/log_pattern_display391.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: log_pattern_display391\nconstants: {INSPECT_STORES: '%d'}\ninvariants: [safety]\nactions:\n", spec, inspectStores)
	for _, a := range []string{"local", "read", "readPair", "save", "inspect", "finish"} {
		binding += "  " + a + ": {expect: 'phase == {{.obs.phase}} and reads == {{.obs.reads}} and stores == {{.obs.stores}} and configs == {{.obs.configs}} and captures == {{.obs.captures}} and finished == {{.obs.finished}}'}\n"
	}
	dir := t.TempDir()
	file := filepath.Join(dir, "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	c, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: dir})
	if e != nil {
		t.Fatal(e)
	}
	result, e := c.Check(t.Context(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	label := "actual"
	if mutate {
		label = "local-read-mutant"
	}
	steps, _ := json.Marshal(sink.Steps())
	if e = artifact("log-pattern-display391-"+label+"-steps.json", steps); e != nil {
		t.Fatal(e)
	}
	if e = artifact("log-pattern-display391-"+label+"-model.txt", []byte(result.Summary())); e != nil {
		t.Fatal(e)
	}
	return result.OK()
}

// Resource-timing entries were unavailable in the actor browser. Admit its
// actual physical HTTP source tail instead, independently of that counter.
func admitActorSQL391(r patternLedger391) error {
	if len(r.Facts) == 0 {
		return fmt.Errorf("missing main source boundary")
	}
	start := r.Facts[len(r.Facts)-1].After.SQL
	if start < 0 || start >= len(r.ProxyCalls) {
		return fmt.Errorf("missing actual actor SQL tail")
	}
	expected := map[string]int{}
	key := func(sql string, params map[string]string) string {
		wire, _ := json.Marshal(params)
		return sql + "\x00" + string(wire)
	}
	for _, q := range r.Definitions[0].Response.Queries {
		plan, e := exploration.Compile(q)
		if e != nil {
			return e
		}
		for _, sql := range []string{plan.Count, plan.Rows, plan.Groups, plan.Fields, plan.TimeSeries, plan.GroupedTimeSeries} {
			if sql != "" {
				expected[key(sql, plan.Parameters)]++
			}
		}
	}
	for _, call := range r.ProxyCalls[start:] {
		if call.Status != 200 || call.Parameters["readonly"] != "1" {
			return fmt.Errorf("actor source response or readonly guard changed")
		}
		params := map[string]string{}
		for name, value := range call.Parameters {
			if strings.HasPrefix(name, "param_") {
				params[strings.TrimPrefix(name, "param_")] = value
			}
		}
		k := key(call.SQL, params)
		if expected[k] <= 0 {
			return fmt.Errorf("actor SQL/typed parameters differ from saved fixed questions")
		}
		expected[k]--
	}
	for _, remaining := range expected {
		if remaining != 0 {
			return fmt.Errorf("actor source tail incomplete")
		}
	}
	return nil
}
