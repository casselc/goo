package verification

import (
	"hegel.dev/go/hegel"
	"testing"
)

const recordInteractionsStateJS = `import assert from 'node:assert/strict';import fs from 'node:fs';import {createSelection} from '../ui/record-interactions.mjs';
const {size,actions}=JSON.parse(fs.readFileSync(0,'utf8')),owner=createSelection(size),expected=new Set();let busy=false,submitted=false,revision=0;
assert(Object.isFrozen(owner));for(const invalid of [-1,0.5,NaN,Infinity,Number.MAX_SAFE_INTEGER+1,'2',null])assert.throws(()=>createSelection(invalid));
for(const {op,index} of actions){const before=owner.indexes(),lease=owner.revision();if(op<=1){const checked=op===0,allowed=Number.isInteger(index)&&index>=0&&index<size&&!busy&&!submitted&&(!checked||expected.has(index)||expected.size<20);assert.equal(owner.set(index,checked),allowed);if(allowed){if(checked)expected.add(index);else expected.delete(index);revision++}else assert.deepEqual(owner.indexes(),before)}else if(op===2){assert.equal(owner.clear(),!busy);if(!busy){expected.clear();submitted=false;revision++}}else{busy=op===3;submitted=op===4;owner.submission({busy,submitted});assert.equal(owner.revision(),lease,'submission state must preserve in-flight revision lease')}
assert.deepEqual(owner.indexes(),[...expected].sort((a,b)=>a-b));assert(Object.isFrozen(owner.indexes()));assert.equal(owner.revision(),revision);assert.deepEqual(owner.state(),{selected:expected.size,submitted,busy});assert(Object.isFrozen(owner.state()));assert(expected.size<=20);for(let i=0;i<size;i++)assert.equal(owner.has(i),expected.has(i));if(revision!==lease)assert.notEqual(owner.revision(),lease,'accepted membership mutation retires the old revision lease');}
// Occurrences, not object equality, are selected; duplicate telemetry is not a unique event ID.
const repeated={SpanId:'1'.repeat(16)},rows=[repeated,repeated],duplicates=createSelection(rows.length);assert(duplicates.set(1,true));assert.deepEqual(duplicates.indexes(),[1]);assert(duplicates.set(0,true));assert.deepEqual(duplicates.indexes(),[0,1]);
const capped=createSelection(25);for(let i=0;i<20;i++)assert(capped.set(i,true));assert.equal(capped.set(20,true),false);assert(capped.set(19,false));assert(capped.set(20,true));assert.equal(capped.state().selected,20);
`

func TestHegelRecordInteractionSelectionStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		size := hegel.Draw(tc, hegel.Integers(0, 80))
		length := hegel.Draw(tc, hegel.Integers(20, 100))
		mask := hegel.Draw(tc, hegel.Integers(1, 63))
		enabled := []int{}
		for op := 0; op < 6; op++ {
			if mask&(1<<op) != 0 {
				enabled = append(enabled, op)
			}
		}
		actions := make([]map[string]int, length)
		for i := range actions {
			actions[i] = map[string]int{"op": enabled[hegel.Draw(tc, hegel.Integers(0, len(enabled)-1))], "index": hegel.Draw(tc, hegel.Integers(-2, size+2))}
		}
		if e := runNativeHistogramJS(map[string]any{"size": size, "actions": actions}, recordInteractionsStateJS); e != nil {
			tc.Fatal(e)
		}
	}, options(100)...)
}
