package verification

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type all437Raw struct {
	Body        string            `json:"body"`
	TimestampNS string            `json:"timestampNS"`
	Resource    map[string]string `json:"resource"`
	WorkID      string            `json:"workID"`
}
type all437Row struct {
	Body                           string
	TimestampNS                    string `json:"timestamp_ns"`
	ResourceAttributes, Attributes map[string]string
}
type all437Fact struct{ Text, Title string }
type all437Visible struct {
	Key, Body   string
	Record      json.RawMessage
	Environment []all437Fact
}
type all437Held struct {
	Keys         []string
	Find         string
	FromMS, ToMS int64
	Count        int
}
type all437Counters struct{ Source, Writes, Metadata, SQL int }
type all437Step struct {
	Action        string
	Before, After all437Counters
	Calls         []pair428Call
	SQL           []pair428SQL
}
type all437Report struct {
	Version, Status, CaptureStore string
	Cleanup                       bool
	Errors                        []string
	Fixture                       struct {
		Token, Database, Service, Term string
		Clock                          int64
		Total                          int
	}
	Raw           []all437Raw
	Calls         []pair428Call
	SQL, ActorSQL []pair428SQL
	Steps         []all437Step
	Tail          struct {
		Batches []struct {
			Selector     exploration.TailSelector
			FromMS, ToMS int64
			ObservedAt   string
			Limited      bool
			RowCount     int
		}
		Rows []all437Visible
	}
	Fixed struct {
		Question exploration.Query
		Response exploration.Result
		RowFacts []all437Visible
	}
	HeldBefore, HeldAfter all437Held
	Fault                 struct {
		Declared string
		Held     all437Held
		Selector exploration.TailSelector
	}
	View, Recalled, Adopted workbench.ExplorationView
	Capture, CaptureReopen  exploration.CapturePage
	Narrow                  []struct {
		Mode     string
		Question exploration.Query
		Response exploration.Result
	}
	Refused []struct {
		Path        string
		Body        json.RawMessage
		Status, SQL int
		Response    json.RawMessage
	}
	Cardinality struct {
		Question exploration.Query
		Status   int
		Response json.RawMessage
	}
	Bombadil struct {
		Snapshots int
		Completed struct {
			Version                                           string
			Phase, Source, Forbidden, Requests, BeforeRetired int
			All, InputDisabled, FocusReturned, Overflow       bool
			LastTo                                            int64
			Selectors                                         []exploration.TailSelector
		}
	}
}

var all437Database = regexp.MustCompile(`^goo437_[a-f0-9]{12}$`)

