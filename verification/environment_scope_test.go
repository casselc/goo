package verification

import (
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"hegel.dev/go/hegel"
	"strings"
	"testing"
)

func TestHegelEnvironmentSelection(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		resource := map[string]string{}
		values := []string{"", "prod", "unreported", "雪", "prod' OR 1=1 --"}
		current, legacy := hegel.Draw(tc, hegel.Booleans()), hegel.Draw(tc, hegel.Booleans())
		cv, lv := hegel.Draw(tc, hegel.SampledFrom(values)), hegel.Draw(tc, hegel.SampledFrom(values))
		if current {
			resource[identity.EnvironmentKey] = cv
		}
		if legacy {
			resource[identity.LegacyEnvironmentKey] = lv
		}
		mode, value := "unreported", ""
		if current || legacy {
			mode = "named"
			value = lv
			if current {
				value = cv
			}
			if value == "" {
				mode = "empty"
			}
		}
		observation := identity.ObserveEnvironment(resource)
		expected := identity.EnvironmentScope{Mode: mode, Value: value}
		if observation.Scope != expected || !expected.Matches(resource) || observation.Conflict != (current && legacy && cv != lv) {
			tc.Fatal("presence/precedence changed", observation)
		}
		for _, scope := range []identity.EnvironmentScope{{Mode: "unreported"}, {Mode: "empty"}, {Mode: "named", Value: "prod"}, {Mode: "named", Value: "unreported"}, {Mode: "named", Value: "雪"}, {Mode: "named", Value: "prod' OR 1=1 --"}} {
			if scope.Matches(resource) != (scope == expected) {
				tc.Fatal("scope leaked", scope, expected)
			}
			sql, params, e := exploration.EnvironmentPredicate(scope)
			if e != nil {
				tc.Fatal(e)
			}
			if scope.Mode == "named" && (params["environment"] != scope.Value || strings.Contains(sql, scope.Value)) {
				tc.Fatal("value not parameterized")
			}
		}
		if (identity.EnvironmentScope{}).Matches(resource) {
			tc.Fatal("zero scope wildcard")
		}
	}, options(200)...)
}
