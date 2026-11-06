package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type reportedRaw353 struct {
	Trace, ID, Parent, Service string
	Resource                   map[string]string
}
type reportedCase353 struct {
	Key                                   string
	Scope                                 exploration.InventoryQuestion
	InventoryCapture, Captured            exploration.CapturePage
	Evidence                              exploration.ServiceDiscovery
	Question                              exploration.Query
	Raw, Oracle, Wider                    []reportedRaw353
	Result                                exploration.Result
	CaptureRequest                        struct{ Question exploration.CaptureQuestion }
	CaptureBytesBefore, CaptureBytesAfter string
	Facts                                 []serviceContinuationFact352
}
type reportedLedger353 struct {
	Version string
	Fixture struct {
		Token, Environment, Parent, Child, ParentNS, EmptyNS string
		FromMS, ToMS                                         int64
		Rows                                                 []struct {
			Service, Trace, ID, ParentID, Tag string
			NS                                *string
			Env                               map[string]string
			At                                int64
		}
	}
	Cases  []reportedCase353
	Calls  []serviceContinuationCall352
	Errors []string
	Scalar struct {
		Question exploration.Query
		Result   exploration.Result
	}
}

func reportedRawKey353(row reportedRaw353) string {
	resource := map[string]string{}
	for _, key := range []string{"deployment.environment.name", "deployment.environment", "service.namespace"} {
		if value, present := row.Resource[key]; present {
			resource[key] = value
		}
	}
	return continuationJSON352([]any{row.Trace, row.ID, row.Parent, row.Service, resource})
}
func reportedRawKeys353(rows []reportedRaw353) []string {
	keys := []string{}
	for _, row := range rows {
		keys = append(keys, reportedRawKey353(row))
	}
	slices.Sort(keys)
	return keys
}
func reportedFixtureRaw353(row reportedSpan353) reportedRaw353 {
	return reportedRaw353{Trace: row.Trace, ID: row.ID, Parent: row.Parent, Service: row.Service, Resource: row.Resource}
}
func reportedResultIDs353(result exploration.Result) ([]string, error) {
	ids := []string{}
	for _, raw := range result.Rows {
		var row struct{ SpanID string }
		if json.Unmarshal(raw, &row) != nil || row.SpanID == "" {
			return nil, fmt.Errorf("invalid observed child ID")
		}
		ids = append(ids, row.SpanID)
	}
	slices.Sort(ids)
	return ids, nil
}
func reportedSourceCount353(calls []serviceContinuationCall352) int {
	n := 0
	for _, call := range calls {
		if call.Method == "POST" && slices.Contains([]string{"/api/discovery", "/api/explore", "/api/explore/page", "/api/record-aggregates", "/api/inventory", "/api/captures"}, call.Path) {
			n++
		}
	}
	return n
}

