package subjectfacts

import (
	"encoding/json"
	"fmt"
	"sort"
	"unicode/utf8"
)

type Fact map[string]any

type Value struct {
	State string   `json:"state"`
	Count *uint64  `json:"count,omitempty"`
	Ratio *float64 `json:"ratio,omitempty"`
}

type Result struct {
	Dimensions map[string]any   `json:"dimensions"`
	Measures   map[string]Value `json:"measures"`
}

func matches(f Fact, p *Predicate) bool {
	return p == nil || f[p.Field] == p.Eq
}

// Evaluate rejects malformed or duplicate prepared facts instead of silently
// producing a healthier denominator. It validates the selected evaluation before
// population filtering, so an invalid excluded row cannot be hidden by a filter.
func (d *Definition) Evaluate(facts []Fact, evaluationID string, groups []string) ([]Result, error) {
	if err := d.Validate(); err != nil {
		return nil, err
	}
	if evaluationID == "" || !utf8.ValidString(evaluationID) {
		return nil, fmt.Errorf("one nonempty UTF-8 evaluation ID is required")
	}
	if err := d.validateGroups(groups); err != nil {
		return nil, err
	}
	type bucket struct {
		dims   map[string]any
		counts map[string]uint64
	}
	buckets := map[string]*bucket{}
	if len(groups) == 0 {
		buckets["[]"] = &bucket{map[string]any{}, map[string]uint64{}}
	}
	seen := map[string]bool{}
	for i, f := range facts {
		id, ok := f["evaluation_id"].(string)
		if !ok || id == "" || !utf8.ValidString(id) {
			return nil, fmt.Errorf("fact %d has no usable UTF-8 evaluation_id", i)
		}
		if id != evaluationID {
			continue
		}
		for field, typ := range d.Source.Fields {
			v, exists := f[field]
			if !exists {
				return nil, fmt.Errorf("fact %d missing field %s", i, field)
			}
			switch typ {
			case "string":
				s, ok := v.(string)
				if !ok || !utf8.ValidString(s) {
					return nil, fmt.Errorf("fact %d field %s must be UTF-8 string", i, field)
				}
				if values, exists := d.Source.Enums[field]; exists && !contains(values, s) {
					return nil, fmt.Errorf("fact %d field %s has undeclared enum value", i, field)
				}
			case "bool":
				if _, ok := v.(bool); !ok {
					return nil, fmt.Errorf("fact %d field %s must be bool", i, field)
				}
			}
		}
		keyParts := []string{id}
		for _, field := range d.Source.SubjectKey {
			v := f[field].(string)
			if v == "" {
				return nil, fmt.Errorf("fact %d has empty subject-key field %s", i, field)
			}
			keyParts = append(keyParts, v)
		}
		key, _ := json.Marshal(keyParts)
		if seen[string(key)] {
			return nil, fmt.Errorf("duplicate subject in evaluation %q at fact %d", evaluationID, i)
		}
		seen[string(key)] = true
		if !matches(f, d.Population) {
			continue
		}
		dims := map[string]any{}
		groupValues := make([]any, 0, len(groups))
		for _, g := range groups {
			v := f[d.Dimensions[g]]
			dims[g] = v
			groupValues = append(groupValues, v)
		}
		groupKey, _ := json.Marshal(groupValues)
		b := buckets[string(groupKey)]
		if b == nil {
			b = &bucket{dims, map[string]uint64{}}
			buckets[string(groupKey)] = b
		}
		for name, m := range d.Measures {
			if m.Op == "count" && matches(f, m.Where) {
				b.counts[name]++
			}
		}
	}
	keys := names(buckets)
	sort.Strings(keys)
	out := make([]Result, 0, len(keys))
	for _, k := range keys {
		b := buckets[k]
		r := Result{b.dims, map[string]Value{}}
		for name, m := range d.Measures {
			if m.Op == "count" {
				n := b.counts[name]
				r.Measures[name] = Value{State: "ok", Count: &n}
				continue
			}
			state := "ok"
			for _, blocker := range m.UnknownWhen {
				if b.counts[blocker] > 0 {
					state = "unknown"
				}
			}
			if state == "ok" && b.counts[m.Denominator] == 0 {
				state = "no_data"
			}
			v := Value{State: state}
			if state == "ok" {
				x := float64(b.counts[m.Numerator]) / float64(b.counts[m.Denominator])
				v.Ratio = &x
			}
			r.Measures[name] = v
		}
		out = append(out, r)
	}
	return out, nil
}
