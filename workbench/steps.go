package workbench

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
)

// TraceSearchIntentMaxBytes matches the existing trace-search request boundary.
// A saved request must remain searchable without truncation or scope rebinding.
const TraceSearchIntentMaxBytes = exploration.TraceSearchRequestMaxBytes

// Step is authored intent, never a result, receipt or inferred shared population.
type StepScope struct {
	Environment         string                        `json:"environment"`
	EnvironmentSelector *identity.EnvironmentSelector `json:"environmentSelector,omitempty"`
	FromMS              int64                         `json:"fromMs"`
	ToMS                int64                         `json:"toMs"`
	FromNS              string                        `json:"fromNS,omitempty"`
	ToNS                string                        `json:"toNS,omitempty"`
}
type Step struct {
	Version     int                             `json:"version"`
	Kind        string                          `json:"kind"`
	Scope       StepScope                       `json:"scope"`
	View        string                          `json:"view"`
	Record      *exploration.Query              `json:"record,omitempty"`
	Measurement *measurements.Definition        `json:"measurement,omitempty"`
	TraceSearch *exploration.TraceSearchRequest `json:"traceSearch,omitempty"`
}

// UnmarshalJSON keeps step payloads closed and refuses explicit null arms.
// Search intent is a checked request, never an arbitrary result/receipt blob.
func (s *Step) UnmarshalJSON(data []byte) error {
	*s = Step{}
	if !identity.WellFormedJSON(data) {
		return errors.New("question step JSON contains malformed Unicode")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	if fields == nil {
		return errors.New("question step requires an object")
	}
	for key, raw := range fields {
		switch key {
		case "version", "kind", "scope", "view", "record", "measurement", "traceSearch":
		default:
			return fmt.Errorf("unknown question step field %q", key)
		}
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return fmt.Errorf("question step %s must be non-null", key)
		}
	}
	var kind string
	if err := json.Unmarshal(fields["kind"], &kind); err != nil {
		return err
	}
	if kind == "trace-search" {
		for _, key := range []string{"version", "kind", "scope", "view", "traceSearch"} {
			if _, present := fields[key]; !present {
				return fmt.Errorf("trace-search step requires %s", key)
			}
		}
		for _, nested := range []struct {
			key      string
			required []string
		}{
			{"scope", []string{"fromMs", "toMs"}},
			{"traceSearch", []string{"version", "fromMs", "toMs", "conditions", "limit"}},
		} {
			var object map[string]json.RawMessage
			if err := json.Unmarshal(fields[nested.key], &object); err != nil {
				return err
			}
			for _, key := range nested.required {
				if _, present := object[key]; !present {
					return fmt.Errorf("trace-search step %s requires %s", nested.key, key)
				}
			}
		}
		// This new intent arm has no meaningful null values. Omitted optionals are
		// distinct from nulls that would silently acquire Go scalar defaults.
		for _, key := range []string{"scope", "traceSearch"} {
			normalized, err := checkedTraceSearchJSON(fields[key])
			if err != nil {
				return err
			}
			fields[key] = normalized
		}
		var scope map[string]json.RawMessage
		if err := json.Unmarshal(fields["scope"], &scope); err != nil {
			return err
		}
		if _, ok := scope["fromNS"]; ok {
			return errors.New("trace-search step scope does not support native nanosecond bounds")
		}
		if _, ok := scope["toNS"]; ok {
			return errors.New("trace-search step scope does not support native nanosecond bounds")
		}
	}
	if kind == "record" {
		var scope map[string]json.RawMessage
		if err := json.Unmarshal(fields["scope"], &scope); err != nil {
			return err
		}
		if _, ok := scope["fromNS"]; ok {
			return errors.New("record step scope does not support native nanosecond bounds")
		}
		if _, ok := scope["toNS"]; ok {
			return errors.New("record step scope does not support native nanosecond bounds")
		}
	}
	// These selector/scope fields cannot disappear into Go scalar zero values.
	for _, nested := range []struct {
		key     string
		allowed []string
	}{
		{"scope", []string{"environment", "environmentSelector", "fromMs", "toMs", "fromNS", "toNS"}},
		{"traceSearch", []string{"version", "environment", "environmentSelector", "fromMs", "toMs", "conditions", "limit"}},
	} {
		if raw, present := fields[nested.key]; present {
			var object map[string]json.RawMessage
			if err := json.Unmarshal(raw, &object); err != nil {
				return err
			}
			if object == nil {
				return fmt.Errorf("question step %s requires an object", nested.key)
			}
			for key, value := range object {
				allowed := false
				for _, name := range nested.allowed {
					if key == name {
						allowed = true
						break
					}
				}
				if !allowed {
					return fmt.Errorf("unknown question step %s field %q", nested.key, key)
				}
				if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
					return fmt.Errorf("question step %s.%s must be non-null", nested.key, key)
				}
			}
		}
	}
	normalized := data
	if kind == "trace-search" {
		// New search wire follows JSON.parse last-value semantics rather than
		// merging struct-object fields from earlier duplicate occurrences.
		var err error
		normalized, err = json.Marshal(fields)
		if err != nil {
			return err
		}
	}
	type wire Step
	var value wire
	decoder := json.NewDecoder(bytes.NewReader(normalized))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	*s = Step(value)
	return nil
}

