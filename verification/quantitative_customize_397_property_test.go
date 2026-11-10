package verification

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"reflect"
	"testing"

	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

// Generate presentation/reuse choices over the actual admitted source frames.
// This is local semantic exploration, not additional source or browser evidence.
func TestHegelQuantitativeCustomization397(t *testing.T) {
	path := os.Getenv("GOO_DISTRIBUTION_CUSTOMIZE_397_REPORT")
	if path == "" {
		t.Skip("actual quantitative source report required")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var report quantLedger397
	if e = json.Unmarshal(raw, &report); e != nil {
		t.Fatal(e)
	}
	if len(report.Cases) != 3 {
		t.Fatal("three actual quantitative source frames required")
	}
	type vector struct {
		Data         any            `json:"data"`
		Destination  any            `json:"destination"`
		Presentation map[string]any `json:"presentation"`
	}
	vectors := []vector{}
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 100000))
		for _, c := range report.Cases {
			fields := []string{"count"}
			for _, field := range c.Source.Response.ViewDataset.Fields {
				if field.Aggregation != nil && (field.Aggregation.Family == "numeric" || field.Aggregation.Family == "duration") {
					fields = append(fields, field.Key)
				}
			}
			value := fields[seed%len(fields)]
			scale := "max"
			if value == "count" && seed%2 == 0 {
				scale = "population"
			}
			vectors = append(vectors, vector{c.Source.Response, c.DestinationQuestion, map[string]any{"kind": []string{"bar", "table"}[seed%2], "value": value, "scale": scale, "order": []string{"source", "label", "value-desc", "value-asc"}[seed%4], "query": []string{"", "read", "write", "雪"}[seed%4], "requiresComplete": seed%2 == 0}})
		}
	}, options(16)...)
	input, _ := json.Marshal(vectors)
	cmd := exec.CommandContext(t.Context(), "node", "--input-type=module", "-e", quantPropertyJS397)
	cmd.Stdin = bytes.NewReader(input)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, e := cmd.Output()
	if e != nil {
		t.Fatalf("generated actual quantitative configuration: %v %s", e, stderr.String())
	}
	var staged []workbench.ExplorationView
	if e = json.Unmarshal(out, &staged); e != nil || len(staged) != len(vectors) {
		t.Fatalf("invalid generated staged output: %v", e)
	}
	store := workbench.Store{Dir: t.TempDir()}
	for _, view := range staged {
		if e = view.Validate(); e != nil {
			t.Fatal(e)
		}
		saved, e := store.CreateExplorationView(view)
		if e != nil {
			t.Fatal(e)
		}
		held, e := store.LoadExplorationView(saved.ID)
		if e != nil || !reflect.DeepEqual(saved, held) {
			t.Fatal("generated immutable store mismatch", e)
		}
	}
	t.Logf("%d generated local quantitative staging/reuse/store vectors over three actual source frames", len(vectors))
}

const quantPropertyJS397 = `import fs from 'node:fs';import assert from 'node:assert/strict';import {distributionCustomization} from '../ui/distributions.mjs';import {adoptExploration} from '../ui/source-snapshots.mjs';import * as D from '../ui/view-declaration.mjs';globalThis.fetch=()=>{throw Error('Local generated customization read telemetry')};const out=[];for(const v of JSON.parse(fs.readFileSync(0,'utf8'))){const before=JSON.stringify(v),q=v.data.question,data=adoptExploration(v.data,q),view=distributionCustomization(data,v.presentation),d=view.displays[q.signal];assert.deepEqual(view.queries,[q]);assert.equal(d.display.value,v.presentation.value);assert.deepEqual(D.reuse(d,q,v.destination),d);const next=D.update(d,{...D.presentation(d),kind:d.kind.name==='bar'?'table':'bar'},q);assert.equal(next.display.value,d.display.value);assert.deepEqual(next.scope,d.scope);for(const query of ['x\n','雪'.repeat(86),'\ud800'])assert.throws(()=>distributionCustomization(data,{...v.presentation,query}));if(d.display.value!=='count')assert.throws(()=>distributionCustomization(data,{...v.presentation,scale:'population'}));if(d.display.value.startsWith('numeric.'))for(const change of [x=>x.numeric.unit='other',x=>x.numeric.field.kind='resource',x=>x.numeric.field.key='other',x=>x.numeric.type='uint64',x=>x.numeric.parse='lenient']){const destination=structuredClone(v.destination);change(destination);assert.throws(()=>D.reuse(d,q,destination))}if(d.display.value.startsWith('duration.')){const destination=structuredClone(v.destination);delete destination.summary;assert.throws(()=>D.reuse(d,q,destination))}assert.equal(Object.hasOwn(view,'basis'),false);assert.equal(Object.hasOwn(view,'viewDataset'),false);assert.equal(JSON.stringify(v),before);out.push(view)}process.stdout.write(JSON.stringify(out));`
