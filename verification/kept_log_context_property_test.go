package verification

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

func TestHegelKeptLogContext328(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 240000))
		mode := hegel.Draw(tc, hegel.SampledFrom([]string{"legacy", "current", "named", "empty", "unreported"}))
		fixtures, e := keptContinuationFixtures(seed, mode, "logs", false)
		if e != nil {
			tc.Fatal(e)
			return
		}
		spans, e := keptContinuationFixtures(seed, mode, "spans", false)
		if e != nil {
			tc.Fatal(e)
			return
		}
		input, _ := json.Marshal(map[string]any{"values": fixtures, "spans": spans})
		cmd := exec.Command("node", "--input-type=module", "-e", keptLog328JS)
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
		if len(views) != 24 {
			tc.Fatal("lost member/origin contexts")
			return
		}
		for _, v := range views {
			if e = v.Validate(); e != nil {
				tc.Fatal(e)
				return
			}
			q := v.Queries[0]
			paged := q
			paged.Offset = 100
			if _, err := exploration.Compile(paged); err != nil {
				tc.Fatal(err)
				return
			}
			plan, compileErr := exploration.Compile(q)
			if compileErr != nil {
				e = compileErr
				tc.Fatal(e)
				return
			}
			// Compilation is admission and SQL generation, not execution against a source.
			// Require resource-key presence in SQL so present-empty cannot become missing.
			if !strings.Contains(plan.Rows, "mapContains(ResourceAttributes,") || !strings.Contains(plan.Rows, "ServiceName = {") {
				tc.Fatal("compiled context lost qualified membership")
				return
			}
			if q.Signal != "logs" || q.Offset != 0 || len(q.Where.All) != 2 || q.Where.All[0].Field.Key != "service" || q.Where.All[1].Field.Kind != "resource" || q.Where.All[1].Field.Key != "service.namespace" {
				tc.Fatal("context widened its service identity")
				return
			}
		}
	}, options(30)...)
}

