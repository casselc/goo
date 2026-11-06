/* Closed source-read diagnostics. Hosts retain query ownership and explicit retries. */
export const reasons=Object.freeze({cancelled:'Source read cancelled.',timeout:'Source read timed out.',budget:'Source read exceeded a limit.',malformed:'Source response rejected.',unavailable:'Source unavailable.'});
const object=value=>value!==null&&typeof value==='object'&&!Array.isArray(value);
export function refusal(value){
 if(!object(value)||Object.keys(value).some(key=>!['kind','detail'].includes(key))||!Object.hasOwn(reasons,value.kind)||typeof value.detail!=='string'||!value.detail.trim()||!value.detail.isWellFormed()||new TextEncoder().encode(value.detail).length>256||/\p{Cc}/u.test(value.detail))throw Error('Unsupported source refusal.');
 return {kind:value.kind,detail:value.detail};
}
export class SourceReadError extends Error{
 constructor(value){const checked=refusal(value);super(reasons[checked.kind]);this.name='SourceReadError';this.refusal=checked;}
}
export function responseError(reply,fallback='Read unavailable.'){
 const data=reply?.data;
 if(reply?.status===502&&data?.refusal!==undefined){try{const checked=refusal(data.refusal);if(data.error!==reasons[checked.kind])throw Error('Unrecognized refusal reason.');return new SourceReadError(checked)}catch{return new SourceReadError({kind:'unavailable',detail:'The source refusal response was not recognized.'})}}
 if(reply?.status>=500)return new SourceReadError({kind:'unavailable',detail:'The read did not return a recognized source response.'});
 return Error(typeof data?.error==='string'&&data.error?data.error:fallback);
}
export function failure(error){return {error:error?.message||'Read unavailable.',...(error instanceof SourceReadError?{refusal:refusal(error.refusal)}:{})};}
