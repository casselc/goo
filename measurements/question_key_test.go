package measurements

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/casselc/goo/identity"
)

func TestQuestionKeyPreservesFixedRepresentationAndDoesNotMutate(t *testing.T) {
	d := definition(t, "checkout-calls")
	keys := map[string]bool{}
	for _, exact := range []bool{false, true} {
		for _, ns := range []*identity.Namespace{nil, {Mode: "missing"}, {Mode: "value", Value: ""}, {Mode: "value", Value: "team"}} {
			for _, attributes := range []map[string]string{nil, {}, {"custom.attr": ""}} {
				d.Source.Namespace = ns
				d.Source.Attributes = attributes
				q := Question{Definition: d, FromMS: 100000, ToMS: 102001}
				if exact {
					q.FromNS = "100000000123"
					q.ToNS = "102000000123"
				}
				before, err := json.Marshal(q)
				if err != nil {
					t.Fatal(err)
				}
				key, err := q.Key()
				if err != nil || key != Hash(q) {
					t.Fatal("fixed identity changed", key, err)
				}
				after, err := json.Marshal(q)
				if err != nil || !bytes.Equal(before, after) {
					t.Fatal("key mutated question", err)
				}
				keys[key] = true
			}
		}
	}
	// At least exact and enclosing-ms windows and literal namespace modes remain distinct.
	if len(keys) < 8 {
		t.Fatal("collapsed fixed representation", len(keys))
	}
	q := Question{Definition: d, FromMS: 100000, ToMS: 102001, FromNS: "100000000123", ToNS: "102000000123"}
	first, err := q.Key()
	if err != nil {
		t.Fatal(err)
	}
	q.ToNS = "102000000124"
	next, err := q.Key()
	if err != nil || first == next {
		t.Fatal("nanosecond boundary missing from identity", err)
	}
}

func TestQuestionKeyRejectsInvalidScopeWithoutIdentifier(t *testing.T) {
	valid := Question{Definition: definition(t, "checkout-calls"), FromMS: 100000, ToMS: 102001}
	cases := []Question{{}, valid, valid, valid, valid, valid, valid, valid}
	cases[1].Definition.Name = ""
	cases[2].ToMS = cases[2].FromMS
	cases[3].FromMS = -1
	cases[4].ToMS = cases[4].FromMS + 7200001
	cases[5].FromNS = "100000000123"
	cases[6].FromNS = "0100000000123"
	cases[6].ToNS = "102000000123"
	cases[7].FromNS = "100001000123"
	cases[7].ToNS = "102000000123"
	for _, q := range cases {
		if key, err := q.Key(); err == nil || key != "" {
			t.Fatal("invalid question acquired a key", key, err)
		}
	}
}
