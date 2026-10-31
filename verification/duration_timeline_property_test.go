package verification

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
	"testing"
)

const durationTimelineParityJS = `import assert from 'node:assert/strict';import fs from 'node:fs';import * as C from '../ui/view-composition.mjs';const {composition,source,destination,accepted}=JSON.parse(fs.readFileSync(0,'utf8'));let out;try{out=C.reuse(composition,source,destination)}catch{}assert.equal(!!out,accepted);if(out){assert.deepEqual(out,composition);C.timeline(out).display.title='changed';assert.notEqual(C.timeline(composition).display.title,'changed')}`

func TestHegelDurationTimelineSupportReuse(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 10000))
		mode := hegel.Draw(tc, hegel.Integers(0, 5))
		summary := hegel.Draw(tc, hegel.Booleans())
		raw, _ := json.Marshal(compositionWire(seed))
		var v workbench.ExplorationView
		if e := json.Unmarshal(raw, &v); e != nil {
			tc.Fatal(e)
		}
		q := v.Queries[0]
		q.Signal = []string{"spans", "logs"}[(seed/4)%2]
		q.BucketMS = 10000
		q.FromMS = 1001
		q.ToMS = 61001
		q.Numeric = nil
		q.Summary = ""
		q.SpanStatus = false
		if summary && q.Signal == "spans" {
			q.Summary = "span-duration"
		}
		for i := range v.Composition.Children {
			if v.Composition.Children[i].Aggregate != nil {
				v.Composition.Children[i].Aggregate.Scope.Signal = q.Signal
				v.Composition.Children[i].Aggregate.Display.Value = "count"
			}
			if v.Composition.Children[i].Records != nil {
				v.Composition.Children[i].Records.Scope.Signal = q.Signal
			}
		}
		measure := []string{"count", "duration.mean", "duration.p95", "rate"}[seed%4]
		timeline := &workbench.TimelineViewDeclaration{Version: "goo.view/1", Scope: workbench.ViewDeclarationScope{Kind: "record-query", Signal: q.Signal}, Kind: workbench.TimelineViewKind{Name: []string{"time-bar", "table"}[(seed/8)%2], Version: 1}, Display: workbench.TimelineViewDisplay{Value: measure, RequiresComplete: seed/32%2 == 0}}
		v.Composition.Children = append(v.Composition.Children, workbench.ViewCompositionChild{Timeline: timeline, Hidden: seed%3 == 0})
		destRaw, _ := json.Marshal(q)
		var dest exploration.Query
		_ = json.Unmarshal(destRaw, &dest)
		dest.FromMS += 1000
		dest.ToMS += 1000
		dest.Environment = "current"
		switch mode {
		case 1:
			dest.Summary = ""
		case 2:
			if dest.Signal == "spans" {
				dest.Signal = "logs"
				dest.Summary = ""
			} else {
				dest.Signal = "spans"
			}
		case 3:
			dest.BucketMS = 60000
		case 4:
			if dest.Signal == "spans" {
				dest.Summary = "span-duration"
			}
		}
		before, _ := json.Marshal(v.Composition)
		out, e := workbench.ReuseViewComposition(*v.Composition, q, dest)
		supported := measure == "count" || (q.Signal == "spans" && q.Summary == "span-duration" && (measure == "duration.mean" || measure == "duration.p95"))
		expected := supported && q.Signal == dest.Signal && q.BucketMS == dest.BucketMS && (measure == "count" || dest.Summary == "span-duration")
		if (e == nil) != expected {
			tc.Fatal("duration capability differs", measure, q, dest, e)
		}
		if e := runNativeHistogramJS(map[string]any{"composition": v.Composition, "source": q, "destination": dest, "accepted": e == nil}, durationTimelineParityJS); e != nil {
			tc.Fatal(e)
		}
		if e == nil {
			out.Children[len(out.Children)-1].Timeline.Display.Title = "changed"
		}
		after, _ := json.Marshal(v.Composition)
		if string(before) != string(after) {
			tc.Fatal("timeline aliases saved declaration")
		}
	}, options(100)...)
}
