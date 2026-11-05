package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

// Exercise the existing completed-window planner and fixed comparison grammar,
// rather than a second implementation of their scope and clock arithmetic.
func TestHegelActivityWindowComparison333(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		signal := "logs"
		if seed%2 == 0 {
			signal = "spans"
		}
		mode := []string{"named", "empty", "unreported"}[seed%3]
		scope := map[string]any{"mode": mode}
		if mode == "named" {
			scope["value"] = fmt.Sprintf("prod-'雪%d", seed)
		}
		policy := "effective"
		if seed%2 == 0 && mode == "named" {
			policy = "current-only"
		}
		selector := map[string]any{"signal": signal, "environmentSelector": map[string]any{"version": "goo/environment-selector/v1", "policy": policy, "scope": scope}, "where": map[string]any{"field": map[string]any{"kind": "attribute", "key": "tenant"}, "op": "eq", "value": fmt.Sprintf("null: \"'雪%d", seed)}}
		namespace := map[string]any{"field": map[string]any{"kind": "resource", "key": "service.namespace"}, "op": "eq", "value": ""}
		if seed%3 == 0 {
			namespace["op"] = "missing"
			delete(namespace, "value")
		} else if seed%3 == 2 {
			namespace["value"] = "team-'雪"
		}
		selector["where"] = map[string]any{"all": []any{selector["where"], namespace}}
		input, _ := json.Marshal(map[string]any{"active": true, "clean": true, "phase": "paused", "selector": selector, "readWindow": map[string]int64{"fromMs": 1790000000000 + int64(seed)*1000, "toMs": 1790000060000 + int64(seed)*1000}})
		script := `import fs from 'node:fs';import assert from 'node:assert/strict';import * as A from '../ui/activity-explanation.mjs';import * as S from '../ui/stream-entry.mjs';import * as C from '../ui/view-composition.mjs';import * as D from '../ui/view-declaration.mjs';import * as P from '../ui/captured-count-comparison.mjs';
const context=JSON.parse(fs.readFileSync(0,'utf8')),before=JSON.stringify(context);globalThis.fetch=()=>{throw Error('planner attempted IO')};const view=S.previousWindowView(context),q=view.queries[0],composition=view.composition,b=C.comparison(composition).baselineQuery;assert.deepEqual(q.groupBy,[{kind:'column',key:'service'}]);assert.deepEqual(q.environmentSelector,context.selector.environmentSelector);assert.deepEqual(q.where,context.selector.where);assert.equal(q.fromMs,context.readWindow.fromMs);assert.equal(b.toMs,q.fromMs);assert.equal(b.fromMs,q.fromMs-60000);assert.equal(b.offset,0);assert.equal(q.offset,0);assert.equal(JSON.stringify(context),before);assert.deepEqual(S.previousWindowView({...context,rows:[{record:{count:'999999999'}}],count:'999999',sourceBasis:'not-authority'}),view);for(const phase of ['loading','reconnecting','stale','stopped','no_data',undefined])assert.throws(()=>S.previousWindowView({...context,phase}));for(const change of [{active:false},{clean:false},{active:undefined},{clean:undefined}])assert.throws(()=>S.previousWindowView({...context,...change}));for(const phase of ['unknown','partial','paused'])assert.deepEqual(S.previousWindowView({...context,phase}),view);
const origin={epoch:7,selector:structuredClone(context.selector),readWindow:structuredClone(context.readWindow)};assert.deepEqual(S.plotWindow(context,origin,7),A.windowQuestion(context));assert.throws(()=>S.plotWindow(context,origin,8));const foreign=structuredClone(context);foreign.selector.environmentSelector.policy=foreign.selector.environmentSelector.policy==='effective'?'current-only':'effective';assert.throws(()=>S.plotWindow(foreign,origin,7));
for(const w of [null,{fromMs:q.fromMs,toMs:q.toMs+1},{fromMs:-1,toMs:59999},{fromMs:1.5,toMs:60001.5},{fromMs:q.toMs,toMs:q.fromMs}])assert.throws(()=>S.previousWindowView({...context,readWindow:w}));assert.throws(()=>S.previousWindowView({...context,readWindow:{fromMs:59999,toMs:119999}}));
const newer=structuredClone(context);newer.readWindow.fromMs+=60000;newer.readWindow.toMs+=60000;const fresh=S.previousWindowView(newer);assert.equal(fresh.queries[0].fromMs,q.toMs);assert.equal(C.comparison(fresh.composition).baselineQuery.toMs,q.toMs);assert.equal(q.fromMs,context.readWindow.fromMs);for(const edit of [x=>x.fromMs--,x=>x.toMs++,x=>x.environmentSelector.policy=x.environmentSelector.policy==='effective'?'current-only':'effective']){const bad=structuredClone(composition);edit(C.comparison(bad).baselineQuery);assert.throws(()=>C.validate(bad,q))}process.stdout.write(JSON.stringify(view));`
		cmd := exec.Command("node", "--input-type=module", "-e", script)
		cmd.Stdin = bytes.NewReader(input)
		out, err := cmd.CombinedOutput()
		if err != nil {
			tc.Fatal(err, string(out))
			return
		}
		var view workbench.ExplorationView
		if err = json.Unmarshal(out, &view); err != nil {
			tc.Fatal(err, string(out))
			return
		}
		if err = view.Validate(); err != nil {
			tc.Fatal(err)
			return
		}
		if _, err = exploration.Compile(view.Queries[0]); err != nil {
			tc.Fatal(err)
			return
		}
		for _, child := range view.Composition.Children {
			if child.Comparison != nil {
				if _, err = exploration.Compile(*child.Comparison.BaselineQuery); err != nil {
					tc.Fatal(err)
					return
				}
			}
		}
	}, options(30)...)
}
