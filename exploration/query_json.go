package exploration

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"reflect"

	"github.com/casselc/goo/identity"
)

// UnmarshalJSON keeps record intent literal across HTTP, saved configuration and
// browser links. Known field names are case-sensitive and explicit null values
// cannot disappear into omitted defaults. Missing fields retain ordinary Go zero
// values; Validate remains responsible for the query's execution semantics.
func (q *Query) UnmarshalJSON(data []byte) error {
	*q = Query{}
	if !identity.WellFormedJSON(data) {
		return errors.New("record question contains malformed Unicode")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	if decoder.Decode(new(any)) != io.EOF {
		return errors.New("expected one record question")
	}

	type wire Query
	// The existing trace-search shape guard checks struct names, nested arms and
	// nulls without imposing trace-search semantics on this ordinary query.
	if !strictTraceSearchValue(value, reflect.TypeFor[wire](), 0) {
		return errors.New("record question requires known non-null fields")
	}
	if where, present := value.(map[string]any)["where"]; present && !strictRecordPredicateShape(where) {
		return errors.New("record predicate requires one closed branch or leaf")
	}
	if edge, present := value.(map[string]any)["reportedEdge"]; present {
		for _, name := range []string{"parent", "child"} {
			side, ok := edge.(map[string]any)[name].(map[string]any)
			if !ok {
				return errors.New("record edge requires explicit qualified endpoints")
			}
			if _, ok := side["service"]; !ok {
				return errors.New("record edge requires an explicit service identity")
			}
		}
	}
	// Decode the parsed value so duplicate object members have JSON.parse's
	// last-value semantics instead of encoding/json's struct-object merging.
	canonical, err := json.Marshal(value)
	if err != nil {
		return err
	}
	var out wire
	if err := json.Unmarshal(canonical, &out); err != nil {
		return err
	}
	*q = Query(out)
	return nil
}

// strictRecordPredicateShape preserves the predicate union before Go zero values
// can erase an explicitly supplied empty op/value or merge two different arms.
// The preceding typed shape guard already bounds recursion and rejects nulls.
func strictRecordPredicateShape(value any) bool {
	object, ok := value.(map[string]any)
	if !ok {
		return false
	}
	all, hasAll := object["all"]
	anyChildren, hasAny := object["any"]
	if hasAll || hasAny {
		if len(object) != 1 {
			return false
		}
		children := all
		if hasAny {
			children = anyChildren
		}
		list, ok := children.([]any)
		if !ok {
			return false
		}
		for _, child := range list {
			if !strictRecordPredicateShape(child) {
				return false
			}
		}
		return true
	}
	_, hasField := object["field"]
	_, hasNumeric := object["numeric"]
	if hasField == hasNumeric {
		return false
	}
	for key := range object {
		if key != "field" && key != "numeric" && key != "op" && key != "value" {
			return false
		}
	}
	return true
}
