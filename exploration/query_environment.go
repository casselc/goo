package exploration

import (
	"fmt"
	"github.com/casselc/goo/identity"
)

// EnvironmentSelection preserves the historical current-key-only name or an
// explicitly versioned selector. Neither omission nor an empty name is a wildcard.
func (q Query) EnvironmentSelection() (identity.EnvironmentSelection, error) {
	if q.EnvironmentSelector == nil {
		selection, err := identity.CurrentEnvironment(q.Environment)
		if err != nil {
			return identity.EnvironmentSelection{}, fmt.Errorf("explicit environment required")
		}
		return selection, nil
	}
	if q.Environment != "" {
		return identity.EnvironmentSelection{}, fmt.Errorf("choose environment or environmentSelector, not both")
	}
	return q.EnvironmentSelector.Selection()
}
func (q Query) SameEnvironment(other Query) bool {
	a, err := q.EnvironmentSelection()
	if err != nil {
		return false
	}
	b, err := other.EnvironmentSelection()
	return err == nil && a.Equal(b)
}
func copyEnvironmentSelector(in *identity.EnvironmentSelector) *identity.EnvironmentSelector {
	if in == nil {
		return nil
	}
	value := *in
	return &value
}

// WithEnvironmentFrom detaches the complete selector when constructing a query.
// It does not validate the rest of either question or read telemetry.
func (q Query) WithEnvironmentFrom(origin Query) Query {
	q.Environment = origin.Environment
	q.EnvironmentSelector = copyEnvironmentSelector(origin.EnvironmentSelector)
	return q
}
