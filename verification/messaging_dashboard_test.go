package verification

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"hegel.dev/go/hegel"
)

func TestHegelMessagingDashboardRolePopulations(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		start := hegel.Draw(tc, hegel.Integers(1, 1000000))
		service := fmt.Sprintf("worker'\\n-%d", seed)
		namespace := fmt.Sprintf("commerce-π-%d", seed)
		for _, role := range []struct{ profile, kind string }{{"messaging-consumer", "Consumer"}, {"messaging-producer", "Producer"}} {
			for _, ns := range []*identity.Namespace{nil, {Mode: "missing"}, {Mode: "value", Value: ""}, {Mode: "value", Value: namespace}} {
				in := exploration.DashboardRequest{Profile: role.profile, Scope: exploration.InventoryQuestion{Environment: "test", FromMS: int64(start), ToMS: int64(start + 60000), Namespace: ns}, Service: service, Signals: []string{"spans", "logs"}, Fields: []exploration.Field{{Kind: "attribute", Key: "messaging.system"}, {Kind: "attribute", Key: "messaging.destination.name"}, {Kind: "attribute", Key: "messaging.operation.name"}}}
				before, _ := json.Marshal(in)
				p, e := exploration.PlanServiceDashboard(in, nil)
				after, _ := json.Marshal(in)
				if e != nil || p.Profile != role.profile || string(before) != string(after) {
					tc.Errorf("messaging intent/input %v", e)
					return
				}
				for _, w := range p.Widgets {
					if w.Query == nil {
						continue
					}
					q := w.Query
					if q.Environment != in.Scope.Environment || q.FromMS != in.Scope.FromMS || q.ToMS != in.Scope.ToMS || q.Offset != 0 || q.Limit != 100 {
						tc.Errorf("messaging scope drift")
						return
					}
					for _, s := range []string{service, "other"} {
						for _, rn := range []*string{nil, lpcPointer(""), lpcPointer(namespace), lpcPointer("other")} {
							for _, kind := range []string{"Consumer", "Producer", "Server", "Client", "consumer", ""} {
								for _, system := range []*string{nil, lpcPointer(""), lpcPointer("kafka"), lpcPointer(" ")} {
									fields := map[exploration.Field]string{{Kind: "column", Key: "service"}: s, {Kind: "column", Key: "kind"}: kind, {Kind: "attribute", Key: "messaging.destination.name"}: "orders", {Kind: "attribute", Key: "messaging.operation.name"}: "process"}
									if rn != nil {
										fields[exploration.Field{Kind: "resource", Key: "service.namespace"}] = *rn
									}
									if system != nil {
										fields[exploration.Field{Kind: "attribute", Key: "messaging.system"}] = *system
									}
									want := s == service
									if ns != nil {
										if ns.Mode == "missing" {
											want = want && rn == nil
										} else {
											want = want && rn != nil && *rn == ns.Value
										}
									}
									if q.Signal == "spans" {
										want = want && kind == role.kind && system != nil && *system != ""
									} else if q.Signal != "logs" {
										tc.Errorf("unexpected signal")
										return
									}
									if hdPredicate(q.Where, fields) != want {
										tc.Errorf("role population %s kind=%q", role.profile, kind)
										return
									}
								}
							}
						}
					}
				}
				in.Fields[0], in.Fields[2] = in.Fields[2], in.Fields[0]
				in.Signals[0], in.Signals[1] = in.Signals[1], in.Signals[0]
				again, e := exploration.PlanServiceDashboard(in, nil)
				if e != nil || !reflect.DeepEqual(p, again) {
					tc.Errorf("messaging discovery order changed plan")
					return
				}
			}
		}
	}, options(12)...)
}
func TestMessagingDashboardCandidatePrecedenceAndDimensions(t *testing.T) {
	base := exploration.DashboardRequest{Profile: "auto", Scope: exploration.InventoryQuestion{Environment: "test", FromMS: 1000, ToMS: 61000}, Service: "worker", Signals: []string{"spans"}, Fields: []exploration.Field{{Kind: "attribute", Key: "messaging.system"}}}
	for _, c := range []struct {
		name               string
		consumer, producer bool
		want               string
	}{{"flat key unknown", false, false, "all-spans"}, {"consumer", true, false, "messaging-consumer"}, {"producer", false, true, "messaging-producer"}, {"both roles", true, true, "messaging-consumer"}} {
		t.Run(c.name, func(t *testing.T) {
			in := base
			in.MessagingConsumerObserved = c.consumer
			in.MessagingProducerObserved = c.producer
			p, e := exploration.PlanServiceDashboard(in, nil)
			if e != nil || p.Profile != c.want {
				t.Fatal(p, e)
			}
		})
	}
	in := base
	in.MessagingConsumerObserved = true
	in.MessagingProducerObserved = true
	for _, key := range []string{"messaging.system", "messaging.operation.name", "messaging.destination.name"} {
		if key != "messaging.system" {
			in.Fields = append(in.Fields, exploration.Field{Kind: "attribute", Key: key})
		}
		p, e := exploration.PlanServiceDashboard(in, nil)
		if e != nil {
			t.Fatal(e)
		}
		found := false
		for _, w := range p.Widgets {
			if w.Query != nil && len(w.Query.GroupBy) == 1 && w.Query.GroupBy[0] == (exploration.Field{Kind: "attribute", Key: key}) {
				found = true
			}
		}
		if !found {
			t.Fatalf("missing literal dimension priority %s", key)
		}
	}
	in.Fields = append(in.Fields, exploration.Field{Kind: "attribute", Key: "db.system.name"})
	in.DBClientObserved = true
	p, e := exploration.PlanServiceDashboard(in, nil)
	if e != nil || p.Profile != "db-client" {
		t.Fatal("DB precedence", p, e)
	}
	in.Fields = append(in.Fields, exploration.Field{Kind: "attribute", Key: "rpc.system"}, exploration.Field{Kind: "attribute", Key: "rpc.method"})
	in.RPCServerObserved = true
	p, e = exploration.PlanServiceDashboard(in, nil)
	if e != nil || p.Profile != "rpc-server" {
		t.Fatal("RPC precedence", p, e)
	}
	in.Fields = append(in.Fields, exploration.Field{Kind: "attribute", Key: "http.request.method"})
	in.HTTPServerObserved = true
	p, e = exploration.PlanServiceDashboard(in, nil)
	if e != nil || p.Profile != "http-server" {
		t.Fatal("HTTP precedence", p, e)
	}
	in.Profile = "messaging-producer"
	p, e = exploration.PlanServiceDashboard(in, nil)
	if e != nil || p.Profile != "messaging-producer" {
		t.Fatal("explicit producer lost", p, e)
	}
	for _, f := range []exploration.Field{{Kind: "resource", Key: "messaging.system"}, {Kind: "attribute", Key: "messaging.destination.name"}} {
		bad := base
		bad.Fields = []exploration.Field{f}
		bad.MessagingConsumerObserved = true
		p, e := exploration.PlanServiceDashboard(bad, nil)
		if e != nil || p.Profile != "all-spans" {
			t.Fatal("hint/resource key promoted system", p, e)
		}
		bad.Profile = "messaging-consumer"
		if _, e := exploration.PlanServiceDashboard(bad, nil); e == nil {
			t.Fatal("explicit unsupported messaging accepted")
		}
	}
}
func TestMessagingDashboardNativeConsumerUnitIdentity(t *testing.T) {
	scope := exploration.InventoryQuestion{Environment: "test", FromMS: 1000, ToMS: 61000, Namespace: &identity.Namespace{Mode: "value", Value: "commerce"}}
	for _, unit := range []string{"s", "ms"} {
		in := exploration.DashboardRequest{Profile: "auto", Scope: scope, Service: "worker", Signals: []string{"metrics"}}
		consumer := hdInstrument("zzz-process", "messaging.process.duration", "histogram", unit, scope, "worker")
		generic := hdInstrument("aaa-gauge", "custom.capacity", "gauge", "1", scope, "worker")
		in.Instruments = append(in.Instruments, generic, consumer)
		p, e := exploration.PlanServiceDashboard(in, nil)
		if e != nil {
			t.Fatal(e)
		}
		want := "all-spans"
		if unit == "s" {
			want = "messaging-consumer"
		}
		if p.Profile != want {
			t.Fatalf("unit %q promoted %q", unit, p.Profile)
		}
		if unit == "s" && p.NativeOptions[0].Instrument.Metric != "messaging.process.duration" {
			t.Fatal("native consumer did not outrank generic")
		}
		for _, w := range p.Widgets {
			if w.Query != nil && w.Query.Signal == "spans" {
				t.Fatal("invented span population")
			}
			if w.Measurement != nil && (w.Grain != "per-stream" || !w.Measurement.Definition.Source.Namespace.Equal(scope.Namespace) || w.Measurement.FromMS != scope.FromMS || w.Measurement.ToMS != scope.ToMS) {
				t.Fatal("native scope/grain")
			}
			if w.Measurement != nil && w.Measurement.Definition.Source.Metric == "messaging.process.duration" && w.Measurement.Definition.Source.Unit != unit {
				t.Fatal("native unit converted")
			}
		}
		in.Instruments[0], in.Instruments[1] = in.Instruments[1], in.Instruments[0]
		for i := range in.Instruments {
			in.Instruments[i].Records += 999
		}
		again, e := exploration.PlanServiceDashboard(in, nil)
		if e != nil {
			t.Fatal(e)
		}
		for i := range p.NativeOptions {
			if p.NativeOptions[i].Key != again.NativeOptions[i].Key || p.NativeOptions[i].Selected != again.NativeOptions[i].Selected {
				t.Fatal("native default order changed")
			}
		}
	}
}
func TestMessagingDashboardObservedPopulations(t *testing.T) {
	path := os.Getenv("GOO_MESSAGING_DASHBOARD_LEDGER")
	if path == "" {
		t.Skip("set GOO_MESSAGING_DASHBOARD_LEDGER to actual190 ledger")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var facts struct{ Records []hdLiveRecord }
	if e = json.Unmarshal(raw, &facts); e != nil {
		t.Fatal(e)
	}
	if len(facts.Records) <= 100 {
		t.Fatal("fixture must exceed page")
	}
	hdCheckObservedLedger(t, path, len(facts.Records))
}
