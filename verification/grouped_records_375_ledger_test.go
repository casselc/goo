package verification

import (
	"context"
	"encoding/json"
	"net/url"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type groupedRect375 struct{ Left, Right, Top, Bottom, Width, Height float64 }
type groupedControl375 struct {
	Text, Href, Tag string
	Rect            groupedRect375
}
type groupedGroup375 struct {
	G0, G1 *string
	Count  string
}
type groupedLayout375 struct {
	Width         int
	Group         groupedGroup375
	Viewport      struct{ Width, Height, DocumentWidth float64 }
	Row           groupedRect375
	TableViewport struct{ ScrollLeft, ScrollWidth, ClientWidth float64 }
	Cells         []struct {
		Text, Label, Header, HeaderScope, HeaderID string
		Headers                                    *string
		Index                                      int
		Rect                                       groupedRect375
		NoClip                                     bool
		Links, Controls                            []groupedControl375
	}
}
type groupedRaw375 struct {
	ID, TraceID, Service, TimestampNS, Status, Severity string
	SeverityNumber                                      int
	Resources, Attributes                               map[string]string
	EventNames                                          []string
	EventAttributes                                     []map[string]string
}
type groupedSnapshot375 struct {
	DOMID           int
	Text, Basis     string
	Links, Controls []struct {
		ID              int
		Href, Text, Tag string
	}
}
type groupedOperation375 struct {
	Request, ExpectedQuestion exploration.Query
	Response                  exploration.Result
	Status                    int
	OracleIDs                 []string
	Type                      *string
	Focus                     struct{ Focused, Visible, CenterHit bool }
}
type groupedCounters375 struct{ Source, SQL, Store int }
type groupedFact375 struct {
	Name          string
	Before, After groupedCounters375
}
type groupedContext375 struct {
	Name                                                               string
	FixtureIndex                                                       int
	Triage                                                             bool
	Question                                                           exploration.Query
	Parent, Compare                                                    groupedOperation375
	Children                                                           []groupedOperation375
	ParentBefore, ParentAfter, ReflowReturnDesktop, ReflowReturnNarrow groupedSnapshot375
	Layouts                                                            []groupedLayout375
	Oracle                                                             struct {
		IDs    []string
		Groups []struct {
			Values []*string
			Count  string
			IDs    []string
		}
	}
	LocalSelection *struct {
		Intent       struct{ Type, DatasetID, Revision, ScopeID, Generation, Key string }
		SameSnapshot bool
		Status       string
	}
	StaleActions []struct {
		Before, After json.RawMessage
		Event         struct{ Connected, DefaultPrevented bool }
	}
	Facts []groupedFact375
}
type groupedAX375 struct {
	Context               string
	Width, TableBackendID int
	Headings, HeaderNames []string
	Nodes                 []struct {
		Ignored          bool
		BackendDOMNodeID int
		Role, Name       struct{ Value string }
	}
}
type groupedReport375 struct {
	Version  string
	Baseline bool
	Errors   []string
	Fixtures []typedTriageFixture370
	Raw      [][]groupedRaw375
	Contexts []groupedContext375
	AX       []groupedAX375
}

func groupedRawRecord375(r groupedRaw375, signal string) (etRecord, bool) {
	ns, e := strconv.ParseInt(r.TimestampNS, 10, 64)
	if e != nil || ns%1000000 != 0 || len(r.EventNames) != len(r.EventAttributes) {
		return etRecord{}, false
	}
	events := []etEvent{}
	for i, name := range r.EventNames {
		events = append(events, etEvent{Name: name, Attributes: r.EventAttributes[i]})
	}
	return etRecord{ID: r.ID, TraceID: r.TraceID, Service: r.Service, Environment: r.Resources["deployment.environment.name"], Signal: signal, TimeMS: ns / 1000000, Attributes: r.Attributes, Resource: r.Resources, Status: r.Status, Severity: r.Severity, SeverityNumber: strconv.Itoa(r.SeverityNumber), Events: events}, true
}
func groupedMembers375(q exploration.Query, raw []groupedRaw375) ([]etRecord, bool) {
	rows := []etRecord{}
	for _, r := range raw {
		row, ok := groupedRawRecord375(r, q.Signal)
		if !ok {
			return nil, false
		}
		scope := false
		if q.EnvironmentSelector != nil {
			scope = typedTriageEnvironment370(q, r.Resources)
		} else {
			value, present := r.Resources["deployment.environment.name"]
			scope = present && value == q.Environment
		}
		if !scope || row.TimeMS < q.FromMS || row.TimeMS >= q.ToMS {
			continue
		}
		yes, e := etMatch(q.Where, row)
		if e != nil {
			return nil, false
		}
		if yes {
			rows = append(rows, row)
		}
	}
	return rows, true
}
func groupedIDs375(rows []etRecord) []string {
	ids := make([]string, len(rows))
	for i, r := range rows {
		ids[i] = r.ID
	}
	slices.Sort(ids)
	return ids
}
func groupedGroups375(rows []etRecord, q exploration.Query) (map[string]string, bool) {
	counts := map[string]int{}
	for _, r := range rows {
		values := []*string{}
		for _, field := range q.GroupBy {
			v, e := etField(field, r)
			if e != nil {
				return nil, false
			}
			values = append(values, v)
		}
		counts[savedJSON360(values)]++
	}
	out := map[string]string{}
	for key, n := range counts {
		out[key] = strconv.Itoa(n)
	}
	return out, true
}
func groupedResult375(op groupedOperation375, q exploration.Query, raw []groupedRaw375) bool {
	rows, ok := groupedMembers375(q, raw)
	if !ok || op.Status != 200 || !nhLiveSameQuery(q, op.Request) || !nhLiveSameQuery(q, op.Response.Question) || op.Response.Basis == "" || op.Response.Count != strconv.Itoa(len(rows)) {
		return false
	}
	ids, e := reportedResultIDs353(op.Response)
	if e != nil || !slices.Equal(ids, groupedIDs375(rows)) {
		return false
	}
	groups, ok := groupedGroups375(rows, q)
	if !ok || len(groups) != len(op.Response.Groups) {
		return false
	}
	seen := map[string]bool{}
	for _, raw := range op.Response.Groups {
		var g map[string]json.RawMessage
		if json.Unmarshal(raw, &g) != nil {
			return false
		}
		values := []*string{}
		for i := range q.GroupBy {
			var v *string
			if json.Unmarshal(g["g"+strconv.Itoa(i)], &v) != nil {
				return false
			}
			values = append(values, v)
		}
		key := savedJSON360(values)
		var count string
		if json.Unmarshal(g["count"], &count) != nil || groups[key] != count || seen[key] {
			return false
		}
		seen[key] = true
	}
	return true
}
func groupedChild375(q exploration.Query, service string, typ *string, inspect bool) exploration.Query {
	raw, _ := json.Marshal(q)
	var child exploration.Query
	_ = json.Unmarshal(raw, &child)
	predicates := []exploration.Predicate{}
	if child.Where != nil {
		if len(child.Where.All) > 0 {
			predicates = append(predicates, child.Where.All...)
		} else {
			predicates = append(predicates, *child.Where)
		}
	}
	predicates = append(predicates, exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: service})
	child.GroupBy = []exploration.Field{q.GroupBy[1]}
	child.Offset = 0
	child.GroupTimeSeries = false
	if inspect {
		p := exploration.Predicate{Field: &child.GroupBy[0], Op: "missing"}
		if typ != nil {
			p.Op = "eq"
			p.Value = *typ
		}
		predicates = append(predicates, p)
	}
	child.Where = &exploration.Predicate{All: predicates}
	return child
}
func groupedLink375(href string) (exploration.Query, bool) {
	var q exploration.Query
	if !strings.HasPrefix(href, "#step=") {
		return q, false
	}
	raw, e := url.PathUnescape(strings.TrimPrefix(href, "#step="))
	if e != nil {
		return q, false
	}
	var step struct {
		Version int
		Kind    string
		Record  exploration.Query
	}
	if json.Unmarshal([]byte(raw), &step) != nil || step.Version != 1 || step.Kind != "record" {
		return q, false
	}
	return step.Record, true
}
func groupedLabel375(value *string, field exploration.Field, triage bool) string {
	if value == nil {
		if !triage {
			return "Missing"
		}
		if field.Key == "exception_type" {
			return "No unique reported exception type"
		}
		return "Missing key"
	}
	if *value == "" {
		return "Present empty string"
	}
	if !triage || slices.Contains([]string{"Missing key", "Present empty string", "No unique reported exception type"}, *value) {
		return strconv.Quote(*value)
	}
	return *value
}
func groupedInside375(r groupedRect375, width, height float64) bool {
	return r.Width > 0 && r.Height > 0 && r.Left >= -0.01 && r.Top >= -0.01 && r.Right <= width+1 && r.Bottom <= height+1 && r.Right > r.Left && r.Bottom > r.Top
}
func groupedGeometry375(l groupedLayout375, c groupedContext375) bool {
	if len(c.Question.GroupBy) != 2 || len(l.Cells) != 4 || l.Width != int(l.Viewport.Width) || l.TableViewport.ScrollLeft != 0 || l.TableViewport.ScrollWidth > l.TableViewport.ClientWidth+1 || l.Viewport.DocumentWidth > l.Viewport.Width || !groupedInside375(l.Row, l.Viewport.Width, l.Viewport.Height) {
		return false
	}
	values := []*string{l.Group.G0, l.Group.G1}
	count := l.Group.Count
	if !c.Triage {
		count += " records"
	}
	for i, cell := range l.Cells {
		if cell.Index != i || cell.Label == "" || cell.Label != cell.Header || cell.HeaderScope != "col" || !cell.NoClip || !groupedInside375(cell.Rect, l.Viewport.Width, l.Viewport.Height) {
			return false
		}
		if i < 2 && (cell.Text != groupedLabel375(values[i], c.Question.GroupBy[i], c.Triage) || cell.Header != c.Question.GroupBy[i].Kind+":"+c.Question.GroupBy[i].Key) {
			return false
		}
		if i == 2 && cell.Text != count {
			return false
		}
		for j := 0; j < i; j++ {
			other := l.Cells[j].Rect
			if min(cell.Rect.Right, other.Right)-max(cell.Rect.Left, other.Left) > 1 && min(cell.Rect.Bottom, other.Bottom)-max(cell.Rect.Top, other.Top) > 1 {
				return false
			}
		}
		for _, control := range cell.Controls {
			if !groupedInside375(control.Rect, l.Viewport.Width, l.Viewport.Height) || control.Rect.Left < cell.Rect.Left-1 || control.Rect.Right > cell.Rect.Right+1 || control.Rect.Top < cell.Rect.Top-1 || control.Rect.Bottom > cell.Rect.Bottom+1 {
				return false
			}
			if l.Width == 390 {
				minimum := 24.
				if c.Triage {
					minimum = 28
				}
				if control.Rect.Height < minimum {
					return false
				}
			}
		}
	}
	action := l.Cells[3]
	if c.Triage {
		if len(action.Links) != 2 || action.Links[0].Text != "Inspect group" || action.Links[1].Text != "Compare within service" || l.Group.G0 == nil {
			return false
		}
		for i, link := range action.Links {
			q, ok := groupedLink375(link.Href)
			if !ok || !nhLiveSameQuery(q, groupedChild375(c.Question, *l.Group.G0, l.Group.G1, i == 0)) {
				return false
			}
		}
	} else if len(action.Links) != 0 || len(action.Controls) != 1 || action.Controls[0].Text != "Select" {
		return false
	}
	return true
}
func groupedAdmission375(c groupedContext375, r groupedReport375) bool {
	if c.FixtureIndex < 0 || c.FixtureIndex >= len(r.Raw) || c.FixtureIndex >= len(r.Fixtures) {
		return false
	}
	raw := r.Raw[c.FixtureIndex]
	f := r.Fixtures[c.FixtureIndex]
	if !groupedResult375(c.Parent, c.Question, raw) || c.ParentBefore.DOMID == 0 || c.ParentBefore.Basis != c.Parent.Response.Basis || !reflect.DeepEqual(c.ParentBefore, c.ParentAfter) || !reflect.DeepEqual(c.ParentBefore, c.ReflowReturnDesktop) || !reflect.DeepEqual(c.ParentBefore, c.ReflowReturnNarrow) {
		return false
	}
	rows, ok := groupedMembers375(c.Question, raw)
	if !ok {
		return false
	}
	ids := slices.Clone(c.Oracle.IDs)
	slices.Sort(ids)
	if !slices.Equal(ids, groupedIDs375(rows)) {
		return false
	}
	groups, ok := groupedGroups375(rows, c.Question)
	if !ok || len(groups) != len(c.Oracle.Groups) || len(c.Layouts) != 2*len(groups) {
		return false
	}
	for _, g := range c.Oracle.Groups {
		if groups[savedJSON360(g.Values)] != g.Count {
			return false
		}
	}
	seen := map[string]bool{}
	for _, layout := range c.Layouts {
		key := strconv.Itoa(layout.Width) + ":" + savedJSON360([]*string{layout.Group.G0, layout.Group.G1})
		if seen[key] || groups[savedJSON360([]*string{layout.Group.G0, layout.Group.G1})] != layout.Group.Count || !groupedGeometry375(layout, c) {
			return false
		}
		seen[key] = true
	}
	if !c.Triage {
		l := c.LocalSelection
		if len(c.Children) != 0 || len(c.StaleActions) != 0 || l == nil || !l.SameSnapshot || l.Intent.Type != "select" || l.Intent.Revision != c.Parent.Response.Basis || l.Intent.Generation != c.Parent.Response.Basis || l.Intent.ScopeID == "" || l.Intent.DatasetID != l.Intent.ScopeID+":groups" || !strings.HasPrefix(l.Intent.Key, "category:") || l.Status == "" {
			return false
		}
	} else {
		children := 1
		if c.Name == "typed-span-triage" {
			children = 4
		}
		if len(c.Children) != children || len(c.StaleActions) != children {
			return false
		}
		for _, child := range c.Children {
			wanted := groupedChild375(c.Question, f.TargetService, child.Type, true)
			if !nhLiveSameQuery(child.ExpectedQuestion, wanted) || !groupedResult375(child, wanted, raw) || !child.Focus.Focused || !child.Focus.Visible || !child.Focus.CenterHit {
				return false
			}
			members, _ := groupedMembers375(wanted, raw)
			ids := slices.Clone(child.OracleIDs)
			slices.Sort(ids)
			if !slices.Equal(ids, groupedIDs375(members)) {
				return false
			}
		}
		if !groupedResult375(c.Compare, groupedChild375(c.Question, f.TargetService, nil, false), raw) {
			return false
		}
		for _, stale := range c.StaleActions {
			if !stale.Event.Connected || !stale.Event.DefaultPrevented || len(stale.Before) == 0 || savedJSON360(stale.Before) != savedJSON360(stale.After) {
				return false
			}
		}
	}
	names := []string{"parent-run", "layout-1280", "layout-390", "local-return-reflow"}
	if !c.Triage {
		names = append(names, "ordinary-local-select")
	} else {
		for range c.Children {
			names = append(names, "keyboard-inspect-stage", "inspect-run", "held-back", "stale-connected-callback")
		}
		names = append(names, "keyboard-compare-stage", "compare-run", "compare-held-back")
	}
	if len(names) != len(c.Facts) {
		return false
	}
	for i, f := range c.Facts {
		if f.Name != names[i] || f.After.Store != f.Before.Store || i > 0 && f.Before != c.Facts[i-1].After {
			return false
		}
		read := strings.HasSuffix(f.Name, "-run")
		if !read && f.After.SQL != f.Before.SQL {
			return false
		}
	}
	return true
}
func groupedAXAdmission375(r groupedReport375) bool {
	if len(r.AX) != 6 {
		return false
	}
	seen := map[string]bool{}
	for _, ax := range r.AX {
		key := ax.Context + ":" + strconv.Itoa(ax.Width)
		if seen[key] || len(ax.Nodes) == 0 || ax.TableBackendID <= 0 || ax.Nodes[0].Ignored || ax.Nodes[0].Role.Value != "table" || ax.Nodes[0].BackendDOMNodeID != ax.TableBackendID || ax.Nodes[0].Name.Value == "" || len(ax.Headings) != 4 || len(ax.HeaderNames) != 4 {
			return false
		}
		seen[key] = true
		headers := []string{}
		roles := map[string]int{}
		for _, node := range ax.Nodes {
			if !node.Ignored {
				roles[node.Role.Value]++
				if node.Role.Value == "columnheader" {
					headers = append(headers, node.Name.Value)
				}
			}
		}
		if !slices.Equal(headers, ax.HeaderNames) || roles["row"] == 0 || roles["cell"] == 0 {
			return false
		}
		for i, h := range headers {
			if !strings.EqualFold(h, ax.Headings[i]) {
				return false
			}
		}
		index := slices.IndexFunc(r.Contexts, func(c groupedContext375) bool { return c.Name == ax.Context })
		if index < 0 || (ax.Width != 390 && ax.Width != 1280) {
			return false
		}
		c := r.Contexts[index]
		for _, layout := range c.Layouts {
			if layout.Width == ax.Width {
				for i, cell := range layout.Cells {
					if cell.Header != ax.Headings[i] {
						return false
					}
				}
			}
		}
		if c.Triage && roles["link"] == 0 || !c.Triage && roles["button"] == 0 {
			return false
		}
	}
	return true
}
func groupedReplay375(t *testing.T, label string, facts []groupedFact375) bool {
	t.Helper()
	if len(facts) == 0 {
		t.Fatal("missing observed group facts")
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	base := facts[0].Before.Source
	depth := 0
	for _, f := range facts {
		action := "history"
		switch f.Name {
		case "parent-run", "inspect-run", "compare-run":
			action = "read"
		case "keyboard-inspect-stage", "keyboard-compare-stage":
			action = "stage"
			depth++
		case "held-back", "compare-held-back":
			action = "back"
			depth--
		}
		rec.Record(context.Background(), action, nil, "obs.reads", f.After.Source-base, "obs.depth", depth, "obs.coherent", true)
	}
	raw, e := json.Marshal(sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	if e = artifact("grouped375-"+label+"-steps.json", raw); e != nil {
		t.Fatal(e)
	}
	checker, e := validate.NewChecker("model/population_drilldown.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	report, e := checker.Check(t.Context(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	if e = artifact("grouped375-"+label+"-model.txt", []byte(report.Summary())); e != nil {
		t.Fatal(e)
	}
	return report.OK()
}
func TestQuintGroupedRecordsObserved375(t *testing.T) {
	path := os.Getenv("GOO_GROUPED_RECORDS_375_REPORT")
	if path == "" {
		t.Skip("actual grouped record journey opt-in")
	}
	baselinePath := os.Getenv("GOO_GROUPED_RECORDS_375_BASELINE_REPORT")
	if baselinePath == "" {
		t.Fatal("matched actual baseline report required")
	}
	read := func(path string) groupedReport375 {
		raw, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		var r groupedReport375
		if json.Unmarshal(raw, &r) != nil {
			t.Fatal("malformed actual group report")
		}
		return r
	}
	r, old := read(path), read(baselinePath)
	if r.Version != "goo.groupedRecords375/1" || old.Version != r.Version || r.Baseline || !old.Baseline || len(r.Errors) != 0 || len(old.Errors) != 0 || len(r.Contexts) != 3 || len(old.Contexts) != 3 || savedJSON360(r.Fixtures) != savedJSON360(old.Fixtures) || savedJSON360(r.Raw) != savedJSON360(old.Raw) || !groupedAXAdmission375(r) {
		t.Fatal("actual grouped fixture/AX admission")
	}
	for i, c := range r.Contexts {
		if !groupedAdmission375(c, r) {
			t.Fatalf("context %s membership/layout/lifecycle admission", c.Name)
		}
		before := old.Contexts[i]
		if before.Name != c.Name || !nhLiveSameQuery(before.Question, c.Question) || before.Parent.Response.Basis != c.Parent.Response.Basis {
			t.Fatal("unmatched baseline")
		}
		overflow := false
		for _, l := range before.Layouts {
			if l.Width == 390 && l.TableViewport.ScrollWidth > l.TableViewport.ClientWidth+1 {
				overflow = true
			}
		}
		if !overflow {
			t.Fatal("baseline did not observe narrow overflow")
		}
		if !groupedReplay375(t, c.Name, c.Facts) {
			t.Fatal("observed acquisition/navigation", c.Name)
		}
	}
	c := r.Contexts[1]
	facts := slices.Clone(c.Facts)
	facts[4].After.Source++
	if groupedReplay375(t, "bad-stage-read", facts) {
		t.Fatal("stage read mutant survived")
	}
	geometry := c
	geometry.Layouts = slices.Clone(c.Layouts)
	i := slices.IndexFunc(geometry.Layouts, func(l groupedLayout375) bool { return l.Width == 390 })
	geometry.Layouts[i].Cells = slices.Clone(geometry.Layouts[i].Cells)
	geometry.Layouts[i].Cells[3].Rect.Right = 500
	if groupedAdmission375(geometry, r) {
		t.Fatal("clipped action mutant survived")
	}
	moved := c
	moved.Layouts = slices.Clone(c.Layouts)
	moved.Layouts[i].Cells = slices.Clone(c.Layouts[i].Cells)
	moved.Layouts[i].Cells[3].Controls = slices.Clone(c.Layouts[i].Cells[3].Controls)
	moved.Layouts[i].Cells[3].Controls[0].Rect.Top = c.Layouts[i].Cells[0].Rect.Top
	if groupedAdmission375(moved, r) {
		t.Fatal("moved action target mutant survived")
	}
	query := c
	query.Children = slices.Clone(c.Children)
	query.Children[0].Request = c.Question
	if groupedAdmission375(query, r) {
		t.Fatal("changed child query mutant survived")
	}
}
