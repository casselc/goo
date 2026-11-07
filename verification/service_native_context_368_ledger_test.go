package verification

import (
	"context"
	"encoding/json"
	"math/big"
	"os"
	"slices"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type serviceNativeReport368 struct {
	Version                   string
	Measurement               measurements.Result
	SelectedStream            metrics.Stream
	ContextPlan               exploration.NativeRecordContext
	SourceBefore, SourceAfter json.RawMessage
	Run, Refused, Pending     struct {
		Request                     exploration.Query
		Response                    json.RawMessage
		Status                      int
		Aborted                     bool
		AbortsBefore, AbortsAfter   int
		BeforeRelease, AfterRelease json.RawMessage
	}
	OldActions struct{ Before, After json.RawMessage }
	Oracle     struct {
		Count string
		Rows  []struct {
			ID          string
			TimestampNS string `json:"timestamp_ns"`
		}
	}
	Fixture struct {
		Environment string
		Records     []struct {
			ID, Service, Namespace, Instance, TimestampNS string
			Env                                           map[string]string
		}
	}
	Facts []savedNativeFact364
}

func serviceNativeAdmission368(r serviceNativeReport368) bool {
	m := r.Measurement
	if r.Version != "goo.serviceNativeContext368/1" || m.Question.Definition.Kind != "histogram" || m.Histograms == nil || m.ReadAt == "" || m.Basis == "" || len(r.SourceBefore) == 0 || savedJSON360(r.SourceBefore) != savedJSON360(r.SourceAfter) || len(r.OldActions.Before) == 0 || savedJSON360(r.OldActions.Before) != savedJSON360(r.OldActions.After) {
		return false
	}
	matched := 0
	for _, series := range m.Histograms.Series {
		if savedJSON360(series.Stream) != savedJSON360(r.SelectedStream) {
			continue
		}
		for _, interval := range series.Intervals {
			if interval.FromNS != r.ContextPlan.FromNS || interval.ToNS != r.ContextPlan.ToNS {
				continue
			}
			plan, err := exploration.PlanNativeRecordContext(m.Question, series.Stream, r.ContextPlan.Question.Signal, interval.FromNS, interval.ToNS)
			if err == nil && savedJSON360(plan) == savedJSON360(r.ContextPlan) {
				matched++
			}
		}
	}
	if matched != 1 || r.ContextPlan.Grain != "interval-context" || !nhLiveSameQuery(r.ContextPlan.Question, r.Run.Request) || !nhLiveSameQuery(r.Run.Request, r.Refused.Request) || !nhLiveSameQuery(r.Run.Request, r.Pending.Request) || r.Run.Status != 200 || r.Refused.Status < 500 || r.Pending.Status != 200 || !r.Pending.Aborted || r.Pending.AbortsAfter <= r.Pending.AbortsBefore || len(r.Pending.BeforeRelease) == 0 || savedJSON360(r.Pending.BeforeRelease) != savedJSON360(r.Pending.AfterRelease) {
		return false
	}
	var result exploration.Result
	if json.Unmarshal(r.Run.Response, &result) != nil || !nhLiveSameQuery(result.Question, r.Run.Request) || result.Count != r.Oracle.Count {
		return false
	}
	got, err := reportedResultIDs353(result)
	if err != nil {
		return false
	}
	want := []string{}
	for _, row := range r.Oracle.Rows {
		want = append(want, row.ID)
	}
	slices.Sort(want)
	if !slices.Equal(got, want) {
		return false
	}
	from, ok := new(big.Int).SetString(r.ContextPlan.FromNS, 10)
	if !ok {
		return false
	}
	to, ok := new(big.Int).SetString(r.ContextPlan.ToNS, 10)
	if !ok {
		return false
	}
	ids := []string{}
	resource := r.SelectedStream.Resource
	for _, row := range r.Fixture.Records {
		at, valid := new(big.Int).SetString(row.TimestampNS, 10)
		if !valid {
			return false
		}
		current, present := row.Env["deployment.environment.name"]
		match := present && current == r.Fixture.Environment
		if selector := m.Question.Definition.EnvironmentSelector; selector != nil && selector.Policy == "effective" && !present {
			match = row.Env["deployment.environment"] == r.Fixture.Environment
		}
		if match && row.Service == resource["service.name"] && row.Namespace == resource["service.namespace"] && row.Instance == resource["service.instance.id"] && at.Cmp(from) >= 0 && at.Cmp(to) < 0 {
			ids = append(ids, row.ID)
		}
	}
	slices.Sort(ids)
	if !slices.Equal(ids, want) || r.Oracle.Count != strconv.Itoa(len(ids)) {
		return false
	}
	names := []string{"open-select", "stage", "run", "return", "stage-refused", "run-refused", "return-refused", "stage-pending-run", "return-pending", "old-actions"}
	reads := []int{0, 0, 1, 0, 0, 1, 0, 1, 0, 0}
	if len(r.Facts) != len(names) {
		return false
	}
	for i, f := range r.Facts {
		if f.Name != names[i] || f.After.Source-f.Before.Source != reads[i] || f.After.Store != f.Before.Store || f.After.SQL < f.Before.SQL || reads[i] == 0 && f.After.SQL != f.Before.SQL {
			return false
		}
		// The independent raw SQL oracle runs between open/select and Stage.
		// It is not a product read; product source/store counters must remain
		// continuous while the proxy's cumulative SQL counter may advance.
		if i > 0 && (f.Before.Source != r.Facts[i-1].After.Source || f.Before.Store != r.Facts[i-1].After.Store || f.Before.SQL < r.Facts[i-1].After.SQL) {
			return false
		}
	}
	return true
}

// These are actual observed boundaries and cumulative read counts. There is no
// declared successful action script after admission. The combined pending macro
// splits at its recorded pre-read counter, as in the existing native context path.
func serviceNativeObservedReplay368(t *testing.T, label string, facts []savedNativeFact364) bool {
	t.Helper()
	if len(facts) == 0 {
		t.Fatal("empty service native observed facts")
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	baseline, depth := facts[0].Before.Source, 0
	record := func(action string, source int) {
		rec.Record(context.Background(), action, nil, "obs.reads", source-baseline, "obs.depth", depth, "obs.coherent", true)
	}
	for _, f := range facts {
		switch f.Name {
		case "open-select", "old-actions":
			record("history", f.After.Source)
		case "stage", "stage-refused":
			depth++
			record("stage", f.After.Source)
		case "run", "run-refused":
			record("read", f.After.Source)
		case "return", "return-refused", "return-pending":
			depth--
			record("back", f.After.Source)
		case "stage-pending-run":
			depth++
			record("stage", f.Before.Source)
			record("read", f.After.Source)
		default:
			t.Fatal("unknown service native observed fact", f.Name)
		}
	}
	encoded, err := json.Marshal(sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	if err = artifact("service-native368-"+label+"-steps.json", encoded); err != nil {
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
	if err = artifact("service-native368-"+label+"-model.txt", []byte(report.Summary())); err != nil {
		t.Fatal(err)
	}
	return report.OK()
}

func TestQuintServiceNativeContextObserved368(t *testing.T) {
	path := os.Getenv("GOO_SERVICE_NATIVE_CONTEXT_368_REPORT")
	if path == "" {
		t.Skip("actual service native context report opt-in")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var report struct {
		Version string
		Cases   []serviceNativeReport368
		Fixture json.RawMessage
	}
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Cases) != 1 {
		t.Fatal("missing actual service native context case")
	}
	r := report.Cases[0]
	r.Version = report.Version
	if err = json.Unmarshal(report.Fixture, &r.Fixture); err != nil {
		t.Fatal(err)
	}
	if !serviceNativeAdmission368(r) {
		t.Fatal("actual held service histogram/context/owner/source/raw oracle admission")
	}
	bad := r
	bad.ContextPlan.Question.ToMS++
	if serviceNativeAdmission368(bad) {
		t.Fatal("changed contextual question admitted")
	}
	bad = r
	bad.SourceAfter = json.RawMessage(`{"changed":true}`)
	if serviceNativeAdmission368(bad) {
		t.Fatal("replaced service origin admitted")
	}
	bad = r
	bad.ContextPlan.FromNS = "1"
	if serviceNativeAdmission368(bad) {
		t.Fatal("foreign export interval admitted")
	}
	if !serviceNativeObservedReplay368(t, "actual", r.Facts) {
		t.Fatal("actual service native context lifecycle")
	}
	counter := slices.Clone(r.Facts)
	counter[1].After.Source++
	if serviceNativeObservedReplay368(t, "bad-stage-read", counter) {
		t.Fatal("model accepted Stage source read")
	}
	order := slices.Clone(r.Facts)
	order[1], order[3] = order[3], order[1]
	if serviceNativeObservedReplay368(t, "bad-return-order", order) {
		t.Fatal("model accepted Return before Stage")
	}
}
