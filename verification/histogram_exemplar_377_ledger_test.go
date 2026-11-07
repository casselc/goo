package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"net/url"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type exemplarCounters377 struct{ Source, SQL, Writes int }
type exemplarFact377 struct {
	Name          string
	Before, After exemplarCounters377
}
type exemplarTrace377 struct {
	TraceID     string `json:"TraceId"`
	SpanID      string `json:"SpanId"`
	Service     string `json:"ServiceName"`
	Name        string `json:"SpanName"`
	TimestampNS string
	Resource    map[string]string `json:"ResourceAttributes"`
}
type exemplarRawPoint377 struct {
	Scope                    string            `json:"ScopeName"`
	Version                  string            `json:"ScopeVersion"`
	Schema                   string            `json:"ScopeSchemaUrl"`
	Metric                   string            `json:"MetricName"`
	Unit                     string            `json:"MetricUnit"`
	Resource                 map[string]string `json:"ResourceAttributes"`
	Attributes               map[string]string
	StartNS, TimeNS          string
	TraceIDs, SpanIDs, Times []string
}
type exemplarJourney377 struct {
	Version, Status string
	Fixture         struct {
		Question                        measurements.Question
		TraceID, Namespace, Environment string
	}
	Parent, ReactivatedParent struct {
		Request  measurements.Question
		Response measurements.Result
	}
	Child struct {
		Request  exploration.Query
		Response exploration.Result
	}
	SelectedReport struct {
		EndNS, SourceRow, Text string
		Rows                   []struct{ Text, Link string }
	}
	StagedQuestion            exploration.Query
	StagedHref, TraceOnlyHref string
	RawPoints                 []exemplarRawPoint377
	RawTraces, ExpectedTraces []exemplarTrace377
	HeldBefore, HeldAfter     struct {
		Text, EndNS, SourceRow, Focused string
		Open, SamePanel, SameParent     bool
		ScrollY                         float64
	}
	Facts          []exemplarFact377
	NativeCounters exemplarCounters377
	ProxyCalls     []struct {
		SQL        string
		Parameters map[string]string
		Status     int
	}
	Errors []string
}

