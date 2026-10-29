package exploration

import "github.com/casselc/goo/identity"

// environmentSelectionPredicate is the single record environment lowering
// boundary. Validate before reading policy/scope: a zero selector must not turn
// into a broad read. Keep current-only SQL byte-stable for existing plans.
func environmentSelectionPredicate(selection identity.EnvironmentSelection) (string, map[string]string, error) {
	if err := selection.Validate(); err != nil {
		return "", nil, err
	}
	scope := selection.Scope()
	if selection.CurrentOnly() {
		return "mapContains(ResourceAttributes, 'deployment.environment.name') AND ResourceAttributes['deployment.environment.name'] = {environment:String}", map[string]string{"environment": scope.Value}, nil
	}
	current := "mapContains(ResourceAttributes,'" + identity.EnvironmentKey + "')"
	legacy := "mapContains(ResourceAttributes,'" + identity.LegacyEnvironmentKey + "')"
	params := map[string]string{}
	if scope.Mode == "unreported" {
		return "(NOT " + current + " AND NOT " + legacy + ")", params, nil
	}
	value := "''"
	if scope.Mode == "named" {
		value = "{environment:String}"
		params["environment"] = scope.Value
	}
	return "((" + current + " AND ResourceAttributes['" + identity.EnvironmentKey + "']=" + value + ") OR (NOT " + current + " AND " + legacy + " AND ResourceAttributes['" + identity.LegacyEnvironmentKey + "']=" + value + "))", params, nil
}
