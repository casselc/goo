package verification

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"github.com/casselc/goo/workbench"
	"os"
	"reflect"
	"strings"
	"testing"
)

type counterRead389 struct {
	Request  measurements.Question
	Response measurements.Result
	Status   int
}
type counterLedger389 struct {
	ProxyCalls []struct {
		SQL        string
		Parameters map[string]string
		Status     int
	}
	InitialSelectionAfterRun string
	PriorSelection           struct {
		SelectedKey string
		Data        json.RawMessage
	}
	Version string
	Errors  []string
	Fixture struct {
		Source, Target, Decimal struct{ Question measurements.Question }
	}
	SourceOracle, DestinationOracle, DecimalOracle                []counterRaw389
	SourceRun, DestinationRun, DestinationBeforeApply, DecimalRun counterRead389
	Save                                                          struct {
		Request  workbench.NativeViewInput
		Response workbench.NativeView
		Status   int
	}
	Inspect, WrongUnitInspect, WrongDomainInspect, WrongUnitSeed, WrongDomainSeed struct {
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
	WrongUnitReuseRefusal, WrongDomainReuseRefusal struct{ Status int }
	StoredArtifactBefore, StoredArtifactAfter      string
	Facts                                          []gaugeFact386
	Timeline                                       struct {
		Points []struct {
			FromNS, ToNS, State string
			Value               *string
		}
	}
	TableRows [][]string
	Calls     []struct {
		Path, Method string
		Status       int
	}
	Actor struct {
		Snapshots int
		Complete  struct {
			Observed, BudgetSafe, Selected, Current, IntervalsExact, PointsExact, RailSafe bool
			Kind, Phase                                                                    string
		}
	}
}

func admitCounterRead389(read counterRead389, q measurements.Question, raw []counterRaw389) error {
	if read.Status != 200 || !reflect.DeepEqual(read.Request, q) || !reflect.DeepEqual(read.Response.Question, q) {
		return fmt.Errorf("counter question changed")
	}
	if e := measurements.ValidateResult(&read.Response); e != nil {
		return e
	}
	expected, e := counterReference389(q, raw)
	if e != nil {
		return e
	}
	if read.Response.Counters == nil || len(read.Response.Counters.Series) != 1 || !reflect.DeepEqual(expected, read.Response.Counters.Series[0].Intervals) {
		return fmt.Errorf("raw counter interval oracle disagrees")
	}
	matched := false
	for _, row := range raw {
		selected := gaugeRaw386{AtNS: row.AtNS, ServiceName: row.ServiceName, MetricName: row.MetricName, MetricUnit: row.MetricUnit, ScopeName: row.ScopeName, ScopeVersion: row.ScopeVersion, ScopeSchemaUrl: row.ScopeSchemaUrl, ResourceSchemaUrl: row.ResourceSchemaUrl, ResourceAttributes: row.ResourceAttributes, Attributes: row.Attributes}
		if gaugeSelect386(q, selected) {
			stream := metrics.Stream{ResourceSchemaURL: row.ResourceSchemaUrl, Resource: row.ResourceAttributes, Attributes: row.Attributes, Scope: row.ScopeName, Version: row.ScopeVersion, SchemaURL: row.ScopeSchemaUrl, Name: row.MetricName, Unit: row.MetricUnit}
			if !reflect.DeepEqual(stream, read.Response.Counters.Series[0].Stream) {
				return fmt.Errorf("raw stream identity differs")
			}
			matched = true
		}
	}
	if !matched {
		return fmt.Errorf("raw counter stream absent")
	}
	if read.Response.Count != fmt.Sprint(len(expected)+1) {
		return fmt.Errorf("raw source membership changed")
	}
	return nil
}
func admitCounter389(r counterLedger389) error {
	fail := func(s string) error { return fmt.Errorf("counter389: %s", s) }
	if r.Version != "goo.nativeCounterView389/1" || len(r.Errors) != 0 {
		return fail("report incomplete")
	}
	for _, c := range []struct {
		read counterRead389
		q    measurements.Question
		raw  []counterRaw389
	}{{r.SourceRun, r.Fixture.Source.Question, r.SourceOracle}, {r.DestinationRun, r.Fixture.Target.Question, r.DestinationOracle},
		{r.DestinationBeforeApply, r.Fixture.Target.Question, r.DestinationOracle}, {r.DecimalRun, r.Fixture.Decimal.Question, r.DecimalOracle}} {
		if e := admitCounterRead389(c.read, c.q, c.raw); e != nil {
			return e
		}
	}
	if len(r.SourceOracle) != 3 || len(r.DestinationOracle) != 11 || len(r.DecimalOracle) != 3 {
		return fail("raw decoys or exports missing")
	}
	if r.Save.Status != 201 || r.Save.Response.Validate() != nil || r.Inspect.Status != 200 || !reflect.DeepEqual(r.Inspect.Response, r.Save.Response) || !reflect.DeepEqual(r.Save.Request.Question, r.Fixture.Source.Question) || !reflect.DeepEqual(r.Save.Response.Question, r.Fixture.Source.Question) {
		return fail("save/load identity")
	}
	if r.Save.Response.Declaration.Selection.Kind != "evaluated-intervals" || r.Save.Response.Declaration.Display.Value != "counter-rate" || r.Save.Response.Declaration.Display.RequiresComplete {
		return fail("display authority")
	}
	if r.Reuse.Status != 200 || r.Reuse.Request.ID != r.Save.Response.ID || r.Reuse.Response.ID != "" || !reflect.DeepEqual(r.Reuse.Request.Question, r.Fixture.Target.Question) || !reflect.DeepEqual(r.Reuse.Response.Question, r.Fixture.Target.Question) || !reflect.DeepEqual(r.Reuse.Response.Declaration, r.Save.Response.Declaration) {
		return fail("donor policy transfer")
	}
	if r.StoredArtifactBefore == "" || r.StoredArtifactBefore != r.StoredArtifactAfter {
		return fail("stored bytes changed")
	}
	var stored workbench.NativeView
	if json.Unmarshal([]byte(r.StoredArtifactBefore), &stored) != nil || stored.Validate() != nil || !reflect.DeepEqual(stored, r.Save.Response) {
		return fail("stored artifact integrity")
	}
	input := r.Save.Request
	b, _ := json.Marshal(input)
	sum := sha256.Sum256(b)
	if hex.EncodeToString(sum[:]) != stored.ID {
		return fail("canonical stored hash")
	}
	if r.WrongUnitSeed.Status != 201 || r.WrongDomainSeed.Status != 201 || r.WrongUnitSeed.Response.Validate() != nil || r.WrongDomainSeed.Response.Validate() != nil || !reflect.DeepEqual(r.WrongUnitInspect.Response, r.WrongUnitSeed.Response) || !reflect.DeepEqual(r.WrongDomainInspect.Response, r.WrongDomainSeed.Response) || r.WrongUnitSeed.Response.Question.Definition.Source.Unit == r.Fixture.Target.Question.Definition.Source.Unit || r.WrongDomainSeed.Response.Question.Definition.CounterDomain == r.Fixture.Target.Question.Definition.CounterDomain {
		return fail("negative configuration controls not distinct and valid")
	}
	if r.WrongUnitReuseRefusal.Status != 400 || r.WrongDomainReuseRefusal.Status != 400 || r.WrongUnitInspect.Status != 200 || r.WrongDomainInspect.Status != 200 {
		return fail("incompatible application accepted")
	}
	if r.Browse.Status != 200 {
		return fail("catalog missing")
	}
	found := false
	for _, s := range r.Browse.Response.Items {
		if s.ID == stored.ID {
			found = true
			if s.CounterDomain != "nonnegative-integer" || s.Family != "counter" || s.Selection != "evaluated-intervals" {
				return fail("summary domain")
			}
		}
	}
	if !found {
		return fail("saved catalog member absent")
	}
	expected, _ := counterReference389(r.Fixture.Target.Question, r.DestinationOracle)
	if len(r.Timeline.Points) != len(expected) || len(r.TableRows) != len(expected) {
		return fail("display drops intervals")
	}
	for i, p := range expected {
		actual := r.Timeline.Points[i]
		cells := r.TableRows[i]
		if len(cells) != 6 {
			return fail("counter table columns")
		}
		if p.State == "ok" {
			if cells[2] != p.Delta || cells[3] != fmt.Sprint(*p.PerSecond) || (p.ExactPerSecond == "" && cells[4] != "Not supplied") {
				return fail("table counter magnitude")
			}
		} else if cells[2] != "Unavailable" || cells[3] != "Unavailable" || cells[4] != "Unavailable" || !strings.Contains(cells[5], p.Reason) {
			return fail("table unknown becomes zero or loses reason")
		}

		if actual.FromNS != p.FromNS || actual.ToNS != p.ToNS {
			return fail("painted interval bounds")
		}
		if p.State == "ok" {
			if actual.State != "observed" || actual.Value == nil || *actual.Value != fmt.Sprint(*p.PerSecond) {
				return fail("painted rate")
			}
		} else if actual.State != "unknown" || actual.Value != nil {
			return fail("unknown becomes numeric")
		}
	}
	expectedReads := []measurements.Plan{r.DecimalRun.Response.Plan, r.SourceRun.Response.Plan, r.DestinationBeforeApply.Response.Plan, r.DestinationRun.Response.Plan, r.DestinationRun.Response.Plan}
	if len(r.ProxyCalls) != len(expectedReads) {
		return fail("unexpected physical source read")
	}
	for i, plan := range expectedReads {
		call := r.ProxyCalls[i]
		if call.Status != 200 || call.SQL != plan.SQL {
			return fail("physical source SQL changed")
		}
		for key, value := range plan.Parameters {
			if call.Parameters["param_"+key] != value {
				return fail("physical source parameter changed")
			}
		}
		if call.Parameters["readonly"] != "1" {
			return fail("physical read authority changed")
		}
	}
	metrics := 0
	for _, c := range r.Calls {
		if c.Method == "POST" && c.Path == "/api/measurements" {
			metrics++
		}
		if c.Path == "/api/captures" {
			return fail("unexpected custody")
		}
	}
	if metrics != 3 {
		return fail("unexpected measurement reads")
	}
	if len(r.Facts) != 11 {
		return fail("incomplete interaction prefix")
	}
	expectedNames := []string{"save-configuration", "destination-existing-result", "open-reuse-is-passive", "browse-metadata", "wrong-unit-refuses-apply", "wrong-domain-refuses-apply", "inspect-configuration", "apply-stages-current-question", "explicit-destination-run", "timeline-retains-unknown", "local-kind-switch-retains-selection"}
	if r.Facts[0].Before.Metrics != 1 || r.Facts[len(r.Facts)-1].After.Metrics != 3 {
		return fail("acquisition endpoints")
	}
	reads := 0
	seen := map[string]bool{}
	for i, f := range r.Facts {
		if f.Name != expectedNames[i] {
			return fail("workflow order")
		}
		if seen[f.Name] {
			return fail("duplicate boundary")
		}
		seen[f.Name] = true
		if i > 0 && f.Before.Metrics != r.Facts[i-1].After.Metrics {
			return fail("source counter discontinuity")
		}
		delta := f.After.Metrics - f.Before.Metrics
		if f.Name == "explicit-destination-run" || f.Name == "destination-existing-result" {
			if delta != 1 {
				return fail("Run missing")
			}
			reads++
		} else if delta != 0 || f.After.SQL != f.Before.SQL {
			return fail("local/config operation reads source")
		}
		if f.After.Captures != 0 {
			return fail("capture side effect")
		}
	}
	if reads != 2 || !seen["apply-stages-current-question"] || !seen["local-kind-switch-retains-selection"] {
		return fail("required workflow incomplete")
	}
	var prior struct {
		Points []struct {
			FromNS, ToNS, State string
			Value               *string
		}
	}
	if json.Unmarshal(r.PriorSelection.Data, &prior) != nil || !reflect.DeepEqual(prior.Points, r.Timeline.Points) {
		return fail("prior destination evidence differs")
	}
	if r.PriorSelection.SelectedKey != "0" || r.InitialSelectionAfterRun != "7" || len(r.PriorSelection.Data) == 0 {
		return fail("existing destination selection absent")
	}
	if r.Actor.Snapshots == 0 || r.Actor.Complete.Phase != "complete" || !r.Actor.Complete.Observed || !r.Actor.Complete.BudgetSafe || !r.Actor.Complete.Selected || !r.Actor.Complete.Current || !r.Actor.Complete.IntervalsExact || !r.Actor.Complete.PointsExact || !r.Actor.Complete.RailSafe || r.Actor.Complete.Kind != "timeline" {
		return fail("actor outcome missing")
	}
	return nil
}
func TestQuintNativeCounterViewObserved389(t *testing.T) {
	path := os.Getenv("GOO_NATIVE_COUNTER_VIEW_389_REPORT")
	if path == "" {
		t.Skip("actual counter view journey required")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var r counterLedger389
	if e = json.Unmarshal(raw, &r); e != nil {
		t.Fatal(e)
	}
	if e = admitCounter389(r); e != nil {
		t.Fatal(e)
	}
	modelFacts := append([]gaugeFact386(nil), r.Facts...)
	for i := range modelFacts {
		if modelFacts[i].Name == "destination-existing-result" {
			modelFacts[i].Name = "explicit-destination-run"
		}
	}
	if !replayGauge386(t, "counter389-actual", modelFacts, false) {
		t.Fatal("measured model rejected")
	}
	mutations := []func(*counterLedger389){func(r *counterLedger389) { r.StoredArtifactAfter += " " },
		func(r *counterLedger389) { r.TableRows[0][4] = "Not recorded" },
		func(r *counterLedger389) { r.ProxyCalls[0].Parameters["param_p0"] = "0" },
		func(r *counterLedger389) { r.TableRows[1][2] = "Not recorded" },
		func(r *counterLedger389) { r.Reuse.Response.Question = r.Fixture.Source.Question },
		func(r *counterLedger389) { r.WrongDomainReuseRefusal.Status = 200 },
		func(r *counterLedger389) { r.Timeline.Points[1].Value = new("0") },
		func(r *counterLedger389) { r.Actor.Complete.Selected = false },
		func(r *counterLedger389) { r.Facts[0].After.Metrics++ },
		func(r *counterLedger389) { r.DestinationOracle[0].Value = "999" }}
	for i, mutate := range mutations {
		var bad counterLedger389
		json.Unmarshal(raw, &bad)
		mutate(&bad)
		if admitCounter389(bad) == nil {
			t.Fatalf("semantic mutant %d admitted", i)
		}
	}
	if replayGauge386(t, "counter389-apply-read-mutant", modelFacts, true) {
		t.Fatal("Apply source read escaped model")
	}
}
