package prebuilt

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func packageFixture(t *testing.T) (Inventory, Binding, []HTTPRequest) {
	t.Helper()
	inf, e := os.Open("../examples/packages/inventory.json")
	if e != nil {
		t.Fatal(e)
	}
	defer inf.Close()
	in, e := ReadInventory(inf)
	if e != nil {
		t.Fatal(e)
	}
	bf, e := os.Open("../examples/packages/http.yaml")
	if e != nil {
		t.Fatal(e)
	}
	defer bf.Close()
	b, e := ReadBinding(bf)
	if e != nil {
		t.Fatal(e)
	}
	data, e := os.ReadFile("../examples/packages/requests.json")
	if e != nil {
		t.Fatal(e)
	}
	var req []HTTPRequest
	if e = json.Unmarshal(data, &req); e != nil {
		t.Fatal(e)
	}
	return in, b, req
}
func TestExpandedHTTPUsesSharedModelAndUnknownGuard(t *testing.T) {
	in, b, req := packageFixture(t)
	expanded, err := Expand(in, b)
	if err != nil {
		t.Fatal(err)
	}
	p, err := expanded.PrepareHTTP(req, "basis-1")
	if err != nil {
		t.Fatal(err)
	}
	result, err := expanded.Measurement.Evaluate(p.Facts, "basis-1", nil)
	if err != nil {
		t.Fatal(err)
	}
	m := result[0].Measures
	if *m["requests"].Count != 4 || *m["successful"].Count != 2 || *m["failures"].Count != 1 || *m["unknown"].Count != 1 || m["success_ratio"].State != "unknown" {
		t.Fatalf("unexpected measures: %+v", m)
	}
	grouped, err := expanded.Measurement.Evaluate(p.Facts, "basis-1", []string{"route", "method"})
	if err != nil {
		t.Fatal(err)
	}
	if len(grouped) != 2 || len(expanded.Plans) != 2 || len(expanded.Views) != 1 {
		t.Fatal("missing shared artifacts")
	}
	if expanded.Revision == "" || expanded.ProfileRevision == "" || p.ExpansionRevision != expanded.Revision {
		t.Fatal("missing provenance")
	}
	if !strings.Contains(expanded.Plans["databricks"].Status, "prepared") {
		t.Fatal("backend must require prepared source")
	}
}
func TestBindingAndPreparationRejectUnsafeAssumptions(t *testing.T) {
	in, b, req := packageFixture(t)
	b.Policy.MissingStatus = "success"
	if _, err := Expand(in, b); err == nil {
		t.Fatal("implicit health")
	}
	b.Policy.MissingStatus = "unknown"
	b.Profile = "otel.database.pool"
	if _, err := Expand(in, b); err == nil {
		t.Fatal("unsupported expansion")
	}
	b.Profile = "otel.http.server.spans"
	in.Scopes[0].Signals[0].SchemaURL = ""
	if _, err := Expand(in, b); err == nil {
		t.Fatal("unrecognized source")
	}
	in.Scopes[0].Signals[0].SchemaURL = "https://opentelemetry.io/schemas/1.37.0"
	out, err := Expand(in, b)
	if err != nil {
		t.Fatal(err)
	}
	duplicate := append(append([]HTTPRequest{}, req...), req[0])
	if _, err = out.PrepareHTTP(duplicate, "e"); err == nil {
		t.Fatal("duplicate subject")
	}
	mutated, err := Expand(in, b)
	if err != nil {
		t.Fatal(err)
	}
	mutated.Binding.Policy.MissingStatus = "success"
	if _, err = mutated.PrepareHTTP(req, "e"); err == nil {
		t.Fatal("mutated expansion reused identity")
	}
	req[0].Kind = "client"
	if _, err = out.PrepareHTTP(req, "e"); err == nil {
		t.Fatal("client span accepted")
	}
	req[0].Kind = "server"
	req[0].Resource["service.name"] = "other"
	if _, err = out.PrepareHTTP(req, "e"); err == nil {
		t.Fatal("scope leakage")
	}
}
func TestExpansionIdentityAndUnboundTargets(t *testing.T) {
	in, b, _ := packageFixture(t)
	a, e := Expand(in, b)
	if e != nil {
		t.Fatal(e)
	}
	in.Scopes[0].Signals = append(in.Scopes[0].Signals, in.Scopes[0].Signals[0])
	same, e := Expand(in, b)
	if e != nil {
		t.Fatal(e)
	}
	if a.Revision != same.Revision {
		t.Fatal("duplicate changed expansion")
	}
	b.Policy.FailureStatusRanges = []StatusRange{{Min: 400, Max: 599}}
	changed, e := Expand(in, b)
	if e != nil {
		t.Fatal(e)
	}
	if changed.Revision == a.Revision {
		t.Fatal("policy identity lost")
	}
	delete(b.PreparedTables, "databricks")
	single, e := Expand(in, b)
	if e != nil {
		t.Fatal(e)
	}
	if len(single.Plans) != 1 {
		t.Fatal("unbound target emitted plan")
	}
}
