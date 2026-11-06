package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
	"slices"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"hegel.dev/go/hegel"
)

type rpcRaw357 struct {
	At         int64             `json:"timestamp_ms"`
	ID         string            `json:"SpanId"`
	Service    string            `json:"ServiceName"`
	Kind       string            `json:"SpanKind"`
	Resource   map[string]string `json:"ResourceAttributes"`
	Attributes map[string]string `json:"SpanAttributes"`
}
type rpcFixture357 struct {
	Mode    int
	Scope   exploration.InventoryQuestion
	Service string
	Rows    []rpcRaw357
}

func rpcFixtureFor357(seed, mode, namespace int) rpcFixture357 {
	f := rpcFixture357{Mode: mode, Service: fmt.Sprintf("rpc'π-%d", seed), Scope: exploration.InventoryQuestion{Environment: "prod", FromMS: 1000 + int64(seed), ToMS: 61000 + int64(seed)}}
	switch namespace {
	case 1:
		f.Scope.Namespace = &identity.Namespace{Mode: "missing"}
	case 2:
		f.Scope.Namespace = &identity.Namespace{Mode: "value", Value: ""}
	case 3:
		f.Scope.Namespace = &identity.Namespace{Mode: "value", Value: "team"}
	}
	if mode > 0 {
		f.Scope.Environment = ""
		s := identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}}
		if mode == 1 {
			s.Policy = "current-only"
		}
		if mode == 3 {
			s.Scope = identity.EnvironmentScope{Mode: "empty"}
		}
		if mode == 4 {
			s.Scope = identity.EnvironmentScope{Mode: "unreported"}
		}
		f.Scope.EnvironmentSelector = &s
	}
	envs := []map[string]string{{"deployment.environment.name": "prod"}, {"deployment.environment": "prod"}, {"deployment.environment.name": ""}, {}, {"deployment.environment.name": "foreign", "deployment.environment": "prod"}, {"deployment.environment.name": "", "deployment.environment": "prod"}}
	add := func(env map[string]string, attrs map[string]string, kind string) {
		resource := map[string]string{}
		for k, v := range env {
			resource[k] = v
		}
		if namespace != 1 {
			resource["service.namespace"] = "team"
			if namespace == 2 {
				resource["service.namespace"] = ""
			}
		}
		f.Rows = append(f.Rows, rpcRaw357{At: f.Scope.FromMS + 100, ID: fmt.Sprintf("%016x", len(f.Rows)+1), Service: f.Service, Kind: kind, Resource: resource, Attributes: attrs})
	}
	for _, env := range envs {
		for _, method := range []*string{nil, lpcPointer(""), lpcPointer("_OTHER"), lpcPointer("package.Service/Call")} {
			a := map[string]string{"rpc.system.name": "grpc"}
			if method != nil {
				a["rpc.method"] = *method
			}
			add(env, a, "Server")
		}
		add(env, map[string]string{"rpc.system": "grpc", "rpc.method": "Legacy/Call"}, "Server")
		add(env, map[string]string{"rpc.system.name": "grpc", "rpc.system": "legacy", "rpc.method": "Dual/Call"}, "Server")
		add(env, map[string]string{"rpc.system.name": "", "rpc.system": "legacy", "rpc.method": "Legacy/Only"}, "Server")
		add(env, map[string]string{"rpc.system.name": "grpc", "rpc.method": "Client/Call"}, "Client")
		add(env, map[string]string{"rpc.system.name": "grpc"}, "server")
		add(env, map[string]string{"rpc.system.name": " "}, "Server")
	}
	base := f.Rows[0]
	for i := range 4 {
		raw, _ := json.Marshal(base)
		var r rpcRaw357
		json.Unmarshal(raw, &r)
		r.ID = fmt.Sprintf("%016x", len(f.Rows)+1)
		switch i {
		case 0:
			r.Service = "foreign"
		case 1:
			r.At = f.Scope.ToMS
		case 2:
			r.At = f.Scope.FromMS - 1
		case 3:
			r.Resource["service.namespace"] = "foreign"
		}
		f.Rows = append(f.Rows, r)
	}
	return f
}

