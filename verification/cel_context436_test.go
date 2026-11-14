package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"github.com/casselc/goo/querycel"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

// The browser supplies observed requests, complete source facts and held snapshots.
// Native and CEL planners are recomputed here; the model projects only navigation
// depth and API source-request counts, not native metric causality or completeness.
type celContext436Counters struct{ Source, Writes, Posts int }
type celContext436Step struct {
	Action        string
	Before, After celContext436Counters
}
type celContext436Fact struct {
	ID, Service, Instance, TimestampNS, Name string
	Namespace                                *string
	Env                                      map[string]string
	SeverityNumber                           uint64
	SeverityText                             string
}
type celContext436Case struct {
	Kind, Signal   string
	Original       exploration.Query
	CheckRequest   querycel.PlanRequest
	Plan           querycel.RecordPlan
	Measurement    measurements.Result
	SelectedStream metrics.Stream
	ContextPlan    exploration.NativeRecordContext
	ServicePlan    exploration.DashboardPlan
	Run            struct {
		Request  exploration.Query
		Response exploration.Result
		Status   int
	}
	Oracle struct {
		Count    string
		Rows     []json.RawMessage
		Selected []string
	}
	Presentation              []struct{ Text, Band, Tone string }
	ParentBefore, ParentAfter json.RawMessage
	Steps                     []celContext436Step
	Retired                   struct {
		Before, After celContext436Counters
		Invoked       bool
	}
}
type celContext436Report struct {
	Version string
	Status  string
	Cleanup bool
	Errors  []string
	Fixture struct {
		Environment, FromNS, ToNS string
		Records                   []celContext436Fact
	}
	Calls    []struct{ Path, Method string }
	Contexts []celContext436Case
}

func celContext436Same(a, b any) bool {
	aa, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bb, err := json.Marshal(b)
	return err == nil && string(aa) == string(bb)
}
func celContext436Snapshot(raw json.RawMessage) any {
	var v any
	if json.Unmarshal(raw, &v) != nil {
		return nil
	}
	return v
}
func celContext436Interval(c celContext436Case) bool {
	if c.Measurement.Basis == "" || c.Measurement.ReadAt == "" || c.ContextPlan.Grain != "interval-context" {
		return false
	}
	matched := 0
	if c.Measurement.Counters != nil {
		for _, s := range c.Measurement.Counters.Series {
			if celContext436Same(s.Stream, c.SelectedStream) {
				for _, i := range s.Intervals {
					if i.FromNS == c.ContextPlan.FromNS && i.ToNS == c.ContextPlan.ToNS {
						matched++
					}
				}
			}
		}
	}
	if c.Measurement.Histograms != nil {
		for _, s := range c.Measurement.Histograms.Series {
			if celContext436Same(s.Stream, c.SelectedStream) {
				for _, i := range s.Intervals {
					if i.FromNS == c.ContextPlan.FromNS && i.ToNS == c.ContextPlan.ToNS {
						matched++
					}
				}
			}
		}
	}
	if matched != 1 {
		return false
	}
	expected, err := exploration.PlanNativeRecordContext(c.Measurement.Question, c.SelectedStream, c.Signal, c.ContextPlan.FromNS, c.ContextPlan.ToNS)
	return err == nil && celContext436Same(expected, c.ContextPlan) && celContext436Same(expected.Question, c.Original)
}
func celContext436IDs(rows []json.RawMessage) ([]string, bool) {
	ids := []string{}
	seen := map[string]bool{}
	for _, raw := range rows {
		var row struct {
			ID     string `json:"id"`
			SpanID string `json:"SpanId"`
		}
		if json.Unmarshal(raw, &row) != nil {
			return nil, false
		}
		id := row.ID
		if id == "" {
			id = row.SpanID
		}
		if id == "" || seen[id] {
			return nil, false
		}
		seen[id] = true
		ids = append(ids, id)
	}
	slices.Sort(ids)
	return ids, true
}

