package exploration

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/identity"
	"strings"
	"testing"
)

// Frozen before the shared environment lowering extraction. Full serialized
// plans include row/count/group/field/time SQL and all parameter names/values.
func TestCurrentEnvironmentPlanCompatibility248(t *testing.T) {
	hashes := map[string]string{"logs": "80b513bb6893eaff3913ad0aa889a718edf132a99c73ee6f51b385a123f96fb9", "spans": "95162515ce79f97113b85ad356e694c312e51a001d8a97d881b5e03be214936d"}
	for signal, want := range hashes {
		q := Query{Signal: signal, Environment: "prod'雪", FromMS: 60000, ToMS: 120000, Limit: 20, Offset: 3, BucketMS: 10000, GroupBy: []Field{{Kind: "resource", Key: "service.namespace"}}, Where: &Predicate{Field: &Field{Kind: "attribute", Key: "tenant"}, Op: "eq", Value: "blue"}}
		plan, err := Compile(q)
		if err != nil {
			t.Fatal(err)
		}
		b, err := json.Marshal(plan)
		if err != nil {
			t.Fatal(err)
		}
		if got := fmt.Sprintf("%x", sha256.Sum256(b)); got != want {
			t.Fatalf("%s plan changed: %s", signal, got)
		}
	}
}

func TestEnvironmentLoweringRefusesZeroAndPreservesPolicy(t *testing.T) {
	if sql, params, err := environmentSelectionPredicate(identity.EnvironmentSelection{}); err == nil || sql != "" || params != nil {
		t.Fatal("zero selection lowered", sql, params, err)
	}
	const name = "prod' OR 1=1 -- 雪"
	current, _ := identity.CurrentEnvironment(name)
	effective, _ := identity.EffectiveEnvironment(identity.EnvironmentScope{Mode: "named", Value: name})
	a, ap, err := environmentSelectionPredicate(current)
	if err != nil {
		t.Fatal(err)
	}
	b, bp, err := environmentSelectionPredicate(effective)
	if err != nil {
		t.Fatal(err)
	}
	if a == b || strings.Contains(a, "'deployment.environment'") || !strings.Contains(b, "'deployment.environment'") {
		t.Fatal("policy conflation", a, b)
	}
	for _, sql := range []string{a, b} {
		if strings.Contains(sql, name) {
			t.Fatal("unbound name", sql)
		}
	}
	if ap["environment"] != name || bp["environment"] != name {
		t.Fatal("bound value changed")
	}
	// Zero checked selections must fail even if an old binding exists.
	c := compiler{signal: "logs", params: map[string]string{}}
	if _, _, err := c.source(Query{Signal: "logs"}, identity.EnvironmentSelection{}); err == nil {
		t.Fatal("unbound compiler source accepted")
	}
}

func TestCheckedSourceBindings249(t *testing.T) {
	current, _ := identity.CurrentEnvironment("prod")
	missing, _ := identity.EffectiveEnvironment(identity.EnvironmentScope{Mode: "unreported"})
	empty, _ := identity.EffectiveEnvironment(identity.EnvironmentScope{Mode: "empty"})
	for _, tc := range []struct {
		selection identity.EnvironmentSelection
		bindings  map[string]string
		ok        bool
	}{
		{current, nil, true}, {current, map[string]string{"environment": "prod"}, true}, {current, map[string]string{"environment": "other"}, false},
		{missing, nil, true}, {empty, nil, true}, {missing, map[string]string{"environment": "prod"}, false}, {empty, map[string]string{"environment": ""}, false},
		{identity.EnvironmentSelection{}, map[string]string{"environment": "prod"}, false},
	} {
		c := compiler{signal: "logs", params: tc.bindings}
		_, _, err := c.source(Query{Signal: "logs"}, tc.selection)
		if (err == nil) != tc.ok {
			t.Fatal("checked binding admission changed", tc, err)
		}
		if err == nil && tc.selection.CurrentOnly() && c.params["environment"] != "prod" {
			t.Fatal("binding not derived from checked selection")
		}
	}
}
