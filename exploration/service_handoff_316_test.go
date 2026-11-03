package exploration

import (
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"reflect"
	"testing"
)

// The service browser's legacy scope is a lossless encoding only of a checked
// current-only named environment. Compare executable plans, not display labels.
func TestServiceHandoffCurrentNamedPlanParity316(t *testing.T) {
	for _, ns := range []*identity.Namespace{nil, {Mode: "missing"}, {Mode: "value"}, {Mode: "value", Value: "team'雪"}} {
		scope := InventoryQuestion{Environment: "prod'雪", FromMS: 100000, ToMS: 160000, Namespace: ns}
		for _, signal := range []string{"logs", "spans"} {
			legacy := DashboardRecordQuestion(scope, "api", signal)
			selected, _ := identity.CurrentEnvironment(scope.Environment)
			selector, _ := selected.Selector()
			typed := legacy
			typed.Environment = ""
			typed.EnvironmentSelector = &selector
			if !legacy.SameEnvironment(typed) {
				t.Fatal("checked environment changed")
			}
			a, err := Compile(legacy)
			if err != nil {
				t.Fatal(err)
			}
			b, err := Compile(typed)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(a, b) {
				t.Fatal("typed current selector changed executable plan")
			}
			effective, _ := identity.EffectiveEnvironment(selected.Scope())
			other, _ := effective.Selector()
			typed.EnvironmentSelector = &other
			if legacy.SameEnvironment(typed) {
				t.Fatal("effective fallback conflated with current-only")
			}
		}
		if _, err := CompileInventory(scope); err != nil {
			t.Fatal(err)
		}
		if _, err := CompileServiceDiscovery(scope); err != nil {
			t.Fatal(err)
		}
		if err := (measurements.DiscoveryQuestion{Environment: scope.Environment, Service: "api", Namespace: ns, FromMS: scope.FromMS, ToMS: scope.ToMS}).Validate(); err != nil {
			t.Fatal(err)
		}
	}
	missing := &identity.Namespace{Mode: "missing"}
	empty := &identity.Namespace{Mode: "value"}
	if missing.Equal(empty) || missing.Equal(nil) || empty.Equal(nil) {
		t.Fatal("namespace identities conflated")
	}
	if !missing.Matches(map[string]string{}) || missing.Matches(map[string]string{identity.NamespaceKey: ""}) || !empty.Matches(map[string]string{identity.NamespaceKey: ""}) || empty.Matches(map[string]string{}) {
		t.Fatal("namespace reporting distinction lost")
	}
	// An empty reported ServiceName is a record identity, but native metric
	// bindings currently require a nonempty service. Do not turn it into missing.
	scope := InventoryQuestion{Environment: "prod", FromMS: 100000, ToMS: 160000}
	if _, err := Compile(DashboardRecordQuestion(scope, "", "spans")); err != nil {
		t.Fatal(err)
	}
	if err := (measurements.DiscoveryQuestion{Environment: "prod", Service: "", FromMS: scope.FromMS, ToMS: scope.ToMS}).Validate(); err == nil {
		t.Fatal("empty metric service unexpectedly admitted")
	}
}
