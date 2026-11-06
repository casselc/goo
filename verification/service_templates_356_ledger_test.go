package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type serviceTemplateCounter356 struct{ Calls, Source, Expands, Saves, Plans int }
type serviceTemplateFact356 struct {
	Name          string
	Before, After serviceTemplateCounter356
}
type serviceTemplateCaseObserved356 struct {
	Key                       string
	Scope                     exploration.InventoryQuestion
	Plan                      exploration.DashboardPlan
	Widget                    exploration.DashboardWidget
	ParentBefore, ParentAfter []json.RawMessage
	Expansion                 struct {
		Request  workbench.ViewTemplateExpansionRequest
		Response workbench.ViewTemplateExpansion
	}
	Run struct {
		Request  exploration.Query
		Response exploration.Result
	}
	Oracle []struct{ ID string }
	Facts  []serviceTemplateFact356
}
type serviceTemplateLedger356 struct {
	Version string
	Fixture struct {
		Environment, Service, Namespace string
		FromMS, ToMS                    int64
		Rows                            []struct {
			ID, Service, Namespace string
			Env                    map[string]string
			At                     int64
		}
	}
	Templates struct{ Primary workbench.ExplorationView }
	Cases     []serviceTemplateCaseObserved356
	Calls     []serviceContinuationCall352
	Errors    []string
}

var serviceTemplateActions356 = []string{"service_context", "use_template", "inspect_template", "stage_template", "run_population", "return_service"}
var serviceTemplateFactNames356 = []string{"service-context", "use-template", "inspect-template", "stage-template", "run-population", "return-service"}

func serviceTemplateSource356(c serviceContinuationCall352) bool {
	return c.Method == "POST" && slices.Contains([]string{"/api/discovery", "/api/explore", "/api/explore/page", "/api/inventory", "/api/metric-discovery", "/api/record-aggregates", "/api/measurements"}, c.Path)
}
func serviceTemplateObservedAdmission356(l serviceTemplateLedger356, c serviceTemplateCaseObserved356) error {
	if continuationJSON352(c.Plan.Scope) != continuationJSON352(c.Scope) || c.Plan.Service != l.Fixture.Service {
		return fmt.Errorf("service plan destination identity")
	}
	q := c.Widget.Query
	if q == nil || c.Widget.State != "supported" || q.Offset != 0 || len(q.GroupBy) != 1 || q.Numeric != nil || q.NumericHistogram != nil || q.ReportedEdge != nil || q.GroupTimeSeries {
		return fmt.Errorf("ordinary completed destination capability")
	}
	if err := q.Validate(); err != nil {
		return err
	}
	if q.FromMS != c.Scope.FromMS || q.ToMS != c.Scope.ToMS || q.Environment != c.Scope.Environment || continuationJSON352(q.EnvironmentSelector) != continuationJSON352(c.Scope.EnvironmentSelector) {
		return fmt.Errorf("destination policy/window")
	}
	if c.Expansion.Request.Source == nil || c.Expansion.Request.Query == nil || continuationJSON352(*c.Expansion.Request.Query) != continuationJSON352(*q) {
		return fmt.Errorf("whole destination expansion request")
	}
	if continuationJSON352(*c.Expansion.Request.Source) != continuationJSON352(l.Templates.Primary) {
		return fmt.Errorf("selected saved definition identity")
	}
	expected, err := workbench.ExpandViewTemplate(c.Expansion.Request)
	if err != nil {
		return err
	}
	if continuationJSON352(expected) != continuationJSON352(c.Expansion.Response) {
		return fmt.Errorf("selected definition expansion")
	}
	if len(c.Expansion.Response.View.Queries) != 1 || continuationJSON352(c.Expansion.Response.View.Queries[0]) != continuationJSON352(*q) || continuationJSON352(c.Run.Request) != continuationJSON352(*q) || continuationJSON352(c.Run.Response.Question) != continuationJSON352(*q) {
		return fmt.Errorf("specimen or stage query leak")
	}
	if len(c.ParentBefore) < 2 || continuationJSON352(c.ParentBefore) != continuationJSON352(c.ParentAfter) {
		return fmt.Errorf("retained parent DOM/receipts/clocks")
	}
	mode := 0
	name := c.Scope.Environment
	if selected := c.Scope.EnvironmentSelector; selected != nil {
		name = selected.Scope.Value
		mode = 1
		if selected.Policy == "effective" {
			mode = 2
			if selected.Scope.Mode == "empty" {
				mode = 3
			}
			if selected.Scope.Mode == "unreported" {
				mode = 4
			}
		}
	}
	expectedIDs := []string{}
	for _, r := range l.Fixture.Rows {
		resource := map[string]string{}
		for k, v := range r.Env {
			resource[k] = v
		}
		resource["service.namespace"] = r.Namespace
		if r.Service == l.Fixture.Service && r.At >= q.FromMS && r.At < q.ToMS && continuationEnvironment352(mode, name, resource) && continuationNamespace352(c.Scope.Namespace, resource) {
			expectedIDs = append(expectedIDs, r.ID)
		}
	}
	slices.Sort(expectedIDs)
	got, err := reportedResultIDs353(c.Run.Response)
	if err != nil {
		return err
	}
	oracle := []string{}
	for _, r := range c.Oracle {
		oracle = append(oracle, r.ID)
	}
	slices.Sort(oracle)
	if len(expectedIDs) == 0 || !slices.Equal(got, expectedIDs) || !slices.Equal(oracle, expectedIDs) || c.Run.Response.Count != fmt.Sprint(len(expectedIDs)) {
		return fmt.Errorf("independent raw OTel destination membership")
	}
	if len(c.Facts) != len(serviceTemplateFactNames356) {
		return fmt.Errorf("incomplete actual gesture sequence")
	}
	for i, f := range c.Facts {
		if f.Name != serviceTemplateFactNames356[i] || f.Before.Calls < 0 || f.After.Calls < f.Before.Calls || f.After.Calls > len(l.Calls) {
			return fmt.Errorf("gesture boundaries")
		}
		want := 0
		if i == 4 {
			want = 1
		}
		actual := 0
		for _, call := range l.Calls[f.Before.Calls:f.After.Calls] {
			if serviceTemplateSource356(call) {
				actual++
				if i != 4 || call.Path != "/api/explore" {
					return fmt.Errorf("configuration acquired source")
				}
				var request exploration.Query
				if json.Unmarshal(call.Body, &request) != nil || continuationJSON352(request) != continuationJSON352(*q) {
					return fmt.Errorf("actual Run descriptor")
				}
			}
		}
		if actual != want || f.After.Source-f.Before.Source != want || f.After.Plans != f.Before.Plans || f.After.Saves != f.Before.Saves {
			return fmt.Errorf("explicit-only source/plan/save budget")
		}
		if i > 0 && f.Before.Source != c.Facts[i-1].After.Source {
			return fmt.Errorf("unaccounted telemetry between gestures")
		}
	}
	return nil
}
func serviceTemplateModelCheck356(t *testing.T, label string, steps []string, reads []int, coherent bool) bool {
	t.Helper()
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink, qobs.WithActor(label))
	for i, a := range steps {
		rec.Record(context.Background(), a, nil, "obs.reads", reads[i], "obs.coherent", coherent || i != len(steps)-1)
	}
	checker, err := validate.NewChecker("model/service_templates_356.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	report, err := checker.Check(context.Background(), sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	artifact("service-templates356-"+label+".txt", []byte(report.Summary()))
	return report.OK()
}

