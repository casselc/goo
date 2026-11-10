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

// One finite authored-wire corpus is consumed by Go decoding/validation, the
// actual JavaScript declaration validator, and the generated API schemas.
// This proves admission parity, not actual source membership or UI rendering.
func TestQuantitativeView397WireParity(t *testing.T) {
	raw, err := os.ReadFile("../tests/fixtures/quantitative-view-397-wire.json")
	if err != nil {
		t.Fatal(err)
	}
	var vectors []struct {
		Name        string            `json:"name"`
		Question    exploration.Query `json:"question"`
		Declaration json.RawMessage   `json:"declaration"`
		Accepted    bool              `json:"accepted"`
	}
	if err = json.Unmarshal(raw, &vectors); err != nil {
		t.Fatal(err)
	}
	store := workbench.Store{Dir: t.TempDir()}
	for _, v := range vectors {
		var d workbench.ViewDeclaration
		err = json.Unmarshal(v.Declaration, &d)
		if err == nil {
			err = d.Validate(v.Question)
		}
		if (err == nil) != v.Accepted {
			t.Fatalf("%s: accepted=%v want=%v error=%v", v.Name, err == nil, v.Accepted, err)
		}
		if !v.Accepted {
			continue
		}
		view := workbench.ExplorationView{Version: 1, Title: "Quantitative contract", Queries: []exploration.Query{v.Question}, Layout: "charts-records", Columns: []string{"utc", "service", "summary", "context"}, Displays: map[string]workbench.ViewDeclaration{v.Question.Signal: d}}
		saved, e := store.CreateExplorationView(view)
		if e != nil {
			t.Fatalf("%s store: %v", v.Name, e)
		}
		held, e := store.LoadExplorationView(saved.ID)
		if e != nil || !reflect.DeepEqual(saved, held) {
			t.Fatalf("%s immutable round trip: %v", v.Name, e)
		}
	}
	cmd := exec.CommandContext(t.Context(), "node", "--input-type=module", "-e", `import fs from 'node:fs';import assert from 'node:assert/strict';import {validate} from '../ui/view-declaration.mjs';for(const v of JSON.parse(fs.readFileSync(0,'utf8'))){let accepted=true;try{validate(v.declaration,v.question)}catch{accepted=false}assert.equal(accepted,v.accepted,v.name)}process.stdout.write('shared quantitative declarations agree');`)
	cmd.Stdin = bytes.NewReader(raw)
	if out, e := cmd.CombinedOutput(); e != nil {
		t.Fatalf("JS quantitative wire: %v %s", e, out)
	}
	t.Logf("%d finite quantitative Go/JS authored-wire vectors, accepted store round trips", len(vectors))
}
