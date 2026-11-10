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
	"hegel.dev/go/hegel"
)

// Generated local authoring operates on actual admitted source frames. It is
// configuration algebra, not additional telemetry or a substitute for wrappers.
func TestHegelSummaryView400(t *testing.T) {
	path := os.Getenv("GOO_SUMMARY_VIEW_400_REPORT")
	if path == "" {
		t.Skip("actual summary journey report required")
	}
	wire, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var report struct {
		Cases []struct {
			Source              customizeRead394
			DestinationQuestion exploration.Query
		}
	}
	if err = json.Unmarshal(wire, &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Cases) != 3 {
		t.Fatal("three actual ordinary summary frames required")
	}
	type vector struct {
		Data        exploration.Result `json:"data"`
		Destination exploration.Query  `json:"destination"`
		Fields      []string           `json:"fields"`
		Title       string             `json:"title"`
	}
	vectors := []vector{}
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 100000))
		for _, c := range report.Cases {
			if c.Source.Response.SummaryDataset == nil {
				t.Fatal("missing actual summary dataset")
			}
			all := []string{}
			for _, f := range c.Source.Response.SummaryDataset.Fields {
				all = append(all, f.Key)
			}
			limit := 1 + seed%min(6, len(all))
			chosen := []string{}
			for i := 0; i < limit; i++ {
				chosen = append(chosen, all[(i+seed)%len(all)])
			}
			vectors = append(vectors, vector{c.Source.Response, c.DestinationQuestion, chosen, []string{"", "Relevant facts", "雪 summary"}[seed%3]})
		}
	}, options(16)...)
	input, _ := json.Marshal(vectors)
	cmd := exec.CommandContext(t.Context(), "node", "--input-type=module", "-e", summaryPropertyJS400)
	cmd.Stdin = bytes.NewReader(input)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("local summary generation: %v %s", err, stderr.String())
	}
	var views []workbench.ExplorationView
	if err = json.Unmarshal(out, &views); err != nil || len(views) != len(vectors) {
		t.Fatalf("invalid generated views: %v", err)
	}
	store := workbench.Store{Dir: t.TempDir()}
	for _, view := range views {
		if err = view.Validate(); err != nil {
			t.Fatal(err)
		}
		saved, e := store.CreateExplorationView(view)
		if e != nil {
			t.Fatal(e)
		}
		held, e := store.LoadExplorationView(saved.ID)
		if e != nil || !reflect.DeepEqual(saved, held) {
			t.Fatal("immutable summary round trip differs", e)
		}
	}
	t.Logf("%d generated local summary field/order/reuse/store vectors over three actual source frames", len(vectors))
}

const summaryPropertyJS400 = `import fs from 'node:fs';import assert from 'node:assert/strict';import * as S from '../ui/summary-declaration.mjs';import * as C from '../ui/view-composition.mjs';import {summaryCustomization} from '../ui/distributions.mjs';import {adoptExploration} from '../ui/source-snapshots.mjs';globalThis.fetch=()=>{throw Error('Local summary authoring read telemetry')};const out=[];for(const v of JSON.parse(fs.readFileSync(0,'utf8'))){const before=JSON.stringify(v),q=v.data.question,data=adoptExploration(v.data,q),view=summaryCustomization(data);assert.deepEqual(view.queries,[q]);const selected=S.validate({...S.create(q),display:{fields:v.fields,title:v.title}},q);view.composition=C.setSummary(view.composition,selected,q);assert.deepEqual(C.summary(view.composition).display.fields,v.fields);const reused=C.reuse(view.composition,q,v.destination);assert.deepEqual(reused,view.composition);const reversed=S.validate({...selected,display:{...selected.display,fields:[...v.fields].reverse()}},q);assert.deepEqual(reversed.display.fields,[...v.fields].reverse());for(const fields of [[],['count','count'],Array(7).fill('count'),['invented'],null])assert.throws(()=>S.validate({...selected,display:{fields}},q));for(const title of ['x'+String.fromCharCode(10),'雪'.repeat(41),String.fromCharCode(0xd800)])assert.throws(()=>S.validate({...selected,display:{fields:v.fields,title}},q));for(const extra of [{snapshot:{}},{basis:'donor'},{query:q}])assert.throws(()=>S.validate({...selected,...extra},q));if(v.fields.some(f=>f.startsWith('numeric.')))for(const mutate of [x=>x.numeric.unit='other',x=>x.numeric.field.kind='resource',x=>x.numeric.field.key='other',x=>x.numeric.type='uint64',x=>x.numeric.parse='lenient']){const dest=structuredClone(v.destination);mutate(dest);assert.throws(()=>S.reuse(selected,q,dest))}if(v.fields.some(f=>f.startsWith('duration.'))){const dest=structuredClone(v.destination);delete dest.summary;assert.throws(()=>S.reuse(selected,q,dest))}if(v.fields.some(f=>f.startsWith('status.'))){const dest=structuredClone(v.destination);delete dest.spanStatus;assert.throws(()=>S.reuse(selected,q,dest))}assert.equal(JSON.stringify(v),before);assert.equal(Object.hasOwn(view,'snapshot'),false);assert.equal(Object.hasOwn(view,'basis'),false);out.push(view)}process.stdout.write(JSON.stringify(out));`
