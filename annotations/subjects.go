package annotations

import (
	"fmt"
	"strconv"
)

// ScopedSubject is an explicit local business identity, distinct from trace IDs.
// Instance is intentionally absent: one subject can have invocations on multiple
// instances, though the current source query selects only one instance.
type ScopedSubject struct {
	Environment string `json:"environment"`
	Service     string `json:"service"`
	Namespace   string `json:"namespace"`
	Tenant      string `json:"tenant"`
	ID          string `json:"id"`
}
type SubjectEvent struct {
	TimeNS     string         `json:"timeNS"`
	Name       string         `json:"name"`
	Outcome    string         `json:"outcome,omitempty"`
	Subject    *ScopedSubject `json:"subject,omitempty"`
	Diagnostic string         `json:"diagnostic,omitempty"`
}

// SubjectEvents normalizes declared span events. Unbound/malformed identity
// disables that relation, rather than turning an operation into business failure.
func (m Manifest) SubjectEvents(source OperationSource, row OperationRow) ([]SubjectEvent, error) {
	result := []SubjectEvent{}
	if source.SubjectNamespace != "" && (source.Service == "" || source.Environment == "") {
		return nil, fmt.Errorf("subject binding requires service and environment")
	}
	if len(row.EventTimes) != len(row.EventNames) || len(row.EventNames) != len(row.EventAttributes) {
		return nil, fmt.Errorf("mismatched span event arrays")
	}
	declared := map[string]bool{}
	for _, d := range m.Declarations {
		if d.Kind == "event" {
			declared[d.Name] = true
		}
	}
	start, err := strconv.ParseInt(row.TimestampNS, 10, 64)
	if err != nil || start < 0 {
		return nil, fmt.Errorf("invalid event parent timestamp")
	}
	duration, err := strconv.ParseUint(row.DurationNS, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid event parent duration")
	}
	for i, name := range row.EventNames {
		ns, err := strconv.ParseInt(row.EventTimes[i], 10, 64)
		if err != nil || ns < start || uint64(ns-start) > duration {
			return nil, fmt.Errorf("event timestamp outside parent span extent")
		}
		a := row.EventAttributes[i]
		event := SubjectEvent{TimeNS: row.EventTimes[i], Name: name, Outcome: a["goo.subject.outcome"]}
		switch {
		case !declared[name] || a["goo.event.name"] != name:
			event.Diagnostic = "not-a-declared-subject-event"
		case source.SubjectNamespace == "":
			event.Diagnostic = "subject-binding-required"
		case a["goo.subject.namespace"] != source.SubjectNamespace:
			event.Diagnostic = "missing-or-mismatched-subject-namespace"
		case a["goo.subject.tenant"] == "" || len(a["goo.subject.tenant"]) > 128:
			event.Diagnostic = "invalid-subject-tenant"
		case a["goo.subject.id"] == "" || len(a["goo.subject.id"]) > 256 || len(event.Outcome) > 128:
			event.Diagnostic = "invalid-subject-identity-or-outcome"
		default:
			event.Subject = &ScopedSubject{Environment: source.Environment, Service: source.Service, Namespace: source.SubjectNamespace, Tenant: a["goo.subject.tenant"], ID: a["goo.subject.id"]}
		}
		result = append(result, event)
	}
	return result, nil
}
