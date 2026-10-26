package main

import (
	"encoding/json"
	"testing"

	"github.com/casselc/goo/measurements"
)

func TestPodYAMLPublishReopenCanonicalResourceServicesAndAttributes(t *testing.T) {
	t.Chdir("../..")
	var in measurements.PodDashboardRequest
	if err := json.Unmarshal([]byte(podMetricHTTPPlanBody(t, nil)), &in); err != nil {
		t.Fatal(err)
	}
	plan, err := measurements.PlanPodDashboard(in)
	if err != nil || len(plan.Widgets) != 2 {
		t.Fatal(plan, err)
	}
	definition := plan.Widgets[0].Question.Definition
	if definition.Source.Resource == nil || definition.Source.Resource.Kind != "kubernetes-pod" {
		t.Fatal("missing pod authority")
	}
	if definition.Source.ExactStream == nil || len(definition.Source.ExactStream.Attributes) != 0 {
		t.Fatal("fixture must be dimensionless")
	}
	// Actual publisher, duplicate publication, fresh catalog, disk YAML and revision
	// checks are identical for both resource authorities. Direct JSON stays separate.
	verifyHostYAMLPublicationVariants(t, definition, true)
}
