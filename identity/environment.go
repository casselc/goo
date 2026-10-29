package identity

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const EnvironmentKey = "deployment.environment.name"
const LegacyEnvironmentKey = "deployment.environment"

// EnvironmentScope is explicit effective-label selection. Its zero value is
// invalid; it never means all environments or changes legacy named questions.
type EnvironmentScope struct {
	Mode  string `json:"mode" yaml:"mode"`
	Value string `json:"value,omitempty" yaml:"value,omitempty"`
}

func (s EnvironmentScope) Validate() error {
	if s.Mode != "named" && s.Mode != "empty" && s.Mode != "unreported" || s.Mode == "named" && s.Value == "" || s.Mode != "named" && s.Value != "" || len(s.Value) > 128 || !utf8.ValidString(s.Value) || strings.ContainsRune(s.Value, 0) {
		return fmt.Errorf("environment scope requires explicit named, empty or unreported mode and a bounded named value")
	}
	return nil
}
func (s EnvironmentScope) Equal(other EnvironmentScope) bool {
	return s.Validate() == nil && other.Validate() == nil && s == other
}
func (s EnvironmentScope) Matches(resource map[string]string) bool {
	if s.Validate() != nil {
		return false
	}
	o := ObserveEnvironment(resource)
	return s.Mode == o.Scope.Mode && s.Value == o.Scope.Value
}

type EnvironmentObservation struct {
	Scope      EnvironmentScope `json:"scope"`
	Current    *string          `json:"current"`
	Legacy     *string          `json:"legacy"`
	WinningKey string           `json:"winningKey"`
	Conflict   bool             `json:"conflict"`
}

// Presence wins over truthiness: current empty masks legacy named. Neither
// attribute being present is distinct from a reported empty label.
func ObserveEnvironment(resource map[string]string) EnvironmentObservation {
	o := EnvironmentObservation{Scope: EnvironmentScope{Mode: "unreported"}}
	current, c := resource[EnvironmentKey]
	legacy, l := resource[LegacyEnvironmentKey]
	if c {
		o.Current = &current
	}
	if l {
		o.Legacy = &legacy
	}
	o.Conflict = c && l && current != legacy
	value := ""
	if c {
		value = current
		o.WinningKey = EnvironmentKey
	} else if l {
		value = legacy
		o.WinningKey = LegacyEnvironmentKey
	} else {
		return o
	}
	o.Scope = EnvironmentScope{Mode: "named", Value: value}
	if value == "" {
		o.Scope.Mode = "empty"
	}
	return o
}
