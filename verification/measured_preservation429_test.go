package verification

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type measured429Counters struct{ Source, Writes, Metadata, SQL int }
type measured429Step struct {
	Action        string
	Before, After measured429Counters
	Calls         []pair428Call
	SQL           []pair428SQL
}
type measured429Metric struct {
	State       string
	Width       int
	FirstRecord *struct {
		DocumentY, Height float64
		Count             int
	}
	VisibleActions struct {
		Count, EnabledCount int
		Names               []string
	}
	ModalHeight, SaveY, RecentGap *float64
	Overflow                      bool
}
type measured429Read struct {
	Request  exploration.Query
	Response exploration.Result
}
type measured429Report struct {
	Version, Phase, Status           string
	Cleanup                          bool
	QuickFilterClosedCallbackRefused bool
	Errors                           []string
	Fixture                          struct {
		Token, Environment         string
		Clock                      int64
		Selection, Baseline, Fixed exploration.Query
		Counter                    measurements.Question
		View                       workbench.ExplorationView
		Activity                   json.RawMessage
	}
	Raw               []pair428Raw
	Calls, ActorCalls []pair428Call
	SQL, ActorSQL     []pair428SQL
	Steps             []measured429Step
	Metrics           []measured429Metric
	Results           struct {
		Fixed         measured429Read
		ActivityFixed measured429Read
		Comparison    struct {
			Selection measured429Read
			Baseline  struct {
				Request  exploration.Query
				Response exploration.RecordAggregatesResult
			}
			Comparison struct {
				Request  exploration.ComparisonRequest
				Response exploration.CohortComparison
			}
		}
		ExplicitCapture exploration.CapturePage
	}
	Held struct {
		Before, After          json.RawMessage
		DOMSame, FocusReturned bool
		RowMutations           int
	}
	Counter struct {
		Request  measurements.Question
		Response measurements.Result
		Text     string
	}
	Bombadil struct {
		Snapshots int
		Completed struct {
			Phase, Source, Forbidden, Requests, BeforeRetired int
			HiddenQuestion, HiddenDestinationNote             bool
			FocusReturned                                     bool
		}
	}
}
type measured429Comparison struct {
	Version       string
	Before, After measured429Report
}

var measured429Actions = []string{"stage-activity", "start-activity", "find-and-stage-fixed", "run-activity-fixed", "stage-exact-fixed", "run-fixed", "measure-fixed-applied", "measure-fixed-edit", "return-to-paused-activity", "open-saved-definition", "exact-recent-recall", "run-paired-comparison", "measure-paired-applied", "preview-capture-only", "repeated-local-preserve", "confirm-pair-capture", "run-native-counter"}
var measured429States = []string{"fixed-applied", "fixed-edit", "recent", "pair-applied", "capture-only"}

