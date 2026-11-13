'use strict';
// Count clones, not recursively serialized bytes, inside the timed owner action.
// This hook observes actual Activity code; it never replaces a query or its data.
async function install(page){await page.addInitScript(()=>{
 const original=window.structuredClone.bind(window);let active=null;
 const item=value=>!!value&&typeof value==='object'&&typeof value.key==='string'&&value.record&&value.readWindow;
 window.structuredClone=function(value,...rest){
  if(active){active.cloneCalls++;if(item(value))active.retainedRowsCopied++;else if(Array.isArray(value)&&value.length&&value.every(item))active.retainedRowsCopied+=value.length;else if(Array.isArray(value?.rows)&&value.rows.length&&value.rows.every(item))active.retainedRowsCopied+=value.rows.length;
   if(value&&typeof value==='object'&&(typeof value.Body==='string'||typeof value.SpanName==='string'))active.recordValuesCopied++;
  }
  return original(value,...rest);
 };
 window.__snapshotProfile423={begin(name){if(active)throw Error('Profile already active');active={name,cloneCalls:0,retainedRowsCopied:0,recordValuesCopied:0,start:performance.now()};},end(){if(!active)throw Error('Profile is not active');const out={...active,durationMs:performance.now()-active.start};delete out.start;active=null;return out;}};
 });}
async function click(page,control,name){return control.evaluate((element,name)=>{const p=window.__snapshotProfile423;p.begin(name);try{element.click();return p.end()}catch(error){p.end();throw error}},name)}
async function input(page,control,name,value){return control.evaluate((element,{name,value})=>{const p=window.__snapshotProfile423;p.begin(name);try{if(element.type==='checkbox'){element.checked=value;element.dispatchEvent(new Event('change',{bubbles:true}))}else{element.value=value;element.dispatchEvent(new Event('input',{bubbles:true}))}return p.end()}catch(error){p.end();throw error}},{name,value})}
module.exports={install,click,input};
