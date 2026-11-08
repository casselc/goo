package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"unicode"

	"github.com/casselc/goo/measurements"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type nativeRecoveryCounters378 struct{ Source, SQL, Writes int }
type nativeRecoveryFact378 struct {
	Name          string
	Before, After nativeRecoveryCounters378
}
type nativeRecoveryReply378 struct {
	Kind     string
	Status   int
	Request  measurements.Question
	Response json.RawMessage
}
type nativeRecoveryCase378 struct{ Failed, Recovered nativeRecoveryReply378 }
type nativeRecoveryPoint378 struct {
	ScopeName, ScopeVersion, ScopeSchemaURL, MetricName, MetricUnit string
	ResourceAttributes, Attributes                                  map[string]string
	StartNS, EndNS, Count                                           string
}
type nativeRecoveryReport378 struct {
	Version, Status string
	Fixture         struct{ Question measurements.Question }
	Cases           []nativeRecoveryCase378
	Facts           []nativeRecoveryFact378
	RawPoints       []nativeRecoveryPoint378
	Errors          []string
	NativeCounters  nativeRecoveryCounters378
	Cancelled       struct {
		Call   struct{ Closed bool }
		Status string
	}
	Late struct {
		Status            int
		Request           measurements.Question
		Response          json.RawMessage
		ReplacementStatus string
	}
}

