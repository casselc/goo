package verification

import (
	"bytes"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
	"os"
	"os/exec"
	"reflect"
	"testing"
)

// Generated local combinations consume actual admitted raw-span frames. This
// explores configuration algebra, not additional source acquisition evidence.
func TestHegelGroupedMean398(t *testing.T) {
	path := os.Getenv("GOO_GROUPED_MEAN_398_REPORT")
	if path == "" {
		t.Skip("actual grouped mean source frames required")
	}
	wire, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var r meanLedger398
	if e = json.Unmarshal(wire, &r); e != nil {
		t.Fatal(e)
	}
	if len(r.Cases) != 1 {
		t.Fatal("one actual grouped mean case required")
	}
	type vector struct {
		Data        exploration.Result                `json:"data"`
		Destination exploration.Query                 `json:"destination"`
		Declaration workbench.TimelineViewDeclaration `json:"declaration"`
	}
	vectors := []vector{}
	c := r.Cases[0]
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 100000))
		for _, read := range []customizeRead394{c.Source, c.ChildRun} {
			d := workbench.TimelineViewDeclaration{Version: "goo.view/1", Scope: workbench.ViewDeclarationScope{Kind: "record-query", Signal: "spans"}, Kind: workbench.TimelineViewKind{Name: []string{"small-multiples", "table", "heatmap"}[seed%3], Version: 1}, Display: workbench.TimelineViewDisplay{Profile: "grouped-duration-mean-v1", Value: "duration.mean", Query: []string{"", "雪", "<literal>"}[seed%3], Order: []string{"source", "label"}[seed%2], RequiresComplete: seed%2 == 0}}
			if e := d.Validate(read.Request); e != nil {
				tc.Fatal(e)
			}
			next, e := workbench.ReuseTimelineViewDeclaration(d, read.Request, c.DestinationQuestion)
			if e != nil || next != d {
				tc.Fatal("mean reuse changed declaration", e)
			}
			vectors = append(vectors, vector{read.Response, c.DestinationQuestion, d})
		}
	}, options(16)...)
	input, _ := json.Marshal(vectors)
	cmd := exec.CommandContext(t.Context(), "node", "--input-type=module", "-e", meanPropertyJS398)
	cmd.Stdin = bytes.NewReader(input)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, e := cmd.Output()
	if e != nil {
		t.Fatalf("generated mean contract: %v %s", e, stderr.String())
	}
	var views []workbench.ExplorationView
	if e = json.Unmarshal(out, &views); e != nil || len(views) != len(vectors) {
		t.Fatal("generated mean output", e)
	}
	store := workbench.Store{Dir: t.TempDir()}
	for i, v := range views {
		if !reflect.DeepEqual(v.Queries, []exploration.Query{vectors[i].Data.Question}) {
			t.Fatal("mean composition changed question")
		}
		saved, e := store.CreateExplorationView(v)
		if e != nil {
			t.Fatal(e)
		}
		held, e := store.LoadExplorationView(saved.ID)
		if e != nil || !reflect.DeepEqual(saved, held) {
			t.Fatal("mean immutable roundtrip", e)
		}
	}
	t.Logf("%d generated local mean Go/JS staging/reuse/store vectors, unchanged scale and rejecting support mutations", len(vectors))
}

const meanPropertyJS398 = `import fs from 'node:fs';import assert from 'node:assert/strict';import * as T from '../ui/timeline-declaration.mjs';import * as C from '../ui/view-composition.mjs';import {validateSnapshot,validateDisplay} from '../ui/view-model.mjs';import {adoptExploration} from '../ui/source-snapshots.mjs';import {hasGuarantee} from '../ui/snapshots.mjs';import {ceiling,groupRows} from '../ui/grouped-timeline-view.mjs';import {distributionCustomization} from '../ui/distributions.mjs';globalThis.fetch=()=>{throw Error('Local mean configuration read telemetry')};const views=[];for(const v of JSON.parse(fs.readFileSync(0,'utf8'))){const before=JSON.stringify(v),q=v.data.question,data=adoptExploration(v.data,q),s=data.groupedTimeDataset,d=v.declaration;assert.equal(validateSnapshot(s).ok,true);assert.equal(hasGuarantee(s,'grouped-intervals'),true);assert.deepEqual(T.reuse(d,q,v.destination),d);const display={kind:d.kind.name,value:'duration.mean',query:d.display.query,order:d.display.order,requiresComplete:d.display.requiresComplete};assert.equal(validateDisplay(s,display).ok,true);const full=ceiling(s,display);for(const query of ['','雪','no-match']){groupRows(s,{...display,query});assert.equal(ceiling(s,{...display,query}),full)}for(const edit of [x=>x.summary='',x=>x.signal='logs',x=>x.bucketMs=10000,x=>x.groupBy[0].kind='resource',x=>x.groupBy[0].key='other']){const target=structuredClone(v.destination);edit(target);assert.throws(()=>T.reuse(d,q,target))}for(const edit of [x=>x.display.value='count',x=>x.display.value='duration.p95',x=>x.display.profile='grouped-count-v1',x=>x.display.query='雪'.repeat(86),x=>x.display.order='value-desc']){const bad=structuredClone(d);edit(bad);assert.throws(()=>T.validate(bad,q))}for(const edit of [x=>{const row=x.rows.find(r=>r.state==='observed');row.values['duration.mean']='999999999'},x=>{const row=x.rows.find(r=>r.state==='no_data');row.values['duration.mean']='0'},x=>{const row=x.rows.find(r=>r.state==='observed');row.cells['duration.mean'].support.valid='0'}]){const bad=structuredClone(s);edit(bad);assert.equal(validateSnapshot(bad).ok,false)}const view=distributionCustomization(data,d);C.view(view);assert.deepEqual(view.queries,[q]);assert.equal(view.composition.children[0].timeline.display.value,'duration.mean');assert.equal(JSON.stringify(v),before);views.push(view)}process.stdout.write(JSON.stringify(views));`
