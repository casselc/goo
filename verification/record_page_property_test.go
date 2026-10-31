package verification

import (
	"hegel.dev/go/hegel"
	"testing"
)

const recordPageAdmissionJS = `import assert from 'node:assert/strict';import fs from 'node:fs';import * as S from '../ui/source-snapshots.mjs';const {seed,offset,limit}=JSON.parse(fs.readFileSync(0,'utf8'));const q={signal:seed%2?'spans':'logs',environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:seed%3?{mode:'named',value:'prod'}:{mode:'unreported',value:''}},fromMs:1000,toMs:61000,limit,offset,groupBy:[{kind:'column',key:'service'}],bucketMs:1000};const rows=Array.from({length:Math.min(limit,seed%5+1)},(_,i)=>({timestamp_ns:String(1000000000+i),...(q.signal==='spans'?{duration_ns:'0'}:{}),attributes:{empty:'',cohort:String(seed)}}));const data={question:q,basis:'a'.repeat(64),capturedAt:'2026-10-31T00:00:00Z',count:String(offset+rows.length+1),state:'observed',rows,hasMore:true,resultToken:'b'.repeat(64),warnings:['independent reads']};const admitted=S.adoptRecordPage(data,q),page=S.getRecordPage(admitted);assert.equal(page.quality.extent,'bounded-page');assert.equal(page.quality.upstream,'unknown');assert.equal(page.page.offset,offset);assert.equal(page.page.returned,rows.length);assert.equal(page.page.hasMore,true);assert.equal(admitted.count,data.count);assert.equal(admitted.resultToken,data.resultToken);assert.equal(admitted.viewDataset,undefined);assert.equal(admitted.timeDataset,undefined);assert.equal(page.complete,undefined);const originalLength=rows.length;data.rows.push({timestamp_ns:'1'});assert.equal(page.rows.length,originalLength);const bad={...data,rows:[]};assert.throws(()=>S.adoptRecordPage({...bad,viewDataset:{}},q));assert.throws(()=>S.adoptRecordPage({...bad,count:'18446744073709551616'},q));assert.throws(()=>S.adoptRecordPage({...bad,question:{...q,offset:(offset+1)%1001}},q));assert.throws(()=>S.adoptRecordPage({...bad,resultToken:'held'},q));`

func TestHegelRecordPageAdmissionKeepsBoundedIndependentMeaning(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 1000))
		offset := hegel.Draw(tc, hegel.Integers(0, 1000))
		limit := hegel.Draw(tc, hegel.Integers(1, 20))
		if err := runNativeHistogramJS(map[string]any{"seed": seed, "offset": offset, "limit": limit}, recordPageAdmissionJS); err != nil {
			tc.Fatal(err)
		}
	}, options(100)...)
}
