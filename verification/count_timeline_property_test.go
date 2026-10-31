package verification

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
	"testing"
)

const timelineParityJS = `import assert from 'node:assert/strict';import fs from 'node:fs';import * as C from '../ui/view-composition.mjs';const {composition,source,destination,accepted}=JSON.parse(fs.readFileSync(0,'utf8'));let out;try{out=C.reuse(composition,source,destination)}catch{}assert.equal(!!out,accepted);if(out){assert.deepEqual(out,composition);C.timeline(out).display.title='changed';assert.notEqual(C.timeline(composition).display.title,'changed')}`

func TestHegelCountTimelineWidthReuse(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 1000))
		mismatch := hegel.Draw(tc, hegel.Booleans())
		raw, _ := json.Marshal(compositionWire(seed))
		var view workbench.ExplorationView
		if e := json.Unmarshal(raw, &view); e != nil {
			tc.Fatal(e)
		}
		source := view.Queries[0]
		widths := []int64{1000, 10000, 60000, 300000}
		source.BucketMS = widths[seed%4]
		source.FromMS = 1000
		source.ToMS = 1000 + source.BucketMS*int64(seed%int(min(int64(239), 7200000/source.BucketMS))+1)
		timeline := &workbench.TimelineViewDeclaration{Version: "goo.view/1", Scope: workbench.ViewDeclarationScope{Kind: "record-query", Signal: source.Signal}, Kind: workbench.TimelineViewKind{Name: "time-bar", Version: 1}, Display: workbench.TimelineViewDisplay{Value: "count"}}
		view.Composition.Children = append(view.Composition.Children, workbench.ViewCompositionChild{Timeline: timeline, Hidden: seed%2 == 0})
		destRaw, _ := json.Marshal(source)
		var destination exploration.Query
		_ = json.Unmarshal(destRaw, &destination)
		destination.Environment = "current"
		destination.FromMS += 1000
		destination.ToMS += 1000
		if mismatch {
			destination.BucketMS = widths[(seed+1)%4]
		}
		before, _ := json.Marshal(view.Composition)
		out, e := workbench.ReuseViewComposition(*view.Composition, source, destination)
		if (e == nil) == mismatch {
			tc.Fatal("width compatibility differs", e)
		}
		if e := runNativeHistogramJS(map[string]any{"composition": view.Composition, "source": source, "destination": destination, "accepted": e == nil}, timelineParityJS); e != nil {
			tc.Fatal(e)
		}
		if e == nil {
			out.Children[len(out.Children)-1].Timeline.Display.Title = "changed"
		}
		after, _ := json.Marshal(view.Composition)
		if string(before) != string(after) {
			tc.Fatal("timeline aliases source")
		}
	}, options(100)...)
}
