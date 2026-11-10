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
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type retainedCounters401 struct{ Source, SQL, Metrics, Records, Stores, Writes int }
type retainedFact401 struct {
	Name          string
	Before, After retainedCounters401
	Setup         bool
}
type retainedAttempt401 struct {
	Request                   exploration.Query
	Response                  json.RawMessage
	Status                    int
	AbortsBefore, AbortsAfter int
}
type retainedCase401 struct {
	Key, Surface, Signal, Policy, AtNS  string
	Namespace                           *string
	Point                               int
	Capture                             exploration.CapturePage
	StoredBytesBefore, StoredBytesAfter string
	Measurement                         measurements.Result
	SelectedStream                      metrics.Stream
	Range                               struct{ FromNS, ToNS, AtNS, RadiusNS string }
	ExpectedQuestion                    exploration.Query
	ContextPlan                         struct {
		Question           exploration.Query
		AtNS, FromNS, ToNS string
	}
	RequestedWidget           struct{ Measurement *measurements.Question }
	Plan                      exploration.DashboardPlan
	ParentBefore, ParentAfter json.RawMessage
	Citation                  struct {
		Measurement             measurements.Question
		Stream                  metrics.Stream
		Basis, ReadAt, Relation string
		Observation             struct{ AtNS, RadiusNS string }
	}
	Oracle                []observationRaw388
	Attempts              []retainedAttempt401
	Pending               *retainedAttempt401
	OldProbe, HiddenProbe struct{ Hash, After string }
	Facts                 []retainedFact401
}
type retainedLedger401 struct {
	Version  string
	Errors   []string
	SetupSQL int
	Fixture  struct {
		Service, Instance string
		Records           []json.RawMessage
	}
	Raw struct {
		Spans, Logs []observationRaw388
		Metrics     []gaugeRaw386
	}
	Cases []retainedCase401
	Facts []retainedFact401
	Calls []struct {
		Path, Method   string
		Body, Response json.RawMessage
		Status         int
	}
	ProxyCalls []struct {
		SQL        string
		Parameters map[string]string
		Status     int
	}
}

