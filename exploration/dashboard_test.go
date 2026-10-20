package exploration

import (
	"encoding/json"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"reflect"
	"strings"
	"testing"
)

func TestDashboardPlanScopeCapabilitiesAndDetachment(t *testing.T) {
	in := DashboardRequest{Scope: InventoryQuestion{Environment: "prod", FromMS: 1000, ToMS: 61000}, Service: "api'\\n", Signals: []string{"logs", "spans"}, Fields: []Field{{Kind: "attribute", Key: "db.system"}, {Kind: "attribute", Key: "http.route"}}}
	before, _ := json.Marshal(in)
	p, e := PlanServiceDashboard(in, []DashboardDeclaration{{Name: "x", Revision: "r", Service: in.Service, Environment: "other"}})
	if e != nil || len(p.Widgets) != 3 || p.Widgets[2].Query.GroupBy[0].Key != "http.route" || len(p.Declarations) != 0 {
		t.Fatal(p, e)
	}
	for _, w := range p.Widgets {
		if w.Query == nil || w.Query.Validate() != nil || w.Query.Where.Value != in.Service || w.Query.FromMS != in.Scope.FromMS {
			t.Fatal(w)
		}
	}
	in.Fields[0], in.Fields[1] = in.Fields[1], in.Fields[0]
	in.Signals[0], in.Signals[1] = in.Signals[1], in.Signals[0]
	again, e := PlanServiceDashboard(in, nil)
	if e != nil || !reflect.DeepEqual(p, again) {
		t.Fatal("order changes plan")
	}
	in.Fields[0].Key = "changed"
	after, _ := json.Marshal(in)
	if string(before) == string(after) || p.Widgets[2].Query.GroupBy[0].Key != "http.route" {
		t.Fatal("aliasing")
	}
}
func TestDashboardMissingSignalsBudgetAndNativeIdentity(t *testing.T) {
	in := DashboardRequest{Scope: InventoryQuestion{Environment: "prod", FromMS: 1000, ToMS: 61000}, Service: "api"}
	p, e := PlanServiceDashboard(in, nil)
	if e != nil || p.Widgets[0].State != "not_observed" || p.Widgets[0].Query != nil {
		t.Fatal(p, e)
	}
	in.Signals = []string{"metrics"}
	for _, name := range []string{"c", "b", "a"} {
		d := measurements.Definition{APIVersion: measurements.Version, Name: name, Title: name, Kind: "gauge", Environment: "prod", LookbackMinutes: 1, MaxGapMS: 60000, Grain: "per-stream", Source: measurements.Binding{Services: []string{"api"}, Scope: "sdk", Metric: name, Unit: "1"}}
		in.Instruments = append(in.Instruments, measurements.Instrument{Kind: "gauge", Metric: name, Unit: "1", Scope: "sdk", Supported: true, Definition: &d})
	}
	p, e = PlanServiceDashboard(in, nil)
	if e != nil || len(p.Widgets) != 4 || p.Widgets[2].Measurement.Definition.Source.Metric != "a" {
		t.Fatal(p, e)
	}
	in.Instruments[0].Definition.Source.Unit = "s"
	if _, e = PlanServiceDashboard(in, nil); e == nil {
		t.Fatal("mixed unit metadata accepted")
	}
	in.Instruments = nil
	in.Fields = make([]Field, 129)
	if _, e = PlanServiceDashboard(in, nil); e == nil {
		t.Fatal("field budget ignored")
	}
	in.Fields = nil
	in.Service = strings.Repeat("x", 129)
	if _, e = PlanServiceDashboard(in, nil); e == nil {
		t.Fatal("service budget ignored")
	}
}

func TestDashboardRejectsUnsupportedNativeTemporalityAndKind(t *testing.T) {
	in := DashboardRequest{Scope: InventoryQuestion{Environment: "prod", FromMS: 1000, ToMS: 61000}, Service: "api", Signals: []string{"metrics"}}
	for _, bad := range []struct {
		kind        string
		temporality int
		monotonic   bool
	}{{"gauge", 1, false}, {"counter", 2, false}, {"histogram", 3, false}, {"lifecycle", 0, false}} {
		d := measurements.Definition{APIVersion: measurements.Version, Name: "metric", Title: "Metric", Kind: bad.kind, Environment: "prod", LookbackMinutes: 1, MaxGapMS: 60000, Grain: "per-stream", Source: measurements.Binding{Services: []string{"api"}, Scope: "sdk", Metric: "metric", Unit: "ms"}}
		in.Instruments = []measurements.Instrument{{Kind: bad.kind, Temporality: bad.temporality, Monotonic: bad.monotonic, Metric: "metric", Unit: "ms", Scope: "sdk", Supported: true, Definition: &d}}
		if _, err := PlanServiceDashboard(in, nil); err == nil {
			t.Fatal("unsupported metadata accepted", bad)
		}
	}
}

