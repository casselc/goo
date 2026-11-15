import {test} from 'node:test';
import assert from 'node:assert/strict';
import {Element} from './support/raw-observation-dom.mjs';
import {mount} from '../ui/filter-editor.mjs';
class Native extends Element {
 before(...nodes){for(const node of nodes)this.parentElement?.insertBefore(node,this)}
}
test('450 shared editor exposes counted-value actions only while a source owner is registered',()=>{
 globalThis.Node=Native;const body=new Native('body');body.root=true;globalThis.document={body,hidden:false,activeElement:null,createElement:tag=>new Native(tag)};
 const host=new Native('section');body.append(host);const field={kind:'column',key:'service'},where={field,op:'eq',value:'worker'},loads=[];
 const editor=mount(host,{where}),buttons=()=>host.querySelectorAll('button').filter(b=>b.textContent==='Load values for this scope');
 assert.equal(buttons().length,1);assert.equal(buttons()[0].hidden,true);buttons()[0].onclick();assert.equal(loads.length,0);
 editor.setValueLoader(value=>loads.push(value));assert.equal(buttons()[0].hidden,false);buttons()[0].onclick();assert.deepEqual(loads,[field]);
 editor.set({field,op:'exists'});assert.equal(buttons()[0].hidden,false);buttons()[0].onclick();assert.equal(loads.length,2);
 editor.setValueLoader(null);assert.equal(buttons()[0].hidden,true);buttons()[0].onclick();assert.equal(loads.length,2);editor.dispose();
});
