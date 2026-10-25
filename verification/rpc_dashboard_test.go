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

func TestHegelRPCDashboardRecordPopulation(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		start := hegel.Draw(tc, hegel.Integers(1, 1000000))
		service := fmt.Sprintf("rpc'\\n-%d", seed)
		for _, ns := range []*identity.Namespace{nil, {Mode: "missing"}, {Mode: "value", Value: ""}, {Mode: "value", Value: fmt.Sprintf("payments-π-%d", seed)}} {
			in := exploration.DashboardRequest{Profile: "rpc-server", Scope: exploration.InventoryQuestion{Environment: "test", FromMS: int64(start), ToMS: int64(start + 60000), Namespace: ns}, Service: service, Signals: []string{"spans", "logs"}, Fields: []exploration.Field{{Kind: "attribute", Key: "rpc.system"}, {Kind: "attribute", Key: "rpc.method"}}}
			before, _ := json.Marshal(in)
			p, e := exploration.PlanServiceDashboard(in, nil)
			after, _ := json.Marshal(in)
			if e != nil || p.Profile != "rpc-server" || string(before) != string(after) {
				tc.Errorf("RPC intent/input: %v", e)
				return
			}
			for _, w := range p.Widgets {
				if w.Query == nil {
					continue
				}
				q := w.Query
				if q.Environment != in.Scope.Environment || q.FromMS != in.Scope.FromMS || q.ToMS != in.Scope.ToMS || q.Limit != 100 || q.Offset != 0 {
					tc.Errorf("RPC scope drift")
					return
				}
				if q.Signal != "spans" && q.Signal != "logs" {
					tc.Errorf("unexpected record question")
					return
				}
				for _, actualService := range []string{service, "other"} {
					for _, reportedNS := range []*string{nil, lpcPointer(""), lpcPointer(fmt.Sprintf("payments-π-%d", seed)), lpcPointer("other")} {
						for _, kind := range []string{"Server", "Client", "Internal", "server", ""} {
							for _, system := range []*string{nil, lpcPointer(""), lpcPointer("grpc"), lpcPointer(" ")} {
								for _, method := range []*string{nil, lpcPointer(""), lpcPointer("GetOrder"), lpcPointer(" ")} {
									fields := map[exploration.Field]string{{Kind: "column", Key: "service"}: actualService, {Kind: "column", Key: "kind"}: kind, {Kind: "attribute", Key: "rpc.system.name"}: "grpc"}
									if reportedNS != nil {
										fields[exploration.Field{Kind: "resource", Key: "service.namespace"}] = *reportedNS
									}
									if system != nil {
										fields[exploration.Field{Kind: "attribute", Key: "rpc.system"}] = *system
									}
									if method != nil {
										fields[exploration.Field{Kind: "attribute", Key: "rpc.method"}] = *method
									}
									want := actualService == service
									if ns != nil {
										if ns.Mode == "missing" {
											want = want && reportedNS == nil
										} else {
											want = want && reportedNS != nil && *reportedNS == ns.Value
										}
									}
									if q.Signal == "spans" {
										want = want && kind == "Server" && system != nil && *system != "" && method != nil && *method != ""
									}
									if hdPredicate(q.Where, fields) != want {
										tc.Errorf("RPC independent population mismatch kind=%q system=%v method=%v", kind, system, method)
										return
									}
								}
							}
						}
					}
				}
			}
			in.Fields[0], in.Fields[1] = in.Fields[1], in.Fields[0]
			in.Signals[0], in.Signals[1] = in.Signals[1], in.Signals[0]
			again, e := exploration.PlanServiceDashboard(in, nil)
			if e != nil || !reflect.DeepEqual(p, again) {
				tc.Errorf("RPC metadata order changed plan")
				return
			}
		}
	}, options(16)...)
}
func TestRPCDashboardCandidateAndPriority(t *testing.T) {
	base := exploration.DashboardRequest{Profile: "auto", Scope: exploration.InventoryQuestion{Environment: "test", FromMS: 1000, ToMS: 61000}, Service: "rpc-api", Signals: []string{"spans"}}
	for _, c := range []struct {
		name   string
		fields []exploration.Field
		hint   bool
		want   string
	}{
		{"no metadata", nil, false, "all-spans"},
		{"flat pair unknown", []exploration.Field{{Kind: "attribute", Key: "rpc.system"}, {Kind: "attribute", Key: "rpc.method"}}, false, "all-spans"},
		{"observed pair", []exploration.Field{{Kind: "attribute", Key: "rpc.system"}, {Kind: "attribute", Key: "rpc.method"}}, true, "rpc-server"},
		{"wrong current alias", []exploration.Field{{Kind: "attribute", Key: "rpc.system.name"}, {Kind: "attribute", Key: "rpc.method"}}, true, "all-spans"},
		{"missing method", []exploration.Field{{Kind: "attribute", Key: "rpc.system"}}, true, "all-spans"},
		{"resource spoof", []exploration.Field{{Kind: "resource", Key: "rpc.system"}, {Kind: "attribute", Key: "rpc.method"}}, true, "all-spans"},
	} {
		t.Run(c.name, func(t *testing.T) {
			in := base
			in.Fields = c.fields
			in.RPCServerObserved = c.hint
			p, e := exploration.PlanServiceDashboard(in, nil)
			if e != nil || p.Profile != c.want {
				t.Fatal(p, e)
			}
		})
	}
	both := base
	both.Fields = []exploration.Field{{Kind: "attribute", Key: "http.request.method"}, {Kind: "attribute", Key: "rpc.system"}, {Kind: "attribute", Key: "rpc.method"}}
	both.HTTPServerObserved = true
	both.RPCServerObserved = true
	p, e := exploration.PlanServiceDashboard(both, nil)
	if e != nil || p.Profile != "http-server" {
		t.Fatal("HTTP must retain auto priority", p, e)
	}
	both.Profile = "rpc-server"
	p, e = exploration.PlanServiceDashboard(both, nil)
	if e != nil || p.Profile != "rpc-server" {
		t.Fatal("explicit RPC lost", p, e)
	}
	base.Profile = "rpc-server"
	if _, e := exploration.PlanServiceDashboard(base, nil); e == nil {
		t.Fatal("explicit absent RPC accepted")
	}
}
func TestRPCDashboardNativeUnitsAndPriority(t *testing.T) {
	scope := exploration.InventoryQuestion{Environment: "test", FromMS: 1000, ToMS: 61000, Namespace: &identity.Namespace{Mode: "value", Value: "payments"}}
	for _, unit := range []string{"ms", "s"} {
		in := exploration.DashboardRequest{Profile: "auto", Scope: scope, Service: "api", Signals: []string{"metrics"}}
		rpc := hdInstrument("zzz-rpc", "rpc.server.duration", "histogram", unit, scope, "api")
		generic := hdInstrument("aaa-gauge", "custom.capacity", "gauge", "1", scope, "api")
		in.Instruments = append(in.Instruments, generic, rpc)
		p, e := exploration.PlanServiceDashboard(in, nil)
		if e != nil {
			t.Fatal(e)
		}
		want := "all-spans"
		if unit == "ms" {
			want = "rpc-server"
		}
		if p.Profile != want {
			t.Fatalf("unit %q selected %q", unit, p.Profile)
		}
		if unit == "ms" && len(p.NativeOptions) > 0 && p.NativeOptions[0].Instrument.Metric != "rpc.server.duration" {
			t.Fatal("RPC did not outrank generic")
		}
		for _, w := range p.Widgets {
			if w.Query != nil && w.Query.Signal == "spans" {
				t.Fatal("invented spans without inventory")
			}
			if w.Measurement != nil && (w.Grain != "per-stream" || !w.Measurement.Definition.Source.Namespace.Equal(scope.Namespace) || w.Measurement.FromMS != scope.FromMS || w.Measurement.ToMS != scope.ToMS) {
				t.Fatal("native grain/scope changed")
			}
			if w.Measurement != nil && w.Measurement.Definition.Source.Metric == "rpc.server.duration" && w.Measurement.Definition.Source.Unit != unit {
				t.Fatal("RPC unit converted")
			}
		}
		in.Instruments[0], in.Instruments[1] = in.Instruments[1], in.Instruments[0]
		for i := range in.Instruments {
			in.Instruments[i].Records += 999
		}
		again, e := exploration.PlanServiceDashboard(in, nil)
		if e != nil || len(p.NativeOptions) != len(again.NativeOptions) {
			t.Fatal(e)
		}
		for i := range p.NativeOptions {
			if p.NativeOptions[i].Key != again.NativeOptions[i].Key || p.NativeOptions[i].Selected != again.NativeOptions[i].Selected {
				t.Fatal("RPC native metadata/count ordering changed")
			}
		}
	}
}

// Reuses the independent raw-fact oracle, not the production planner/evaluator.
func TestRPCDashboardObservedPopulations(t *testing.T) {
	path := os.Getenv("GOO_RPC_DASHBOARD_LEDGER")
	if path == "" {
		t.Skip("set GOO_RPC_DASHBOARD_LEDGER to an actual 188 ledger")
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
		t.Fatal("full-population fixture must exceed inspection page")
	}
	hdCheckObservedLedger(t, path, len(facts.Records))
}
