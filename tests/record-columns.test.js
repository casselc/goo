const{test}=require('node:test'),assert=require('node:assert/strict'),{fields,value,validate}=require('../ui/record-columns.mjs');
test('qualified supplied fields retain absence, empty strings and exact integer text',()=>{const r={Body:'log',ResourceAttributes:{region:'resource'},LogAttributes:{region:'',large:'9007199254740993'}};assert.equal(fields([r]).filter(f=>f.key==='region').length,2);assert.deepEqual(value(r,{kind:'attribute',key:'region'}),{present:true,value:''});assert.deepEqual(value(r,{kind:'attribute',key:'absent'}),{present:false});assert.equal(value(r,{kind:'attribute',key:'large'}).value,'9007199254740993');assert.equal(value(r,{kind:'resource',key:'region'}).value,'resource')});
test('display layout detaches ordered fields and preserves required core columns',()=>{const v={columns:['summary','utc'],fieldColumns:[{kind:'resource',key:'x'},{kind:'attribute',key:'x'}]},out=validate(v);out.fieldColumns[0].key='changed';assert.equal(v.fieldColumns[0].key,'x');for(const layout of [{columns:['utc']},{columns:['utc','summary','utc']},{columns:['utc','summary'],fieldColumns:[{kind:'attribute',key:'x'},{kind:'attribute',key:'x'}]},{columns:['utc','summary'],fieldColumns:Array.from({length:9},(_,i)=>({kind:'attribute',key:String(i)}))}])assert.throws(()=>validate(layout))});


test('qualified record columns preserve well-formed Unicode under the Go UTF-8 byte bound',()=>{
 const C=require('../ui/record-columns.mjs'),field={kind:'attribute',key:'team.🚀'};
 assert.deepEqual(C.validate({columns:['utc','summary'],fieldColumns:[field]}).fieldColumns,[field]);
 assert(C.fields([{SpanAttributes:{'team.🚀':'launch'}}]).some(f=>f.kind==='attribute'&&f.key==='team.🚀'));
 assert.deepEqual(C.validate({columns:['utc','summary'],fieldColumns:[{kind:'attribute',key:'🚀'.repeat(32)}]}).fieldColumns[0].key,'🚀'.repeat(32));
 for(const key of ['\ud800','x\0y','🚀'.repeat(33)])assert.throws(()=>C.validate({columns:['utc','summary'],fieldColumns:[{kind:'attribute',key}]}));
});

test('shared qualified-field presentation preserves time units and missing/empty/null distinctions',()=>{
 const C=require('../ui/record-columns.mjs'),record={timestamp_ns:'1793448000123456789',duration_ns:'9000000',Attributes:{x:'',nil:null,nested:{count:0}},ResourceAttributes:{x:'resource'}};
 assert.deepEqual(C.value(record,{kind:'column',key:'timestamp_ns'}),{present:true,value:record.timestamp_ns});
 assert.deepEqual(C.presentation(record,{kind:'column',key:'timestamp_ns'}),{state:'present',type:'string',text:'2026-10-31T12:00:00.123456789Z',title:'1793448000123456789 ns'});
 assert.equal(C.presentation(record,{kind:'column',key:'duration_ns'}).text,'9.000 ms');
 assert.equal(C.presentation(record,{kind:'attribute',key:'x'}).state,'empty');assert.equal(C.presentation(record,{kind:'resource',key:'x'}).text,'resource');
 assert.equal(C.presentation(record,{kind:'attribute',key:'nil'}).state,'null');assert.equal(C.presentation(record,{kind:'resource',key:'nil'}).state,'missing');
 assert.equal(C.presentation(record,{kind:'attribute',key:'nested'}).text,'{"count":0}');
 assert.equal(C.presentation({timestamp_ns:'invalid'},{kind:'column',key:'timestamp_ns'}).text,'Timestamp unavailable');assert.equal(C.presentation({timestamp_ns:'9'.repeat(33)},{kind:'column',key:'timestamp_ns'}).text,'Timestamp unavailable');
 assert.equal(C.presentation({duration_ns:'-1'},{kind:'column',key:'duration_ns'}).text,'Duration unavailable');
});


test('literal state words and JSON-looking strings stay distinct from typed or missing values',()=>{
 const C=require('../ui/record-columns.mjs'),field={kind:'attribute',key:'v'},p=v=>C.presentation({Attributes:{v}},field);
 for(const value of ['Missing','null','Null','Present empty string','true','false','0','{"count":0}','"Missing"']){assert.equal(p(value).text,JSON.stringify(value));assert.equal(p(value).type,'string');}
 assert.equal(p(null).text,'Null');assert.equal(p(null).type,'null');assert.equal(p(false).text,'false');assert.equal(p(false).type,'boolean');assert.equal(p(0).text,'0');assert.equal(p(0).type,'number');assert.equal(C.presentation({},field).state,'missing');
});

test('column changes are admitted before mutating layout or projected headers',()=>{
 const C=require('../ui/record-columns.mjs'),oldDocument=global.document;
 class Element{
  constructor(tag){this.tag=tag;this.children=[];this.dataset={};this.listeners=new Map()}
  append(...nodes){for(const child of nodes){child.remove?.();child.parent=this;this.children.push(child)}}
  remove(){if(this.parent){this.parent.children=this.parent.children.filter(child=>child!==this);this.parent=null}}
  addEventListener(name,fn){this.listeners.set(name,fn)}
  querySelectorAll(){return []}
 }
 global.document={createElement:tag=>new Element(tag)};
 try{
  const host=new Element('div'),head=new Element('tr'),table=new Element('table');table.tHead={rows:[head]};
  for(const label of ['UTC','Service','Summary','Context']){const cell=new Element('th');cell.textContent=label;head.append(cell)}
  let allowed=false,changes=0;const control=C.mount(host,table,{canChange:()=>allowed,onChange:()=>changes++}),picker=host.children[0],before=[...head.children],next={columns:['summary','utc'],fieldColumns:[{kind:'attribute',key:'region'}]};
  picker.listeners.get('goo:record-columns')({detail:next});assert.deepEqual(control.layout(),{columns:['utc','service','summary','context'],fieldColumns:[]});assert.deepEqual(head.children,before);assert.equal(changes,0);assert.deepEqual(picker.model.layout,control.layout());
  allowed=true;picker.listeners.get('goo:record-columns')({detail:next});assert.deepEqual(control.layout(),next);assert.equal(changes,1);assert.equal(head.children[0].textContent,'Summary');assert.equal(head.children[1].textContent,'UTC');assert.equal(head.children.at(-1).textContent,'attribute:region');
 }finally{global.document=oldDocument}
});