func all437IsAll(q exploration.Query) bool {
	s, e := q.EnvironmentSelection()
	return e == nil && s.All() && q.Signal == "logs" && q.Environment == "" && q.Validate() == nil
}
func all437CanonicalRow(raw json.RawMessage) (all437Row, error) {
	var row all437Row
	err := json.Unmarshal(raw, &row)
	return row, err
}
func all437InWindow(raw all437Raw, q exploration.Query) bool {
	ns, e := strconv.ParseInt(raw.TimestampNS, 10, 64)
	return e == nil && ns >= q.FromMS*1000000 && ns < q.ToMS*1000000
}
func all437Where(p *exploration.Predicate, raw all437Raw) (bool, bool) {
	if p == nil {
		return true, true
	}
	if p.Numeric != nil {
		return false, false
	}
	if len(p.All) > 0 || len(p.Any) > 0 {
		parts := p.All
		all := true
		if len(p.Any) > 0 {
			parts = p.Any
			all = false
		}
		out := all
		for i := range parts {
			matched, known := all437Where(&parts[i], raw)
			if !known {
				return false, false
			}
			if all {
				out = out && matched
			} else {
				out = out || matched
			}
		}
		return out, true
	}
	if p.Field == nil {
		return false, false
	}
	value, present := "", false
	switch p.Field.Kind {
	case "column":
		if p.Field.Key == "body" {
			value, present = raw.Body, true
		} else if p.Field.Key == "service" {
			value, present = raw.Resource["service.name"]
		}
	case "resource":
		value, present = raw.Resource[p.Field.Key]
	case "attribute":
		if p.Field.Key == "work.id" {
			value, present = raw.WorkID, true
		}
	}
	switch p.Op {
	case "exists":
		return present, true
	case "missing":
		return !present, true
	case "eq":
		return present && value == p.Value, true
	case "ne":
		return present && value != p.Value, true
	case "contains":
		return present && strings.Contains(value, p.Value), true
	default:
		return false, false
	}
}
func all437Oracle(raw []all437Raw, q exploration.Query) ([]all437Raw, bool) {
	selector, e := q.EnvironmentSelection()
	if e != nil {
		return nil, false
	}
	var out []all437Raw
	for _, row := range raw {
		want := true
		if !selector.All() {
			observed := all437Observed(row.Resource)
			want = observed == selector.Scope()
			if selector.CurrentOnly() {
				value, exists := row.Resource[identity.EnvironmentKey]
				want = exists && value == selector.Scope().Value
			}
		}
		matched, known := all437Where(q.Where, row)
		if !known {
			return nil, false
		}
		if want && matched && all437InWindow(row, q) {
			out = append(out, row)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		a, _ := strconv.ParseInt(out[i].TimestampNS, 10, 64)
		b, _ := strconv.ParseInt(out[j].TimestampNS, 10, 64)
		return a < b
	})
	return out, true
}
func all437Rows(raw []all437Raw, q exploration.Query, rows []json.RawMessage, want int) bool {
	expected, ok := all437Oracle(raw, q)
	if !ok || len(rows) != min(want, len(expected)) {
		return false
	}
	seen := map[string]bool{}
	for i, encoded := range rows {
		row, e := all437CanonicalRow(encoded)
		if e != nil || i >= len(expected) || row.Body != expected[i].Body || row.TimestampNS != expected[i].TimestampNS || !pair428Same(row.ResourceAttributes, expected[i].Resource) || row.Attributes["work.id"] != expected[i].WorkID || seen[row.Body] {
			return false
		}
		seen[row.Body] = true
	}
	return true
}
func all437FactMatches(f all437Fact, resource map[string]string) bool {
	observed := all437Observed(resource)
	label := "Unreported environment"
	if observed.Mode == "empty" {
		label = "Empty environment label"
	} else if observed.Mode == "named" {
		encoded, _ := json.Marshal(observed.Value)
		label = "Environment " + string(encoded)
	}
	var current, legacy any
	if value, ok := resource[identity.EnvironmentKey]; ok {
		current = value
	}
	if value, ok := resource[identity.LegacyEnvironmentKey]; ok {
		legacy = value
	}
	c, _ := json.Marshal(current)
	l, _ := json.Marshal(legacy)
	displayMatches := f.Text == label
	if observed.Mode == "named" {
		encoded, _ := json.Marshal(observed.Value)
		displayMatches = displayMatches || f.Text == "Env "+string(encoded)
	}
	return displayMatches && strings.Contains(f.Title, "Current: "+string(c)) && strings.Contains(f.Title, "legacy: "+string(l)) && (current == nil || legacy == nil || current == legacy || strings.Contains(f.Title, "current key wins"))
}
func all437Dataset(r exploration.Result) bool {
	for _, d := range []*exploration.ViewDataset{r.ViewDataset, r.TimeDataset, r.GroupedTimeDataset, r.SummaryDataset, r.HistogramDataset} {
		if d != nil && (d.Quality.Sampling != "unknown" || d.Scope.Signal != "logs" || !pair428Same(d.Scope.Selector, r.Question)) {
			return false
		}
	}
	return true
}
func all437Result(raw []all437Raw, q exploration.Query, r exploration.Result) bool {
	expected, ok := all437Oracle(raw, q)
	_, clockErr := time.Parse(time.RFC3339Nano, r.CapturedAt)
	p, e := exploration.Compile(q)
	return ok && e == nil && clockErr == nil && pair428Same(q, r.Question) && r.Count == strconv.Itoa(len(expected)) && r.State == "observed" && len(r.Basis) == 64 && all437Rows(raw, q, r.Rows, q.Limit) && pair428Same(p, r.Plan) && all437Dataset(r)
}
func all437CallKey(c pair428Call) string {
	var body any
	if len(c.Body) > 0 {
		if json.Unmarshal(c.Body, &body) != nil {
			return "bad"
		}
	}
	canonical, _ := json.Marshal(body)
	return c.Method + " " + c.Path + " " + string(canonical)
}
func all437SourceAction(action, path string) bool {
	switch path {
	case "/api/tail":
		return action == "start-all" || action == "source-access-refusal-keeps-all"
	case "/api/explore":
		return action == "run-fixed-all" || action == "run-before-capture" || strings.HasPrefix(action, "read-all-for-") || strings.HasPrefix(action, "read-reported-")
	case "/api/explore/counted-values":
		return action == "high-cardinality-values-refusal"
	case "/api/captures":
		return action == "capture-all-fixed-question"
	}
	return false
}
func all437SQLMatches(call pair428SQL, p exploration.Plan) bool {
	if call.SQL == "" || !slices.Contains([]string{p.Rows, p.Count, p.Groups, p.Fields, p.TimeSeries, p.GroupedTimeSeries}, call.SQL) {
		return false
	}
	for key, value := range p.Parameters {
		if call.Parameters["param_"+key] != value {
			return false
		}
	}
	return true
}
func all437SQLAllowed(call pair428SQL, plans []exploration.Plan, tail bool) bool {
	if strings.Contains(call.SQL, "system.columns") || strings.Contains(call.SQL, "system.tables") {
		return strings.HasPrefix(strings.TrimSpace(call.SQL), "SELECT ") && strings.Contains(call.SQL, "otel_logs")
	}
	for _, p := range plans {
		if all437SQLMatches(call, p) {
			return true
		}
	}
	if tail {
		to, e := strconv.ParseInt(call.Parameters["param_to"], 10, 64)
		from, f := strconv.ParseInt(call.Parameters["param_from"], 10, 64)
		if e != nil || f != nil || to-from != exploration.TailWindowMS {
			return false
		}
		wire, _ := all437Selector()
		q := exploration.TailSelector{Signal: "logs", EnvironmentSelector: &wire}.Query(to)
		p, e := exploration.Compile(q)
		if e != nil {
			return false
		}
		p.Rows = strings.Replace(p.Rows, " LIMIT 101 OFFSET 0 ", fmt.Sprintf(" LIMIT %d OFFSET 0 ", exploration.TailReadLimit+1), 1)
		p.Rows = strings.Replace(p.Rows, " ORDER BY Timestamp,", " ORDER BY Timestamp DESC,", 1)
		return all437SQLMatches(call, p)
	}
	return false
}

