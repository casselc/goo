package verification

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type gaugeRaw386 struct {
	AtNS, Value, ServiceName, MetricName, MetricUnit, ScopeName, ScopeVersion, ScopeSchemaUrl, ResourceSchemaUrl string
	Flags                                                                                                        int
	ResourceAttributes, Attributes                                                                               map[string]string
}
type gaugeSample386 struct {
	AtNS, State, Reason string
	Value               *json.RawMessage
}
type gaugeResult386 struct {
	Question             measurements.Question
	Count, ReadAt, Basis string
	Gauges               *struct {
		FromNS, ToNS string
		Series       []struct {
			Stream  metrics.Stream
			Samples []gaugeSample386
		}
	}
}
type gaugeRead386 struct {
	Request  measurements.Question
	Response gaugeResult386
	Status   int
}
type gaugeFact386 struct {
	Name          string
	Before, After nativeReuseCounters371
}
type gaugeRect386 struct{ X, Y, Width, Height float64 }
type gaugeMarker386 struct {
	Key, State, FromNS, ToNS, Left, Top, Pressed string
	Value                                        *string
	Rect                                         gaugeRect386
}
type gaugeRail386 struct {
	Label, Text                  string
	NumericPlot, UnavailablePlot gaugeRect386
	Numeric, Unavailable         []gaugeMarker386
}
type gaugeLedger386 struct {
	TimelineRail, TimelineRailNarrow gaugeRail386
	Version                          string
	Errors                           []string
	Fixture                          struct {
		Source, Target struct {
			Question measurements.Question
			Stream   metrics.Stream
		}
	}
	SourceOracle, DestinationOracle []gaugeRaw386
	SourceRun, DestinationRun       gaugeRead386
	Save                            struct {
		Request  workbench.NativeViewInput
		Response workbench.NativeView
		Status   int
	}
	Inspect, HeldInspect, WrongUnitInspect struct {
		Response workbench.NativeView
		Status   int
	}
	Reuse struct {
		Request  workbench.NativeViewReuseRequest
		Response workbench.NativeView
		Status   int
	}
	Browse struct {
		Response workbench.NativeViewList
		Status   int
	}
	WrongUnitSeed struct {
		Response workbench.NativeView
		Status   int
	}
	CounterRefusal, MissingView struct{ Status int }
	StaleInspect                struct {
		CurrentID, InspectText string
		ApplyDisabled          bool
	}
	TableRows [][]string
	Timeline  struct {
		FromNS, ToNS, Kind, Unit string
		Total, Omitted           int
		Points                   []struct {
			FromNS, ToNS, State string
			Value               *string
		}
	}
	StoredArtifactBefore, StoredArtifactAfter string
	Facts                                     []gaugeFact386
	Calls                                     []struct {
		Path, Method string
		Body         json.RawMessage
		Status       int
	}
	ProxyCalls []struct {
		SQL    string
		Status int
	}
}

