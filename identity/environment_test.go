package identity

import "testing"

func TestEnvironmentPresencePrecedenceAndNoWildcard(t *testing.T) {
	values := []*string{nil, ptrEnvironment(""), ptrEnvironment("prod"), ptrEnvironment("other")}
	scopes := []EnvironmentScope{{Mode: "unreported"}, {Mode: "empty"}, {Mode: "named", Value: "prod"}, {Mode: "named", Value: "other"}}
	for _, current := range values {
		for _, legacy := range values {
			resource := map[string]string{}
			if current != nil {
				resource[EnvironmentKey] = *current
			}
			if legacy != nil {
				resource[LegacyEnvironmentKey] = *legacy
			}
			got := ObserveEnvironment(resource)
			winner := current
			key := EnvironmentKey
			if winner == nil {
				winner = legacy
				key = LegacyEnvironmentKey
			}
			mode, value := "unreported", ""
			if winner != nil {
				mode = "named"
				value = *winner
				if value == "" {
					mode = "empty"
				}
			} else {
				key = ""
			}
			if got.Scope != (EnvironmentScope{Mode: mode, Value: value}) || got.WinningKey != key || got.Conflict != (current != nil && legacy != nil && *current != *legacy) {
				t.Fatalf("wrong precedence: %+v", got)
			}
			matched := 0
			for _, scope := range scopes {
				if scope.Matches(resource) {
					matched++
					if scope != got.Scope {
						t.Fatal("different selection matched", scope, got)
					}
				}
			}
			if matched != 1 {
				t.Fatal("missing/empty/named conflated", matched, got)
			}
		}
	}
	for _, scope := range []EnvironmentScope{{}, {Mode: "any"}, {Mode: "unreported", Value: "prod"}, {Mode: "empty", Value: "x"}, {Mode: "named"}, {Mode: "named", Value: "\xff"}} {
		if scope.Validate() == nil || scope.Matches(nil) || scope.Equal(scope) {
			t.Fatal("invalid selector acts as wildcard", scope)
		}
	}
	named := EnvironmentScope{Mode: "named", Value: "unreported"}
	if named.Matches(nil) || !named.Matches(map[string]string{EnvironmentKey: "unreported"}) {
		t.Fatal("magic string overload")
	}
}
func ptrEnvironment(s string) *string { return &s }
