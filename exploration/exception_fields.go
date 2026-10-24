package exploration

import "fmt"

// ExceptionEventPresent concerns a reported event name, not span failure, health,
// uncaught behavior, event uniqueness or completeness of exception telemetry.
func ExceptionEventPresent(names []string) bool {
	for _, name := range names {
		if name == "exception" {
			return true
		}
	}
	return false
}

// ReferenceExceptionFields projects one span record, never one record per event.
// A type is present only when exactly one distinct value was reported on matching
// exception events. Present empty and absent remain distinct; ambiguity is null.
func ReferenceExceptionFields(names []string, attributes []map[string]string) (map[Field]string, error) {
	if len(names) != len(attributes) {
		return nil, fmt.Errorf("exception event names and attributes must align")
	}
	event := "no"
	if ExceptionEventPresent(names) {
		event = "yes"
	}
	fields := map[Field]string{{Kind: "column", Key: "exception_event"}: event}
	types := map[string]bool{}
	for i, name := range names {
		if name == "exception" {
			if value, ok := attributes[i]["exception.type"]; ok {
				types[value] = true
			}
		}
	}
	state := "unreported"
	if len(types) > 1 {
		state = "multiple"
	} else if len(types) == 1 {
		state = "single"
		for value := range types {
			fields[Field{Kind: "column", Key: "exception_type"}] = value
			if value == "" {
				state = "empty"
			}
		}
	}
	fields[Field{Kind: "column", Key: "exception_type_state"}] = state
	return fields, nil
}

func exceptionTypesSQL() string {
	// arrayZip deliberately refuses malformed source-array alignment. Spark's
	// padding zip is not interchangeable without an explicit alignment contract.
	return "arrayDistinct(arrayMap(e -> e.2['exception.type'], arrayFilter(e -> e.1 = 'exception' AND mapContains(e.2, 'exception.type'), arrayZip(`Events.Name`, `Events.Attributes`))))"
}
func exceptionFieldSQL(key string) (value, present string, ok bool) {
	switch key {
	case "exception_event":
		return "if(has(`Events.Name`, 'exception'), 'yes', 'no')", "1", true
	case "exception_type":
		types := exceptionTypesSQL()
		return "arrayElement(" + types + ", 1)", "length(" + types + ") = 1", true
	case "exception_type_state":
		types := exceptionTypesSQL()
		return "multiIf(length(" + types + ") = 0, 'unreported', length(" + types + ") > 1, 'multiple', arrayElement(" + types + ", 1) = '', 'empty', 'single')", "1", true
	}
	return "", "", false
}

// CanBeMissing identifies map absence or the single nullable derived column.
// Every other fixed column is present in a valid bound source record.
func (f Field) CanBeMissing() bool {
	return f.Kind == "attribute" || f.Kind == "resource" || f.Kind == "column" && f.Key == "exception_type"
}
func validateGroupDimension(f Field, value *string) error {
	if value == nil && !f.CanBeMissing() {
		return fmt.Errorf("fixed column %s cannot be missing", f.Key)
	}
	return nil
}

func queryReferencesColumn(q Query, key string) bool {
	for _, field := range q.GroupBy {
		if field == (Field{Kind: "column", Key: key}) {
			return true
		}
	}
	var walk func(*Predicate) bool
	walk = func(p *Predicate) bool {
		if p == nil {
			return false
		}
		if p.Field != nil && *p.Field == (Field{Kind: "column", Key: key}) {
			return true
		}
		for i := range p.All {
			if walk(&p.All[i]) {
				return true
			}
		}
		for i := range p.Any {
			if walk(&p.Any[i]) {
				return true
			}
		}
		return false
	}
	return walk(q.Where)
}
func (c *compiler) requestedDerivedColumns(q Query) string {
	out := ""
	if q.Signal == "spans" {
		for _, key := range []string{"exception_event", "exception_type", "exception_type_state"} {
			if queryReferencesColumn(q, key) {
				value, present, _ := exceptionFieldSQL(key)
				if key == "exception_type" {
					value = "if((" + present + "), CAST(" + value + " AS Nullable(String)), NULL)"
				}
				out += ", " + value + " AS " + key
			}
		}
	}
	if q.Signal == "logs" && queryReferencesColumn(q, "severity_number") {
		out += ", toString(toUInt64(SeverityNumber)) AS severity_number"
	}
	return out
}
