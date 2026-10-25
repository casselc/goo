package exploration

import (
	"github.com/casselc/goo/measurements"
	"testing"
)

func dashboardDBInput() DashboardRequest {
	in := dashboardHTTPInput()
	in.HTTPServerObserved = false
	in.DBClientObserved = true
	in.Fields = []Field{{Kind: "attribute", Key: "db.system.name"}, {Kind: "attribute", Key: "db.operation.name"}}
	return in
}
func TestDashboardDBCurrentClientPopulationAndDimensions(t *testing.T) {
	for _, dimension := range []string{"db.operation.name", "db.namespace", "db.system.name"} {
		in := dashboardDBInput()
		in.Fields = []Field{{Kind: "attribute", Key: "db.system.name"}, {Kind: "attribute", Key: dimension}, {Kind: "attribute", Key: "db.query.summary"}, {Kind: "attribute", Key: "db.statement"}}
		p, e := PlanServiceDashboard(in, nil)
		if e != nil || p.Profile != "db-client" {
			t.Fatal(p, e)
		}
		for _, w := range p.Widgets {
			if w.Query == nil || w.Query.Signal != "spans" {
				continue
			}
			if w.Family != "db-client" || w.Grain != "client-span-record" {
				t.Fatal(w)
			}
			if _, e := Compile(*w.Query); e != nil {
				t.Fatal(e)
			}
			if w.Key == "dimension" && w.Query.GroupBy[0] != (Field{Kind: "attribute", Key: dimension}) {
				t.Fatal(w)
			}
			for _, kind := range []string{"Client", "Server", "Internal", "SPAN_KIND_CLIENT", ""} {
				for _, system := range []string{"postgresql", "", " "} {
					v := map[Field]string{{Kind: "column", Key: "service"}: "api", {Kind: "column", Key: "kind"}: kind, {Kind: "attribute", Key: "db.system.name"}: system, {Kind: "attribute", Key: "db.system"}: "mysql"}
					if got, want := Match(w.Query.Where, v), kind == "Client" && system != ""; got != want {
						t.Fatalf("%q %q", kind, system)
					}
					// Optional dimensions do not restrict the population, even missing/empty.
					v[Field{Kind: "attribute", Key: dimension}] = ""
					if dimension != "db.system.name" && Match(w.Query.Where, v) != (kind == "Client" && system != "") {
						t.Fatal("dimension became admission condition")
					}
					delete(v, Field{Kind: "attribute", Key: "db.system.name"})
					if Match(w.Query.Where, v) {
						t.Fatal("legacy alias admitted")
					}
				}
			}
		}
	}
}
func TestDashboardDBHintsPriorityAndNoAlias(t *testing.T) {
	in := dashboardDBInput()
	in.DBClientObserved = false
	p, e := PlanServiceDashboard(in, nil)
	if e != nil || p.Profile != "all-spans" {
		t.Fatal(p, e)
	}
	in.Profile = "db-client"
	p, e = PlanServiceDashboard(in, nil)
	if e != nil || p.Profile != "db-client" {
		t.Fatal(p, e)
	}
	in.Profile = "auto"
	in.DBClientObserved = true
	in.RPCServerObserved = true
	in.Fields = append(in.Fields, Field{Kind: "attribute", Key: "rpc.system"}, Field{Kind: "attribute", Key: "rpc.method"})
	p, e = PlanServiceDashboard(in, nil)
	if e != nil || p.Profile != "rpc-server" {
		t.Fatal(p, e)
	}
	in.HTTPServerObserved = true
	in.Fields = append(in.Fields, Field{Kind: "attribute", Key: "http.request.method"})
	p, e = PlanServiceDashboard(in, nil)
	if e != nil || p.Profile != "http-server" {
		t.Fatal(p, e)
	}
	for _, f := range []Field{{Kind: "attribute", Key: "db.system"}, {Kind: "resource", Key: "db.system.name"}, {Kind: "attribute", Key: "db.query.summary"}} {
		in := dashboardDBInput()
		in.Fields = []Field{f}
		p, e = PlanServiceDashboard(in, nil)
		if e != nil || p.Profile != "all-spans" {
			t.Fatal(p, e)
		}
		in.Profile = "db-client"
		if _, e = PlanServiceDashboard(in, nil); e == nil {
			t.Fatal("noncurrent candidate admitted")
		}
	}
}
func TestDashboardDBNativeExactSecondsAndBoundedPriority(t *testing.T) {
	in := dashboardDBInput()
	in.Fields = nil
	in.DBClientObserved = false
	in.Signals = append(in.Signals, "metrics")
	db := dashboardTestInstrument("z-db", "db.client.operation.duration", "histogram", "s")
	in.Instruments = []measurements.Instrument{dashboardTestInstrument("a-gauge", "load", "gauge", "1"), db, dashboardTestInstrument("b-gauge", "other", "gauge", "1")}
	p, e := PlanServiceDashboard(in, nil)
	if e != nil || p.Profile != "db-client" || p.NativeOptions[0].Instrument.Metric != "db.client.operation.duration" {
		t.Fatal(p, e)
	}
	selected := 0
	for _, option := range p.NativeOptions {
		if option.Selected {
			selected++
		}
	}
	if selected != 2 || p.NativeLimit != 2 {
		t.Fatal("native budget changed")
	}
	for _, bad := range []measurements.Instrument{dashboardTestInstrument("old", "db.client.duration", "histogram", "ms"), dashboardTestInstrument("unit", "db.client.operation.duration", "histogram", "ms")} {
		if dashboardNativeDB(in.Scope, in.Service, bad) {
			t.Fatal("unsupported native binding admitted")
		}
	}
	d := *db.Definition
	d.Environment = "other"
	db.Definition = &d
	if dashboardNativeDB(in.Scope, in.Service, db) {
		t.Fatal("scope mismatch admitted")
	}
}
