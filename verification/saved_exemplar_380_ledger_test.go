package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"regexp"
	"slices"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type savedExemplarCase380 struct {
	Question                            measurements.Question
	Capture                             exploration.CapturePage
	Stream                              metrics.Stream
	ExemplarEndNS                       string
	ExemplarIndex                       int
	ContextPlan                         exploration.HistogramExemplarContext
	Citation                            json.RawMessage
	StoredBytesBefore, StoredBytesAfter string
	SourceBefore, SourceAfter           json.RawMessage
	OldActions                          struct{ Before, After json.RawMessage }
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
	Facts []savedNativeFact364
}
type savedExemplarCall380 struct{ Method, Path string }

type savedExemplarLedger380 struct {
	Version   string
	Cases     []savedExemplarCase380
	RawTraces []exemplarTrace377
	Calls     []savedExemplarCall380
	Errors    []string
}

func savedExemplarAdmission380(r savedExemplarLedger380, c savedExemplarCase380) error {
	fail := func(message string) error { return fmt.Errorf("saved exemplar380: %s", message) }
	measurement := c.Capture.Measurement
	if len(r.Errors) != 0 || measurement == nil || measurement.Histograms == nil || measurement.Basis == "" || measurement.ReadAt == "" || !peLiveCaptureHash(c.Capture) {
		return fail("authoritative capture unavailable")
	}
	if savedJSON360(measurement.Question) != savedJSON360(c.Question) || savedJSON360(c.Capture.Receipt.Question.Measurement) != savedJSON360(c.Question) {
		return fail("capture question changed")
	}
	if c.StoredBytesBefore == "" || c.StoredBytesBefore != c.StoredBytesAfter || len(c.SourceBefore) == 0 || len(c.SourceAfter) == 0 || savedJSON360(c.SourceBefore) != savedJSON360(c.SourceAfter) || len(c.OldActions.Before) == 0 || savedJSON360(c.OldActions.Before) != savedJSON360(c.OldActions.After) {
		return fail("custody or held DOM changed")
	}

	var stored exploration.SavedCapture
	if json.Unmarshal([]byte(c.StoredBytesBefore), &stored) != nil {
		return fail("stored capture bytes are invalid")
	}
	expectedStored := exploration.SavedCapture{CaptureReceipt: c.Capture.Receipt, Measurement: measurement}
	if savedJSON360(stored) != savedJSON360(expectedStored) {
		return fail("stored custody differs from displayed capture")
	}
	var held struct {
		SameParent, SamePanel, Open bool
		EndNS, SourceRow, Hash      string
	}
	if json.Unmarshal(c.SourceBefore, &held) != nil || !held.SameParent || !held.SamePanel || !held.Open || held.EndNS != c.ExemplarEndNS || held.Hash != "#capture="+c.Capture.Receipt.ID {
		return fail("actual held capture/export ownership unavailable")
	}
	plan, err := exploration.PlanHistogramExemplarContext(measurement, c.Stream, c.ExemplarEndNS, c.ExemplarIndex)
	if err != nil {
		return fail("retained export admission: " + err.Error())
	}
	if held.SourceRow != strconv.Itoa(plan.SourceRow) {
		return fail("actual selected source occurrence differs")
	}
	if savedJSON360(plan) != savedJSON360(c.ContextPlan) {
		return fail("reported occurrence or planned context changed")
	}
	reference := map[string]any{"sourceRow": plan.SourceRow, "exportStartNS": plan.ExportStartNS, "exportEndNS": plan.ExportEndNS, "exemplarIndex": plan.ExemplarIndex, "timeNS": plan.TimeNS, "traceId": plan.TraceID}
	if plan.SpanID != "" {
		reference["spanId"] = plan.SpanID
	}
	expectedCitation := map[string]any{"receipt": c.Capture.Receipt, "measurement": measurement.Question, "stream": c.Stream, "fromNS": plan.FromNS, "toNS": plan.ToNS, "basis": measurement.Basis, "readAt": measurement.ReadAt, "relation": "reported-exemplar", "reference": reference}
	if savedJSON360(c.Citation) != savedJSON360(expectedCitation) {
		return fail("capture/export/sample citation changed")
	}
	if !nhLiveSameQuery(plan.Question, c.Run.Request) || !nhLiveSameQuery(plan.Question, c.Refused.Request) || !nhLiveSameQuery(plan.Question, c.Pending.Request) || c.Run.Status != 200 || c.Refused.Status != 502 || c.Pending.Status != 200 || !c.Pending.Aborted || c.Pending.AbortsAfter <= c.Pending.AbortsBefore {
		return fail("fresh read, refusal or cancellation changed")
	}
	var result exploration.Result
	if json.Unmarshal(c.Run.Response, &result) != nil || !nhLiveSameQuery(result.Question, plan.Question) {
		return fail("fresh result scope changed")
	}
	var refused struct{ Refusal struct{ Kind string } }
	if json.Unmarshal(c.Refused.Response, &refused) != nil || refused.Refusal.Kind != "unavailable" {
		return fail("structured refused read changed")
	}
	lo, ok := new(big.Int).SetString(plan.FromNS, 10)
	if !ok {
		return fail("invalid original lower bound")
	}
	hi, ok := new(big.Int).SetString(plan.ToNS, 10)
	if !ok {
		return fail("invalid original upper bound")
	}
	namespace := c.Question.Definition.Source.Namespace
	if namespace == nil {
		namespace = &identity.Namespace{Mode: "missing"}
		if value, present := c.Stream.Resource[identity.NamespaceKey]; present {
			namespace = &identity.Namespace{Mode: "value", Value: value}
		}
	}
	ids := []string{}
	services := map[string]bool{}
	for _, row := range r.RawTraces {
		at, valid := new(big.Int).SetString(row.TimestampNS, 10)
		if !valid {
			return fail("invalid independent trace time")
		}
		value, present := row.Resource[identity.NamespaceKey]
		nsMatches := namespace.Mode == "missing" && !present || namespace.Mode == "value" && present && value == namespace.Value
		if row.TraceID == plan.TraceID && nsMatches && exemplarEnvironment377(c.Question.Definition, row.Resource) && at.Cmp(lo) >= 0 && at.Cmp(hi) < 0 {
			ids = append(ids, row.SpanID)
			services[row.Service] = true
		}
	}
	if len(r.RawTraces) != 4 || len(ids) != 2 || len(services) != 2 {
		return fail("independent distributed trace/decoy membership changed")
	}
	slices.Sort(ids)
	want := []string{}
	for _, row := range c.Oracle.Rows {
		want = append(want, row.ID)
	}
	slices.Sort(want)
	got, idsErr := reportedResultIDs353(result)
	if idsErr != nil || !slices.Equal(ids, want) || !slices.Equal(ids, got) || c.Oracle.Count != strconv.Itoa(len(ids)) || result.Count != c.Oracle.Count {
		return fail("fresh result differs from independent scoped membership")
	}
	names := []string{"open-select", "inert-reference", "stage", "run", "return", "old-actions", "stage-refused", "run-refused", "return-refused", "stage-pending-run", "return-pending"}
	reads := []int{0, 0, 0, 1, 0, 0, 0, 1, 0, 1, 0}
	if len(c.Facts) != len(names) {
		return fail("actual ownership facts missing")
	}
	for index, f := range c.Facts {
		if f.Name != names[index] || f.After.Source-f.Before.Source != reads[index] || f.After.SQL < f.Before.SQL || f.After.Calls < f.Before.Calls || f.After.Store < f.Before.Store {
			return fail("observed lifecycle budget changed")
		}
		if index > 0 && f.After.Store != f.Before.Store {
			return fail("stored capture reread on continuation")
		}
		if index > 0 && reads[index] == 0 && f.After.SQL != f.Before.SQL {
			return fail("passive stage/return performed source I/O")
		}
		if f.Before.Calls < 0 || f.After.Calls > len(r.Calls) {
			return fail("actual API call slice missing")
		}
		calls := r.Calls[f.Before.Calls:f.After.Calls]
		if index == 0 {
			if f.After.Store-f.Before.Store != 1 {
				return fail("capture open did not read exactly one stored capture")
			}
			for _, call := range calls {
				if call.Method != "GET" {
					return fail("capture open made a non-read request")
				}
			}
		} else if f.Name == "return-pending" {
			// A pending fresh query needs an upstream cancellation request. This
			// is not a telemetry/store read. The report also measures its abort.
			if len(calls) != 1 || calls[0].Method != "POST" || !regexp.MustCompile(`^/api/queries/[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}/cancel$`).MatchString(calls[0].Path) {
				return fail("pending return cancellation changed")
			}
		} else {
			expected := map[savedExemplarCall380]int{}
			if f.Name == "stage" || f.Name == "stage-refused" || f.Name == "stage-pending-run" {
				expected[savedExemplarCall380{"GET", "/api/recipes"}] = 1
				expected[savedExemplarCall380{"GET", "/api/retained-evidence/status"}] = 1
			}
			if reads[index] > 0 {
				expected[savedExemplarCall380{"POST", "/api/explore"}] = 1
			}
			for _, call := range calls {
				if expected[call] != 1 {
					return fail("unexpected or duplicate ancillary API call")
				}
				delete(expected, call)
			}
			if len(expected) != 0 {
				return fail("recorded expected API call missing")
			}
		}
		if reads[index] > 0 && f.After.SQL <= f.Before.SQL {
			return fail("fresh source call lacks actual source read")
		}
		if index > 0 && f.Before != c.Facts[index-1].After {
			return fail("actual lifecycle counters discontinuous")
		}
	}
	if c.Facts[len(c.Facts)-1].After.Calls != len(r.Calls) {
		return fail("unclassified trailing API calls")
	}
	return nil
}

