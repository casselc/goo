async function openFilters(page){const panel=page.locator('#filterComposer');if(!await panel.evaluate(n=>n.open))await panel.locator(':scope > summary').click()}
async function openMeasures(page){const panel=page.locator('#measureComposer');if(!await panel.evaluate(n=>n.open))await panel.locator(':scope > summary').click()}
module.exports={openFilters,openMeasures};
