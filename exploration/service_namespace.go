package exploration

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/identity"
)

func namespaceColumn(column string) string {
	return "if(mapContains(" + column + ",'service.namespace'),toNullable(" + column + "['service.namespace']),NULL)"
}
func namespaceFilter(n *identity.Namespace, params map[string]string) string {
	if n == nil {
		return ""
	}
	if n.Mode == "missing" {
		return " AND NOT mapContains(ResourceAttributes,'service.namespace')"
	}
	params["namespace"] = n.Value
	return " AND mapContains(ResourceAttributes,'service.namespace') AND ResourceAttributes['service.namespace']={namespace:String}"
}
func decodeNamespace(raw json.RawMessage) (*string, error) {
	var n *string
	if len(raw) == 0 {
		return nil, nil
	}
	if json.Unmarshal(raw, &n) != nil {
		return nil, fmt.Errorf("invalid namespace")
	}
	if n != nil {
		if e := identity.FromValue(n).Validate(); e != nil {
			return nil, e
		}
	}
	return n, nil
}
func namespaceKey(n *string) string { b, _ := json.Marshal(n); return string(b) }
