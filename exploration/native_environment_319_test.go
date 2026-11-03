package exploration

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
)

func nativeGaugeEnvironment319(mode string) measurements.Question {
	d := *dashboardTestInstrument("typed-gauge", "queue.depth", "gauge", "items").Definition
	d.Environment = ""
	d.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: mode}}
	if mode == "named" {
		d.EnvironmentSelector.Scope.Value = "prod"
	}
	d.Source.Namespace = &identity.Namespace{Mode: "value", Value: "team"}
	return measurements.Question{Definition: d, FromMS: 100000, ToMS: 160000}
}

func TestTypedNativeGaugeRetainedScope319(t *testing.T) {
	for _, mode := range []string{"named", "empty", "unreported"} {
		t.Run(mode, func(t *testing.T) {
			parent := nativeGaugeEnvironment319(mode)
			want := *parent.Definition.EnvironmentSelector
			capture := CaptureQuestion{Kind: "measurement", Measurement: &parent}
			normalized := capture.Normalized()
			scope, err := capture.Scope()
			if err != nil || scope.Environment != "" || scope.EnvironmentSelector == nil || *scope.EnvironmentSelector != want {
				t.Fatal("native scope dropped environment policy", err)
			}
			summary := summarizeCapture(CaptureReceipt{ID: "native-typed", Title: "Gauge", CapturedAt: "2026-11-03T12:00:00Z", Question: normalized})
			if summary.Environment != nil || summary.EnvironmentSelector == nil || *summary.EnvironmentSelector != want {
				t.Fatal("typed population became an invented named environment", summary)
			}
			discovery := measurements.DiscoveryQuestion{EnvironmentSelector: parent.Definition.EnvironmentSelector, Service: "api", Namespace: parent.Definition.Source.Namespace, FromMS: parent.FromMS, ToMS: parent.ToMS}
			if _, err := DescribeDiscoveryTransform(capture, discovery); err != nil {
				t.Fatal("same-policy discovery transform refused", err)
			}
			wrong := discovery
			wrong.EnvironmentSelector = nil
			wrong.Environment = "prod"
			if _, err := DescribeDiscoveryTransform(capture, wrong); err == nil {
				t.Fatal("discovery replaced policy with the same-looking name")
			}
			encoded, err := json.Marshal(normalized)
			if err != nil {
				t.Fatal(err)
			}
			var restored CaptureQuestion
			if err := json.Unmarshal(encoded, &restored); err != nil {
				t.Fatal(err)
			}
			parent.Definition.EnvironmentSelector.Policy = "broken"
			parent.Definition.Source.Namespace.Value = "changed"
			for _, held := range []CaptureQuestion{normalized, restored} {
				if err := held.Validate(); err != nil || *held.Measurement.Definition.EnvironmentSelector != want || held.Measurement.Definition.Source.Namespace.Value != "team" {
					t.Fatal("retained question aliases its caller", err)
				}
			}
			if *scope.EnvironmentSelector != want || scope.Binding.Namespace.Value != "team" || *summary.EnvironmentSelector != want {
				t.Fatal("scope/catalog metadata aliases its question")
			}
		})
	}
}

func TestTypedGaugeRecordContext319(t *testing.T) {
	for _, mode := range []string{"named", "empty", "unreported"} {
		t.Run(mode, func(t *testing.T) {
			parent := nativeGaugeEnvironment319(mode)
			want := *parent.Definition.EnvironmentSelector
			resource := map[string]string{"service.name": "api", "service.instance.id": "instance-1", identity.NamespaceKey: "team"}
			switch mode {
			case "named":
				resource[identity.LegacyEnvironmentKey] = "prod"
			case "empty":
				resource[identity.EnvironmentKey] = ""
				resource[identity.LegacyEnvironmentKey] = "prod"
			}
			binding := parent.Definition.Source
			stream := metrics.Stream{Resource: resource, Scope: binding.Scope, Version: binding.Version, SchemaURL: binding.SchemaURL, Name: binding.Metric, Unit: binding.Unit}
			before, _ := json.Marshal(stream)
			for _, signal := range []string{"logs", "spans"} {
				context, err := PlanNativeRecordContext(parent, stream, signal, "120000000123", "120000000124")
				if err != nil {
					t.Fatal(err)
				}
				if context.Grain != "point-context" || context.Question.Environment != "" || context.Question.EnvironmentSelector == nil || *context.Question.EnvironmentSelector != want || context.Question.FromMS != 120000 || context.Question.ToMS != 120001 {
					t.Fatal("record projection lost exact point or policy", context)
				}
				if _, err := Compile(context.Question); err != nil {
					t.Fatal(err)
				}
				context.Question.EnvironmentSelector.Policy = "changed"
				if *parent.Definition.EnvironmentSelector != want {
					t.Fatal("record context aliases native definition")
				}
			}
			after, _ := json.Marshal(stream)
			if !reflect.DeepEqual(before, after) {
				t.Fatal("record projection mutated exact raw stream identity")
			}
			bad := stream
			bad.Resource = map[string]string{}
			for key, value := range resource {
				bad.Resource[key] = value
			}
			bad.Resource[identity.EnvironmentKey] = "other"
			if _, err := PlanNativeRecordContext(parent, bad, "logs", "120000000123", "120000000124"); err == nil {
				t.Fatal("losing legacy label or absent-policy mismatch entered context")
			}
		})
	}
}
