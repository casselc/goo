package annotations

import (
	"fmt"
	"sort"
)

// Observation is a normalized completed span, not a metric point or business subject.
// Adapters must preserve scope provenance and must not truncate before reporting totals.
type Observation struct {
	Resource     string `json:"resource"`
	Scope        string `json:"scope"`
	ScopeVersion string `json:"scopeVersion"`
	SchemaURL    string `json:"schemaURL"`
	TraceID      string `json:"traceID"`
	SpanID       string `json:"spanID"`
	Name         string `json:"name"`
	Role         string `json:"role"`
	Outcome      string `json:"outcome"`
	DurationNS   uint64 `json:"durationNS"`
}
type OperationOverview struct {
	Resource string `json:"resource"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	Status   string `json:"status"`
	Calls    uint64 `json:"calls"`
	Errors   uint64 `json:"errors"`
	Panics   uint64 `json:"panics"`
	// Durations are kept as samples: no averaged percentiles or lossy numeric conversion.
	DurationsNS []uint64 `json:"durationsNS"`
}
type Conformance struct {
	APIVersion       string              `json:"apiVersion"`
	ManifestRevision string              `json:"manifestRevision"`
	Basis            string              `json:"basis"`
	Operations       []OperationOverview `json:"operations"`
	Warnings         []string            `json:"warnings"`
}

// Observe evaluates one explicitly selected resource scope. Missing declarations
// remain unobserved; observed compatibility never asserts complete instrumentation.
func (m Manifest) Observe(resource, basis string, spans []Observation) (Conformance, error) {
	result := Conformance{APIVersion: "goo/operation-observations/v1alpha1", ManifestRevision: m.Revision, Basis: basis,
		Operations: []OperationOverview{}, Warnings: []string{"Selected completed spans only; invocation counts are not business-subject counts.", "Sampling, source coverage, active invocations and completeness are unknown; no rates or SLOs are inferred."}}
	if resource == "" || basis == "" || m.APIVersion != "goo/instrumentation-manifest/v1alpha1" || m.Revision == "" {
		return result, fmt.Errorf("require a versioned manifest, resource and evidence basis")
	}
	byKey := map[string]*OperationOverview{}
	key := func(name, role string) string { return name + "\x00" + role }
	for _, d := range m.Declarations {
		if d.Kind == "operation" {
			k := key(d.Name, d.Role)
			if byKey[k] == nil {
				byKey[k] = &OperationOverview{Resource: resource, Name: d.Name, Role: d.Role, Status: "unobserved", DurationsNS: []uint64{}}
			}
		}
	}
	seen := map[string]Observation{}
	for _, s := range spans {
		if s.Resource != resource {
			return result, fmt.Errorf("observation crosses selected resource scope")
		}
		if s.Scope != m.Scope || s.ScopeVersion != m.ScopeVersion || s.SchemaURL != m.SchemaURL {
			return result, fmt.Errorf("observation scope/version/schema mismatch")
		}
		if !hexID(s.TraceID, 32) || !hexID(s.SpanID, 16) {
			return result, fmt.Errorf("invalid trace/span identity")
		}
		id := s.TraceID + "/" + s.SpanID
		if old, ok := seen[id]; ok {
			if old != s {
				return result, fmt.Errorf("conflicting span identity %s", id)
			}
			continue
		}
		seen[id] = s
		row := byKey[key(s.Name, s.Role)]
		if row == nil {
			return result, fmt.Errorf("undeclared operation or role: %s/%s", s.Name, s.Role)
		}
		switch s.Outcome {
		case "ok":
		case "error":
			row.Errors++
		case "panic":
			row.Panics++
		default:
			return result, fmt.Errorf("unsupported operation outcome %q", s.Outcome)
		}
		row.Calls++
		row.Status = "observed-compatible"
		row.DurationsNS = append(row.DurationsNS, s.DurationNS)
	}
	keys := make([]string, 0, len(byKey))
	for k := range byKey {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		row := byKey[k]
		sort.Slice(row.DurationsNS, func(i, j int) bool { return row.DurationsNS[i] < row.DurationsNS[j] })
		result.Operations = append(result.Operations, *row)
	}
	return result, nil
}
func hexID(s string, n int) bool {
	if len(s) != n {
		return false
	}
	nonzero := false
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
		nonzero = nonzero || c != '0'
	}
	return nonzero
}
