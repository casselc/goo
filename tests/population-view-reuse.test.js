const {test}=require('node:test'),assert=require('node:assert/strict');
const {reuseRequestCurrent}=require('../ui/population-workspace.mjs');
const context=()=>({active:true,visible:true,generation:3,draftRevision:7,libraryGeneration:2,serial:4,questionKey:'checked destination'});
test('an inspected display belongs to one exact destination draft and catalog candidate',()=>{
 const expected=context();assert.equal(reuseRequestCurrent(expected,context()),true);
 for(const change of [{active:false},{visible:false},{generation:4},{draftRevision:8},{libraryGeneration:3},{serial:5},{questionKey:'another destination'}])assert.equal(reuseRequestCurrent(expected,{...context(),...change}),false,JSON.stringify(change));
 assert.deepEqual(expected,context());
});
test('out-of-order configuration replies cannot replace a newer compatible candidate',async()=>{
 let releaseA,releaseB;const a=new Promise(resolve=>releaseA=resolve),b=new Promise(resolve=>releaseB=resolve),publications=[];let current=context();
 const first={...current},publish=async(promise,expected)=>{const value=await promise;if(reuseRequestCurrent(expected,current))publications.push(value)};
 const older=publish(a,first);current={...current,serial:5};const newer=publish(b,{...current});releaseB('newer display');await newer;releaseA('older display');await older;
 assert.deepEqual(publications,['newer display']);
 // Even a response with the same query text loses ownership after a draft edit.
 const sameQuery={...current};current={...current,draftRevision:current.draftRevision+1};assert.equal(reuseRequestCurrent(sameQuery,current),false);
});

const {compositionReuseReason}=require('../ui/population-workspace.mjs');
test('whole-composition application waits for one fixed owner, explicit tail removal and page completion',()=>{
 assert.equal(compositionReuseReason({queryCount:1}), '');
 for(const queryCount of [0,2])assert.match(compositionReuseReason({queryCount}),/one current query/);
 assert.match(compositionReuseReason({queryCount:1,tailConfigured:true}),/Clear live configuration/);
 assert.match(compositionReuseReason({queryCount:1,pagePending:true}),/record page to finish/);
 const state={queryCount:1,tailConfigured:true,pagePending:true};compositionReuseReason(state);assert.deepEqual(state,{queryCount:1,tailConfigured:true,pagePending:true});
});


test('chosen destination baseline is part of the private preview lease',()=>{
 const expected={active:true,visible:true,generation:1,draftRevision:2,libraryGeneration:3,serial:4,questionKey:'same target',baselineCapture:'a'.repeat(64)};
 assert.equal(reuseRequestCurrent(expected,{...expected}),true);assert.equal(reuseRequestCurrent(expected,{...expected,baselineCapture:'b'.repeat(64)}),false);assert.equal(reuseRequestCurrent(expected,{...expected,baselineCapture:''}),false);
});

const {selectionCommandCurrent}=require('../ui/population-workspace.mjs');
test('every selection command belongs to one tray owner, draft and configured question snapshot',()=>{
 const owner={},expected={owner,generation:7,draftRevision:9,selectionKey:'selected-values',questionKey:'fixed-query'},actual={...expected,active:true,visible:true,connected:true,child:false,busy:false,countPending:false,sourceDraft:false,queryDraft:false,recordPending:false,pagePending:false};assert.equal(selectionCommandCurrent(expected,actual),true);
 for(const change of [{owner:{}},{generation:8},{draftRevision:10},{selectionKey:'other'},{questionKey:'new-window'},{active:false},{visible:false},{connected:false},{child:true},{busy:true},{countPending:true},{sourceDraft:true},{queryDraft:true},{recordPending:true},{pagePending:true}])assert.equal(selectionCommandCurrent(expected,{...actual,...change}),false,JSON.stringify(change));
 // Toggling away and back to identical values still issues a different owner.
 assert.equal(selectionCommandCurrent(expected,{...actual,owner:{}}),false);assert.equal(selectionCommandCurrent(expected,actual),true);
});
