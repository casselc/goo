import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readdir,readFile} from 'node:fs/promises';

test('application owners import without DOM, source reads or namespace publication',async()=>{
  const before=new Set(Object.getOwnPropertyNames(globalThis)),fetch=globalThis.fetch;
  let reads=0;globalThis.fetch=()=>{reads++;throw Error('Owner import must not read sources')};
  try{
    for(const file of (await readdir(new URL('../ui/',import.meta.url))).filter(p=>p.endsWith('.mjs')&&p!=='app.mjs')){
      const source=await readFile(new URL('../ui/'+file,import.meta.url),'utf8');
      assert.doesNotMatch(source,/\brequire\s*\(|module\.exports|(?:global|globalThis|window)\.Goo/);
      await import('../ui/'+file);
    }
    assert.equal(reads,0);
    assert.deepEqual(Object.getOwnPropertyNames(globalThis).filter(key=>!before.has(key)),[]);
  }finally{globalThis.fetch=fetch}
});

test('HTML uses one explicit application entry and the compatibility bridge is gone',async()=>{
  const html=await readFile(new URL('../ui/index.html',import.meta.url),'utf8');
  assert.deepEqual([...html.matchAll(/<script[^>]*src="([^"]+)"/g)].map(m=>m[1]),['/app.mjs']);
  assert.match(html,/<script type="module" src="\/app\.mjs"><\/script>/);
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
