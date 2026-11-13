/* Exact versioned record identity; source classification remains authoritative. */
export const field=Object.freeze({kind:'column',key:'exception_fingerprint_v1'});
export const reasons=Object.freeze({none:'No named exception event','both-missing':'Type and stack not reported','missing-type':'Type not reported','missing-stack':'Stack not reported','both-empty':'Type and stack present empty','empty-type':'Type present empty','empty-stack':'Stack present empty',multiple:'Conflicting exception event tuples',oversize:'Exception input exceeds bound','invalid-utf8':'Invalid UTF-8 exception input'});
export function decode(value){
 if(typeof value!=='string')throw Error('Choose an exact exception fingerprint v1.');
 if(/^ef1:c:[a-f0-9]{64}$/.test(value))return {id:value,classified:true,state:'classified',reason:'',label:'Exact reported type + stack'};
 const reason=value.startsWith('ef1:u:')?value.slice(6):'';
 if(Object.hasOwn(reasons,reason))return {id:value,classified:false,state:'unclassified',reason,label:reasons[reason]};
 throw Error('Invalid exception fingerprint v1.');
}
