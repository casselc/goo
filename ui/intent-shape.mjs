/* Closed supplied configuration; no coercion, normalization or source access. */
export const bytes=value=>new TextEncoder().encode(value).length;
export function object(value,keys,message='Unsupported intent fields.'){
 if(!value||typeof value!=='object'||Array.isArray(value)||![Object.prototype,null].includes(Object.getPrototypeOf(value))||Reflect.ownKeys(value).some(key=>typeof key!=='string'||!keys.includes(key)||!Object.hasOwn(Object.getOwnPropertyDescriptor(value,key),'value')||!Object.getOwnPropertyDescriptor(value,key).enumerable))throw Error(message);
 return value;
}
export function text(value,max,{empty=true}={}){return typeof value==='string'&&value.isWellFormed()&&(empty||value.length>0)&&bytes(value)<=max&&!value.includes('\0');}

export function array(value){
 if(!Array.isArray(value)||Object.getPrototypeOf(value)!==Array.prototype||Reflect.ownKeys(value).length!==value.length+1)throw Error('Use a dense intent array.');
 for(let i=0;i<value.length;i++){const d=Object.getOwnPropertyDescriptor(value,String(i));if(!d||!Object.hasOwn(d,'value')||!d.enumerable)throw Error('Use a dense intent array.');}
 return value;
}