// Independent membership follows the fixture's explicit producer/interval and
// numeric severity facts. It never treats FATAL text as numeric Error.
func celContext436Expected(r celContext436Report, c celContext436Case, refined bool) ([]string, bool) {
	from, err := strconv.ParseUint(c.ContextPlan.FromNS, 10, 64)
	if err != nil {
		return nil, false
	}
	to, err := strconv.ParseUint(c.ContextPlan.ToNS, 10, 64)
	if err != nil {
		return nil, false
	}
	ns, wantsNS := c.SelectedStream.Resource["service.namespace"]
	ids := []string{}
	infoFatal := false
	for _, f := range r.Fixture.Records {
		at, e := strconv.ParseUint(f.TimestampNS, 10, 64)
		if e != nil {
			return nil, false
		}
		env, ok := f.Env["deployment.environment.name"]
		if !ok {
			env = f.Env["deployment.environment"]
		}
		if env != r.Fixture.Environment || f.Service != c.SelectedStream.Resource["service.name"] || f.Instance != c.SelectedStream.Resource["service.instance.id"] || at < from || at >= to {
			continue
		}
		if wantsNS {
			if f.Namespace == nil || *f.Namespace != ns {
				continue
			}
		} else if f.Namespace != nil {
			continue
		}
		if refined && (f.SeverityNumber > 9 || !(strings.Contains(f.Name, "inside-") || f.Name == "at-lower")) {
			continue
		}
		if f.SeverityNumber == 9 && f.SeverityText == "FATAL" {
			infoFatal = true
		}
		ids = append(ids, f.ID)
	}
	slices.Sort(ids)
	return ids, infoFatal
}
func celContext436Rows(r celContext436Report, c celContext436Case) bool {
	facts := map[string]celContext436Fact{}
	for _, f := range r.Fixture.Records {
		facts[f.ID] = f
	}
	for _, raw := range c.Oracle.Rows {
		var row struct {
			SpanID                          string `json:"SpanId"`
			Body, SeverityText, ServiceName string
			SeverityNumber                  json.RawMessage
			ResourceAttributes              map[string]string
			TimestampNS                     string `json:"timestamp_ns"`
		}
		if json.Unmarshal(raw, &row) != nil {
			return false
		}
		f, ok := facts[row.SpanID]
		if !ok {
			return false
		}
		var number uint64
		if json.Unmarshal(row.SeverityNumber, &number) != nil {
			var value string
			if json.Unmarshal(row.SeverityNumber, &value) != nil {
				return false
			}
			n, e := strconv.ParseUint(value, 10, 64)
			if e != nil {
				return false
			}
			number = n
		}
		if row.Body != f.Name || row.ServiceName != f.Service || row.SeverityText != f.SeverityText || number != f.SeverityNumber || row.TimestampNS != f.TimestampNS || row.ResourceAttributes["service.instance.id"] != f.Instance {
			return false
		}
		ns, known := row.ResourceAttributes["service.namespace"]
		if f.Namespace == nil && known || f.Namespace != nil && (!known || ns != *f.Namespace) {
			return false
		}
		for _, key := range []string{"deployment.environment.name", "deployment.environment"} {
			value, known := row.ResourceAttributes[key]
			expected, present := f.Env[key]
			if known != present || value != expected {
				return false
			}
		}
	}
	return true
}

func celContext436Calls(r celContext436Report, c celContext436Case) bool {
	posts := []string{}
	for _, call := range r.Calls {
		if call.Method == "POST" {
			posts = append(posts, call.Path)
		}
	}
	for _, step := range c.Steps {
		lo, hi := step.Before.Posts, step.After.Posts
		if lo < 0 || hi < lo || hi > len(posts) {
			return false
		}
		plans, runs := 0, 0
		for _, path := range posts[lo:hi] {
			cancel := strings.HasPrefix(path, "/api/queries/") && strings.HasSuffix(path, "/cancel")
			if cancel && step.Action != "retired" {
				continue
			}
			if step.Action == "check" && (path == "/api/cel/context" || path == "/api/cel/plan") {
				if path == "/api/cel/plan" {
					plans++
				}
				continue
			}
			if step.Action == "run" && path == "/api/explore" {
				runs++
				continue
			}
			return false
		}
		if step.Action == "check" && plans != 1 || step.Action == "run" && runs != 1 {
			return false
		}
	}
	return true
}

