package verification

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
)

type rpcTemplateFact367 struct {
	Name          string
	Before, After struct{ Calls, Source, SQL int }
}
type rpcTemplateReport367 struct {
	Version string
	Fixture struct {
		Rows []struct {
			ID, Service, Namespace string
			Env                    map[string]string
			At                     int64
			Kind                   json.RawMessage
			Attributes             map[string]string
		}
	}
	Source struct {
		Plan      exploration.DashboardPlan
		Widget    exploration.DashboardWidget
		Customize struct {
			Request       workbench.ViewTemplateExpansionRequest
			Response      workbench.ViewTemplateExpansion
			Before, After struct{ Source, SQL int }
		}
		EditedComposition workbench.ViewComposition
		Saved             workbench.ExplorationView
		Save              struct{ Before, After struct{ Source, SQL int } }
	}
	Target struct {
		Plan      exploration.DashboardPlan
		Widget    exploration.DashboardWidget
		Expansion struct {
			Request  workbench.ViewTemplateExpansionRequest
			Response workbench.ViewTemplateExpansion
		}
		StagedComposition workbench.ViewComposition
		Run               struct {
			Request  exploration.Query
			Response exploration.Result
		}
		Oracle struct {
			Count string
			Rows  []struct{ ID string }
		}
		ParentBefore, ParentAfter []json.RawMessage
	}
	TemplateAfter workbench.ExplorationView
	Stale         struct {
		Status                                    int
		BeforeOldApply, AfterOldApply, AfterReply json.RawMessage
		ClientPlan                                exploration.DashboardPlan
	}
	Refusals []struct {
		Kind          string
		Status        int
		Before, After struct{ Source, SQL int }
	}
	Facts []rpcTemplateFact367
	Calls []serviceContinuationCall352
}