func gaugeValue386(raw *json.RawMessage) *string {
	if raw == nil {
		return nil
	}
	var s string
	if json.Unmarshal(*raw, &s) != nil {
		s = string(*raw)
	}
	return &s
}
func gaugeSelect386(q measurements.Question, row gaugeRaw386) bool {
	d := q.Definition
	b := d.Source
	if !slices.Contains(b.Services, row.ServiceName) || !continuationNamespace352(b.Namespace, row.ResourceAttributes) || row.MetricName != b.Metric || row.MetricUnit != b.Unit || row.ScopeName != b.Scope || row.ScopeVersion != b.Version || row.ScopeSchemaUrl != b.SchemaURL {
		return false
	}
	selection := d.EnvironmentSelector
	mode, name := 0, d.Environment
	if selection != nil {
		name = selection.Scope.Value
		if selection.Policy == "effective" {
			mode = 2
		}
		if selection.Scope.Mode == "empty" {
			mode = 3
		} else if selection.Scope.Mode == "unreported" {
			mode = 4
		}
	}
	if !continuationEnvironment352(mode, name, row.ResourceAttributes) {
		return false
	}
	for k, v := range b.Attributes {
		if value, ok := row.Attributes[k]; !ok || value != v {
			return false
		}
	}
	if s := b.ExactStream; s != nil {
		if !reflect.DeepEqual(s.Resource, row.ResourceAttributes) || !reflect.DeepEqual(s.Attributes, row.Attributes) || s.ResourceSchemaURL != row.ResourceSchemaUrl {
			return false
		}
	}
	from, to := q.FromNS, q.ToNS
	if from == "" {
		from = strconv.FormatInt(q.FromMS*1000000, 10)
	}
	if to == "" {
		to = strconv.FormatInt(q.ToMS*1000000, 10)
	}
	at, ok := new(big.Int).SetString(row.AtNS, 10)
	lo, ok2 := new(big.Int).SetString(from, 10)
	hi, ok3 := new(big.Int).SetString(to, 10)
	return ok && ok2 && ok3 && at.Cmp(lo) >= 0 && at.Cmp(hi) < 0
}
func gaugeReference386(q measurements.Question, rows []gaugeRaw386) []gaugeSample386 {
	selected := []gaugeRaw386{}
	for _, row := range rows {
		if gaugeSelect386(q, row) {
			selected = append(selected, row)
		}
	}
	slices.SortFunc(selected, func(a, b gaugeRaw386) int {
		x, _ := new(big.Int).SetString(a.AtNS, 10)
		y, _ := new(big.Int).SetString(b.AtNS, 10)
		return x.Cmp(y)
	})
	out := []gaugeSample386{}
	end := q.ToNS
	if end == "" {
		end = strconv.FormatInt(q.ToMS*1000000, 10)
	}
	hi, _ := new(big.Int).SetString(end, 10)
	age := new(big.Int).Mul(big.NewInt(q.Definition.MaxGapMS), big.NewInt(1000000))
	for _, row := range selected {
		sample := gaugeSample386{AtNS: row.AtNS, State: "observed"}
		if row.Flags&1 != 0 {
			sample.State = "unknown"
			sample.Reason = "no recorded value"
		} else {
			v := json.RawMessage(strconv.Quote(row.Value))
			sample.Value = &v
			at, _ := new(big.Int).SetString(row.AtNS, 10)
			if new(big.Int).Sub(hi, at).Cmp(age) > 0 {
				sample.State = "stale"
				sample.Reason = "older than declared freshness"
			}
		}
		out = append(out, sample)
	}
	return out
}
func admitGaugeRead386(read gaugeRead386, q measurements.Question, rows []gaugeRaw386) error {
	if len(rows) == 0 {
		return fmt.Errorf("gauge386: physical bank missing")
	}
	if read.Status != 200 || !reflect.DeepEqual(read.Request, q) || !reflect.DeepEqual(read.Response.Question, q) || read.Response.Gauges == nil || len(read.Response.Gauges.Series) != 1 || read.Response.Basis == "" || read.Response.ReadAt == "" {
		return fmt.Errorf("gauge386: missing exact actual measurement")
	}
	expected := gaugeReference386(q, rows)
	actual := read.Response.Gauges.Series[0]
	if read.Response.Count != strconv.Itoa(len(expected)) || len(actual.Samples) != len(expected) || read.Response.Gauges.FromNS != q.FromNS || read.Response.Gauges.ToNS != q.ToNS {
		return fmt.Errorf("gauge386: population/window changed")
	}
	for i, want := range expected {
		got := actual.Samples[i]
		if got.AtNS != want.AtNS || got.State != want.State || got.Reason != want.Reason || !reflect.DeepEqual(gaugeValue386(got.Value), gaugeValue386(want.Value)) {
			return fmt.Errorf("gauge386: raw observation/value/freshness changed at %d", i)
		}
	}
	if !reflect.DeepEqual(actual.Stream.Resource, rows[0].ResourceAttributes) && q.Definition.Source.ExactStream == nil {
		return fmt.Errorf("gauge386: donor stream changed")
	}
	if q.Definition.Source.ExactStream != nil && !reflect.DeepEqual(actual.Stream, *q.Definition.Source.ExactStream) {
		return fmt.Errorf("gauge386: exact destination stream changed")
	}
	return nil
}
func admitGauge386(r gaugeLedger386) error {
	fail := func(s string) error { return fmt.Errorf("native gauge386: %s", s) }
	if r.Version != "goo.nativeGaugeView386/1" || len(r.Errors) > 0 {
		return fail("wrong report/errors")
	}
	q := r.Fixture.Target.Question
	if e := admitGaugeRead386(r.SourceRun, r.Fixture.Source.Question, r.SourceOracle); e != nil {
		return e
	}
	if e := admitGaugeRead386(r.DestinationRun, q, r.DestinationOracle); e != nil {
		return e
	}
	if len(r.DestinationOracle) != 6 || len(gaugeReference386(q, r.DestinationOracle)) != 4 {
		return fail("source decoys/oracle missing")
	}
	if r.Save.Status != 201 || r.Save.Response.Validate() != nil || r.Save.Response.Question.Definition.Kind != "gauge" || r.Save.Response.Declaration.Selection.Kind != "retained-observations" || r.Save.Response.Declaration.Kind.Name != "table" || r.Save.Response.Declaration.Display.RequiresComplete {
		return fail("gauge declaration changed")
	}
	if !reflect.DeepEqual(r.Save.Request.Question, r.Fixture.Source.Question) || !reflect.DeepEqual(r.Save.Response.Question, r.Fixture.Source.Question) || r.Save.Request.Title != r.Save.Response.Title || !reflect.DeepEqual(r.Save.Request.Declaration, r.Save.Response.Declaration) || r.Inspect.Status != 200 || !reflect.DeepEqual(r.Inspect.Response, r.Save.Response) {
		return fail("configuration authority changed")
	}
	if r.Reuse.Status != 200 || r.Reuse.Request.ID != r.Save.Response.ID || r.Reuse.Response.ID != "" || r.Reuse.Response.Title != r.Save.Response.Title || !reflect.DeepEqual(r.Reuse.Request.Question, q) || !reflect.DeepEqual(r.Reuse.Response.Question, q) || !reflect.DeepEqual(r.Reuse.Response.Declaration, r.Save.Response.Declaration) {
		return fail("donor data/scope entered destination")
	}
	if r.WrongUnitSeed.Status != 201 || r.WrongUnitInspect.Status != 200 || r.WrongUnitSeed.Response.Validate() != nil || !reflect.DeepEqual(r.WrongUnitInspect.Response, r.WrongUnitSeed.Response) || r.WrongUnitInspect.Response.Question.Definition.Source.Unit == q.Definition.Source.Unit || r.CounterRefusal.Status != 400 || r.MissingView.Status != 404 {
		return fail("unsupported family/unit/missing configuration not refused")
	}
	if r.Browse.Status != 200 {
		return fail("metadata absent")
	}
	found := false
	for _, item := range r.Browse.Response.Items {
		if item.ID == r.Save.Response.ID {
			found = item.Family == "gauge" && item.Selection == "retained-observations" && item.Kind == "table" && item.Unit == q.Definition.Source.Unit
		}
	}
	if !found {
		return fail("typed metadata missing")
	}
	if r.StoredArtifactBefore == "" || r.StoredArtifactBefore != r.StoredArtifactAfter {
		return fail("immutable bytes changed")
	}
	var stored workbench.NativeView
	if json.Unmarshal([]byte(r.StoredArtifactBefore), &stored) != nil || !reflect.DeepEqual(stored, r.Save.Response) {
		return fail("physical artifact changed")
	}
	wire, _ := json.Marshal(workbench.NativeViewInput{Version: stored.Version, Title: stored.Title, Question: stored.Question, Declaration: stored.Declaration})
	sum := sha256.Sum256(wire)
	if hex.EncodeToString(sum[:]) != stored.ID {
		return fail("physical digest changed")
	}
	expected := gaugeReference386(q, r.DestinationOracle)
	if r.Timeline.FromNS != q.FromNS || r.Timeline.ToNS != q.ToNS || r.Timeline.Kind != "gauge" || r.Timeline.Unit != q.Definition.Source.Unit || r.Timeline.Total != len(expected) || r.Timeline.Omitted != 0 || len(r.Timeline.Points) != len(expected) || len(r.TableRows) != len(expected) {
		return fail("retained projection changed")
	}
	for i, want := range expected {
		p := r.Timeline.Points[i]
		v := gaugeValue386(want.Value)
		if p.FromNS != want.AtNS || p.ToNS != want.AtNS || p.State != want.State || !reflect.DeepEqual(p.Value, v) {
			return fail("timeline inserted/interpolated values")
		}
		cells := r.TableRows[i]
		if len(cells) != 4 || cells[2] != want.State || cells[3] != func() string {
			if want.Reason == "" {
				return "Not recorded"
			}
			return want.Reason
		}() || cells[0] == want.AtNS || !strings.Contains(cells[0], "20") {
			return fail("table observation semantics changed")
		}
		if v != nil && cells[1] != *v || v == nil && cells[1] == "0" {
			return fail("table invented zero or lost recorded zero")
		}
	}
	for _, rail := range []gaugeRail386{r.TimelineRail, r.TimelineRailNarrow} {
		if e := admitGaugeRail386(rail, q, expected); e != nil {
			return e
		}
	}
	if r.HeldInspect.Status != 200 || !reflect.DeepEqual(r.HeldInspect.Response, r.Save.Response) || r.StaleInspect.CurrentID != strings.Repeat("f", 64) || !r.StaleInspect.ApplyDisabled || r.StaleInspect.InspectText != "" {
		return fail("late metadata became current")
	}
	// No other telemetry endpoint may hide outside the measurement counter.
	measurementsSeen := 0
	for _, call := range r.Calls {
		if call.Method == "POST" && slices.Contains([]string{"/api/explore", "/api/explore/page", "/api/record-aggregates", "/api/tail", "/api/discovery", "/api/inventory", "/api/metric-discovery", "/api/compare"}, call.Path) {
			return fail("unmeasured telemetry read")
		}
		if strings.HasPrefix(call.Path, "/api/captures") {
			return fail("implicit evidence capture")
		}
		if call.Method == "POST" && call.Path == "/api/measurements" {
			measurementsSeen++
			var question measurements.Question
			if json.Unmarshal(call.Body, &question) != nil || !reflect.DeepEqual(question, q) && !reflect.DeepEqual(question, r.Fixture.Source.Question) {
				return fail("foreign measurement read")
			}
		}
	}
	if measurementsSeen != 2 {
		return fail("actual measurement calls missing")
	}
	expectedMetadata := map[string]int{"save-configuration": 1, "open-reuse-is-passive": 0, "browse-metadata": 1, "wrong-unit-refuses-apply": 1, "inspect-configuration": 1, "apply-stages-current-question": 1, "explicit-destination-run": 0, "timeline-retains-unknown": 0, "local-kind-switch-retains-selection": 0, "late-inspect-is-not-current": 1, "missing-configuration-is-not-data": 1}
	if len(r.Facts) != len(expectedMetadata) {
		return fail("gesture sequence missing")
	}
	seen := map[string]bool{}
	for index, f := range r.Facts {
		if index > 0 && f.Before != r.Facts[index-1].After {
			return fail("measured counter discontinuity")
		}
		m, ok := expectedMetadata[f.Name]
		if !ok || seen[f.Name] {
			return fail("unexpected/duplicate gesture")
		}
		seen[f.Name] = true
		n := 0
		if f.Name == "explicit-destination-run" {
			n = 1
		}
		if f.After.Metrics-f.Before.Metrics != n || f.After.Metadata-f.Before.Metadata != m || f.After.Captures != f.Before.Captures || f.After.Captures != 0 {
			return fail("implicit telemetry/capture")
		}
		if n == 0 && f.After.SQL != f.Before.SQL || n > 0 && f.After.SQL <= f.Before.SQL {
			return fail("source SQL budget changed")
		}
		if f.Before.SQL < 0 || f.After.SQL > len(r.ProxyCalls) {
			return fail("physical source range invalid")
		}
	}
	return nil
}

