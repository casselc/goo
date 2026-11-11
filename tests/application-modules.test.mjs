import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readdir,readFile} from 'node:fs/promises';

// Page bootstraps require their own document. Every other module remains safe
// to import before a host mounts an application owner or supplied display.
const explicitEntries=new Set(['app.mjs','display-examples.mjs']);

test('application owners import without DOM, source reads or namespace publication',async()=>{
  const before=new Set(Object.getOwnPropertyNames(globalThis)),fetch=globalThis.fetch;
  let reads=0;globalThis.fetch=()=>{reads++;throw Error('Owner import must not read sources')};
  try{
    for(const file of (await readdir(new URL('../ui/',import.meta.url))).filter(p=>p.endsWith('.mjs')&&!explicitEntries.has(p))){
      const source=await readFile(new URL('../ui/'+file,import.meta.url),'utf8');
      assert.doesNotMatch(source,/\brequire\s*\(|module\.exports|(?:global|globalThis|window)\.Goo/);
      assert.doesNotMatch(source,/(?:\bfrom\s*|\bimport\s*(?:\(\s*)?)['"](?:\.\/|\/)?display-examples\.mjs['"]/,file+' must not import the standalone reference bootstrap');
      await import('../ui/'+file);
    }
    assert.equal(reads,0);
    assert.deepEqual(Object.getOwnPropertyNames(globalThis).filter(key=>!before.has(key)),[]);
  }finally{globalThis.fetch=fetch}
});

test('HTML uses explicit application/reference entries and the compatibility bridge is gone',async()=>{
  const html=await readFile(new URL('../ui/index.html',import.meta.url),'utf8');
  assert.deepEqual([...html.matchAll(/<script[^>]*src="([^"]+)"/g)].map(m=>m[1]),['/app.mjs']);
  assert.match(html,/<script type="module" src="\/app\.mjs"><\/script>/);
  const reference=await readFile(new URL('../ui/display-examples.html',import.meta.url),'utf8');
  assert.deepEqual([...reference.matchAll(/<script[^>]*src="([^"]+)"/g)].map(m=>m[1]),['/display-examples.mjs']);
  assert.match(reference,/<script type="module" src="\/display-examples\.mjs"><\/script>/);
  const entry=await readFile(new URL('../ui/app.mjs',import.meta.url),'utf8');
  assert.doesNotMatch(entry,/(?:\bfrom\s*|\bimport\s*(?:\(\s*)?)['"](?:\.\/|\/)?display-examples\.mjs['"]/);
  assert.equal((await readdir(new URL('../ui/',import.meta.url))).includes('app-display.mjs'),false);
});

test('investigation exports are stable entrypoints before an editor mounts',async()=>{
 const documents=await import('../ui/documents.mjs');
 assert.equal(documents.context(),null);
 for(const key of ['startRetained','stageRetained','stageFinding','stageQuestion','stageComparison','linkCapture','attach']){
  assert.equal(typeof documents[key],'function');
  assert.throws(()=>documents[key](),/editor is not mounted/);
 }
});

test('application entry uses native examples without the retired Python query surface',async()=>{
  const [entry,html]=await Promise.all(['app.mjs','index.html'].map(file=>readFile(new URL('../ui/'+file,import.meta.url),'utf8')));
  assert.doesNotMatch(entry,/['"]\/api\/(?:query|explain|package)['"]/);
  assert.doesNotMatch(entry,/workflows\.mjs|legacyExamples/);
  assert.doesNotMatch(html,/id="(?:legacyExamples|workflows|view|group|scenario|runExpression|expression|package)"/);
  for(const id of ['minutes','exampleQuestions','guidedMeasurements','prebuilts','runOperations','runHttpPackage']){
    assert.ok(html.includes('id="'+id+'"'),id+' must remain available');
  }
  for(const path of ['/api/example-questions','/api/prebuilts','/api/operations','/api/http-package']){
    assert.ok(entry.includes(path),path+' must remain available');
  }
});
