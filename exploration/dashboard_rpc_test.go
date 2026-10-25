package exploration

import (
	"encoding/json"
	"github.com/casselc/goo/measurements"
	"strings"
	"testing"
)

func dashboardRPCInput() DashboardRequest {
	in := dashboardHTTPInput()
	in.HTTPServerObserved = false
	in.RPCServerObserved = true
	in.Fields = []Field{{Kind: "attribute", Key: "rpc.system"}, {Kind: "attribute", Key: "rpc.method"}}
	return in
}
func TestDashboardRPCLegacyExactPopulation(t *testing.T) {
	in := dashboardRPCInput()
	p, err := PlanServiceDashboard(in, nil)
	if err != nil || p.Profile != "rpc-server" {
		t.Fatal(p, err)
	}
	for _, w := range p.Widgets {
		if w.Query == nil || w.Query.Signal != "spans" {
			continue
		}
		if w.Family != "rpc-server" || w.Grain != "server-span-record" {
			t.Fatal(w)
		}
		if _, err := Compile(*w.Query); err != nil {
			t.Fatal(err)
		}
		for _, kind := range []string{"Server", "Client", "Internal", "SPAN_KIND_SERVER", ""} {
			for _, system := range []string{"grpc", "", " "} {
				for _, method := range []string{"Get", "", " "} {
					values := map[Field]string{{Kind: "column", Key: "service"}: "api", {Kind: "column", Key: "kind"}: kind, {Kind: "attribute", Key: "rpc.system"}: system, {Kind: "attribute", Key: "rpc.method"}: method}
					if got, want := Match(w.Query.Where, values), kind == "Server" && system != "" && method != ""; got != want {
						t.Fatalf("%s %q %q", kind, system, method)
					}
					delete(values, Field{Kind: "attribute", Key: "rpc.system"})
					if Match(w.Query.Where, values) {
						t.Fatal("missing system admitted")
					}
					values[Field{Kind: "attribute", Key: "rpc.system"}] = system
					delete(values, Field{Kind: "attribute", Key: "rpc.method"})
					if Match(w.Query.Where, values) {
						t.Fatal("missing method admitted")
					}
				}
			}
		}
		if w.Key == "dimension" && w.Query.GroupBy[0] != (Field{Kind: "attribute", Key: "rpc.method"}) {
			t.Fatal(w)
		}
	}
}
func TestDashboardRPCCandidatesPriorityAndLegacyOmission(t *testing.T) {
	in := dashboardRPCInput()
	in.RPCServerObserved = false
	p, e := PlanServiceDashboard(in, nil)
	if e != nil || p.Profile != "all-spans" {
		t.Fatal(p, e)
	}
	in.Profile = "rpc-server"
	p, e = PlanServiceDashboard(in, nil)
	if e != nil || p.Profile != "rpc-server" {
		t.Fatal(p, e)
	}
	in.Profile = "auto"
	in.RPCServerObserved = true
	in.HTTPServerObserved = true
	in.Fields = append(in.Fields, Field{Kind: "attribute", Key: "http.request.method"})
	p, e = PlanServiceDashboard(in, nil)
	if e != nil || p.Profile != "http-server" {
		t.Fatal(p, e)
	}
	for _, fields := range [][]Field{{{Kind: "attribute", Key: "rpc.system.name"}, {Kind: "attribute", Key: "rpc.method"}}, {{Kind: "resource", Key: "rpc.system"}, {Kind: "attribute", Key: "rpc.method"}}, {{Kind: "attribute", Key: "rpc.system"}}} {
		in := dashboardRPCInput()
		in.Fields = fields
		p, e = PlanServiceDashboard(in, nil)
		if e != nil || p.Profile != "all-spans" {
			t.Fatal(p, e)
		}
		in.Profile = "rpc-server"
		if _, e = PlanServiceDashboard(in, nil); e == nil {
			t.Fatal("unsupported profile admitted")
		}
	}
	in = dashboardRPCInput()
	in.Profile = ""
	p, e = PlanServiceDashboard(in, nil)
	if e != nil {
		t.Fatal(e)
	}
	b, _ := json.Marshal(p)
	if strings.Contains(string(b), "families") || strings.Contains(string(b), "rpc-server") {
		t.Fatal("legacy changed")
	}
}
func TestDashboardRPCNativeExactLegacyAndPriority(t *testing.T) {
	in := dashboardRPCInput()
	in.Fields = nil
	in.RPCServerObserved = false
	in.Signals = append(in.Signals, "metrics")
	rpc := dashboardTestInstrument("z-rpc", "rpc.server.duration", "histogram", "ms")
	in.Instruments = []measurements.Instrument{dashboardTestInstrument("a-gauge", "load", "gauge", "1"), rpc, dashboardTestInstrument("b-gauge", "other", "gauge", "1")}
	p, e := PlanServiceDashboard(in, nil)
	if e != nil || p.Profile != "rpc-server" || p.NativeOptions[0].Instrument.Metric != "rpc.server.duration" || len(p.NativeOptions) != 3 {
		t.Fatal(p, e)
	}
	for _, bad := range []measurements.Instrument{dashboardTestInstrument("modern", "rpc.server.call.duration", "histogram", "s"), dashboardTestInstrument("wrong-unit", "rpc.server.duration", "histogram", "s")} {
		if dashboardNativeRPC(in.Scope, in.Service, bad) {
			t.Fatal("unapproved binding eligible")
		}
	}
	corrupt := rpc
	d := *rpc.Definition
	d.Source.Metric = "other"
	corrupt.Definition = &d
	if dashboardNativeRPC(in.Scope, in.Service, corrupt) {
		t.Fatal("identity mismatch eligible")
	}
}