// Placement is measured from actual DOM parents and rectangles, independently
// from the point data: an unknown point at a numeric maximum is not acceptable.
func admitGaugeRail386(rail gaugeRail386, q measurements.Question, expected []gaugeSample386) error {
	fail := func(s string) error { return fmt.Errorf("gauge386 unavailable rail: %s", s) }
	if rail.Label != "Unavailable observations positioned by time" || rail.Text != "Unavailable observations · no numeric value" || len(rail.Numeric) != 3 || len(rail.Unavailable) != 1 {
		return fail("label or placement population changed")
	}
	p, u := rail.NumericPlot, rail.UnavailablePlot
	if p.Width <= 0 || p.Height <= 0 || u.Width <= 0 || u.Height <= 0 || math.Abs(p.X-u.X) > 1 || math.Abs(p.Width-u.Width) > 1 || u.Y < p.Y+p.Height-1 {
		return fail("rail intersects numeric domain or changes time axis")
	}
	lo, _ := new(big.Int).SetString(q.FromNS, 10)
	hi, _ := new(big.Int).SetString(q.ToNS, 10)
	width := new(big.Int).Sub(hi, lo)
	for index, point := range append(slices.Clone(rail.Numeric), rail.Unavailable...) {
		key, e := strconv.Atoi(point.Key)
		if e != nil || key != index || key >= len(expected) {
			return fail("marker key changed")
		}
		want := expected[key]
		if point.FromNS != want.AtNS || point.ToNS != want.AtNS || point.State != want.State || !reflect.DeepEqual(point.Value, gaugeValue386(want.Value)) {
			return fail("marker data changed")
		}
		at, _ := new(big.Int).SetString(want.AtNS, 10)
		ratio, _ := new(big.Rat).SetFrac(new(big.Int).Sub(at, lo), width).Float64()
		left, e := strconv.ParseFloat(strings.TrimSuffix(point.Left, "%"), 64)
		if e != nil || math.Abs(left-100*ratio) > 0.0011 {
			return fail("reported time X changed")
		}
		if want.State == "unknown" {
			if point.Value != nil || point.Top != "" || point.Pressed != "true" {
				return fail("unknown has numeric position/value or lost selection")
			}
			center := point.Rect.Y + point.Rect.Height/2
			if center < p.Y+p.Height || center < u.Y-1 || center > u.Y+u.Height+1 {
				return fail("unavailable marker remains in numeric plot")
			}
		}
		if math.Abs(point.Rect.X+point.Rect.Width/2-(p.X+p.Width*ratio)) > 1 {
			return fail("marker geometry changed time X")
		}
	}
	return nil
}
func replayGauge386(t *testing.T, label string, facts []gaugeFact386, mutant bool) bool {
	t.Helper()
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	base := facts[0].Before.Metrics
	for _, f := range facts {
		action := "history"
		if f.Name == "explicit-destination-run" {
			action = "read"
		}
		reads := f.After.Metrics - base
		if mutant && f.Name == "apply-stages-current-question" {
			reads++
		}
		rec.Record(context.Background(), action, nil, "obs.reads", reads, "obs.captures", 0)
	}
	wire, _ := json.Marshal(sink.Steps())
	if e := artifact("native-gauge386-"+label+"-steps.json", wire); e != nil {
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
	if e = artifact("native-gauge386-"+label+"-model.txt", []byte(result.Summary())); e != nil {
		t.Fatal(e)
	}
	return result.OK()
}
func TestQuintNativeGaugeViewObserved386(t *testing.T) {
	path := os.Getenv("GOO_NATIVE_GAUGE_VIEW_386_REPORT")
	if path == "" {
		t.Skip("actual native gauge display report required")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var r gaugeLedger386
	if json.Unmarshal(raw, &r) != nil {
		t.Fatal("invalid report")
	}
	if e = admitGauge386(r); e != nil {
		t.Fatal(e)
	}
	if !replayGauge386(t, "actual", r.Facts, false) {
		t.Fatal("actual measured facts violate model")
	}

	mutations := []func(*gaugeLedger386){
		func(r *gaugeLedger386) { r.StoredArtifactAfter += " " },
		func(r *gaugeLedger386) { r.Reuse.Response.Question = r.Fixture.Source.Question },
		func(r *gaugeLedger386) { r.Reuse.Response.Question.Definition.Source.Unit = "By" },
		func(r *gaugeLedger386) { r.Reuse.Response.Question.Definition.Kind = "histogram" },
		func(r *gaugeLedger386) { r.Reuse.Response.Question.Definition.MaxGapMS++ },
		func(r *gaugeLedger386) { r.Save.Response.Declaration.Display.RequiresComplete = true },
		func(r *gaugeLedger386) { r.DestinationRun.Response.Gauges.Series[0].Samples[0].State = "observed" },
		func(r *gaugeLedger386) {
			zero := json.RawMessage(`0`)
			r.DestinationRun.Response.Gauges.Series[0].Samples[3].Value = &zero
		},
		func(r *gaugeLedger386) { r.Timeline.Points[1].Value = nil },
		func(r *gaugeLedger386) { r.Facts[5].After.Metrics++ },
		func(r *gaugeLedger386) { r.StaleInspect.ApplyDisabled = false },
		func(r *gaugeLedger386) { v := "0"; r.TimelineRail.Unavailable[0].Value = &v },
		func(r *gaugeLedger386) { r.TimelineRail.Unavailable[0].Rect.Y = r.TimelineRail.NumericPlot.Y },
		func(r *gaugeLedger386) { r.TimelineRail.Unavailable[0].Rect.X += 14 },
	}

	for i, mutate := range mutations {
		var bad gaugeLedger386
		json.Unmarshal(raw, &bad)
		mutate(&bad)
		if admitGauge386(bad) == nil {
			t.Fatalf("semantic mutant%d survived", i)
		}
	}
	if replayGauge386(t, "read-on-apply-mutant", r.Facts, true) {
		t.Fatal("read-on-apply mutant survived model")
	}
}
