'use strict';
const {test,expect}=require('@playwright/test');
const fs=require('node:fs/promises');
const path=require('node:path');
const base=process.env.GOO_DISPLAY_EXAMPLES_BASE||'http://127.0.0.1:8088';

function importedFacts(snapshot,fixture){
  expect(snapshot.id).toMatch(/^goo\.local\/[1-9]\d*$/);
  expect(snapshot.revision).toBe(snapshot.id);
  expect(snapshot.id).not.toBe(fixture.id);
  expect(Object.hasOwn(snapshot,'completeness')).toBe(false);
  const actual=structuredClone(snapshot),expected=structuredClone(fixture);
  delete actual.id;delete actual.revision;delete expected.id;delete expected.revision;
  expect(actual).toEqual(expected);
  return snapshot;
}


test('408 public displays upgrade attributes and pre-registration properties without application effects',async({page})=>{
  const errors=[],requests=[];
  page.on('pageerror',error=>errors.push(error.message));
  page.on('request',request=>requests.push(new URL(request.url()).pathname));
  await page.addInitScript(()=>{
    const attempts=[];Object.defineProperty(window,'__display408Effects',{value:attempts});
    const refuse=operation=>{attempts.push(operation);throw Error('Passive reference attempted '+operation)};
    for(const name of ['fetch','XMLHttpRequest','WebSocket','EventSource'])
      Object.defineProperty(window,name,{configurable:true,value:function(){return refuse(name)}});
    for(const name of ['getItem','setItem','removeItem','clear'])
      Storage.prototype[name]=function(){return refuse('storage '+name)};
    for(const name of ['localStorage','sessionStorage','indexedDB'])
      Object.defineProperty(window,name,{configurable:true,get(){return refuse(name)}});
  });
  let release;
  const gate=new Promise(resolve=>{release=resolve});
  await page.route('**/display.mjs',async route=>{await gate;await route.continue()});
  await page.goto(base+'/display-examples',{waitUntil:'commit'});
  const summary=page.locator('#display408-summary'),ranked=page.locator('#display408-ranked');
  try{
    await expect(ranked).toBeAttached();
    await expect(summary).toHaveText('20 captured records · synthetic supplied summary.');
    expect(await page.evaluate(()=>({chart:customElements.get('goo-chart')===undefined,summary:customElements.get('goo-measure-summary')===undefined,ownSnapshot:Object.hasOwn(document.querySelector('#display408-ranked'),'snapshot'),ownDisplay:Object.hasOwn(document.querySelector('#display408-ranked'),'display')}))).toEqual({chart:true,summary:true,ownSnapshot:true,ownDisplay:true});
    const supplied=await page.evaluate(()=>({summary:JSON.parse(document.querySelector('#display408-summary-data').textContent),groups:JSON.parse(document.querySelector('#display408-ranked-data').textContent)}));
    expect(JSON.parse(await summary.getAttribute('snapshot'))).toEqual(supplied.summary);
    release();
    await expect(page.locator('#display408-status')).toHaveText('Supplied examples ready.');
    await expect(summary.locator('.measure-value')).toHaveText('20 records');
    const table=page.locator('#display408-summary-table');
    await expect(table.locator('tbody tr')).toHaveCount(1);
    const accepted={
      table:importedFacts(await table.evaluate(element=>element.snapshot),supplied.summary),
      summary:importedFacts(await summary.evaluate(element=>element.snapshot),supplied.summary),
      groups:importedFacts(await ranked.evaluate(element=>element.snapshot),supplied.groups),
    };
    expect(new Set(Object.values(accepted).map(snapshot=>snapshot.id)).size).toBe(3);
    expect(await ranked.evaluate(element=>({ownSnapshot:Object.hasOwn(element,'snapshot'),ownDisplay:Object.hasOwn(element,'display'),frozen:Object.isFrozen(element.snapshot.rows[0].values)}))).toEqual({ownSnapshot:false,ownDisplay:false,frozen:true});
    await expect(ranked.locator('.scope-view-bar')).toHaveCount(3);
    await expect(ranked).toContainText('unknown groups');
    await expect(ranked).toContainText('unknown upstream sampling');
    await ranked.locator('.scope-view-bar[data-key="api"]').focus();
    await page.keyboard.press('Enter');
    await expect(page.locator('#display408-status')).toHaveText('Selected api · 12 supplied records.');
    await page.getByRole('combobox',{name:'View',exact:true}).selectOption('table');
    await expect(page.locator('#display408-status')).toHaveText('Table view · same supplied groups.');
    await expect(ranked.locator('tbody tr')).toHaveCount(3);
    expect(await ranked.evaluate(element=>element.snapshot)).toEqual(accepted.groups);
    await ranked.getByRole('button',{name:'Select "api"',exact:true}).click();
    await expect(page.locator('#display408-status')).toHaveText('Selected api · 12 supplied records.');

    // Existing upgraded elements also accept data while detached, without sharing
    // mutable caller objects or acquiring source/coverage authority.
    const detached=await page.evaluate(()=>{
      const source=JSON.parse(document.querySelector('#display408-ranked-data').textContent),original=structuredClone(source),display={kind:'table',category:'category',value:'count',title:'Detached exact table'},element=document.createElement('goo-chart');
      element.id='display408-detached-check';element.snapshot=source;element.display=display;
      source.id='mutated caller';source.rows[0].values.count='999';display.title='mutated caller';
      document.querySelector('main').append(element);
      return {original,snapshot:element.snapshot,display:element.display};
    });
    importedFacts(detached.snapshot,detached.original);
    expect(Object.values(accepted).map(snapshot=>snapshot.id)).not.toContain(detached.snapshot.id);
    expect(detached.display.title).toBe('Detached exact table');
    const extra=page.locator('#display408-detached-check');
    await expect(extra.locator('tbody tr')).toHaveCount(3);
    await extra.evaluate(element=>{element.remove();element.display={kind:'bar',category:'category',value:'count',title:'Reattached supplied groups'};document.querySelector('main').append(element)});
    await expect(extra.locator('.scope-view-bar')).toHaveCount(3);
    expect(await extra.evaluate(element=>element.snapshot)).toEqual(detached.snapshot);
    await extra.evaluate(element=>element.display={...element.display,requiresComplete:true});
    await expect(extra.getByRole('status')).toContainText('Complete');
    await expect(extra.locator('.scope-view-bar')).toHaveCount(0);
    expect(await extra.evaluate(element=>element.snapshot)).toEqual(detached.snapshot);
    await extra.evaluate(element=>element.remove());
    await summary.evaluate(element=>element.setAttribute('display',JSON.stringify({fields:['count'],title:'Updated attribute summary'})));
    await expect(summary).toContainText('Updated attribute summary');
    expect(await summary.evaluate(element=>element.snapshot)).toEqual(accepted.summary);
    await summary.evaluate(element=>element.setAttribute('display',JSON.stringify({fields:['count'],title:'Supplied summary'})));
    await page.getByRole('combobox',{name:'View',exact:true}).selectOption('bar');
    expect(await page.evaluate(()=>Object.keys(globalThis).filter(key=>/^Goo/.test(key)))).toEqual([]);
    expect(await page.evaluate(()=>window.__display408Effects)).toEqual([]);
    expect(requests.filter(url=>url.startsWith('/api/'))).toEqual([]);
    const forbidden=new Set(['/app.mjs','/queries.mjs','/population-workspace.mjs','/exploration.mjs','/source-snapshots.mjs']);
    expect(requests.filter(url=>forbidden.has(url))).toEqual([]);
    expect(errors).toEqual([]);
    await fs.mkdir('docs/journeys/screenshots',{recursive:true});
    await page.screenshot({path:'docs/journeys/screenshots/408-display-embedding.png',fullPage:true});
    await page.setViewportSize({width:390,height:844});
    expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
    await page.screenshot({path:'docs/journeys/screenshots/408-display-embedding-390.png',fullPage:true});
    const directory=path.resolve('artifacts/display408');await fs.mkdir(directory,{recursive:true});
    await fs.writeFile(path.join(directory,'summary.json'),JSON.stringify({version:'goo.display408/1',fixtures:'synthetic supplied data; no source observations',summary:supplied.summary,groups:supplied.groups,accepted,requests,errors,checks:['SSR fallback','pre-registration properties','detached properties','caller mutation isolation','local Bars/Table','keyboard selection intent','attribute update','imported completeness refusal','no application effects','desktop and390px']},null,2));
  }finally{release()}
});
