package verification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
)

type recall427Intent struct {
	Kind  string `json:"kind"`
	ID    string `json:"id"`
	Title string `json:"title"`
}
type recall427Entry struct {
	Intent recall427Intent `json:"intent"`
	UsedAt int64           `json:"usedAt"`
	Pinned bool            `json:"pinned"`
}
type recall427Operation struct {
	Action string `json:"action"`
	ID     string `json:"id"`
	Title  string `json:"title"`
}

// Real journal state versus an independent bounded reference model. This does
// not fetch a definition or claim a browser controller's invocation authority.
func TestHegelExactViewRecall427JournalState(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		clock := int64(1000000)
		entries := []recall427Entry{}
		var operations []recall427Operation
		var expected [][]recall427Entry
		bound := func() {
			entries = slices.DeleteFunc(entries, func(e recall427Entry) bool { return !e.Pinned && clock-e.UsedAt > 30*86400000 })
			slices.SortStableFunc(entries, func(a, b recall427Entry) int {
				if a.UsedAt > b.UsedAt {
					return -1
				}
				if a.UsedAt < b.UsedAt {
					return 1
				}
				return 0
			})
			entries = append(append([]recall427Entry{}, filterRecall427(entries, true)...), filterRecall427(entries, false)...)
		}
		for range 48 {
			op := recall427Operation{Action: hegel.Draw(tc, hegel.SampledFrom([]string{"remember", "pin", "unpin", "forget", "clear", "reload", "expire"})), ID: fmt.Sprintf("%064x", hegel.Draw(tc, hegel.Integers(1, 9))), Title: hegel.Draw(tc, hegel.SampledFrom([]string{"Saved logs", "雪 records", "Comparison", "<literal>"}))}
			operations = append(operations, op)
			if op.Action == "expire" {
				clock += 31 * 86400000
			} else {
				clock++
			}
			bound()
			index := slices.IndexFunc(entries, func(e recall427Entry) bool { return e.Intent.ID == op.ID })
			switch op.Action {
			case "remember", "pin":
				pinned := op.Action == "pin" || index >= 0 && entries[index].Pinned
				if index >= 0 {
					entries = slices.Delete(entries, index, index+1)
				}
				entries = append(entries, recall427Entry{recall427Intent{"saved-view", op.ID, op.Title}, clock, pinned})
			case "unpin":
				if index >= 0 {
					entries[index].Pinned = false
				}
			case "forget":
				if index >= 0 {
					entries = slices.Delete(entries, index, index+1)
				}
			case "clear":
				entries = filterRecall427(entries, true)
			}
			bound()
			expected = append(expected, append([]recall427Entry{}, entries...))
		}
		input, _ := json.Marshal(operations)
		cmd := exec.Command("node", "scripts/exact-view-recall427.mjs")
		cmd.Stdin = bytes.NewReader(input)
		output, err := cmd.CombinedOutput()
		if err != nil {
			tc.Fatal(err, string(output))
		}
		var observed struct {
			Snapshots []struct {
				Items  []recall427Entry
				Stored []struct {
					Version int
					Items   []map[string]json.RawMessage
				}
			}
			Unchanged bool
			Forbidden []bool
		}
		if err = json.Unmarshal(output, &observed); err != nil {
			tc.Fatal(err, string(output))
		}
		if len(observed.Snapshots) != len(expected) || !observed.Unchanged || len(observed.Forbidden) != 7 || slices.Contains(observed.Forbidden, false) {
			tc.Fatal("journal reference identity/admission contract failed")
		}
		for i, snapshot := range observed.Snapshots {
			if !reflect.DeepEqual(snapshot.Items, expected[i]) {
				tc.Fatalf("step%d %s: got%+v want%+v", i, operations[i].Action, snapshot.Items, expected[i])
			}
			for _, stored := range snapshot.Stored {
				if stored.Version != 1 {
					tc.Fatal("stored version changed")
				}
				for _, entry := range stored.Items {
					var intent map[string]json.RawMessage
					if err = json.Unmarshal(entry["intent"], &intent); err != nil {
						tc.Fatal(err)
					}
					if len(intent) != 3 || intent["kind"] == nil || intent["id"] == nil || intent["title"] == nil {
						tc.Fatal("stored source or query instead of immutable reference")
					}
				}
			}
		}
	}, options(40)...)
}
func filterRecall427(entries []recall427Entry, pinned bool) []recall427Entry {
	out := []recall427Entry{}
	for _, entry := range entries {
		if entry.Pinned == pinned {
			out = append(out, entry)
		}
	}
	return out
}

