package measurements

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/casselc/goo/identity"
	"gopkg.in/yaml.v3"
)

func copyEnvironmentSelector(value *identity.EnvironmentSelector) *identity.EnvironmentSelector {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
func environmentSelection(name string, selector *identity.EnvironmentSelector) (identity.EnvironmentSelection, error) {
	if selector == nil {
		return identity.CurrentEnvironment(name)
	}
	if name != "" {
		return identity.EnvironmentSelection{}, fmt.Errorf("choose environment or environmentSelector, not both")
	}
	return selector.Selection()
}
func (d Definition) EnvironmentSelection() (identity.EnvironmentSelection, error) {
	return environmentSelection(d.Environment, d.EnvironmentSelector)
}
func (d Definition) SameEnvironment(other Definition) bool {
	a, e := d.EnvironmentSelection()
	if e != nil {
		return false
	}
	b, e := other.EnvironmentSelection()
	return e == nil && a.Equal(b)
}
func (d Definition) WithEnvironmentFrom(origin Definition) Definition {
	d.Environment = origin.Environment
	d.EnvironmentSelector = copyEnvironmentSelector(origin.EnvironmentSelector)
	return d
}
func (q DiscoveryQuestion) EnvironmentSelection() (identity.EnvironmentSelection, error) {
	return environmentSelection(q.Environment, q.EnvironmentSelector)
}
func (q DiscoveryQuestion) WithEnvironmentFrom(origin Definition) DiscoveryQuestion {
	q.Environment = origin.Environment
	q.EnvironmentSelector = copyEnvironmentSelector(origin.EnvironmentSelector)
	return q
}

func (q DiscoveryQuestion) clone() DiscoveryQuestion {
	q.EnvironmentSelector = copyEnvironmentSelector(q.EnvironmentSelector)
	q.Namespace = q.Namespace.Clone()
	return q
}
func (q Question) detached() (Question, error) {
	raw, e := json.Marshal(q)
	if e != nil {
		return Question{}, e
	}
	var copy Question
	e = json.Unmarshal(raw, &copy)
	return copy, e
}

const typedServiceNativeReason = "Typed environment native evaluation supports service-bound per-stream gauges, monotonic counters and explicit duration histograms. Stock, state, lifecycle and resource bindings remain unavailable."

func typedServiceKind(kind string) bool {
	return kind == "gauge" || kind == "counter" || kind == "histogram"
}

// Lower only a checked logical selection into the metric adapter. Full resource
// identity remains untouched; selecting an effective label never merges streams.
func metricEnvironmentPredicate(selection identity.EnvironmentSelection, param func(string, string) string) (string, error) {
	if e := selection.Validate(); e != nil {
		return "", e
	}
	scope := selection.Scope()
	current := "mapContains(ResourceAttributes,'deployment.environment.name')"
	legacy := "mapContains(ResourceAttributes,'deployment.environment')"
	if selection.CurrentOnly() {
		return current + " AND ResourceAttributes['deployment.environment.name']=" + param(scope.Value, "String"), nil
	}
	if scope.Mode == "unreported" {
		return "(NOT " + current + " AND NOT " + legacy + ")", nil
	}
	value := "''"
	if scope.Mode == "named" {
		value = param(scope.Value, "String")
	}
	return "((" + current + " AND ResourceAttributes['deployment.environment.name']=" + value + ") OR (NOT " + current + " AND " + legacy + " AND ResourceAttributes['deployment.environment']=" + value + "))", nil
}
func guardEnvironmentJSON(data []byte, allowed ...string) error {
	var fields map[string]json.RawMessage
	if e := json.Unmarshal(data, &fields); e != nil {
		return e
	}
	for key := range fields {
		known := false
		for _, name := range allowed {
			known = known || name == key
		}
		if !known {
			return fmt.Errorf("json: unknown field %q", key)
		}
	}
	for _, key := range []string{"environment", "environmentSelector"} {
		if raw, ok := fields[key]; ok && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return fmt.Errorf("%s must be omitted or non-null", key)
		}
	}
	return nil
}
func (d *Definition) UnmarshalJSON(data []byte) error {
	if e := guardEnvironmentJSON(data, "counterDomain", "stateProfile", "stockDomain", "histogramBucketMs", "apiVersion", "name", "title", "kind", "environment", "environmentSelector", "lookbackMinutes", "maxGapMs", "grain", "source", "lifecycle"); e != nil {
		return e
	}
	type wire Definition
	var value wire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if e := decoder.Decode(&value); e != nil {
		return e
	}
	*d = Definition(value)
	return nil
}
func (q *DiscoveryQuestion) UnmarshalJSON(data []byte) error {
	if e := guardEnvironmentJSON(data, "namespace", "environment", "environmentSelector", "service", "fromMs", "toMs"); e != nil {
		return e
	}
	type wire DiscoveryQuestion
	var value wire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if e := decoder.Decode(&value); e != nil {
		return e
	}
	*q = DiscoveryQuestion(value)
	return nil
}

func validateSelectorYAML(node *yaml.Node, scope bool) error {
	if node.Kind != yaml.MappingNode || node.Tag != "!!map" {
		return fmt.Errorf("environmentSelector requires a plain mapping")
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		key, value := node.Content[i], node.Content[i+1]
		if key.Kind != yaml.ScalarNode || key.Tag != "!!str" {
			return fmt.Errorf("environmentSelector requires string field names")
		}
		if !scope && key.Value == "scope" {
			if e := validateSelectorYAML(value, true); e != nil {
				return e
			}
		} else if value.Kind != yaml.ScalarNode || value.Tag != "!!str" {
			return fmt.Errorf("environmentSelector fields require literal strings")
		}
	}
	return nil
}
