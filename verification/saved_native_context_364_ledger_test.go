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

type savedNativeState364 struct{ Calls, Source, Store, SQL int }
type savedNativeFact364 struct {
	Name          string
	Before, After savedNativeState364
}
type savedNativeCase364 struct {
	Kind                                string
	Question                            measurements.Question
	Capture                             exploration.CapturePage
	StoredBytesBefore, StoredBytesAfter string
	SourceBefore, SourceAfter           json.RawMessage
	Interval                            struct{ FromNS, ToNS string }
	ContextPlan                         exploration.NativeRecordContext
	Run, Refused, Pending               struct {
		Request                   exploration.Query
		Response                  json.RawMessage
		Status                    int
		AbortsBefore, AbortsAfter int
		Aborted                   bool
	}
	Oracle struct {
		Count string
		Rows  []struct {
			ID          string
			TimestampNS string `json:"timestamp_ns"`
		}
	}
	OldActions struct{ Before, After json.RawMessage }
	Facts      []savedNativeFact364
}

func savedNativeAdmission364(c savedNativeCase364) bool {
	if c.Kind != "counter" && c.Kind != "histogram" {
		return false
	}
	if len(c.SourceBefore) == 0 || len(c.SourceAfter) == 0 || len(c.OldActions.Before) == 0 || len(c.OldActions.After) == 0 || c.Capture.Measurement == nil || c.StoredBytesBefore == "" || c.StoredBytesBefore != c.StoredBytesAfter || savedJSON360(c.SourceBefore) != savedJSON360(c.SourceAfter) || savedJSON360(c.OldActions.Before) != savedJSON360(c.OldActions.After) {
		return false
	}
	if !peLiveCaptureHash(c.Capture) {
		return false
	}
	if c.Capture.Measurement.ReadAt == "" || c.Capture.Measurement.Basis == "" {
		return false
	}
	if savedJSON360(c.Capture.Measurement.Question) != savedJSON360(c.Question) || savedJSON360(c.Capture.Receipt.Question.Measurement) != savedJSON360(c.Question) {
		return false
	}
	matched := 0
	checkInterval := func(stream metrics.Stream, from, to string) {
		if from != c.Interval.FromNS || to != c.Interval.ToNS {
			return
		}
		planned, err := exploration.PlanNativeRecordContext(c.Question, stream, c.ContextPlan.Question.Signal, from, to)
		if err == nil && savedJSON360(planned) == savedJSON360(c.ContextPlan) {
			matched++
		}
	}
	if c.Kind == "counter" && c.Capture.Measurement.Counters != nil {
		for _, series := range c.Capture.Measurement.Counters.Series {
			for _, interval := range series.Intervals {
				checkInterval(series.Stream, interval.FromNS, interval.ToNS)
			}
		}
	}
	if c.Kind == "histogram" && c.Capture.Measurement.Histograms != nil {
		for _, series := range c.Capture.Measurement.Histograms.Series {
			for _, interval := range series.Intervals {
				checkInterval(series.Stream, interval.FromNS, interval.ToNS)
			}
		}
	}
	if matched != 1 {
		return false
	}
	if c.ContextPlan.FromNS != c.Interval.FromNS || c.ContextPlan.ToNS != c.Interval.ToNS || c.ContextPlan.Grain != "interval-context" {
		return false
	}
	if !nhLiveSameQuery(c.ContextPlan.Question, c.Run.Request) || !nhLiveSameQuery(c.Run.Request, c.Refused.Request) || !nhLiveSameQuery(c.Run.Request, c.Pending.Request) || c.Run.Status != 200 || c.Refused.Status < 500 || c.Pending.Status != 200 {
		return false
	}
	var result exploration.Result
	if json.Unmarshal(c.Run.Response, &result) != nil || !nhLiveSameQuery(result.Question, c.Run.Request) || result.Count != c.Oracle.Count {
		return false
	}
	got, idsErr := reportedResultIDs353(result)
	if idsErr != nil {
		return false
	}
	want := []string{}
	for _, row := range c.Oracle.Rows {
		want = append(want, row.ID)
	}
	slices.Sort(want)
	if !slices.Equal(got, want) {
		return false
	}
	if !c.Pending.Aborted || c.Pending.AbortsAfter <= c.Pending.AbortsBefore {
		return false
	}
	names := []string{"open-select", "stage", "run", "return", "stage-refused", "run-refused", "return-refused", "stage-pending-run", "return-pending", "old-actions"}
	reads := []int{0, 0, 1, 0, 0, 1, 0, 1, 0, 0}
	if len(c.Facts) != len(names) {
		return false
	}
	for i, f := range c.Facts {
		if f.Name != names[i] || f.After.Source-f.Before.Source != reads[i] || f.After.SQL < f.Before.SQL || reads[i] == 0 && f.After.SQL != f.Before.SQL {
			return false
		}
		if i > 0 && f.After.Store != f.Before.Store {
			return false
		}
	}
	return true
}
func savedNativeReplay364(t *testing.T, mutant int) bool {
	t.Helper()
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	reads, depth := 0, 0
	actions := []string{"history", "stage", "read", "back", "stage", "read", "back", "stage", "read", "back", "history"}
	for i, action := range actions {
		if action == "stage" {
			depth++
		}
		if action == "back" {
			depth--
		}
		if action == "read" {
			reads++
		}
		observed := reads
		if mutant == 1 && i == 1 {
			observed++
		}
		if mutant == 2 && i == 0 {
			action = "back"
		}
		rec.Record(context.Background(), action, nil, "obs.reads", observed, "obs.depth", depth, "obs.coherent", true)
	}
	checker, err := validate.NewChecker("model/population_drilldown.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	report, err := checker.Check(t.Context(), sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	return report.OK()
}

// Project recorded macrosteps, retaining their actual cumulative source counts.
// The combined stage-pending-run has no intermediate snapshot: stage uses its
// recorded Before.Source and read its After.Source, a logical macrostep split.
func savedNativeObservedReplay364(t *testing.T, label string, facts []savedNativeFact364) bool {
	t.Helper()
	if len(facts) == 0 {
		t.Fatal("empty actual saved context facts")
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	baseline, depth := facts[0].Before.Source, 0
	record := func(action string, source int) {
		rec.Record(context.Background(), action, nil, "obs.reads", source-baseline, "obs.depth", depth, "obs.coherent", true)
	}
	for _, f := range facts {
		switch f.Name {
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
		case "open-select", "old-actions":
			record("history", f.After.Source)
		default:
			t.Fatal("unknown actual fact", f.Name)
		}
	}
	encoded, err := json.Marshal(sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	if err = artifact("saved-native364-"+label+"-steps.json", encoded); err != nil {
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
	if err = artifact("saved-native364-"+label+"-model.txt", []byte(report.Summary())); err != nil {
		t.Fatal(err)
	}
	return report.OK()
}
func TestQuintSavedNativeContextProtocol364(t *testing.T) {
	if !savedNativeReplay364(t, 0) || savedNativeReplay364(t, 1) || savedNativeReplay364(t, 2) {
		t.Fatal("saved context source/depth positive or rejecting controls")
	}
}
func TestQuintSavedNativeContextObserved364(t *testing.T) {
	path := os.Getenv("GOO_SAVED_NATIVE_CONTEXT_364_REPORT")
	if path == "" {
		t.Skip("actual saved-native context report opt-in")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var report struct {
		Version string
		Cases   []savedNativeCase364
		Fixture struct {
			Environment string
			Records     []struct {
				ID, Service, Namespace, Instance, TimestampNS string
				Env                                           map[string]string
			}
		}
	}
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if report.Version != "goo.savedNativeContext364/1" || len(report.Cases) != 2 {
		t.Fatal("missing actual saved native cases")
	}
	for _, c := range report.Cases {
		from, ok := new(big.Int).SetString(c.Interval.FromNS, 10)
		if !ok {
			t.Fatal("invalid native interval")
		}
		to, ok := new(big.Int).SetString(c.Interval.ToNS, 10)
		if !ok {
			t.Fatal("invalid native interval")
		}
		ids := []string{}
		for _, row := range report.Fixture.Records {
			at, valid := new(big.Int).SetString(row.TimestampNS, 10)
			if !valid {
				t.Fatal("invalid raw timestamp")
			}
			current, present := row.Env["deployment.environment.name"]
			match := present && current == report.Fixture.Environment
			if c.Question.Definition.EnvironmentSelector.Policy == "effective" && !present {
				match = row.Env["deployment.environment"] == report.Fixture.Environment
			}
			if match && row.Service == "native-worker" && row.Namespace == "inventory" && row.Instance == "worker-a" && at.Cmp(from) >= 0 && at.Cmp(to) < 0 {
				ids = append(ids, row.ID)
			}
		}
		want := []string{}
		for _, row := range c.Oracle.Rows {
			want = append(want, row.ID)
		}
		slices.Sort(ids)
		slices.Sort(want)
		if !slices.Equal(ids, want) || c.Oracle.Count != strconv.Itoa(len(ids)) {
			t.Fatal("independent raw resource/environment/half-open membership")
		}

		if !savedNativeAdmission364(c) {
			t.Fatal("saved capture/context/ownership/source admission", c.Kind)
		}
		bad := c
		bad.Run.Request.ToMS++
		if savedNativeAdmission364(bad) {
			t.Fatal("changed contextual question admitted")
		}
		bad = c
		bad.StoredBytesAfter += "x"
		if savedNativeAdmission364(bad) {
			t.Fatal("capture byte mutation admitted")
		}
		bad = c
		bad.SourceAfter = json.RawMessage(`{"changed":true}`)
		if savedNativeAdmission364(bad) {
			t.Fatal("retained origin mutation admitted")
		}
		if !savedNativeObservedReplay364(t, c.Kind, c.Facts) {
			t.Fatal("actual admitted saved context lifecycle")
		}
		counter := slices.Clone(c.Facts)
		counter[1].After.Source++
		if savedNativeObservedReplay364(t, c.Kind+"-bad-counter", counter) {
			t.Fatal("model accepted mutated actual Stage source counter")
		}
		order := slices.Clone(c.Facts)
		order[1], order[3] = order[3], order[1]
		if savedNativeObservedReplay364(t, c.Kind+"-bad-order", order) {
			t.Fatal("model accepted mutated actual return-before-stage")
		}

	}
}
