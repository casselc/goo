package annotations

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestOperationPackagePlanning(t *testing.T) {
	p, err := LoadOperationPackage("../config/operations/checkout.yaml")
	if err != nil {
		t.Fatal(err)
	}
	again, err := p.Definition.Bind(p.Manifest)
	if err != nil || again.Revision != p.Revision {
		t.Fatal("nondeterministic plan", err)
	}
	before, _ := json.Marshal(p.Definition)
	_, err = p.Definition.Bind(p.Manifest)
	after, _ := json.Marshal(p.Definition)
	if err != nil || string(before) != string(after) {
		t.Fatal("planner mutated definition")
	}
	reports, err := LoadOperationPackage("../config/operations/reports.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if reports.Source.Service == p.Source.Service || reports.Source.SubjectNamespace == p.Source.SubjectNamespace || reflect.DeepEqual(reports.Definition.Views, p.Definition.Views) {
		t.Fatal("second package is not independently bound")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("param_service") != p.Source.Service {
			t.Error("wrong resource")
		}
	}))
	defer server.Close()
	observed, err := p.Fetch(context.Background(), server.Client(), server.URL, "", "", 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if observed.Package.Revision != p.Revision || observed.Conformance.Basis == "" {
		t.Fatal("missing package provenance")
	}
	p.Source.Service = "changed"
	if _, err = p.Fetch(context.Background(), server.Client(), server.URL, "", "", 0, 1); err == nil {
		t.Fatal("accepted mutated plan")
	}
}
func TestOperationPackageRejectsUnsupportedContracts(t *testing.T) {
	data, err := os.ReadFile("../config/operations/checkout.yaml")
	if err != nil {
		t.Fatal(err)
	}
	m, err := ScanDir("../examples/operation-app")
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(string) string{
		func(s string) string { return strings.Replace(s, "role: server", "role: client", 1) },
		func(s string) string { return strings.Replace(s, "version: 0.1.0", "version: 9.0.0", 1) },
		func(s string) string { return strings.Replace(s, "sampling: unverified", "sampling: complete", 1) },
		func(s string) string { return strings.Replace(s, "tenant, id", "id", 1) },
		func(s string) string { return strings.Replace(s, "trace-waterfall", "unknown-view", 1) },
		func(s string) string { return strings.Replace(s, "reservation.terminal", "undeclared.event", 1) },
		func(s string) string { return s + "unknown: true\n" },
		func(s string) string { return s + "name: duplicate\n" },
		func(s string) string { return s + "---\n{}\n" },
		func(s string) string {
			return strings.Replace(s, "title: Checkout operations", "title: &alias Checkout operations\nunused: *alias", 1)
		},
	} {
		definition, err := ReadOperationPackage(strings.NewReader(change(string(data))))
		if err == nil {
			_, err = definition.Bind(m)
		}
		if err == nil {
			t.Fatal("accepted unsupported package")
		}
	}
	// Declaration ordering is not semantics; source binding and enabled views are.
	p, err := ReadOperationPackage(strings.NewReader(string(data)))
	if err != nil {
		t.Fatal(err)
	}
	a, err := p.Bind(m)
	if err != nil {
		t.Fatal(err)
	}
	p.Operations[0], p.Operations[1] = p.Operations[1], p.Operations[0]
	b, err := p.Bind(m)
	if err != nil || a.Revision != b.Revision {
		t.Fatal("declaration order changed semantic revision")
	}
	p.Resource.Service = "different"
	c, err := p.Bind(m)
	if err != nil || c.Revision == a.Revision {
		t.Fatal("resource omitted from revision")
	}
}