type all437PlannedSQL struct {
	text string
	plan exploration.Plan
}

func all437ReadSQL(p exploration.Plan) []all437PlannedSQL {
	out := []all437PlannedSQL{}
	for _, text := range []string{p.Count, p.Rows, p.Groups, p.Fields, p.TimeSeries, p.GroupedTimeSeries} {
		if text != "" {
			out = append(out, all437PlannedSQL{text, p})
		}
	}
	return out
}
func all437SourceSettings(call pair428SQL) bool {
	return call.Parameters["readonly"] == "1" && call.Parameters["max_execution_time"] == "10" && call.Parameters["max_result_rows"] == "1002" && call.Parameters["max_memory_usage"] == "268435456" && call.Parameters["max_bytes_to_read"] == "268435456"
}

func all437Admission(r all437Report) string {
	if r.Version != "goo.allEnvironments437/1" || r.Status != "passed" || !r.Cleanup || len(r.Errors) > 0 || !all437Database.MatchString(r.Fixture.Database) || r.Fixture.Total != 1005 || len(r.Raw) != 1005 {
		return "incomplete actual bank/report"
	}
	rawByBody := map[string]all437Raw{}
	rawIDs := map[string]bool{}
	classes := map[identity.EnvironmentScope]int{}
	for _, row := range r.Raw {
		if row.Body == "" || !strings.Contains(row.Body, r.Fixture.Term) || row.Resource["all437.bank"] != r.Fixture.Token || row.WorkID == "" || rawIDs[row.WorkID] || rawByBody[row.Body].Body != "" {
			return "raw oracle bank identity"
		}
		rawByBody[row.Body] = row
		rawIDs[row.WorkID] = true
		classes[all437Observed(row.Resource)]++
	}
	if classes[identity.EnvironmentScope{Mode: "empty"}] != 201 || classes[identity.EnvironmentScope{Mode: "unreported"}] != 201 || classes[identity.EnvironmentScope{Mode: "named", Value: "all"}] != 201 || len(classes) != 5 {
		return "raw precedence/coverage"
	}
	q := r.Fixed.Question
	if !all437IsAll(q) || !all437Result(r.Raw, q, r.Fixed.Response) || r.Fixed.Response.Count != "1005" || len(r.Fixed.Response.Rows) != 100 {
		return "fixed All scope/source oracle"
	}
	if len(r.Fixed.RowFacts) != 100 {
		return "fixed environment facts absent"
	}
	for i, f := range r.Fixed.RowFacts {
		row, e := all437CanonicalRow(r.Fixed.Response.Rows[i])
		if e != nil || len(f.Environment) != 1 || !all437FactMatches(f.Environment[0], row.ResourceAttributes) {
			return "fixed environment fact/precedence"
		}
	}
	if len(r.Tail.Rows) != 100 || len(r.Tail.Batches) < 1 {
		return "tail bounded positive witness"
	}
	boundedBatch := false
	for _, b := range r.Tail.Batches {
		boundedBatch = boundedBatch || b.Limited && b.RowCount == exploration.TailReadLimit
		sel, e := b.Selector.Query(b.ToMS).EnvironmentSelection()
		_, clockErr := time.Parse(time.RFC3339Nano, b.ObservedAt)
		if e != nil || !sel.All() || b.ToMS-b.FromMS != exploration.TailWindowMS || b.RowCount < 0 || b.RowCount > exploration.TailReadLimit || clockErr != nil {
			return "tail cap/window/All selector"
		}
	}
	if !boundedBatch {
		return "missing positive capped source batch"
	}
	seenTail := map[string]bool{}
	for _, visible := range r.Tail.Rows {
		row, e := all437CanonicalRow(visible.Record)
		raw, exists := rawByBody[row.Body]
		if e != nil || !exists || visible.Key == "" || seenTail[visible.Key] || row.TimestampNS != raw.TimestampNS || !pair428Same(row.ResourceAttributes, raw.Resource) || len(visible.Environment) != 1 || !all437FactMatches(visible.Environment[0], raw.Resource) {
			return "tail actual row/fact"
		}
		within := false
		for _, batch := range r.Tail.Batches {
			within = within || all437InWindow(raw, batch.Selector.Query(batch.ToMS))
		}
		if !within {
			return "tail row outside declared server window"
		}
		seenTail[visible.Key] = true
	}
	if len(r.HeldBefore.Keys) != 100 || !pair428Same(r.HeldBefore, r.HeldAfter) || r.HeldBefore.Find != r.Fixture.Term || r.Fault.Declared == "" || !pair428Same(r.Fault.Held.Keys, r.HeldBefore.Keys) {
		return "held return/refusal custody"
	}
	faultSel, e := r.Fault.Selector.Query(q.ToMS).EnvironmentSelection()
	if e != nil || !faultSel.All() {
		return "source refusal broadened/replaced scope"
	}
	id, e := recall427ViewID(r.View)
	if e != nil || id != r.View.ID || len(r.View.Queries) != 1 || !pair428Same(r.View.Queries[0], q) || !pair428Same(r.View, r.Recalled) {
		return "definition/Recent lost exact All"
	}
	adopted, saved := r.Adopted, r.View
	adopted.ID, saved.ID = "", ""
	if adopted.Validate() != nil || !pair428Same(adopted, saved) {
		return "actual UI export differs from recalled definition"
	}
	stored, e := (exploration.CaptureStore{Dir: r.CaptureStore}).Load(r.Capture.Receipt.ID)
	if e != nil || stored.Explore == nil || stored.Question.Explore == nil || !all437IsAll(*stored.Question.Explore) || !pair428Same(*stored.Question.Explore, (exploration.CaptureQuestion{Kind: "explore", Explore: &q}).Normalized().Explore) || stored.Explore.Count != "1005" || !all437Rows(r.Raw, *stored.Question.Explore, stored.Explore.Rows, 1000) || r.Capture.RetainedRecords != 1000 || !r.Capture.RecordsLimited || !pair428Same(r.Capture, r.CaptureReopen) || !pair428Ordered(stored.Explore.CapturedAt, stored.CapturedAt) {
		return "capture immutable/full-count/row-bank scope"
	}
	capturePlan, e := exploration.Compile(*stored.Question.Explore)
	if e != nil {
		return "stored question could not compile"
	}
	capturePlan.Rows = strings.Replace(capturePlan.Rows, " LIMIT 101 OFFSET 0 ", fmt.Sprintf(" LIMIT %d OFFSET 0 ", exploration.CaptureRowLimit+1), 1)
	if !pair428Same(capturePlan, stored.Explore.Plan) {
		return "capture source plan differs from exact All question"
	}
	page, e := stored.Page("")
	if e != nil || !pair428Same(page, r.Capture) {
		return "stored page differs from actual reopening"
	}
	wantedNarrow := map[string]identity.EnvironmentScope{"empty": {Mode: "empty"}, "unreported": {Mode: "unreported"}, "named-all": {Mode: "named", Value: "all"}}
	if len(r.Narrow) != len(wantedNarrow) {
		return "narrowing witnesses missing"
	}
	for _, n := range r.Narrow {
		scope, exists := wantedNarrow[n.Mode]
		s, e := n.Question.EnvironmentSelection()
		if !exists || e != nil || s.All() || s.CurrentOnly() || s.Scope() != scope || n.Question.FromMS != q.FromMS || n.Question.ToMS != q.ToMS || !pair428Same(n.Question.Where, q.Where) || !all437Result(r.Raw, n.Question, n.Response) || n.Response.Count != "201" {
			return "reported narrowing changed parent/policy"
		}
		delete(wantedNarrow, n.Mode)
	}
	if len(r.Refused) < 4 {
		return "unsupported route probes missing"
	}
	paths := map[string]bool{}
	for _, f := range r.Refused {
		var reason struct{ Error string }
		json.Unmarshal(f.Response, &reason)
		if f.Status != 400 || f.SQL != 0 || !strings.Contains(strings.ToLower(reason.Error), "all") {
			return "unsupported family read/refusal"
		}
		paths[f.Path] = true
	}
	for _, path := range []string{"/api/explore", "/api/inventory", "/api/trace-search", "/api/metric-discovery"} {
		if !paths[path] {
			return "unsupported route absent"
		}
	}
	var refusal struct {
		Refusal struct{ Kind string }
		Groups  json.RawMessage
	}
	if json.Unmarshal(r.Cardinality.Response, &refusal) != nil || r.Cardinality.Status != 502 || refusal.Refusal.Kind != "budget" || len(refusal.Groups) > 0 || exploration.ValidateCountedValuesQuery(r.Cardinality.Question) != nil || !r.Cardinality.Question.SameEnvironment(q) || r.Cardinality.Question.FromMS != q.FromMS || r.Cardinality.Question.ToMS != q.ToMS || !pair428Same(r.Cardinality.Question.Where, q.Where) || !pair428Same(r.Cardinality.Question.GroupBy, []exploration.Field{{Kind: "attribute", Key: "work.id"}}) {
		return "cardinality refusal lost bounds or returned partial groups"
	}
	c := r.Bombadil.Completed
	if r.Bombadil.Snapshots < 2 || c.Version != "goo.all-environments-actor/1" || c.Phase != 7 || c.Source != 1 || c.Forbidden != 0 || !c.All || !c.InputDisabled || !c.FocusReturned || c.Overflow || c.Requests != c.BeforeRetired || c.LastTo <= 0 || len(c.Selectors) != 1 {
		return "bounded actor incomplete/authority"
	}
	actorSelected, e := c.Selectors[0].Query(c.LastTo).EnvironmentSelection()
	if e != nil || !actorSelected.All() || (c.Selectors[0].Signal != "" && c.Selectors[0].Signal != "logs") {
		return "actor wrong source scope"
	}
	plans := []exploration.Plan{r.Fixed.Response.Plan, stored.Explore.Plan}
	for _, n := range r.Narrow {
		plans = append(plans, n.Response.Plan)
	}
	cp, e := exploration.Compile(r.Cardinality.Question)
	if e != nil {
		return "cardinality plan invalid"
	}
	plans = append(plans, cp)
	usedSQL := map[int]bool{}
	seenActions := map[string]bool{}
	ownedCalls := map[string]int{}
	for _, step := range r.Steps {
		seenActions[step.Action] = true
		if step.After.SQL-step.Before.SQL != len(step.SQL) {
			return "SQL counter mismatch"
		}
		source, writes := 0, 0
		wantedSQL := []all437PlannedSQL{}
		tailInvocation := false
		for _, call := range step.Calls {
			if pair428Metadata(call) {
				if call.Path == "/api/view-artifacts/encode" && step.Action == "export-recalled-all" {
					var reply, request struct {
						Artifact struct{ View workbench.ExplorationView }
					}
					if call.Status != 200 || json.Unmarshal(call.Body, &request) != nil || request.Artifact.View.Validate() != nil || !pair428Same(request.Artifact.View, r.Adopted) || json.Unmarshal(call.Response, &reply) != nil || !pair428Same(reply.Artifact.View, r.Adopted) {
						return "actual export response did not supply adopted view"
					}
				}
				continue
			}
			if call.Method != "POST" || !all437SourceAction(step.Action, call.Path) {
				return "unowned step API " + step.Action + " " + call.Path
			}
			switch call.Path {
			case "/api/tail":
				var selector exploration.TailSelector
				if json.Unmarshal(call.Body, &selector) != nil || !all437IsAll(selector.Query(q.ToMS)) || selector.Where != nil || call.Status != 200 {
					return "tail request differs from explicit All"
				}
				tailInvocation = true
			case "/api/explore":
				expected := q
				if strings.HasPrefix(step.Action, "read-reported-") {
					mode := strings.TrimPrefix(step.Action, "read-reported-")
					for _, n := range r.Narrow {
						if n.Mode == mode {
							expected = n.Question
						}
					}
				}
				var requested exploration.Query
				var response exploration.Result
				if json.Unmarshal(call.Body, &requested) != nil || call.Status != 200 || !pair428Same(requested, expected) || json.Unmarshal(call.Response, &response) != nil || !all437Result(r.Raw, requested, response) {
					return "source request/response differs from owned fixed question"
				}
				plan, e := exploration.Compile(requested)
				if e != nil {
					return "owned query could not compile"
				}
				wantedSQL = append(wantedSQL, all437ReadSQL(plan)...)
			case "/api/explore/counted-values":
				var requested exploration.Query
				if json.Unmarshal(call.Body, &requested) != nil || call.Status != 502 || !pair428Same(requested, r.Cardinality.Question) || !pair428Same(call.Response, r.Cardinality.Response) {
					return "cardinality invocation differs from its report"
				}
				wantedSQL = append(wantedSQL, all437PlannedSQL{cp.Count, cp}, all437PlannedSQL{cp.Groups, cp})
			case "/api/captures":
				var request struct {
					Question exploration.CaptureQuestion
					Title    string
				}
				if json.Unmarshal(call.Body, &request) != nil || request.Question.Kind != "explore" || !pair428Same(request.Question.Normalized(), stored.Question) || request.Title != stored.Title || call.Status != 201 || !pair428Same(call.Response, r.Capture) {
					return "capture invocation differs from stored All question"
				}
				wantedSQL = append(wantedSQL, all437ReadSQL(stored.Explore.Plan)...)
			}
			ownedCalls[all437CallKey(call)]++
			if call.Path == "/api/captures" {
				writes++
			} else {
				source++
			}
		}
		if source+writes > 1 {
			return "one invocation unexpectedly repeated source/publication"
		}
		if !tailInvocation {
			if len(step.SQL) != len(wantedSQL) {
				return "source read cost differs from owned invocation " + step.Action
			}
			for i, sql := range step.SQL {
				if sql.SQL != wantedSQL[i].text || !all437SQLMatches(sql, wantedSQL[i].plan) {
					return "source SQL differs from owned request order " + step.Action
				}
			}
		}
		if step.After.Source-step.Before.Source != source || step.After.Writes-step.Before.Writes != writes {
			return "API counters disagree"
		}
		if source == 0 && writes == 0 && len(step.SQL) != 0 {
			return "hidden SQL in local action"
		}
		sqlPlans := plans
		if tailInvocation {
			sqlPlans = nil
		}
		for _, sql := range step.SQL {
			if usedSQL[sql.Index] || sql.Parameters["database"] != r.Fixture.Database || !sql.Closed || !all437SourceSettings(sql) || !all437SQLAllowed(sql, sqlPlans, tailInvocation) {
				return "unowned/changed source SQL " + step.Action
			}
			usedSQL[sql.Index] = true
			if step.Action == "source-access-refusal-keeps-all" {
				if sql.Status != 403 || sql.Fault == nil {
					return "missing actual declared access refusal"
				}
			} else if sql.Status != 200 || sql.Fault != nil {
				return "unexpected source failure"
			}
		}
	}
	for _, action := range []string{"blank-is-not-all", "choose-all-explicitly", "spans-do-not-inherit-wide-permission", "start-all", "pause-all", "wide-row-context-refuses", "stage-fixed-all", "run-fixed-all", "return-to-held-all", "source-access-refusal-keeps-all", "high-cardinality-values-refusal", "capture-preview-is-local", "capture-all-fixed-question", "cold-stored-reopen", "open-exact-all-definition", "cold-recent-recalls-all", "export-recalled-all", "stage-reported-empty", "stage-reported-unreported", "stage-reported-named-all"} {
		if !seenActions[action] {
			return "missing observed action " + action
		}
	}
	for _, call := range r.Calls {
		if pair428Metadata(call) {
			continue
		}
		key := all437CallKey(call)
		if ownedCalls[key] == 0 {
			return "global API lacks owning invocation " + call.Path
		}
		ownedCalls[key]--
	}
	for _, remaining := range ownedCalls {
		if remaining != 0 {
			return "step API not in actual global ledger"
		}
	}
	for _, sql := range r.ActorSQL {
		if usedSQL[sql.Index] || sql.Parameters["database"] != r.Fixture.Database || !sql.Closed || !all437SourceSettings(sql) || sql.Status != 200 || sql.Fault != nil || !all437SQLAllowed(sql, nil, true) {
			return "actor SQL authority"
		}
		usedSQL[sql.Index] = true
	}
	if len(r.ActorSQL) == 0 || len(usedSQL) != len(r.SQL) {
		return "unowned global SQL"
	}
	for _, sql := range r.SQL {
		if !usedSQL[sql.Index] {
			return "global SQL absent from owner"
		}
	}
	return ""
}