func TestNamespaceDashboardAndDeclaredBindingQuestion(t *testing.T) {
	ns := &identity.Namespace{Mode: "value", Value: "payments"}
	in := DashboardRequest{Scope: InventoryQuestion{Namespace: ns, Environment: "prod", FromMS: 1000, ToMS: 61000}, Service: "api", Signals: []string{"spans"}}
	d := DashboardDeclaration{Name: "contract", Title: "Contract", Revision: strings.Repeat("a", 64), Service: "api", Environment: "prod", Instance: "one", Scope: "goo.sdk", Version: "1", SchemaURL: "goo://v1", Operations: []DashboardOperation{{Name: "work", Role: "server"}}}
	p, e := PlanServiceDashboard(in, []DashboardDeclaration{d})
	if e != nil {
		t.Fatal(e)
	}
	w := p.Widgets[len(p.Widgets)-1]
	if w.State != "supported" || w.Query == nil || p.Declarations[0].Match != "namespace-unspecified" {
		t.Fatal(p)
	}
	plan, e := Compile(*w.Query)
	if e != nil || !strings.Contains(plan.Count, "ScopeVersion") || strings.Contains(plan.Count, "ScopeSchemaUrl") || strings.Contains(plan.Count, "payments") || !strings.Contains(plan.Count, "mapContains") {
		t.Fatal(plan, e)
	}
	resource := map[Field]string{{Kind: "column", Key: "service"}: "api", {Kind: "resource", Key: "service.namespace"}: "payments", {Kind: "resource", Key: "service.instance.id"}: "one", {Kind: "column", Key: "scope"}: "goo.sdk", {Kind: "column", Key: "scope_version"}: "1", {Kind: "attribute", Key: "goo.operation.name"}: "work", {Kind: "attribute", Key: "goo.operation.role"}: "server"}
	if !Match(w.Query.Where, resource) {
		t.Fatal("valid binding rejected")
	}
	resource[Field{Kind: "resource", Key: "service.namespace"}] = "other"
	if Match(w.Query.Where, resource) {
		t.Fatal("namespace leak")
	}
	ns.Value = "changed"
	if p.Scope.Namespace.Value != "payments" {
		t.Fatal("plan aliases selector")
	}
	d.Namespace = &identity.Namespace{Mode: "missing"}
	p, e = PlanServiceDashboard(in, []DashboardDeclaration{d})
	if e != nil || p.Widgets[len(p.Widgets)-1].State != "unsupported" {
		t.Fatal("mismatched declaration executable", p, e)
	}
}

func TestDashboardExplicitNativeSelection(t *testing.T) {
	in := DashboardRequest{Scope: InventoryQuestion{Environment: "prod", FromMS: 1000, ToMS: 61000}, Service: "api", Signals: []string{"metrics"}}
	for _, name := range []string{"a", "b", "c", "d"} {
		d := measurements.Definition{APIVersion: measurements.Version, Name: name, Title: name, Kind: "gauge", Environment: "prod", LookbackMinutes: 1, MaxGapMS: 60000, Grain: "per-stream", Source: measurements.Binding{Services: []string{"api"}, Scope: "sdk", Metric: name, Unit: "1"}}
		in.Instruments = append(in.Instruments, measurements.Instrument{Kind: "gauge", Metric: name, Unit: "1", Scope: "sdk", Supported: true, Definition: &d})
	}
	defaults, e := PlanServiceDashboard(in, nil)
	if e != nil {
		t.Fatal(e)
	}
	if defaults.NativeLimit != 2 || len(defaults.NativeOptions) != 4 {
		t.Fatal(defaults)
	}
	selected := []string{defaults.NativeOptions[3].Key, defaults.NativeOptions[2].Key}
	in.NativeSelection = &selected
	p, e := PlanServiceDashboard(in, nil)
	if e != nil {
		t.Fatal(e)
	}
	if len(p.Widgets) != 4 || p.Widgets[2].Measurement.Definition.Source.Metric != "c" || p.Widgets[3].Measurement.Definition.Source.Metric != "d" {
		t.Fatal(p)
	}
	selected[0], selected[1] = selected[1], selected[0]
	in.Instruments[0], in.Instruments[3] = in.Instruments[3], in.Instruments[0]
	again, e := PlanServiceDashboard(in, nil)
	if e != nil || !reflect.DeepEqual(p, again) {
		t.Fatal("ordering changed explicit selection", e)
	}
	for _, bad := range [][]string{{selected[0], selected[0]}, {strings.Repeat("f", 64)}, {selected[0], selected[1], defaults.NativeOptions[0].Key}} {
		in.NativeSelection = &bad
		if _, e := PlanServiceDashboard(in, nil); e == nil {
			t.Fatal("invalid selection accepted", bad)
		}
	}
	empty := []string{}
	in.NativeSelection = &empty
	p, e = PlanServiceDashboard(in, nil)
	if e != nil || len(p.Widgets) != 2 {
		t.Fatal("empty selection caused native reads", p, e)
	}
	// A key is a scoped definition, never just a metric name or discovery row index.
	in.NativeSelection = &selected
	in.Scope.Environment = "other"
	if _, e := PlanServiceDashboard(in, nil); e == nil {
		t.Fatal("cross-environment selection accepted")
	}
}