func admitRetained401(r retainedLedger401) error {
	fail := func(s string) error { return fmt.Errorf("retained401: %s", s) }
	if r.Version != "goo.retainedGauge401/1" || len(r.Errors) != 0 || len(r.Cases) != 6 || len(r.Facts) != 39 || len(r.Raw.Metrics) != 16 || len(r.Raw.Spans) != len(r.Fixture.Records) || len(r.Raw.Logs) != len(r.Fixture.Records) {
		return fail("incomplete actual source/gesture bank")
	}
	seen := map[string]bool{}
	setupCosts := map[string]retainedCounters401{}
	childSQL := map[string]int{}
	expectedFacts := []retainedFact401{}
	for _, c := range r.Cases {
		if seen[c.Key] || !slices.Contains([]string{"service", "saved"}, c.Surface) {
			return fail("duplicate or foreign retained owner")
		}
		seen[c.Key] = true
		suffix := "named"
		point := 0
		signal := "spans"
		policy := "current-only"
		if c.Namespace == nil {
			suffix = "missing"
			point = 3
			signal = "logs"
			policy = "effective"
		} else if *c.Namespace == "" {
			suffix = "empty"
			point = 1
			policy = "effective"
		}
		if c.Key != c.Surface+"-"+suffix || c.Point != point || c.Signal != signal || c.Policy != policy {
			return fail("required namespace/zero/unknown matrix changed")
		}
		m := c.Measurement
		q := m.Question
		selection, selectionErr := q.Definition.EnvironmentSelection()
		if measurements.ValidateResult(&m) != nil || m.Gauges == nil || m.Basis == "" || m.ReadAt == "" || selectionErr != nil || selection.CurrentOnly() != (policy == "current-only") {
			return fail("invalid retained native result/policy")
		}
		if !peLiveCaptureHash(c.Capture) || c.StoredBytesBefore != c.StoredBytesAfter {
			return fail("immutable actual receipt custody changed")
		}
		var stored exploration.SavedCapture
		if json.Unmarshal([]byte(c.StoredBytesBefore), &stored) != nil {
			return fail("stored capture absent")
		}
		page, e := stored.Page("")
		if e != nil || !same401(page, &c.Capture) {
			return fail("stored immutable bytes differ from visited page")
		}
		if c.Surface == "saved" {
			if c.Capture.Measurement == nil || !same401(c.Capture.Measurement, &m) || c.Capture.Receipt.Question.Kind != "measurement" {
				return fail("saved result/question detached from receipt")
			}
		} else if c.Capture.Inventory == nil || c.RequestedWidget.Measurement == nil || !reflect.DeepEqual(*c.RequestedWidget.Measurement, q) {
			return fail("service result differs from actual requested widget")
		}
		selected := 0
		for _, series := range m.Gauges.Series {
			scoped := q
			scoped.Definition.Source.ExactStream = &series.Stream
			want := gaugeReference386(scoped, r.Raw.Metrics)
			if len(want) != 4 || len(series.Samples) != 4 {
				return fail("physical sample membership changed")
			}
			for i, p := range series.Samples {
				var value *string
				if p.Value != nil {
					x := strconv.FormatFloat(*p.Value, 'g', -1, 64)
					value = &x
				}
				if p.AtNS != want[i].AtNS || p.State != want[i].State || p.Reason != want[i].Reason || !reflect.DeepEqual(value, gaugeValue386(want[i].Value)) {
					return fail("physical zero/unknown/timestamp/freshness changed")
				}
			}
			if reflect.DeepEqual(series.Stream, c.SelectedStream) {
				selected++
				p := series.Samples[c.Point]
				if p.AtNS != c.AtNS || point == 1 && (p.Value == nil || *p.Value != 0) || point == 3 && (p.Value != nil || p.State != "unknown") {
					return fail("selected timestamp/zero/unknown changed")
				}
			}
		}
		if selected != 1 {
			return fail("selected stream not unique")
		}
		ns, present := c.SelectedStream.Resource[identityNamespace388]
		if c.Namespace == nil && present || c.Namespace != nil && (!present || ns != *c.Namespace) {
			return fail("missing/empty/named namespace collapsed")
		}
		loText, hiText := q.BoundsNS()
		lo, ok := new(big.Int).SetString(loText, 10)
		hi, ok2 := new(big.Int).SetString(hiText, 10)
		at, ok3 := new(big.Int).SetString(c.AtNS, 10)
		if !ok || !ok2 || !ok3 {
			return fail("invalid exact timestamps")
		}
		from := new(big.Int).Sub(new(big.Int).Set(at), big.NewInt(60000000000))
		to := new(big.Int).Add(new(big.Int).Set(at), big.NewInt(60000000000))
		if from.Cmp(lo) < 0 {
			from.Set(lo)
		}
		if to.Cmp(hi) > 0 {
			to.Set(hi)
		}
		p, e := exploration.PlanNativeObservationContext(&m, c.SelectedStream, c.Signal, c.AtNS)
		if e != nil || !reflect.DeepEqual(p.Question, c.ExpectedQuestion) || !reflect.DeepEqual(c.ContextPlan.Question, p.Question) || c.Range.FromNS != from.String() || c.Range.ToNS != to.String() || c.Range.AtNS != c.AtNS || c.Range.RadiusNS != "60000000000" || c.ContextPlan.FromNS != from.String() || c.ContextPlan.ToNS != to.String() || c.ContextPlan.AtNS != c.AtNS {
			return fail("clipped exact observation scope changed")
		}
		if !reflect.DeepEqual(c.Citation.Measurement, q) || !reflect.DeepEqual(c.Citation.Stream, c.SelectedStream) || c.Citation.Basis != m.Basis || c.Citation.ReadAt != m.ReadAt || c.Citation.Relation != "observation-context" || c.Citation.Observation.AtNS != c.AtNS || c.Citation.Observation.RadiusNS != "60000000000" || !sameTransport384(c.ParentBefore, c.ParentAfter) || c.OldProbe.Hash != c.OldProbe.After || c.HiddenProbe.Hash != c.HiddenProbe.After {
			return fail("retained parent/citation/retired callback changed")
		}
		rows := r.Raw.Spans
		if c.Signal == "logs" {
			rows = r.Raw.Logs
		}
		oracleQuestion := q
		if oracleQuestion.Definition.EnvironmentSelector == nil {
			selector, err := selection.Selector()
			if err != nil {
				return fail("invalid private oracle environment policy")
			}
			oracleQuestion.Definition.EnvironmentSelector = &selector
		}
		ids := observationMembers388(rows, oracleQuestion, r.Fixture.Service, r.Fixture.Instance, c.Namespace, from.String(), to.String())
		got := []string{}
		for _, row := range c.Oracle {
			got = append(got, row.SpanID)
		}
		slices.Sort(got)
		if len(ids) < 2 || !slices.Equal(ids, got) {
			return fail("raw record oracle changed")
		}
		attempts := append([]retainedAttempt401{}, c.Attempts...)
		if len(attempts) != 1 {
			return fail("explicit normal run missing")
		}
		if c.Key == "saved-missing" {
			if c.Pending == nil || c.Pending.AbortsAfter <= c.Pending.AbortsBefore {
				return fail("pending child was not aborted on retained return")
			}
			attempts = append(attempts, *c.Pending)
		} else if c.Pending != nil {
			return fail("unexpected pending read")
		}
		for _, a := range attempts {
			var child exploration.Result
			if a.Status != 200 || !reflect.DeepEqual(a.Request, p.Question) || json.Unmarshal(a.Response, &child) != nil || !reflect.DeepEqual(child.Question, p.Question) || child.Count != strconv.Itoa(len(ids)) {
				return fail("actual explicit child request/result changed")
			}
			actual, e := reportedResultIDs353(child)
			if e != nil || !slices.Equal(actual, ids) {
				return fail("foreign/outside/boundary records admitted")
			}
		}
		phases := []string{"open-select", "hidden-inert", "stage", "run", "return", "retired-stage"}
		if c.Key == "saved-missing" {
			phases = append(phases, "stage", "pending-run", "return-pending")
		}
		if len(c.Facts) != len(phases) {
			return fail("gesture sequence incomplete")
		}
		for i, phase := range phases {
			if c.Facts[i].Name != c.Key+"-"+phase {
				return fail("gesture sequence changed")
			}
		}
		var parent struct {
			Selected          string
			Open              bool
			DomID, TimelineID int
		}
		if json.Unmarshal(c.ParentBefore, &parent) != nil || parent.Selected != strconv.Itoa(point) || !parent.Open || parent.DomID <= 0 || parent.TimelineID <= 0 {
			return fail("retained selection/DOM absent")
		}
		sourceMatch, visitMatch := false, false
		for _, call := range r.Calls {
			if call.Method == "GET" && call.Path == "/api/captures/"+c.Capture.Receipt.ID && call.Status == 200 && same401(call.Response, c.Capture) {
				visitMatch = true
			}
			if c.Surface == "service" && call.Path == "/api/measurements" && call.Method == "POST" && call.Status == 200 && same401(call.Body, q) && same401(call.Response, m) {
				sourceMatch = true
			}
		}
		if !visitMatch || c.Surface == "service" && !sourceMatch {
			return fail("visited receipt/successful native request missing")
		}
		if c.Surface == "service" {
			nativeWidgets, recordWidgets := 0, 0
			for _, w := range c.Plan.Widgets {
				if w.State != "supported" {
					continue
				}
				path := "/api/record-aggregates"
				var question any
				if w.Measurement != nil {
					nativeWidgets++
					path = "/api/measurements"
					question = w.Measurement
				} else if w.Query != nil {
					recordWidgets++
					question = w.Query
				} else {
					return fail("supported widget has no question")
				}
				found := false
				for _, call := range r.Calls {
					if call.Method == "POST" && call.Path == path && call.Status == 200 && same401(call.Body, question) {
						found = true
					}
				}
				if !found {
					return fail("declared widget read absent")
				}
			}
			setupCosts[c.Key+"-open-select"] = retainedCounters401{Source: 5 + nativeWidgets + recordWidgets, Metrics: nativeWidgets, Records: 1, Stores: 1}
		}
		for _, a := range attempts {
			compiled, err := exploration.Compile(a.Request)
			if err != nil {
				return fail("child compile failed")
			}
			for _, sql := range []string{compiled.Count, compiled.Rows, compiled.Fields} {
				if sql == "" {
					return fail("expected child statement absent")
				}
				childSQL[retainedStatement401(sql, compiled.Parameters, false)]++
			}
		}
		expectedFacts = append(expectedFacts, c.Facts...)
	}
	if !same401(expectedFacts, r.Facts) {
		return fail("case and global event order differ")
	}
	for i, f := range r.Facts {
		if i > 0 && f.Before != r.Facts[i-1].After {
			return fail("counter discontinuity")
		}
		d := retainedDelta401(f)
		open := strings.HasSuffix(f.Name, "-open-select")
		run := strings.HasSuffix(f.Name, "-run")
		service := strings.HasPrefix(f.Name, "service-")
		if d.Source < 0 || d.SQL < 0 || d.Metrics < 0 || d.Records < 0 || d.Stores < 0 || d.Writes != 0 || f.Before.Writes != 0 {
			return fail("negative counter or config/capture write")
		}
		if open {
			if d.Stores != 1 || f.Setup != service {
				return fail("visit budget changed")
			}
			if service {
				if d.Source != setupCosts[f.Name].Source || d.Metrics != setupCosts[f.Name].Metrics || d.Records != setupCosts[f.Name].Records || d.SQL < 1 {
					return fail("service setup source budget absent")
				}
			} else if d.Source != 0 || d.SQL != 0 || d.Metrics != 0 || d.Records != 0 {
				return fail("saved visit reread native source")
			}
		} else if run {
			if f.Setup || d.Source != 1 || d.Records != 1 || d.Metrics != 0 || d.Stores != 0 || d.SQL != 3 {
				return fail("explicit record read budget changed")
			}
		} else if f.Setup || d != (retainedCounters401{}) {
			return fail("local stage/return performed IO")
		}
	}
	last := r.Facts[len(r.Facts)-1].After
	if last.Stores != 6 || last.Records != 10 || last.SQL != len(r.ProxyCalls) || r.Facts[0].Before.SQL != r.SetupSQL {
		return fail("final physical/store accounting changed")
	}
	source, stores, records, metricsCount := 0, 0, 0, 0
	for _, call := range r.Calls {
		if call.Method == "GET" && strings.HasPrefix(call.Path, "/api/captures/") && len(call.Path) == len("/api/captures/")+64 {
			stores++
		}
		if call.Method != "POST" {
			continue
		}
		switch call.Path {
		case "/api/captures", "/api/documents", "/api/native-views":
			return fail("unexpected write")
		case "/api/measurements":
			source++
			metricsCount++
		case "/api/explore":
			source++
			records++
		case "/api/discovery", "/api/inventory", "/api/metric-discovery", "/api/record-aggregates":
			source++
		case "/api/explore/page", "/api/tail":
			return fail("unowned source family")
		}
	}
	if source != last.Source || stores != last.Stores || records != last.Records || metricsCount != last.Metrics {
		return fail("actual requests differ from observed budgets")
	}
	physical := map[string]int{}
	for _, call := range r.ProxyCalls {
		if call.Status == 200 {
			physical[retainedStatement401(call.SQL, call.Parameters, true)]++
		}
	}
	for sql, count := range childSQL {
		if physical[sql] < count {
			return fail("compiled child count/rows/fields physical statement absent")
		}
	}
	return nil
}

