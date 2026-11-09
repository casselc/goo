package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
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

type observationRaw388 struct {
	ServiceName, TraceID, SpanID, AtNS string
	ResourceAttributes, Attributes     map[string]string
}
type observationCounters388 struct{ Source, SQL, Metrics, Records, Store, Writes int }
type observationFact388 struct {
	Name          string
	Before, After observationCounters388
}
type observationCase388 struct {
	Key, Signal, AtNS string
	Namespace         *string
	Point, PointIndex int
	SelectedPoint     struct {
		Key, FromNS, ToNS, State string
		Value                    *string
	}
	Range                             struct{ FromNS, ToNS, AtNS, RadiusNS string }
	ExpectedQuestion                  exploration.Query
	Oracle, PointMembers, FullMembers []observationRaw388
	ParentBefore, ParentAfter         json.RawMessage
	Citation                          struct {
		Question      measurements.Question
		Stream        metrics.Stream
		Basis, ReadAt string
		Observation   struct{ AtNS, RadiusNS string }
	}
	Attempts []struct {
		Request  exploration.Query
		Response json.RawMessage
		Status   int
	}
}
type observationLedger388 struct {
	Version string
	Errors  []string
	Fixture struct {
		Service, Instance, Environment, Named string
		Question                              measurements.Question
		Records                               []struct{ ID string }
	}
	Raw struct {
		Spans, Logs []observationRaw388
		Metrics     []gaugeRaw386
	}
	Native struct {
		Request  measurements.Question
		Response measurements.Result
		Status   int
	}
	Cases []observationCase388
	Calls []struct {
		Path, Method string
		Body         json.RawMessage
		Status       int
	}
	Facts      []observationFact388
	ProxyCalls []struct {
		SQL    string
		Status int
	}
}

func observationMembers388(rows []observationRaw388, q measurements.Question, service, instance string, namespace *string, from, to string) []string {
	lo, _ := new(big.Int).SetString(from, 10)
	hi, _ := new(big.Int).SetString(to, 10)
	ids := []string{}
	for _, r := range rows {
		if r.ServiceName != service || r.ResourceAttributes["service.instance.id"] != instance || !observationEnv388(q.Definition.EnvironmentSelector, r.ResourceAttributes) {
			continue
		}
		ns, present := r.ResourceAttributes[identityNamespace388]
		if namespace == nil && present || namespace != nil && (!present || ns != *namespace) {
			continue
		}
		at, ok := new(big.Int).SetString(r.AtNS, 10)
		if ok && at.Cmp(lo) >= 0 && at.Cmp(hi) < 0 {
			ids = append(ids, r.SpanID)
		}
	}
	slices.Sort(ids)
	return ids
}

const identityNamespace388 = "service.namespace"