// These predicates are literal from pinned OTel v1.45.0, not derived from the
// product descriptor. Current method is conditional: missing/empty stay members.
func rpcMember357(f rpcFixture357, r rpcRaw357, current bool) bool {
	if r.Service != f.Service || r.At < f.Scope.FromMS || r.At >= f.Scope.ToMS || r.Kind != "Server" || !continuationEnvironment352(f.Mode, "prod", r.Resource) || !continuationNamespace352(f.Scope.Namespace, r.Resource) {
		return false
	}
	if current {
		return r.Attributes["rpc.system.name"] != ""
	}
	return r.Attributes["rpc.system"] != "" && r.Attributes["rpc.method"] != ""
}
func rpcQuestion357(f rpcFixture357) exploration.Query {
	q := exploration.Query{Signal: "spans", Environment: f.Scope.Environment, EnvironmentSelector: f.Scope.EnvironmentSelector, FromMS: f.Scope.FromMS, ToMS: f.Scope.ToMS, Limit: 100, Where: &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: f.Service}}
	return q
}
func rpcHints357(f rpcFixture357) (bool, bool, error) {
	in, _ := json.Marshal(map[string]any{"scope": f.Scope, "service": f.Service, "question": rpcQuestion357(f), "rows": f.Rows})
	cmd := exec.Command("node", "current_rpc_357_fixture.mjs")
	cmd.Stdin = bytes.NewReader(in)
	out, err := cmd.Output()
	if err != nil {
		if fail, ok := err.(*exec.ExitError); ok {
			return false, false, fmt.Errorf("actual RPC witness: %s", fail.Stderr)
		}
		return false, false, err
	}
	var hints struct{ Current, Legacy, Complete bool }
	if json.Unmarshal(out, &hints) != nil || !hints.Complete {
		return false, false, fmt.Errorf("incomplete actual witness")
	}
	return hints.Current, hints.Legacy, nil
}
func rpcPlan357(f rpcFixture357, profile string, current, legacy bool, instruments []measurements.Instrument) (*exploration.DashboardPlan, error) {
	return exploration.PlanServiceDashboard(exploration.DashboardRequest{Scope: f.Scope, Service: f.Service, Profile: profile, Signals: []string{"spans", "logs", "metrics"}, Fields: []exploration.Field{{Kind: "attribute", Key: "rpc.system.name"}, {Kind: "attribute", Key: "rpc.system"}, {Kind: "attribute", Key: "rpc.method"}}, RPCServerCurrentObserved: current, RPCServerObserved: legacy, Instruments: instruments}, nil)
}
func TestHegelCurrentRPC357(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		f := rpcFixtureFor357(hegel.Draw(tc, hegel.Integers(1, 10000)), hegel.Draw(tc, hegel.Integers(0, 4)), hegel.Draw(tc, hegel.Integers(0, 3)))
		current, legacy, err := rpcHints357(f)
		if err != nil {
			tc.Fatal(err)
		}
		wantCurrent, wantLegacy := false, false
		for _, row := range f.Rows[:min(100, len(f.Rows))] {
			wantCurrent = wantCurrent || rpcMember357(f, row, true)
			wantLegacy = wantLegacy || rpcMember357(f, row, false)
		}
		if current != wantCurrent || legacy != wantLegacy {
			tc.Fatal("independent bounded raw witness mismatch")
		}
		for _, profile := range []string{"rpc-server-current", "rpc-server", "all-spans", "auto"} {
			plan, err := rpcPlan357(f, profile, current, legacy, nil)
			if err != nil {
				tc.Fatal(err)
			}
			want := profile
			if want == "auto" {
				want = "all-spans"
				if legacy {
					want = "rpc-server"
				}
				if current {
					want = "rpc-server-current"
				}
			}
			if plan.Profile != want {
				tc.Fatal("manual selection or current/legacy precedence", plan.Profile, want)
			}
			for _, w := range plan.Widgets {
				if w.Query == nil {
					continue
				}
				q := w.Query
				if q.FromMS != f.Scope.FromMS || q.ToMS != f.Scope.ToMS || q.Environment != f.Scope.Environment || continuationJSON352(q.EnvironmentSelector) != continuationJSON352(f.Scope.EnvironmentSelector) || q.Offset != 0 {
					tc.Fatal("scope/profile drift")
				}
				if q.Signal != "spans" {
					continue
				}
				for _, row := range f.Rows {
					values := map[exploration.Field]string{{Kind: "column", Key: "service"}: row.Service, {Kind: "column", Key: "kind"}: row.Kind}
					for k, v := range row.Attributes {
						values[exploration.Field{Kind: "attribute", Key: k}] = v
					}
					for k, v := range row.Resource {
						values[exploration.Field{Kind: "resource", Key: k}] = v
					}
					actual := hdPredicate(q.Where, values) && row.At >= q.FromMS && row.At < q.ToMS && continuationEnvironment352(f.Mode, "prod", row.Resource)
					expected := rpcMember357(f, row, want == "rpc-server-current")
					if want == "all-spans" {
						expected = row.Service == f.Service && row.At >= q.FromMS && row.At < q.ToMS && continuationEnvironment352(f.Mode, "prod", row.Resource) && continuationNamespace352(f.Scope.Namespace, row.Resource)
					}
					if actual != expected {
						tc.Fatal("raw membership scope/role/current-vs-legacy", want, row.ID)
					}
				}
				if w.Key == "dimension" && want != "all-spans" && !reflect.DeepEqual(q.GroupBy, []exploration.Field{{Kind: "attribute", Key: "rpc.method"}}) {
					tc.Fatal("RPC method grouping changed missing/empty semantics")
				}
			}
		}
		// Explicit published precedence, independently enumerated rather than
		// read back from product family descriptors.
		for stage, want := range []string{"http-server", "rpc-server-current", "rpc-server", "db-client", "messaging-consumer", "messaging-producer", "all-spans"} {
			request := exploration.DashboardRequest{Scope: f.Scope, Service: f.Service, Profile: "auto", Signals: []string{"spans"}, Fields: []exploration.Field{{Kind: "attribute", Key: "http.request.method"}, {Kind: "attribute", Key: "rpc.system.name"}, {Kind: "attribute", Key: "rpc.system"}, {Kind: "attribute", Key: "rpc.method"}, {Kind: "attribute", Key: "db.system.name"}, {Kind: "attribute", Key: "messaging.system"}}, HTTPServerObserved: stage <= 0, RPCServerCurrentObserved: stage <= 1, RPCServerObserved: stage <= 2, DBClientObserved: stage <= 3, MessagingConsumerObserved: stage <= 4, MessagingProducerObserved: stage <= 5}
			plan, e := exploration.PlanServiceDashboard(request, nil)
			if e != nil || plan.Profile != want {
				tc.Fatal("independent auto family precedence", stage, want, plan.Profile, e)
			}
		}
		// Metadata alone may offer a candidate but cannot make it the automatic lens.
		unknown, err := rpcPlan357(f, "auto", false, false, nil)
		if err != nil || unknown.Profile != "all-spans" {
			tc.Fatal("metadata mistaken for same-row witness", err)
		}
		// Native priority is independently literal: correct name/unit/type plus existing
		// binder support, never a semconv conformance or cumulative-only assertion.
		scalar := f.Scope
		if scalar.EnvironmentSelector != nil {
			scalar.Environment = "prod"
			scalar.EnvironmentSelector = nil
		}
		for _, tempo := range []int{0, 1, 2, 3} {
			for _, variant := range []int{0, 1, 2, 3, 4, 5} {
				i := hdInstrument("rpc", "rpc.server.call.duration", "histogram", "s", scalar, f.Service)
				i.Temporality = tempo
				switch variant {
				case 1:
					i.Metric = "rpc.server.duration"
					i.Unit = "ms"
					i.Definition.Source.Unit = i.Unit
					i.Definition.Source.Metric = i.Metric
				case 2:
					i.Unit = "ms"
					i.Definition.Source.Unit = i.Unit
				case 3:
					i.Kind = "gauge"
					i.Definition.Kind = "gauge"
				case 4:
					i.Supported = false
				case 5:
					i.Definition.Source.Services = []string{"foreign"}
				}
				nativeFixture := f
				nativeFixture.Scope = scalar
				plan, e := rpcPlan357(nativeFixture, "auto", false, false, []measurements.Instrument{i})
				engineValid := variant == 4 || variant != 5 && ((variant == 3 && tempo == 0) || (variant != 3 && (tempo == 1 || tempo == 2)))
				if (e == nil) != engineValid {
					tc.Fatal("independent existing engine admission", tempo, variant, e)
				}
				if !engineValid {
					continue
				}
				nativeWant := variant == 0 && (tempo == 1 || tempo == 2)
				if (plan.Profile == "rpc-server-current") != nativeWant {
					tc.Fatal("literal current native eligibility", tempo, variant, plan.Profile)
				}
				if variant == 1 && plan.Profile != "rpc-server" {
					tc.Fatal("legacy millisecond instrument must remain independent", tempo, plan.Profile)
				}
			}
		}
	}, options(40)...)
}

func rpcIDs357(f rpcFixture357, current bool) []string {
	out := []string{}
	for _, r := range f.Rows {
		if rpcMember357(f, r, current) {
			out = append(out, r.ID)
		}
	}
	slices.Sort(out)
	return out
}
func rpcCount357(f rpcFixture357, current bool) string {
	return strconv.Itoa(len(rpcIDs357(f, current)))
}
