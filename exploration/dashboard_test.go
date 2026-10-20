package exploration

import (
	"encoding/json"
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
