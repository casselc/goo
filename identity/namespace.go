// Package identity describes logical OTel service selectors without storage bindings.
package identity

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const NamespaceKey = "service.namespace"

type Namespace struct {
	Mode  string `json:"mode" yaml:"mode"`
	Value string `json:"value,omitempty" yaml:"value,omitempty"`
}

func (n *Namespace) Validate() error {
	if n == nil {
		return nil
	}
	if n.Mode != "missing" && n.Mode != "value" || n.Mode == "missing" && n.Value != "" || !utf8.ValidString(n.Value) || len(n.Value) > 128 || strings.ContainsRune(n.Value, 0) {
		return fmt.Errorf("namespace requires missing or value mode with bounded literal value")
	}
	return nil
}
func (n *Namespace) Matches(resource map[string]string) bool {
	if n == nil {
		return true
	}
	v, ok := resource[NamespaceKey]
	if n.Mode == "missing" {
		return !ok
	}
	return ok && v == n.Value
}
func (n *Namespace) Equal(other *Namespace) bool {
	if n == nil || other == nil {
		return n == nil && other == nil
	}
	return *n == *other
}
func (n *Namespace) Clone() *Namespace {
	if n == nil {
		return nil
	}
	v := *n
	return &v
}
func FromValue(value *string) *Namespace {
	if value == nil {
		return &Namespace{Mode: "missing"}
	}
	return &Namespace{Mode: "value", Value: *value}
}
