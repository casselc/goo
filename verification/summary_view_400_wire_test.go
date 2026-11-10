package verification

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
)

// Both leaf and enclosing carrier run through actual consumers. This guards the
// parent-schema admission gap found during 398 rather than testing a leaf alone.
func TestSummaryView400WireParity(t *testing.T) {
	leaf, err := os.ReadFile("../tests/fixtures/summary-view-400-wire.json")
	if err != nil {
		t.Fatal(err)
	}
	carriers, err := os.ReadFile("../tests/fixtures/summary-view-400-carriers.json")
	if err != nil {
		t.Fatal(err)
	}
	var vectors []struct {
		Name        string
		Question    exploration.Query
		Declaration json.RawMessage
		Accepted    bool
	}
	if err = json.Unmarshal(leaf, &vectors); err != nil {
		t.Fatal(err)
	}
	for _, v := range vectors {
		var d workbench.SummaryViewDeclaration
		err = json.Unmarshal(v.Declaration, &d)
		if err == nil {
			err = d.Validate(v.Question)
		}
		if (err == nil) != v.Accepted {
			t.Fatalf("%s leaf accepted=%v want=%v: %v", v.Name, err == nil, v.Accepted, err)
		}
	}
	var envelopes []struct {
		Name     string
		View     json.RawMessage
		Accepted bool
	}
	if err = json.Unmarshal(carriers, &envelopes); err != nil {
		t.Fatal(err)
	}
	store := workbench.Store{Dir: t.TempDir()}
	for _, v := range envelopes {
		var view workbench.ExplorationView
		err = json.Unmarshal(v.View, &view)
		if err == nil {
			err = view.Validate()
		}
		if (err == nil) != v.Accepted {
			t.Fatalf("%s carrier accepted=%v want=%v: %v", v.Name, err == nil, v.Accepted, err)
		}
		if !v.Accepted {
			continue
		}
		saved, e := store.CreateExplorationView(view)
		if e != nil {
			t.Fatalf("%s store: %v", v.Name, e)
		}
		held, e := store.LoadExplorationView(saved.ID)
		if e != nil || !reflect.DeepEqual(saved, held) {
			t.Fatalf("%s immutable carrier: %v", v.Name, e)
		}
	}
	input, _ := json.Marshal(map[string]json.RawMessage{"leaf": leaf, "carriers": carriers})
	cmd := exec.CommandContext(t.Context(), "node", "--input-type=module", "-e", summaryWireJS400)
	cmd.Stdin = bytes.NewReader(input)
	if out, e := cmd.CombinedOutput(); e != nil {
		t.Fatalf("JS summary leaf/carrier: %v %s", e, out)
	}
	t.Logf("%d leaf and %d enclosing summary Go/JS vectors; accepted immutable-store round trips", len(vectors), len(envelopes))
}

const summaryWireJS400 = `import fs from 'node:fs';import assert from 'node:assert/strict';import * as S from '../ui/summary-declaration.mjs';import * as C from '../ui/view-composition.mjs';import {stagedView} from '../ui/view-template-handoff.mjs';const x=JSON.parse(fs.readFileSync(0,'utf8'));for(const v of x.leaf){let ok=true;try{S.validate(v.declaration,v.question)}catch{ok=false}assert.equal(ok,v.accepted,v.name)}for(const v of x.carriers){let ok=true;try{C.validate(v.view.composition,v.view.queries[0]);stagedView(v.view,v.view.queries[0])}catch{ok=false}assert.equal(ok,v.accepted,v.name)}process.stdout.write('summary leaves and complete carriers agree');`
