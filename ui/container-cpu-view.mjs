/* CPU is an exported fractional-core gauge, not a counter rate or capacity verdict. */
import * as ContainerMemoryView from './container-memory-view.mjs';
const global=globalThis;
const base=ContainerMemoryView;
function validValue(value){return typeof value==='string'&&value.length<=512&&/^(0|[1-9][0-9]*)(\.[0-9]*[1-9])?$/.test(value)&&Number.isFinite(Number(value))&&(value==='0'||Number(value)>0)}
function format(value){if(!validValue(value))throw Error('Invalid exported CPU cores.');const [whole,part='']=value.split('.');if(whole.length<=7&&part.length<=6)return value+' cores';const digits=(whole+part).replace(/^0+/,'');if(!digits)return '0 cores';const exponent=whole==='0'?-part.search(/[1-9]/)-1:whole.length-1,tail=digits.slice(1,4).replace(/0+$/,'');return (digits.length>4?'≈':'')+digits[0]+(tail?'.'+tail:'')+'e'+exponent+' cores'}
const api=base.create({family:'cpu',label:'CPU',primary:'usage',primaryKey:'usage',primaryLabel:'Usage',unit:'{cpu}',heading:'Reported CPU cores',roleData:'cpuRole',validValue,format,percentLimit:1024,scaleNote:'Shared core scale; independently exported observations are not an aligned snapshot. Usage is already averaged by the receiver, not a counter rate, saturation or health.'});
Object.assign(api,{validValue,format});
const roles=api.roles;
const identity=api.identity;
const validate=api.validate;
const percent=api.percent;
const render=api.render;
const exportedValidValue=api.validValue;
const exportedFormat=api.format;
export {roles,identity,validate,percent,render,exportedValidValue as validValue,exportedFormat as format};
