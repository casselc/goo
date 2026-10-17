package prebuilt

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestCatalogAndPublishedLibraryAgree(t *testing.T) {
	b, err := os.ReadFile("../definitions/prebuilts.json")
	if err != nil {
		t.Fatal(err)
	}
	var published Registry
	if err = json.Unmarshal(b, &published); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(published, Catalog()) {
		t.Fatal("regenerate definitions/prebuilts.json with goo catalog")
	}
	seen := map[string]bool{}
	for _, p := range Catalog().Profiles {
		if p.ID == "" || seen[p.ID] || len(p.Requirements) == 0 || len(p.Caveats) == 0 {
			t.Fatalf("invalid profile %s", p.ID)
		}
		seen[p.ID] = true
		for _, req := range p.Requirements {
			if len(req.AnyOf) == 0 {
				t.Fatal("empty requirement")
			}
			for _, c := range req.AnyOf {
				if c.SchemaURL == "" || c.Kind == "" || (c.Name == "" && !strings.HasPrefix(c.Kind, "span/")) || c.Unit == "" {
					t.Fatalf("unversioned contract: %+v", c)
				}
			}
		}
	}
}
func TestUnitsVersionAndCoobservedDimensionsAreRequired(t *testing.T) {
	expected := Catalog().Profiles[0].Requirements[0].AnyOf[0]
	base := Scope{ID: "a", Resource: map[string]string{"service.name": "checkout"}, Signals: []Contract{expected}}
	run := func(scope Scope) string {
		t.Helper()
		p, e := Discover(Inventory{APIVersion: "goo/inventory/v1alpha1", Basis: "fixture", Scopes: []Scope{scope}})
		if e != nil {
			t.Fatal(e)
		}
		for _, x := range p {
			if x.Profile == "otel.http.server" {
				return x.Status
			}
		}
		t.Fatal("missing profile")
		return ""
	}
	if run(base) != "compatible-shape" {
		t.Fatal("expected matching contract")
	}
	for _, field := range []string{"unit", "schema"} {
		wrong := expected
		if field == "unit" {
			wrong.Unit = "ms"
		} else {
			wrong.SchemaURL = ""
		}
		s := base
		s.Signals = []Contract{wrong}
		if run(s) == "compatible-shape" {
			t.Fatal("guessed unit/version")
		}
	}
	a, b := expected, expected
	a.Attributes = []string{"http.route"}
	b.Attributes = []string{"http.request.method", "http.response.status_code"}
	split := base
	split.Signals = []Contract{a, b}
	if run(split) == "compatible-shape" {
		t.Fatal("merged unrelated series attributes")
	}
}
func TestStrictInventory(t *testing.T) {
	for _, source := range []string{`{"apiVersion":"bad","basis":"x"}`, `{"apiVersion":"goo/inventory/v1alpha1","basis":"x","extra":1}`, `{"apiVersion":"goo/inventory/v1alpha1","basis":"x"} {}`} {
		if _, err := ReadInventory(strings.NewReader(source)); err == nil {
			t.Fatal("accepted invalid inventory", source)
		}
	}
}