// Replay the recorded macrosteps, not a generated success trace. The combined
// pending stage/read uses its recorded before/after source counts as in 364.
func savedExemplarReplay380(t *testing.T, label string, facts []savedNativeFact364) bool {
	t.Helper()
	if len(facts) == 0 {
		t.Fatal("empty actual saved exemplar facts")
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	baseline, depth := facts[0].Before.Source, 0
	record := func(action string, source int) {
		rec.Record(context.Background(), action, nil, "obs.reads", source-baseline, "obs.depth", depth)
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
		case "open-select", "inert-reference", "old-actions":
			record("history", f.After.Source)
		default:
			t.Fatal("unknown recorded saved exemplar fact", f.Name)
		}
	}
	encoded, err := json.Marshal(sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	if err = artifact("saved-exemplar380-"+label+"-steps.json", encoded); err != nil {
		t.Fatal(err)
	}
	checker, err := validate.NewChecker("model/saved_exemplar_380.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	report, err := checker.Check(t.Context(), sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	if err = artifact("saved-exemplar380-"+label+"-model.txt", []byte(report.Summary())); err != nil {
		t.Fatal(err)
	}
	return report.OK()
}

func TestQuintSavedExemplarObserved380(t *testing.T) {
	path := os.Getenv("GOO_SAVED_EXEMPLAR_380_REPORT")
	if path == "" {
		t.Skip("actual saved exemplar report opt-in")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var report savedExemplarLedger380
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if report.Version != "goo.savedExemplars380/1" || len(report.Cases) != 1 {
		t.Fatal("missing actual saved exemplar case")
	}
	c := report.Cases[0]
	if err = savedExemplarAdmission380(report, c); err != nil {
		t.Fatal(err)
	}
	clone := func() savedExemplarCase380 {
		var value savedExemplarLedger380
		if err := json.Unmarshal(raw, &value); err != nil {
			t.Fatal(err)
		}
		return value.Cases[0]
	}
	mutations := []struct {
		name   string
		change func(*savedExemplarCase380)
	}{
		{"trace", func(c *savedExemplarCase380) { c.Run.Request.Where.All[0].Value = "ffffffffffffffffffffffffffffffff" }},
		{"namespace", func(c *savedExemplarCase380) { c.Run.Request.Where.All[1].Value = "foreign" }},
		{"occurrence", func(c *savedExemplarCase380) { c.ContextPlan.SourceRow++ }},
		{"sample", func(c *savedExemplarCase380) { c.ExemplarIndex++ }},
		{"custody", func(c *savedExemplarCase380) { c.StoredBytesAfter += "x" }},
		{"foreign-custody", func(c *savedExemplarCase380) { c.StoredBytesBefore = `{}`; c.StoredBytesAfter = `{}` }},
		{"held", func(c *savedExemplarCase380) { c.SourceAfter = json.RawMessage(`{"changed":true}`) }},
		{"citation", func(c *savedExemplarCase380) { c.Citation = json.RawMessage(`{"foreign":true}`) }},
		{"passive-source", func(c *savedExemplarCase380) { c.Facts[2].After.Source++ }},
		{"passive-store", func(c *savedExemplarCase380) { c.Facts[4].After.Store++ }},
	}
	for _, mutation := range mutations {
		bad := clone()
		mutation.change(&bad)
		if savedExemplarAdmission380(report, bad) == nil {
			t.Fatal("admitted actual-ledger mutation", mutation.name)
		}
	}

	for _, name := range []string{"ancillary-duplicate", "ancillary-addition", "foreign-cancel"} {
		var changed savedExemplarLedger380
		if err = json.Unmarshal(raw, &changed); err != nil {
			t.Fatal(err)
		}
		switch name {
		case "ancillary-duplicate":
			changed.Calls[changed.Cases[0].Facts[2].Before.Calls+1] = savedExemplarCall380{"GET", "/api/recipes"}
		case "ancillary-addition":
			changed.Calls = append(changed.Calls, savedExemplarCall380{"GET", "/api/recipes"})
		case "foreign-cancel":
			changed.Calls[len(changed.Calls)-1].Path = "/api/queries/not-a-query/cancel"
		}
		if savedExemplarAdmission380(changed, changed.Cases[0]) == nil {
			t.Fatal("admitted actual API call mutation", name)
		}
	}
	if !savedExemplarReplay380(t, "actual", c.Facts) {
		t.Fatal("actual saved exemplar lifecycle failed model")
	}
	counter := slices.Clone(c.Facts)
	counter[2].After.Source++
	if savedExemplarReplay380(t, "passive-source", counter) {
		t.Fatal("model accepted passive source read")
	}
	order := slices.Clone(c.Facts)
	order[2], order[4] = order[4], order[2]
	if savedExemplarReplay380(t, "return-before-stage", order) {
		t.Fatal("model accepted return before stage")
	}
}
