package annotations

import (
	"github.com/casselc/goo/identity"
	"strings"
	"testing"
)

func TestOperationNamespaceBinding(t *testing.T) {
	m, e := ScanDir("../examples/operation-app")
	if e != nil {
		t.Fatal(e)
	}
	for _, ns := range []*identity.Namespace{{Mode: "missing"}, {Mode: "value"}, {Mode: "value", Value: "payments'"}} {
		source := OperationSource{Service: "api", Environment: "prod", Instance: "one", DeclaredSchema: m.SchemaURL, Namespace: ns}
		sql, params, e := m.OperationQuery(source, 1, 2)
		if e != nil || !strings.Contains(sql, "mapContains") || params["namespace"] != ns.Value {
			t.Fatal(sql, params, e)
		}
		row := OperationRow{TimestampNS: "1000000", DurationNS: "1", TraceID: strings.Repeat("1", 32), SpanID: strings.Repeat("2", 16), Kind: "Server", Scope: m.Scope, Version: m.ScopeVersion, Resource: map[string]string{"service.name": "api", "deployment.environment.name": "prod", "service.instance.id": "one"}, Attributes: map[string]string{"goo.operation.name": "checkout.reserve", "goo.operation.role": "server", "goo.operation.outcome": "ok"}}
		if ns.Mode == "value" {
			row.Resource[identity.NamespaceKey] = ns.Value
		}
		if _, _, e = m.ObserveRows(source, 1, 2, []OperationRow{row}); e != nil {
			t.Fatal(e)
		}
		row.Resource[identity.NamespaceKey] = "foreign"
		if _, _, e = m.ObserveRows(source, 1, 2, []OperationRow{row}); e == nil {
			t.Fatal("foreign namespace accepted")
		}
	}
}
