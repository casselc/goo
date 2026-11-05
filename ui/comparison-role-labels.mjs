/* Passive operand names; labels confer no acquisition or evidence authority. */
export function comparisonRoleLabels(value){
 if(value===undefined)return Object.freeze(['Operand A','Operand B']);
 if(!Array.isArray(value)||Object.getPrototypeOf(value)!==Array.prototype||value.length!==2||Reflect.ownKeys(value).length!==3)throw TypeError('Supply two plain operand labels.');
 const labels=[];for(let index=0;index<2;index++){const d=Object.getOwnPropertyDescriptor(value,index);if(!d?.enumerable||!Object.hasOwn(d,'value')||typeof d.value!=='string'||!d.value.trim()||!d.value.isWellFormed()||new TextEncoder().encode(d.value).length>120||/\p{Cc}/u.test(d.value))throw TypeError('Supply bounded operand labels.');labels.push(d.value)}return Object.freeze(labels);
}
