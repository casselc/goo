package identity

import "fmt"

const (
	currentEnvironmentPolicy uint8 = iota + 1
	effectiveEnvironmentPolicy
)

// EnvironmentSelection is an immutable, checked record-population selector.
// It deliberately has no wire representation. Existing named questions and
// effective-label adoption are different policies, even for the same name.
// Use the constructors; the zero value is invalid and never matches records.
type EnvironmentSelection struct {
	policy uint8
	scope  EnvironmentScope
}

// CurrentEnvironment retains the current-key-only rule of existing questions.
func CurrentEnvironment(name string) (EnvironmentSelection, error) {
	scope := EnvironmentScope{Mode: "named", Value: name}
	if err := scope.Validate(); err != nil {
		return EnvironmentSelection{}, err
	}
	return EnvironmentSelection{policy: currentEnvironmentPolicy, scope: scope}, nil
}

// EffectiveEnvironment uses current-key presence, then legacy fallback, and
// distinguishes missing labels from present-empty values.
func EffectiveEnvironment(scope EnvironmentScope) (EnvironmentSelection, error) {
	if err := scope.Validate(); err != nil {
		return EnvironmentSelection{}, err
	}
	return EnvironmentSelection{policy: effectiveEnvironmentPolicy, scope: scope}, nil
}

func (s EnvironmentSelection) Validate() error {
	if s.policy != currentEnvironmentPolicy && s.policy != effectiveEnvironmentPolicy || s.policy == currentEnvironmentPolicy && s.scope.Mode != "named" {
		return fmt.Errorf("explicit current-only or effective environment selection required")
	}
	return s.scope.Validate()
}

// Scope returns a copy; changing it cannot change this selection.
func (s EnvironmentSelection) Scope() EnvironmentScope { return s.scope }

// CurrentOnly reports the policy, not whether the current label was observed.
func (s EnvironmentSelection) CurrentOnly() bool { return s.policy == currentEnvironmentPolicy }

func (s EnvironmentSelection) Equal(other EnvironmentSelection) bool {
	return s.Validate() == nil && other.Validate() == nil && s == other
}

func (s EnvironmentSelection) Matches(resource map[string]string) bool {
	if s.Validate() != nil {
		return false
	}
	if s.CurrentOnly() {
		value, present := resource[EnvironmentKey]
		return present && value == s.scope.Value
	}
	return s.scope.Matches(resource)
}

// Refuse accidental JSON/hash adoption. Private semantic fields must never
// silently serialize as {}, erasing the selection policy from an identity.
func (EnvironmentSelection) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("environment selection has no JSON format; use the owning versioned wire type")
}

func (*EnvironmentSelection) UnmarshalJSON([]byte) error {
	return fmt.Errorf("environment selection has no JSON format; use the owning versioned wire type")
}

// Semantic values have no YAML representation either. Use EnvironmentSelector
// at an explicitly versioned wire boundary instead of losing private policy.
func (EnvironmentSelection) MarshalYAML() (any, error) {
	return nil, fmt.Errorf("environment selection has no YAML format; use the owning versioned wire type")
}
func (*EnvironmentSelection) UnmarshalYAML(func(any) error) error {
	return fmt.Errorf("environment selection has no YAML format; use the owning versioned wire type")
}
