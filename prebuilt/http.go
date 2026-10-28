package prebuilt

import (
	"fmt"
	"github.com/casselc/goo/subjectfacts"
)

type HTTPRequest struct {
	Kind     string            `json:"kind"`
	Resource map[string]string `json:"resource"`
	TraceID  string            `json:"traceId"`
	SpanID   string            `json:"spanId"`
	Route    string            `json:"route"`
	Method   string            `json:"method"`
	Status   *int              `json:"status"`
}
type HTTPPrepared struct {
	ExpansionRevision string              `json:"expansionRevision"`
	EvaluationID      string              `json:"evaluationId"`
	InputDigest       string              `json:"inputDigest"`
	Evidence          string              `json:"evidence"`
	Facts             []subjectfacts.Fact `json:"facts"`
}

// PrepareHTTP accepts already selected normalized requests, not raw OTLP. Resource
// mismatches or duplicate identities are rejected; incomplete outcomes stay unknown.
//
//quint:action action:prepare_http when:ok actor:evaluation obs.selected:len(requests) obs.facts:len(result.Facts)
func (e *Expansion) PrepareHTTP(requests []HTTPRequest, evaluation string) (result *HTTPPrepared, err error) {
	if e == nil {
		return nil, fmt.Errorf("expansion is required")
	}
	check := *e
	check.Revision = ""
	if e.Revision == "" || digest(check) != e.Revision {
		return nil, fmt.Errorf("expansion content no longer matches its revision")
	}
	if evaluation == "" {
		return nil, fmt.Errorf("evaluation ID is required")
	}
	out := &HTTPPrepared{ExpansionRevision: e.Revision, EvaluationID: evaluation, InputDigest: digest(requests), Evidence: "unverified-input-selection-and-completeness", Facts: []subjectfacts.Fact{}}
	for i, r := range requests {
		if r.Kind != "server" {
			return nil, fmt.Errorf("request %d must be a normalized server span", i)
		}
		for key, value := range e.Resource {
			if r.Resource[key] != value {
				return nil, fmt.Errorf("request %d mismatches resource %s", i, key)
			}
		}
		if r.TraceID == "" || r.SpanID == "" {
			return nil, fmt.Errorf("request %d requires scoped trace/span identity", i)
		}
		route := r.Route
		if route == "" {
			route = "(unknown route)"
		} else if route == "(unknown route)" {
			return nil, fmt.Errorf("reserved unknown route label")
		}
		method := r.Method
		if method == "" {
			method = "(unknown method)"
		}
		class := "unknown"
		if r.Status != nil && *r.Status >= 200 && *r.Status <= 599 && r.Method != "" {
			class = "success"
			for _, band := range e.Binding.Policy.FailureStatusRanges {
				if *r.Status >= band.Min && *r.Status <= band.Max {
					class = "failure"
				}
			}
		}
		out.Facts = append(out.Facts, subjectfacts.Fact{"evaluation_id": evaluation, "scope_id": e.Binding.Scope, "trace_id": r.TraceID, "span_id": r.SpanID, "route": route, "method": method, "classification": class})
	}
	if _, err := e.Measurement.Evaluate(out.Facts, evaluation, nil); err != nil {
		return nil, err
	}
	return out, nil
}
