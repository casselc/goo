package subjectfacts_test

import (
	"fmt"
	"strings"

	"github.com/casselc/goo/subjectfacts"
)

// The source adapter prepares one typed fact per scoped subject and evaluation.
// This example needs no database, telemetry exporter or evidence store.
func ExampleDefinition_Evaluate() {
	definition, err := subjectfacts.Read(strings.NewReader(`
apiVersion: goo/v1alpha1
name: request_results
source:
  kind: subject-facts
  subjectKey: [tenant, request_id]
  fields: {evaluation_id: string, tenant: string, request_id: string, outcome: string}
  enums: {outcome: [success, failure, unknown]}
  bindings:
    clickhouse: request_facts
    databricks: telemetry.semantic.request_facts
measures:
  requests: {op: count}
  successful: {op: count, where: {field: outcome, eq: success}}
  unknown: {op: count, where: {field: outcome, eq: unknown}}
  success_ratio: {op: ratio, numerator: successful, denominator: requests, unknownWhen: [unknown]}
`))
	if err != nil {
		panic(err)
	}
	facts := []subjectfacts.Fact{
		{"evaluation_id": "batch-1", "tenant": "shop", "request_id": "a", "outcome": "success"},
		{"evaluation_id": "batch-1", "tenant": "shop", "request_id": "b", "outcome": "failure"},
	}
	rows, err := definition.Evaluate(facts, "batch-1", nil)
	if err != nil {
		panic(err)
	}
	ratio := rows[0].Measures["success_ratio"]
	fmt.Println(ratio.State, *ratio.Ratio)
	// Compilation produces inspectable prepared-source artifacts, not a deployment.
	for _, target := range []string{"clickhouse", "databricks"} {
		plan, err := definition.Compile(target, nil)
		if err != nil {
			panic(err)
		}
		fmt.Println(plan.Target, plan.Status)
	}
	// Output:
	// ok 0.5
	// clickhouse requires-prepared-source
	// databricks requires-prepared-source
}