// The six raw OTLP observations are independent of the rendered page and limits.
func measured429Count(r measured429Report, q exploration.Query) int {
	count := 0
	for _, raw := range r.Raw {
		ns, e := strconv.ParseInt(raw.TimestampNS, 10, 64)
		if e != nil || ns < q.FromMS*1000000 || ns >= q.ToMS*1000000 {
			continue
		}
		values := map[exploration.Field]string{{Kind: "column", Key: "body"}: raw.Body, {Kind: "column", Key: "service"}: raw.ResourceAttributes["service.name"]}
		for k, v := range raw.ResourceAttributes {
			values[exploration.Field{Kind: "resource", Key: k}] = v
		}
		for k, v := range raw.LogAttributes {
			values[exploration.Field{Kind: "attribute", Key: k}] = v
		}
		if exploration.Match(q.Where, values) {
			count++
		}
	}
	return count
}
func measured429Clock(s string) bool { _, e := time.Parse(time.RFC3339Nano, s); return e == nil }
func measured429Scope(r measured429Report) string {
	if len(r.Raw) != 6 || r.Fixture.Token == "" || r.Fixture.Clock <= 0 || r.Fixture.View.Validate() != nil || len(r.Fixture.View.Queries) != 1 || !pair428Same(r.Fixture.View.Queries[0], r.Fixture.Selection) {
		return "fixture/view"
	}
	ids := map[string]bool{}
	for _, raw := range r.Raw {
		id := raw.LogAttributes["work.id"]
		if id == "" || ids[id] || raw.ResourceAttributes["measure429.bank"] != r.Fixture.Token || raw.ResourceAttributes["deployment.environment.name"] != r.Fixture.Environment {
			return "raw membership"
		}
		ids[id] = true
	}
	if !pair428Same(r.Results.Fixed.Request, r.Fixture.Fixed) || r.Results.ActivityFixed.Response.Count != "3" || !pair428Same(r.Results.ActivityFixed.Request, r.Results.ActivityFixed.Response.Question) {
		return "separate fixed/Activity scope"
	}
	c := r.Results.Comparison
	for _, q := range []exploration.Query{r.Fixture.Selection, r.Fixture.Baseline, r.Results.Fixed.Request, c.Selection.Request, c.Baseline.Request} {
		if q.Validate() != nil || !q.SameEnvironment(r.Fixture.Selection) || q.FromMS != r.Fixture.Selection.FromMS || q.ToMS != r.Fixture.Selection.ToMS || measured429Count(r, q) != 3 {
			return "fixed scope/membership"
		}
	}
	for _, read := range []measured429Read{r.Results.Fixed, c.Selection} {
		ids := []string{}
		for _, raw := range read.Response.Rows {
			var row struct{ Attributes map[string]string }
			if json.Unmarshal(raw, &row) != nil {
				return "record membership shape"
			}
			ids = append(ids, row.Attributes["work.id"])
		}
		want := []string{}
		for _, raw := range r.Raw {
			values := map[exploration.Field]string{{Kind: "column", Key: "body"}: raw.Body}
			for k, v := range raw.ResourceAttributes {
				values[exploration.Field{Kind: "resource", Key: k}] = v
			}
			for k, v := range raw.LogAttributes {
				values[exploration.Field{Kind: "attribute", Key: k}] = v
			}
			if exploration.Match(read.Request.Where, values) {
				want = append(want, raw.LogAttributes["work.id"])
			}
		}
		slices.Sort(ids)
		slices.Sort(want)
		if !slices.Equal(ids, want) {
			return "actual record membership"
		}
		if !pair428Same(read.Request, read.Response.Question) || read.Response.Count != "3" || read.Response.State != "observed" || len(read.Response.Rows) != 3 || !recall427ID.MatchString(read.Response.Basis) || !measured429Clock(read.Response.CapturedAt) {
			return "applied read"
		}
	}
	if !pair428Same(c.Selection.Request, r.Fixture.Selection) || !pair428Same(c.Baseline.Request, r.Fixture.Baseline) || !pair428Same(c.Baseline.Request, c.Baseline.Response.Question) || c.Baseline.Response.Count != "3" || c.Baseline.Response.State != "observed" || !measured429Clock(c.Baseline.Response.CapturedAt) {
		return "comparison operands"
	}
	cohort, e := exploration.CohortFromResult(&c.Selection.Response)
	if e != nil || !pair428Same(cohort, c.Comparison.Request.Selection) {
		return "selection cohort"
	}
	base := c.Comparison.Request.Baseline
	if !pair428Same(base.Question, r.Fixture.Baseline) || base.Count != "3" || base.Basis != c.Baseline.Response.Basis || base.CapturedAt != c.Baseline.Response.CapturedAt {
		return "baseline cohort"
	}
	decision, e := exploration.CompareCohorts(c.Comparison.Request)
	if e != nil || !pair428Same(decision, c.Comparison.Response) {
		return "cohort decision"
	}
	p := r.Results.ExplicitCapture
	want := exploration.CaptureQuestion{Kind: "comparison", Comparison: &exploration.ComparisonQuestion{Baseline: r.Fixture.Baseline, Selection: r.Fixture.Selection}}
	if !pair428Same(p.Receipt.Question, want.Normalized()) || p.Comparison == nil || p.Comparison.Request.Baseline.Count != "3" || p.Comparison.Request.Selection.Count != "3" || !recall427ID.MatchString(p.Receipt.ID) || !measured429Clock(p.Receipt.CapturedAt) {
		return "explicit pair/order"
	}
	return ""
}
func measured429Key(m measured429Metric) string { return fmt.Sprintf("%s/%d", m.State, m.Width) }
func measured429Geometry(r measured429Report) string {
	if len(r.Metrics) != len(measured429States)*3 {
		return "measurement coverage"
	}
	seen := map[string]bool{}
	finite := func(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }
	for _, m := range r.Metrics {
		key := measured429Key(m)
		if seen[key] || !slices.Contains(measured429States, m.State) || !slices.Contains([]int{320, 390, 1440}, m.Width) || m.Overflow || m.VisibleActions.Count < 1 || m.VisibleActions.Count != len(m.VisibleActions.Names) || m.VisibleActions.EnabledCount < 0 || m.VisibleActions.EnabledCount > m.VisibleActions.Count {
			return "measurement identity/actions/overflow"
		}
		seen[key] = true
		if m.FirstRecord != nil && (!finite(m.FirstRecord.DocumentY) || !finite(m.FirstRecord.Height) || m.FirstRecord.DocumentY < 0 || m.FirstRecord.Height <= 0 || m.FirstRecord.Count != 3) {
			return "first-record measurement"
		}
		if m.State == "capture-only" && (m.ModalHeight == nil || m.SaveY == nil || *m.ModalHeight <= 0) {
			return "missing modal measurement"
		}
		for _, v := range []*float64{m.ModalHeight, m.SaveY, m.RecentGap} {
			if v != nil && !finite(*v) {
				return "nonfinite layout"
			}
		}
	}
	return ""
}
func measured429Metadata(call pair428Call) bool {
	return call.Method == "POST" && call.Path == "/api/measurement-packages/plan" || pair428Metadata(call) || call.Method == "GET" && strings.HasPrefix(call.Path, "/api/activity-views/") && recall427ID.MatchString(strings.TrimPrefix(call.Path, "/api/activity-views/"))
}
func measured429Body[T any](raw json.RawMessage, want T) bool {
	var actual T
	return json.Unmarshal(raw, &actual) == nil && pair428Same(actual, want)
}
func measured429Authority(r measured429Report) string {
	seenSQL := map[int]bool{}
	account := func(calls []pair428SQL) bool {
		for _, call := range calls {
			if call.Index < 0 || call.Index >= len(r.SQL) || seenSQL[call.Index] || !pair428Same(call, r.SQL[call.Index]) || !call.Closed || call.Status != 200 || call.Fault != nil {
				return false
			}
			seenSQL[call.Index] = true
		}
		return true
	}
	seen := map[string]bool{}
	for _, step := range r.Steps {
		if seen[step.Action] || !slices.Contains(measured429Actions, step.Action) {
			return "unknown/duplicate action"
		}
		seen[step.Action] = true
		if !account(step.SQL) || step.After.SQL-step.Before.SQL != len(step.SQL) {
			return "actual SQL ownership"
		}
		source, writes, metadata := 0, 0, 0
		for _, call := range step.Calls {
			if call.Method == "POST" && call.Path == "/api/measurement-packages/plan" && step.Action != "run-native-counter" {
				return "plan outside counter context"
			}
			if call.Method == "GET" {
				metadata++
			}
			if call.Method == "POST" && call.Path == "/api/compare" && (step.Action != "run-paired-comparison" || !measured429Body(call.Body, r.Results.Comparison.Comparison.Request)) {
				return "comparison request mismatch"
			}
			if measured429Metadata(call) {
				continue
			}
			switch call.Method + " " + call.Path {
			case "POST /api/tail":
				if step.Action != "start-activity" {
					return "tail outside Start"
				}
				source++
			case "POST /api/explore":
				if step.Action != "run-activity-fixed" && step.Action != "run-fixed" && step.Action != "run-paired-comparison" {
					return "Explore outside Run"
				}
				want := r.Results.Fixed.Request
				if step.Action == "run-activity-fixed" {
					want = r.Results.ActivityFixed.Request
				}
				if step.Action == "run-paired-comparison" {
					want = r.Results.Comparison.Selection.Request
				}
				if !measured429Body(call.Body, want) {
					return "Explore request mismatch"
				}
				source++
			case "POST /api/record-aggregates":
				if step.Action != "run-paired-comparison" {
					return "aggregate outside Run"
				}
				if !measured429Body(call.Body, r.Results.Comparison.Baseline.Request) {
					return "aggregate request mismatch"
				}
				source++
			case "POST /api/measurements":
				if step.Action != "run-native-counter" {
					return "measurement outside Run"
				}
				if !measured429Body(call.Body, r.Counter.Request) {
					return "measurement request mismatch"
				}
				source++
			case "POST /api/captures":
				if step.Action != "confirm-pair-capture" {
					return "capture outside Confirm"
				}
				var body struct{ Question exploration.CaptureQuestion }
				if json.Unmarshal(call.Body, &body) != nil || !pair428Same(body.Question.Normalized(), r.Results.ExplicitCapture.Receipt.Question) {
					return "capture request mismatch"
				}
				writes++
			default:
				return "unknown API authority " + call.Method + " " + call.Path
			}
		}
		if step.After.Source-step.Before.Source != source || step.After.Writes-step.Before.Writes != writes || step.After.Metadata-step.Before.Metadata != metadata {
			return "actual request counters"
		}
		if source == 0 && writes == 0 && len(step.SQL) != 0 {
			return "hidden local SQL"
		}
	}
	if len(seen) != len(measured429Actions) || !account(r.ActorSQL) || len(seenSQL) != len(r.SQL) {
		return "incomplete observed authority"
	}
	for _, call := range r.Calls {
		if measured429Metadata(call) {
			continue
		}
		found := false
		for _, step := range r.Steps {
			for _, owned := range step.Calls {
				if call.Method == owned.Method && call.Path == owned.Path && pair428Same(call.Body, owned.Body) {
					found = true
				}
			}
		}
		if !found {
			return "unowned API call"
		}
	}
	return ""
}

