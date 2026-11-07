package verification

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

func rpcTemplateQuery367(seed, mode int, service string) exploration.Query {
	q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: int64(1000 + seed), ToMS: int64(61000 + seed), Limit: 37, Summary: "span-duration", SpanStatus: true, GroupBy: []exploration.Field{{Kind: "attribute", Key: "rpc.method"}}, Where: &exploration.Predicate{All: []exploration.Predicate{
		{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: service},
		{Field: &exploration.Field{Kind: "column", Key: "kind"}, Op: "eq", Value: "Server"},
		{Field: &exploration.Field{Kind: "attribute", Key: "rpc.system.name"}, Op: "ne"},
	}}}
	if mode > 0 {
		selector := identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: []string{"", "named", "named", "empty", "unreported"}[mode]}}
		if mode == 1 {
			selector.Policy = "current-only"
		}
		if selector.Scope.Mode == "named" {
			selector.Scope.Value = "destination-π"
		}
		q.Environment = ""
		q.EnvironmentSelector = &selector
	}
	return q
}
func rpcTemplateCopy367(q exploration.Query) exploration.Query {
	raw, _ := json.Marshal(q)
	var out exploration.Query
	json.Unmarshal(raw, &out)
	return out
}
func TestHegelCurrentRPCTemplate367(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 100000))
		mode := hegel.Draw(tc, hegel.Integers(0, 4))
		ns := hegel.Draw(tc, hegel.Integers(0, 2))
		service := fmt.Sprintf("worker'π-%d", seed)
		q := rpcTemplateQuery367(seed, mode, service)
		namespace := exploration.Predicate{Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "missing"}
		if ns > 0 {
			namespace.Op = "eq"
			if ns == 2 {
				namespace.Value = fmt.Sprintf("team-%d", seed)
			}
		}
		q.Where.All = append(q.Where.All, namespace, exploration.Predicate{Any: []exploration.Predicate{{Field: &exploration.Field{Kind: "attribute", Key: "release"}, Op: "missing"}, {Field: &exploration.Field{Kind: "attribute", Key: "release"}, Op: "eq", Value: ""}}})
		widget := exploration.DashboardWidget{Key: "dimension", Family: "rpc-server-current", Grain: "server-span-record", State: "supported", Query: &q}
		before := continuationJSON352(widget)
		expanded, err := workbench.ExpandViewTemplate(workbench.ViewTemplateExpansionRequest{Template: "otel.rpc-server-current-methods", Version: 1, Widget: widget})
		if err != nil {
			tc.Fatal(err)
		}
		if !reflect.DeepEqual(expanded.View.Queries, []exploration.Query{q}) || expanded.View.Template == nil || expanded.View.Template.Profile != "otel.rpc-server-current" || expanded.Template.ID != "otel.rpc-server-current-methods" || continuationJSON352(widget) != before {
			tc.Fatal("shipped RPC template changed whole destination or profile")
		}
		// Raw membership oracle is the official current incoming RPC predicate, not
		// the shipped descriptor: method is optional and legacy system is insufficient.
		for _, role := range []string{"Server", "Client", "Producer", ""} {
			for _, system := range []string{"grpc", "", "MISSING"} {
				for _, method := range []string{"pkg.Service/Call", "_OTHER", "", "MISSING"} {
					fields := map[exploration.Field]string{{Kind: "column", Key: "service"}: service, {Kind: "column", Key: "kind"}: role}
					if system != "MISSING" {
						fields[exploration.Field{Kind: "attribute", Key: "rpc.system.name"}] = system
					}
					if method != "MISSING" {
						fields[exploration.Field{Kind: "attribute", Key: "rpc.method"}] = method
					}
					if ns > 0 {
						fields[exploration.Field{Kind: "resource", Key: "service.namespace"}] = namespace.Value
					}
					fields[exploration.Field{Kind: "attribute", Key: "rpc.system"}] = "legacy-present"
					if hdPredicate(expanded.View.Queries[0].Where, fields) != (role == "Server" && system != "" && system != "MISSING") {
						tc.Fatal("current RPC witness/method optionality changed", role, system, method)
					}
				}
			}
		}
		source := expanded.View
		source.Title = "Reusable current RPC composition"
		source.Queries = []exploration.Query{rpcTemplateQuery367(seed, 0, "specimen-only-service")}
		destination := rpcTemplateCopy367(q)
		destination.FromMS += 200000
		destination.ToMS += 200000
		destination.Limit = 61
		destination.Offset = 200
		destination.GroupBy = []exploration.Field{{Kind: []string{"attribute", "resource"}[seed%2], Key: "rpc.method"}}
		destination.Where.All = append(destination.Where.All, exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "region"}, Op: "eq", Value: "destination-only"})
		inputBefore := continuationJSON352([]any{source, destination})
		out, err := workbench.ExpandViewTemplate(workbench.ViewTemplateExpansionRequest{Source: &source, Query: &destination})
		if err != nil {
			tc.Fatal("reuse current RPC authored composition", err)
		}
		expected := rpcTemplateCopy367(destination)
		expected.Offset = 0
		if !reflect.DeepEqual(out.View.Queries, []exploration.Query{expected}) || !reflect.DeepEqual(out.View.Composition, source.Composition) || continuationJSON352([]any{source, destination}) != inputBefore {
			tc.Fatal("specimen leaked scope/predicate/group or changed display")
		}
		out.View.Queries[0].Where.All[0].Value = "caller edit"
		out.View.Composition.Children[0].Facet.Display.Title = "caller edit"
		if continuationJSON352([]any{source, destination}) != inputBefore {
			tc.Fatal("expanded output aliases source/destination")
		}
		for variant := range 7 {
			bad := rpcTemplateCopy367(destination)
			switch variant {
			case 0:
				bad.Where.All[1].Value = "Client"
			case 1:
				bad.Where.All[2].Field.Key = "rpc.system"
			case 2:
				bad.Where.All[2].Op = "exists"
			case 3:
				bad.Where.All[0].Value = ""
			case 4:
				bad.Signal = "logs"
				bad.Summary = ""
				bad.SpanStatus = false
			case 5:
				bad.Numeric = &exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work"}, Type: "int64", Parse: "canonical-decimal", Unit: "items"}
			case 6:
				bad.GroupTimeSeries = true
				bad.BucketMS = 1000
			}
			if _, err := workbench.ExpandViewTemplate(workbench.ViewTemplateExpansionRequest{Source: &source, Query: &bad}); err == nil {
				tc.Fatal("authored current RPC profile admitted unsupported witness/projection", variant)
			}
		}
		legacy := widget
		legacy.Family = "rpc-server"
		if _, err := workbench.ExpandViewTemplate(workbench.ViewTemplateExpansionRequest{Template: "otel.rpc-server-current-methods", Version: 1, Widget: legacy}); err == nil {
			tc.Fatal("legacy family silently upgraded")
		}
		native := widget
		native.Query = nil
		native.Measurement = &measurements.Question{}
		if _, err := workbench.ExpandViewTemplate(workbench.ViewTemplateExpansionRequest{Template: "otel.rpc-server-current-methods", Version: 1, Widget: native}); err == nil {
			tc.Fatal("native measurement became record template")
		}
	}, options(40)...)
}