// Only compiler bind values form statement identity. Source safety settings are
// separately supplied transport controls and cannot substitute for param_*.
func retainedStatement401(sql string, parameters map[string]string, transport bool) string {
	checked := map[string]string{}
	for key, value := range parameters {
		if transport {
			if !strings.HasPrefix(key, "param_") {
				continue
			}
			key = strings.TrimPrefix(key, "param_")
		}
		checked[key] = value
	}
	raw, _ := json.Marshal(struct {
		SQL        string
		Parameters map[string]string
	}{sql, checked})
	return string(raw)
}
func same401(a, b any) bool {
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	return sameTransport384(x, y)
}
func retainedDelta401(f retainedFact401) retainedCounters401 {
	return retainedCounters401{f.After.Source - f.Before.Source, f.After.SQL - f.Before.SQL, f.After.Metrics - f.Before.Metrics, f.After.Records - f.Before.Records, f.After.Stores - f.Before.Stores, f.After.Writes - f.Before.Writes}
}

func TestQuintRetainedGaugeObserved401(t *testing.T) {
	path := os.Getenv("GOO_RETAINED_GAUGE_401_REPORT")
	if path == "" {
		t.Skip("actual retained gauge journey required")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var r retainedLedger401
	if json.Unmarshal(raw, &r) != nil {
		t.Fatal("invalid retained gauge report")
	}
	if e = admitRetained401(r); e != nil {
		t.Fatal(e)
	}
	mutants := []func(*retainedLedger401){func(r *retainedLedger401) { r.Cases[0].Range.FromNS = "0" }, func(r *retainedLedger401) { r.Cases[0].AtNS = "1" }, func(r *retainedLedger401) { r.Cases[0].Capture.Receipt.ID = strings.Repeat("a", 64) }, func(r *retainedLedger401) { r.Cases[1].Citation.Basis = "foreign" }, func(r *retainedLedger401) { r.Cases[2].ParentAfter = json.RawMessage(`{"changed":true}`) }, func(r *retainedLedger401) { r.Cases[3].Attempts[0].Request.ToMS++ }, func(r *retainedLedger401) { r.Cases[4].SelectedStream.Resource[identityNamespace388] = "" }, func(r *retainedLedger401) { r.Cases[5].StoredBytesAfter = "{}" }, func(r *retainedLedger401) { r.Facts[2].After.Source++ }, func(r *retainedLedger401) {
		last := len(r.ProxyCalls) - 1
		for key := range r.ProxyCalls[last].Parameters {
			if strings.HasPrefix(key, "param_") {
				r.ProxyCalls[last].Parameters[key] = "foreign-bind-value"
				return
			}
		}
	}}
	for i, mutate := range mutants {
		var bad retainedLedger401
		json.Unmarshal(raw, &bad)
		mutate(&bad)
		if admitRetained401(bad) == nil {
			t.Fatalf("semantic mutant %d admitted", i)
		}
	}
	if !replayRetained401(t, r, false) || replayRetained401(t, r, true) {
		t.Fatal("observed model or staging-read negative control incorrect")
	}
}
func replayRetained401(t *testing.T, r retainedLedger401, mutate bool) bool {
	t.Helper()
	reads, stores := []int{}, []int{}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, f := range r.Facts {
		d := retainedDelta401(f)
		read, store := 0, 0
		if strings.HasSuffix(f.Name, "-open-select") {
			store = 1
			if f.Setup {
				read = d.Source
			}
		} else if strings.HasSuffix(f.Name, "-run") {
			read = 1
		}
		reads = append(reads, read)
		stores = append(stores, store)
		source := f.After.Source
		if mutate && f.Name == "saved-named-stage" {
			source++
		}
		rec.Record(context.Background(), "step", nil, "obs.phase", i+1, "obs.reads", source, "obs.stores", f.After.Stores, "obs.writes", f.After.Writes)
	}
	list := func(v []int) string { b, _ := json.Marshal(v); return string(b) }
	spec, e := filepath.Abs("model/retained_gauge401.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: retained_gauge401\nconstants: {READS: '%s', STORES: '%s'}\ninvariants: [safety]\nactions:\n  step: {expect: 'phase == {{.obs.phase}} and reads == {{.obs.reads}} and stores == {{.obs.stores}} and writes == {{.obs.writes}}'}\n", spec, list(reads), list(stores))
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

// The actor is a separate real Saved capture and driver campaign. Its final
// observation is checked against the main run's physical records, not its own
// expected IDs, and is deliberately not presented as an independently ordered
// cross-host trace.
func TestBombadilRetainedGaugeObserved401(t *testing.T) {
	mainPath, actorPath := os.Getenv("GOO_RETAINED_GAUGE_401_REPORT"), os.Getenv("GOO_RETAINED_GAUGE_401_ACTOR")
	if mainPath == "" || actorPath == "" {
		t.Skip("actual retained gauge main and actor reports required")
	}
	b, e := os.ReadFile(mainPath)
	if e != nil {
		t.Fatal(e)
	}
	var main retainedLedger401
	if json.Unmarshal(b, &main) != nil {
		t.Fatal("invalid main report")
	}
	if e = admitRetained401(main); e != nil {
		t.Fatal(e)
	}
	b, e = os.ReadFile(actorPath)
	if e != nil {
		t.Fatal(e)
	}
	var a struct {
		Version             string
		Question            measurements.Question
		ExpectedQuestion    exploration.Query
		Capture             exploration.CapturePage
		SQLBefore, SQLAfter int
		Setup               struct {
			Status   int
			Response exploration.CapturePage
			Request  struct{ Question exploration.CaptureQuestion }
		}
		Actor struct {
			Snapshots int
			Phases    []string
			Complete  struct {
				Version, Phase, Receipt string
				Loaded, Back, Overflow  bool
				Timeline                struct {
					Data     json.RawMessage
					Selected string
				}
				Context json.RawMessage
				Held    struct {
					Timeline          json.RawMessage
					Selected, Receipt string
					Context           json.RawMessage
				}
				Calls []struct {
					Path, Method string
					Request      exploration.Query
					Response     json.RawMessage
					Status       int
				}
			}
		}
	}
	if json.Unmarshal(b, &a) != nil || a.Version != "goo.retainedGauge401ActorReport/1" || a.Setup.Status != 201 || !same401(a.Setup.Response, a.Capture) || !peLiveCaptureHash(a.Capture) || a.Capture.Measurement == nil || !reflect.DeepEqual(a.Question, a.Capture.Measurement.Question) || a.Setup.Request.Question.Measurement == nil || !reflect.DeepEqual(*a.Setup.Request.Question.Measurement, a.Question) {
		t.Fatal("actor source receipt/setup changed")
	}
	var c *retainedCase401
	for i := range main.Cases {
		if main.Cases[i].Key == "saved-missing" {
			c = &main.Cases[i]
		}
	}
	if c == nil || !reflect.DeepEqual(a.Question, c.Measurement.Question) || !reflect.DeepEqual(a.ExpectedQuestion, c.ExpectedQuestion) {
		t.Fatal("actor scope differs from independently checked bank")
	}
	m := a.Capture.Measurement
	if measurements.ValidateResult(m) != nil || m.Gauges == nil || len(m.Gauges.Series) != 1 {
		t.Fatal("actor native result invalid")
	}
	p, e := exploration.PlanNativeObservationContext(m, m.Gauges.Series[0].Stream, "logs", c.AtNS)
	if e != nil || !reflect.DeepEqual(p.Question, c.ExpectedQuestion) {
		t.Fatal("actor observation projection changed")
	}
	s := a.Actor.Complete
	if s.Version != "goo.retainedGauge401Actor/1" || s.Phase != "complete" || !s.Loaded || s.Back || s.Overflow || s.Timeline.Selected != "3" || s.Held.Selected != "3" || s.Receipt != s.Held.Receipt || !sameTransport384(s.Timeline.Data, s.Held.Timeline) || !sameTransport384(s.Context, s.Held.Context) || a.SQLAfter-a.SQLBefore != 3 {
		t.Fatal("actor retained return or source budget changed")
	}
	for _, phase := range []string{"stream", "selected", "staged", "read", "back", "complete"} {
		if !slices.Contains(a.Actor.Phases, phase) {
			t.Fatalf("actor missing %s", phase)
		}
	}
	reads := 0
	for _, call := range s.Calls {
		if call.Method == "GET" && strings.HasPrefix(call.Path, "/api/captures/") {
			t.Fatal("actor reread receipt")
		}
		if call.Method != "POST" {
			continue
		}
		if call.Path != "/api/explore" {
			t.Fatal("actor unowned source/write")
		}
		reads++
		var child exploration.Result
		if call.Status != 200 || !reflect.DeepEqual(call.Request, c.ExpectedQuestion) || json.Unmarshal(call.Response, &child) != nil || !reflect.DeepEqual(child.Question, c.ExpectedQuestion) {
			t.Fatal("actor child changed")
		}
		ids := observationMembers388(main.Raw.Logs, a.Question, main.Fixture.Service, main.Fixture.Instance, c.Namespace, c.Range.FromNS, c.Range.ToNS)
		got, e := reportedResultIDs353(child)
		if e != nil || !slices.Equal(got, ids) || child.Count != strconv.Itoa(len(ids)) {
			t.Fatal("actor foreign/outside/absent records")
		}
	}
	if reads != 1 {
		t.Fatal("actor must explicitly run exactly once")
	}
}
