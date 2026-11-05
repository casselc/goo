const {test}=require('node:test'),assert=require('node:assert/strict'),{mount}=require('../ui/attributes.mjs');
class Element{
 constructor(tag){this.tagName=tag;this.children=[];this.dataset={};this.value='';this.textContent='';this.attributes={};}
 append(...children){this.children.push(...children);if(this.tagName==='select'&&!this.value)this.value=children.find(child=>child?.value)?.value||''}
 prepend(...children){this.children.unshift(...children)}
 replaceChildren(...children){this.children=children}
 setAttribute(key,value){this.attributes[key]=value}getAttribute(key){return this.attributes[key]}
}
const descendants=node=>[node,...node.children.filter(child=>child instanceof Element).flatMap(descendants)];
test('attribute inspection reports host refusal truthfully while ordinary void/true callbacks stage normally',()=>{
 const previous=global.document;global.document={createElement:tag=>new Element(tag)};
 try{for(const accepted of [false,undefined,true]){let observed=null;const host=new Element('section');mount(host,{Attributes:{'http.route':'/checkout'}},{raw:false,onFacet:(field,value,mode)=>{observed={field,value,mode};return accepted}});const all=descendants(host),button=all.find(node=>node.getAttribute('aria-label')==='Include attribute:http.route from record'),status=all.find(node=>node.className==='attribute-feedback');assert(button);assert.equal(button.disabled,false);button.onclick();assert.deepEqual(observed,{field:{kind:'attribute',key:'http.route'},value:'/checkout',mode:'include'});assert.equal(status.textContent,accepted===false?'Not staged.':'Staged include for attribute:http.route. Apply or Open in the selection tray; displayed measurements are unchanged.');}}
 finally{global.document=previous}
});
