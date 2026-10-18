package investigations

import (
	"encoding/json"
	"fmt"
	"regexp"
)

var decimal = regexp.MustCompile(`^[0-9]{1,20}$`)
var traceID = regexp.MustCompile(`^[a-fA-F0-9]{32}$`)

// ValidateSelection constrains the persisted navigation protocol, not query semantics.
func ValidateSelection(raw json.RawMessage) error {
	if len(raw) == 0 {
		return nil
	}
	if len(raw) > 8192 {
		return fmt.Errorf("selection exceeds 8 KiB")
	}
	var value map[string]json.RawMessage
	if json.Unmarshal(raw, &value) != nil || value == nil {
		return fmt.Errorf("selection must be an object")
	}
	for k, v := range value {
		switch k {
		case "filter":
			var filter map[string]string
			if json.Unmarshal(v, &filter) != nil || filter == nil {
				return fmt.Errorf("filter must contain strings")
			}
			for name, s := range filter {
				switch name {
				case "fromNS", "toNS", "minNS", "maxNS":
					if !decimal.MatchString(s) {
						return fmt.Errorf("invalid selection timestamp or duration")
					}
				case "operation", "role", "outcome":
					if len(s) > 256 {
						return fmt.Errorf("selection value too long")
					}
				case "traceID":
					if !traceID.MatchString(s) {
						return fmt.Errorf("invalid trace ID")
					}
				default:
					return fmt.Errorf("unsupported selection filter")
				}
			}
		case "ratio", "comparison":
			var s string
			if json.Unmarshal(v, &s) != nil || len(s) > 4096 {
				return fmt.Errorf("invalid panel selection")
			}

		case "traceID":
			var s string
			if json.Unmarshal(v, &s) != nil || !traceID.MatchString(s) {
				return fmt.Errorf("invalid trace ID")
			}
		case "subject":
			var subject map[string]string
			if json.Unmarshal(v, &subject) != nil || len(subject) != 5 {
				return fmt.Errorf("require scoped subject")
			}
			for _, name := range []string{"environment", "service", "namespace", "tenant", "id"} {
				if subject[name] == "" || len(subject[name]) > 1024 {
					return fmt.Errorf("invalid scoped subject")
				}
			}
		default:
			return fmt.Errorf("unsupported selection field")
		}
	}
	return nil
}
