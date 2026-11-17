// Existing operation workflow now lives within Services.
async function openOperations(page, base) {
  if (base) await page.goto(base + '/#services');
  else {
    const services = page.locator('[data-workspace="services"]');
    if (!await services.isVisible()) await page.getByRole('button', {name:'Menu',exact:true}).click();
    await services.click();
  }
  await page.getByRole('tab', {name:'Operations',exact:true}).click();
}
module.exports = {openOperations};