func reportedObservedCase353(l reportedLedger353, c reportedCase353) (string, error) {
	phase := "discover"
	f := l.Fixture
	mode, count, widerCount := 0, 0, 0
	childNS := &identity.Namespace{Mode: "missing"}
	switch c.Key {
	case "current":
		mode, count, widerCount = 1, 2, 7
	case "effective":
		mode, count, widerCount = 2, 3, 10
	case "empty":
		mode, count, widerCount = 3, 1, 3
		childNS = &identity.Namespace{Mode: "value", Value: f.EmptyNS}
	case "unreported":
		mode, count, widerCount = 4, 1, 3
	default:
		return phase, fmt.Errorf("unknown policy case")
	}
	selector := identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: f.Environment}}
	if mode == 1 {
		selector.Policy = "current-only"
	} else if mode == 3 {
		selector.Scope = identity.EnvironmentScope{Mode: "empty"}
	} else if mode == 4 {
		selector.Scope = identity.EnvironmentScope{Mode: "unreported"}
	}
	if c.Scope.Validate() != nil || c.Scope.Environment != "" || c.Scope.EnvironmentSelector == nil || continuationJSON352(*c.Scope.EnvironmentSelector) != continuationJSON352(selector) || c.Scope.FromMS != f.FromMS || c.Scope.ToMS != f.ToMS || c.Scope.Namespace != nil || c.Scope.Service != nil || continuationJSON352(c.Evidence.Question) != continuationJSON352(c.Scope) || c.Evidence.Identity != "otel-service-namespace/v1" || !serviceContinuationHash352(c.Evidence.Basis) {
		return phase, fmt.Errorf("discovery changed original checked scope")
	}
	if _, err := time.Parse(time.RFC3339Nano, c.Evidence.CapturedAt); err != nil {
		return phase, fmt.Errorf("missing discovery clock")
	}
	if c.InventoryCapture.Inventory == nil || c.InventoryCapture.Receipt.Question.Kind != "inventory" || continuationJSON352(c.InventoryCapture.Inventory.Question) != continuationJSON352(c.Scope) {
		return phase, fmt.Errorf("original captured inventory lost fixed scope")
	}
	oracleFixture := reportedFixture353{Mode: mode, Environment: f.Environment, Parent: f.Parent, Child: f.Child, ParentNS: &identity.Namespace{Mode: "value", Value: f.ParentNS}, ChildNS: childNS, Scope: c.Scope}
	for _, row := range f.Rows {
		ns := &identity.Namespace{Mode: "missing"}
		if row.NS != nil {
			ns = &identity.Namespace{Mode: "value", Value: *row.NS}
		}
		oracleFixture.Rows = append(oracleFixture.Rows, reportedSpan353{At: row.At, Service: row.Service, Trace: row.Trace, ID: row.ID, Parent: row.ParentID, Resource: reportedResource353(row.Env, ns), Tag: row.Tag})
	}
	expectedRaw, expectedWide := []reportedRaw353{}, []reportedRaw353{}
	for _, row := range oracleFixture.Rows {
		if row.At < f.FromMS || row.At >= f.ToMS || !continuationEnvironment352(mode, f.Environment, row.Resource) {
			continue
		}
		raw := reportedFixtureRaw353(row)
		expectedRaw = append(expectedRaw, raw)
		if row.Service == f.Child && continuationNamespace352(childNS, row.Resource) {
			expectedWide = append(expectedWide, raw)
		}
	}
	selected := reportedOracle353(oracleFixture)
	expectedExact := []reportedRaw353{}
	for _, row := range selected {
		expectedExact = append(expectedExact, reportedFixtureRaw353(row))
	}
	if len(selected) != count || len(expectedWide) != widerCount || !slices.Equal(reportedRawKeys353(c.Raw), reportedRawKeys353(expectedRaw)) || !slices.Equal(reportedRawKeys353(c.Wider), reportedRawKeys353(expectedWide)) || !slices.Equal(reportedRawKeys353(c.Oracle), reportedRawKeys353(expectedExact)) {
		return phase, fmt.Errorf("raw source universe/parent owner set disagrees with independent seeded oracle")
	}
	phase = "investigate"
	want := exploration.Query{Signal: "spans", EnvironmentSelector: &selector, FromMS: f.FromMS, ToMS: f.ToMS, Limit: 100, Summary: "span-duration", SpanStatus: true, GroupBy: []exploration.Field{{Kind: "column", Key: "name"}}, ReportedEdge: &exploration.ReportedEdge{Parent: exploration.EdgeService{Service: f.Parent, Namespace: oracleFixture.ParentNS.Clone()}, Child: exploration.EdgeService{Service: f.Child, Namespace: childNS.Clone()}}}
	if c.Question.Validate() != nil || continuationJSON352(c.Question) != continuationJSON352(want) || continuationJSON352(c.Result.Question) != continuationJSON352(want) || c.Result.Count != strconv.Itoa(count) || len(c.Result.Rows) != count || c.Result.HasMore || !serviceContinuationHash352(c.Result.Basis) {
		return phase, fmt.Errorf("exact question/result changed whole descriptor/profile or record count")
	}
	ids, err := reportedResultIDs353(c.Result)
	if err != nil {
		return phase, err
	}
	expectedIDs := []string{}
	for _, row := range selected {
		expectedIDs = append(expectedIDs, row.ID)
	}
	slices.Sort(expectedIDs)
	if !slices.Equal(ids, expectedIDs) {
		return phase, fmt.Errorf("exact returned IDs disagree with raw-owner oracle")
	}
	for _, raw := range c.Result.Rows {
		var row serviceContinuationRow352
		if json.Unmarshal(raw, &row) != nil || row.Service != f.Child || !continuationEnvironment352(mode, f.Environment, row.Resource) || !continuationNamespace352(childNS, row.Resource) {
			return phase, fmt.Errorf("exact child row violates selected scope")
		}
		at, err := strconv.ParseInt(row.TimeNS, 10, 64)
		if err != nil || at < f.FromMS*1000000 || at >= f.ToMS*1000000 {
			return phase, fmt.Errorf("exact child outside half-open window")
		}
	}
	if c.Result.Duration == nil || c.Result.Duration.Count != strconv.Itoa(count) || c.Result.Duration.SumNS != strconv.Itoa(count*2000000) || c.Result.Duration.P95NS != "2000000" || c.Result.SpanStatus == nil || c.Result.SpanStatus.Count != strconv.Itoa(count) || c.Result.SpanStatus.OK != strconv.Itoa(count) {
		return phase, fmt.Errorf("exact duration/status profile was not observed")
	}
	runClock, err := time.Parse(time.RFC3339Nano, c.Result.CapturedAt)
	if err != nil {
		return phase, fmt.Errorf("missing exact read clock")
	}
	phase = "capture"
	if c.CaptureRequest.Question.Kind != "explore" || c.CaptureRequest.Question.Explore == nil || continuationJSON352(*c.CaptureRequest.Question.Explore) != continuationJSON352(want) || c.Captured.Explore == nil || c.Captured.Receipt.Question.Explore == nil || continuationJSON352(*c.Captured.Receipt.Question.Explore) != continuationJSON352(want) || continuationJSON352(c.Captured.Explore.Question) != continuationJSON352(want) || c.Captured.Explore.Count != strconv.Itoa(count) || continuationJSON352(c.Captured.Explore.Duration) != continuationJSON352(c.Result.Duration) {
		return phase, fmt.Errorf("capture lost exact checked question/profile")
	}
	capturedIDs, err := reportedResultIDs353(*c.Captured.Explore)
	if err != nil || !slices.Equal(capturedIDs, expectedIDs) {
		return phase, fmt.Errorf("captured exact IDs changed")
	}
	captureClock, err := time.Parse(time.RFC3339Nano, c.Captured.Explore.CapturedAt)
	if err != nil || !captureClock.After(runClock) {
		return phase, fmt.Errorf("explicit capture reread lacks independent clock")
	}
	custody, err := time.Parse(time.RFC3339Nano, c.Captured.Receipt.CapturedAt)
	if err != nil || custody.Before(captureClock) {
		return phase, fmt.Errorf("custody and source clocks disagree")
	}
	var stored exploration.SavedCapture
	if c.CaptureBytesBefore == "" || c.CaptureBytesAfter != c.CaptureBytesBefore || json.Unmarshal([]byte(c.CaptureBytesBefore), &stored) != nil || stored.ID != c.Captured.Receipt.ID || stored.Explore == nil || continuationJSON352(stored.Question) != continuationJSON352(c.Captured.Receipt.Question) || continuationJSON352(stored.Explore.Question) != continuationJSON352(want) {
		return phase, fmt.Errorf("stored immutable exact capture changed on reopen")
	}
	names := []string{"reopen-inventory", "display-graph", "stage-exact", "run-exact", "capture-exact", "reopen-exact", "return-services"}
	if len(c.Facts) != len(names) {
		return phase, fmt.Errorf("missing observed operation boundaries")
	}
	previous := -1
	for i, fact := range c.Facts {
		if fact.Name != names[i] || fact.CallsBefore < 0 || fact.CallsAfter < fact.CallsBefore || fact.CallsAfter > len(l.Calls) || fact.CallsBefore < previous || fact.Before != reportedSourceCount353(l.Calls[:fact.CallsBefore]) || fact.After != reportedSourceCount353(l.Calls[:fact.CallsAfter]) {
			return phase, fmt.Errorf("raw operation accounting mismatch: %s", fact.Name)
		}
		previous = fact.CallsAfter
		delta := 0
		if fact.Name == "run-exact" || fact.Name == "capture-exact" {
			delta = 1
		}
		if fact.After-fact.Before != delta {
			return phase, fmt.Errorf("hidden/missing source read: %s", fact.Name)
		}
	}
	return "", nil
}

