package measurements

import (
	"encoding/json"
	"fmt"
	"maps"

	"github.com/casselc/goo/metrics"
)

// An exact stream is an explicit package binding, not a presentation filter.
// Identity excludes temporality: epoch/temporality changes must still reach the
// native evaluator rather than silently disappearing from source evidence.
func validateExactStream(d Definition) error {
	s := d.Source.ExactStream
	if s == nil {
		return nil
	}
	b := d.Source
	encoded, _ := json.Marshal(s)
	if len(encoded) > 16384 || len(s.Resource) > 64 || len(s.Attributes) > 64 || !text(s.ResourceSchemaURL, 1024) {
		return fmt.Errorf("exact stream identity exceeds bounded package limits")
	}
	for _, values := range []map[string]string{s.Resource, s.Attributes} {
		for k, v := range values {
			if k == "" || !text(k, 256) || !text(v, 2048) {
				return fmt.Errorf("invalid exact stream attribute")
			}
		}
	}
	serviceOK := false
	for _, name := range b.Services {
		serviceOK = serviceOK || s.Resource["service.name"] == name
	}
	identityOK := serviceOK && s.Resource["service.instance.id"] != ""
	if r := b.Resource; r != nil {
		identityOK = r.AttributeKey() != "" && r.ID != "" && s.Resource[r.AttributeKey()] == r.ID
	}
	if !identityOK || s.Resource["deployment.environment.name"] != d.Environment || !b.Namespace.Matches(s.Resource) || s.Scope != b.Scope || s.Version != b.Version || s.SchemaURL != b.SchemaURL || s.Name != b.Metric || s.Unit != b.Unit {
		return fmt.Errorf("exact stream conflicts with metric source binding")
	}
	for k, v := range b.Attributes {
		actual, ok := s.Attributes[k]
		if !ok || actual != v {
			return fmt.Errorf("exact stream conflicts with attribute selector")
		}
	}
	return nil
}

func sameExactStream(a, b metrics.Stream) bool {
	return a.ResourceSchemaURL == b.ResourceSchemaURL && maps.Equal(a.Resource, b.Resource) && a.Scope == b.Scope && a.Version == b.Version && a.SchemaURL == b.SchemaURL && a.Name == b.Name && a.Unit == b.Unit && maps.Equal(a.Attributes, b.Attributes)
}