func rpcTemplateAdmission367(r rpcTemplateReport367) error {
	a, b := r.Source.Widget, r.Target.Widget
	if a.Family != "rpc-server-current" || b.Family != a.Family || a.Grain != "server-span-record" || b.Grain != a.Grain || a.Query == nil || b.Query == nil {
		return fmt.Errorf("current RPC record family")
	}
	if r.Source.Saved.Template == nil || r.Source.Saved.Template.Profile != "otel.rpc-server-current" || len(r.Source.Saved.Queries) != 1 || !nhLiveSameQuery(r.Source.Saved.Queries[0], *a.Query) || savedJSON360(r.Source.Saved.Composition) != savedJSON360(r.Source.EditedComposition) {
		return fmt.Errorf("saved authored specimen/composition identity")
	}
	if r.Source.Customize.Request.Template != "otel.rpc-server-current-methods" || r.Source.Customize.Request.Widget.Query == nil || !nhLiveSameQuery(*r.Source.Customize.Request.Widget.Query, *a.Query) {
		return fmt.Errorf("source shipped current RPC request")
	}
	customized, err := workbench.ExpandViewTemplate(r.Source.Customize.Request)
	if err != nil || savedJSON360(customized) != savedJSON360(r.Source.Customize.Response) {
		return fmt.Errorf("source shipped template expansion")
	}
	if savedJSON360(r.TemplateAfter) != savedJSON360(r.Source.Saved) || r.Source.Customize.Before.Source != r.Source.Customize.After.Source || r.Source.Customize.Before.SQL != r.Source.Customize.After.SQL || r.Source.Save.Before.Source != r.Source.Save.After.Source || r.Source.Save.Before.SQL != r.Source.Save.After.SQL || r.Source.Customize.After.Source != r.Source.Save.Before.Source {
		return fmt.Errorf("source authoring read telemetry or mutated saved definition")
	}
	if r.Stale.Status != 200 || r.Stale.ClientPlan.Profile != "all-spans" || len(r.Stale.BeforeOldApply) == 0 || savedJSON360(r.Stale.BeforeOldApply) != savedJSON360(r.Stale.AfterOldApply) || savedJSON360(r.Stale.BeforeOldApply) != savedJSON360(r.Stale.AfterReply) {
		return fmt.Errorf("retired factory/Apply acquired replacement owner")
	}
	if len(r.Refusals) != 3 {
		return fmt.Errorf("missing actual role refusal controls")
	}
	for _, refusal := range r.Refusals {
		if refusal.Status != 400 || refusal.Before.Source != refusal.After.Source || refusal.Before.SQL != refusal.After.SQL {
			return fmt.Errorf("counterfactual acquired telemetry")
		}
	}
	q := b.Query
	if q.FromMS != r.Target.Plan.Scope.FromMS || q.ToMS != r.Target.Plan.Scope.ToMS || q.Environment != r.Target.Plan.Scope.Environment || savedJSON360(q.EnvironmentSelector) != savedJSON360(r.Target.Plan.Scope.EnvironmentSelector) || len(q.GroupBy) != 1 || q.GroupBy[0] != (exploration.Field{Kind: "attribute", Key: "rpc.method"}) {
		return fmt.Errorf("target policy/window/group identity")
	}
	request := r.Target.Expansion.Request
	if request.Source == nil || request.Query == nil || savedJSON360(*request.Source) != savedJSON360(r.Source.Saved) || !nhLiveSameQuery(*request.Query, *q) {
		return fmt.Errorf("selected specimen or whole destination request")
	}
	expanded, err := workbench.ExpandViewTemplate(request)
	if err != nil || savedJSON360(expanded) != savedJSON360(r.Target.Expansion.Response) {
		return fmt.Errorf("actual Go authored expansion identity: %v", err)
	}
	if len(expanded.View.Queries) != 1 || !nhLiveSameQuery(expanded.View.Queries[0], *q) || savedJSON360(expanded.View.Composition) != savedJSON360(r.Source.EditedComposition) || savedJSON360(r.Target.StagedComposition) != savedJSON360(r.Source.EditedComposition) || !nhLiveSameQuery(r.Target.Run.Request, *q) || !nhLiveSameQuery(r.Target.Run.Response.Question, *q) {
		return fmt.Errorf("specimen leaked into staged/Run scope or display")
	}
	if len(r.Target.ParentBefore) < 2 || savedJSON360(r.Target.ParentBefore) != savedJSON360(r.Target.ParentAfter) {
		return fmt.Errorf("retained service DOM/clock receipts")
	}
	mode, name := 0, q.Environment
	if s := q.EnvironmentSelector; s != nil {
		name = s.Scope.Value
		mode = 1
		if s.Policy == "effective" {
			mode = 2
			if s.Scope.Mode == "empty" {
				mode = 3
			}
			if s.Scope.Mode == "unreported" {
				mode = 4
			}
		}
	}
	expected := []string{}
	for _, row := range r.Fixture.Rows {
		resource := map[string]string{}
		for k, v := range row.Env {
			resource[k] = v
		}
		resource["service.namespace"] = row.Namespace
		server := string(row.Kind) == "2" || string(row.Kind) == `"Server"`
		if row.Service == r.Target.Plan.Service && row.At >= q.FromMS && row.At < q.ToMS && server && row.Attributes["rpc.system.name"] != "" && continuationEnvironment352(mode, name, resource) && continuationNamespace352(r.Target.Plan.Scope.Namespace, resource) {
			expected = append(expected, row.ID)
		}
	}
	got, err := reportedResultIDs353(r.Target.Run.Response)
	if err != nil {
		return err
	}
	oracle := []string{}
	for _, row := range r.Target.Oracle.Rows {
		oracle = append(oracle, row.ID)
	}
	slices.Sort(expected)
	slices.Sort(oracle)
	if len(expected) == 0 || !slices.Equal(expected, got) || !slices.Equal(expected, oracle) || r.Target.Oracle.Count != fmt.Sprint(len(expected)) || r.Target.Run.Response.Count != fmt.Sprint(len(expected)) {
		return fmt.Errorf("independent raw current RPC membership (method optional)")
	}
	if len(r.Facts) != len(serviceTemplateFactNames356) {
		return fmt.Errorf("incomplete actual target journey")
	}
	for i, f := range r.Facts {
		if f.Name != serviceTemplateFactNames356[i] || f.Before.Calls < 0 || f.After.Calls < f.Before.Calls || f.After.Calls > len(r.Calls) {
			return fmt.Errorf("actual fact boundaries")
		}
		want := 0
		if f.Name == "run-population" {
			want = 1
		}
		count := 0
		for _, call := range r.Calls[f.Before.Calls:f.After.Calls] {
			if serviceTemplateSource356(call) {
				count++
				var actual exploration.Query
				if call.Path != "/api/explore" || json.Unmarshal(call.Body, &actual) != nil || !nhLiveSameQuery(actual, *q) {
					return fmt.Errorf("source read escaped explicit whole question")
				}
			}
		}
		if count != want || f.After.Source-f.Before.Source != want || want == 0 && f.After.SQL != f.Before.SQL {
			return fmt.Errorf("implicit telemetry outside Run")
		}
	}
	return nil
}
func rpcTemplateReplay367(t *testing.T, label string, facts []rpcTemplateFact367) bool {
	t.Helper()
	if len(facts) == 0 {
		t.Fatal("empty actual template facts")
	}
	steps := []string{}
	reads := []int{}
	base := facts[0].Before.Source
	for _, f := range facts {
		steps = append(steps, strings.ReplaceAll(f.Name, "-", "_"))
		reads = append(reads, f.After.Source-base)
	}
	return serviceTemplateModelCheck356(t, "rpc367-"+label, steps, reads, true)
}
func TestQuintCurrentRPCTemplateObserved367(t *testing.T) {
	path := os.Getenv("GOO_CURRENT_RPC_TEMPLATE_367_REPORT")
	if path == "" {
		t.Skip("actual current RPC authored-template report opt-in")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var report rpcTemplateReport367
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if report.Version != "goo.rpcRoleTemplate367/1" {
		t.Fatal("missing actual RPC template report")
	}
	if err = rpcTemplateAdmission367(report); err != nil {
		t.Fatal(err)
	}
	if !rpcTemplateReplay367(t, "observed", report.Facts) {
		t.Fatal("actual template lifecycle rejected")
	}
	counter := slices.Clone(report.Facts)
	counter[1].After.Source++
	if rpcTemplateReplay367(t, "counter-mutant", counter) {
		t.Fatal("implicit read accepted")
	}
	order := slices.Clone(report.Facts)
	order[1], order[4] = order[4], order[1]
	if rpcTemplateReplay367(t, "order-mutant", order) {
		t.Fatal("Run before local Apply accepted")
	}
	bad := report
	query := rpcTemplateCopy367(*bad.Target.Widget.Query)
	query.ToMS++
	bad.Target.Widget.Query = &query
	if rpcTemplateAdmission367(bad) == nil {
		t.Fatal("foreign destination window admitted")
	}
}
