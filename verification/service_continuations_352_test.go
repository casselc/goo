package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

// Deliberately independent of EnvironmentSelection.Matches: this is a reference
// over raw OTel map presence, including current/legacy conflicts and empty values.
func continuationEnvironment352(mode int, name string, resource map[string]string) bool {
	current, hasCurrent := resource["deployment.environment.name"]
	legacy, hasLegacy := resource["deployment.environment"]
	if mode == 0 || mode == 1 {
		return hasCurrent && current == name
	}
	if mode == 4 {
		return !hasCurrent && !hasLegacy
	}
	value, present := current, hasCurrent
	if !present {
		value, present = legacy, hasLegacy
	}
	if mode == 3 {
		return present && value == ""
	}
	return present && value == name
}

func continuationNamespace352(want *identity.Namespace, resource map[string]string) bool {
	if want == nil {
		return true
	}
	value, present := resource["service.namespace"]
	if want.Mode == "missing" {
		return !present
	}
	return present && value == want.Value
}

func continuationJSON352(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(raw)
}

func TestHegelServiceContinuations352(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		mode := hegel.Draw(tc, hegel.Integers(0, 4))
		nsMode := hegel.Draw(tc, hegel.Integers(0, 2))
		qualified := hegel.Draw(tc, hegel.Booleans())
		scopedNamespace := hegel.Draw(tc, hegel.Booleans())
		emptyService := hegel.Draw(tc, hegel.Booleans())
		name, service := fmt.Sprintf("env-'雪-%d", seed), fmt.Sprintf("child-'雪-%d", seed)
		if emptyService {
			service = ""
		}
		from, to := int64(1800000000000+seed*1000), int64(1800000060000+seed*1000)
		scope := exploration.InventoryQuestion{Environment: name, FromMS: from, ToMS: to}
		if mode > 0 {
			scope.Environment = ""
			scope.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: name}}
			switch mode {
			case 1:
				scope.EnvironmentSelector.Policy = "current-only"
			case 3:
				scope.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "empty"}
			case 4:
				scope.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "unreported"}
			}
		}
		namespace := &identity.Namespace{Mode: "missing"}
		edge := map[string]any{"parent": "upstream-" + fmt.Sprint(seed), "child": service, "kind": "edge", "count": "7"}
		if nsMode > 0 {
			namespace.Mode = "value"
			if nsMode == 2 {
				namespace.Value = fmt.Sprintf("team-'雪-%d", seed)
			}
			edge["childNamespace"], edge["parentNamespace"] = namespace.Value, namespace.Value
		}
		if scopedNamespace {
			scope.Namespace = namespace.Clone()
		}
		selectedNamespace := scope.Namespace
		if qualified {
			selectedNamespace = namespace
		}
		savedWhere := &exploration.Predicate{All: []exploration.Predicate{
			{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "saved-other-service"},
			{Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "eq", Value: "saved-other-namespace"},
			{Field: &exploration.Field{Kind: "column", Key: "body"}, Op: "contains", Value: "saved-only predicate"},
		}}
		view := workbench.ActivityView{Version: 1, ID: strings.Repeat("a", 64), Title: "Independent saved view", Selector: exploration.TailSelector{Signal: "logs", Environment: scope.Environment, EnvironmentSelector: scope.EnvironmentSelector, Where: savedWhere}, Columns: []string{"utc", "service", "summary"}}
		foreign := map[string]any{"environmentSelector": identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: name}}}
		if mode == 2 {
			foreign = map[string]any{"environment": name}
		} else if mode == 3 || mode == 4 {
			foreign = map[string]any{"environmentSelector": identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: name}}}
		}
		resources := []map[string]string{
			{}, {"deployment.environment.name": name}, {"deployment.environment": name},
			{"deployment.environment.name": "foreign", "deployment.environment": name},
			{"deployment.environment.name": "", "deployment.environment": name},
			{"deployment.environment": ""},
			{"deployment.environment.name": name, "deployment.environment": "foreign"},
		}
		input, _ := json.Marshal(map[string]any{"scope": scope, "edge": edge, "qualified": qualified, "view": view, "foreignScope": foreign, "resources": resources})
		cmd := exec.CommandContext(t.Context(), "node", "service_continuations_352_fixture.mjs")
		cmd.Stdin = bytes.NewReader(input)
		raw, err := cmd.CombinedOutput()
		if err != nil {
			tc.Fatal("actual UI continuation helpers", err, string(raw))
			return
		}
		var got struct {
			Wider, Exact                           exploration.Query
			ExactRefused, OwnPolicy, ForeignPolicy bool
			Unchanged, Detached                    bool
			ExactReason                            string
			EnvironmentMatches                     []bool
			Filter                                 map[string]string
			Saved                                  workbench.ActivityView
		}
		if err = json.Unmarshal(raw, &got); err != nil || !got.Unchanged || !got.Detached || !got.OwnPolicy || got.ForeignPolicy {
			tc.Fatal("scope, saved selector or caller detachment lost", err, string(raw))
			return
		}
		if err = got.Wider.Validate(); err != nil {
			tc.Fatal("Go rejected actual wider child query", err)
			return
		}
		if got.Wider.Environment != scope.Environment || continuationJSON352(got.Wider.EnvironmentSelector) != continuationJSON352(scope.EnvironmentSelector) || got.Wider.FromMS != from || got.Wider.ToMS != to || got.Wider.Signal != "spans" || got.Wider.Limit != 100 || got.Wider.Offset != 0 || got.Wider.ReportedEdge != nil {
			tc.Fatal("wider child is not the fixed original policy/window", got.Wider)
			return
		}
		compiled, err := exploration.Compile(got.Wider)
		if err != nil || strings.Contains(compiled.Count+compiled.Rows+compiled.Groups, name) || service != "" && strings.Contains(compiled.Count+compiled.Rows+compiled.Groups, service) {
			tc.Fatal("literal scope interpolated or query did not compile", err)
			return
		}
		selection, _ := got.Wider.EnvironmentSelection()
		for i, resource := range resources {
			wantEnvironment := continuationEnvironment352(mode, name, resource)
			if got.EnvironmentMatches[i] != wantEnvironment || selection.Matches(resource) != wantEnvironment {
				tc.Fatal("current/legacy policy changed", mode, resource)
				return
			}
			for _, candidateNamespace := range []map[string]string{{}, {"service.namespace": ""}, {"service.namespace": namespace.Value}, {"service.namespace": "foreign"}} {
				combined := map[string]string{}
				for key, value := range resource {
					combined[key] = value
				}
				for key, value := range candidateNamespace {
					combined[key] = value
				}
				for _, candidateService := range []string{service, "foreign-service"} {
					fields := map[exploration.Field]string{{Kind: "column", Key: "service"}: candidateService}
					if value, present := combined["service.namespace"]; present {
						fields[exploration.Field{Kind: "resource", Key: "service.namespace"}] = value
					}
					for _, timestamp := range []int64{from - 1, from, to - 1, to} {
						actual := selection.Matches(combined) && exploration.Match(got.Wider.Where, fields) && timestamp >= got.Wider.FromMS && timestamp < got.Wider.ToMS
						want := wantEnvironment && candidateService == service && continuationNamespace352(selectedNamespace, combined) && timestamp >= from && timestamp < to
						if actual != want {
							tc.Fatal("child population differs from independent membership", mode, combined, candidateService, timestamp)
							return
						}
					}
				}
			}
		}
		if mode > 0 {
			if !got.ExactRefused || got.ExactReason == "" || got.Exact.ReportedEdge != nil {
				tc.Fatal("unsupported typed exact edge was constructed or silently projected", got.ExactRefused, got.ExactReason)
				return
			}
			exact := got.Wider
			exact.ReportedEdge = &exploration.ReportedEdge{Parent: exploration.EdgeService{Service: edge["parent"].(string), Namespace: namespace.Clone()}, Child: exploration.EdgeService{Service: service, Namespace: namespace.Clone()}}
			if exact.Validate() == nil {
				tc.Fatal("backend unexpectedly admitted unsupported typed exact edge")
				return
			}
		} else if got.ExactRefused || got.Exact.Validate() != nil || got.Exact.ReportedEdge == nil || got.Exact.Environment != name || got.Exact.FromMS != from || got.Exact.ToMS != to {
			tc.Fatal("supported scalar exact edge changed", got.ExactReason)
			return
		}
		if got.Saved.Validate() != nil || continuationJSON352(got.Saved.Selector) != continuationJSON352(view.Selector) || len(got.Filter) != 1 {
			tc.Fatal("saved view inherited the service filter or fixed window")
			return
		}
		if mode == 0 {
			if got.Filter["environment"] != name {
				tc.Fatal("scalar catalog scope changed")
			}
		} else {
			var selector identity.EnvironmentSelector
			if json.Unmarshal([]byte(got.Filter["environmentSelector"]), &selector) != nil || continuationJSON352(selector) != continuationJSON352(*scope.EnvironmentSelector) {
				tc.Fatal("typed Activity catalog filter lost policy")
			}
		}
	}, options(40)...)
}