const keptLog328JS = `
import assert from 'node:assert/strict';import fs from 'node:fs';import * as K from '../ui/kept-saved.mjs';import * as D from '../ui/discovery.mjs';import * as E from '../ui/record-environment.mjs';import * as C from '../ui/view-composition.mjs';import * as S from '../ui/question-session.mjs';
const {values,spans}=JSON.parse(fs.readFileSync(0,'utf8')),views=[];
for(const value of values){const before=JSON.stringify(value),tail=value.version==='goo/kept-observations/v1',origin=tail?value.selector:value.question;
for(let index=0;index<2;index++){const m=value.members[index],r=m.record,resource=r.ResourceAttributes,from=tail?m.fromMs:origin.fromMs,to=tail?m.toMs:origin.toMs,center=Number(BigInt(r.timestamp_ns)/1000000n),v=K.surroundingView(value,index),q=v.queries[0],has=Object.hasOwn(resource,'service.namespace'),ns=resource['service.namespace'];assert.equal(q.signal,'logs');assert.deepEqual(E.questionFields(q),E.questionFields(origin));assert.equal(q.fromMs,Math.max(from,center-30000));assert.equal(q.toMs,Math.min(to,center+30001));assert(q.fromMs>=from&&q.toMs<=to);assert(BigInt(q.fromMs)*1000000n<=BigInt(r.timestamp_ns)&&BigInt(r.timestamp_ns)<BigInt(q.toMs)*1000000n);assert.equal(q.offset,0);assert.equal(q.limit,100);assert.deepEqual(q.where,{all:[{field:{kind:'column',key:'service'},op:'eq',value:r.ServiceName},{field:{kind:'resource',key:'service.namespace'},op:has?'eq':'missing',...(has?{value:ns}:{})}]});for(const key of ['groupBy','summary','spanStatus','numeric','numericHistogram','bucketMs','groupTimeSeries','reportedEdge'])assert(!q[key]);assert(C.recordOnly(v.composition));assert.equal(C.records(v.composition).kind.name,'record-table');assert(!r.TraceId,'fixture must really be untraced');views.push(v);
const matches=(service,attrs)=>service===q.where.all[0].value&&(q.where.all[1].op==='missing'?!Object.hasOwn(attrs,'service.namespace'):Object.hasOwn(attrs,'service.namespace')&&attrs['service.namespace']===q.where.all[1].value);assert(matches(r.ServiceName,resource));assert(!matches('different-service',resource));for(const other of [undefined,'','different-namespace']){const attrs={...resource};if(other===undefined)delete attrs['service.namespace'];else attrs['service.namespace']=other;assert.equal(matches(r.ServiceName,attrs),has?other===ns:other===undefined)}
const fixed={...origin,fromMs:from,toMs:to};assert.deepEqual(D.surroundingQuestion(fixed,r),q);
for(const radius of [1000,30000,300000]){const direct=D.surroundingQuestion(fixed,r,radius,100);assert.equal(direct.fromMs,Math.max(from,center-radius));assert.equal(direct.toMs,Math.min(to,center+radius+1));assert.equal(direct.offset,100);assert.equal(direct.limit,100);assert.deepEqual(direct.where,q.where);assert.deepEqual(E.questionFields(direct),E.questionFields(origin));const admitted=structuredClone(v);admitted.queries=[{...direct,offset:0}];views.push(admitted);}
for(const change of [r=>delete r.ServiceName,r=>r.ServiceName='',r=>r.ServiceName=42,r=>delete r.ResourceAttributes,r=>r.ResourceAttributes=null,r=>r.ResourceAttributes=[],r=>r.ResourceAttributes='unknown',r=>r.ResourceAttributes['service.namespace']=null,r=>r.ResourceAttributes['service.namespace']=42,r=>r.timestamp_ns=String(BigInt(to)*1000000n)]){const bad=structuredClone(r);change(bad);assert.throws(()=>D.surroundingQuestion(fixed,bad));}
for(const radius of [0,999,300001,NaN,1000.5,'1000'])assert.throws(()=>D.surroundingQuestion(fixed,r,radius));for(const offset of [-1,1001,0.5,'0',NaN])assert.throws(()=>D.surroundingQuestion(fixed,r,30000,offset));
const directDetached=D.surroundingQuestion(fixed,r);directDetached.where.all[0].value='changed';assert.equal(JSON.stringify(value),before);assert.deepEqual(D.surroundingQuestion(fixed,r),q);
const independent=K.surroundingView(value,index);q.where.all[0].value='changed';assert.equal(JSON.stringify(value),before);assert.equal(independent.queries[0].where.all[0].value,r.ServiceName);q.where.all[0].value=r.ServiceName;
for(const change of [r=>delete r.ServiceName,r=>r.ServiceName='',r=>delete r.ResourceAttributes,r=>r.ResourceAttributes=null,r=>r.ResourceAttributes=[],r=>r.ResourceAttributes['service.namespace']=null,r=>r.ResourceAttributes['service.namespace']=42,r=>delete r.timestamp_ns,r=>r.timestamp_ns='not-a-time',r=>r.timestamp_ns=String(BigInt(to)*1000000n)]){const bad=structuredClone(value);change(bad.members[index].record);assert.throws(()=>K.surroundingView(bad,index));}
}
for(const index of [-1,2,0.5,'0',null])assert.throws(()=>K.surroundingView(value,index));const foreign=structuredClone(value);foreign.members[0].record.ResourceAttributes['deployment.environment.name']='foreign';assert.throws(()=>K.surroundingView(foreign,0));const edited=structuredClone(value),planned=K.surroundingView(edited,0),expected=S.identity(planned.queries[0]);edited.members[0].record.ServiceName='changed';if(tail)edited.members[0].fromMs--;else edited.question.fromMs--;assert.equal(S.identity(planned.queries[0]),expected);assert.equal(JSON.stringify(value),before);
}
for(const value of spans)assert.throws(()=>K.surroundingView(value,0));process.stdout.write(JSON.stringify(views));
`
