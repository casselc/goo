/* Host presentation defaults over the shared passive resource renderer. */
import * as ResourceNativeView from './resource-native-view.mjs';
const global=globalThis;
const view=ResourceNativeView,api={...view,render:(host,widget,result,options={})=>view.render(host,widget,result,{compact:true,...options})};
const byteSI=api.byteSI;
const byteRate=api.byteRate;
const invalidValue=api.invalidValue;
const timeline=api.timeline;
const display=api.display;
const render=api.render;
export {byteSI,byteRate,invalidValue,timeline,display,render};
