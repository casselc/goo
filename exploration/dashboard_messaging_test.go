package exploration

import (
	"github.com/casselc/goo/measurements"
	"testing"
)

func dashboardMessagingInput() DashboardRequest {
	in := dashboardHTTPInput()
	in.HTTPServerObserved = false
	in.MessagingConsumerObserved = true
	in.Fields = []Field{{Kind: "attribute", Key: "messaging.system"}}
	return in
}
func TestDashboardMessagingRoleAndSystemPopulation(t *testing.T) {
	for _, role := range []string{"consumer", "producer"} {
		for _, dimension := range []string{"messaging.destination.name", "messaging.operation.name", "messaging.system"} {
			in := dashboardMessagingInput()
			in.Profile = "messaging-" + role
			in.Fields = append(in.Fields, Field{Kind: "attribute", Key: dimension})
			p, e := PlanServiceDashboard(in, nil)
			if e != nil || p.Profile != in.Profile {
				t.Fatal(p, e)
			}
			for _, w := range p.Widgets {
				if w.Query == nil || w.Query.Signal != "spans" {
					continue
				}
				if w.Family != in.Profile || w.Grain != role+"-span-record" {
					t.Fatal(w)
				}
				if _, e := Compile(*w.Query); e != nil {
					t.Fatal(e)
				}
				if w.Key == "dimension" && w.Query.GroupBy[0] != (Field{Kind: "attribute", Key: dimension}) {
					t.Fatal(w)
				}
				for _, kind := range []string{"Consumer", "Producer", "Client", "Server", "Internal", "SPAN_KIND_CONSUMER", ""} {
					for _, system := range []string{"kafka", "", " "} {
						v := map[Field]string{{Kind: "column", Key: "service"}: "api", {Kind: "column", Key: "kind"}: kind, {Kind: "attribute", Key: "messaging.system"}: system}
						wantKind := "Consumer"
						if role == "producer" {
							wantKind = "Producer"
						}
						want := kind == wantKind && system != ""
						if Match(w.Query.Where, v) != want {
							t.Fatalf("%s %q %q", role, kind, system)
						}
						if dimension != "messaging.system" {
							v[Field{Kind: "attribute", Key: dimension}] = ""
							if Match(w.Query.Where, v) != want {
								t.Fatal("optional dimension narrows")
							}
						}
						delete(v, Field{Kind: "attribute", Key: "messaging.system"})
						if Match(w.Query.Where, v) {
							t.Fatal("missing system admitted")
						}
					}
				}
			}
		}
	}
}
func TestDashboardMessagingHintsAndPrecedence(t *testing.T) {
	in := dashboardMessagingInput()
	in.MessagingConsumerObserved = false
	p, e := PlanServiceDashboard(in, nil)
	if e != nil || p.Profile != "all-spans" {
		t.Fatal(p, e)
	}
	in.MessagingProducerObserved = true
	p, e = PlanServiceDashboard(in, nil)
	if e != nil || p.Profile != "messaging-producer" {
		t.Fatal(p, e)
	}
	in.MessagingConsumerObserved = true
	p, e = PlanServiceDashboard(in, nil)
	if e != nil || p.Profile != "messaging-consumer" {
		t.Fatal(p, e)
	}
	in.DBClientObserved = true
	in.Fields = append(in.Fields, Field{Kind: "attribute", Key: "db.system.name"})
	p, e = PlanServiceDashboard(in, nil)
	if e != nil || p.Profile != "db-client" {
		t.Fatal(p, e)
	}
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
	in = dashboardMessagingInput()
	in.Fields = []Field{{Kind: "resource", Key: "messaging.system"}, {Kind: "attribute", Key: "messaging.destination.name"}}
	p, e = PlanServiceDashboard(in, nil)
	if e != nil || p.Profile != "all-spans" {
		t.Fatal(p, e)
	}
	in.Profile = "messaging-producer"
	if _, e = PlanServiceDashboard(in, nil); e == nil {
		t.Fatal("unsupported candidate")
	}
}
func TestDashboardMessagingNativeProcessOnlySeconds(t *testing.T) {
	in := dashboardMessagingInput()
	in.Fields = nil
	in.MessagingConsumerObserved = false
	in.Signals = append(in.Signals, "metrics")
	process := dashboardTestInstrument("z-process", "messaging.process.duration", "histogram", "s")
	in.Instruments = []measurements.Instrument{dashboardTestInstrument("a-load", "load", "gauge", "1"), process, dashboardTestInstrument("b-other", "other", "gauge", "1")}
	p, e := PlanServiceDashboard(in, nil)
	if e != nil || p.Profile != "messaging-consumer" || p.NativeOptions[0].Instrument.Metric != "messaging.process.duration" {
		t.Fatal(p, e)
	}
	count := 0
	for _, o := range p.NativeOptions {
		if o.Selected {
			count++
		}
	}
	if count != 2 {
		t.Fatal("native read budget")
	}
	in.Profile = "messaging-producer"
	if _, e = PlanServiceDashboard(in, nil); e == nil {
		t.Fatal("process metric admitted producer")
	}
	for _, bad := range []measurements.Instrument{dashboardTestInstrument("publish", "messaging.publish.duration", "histogram", "s"), dashboardTestInstrument("unit", "messaging.process.duration", "histogram", "ms")} {
		if dashboardNativeMessaging(in.Scope, in.Service, bad) {
			t.Fatal("unapproved native binding")
		}
	}
	d := *process.Definition
	d.Source.Metric = "other"
	process.Definition = &d
	if dashboardNativeMessaging(in.Scope, in.Service, process) {
		t.Fatal("identity mismatch")
	}
}

func TestDashboardMessagingNativeUnrelatedProfilesRemainGeneric(t *testing.T) {
	for _, profile := range []string{"all-spans", "http-server", "rpc-server", "db-client", "messaging-producer"} {
		in := dashboardMessagingInput()
		in.Profile = profile
		in.Fields = append(in.Fields, Field{Kind: "attribute", Key: "http.request.method"}, Field{Kind: "attribute", Key: "rpc.system"}, Field{Kind: "attribute", Key: "rpc.method"}, Field{Kind: "attribute", Key: "db.system.name"})
		in.Signals = append(in.Signals, "metrics")
		in.Instruments = []measurements.Instrument{dashboardTestInstrument("process", "messaging.process.duration", "histogram", "s")}
		p, err := PlanServiceDashboard(in, nil)
		if err != nil {
			t.Fatal(err)
		}
		for _, widget := range p.Widgets {
			if widget.Measurement != nil && (widget.Family != "native" || widget.Reason != "") {
				t.Fatalf("profile %s retagged existing native binding: %+v", profile, widget)
			}
		}
	}
}
