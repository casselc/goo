// Reusable authoring belongs to Saved; preserve its selected family on return.
async function openDefinitions(page, base) {
  if (base) await page.goto(base + '/#investigations');
  else {
    const saved = page.locator('[data-workspace=investigations]');
    if (!await saved.isVisible()) await page.locator('.entry-menu').click();
    await saved.click();
  }
  await page.locator('#savedDefinitionsTab').click();
}
module.exports = {openDefinitions};