// Browser acceptance schema is intentionally supplied by the real journey. Its
// full definitions are re-admitted here rather than trusting a display caption.
type recall427Counters struct{ Source, Writes, Metadata, Posts int }
type recall427Call struct {
	Method, Path string
	Body         json.RawMessage
}
type recall427Step struct {
	Action                    string
	Before, After             recall427Counters
	ExpectedID                string
	ActualDefinition          *workbench.ExplorationView
	Calls                     []recall427Call
	SourceBefore, SourceAfter *string
}
type recall427Report struct {
	Version, Status                      string
	Cleanup                              bool
	KeepFocusRestored, AdoptFocusHeading bool
	Errors                               []string
	Calls                                []recall427Call
	Views                                []workbench.ExplorationView
	Steps                                []recall427Step
	Journal                              json.RawMessage
	Raw                                  []struct {
		Body, SeverityText string
		SeverityNumber     uint64
	}
	Run struct {
		Requests  []recall427Call
		Responses []struct {
			Path    string
			Request json.RawMessage
			Result  json.RawMessage
		}
	}
}

func recall427ViewID(view workbench.ExplorationView) (string, error) {
	if err := view.Validate(); err != nil {
		return "", err
	}
	view.ID = ""
	return workbench.ViewContentID(view)
}

var recall427ID = regexp.MustCompile(`^[a-f0-9]{64}$`)