func celContext436Admission(r celContext436Report, c celContext436Case) bool {
	if r.Version != "goo.celContextEntry4366/1" || !slices.Contains([]string{"native", "service"}, c.Kind) || c.Signal != "logs" || !celContext436Interval(c) || !celContext436Calls(r, c) {
		return false
	}
	if c.Kind == "native" && c.Measurement.Question.Definition.Kind != "counter" || c.Kind == "service" && c.Measurement.Question.Definition.Kind != "histogram" {
		return false
	}
	if c.Kind == "service" {
		found := 0
		for _, widget := range c.ServicePlan.Widgets {
			if widget.Measurement != nil && celContext436Same(widget.Measurement, c.Measurement.Question) {
				found++
			}
		}
		if found != 1 || c.ServicePlan.Service != c.SelectedStream.Resource["service.name"] {
			return false
		}
	}
	if !celContext436Rows(r, c) {
		return false
	}
	if c.CheckRequest.Expression != `log.severity_number <= 9u && (log.body.contains("inside-") || log.body == "at-lower")` || !celContext436Same(c.CheckRequest.Base, c.Original) {
		return false
	}
	plan, err := querycel.PlanRecord(c.CheckRequest)
	if err != nil || !celContext436Same(plan, c.Plan) || !celContext436Same(c.Run.Request, plan.Question) || !celContext436Same(c.Run.Response.Question, c.Run.Request) || c.Run.Status != 200 || c.Run.Response.Basis == "" {
		return false
	}
	if len(c.Presentation) != 100 {
		return false
	}
	for _, label := range c.Presentation {
		if label.Text != "Info · producer: FATAL" || label.Band != "info" || label.Tone != "neutral" {
			return false
		}
	}
	want, witness := celContext436Expected(r, c, true)
	parent, _ := celContext436Expected(r, c, false)
	if !witness || len(want) <= 100 || c.Oracle.Count != strconv.Itoa(len(want)) || c.Run.Response.Count != c.Oracle.Count || len(c.Run.Response.Rows) != 100 || !c.Run.Response.HasMore {
		return false
	}
	oracle, ok := celContext436IDs(c.Oracle.Rows)
	if !ok || !slices.Equal(oracle, parent) {
		return false
	}
	selected := slices.Clone(c.Oracle.Selected)
	slices.Sort(selected)
	if !slices.Equal(selected, want) {
		return false
	}
	actual, ok := celContext436IDs(c.Run.Response.Rows)
	if !ok {
		return false
	}
	for _, id := range actual {
		if !slices.Contains(want, id) {
			return false
		}
	}
	before, after := celContext436Snapshot(c.ParentBefore), celContext436Snapshot(c.ParentAfter)
	return before != nil && reflect.DeepEqual(before, after) && c.Retired.Invoked && c.Retired.Before == c.Retired.After && len(c.Steps) > 0 && c.Steps[len(c.Steps)-1].Before == c.Retired.Before && c.Steps[len(c.Steps)-1].After == c.Retired.After
}

