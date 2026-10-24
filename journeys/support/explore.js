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