func checkedTraceSearchJSON(raw json.RawMessage) (json.RawMessage, error) {
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	var visit func(any, int) bool
	visit = func(value any, depth int) bool {
		if depth > 32 {
			return false
		}
		switch value := value.(type) {
		case nil:
			return false
		case []any:
			for _, item := range value {
				if !visit(item, depth+1) {
					return false
				}
			}
		case map[string]any:
			// Search scope/request objects are all typed; user-selected field
			// names are string values, never arbitrary object keys. This closes
			// encoding/json's case-insensitive struct aliases as well as nulls.
			for key, item := range value {
				switch key {
				case "version", "environment", "environmentSelector", "scope", "policy", "mode", "fromMs", "toMs", "conditions", "limit", "all", "any", "field", "numeric", "op", "value", "kind", "key", "type", "unit", "parse", "scale":
				default:
					return false
				}
				if !visit(item, depth+1) {
					return false
				}
			}
		}
		return true
	}
	if !visit(value, 0) {
		return nil, errors.New("trace-search step scope and request require known non-null fields")
	}
	return json.Marshal(value)
}

func (s Step) Validate() error {
	if s.Version != 1 {
		return errors.New("unsupported question step version")
	}
	switch s.Kind {
	case "record":
		if s.Scope.FromNS != "" || s.Scope.ToNS != "" || s.Record == nil || s.Measurement != nil || s.TraceSearch != nil || (s.View != "Records" && s.View != "Distribution") {
			return errors.New("record step requires one record question and supported view")
		}
		q := s.Record
		scope := exploration.Query{Environment: s.Scope.Environment, EnvironmentSelector: s.Scope.EnvironmentSelector}
		if !q.SameEnvironment(scope) || q.FromMS != s.Scope.FromMS || q.ToMS != s.Scope.ToMS {
			return errors.New("record question differs from step scope")
		}
		return q.Validate()
	case "measurement":
		if s.Measurement == nil || s.Record != nil || s.TraceSearch != nil || s.View != "Overview" || !s.Measurement.SameEnvironment(measurements.Definition{Environment: s.Scope.Environment, EnvironmentSelector: s.Scope.EnvironmentSelector}) {
			return errors.New("measurement step requires one definition matching scope")
		}
		return (measurements.Question{Definition: *s.Measurement, FromMS: s.Scope.FromMS, ToMS: s.Scope.ToMS, FromNS: s.Scope.FromNS, ToNS: s.Scope.ToNS}).Validate()
	case "trace-search":
		if s.TraceSearch == nil || s.Record != nil || s.Measurement != nil || s.View != "Traces" || s.Scope.FromNS != "" || s.Scope.ToNS != "" {
			return errors.New("trace-search step requires one search request and supported view")
		}
		q := s.TraceSearch
		if !(exploration.Query{Environment: q.Environment, EnvironmentSelector: q.EnvironmentSelector}).SameEnvironment(exploration.Query{Environment: s.Scope.Environment, EnvironmentSelector: s.Scope.EnvironmentSelector}) || q.FromMS != s.Scope.FromMS || q.ToMS != s.Scope.ToMS {
			return errors.New("trace-search request differs from step scope")
		}
		if err := q.Validate(); err != nil {
			return err
		}
		encoded, err := json.Marshal(q)
		if err != nil {
			return err
		}
		if len(encoded) > TraceSearchIntentMaxBytes {
			return errors.New("trace-search intent exceeds the 64 KiB request limit")
		}
		return nil
	default:
		return errors.New("unsupported question step kind")
	}
}