func TestQuintCELContext436Observed(t *testing.T) {
	name := os.Getenv("GOO_CEL_CONTEXT_4366_REPORT")
	if name == "" {
		t.Skip("requires actual Services/native CEL context browser report")
	}
	raw, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	var observed celContext436Report
	if err = json.Unmarshal(raw, &observed); err != nil {
		t.Fatal(err)
	}
	if observed.Status != "passed" || !observed.Cleanup || len(observed.Errors) != 0 {
		for _, c := range observed.Contexts {
			want, witness := celContext436Expected(observed, c, true)
			t.Logf("partial %s: admission=%v native=%v sourceRows=%v expected=%d InfoFATAL=%v steps=%d", c.Kind, celContext436Admission(observed, c), celContext436Interval(c), celContext436Rows(observed, c), len(want), witness, len(c.Steps))
		}
		t.Fatal("require completed, clean real browser run")
	}
	if len(observed.Contexts) != 2 {
		t.Fatal("require actual native and service contexts")
	}
	kinds := map[string]bool{}
	for _, c := range observed.Contexts {
		if kinds[c.Kind] || !celContext436Admission(observed, c) {
			t.Fatalf("native/CEL/source/held-context admission failed for %s", c.Kind)
		}
		kinds[c.Kind] = true
	}
	checker, err := validate.NewChecker("model/population_drilldown.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	check := func(r celContext436Report, c celContext436Case) bool {
		coherent := celContext436Admission(r, c)
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		depth, reads := 0, 0
		expectedActions := []string{"enter", "check", "run", "inspect", "return", "retired"}
		if len(c.Steps) != len(expectedActions) {
			coherent = false
		}
		previous := celContext436Counters{}
		for i, step := range c.Steps {
			if i >= len(expectedActions) || step.Action != expectedActions[i] {
				coherent = false
			}
			if i > 0 && (previous.Source != step.Before.Source || previous.Writes != step.Before.Writes) {
				coherent = false
			}
			previous = step.After
			if step.After.Posts < step.Before.Posts || step.Action == "check" && step.After.Posts <= step.Before.Posts {
				coherent = false
			}
			action := "history"
			delta := step.After.Source - step.Before.Source
			if step.Before.Writes != step.After.Writes || step.Before.Source < 0 || step.After.Source < 0 {
				coherent = false
			}
			switch step.Action {
			case "enter":
				action = "stage"
				depth++
			case "run":
				action = "read"
				if delta != 1 {
					coherent = false
				}
			case "return":
				action = "back"
				depth--
			case "check", "retired", "mode", "edit", "invalid", "inspect":
			default:
				coherent = false
			}
			if step.Action != "run" && delta != 0 {
				coherent = false
			}
			reads += delta
			rec.Record(context.Background(), action, nil, "obs.reads", reads, "obs.depth", depth, "obs.coherent", coherent)
		}
		if depth != 0 || reads != 1 {
			coherent = false
			rec.Record(context.Background(), "history", nil, "obs.reads", reads, "obs.depth", depth, "obs.coherent", false)
		}
		result, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return result.OK()
	}
	for _, c := range observed.Contexts {
		if !check(observed, c) {
			t.Fatalf("actual %s navigation/request trace refused", c.Kind)
		}
	}
	mutants := []string{"hidden-read", "parent", "bounds", "severity", "native-producer", "retired-read", "custody", "page-total", "severity-display", "service-origin"}
	for _, mutation := range mutants {
		var r celContext436Report
		json.Unmarshal(raw, &r)
		c := &r.Contexts[0]
		if mutation == "service-origin" {
			c = &r.Contexts[1]
		}
		switch mutation {
		case "hidden-read":
			for i := range c.Steps {
				if c.Steps[i].Action == "check" {
					c.Steps[i].After.Source++
					break
				}
			}
		case "parent":
			c.CheckRequest.Base.Where = nil
		case "bounds":
			c.Run.Response.Question.ToMS--
		case "severity":
			c.CheckRequest.Expression = `log.severity_text == "FATAL"`
		case "native-producer":
			c.SelectedStream.Resource["service.instance.id"] = "other-instance"
		case "retired-read":
			c.Retired.After.Source++
		case "service-origin":
			c.ServicePlan.Service = "foreign-service"
		case "severity-display":
			c.Presentation[0].Band = "fatal"
		case "page-total":
			c.Run.Response.Count = strconv.Itoa(len(c.Run.Response.Rows))
		case "custody":
			c.ParentAfter = json.RawMessage(`{"forged":true}`)
		}
		if check(r, *c) {
			t.Fatal("incorrect observed context admitted", mutation)
		}
	}
	t.Logf("admitted two real context traces; ten semantic/authority controls rejected (%s)", fmt.Sprint(kinds))
}
