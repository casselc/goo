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

func TestHegelDBDashboardPopulation(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		start := hegel.Draw(tc, hegel.Integers(1, 1000000))
		service := fmt.Sprintf("db'\\n-%d", seed)
		namespace := fmt.Sprintf("payments-π-%d", seed)
		for _, ns := range []*identity.Namespace{nil, {Mode: "missing"}, {Mode: "value", Value: ""}, {Mode: "value", Value: namespace}} {
			in := exploration.DashboardRequest{Profile: "db-client", Scope: exploration.InventoryQuestion{Environment: "test", FromMS: int64(start), ToMS: int64(start + 60000), Namespace: ns}, Service: service, Signals: []string{"spans", "logs"}, Fields: []exploration.Field{{Kind: "attribute", Key: "db.system.name"}, {Kind: "attribute", Key: "db.operation.name"}, {Kind: "attribute", Key: "db.namespace"}}}
			before, _ := json.Marshal(in)
			p, e := exploration.PlanServiceDashboard(in, nil)
			after, _ := json.Marshal(in)
			if e != nil || p.Profile != "db-client" || string(before) != string(after) {
				tc.Errorf("DB intent/input: %v", e)
				return
			}
			for _, w := range p.Widgets {
				if w.Query == nil {
					continue
				}
				q := w.Query
				if q.Environment != in.Scope.Environment || q.FromMS != in.Scope.FromMS || q.ToMS != in.Scope.ToMS || q.Offset != 0 || q.Limit != 100 {
					tc.Errorf("DB scope drift")
					return
				}
				for _, actualService := range []string{service, "other"} {
					for _, resourceNS := range []*string{nil, lpcPointer(""), lpcPointer(namespace), lpcPointer("other")} {
						for _, kind := range []string{"Client", "Server", "Internal", "client", ""} {
							for _, system := range []*string{nil, lpcPointer(""), lpcPointer("postgresql"), lpcPointer(" ")} {
								fields := map[exploration.Field]string{{Kind: "column", Key: "service"}: actualService, {Kind: "column", Key: "kind"}: kind, {Kind: "attribute", Key: "db.system"}: "postgresql", {Kind: "attribute", Key: "db.operation.name"}: "SELECT", {Kind: "attribute", Key: "db.namespace"}: "orders"}
								if resourceNS != nil {
									fields[exploration.Field{Kind: "resource", Key: "service.namespace"}] = *resourceNS
								}
								if system != nil {
									fields[exploration.Field{Kind: "attribute", Key: "db.system.name"}] = *system
								}
								want := actualService == service
								if ns != nil {
									if ns.Mode == "missing" {
										want = want && resourceNS == nil
									} else {
										want = want && resourceNS != nil && *resourceNS == ns.Value
									}
								}
								if q.Signal == "spans" {
									want = want && kind == "Client" && system != nil && *system != ""
								} else if q.Signal != "logs" {
									tc.Errorf("unexpected signal")
									return
								}
								if hdPredicate(q.Where, fields) != want {
									tc.Errorf("DB population mismatch kind=%q", kind)
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
				tc.Errorf("DB metadata order changed plan")
				return
			}
		}
	}, options(16)...)
}
func TestDBDashboardCandidatesDimensionsAndPrecedence(t *testing.T) {
	base := exploration.DashboardRequest{Profile: "auto", Scope: exploration.InventoryQuestion{Environment: "test", FromMS: 1000, ToMS: 61000}, Service: "api", Signals: []string{"spans"}}
	for _, c := range []struct {
		name  string
		field exploration.Field
		hint  bool
		want  string
	}{
		{"current key unknown", exploration.Field{Kind: "attribute", Key: "db.system.name"}, false, "all-spans"},
		{"positive current", exploration.Field{Kind: "attribute", Key: "db.system.name"}, true, "db-client"},
		{"legacy alone", exploration.Field{Kind: "attribute", Key: "db.system"}, true, "all-spans"},
		{"resource spoof", exploration.Field{Kind: "resource", Key: "db.system.name"}, true, "all-spans"},
	} {
		t.Run(c.name, func(t *testing.T) {
			in := base
			in.Fields = []exploration.Field{c.field}
			in.DBClientObserved = c.hint
			p, e := exploration.PlanServiceDashboard(in, nil)
			if e != nil || p.Profile != c.want {
				t.Fatal(p, e)
			}
		})
	}
	in := base
	in.DBClientObserved = true
	in.Fields = []exploration.Field{{Kind: "attribute", Key: "db.system.name"}}
	for _, key := range []string{"db.system.name", "db.namespace", "db.operation.name"} {
		if key != "db.system.name" {
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
	in.Fields = append(in.Fields, exploration.Field{Kind: "attribute", Key: "rpc.system"}, exploration.Field{Kind: "attribute", Key: "rpc.method"})
	in.RPCServerObserved = true
	p, e := exploration.PlanServiceDashboard(in, nil)
	if e != nil || p.Profile != "rpc-server" {
		t.Fatal("RPC precedence", p, e)
	}
	in.Fields = append(in.Fields, exploration.Field{Kind: "attribute", Key: "http.request.method"})
	in.HTTPServerObserved = true
	p, e = exploration.PlanServiceDashboard(in, nil)
	if e != nil || p.Profile != "http-server" {
		t.Fatal("HTTP precedence", p, e)
	}
	in.Profile = "db-client"
	p, e = exploration.PlanServiceDashboard(in, nil)
	if e != nil || p.Profile != "db-client" {
		t.Fatal("explicit DB lost", p, e)
	}
	base.Profile = "db-client"
	if _, e := exploration.PlanServiceDashboard(base, nil); e == nil {
		t.Fatal("unsupported DB accepted")
	}
}
func TestDBDashboardNativeExactUnitAndIdentity(t *testing.T) {
	scope := exploration.InventoryQuestion{Environment: "test", FromMS: 1000, ToMS: 61000, Namespace: &identity.Namespace{Mode: "value", Value: "payments"}}
	for _, unit := range []string{"s", "ms"} {
		in := exploration.DashboardRequest{Profile: "auto", Scope: scope, Service: "api", Signals: []string{"metrics"}}
		db := hdInstrument("zzz-db", "db.client.operation.duration", "histogram", unit, scope, "api")
		gauge := hdInstrument("aaa-gauge", "custom.capacity", "gauge", "1", scope, "api")
		in.Instruments = append(in.Instruments, gauge, db)
		p, e := exploration.PlanServiceDashboard(in, nil)
		if e != nil {
			t.Fatal(e)
		}
		want := "all-spans"
		if unit == "s" {
			want = "db-client"
		}
		if p.Profile != want {
			t.Fatalf("unit %q selected %q", unit, p.Profile)
		}
		if unit == "s" && p.NativeOptions[0].Instrument.Metric != "db.client.operation.duration" {
			t.Fatal("eligible DB did not outrank generic")
		}
		for _, w := range p.Widgets {
			if w.Query != nil && w.Query.Signal == "spans" {
				t.Fatal("invented spans without discovery")
			}
			if w.Measurement != nil && (w.Grain != "per-stream" || !w.Measurement.Definition.Source.Namespace.Equal(scope.Namespace) || w.Measurement.FromMS != scope.FromMS || w.Measurement.ToMS != scope.ToMS) {
				t.Fatal("native scope/grain")
			}
			if w.Measurement != nil && w.Measurement.Definition.Source.Metric == "db.client.operation.duration" && w.Measurement.Definition.Source.Unit != unit {
				t.Fatal("native unit conversion")
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
				t.Fatal("DB order changed with counts/permutation")
			}
		}
	}
}
func TestDBDashboardObservedPopulations(t *testing.T) {
	path := os.Getenv("GOO_DB_DASHBOARD_LEDGER")
	if path == "" {
		t.Skip("set GOO_DB_DASHBOARD_LEDGER to actual189 ledger")
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
		t.Fatal("fixture must exceed inspection page")
	}
	hdCheckObservedLedger(t, path, len(facts.Records))
}