// Compare only immutable questions. Rolling Activity is independently exercised
// per phase and must not be made comparable by dropping its actual server bounds.
func measured429SameFixture(a, b measured429Report) bool {
	return a.Fixture.Token == b.Fixture.Token && a.Fixture.Environment == b.Fixture.Environment && a.Fixture.Clock == b.Fixture.Clock && pair428Same(a.Fixture.Selection, b.Fixture.Selection) && pair428Same(a.Fixture.Baseline, b.Fixture.Baseline) && pair428Same(a.Fixture.Fixed, b.Fixture.Fixed) && pair428Same(a.Fixture.View, b.Fixture.View) && pair428Same(a.Fixture.Counter, b.Fixture.Counter)
}
func measured429Counter(r measured429Report) bool {
	c := r.Counter
	if !pair428Same(c.Request, r.Fixture.Counter) || !pair428Same(c.Request, c.Response.Question) || !measured429Clock(c.Response.ReadAt) || c.Response.Counters == nil || c.Response.Counters.ValueDomain != "nonnegative-decimal" || len(c.Response.Counters.Series) != 1 || len(c.Response.Rows) != 1 || c.Request.Definition.Source.Unit != "s" || c.Request.Definition.CounterDomain != "nonnegative-decimal" {
		return false
	}
	for _, interval := range c.Response.Counters.Series[0].Intervals {
		from, e := strconv.ParseInt(interval.FromNS, 10, 64)
		to, f := strconv.ParseInt(interval.ToNS, 10, 64)
		if e == nil && f == nil && to-from == 3600000000 && interval.State == "ok" && interval.Delta == "0.001" && interval.ExactPerSecond == "1/3600" && interval.PerSecond != nil && math.Abs(*interval.PerSecond-1.0/3600) < 1e-16 {
			return true
		}
	}
	return false
}
func measured429Actor(r measured429Report) bool {
	if r.Phase == "before" {
		return len(r.ActorSQL) == 0 && len(r.ActorCalls) == 0
	}
	b := r.Bombadil.Completed
	if r.Bombadil.Snapshots < 1 || b.Phase != 6 || b.Source != 2 || b.Forbidden != 0 || b.Requests != b.BeforeRetired || !b.FocusReturned || !b.HiddenQuestion || !b.HiddenDestinationNote || len(r.ActorSQL) != 6 {
		return false
	}
	source := 0
	for _, call := range r.ActorCalls {
		if measured429Metadata(call) {
			continue
		}
		if call.Method != "POST" || !slices.Contains([]string{"/api/explore", "/api/record-aggregates"}, call.Path) {
			return false
		}
		source++
	}
	if source != 2 {
		return false
	}
	// Every actor query must match an original selected or baseline plan; there is
	// no generic source allowance for a local dialog or retired callback.
	type expectedSQL struct {
		text       string
		parameters map[string]string
	}
	expected := []expectedSQL{}
	for i, q := range []exploration.Query{r.Fixture.Selection, r.Fixture.Baseline} {
		p, e := exploration.Compile(q)
		if e != nil {
			return false
		}
		queries := []string{p.Count, p.Rows, p.Groups, p.Fields}
		if i == 1 {
			queries = []string{p.Count, p.Groups}
		}
		for _, sql := range queries {
			expected = append(expected, expectedSQL{sql, p.Parameters})
		}
	}
	used := map[int]bool{}
	for _, call := range r.ActorSQL {
		found := -1
		for i, want := range expected {
			if used[i] || want.text != call.SQL {
				continue
			}
			same := true
			for k, v := range want.parameters {
				same = same && call.Parameters["param_"+k] == v
			}
			if same {
				found = i
				break
			}
		}
		if found < 0 {
			return false
		}
		used[found] = true
	}
	return true
}
func measured429Admission(r measured429Comparison) string {
	if r.Version != "goo.measuredPreservation429.comparison/1" || r.Before.Phase != "before" || r.After.Phase != "after" || !measured429SameFixture(r.Before, r.After) || !pair428Same(r.Before.Raw, r.After.Raw) {
		return "different fixture/scope"
	}
	for _, phase := range []measured429Report{r.Before, r.After} {
		if phase.Version != "goo.measuredPreservation429/1" || phase.Status != "passed" || !phase.Cleanup || len(phase.Errors) > 0 || !phase.Held.DOMSame || !phase.Held.FocusReturned || phase.Held.RowMutations != 0 || len(phase.Held.Before) == 0 || !pair428Same(phase.Held.Before, phase.Held.After) {
			return "incomplete/changed held visit"
		}
		if !measured429Counter(phase) || !measured429Actor(phase) {
			return phase.Phase + ": counter/actor witness"
		}
		for _, check := range []func(measured429Report) string{measured429Scope, measured429Geometry, measured429Authority} {
			if reason := check(phase); reason != "" {
				return phase.Phase + ": " + reason
			}
		}
	}
	if !strings.Contains(r.Before.Counter.Text, "0.0002777777777777778") || strings.Contains(r.After.Counter.Text, "0.000 s/s") || !strings.Contains(r.After.Counter.Text, "e-4") && !strings.Contains(r.After.Counter.Text, "0.000278") {
		return "positive counter display"
	}
	if !pair428Same(r.Before.Counter.Response.Rows, r.After.Counter.Response.Rows) || !pair428Same(r.Before.Counter.Response.Counters, r.After.Counter.Response.Counters) {
		return "different counter samples/intervals"
	}
	a, b := r.Before.Results, r.After.Results
	if !pair428Same(a.Fixed.Request, b.Fixed.Request) || a.Fixed.Response.Basis != b.Fixed.Response.Basis || !pair428Same(a.Fixed.Response.Rows, b.Fixed.Response.Rows) || a.Comparison.Selection.Response.Basis != b.Comparison.Selection.Response.Basis || a.Comparison.Baseline.Response.Basis != b.Comparison.Baseline.Response.Basis {
		return "different applied evidence"
	}
	before := map[string]measured429Metric{}
	for _, m := range r.Before.Metrics {
		before[measured429Key(m)] = m
	}
	improved := false
	for _, after := range r.After.Metrics {
		prev := before[measured429Key(after)]
		if (prev.FirstRecord == nil) != (after.FirstRecord == nil) || prev.FirstRecord != nil && prev.FirstRecord.Count != after.FirstRecord.Count {
			return "different measured record state"
		}
		improved = improved || after.VisibleActions.Count < prev.VisibleActions.Count || prev.FirstRecord != nil && after.FirstRecord.DocumentY < prev.FirstRecord.DocumentY || prev.ModalHeight != nil && after.ModalHeight != nil && *after.ModalHeight < *prev.ModalHeight
	}
	if !r.After.QuickFilterClosedCallbackRefused {
		return "closed quick-filter callback not refused"
	}
	if !improved {
		return "no measured consolidation"
	}
	return ""
}
func TestQuintMeasuredPreservation429Observed(t *testing.T) {
	file := os.Getenv("GOO_MEASURED_PRESERVATION_429_REPORT")
	if file == "" {
		t.Skip("requires completed actual before/after measured preservation reports")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var report measured429Comparison
	if e = json.Unmarshal(raw, &report); e != nil {
		t.Fatal(e)
	}
	if reason := measured429Admission(report); reason != "" {
		for _, phase := range []measured429Report{report.Before, report.After} {
			t.Logf("%s status=%s cleanup=%v steps=%d metrics=%d SQL=%d scope=%q geometry=%q authority=%q counter=%v actor=%v", phase.Phase, phase.Status, phase.Cleanup, len(phase.Steps), len(phase.Metrics), len(phase.SQL), measured429Scope(phase), measured429Geometry(phase), measured429Authority(phase), measured429Counter(phase), measured429Actor(phase))
		}
		t.Fatal("actual measured consolidation admission:", reason)
	}
	checker, e := validate.NewChecker("model/population_drilldown.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	check := func(r measured429Comparison) bool {
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		reads := 0
		coherent := measured429Admission(r) == ""
		for _, phase := range []measured429Report{r.Before, r.After} {
			for _, step := range phase.Steps {
				n := step.After.SQL - step.Before.SQL
				if n <= 0 {
					rec.Record(t.Context(), "history", nil, "obs.reads", reads, "obs.depth", 0, "obs.coherent", coherent)
				} else {
					for range n {
						reads++
						rec.Record(t.Context(), "read", nil, "obs.reads", reads, "obs.depth", 0, "obs.coherent", coherent)
					}
				}
			}
		}
		verdict, e := checker.Check(t.Context(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return verdict.OK()
	}
	if !check(report) {
		t.Fatal("actual source/local trace rejected")
	}
	for i := 0; i < 8; i++ {
		var bad measured429Comparison
		encoded, _ := json.Marshal(report)
		if e := json.Unmarshal(encoded, &bad); e != nil {
			t.Fatal(e)
		}
		switch i {
		case 0:
			bad.After.Fixture.Selection.Where = nil
		case 1:
			for j := range bad.After.Metrics {
				if bad.After.Metrics[j].FirstRecord != nil {
					bad.After.Metrics[j].FirstRecord.Count++
					break
				}
			}
		case 2:
			bad.After.Steps[0].After.SQL++
		case 3:
			bad.After.Steps[0].Calls = append(bad.After.Steps[0].Calls, pair428Call{Method: "POST", Path: "/api/trace-search"})
		case 4:
			bad.After.Held.After = json.RawMessage(`{"changed":true}`)
		case 5:
			bad.After.Results.Fixed.Response.Count = "2"
		case 6:
			bad.After.Results.ExplicitCapture.Receipt.Question.Comparison.Baseline = bad.After.Fixture.Selection
		case 7:
			bad.After.QuickFilterClosedCallbackRefused = false
		}
		if check(bad) {
			t.Fatal("measurement/scope/authority mutant accepted", i)
		}
	}
	t.Logf("admitted %d matched measurements and %d actual before/after steps; rejected eight measurement/scope/authority controls", len(report.After.Metrics), len(report.Before.Steps)+len(report.After.Steps))
}
