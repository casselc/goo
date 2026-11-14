package verification

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type pair428Call struct {
	Method, Path string
	Body         json.RawMessage
	Status       int
	Response     json.RawMessage
}
type pair428SQL struct {
	Index      int
	SQL        string
	Parameters map[string]string
	Fault      *string
	Status     int
	Closed     bool
}
type pair428Counters struct{ Source, Writes, Captures, SQL int }
type pair428Step struct {
	Action        string
	Before, After pair428Counters
	Calls         []pair428Call
	SQL           []pair428SQL
}
type pair428Raw struct {
	Body                              string
	TimestampNS                       string `json:"timestamp_ns"`
	SeverityNumber                    int
	ResourceAttributes, LogAttributes map[string]string
}
type pair428Report struct {
	Version, Status, Token string
	Clock                  int64
	Cleanup                bool
	Errors                 []string
	Baseline, Selection    exploration.Query
	RawBefore, RawAfter    []pair428Raw
	Capture                struct {
		Request struct {
			Title    string
			Question exploration.CaptureQuestion
		}
		Response exploration.CapturePage
	}
	Held struct {
		Baseline, Selection struct {
			Request  exploration.Query
			Response exploration.Result
		}
		Comparison struct {
			Request  exploration.ComparisonRequest
			Response exploration.CohortComparison
		}
	}
	Stored                 struct{ Pair, Baseline, Selection exploration.SavedCapture }
	ReopenedPair, ColdPair exploration.CapturePage
	Children               struct{ Baseline, Selection exploration.CapturePage }
	Attachment             struct {
		Document  workbench.Document
		Reference string
	}
	Refusal struct {
		Status int
		Body   json.RawMessage
	}
	DestinationFocus, KeyboardConfirmReachable, DepartedFocusRetained bool
	PendingAttachment                                                 struct {
		Reference string
		Retained  bool
	}
	DepartedCapture exploration.CapturePage
	Calls           []pair428Call
	SQL             []pair428SQL
	ActorSQL        []pair428SQL
	Steps           []pair428Step
	Bombadil        struct {
		Snapshots int
		Completed struct {
			Phase, Forbidden, Source, Requests, BeforeRetired int
			FocusReturned                                     bool
		}
	}
}

