package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"hegel.dev/go/hegel"
)

type httpRaw385 struct {
	At           int64             `json:"timestamp_ms,string"`
	ID           string            `json:"SpanId"`
	TraceID      string            `json:"TraceId,omitempty"`
	ParentSpanID string            `json:"ParentSpanId,omitempty"`
	Operation    string            `json:"SpanName,omitempty"`
	TimestampNS  string            `json:"timestamp_ns,omitempty"`
	Service      string            `json:"ServiceName"`
	Kind         string            `json:"SpanKind"`
	Resource     map[string]string `json:"ResourceAttributes"`
	Attributes   map[string]string `json:"SpanAttributes"`
	Status       string            `json:"StatusCode"`
	DurationNS   string            `json:"duration_ns"`
}
type httpFixture385 struct {
	Mode    int                           `json:"mode"`
	Scope   exploration.InventoryQuestion `json:"scope"`
	Service string                        `json:"service"`
	Rows    []httpRaw385                  `json:"rows"`
}

func httpFixtureFor385(seed, mode, namespace int) httpFixture385 {
	f := httpFixture385{Mode: mode, Service: fmt.Sprintf("caller'π\\\n-%d", seed), Scope: exploration.InventoryQuestion{Environment: "prod", FromMS: int64(seed) + 1000, ToMS: int64(seed) + 61000}}
	switch namespace {
	case 1:
		f.Scope.Namespace = &identity.Namespace{Mode: "missing"}
	case 2:
		f.Scope.Namespace = &identity.Namespace{Mode: "value", Value: ""}
	case 3:
		f.Scope.Namespace = &identity.Namespace{Mode: "value", Value: "team'π\n"}
	}
	if mode > 0 {
		f.Scope.Environment = ""
		sel := identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}}
		if mode == 1 {
			sel.Policy = "current-only"
		}
		if mode == 3 {
			sel.Scope = identity.EnvironmentScope{Mode: "empty"}
		}
		if mode == 4 {
			sel.Scope = identity.EnvironmentScope{Mode: "unreported"}
		}
		f.Scope.EnvironmentSelector = &sel
	}
	envs := []map[string]string{{"deployment.environment.name": "prod"}, {"deployment.environment": "prod"}, {"deployment.environment.name": ""}, {}, {"deployment.environment.name": "foreign", "deployment.environment": "prod"}, {"deployment.environment.name": "", "deployment.environment": "prod"}}
	methods := []*string{nil, lpcPointer(""), lpcPointer("GET"), lpcPointer("POST"), lpcPointer("_OTHER"), lpcPointer(" "), lpcPointer("WEIRD'π\n")}
	for _, env := range envs {
		for _, method := range methods {
			for _, kind := range []string{"Client", "Server", "Internal", "client", "CLIENT", "", "Consumer"} {
				resource := map[string]string{}
				for k, v := range env {
					resource[k] = v
				}
				switch namespace {
				case 0:
					if len(f.Rows)%2 == 0 {
						resource["service.namespace"] = ""
					}
				case 2:
					resource["service.namespace"] = ""
				case 3:
					resource["service.namespace"] = f.Scope.Namespace.Value
				}
				attrs := map[string]string{"http.method": "LEGACY", "url.full": "https://not-an-instrumented-service.example/path"}
				if method != nil {
					attrs["http.request.method"] = *method
				}
				switch len(f.Rows) % 4 {
				case 1:
					attrs["server.address"] = ""
				case 2:
					attrs["server.address"] = "peer'π\n.example"
				case 3:
					attrs["server.address"] = "Missing key"
				}
				switch len(f.Rows) % 8 {
				case 1:
					attrs["server.port"] = ""
				case 2:
					attrs["server.port"] = "443"
				case 3:
					attrs["server.port"] = "invalid"
				case 4:
					attrs["server.port"] = "0"
				case 5:
					attrs["server.port"] = "0443"
				case 6:
					attrs["server.port"] = "443 "
				case 7:
					attrs["server.port"] = "65536"
				}
				switch len(f.Rows) % 5 {
				case 1:
					attrs["http.response.status_code"] = ""
				case 2:
					attrs["http.response.status_code"] = "200"
				case 3:
					attrs["http.response.status_code"] = "503"
				case 4:
					attrs["http.response.status_code"] = "unknown"
				}
				statuses := []string{"Error", "Ok", "Unset", "", "ERROR"}
				n := len(f.Rows)
				f.Rows = append(f.Rows, httpRaw385{At: f.Scope.FromMS + int64(n%1000), ID: fmt.Sprintf("%016x", n+1), Service: f.Service, Kind: kind, Resource: resource, Attributes: attrs, Status: statuses[n%len(statuses)], DurationNS: strconv.Itoa(1000000 + n)})
			}
		}
	}
	// Explicit out-of-scope decoys retain the same conventional attributes.
	for index := range 5 {
		r := f.Rows[14]
		r.ID = fmt.Sprintf("%016x", len(f.Rows)+1)
		r.Resource = map[string]string{}
		for k, v := range f.Rows[14].Resource {
			r.Resource[k] = v
		}
		switch index {
		case 0:
			r.Service = "foreign"
		case 1:
			r.At = f.Scope.FromMS - 1
		case 2:
			r.At = f.Scope.ToMS
		case 3:
			r.Resource["service.namespace"] = "foreign"
		case 4:
			r.Kind = "Server"
		}
		f.Rows = append(f.Rows, r)
	}
	// Exact duplicated source records remain records. A large bounded prefix
	// must never redefine the full count or deduplicate by SpanId.
	for _, row := range f.Rows {
		if httpMember385(f, row) {
			for range 105 {
				f.Rows = append(f.Rows, row)
			}
			break
		}
	}
	return f
}

