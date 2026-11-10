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
func TestGroupedMean398WireParity(t *testing.T) {
	raw, err := os.ReadFile("../tests/fixtures/grouped-mean-398-wire.json")
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
		var d workbench.TimelineViewDeclaration
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
		view := workbench.ExplorationView{Version: 1, Title: "Grouped mean contract", Queries: []exploration.Query{v.Question}, Composition: &workbench.ViewComposition{Version: "goo.composition/1", Kind: workbench.CompositionKind{Name: "stack", Version: 1}, Children: []workbench.ViewCompositionChild{{Timeline: &d}, {Records: &workbench.RecordViewDeclaration{Version: "goo.view/1", Scope: workbench.ViewDeclarationScope{Kind: "record-query", Signal: v.Question.Signal}, Kind: workbench.RecordViewKind{Name: "record-table", Version: 1}, Display: workbench.RecordViewDisplay{Columns: []string{"utc", "summary"}}}}}}}
		saved, e := store.CreateExplorationView(view)
		if e != nil {
			t.Fatalf("%s store: %v", v.Name, e)
		}
		held, e := store.LoadExplorationView(saved.ID)
		if e != nil || !reflect.DeepEqual(saved, held) {
			t.Fatalf("%s immutable round trip: %v", v.Name, e)
		}
	}
	cmd := exec.CommandContext(t.Context(), "node", "--input-type=module", "-e", `import fs from 'node:fs';import assert from 'node:assert/strict';import {validate} from '../ui/timeline-declaration.mjs';for(const v of JSON.parse(fs.readFileSync(0,'utf8'))){let accepted=true;try{validate(v.declaration,v.question)}catch{accepted=false}assert.equal(accepted,v.accepted,v.name)}process.stdout.write('shared grouped mean declarations agree');`)
	cmd.Stdin = bytes.NewReader(raw)
	if out, e := cmd.CombinedOutput(); e != nil {
		t.Fatalf("JS grouped mean wire: %v %s", e, out)
	}
	t.Logf("%d finite grouped mean Go/JS authored-wire vectors, accepted store round trips", len(vectors))
}
