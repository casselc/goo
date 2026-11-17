package exploration

import (
	"reflect"
	"testing"

	"github.com/casselc/goo/identity"
)

func TestHTTPRouteMethodDiagnosisKeepsScopeAndRecordPopulation(t *testing.T) {
	for _, namespace := range []*identity.Namespace{nil, {Mode: "missing"}, {Mode: "value", Value: ""}, {Mode: "value", Value: "team'雪"}} {
		input := dashboardHTTPInput()
		input.Scope.Namespace = namespace
		plan, err := PlanServiceDashboard(input, nil)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, widget := range plan.Widgets {
			if widget.Key != "dimension" {
				continue
			}
			found = true
			want := []Field{{Kind: "attribute", Key: "http.route"}, {Kind: "attribute", Key: "http.request.method"}}
			if !reflect.DeepEqual(widget.Query.GroupBy, want) {
				t.Fatal(widget.Query.GroupBy)
			}
			if widget.Family != "http-server" || widget.Grain != "server-span-record" || widget.Query.FromMS != input.Scope.FromMS || widget.Query.ToMS != input.Scope.ToMS {
				t.Fatal(widget)
			}
			if _, err := Compile(*widget.Query); err != nil {
				t.Fatal(err)
			}
			for _, method := range []string{"GET", "POST", "_OTHER", "雪'"} {
				values := map[Field]string{{Kind: "column", Key: "service"}: input.Service, {Kind: "column", Key: "kind"}: "Server", {Kind: "attribute", Key: "http.request.method"}: method}
				if namespace != nil && namespace.Mode == "value" {
					values[Field{Kind: "resource", Key: "service.namespace"}] = namespace.Value
				}
				if !Match(widget.Query.Where, values) {
					t.Fatalf("route absence discarded an otherwise matching method %q", method)
				}
				values[Field{Kind: "column", Key: "kind"}] = "Client"
				if Match(widget.Query.Where, values) {
					t.Fatal("client included")
				}
			}
		}
		if !found {
			t.Fatal("no shared route/method widget")
		}
	}
}