// Literal convention boundary, independent from product descriptors and evaluator.
// Destination and response attributes are observations, never membership or health.
func httpMember385(f httpFixture385, r httpRaw385) bool {
	name := f.Scope.Environment
	if f.Scope.EnvironmentSelector != nil && f.Scope.EnvironmentSelector.Scope.Mode == "named" {
		name = f.Scope.EnvironmentSelector.Scope.Value
	}
	return r.Service == f.Service && r.Kind == "Client" && r.At >= f.Scope.FromMS && r.At < f.Scope.ToMS && continuationEnvironment352(f.Mode, name, r.Resource) && continuationNamespace352(f.Scope.Namespace, r.Resource) && r.Attributes["http.request.method"] != ""
}
func httpValues385(r httpRaw385) map[exploration.Field]string {
	v := map[exploration.Field]string{{Kind: "column", Key: "service"}: r.Service, {Kind: "column", Key: "kind"}: r.Kind, {Kind: "column", Key: "status"}: r.Status}
	if r.TraceID != "" {
		v[exploration.Field{Kind: "column", Key: "trace_id"}] = r.TraceID
	}
	if r.ID != "" {
		v[exploration.Field{Kind: "column", Key: "span_id"}] = r.ID
	}
	for k, s := range r.Resource {
		v[exploration.Field{Kind: "resource", Key: k}] = s
	}
	for k, s := range r.Attributes {
		v[exploration.Field{Kind: "attribute", Key: k}] = s
	}
	return v
}
func httpIDs385(f httpFixture385) []string {
	ids := []string{}
	for _, r := range f.Rows {
		if httpMember385(f, r) {
			ids = append(ids, r.ID)
		}
	}
	slices.Sort(ids)
	return ids
}
func httpPlan385(f httpFixture385, profile string, instruments []measurements.Instrument) (*exploration.DashboardPlan, error) {
	return exploration.PlanServiceDashboard(exploration.DashboardRequest{Scope: f.Scope, Service: f.Service, Profile: profile, Signals: []string{"spans", "logs", "metrics"}, Fields: []exploration.Field{{Kind: "attribute", Key: "http.request.method"}, {Kind: "attribute", Key: "server.address"}, {Kind: "attribute", Key: "server.port"}, {Kind: "attribute", Key: "http.response.status_code"}}, Instruments: instruments}, nil)
}
func httpGroupCounts385(f httpFixture385, key string) map[string]int {
	out := map[string]int{}
	for _, r := range f.Rows {
		if !httpMember385(f, r) {
			continue
		}
		var v *string
		if raw, present := r.Attributes[key]; present {
			v = lpcPointer(raw)
		}
		out[continuationJSON352(v)]++
	}
	return out
}
func TestHegelHTTPClientProfile385(t *testing.T) {
	fixtures := []httpFixture385{}
	hegel.Test(t, func(tc *hegel.T) {
		f := httpFixtureFor385(hegel.Draw(tc, hegel.Integers(1, 10000)), hegel.Draw(tc, hegel.Integers(0, 4)), hegel.Draw(tc, hegel.Integers(0, 3)))
		fixtures = append(fixtures, f)
		before := continuationJSON352(f)
		p, e := httpPlan385(f, "http-client", nil)
		if e != nil || p.Profile != "http-client" || continuationJSON352(f) != before {
			tc.Fatal("explicit current Client profile or detached input", e)
		}
		dimensions := map[string]string{"dimension": "server.address", "http-client-methods": "http.request.method", "http-client-ports": "server.port", "http-client-responses": "http.response.status_code"}
		found := map[string]bool{}
		for _, w := range p.Widgets {
			if w.Query == nil || w.Query.Signal != "spans" {
				continue
			}
			q := w.Query
			if q.Validate() != nil || q.Environment != f.Scope.Environment || continuationJSON352(q.EnvironmentSelector) != continuationJSON352(f.Scope.EnvironmentSelector) || q.FromMS != f.Scope.FromMS || q.ToMS != f.Scope.ToMS || q.Offset != 0 || q.Limit != 100 {
				tc.Fatal("Client fixed scope or inspection bound changed")
			}
			for _, r := range f.Rows {
				actual := hdPredicate(q.Where, httpValues385(r)) && r.At >= q.FromMS && r.At < q.ToMS && continuationEnvironment352(f.Mode, "prod", r.Resource)
				if actual != httpMember385(f, r) {
					tc.Fatal("Client method/kind/namespace/environment membership drift", w.Key, r.ID)
				}
			}
			if key, ok := dimensions[w.Key]; ok {
				found[w.Key] = true
				if !reflect.DeepEqual(q.GroupBy, []exploration.Field{{Kind: "attribute", Key: key}}) {
					tc.Fatal("qualified destination/method/port/response grouping drift", w.Key)
				}
			}
			// A smaller inspection page never changes the full population predicate.
			small := *q
			small.Limit = 1
			for _, r := range f.Rows {
				if hdPredicate(small.Where, httpValues385(r)) != hdPredicate(q.Where, httpValues385(r)) {
					tc.Fatal("list limit became membership")
				}
			}
			compiled, e := exploration.Compile(*q)
			if e != nil {
				tc.Fatal(e)
			}
			if strings.Contains(compiled.Count, f.Service) {
				tc.Fatal("literal service interpolated into SQL")
			}
			parameterized := false
			for _, v := range compiled.Parameters {
				parameterized = parameterized || v == f.Service
			}
			if !parameterized {
				tc.Fatal("service literal not parameterized")
			}
		}
		if len(found) != len(dimensions) {
			tc.Fatal("missing separate qualified group observations", found)
		}
		// New Client observations do not change published auto precedence.
		auto, e := httpPlan385(f, "auto", nil)
		if e != nil || auto.Profile != "all-spans" {
			tc.Fatal("explicit Client lens changed automatic profile", e)
		}
		// Metadata from old HTTP alone cannot grant the current lens.
		legacy := exploration.DashboardRequest{Scope: f.Scope, Service: f.Service, Profile: "http-client", Signals: []string{"spans"}, Fields: []exploration.Field{{Kind: "attribute", Key: "http.method"}}}
		for _, key := range []string{"http.method", "HTTP.REQUEST.METHOD", "http.request.Method", "http.request.method ", "server.address"} {
			legacy.Fields = []exploration.Field{{Kind: "attribute", Key: key}}
			if _, e := exploration.PlanServiceDashboard(legacy, nil); e == nil {
				tc.Fatal("foreign metadata key silently admitted current profile", key)
			}
		}
		bad := legacy
		bad.Fields = []exploration.Field{{Kind: "attribute", Key: "http.request.method"}}
		bad.Service = "caller\x00foreign"
		if _, e := exploration.PlanServiceDashboard(bad, nil); e == nil {
			tc.Fatal("NUL service accepted")
		}
	}, options(100)...)
	in, _ := json.Marshal(fixtures)
	cmd := exec.Command("node", "http_client_385_fixture.mjs")
	cmd.Stdin = bytes.NewReader(in)
	out, e := cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("mandatory batched raw Client witness: %v\n%s", e, out)
	}
	t.Log(string(out))
}

func TestHTTPClientNativeProfileRemainsGeneric385(t *testing.T) {
	scope := exploration.InventoryQuestion{Environment: "prod", FromMS: 1000, ToMS: 61000}
	instrument := hdInstrument("client-duration", "http.client.request.duration", "histogram", "s", scope, "caller")
	instrument.CapabilityPlan = &measurements.InstrumentCapabilityPlan{Profile: "http-client-duration", Eligible: true, ShapeMatches: true}
	in := exploration.DashboardRequest{Scope: scope, Service: "caller", Profile: "auto", Signals: []string{"metrics"}, Instruments: []measurements.Instrument{instrument}}
	plan, e := exploration.PlanServiceDashboard(in, nil)
	if e != nil || plan.Profile != "all-spans" {
		t.Fatal("native metadata promoted a client convention profile", e)
	}
	in.Profile = "http-client"
	if _, e = exploration.PlanServiceDashboard(in, nil); e == nil {
		t.Fatal("native client histogram granted record-profile admission without span metadata")
	}
}