func TestQuintReportedEdgeObserved353(t *testing.T) {
	name := os.Getenv("GOO_REPORTED_EDGE_353_REPORT")
	if name == "" {
		t.Skip("native exact-edge report opt-in")
	}
	raw, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	var ledger reportedLedger353
	if json.Unmarshal(raw, &ledger) != nil {
		t.Fatal("invalid observed exact-edge report")
	}
	if ledger.Version != "goo.reportedEdge353/1" || len(ledger.Cases) != 4 || len(ledger.Errors) != 0 || len(ledger.Fixture.Rows) == 0 {
		t.Fatal("incomplete observed campaign")
	}
	seen := map[string]bool{}
	for _, c := range ledger.Cases {
		if seen[c.Key] {
			t.Fatal("duplicate policy case")
		}
		seen[c.Key] = true
	}
	check := func(label string, c reportedCase353, want bool) {
		t.Helper()
		failed, admission := reportedObservedCase353(ledger, c)
		if (admission == nil) != want {
			t.Fatalf("%s independent admission: %v", label, admission)
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink, qobs.WithActor("edge353-"+c.Key), qobs.WithProcess("actual-browser"))
		for _, action := range []string{"discover", "investigate", "capture"} {
			coherent := action != failed
			rec.Record(context.Background(), action, nil, "obs.coherent", coherent)
			if !coherent {
				break
			}
		}
		checker, err := validate.NewChecker("model/reported_edge.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if err != nil {
			t.Fatal(err)
		}
		report, err := checker.Check(context.Background(), sink.Steps())
		if err != nil {
			t.Fatal(err)
		}
		artifact("reported-edge353-"+label+".txt", []byte(report.Summary()))
		if report.OK() != want {
			t.Fatalf("%s admission=%v: %s", label, admission, report.Summary())
		}
	}
	for _, c := range ledger.Cases {
		check("actual-"+c.Key, c, true)
	}
	current := ledger.Cases[0]
	scalar := ledger.Scalar.Question
	expectedScalar := current.Question
	expectedScalar.Environment, expectedScalar.EnvironmentSelector = ledger.Fixture.Environment, nil
	ids, _ := reportedResultIDs353(ledger.Scalar.Result)
	expectedIDs, _ := reportedResultIDs353(current.Result)
	if scalar.Validate() != nil || continuationJSON352(scalar) != continuationJSON352(expectedScalar) || continuationJSON352(ledger.Scalar.Result.Question) != continuationJSON352(expectedScalar) || !slices.Equal(ids, expectedIDs) || ledger.Scalar.Result.Count != current.Result.Count {
		t.Fatal("scalar/current-only observed inequivalence")
	}
	mutants := []func(*reportedCase353){
		func(c *reportedCase353) { c.Scope.EnvironmentSelector.Policy = "effective" },
		func(c *reportedCase353) { c.Question.ReportedEdge.Parent.Namespace.Value = "foreign" },
		func(c *reportedCase353) { c.Result.Question.EnvironmentSelector = nil },
		func(c *reportedCase353) { c.Result.Question.Summary = "" },
		func(c *reportedCase353) { c.Oracle = c.Oracle[:0] },
		func(c *reportedCase353) { c.Raw = c.Raw[:len(c.Raw)-1] },
		func(c *reportedCase353) { c.Captured.Receipt.Question.Explore.ReportedEdge = nil },
		func(c *reportedCase353) { c.CaptureBytesAfter += " " },
		func(c *reportedCase353) { c.Facts[2].After++ },
		func(c *reportedCase353) { c.Facts = c.Facts[:6] },
	}
	for i, mutate := range mutants {
		var fresh reportedLedger353
		json.Unmarshal(raw, &fresh)
		changed := fresh.Cases[0]
		mutate(&changed)
		check(fmt.Sprintf("negative-%d", i), changed, false)
	}
}
