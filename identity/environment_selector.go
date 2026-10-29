package identity

import (
	"encoding/json"
	"fmt"
	"strconv"
	"unicode/utf8"
)

const EnvironmentSelectorVersion = "goo/environment-selector/v1"

// EnvironmentSelector is a versioned wire value, separate from the immutable
// checked EnvironmentSelection. It does not extend existing Query/Tail inputs.
type EnvironmentSelector struct {
	Version string           `json:"version" yaml:"version"`
	Policy  string           `json:"policy" yaml:"policy"`
	Scope   EnvironmentScope `json:"scope" yaml:"scope"`
}

func (w EnvironmentSelector) Selection() (EnvironmentSelection, error) {
	if w.Version != EnvironmentSelectorVersion {
		return EnvironmentSelection{}, fmt.Errorf("unsupported environment selector version")
	}
	switch w.Policy {
	case "current-only":
		if w.Scope.Mode != "named" {
			return EnvironmentSelection{}, fmt.Errorf("current-only selection requires named scope")
		}
		return CurrentEnvironment(w.Scope.Value)
	case "effective":
		return EffectiveEnvironment(w.Scope)
	default:
		return EnvironmentSelection{}, fmt.Errorf("explicit environment selection policy required")
	}
}
func (w EnvironmentSelector) Validate() error { _, err := w.Selection(); return err }

func (s EnvironmentSelection) Selector() (EnvironmentSelector, error) {
	if err := s.Validate(); err != nil {
		return EnvironmentSelector{}, err
	}
	policy := "effective"
	if s.CurrentOnly() {
		policy = "current-only"
	}
	return EnvironmentSelector{Version: EnvironmentSelectorVersion, Policy: policy, Scope: s.Scope()}, nil
}

// UnmarshalJSON refuses unknown fields and explicit nulls instead of allowing
// them to disappear into Go string zero values. Semantic selection stays checked.
func (w *EnvironmentSelector) UnmarshalJSON(data []byte) error {
	*w = EnvironmentSelector{}
	if !wellFormedSelectorJSON(data) {
		return fmt.Errorf("environment selector JSON contains malformed Unicode")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for key := range fields {
		if key != "version" && key != "policy" && key != "scope" {
			return fmt.Errorf("unknown environment selector field %q", key)
		}
	}
	var scopeFields map[string]json.RawMessage
	if err := json.Unmarshal(fields["scope"], &scopeFields); err != nil {
		return err
	}
	for key := range scopeFields {
		if key != "mode" && key != "value" {
			return fmt.Errorf("unknown environment scope field %q", key)
		}
	}
	// Decode from the parsed maps, so duplicate members have JSON.parse's
	// last-value semantics instead of encoding/json's struct-object merging.
	readString := func(fields map[string]json.RawMessage, key string) (string, error) {
		var value *string
		if err := json.Unmarshal(fields[key], &value); err != nil {
			return "", err
		}
		if value == nil {
			return "", fmt.Errorf("environment selector %s must be a string", key)
		}
		return *value, nil
	}
	version, err := readString(fields, "version")
	if err != nil {
		return err
	}
	policy, err := readString(fields, "policy")
	if err != nil {
		return err
	}
	mode, err := readString(scopeFields, "mode")
	if err != nil {
		return err
	}
	name := ""
	if _, present := scopeFields["value"]; present {
		name, err = readString(scopeFields, "value")
		if err != nil {
			return err
		}
	}
	candidate := EnvironmentSelector{Version: version, Policy: policy, Scope: EnvironmentScope{Mode: mode, Value: name}}
	if err := candidate.Validate(); err != nil {
		return err
	}
	*w = candidate
	return nil
}

// encoding/json otherwise replaces invalid UTF-8 and unpaired surrogate escapes
// before field validation. Reject those lossy representations, preserving valid
// surrogate pairs and literal U+FFFD. JSON grammar is checked separately above.
func wellFormedSelectorJSON(data []byte) bool {
	if !utf8.Valid(data) {
		return false
	}
	quoted := false
	for i := 0; i < len(data); i++ {
		if data[i] == '"' {
			quoted = !quoted
			continue
		}
		if !quoted || data[i] != '\\' {
			continue
		}
		if i+1 >= len(data) {
			return false
		}
		if data[i+1] != 'u' {
			i++
			continue
		}
		if i+6 > len(data) {
			return false
		}
		code, err := strconv.ParseUint(string(data[i+2:i+6]), 16, 16)
		if err != nil {
			return false
		}
		if code >= 0xdc00 && code <= 0xdfff {
			return false
		}
		if code >= 0xd800 && code <= 0xdbff {
			if i+12 > len(data) || data[i+6] != '\\' || data[i+7] != 'u' {
				return false
			}
			low, err := strconv.ParseUint(string(data[i+8:i+12]), 16, 16)
			if err != nil || low < 0xdc00 || low > 0xdfff {
				return false
			}
			i += 11
		} else {
			i += 5
		}
	}
	return true
}