// This is a model/toolchain unit control, not claimed browser evidence. It
// exercises generated imports before the required native report gate.
func TestQuintServiceTemplateProtocol356(t *testing.T) {
	if !serviceTemplateModelCheck356(t, "protocol-complete", serviceTemplateActions356, []int{0, 0, 0, 0, 1, 1}, true) {
		t.Fatal("complete protocol refused")
	}
	if serviceTemplateModelCheck356(t, "protocol-implicit-read", serviceTemplateActions356, []int{0, 1, 1, 1, 2, 2}, true) {
		t.Fatal("implicit read accepted")
	}
	if serviceTemplateModelCheck356(t, "protocol-run-before-stage", []string{"service_context", "use_template", "inspect_template", "run_population", "return_service"}, []int{0, 0, 0, 1, 1}, true) {
		t.Fatal("unstaged read accepted")
	}
}
func TestQuintServiceTemplatesObserved356(t *testing.T) {
	path := os.Getenv("GOO_SERVICE_TEMPLATES_356_REPORT")
	if path == "" {
		t.Skip("explicit native observed report required")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var l serviceTemplateLedger356
	if err = json.Unmarshal(raw, &l); err != nil {
		t.Fatal(err)
	}
	if l.Version != "goo.serviceTemplates356/1" || len(l.Cases) != 4 || len(l.Errors) != 0 {
		t.Fatal("incomplete actual native campaign")
	}
	check := func(label string, c serviceTemplateCaseObserved356, want bool) {
		t.Helper()
		admission := serviceTemplateObservedAdmission356(l, c)
		if (admission == nil) != want {
			t.Fatalf("%s admission: %v", label, admission)
		}
		reads := make([]int, len(c.Facts))
		if len(c.Facts) == 6 {
			base := c.Facts[0].Before.Source
			for i, f := range c.Facts {
				reads[i] = f.After.Source - base
			}
		} else {
			reads = []int{0, 0, 0, 0, 1, 1}
		}
		if serviceTemplateModelCheck356(t, label, serviceTemplateActions356, reads, admission == nil) != want {
			t.Fatalf("%s model verdict", label)
		}
	}
	for _, c := range l.Cases {
		check("actual-"+c.Key, c, true)
	}
	mutants := []func(*serviceTemplateCaseObserved356){func(c *serviceTemplateCaseObserved356) { c.Run.Request.FromMS++ }, func(c *serviceTemplateCaseObserved356) {
		c.Expansion.Request.Query.GroupBy[0].Kind = "column"
		c.Expansion.Request.Query.GroupBy[0].Key = "service"
	}, func(c *serviceTemplateCaseObserved356) {
		c.Expansion.Response.View.Queries[0] = c.Expansion.Request.Source.Queries[0]
	}, func(c *serviceTemplateCaseObserved356) { c.ParentAfter = nil }, func(c *serviceTemplateCaseObserved356) { c.Run.Response.Rows = c.Run.Response.Rows[:0] }, func(c *serviceTemplateCaseObserved356) { c.Facts[1].After.Source++ }}
	for i, mutate := range mutants {
		var fresh serviceTemplateLedger356
		json.Unmarshal(raw, &fresh)
		c := fresh.Cases[0]
		mutate(&c)
		check(fmt.Sprintf("poison-%d", i), c, false)
	}
}
