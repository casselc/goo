package identity

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEnvironmentSelectionPolicies(t *testing.T) {
	current, err := CurrentEnvironment("prod")
	if err != nil {
		t.Fatal(err)
	}
	effective, err := EffectiveEnvironment(EnvironmentScope{Mode: "named", Value: "prod"})
	if err != nil {
		t.Fatal(err)
	}
	if current.Equal(effective) || !current.Equal(current) || !effective.Equal(effective) {
		t.Fatal("policy lost from equality")
	}
	values := []*string{nil, ptrEnvironment(""), ptrEnvironment("prod"), ptrEnvironment("other")}
	for _, c := range values {
		for _, l := range values {
			resource := map[string]string{}
			if c != nil {
				resource[EnvironmentKey] = *c
			}
			if l != nil {
				resource[LegacyEnvironmentKey] = *l
			}
			if current.Matches(resource) != (c != nil && *c == "prod") {
				t.Fatal("current-only changed", resource)
			}
			want := c != nil && *c == "prod" || c == nil && l != nil && *l == "prod"
			if effective.Matches(resource) != want {
				t.Fatal("fallback changed", resource)
			}
			if (EnvironmentSelection{}).Matches(resource) {
				t.Fatal("zero value matches")
			}
		}
	}
	copy := effective.Scope()
	copy.Value = "other"
	if effective.Scope().Value != "prod" {
		t.Fatal("scope accessor mutates selection")
	}
	if (EnvironmentSelection{}).Validate() == nil || (EnvironmentSelection{}).Equal(EnvironmentSelection{}) {
		t.Fatal("zero selection valid")
	}
	for _, name := range []string{"", strings.Repeat("x", 129), "bad\x00name", "\xff"} {
		selection, err := CurrentEnvironment(name)
		if err == nil || selection.Validate() == nil || selection.Matches(nil) {
			t.Fatal("invalid name accepted", name)
		}
	}
	for _, scope := range []EnvironmentScope{{}, {Mode: "any"}, {Mode: "named"}, {Mode: "empty", Value: "prod"}} {
		selection, err := EffectiveEnvironment(scope)
		if err == nil || selection.Validate() == nil {
			t.Fatal("invalid effective scope accepted", scope)
		}
	}
	for _, scope := range []EnvironmentScope{{Mode: "empty"}, {Mode: "unreported"}, {Mode: "named", Value: "雪"}} {
		selection, err := EffectiveEnvironment(scope)
		if err != nil || !selection.Scope().Equal(scope) || selection.CurrentOnly() {
			t.Fatal("effective selector changed", scope)
		}
	}
}

func TestEnvironmentSelectionHasNoJSONIdentity(t *testing.T) {
	current, _ := CurrentEnvironment("prod")
	effective, _ := EffectiveEnvironment(EnvironmentScope{Mode: "named", Value: "prod"})
	for _, selection := range []EnvironmentSelection{{}, current, effective} {
		if _, err := json.Marshal(selection); err == nil {
			t.Fatal("selection silently serialized without policy")
		}
	}
	for _, data := range []string{`{}`, `null`, `{"environment":"prod"}`} {
		var selection EnvironmentSelection
		if err := json.Unmarshal([]byte(data), &selection); err == nil || selection.Validate() == nil {
			t.Fatal("unversioned selection accepted", data)
		}
	}
}
