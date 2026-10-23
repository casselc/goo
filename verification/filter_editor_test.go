package verification

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
	"strconv"
	"testing"
)

// Saved predicates preserve membership through omitempty, including absent keys.
func TestFilterEditorRoundTripProperties(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		n := hegel.Draw(tc, hegel.Integers(0, 1000000))
		depth := hegel.Draw(tc, hegel.Integers(0, 3))
		f := exploration.Field{Kind: "attribute", Key: "release"}
		p := exploration.Predicate{Any: []exploration.Predicate{{Field: &f, Op: "missing"}, {Field: &f, Op: "eq", Value: ""}}}
		for i := 0; i < depth; i++ {
			if (n+i)%2 == 0 {
				p = exploration.Predicate{All: []exploration.Predicate{p}}
			} else {
				p = exploration.Predicate{Any: []exploration.Predicate{p}}
			}
		}
		q := exploration.Query{Signal: "logs", Environment: "prod", FromMS: 0, ToMS: 60000, Limit: 100, Where: &p}
		if e := q.Validate(); e != nil {
			t.Fatal(e)
		}
		b, e := json.Marshal(q)
		if e != nil {
			t.Fatal(e)
		}
		var restored exploration.Query
		if e = json.Unmarshal(b, &restored); e != nil {
			t.Fatal(e)
		}
		for _, record := range []map[exploration.Field]string{{}, {f: ""}, {f: strconv.Itoa(n)}} {
			if exploration.Match(q.Where, record) != exploration.Match(restored.Where, record) {
				t.Fatal("changed membership after serialization")
			}
		}
		if !exploration.Match(restored.Where, map[exploration.Field]string{}) || !exploration.Match(restored.Where, map[exploration.Field]string{f: ""}) || exploration.Match(restored.Where, map[exploration.Field]string{f: "canary"}) {
			t.Fatal("missing and empty union changed")
		}
	}, options(80)...)
}
