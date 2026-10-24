package workbench

import (
	"encoding/json"
	"os"
	"os/exec"
	"reflect"
	"testing"

	"github.com/casselc/goo/measurements"
)

func TestSharedStepProtocolFixtures(t *testing.T) {
	raw, e := os.ReadFile("../testdata/protocol/question-steps.json")
	if e != nil {
		t.Fatal(e)
	}
	var f struct {
		Version int
		Cases   []struct {
			Name               string
			Input, Canonical   json.RawMessage
			Valid              bool
			DefinitionRevision string
		}
	}
	if e = json.Unmarshal(raw, &f); e != nil {
		t.Fatal(e)
	}
	if f.Version != 1 || len(f.Cases) != 16 {
		t.Fatal("fixture coverage changed")
	}
	script := `const fs=require('node:fs'),assert=require('node:assert/strict'),w=require('../ui/question-protocol'),compat=require('../ui/workspace'),f=JSON.parse(fs.readFileSync('../testdata/protocol/question-steps.json'));for(const key of Object.keys(w))assert.equal(w[key],compat[key]);console.log(JSON.stringify(f.cases.map(c=>{try{return {valid:true,step:w.readStepLink(w.stepLink(c.input))}}catch{return {valid:false}}})));`
	out, e := exec.Command("node", "-e", script).CombinedOutput()
	if e != nil {
		t.Fatalf("JS protocol worker: %v %s", e, out)
	}
	var browser []struct {
		Valid bool
		Step  Step
	}
	if e = json.Unmarshal(out, &browser); e != nil || len(browser) != len(f.Cases) {
		t.Fatal("browser fixture output", e)
	}
	for i, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var s Step
			if e := json.Unmarshal(c.Input, &s); e != nil {
				if c.Valid || browser[i].Valid {
					t.Fatal("invalid wire binding accepted", e)
				}
				return
			}
			if valid := s.Validate() == nil; valid != c.Valid || browser[i].Valid != c.Valid {
				t.Fatal("Go/JS validity changed")
			}
			if !c.Valid {
				return
			}
			wire, _ := json.Marshal(s)
			var a, b any
			json.Unmarshal(wire, &a)
			json.Unmarshal(c.Canonical, &b)
			if !reflect.DeepEqual(a, b) {
				t.Fatalf("Go canonical wire changed: %s", wire)
			}
			if !reflect.DeepEqual(s, browser[i].Step) {
				t.Fatal("browser roundtrip changed typed scope or binding")
			}
			if s.Measurement != nil && (measurements.Hash(s.Measurement) != c.DefinitionRevision || measurements.Hash(browser[i].Step.Measurement) != c.DefinitionRevision) {
				t.Fatal("native definition digest changed")
			}
		})
	}
}
