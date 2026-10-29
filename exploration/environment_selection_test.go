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
	// source callers such as trace search intentionally omit q.Environment;
	// their checked shared parameter is authoritative. Missing bindings fail closed.
	c := compiler{signal: "logs", params: map[string]string{}}
	if _, _, err := c.source(Query{Signal: "logs"}); err == nil {
		t.Fatal("unbound compiler source accepted")
	}
}
