package identity

import "testing"

func TestMissingEmptyLiteralAndWildcard(t *testing.T) {
	for _, resource := range []map[string]string{{}, {NamespaceKey: ""}, {NamespaceKey: "payments"}} {
		if !(*Namespace)(nil).Matches(resource) {
			t.Fatal("wildcard")
		}
	}
	missing := Namespace{Mode: "missing"}
	empty := Namespace{Mode: "value"}
	named := Namespace{Mode: "value", Value: "payments"}
	if !missing.Matches(nil) || missing.Matches(map[string]string{NamespaceKey: ""}) || empty.Matches(nil) || !empty.Matches(map[string]string{NamespaceKey: ""}) || !named.Matches(map[string]string{NamespaceKey: "payments"}) || named.Equal(&empty) {
		t.Fatal("namespace conflation")
	}
	for _, n := range []*Namespace{{Mode: "any"}, {Mode: "missing", Value: "x"}, {Mode: "value", Value: "a\x00b"}} {
		if n.Validate() == nil {
			t.Fatal("invalid namespace accepted")
		}
	}
}