func exemplarEnvironment377(d measurements.Definition, resource map[string]string) bool {
	if d.EnvironmentSelector == nil {
		value, present := resource[identity.EnvironmentKey]
		return present && value == d.Environment
	}
	selection := d.EnvironmentSelector
	value, present := resource[identity.EnvironmentKey]
	if selection.Policy == "effective" && !present {
		value, present = resource[identity.LegacyEnvironmentKey]
	}
	switch selection.Scope.Mode {
	case "named":
		return present && value == selection.Scope.Value
	case "empty":
		return present && value == ""
	case "unreported":
		return !present
	}
	return false
}
func exemplarLink377(href string) (exploration.Query, error) {
	var value struct {
		Version int
		Kind    string
		Record  exploration.Query
	}
	if !strings.HasPrefix(href, "#step=") {
		return value.Record, fmt.Errorf("missing record step link")
	}
	raw, err := url.PathUnescape(strings.TrimPrefix(href, "#step="))
	if err != nil {
		return value.Record, err
	}
	if err = json.Unmarshal([]byte(raw), &value); err != nil {
		return value.Record, err
	}
	if value.Version != 1 || value.Kind != "record" {
		return value.Record, fmt.Errorf("not a reported record question")
	}
	return value.Record, value.Record.Validate()
}
func exemplarTraceKeys377(rows []exemplarTrace377) []string {
	out := []string{}
	for _, row := range rows {
		out = append(out, strings.Join([]string{row.TraceID, row.SpanID, row.Service, row.TimestampNS}, "\x00"))
	}
	slices.Sort(out)
	return out
}
func histogramExemplarAdmission377(r exemplarJourney377) error {
	fail := func(reason string) error { return fmt.Errorf("377 observed admission: %s", reason) }
	if r.Version != "goo.histogramExemplarJourney377/1" || r.Status != "passed" || len(r.Errors) != 0 || len(r.Facts) == 0 {
		return fail("only a complete successful native report is admissible")
	}
	if !reflect.DeepEqual(r.Parent.Request, r.Fixture.Question) || !reflect.DeepEqual(r.Parent.Response.Question, r.Fixture.Question) || !reflect.DeepEqual(r.ReactivatedParent.Request, r.Fixture.Question) || !reflect.DeepEqual(r.ReactivatedParent.Response.Question, r.Fixture.Question) {
		return fail("original native question changed")
	}
	for _, result := range []*measurements.Result{&r.Parent.Response, &r.ReactivatedParent.Response} {
		if err := measurements.ValidateResult(result); err != nil {
			return fail("retained native replay: " + err.Error())
		}
		plan, err := measurements.Compile(result.Question)
		if err != nil || !reflect.DeepEqual(plan, result.Plan) {
			return fail("actual native plan differs from its scoped compiler")
		}
	}
	if r.Parent.Response.Count != "2" || len(r.Parent.Response.Rows) != 2 || r.ReactivatedParent.Response.Count != "2" || len(r.ReactivatedParent.Response.Rows) != 2 || len(r.RawPoints) != 3 {
		return fail("scoped native exports or independent decoy inventory differ")
	}
	if r.Parent.Response.HistogramExemplars == nil || r.ReactivatedParent.Response.HistogramExemplars == nil {
		return fail("retained reported export metadata missing")
	}
	findExport := func(result *measurements.Result) (*measurements.HistogramExemplarExport, error) {
		var selected *measurements.HistogramExemplarExport
		for i := range result.HistogramExemplars.Exports {
			export := &result.HistogramExemplars.Exports[i]
			if export.Stream.Attributes["probe.case"] == "primary" {
				if selected != nil {
					return nil, fail("ambiguous primary export")
				}
				selected = export
			}
		}
		if selected == nil || selected.State != "available" || selected.Total != "6" || len(selected.Items) != 6 {
			return nil, fail("missing complete retained six-reference source export")
		}
		return selected, nil
	}
	first, err := findExport(&r.Parent.Response)
	if err != nil {
		return err
	}
	selected, err := findExport(&r.ReactivatedParent.Response)
	if err != nil {
		return err
	}
	if r.SelectedReport.EndNS != first.EndNS || r.SelectedReport.SourceRow != strconv.Itoa(first.SourceRow) || len(r.SelectedReport.Rows) != 6 {
		return fail("selected inspector belongs to another raw export")
	}
	matches := 0
	for _, point := range r.RawPoints {
		if point.TimeNS != first.EndNS || point.Attributes["probe.case"] != "primary" {
			continue
		}
		matches++
		if point.StartNS != first.StartNS || point.Scope != first.Stream.Scope || point.Version != first.Stream.Version || point.Schema != first.Stream.SchemaURL || point.Metric != first.Stream.Name || point.Unit != first.Stream.Unit || !reflect.DeepEqual(point.Resource, first.Stream.Resource) || !reflect.DeepEqual(point.Attributes, first.Stream.Attributes) || len(point.TraceIDs) != 6 || len(point.SpanIDs) != 6 || len(point.Times) != 6 {
			return fail("independent source export identity or ordinal arrays differ")
		}
		for index, item := range first.Items {
			if item.Index != index || item.TraceID != point.TraceIDs[index] || item.SpanID != point.SpanIDs[index] || item.TimeNS != point.Times[index] {
				return fail("captured reference differs from the raw source ordinal")
			}
		}
	}
	if matches != 1 {
		return fail("raw source export association is not unique")
	}
	contextPlan, err := exploration.PlanHistogramExemplarContext(&r.ReactivatedParent.Response, selected.Stream, selected.EndNS, 0)
	if err != nil {
		return fail("selected reported reference cannot be projected: " + err.Error())
	}
	question := r.Fixture.Question
	from, to := question.BoundsNS()
	upper, ok := new(big.Int).SetString(to, 10)
	if !ok {
		return fail("invalid exact original bound")
	}
	upper.Sub(upper, big.NewInt(1))
	namespace := question.Definition.Source.Namespace
	if namespace == nil {
		namespace = &identity.Namespace{Mode: "missing"}
		if value, present := selected.Stream.Resource[identity.NamespaceKey]; present {
			namespace = &identity.Namespace{Mode: "value", Value: value}
		}
	}
	namespaceLeaf := exploration.Predicate{Field: &exploration.Field{Kind: "resource", Key: identity.NamespaceKey}, Op: "missing"}
	if namespace.Mode == "value" {
		namespaceLeaf.Op = "eq"
		namespaceLeaf.Value = namespace.Value
	}
	expected := exploration.Query{Signal: "spans", Environment: question.Definition.Environment, EnvironmentSelector: question.Definition.EnvironmentSelector, FromMS: question.FromMS, ToMS: question.ToMS, Limit: 100, Where: &exploration.Predicate{All: []exploration.Predicate{{Field: &exploration.Field{Kind: "column", Key: "trace_id"}, Op: "eq", Value: selected.Items[0].TraceID}, namespaceLeaf, {Field: &exploration.Field{Kind: "column", Key: "timestamp_ns"}, Op: "gte", Value: from}, {Field: &exploration.Field{Kind: "column", Key: "timestamp_ns"}, Op: "lte", Value: upper.String()}}}}
	if expected.Where.All[0].Value != r.Fixture.TraceID || !reflect.DeepEqual(contextPlan.Question, expected) || !reflect.DeepEqual(r.StagedQuestion, expected) || !reflect.DeepEqual(r.Child.Request, expected) || !reflect.DeepEqual(r.Child.Response.Question, expected) {
		return fail("trace/namespace/original exact window changed, or metric/span predicates leaked")
	}
	linked, err := exemplarLink377(r.StagedHref)
	if err != nil || !reflect.DeepEqual(linked, expected) {
		return fail("staged native link changed the exact child question")
	}
	traceOnly, err := exploration.PlanHistogramExemplarContext(&r.Parent.Response, first.Stream, first.EndNS, 5)
	if err != nil || traceOnly.SpanID != "" {
		return fail("valid trace-only reference lost authority or invented span focus")
	}
	linked, err = exemplarLink377(r.TraceOnlyHref)
	if err != nil || !reflect.DeepEqual(linked, traceOnly.Question) {
		return fail("trace-only staged question differs")
	}
	for index, item := range first.Items {
		valid := len(item.TraceID) == 32 && strings.Trim(item.TraceID, "0") != ""
		if valid {
			linked, err = exemplarLink377(r.SelectedReport.Rows[index].Link)
			if err != nil || !reflect.DeepEqual(linked, expected) {
				return fail("reported ordinal link changed its scoped trace")
			}
		} else if r.SelectedReport.Rows[index].Link != "" {
			return fail("unreported trace identity acquired an action")
		}
	}
	if len(r.RawTraces) != 4 {
		return fail("independent distributed trace and decoy rows missing")
	}
	lo, _ := new(big.Int).SetString(from, 10)
	hi, _ := new(big.Int).SetString(to, 10)
	oracle := []exemplarTrace377{}
	for _, row := range r.RawTraces {
		at, valid := new(big.Int).SetString(row.TimestampNS, 10)
		if !valid {
			return fail("raw trace timestamp unavailable")
		}
		value, present := row.Resource[identity.NamespaceKey]
		nsMatches := namespace.Mode == "missing" && !present || namespace.Mode == "value" && present && value == namespace.Value
		if row.TraceID == r.Fixture.TraceID && nsMatches && exemplarEnvironment377(question.Definition, row.Resource) && at.Cmp(lo) >= 0 && at.Cmp(hi) < 0 {
			oracle = append(oracle, row)
		}
	}
	if len(oracle) != 2 || oracle[0].Service == oracle[1].Service || !reflect.DeepEqual(exemplarTraceKeys377(oracle), exemplarTraceKeys377(r.ExpectedTraces)) || r.Child.Response.Count != "2" || len(r.Child.Response.Rows) != 2 {
		return fail("distributed peers or independent scoped trace membership differ")
	}
	actual := []exemplarTrace377{}
	for _, raw := range r.Child.Response.Rows {
		var row struct {
			TraceID     string            `json:"TraceId"`
			SpanID      string            `json:"SpanId"`
			Service     string            `json:"ServiceName"`
			TimestampNS string            `json:"timestamp_ns"`
			Resource    map[string]string `json:"ResourceAttributes"`
		}
		if json.Unmarshal(raw, &row) != nil {
			return fail("child record row unavailable")
		}
		actual = append(actual, exemplarTrace377{TraceID: row.TraceID, SpanID: row.SpanID, Service: row.Service, TimestampNS: row.TimestampNS, Resource: row.Resource})
	}
	if !reflect.DeepEqual(exemplarTraceKeys377(actual), exemplarTraceKeys377(oracle)) {
		return fail("child rows do not match raw source IDs and exact timestamps")
	}
	if !r.HeldAfter.SamePanel || !r.HeldAfter.SameParent || r.HeldAfter.Text != r.HeldBefore.Text || r.HeldAfter.EndNS != r.HeldBefore.EndNS || r.HeldAfter.SourceRow != r.HeldBefore.SourceRow || r.HeldBefore.EndNS != selected.EndNS || r.HeldBefore.SourceRow != strconv.Itoa(selected.SourceRow) || !r.HeldBefore.Open || !r.HeldAfter.Open || math.Abs(r.HeldAfter.ScrollY-r.HeldBefore.ScrollY) > 2 || r.HeldAfter.Focused != "Inspect reported trace" {
		return fail("held source export DOM/focus/scope was replaced")
	}
	seen := map[string]int{}
	previous := exemplarCounters377{}
	for _, fact := range r.Facts {
		if fact.Before != previous || fact.Before.Writes != 0 || fact.After.Writes != 0 {
			return fail("unobserved reads/writes between boundaries")
		}
		delta := fact.After.Source - fact.Before.Source
		if fact.Name == "explicit-source-run" || fact.Name == "explicit-source-reactivation" || fact.Name == "explicit-reported-trace-run" {
			expectedSQL := 1
			if fact.Name == "explicit-reported-trace-run" {
				expectedSQL = 3
			}
			if delta != 1 || fact.After.SQL-fact.Before.SQL != expectedSQL {
				return fail("explicit run read budget changed")
			}
		} else if delta != 0 || fact.After.SQL != fact.Before.SQL {
			return fail("passive/staged/held boundary acquired source data")
		}
		seen[fact.Name]++
		previous = fact.After
	}
	for _, name := range []string{"measurement-preview-is-passive", "explicit-source-run", "empty-source-export-inspection", "primary-source-export-inspection", "trace-only-reference-stage", "trace-only-held-return", "explicit-source-reactivation", "reported-trace-stage", "explicit-reported-trace-run", "held-source-export-return", "retired-reported-trace-callback"} {
		if seen[name] != 1 {
			return fail("missing/duplicated actual boundary " + name)
		}
	}
	if previous.Source != 3 || previous.SQL != 5 || r.NativeCounters != previous || len(r.ProxyCalls) < 5 {
		return fail("main parent2/child1 acquisition budget changed")
	}
	childPlan, err := exploration.Compile(expected)
	if err != nil || !reflect.DeepEqual(childPlan, r.Child.Response.Plan) {
		return fail("actual child source plan differs from exact scoped compiler")
	}
	expectedSQL := map[string]int{r.Parent.Response.Plan.SQL: 1}
	expectedSQL[r.ReactivatedParent.Response.Plan.SQL]++
	for _, sql := range []string{childPlan.Rows, childPlan.Count, childPlan.Fields} {
		expectedSQL[sql]++
	}
	for _, call := range r.ProxyCalls[:5] {
		if call.Status != 200 || expectedSQL[call.SQL] == 0 || call.Parameters["readonly"] != "1" {
			return fail("unexpected physical source acquisition")
		}
		expectedSQL[call.SQL]--
	}
	for _, count := range expectedSQL {
		if count != 0 {
			return fail("missing declared physical source query")
		}
	}
	return nil
}
func histogramExemplarReplay377(t *testing.T, label string, facts []exemplarFact377) bool {
	t.Helper()
	sink := &qobs.MemorySink{}
	recorder := qobs.NewRecorder(sink)
	depth := 0
	for _, fact := range facts {
		action := ""
		switch fact.Name {
		case "explicit-source-run", "explicit-source-reactivation", "explicit-reported-trace-run":
			action = "read"
		case "trace-only-reference-stage", "reported-trace-stage":
			action = "stage"
			depth++
		case "trace-only-held-return", "held-source-export-return":
			action = "back"
			depth--
		case "measurement-preview-is-passive", "empty-source-export-inspection", "primary-source-export-inspection", "retired-reported-trace-callback":
			action = "history"
		default:
			t.Fatal("unmapped actual exemplar boundary", fact.Name)
		}
		recorder.Record(context.Background(), action, nil, "obs.reads", fact.After.Source, "obs.depth", depth, "obs.coherent", true)
	}
	encoded, err := json.Marshal(sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	if err = artifact("histogram-exemplar377-"+label+"-steps.json", encoded); err != nil {
		t.Fatal(err)
	}
	checker, err := validate.NewChecker("model/population_drilldown.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	report, err := checker.Check(t.Context(), sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	if err = artifact("histogram-exemplar377-"+label+"-model.txt", []byte(report.Summary())); err != nil {
		t.Fatal(err)
	}
	return report.OK()
}
func TestQuintHistogramExemplarObserved377(t *testing.T) {
	path := os.Getenv("GOO_HISTOGRAM_EXEMPLAR_377_REPORT")
	if path == "" {
		t.Skip("actual reported histogram exemplar journey opt-in")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var report exemplarJourney377
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if err = histogramExemplarAdmission377(report); err != nil {
		t.Fatal(err)
	}
	clone := func() exemplarJourney377 {
		var value exemplarJourney377
		if err := json.Unmarshal(raw, &value); err != nil {
			t.Fatal(err)
		}
		return value
	}
	changed := clone()
	changed.Child.Request.Where.All[0].Value = strings.Repeat("f", 32)
	if histogramExemplarAdmission377(changed) == nil {
		t.Fatal("changed trace question admitted")
	}
	changed = clone()
	changed.Child.Request.Where.All[1].Value = "other-namespace"
	if histogramExemplarAdmission377(changed) == nil {
		t.Fatal("broadened/changed namespace admitted")
	}
	changed = clone()
	changed.SelectedReport.SourceRow = "999"
	if histogramExemplarAdmission377(changed) == nil {
		t.Fatal("foreign source export association admitted")
	}
	if !histogramExemplarReplay377(t, "actual", report.Facts) {
		t.Fatal("actual observed stage/read/held boundaries violate existing navigation model")
	}
	facts := slices.Clone(report.Facts)
	mutating := false
	for index := range facts {
		if facts[index].Name == "reported-trace-stage" {
			mutating = true
			facts[index].After.Source++
			continue
		}
		if mutating {
			facts[index].Before.Source++
			facts[index].After.Source++
		}
	}
	if histogramExemplarReplay377(t, "extra-stage-read", facts) {
		t.Fatal("extra stage source read admitted by navigation model")
	}
}
