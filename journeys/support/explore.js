async function openFilters(page){const panel=page.locator('#filterComposer');if(!await panel.evaluate(n=>n.open))await panel.locator(':scope > summary').click()}
async function openMeasures(page){const panel=page.locator('#measureComposer');if(!await panel.evaluate(n=>n.open))await panel.locator(':scope > summary').click()}
// Read each supplied row through the actual local inspector, without querying again.
async function inspectedRecords(page){const rows=page.locator('#exploreRows .record-row'),records=[];for(let i=0;i<await rows.count();i++){await rows.nth(i).getByRole('button',{name:'Inspect record',exact:true}).click();const inspector=page.locator('#exploreRows .ui-inspector');records.push(JSON.parse(await inspector.locator('.raw-record > pre').textContent()));await inspector.getByRole('button',{name:'Close selected item details',exact:true}).click()}return records}
async function addFilter(page,field,op,value=''){
 await openFilters(page);const editor=page.locator('#exploreFilterEditor');await editor.getByRole('button',{name:'Add condition',exact:true}).click();
 const namespace=editor.locator('select[aria-label$=" namespace"]').last(),id=(await namespace.getAttribute('aria-label')).match(/Condition (\d+)/)[1];
 await namespace.selectOption(field.kind);await editor.getByLabel('Condition '+id+' field',{exact:true}).fill(field.key);await editor.getByLabel('Condition '+id+' operator',{exact:true}).selectOption(op);if(!['exists','missing'].includes(op))await editor.getByLabel('Condition '+id+' value',{exact:true}).fill(value);
 return {editor,id};
}
async function clearFilters(page){await openFilters(page);const clear=page.locator('#exploreFilterEditor').getByRole('button',{name:'Clear filters',exact:true});if(await clear.isEnabled())await clear.click()}
async function matchFilters(page,value){await openFilters(page);await page.locator('#exploreFilterEditor').getByRole('combobox',{name:'Group 1 match',exact:true}).selectOption(value)}
async function removeFilter(page,field,op){await openFilters(page);const editor=page.locator('#exploreFilterEditor'),keys=editor.locator('input[aria-label$=" field"]');for(let i=0;i<await keys.count();i++){const key=keys.nth(i),id=(await key.getAttribute('aria-label')).match(/Condition (\d+)/)[1];if(await key.inputValue()===field.key&&await editor.getByLabel('Condition '+id+' namespace',{exact:true}).inputValue()===field.kind&&await editor.getByLabel('Condition '+id+' operator',{exact:true}).inputValue()===op){await editor.getByRole('button',{name:'Remove condition '+id,exact:true}).click();return}}throw Error('Filter not found')}
module.exports={removeFilter,openFilters,openMeasures,inspectedRecords,addFilter,clearFilters,matchFilters};

async function chooseGrouping(page,field){await openFilters(page);const control=page.locator('#exploreFilterEditor goo-grouping-control'),pick=control.getByRole('combobox',{name:'Choose grouping field',exact:true});if((await control.evaluate(el=>el.data.grouping.length))>0)await pick.selectOption('clear');if(!field)return;const value=JSON.stringify(field);if(await pick.locator('option').evaluateAll((xs,v)=>xs.some(x=>x.value===v),value))await pick.selectOption(value);else{const details=control.locator('.grouping-custom');if(!await details.evaluate(e=>e.open))await details.locator('summary').click();await control.getByRole('textbox',{name:'Custom grouping field'}).fill(field.kind+':'+field.key);await control.getByRole('button',{name:'Add field',exact:true}).click()}}
module.exports.chooseGrouping=chooseGrouping;

// Open only the active fixed-question preservation disclosure; other capture surfaces keep their own controls.
async function openPreserve(page,container=page){const panel=container.locator('.question-preserve').filter({visible:true});if(!await panel.evaluate(e=>e.open))await panel.locator(':scope > summary').click()}
module.exports.openPreserve=openPreserve;

// Record definitions stage a fixed question. The caller owns the subsequent explicit Run.
async function stageRecordPackage(page,button=page.locator('#packageUse')){
 const {expect}=require('@playwright/test'),attempts=[],observe=r=>{const path=new URL(r.url()).pathname;if(r.method()==='POST'&&path.startsWith('/api/')&&path!=='/api/telemetry-packages/plan')attempts.push(path)};
 const text=(await button.textContent()).trim();expect(text).toMatch(/^Stage /);const planned=text==='Stage fixed question in Explore'?null:page.waitForResponse(r=>r.request().method()==='POST'&&new URL(r.url()).pathname==='/api/telemetry-packages/plan');page.on('request',observe);
 try{await button.click();if(planned)expect((await planned).ok()).toBe(true);await expect(page.locator('#exploreStatus')).toContainText('Fixed question staged.');await expect(page.locator('#explorePackageScope')).toBeVisible();await page.evaluate(()=>new Promise(resolve=>requestAnimationFrame(()=>requestAnimationFrame(resolve))));expect(attempts).toEqual([])}finally{page.off('request',observe)}
}
module.exports.stageRecordPackage=stageRecordPackage;

async function openRecordPackage(page){const panel=page.locator('summary').filter({hasText:/^Reusable question package$/}).locator('..');if(!await panel.evaluate(e=>e.open))await panel.locator(':scope > summary').click()}
module.exports.openRecordPackage=openRecordPackage;