func TestQuintAllEnvironments437Observed(t *testing.T) {
	file := os.Getenv("GOO_ALL_ENVIRONMENTS_437_REPORT")
	if file == "" {
		t.Skip("requires completed actual All-environments browser report")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var report all437Report
	if e = json.Unmarshal(raw, &report); e != nil {
		t.Fatal(e)
	}
	if reason := all437Admission(report); reason != "" {
		t.Fatal("actual All-environments admission:", reason)
	}
	checker, e := validate.NewChecker("model/population_drilldown.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	check := func(r all437Report) bool {
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		reads := 0
		coherent := all437Admission(r) == ""
		for _, step := range r.Steps {
			n := len(step.SQL)
			if n == 0 {
				rec.Record(t.Context(), "history", nil, "obs.reads", reads, "obs.depth", 0, "obs.coherent", coherent)
			} else {
				for range n {
					reads++
					rec.Record(t.Context(), "read", nil, "obs.reads", reads, "obs.depth", 0, "obs.coherent", coherent)
				}
			}
		}
		for range r.ActorSQL {
			reads++
			rec.Record(t.Context(), "read", nil, "obs.reads", reads, "obs.depth", 0, "obs.coherent", coherent)
		}
		v, e := checker.Check(t.Context(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return v.OK()
	}
	if !check(report) {
		t.Fatal("actual source/local trace rejected")
	}
	for mutant := range 8 {
		encoded, _ := json.Marshal(report)
		var bad all437Report
		if e = json.Unmarshal(encoded, &bad); e != nil {
			t.Fatal(e)
		}
		switch mutant {
		case 0:
			bad.Fixed.Response.Question.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "named", Value: "all"}
		case 1:
			row, _ := all437CanonicalRow(bad.Fixed.Response.Rows[0])
			for i := range bad.Raw {
				if bad.Raw[i].Body == row.Body {
					bad.Raw[i].Resource[identity.EnvironmentKey] = "changed"
					break
				}
			}
		case 2:
			bad.Fixed.Response.Rows = bad.Fixed.Response.Rows[1:]
		case 3:
			bad.Recalled.Queries[0].EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "unreported"}
		case 4:
			bad.Capture.RetainedRecords = 1005
		case 5:
			bad.SQL[0].Parameters["database"] = "default"
			for i := range bad.Steps {
				for j := range bad.Steps[i].SQL {
					if bad.Steps[i].SQL[j].Index == bad.SQL[0].Index {
						bad.Steps[i].SQL[j].Parameters["database"] = "default"
					}
				}
			}
		case 6:
			bad.Refused[0].SQL = 1
		case 7:
			bad.HeldAfter.Keys[0] = "retired-replacement"
		}
		if check(bad) {
			t.Fatalf("negative control %d accepted", mutant)
		}
	}
	t.Logf("admitted %d actual steps, %d configured-source SQL requests, 1005 raw observations and eight rejecting controls", len(report.Steps), len(report.SQL))
}
