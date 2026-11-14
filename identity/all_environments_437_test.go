package identity

import (
	"encoding/json"
	"testing"
)

func TestAllEnvironments437IsSelectionNotObservation(t *testing.T) {
	all := AllEnvironments()
	if all.Validate() != nil || !all.All() || all.CurrentOnly() || !all.Equal(AllEnvironments()) {
		t.Fatal("explicit All policy is invalid")
	}
	if (EnvironmentScope{Mode: "all"}).Validate() == nil {
		t.Fatal("All became an observed environment")
	}
	if _, err := EffectiveEnvironment(EnvironmentScope{Mode: "all"}); err == nil {
		t.Fatal("observed-scope constructor admitted All")
	}
	for _, resource := range []map[string]string{nil, {}, {EnvironmentKey: ""}, {LegacyEnvironmentKey: "prod"}, {EnvironmentKey: "", LegacyEnvironmentKey: "prod"}, {EnvironmentKey: "all"}, {EnvironmentKey: "prod", LegacyEnvironmentKey: "other"}} {
		if !all.Matches(resource) || ObserveEnvironment(resource).Scope.Mode == "all" {
			t.Fatal("All changed observation classification", resource)
		}
	}
	literal, err := CurrentEnvironment("all")
	if err != nil || literal.All() || literal.Equal(all) || literal.Matches(map[string]string{EnvironmentKey: "prod"}) || !literal.Matches(map[string]string{EnvironmentKey: "all"}) {
		t.Fatal("literal environment name was broadened")
	}
	copy := all.Scope()
	copy.Mode = "named"
	if !all.All() || all.Validate() != nil {
		t.Fatal("scope copy mutated checked selection")
	}
	if _, err := json.Marshal(all); err == nil {
		t.Fatal("private selection acquired an accidental wire format")
	}
}