func nativeRecoverySame378(a, b any) bool { return reflect.DeepEqual(a, b) }
func nativeRecoveryFactByName378(report nativeRecoveryReport378, name string) (nativeRecoveryFact378, bool) {
	for _, fact := range report.Facts {
		if fact.Name == name {
			return fact, true
		}
	}
	return nativeRecoveryFact378{}, false
}
func nativeRecoveryPassive378(fact nativeRecoveryFact378) bool {
	return fact.After.Source == fact.Before.Source && fact.After.SQL == fact.Before.SQL && fact.After.Writes == fact.Before.Writes
}
func nativeRecoveryResult378(report nativeRecoveryReport378, reply nativeRecoveryReply378) error {
	if reply.Status != 200 || !nativeRecoverySame378(reply.Request, report.Fixture.Question) {
		return fmt.Errorf("recovery request/status changed")
	}
	var result measurements.Result
	if err := json.Unmarshal(reply.Response, &result); err != nil {
		return err
	}
	if err := measurements.ValidateResult(&result); err != nil {
		return err
	}
	if !nativeRecoverySame378(result.Question, report.Fixture.Question) || result.Count != "2" || len(result.Rows) != 2 {
		return fmt.Errorf("recovery result question/population changed")
	}
	// Independently compare retained physical identities/counts with the actual raw
	// ClickHouse oracle, rather than deriving expectations through the Goo planner.
	for _, raw := range result.Rows {
		var row struct {
			Resource, Attributes       map[string]string
			Scope, Version, Name, Unit string
			SchemaURL                  string `json:"schema_url"`
			StartNS                    string `json:"start_ns"`
			EndNS                      string `json:"end_ns"`
			Count                      string
		}
		if err := json.Unmarshal(raw, &row); err != nil {
			return err
		}
		matched := 0
		for _, point := range report.RawPoints {
			if row.Scope == point.ScopeName && row.Version == point.ScopeVersion && row.SchemaURL == point.ScopeSchemaURL && row.Name == point.MetricName && row.Unit == point.MetricUnit && row.StartNS == point.StartNS && row.EndNS == point.EndNS && row.Count == point.Count && nativeRecoverySame378(row.Resource, point.ResourceAttributes) && nativeRecoverySame378(row.Attributes, point.Attributes) {
				matched++
			}
		}
		if matched != 1 {
			return fmt.Errorf("recovery row not uniquely present in independent source oracle")
		}
	}
	return nil
}
func nativeRecoveryRefusal378(report nativeRecoveryReport378, reply nativeRecoveryReply378, kind string) error {
	if reply.Status != 502 || reply.Kind != kind || !nativeRecoverySame378(reply.Request, report.Fixture.Question) {
		return fmt.Errorf("failed fixed question/status/kind changed")
	}
	var body map[string]json.RawMessage
	if err := json.Unmarshal(reply.Response, &body); err != nil || len(body) != 2 {
		return fmt.Errorf("unrecognized source refusal envelope")
	}
	var refusal map[string]string
	var reason string
	if json.Unmarshal(body["error"], &reason) != nil || json.Unmarshal(body["refusal"], &refusal) != nil || len(refusal) != 2 || refusal["kind"] != kind {
		return fmt.Errorf("unrecognized closed source refusal")
	}
	expected := map[string]string{"unavailable": "Source unavailable.", "malformed": "Source response rejected."}[kind]
	detail := refusal["detail"]
	if reason != expected || detail == "" || len(detail) > 256 || strings.Contains(string(reply.Response), "PRIVATE") || strings.IndexFunc(detail, unicode.IsControl) >= 0 {
		return fmt.Errorf("unsafe source refusal reason/detail")
	}
	return nil
}
func nativeRecoveryAdmission378(report nativeRecoveryReport378) error {
	if report.Version != "goo.nativeRecovery378/1" || report.Status != "passed" || len(report.Errors) != 0 || len(report.Cases) != 2 || len(report.RawPoints) != 2 {
		return fmt.Errorf("incomplete actual native recovery journey")
	}
	question := report.Fixture.Question
	if err := question.Validate(); err != nil {
		return err
	}
	if question.FromNS == "" || question.ToNS == "" || question.Definition.EnvironmentSelector == nil || question.Definition.Source.Namespace == nil {
		return fmt.Errorf("missing fixed nanosecond/typed namespace question")
	}
	for index, kind := range []string{"unavailable", "malformed"} {
		c := report.Cases[index]
		if err := nativeRecoveryRefusal378(report, c.Failed, kind); err != nil {
			return err
		}
		if err := nativeRecoveryResult378(report, c.Recovered); err != nil {
			return err
		}
		for _, name := range []string{"inspect-" + kind + "-details", "hidden-inert-disabled-detached-retry-" + kind, "retired-after-success-" + kind} {
			fact, ok := nativeRecoveryFactByName378(report, name)
			if !ok || !nativeRecoveryPassive378(fact) {
				return fmt.Errorf("missing or active passive recovery boundary %s", name)
			}
		}
	}
	for _, fact := range report.Facts {
		delta := 0
		if strings.HasSuffix(fact.Name, "-source-refusal") || strings.HasPrefix(fact.Name, "explicit-exact-retry-") || fact.Name == "cancel-retire-held-retry" || fact.Name == "real-success-held-before-delivery" {
			delta = 1
		}
		if fact.After.Source-fact.Before.Source != delta || fact.After.Writes != fact.Before.Writes || fact.After.SQL < fact.Before.SQL || delta == 0 && fact.After.SQL != fact.Before.SQL {
			return fmt.Errorf("unexpected source/store work at %s", fact.Name)
		}
	}
	for _, name := range []string{"editing-retires-retry", "navigation-retires-retry", "late-success-cannot-replace-owner"} {
		fact, ok := nativeRecoveryFactByName378(report, name)
		if !ok || !nativeRecoveryPassive378(fact) {
			return fmt.Errorf("missing retired owner boundary %s", name)
		}
	}
	if !report.Cancelled.Call.Closed || !strings.Contains(strings.ToLower(report.Cancelled.Status), "cancel") {
		return fmt.Errorf("held cancellation not observed")
	}
	if report.Late.ReplacementStatus == "" {
		return fmt.Errorf("late replacement owner not observed")
	}
	if err := nativeRecoveryResult378(report, nativeRecoveryReply378{Status: report.Late.Status, Request: report.Late.Request, Response: report.Late.Response}); err != nil {
		return fmt.Errorf("held real-success observation: %w", err)
	}
	if report.NativeCounters.Source != 10 || report.NativeCounters.Writes != 0 {
		return fmt.Errorf("unexpected complete native acquisition/store budget")
	}
	return nil
}
func nativeRecoveryReplay378(t *testing.T, report nativeRecoveryReport378, caseIndex int, poisonDetails bool) bool {
	t.Helper()
	c := report.Cases[caseIndex]
	kind := c.Failed.Kind
	details, _ := nativeRecoveryFactByName378(report, "inspect-"+kind+"-details")
	if poisonDetails {
		details.After.Source++
	}
	guards, _ := nativeRecoveryFactByName378(report, "hidden-inert-disabled-detached-retry-"+kind)
	retired, _ := nativeRecoveryFactByName378(report, "retired-after-success-"+kind)
	// Every projected event comes from the recorded request/response or passive
	// boundary for this actual recovery. No widget plan or success payload is made.
	events := []struct {
		action   string
		coherent bool
	}{
		{"start", nativeRecoverySame378(c.Failed.Request, report.Fixture.Question)},
		{"source_failed", c.Failed.Status == 502},
		{"refused", nativeRecoveryPassive378(details)},
		{"refused", nativeRecoveryPassive378(guards)},
		{"retry", nativeRecoverySame378(c.Recovered.Request, c.Failed.Request)},
		{"succeed", c.Recovered.Status == 200 && nativeRecoveryResult378(report, c.Recovered) == nil},
		{"refused", nativeRecoveryPassive378(retired)},
		{"finish", report.Status == "passed"},
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink, qobs.WithActor(kind))
	for _, event := range events {
		rec.Record(context.Background(), event.action, nil, "obs.coherent", event.coherent)
	}
	checker, err := validate.NewChecker("model/widget_recovery_354.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	result, err := checker.Check(context.Background(), sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	return result.OK()
}
func TestQuintNativeRecoveryObserved378(t *testing.T) {
	path := os.Getenv("GOO_NATIVE_RECOVERY_378_REPORT")
	if path == "" {
		t.Skip("actual native recovery journey opt-in")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var report nativeRecoveryReport378
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if err = nativeRecoveryAdmission378(report); err != nil {
		t.Fatal(err)
	}
	clone := func() nativeRecoveryReport378 {
		var r nativeRecoveryReport378
		if err := json.Unmarshal(raw, &r); err != nil {
			t.Fatal(err)
		}
		return r
	}
	changed := clone()
	changed.Cases[0].Recovered.Request.FromMS++
	if nativeRecoveryAdmission378(changed) == nil {
		t.Fatal("changed retry window admitted")
	}
	changed = clone()
	for i := range changed.Facts {
		if changed.Facts[i].Name == "inspect-unavailable-details" {
			changed.Facts[i].After.Source++
			break
		}
	}
	if nativeRecoveryAdmission378(changed) == nil {
		t.Fatal("passive source acquisition admitted")
	}
	for i := range report.Cases {
		if !nativeRecoveryReplay378(t, report, i, false) {
			t.Fatal("actual recovery rejected", i)
		}
	}
	if nativeRecoveryReplay378(t, report, 0, true) {
		t.Fatal("model admitted source work during details disclosure")
	}
}