func recall427Journal(raw json.RawMessage) bool {
	var journal struct {
		Version int
		Items   []struct {
			Intent map[string]json.RawMessage
			UsedAt int64
			Pinned bool
		}
	}
	if json.Unmarshal(raw, &journal) != nil || journal.Version != 1 || len(journal.Items) > 40 {
		return false
	}
	refs := 0
	for _, entry := range journal.Items {
		var kind string
		if json.Unmarshal(entry.Intent["kind"], &kind) != nil {
			return false
		}
		if kind != "saved-view" {
			continue
		}
		refs++
		if len(entry.Intent) != 3 {
			return false
		}
		var id, title string
		if json.Unmarshal(entry.Intent["id"], &id) != nil || json.Unmarshal(entry.Intent["title"], &title) != nil || !recall427ID.MatchString(id) || title == "" || len(title) > 120 || !utf8.ValidString(title) {
			return false
		}
	}
	return refs >= 2
}
func recall427Pure(call recall427Call) bool {
	if call.Method == "GET" {
		if call.Path == "/api/exploration-views" || strings.HasPrefix(call.Path, "/api/exploration-views/") && recall427ID.MatchString(strings.TrimPrefix(call.Path, "/api/exploration-views/")) {
			return true
		}
		return slices.Contains([]string{"/api/ui-config", "/api/capabilities", "/api/recipes", "/api/measurement-packages", "/api/telemetry-packages", "/api/retained-evidence/status"}, call.Path)
	}
	if call.Method != "POST" {
		return false
	}
	return slices.Contains([]string{"/api/cel/context", "/api/cel/plan", "/api/view-artifacts/encode", "/api/view-artifacts/decode", "/api/compare"}, call.Path) || regexp.MustCompile(`^/api/queries/[^/]+/cancel$`).MatchString(call.Path)
}
func recall427Admission(report recall427Report) bool {
	if report.Version != "goo.savedViewRecall427/1" || report.Status != "passed" || !report.Cleanup || len(report.Errors) != 0 || len(report.Views) != 2 || len(report.Steps) < 10 || !recall427Journal(report.Journal) {
		return false
	}
	views := map[string]workbench.ExplorationView{}
	authored, comparison := false, false
	for _, view := range report.Views {
		id, err := recall427ViewID(view)
		if err != nil || id != view.ID || views[id].ID != "" {
			return false
		}
		views[id] = view
		authored = authored || len(view.Authoring) > 0
		if view.Composition != nil {
			for _, child := range view.Composition.Children {
				comparison = comparison || child.Comparison != nil && child.Comparison.BaselineQuery != nil
			}
		}
	}
	if !authored || !comparison {
		return false
	}
	globalSource := 0
	for _, call := range report.Calls {
		if call.Method == "POST" && slices.Contains([]string{"/api/explore", "/api/record-aggregates"}, call.Path) {
			globalSource++
			matched := false
			for _, expected := range report.Run.Requests {
				if call.Path == expected.Path && recall427Same(call.Body, expected.Body) {
					matched = true
				}
			}
			if !matched {
				return false
			}
		} else if !recall427Pure(call) {
			return false
		}
	}
	if globalSource != 2 {
		return false
	}
	completed := map[string]bool{}
	exported := map[string]bool{}
	runSeen := false
	for _, step := range report.Steps {
		completed[step.Action] = true
		if step.Calls == nil || step.After.Source < step.Before.Source || step.After.Writes < step.Before.Writes || step.After.Posts < step.Before.Posts {
			return false
		}
		delta := step.After.Source - step.Before.Source
		if step.Action == "explicit-comparison-run" {
			if delta != 2 || step.After.Writes != step.Before.Writes {
				return false
			}
			runSeen = true
			for _, call := range step.Calls {
				if !(call.Method == "POST" && (call.Path == "/api/explore" || call.Path == "/api/record-aggregates")) && !recall427Pure(call) {
					return false
				}
			}
		} else {
			if delta != 0 || step.After.Writes != step.Before.Writes {
				return false
			}
			for _, call := range step.Calls {
				if !recall427Pure(call) {
					return false
				}
			}
		}
		if step.ActualDefinition != nil {
			id, err := recall427ViewID(*step.ActualDefinition)
			if err != nil || id != step.ExpectedID || views[id].ID == "" {
				return false
			}
			if strings.HasPrefix(step.Action, "export-") {
				exported[id] = true
			}
		}
		if strings.Contains(step.Action, "keep") && step.SourceBefore != nil && (step.SourceAfter == nil || *step.SourceBefore != *step.SourceAfter) {
			return false
		}
	}
	for _, required := range []string{"pin-and-persist", "stage-cel-from-activity", "export-cel-exact", "stage-comparison", "export-two-operands-exact", "filter-and-unpin", "refuse-missing", "refuse-corrupt", "refuse-old-library", "metadata-aba-success", "metadata-aba-error", "stage-valid-after-faults", "dirty-recall-keep", "dirty-recall-adopt", "forget-reference"} {
		if !completed[required] {
			return false
		}
	}
	if !report.KeepFocusRestored || !report.AdoptFocusHeading || !runSeen || len(exported) != 2 {
		return false
	}
	if len(report.Run.Requests) != 2 || len(report.Run.Responses) < 3 || len(report.Raw) != 5 {
		return false
	}
	var pairView workbench.ExplorationView
	var baseline *exploration.Query
	for _, view := range report.Views {
		if view.Composition != nil {
			for _, child := range view.Composition.Children {
				if child.Comparison != nil && child.Comparison.BaselineQuery != nil {
					pairView = view
					baseline = child.Comparison.BaselineQuery
				}
			}
		}
	}
	expected := []exploration.Query{pairView.Queries[0], *baseline}
	seen := map[int]bool{}
	sourceResults := map[int]exploration.Cohort{}
	pureComparisons := 0
	for _, call := range report.Run.Requests {
		if call.Method != "POST" || !slices.Contains([]string{"/api/explore", "/api/record-aggregates"}, call.Path) {
			return false
		}
		var q exploration.Query
		if json.Unmarshal(call.Body, &q) != nil {
			return false
		}
		which := -1
		for j, want := range expected {
			if recall427Same(q, want) {
				which = j
			}
		}
		if which < 0 || seen[which] {
			return false
		}
		seen[which] = true
		var found *struct {
			Path    string
			Request json.RawMessage
			Result  json.RawMessage
		}
		for i := range report.Run.Responses {
			r := &report.Run.Responses[i]
			var rq exploration.Query
			if r.Path == call.Path && json.Unmarshal(r.Request, &rq) == nil && recall427Same(rq, q) {
				found = r
				break
			}
		}
		if found == nil {
			return false
		}
		var result exploration.Cohort
		if json.Unmarshal(found.Result, &result) != nil || !recall427Same(result.Question, q) || result.State != "observed" {
			return false
		}
		if call.Path == "/api/record-aggregates" {
			var aggregate exploration.RecordAggregatesResult
			if json.Unmarshal(found.Result, &aggregate) != nil || aggregate.Version != exploration.RecordAggregatesVersion || exploration.ValidateRecordAggregatesQuery(q) != nil {
				return false
			}
		}
		if q.Where == nil || len(q.Where.All) != 2 || q.Where.All[1].Field == nil || q.Where.All[1].Field.Key != "body" || q.Where.All[1].Op != "contains" {
			return false
		}
		count := 0
		for _, row := range report.Raw {
			if strings.Contains(row.Body, q.Where.All[1].Value) {
				count++
			}
		}
		if result.Count != fmt.Sprint(count) {
			return false
		}
		sourceResults[which] = result
	}
	for _, response := range report.Run.Responses {
		if response.Path != "/api/compare" {
			continue
		}
		pureComparisons++
		var request exploration.ComparisonRequest
		var actual exploration.CohortComparison
		if json.Unmarshal(response.Request, &request) != nil || json.Unmarshal(response.Result, &actual) != nil || !recall427Same(request.Selection, sourceResults[0]) || !recall427Same(request.Baseline, sourceResults[1]) {
			return false
		}
		computed, err := exploration.CompareCohorts(request)
		if err != nil || !recall427Same(computed, actual) {
			return false
		}
	}
	if pureComparisons != 1 {
		return false
	}
	return true
}
func recall427Same(a, b any) bool { return bytes.Equal(mustJSONRecall427(a), mustJSONRecall427(b)) }
func recall427Clone(report recall427Report) recall427Report {
	var out recall427Report
	if err := json.Unmarshal(mustJSONRecall427(report), &out); err != nil {
		panic(err)
	}
	return out
}
func TestQuintExactViewRecall427Observed(t *testing.T) {
	file := os.Getenv("GOO_EXACT_VIEW_RECALL_427_REPORT")
	if file == "" {
		t.Skip("requires complete actual saved-view recall browser report")
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var report recall427Report
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if !recall427Admission(report) {
		t.Logf("report version=%s status=%s cleanup=%v errors=%v views=%d steps=%d journal=%v runrequests=%d responses=%d raw=%d", report.Version, report.Status, report.Cleanup, report.Errors, len(report.Views), len(report.Steps), recall427Journal(report.Journal), len(report.Run.Requests), len(report.Run.Responses), len(report.Raw))
		for _, v := range report.Views {
			id, e := recall427ViewID(v)
			t.Logf("view %s expected=%s actual=%s err=%v", v.Title, v.ID, id, e)
		}
		for _, step := range report.Steps {
			paths := []string{}
			for _, call := range step.Calls {
				paths = append(paths, call.Method+" "+call.Path)
			}
			id := ""
			var e error
			if step.ActualDefinition != nil {
				id, e = recall427ViewID(*step.ActualDefinition)
			}
			t.Logf("step %s source=%d writes=%d calls=%v exact=%s/%s err=%v", step.Action, step.After.Source-step.Before.Source, step.After.Writes-step.Before.Writes, paths, step.ExpectedID, id, e)
		}
		for _, r := range report.Run.Responses {
			var result struct{ State, Count string }
			_ = json.Unmarshal(r.Result, &result)
			t.Logf("response %s state=%s count=%s requestBytes=%d resultBytes=%d", r.Path, result.State, result.Count, len(r.Request), len(r.Result))
		}
		t.Fatal("actual definition/reference/source admission failed")
	}
	checker, err := validate.NewChecker("model/population_drilldown.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	check := func(r recall427Report) bool {
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		reads := 0
		coherent := recall427Admission(r)
		for _, step := range r.Steps {
			delta := step.After.Source - step.Before.Source
			if step.Action == "explicit-comparison-run" {
				for range delta {
					reads++
					rec.Record(context.Background(), "read", nil, "obs.reads", reads, "obs.depth", 0, "obs.coherent", coherent)
				}
			} else {
				rec.Record(context.Background(), "history", nil, "obs.reads", reads, "obs.depth", 0, "obs.coherent", coherent)
			}
		}
		result, err := checker.Check(context.Background(), sink.Steps())
		if err != nil {
			t.Fatal(err)
		}
		return result.OK()
	}
	if !check(report) {
		t.Fatal("actual recall trace rejected")
	}
	for mutation := 0; mutation < 8; mutation++ {
		bad := recall427Clone(report)
		switch mutation {
		case 0:
			bad.Steps[0].After.Source++
		case 1:
			bad.Steps[0].Calls = append(bad.Steps[0].Calls, recall427Call{Method: "POST", Path: "/api/trace-search"})
		case 2:
			bad.Steps[0].After.Writes++
		case 3:
			var journal map[string]any
			json.Unmarshal(bad.Journal, &journal)
			for _, value := range journal["items"].([]any) {
				entry := value.(map[string]any)
				intent := entry["intent"].(map[string]any)
				if intent["kind"] == "saved-view" {
					intent["source"] = "copied private source"
					break
				}
			}
			bad.Journal = mustJSONRecall427(journal)
		case 4:
			for i, step := range bad.Steps {
				if step.ActualDefinition != nil {
					bad.Steps[i].ActualDefinition.Queries[0].ToMS++
					break
				}
			}
		case 5:
			for i, step := range bad.Steps {
				if step.ActualDefinition != nil && len(step.ActualDefinition.Authoring) > 0 {
					for key, a := range bad.Steps[i].ActualDefinition.Authoring {
						a.Source += " // altered"
						bad.Steps[i].ActualDefinition.Authoring[key] = a
					}
					break
				}
			}
		case 6:
			bad.Run.Requests = bad.Run.Requests[:1]
		case 7:
			for _, step := range bad.Steps {
				if step.ActualDefinition != nil && step.ActualDefinition.Composition != nil {
					for _, child := range step.ActualDefinition.Composition.Children {
						if child.Comparison != nil && child.Comparison.BaselineQuery != nil {
							child.Comparison.BaselineQuery.Where = nil
						}
					}
				}
			}
		}
		if check(bad) {
			t.Fatalf("recall mutation %d accepted", mutation)
		}
	}
	t.Logf("admitted %d actual source/local steps, two full definitions/operands; rejected eight semantic/authority mutations", len(report.Steps))
}
func mustJSONRecall427(value any) []byte {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return raw
}
