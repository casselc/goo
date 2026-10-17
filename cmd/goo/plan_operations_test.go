package main

import (
	"bytes"
	"encoding/json"
	"github.com/casselc/goo/annotations"
	"testing"
)

func TestPlanOperationsCLI(t *testing.T) {
	for _, name := range []string{"checkout", "reports"} {
		var out bytes.Buffer
		if err := run([]string{"plan-operations", "-package", "../../config/operations/" + name + ".yaml"}, &out); err != nil {
			t.Fatal(err)
		}
		var plan annotations.OperationPlan
		if err := json.Unmarshal(out.Bytes(), &plan); err != nil {
			t.Fatal(err)
		}
		if plan.Definition.Name != name || plan.Revision == "" {
			t.Fatal("missing planned contract")
		}
	}
}
