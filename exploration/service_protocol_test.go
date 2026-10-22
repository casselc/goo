package exploration

import (
	"encoding/json"
	"github.com/casselc/goo/identity"
	"os"
	"reflect"
	"testing"
)

func TestServiceNamespaceProtocolFixtures(t *testing.T) {
	raw, e := os.ReadFile("../testdata/protocol/service-namespaces.json")
	if e != nil {
		t.Fatal(e)
	}
	var f struct {
		Version   int
		Resources []map[string]string
		Cases     []struct {
			Name                 string
			Namespace            *identity.Namespace
			Canonical, Predicate json.RawMessage
			Key                  string
			Matches              []bool
		}
	}
	if e = json.Unmarshal(raw, &f); e != nil {
		t.Fatal(e)
	}
	if f.Version != 1 || len(f.Cases) != 8 {
		t.Fatal("fixture coverage changed")
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			if e := c.Namespace.Validate(); e != nil {
				t.Fatal(e)
			}
			compare := func(value any, want json.RawMessage) {
				t.Helper()
				b, _ := json.Marshal(value)
				var a, z any
				json.Unmarshal(b, &a)
				json.Unmarshal(want, &z)
				if !reflect.DeepEqual(a, z) {
					t.Fatalf("wire changed: %s != %s", b, want)
				}
			}
			compare(c.Namespace, c.Canonical)
			p := NamespacePredicate(c.Namespace)
			compare(p, c.Predicate)
			mode, value := "any", ""
			if c.Namespace != nil {
				mode, value = c.Namespace.Mode, c.Namespace.Value
			}
			key, _ := json.Marshal([]string{"api", mode, value})
			if string(key) != c.Key {
				t.Fatal("service identity key changed")
			}
			if len(c.Matches) != len(f.Resources) {
				t.Fatal("missing match cases")
			}
			for i, r := range f.Resources {
				if c.Namespace.Matches(r) != c.Matches[i] {
					t.Fatal("namespace presence changed", i)
				}
				values := map[Field]string{}
				for k, v := range r {
					values[Field{Kind: "resource", Key: k}] = v
				}
				if Match(p, values) != c.Matches[i] {
					t.Fatal("record predicate differs from service selector", i)
				}
			}
		})
	}
}