func pair428Same(a, b any) bool {
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	var ax, by any
	if json.Unmarshal(x, &ax) != nil || json.Unmarshal(y, &by) != nil {
		return false
	}
	return reflect.DeepEqual(ax, by)
}
func pair428Clone(r pair428Report) pair428Report {
	var out pair428Report
	raw, _ := json.Marshal(r)
	if err := json.Unmarshal(raw, &out); err != nil {
		panic(err)
	}
	return out
}
func pair428Hash(c exploration.SavedCapture) string {
	c.ID = ""
	raw, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
func pair428Ordered(earlier, later string) bool {
	a, e := time.Parse(time.RFC3339Nano, earlier)
	b, f := time.Parse(time.RFC3339Nano, later)
	return e == nil && f == nil && !a.After(b)
}
func pair428Question(q exploration.Query, token string, clock int64, selected bool) bool {
	if q.Validate() != nil || q.Signal != "logs" || q.Environment != "" || q.FromMS != clock-60000 || q.ToMS != clock+1000 || q.Offset != 0 || q.Limit != 100 || !reflect.DeepEqual(q.GroupBy, []exploration.Field{{Kind: "column", Key: "service"}}) || q.EnvironmentSelector == nil {
		return false
	}
	env := q.EnvironmentSelector
	if env.Policy != "effective" || env.Scope.Mode != "named" || env.Scope.Value != "pair428-"+token {
		return false
	}
	op := "ne"
	if selected {
		op = "eq"
	}
	wanted := map[string]any{"all": []any{map[string]any{"field": map[string]any{"kind": "resource", "key": "pair428.bank"}, "op": "eq", "value": token}, map[string]any{"field": map[string]any{"kind": "attribute", "key": "cohort"}, "op": op, "value": "canary"}}}
	return pair428Same(q.Where, wanted)
}
func pair428Facts(raw []pair428Raw, token string, clock int64) (map[string]pair428Raw, int, int, bool) {
	ids := map[string]pair428Raw{}
	a, b := 0, 0
	for _, r := range raw {
		id := r.LogAttributes["work.id"]
		cohort := r.LogAttributes["cohort"]
		ns, err := strconv.ParseInt(r.TimestampNS, 10, 64)
		if id == "" || ids[id].Body != "" || err != nil || ns < (clock-60000)*1000000 || ns >= (clock+1000)*1000000 || r.ResourceAttributes["pair428.bank"] != token || r.ResourceAttributes["service.name"] != "pair-worker" || r.ResourceAttributes["deployment.environment.name"] != "pair428-"+token || r.SeverityNumber != 9 || r.Body != "pair "+cohort+" "+id {
			return nil, 0, 0, false
		}
		switch cohort {
		case "canary":
			a++
		case "stable":
			b++
		default:
			return nil, 0, 0, false
		}
		ids[id] = r
	}
	return ids, a, b, true
}
func pair428Capture(r pair428Report) string {
	if !pair428Question(r.Baseline, r.Token, r.Clock, false) || !pair428Question(r.Selection, r.Token, r.Clock, true) {
		return "configured operands changed"
	}
	before, sa, ba, ok := pair428Facts(r.RawBefore, r.Token, r.Clock)
	if !ok || sa != 2 || ba != 3 {
		return "original source facts"
	}
	after, sb, bb, ok := pair428Facts(r.RawAfter, r.Token, r.Clock)
	if !ok || sb != 3 || bb != 3 {
		return "late source facts"
	}
	for id, record := range before {
		if !pair428Same(record, after[id]) {
			return "late arrival changed predecessor facts"
		}
	}
	q := exploration.CaptureQuestion{Kind: "comparison", Comparison: &exploration.ComparisonQuestion{Baseline: r.Baseline, Selection: r.Selection}}
	if q.Validate() != nil || !pair428Same(r.Capture.Request.Question, q) || !pair428Same(r.Capture.Response.Receipt.Question, q.Normalized()) {
		return "capture request order or original scope"
	}
	pair := r.Stored.Pair
	api := r.Capture.Response
	if pair.Question.Kind != "comparison" || pair.Comparison == nil || pair428Hash(pair) != pair.ID || pair.ID != api.Receipt.ID || !pair428Same(pair.CaptureReceipt, api.Receipt) || !pair428Same(pair.Comparison, api.Comparison) || api.Explore != nil || api.Inventory != nil {
		return "pair content or API identity"
	}
	expected, err := exploration.CompareCohorts(pair.Comparison.Request)
	if err != nil || !pair428Same(expected, pair.Comparison.Result) {
		return "stored comparison decision"
	}
	for _, side := range []struct {
		stored exploration.SavedCapture
		page   exploration.CapturePage
		q      exploration.Query
		cohort exploration.Cohort
		id     string
		held   exploration.Result
		count  int
	}{{r.Stored.Baseline, r.Children.Baseline, r.Baseline, pair.Comparison.Request.Baseline, pair.Comparison.BaselineID, r.Held.Baseline.Response, bb}, {r.Stored.Selection, r.Children.Selection, r.Selection, pair.Comparison.Request.Selection, pair.Comparison.SelectionID, r.Held.Selection.Response, sb}} {
		c := side.stored
		if c.ID != side.id || pair428Hash(c) != c.ID || c.Explore == nil || c.Question.Kind != "explore" || !pair428Same(c.Question, exploration.CaptureQuestion{Kind: "explore", Explore: &side.q}.Normalized()) || c.Explore.Count != strconv.Itoa(side.count) || c.Explore.State != "observed" || len(c.Explore.Rows) != side.count {
			return "ordered child content or full count"
		}
		// Check each small stored bank against independent raw OTLP source facts.
		gotIDs := map[string]bool{}
		for _, raw := range c.Explore.Rows {
			var row struct {
				Body, TimestampNS string
				Attributes        map[string]string
				SeverityNumber    int
			}
			var decoded map[string]json.RawMessage
			if json.Unmarshal(raw, &decoded) != nil {
				return "stored row shape"
			}
			if json.Unmarshal(decoded["Body"], &row.Body) != nil || json.Unmarshal(decoded["timestamp_ns"], &row.TimestampNS) != nil || json.Unmarshal(decoded["Attributes"], &row.Attributes) != nil || json.Unmarshal(decoded["SeverityNumber"], &row.SeverityNumber) != nil {
				return "stored row shape"
			}
			id := row.Attributes["work.id"]
			fact, exists := after[id]
			selected := side.q.Where.All[1].Op == "eq"
			if !exists || gotIDs[id] || row.SeverityNumber != fact.SeverityNumber || row.Body != fact.Body || row.TimestampNS != fact.TimestampNS || (fact.LogAttributes["cohort"] == "canary") != selected {
				return "stored bank differs from independent raw membership"
			}
			gotIDs[id] = true
		}
		page, err := c.Page("")
		if err != nil || !pair428Same(page, side.page) {
			return "child API differs from stored evidence"
		}
		cohort, err := exploration.CohortFromResult(c.Explore)
		if err != nil || !pair428Same(cohort, side.cohort) {
			return "pair cohort differs from child"
		}
		if !pair428Ordered(side.held.CapturedAt, c.Explore.CapturedAt) || !pair428Ordered(c.Explore.CapturedAt, c.CapturedAt) || !pair428Ordered(c.CapturedAt, pair.CapturedAt) {
			return "fresh source and receipt clock order"
		}
	}
	if !pair428Same(api, r.ReopenedPair) || !pair428Same(api, r.ColdPair) {
		return "source-free reopen changed pair"
	}
	if r.Held.Selection.Response.Count != "2" || r.Held.Baseline.Response.Count != "3" || !pair428Same(r.Held.Selection.Request, r.Selection) || !pair428Same(r.Held.Baseline.Request, r.Baseline) {
		return "held comparison or scope"
	}
	heldWant, heldErr := exploration.CompareCohorts(r.Held.Comparison.Request)
	if heldErr != nil || !pair428Same(heldWant, r.Held.Comparison.Response) || r.Held.Comparison.Request.Baseline.Count != "3" || r.Held.Comparison.Request.Selection.Count != "2" || !pair428Same(r.Held.Comparison.Request.Baseline.Question, r.Baseline) || !pair428Same(r.Held.Comparison.Request.Selection.Question, r.Selection) || r.Held.Comparison.Request.Baseline.Basis != r.Held.Baseline.Response.Basis || r.Held.Comparison.Request.Selection.Basis != r.Held.Selection.Response.Basis {
		return "held comparison differs from original source cohorts"
	}
	if pair.Comparison.Result.ViewComparison == nil || pair.Comparison.Result.ViewComparison.Baseline.Completeness != nil || pair.Comparison.Result.ViewComparison.Selection.Completeness != nil {
		return "invented source coverage"
	}
	if r.Attachment.Reference != pair.ID || r.Attachment.Document.DocumentInput.Validate() != nil {
		return "capture attachment"
	}
	attached := false
	for _, card := range r.Attachment.Document.Cards {
		if card.Kind == "capture" && card.Reference == pair.ID {
			attached = true
		}
		if card.Kind == "comparison" && card.Reference == pair.ID {
			return "general pair relabeled specialized receipt"
		}
	}
	if !attached {
		return "missing parent capture card"
	}
	if r.Refusal.Status < 400 || r.DepartedCapture.Receipt.Question.Kind != "comparison" || r.DepartedCapture.Comparison == nil || !r.PendingAttachment.Retained || r.PendingAttachment.Reference != r.DepartedCapture.Receipt.ID || !pair428Same(r.DepartedCapture.Receipt.Question, q.Normalized()) {
		return "missing source refusal or departed publication"
	}
	return ""
}
func pair428Metadata(call pair428Call) bool {
	if call.Method == "GET" {
		for _, path := range []string{"/api/ui-config", "/api/capabilities", "/api/recipes", "/api/measurement-packages", "/api/telemetry-packages", "/api/retained-evidence/status", "/api/captures", "/api/captures/search", "/api/documents", "/api/exploration-views"} {
			if call.Path == path {
				return true
			}
		}
		for _, prefix := range []string{"/api/captures/", "/api/documents/", "/api/exploration-views/"} {
			if strings.HasPrefix(call.Path, prefix) && recall427ID.MatchString(strings.TrimPrefix(call.Path, prefix)) {
				return true
			}
		}
		if strings.HasPrefix(call.Path, "/api/captures/") && strings.HasSuffix(call.Path, "/receipt") && recall427ID.MatchString(strings.TrimSuffix(strings.TrimPrefix(call.Path, "/api/captures/"), "/receipt")) {
			return true
		}
		return false
	}
	return call.Method == "POST" && (slices.Contains([]string{"/api/compare", "/api/cel/context", "/api/cel/plan", "/api/view-artifacts/encode", "/api/view-artifacts/decode"}, call.Path) || strings.HasPrefix(call.Path, "/api/queries/") && strings.HasSuffix(call.Path, "/cancel"))
}
func pair428Admission(r pair428Report) string {
	if r.Version != "goo.pairedCapture428/1" || r.Status != "passed" || !r.Cleanup || len(r.Errors) > 0 || r.Bombadil.Snapshots < 1 || r.Bombadil.Completed.Phase != 6 || r.Bombadil.Completed.Forbidden != 0 || r.Bombadil.Completed.Source != 2 || !r.Bombadil.Completed.FocusReturned || r.Bombadil.Completed.Requests != r.Bombadil.Completed.BeforeRetired || !r.DestinationFocus || !r.KeyboardConfirmReachable || !r.DepartedFocusRetained {
		return "incomplete actual run"
	}
	if reason := pair428Capture(r); reason != "" {
		return reason
	}
	// Every actual proxy request belongs to an observed main step or the bounded
	// actor's one explicit Run. API-only counters cannot hide server-side reads.
	seenSQL := map[int]bool{}
	account := func(calls []pair428SQL) bool {
		for _, call := range calls {
			if call.Index < 0 || call.Index >= len(r.SQL) || seenSQL[call.Index] || !pair428Same(call, r.SQL[call.Index]) {
				return false
			}
			seenSQL[call.Index] = true
		}
		return true
	}
	for _, step := range r.Steps {
		if !account(step.SQL) {
			return "SQL step ownership"
		}
	}
	if !account(r.ActorSQL) || len(seenSQL) != len(r.SQL) {
		return "unowned actual SQL"
	}
	selectedPlan, err := exploration.Compile(r.Selection)
	if err != nil {
		return "actor selection plan"
	}
	baselinePlan, err := exploration.Compile(r.Baseline)
	if err != nil {
		return "actor baseline plan"
	}
	wantedSQL := []string{selectedPlan.Count, selectedPlan.Rows, selectedPlan.Groups, selectedPlan.Fields, baselinePlan.Count, baselinePlan.Groups}
	if len(r.ActorSQL) != len(wantedSQL) {
		return "actor source read budget"
	}
	for _, call := range r.ActorSQL {
		at := slices.Index(wantedSQL, call.SQL)
		if at < 0 || call.Status != 200 || !call.Closed || call.Fault != nil {
			return "actor unapproved source SQL"
		}
		wantedSQL = slices.Delete(wantedSQL, at, at+1)
	}
	seen := map[string]bool{}
	for _, step := range r.Steps {
		seen[step.Action] = true
		if step.Calls == nil || step.SQL == nil || step.After.SQL-step.Before.SQL != len(step.SQL) {
			return "missing actual step calls/SQL"
		}
		source, writes, captures := 0, 0, 0
		for _, call := range step.Calls {
			if pair428Metadata(call) {
				continue
			}
			if call.Method != "POST" {
				return "unknown step API authority: " + call.Method + " " + call.Path
			}
			switch call.Path {
			case "/api/explore", "/api/record-aggregates":
				if !slices.Contains([]string{"run-held-comparison", "source-run-recovery", "run-before-source-refusal", "run-before-pending-departure"}, step.Action) {
					return "telemetry outside explicit Run"
				}
				source++
			case "/api/captures":
				if !slices.Contains([]string{"confirm-fresh-pair", "source-refusal-retains-current", "pending-departure-does-not-adopt"}, step.Action) {
					return "capture outside explicit confirmation"
				}
				writes++
				captures++
			case "/api/documents":
				if step.Action != "attach-parent-capture-card" {
					return "document write outside explicit Save"
				}
				writes++
			default:
				return "unknown step POST " + call.Path
			}
		}
		if step.After.Source-step.Before.Source != source || step.After.Writes-step.Before.Writes != writes || step.After.Captures-step.Before.Captures != captures {
			return "step counters disagree with real requests"
		}
		if source == 0 && captures == 0 && len(step.SQL) > 0 {
			return "hidden SQL during local action"
		}
		for _, sql := range step.SQL {
			if !sql.Closed || sql.SQL == "" || sql.Status < 200 {
				return "unfinished SQL"
			}
			if step.Action == "source-refusal-retains-current" {
				if sql.Status < 400 || sql.Fault == nil {
					return "source refusal not observed"
				}
			} else if sql.Status != 200 || sql.Fault != nil {
				return "unexpected source fault"
			}
		}
		if step.Action == "confirm-fresh-pair" {
			if captures != 1 || source != 0 || len(step.SQL) != 8 {
				return "fresh pair did not independently read two children"
			}
			i := 0
			for _, child := range []exploration.SavedCapture{r.Stored.Baseline, r.Stored.Selection} {
				for _, want := range []string{child.Explore.Plan.Count, child.Explore.Plan.Rows, child.Explore.Plan.Groups, child.Explore.Plan.Fields} {
					if step.SQL[i].SQL != want {
						return "fresh child SQL plan/order"
					}
					for key, value := range child.Explore.Plan.Parameters {
						if step.SQL[i].Parameters["param_"+key] != value {
							return "fresh child SQL parameter drift"
						}
					}
					i++
				}
			}
		}
	}
	for _, required := range []string{"stage-fixed-comparison", "run-held-comparison", "open-and-cancel-pair-preview", "confirm-fresh-pair", "attach-parent-capture-card", "reopen-stored-pair", "open-stored-baseline", "open-stored-selection", "cold-reopen-pair", "source-refusal-retains-current", "pending-departure-does-not-adopt"} {
		if !seen[required] {
			return "missing observed action " + required
		}
	}
	for _, call := range r.Calls {
		if pair428Metadata(call) {
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
			return "unowned API authority " + call.Path
		}
	}
	return ""
}
func TestQuintPairedCapture428Observed(t *testing.T) {
	file := os.Getenv("GOO_PAIRED_CAPTURE_428_REPORT")
	if file == "" {
		t.Skip("requires complete actual paired-capture browser report")
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var report pair428Report
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if reason := pair428Admission(report); reason != "" {
		t.Fatalf("actual paired-capture admission: %s (capture checks: %s)", reason, pair428Capture(report))
	}
	checker, err := validate.NewChecker("model/population_drilldown.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	check := func(r pair428Report) bool {
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		reads := 0
		coherent := pair428Admission(r) == ""
		for _, step := range r.Steps {
			delta := step.After.SQL - step.Before.SQL
			if delta == 0 {
				rec.Record(t.Context(), "history", nil, "obs.reads", reads, "obs.depth", 0, "obs.coherent", coherent)
			} else {
				for range delta {
					reads++
					rec.Record(t.Context(), "read", nil, "obs.reads", reads, "obs.depth", 0, "obs.coherent", coherent)
				}
			}
		}
		result, err := checker.Check(t.Context(), sink.Steps())
		if err != nil {
			t.Fatal(err)
		}
		return result.OK()
	}
	if !check(report) {
		t.Fatal("actual paired capture trace rejected")
	}
	for mutant := 0; mutant < 8; mutant++ {
		bad := pair428Clone(report)
		switch mutant {
		case 0:
			bad.Capture.Request.Question.Comparison.Baseline = bad.Selection
		case 1:
			bad.Stored.Pair.Comparison.BaselineID = bad.Stored.Pair.Comparison.SelectionID
		case 2:
			bad.Stored.Selection.Explore.Count = "2"
		case 3:
			bad.ColdPair.Receipt.Question.Comparison.Selection.ToMS++
		case 4:
			bad.Steps[0].After.SQL++
		case 5:
			bad.Steps[0].Calls = append(bad.Steps[0].Calls, pair428Call{Method: "POST", Path: "/api/trace-search"})
		case 6:
			bad.Attachment.Reference = bad.Stored.Selection.ID
		case 7:
			bad.RawAfter = bad.RawBefore
		}
		if check(bad) {
			t.Fatal("paired capture mutation accepted", mutant)
		}
	}
	t.Logf("admitted %d actual steps, ordered fresh children and immutable pair; rejected eight semantic/authority mutations", len(report.Steps))
}