func admitObservationLedger388(r observationLedger388) error {
	fail := func(s string) error { return fmt.Errorf("observation388: %s", s) }
	if r.Version != "goo.gaugeObservation388/1" || len(r.Errors) > 0 || r.Native.Status != 200 || len(r.Cases) != 3 {
		return fail("complete ordinary measurement journey absent")
	}
	result := r.Native.Response
	q := result.Question
	if !reflect.DeepEqual(q, r.Fixture.Question) || !reflect.DeepEqual(r.Native.Request, q) || measurements.ValidateResult(&result) != nil || result.Gauges == nil || len(result.Gauges.Series) != 3 || result.Count != "12" || len(r.Raw.Metrics) != 16 {
		return fail("native source question/replay changed")
	}
	if len(r.Raw.Spans) != len(r.Fixture.Records) || len(r.Raw.Logs) != len(r.Fixture.Records) {
		return fail("complete source record controls absent")
	}
	// Each evaluated stream is independently checked against physical gauge rows.
	for _, series := range result.Gauges.Series {
		scoped := q
		scoped.Definition.Source.ExactStream = &series.Stream
		expected := gaugeReference386(scoped, r.Raw.Metrics)
		if len(expected) != 4 || len(series.Samples) != 4 {
			return fail("native stream membership changed")
		}
		for i, want := range expected {
			sample := series.Samples[i]
			var value *string
			if sample.Value != nil {
				v := strconv.FormatFloat(*sample.Value, 'g', -1, 64)
				value = &v
			}
			if sample.AtNS != want.AtNS || sample.State != want.State || sample.Reason != want.Reason || !reflect.DeepEqual(value, gaugeValue386(want.Value)) {
				return fail("physical zero/unknown/freshness mismatch")
			}
		}
	}
	loText, hiText := q.BoundsNS()
	lo, _ := new(big.Int).SetString(loText, 10)
	hi, _ := new(big.Int).SetString(hiText, 10)
	radius := big.NewInt(observationRadius388)
	caseNames := map[string]bool{}
	readAttempts := 0
	for _, c := range r.Cases {
		if caseNames[c.Key] {
			return fail("duplicate context case")
		}
		caseNames[c.Key] = true
		if !sameTransport384(c.ParentBefore, c.ParentAfter) || !reflect.DeepEqual(c.Citation.Question, q) || c.Citation.Basis != result.Basis || c.Citation.ReadAt != result.ReadAt || c.Citation.Observation.AtNS != c.AtNS || c.Citation.Observation.RadiusNS != "60000000000" {
			return fail("held source/citation changed")
		}
		var selected *metrics.GaugeSeries
		for i, series := range result.Gauges.Series {
			ns, present := series.Stream.Resource[identityNamespace388]
			if c.Namespace == nil && !present || c.Namespace != nil && present && ns == *c.Namespace {
				if selected != nil {
					return fail("namespace did not identify unique retained stream")
				}
				selected = &result.Gauges.Series[i]
			}
		}
		if selected == nil {
			return fail("selected stream missing")
		}
		if !reflect.DeepEqual(c.Citation.Stream, selected.Stream) {
			return fail("citation changed selected stream identity")
		}
		matches := 0
		var sample metrics.GaugeSample
		for _, p := range selected.Samples {
			if p.AtNS == c.AtNS {
				matches++
				sample = p
			}
		}
		if matches != 1 {
			return fail("selected timestamp not a retained occurrence")
		}
		at, ok := new(big.Int).SetString(c.AtNS, 10)
		if !ok {
			return fail("reported timestamp malformed")
		}
		from := new(big.Int).Sub(new(big.Int).Set(at), radius)
		if from.Cmp(lo) < 0 {
			from.Set(lo)
		}
		to := new(big.Int).Add(new(big.Int).Set(at), radius)
		if to.Cmp(hi) > 0 {
			to.Set(hi)
		}
		if c.Range.FromNS != from.String() || c.Range.ToNS != to.String() || c.Range.AtNS != c.AtNS || c.Range.RadiusNS != "60000000000" || c.SelectedPoint.FromNS != c.AtNS || c.SelectedPoint.ToNS != c.AtNS || c.SelectedPoint.State != sample.State {
			return fail("clipping or selected-point meaning changed")
		}
		if sample.Value == nil {
			if c.SelectedPoint.Value != nil {
				return fail("unknown sample invented magnitude")
			}
		} else if c.SelectedPoint.Value == nil || *c.SelectedPoint.Value != strconv.FormatFloat(*sample.Value, 'g', -1, 64) {
			return fail("selected real zero/value lost")
		}
		plan, e := exploration.PlanNativeObservationContext(&result, selected.Stream, c.Signal, c.AtNS)
		if e != nil || !reflect.DeepEqual(plan.Question, c.ExpectedQuestion) || plan.FromNS != from.String() || plan.ToNS != to.String() || plan.Relation != "observation-context" {
			return fail("exact staged projection changed")
		}
		rows := r.Raw.Spans
		if c.Signal == "logs" {
			rows = r.Raw.Logs
		} else if c.Signal != "spans" {
			return fail("unsupported child signal")
		}
		ids := observationMembers388(rows, q, r.Fixture.Service, r.Fixture.Instance, c.Namespace, from.String(), to.String())
		pointIDs := observationMembers388(rows, q, r.Fixture.Service, r.Fixture.Instance, c.Namespace, c.AtNS, new(big.Int).Add(at, big.NewInt(1)).String())
		fullIDs := observationMembers388(rows, q, r.Fixture.Service, r.Fixture.Instance, c.Namespace, loText, hiText)
		if len(ids) <= len(pointIDs) || len(ids) >= len(fullIDs) {
			return fail("nearby context indistinguishable from 1ns point/full native window")
		}
		reported := func(rows []observationRaw388) []string {
			ids := []string{}
			for _, row := range rows {
				ids = append(ids, row.SpanID)
			}
			slices.Sort(ids)
			return ids
		}
		if !slices.Equal(reported(c.Oracle), ids) || !slices.Equal(reported(c.PointMembers), pointIDs) || !slices.Equal(reported(c.FullMembers), fullIDs) {
			return fail("independent raw membership differs from browser bank")
		}
		wantedAttempts := 1
		if c.Signal == "logs" {
			wantedAttempts = 2
		}
		if len(c.Attempts) != wantedAttempts {
			return fail("complete failure/retry sequence absent")
		}
		for i, attempt := range c.Attempts {
			readAttempts++
			if !reflect.DeepEqual(attempt.Request, c.ExpectedQuestion) {
				return fail("read drifted from staged scope")
			}
			if i == 0 && wantedAttempts == 2 {
				if attempt.Status != 502 {
					return fail("planted source refusal missing")
				}
				continue
			}
			if attempt.Status != 200 {
				return fail("explicit nearby read failed")
			}
			var child exploration.Result
			if json.Unmarshal(attempt.Response, &child) != nil || !reflect.DeepEqual(child.Question, c.ExpectedQuestion) || child.Count != strconv.Itoa(len(ids)) {
				return fail("child scope/full count changed")
			}
			actual, e := reportedResultIDs353(child)
			if e != nil || !slices.Equal(actual, ids) {
				return fail("foreign/outside/boundary child records admitted")
			}
		}
	}
	if !caseNames["named-stale-lower"] || !caseNames["missing-unknown-upper"] || !caseNames["empty-stale-zero-middle"] || readAttempts != 4 {
		return fail("required clipping/namespace/unknown/zero cases missing")
	}
	expected := map[string]int{"native-run": 1}
	for _, key := range []string{"named-stale-lower", "missing-unknown-upper", "empty-stale-zero-middle"} {
		expected[key+"-select"] = 0
		expected[key+"-run"] = 1
		expected[key+"-return"] = 0
		expected[key+"-stage"] = 0
	}
	// The failed visit and successful retry both stage the same child question.
	expected["missing-unknown-upper-refused-run"] = 1
	expected["missing-unknown-upper-refused-return"] = 0
	if len(r.Facts) != 16 {
		return fail("measured journey is only a prefix")
	}
	seen := map[string]int{}
	reads := 0
	for i, f := range r.Facts {
		n, ok := expected[f.Name]

		if !ok {
			return fail("unknown measured action")
		}
		seen[f.Name]++
		if seen[f.Name] > 1 && !strings.HasSuffix(f.Name, "-stage") {
			return fail("duplicate measured action")
		}
		if i > 0 && f.Before != r.Facts[i-1].After {
			return fail("counter discontinuity")
		}
		if f.After.Source-f.Before.Source != n || f.After.Store != 0 || f.Before.Store != 0 || f.After.Writes != 0 || f.Before.Writes != 0 {
			return fail("implicit source/store/write")
		}
		if n == 0 && f.After.SQL != f.Before.SQL || n > 0 && f.After.SQL <= f.Before.SQL {
			return fail("physical source budget mismatch")
		}
		if f.Name == "native-run" {
			if f.After.Metrics-f.Before.Metrics != 1 || f.After.Records != f.Before.Records {
				return fail("native setup read miscount")
			}
		} else if f.After.Metrics != f.Before.Metrics || f.After.Records-f.Before.Records != n {
			return fail("child read miscount")
		}
		reads += n
	}
	for name := range expected {
		visits := 1
		if name == "missing-unknown-upper-stage" {
			visits = 2
		}
		if seen[name] != visits {
			return fail("complete gesture/failed retry sequence absent")
		}
	}
	if reads != 5 {
		return fail("source read/refusal/retry budget changed")
	}
	// No uncounted telemetry family may hide outside the supplied counters.
	sourceCalls := 0
	for _, call := range r.Calls {
		if call.Method == "POST" {
			if call.Path == "/api/measurements" || call.Path == "/api/explore" {
				sourceCalls++
			} else if slices.Contains([]string{"/api/explore/page", "/api/record-aggregates", "/api/discovery", "/api/inventory", "/api/metric-discovery", "/api/tail", "/api/captures", "/api/documents", "/api/native-views"}, call.Path) {
				return fail("unmeasured source/store write")
			}
		}
	}
	if sourceCalls != 5 {
		return fail("raw source requests differ from measured events")
	}
	return nil
}
func replayObservation388(t *testing.T, label string, facts []observationFact388, mutate bool) bool {
	t.Helper()
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	base := facts[0].Before.Source
	for _, f := range facts {
		action := "history"
		if f.After.Source > f.Before.Source {
			action = "read"
		}
		reads := f.After.Source - base
		if mutate && f.Name == "named-stale-lower-stage" {
			reads++
		}
		rec.Record(context.Background(), action, nil, "obs.reads", reads, "obs.captures", f.After.Writes)
	}
	wire, _ := json.Marshal(sink.Steps())
	if e := artifact("gauge-observation388-"+label+"-steps.json", wire); e != nil {
		t.Fatal(e)
	}
	checker, e := validate.NewChecker("model/captured_trace_search_383.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	result, e := checker.Check(t.Context(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	if e = artifact("gauge-observation388-"+label+"-model.txt", []byte(result.Summary())); e != nil {
		t.Fatal(e)
	}
	return result.OK()
}
func TestQuintNativeObservationObserved388(t *testing.T) {
	path := os.Getenv("GOO_GAUGE_OBSERVATION_388_REPORT")
	if path == "" {
		t.Skip("actual gauge observation journey required")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var r observationLedger388
	if json.Unmarshal(raw, &r) != nil {
		t.Fatal("invalid actual observation report")
	}
	if e = admitObservationLedger388(r); e != nil {
		t.Fatal(e)
	}
	if !replayObservation388(t, "actual", r.Facts, false) {
		t.Fatal("actual observation budget violates model")
	}
	mutations := []func(*observationLedger388){
		func(r *observationLedger388) { r.Cases[0].Range.FromNS = strconv.FormatInt(0, 10) },
		func(r *observationLedger388) { r.Cases[0].ExpectedQuestion.FromMS++ },
		func(r *observationLedger388) { r.Cases[1].Namespace = func() *string { x := ""; return &x }() },
		func(r *observationLedger388) { r.Cases[0].ParentAfter = json.RawMessage(`{"changed":true}`) },
		func(r *observationLedger388) {
			r.Cases[1].SelectedPoint.Value = func() *string { x := "0"; return &x }()
		},
		func(r *observationLedger388) { r.Cases[2].SelectedPoint.Value = nil },
		func(r *observationLedger388) { r.Cases[0].Citation.Basis = "foreign" },
		func(r *observationLedger388) { r.Cases[0].AtNS = "not-reported" },
		func(r *observationLedger388) { r.Cases[0].Attempts[0].Request.ToMS++ },
		func(r *observationLedger388) { r.Facts[2].After.Source++ },
	}
	for i, mutate := range mutations {
		var bad observationLedger388
		json.Unmarshal(raw, &bad)
		mutate(&bad)
		if admitObservationLedger388(bad) == nil {
			t.Fatalf("semantic mutant%d admitted", i)
		}
	}
	if replayObservation388(t, "stage-read-mutant", r.Facts, true) {
		t.Fatal("stage source read escaped model")
	}
}
