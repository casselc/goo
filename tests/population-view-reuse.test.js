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
