package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

// The generated action and resize are checked independently of source acquisition.
// Adjacent windows are fixed recipes, never retained observations or rate claims.
func TestHegelPreviousWindowGoJS321(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		width := int64(hegel.Draw(tc, hegel.Integers(1, 120))) * 1000
		from := int64(1790000000000) + int64(seed)*1000
		mode := hegel.Draw(tc, hegel.SampledFrom([]string{"current", "named", "empty", "unreported"}))
		scope := identity.EnvironmentScope{Mode: mode}
		if mode == "current" || mode == "named" {
			scope.Mode = "named"
			scope.Value = fmt.Sprintf("prod-%d'雪", seed%11)
		}
		sel, e := identity.EffectiveEnvironment(scope)
		if mode == "current" {
			sel, e = identity.CurrentEnvironment(scope.Value)
		}
		if e != nil {
			tc.Fatal(e)
			return
		}
		wire, _ := sel.Selector()
		ns := identity.Namespace{Mode: "missing"}
		if seed%3 != 0 {
			ns.Mode = "value"
			if seed%3 == 2 {
				ns.Value = "team'雪"
			}
		}
		signal := "logs"
		if seed%2 == 0 {
			signal = "spans"
		}
		q := exploration.Query{Signal: signal, EnvironmentSelector: &wire, FromMS: from, ToMS: from + width, Limit: 100, GroupBy: []exploration.Field{{Kind: "attribute", Key: "route"}}, Where: &exploration.Predicate{All: []exploration.Predicate{*exploration.NamespacePredicate(&ns), {Field: &exploration.Field{Kind: "attribute", Key: "tenant"}, Op: "eq", Value: fmt.Sprint(seed)}}}}
		input, _ := json.Marshal(map[string]any{"q": q, "range": map[string]int64{"fromMs": from + 10000, "toMs": from + 10000 + width/2 + 1}})
		script := `import fs from 'node:fs';import assert from 'node:assert/strict';import * as C from '../ui/view-composition.mjs';import * as P from '../ui/captured-count-comparison.mjs';
 const {q,range}=JSON.parse(fs.readFileSync(0,'utf8'));const scope={kind:'record-query',signal:q.signal},a={version:'goo.view/1',scope,kind:{name:'bar',version:1},display:{category:'category',value:'count',title:'Literal title'}};const c=C.create(q,a,{columns:['utc','summary'],fieldColumns:[{kind:'resource',key:'service.namespace'}]}),before=JSON.stringify({q,c});
 const next={query:structuredClone(q),composition:C.setComparison(c,P.createQuery(q,P.previousQuery(q)),q)},pair=C.comparison(next.composition),width=q.toMs-q.fromMs;
 assert.equal(pair.baselineQuery.fromMs,q.fromMs-width);assert.equal(pair.baselineQuery.toMs,q.fromMs);assert.deepEqual(next.query,q);
 const strip=q=>{q=structuredClone(q);delete q.fromMs;delete q.toMs;q.offset=0;return q};assert.deepEqual(strip(pair.baselineQuery),strip(q));
 const rebased=C.rebaseFixedComparison(next.composition,next.query,range),b=C.comparison(rebased.composition).baselineQuery;assert.equal(b.toMs,range.fromMs);assert.equal(b.fromMs,range.fromMs-(range.toMs-range.fromMs));assert.deepEqual(strip(b),strip(q));assert.deepEqual(C.records(rebased.composition),C.records(c));assert.deepEqual(C.aggregate(rebased.composition),C.aggregate(c));
 for(const edit of [b=>b.fromMs--,b=>b.toMs++,b=>b.environmentSelector.policy=b.environmentSelector.policy==='effective'?'current-only':'effective',b=>b.groupBy[0].key='other',b=>b.numeric={field:{kind:'attribute',key:'n'}}]){const bad=structuredClone(next.composition);edit(C.comparison(bad).baselineQuery);assert.throws(()=>C.validate(bad,q));}
 const early={...q,fromMs:width-1,toMs:2*width-1};assert.throws(()=>P.previousQuery(early));assert.throws(()=>C.rebaseFixedComparison(next.composition,q,{fromMs:1,toMs:width+2}));assert.equal(JSON.stringify({q,c}),before);
 const views=[next,rebased].map((v,i)=>({version:1,title:'Previous '+i,queries:[v.query],composition:v.composition}));const original=pair.baselineQuery.where.all[1].value;pair.baselineQuery.where.all[1].value='changed';assert.equal(JSON.stringify({q,c}),before);assert.equal(C.comparison(rebased.composition).baselineQuery.where.all[1].value,original);pair.baselineQuery.where.all[1].value=original;process.stdout.write(JSON.stringify(views));`
		cmd := exec.Command("node", "--input-type=module", "-e", script)
		cmd.Stdin = bytes.NewReader(input)
		out, e := cmd.CombinedOutput()
		if e != nil {
			tc.Fatal(e, string(out))
			return
		}
		var views []workbench.ExplorationView
		if e = json.Unmarshal(out, &views); e != nil {
			tc.Fatal(e, string(out))
			return
		}
		for _, view := range views {
			if e = view.Validate(); e != nil {
				tc.Fatal(e)
				return
			}
			artifact := workbench.ViewArtifact{Version: workbench.ViewArtifactVersion, Namespace: "team", Name: fmt.Sprintf("previous-%d", seed), Revision: "1", View: view}
			encoded, err := workbench.EncodeViewArtifact(artifact)
			if err != nil {
				tc.Fatal(err)
				return
			}
			decoded, err := workbench.DecodeViewArtifact(encoded.YAML)
			if err != nil || decoded.YAML != encoded.YAML {
				tc.Fatal("portable pair changed", err)
				return
			}
			for _, child := range view.Composition.Children {
				if child.Comparison == nil {
					continue
				}
				for _, change := range []func(*exploration.Query){
					func(b *exploration.Query) { b.FromMS-- },
					func(b *exploration.Query) { b.ToMS++ },
					func(b *exploration.Query) {
						b.EnvironmentSelector.Scope.Value = "foreign"
						b.EnvironmentSelector.Scope.Mode = "named"
					},
					func(b *exploration.Query) { b.GroupBy[0].Key = "other" },
				} {
					raw, _ := json.Marshal(child.Comparison)
					var bad workbench.CountComparisonViewDeclaration
					if err := json.Unmarshal(raw, &bad); err != nil {
						tc.Fatal(err)
						return
					}
					change(bad.BaselineQuery)
					if bad.Validate(view.Queries[0]) == nil {
						tc.Fatal("invalid window/policy/profile baseline admitted")
						return
					}
				}
			}
		}
	}, options(30)...)
}
