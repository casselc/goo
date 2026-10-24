package workbench

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
)

func TestTypedNumericPredicateProtocolFixtures(t *testing.T) {
	raw, e := os.ReadFile("../testdata/protocol/numeric-predicates.json")
	if e != nil {
		t.Fatal(e)
	}
	var fixture struct {
		Version int
		Cases   []struct {
			Name                 string
			Predicate, Canonical json.RawMessage
			Valid                bool
		}
	}
	if json.Unmarshal(raw, &fixture) != nil || fixture.Version != 1 || len(fixture.Cases) < 40 {
		t.Fatal("numeric protocol fixture invalid")
	}
	script := `const fs=require('node:fs'),p=require('../ui/question-protocol'),f=require('../ui/filter-editor');const x=JSON.parse(fs.readFileSync('../testdata/protocol/numeric-predicates.json'));console.log(JSON.stringify(x.cases.map(c=>{try{const q={signal:'spans',environment:'test',fromMs:1000,toMs:61000,limit:100,offset:0,where:c.predicate};p.validateStep(p.recordStep(q));f.validate(q.where,'spans');return {valid:true,step:p.readStepLink(p.stepLink(p.recordStep(q)))}}catch{return {valid:false}}})))`
	out, e := exec.Command("node", "-e", script).CombinedOutput()
	if e != nil {
		t.Fatalf("numeric browser worker: %v %s", e, out)
	}
	var browser []struct {
		Valid bool
		Step  Step
	}
	if json.Unmarshal(out, &browser) != nil || len(browser) != len(fixture.Cases) {
		t.Fatal("numeric browser worker response")
	}
	for i, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var predicate exploration.Predicate
			decoder := json.NewDecoder(bytes.NewReader(c.Predicate))
			decoder.DisallowUnknownFields()
			e := decoder.Decode(&predicate)
			q := exploration.Query{Signal: "spans", Environment: "test", FromMS: 1000, ToMS: 61000, Limit: 100, Where: &predicate}
			valid := e == nil && q.Validate() == nil
			if valid != c.Valid || browser[i].Valid != c.Valid {
				t.Fatalf("numeric Go/browser validity differs: Go=%v browser=%v expected=%v error=%v", valid, browser[i].Valid, c.Valid, e)
			}
			if !c.Valid {
				return
			}
			wire, _ := json.Marshal(predicate)
			var a, b any
			json.Unmarshal(wire, &a)
			json.Unmarshal(c.Canonical, &b)
			if !reflect.DeepEqual(a, b) {
				t.Fatalf("numeric canonical wire changed: %s", wire)
			}
			if browser[i].Step.Validate() != nil || browser[i].Step.Record == nil || !reflect.DeepEqual(*browser[i].Step.Record, q) {
				t.Fatal("numeric browser round trip changed binding, value or scope")
			}
		})
	}
}

// Standard encoding/json replaces unpaired surrogate escapes. The browser rejects
// those authored strings before transport; this is an explicit boundary control.
func TestTypedNumericUnicodeDecodeBoundary(t *testing.T) {
	raw, e := os.ReadFile("../testdata/protocol/numeric-predicate-unicode.json")
	if e != nil {
		t.Fatal(e)
	}
	var fixture struct {
		Version int
		Cases   []struct {
			Name, Wire string
			Decoded    json.RawMessage
		}
	}
	if json.Unmarshal(raw, &fixture) != nil || fixture.Version != 1 || len(fixture.Cases) != 2 {
		t.Fatal("Unicode boundary fixture")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var actual, want exploration.Predicate
			if json.Unmarshal([]byte(c.Wire), &actual) != nil || json.Unmarshal(c.Decoded, &want) != nil || !reflect.DeepEqual(actual, want) {
				t.Fatal("Go JSON decode boundary changed")
			}
			q := exploration.Query{Signal: "spans", Environment: "test", FromMS: 1000, ToMS: 61000, Limit: 100, Where: &actual}
			if q.Validate() != nil {
				t.Fatal("Decoded literal replacement should remain valid Unicode")
			}
		})
	}
}
