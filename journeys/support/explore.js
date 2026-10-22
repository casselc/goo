async function openFilters(page){const panel=page.locator('#filterComposer');if(!await panel.evaluate(n=>n.open))await panel.locator(':scope > summary').click()}
async function openMeasures(page){const panel=page.locator('#measureComposer');if(!await panel.evaluate(n=>n.open))await panel.locator(':scope > summary').click()}
// Read each supplied row through the actual local inspector, without querying again.
async function inspectedRecords(page){const rows=page.locator('#exploreRows .record-row'),records=[];for(let i=0;i<await rows.count();i++){await rows.nth(i).getByRole('button',{name:'Inspect record',exact:true}).click();const inspector=page.locator('#exploreRows .ui-inspector');records.push(JSON.parse(await inspector.locator('.raw-record > pre').textContent()));await inspector.getByRole('button',{name:'Close selected item details',exact:true}).click()}return records}
module.exports={openFilters,openMeasures,inspectedRecords};
